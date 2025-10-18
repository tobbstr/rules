package rules

import (
	"reflect"
	"sync"
	"time"
)

// Domain represents a business domain (strongly typed).
type Domain string

// String returns the domain as a string.
func (d Domain) String() string {
	return string(d)
}

// RegisteredRule wraps a rule with its registration metadata.
type RegisteredRule struct {
	// Rule is the actual rule (type-erased)
	Rule any

	// Domains lists which domain(s) this rule belongs to
	Domains []Domain

	// Group is the human-readable group name (optional)
	Group string

	// Description is the documentation description (optional)
	// This is the single source of truth for rule descriptions
	Description string

	// Metadata provides additional rule information
	Metadata *RuleMetadata

	// RegisteredAt tracks when the rule was registered
	RegisteredAt time.Time
}

// RuleMetadata provides additional rule information.
type RuleMetadata struct {
	// Tags categorize the rule (beyond domain)
	Tags []string

	// Owner identifies who maintains the rule
	Owner string

	// Version tracks rule version
	Version string

	// CreatedAt tracks when the rule was created
	CreatedAt time.Time

	// UpdatedAt tracks last modification
	UpdatedAt time.Time

	// RelatedRules links to related rules
	RelatedRules []string

	// Dependencies lists other domains this rule depends on
	Dependencies []Domain
}

// Registry manages rule registration and retrieval.
type Registry interface {
	// Register adds a rule to the registry with domain tags and/or group
	Register(rule any, opts ...RegistrationOption) error

	// AllRules returns all registered rules
	AllRules() []RegisteredRule

	// RulesByDomain returns rules for a specific domain
	RulesByDomain(domain Domain) []RegisteredRule

	// RulesByDomains returns rules matching any of the specified domains
	RulesByDomains(domains ...Domain) []RegisteredRule

	// RulesByGroup returns rules belonging to a named group
	RulesByGroup(groupName string) []RegisteredRule

	// Domains returns all registered domain names
	Domains() []Domain

	// Groups returns all registered group names
	Groups() []string

	// UpdateDescription updates the description for an already registered rule
	UpdateDescription(rule any, description string) error

	// UpdateMetadata updates or adds metadata to an already registered rule
	UpdateMetadata(rule any, metadata RuleMetadata) error

	// GetDescription retrieves the description for a rule
	GetDescription(rule any) string

	// Clear removes all registered rules (useful for testing)
	Clear()
}

// registrationConfig holds configuration for rule registration.
type registrationConfig struct {
	domains     []Domain
	group       string
	description string
	metadata    *RuleMetadata
}

// RegistrationOption configures rule registration.
type RegistrationOption func(*registrationConfig)

// WithDomain tags the rule with a single domain.
func WithDomain(d Domain) RegistrationOption {
	return func(c *registrationConfig) {
		c.domains = append(c.domains, d)
	}
}

// WithDomains tags the rule with multiple domains (no group name).
func WithDomains(domains ...Domain) RegistrationOption {
	return func(c *registrationConfig) {
		c.domains = append(c.domains, domains...)
	}
}

// WithGroup tags the rule with a meaningful name and multiple domains.
// This is the preferred method for cross-domain rules.
func WithGroup(name string, domains ...Domain) RegistrationOption {
	return func(c *registrationConfig) {
		c.group = name
		c.domains = append(c.domains, domains...)
	}
}

// WithRegistrationMetadata attaches metadata to the rule during registration.
func WithRegistrationMetadata(metadata RuleMetadata) RegistrationOption {
	return func(c *registrationConfig) {
		c.metadata = &metadata
	}
}

// WithRegistrationDescription sets the description during registration.
func WithRegistrationDescription(description string) RegistrationOption {
	return func(c *registrationConfig) {
		c.description = description
	}
}

// defaultRegistry is a thread-safe registry implementation.
type defaultRegistry struct {
	mu    sync.RWMutex
	rules map[uintptr]*RegisteredRule // pointer address as key
}

// NewRegistry creates a new registry instance.
func NewRegistry() Registry {
	return &defaultRegistry{
		rules: make(map[uintptr]*RegisteredRule),
	}
}

// Register adds a rule to the registry.
func (r *defaultRegistry) Register(rule any, opts ...RegistrationOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	config := &registrationConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Deduplicate domains
	domains := deduplicateDomains(config.domains)

	// Get pointer address for lookup
	ptr := getRulePointer(rule)

	// Check if already registered
	if existing, ok := r.rules[ptr]; ok {
		// Update existing registration
		if len(domains) > 0 {
			existing.Domains = domains
		}
		if config.group != "" {
			existing.Group = config.group
		}
		if config.description != "" {
			existing.Description = config.description
		}
		if config.metadata != nil {
			existing.Metadata = config.metadata
		}
		return nil
	}

	// Create new registration
	registered := &RegisteredRule{
		Rule:         rule,
		Domains:      domains,
		Group:        config.group,
		Description:  config.description,
		Metadata:     config.metadata,
		RegisteredAt: time.Now(),
	}

	r.rules[ptr] = registered
	return nil
}

// AllRules returns all registered rules.
func (r *defaultRegistry) AllRules() []RegisteredRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]RegisteredRule, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, *rule)
	}
	return result
}

// RulesByDomain returns rules for a specific domain.
func (r *defaultRegistry) RulesByDomain(domain Domain) []RegisteredRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []RegisteredRule
	for _, rule := range r.rules {
		for _, d := range rule.Domains {
			if d == domain {
				result = append(result, *rule)
				break
			}
		}
	}
	return result
}

// RulesByDomains returns rules matching any of the specified domains.
func (r *defaultRegistry) RulesByDomains(domains ...Domain) []RegisteredRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domainSet := make(map[Domain]bool)
	for _, d := range domains {
		domainSet[d] = true
	}

	var result []RegisteredRule
	for _, rule := range r.rules {
		for _, d := range rule.Domains {
			if domainSet[d] {
				result = append(result, *rule)
				break
			}
		}
	}
	return result
}

// RulesByGroup returns rules belonging to a named group.
func (r *defaultRegistry) RulesByGroup(groupName string) []RegisteredRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []RegisteredRule
	for _, rule := range r.rules {
		if rule.Group == groupName {
			result = append(result, *rule)
		}
	}
	return result
}

// Domains returns all registered domain names.
func (r *defaultRegistry) Domains() []Domain {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domainSet := make(map[Domain]bool)
	for _, rule := range r.rules {
		for _, d := range rule.Domains {
			domainSet[d] = true
		}
	}

	result := make([]Domain, 0, len(domainSet))
	for d := range domainSet {
		result = append(result, d)
	}
	return result
}

// Groups returns all registered group names.
func (r *defaultRegistry) Groups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groupSet := make(map[string]bool)
	for _, rule := range r.rules {
		if rule.Group != "" {
			groupSet[rule.Group] = true
		}
	}

	result := make([]string, 0, len(groupSet))
	for g := range groupSet {
		result = append(result, g)
	}
	return result
}

// UpdateDescription updates the description for an already registered rule.
func (r *defaultRegistry) UpdateDescription(rule any, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ptr := getRulePointer(rule)
	if registered, ok := r.rules[ptr]; ok {
		registered.Description = description
		return nil
	}

	return nil // Silently succeed if not found
}

// UpdateMetadata updates or adds metadata to an already registered rule.
func (r *defaultRegistry) UpdateMetadata(rule any, metadata RuleMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ptr := getRulePointer(rule)
	if registered, ok := r.rules[ptr]; ok {
		registered.Metadata = &metadata
		return nil
	}

	return nil // Silently succeed if not found
}

// GetDescription retrieves the description for a rule.
func (r *defaultRegistry) GetDescription(rule any) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ptr := getRulePointer(rule)
	if registered, ok := r.rules[ptr]; ok {
		return registered.Description
	}

	return ""
}

// Clear removes all registered rules.
func (r *defaultRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make(map[uintptr]*RegisteredRule)
}

// DefaultRegistry is the global registry instance.
var DefaultRegistry = NewRegistry()

// Register adds a rule to the default registry.
func Register(rule any, opts ...RegistrationOption) error {
	return DefaultRegistry.Register(rule, opts...)
}

// AllRules returns all rules from the default registry.
func AllRules() []RegisteredRule {
	return DefaultRegistry.AllRules()
}

// RulesByDomain returns rules for a domain from the default registry.
func RulesByDomain(domain Domain) []RegisteredRule {
	return DefaultRegistry.RulesByDomain(domain)
}

// RulesByDomains returns rules for domains from the default registry.
func RulesByDomains(domains ...Domain) []RegisteredRule {
	return DefaultRegistry.RulesByDomains(domains...)
}

// RulesByGroup returns rules for a group from the default registry.
func RulesByGroup(groupName string) []RegisteredRule {
	return DefaultRegistry.RulesByGroup(groupName)
}

// UpdateDescription updates the description in the default registry.
func UpdateDescription(rule any, description string) error {
	return DefaultRegistry.UpdateDescription(rule, description)
}

// UpdateMetadata updates metadata in the default registry.
func UpdateMetadata(rule any, metadata RuleMetadata) error {
	return DefaultRegistry.UpdateMetadata(rule, metadata)
}

// GetDescription retrieves description from the default registry.
func GetDescription(rule any) string {
	return DefaultRegistry.GetDescription(rule)
}

// Helper functions

// getRulePointer gets the pointer address of a rule for lookup.
func getRulePointer(rule any) uintptr {
	// Use reflect to get the pointer value
	// This works because rules are typically package-level variables
	v := reflect.ValueOf(rule)
	if v.Kind() == reflect.Ptr {
		return v.Pointer()
	}
	// For interface values, get the underlying pointer
	return v.Pointer()
}

// deduplicateDomains removes duplicate domains from a slice.
func deduplicateDomains(domains []Domain) []Domain {
	if len(domains) == 0 {
		return domains
	}

	seen := make(map[Domain]bool)
	result := make([]Domain, 0, len(domains))

	for _, d := range domains {
		if !seen[d] {
			seen[d] = true
			result = append(result, d)
		}
	}

	return result
}
