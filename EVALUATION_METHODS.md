# Rule Evaluation Methods - Quick Reference

This guide helps you choose the right evaluation method for your use case.

## Decision Matrix

| Use Case | Method | Performance | What You Get |
|----------|--------|-------------|--------------|
| Production validation | `EvaluateFast()` | ⚡⚡⚡⚡⚡ 22 ns | bool + error |
| Health checks / APIs | `rule.Evaluate()` | ⚡⚡⚡⚡ 25 ns | bool + error |
| With metrics/timing | `Evaluate()` | ⚡⚡⚡ 140 ns | + duration |
| Fast debugging | `EvaluateDetailedShortCircuit()` | ⚡⚡ 432 ns | + partial tree |
| Complete debugging | `EvaluateDetailed()` | ⚡ 858 ns | + full tree |

## Method Details

### 1. Direct Rule Evaluation (Fastest)
```go
satisfied, err := rule.Evaluate(input)
```
- **Speed:** ~25 ns/op
- **Memory:** 0 allocations
- **Returns:** `bool, error`
- **Use when:** You just need a yes/no answer

### 2. EvaluateFast() (New! Almost as fast)
```go
evaluator := rules.NewEvaluator(rule)
satisfied, err := evaluator.EvaluateFast(input)
```
- **Speed:** ~22 ns/op
- **Memory:** 0 allocations  
- **Returns:** `bool, error`
- **Use when:** Using Evaluator but don't need timing

### 3. Evaluate() (With timing)
```go
evaluator := rules.NewEvaluator(rule)
result := evaluator.Evaluate(input)
fmt.Printf("Satisfied: %v, Took: %v\n", result.Satisfied, result.Duration)
```
- **Speed:** ~140 ns/op
- **Memory:** 0 allocations
- **Returns:** `Result` with timing info
- **Use when:** You need to measure/log evaluation time

### 4. EvaluateDetailedShortCircuit() (New! Fast + some details)
```go
result := evaluator.EvaluateDetailedShortCircuit(input)
// result.Children contains evaluated children (not all)
fmt.Println(result.String())
```
- **Speed:** ~432 ns/op
- **Memory:** 1 allocation
- **Returns:** `Result` with partial child tree
- **Behavior:**
  - AND: Stops at first failing child
  - OR: Stops at first succeeding child
- **Use when:** Debugging but need better performance

### 5. EvaluateDetailed() (Complete tree)
```go
result := evaluator.EvaluateDetailed(input)
// result.Children contains ALL child results
for _, child := range result.Children {
    fmt.Printf("%s: %v\n", child.RuleName, child.Satisfied)
}
```
- **Speed:** ~858 ns/op
- **Memory:** 1 allocation (was 3 before optimization!)
- **Returns:** `Result` with complete child tree
- **Use when:** Need complete evaluation details for debugging/auditing

## Performance Comparison

### Throughput (evaluations/second/core)

| Method | Ops/Sec | Best For |
|--------|---------|----------|
| `rule.Evaluate()` | 45M | Direct validation |
| `EvaluateFast()` | 45M | High-throughput APIs |
| `Evaluate()` | 7M | Monitored validation |
| `EvaluateDetailedShortCircuit()` | 2.3M | Quick debugging |
| `EvaluateDetailed()` | 1.2M | Full debugging |

## Example Usage

### Production API Endpoint
```go
func validateOrder(order Order) error {
    evaluator := rules.NewEvaluator(orderValidationRule)
    
    // Use fastest method
    satisfied, err := evaluator.EvaluateFast(order)
    if err != nil {
        return fmt.Errorf("validation error: %w", err)
    }
    if !satisfied {
        return errors.New("order validation failed")
    }
    return nil
}
```

### With Monitoring
```go
func validateOrderWithMetrics(order Order) error {
    evaluator := rules.NewEvaluator(orderValidationRule)
    
    result := evaluator.Evaluate(order)
    
    // Log timing
    metrics.RecordDuration("order_validation", result.Duration)
    
    if result.Error != nil {
        return result.Error
    }
    if !result.Satisfied {
        return errors.New("validation failed")
    }
    return nil
}
```

### Debugging / Development
```go
func debugValidation(order Order) {
    evaluator := rules.NewEvaluator(orderValidationRule)
    
    // Get complete details
    result := evaluator.EvaluateDetailed(order)
    
    // Pretty print the tree
    fmt.Println(result.String())
    // Output:
    // ✓ order validation (took 1.5ms)
    //   ✓ has items (took 100ns)
    //   ✗ minimum amount (took 200ns)
    //   ✓ valid country (took 150ns)
}
```

### Quick Failure Analysis
```go
func quickDebug(order Order) {
    evaluator := rules.NewEvaluator(orderValidationRule)
    
    // Get partial tree (faster than EvaluateDetailed)
    result := evaluator.EvaluateDetailedShortCircuit(order)
    
    if !result.Satisfied {
        // See which rule failed without evaluating all children
        log.Debugf("Validation failed at: %s", result.Children[len(result.Children)-1].RuleName)
    }
}
```

## Migration Guide

If you're currently using:

### `rule.Evaluate(input)` → No change needed!
Already optimal for most cases.

### `evaluator.Evaluate(input)` → Consider switching
```go
// If you don't use result.Duration:
- result := evaluator.Evaluate(input)
+ satisfied, err := evaluator.EvaluateFast(input)
```

### `evaluator.EvaluateDetailed(input)` → Consider switching
```go
// If you only need to see failing rules:
- result := evaluator.EvaluateDetailed(input)
+ result := evaluator.EvaluateDetailedShortCircuit(input)
// 2x faster!
```

## Performance Tips

1. **Reuse Evaluators**: Create once, use many times
   ```go
   var orderEvaluator = rules.NewEvaluator(orderRule)
   
   func validate(order Order) error {
       satisfied, err := orderEvaluator.EvaluateFast(order)
       // ...
   }
   ```

2. **Use appropriate method**: Don't use detailed evaluation in production
   ```go
   // ❌ BAD: Unnecessary overhead in production
   result := evaluator.EvaluateDetailed(order)
   
   // ✅ GOOD: Fast path for production
   satisfied, err := evaluator.EvaluateFast(order)
   ```

3. **Short-circuit for debugging**: Get 50% speedup over full detailed
   ```go
   // ✅ GOOD: Faster debugging
   result := evaluator.EvaluateDetailedShortCircuit(order)
   ```

4. **Direct evaluation when possible**: Skip Evaluator wrapper
   ```go
   // Fastest - no wrapper overhead
   satisfied, err := rule.Evaluate(input)
   ```

## Memory Profile

| Method | Allocations | Bytes Allocated |
|--------|-------------|-----------------|
| `rule.Evaluate()` | 0 | 0 |
| `EvaluateFast()` | 0 | 0 |
| `Evaluate()` | 0 | 0 |
| `EvaluateDetailedShortCircuit()` | 1 | 224 |
| `EvaluateDetailed()` | 1 | 224 |

(For AND rule with 3 children. Scales with rule complexity.)

## See Also

- [PERFORMANCE_IMPROVEMENTS.md](PERFORMANCE_IMPROVEMENTS.md) - Detailed optimization report
- [OPTIMIZATION_SUMMARY.md](OPTIMIZATION_SUMMARY.md) - Complete summary with benchmarks
- [README.md](README.md) - Full library documentation

