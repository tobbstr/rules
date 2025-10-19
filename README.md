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
- **Well-tested**: 94.4% test coverage with 239 test cases

## Installation

```bash
go get github.com/tobbstr/rules
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/tobbstr/rules"
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

## Use Cases

### When to Use This Library

This library is an excellent fit for:

#### 1. **Business Rule Validation**
Complex validation logic for orders, applications, requests, or any domain objects with multiple interdependent conditions.
- E-commerce order eligibility (minimum amounts, shipping restrictions, customer tiers)
- Loan and credit approval systems (credit scores, income requirements, debt ratios)
- Insurance underwriting rules (risk factors, coverage limits, exclusions)
- Regulatory compliance checks (KYC, AML, data residency)

#### 2. **Access Control & Authorization**
Multi-factor authorization decisions that go beyond simple role checks.
- Role-based access control (RBAC) with conditional requirements
- Attribute-based access control (ABAC) with complex policies
- Resource-specific permissions with contextual rules
- MFA and security posture requirements

#### 3. **Dynamic Filtering & Matching**
Building configurable filters from user preferences or system configuration.
- Product catalog filtering (price, category, ratings, availability)
- Search result refinement with multiple criteria
- Recommendation system eligibility rules
- Content moderation policies

#### 4. **Workflow & State Validation**
Validating whether entities can transition between states or proceed in workflows.
- Order lifecycle validation (can ship? can cancel? can refund?)
- User onboarding progress gates
- Multi-step form validation with dependencies
- Feature flag and experiment eligibility

#### 5. **Pricing & Discount Rules**
Determining eligibility for promotions, discounts, or special pricing.
- Customer tier-based pricing
- Volume discount eligibility
- Promotional campaign rules
- Dynamic pricing conditions

#### 6. **Audit & Compliance**
When you need detailed records of why decisions were made.
- Detailed evaluation trails for auditing
- Compliance documentation generation
- Decision explanation for transparency
- Regulatory reporting requirements

#### 7. **Testing & Validation**
As part of your test suite for complex domain logic.
- Table-driven tests with hierarchical rule evaluation
- Integration tests for business logic
- Regression testing of rule changes
- Documentation of business requirements

### When NOT to Use This Library

This library may not be the best choice for:

#### 1. **Simple Validation**
For straightforward single-condition checks, this library adds unnecessary overhead.
❌ **Don't use**: Checking if `age >= 18` as a standalone operation  
✅ **Use instead**: Direct conditional logic or simple validator functions

#### 2. **High-Frequency, Low-Latency Operations**
Performance-critical hot paths where every microsecond matters.
❌ **Don't use**: Request routing, packet filtering, tight game loops  
✅ **Use instead**: Optimized direct code, lookup tables, or specialized libraries

The library's detailed evaluation and hierarchical structure add overhead (typically microseconds per rule, but can accumulate).

#### 3. **Rules Requiring Side Effects**
Rules that need to modify state, make database calls, or trigger actions.
❌ **Don't use**: Updating inventory, sending emails, logging transactions  
✅ **Use instead**: Command pattern, service layer methods, event handlers

Rules should be pure predicates that evaluate conditions, not perform actions.

#### 4. **Frequently Changing Rules**
Business logic that changes multiple times per day or needs no-code editing.
❌ **Don't use**: A/B test variants, ML model outputs, user-customizable logic  
✅ **Use instead**: Rule engines with UI (Drools, Easy Rules), feature flag systems

This library requires code changes and deployments to modify rules.

#### 5. **Complex State Machines**
Systems with many states, transitions, and temporal logic.
❌ **Don't use**: Workflow orchestration, saga patterns, multi-step transactions  
✅ **Use instead**: State machine libraries, workflow engines (Temporal, Cadence)

While you can represent states as rules, dedicated state machine libraries are more appropriate.

#### 6. **Database Query Optimization**
Translating business rules to efficient database queries.
❌ **Don't use**: Filtering millions of records, complex joins, aggregations  
✅ **Use instead**: Query builders, ORMs, database-native features

Evaluate rules in code only after retrieving relevant data.

#### 7. **Real-Time Stream Processing**
Processing high-volume event streams with complex event patterns.
❌ **Don't use**: IoT data processing, log aggregation, metrics pipelines  
✅ **Use instead**: Stream processing frameworks (Flink, Kafka Streams)

The library evaluates individual inputs, not event patterns over time.

#### 8. **Machine Learning Integration**
Rules that depend on ML model predictions or probabilistic decisions.
❌ **Don't use**: Primary decision mechanism alongside ML scores  
✅ **Consider carefully**: May work for eligibility checks before/after ML inference

If your logic is mostly ML-driven, a simpler approach may suffice.

#### 9. **Complex Algorithmic Logic**
Algorithms with loops, recursion, or intermediate computations.
❌ **Don't use**: Graph traversal, optimization problems, data transformations  
✅ **Use instead**: Direct algorithmic implementation

Rules are for evaluating conditions, not computing results.

#### 10. **Embedded Systems or Resource-Constrained Environments**
Environments with strict memory or CPU constraints.
❌ **Don't use**: Microcontrollers, edge devices, mobile apps with tight budgets  
✅ **Consider carefully**: Evaluate overhead; the library uses reflection and generics

### Decision Guide

**Use this library when:**
- ✅ You have complex, hierarchical business logic
- ✅ Rules compose from multiple conditions (AND, OR, NOT)
- ✅ You need detailed evaluation results for auditing
- ✅ Business logic needs clear documentation
- ✅ Rules are relatively stable (hours to weeks between changes)
- ✅ Performance overhead of microseconds per rule is acceptable
- ✅ You want type-safe, compile-time checked rules

**Don't use this library when:**
- ❌ Simple if/else statements suffice
- ❌ Rules need no-code editing by non-developers
- ❌ Every nanosecond of performance matters
- ❌ Rules require side effects or state mutations
- ❌ You're building a workflow engine or state machine
- ❌ Rules change multiple times per day
- ❌ You need to evaluate rules in database queries

## Core Concepts

### Rules

A `Rule[T]` is an interface that can evaluate a condition against an input 
of type `T`:

```go
type Rule[T any] interface {
    Evaluate(input T) (bool, error)
    Name() string
}
```

**Note**: Rule descriptions are managed through the registry system using 
`WithDescription()`, not as interface methods. This allows descriptions to be 
updated without modifying rule implementations.

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

## Documentation Generation

The rules package includes a powerful documentation generation system that can automatically produce comprehensive documentation from your business rules in multiple formats.

### Key Features

- **Multiple Formats**: Generate Markdown, JSON, HTML, and Mermaid diagrams
- **Domain-Driven Organization**: Group rules by business domains
- **Auto-Registration**: Rules automatically register themselves for documentation
- **Hierarchical Visualization**: Shows parent-child relationships
- **Interactive HTML**: Collapsible sections, search, and navigation
- **Rich Metadata**: Owners, versions, tags, dependencies

### Quick Start

```go
// Define domains
const (
    OrderDomain rules.Domain = "order"
    UserDomain  rules.Domain = "user"
)

// Create rules with domains - they auto-register
minAmount := rules.NewWithDomain(
    "minimum amount",
    OrderDomain,
    func(o Order) (bool, error) {
        return o.Amount >= 100, nil
    },
)
rules.WithDescription(minAmount, "Order must meet minimum amount")

// Generate documentation
md, err := rules.GenerateMarkdown(rules.DocumentOptions{
    Title:         "Business Rules",
    GroupByDomain: true,
})

html, err := rules.GenerateHTML(rules.DocumentOptions{
    Title:           "Business Rules",
    IncludeMetadata: true,
})

mermaid, err := rules.GenerateMermaid(rules.DocumentOptions{
    GroupByDomain: true,
})
```

### Domain-Based Organization

Organize rules by business domains for better maintainability:

```go
// Create domain-specific rules
orderRule := rules.NewWithDomain("valid order", OrderDomain, ...)
userRule := rules.NewWithDomain("active user", UserDomain, ...)

// Generate documentation for specific domains
md, err := rules.GenerateDomainMarkdown(OrderDomain, rules.DocumentOptions{})

// Or multiple domains
md, err := rules.GenerateDomainsMarkdown(
    []rules.Domain{OrderDomain, UserDomain},
    rules.DocumentOptions{GroupByDomain: true},
)
```

### Group-Based Organization

Use groups for cross-domain categorization:

```go
validationRule := rules.NewWithGroup(
    "email validation",
    "Input Validation",  // Group name
    []rules.Domain{UserDomain},
    func(u User) (bool, error) {
        return validateEmail(u.Email), nil
    },
)

// Document by group
md, err := rules.GenerateGroupMarkdown("Input Validation", rules.DocumentOptions{})
```

### Rich Metadata

Add comprehensive metadata to your rules, including requirement traceability:

```go
rule := rules.NewWithDomain("premium order", OrderDomain, ...)

rules.UpdateMetadata(rule, rules.RuleMetadata{
    RequirementID:       "JIRA-1234",  // Link to external requirement
    BusinessDescription: "Premium customers get free expedited shipping " +
                        "on orders over $50",  // Plain English description
    Owner:               "Order Team",
    Version:             "2.0.0",
    Tags:                []string{"premium", "high-value"},
    Dependencies:        []rules.Domain{UserDomain, PaymentDomain},
    RelatedRules:        []string{"VIP discount"},
})

// Generate with metadata
md, err := rules.GenerateMarkdown(rules.DocumentOptions{
    IncludeMetadata: true,
})
```

#### Requirement Traceability

The `RequirementID` and `BusinessDescription` fields enable direct traceability to external requirements:

```go
minimumAmount := rules.NewWithDomain("minimum amount", OrderDomain, ...)

rules.UpdateMetadata(minimumAmount, rules.RuleMetadata{
    RequirementID: "JIRA-1234",
    BusinessDescription: "Standard customers must have a minimum order " +
                        "amount of $100 to qualify for free shipping",
    Owner:     "Order Team",
    Version:   "1.0.0",
    CreatedAt: time.Now(),
})
```

This creates a direct link between your Jira tickets (or other requirement management tools) and the implementing code, making it easy to:
- Verify which requirements have been implemented
- Generate compliance reports
- Track requirement coverage
- Keep documentation synchronized with business requirements

### Domain Inheritance

Hierarchical rules automatically inherit domains from their children:

```go
// Child rules with domains
rule1 := rules.NewWithDomain("check A", OrderDomain, ...)
rule2 := rules.NewWithDomain("check B", UserDomain, ...)

// Parent automatically gets both domains
combined := rules.And("combined check", rule1, rule2)
// combined now has both OrderDomain and UserDomain
```

### Output Formats

#### Markdown
Perfect for documentation sites, READMEs, and wikis:
- Hierarchical structure with headers
- Domain and group sections
- Table of contents
- Metadata display

#### JSON
Machine-readable format for tools and APIs:
- Complete rule hierarchy
- Full metadata serialization
- Domain and group filtering
- Easy integration with other systems

#### HTML
Interactive documentation with:
- Collapsible rule sections
- Real-time search
- Navigation sidebar
- Responsive design
- Type-specific badges

#### Mermaid
Visual diagrams showing:
- Rule hierarchies
- Domain-based subgraphs
- Different shapes per rule type
- Connection arrows
- Color coding

### Filtering Documentation

Control which rules to document:

```go
// Include only specific domains
opts := rules.DocumentOptions{
    IncludeDomains: []rules.Domain{OrderDomain, UserDomain},
}

// Exclude specific domains
opts := rules.DocumentOptions{
    ExcludeDomains: []rules.Domain{InternalDomain},
}

// Limit hierarchy depth
opts := rules.DocumentOptions{
    MaxDepth: 3,  // Only show 3 levels deep
}
```

### Command-Line Tool

Use the included CLI tool to generate documentation:

```bash
# Generate all formats (Markdown, HTML, JSON, Mermaid)
go run ./cmd/gendocs/main.go

# Custom output directory
go run ./cmd/gendocs/main.go -output ./documentation

# Specific formats only
go run ./cmd/gendocs/main.go -formats markdown,html

# Custom title and description
go run ./cmd/gendocs/main.go \
  -title "Order Processing Rules" \
  -description "Business rules for our e-commerce platform"

# See all options
go run ./cmd/gendocs/main.go -help
```

**Available Flags:**
- `-output` - Output directory (default: `docs`)
- `-title` - Documentation title
- `-description` - Documentation description
- `-formats` - Comma-separated formats: `markdown,html,json,mermaid`
- `-group-by-domain` - Group rules by domain (default: `true`)
- `-include-metadata` - Include metadata in docs (default: `true`)

### Keeping Documentation in Sync

To ensure documentation stays synchronized with code, use one of these approaches:

#### Option 1: Pre-commit Hook (Recommended)

Automatically generate documentation before each commit:

```bash
# Create .git/hooks/pre-commit
#!/bin/bash
echo "Generating documentation..."
go run ./cmd/gendocs/main.go || exit 1
git add docs/
echo "✓ Documentation updated"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

#### Option 2: Makefile Integration

```makefile
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@go run ./cmd/gendocs/main.go
	@echo "✓ Documentation generated"

.PHONY: pre-commit
pre-commit: test docs
	@echo "✓ Ready to commit"
```

Developer workflow:
```bash
make pre-commit  # Run tests and generate docs
git add .
git commit -m "Add new rule"
```

#### Option 3: CI Validation

Validate that documentation is up-to-date in CI (doesn't auto-commit):

```yaml
# .github/workflows/ci.yml
- name: Check documentation is current
  run: |
    go run ./cmd/gendocs/main.go
    if ! git diff --exit-code docs/; then
      echo "❌ Documentation is out of date"
      echo "Run 'go run ./cmd/gendocs/main.go' and commit the changes"
      exit 1
    fi
    echo "✓ Documentation is up-to-date"
```

This approach ensures documentation is always committed with code changes, not added by CI.

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/tobbstr/rules"
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
- Consider caching expensive rule evaluations

## License

Same as the parent project.

