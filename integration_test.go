package rules_test

import (
	"testing"
	"time"

	"github.com/tobbstr/rules"
)

// TestRealWorldOrderProcessing demonstrates a complete real-world scenario
func TestRealWorldOrderProcessing(t *testing.T) {
	t.Parallel()

	// Define a realistic order type
	type Order struct {
		ID              string
		CustomerID      string
		TotalAmount     float64
		ItemCount       int
		ShippingCountry string
		IsVIPCustomer   bool
		IsPrimeCustomer bool
		HasPromoCode    bool
		CreatedAt       time.Time
	}

	// Build comprehensive order validation rules
	buildOrderRules := func() rules.Rule[Order] {
		// Basic validation rules
		hasItems := rules.New(
			"has items",
			func(order Order) (bool, error) {
				return order.ItemCount > 0, nil
			},
		)

		hasValidAmount := rules.New(
			"has valid amount",
			func(order Order) (bool, error) {
				return order.TotalAmount > 0, nil
			},
		)

		basicValidation := rules.And(
			"basic validation",
			hasItems,
			hasValidAmount,
		)

		// Shipping rules
		domesticShipping := rules.New(
			"domestic shipping",
			func(order Order) (bool, error) {
				return order.ShippingCountry == "US", nil
			},
		)

		internationalShipping := rules.New(
			"international shipping",
			func(order Order) (bool, error) {
				validCountries := map[string]bool{
					"CA": true, "UK": true, "DE": true, "FR": true,
				}
				return validCountries[order.ShippingCountry], nil
			},
		)

		validShipping := rules.Or(
			"valid shipping",
			domesticShipping,
			internationalShipping,
		)

		// Minimum order rules by customer type
		standardMinimum := rules.New(
			"standard minimum",
			func(order Order) (bool, error) {
				return order.TotalAmount >= 50.0, nil
			},
		)

		// VIP customers - no minimum
		vipCustomer := rules.New(
			"VIP customer",
			func(order Order) (bool, error) {
				return order.IsVIPCustomer, nil
			},
		)

		// Prime customers - reduced minimum
		primeMinimum := rules.New(
			"prime minimum",
			func(order Order) (bool, error) {
				if !order.IsPrimeCustomer {
					return false, nil
				}
				return order.TotalAmount >= 25.0, nil
			},
		)

		// Promo code - reduced minimum
		promoMinimum := rules.New(
			"promo minimum",
			func(order Order) (bool, error) {
				if !order.HasPromoCode {
					return false, nil
				}
				return order.TotalAmount >= 35.0, nil
			},
		)

		// Amount requirement is satisfied if ANY of these are true
		amountRequirement := rules.Or(
			"amount requirement",
			vipCustomer,
			primeMinimum,
			promoMinimum,
			standardMinimum,
		)

		// Final eligibility: all requirements must be met
		return rules.And(
			"order eligible for processing",
			basicValidation,
			validShipping,
			amountRequirement,
		)
	}

	eligibilityRule := buildOrderRules()
	evaluator := rules.NewEvaluator(eligibilityRule)

	// Test cases
	tests := []struct {
		name        string
		order       Order
		wantSuccess bool
	}{
		{
			name: "standard customer - valid order",
			order: Order{
				ID:              "ORD-001",
				CustomerID:      "CUST-001",
				TotalAmount:     75.0,
				ItemCount:       3,
				ShippingCountry: "US",
				IsVIPCustomer:   false,
				IsPrimeCustomer: false,
				HasPromoCode:    false,
			},
			wantSuccess: true,
		},
		{
			name: "standard customer - below minimum",
			order: Order{
				ID:              "ORD-002",
				CustomerID:      "CUST-002",
				TotalAmount:     30.0,
				ItemCount:       1,
				ShippingCountry: "US",
				IsVIPCustomer:   false,
				IsPrimeCustomer: false,
				HasPromoCode:    false,
			},
			wantSuccess: false,
		},
		{
			name: "VIP customer - no minimum required",
			order: Order{
				ID:              "ORD-003",
				CustomerID:      "CUST-003",
				TotalAmount:     10.0,
				ItemCount:       1,
				ShippingCountry: "US",
				IsVIPCustomer:   true,
				IsPrimeCustomer: false,
				HasPromoCode:    false,
			},
			wantSuccess: true,
		},
		{
			name: "Prime customer - meets reduced minimum",
			order: Order{
				ID:              "ORD-004",
				CustomerID:      "CUST-004",
				TotalAmount:     30.0,
				ItemCount:       2,
				ShippingCountry: "CA",
				IsVIPCustomer:   false,
				IsPrimeCustomer: true,
				HasPromoCode:    false,
			},
			wantSuccess: true,
		},
		{
			name: "Promo code - meets reduced minimum",
			order: Order{
				ID:              "ORD-005",
				CustomerID:      "CUST-005",
				TotalAmount:     40.0,
				ItemCount:       2,
				ShippingCountry: "UK",
				IsVIPCustomer:   false,
				IsPrimeCustomer: false,
				HasPromoCode:    true,
			},
			wantSuccess: true,
		},
		{
			name: "invalid shipping country",
			order: Order{
				ID:              "ORD-006",
				CustomerID:      "CUST-006",
				TotalAmount:     100.0,
				ItemCount:       3,
				ShippingCountry: "XX",
				IsVIPCustomer:   false,
				IsPrimeCustomer: false,
				HasPromoCode:    false,
			},
			wantSuccess: false,
		},
		{
			name: "no items",
			order: Order{
				ID:              "ORD-007",
				CustomerID:      "CUST-007",
				TotalAmount:     100.0,
				ItemCount:       0,
				ShippingCountry: "US",
				IsVIPCustomer:   false,
				IsPrimeCustomer: false,
				HasPromoCode:    false,
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use detailed evaluation for comprehensive results
			result := evaluator.EvaluateDetailed(tt.order)

			if result.IsSuccessful() != tt.wantSuccess {
				t.Errorf(
					"Order %s: expected success=%v, got %v\n%s",
					tt.order.ID,
					tt.wantSuccess,
					result.IsSuccessful(),
					result.String(),
				)
			}

			// Verify no errors occurred
			if result.HasError() {
				t.Errorf(
					"Order %s: unexpected error: %v",
					tt.order.ID,
					result.Error,
				)
			}

			// Log detailed results for inspection
			t.Logf("Order %s evaluation:\n%s", tt.order.ID, result.String())
		})
	}
}

// BenchmarkRuleEvaluation benchmarks rule evaluation performance
func BenchmarkRuleEvaluation(b *testing.B) {
	type SimpleInput struct {
		Value int
	}

	rule := rules.And(
		"complex rule",
		rules.New(
			"value > 10",
			func(input SimpleInput) (bool, error) {
				return input.Value > 10, nil
			},
		),
		rules.New(
			"value < 100",
			func(input SimpleInput) (bool, error) {
				return input.Value < 100, nil
			},
		),
		rules.New(
			"value is even",
			func(input SimpleInput) (bool, error) {
				return input.Value%2 == 0, nil
			},
		),
	)

	input := SimpleInput{Value: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate(input)
	}
}

// BenchmarkDetailedEvaluation benchmarks detailed evaluation
func BenchmarkDetailedEvaluation(b *testing.B) {
	type SimpleInput struct {
		Value int
	}

	rule := rules.And(
		"complex rule",
		rules.New(
			"value > 10",
			func(input SimpleInput) (bool, error) {
				return input.Value > 10, nil
			},
		),
		rules.Or(
			"value range",
			rules.New(
				"value < 50",
				func(input SimpleInput) (bool, error) {
					return input.Value < 50, nil
				},
			),
			rules.New(
				"value > 100",
				func(input SimpleInput) (bool, error) {
					return input.Value > 100, nil
				},
			),
		),
	)

	evaluator := rules.NewEvaluator(rule)
	input := SimpleInput{Value: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailed(input)
	}
}
