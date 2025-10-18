package rules

import (
	"encoding/json"
	"time"
)

// JSONRuleDoc represents a rule in JSON documentation format.
type JSONRuleDoc struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Type        string        `json:"type"`
	Domains     []string      `json:"domains,omitempty"`
	Group       string        `json:"group,omitempty"`
	Metadata    *JSONMetadata `json:"metadata,omitempty"`
	Children    []JSONRuleDoc `json:"children,omitempty"`
	Depth       int           `json:"depth"`
}

// JSONMetadata represents rule metadata in JSON format.
type JSONMetadata struct {
	Owner        string   `json:"owner,omitempty"`
	Version      string   `json:"version,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	RelatedRules []string `json:"relatedRules,omitempty"`
}

// JSONDocumentation represents the complete JSON documentation structure.
type JSONDocumentation struct {
	Title         string                   `json:"title"`
	Description   string                   `json:"description,omitempty"`
	GeneratedAt   string                   `json:"generatedAt"`
	Version       string                   `json:"version,omitempty"`
	Domains       []string                 `json:"domains,omitempty"`
	Groups        []string                 `json:"groups,omitempty"`
	Rules         []JSONRuleDoc            `json:"rules"`
	RulesByDomain map[string][]JSONRuleDoc `json:"rulesByDomain,omitempty"`
	RulesByGroup  map[string][]JSONRuleDoc `json:"rulesByGroup,omitempty"`
}

// GenerateJSON generates JSON documentation for all registered rules.
func GenerateJSON(opts DocumentOptions) (string, error) {
	return GenerateJSONFromRules(AllRules(), opts)
}

// GenerateJSONFromRules generates JSON documentation from a list of registered rules.
func GenerateJSONFromRules(rules []RegisteredRule, opts DocumentOptions) (string, error) {
	// Filter rules based on options
	filtered := filterRegisteredRules(rules, opts)

	// Build the documentation structure
	doc := &JSONDocumentation{
		Title:       opts.Title,
		Description: opts.Description,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     "1.0.0",
	}

	if doc.Title == "" {
		doc.Title = "Business Rules Documentation"
	}

	// Collect domains and groups
	domainSet := make(map[string]bool)
	groupSet := make(map[string]bool)

	for _, rule := range filtered {
		for _, domain := range rule.Domains {
			domainSet[string(domain)] = true
		}
		if rule.Group != "" {
			groupSet[rule.Group] = true
		}
	}

	// Convert to sorted slices
	for domain := range domainSet {
		doc.Domains = append(doc.Domains, domain)
	}
	for group := range groupSet {
		doc.Groups = append(doc.Groups, group)
	}

	// Generate rule documentation
	for _, regRule := range filtered {
		ruleDoc := buildJSONRuleDoc(regRule, opts)
		doc.Rules = append(doc.Rules, ruleDoc)
	}

	// Group by domain if requested
	if opts.GroupByDomain {
		doc.RulesByDomain = make(map[string][]JSONRuleDoc)
		doc.RulesByGroup = make(map[string][]JSONRuleDoc)

		for _, regRule := range filtered {
			ruleDoc := buildJSONRuleDoc(regRule, opts)

			// Add to domain groups
			for _, domain := range regRule.Domains {
				domainStr := string(domain)
				doc.RulesByDomain[domainStr] = append(doc.RulesByDomain[domainStr], ruleDoc)
			}

			// Add to group
			if regRule.Group != "" {
				doc.RulesByGroup[regRule.Group] = append(doc.RulesByGroup[regRule.Group], ruleDoc)
			} else if len(regRule.Domains) > 0 {
				// Use primary domain as group if no explicit group
				domainStr := string(regRule.Domains[0])
				doc.RulesByGroup[domainStr] = append(doc.RulesByGroup[domainStr], ruleDoc)
			}
		}
	}

	// Marshal to JSON with indentation
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GenerateDomainJSON generates JSON documentation for a specific domain.
func GenerateDomainJSON(domain Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomain(domain)
	opts.Title = string(domain) + " Domain Rules"
	return GenerateJSONFromRules(rules, opts)
}

// GenerateDomainsJSON generates JSON documentation for multiple domains.
func GenerateDomainsJSON(domains []Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomains(domains...)
	if opts.Title == "" {
		opts.Title = "Multi-Domain Rules"
	}
	return GenerateJSONFromRules(rules, opts)
}

// GenerateGroupJSON generates JSON documentation for a specific group.
func GenerateGroupJSON(groupName string, opts DocumentOptions) (string, error) {
	rules := RulesByGroup(groupName)
	opts.Title = groupName + " Rules"
	return GenerateJSONFromRules(rules, opts)
}

// buildJSONRuleDoc builds a JSON rule documentation structure.
func buildJSONRuleDoc(regRule RegisteredRule, opts DocumentOptions) JSONRuleDoc {
	// Build the rule tree
	node := buildRuleTree(regRule.Rule, &regRule, 0, opts.MaxDepth)

	ruleDoc := JSONRuleDoc{
		Name:        node.Name,
		Description: node.Description,
		Type:        node.Type.String(),
		Depth:       node.Depth,
	}

	// Add domains
	for _, domain := range node.Domains {
		ruleDoc.Domains = append(ruleDoc.Domains, string(domain))
	}

	// Add group
	ruleDoc.Group = node.Group

	// Add metadata if requested
	if opts.IncludeMetadata && node.Metadata != nil {
		ruleDoc.Metadata = buildJSONMetadata(node.Metadata)
	}

	// Add children recursively
	for _, child := range node.Children {
		childDoc := buildJSONRuleDocFromNode(child, opts)
		ruleDoc.Children = append(ruleDoc.Children, childDoc)
	}

	return ruleDoc
}

// buildJSONRuleDocFromNode builds JSON documentation from a rule node.
func buildJSONRuleDocFromNode(node *ruleNode, opts DocumentOptions) JSONRuleDoc {
	ruleDoc := JSONRuleDoc{
		Name:        node.Name,
		Description: node.Description,
		Type:        node.Type.String(),
		Depth:       node.Depth,
	}

	// Add domains
	for _, domain := range node.Domains {
		ruleDoc.Domains = append(ruleDoc.Domains, string(domain))
	}

	// Add group
	ruleDoc.Group = node.Group

	// Add metadata if requested
	if opts.IncludeMetadata && node.Metadata != nil {
		ruleDoc.Metadata = buildJSONMetadata(node.Metadata)
	}

	// Add children recursively
	for _, child := range node.Children {
		childDoc := buildJSONRuleDocFromNode(child, opts)
		ruleDoc.Children = append(ruleDoc.Children, childDoc)
	}

	return ruleDoc
}

// buildJSONMetadata converts RuleMetadata to JSONMetadata.
func buildJSONMetadata(metadata *RuleMetadata) *JSONMetadata {
	jsonMeta := &JSONMetadata{
		Owner:   metadata.Owner,
		Version: metadata.Version,
		Tags:    metadata.Tags,
	}

	if !metadata.CreatedAt.IsZero() {
		jsonMeta.CreatedAt = metadata.CreatedAt.Format(time.RFC3339)
	}

	if !metadata.UpdatedAt.IsZero() {
		jsonMeta.UpdatedAt = metadata.UpdatedAt.Format(time.RFC3339)
	}

	for _, dep := range metadata.Dependencies {
		jsonMeta.Dependencies = append(jsonMeta.Dependencies, string(dep))
	}

	jsonMeta.RelatedRules = metadata.RelatedRules

	return jsonMeta
}
