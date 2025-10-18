# Issues Found in Documentation Generation Plan

## Critical Issues

### 1. Type Inconsistency: Domain vs string ✅ **RESOLVED**

**Problem**: The plan defined `Domain` as a custom type for type safety, but then used `string` inconsistently throughout the API.

**Solution**: Changed ALL domain parameters and fields to use `rules.Domain` type consistently.

**Fixed APIs**:
```go
// ✅ FIXED - now uses Domain type
type RegisteredRule struct {
    Domains []Domain  // ✅ Fixed
}

type RuleMetadata struct {
    Domains      []Domain  // ✅ Fixed
    Dependencies []Domain  // ✅ Fixed
}

type DocumentOptions struct {
    IncludeDomains []Domain  // ✅ Fixed
    ExcludeDomains []Domain  // ✅ Fixed
}

// Registry interface
func RulesByDomain(domain Domain) []RegisteredRule          // ✅ Fixed
func RulesByDomains(domains ...Domain) []RegisteredRule     // ✅ Fixed
func Domains() []Domain                                      // ✅ Fixed

// RegistrationOption functions
func Domain(d Domain) RegistrationOption                     // ✅ Fixed
func Domains(domains ...Domain) RegistrationOption          // ✅ Fixed
func Group(name string, domains ...Domain) RegistrationOption // ✅ Fixed

// Convenience functions
func GenerateDomainMarkdown(domain Domain, ...) (string, error)      // ✅ Fixed
func GenerateDomainsMarkdown(domains []Domain, ...) (string, error)  // ✅ Fixed
```

**Updated Sections**:
- Section 4.3: Registry API
- Section 4.4: DocumentOptions
- Section 4.5: RuleMetadata
- Section 4.6: Convenience functions
- Section 2.6: Examples updated to use typed domains
- Section 11.2, 11.3, 11.4: All examples updated

**Status**: All type inconsistencies fixed. Full type safety achieved.

---

### 2. Hierarchical Rules Not Addressed ✅ **RESOLVED**

**Problem**: Auto-registration only works for simple rules created with `NewWithDomain()` and `NewWithGroup()`. 
Existing combinators (`And()`, `Or()`, `Not()`, `AllOf()`, `AnyOf()`, etc.) have no domain support.

**Solution Chosen**: **Option B: Inherit Domains from Children**

Hierarchical rules automatically:
1. Collect domains from all child rules
2. Deduplicate domains
3. Auto-register if result has at least one domain
4. No domains = no registration (utility rules)

**Example**:
```go
var rule1 = rules.NewWithDomain("check1", OrderDomain, pred1)    // [order]
var rule2 = rules.NewWithDomain("check2", UserDomain, pred2)     // [user]
var combined = rules.And("combined", rule1, rule2)               // [order, user] (auto-inherited, auto-registered)
```

**Updated in Plan**:
- Section 2.3: Added example of domain inheritance
- Section 2.5: New section dedicated to hierarchical rules and domain inheritance
- Section 5.1: Added implementation steps for domain inheritance
- Section 13: Removed from open questions

**Status**: Documentation updated, ready for implementation.

---

### 3. WithDescription() Integration ✅ **RESOLVED**

**Problem**: If a rule is auto-registered, then wrapped with `WithDescription()`, what happens to the registry?

**Solution Chosen**: **Descriptions stored in registry only, not on rules.**

**Key Design Decisions**:
1. **Remove `Description()` from `Rule[T]` interface** - Breaking change, but acceptable
   (no consumers yet)
2. **Single source of truth** - Registry stores descriptions, not rules
3. **No wrapper needed** - `WithDescription()` updates registry entry and returns same rule
4. **Pointer equality lookup** - Registry finds rules by pointer equality

**Implementation**:
```go
// Rule interface - NO Description() method
type Rule[T any] interface {
    Evaluate(input T) (bool, error)
    Name() string
    // Description() removed - stored in registry instead
}

// RegisteredRule in registry has Description field
type RegisteredRule struct {
    Rule        any
    Description string  // Single source of truth
    Domains     []Domain
    Group       string
    // ...
}

// WithDescription updates registry entry, returns same rule
func WithDescription[T any](rule Rule[T], description string) Rule[T] {
    DefaultRegistry.UpdateDescription(rule, description)
    return rule  // Same rule, no wrapper
}

// GetDescription retrieves from registry
func GetDescription(rule any) string {
    return DefaultRegistry.GetDescription(rule)
}
```

**Benefits**:
- No wrapper type needed
- Rules remain truly immutable
- Clean separation: rules = logic, registry = metadata
- Single source of truth for descriptions

**Updated in Plan**:
- Section 3.1: Clarified Rule[T] only has Evaluate() and Name()
- Section 4.2.1: Documented registry-based description storage
- Section 4.3: Added Description field to RegisteredRule
- Section 9.2: Documented breaking change to remove Description() from Rule[T]

**Status**: Fully resolved and documented.

---

### 4. UpdateMetadata() Function Missing Implementation Details ✅ **RESOLVED**

**Problem**: Function signature is:
```go
func UpdateMetadata(rule any, metadata RuleMetadata) error
```

How does it find the right registry entry?

**Solution Chosen**: **Pointer equality lookup**

**Implementation**:
- Registry stores rules by pointer in internal map
- `UpdateMetadata()`, `WithDescription()`, `GetDescription()` all use pointer equality
- Returns error if rule not found (not registered)
- Works correctly because:
  - Rules are typically package-level variables (stable pointers)
  - No wrapper types (WithDescription doesn't wrap, just updates registry)
  - Simple and performant

**Benefits**:
- Simple implementation (pointer map lookup)
- Fast O(1) lookups
- No need for unique IDs
- No complex unwrapping logic

**Updated in Plan**:
- Section 4.2.1: Documented "Lookup is by pointer equality" for all functions
- Added "Registry Lookup Mechanism" explanation

**Status**: Fully resolved and documented.

---

### 5. Thread-Safety Not Explicit for Auto-Registration ⚠️ **NEEDS DOCUMENTATION**

**Problem**: Multiple packages initializing concurrently during program startup. Is auto-registration thread-safe?

```go
// package order - init runs
var MinimumAmountRule = rules.NewWithDomain(...)  // Registers

// package user - init runs concurrently
var ActiveUserRule = rules.NewWithDomain(...)  // Registers

// Are these concurrent registrations safe?
```

**Recommendation**: Explicitly document thread-safety guarantees for auto-registration.

---

## Medium Priority Issues

### 6. Domain.String() Method Unnecessary

**Issue**: Section 4.1 defines:
```go
type Domain string
func (d Domain) String() string
```

Since `Domain` is a string type alias, it already implements `Stringer` interface implicitly. 
The explicit `String()` method is redundant.

**Recommendation**: Remove from plan or clarify if there's a specific reason.

---

### 7. Missing: Query Domains from Rules

**Issue**: Once registered, how do you query what domains a rule belongs to without registry lookup?

**Options**:
1. Add `Domains()` method to `Rule[T]` interface (breaking change)
2. Add optional `DomainProvider` interface
3. Only query via registry

**Recommendation**: Keep it registry-only to avoid changing `Rule[T]` interface.

---

### 8. Example Has Wrong Function Name

**Issue**: Section 11.2, line 1331:
```go
jsonDocs, err := rules.GenerateAllRulesMarkdown(rules.DocumentOptions{
    Format: rules.FormatJSON,  // ❌ Format is JSON but function says Markdown
})
```

**Fix**: Should be `GenerateAllRulesJSON()` or a generic function that handles all formats.

---

## Minor Issues

### 9. TODO Organization Could Be Clearer

**Issue**: The TODO list mixes different concerns in Phase 1.

**Recommendation**: Reorganize Phase 1 as:
- Phase 1A: Core Types (Domain, RegisteredRule, RuleMetadata)
- Phase 1B: Registry Implementation (with tests)
- Phase 1C: Auto-Registration Functions (with tests)
- Phase 1D: Rule Introspection

This makes dependencies clearer and allows parallel work.

---

### 10. RulesByGroup vs GenerateGroupMarkdown

**Issue**: Are both needed? They serve similar but distinct purposes.

**Answer**: Yes, both are needed:
- `RulesByGroup()` - Query function returning rules
- `GenerateGroupMarkdown()` - Convenience function for documentation

This is fine, just confirming intent.

---

## Questions to Answer Before Implementation

1. **Hierarchical Rules**: Which solution (A/B/C/D) for supporting And/Or/Not with domains?

2. **WithDescription Behavior**: Transparent wrapper or registry update?

3. **UpdateMetadata Lookup**: By pointer, ID, or name?

4. **Domain Type Consistency**: Confirm ALL uses of domain should be `Domain` type, not `string`?

5. **Hierarchical Domain Inheritance**: Should `And(rule1, rule2)` auto-inherit domains from children?

6. **Builder Pattern Integration**: How does `rules.Builder[T]` integrate with auto-registration?

7. **Map Function Integration**: How does `Map()` work with domains (it transforms types)?

---

## Actionable Next Steps

### ✅ Completed Before Implementation:

1. ~~**Fix ALL type inconsistencies**~~ ✅ - Changed `string` to `Domain` throughout
2. ~~**Decide on hierarchical rule strategy**~~ ✅ - Using domain inheritance from children
3. ~~**Clarify WithDescription behavior**~~ ✅ - Registry-based, no wrapper, pointer lookup
4. ~~**Specify UpdateMetadata lookup**~~ ✅ - Pointer equality for O(1) lookups
5. ~~**Reorganize TODO Phase 1**~~ ✅ - Added domain inheritance tasks
6. ~~**Update Plan Sections**~~ ✅ - All critical sections updated

### Remaining Before Implementation (Non-Blocking):

7. **Document thread-safety** - Explicit guarantees for auto-registration
8. **Fix example typo** - GenerateAllRulesMarkdown -> proper function name (section 11.2)
9. **Remove Domain.String()** - Redundant method from plan (section 4.1)

### Can Add During Implementation:

10. **Builder pattern integration** - How Builder interacts with domains
11. **Map function domain handling** - How cross-type rules work with domains

### Plan Sections Updated:

- ~~Section 2.5: New section on hierarchical rule domain inheritance~~ ✅
- ~~Section 3.1: Clarified Rule[T] only has Evaluate() and Name()~~ ✅
- ~~Section 4.2.1: Documented registry-based description storage~~ ✅
- ~~Section 4.3: Added Description field to RegisteredRule~~ ✅
- ~~Section 4.3-4.6: Fix all `string` to `Domain` in Registry API~~ ✅
- ~~Section 9.2: Documented breaking change (remove Description())~~ ✅
- ~~Section 13: Removed hierarchical rules from open questions~~ ✅
- Section 4.1: Remove redundant `Domain.String()` method
- Section 11.2: Fix function name typo

---

## Priority Ranking

**✅ Completed (All Blocking Issues Resolved!):**
1. ~~Domain type consistency~~ - All APIs now use `Domain` type
2. ~~Hierarchical rules strategy~~ - Using domain inheritance from children
3. ~~TODO reorganization~~ - Phase 1 split with domain inheritance tasks
4. ~~WithDescription behavior~~ - Registry-based storage, no wrapper, pointer lookup
5. ~~UpdateMetadata lookup~~ - Pointer equality for fast O(1) lookups

**Should Fix Before Implementation:**
6. Thread-safety documentation (explicit guarantees for concurrent init)
7. Example typo (section 11.2 - GenerateAllRulesMarkdown with JSON format)
8. Remove Domain.String() method from plan (redundant)

**Can Document During Implementation:**
9. Builder/Map integration details (how they interact with domains)

**Breaking Changes Accepted:**
- Remove `Description()` method from `Rule[T]` interface (no consumers yet)


