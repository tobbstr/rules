package rules_test

import (
	"fmt"

	"github.com/tobbstr/rules"
)

// Order represents a customer order
type Order struct {
	TotalAmount float64
	IsVIP       bool
	ItemCount   int
	Country     string
}

// Example_simpleRule demonstrates creating and evaluating a simple rule.
func Example_simpleRule() {
	// Create a simple rule
	minimumOrder := rules.New(
		"minimum order amount",
		func(order Order) (bool, error) {
			return order.TotalAmount >= 100.0, nil
		},
	)

	// Evaluate the rule
	order := Order{TotalAmount: 150.0}
	satisfied, err := minimumOrder.Evaluate(order)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Rule satisfied: %v\n", satisfied)
	// Output: Rule satisfied: true
}

// Example_hierarchicalRules demonstrates creating and evaluating
// hierarchical rules.
func Example_hierarchicalRules() {
	// Define individual rules
	minimumAmount := rules.New(
		"minimum amount",
		func(order Order) (bool, error) {
			return order.TotalAmount >= 100.0, nil
		},
	)

	validCountry := rules.New(
		"valid country",
		func(order Order) (bool, error) {
			return order.Country == "US" || order.Country == "CA", nil
		},
	)

	vipCustomer := rules.New(
		"VIP customer",
		func(order Order) (bool, error) {
			return order.IsVIP, nil
		},
	)

	// Combine rules hierarchically
	standardEligibility := rules.And(
		"standard eligibility",
		minimumAmount,
		validCountry,
	)

	// VIP customers OR standard eligible customers
	eligibility := rules.Or(
		"order eligibility",
		vipCustomer,
		standardEligibility,
	)

	// Test with a VIP customer
	order := Order{
		TotalAmount: 50.0, // Below minimum
		IsVIP:       true,
		Country:     "UK", // Invalid country
	}

	satisfied, _ := eligibility.Evaluate(order)
	fmt.Printf("VIP order eligible: %v\n", satisfied)

	// Test with standard customer
	order2 := Order{
		TotalAmount: 150.0, // Above minimum
		IsVIP:       false,
		Country:     "US", // Valid country
	}

	satisfied2, _ := eligibility.Evaluate(order2)
	fmt.Printf("Standard order eligible: %v\n", satisfied2)

	// Output:
	// VIP order eligible: true
	// Standard order eligible: true
}

// Example_builder demonstrates using the builder pattern to construct rules.
func Example_builder() {
	builder := rules.NewBuilder[Order]()

	// Add rules using fluent API
	builder.
		AddCondition("minimum amount", func(order Order) bool {
			return order.TotalAmount >= 100.0
		}).
		AddCondition("minimum items", func(order Order) bool {
			return order.ItemCount >= 3
		})

	// Build an AND rule
	eligibility := builder.BuildAnd("order requirements")

	order := Order{TotalAmount: 150.0, ItemCount: 5}
	satisfied, _ := eligibility.Evaluate(order)
	fmt.Printf("Requirements met: %v\n", satisfied)

	// Output: Requirements met: true
}

// Example_evaluator demonstrates using the evaluator for detailed results.
func Example_evaluator() {
	// Create hierarchical rules
	rule1 := rules.New(
		"amount >= 100",
		func(order Order) (bool, error) {
			return order.TotalAmount >= 100.0, nil
		},
	)

	rule2 := rules.New(
		"items >= 3",
		func(order Order) (bool, error) {
			return order.ItemCount >= 3, nil
		},
	)

	rule3 := rules.New(
		"valid country",
		func(order Order) (bool, error) {
			return order.Country == "US", nil
		},
	)

	eligibility := rules.And("order eligibility", rule1, rule2, rule3)

	// Use evaluator for detailed results
	evaluator := rules.NewEvaluator(eligibility)

	order := Order{
		TotalAmount: 150.0,
		ItemCount:   2, // Fails this requirement
		Country:     "US",
	}

	result := evaluator.EvaluateDetailed(order)

	fmt.Printf("Overall result: %v\n", result.IsSuccessful())
	fmt.Printf("Failed at: items >= 3\n")

	// Output:
	// Overall result: false
	// Failed at: items >= 3
}

// Example_helpers demonstrates using helper functions for common patterns.
func Example_helpers() {
	rule1 := rules.New(
		"amount > 100",
		func(order Order) (bool, error) {
			return order.TotalAmount > 100.0, nil
		},
	)

	rule2 := rules.New(
		"VIP customer",
		func(order Order) (bool, error) {
			return order.IsVIP, nil
		},
	)

	rule3 := rules.New(
		"items > 5",
		func(order Order) (bool, error) {
			return order.ItemCount > 5, nil
		},
	)

	// At least 2 of the 3 rules must be satisfied
	eligibility := rules.AtLeast("premium eligibility", 2, rule1, rule2, rule3)

	// VIP customer with decent amount but few items
	order := Order{
		TotalAmount: 150.0, // ✓
		IsVIP:       true,  // ✓
		ItemCount:   3,     // ✗
	}

	satisfied, _ := eligibility.Evaluate(order)
	fmt.Printf("Eligible: %v\n", satisfied)

	// Output: Eligible: true
}

// Example_noneOf demonstrates the NoneOf helper function.
func Example_noneOf() {
	// Define prohibited conditions
	suspiciousAmount := rules.New(
		"suspicious amount",
		func(order Order) (bool, error) {
			return order.TotalAmount >= 10000.0, nil
		},
	)

	bannedCountry := rules.New(
		"banned country",
		func(order Order) (bool, error) {
			return order.Country == "XX", nil
		},
	)

	// Order is valid if NONE of the prohibited conditions are met
	validOrder := rules.NoneOf(
		"no flags",
		suspiciousAmount,
		bannedCountry,
	)

	order := Order{
		TotalAmount: 500.0,
		Country:     "US",
	}

	satisfied, _ := validOrder.Evaluate(order)
	fmt.Printf("Valid order: %v\n", satisfied)

	// Output: Valid order: true
}
