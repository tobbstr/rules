package rules

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGenerateHTML_Empty(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	opts := DocumentOptions{
		Format: FormatHTML,
		Title:  "Empty Documentation",
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for expected HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}

	if !strings.Contains(html, "Empty Documentation") {
		t.Error("HTML should contain title")
	}

	if !strings.Contains(html, "<style>") {
		t.Error("HTML should contain embedded CSS")
	}

	if !strings.Contains(html, "<script>") {
		t.Error("HTML should contain embedded JavaScript")
	}
}

func TestGenerateHTML_SingleRule(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a simple rule
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})
	WithDescription(rule, "This is a test rule")

	opts := DocumentOptions{
		Format: FormatHTML,
		Title:  "Test Documentation",
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for expected content
	expectedContent := []string{
		"Test Documentation",
		"test rule",
		"SIMPLE",
		"This is a test rule",
		string(TestOrderDomain),
		"rule-card",
		"type-badge",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain %q", expected)
		}
	}
}

func TestGenerateHTML_WithMetadata(t *testing.T) {
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
		Format:          FormatHTML,
		IncludeMetadata: true,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for metadata
	expectedContent := []string{
		"test-team",
		"1.0.0",
		"validation",
		"financial",
		"2024-01-01",
		"2024-01-15",
		"metadata-item",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain metadata %q", expected)
		}
	}
}

func TestGenerateHTML_HierarchicalRules(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create hierarchical rules
	rule1 := NewWithDomain("amount >= 100", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})
	WithDescription(rule1, "Amount must be at least 100")

	rule2 := NewWithDomain("country is US", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Country == "US", nil
	})
	WithDescription(rule2, "Order country must be US")

	andRule := And("combined order rules", rule1, rule2)
	_ = Register(andRule, WithDomain(TestOrderDomain), WithGroup("Order Validation"))
	WithDescription(andRule, "Order must meet all requirements")

	opts := DocumentOptions{
		Format: FormatHTML,
		Title:  "Hierarchical Rules",
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for hierarchical structure
	expectedContent := []string{
		"Order must meet all requirements",
		"Amount must be at least 100",
		"Order country must be US",
		"type-and",
		"rule-children",
		"child-rule",
		"All of these conditions must be satisfied",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain %q", expected)
		}
	}
}

func TestGenerateHTML_DomainGrouping(t *testing.T) {
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
		Format:        FormatHTML,
		GroupByDomain: true,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for domain sections
	expectedContent := []string{
		string(TestOrderDomain),
		string(TestUserDomain),
		"domain-badge",
		"rule-group",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain domain content %q", expected)
		}
	}
}

func TestGenerateHTML_GroupByGroup(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules with explicit groups
	_ = NewWithGroup("validation rule", "Validation Rules", []Domain{TestOrderDomain}, func(o TestOrder) (bool, error) {
		return o.Amount >= 0, nil
	})

	_ = NewWithGroup("security rule", "Security Rules", []Domain{TestOrderDomain}, func(o TestOrder) (bool, error) {
		return o.Country != "", nil
	})

	opts := DocumentOptions{
		Format:        FormatHTML,
		GroupByDomain: true, // This also groups by group
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for groups
	expectedContent := []string{
		"Validation Rules",
		"Security Rules",
		"validation rule",
		"security rule",
		"rule-group",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain group content %q", expected)
		}
	}
}

func TestGenerateHTML_Sidebar(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules
	_ = NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithGroup("grouped rule", "Test Group", []Domain{TestOrderDomain}, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for sidebar elements
	expectedContent := []string{
		"sidebar",
		"search-box",
		"Domains",
		"Groups",
		string(TestOrderDomain),
		"Test Group",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML sidebar should contain %q", expected)
		}
	}
}

func TestGenerateHTML_JavaScriptFunctions(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule
	_ = NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for JavaScript functions
	expectedContent := []string{
		"toggleRule",
		"searchBox",
		"addEventListener",
		"scrollIntoView",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain JavaScript function %q", expected)
		}
	}
}

func TestGenerateHTML_DomainFiltering(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})
	WithDescription(rule1, "Order rule description")

	rule2 := NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})
	WithDescription(rule2, "User rule description")

	// Filter to only include TestOrderDomain
	opts := DocumentOptions{
		Format:         FormatHTML,
		IncludeDomains: []Domain{TestOrderDomain},
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Should contain order rule
	if !strings.Contains(html, "Order rule description") {
		t.Error("HTML should contain order rule")
	}

	// Should NOT contain user rule
	if strings.Contains(html, "User rule description") {
		t.Error("HTML should not contain user rule")
	}
}

func TestGenerateHTML_DomainExclusion(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})
	WithDescription(rule1, "Order rule description")

	rule2 := NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})
	WithDescription(rule2, "User rule description")

	// Exclude TestUserDomain
	opts := DocumentOptions{
		Format:         FormatHTML,
		ExcludeDomains: []Domain{TestUserDomain},
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Should contain order rule
	if !strings.Contains(html, "Order rule description") {
		t.Error("HTML should contain order rule")
	}

	// Should NOT contain user rule
	if strings.Contains(html, "User rule description") {
		t.Error("HTML should not contain user rule")
	}
}

func TestGenerateHTML_MaxDepth(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create deeply nested rules (use New to avoid auto-registration for level3)
	level3 := New("level 3", func(o TestOrder) (bool, error) {
		return true, nil
	})

	level2 := Not("not level 3", level3)
	_ = Register(level2, WithDomain(TestOrderDomain))
	WithDescription(level2, "Level 2 rule")

	level1 := And("and level 2", level2)
	_ = Register(level1, WithDomain(TestOrderDomain))
	WithDescription(level1, "Level 1 rule")

	// Limit depth to 2 (0=root, 1=level2, stop before level3)
	opts := DocumentOptions{
		Format:   FormatHTML,
		MaxDepth: 2,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Should contain level 1 and 2
	if !strings.Contains(html, "Level 1 rule") {
		t.Error("HTML should contain Level 1 rule")
	}
	if !strings.Contains(html, "Level 2 rule") {
		t.Error("HTML should contain Level 2 rule")
	}

	// Should NOT contain level 3 (beyond max depth)
	if strings.Contains(html, "Level 3 rule") {
		t.Error("HTML should not contain Level 3 rule (beyond MaxDepth)")
	}
}

func TestGenerateHTML_ResponsiveCSS(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for responsive CSS
	expectedContent := []string{
		"@media",
		"max-width",
		"flex",
		"grid",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain responsive CSS %q", expected)
		}
	}
}

func TestGenerateHTML_TypeBadges(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules of different types
	simple := NewWithDomain("simple", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	andRule := And("and rule", simple)
	_ = Register(andRule, WithDomain(TestOrderDomain))

	orRule := Or("or rule", simple)
	_ = Register(orRule, WithDomain(TestOrderDomain))

	notRule := Not("not rule", simple)
	_ = Register(notRule, WithDomain(TestOrderDomain))

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for type badges
	expectedContent := []string{
		"type-simple",
		"type-and",
		"type-or",
		"type-not",
		"type-badge",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain type badge %q", expected)
		}
	}
}

func TestGenerateDomainHTML(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateDomainHTML(TestOrderDomain, opts)
	if err != nil {
		t.Fatalf("GenerateDomainHTML() error = %v", err)
	}

	// Should contain order domain in title
	if !strings.Contains(html, string(TestOrderDomain)+" Domain Rules") {
		t.Error("HTML should contain domain in title")
	}

	// Should contain order rule
	ruleName := getRuleName(rule1)
	if !strings.Contains(html, ruleName) {
		t.Error("HTML should contain order rule")
	}
}

func TestGenerateDomainsHTML(t *testing.T) {
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
		Format: FormatHTML,
	}

	html, err := GenerateDomainsHTML([]Domain{TestOrderDomain, TestUserDomain}, opts)
	if err != nil {
		t.Fatalf("GenerateDomainsHTML() error = %v", err)
	}

	// Should contain both domains
	expectedContent := []string{
		string(TestOrderDomain),
		string(TestUserDomain),
		"order rule",
		"user rule",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain %q", expected)
		}
	}
}

func TestGenerateGroupHTML(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different groups
	rule1 := NewWithGroup("validation rule", "Validation", []Domain{TestOrderDomain}, func(o TestOrder) (bool, error) {
		return true, nil
	})

	_ = NewWithGroup("security rule", "Security", []Domain{TestOrderDomain}, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateGroupHTML("Validation", opts)
	if err != nil {
		t.Fatalf("GenerateGroupHTML() error = %v", err)
	}

	// Should contain group in title
	if !strings.Contains(html, "Validation Rules") {
		t.Error("HTML should contain group in title")
	}

	// Should contain validation rule
	ruleName := getRuleName(rule1)
	if !strings.Contains(html, ruleName) {
		t.Error("HTML should contain validation rule")
	}

	// Should NOT contain security rule
	if strings.Contains(html, "security rule") {
		t.Error("HTML should not contain security rule")
	}
}

func TestGenerateHTML_CollapsibleContent(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Check for collapsible elements
	expectedContent := []string{
		"collapsible-content",
		"toggle-icon",
		"onclick",
		fmt.Sprintf("rule-%p", rule),
	}

	for _, expected := range expectedContent {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML should contain collapsible element %q", expected)
		}
	}
}

func TestGenerateHTML_HTMLEscaping(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule with special HTML characters
	rule := NewWithDomain("test <script>alert('xss')</script>", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})
	WithDescription(rule, "Description with <tags> & special chars")

	opts := DocumentOptions{
		Format: FormatHTML,
	}

	html, err := GenerateHTML(opts)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	// Should escape HTML characters
	if strings.Contains(html, "<script>alert") {
		t.Error("HTML should escape script tags")
	}

	if !strings.Contains(html, "&lt;") {
		t.Error("HTML should contain escaped angle brackets")
	}
}
