// Package rules provides a type-safe, expressive library for defining and
// evaluating hierarchical business rules in Go.
//
// # Overview
//
// The rules package allows you to define complex business logic through
// composable, hierarchical rules. Rules can be combined using logical
// operators (AND, OR, NOT) and evaluated against input data in a type-safe
// manner using Go generics.
//
// # Key Features
//
//   - Type-safe rules using Go generics
//   - Hierarchical rule composition (rules can contain other rules)
//   - Context support for cancellation and deadlines
//   - Comprehensive error handling with wrapped errors
//   - Detailed evaluation results with timing information
//   - Fluent builder API for constructing complex rules
//   - Helper functions for common patterns
//
// # Basic Usage
//
// Create a simple rule:
//
//	rule := rules.New(
//	    "minimum amount",
//	    func(order Order) (bool, error) {
//	        return order.Amount >= 100.0, nil
//	    },
//	)
//
//	satisfied, err := rule.Evaluate(order)
//
// # Hierarchical Rules
//
// Combine rules using logical operators:
//
//	rule1 := rules.New("amount > 100", ...)
//	rule2 := rules.New("valid country", ...)
//	rule3 := rules.New("VIP customer", ...)
//
//	// All rules must be satisfied
//	allRules := rules.And("all requirements", rule1, rule2, rule3)
//
//	// At least one rule must be satisfied
//	anyRule := rules.Or("any requirement", rule1, rule2, rule3)
//
//	// Rule must not be satisfied
//	notRule := rules.Not("not VIP", rule3)
//
// # Builder Pattern
//
// Use the builder for a more fluent API:
//
//	builder := rules.NewBuilder[Order]()
//	builder.
//	    AddCondition("amount >= 100", func(o Order) bool {
//	        return o.Amount >= 100
//	    }).
//	    AddCondition("valid country", func(o Order) bool {
//	        return o.Country == "US"
//	    })
//
//	rule := builder.BuildAnd("eligibility")
//
// # Helper Functions
//
// The package provides helper functions for common patterns:
//
//   - Always/Never: Rules that always/never succeed
//   - AllOf/AnyOf/NoneOf: Aliases for AND/OR/NOT patterns
//   - AtLeast: At least N rules must be satisfied
//   - Exactly: Exactly N rules must be satisfied
//   - AtMost: At most N rules must be satisfied
//
// Example:
//
//	rule := rules.AtLeast("premium", 2, rule1, rule2, rule3)
//
// # Detailed Evaluation
//
// Use an Evaluator to get detailed results including child rule results:
//
//	evaluator := rules.NewEvaluator(rule)
//	result := evaluator.EvaluateDetailed(ctx, input)
//
//	// Check results
//	fmt.Println(result.String())  // Pretty-printed tree
//	fmt.Println(result.Duration)  // Evaluation time
//
// # Error Handling
//
// The package follows Go best practices for error handling:
//
//   - All evaluation functions return explicit errors
//   - Errors are wrapped with context using %w
//   - Sentinel errors are provided for common cases
//
// # Error Handling
//
// All evaluation functions return explicit errors:
//
//	satisfied, err := rule.Evaluate(input)
//	if err != nil {
//	    // Handle evaluation errors
//	}
//
// # Best Practices
//
//   - Use descriptive names for rules to aid debugging
//   - Keep individual rule predicates simple and focused
//   - Compose complex logic through hierarchical rules
//   - Use the builder for dynamic rule construction
//   - Handle errors explicitly at evaluation boundaries
package rules
