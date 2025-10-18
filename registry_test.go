package rules

import (
	"testing"
	"time"
)

// Test domains
const (
	TestOrderDomain   = Domain("order")
	TestUserDomain    = Domain("user")
	TestPaymentDomain = Domain("payment")
)

func TestDomain_String(t *testing.T) {
	domain := Domain("test")
	if domain.String() != "test" {
		t.Errorf("Domain.String() = %q, want %q", domain.String(), "test")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return o.Amount >= 100, nil
	})

	err := registry.Register(rule, WithDomain(TestOrderDomain))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rules := registry.AllRules()
	if len(rules) != 1 {
		t.Fatalf("AllRules() returned %d rules, want 1", len(rules))
	}

	if len(rules[0].Domains) != 1 || rules[0].Domains[0] != TestOrderDomain {
		t.Errorf("Rule domains = %v, want [%v]", rules[0].Domains, TestOrderDomain)
	}
}

func TestRegistry_RegisterMultipleDomains(t *testing.T) {
	registry := NewRegistry()

	rule := New("cross-domain rule", func(o Order) (bool, error) {
		return true, nil
	})

	err := registry.Register(
		rule,
		WithDomains(TestOrderDomain, TestUserDomain),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rules := registry.AllRules()
	if len(rules) != 1 {
		t.Fatalf("AllRules() returned %d rules, want 1", len(rules))
	}

	if len(rules[0].Domains) != 2 {
		t.Errorf("Rule has %d domains, want 2", len(rules[0].Domains))
	}
}

func TestRegistry_RegisterWithGroup(t *testing.T) {
	registry := NewRegistry()

	rule := New("eligibility rule", func(o Order) (bool, error) {
		return true, nil
	})

	groupName := "Order Eligibility"
	err := registry.Register(
		rule,
		WithGroup(groupName, TestOrderDomain, TestUserDomain),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rules := registry.AllRules()
	if len(rules) != 1 {
		t.Fatalf("AllRules() returned %d rules, want 1", len(rules))
	}

	if rules[0].Group != groupName {
		t.Errorf("Rule group = %q, want %q", rules[0].Group, groupName)
	}

	if len(rules[0].Domains) != 2 {
		t.Errorf("Rule has %d domains, want 2", len(rules[0].Domains))
	}
}

func TestRegistry_RegisterWithDescription(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	description := "This is a test rule"
	err := registry.Register(
		rule,
		WithDomain(TestOrderDomain),
		WithRegistrationDescription(description),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rules := registry.AllRules()
	if rules[0].Description != description {
		t.Errorf("Rule description = %q, want %q",
			rules[0].Description, description)
	}
}

func TestRegistry_RegisterWithMetadata(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	metadata := RuleMetadata{
		Owner:   "test-team",
		Version: "1.0.0",
		Tags:    []string{"validation", "financial"},
	}

	err := registry.Register(
		rule,
		WithDomain(TestOrderDomain),
		WithRegistrationMetadata(metadata),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rules := registry.AllRules()
	if rules[0].Metadata == nil {
		t.Fatal("Rule metadata is nil")
	}

	if rules[0].Metadata.Owner != metadata.Owner {
		t.Errorf("Metadata owner = %q, want %q",
			rules[0].Metadata.Owner, metadata.Owner)
	}

	if rules[0].Metadata.Version != metadata.Version {
		t.Errorf("Metadata version = %q, want %q",
			rules[0].Metadata.Version, metadata.Version)
	}

	if len(rules[0].Metadata.Tags) != len(metadata.Tags) {
		t.Errorf("Metadata has %d tags, want %d",
			len(rules[0].Metadata.Tags), len(metadata.Tags))
	}
}

func TestRegistry_RulesByDomain(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("order rule", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("user rule", func(u User) (bool, error) {
		return true, nil
	})
	rule3 := New("cross-domain rule", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule1, WithDomain(TestOrderDomain))
	_ = registry.Register(rule2, WithDomain(TestUserDomain))
	_ = registry.Register(rule3, WithDomains(TestOrderDomain, TestUserDomain))

	orderRules := registry.RulesByDomain(TestOrderDomain)
	if len(orderRules) != 2 {
		t.Errorf("RulesByDomain(order) returned %d rules, want 2",
			len(orderRules))
	}

	userRules := registry.RulesByDomain(TestUserDomain)
	if len(userRules) != 2 {
		t.Errorf("RulesByDomain(user) returned %d rules, want 2",
			len(userRules))
	}

	paymentRules := registry.RulesByDomain(TestPaymentDomain)
	if len(paymentRules) != 0 {
		t.Errorf("RulesByDomain(payment) returned %d rules, want 0",
			len(paymentRules))
	}
}

func TestRegistry_RulesByDomains(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("order rule", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("user rule", func(u User) (bool, error) {
		return true, nil
	})
	rule3 := New("payment rule", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule1, WithDomain(TestOrderDomain))
	_ = registry.Register(rule2, WithDomain(TestUserDomain))
	_ = registry.Register(rule3, WithDomain(TestPaymentDomain))

	rules := registry.RulesByDomains(TestOrderDomain, TestUserDomain)
	if len(rules) != 2 {
		t.Errorf("RulesByDomains(order, user) returned %d rules, want 2",
			len(rules))
	}
}

func TestRegistry_RulesByGroup(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("rule 1", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("rule 2", func(o Order) (bool, error) {
		return true, nil
	})
	rule3 := New("rule 3", func(o Order) (bool, error) {
		return true, nil
	})

	group1 := "Group 1"
	group2 := "Group 2"

	_ = registry.Register(rule1, WithGroup(group1, TestOrderDomain))
	_ = registry.Register(rule2, WithGroup(group1, TestUserDomain))
	_ = registry.Register(rule3, WithGroup(group2, TestOrderDomain))

	group1Rules := registry.RulesByGroup(group1)
	if len(group1Rules) != 2 {
		t.Errorf("RulesByGroup(%q) returned %d rules, want 2",
			group1, len(group1Rules))
	}

	group2Rules := registry.RulesByGroup(group2)
	if len(group2Rules) != 1 {
		t.Errorf("RulesByGroup(%q) returned %d rules, want 1",
			group2, len(group2Rules))
	}

	nonexistentRules := registry.RulesByGroup("Nonexistent")
	if len(nonexistentRules) != 0 {
		t.Errorf("RulesByGroup(nonexistent) returned %d rules, want 0",
			len(nonexistentRules))
	}
}

func TestRegistry_Domains(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("rule 1", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("rule 2", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule1, WithDomain(TestOrderDomain))
	_ = registry.Register(rule2, WithDomains(TestUserDomain, TestPaymentDomain))

	domains := registry.Domains()
	if len(domains) != 3 {
		t.Errorf("Domains() returned %d domains, want 3", len(domains))
	}

	domainSet := make(map[Domain]bool)
	for _, d := range domains {
		domainSet[d] = true
	}

	expectedDomains := []Domain{
		TestOrderDomain,
		TestUserDomain,
		TestPaymentDomain,
	}
	for _, expected := range expectedDomains {
		if !domainSet[expected] {
			t.Errorf("Domain %v not found in result", expected)
		}
	}
}

func TestRegistry_Groups(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("rule 1", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("rule 2", func(o Order) (bool, error) {
		return true, nil
	})
	rule3 := New("rule 3", func(o Order) (bool, error) {
		return true, nil
	})

	group1 := "Group 1"
	group2 := "Group 2"

	_ = registry.Register(rule1, WithGroup(group1, TestOrderDomain))
	_ = registry.Register(rule2, WithGroup(group2, TestUserDomain))
	_ = registry.Register(rule3, WithDomain(TestPaymentDomain)) // No group

	groups := registry.Groups()
	if len(groups) != 2 {
		t.Errorf("Groups() returned %d groups, want 2", len(groups))
	}

	groupSet := make(map[string]bool)
	for _, g := range groups {
		groupSet[g] = true
	}

	if !groupSet[group1] || !groupSet[group2] {
		t.Errorf("Expected groups %q and %q not found", group1, group2)
	}
}

func TestRegistry_UpdateDescription(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule, WithDomain(TestOrderDomain))

	newDescription := "Updated description"
	err := registry.UpdateDescription(rule, newDescription)
	if err != nil {
		t.Fatalf("UpdateDescription() error = %v", err)
	}

	rules := registry.AllRules()
	if rules[0].Description != newDescription {
		t.Errorf("Description = %q, want %q",
			rules[0].Description, newDescription)
	}
}

func TestRegistry_UpdateMetadata(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule, WithDomain(TestOrderDomain))

	newMetadata := RuleMetadata{
		Owner:   "updated-team",
		Version: "2.0.0",
	}

	err := registry.UpdateMetadata(rule, newMetadata)
	if err != nil {
		t.Fatalf("UpdateMetadata() error = %v", err)
	}

	rules := registry.AllRules()
	if rules[0].Metadata.Owner != newMetadata.Owner {
		t.Errorf("Metadata owner = %q, want %q",
			rules[0].Metadata.Owner, newMetadata.Owner)
	}
}

func TestRegistry_GetDescription(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	description := "Test description"
	_ = registry.Register(
		rule,
		WithDomain(TestOrderDomain),
		WithRegistrationDescription(description),
	)

	retrieved := registry.GetDescription(rule)
	if retrieved != description {
		t.Errorf("GetDescription() = %q, want %q", retrieved, description)
	}

	// Test with non-registered rule
	unregisteredRule := New("unregistered", func(o Order) (bool, error) {
		return true, nil
	})
	retrieved = registry.GetDescription(unregisteredRule)
	if retrieved != "" {
		t.Errorf("GetDescription(unregistered) = %q, want empty string",
			retrieved)
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	rule1 := New("rule 1", func(o Order) (bool, error) {
		return true, nil
	})
	rule2 := New("rule 2", func(o Order) (bool, error) {
		return true, nil
	})

	_ = registry.Register(rule1, WithDomain(TestOrderDomain))
	_ = registry.Register(rule2, WithDomain(TestUserDomain))

	if len(registry.AllRules()) != 2 {
		t.Fatalf("Expected 2 rules before clear")
	}

	registry.Clear()

	if len(registry.AllRules()) != 0 {
		t.Errorf("AllRules() after Clear() returned %d rules, want 0",
			len(registry.AllRules()))
	}
}

func TestRegistry_RegisteredAt(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	before := time.Now()
	_ = registry.Register(rule, WithDomain(TestOrderDomain))
	after := time.Now()

	rules := registry.AllRules()
	registeredAt := rules[0].RegisteredAt

	if registeredAt.Before(before) || registeredAt.After(after) {
		t.Errorf("RegisteredAt %v is outside expected range [%v, %v]",
			registeredAt, before, after)
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	registry := NewRegistry()

	rule := New("test rule", func(o Order) (bool, error) {
		return true, nil
	})

	// First registration
	_ = registry.Register(rule, WithDomain(TestOrderDomain))

	// Second registration (should update, not duplicate)
	_ = registry.Register(
		rule,
		WithDomains(TestOrderDomain, TestUserDomain),
	)

	rules := registry.AllRules()
	if len(rules) != 1 {
		t.Errorf("AllRules() returned %d rules after duplicate registration, want 1",
			len(rules))
	}

	// Should have updated domains
	if len(rules[0].Domains) != 2 {
		t.Errorf("Rule has %d domains after update, want 2",
			len(rules[0].Domains))
	}
}

func TestDeduplicateDomains(t *testing.T) {
	tests := []struct {
		name     string
		input    []Domain
		expected []Domain
	}{
		{
			name:     "no duplicates",
			input:    []Domain{"a", "b", "c"},
			expected: []Domain{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []Domain{"a", "b", "a", "c", "b"},
			expected: []Domain{"a", "b", "c"},
		},
		{
			name:     "empty",
			input:    []Domain{},
			expected: []Domain{},
		},
		{
			name:     "all same",
			input:    []Domain{"a", "a", "a"},
			expected: []Domain{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateDomains(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("deduplicateDomains() returned %d domains, want %d",
					len(result), len(tt.expected))
			}

			resultSet := make(map[Domain]bool)
			for _, d := range result {
				resultSet[d] = true
			}

			for _, expected := range tt.expected {
				if !resultSet[expected] {
					t.Errorf("Domain %v not found in result", expected)
				}
			}
		})
	}
}
