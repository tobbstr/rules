# Rules Package

A type-safe, expressive library for defining and evaluating hierarchical 
business rules in Go.

## Features

- **Type-safe**: Uses Go generics for compile-time type safety
- **Hierarchical**: Rules can contain other rules, forming complex logic trees
- **Expressive**: Fluent API and helper functions for readable rule definitions
- **Detailed results**: Comprehensive evaluation results with timing and 
  error information
- **Composable**: Combine rules using AND, OR, NOT logical operators
- **Well-tested**: 100% test coverage on public APIs

## Installation

```bash
go get github.com/tobbstr/the/rules
```

## Quick Start

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type Order struct {
    Amount  float64
    IsVIP   bool
    Country string
}

func main() {
    // Define a simple rule
    minimumAmount := rules.New(
        "minimum amount",
        func(order Order) (bool, error) {
            return order.Amount >= 100.0, nil
        },
    )

    // Evaluate the rule
    order := Order{Amount: 150.0}
    satisfied, err := minimumAmount.Evaluate(order)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Rule satisfied: %v\n", satisfied)
}
```

## Core Concepts

### Rules

A `Rule[T]` is an interface that can evaluate a condition against an input 
of type `T`:

```go
type Rule[T any] interface {
    Evaluate(input T) (bool, error)
    Name() string
    Description() string
}
```

### Simple Rules

Create basic rules using the `New` function:

```go
rule := rules.New(
    "value greater than 10",
    func(input int) (bool, error) {
        return input > 10, nil
    },
)
```

### Hierarchical Rules

Combine rules using logical operators:

```go
// AND: All rules must be satisfied
andRule := rules.And("all requirements", rule1, rule2, rule3)

// OR: At least one rule must be satisfied
orRule := rules.Or("any requirement", rule1, rule2, rule3)

// NOT: Rule must not be satisfied
notRule := rules.Not("not premium", premiumRule)
```

## Builder Pattern

Use the builder for a fluent API:

```go
builder := rules.NewBuilder[Order]()

builder.
    AddCondition("minimum amount", func(o Order) bool {
        return o.Amount >= 100
    }).
    AddCondition("valid country", func(o Order) bool {
        return o.Country == "US"
    })

rule := builder.BuildAnd("order requirements")
```

## Helper Functions

The package provides helper functions for common patterns:

### AllOf, AnyOf, NoneOf

```go
// All rules must be satisfied
allRule := rules.AllOf("all", rule1, rule2, rule3)

// At least one rule must be satisfied
anyRule := rules.AnyOf("any", rule1, rule2, rule3)

// None of the rules should be satisfied
noneRule := rules.NoneOf("none", rule1, rule2, rule3)
```

### Quantifiers

```go
// At least N rules must be satisfied
atLeastRule := rules.AtLeast("at least 2", 2, rule1, rule2, rule3)

// Exactly N rules must be satisfied
exactlyRule := rules.Exactly("exactly 2", 2, rule1, rule2, rule3)

// At most N rules must be satisfied
atMostRule := rules.AtMost("at most 2", 2, rule1, rule2, rule3)
```

### Always and Never

```go
// Rule that always succeeds
alwaysRule := rules.Always[Order]("always")

// Rule that never succeeds
neverRule := rules.Never[Order]("never")
```

## Cross-Type Rule Composition

Combine rules that operate on different types using the `Map` function and 
extractors:

### Mapping Rules

Transform a rule from one type to another:

```go
// Rule that operates on User
userRule := rules.New(
    "user is admin",
    func(user User) (bool, error) {
        return user.Role == "admin", nil
    },
)

// Map it to operate on Request (which contains User)
requestRule := rules.Map(
    "request from admin",
    userRule,
    func(req Request) User { return req.User },
)
```

### Combining Different Types

Combine rules from different type hierarchies:

```go
type OrderRequest struct {
    User  User
    Order Order
}

userRule := rules.New("user is active", ...)
orderRule := rules.New("order is valid", ...)

// Combine using extractors
combined := rules.Combine(
    "valid order request",
    userRule,
    func(req OrderRequest) User { return req.User },
    orderRule,
    func(req OrderRequest) Order { return req.Order },
)
```

### Multiple Types

For more than two types:

```go
// Combine 3 different types
combined := rules.Combine3(
    "complete validation",
    userRule,
    func(req CompleteRequest) User { return req.User },
    orderRule,
    func(req CompleteRequest) Order { return req.Order },
    shippingRule,
    func(req CompleteRequest) Shipping { return req.Shipping },
)

// Or use CombineMany for flexibility
combined := rules.CombineMany(
    "all checks",
    rules.Map("user check", userRule, extractUser),
    rules.Map("order check", orderRule, extractOrder),
    rules.Map("payment check", paymentRule, extractPayment),
)
```

## Detailed Evaluation

Use an `Evaluator` to get detailed results:

```go
evaluator := rules.NewEvaluator(complexRule)
result := evaluator.EvaluateDetailed(ctx, input)

// Check the result
if result.IsSuccessful() {
    fmt.Println("All rules satisfied!")
}

// Print detailed results
fmt.Println(result.String())

// Example output:
// ✓ order eligibility (took 123µs)
//   ✓ amount >= 100 (took 45µs)
//   ✗ items >= 3 (took 38µs)
//   ✓ valid country (took 40µs)
```

## Complete Example

```go
package main

import (    "fmt"
    "github.com/tobbstr/the/rules"
)

type Order struct {
    TotalAmount float64
    IsVIP       bool
    ItemCount   int
    Country     string
}

func main() {
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

    // Evaluate with detailed results
    evaluator := rules.NewEvaluator(eligibility)
    
    order := Order{
        TotalAmount: 150.0,
        IsVIP:       false,
        Country:     "US",
    }

    result := evaluator.EvaluateDetailed(order)

    if result.IsSuccessful() {
        fmt.Println("Order is eligible!")
        fmt.Printf("\nDetails:\n%s\n", result.String())
    } else {
        fmt.Println("Order is not eligible")
    }
}
```

## Error Handling

The package follows Go best practices:

- All errors are wrapped with context using `%w`
- Sentinel errors are provided for common cases:
  - `ErrNilRule`: When a nil rule is provided
  - `ErrEmptyRules`: When an empty rules list is provided
  - `ErrEvaluationFailed`: When rule evaluation fails

```go
satisfied, err := rule.Evaluate(input)
if err != nil {
    if errors.Is(err, rules.ErrNilRule) {
        // Handle nil rule error
    }
    // Handle other errors
}
```

## Best Practices

1. **Use descriptive names**: Make debugging easier with clear rule names
2. **Keep predicates simple**: Focus each rule on a single condition
3. **Compose hierarchically**: Build complex logic from simple rules
4. **Log at boundaries**: Log errors at the root of the call stack
6. **Use the builder**: For dynamic rule construction
7. **Test thoroughly**: Use table-driven tests for rule validation
8. **Map for cross-type composition**: Use `Map()` when combining rules 
   that operate on different types

## Performance Considerations

- Rules are evaluated lazily (short-circuit evaluation)
- AND rules stop at first failure
- OR rules stop at first success
- Use context timeouts for expensive predicates
- Consider caching expensive rule evaluations

## License

Same as the parent project.

