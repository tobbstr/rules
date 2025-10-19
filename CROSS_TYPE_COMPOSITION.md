# Cross-Type Rule Composition

This guide explains how to combine rules that operate on different input types.

## Problem

You often need to validate complex requests that contain multiple domain entities:

```go
type OrderRequest struct {
    User    User
    Order   Order
    Payment Payment
}
```

Each entity has its own validation rules, but they were defined for their specific types:
- `Rule[User]` for user validation
- `Rule[Order]` for order validation  
- `Rule[Payment]` for payment validation

**How do you combine them into a single `Rule[OrderRequest]`?**

## Solution: Mapping Functions

Use the `Map()` function to transform rules between types using extractor functions.

### Basic Mapping

Transform a rule from one type to another:

```go
// Rule that operates on User
userRule := rules.New(
    "user is admin",
    func(user User) (bool, error) {
        return user.Role == "admin", nil
    },
)

// Transform to operate on OrderRequest
requestRule := rules.Map(
    "request from admin",
    userRule,
    func(req OrderRequest) User { return req.User }, // Extractor
)

// Now it's a Rule[OrderRequest]
satisfied, err := requestRule.Evaluate(ctx, orderRequest)
```

### Combining Two Types

Use `Combine()` for two different types:

```go
combined := rules.Combine(
    "valid request",
    userRule,                                        // Rule[User]
    func(req OrderRequest) User { return req.User }, // Extract User
    orderRule,                                       // Rule[Order]
    func(req OrderRequest) Order { return req.Order }, // Extract Order
)

// Result: Rule[OrderRequest]
```

### Combining Three Types

Use `Combine3()` for three different types:

```go
combined := rules.Combine3(
    "complete validation",
    userRule,    func(req OrderRequest) User { return req.User },
    orderRule,   func(req OrderRequest) Order { return req.Order },
    paymentRule, func(req OrderRequest) Payment { return req.Payment },
)
```

### Combining Many Types

For more flexibility, use `CombineMany()` with explicit mapping:

```go
combined := rules.CombineMany(
    "all validations",
    rules.Map("user check", userRule,
        func(req OrderRequest) User { return req.User }),
    rules.Map("order check", orderRule,
        func(req OrderRequest) Order { return req.Order }),
    rules.Map("payment check", paymentRule,
        func(req OrderRequest) Payment { return req.Payment }),
    rules.Map("shipping check", shippingRule,
        func(req OrderRequest) Shipping { return req.Shipping }),
)
```

## Complete Example

```go
package main

import (    "fmt"
    "github.com/tobbstr/rules"
)

type User struct {
    Role     string
    IsActive bool
}

type Order struct {
    Amount    float64
    ItemCount int
}

type OrderRequest struct {
    User  User
    Order Order
}

func main() {
    // Define rules for individual types
    activeUserRule := rules.New(
        "user is active",
        func(user User) (bool, error) {
            return user.IsActive, nil
        },
    )

    validOrderRule := rules.New(
        "order is valid",
        func(order Order) (bool, error) {
            return order.Amount > 0 && order.ItemCount > 0, nil
        },
    )

    // Combine into a single rule for OrderRequest
    requestValidation := rules.Combine(
        "valid order request",
        activeUserRule,
        func(req OrderRequest) User { return req.User },
        validOrderRule,
        func(req OrderRequest) Order { return req.Order },
    )

    // Use it
    request := OrderRequest{
        User:  User{IsActive: true, Role: "customer"},
        Order: Order{Amount: 100.0, ItemCount: 3},
    }

    satisfied, err := requestValidation.Evaluate(request)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Request valid: %v\n", satisfied)
}
```

## Pattern: Reusable Extractors

For cleaner code, define extractors as named functions:

```go
// Define extractors
func extractUser(req OrderRequest) User {
    return req.User
}

func extractOrder(req OrderRequest) Order {
    return req.Order
}

func extractPayment(req OrderRequest) Payment {
    return req.Payment
}

// Use them
validation := rules.CombineMany(
    "complete validation",
    rules.Map("user", userRule, extractUser),
    rules.Map("order", orderRule, extractOrder),
    rules.Map("payment", paymentRule, extractPayment),
)
```

## Pattern: Hierarchical Cross-Type Rules

You can combine mapped rules with other logical operators:

```go
// User must be either admin OR active with valid order
authorization := rules.Or(
    "authorized request",
    rules.Map("is admin", adminRule, extractUser),
    rules.And(
        "active with valid order",
        rules.Map("is active", activeRule, extractUser),
        rules.Map("valid order", orderRule, extractOrder),
    ),
)
```

## When to Use This Pattern

Use cross-type composition when:

1. **Separated Domain Models**: Your domain entities are properly separated
   ```go
   // Good separation
   type User struct { ... }
   type Order struct { ... }
   type OrderRequest struct { User, Order }
   ```

2. **Reusable Rules**: You want to define rules once and reuse them
   ```go
   // Define once
   userRule := rules.New("active user", ...)
   
   // Use in multiple contexts
   orderValidation := rules.Combine(..., userRule, ...)
   refundValidation := rules.Combine(..., userRule, ...)
   ```

3. **Aggregate Validation**: You're validating composite/aggregate requests
   ```go
   type CheckoutRequest struct {
       Cart     Cart
       User     User
       Payment  Payment
       Shipping Shipping
   }
   ```

4. **Microservices**: Rules come from different services/domains
   ```go
   // User rules from auth service
   // Order rules from order service
   // Payment rules from payment service
   // Combine them in API gateway
   ```

## Type Safety

The mapping functions are fully type-safe:

```go
// Compile-time error if types don't match
userRule := rules.New(...) // Rule[User]
badMap := rules.Map(
    "bad",
    userRule,
    func(req OrderRequest) Order { return req.Order }, // ❌ Won't compile
)

// Correct
goodMap := rules.Map(
    "good",
    userRule,
    func(req OrderRequest) User { return req.User }, // ✓ Type safe
)
```

## Performance

Mapping adds minimal overhead:
- Single function call per evaluation
- No allocations
- Extractors are typically just field access

```go
// This is very fast
func extractUser(req OrderRequest) User {
    return req.User  // Just a field access
}
```

## Testing Mapped Rules

Test mapped rules like any other rule:

```go
func TestMappedRule(t *testing.T) {
    userRule := rules.New("active", ...)
    mappedRule := rules.Map("mapped", userRule, extractUser)

    tests := []struct {
        name    string
        request OrderRequest
        want    bool
    }{
        {"active user", OrderRequest{User: User{IsActive: true}}, true},
        {"inactive user", OrderRequest{User: User{IsActive: false}}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, _ := mappedRule.Evaluate(tt.request)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## API Reference

### Map Function

```go
func Map[TSource, TTarget any](
    name string,
    rule Rule[TTarget],
    mapper MapperFunc[TSource, TTarget],
) Rule[TSource]
```

Transforms a `Rule[TTarget]` into a `Rule[TSource]`.

### Combine Function

```go
func Combine[TCombined, T1, T2 any](
    name string,
    rule1 Rule[T1],
    extractor1 MapperFunc[TCombined, T1],
    rule2 Rule[T2],
    extractor2 MapperFunc[TCombined, T2],
) Rule[TCombined]
```

Combines two rules from different types into an AND rule.

### Combine3 Function

```go
func Combine3[TCombined, T1, T2, T3 any](
    name string,
    rule1 Rule[T1],
    extractor1 MapperFunc[TCombined, T1],
    rule2 Rule[T2],
    extractor2 MapperFunc[TCombined, T2],
    rule3 Rule[T3],
    extractor3 MapperFunc[TCombined, T3],
) Rule[TCombined]
```

Combines three rules from different types into an AND rule.

### CombineMany Function

```go
func CombineMany[TCombined any](
    name string,
    mappedRules ...Rule[TCombined],
) Rule[TCombined]
```

Combines multiple already-mapped rules into an AND rule.

## Summary

Cross-type rule composition allows you to:
- ✅ Combine rules from different domain entities
- ✅ Maintain type safety throughout
- ✅ Reuse rules across different contexts
- ✅ Keep domain models properly separated
- ✅ Build complex validation logic from simple parts

This makes the rules library suitable for real-world microservices and 
domain-driven design architectures.

