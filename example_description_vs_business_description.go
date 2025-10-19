package rules

import "fmt"

// ExampleRuleMetadata_descriptionVsBusinessDescription demonstrates the difference
// between Description (technical) and BusinessDescription (business context).
func ExampleRuleMetadata_descriptionVsBusinessDescription() {
	type Order struct {
		Amount  float64
		Country string
	}

	// Clear registry for clean example
	DefaultRegistry.Clear()

	const OrderDomain Domain = "order"

	// Create a rule
	minAmountRule := NewWithDomain(
		"minimum order amount",
		OrderDomain,
		func(order Order) (bool, error) {
			return order.Amount >= 100.0, nil
		},
	)

	// Set technical description (WHAT it does) - for developers
	UpdateDescription(minAmountRule,
		"Validates that order amount is at least $100")

	// Set business context (WHY it exists) - for stakeholders
	UpdateMetadata(minAmountRule, RuleMetadata{
		RequirementID: "JIRA-1234",
		BusinessDescription: "Standard customers must have a minimum order " +
			"amount of $100 to qualify for free shipping per our Q1 2025 " +
			"shipping policy approved by the logistics team",
		Owner:   "Order Team",
		Version: "1.0.0",
	})

	// Another example with shipping
	countryRule := NewWithDomain(
		"valid shipping country",
		OrderDomain,
		func(order Order) (bool, error) {
			return order.Country == "US" || order.Country == "CA", nil
		},
	)

	// Technical: what the code checks
	UpdateDescription(countryRule,
		"Checks if order country is US or CA")

	// Business: why we have this restriction
	UpdateMetadata(countryRule, RuleMetadata{
		RequirementID: "JIRA-1235",
		BusinessDescription: "We only ship to US and Canada due to current " +
			"logistics partnerships and customs agreements. Expansion to EU " +
			"is planned for Q3 2025 pending legal review",
		Owner:   "Shipping Team",
		Version: "2.1.0",
	})

	fmt.Println("Rules configured with both technical and business descriptions")

	// Output:
	// Rules configured with both technical and business descriptions
}
