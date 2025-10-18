package rules

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateJSON_Empty(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	opts := DocumentOptions{
		Format: FormatJSON,
		Title:  "Empty JSON Documentation",
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON to verify it's valid
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if doc.Title != "Empty JSON Documentation" {
		t.Errorf("Title = %q, want %q", doc.Title, "Empty JSON Documentation")
	}

	if len(doc.Rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(doc.Rules))
	}
}

func TestGenerateJSON_SingleRule(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a simple rule
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})
	WithDescription(rule, "This is a test rule")

	opts := DocumentOptions{
		Format: FormatJSON,
		Title:  "Test JSON Documentation",
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if doc.Title != "Test JSON Documentation" {
		t.Errorf("Title = %q, want %q", doc.Title, "Test JSON Documentation")
	}

	if len(doc.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(doc.Rules))
	}

	rule0 := doc.Rules[0]
	if rule0.Name != "test rule" {
		t.Errorf("Rule name = %q, want %q", rule0.Name, "test rule")
	}

	if rule0.Description != "This is a test rule" {
		t.Errorf("Rule description = %q, want %q", rule0.Description, "This is a test rule")
	}

	if rule0.Type != "SIMPLE" {
		t.Errorf("Rule type = %q, want %q", rule0.Type, "SIMPLE")
	}

	if len(rule0.Domains) != 1 || rule0.Domains[0] != string(TestOrderDomain) {
		t.Errorf("Rule domains = %v, want [%s]", rule0.Domains, TestOrderDomain)
	}
}

func TestGenerateJSON_WithMetadata(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule with metadata
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	metadata := RuleMetadata{
		Owner:     "test-team",
		Version:   "1.0.0",
		Tags:      []string{"validation", "financial"},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	_ = UpdateMetadata(rule, metadata)

	opts := DocumentOptions{
		Format:          FormatJSON,
		IncludeMetadata: true,
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(doc.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(doc.Rules))
	}

	rule0 := doc.Rules[0]
	if rule0.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}

	if rule0.Metadata.Owner != "test-team" {
		t.Errorf("Metadata owner = %q, want %q", rule0.Metadata.Owner, "test-team")
	}

	if rule0.Metadata.Version != "1.0.0" {
		t.Errorf("Metadata version = %q, want %q", rule0.Metadata.Version, "1.0.0")
	}

	if len(rule0.Metadata.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(rule0.Metadata.Tags))
	}

	if rule0.Metadata.CreatedAt == "" {
		t.Error("Expected CreatedAt to be set")
	}

	if rule0.Metadata.UpdatedAt == "" {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestGenerateJSON_HierarchicalRules(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create hierarchical rules
	rule1 := NewWithDomain("amount check", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})
	WithDescription(rule1, "Check minimum amount")

	rule2 := NewWithDomain("country check", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Country == "US", nil
	})
	WithDescription(rule2, "Check valid country")

	andRule := And("order validation", rule1, rule2)
	WithDescription(andRule, "Validate order requirements")

	opts := DocumentOptions{
		Format: FormatJSON,
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Find the AND rule
	var andRuleDoc *JSONRuleDoc
	for i, r := range doc.Rules {
		if r.Name == "order validation" {
			andRuleDoc = &doc.Rules[i]
			break
		}
	}

	if andRuleDoc == nil {
		t.Fatal("AND rule not found in documentation")
	}

	if andRuleDoc.Type != "AND" {
		t.Errorf("Rule type = %q, want %q", andRuleDoc.Type, "AND")
	}

	if len(andRuleDoc.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(andRuleDoc.Children))
	}

	// Verify children
	childNames := make(map[string]bool)
	for _, child := range andRuleDoc.Children {
		childNames[child.Name] = true
	}

	if !childNames["amount check"] || !childNames["country check"] {
		t.Error("Expected children 'amount check' and 'country check'")
	}
}

func TestGenerateJSON_GroupedByDomain(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	orderRule := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})
	WithDescription(orderRule, "Order domain rule")

	userRule := NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})
	WithDescription(userRule, "User domain rule")

	opts := DocumentOptions{
		Format:        FormatJSON,
		GroupByDomain: true,
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check domains
	if len(doc.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(doc.Domains))
	}

	// Check grouped rules
	if doc.RulesByDomain == nil {
		t.Fatal("Expected RulesByDomain to be populated")
	}

	orderDomainRules, ok := doc.RulesByDomain[string(TestOrderDomain)]
	if !ok || len(orderDomainRules) != 1 {
		t.Errorf("Expected 1 rule in order domain, got %d", len(orderDomainRules))
	}

	userDomainRules, ok := doc.RulesByDomain[string(TestUserDomain)]
	if !ok || len(userDomainRules) != 1 {
		t.Errorf("Expected 1 rule in user domain, got %d", len(userDomainRules))
	}
}

func TestGenerateJSON_WithGroup(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create cross-domain rule with group
	rule := NewWithGroup(
		"eligibility rule",
		"Order Eligibility",
		[]Domain{TestOrderDomain, TestUserDomain},
		func(ctx TestContext) (bool, error) {
			return true, nil
		},
	)
	WithDescription(rule, "Check if order is eligible")

	opts := DocumentOptions{
		Format:        FormatJSON,
		GroupByDomain: true,
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check groups
	if len(doc.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(doc.Groups))
	}

	if doc.Groups[0] != "Order Eligibility" {
		t.Errorf("Group = %q, want %q", doc.Groups[0], "Order Eligibility")
	}

	// Check grouped rules
	if doc.RulesByGroup == nil {
		t.Fatal("Expected RulesByGroup to be populated")
	}

	groupRules, ok := doc.RulesByGroup["Order Eligibility"]
	if !ok || len(groupRules) != 1 {
		t.Errorf("Expected 1 rule in Order Eligibility group, got %d", len(groupRules))
	}

	if groupRules[0].Group != "Order Eligibility" {
		t.Errorf("Rule group = %q, want %q", groupRules[0].Group, "Order Eligibility")
	}
}

func TestGenerateJSON_DomainFilter(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	_ = NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("payment rule", TestPaymentDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	// Filter to only order domain
	opts := DocumentOptions{
		Format:         FormatJSON,
		IncludeDomains: []Domain{TestOrderDomain},
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only have 1 rule
	if len(doc.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(doc.Rules))
	}

	if doc.Rules[0].Name != "order rule" {
		t.Errorf("Rule name = %q, want %q", doc.Rules[0].Name, "order rule")
	}
}

func TestGenerateJSON_ExcludeDomain(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	_ = NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	// Exclude user domain
	opts := DocumentOptions{
		Format:         FormatJSON,
		ExcludeDomains: []Domain{TestUserDomain},
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should only have order rule
	if len(doc.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(doc.Rules))
	}

	if doc.Rules[0].Name != "order rule" {
		t.Errorf("Rule name = %q, want %q", doc.Rules[0].Name, "order rule")
	}
}

func TestGenerateDomainJSON(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	_ = NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatJSON,
	}

	jsonStr, err := GenerateDomainJSON(TestOrderDomain, opts)
	if err != nil {
		t.Fatalf("GenerateDomainJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check title
	if !strings.Contains(doc.Title, "order") {
		t.Errorf("Title should contain 'order', got %q", doc.Title)
	}

	// Should only have order rule
	if len(doc.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(doc.Rules))
	}

	if doc.Rules[0].Name != "order rule" {
		t.Errorf("Rule name = %q, want %q", doc.Rules[0].Name, "order rule")
	}
}

func TestGenerateGroupJSON(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create cross-domain rule with group
	_ = NewWithGroup(
		"eligibility rule",
		"Order Eligibility",
		[]Domain{TestOrderDomain, TestUserDomain},
		func(ctx TestContext) (bool, error) {
			return true, nil
		},
	)

	opts := DocumentOptions{
		Format: FormatJSON,
	}

	jsonStr, err := GenerateGroupJSON("Order Eligibility", opts)
	if err != nil {
		t.Fatalf("GenerateGroupJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check title
	if !strings.Contains(doc.Title, "Order Eligibility") {
		t.Errorf("Title should contain 'Order Eligibility', got %q", doc.Title)
	}

	// Should have the eligibility rule
	if len(doc.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(doc.Rules))
	}

	if doc.Rules[0].Name != "eligibility rule" {
		t.Errorf("Rule name = %q, want %q", doc.Rules[0].Name, "eligibility rule")
	}
}

func TestGenerateJSON_MaxDepth(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create deeply nested rules
	rule1 := NewWithDomain("check 1", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("check 2", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	orRule := Or("level 2", rule1, rule2)
	_ = And("level 1", orRule)

	// Limit depth to 1
	opts := DocumentOptions{
		Format:   FormatJSON,
		MaxDepth: 1,
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Find level 1 rule
	var level1Rule *JSONRuleDoc
	for i, r := range doc.Rules {
		if r.Name == "level 1" {
			level1Rule = &doc.Rules[i]
			break
		}
	}

	if level1Rule != nil {
		// Should not have children due to MaxDepth
		if len(level1Rule.Children) > 0 {
			t.Errorf("Expected no children with MaxDepth=1, got %d", len(level1Rule.Children))
		}
	}
}

func TestGenerateDomainsJSON(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	_ = NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("payment rule", TestPaymentDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatJSON,
	}

	// Generate for order and user domains only
	jsonStr, err := GenerateDomainsJSON([]Domain{TestOrderDomain, TestUserDomain}, opts)
	if err != nil {
		t.Fatalf("GenerateDomainsJSON() error = %v", err)
	}

	// Parse JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should have 2 rules
	if len(doc.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(doc.Rules))
	}

	// Check that payment rule is not included
	for _, rule := range doc.Rules {
		if rule.Name == "payment rule" {
			t.Error("Payment rule should not be included")
		}
	}
}

func TestGenerateJSON_ValidFormat(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a few rules
	_ = NewWithDomain("rule 1", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("rule 2", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format:        FormatJSON,
		GroupByDomain: true,
		Title:         "Valid JSON Test",
		Description:   "Testing JSON format validity",
	}

	jsonStr, err := GenerateJSON(opts)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var doc JSONDocumentation
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify required fields
	if doc.Title == "" {
		t.Error("Title should not be empty")
	}

	if doc.GeneratedAt == "" {
		t.Error("GeneratedAt should not be empty")
	}

	if doc.Version == "" {
		t.Error("Version should not be empty")
	}
}
