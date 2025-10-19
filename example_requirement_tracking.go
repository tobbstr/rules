package rules

import "fmt"

// ExampleRuleMetadata_requirementTracking demonstrates linking rules to external
// requirements like Jira tickets and adding business descriptions.
func ExampleRuleMetadata_requirementTracking() {
	type Order struct {
		TotalAmount float64
		Country     string
	}

	// Clear the registry for clean example output
	DefaultRegistry.Clear()

	const OrderDomain Domain = "order"

	// Create a rule linked to a Jira ticket
	minimumAmount := NewWithDomain(
		"minimum amount check",
		OrderDomain,
		func(order Order) (bool, error) {
			return order.TotalAmount >= 100.0, nil
		},
	)

	// Add requirement traceability and business description
	UpdateMetadata(minimumAmount, RuleMetadata{
		RequirementID: "JIRA-1234",
		BusinessDescription: "Standard customers must have a minimum order " +
			"amount of $100 to qualify for free shipping",
		Owner:   "Order Team",
		Version: "1.0.0",
		Tags:    []string{"financial", "validation"},
	})

	// Create another rule with requirement tracking
	validCountry := NewWithDomain(
		"valid country check",
		OrderDomain,
		func(order Order) (bool, error) {
			validCountries := map[string]bool{"US": true, "CA": true}
			return validCountries[order.Country], nil
		},
	)

	UpdateMetadata(validCountry, RuleMetadata{
		RequirementID: "JIRA-1235",
		BusinessDescription: "Orders can only be shipped to US and Canada " +
			"due to logistics constraints",
		Owner:   "Shipping Team",
		Version: "2.1.0",
		Tags:    []string{"geographic", "shipping"},
	})

	// Print confirmation
	fmt.Println("Requirements implemented:")
	fmt.Println("- JIRA-1234: Minimum order amount validation")
	fmt.Println("- JIRA-1235: Geographic shipping restrictions")

	// Output:
	// Requirements implemented:
	// - JIRA-1234: Minimum order amount validation
	// - JIRA-1235: Geographic shipping restrictions
}
