# Rules Package - Usage Examples

This document provides comprehensive examples of using the rules package in 
various scenarios.

## Table of Contents

1. [Basic Rule Evaluation](#basic-rule-evaluation)
2. [E-commerce Order Validation](#e-commerce-order-validation)
3. [User Access Control](#user-access-control)
4. [Loan Approval System](#loan-approval-system)
5. [Dynamic Rule Construction](#dynamic-rule-construction)
6. [Error Handling](#error-handling)

## Basic Rule Evaluation

The simplest use case: define and evaluate a single rule.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

func main() {
    // Define a rule
    isAdult := rules.New(
        "is adult",
        func(age int) (bool, error) {
            return age >= 18, nil
        },
    )

    // Evaluate
    satisfied, err := isAdult.Evaluate(25)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Is adult: %v\n", satisfied) // true
}
```

## E-commerce Order Validation

A realistic e-commerce example with hierarchical rules.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type Order struct {
    TotalAmount float64
    CustomerID  string
    Country     string
    ItemCount   int
    IsVIP       bool
    IsPrime     bool
}

func buildOrderEligibilityRules() rules.Rule[Order] {
    // Basic eligibility rules
    minimumAmount := rules.New(
        "minimum order amount",
        func(order Order) (bool, error) {
            return order.TotalAmount >= 50.0, nil
        },
    )

    validCountry := rules.New(
        "valid shipping country",
        func(order Order) (bool, error) {
            validCountries := map[string]bool{
                "US": true, "CA": true, "UK": true, "DE": true,
            }
            return validCountries[order.Country], nil
        },
    )

    hasItems := rules.New(
        "has items",
        func(order Order) (bool, error) {
            return order.ItemCount > 0, nil
        },
    )

    // Standard eligibility: amount + country + items
    standardEligibility := rules.And(
        "standard eligibility",
        minimumAmount,
        validCountry,
        hasItems,
    )

    // Premium customers
    vipCustomer := rules.New(
        "VIP customer",
        func(order Order) (bool, error) {
            return order.IsVIP, nil
        },
    )

    primeCustomer := rules.New(
        "Prime customer",
        func(order Order) (bool, error) {
            return order.IsPrime, nil
        },
    )

    premiumCustomer := rules.Or(
        "premium customer",
        vipCustomer,
        primeCustomer,
    )

    // Premium customers bypass amount requirement
    premiumEligibility := rules.And(
        "premium eligibility",
        premiumCustomer,
        validCountry,
        hasItems,
    )

    // Final rule: standard OR premium eligibility
    return rules.Or(
        "order eligible for processing",
        standardEligibility,
        premiumEligibility,
    )
}

func main() {
    eligibilityRule := buildOrderEligibilityRules()
    evaluator := rules.NewEvaluator(eligibilityRule)

    // Test case 1: Standard customer with valid order
    order1 := Order{
        TotalAmount: 75.0,
        Country:     "US",
        ItemCount:   3,
        IsVIP:       false,
        IsPrime:     false,
    }

    result1 := evaluator.EvaluateDetailed(order1)
    fmt.Printf("Order 1 eligible: %v\n", result1.IsSuccessful())
    fmt.Printf("%s\n\n", result1.String())

    // Test case 2: VIP customer with low amount
    order2 := Order{
        TotalAmount: 20.0, // Below minimum
        Country:     "US",
        ItemCount:   1,
        IsVIP:       true,
        IsPrime:     false,
    }

    result2 := evaluator.EvaluateDetailed(order2)
    fmt.Printf("Order 2 eligible: %v\n", result2.IsSuccessful())
    fmt.Printf("%s\n\n", result2.String())

    // Test case 3: Invalid country
    order3 := Order{
        TotalAmount: 100.0,
        Country:     "XX", // Invalid
        ItemCount:   2,
        IsVIP:       false,
        IsPrime:     false,
    }

    result3 := evaluator.EvaluateDetailed(order3)
    fmt.Printf("Order 3 eligible: %v\n", result3.IsSuccessful())
    fmt.Printf("%s\n\n", result3.String())
}
```

## User Access Control

Implementing role-based access control with business rules.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type User struct {
    ID       string
    Role     string
    IsActive bool
    MFAEnabled bool
}

type Resource struct {
    Name         string
    RequiresMFA  bool
    AllowedRoles []string
}

type AccessRequest struct {
    User     User
    Resource Resource
}

func buildAccessControlRules() rules.Rule[AccessRequest] {
    // User must be active
    userActive := rules.New(
        "user is active",
        func(req AccessRequest) (bool, error) {
            return req.User.IsActive, nil
        },
    )

    // User has required role
    hasRole := rules.New(
        "user has required role",
        func(req AccessRequest) (bool, error) {
            for _, role := range req.Resource.AllowedRoles {
                if req.User.Role == role {
                    return true, nil
                }
            }
            return false, nil
        },
    )

    // MFA check (conditional)
    mfaCompliant := rules.New(
        "MFA compliant",
        func(req AccessRequest) (bool, error) {
            // If resource doesn't require MFA, pass
            if !req.Resource.RequiresMFA {
                return true, nil
            }
            // If resource requires MFA, user must have it enabled
            return req.User.MFAEnabled, nil
        },
    )

    // Combine all rules
    return rules.AllOf(
        "access granted",
        userActive,
        hasRole,
        mfaCompliant,
    )
}

func main() {
    accessRule := buildAccessControlRules()

    // Test case 1: Valid access
    req1 := AccessRequest{
        User: User{
            ID:         "user123",
            Role:       "admin",
            IsActive:   true,
            MFAEnabled: true,
        },
        Resource: Resource{
            Name:         "sensitive-data",
            RequiresMFA:  true,
            AllowedRoles: []string{"admin", "manager"},
        },
    }

    granted1, _ := accessRule.Evaluate(req1)
    fmt.Printf("Access request 1: %v\n", granted1)

    // Test case 2: Missing MFA
    req2 := AccessRequest{
        User: User{
            ID:         "user456",
            Role:       "admin",
            IsActive:   true,
            MFAEnabled: false, // Missing MFA
        },
        Resource: Resource{
            Name:         "sensitive-data",
            RequiresMFA:  true,
            AllowedRoles: []string{"admin", "manager"},
        },
    }

    granted2, _ := accessRule.Evaluate(req2)
    fmt.Printf("Access request 2: %v\n", granted2)
}
```

## Loan Approval System

Complex business rules for loan approval with multiple criteria.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type LoanApplication struct {
    CreditScore       int
    AnnualIncome      float64
    DebtToIncomeRatio float64
    EmploymentYears   int
    HasBankruptcy     bool
    RequestedAmount   float64
}

func buildLoanApprovalRules() rules.Rule[LoanApplication] {
    // Credit score tiers
    excellentCredit := rules.New(
        "excellent credit",
        func(app LoanApplication) (bool, error) {
            return app.CreditScore >= 750, nil
        },
    )

    goodCredit := rules.New(
        "good credit",
        func(app LoanApplication) (bool, error) {
            return app.CreditScore >= 700, nil
        },
    )

    fairCredit := rules.New(
        "fair credit",
        func(app LoanApplication) (bool, error) {
            return app.CreditScore >= 650, nil
        },
    )

    // Income requirements
    sufficientIncome := rules.New(
        "sufficient income",
        func(app LoanApplication) (bool, error) {
            return app.AnnualIncome >= 50000, nil
        },
    )

    highIncome := rules.New(
        "high income",
        func(app LoanApplication) (bool, error) {
            return app.AnnualIncome >= 100000, nil
        },
    )

    // Debt ratio
    lowDebt := rules.New(
        "low debt ratio",
        func(app LoanApplication) (bool, error) {
            return app.DebtToIncomeRatio <= 0.36, nil
        },
    )

    // Employment stability
    stableEmployment := rules.New(
        "stable employment",
        func(app LoanApplication) (bool, error) {
            return app.EmploymentYears >= 2, nil
        },
    )

    // No bankruptcy
    noBankruptcy := rules.New(
        "no bankruptcy",
        func(app LoanApplication) (bool, error) {
            return !app.HasBankruptcy, nil
        },
    )

    // Reasonable loan amount
    reasonableLoan := rules.New(
        "reasonable loan amount",
        func(app LoanApplication) (bool, error) {
            maxLoan := app.AnnualIncome * 3
            return app.RequestedAmount <= maxLoan, nil
        },
    )

    // Tier 1: Excellent credit - more lenient
    tier1 := rules.And(
        "tier 1 approval",
        excellentCredit,
        sufficientIncome,
        noBankruptcy,
        reasonableLoan,
    )

    // Tier 2: Good credit - standard requirements
    tier2 := rules.And(
        "tier 2 approval",
        goodCredit,
        sufficientIncome,
        lowDebt,
        stableEmployment,
        noBankruptcy,
        reasonableLoan,
    )

    // Tier 3: Fair credit - strict requirements
    tier3 := rules.And(
        "tier 3 approval",
        fairCredit,
        highIncome,
        lowDebt,
        stableEmployment,
        noBankruptcy,
        reasonableLoan,
    )

    // Approve if any tier is satisfied
    return rules.AnyOf(
        "loan approval",
        tier1,
        tier2,
        tier3,
    )
}

func main() {
    approvalRule := buildLoanApprovalRules()
    evaluator := rules.NewEvaluator(approvalRule)

    // Test case: Excellent credit applicant
    app1 := LoanApplication{
        CreditScore:       780,
        AnnualIncome:      75000,
        DebtToIncomeRatio: 0.40, // Higher debt but excellent credit
        EmploymentYears:   1,    // Less experience but excellent credit
        HasBankruptcy:     false,
        RequestedAmount:   150000,
    }

    result1 := evaluator.EvaluateDetailed(app1)
    fmt.Printf("Application 1 approved: %v\n", result1.IsSuccessful())
    fmt.Printf("%s\n\n", result1.String())
}
```

## Dynamic Rule Construction

Building rules dynamically based on configuration.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type Product struct {
    Price       float64
    Category    string
    InStock     bool
    Rating      float64
    ReviewCount int
}

type FilterConfig struct {
    MaxPrice      float64
    Categories    []string
    MinRating     float64
    MinReviews    int
    InStockOnly   bool
}

func buildDynamicProductFilter(
    config FilterConfig,
) rules.Rule[Product] {
    builder := rules.NewBuilder[Product]()

    // Add price filter if specified
    if config.MaxPrice > 0 {
        builder.AddCondition(
            fmt.Sprintf("price <= %.2f", config.MaxPrice),
            func(p Product) bool {
                return p.Price <= config.MaxPrice
            },
        )
    }

    // Add category filter if specified
    if len(config.Categories) > 0 {
        categoryMap := make(map[string]bool)
        for _, cat := range config.Categories {
            categoryMap[cat] = true
        }

        builder.AddCondition(
            "in allowed categories",
            func(p Product) bool {
                return categoryMap[p.Category]
            },
        )
    }

    // Add rating filter if specified
    if config.MinRating > 0 {
        builder.AddCondition(
            fmt.Sprintf("rating >= %.1f", config.MinRating),
            func(p Product) bool {
                return p.Rating >= config.MinRating
            },
        )
    }

    // Add review count filter if specified
    if config.MinReviews > 0 {
        builder.AddCondition(
            fmt.Sprintf("reviews >= %d", config.MinReviews),
            func(p Product) bool {
                return p.ReviewCount >= config.MinReviews
            },
        )
    }

    // Add stock filter if specified
    if config.InStockOnly {
        builder.AddCondition(
            "in stock",
            func(p Product) bool {
                return p.InStock
            },
        )
    }

    return builder.BuildAnd("product filter")
}

func main() {
    // Create filter configuration
    config := FilterConfig{
        MaxPrice:    100.0,
        Categories:  []string{"Electronics", "Books"},
        MinRating:   4.0,
        MinReviews:  10,
        InStockOnly: true,
    }

    filter := buildDynamicProductFilter(config)

    // Test products
    products := []Product{
        {
            Price:       79.99,
            Category:    "Electronics",
            InStock:     true,
            Rating:      4.5,
            ReviewCount: 50,
        },
        {
            Price:       150.0, // Too expensive
            Category:    "Electronics",
            InStock:     true,
            Rating:      4.8,
            ReviewCount: 100,
        },
        {
            Price:       29.99,
            Category:    "Toys", // Wrong category
            InStock:     true,
            Rating:      4.2,
            ReviewCount: 20,
        },
    }

    for i, product := range products {
        matches, _ := filter.Evaluate(product)
        fmt.Printf("Product %d matches filter: %v\n", i+1, matches)
    }
}
```

## Error Handling

Proper error handling in rule evaluation.

```go
package main

import (
    "errors"
    "fmt"
    "github.com/tobbstr/the/rules"
)

type ValidationInput struct {
    Data  string
    Valid bool
}

func main() {
    // Rule that might fail
    externalValidation := rules.New(
        "external API validation",
        func(input ValidationInput) (bool, error) {
            // Simulate external API call
            // In reality, this might fail due to network issues
            if input.Data == "trigger-error" {
                return false, errors.New("API unavailable")
            }
            return input.Valid, nil
        },
    )

    dataValidation := rules.New(
        "data validation",
        func(input ValidationInput) (bool, error) {
            if input.Data == "" {
                return false, errors.New("empty data not allowed")
            }
            return true, nil
        },
    )

    combinedRule := rules.And(
        "all validations",
        dataValidation,
        externalValidation,
    )

    // Test 1: Error from rule
    input1 := ValidationInput{Data: "trigger-error", Valid: true}
    _, err1 := combinedRule.Evaluate(input1)
    if err1 != nil {
        fmt.Printf("Validation failed: %v\n", err1)
    }

    // Test 2: Empty data error
    input2 := ValidationInput{Data: "", Valid: true}
    _, err2 := combinedRule.Evaluate(input2)
    if err2 != nil {
        fmt.Printf("Validation failed: %v\n", err2)
    }

    // Test 3: Success
    input3 := ValidationInput{Data: "valid-data", Valid: true}
    satisfied, err3 := combinedRule.Evaluate(input3)
    if err3 != nil {
        fmt.Printf("Validation failed: %v\n", err3)
    } else {
        fmt.Printf("Validation succeeded: %v\n", satisfied)
    }
}
```

## Cross-Type Rule Composition

Combining rules that operate on different input types using mapping functions.

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type User struct {
    ID       string
    Role     string
    IsActive bool
}

type Order struct {
    ID          string
    Amount      float64
    ItemCount   int
}

type Payment struct {
    Method      string
    IsProcessed bool
}

type CompleteOrderRequest struct {
    User    User
    Order   Order
    Payment Payment
}

func main() {
    // Define rules for different types
    userRule := rules.New(
        "user is active admin",
        func(user User) (bool, error) {
            return user.Role == "admin" && user.IsActive, nil
        },
    )

    orderRule := rules.New(
        "order is valid",
        func(order Order) (bool, error) {
            return order.Amount > 0 && order.ItemCount > 0, nil
        },
    )

    paymentRule := rules.New(
        "payment is processed",
        func(payment Payment) (bool, error) {
            return payment.IsProcessed, nil
        },
    )

    // Combine rules from different types using Combine3
    completeValidation := rules.Combine3(
        "complete order validation",
        userRule,
        func(req CompleteOrderRequest) User { return req.User },
        orderRule,
        func(req CompleteOrderRequest) Order { return req.Order },
        paymentRule,
        func(req CompleteOrderRequest) Payment { return req.Payment },
    )

    // Alternative: Use CombineMany for more flexibility
    flexibleValidation := rules.CombineMany(
        "flexible validation",
        rules.Map(
            "user validation",
            userRule,
            func(req CompleteOrderRequest) User { return req.User },
        ),
        rules.Map(
            "order validation",
            orderRule,
            func(req CompleteOrderRequest) Order { return req.Order },
        ),
        rules.Map(
            "payment validation",
            paymentRule,
            func(req CompleteOrderRequest) Payment { return req.Payment },
        ),
    )

    // Test the rules
    request := CompleteOrderRequest{
        User: User{
            ID:       "user-123",
            Role:     "admin",
            IsActive: true,
        },
        Order: Order{
            ID:        "order-456",
            Amount:    150.0,
            ItemCount: 3,
        },
        Payment: Payment{
            Method:      "credit_card",
            IsProcessed: true,
        },
    }

    // Evaluate with detailed results
    evaluator := rules.NewEvaluator(completeValidation)
    result := evaluator.EvaluateDetailed(request)

    if result.IsSuccessful() {
        fmt.Println("✓ Order request approved!")
        fmt.Printf("\nValidation details:\n%s\n", result.String())
    } else {
        fmt.Println("✗ Order request rejected")
        fmt.Printf("\nFailure details:\n%s\n", result.String())
    }

    // Test with flexible validation
    satisfied, _ := flexibleValidation.Evaluate(request)
    fmt.Printf("\nFlexible validation: %v\n", satisfied)
}
```

### Advanced Cross-Type Patterns

You can also create reusable extractors:

```go
// Define extractors as functions
func extractUser(req CompleteOrderRequest) User {
    return req.User
}

func extractOrder(req CompleteOrderRequest) Order {
    return req.Order
}

func extractPayment(req CompleteOrderRequest) Payment {
    return req.Payment
}

// Use them in rule composition
validation := rules.CombineMany(
    "complete validation",
    rules.Map("user check", userRule, extractUser),
    rules.Map("order check", orderRule, extractOrder),
    rules.Map("payment check", paymentRule, extractPayment),
)
```

This pattern is especially useful when:
- You have domain models separated by concern
- You want to reuse rules across different aggregate types
- You need to validate composite requests with multiple entities

## Best Practices Summary

1. **Use Descriptive Names**: Make rules self-documenting
2. **Keep Rules Focused**: Each rule should check one condition
3. **Compose Hierarchically**: Build complex logic from simple rules
4. **Handle Context**: Always respect context cancellation
5. **Use Builder for Dynamic Rules**: When rules change based on config
6. **Test Thoroughly**: Use table-driven tests
7. **Log at Boundaries**: Log evaluation results at system edges
8. **Use Detailed Evaluation**: For debugging and auditing
9. **Map for Cross-Type Rules**: Use `Map()` and `Combine()` when working 
   with different types

