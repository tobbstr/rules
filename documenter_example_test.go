package rules_test

import (
	"fmt"
	"log"

	"github.com/tobbstr/rules"
)

// Define domains for different business areas
const (
	OrderDomain    rules.Domain = "order"
	UserDomain     rules.Domain = "user"
	PaymentDomain  rules.Domain = "payment"
	ShippingDomain rules.Domain = "shipping"
)

// DocOrder represents an order in the system for documentation examples
type DocOrder struct {
	ID         string
	Amount     float64
	Currency   string
	CustomerID string
	Items      int
	Country    string
}

// DocUser represents a user in the system for documentation examples
type DocUser struct {
	ID      string
	Email   string
	IsVIP   bool
	IsPrime bool
	Status  string
	Age     int
}

// example_documentationGeneration demonstrates the complete documentation generation workflow
// with multiple domains, hierarchical rules, and various output formats.
// Note: This is a non-testable example due to variable timestamps in output.
func example_documentationGeneration() {
	// Clear registry for clean example
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// ===== ORDER DOMAIN RULES =====

	// Create order validation rules
	minAmount := rules.NewWithDomain("minimum amount", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Amount >= 10.0, nil
	})
	rules.WithDescription(minAmount, "Order must meet minimum amount of $10")

	hasItems := rules.NewWithDomain("has items", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Items > 0, nil
	})
	rules.WithDescription(hasItems, "Order must contain at least one item")

	validCurrency := rules.NewWithDomain("valid currency", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Currency == "USD" || o.Currency == "EUR", nil
	})
	rules.WithDescription(validCurrency, "Order currency must be USD or EUR")

	// Create composite rule
	validOrder := rules.And("valid order", minAmount, hasItems, validCurrency)
	_ = rules.Register(validOrder,
		rules.WithDomain(OrderDomain),
		rules.WithGroup("Order Validation"))
	rules.WithDescription(validOrder, "Order passes all validation checks")

	// ===== USER DOMAIN RULES =====

	// Create user validation rules
	validEmail := rules.NewWithDomain("valid email", UserDomain, func(u DocUser) (bool, error) {
		return len(u.Email) > 0 && len(u.Email) < 255, nil
	})
	rules.WithDescription(validEmail, "User must have a valid email address")

	activeUser := rules.NewWithDomain("active user", UserDomain, func(u DocUser) (bool, error) {
		return u.Status == "active", nil
	})
	rules.WithDescription(activeUser, "User account must be active")

	adultUser := rules.NewWithDomain("adult user", UserDomain, func(u DocUser) (bool, error) {
		return u.Age >= 18, nil
	})
	rules.WithDescription(adultUser, "User must be 18 years or older")

	// Create user validation composite
	validUser := rules.And("valid user", validEmail, activeUser, adultUser)
	_ = rules.Register(validUser,
		rules.WithDomain(UserDomain),
		rules.WithGroup("User Validation"))
	rules.WithDescription(validUser, "User passes all validation checks")

	// ===== PAYMENT DOMAIN RULES =====

	vipDiscount := rules.NewWithDomain("VIP discount eligible", PaymentDomain, func(u DocUser) (bool, error) {
		return u.IsVIP, nil
	})
	rules.WithDescription(vipDiscount, "VIP users receive automatic discount")

	primeShipping := rules.NewWithDomain("Prime free shipping", PaymentDomain, func(u DocUser) (bool, error) {
		return u.IsPrime, nil
	})
	rules.WithDescription(primeShipping, "Prime members get free shipping")

	specialOffer := rules.Or("special offer", vipDiscount, primeShipping)
	_ = rules.Register(specialOffer,
		rules.WithDomain(PaymentDomain),
		rules.WithGroup("Promotions"))
	rules.WithDescription(specialOffer, "User qualifies for special offers")

	// ===== GENERATE DOCUMENTATION IN MULTIPLE FORMATS =====

	fmt.Println("=== Markdown Documentation ===")
	md, err := rules.GenerateMarkdown(rules.DocumentOptions{
		Title:         "Business Rules Documentation",
		Description:   "Complete documentation of all business rules across domains",
		GroupByDomain: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	fmt.Println("\n=== JSON Documentation ===")
	json, err := rules.GenerateJSON(rules.DocumentOptions{
		Title:           "Business Rules",
		GroupByDomain:   true,
		IncludeMetadata: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(json)

	fmt.Println("\n=== HTML Documentation ===")
	html, err := rules.GenerateHTML(rules.DocumentOptions{
		Title:         "Business Rules",
		GroupByDomain: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("HTML Generated:", len(html), "bytes")

	fmt.Println("\n=== Mermaid Diagram ===")
	mermaid, err := rules.GenerateMermaid(rules.DocumentOptions{
		GroupByDomain: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mermaid)
}

// example_domainSpecificDocumentation demonstrates generating documentation for specific domains.
// Note: This is a non-testable example due to variable timestamps in output.
func example_domainSpecificDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create rules in multiple domains
	_ = rules.NewWithDomain("order rule 1", OrderDomain, func(o DocOrder) (bool, error) {
		return true, nil
	})

	_ = rules.NewWithDomain("order rule 2", OrderDomain, func(o DocOrder) (bool, error) {
		return true, nil
	})

	_ = rules.NewWithDomain("user rule 1", UserDomain, func(u DocUser) (bool, error) {
		return true, nil
	})

	// Document only Order domain
	fmt.Println("=== Order Domain Only ===")
	md, err := rules.GenerateDomainMarkdown(OrderDomain, rules.DocumentOptions{
		Title: "Order Domain Rules",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Document multiple specific domains
	fmt.Println("\n=== Multiple Domains ===")
	md, err = rules.GenerateDomainsMarkdown(
		[]rules.Domain{OrderDomain, UserDomain},
		rules.DocumentOptions{
			Title:         "Order and User Rules",
			GroupByDomain: true,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Output:
	// Domain-specific documentation generated
}

// example_groupBasedDocumentation demonstrates organizing rules by groups.
// Note: This is a non-testable example due to variable timestamps in output.
func example_groupBasedDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create rules in different groups
	_ = rules.NewWithGroup(
		"validation rule 1",
		"Input Validation",
		[]rules.Domain{OrderDomain},
		func(o DocOrder) (bool, error) {
			return o.Amount > 0, nil
		},
	)

	_ = rules.NewWithGroup(
		"validation rule 2",
		"Input Validation",
		[]rules.Domain{OrderDomain},
		func(o DocOrder) (bool, error) {
			return o.Items > 0, nil
		},
	)

	_ = rules.NewWithGroup(
		"security rule 1",
		"Security",
		[]rules.Domain{UserDomain},
		func(u DocUser) (bool, error) {
			return u.Status == "active", nil
		},
	)

	// Document specific group
	fmt.Println("=== Input Validation Group ===")
	md, err := rules.GenerateGroupMarkdown("Input Validation", rules.DocumentOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Output:
	// Group-specific documentation generated
}

// example_filteringDocumentation demonstrates filtering rules by domain.
// Note: This is a non-testable example due to variable timestamps in output.
func example_filteringDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create rules in multiple domains
	_ = rules.NewWithDomain("order rule", OrderDomain, func(o DocOrder) (bool, error) {
		return true, nil
	})

	_ = rules.NewWithDomain("user rule", UserDomain, func(u DocUser) (bool, error) {
		return true, nil
	})

	_ = rules.NewWithDomain("payment rule", PaymentDomain, func(p DocOrder) (bool, error) {
		return true, nil
	})

	// Include only specific domains
	fmt.Println("=== Include Order and User ===")
	md, err := rules.GenerateMarkdown(rules.DocumentOptions{
		IncludeDomains: []rules.Domain{OrderDomain, UserDomain},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Exclude specific domains
	fmt.Println("\n=== Exclude Payment ===")
	md, err = rules.GenerateMarkdown(rules.DocumentOptions{
		ExcludeDomains: []rules.Domain{PaymentDomain},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Output:
	// Filtered documentation generated
}

// example_metadataDocumentation demonstrates including metadata in documentation.
// Note: This is a non-testable example due to variable timestamps in output.
func example_metadataDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create a rule with rich metadata
	rule := rules.NewWithDomain("premium order", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Amount >= 100.0, nil
	})

	// Add comprehensive metadata
	_ = rules.UpdateMetadata(rule, rules.RuleMetadata{
		Owner:        "Order Team",
		Version:      "2.0.0",
		Tags:         []string{"premium", "high-value", "priority"},
		Dependencies: []rules.Domain{UserDomain, PaymentDomain},
		RelatedRules: []string{"VIP discount", "expedited shipping"},
	})
	rules.WithDescription(rule, "Identifies premium orders for special handling")

	// Generate documentation with metadata
	md, err := rules.GenerateMarkdown(rules.DocumentOptions{
		Title:           "Premium Order Rules",
		IncludeMetadata: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Output:
	// Documentation with metadata generated
}

// example_hierarchicalRulesDocumentation demonstrates documenting complex rule hierarchies.
// Note: This is a non-testable example due to variable timestamps in output.
func example_hierarchicalRulesDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create leaf rules
	hasStock := rules.NewWithDomain("has stock", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Items > 0, nil
	})
	rules.WithDescription(hasStock, "Product is in stock")

	validPrice := rules.NewWithDomain("valid price", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Amount > 0, nil
	})
	rules.WithDescription(validPrice, "Price is positive")

	validCountry := rules.NewWithDomain("valid country", OrderDomain, func(o DocOrder) (bool, error) {
		return o.Country != "", nil
	})
	rules.WithDescription(validCountry, "Shipping country specified")

	// Build hierarchy
	canShip := rules.And("can ship", hasStock, validCountry)
	_ = rules.Register(canShip, rules.WithDomain(OrderDomain))
	rules.WithDescription(canShip, "Order can be shipped")

	validCheckout := rules.And("valid checkout", validPrice, canShip)
	_ = rules.Register(validCheckout, rules.WithDomain(OrderDomain))
	rules.WithDescription(validCheckout, "Order ready for checkout")

	// Generate hierarchical documentation
	md, err := rules.GenerateMarkdown(rules.DocumentOptions{
		Title: "Order Processing Rules",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Generate Mermaid diagram to visualize hierarchy
	mermaid, err := rules.GenerateMermaid(rules.DocumentOptions{
		GroupByDomain: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n", mermaid)

	// Output:
	// Hierarchical documentation generated
}

// example_depthLimitedDocumentation demonstrates limiting documentation depth.
// Note: This is a non-testable example due to variable timestamps in output.
func example_depthLimitedDocumentation() {
	// Clear registry
	rules.DefaultRegistry.Clear()
	defer rules.DefaultRegistry.Clear()

	// Create deeply nested rules
	leaf := rules.New("leaf rule", func(o DocOrder) (bool, error) {
		return true, nil
	})

	level2 := rules.Not("level 2", leaf)
	_ = rules.Register(level2, rules.WithDomain(OrderDomain))

	level1 := rules.And("level 1", level2)
	_ = rules.Register(level1, rules.WithDomain(OrderDomain))

	// Document with depth limit
	md, err := rules.GenerateMarkdown(rules.DocumentOptions{
		Title:    "Limited Depth Documentation",
		MaxDepth: 2, // Only show 2 levels deep
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)

	// Output:
	// Depth-limited documentation generated
}
