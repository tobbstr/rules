package rules

import (
	"testing"
)

// Test types for domain inheritance tests (using simple types)
type TestOrder struct {
	Amount  float64
	Country string
}

type TestUser struct {
	IsVIP  bool
	Status string
}

// Combined type for cross-domain rules
type TestContext struct {
	Order TestOrder
	User  TestUser
}

func TestNewWithDomain_AutoRegistration(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	if rule == nil {
		t.Fatal("NewWithDomain() returned nil")
	}

	// Verify it was auto-registered
	rules := AllRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 auto-registered rule, got %d", len(rules))
	}

	if len(rules[0].Domains) != 1 || rules[0].Domains[0] != TestOrderDomain {
		t.Errorf("Rule domains = %v, want [%v]", rules[0].Domains, TestOrderDomain)
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestNewWithGroup_AutoRegistration(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	groupName := "Test Group"
	rule := NewWithGroup(
		"cross-domain rule",
		groupName,
		[]Domain{TestOrderDomain, TestUserDomain},
		func(ctx TestContext) (bool, error) {
			return true, nil
		},
	)

	if rule == nil {
		t.Fatal("NewWithGroup() returned nil")
	}

	// Verify it was auto-registered
	rules := AllRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 auto-registered rule, got %d", len(rules))
	}

	if rules[0].Group != groupName {
		t.Errorf("Rule group = %q, want %q", rules[0].Group, groupName)
	}

	if len(rules[0].Domains) != 2 {
		t.Errorf("Rule has %d domains, want 2", len(rules[0].Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestWithDescription_UpdatesRegistry(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule := NewWithDomain("test rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	description := "This is a test description"
	returnedRule := WithDescription(rule, description)

	// Should return the same rule
	if returnedRule != rule {
		t.Error("WithDescription() should return the same rule instance")
	}

	// Verify description was updated in registry
	retrieved := GetDescription(rule)
	if retrieved != description {
		t.Errorf("GetDescription() = %q, want %q", retrieved, description)
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAnd_DomainInheritance_SingleDomain(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	rule2 := NewWithDomain("rule 2", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Country == "US", nil
	})

	// And should inherit and deduplicate domains
	andRule := And("combined", rule1, rule2)

	// Verify And rule was auto-registered with inherited domain
	allRules := AllRules()

	// Find the And rule in registry
	var andRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == andRule {
			andRuleEntry = &r
			break
		}
	}

	if andRuleEntry == nil {
		t.Fatal("And rule not found in registry")
	}

	if len(andRuleEntry.Domains) != 1 {
		t.Errorf("And rule has %d domains, want 1 (deduplicated)", len(andRuleEntry.Domains))
	}

	if andRuleEntry.Domains[0] != TestOrderDomain {
		t.Errorf("And rule domain = %v, want %v", andRuleEntry.Domains[0], TestOrderDomain)
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAnd_DomainInheritance_MultipleDomains(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("order rule", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return ctx.Order.Amount >= 100, nil
	})

	rule2 := NewWithDomain("user rule", TestUserDomain, func(ctx TestContext) (bool, error) {
		return ctx.User.IsVIP, nil
	})

	// And should inherit domains from both rules
	andRule := And("cross-domain and", rule1, rule2)

	// Find the And rule in registry
	allRules := AllRules()
	var andRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == andRule {
			andRuleEntry = &r
			break
		}
	}

	if andRuleEntry == nil {
		t.Fatal("And rule not found in registry")
	}

	if len(andRuleEntry.Domains) != 2 {
		t.Errorf("And rule has %d domains, want 2", len(andRuleEntry.Domains))
	}

	domainSet := make(map[Domain]bool)
	for _, d := range andRuleEntry.Domains {
		domainSet[d] = true
	}

	if !domainSet[TestOrderDomain] || !domainSet[TestUserDomain] {
		t.Errorf("And rule domains = %v, want [order, user]", andRuleEntry.Domains)
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestOr_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("order rule", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return ctx.Order.Amount >= 100, nil
	})

	rule2 := NewWithDomain("user rule", TestUserDomain, func(ctx TestContext) (bool, error) {
		return ctx.User.IsVIP, nil
	})

	orRule := Or("cross-domain or", rule1, rule2)

	// Find the Or rule in registry
	allRules := AllRules()
	var orRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == orRule {
			orRuleEntry = &r
			break
		}
	}

	if orRuleEntry == nil {
		t.Fatal("Or rule not found in registry")
	}

	if len(orRuleEntry.Domains) != 2 {
		t.Errorf("Or rule has %d domains, want 2", len(orRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestNot_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return o.Amount >= 100, nil
	})

	notRule := Not("not order rule", rule)

	// Find the Not rule in registry
	allRules := AllRules()
	var notRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == notRule {
			notRuleEntry = &r
			break
		}
	}

	if notRuleEntry == nil {
		t.Fatal("Not rule not found in registry")
	}

	if len(notRuleEntry.Domains) != 1 || notRuleEntry.Domains[0] != TestOrderDomain {
		t.Errorf("Not rule domains = %v, want [%v]", notRuleEntry.Domains, TestOrderDomain)
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAllOf_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	allOfRule := AllOf("all of", rule1, rule2)

	// AllOf is just an alias for And, so should behave the same
	allRules := AllRules()
	var allOfRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == allOfRule {
			allOfRuleEntry = &r
			break
		}
	}

	if allOfRuleEntry == nil {
		t.Fatal("AllOf rule not found in registry")
	}

	if len(allOfRuleEntry.Domains) != 2 {
		t.Errorf("AllOf rule has %d domains, want 2", len(allOfRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAnyOf_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	anyOfRule := AnyOf("any of", rule1, rule2)

	// AnyOf is just an alias for Or
	allRules := AllRules()
	var anyOfRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == anyOfRule {
			anyOfRuleEntry = &r
			break
		}
	}

	if anyOfRuleEntry == nil {
		t.Fatal("AnyOf rule not found in registry")
	}

	if len(anyOfRuleEntry.Domains) != 2 {
		t.Errorf("AnyOf rule has %d domains, want 2", len(anyOfRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestNoneOf_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	noneOfRule := NoneOf("none of", rule1, rule2)

	// NoneOf creates Not(Or(...)), so should inherit all domains
	allRules := AllRules()
	var noneOfRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == noneOfRule {
			noneOfRuleEntry = &r
			break
		}
	}

	if noneOfRuleEntry == nil {
		t.Fatal("NoneOf rule not found in registry")
	}

	if len(noneOfRuleEntry.Domains) != 2 {
		t.Errorf("NoneOf rule has %d domains, want 2", len(noneOfRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAtLeast_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	atLeastRule := AtLeast("at least 1", 1, rule1, rule2)

	allRules := AllRules()
	var atLeastRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == atLeastRule {
			atLeastRuleEntry = &r
			break
		}
	}

	if atLeastRuleEntry == nil {
		t.Fatal("AtLeast rule not found in registry")
	}

	if len(atLeastRuleEntry.Domains) != 2 {
		t.Errorf("AtLeast rule has %d domains, want 2", len(atLeastRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestExactly_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	exactlyRule := Exactly("exactly 1", 1, rule1, rule2)

	allRules := AllRules()
	var exactlyRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == exactlyRule {
			exactlyRuleEntry = &r
			break
		}
	}

	if exactlyRuleEntry == nil {
		t.Fatal("Exactly rule not found in registry")
	}

	if len(exactlyRuleEntry.Domains) != 2 {
		t.Errorf("Exactly rule has %d domains, want 2", len(exactlyRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestAtMost_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("rule 1", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("rule 2", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	atMostRule := AtMost("at most 1", 1, rule1, rule2)

	allRules := AllRules()
	var atMostRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == atMostRule {
			atMostRuleEntry = &r
			break
		}
	}

	if atMostRuleEntry == nil {
		t.Fatal("AtMost rule not found in registry")
	}

	if len(atMostRuleEntry.Domains) != 2 {
		t.Errorf("AtMost rule has %d domains, want 2", len(atMostRuleEntry.Domains))
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestNestedRules_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	rule1 := NewWithDomain("order rule", TestOrderDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule2 := NewWithDomain("user rule", TestUserDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	rule3 := NewWithDomain("payment rule", TestPaymentDomain, func(ctx TestContext) (bool, error) {
		return true, nil
	})

	// Create nested structure: And(Or(rule1, rule2), rule3)
	orRule := Or("order or user", rule1, rule2)
	andRule := And("complex rule", orRule, rule3)

	// Find the And rule in registry
	allRules := AllRules()
	var andRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == andRule {
			andRuleEntry = &r
			break
		}
	}

	if andRuleEntry == nil {
		t.Fatal("And rule not found in registry")
	}

	// Should have all three domains
	if len(andRuleEntry.Domains) != 3 {
		t.Errorf("Nested And rule has %d domains, want 3", len(andRuleEntry.Domains))
	}

	domainSet := make(map[Domain]bool)
	for _, d := range andRuleEntry.Domains {
		domainSet[d] = true
	}

	expectedDomains := []Domain{TestOrderDomain, TestUserDomain, TestPaymentDomain}
	for _, expected := range expectedDomains {
		if !domainSet[expected] {
			t.Errorf("Domain %v not found in nested rule domains", expected)
		}
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestRulesWithoutDomain_NotRegistered(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	// Rules created with New() (no domain) should not be auto-registered
	rule1 := New("utility rule", func(o TestOrder) (bool, error) {
		return true, nil
	})

	rule2 := New("another utility", func(o TestOrder) (bool, error) {
		return true, nil
	})

	// Combine them with And - should still not register
	andRule := And("utility and", rule1, rule2)

	// Verify nothing was registered
	allRules := AllRules()
	if len(allRules) != 0 {
		t.Errorf("Expected 0 registered rules, got %d", len(allRules))
	}

	// Verify the rules still work
	result, err := andRule.Evaluate(TestOrder{Amount: 100})
	if err != nil {
		t.Errorf("andRule.Evaluate() error = %v", err)
	}
	if !result {
		t.Error("andRule.Evaluate() returned false, want true")
	}

	// Clean up
	DefaultRegistry.Clear()
}

func TestMixedRules_DomainInheritance(t *testing.T) {
	// Clear registry before test
	DefaultRegistry.Clear()

	// Mix of rules with and without domains
	rule1 := NewWithDomain("order rule", TestOrderDomain, func(o TestOrder) (bool, error) {
		return true, nil
	})

	rule2 := New("utility rule", func(o TestOrder) (bool, error) {
		return true, nil
	})

	// And should inherit only the order domain
	andRule := And("mixed", rule1, rule2)

	allRules := AllRules()
	var andRuleEntry *RegisteredRule
	for _, r := range allRules {
		if r.Rule == andRule {
			andRuleEntry = &r
			break
		}
	}

	if andRuleEntry == nil {
		t.Fatal("Mixed And rule not found in registry")
	}

	if len(andRuleEntry.Domains) != 1 || andRuleEntry.Domains[0] != TestOrderDomain {
		t.Errorf("Mixed And rule domains = %v, want [%v]",
			andRuleEntry.Domains, TestOrderDomain)
	}

	// Clean up
	DefaultRegistry.Clear()
}
