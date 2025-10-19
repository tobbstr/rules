package rules

import (
	"strings"
	"testing"
)

func TestGenerateMermaid_Empty(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for expected content
	if !strings.Contains(mermaid, "graph TD") {
		t.Error("Mermaid should contain graph header")
	}

	if !strings.Contains(mermaid, "No rules match filters") {
		t.Error("Mermaid should indicate no rules")
	}
}

func TestGenerateMermaid_SingleRule(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a simple rule
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for expected content
	expectedContent := []string{
		"graph TD",
		"test rule",
		getMermaidNodeID(rule),
	}

	for _, expected := range expectedContent {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("Mermaid should contain %q", expected)
		}
	}
}

func TestGenerateMermaid_WithMetadata(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule
	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	opts := DocumentOptions{
		Format:          FormatMermaid,
		IncludeMetadata: true,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for type in label when metadata is included
	if !strings.Contains(mermaid, "SIMPLE") {
		t.Error("Mermaid should contain SIMPLE type when metadata is included")
	}

	// Verify rule name is present
	nodeID := getMermaidNodeID(rule)
	if !strings.Contains(mermaid, nodeID) {
		t.Error("Mermaid should contain node ID")
	}
}

func TestGenerateMermaid_HierarchicalRules(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create hierarchical rules
	rule1 := NewWithDomain("amount >= 100", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	rule2 := NewWithDomain("country is US", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Country == "US", nil
	})

	andRule := And("order validation", rule1, rule2)
	_ = Register(andRule, WithDomain(TestOrderDomain))

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for all rules
	expectedContent := []string{
		"graph TD",
		getMermaidNodeID(rule1),
		getMermaidNodeID(rule2),
		getMermaidNodeID(andRule),
	}

	for _, expected := range expectedContent {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("Mermaid should contain %q", expected)
		}
	}

	// Check for connections (arrows)
	parentID := getMermaidNodeID(andRule)
	if !strings.Contains(mermaid, parentID+" -->") {
		t.Error("Mermaid should contain connection arrows")
	}
}

func TestGenerateMermaid_DifferentRuleTypes(t *testing.T) {
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
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for different node shapes
	// Simple: [label]
	// AND: [[label]]
	// OR: {label}
	// NOT: [(label)]

	simpleID := getMermaidNodeID(simple)
	andID := getMermaidNodeID(andRule)
	orID := getMermaidNodeID(orRule)
	notID := getMermaidNodeID(notRule)

	if !strings.Contains(mermaid, simpleID+"[") {
		t.Error("Simple rule should use rectangle shape []")
	}

	if !strings.Contains(mermaid, andID+"[[") {
		t.Error("AND rule should use double rectangle shape [[]]")
	}

	if !strings.Contains(mermaid, orID+"{") {
		t.Error("OR rule should use rhombus shape {}")
	}

	if !strings.Contains(mermaid, notID+"[(") {
		t.Error("NOT rule should use stadium shape [()]")
	}
}

func TestGenerateMermaid_DomainGrouping(t *testing.T) {
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
		Format:        FormatMermaid,
		GroupByDomain: true,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for subgraphs
	expectedContent := []string{
		"subgraph",
		string(TestOrderDomain) + " Domain",
		string(TestUserDomain) + " Domain",
		"end",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("Mermaid should contain %q", expected)
		}
	}
}

func TestGenerateMermaid_Styling(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule
	_ = NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for styling definitions
	expectedContent := []string{
		"%% Styling",
		"classDef simpleRule",
		"classDef andRule",
		"classDef orRule",
		"classDef notRule",
		"fill:",
		"stroke:",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("Mermaid should contain styling %q", expected)
		}
	}
}

func TestGenerateMermaid_Connections(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create hierarchical rules with different types
	simple := NewWithDomain("simple", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	andRule := And("and rule", simple)
	_ = Register(andRule, WithDomain(TestOrderDomain))

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for connection section
	if !strings.Contains(mermaid, "%% Rule connections") {
		t.Error("Mermaid should contain connections section")
	}

	// Check for arrow from parent to child
	parentID := getMermaidNodeID(andRule)
	childID := getMermaidNodeID(simple)

	// Should have connection like: R123 --> R456
	if !strings.Contains(mermaid, parentID) || !strings.Contains(mermaid, childID) {
		t.Error("Mermaid should contain parent and child node IDs")
	}

	if !strings.Contains(mermaid, "-->") {
		t.Error("Mermaid should contain connection arrow")
	}
}

func TestGenerateMermaid_DomainFiltering(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	// Filter to only include TestOrderDomain
	opts := DocumentOptions{
		Format:         FormatMermaid,
		IncludeDomains: []Domain{TestOrderDomain},
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Should contain order rule
	orderID := getMermaidNodeID(rule1)
	if !strings.Contains(mermaid, orderID) {
		t.Error("Mermaid should contain order rule")
	}

	// Should NOT contain user rule
	userID := getMermaidNodeID(rule2)
	if strings.Contains(mermaid, userID) {
		t.Error("Mermaid should not contain user rule")
	}
}

func TestGenerateMermaid_DomainExclusion(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules in different domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("user rule", TestUserDomain, func(u TestUser) (bool, error) {
		return true, nil
	})

	// Exclude TestUserDomain
	opts := DocumentOptions{
		Format:         FormatMermaid,
		ExcludeDomains: []Domain{TestUserDomain},
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Should contain order rule
	orderID := getMermaidNodeID(rule1)
	if !strings.Contains(mermaid, orderID) {
		t.Error("Mermaid should contain order rule")
	}

	// Should NOT contain user rule
	userID := getMermaidNodeID(rule2)
	if strings.Contains(mermaid, userID) {
		t.Error("Mermaid should not contain user rule")
	}
}

func TestGenerateMermaid_MaxDepth(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create deeply nested rules (use New to avoid auto-registration)
	level3 := New("level 3", func(o TestOrder) (bool, error) {
		return true, nil
	})

	level2 := Not("not level 3", level3)
	// Don't register level2 separately - it should only appear as a child of level1

	level1 := And("and level 2", level2)
	_ = Register(level1, WithDomain(TestOrderDomain))

	// Limit depth to 2 (0=level1, 1=level2, stop before level3)
	opts := DocumentOptions{
		Format:   FormatMermaid,
		MaxDepth: 2,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Should contain level 1 and 2
	level1ID := getMermaidNodeID(level1)
	level2ID := getMermaidNodeID(level2)

	if !strings.Contains(mermaid, level1ID) {
		t.Error("Mermaid should contain Level 1 rule")
	}
	if !strings.Contains(mermaid, level2ID) {
		t.Error("Mermaid should contain Level 2 rule")
	}

	// Should NOT contain level 3 (beyond max depth)
	level3ID := getMermaidNodeID(level3)
	if strings.Contains(mermaid, level3ID) {
		t.Error("Mermaid should not contain Level 3 rule (beyond MaxDepth)")
	}
}

func TestGenerateDomainMermaid(t *testing.T) {
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
		Format: FormatMermaid,
	}

	mermaid, err := GenerateDomainMermaid(TestOrderDomain, opts)
	if err != nil {
		t.Fatalf("GenerateDomainMermaid() error = %v", err)
	}

	// Should contain order domain subgraph
	if !strings.Contains(mermaid, string(TestOrderDomain)) {
		t.Error("Mermaid should contain order domain")
	}

	// Should contain order rule
	orderID := getMermaidNodeID(rule1)
	if !strings.Contains(mermaid, orderID) {
		t.Error("Mermaid should contain order rule")
	}
}

func TestGenerateDomainsMermaid(t *testing.T) {
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
		Format: FormatMermaid,
	}

	mermaid, err := GenerateDomainsMermaid([]Domain{TestOrderDomain, TestUserDomain}, opts)
	if err != nil {
		t.Fatalf("GenerateDomainsMermaid() error = %v", err)
	}

	// Should contain both domains
	expectedContent := []string{
		string(TestOrderDomain),
		string(TestUserDomain),
		"subgraph",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(mermaid, expected) {
			t.Errorf("Mermaid should contain %q", expected)
		}
	}
}

func TestGenerateGroupMermaid(t *testing.T) {
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
		Format: FormatMermaid,
	}

	mermaid, err := GenerateGroupMermaid("Validation", opts)
	if err != nil {
		t.Fatalf("GenerateGroupMermaid() error = %v", err)
	}

	// Should contain validation rule
	validationID := getMermaidNodeID(rule1)
	if !strings.Contains(mermaid, validationID) {
		t.Error("Mermaid should contain validation rule")
	}
}

func TestGenerateMermaid_EscapeSpecialCharacters(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule with special characters in name
	rule := NewWithDomain("test \"quoted\" rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Should escape quotes
	if !strings.Contains(mermaid, "&quot;") {
		t.Error("Mermaid should escape quotes in labels")
	}

	// Should still contain the rule
	ruleID := getMermaidNodeID(rule)
	if !strings.Contains(mermaid, ruleID) {
		t.Error("Mermaid should contain rule with escaped label")
	}
}

func TestGenerateMermaid_DifferentArrowStyles(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create rules with different parent types
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
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Check for different arrow styles
	// AND uses -->
	// OR uses -.->
	// NOT uses ==>

	if !strings.Contains(mermaid, "-->") {
		t.Error("Mermaid should contain solid arrow (-->)")
	}
}

func TestGenerateMermaid_NoDuplicateNodes(t *testing.T) {
	// Clear registry
	DefaultRegistry.Clear()
	defer DefaultRegistry.Clear()

	// Create a rule that's shared by multiple parents
	shared := NewWithDomain("shared rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	and1 := And("and 1", shared)
	_ = Register(and1, WithDomain(TestOrderDomain))

	and2 := And("and 2", shared)
	_ = Register(and2, WithDomain(TestOrderDomain))

	opts := DocumentOptions{
		Format: FormatMermaid,
	}

	mermaid, err := GenerateMermaid(opts)
	if err != nil {
		t.Fatalf("GenerateMermaid() error = %v", err)
	}

	// Count occurrences of shared rule's node definition
	sharedID := getMermaidNodeID(shared)
	nodeDefPattern := sharedID + "["

	count := strings.Count(mermaid, nodeDefPattern)
	if count > 1 {
		t.Errorf("Shared rule should appear only once in node definitions, found %d times", count)
	}
}

func TestSanitizeMermaidID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "spaces",
			input:    "Order Domain",
			expected: "Order_Domain",
		},
		{
			name:     "hyphens",
			input:    "order-management",
			expected: "order_management",
		},
		{
			name:     "dots",
			input:    "order.shipping",
			expected: "order_shipping",
		},
		{
			name:     "mixed",
			input:    "order-management.shipping domain",
			expected: "order_management_shipping_domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeMermaidID(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeMermaidID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeMermaidLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "quotes",
			input:    `test "quoted" text`,
			expected: `test &quot;quoted&quot; text`,
		},
		{
			name:     "no special chars",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "multiple quotes",
			input:    `"start" and "end"`,
			expected: `&quot;start&quot; and &quot;end&quot;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMermaidLabel(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMermaidLabel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
