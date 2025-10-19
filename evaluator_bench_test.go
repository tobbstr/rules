package rules

import (
	"errors"
	"testing"
)

// Benchmark inputs
type benchInput struct {
	value  int
	status string
	active bool
}

// Benchmark: Simple rule evaluation
func BenchmarkSimpleRuleEvaluation(b *testing.B) {
	rule := New("value check", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})

	input := benchInput{value: 150}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate(input)
	}
}

// Benchmark: AND rule with 3 children
func BenchmarkAndRule3Children(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	andRule := And("all checks", rule1, rule2, rule3)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = andRule.Evaluate(input)
	}
}

// Benchmark: AND rule with 10 children
func BenchmarkAndRule10Children(b *testing.B) {
	rules := make([]Rule[benchInput], 10)
	for i := 0; i < 10; i++ {
		threshold := i * 10
		rules[i] = New("check", func(input benchInput) (bool, error) {
			return input.value > threshold, nil
		})
	}

	andRule := And("all checks", rules...)
	input := benchInput{value: 150}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = andRule.Evaluate(input)
	}
}

// Benchmark: OR rule with 3 children (first matches)
func BenchmarkOrRule3ChildrenFirstMatch(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	orRule := Or("any check", rule1, rule2, rule3)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = orRule.Evaluate(input)
	}
}

// Benchmark: OR rule with 3 children (last matches)
func BenchmarkOrRule3ChildrenLastMatch(b *testing.B) {
	rule1 := New("value > 1000", func(input benchInput) (bool, error) {
		return input.value > 1000, nil
	})
	rule2 := New("status == invalid", func(input benchInput) (bool, error) {
		return input.status == "invalid", nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	orRule := Or("any check", rule1, rule2, rule3)
	input := benchInput{value: 150, status: "valid", active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = orRule.Evaluate(input)
	}
}

// Benchmark: NOT rule
func BenchmarkNotRule(b *testing.B) {
	rule := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	notRule := Not("not check", rule)

	input := benchInput{value: 50}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = notRule.Evaluate(input)
	}
}

// Benchmark: Deeply nested rules (5 levels)
func BenchmarkDeeplyNestedRules(b *testing.B) {
	rule1 := New("value > 0", func(input benchInput) (bool, error) {
		return input.value > 0, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	// Level 1
	level1 := And("level1", rule1, rule2)
	// Level 2
	level2 := Or("level2", level1, rule3)
	// Level 3
	level3 := And("level3", level2, rule1)
	// Level 4
	level4 := Or("level4", level3, rule2)
	// Level 5
	level5 := And("level5", level4, rule3)

	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = level5.Evaluate(input)
	}
}

// Benchmark: Wide rule tree (1 AND with 100 children)
func BenchmarkWideRuleTree(b *testing.B) {
	rules := make([]Rule[benchInput], 100)
	for i := 0; i < 100; i++ {
		threshold := i
		rules[i] = New("check", func(input benchInput) (bool, error) {
			return input.value > threshold, nil
		})
	}

	wideRule := And("wide check", rules...)
	input := benchInput{value: 150}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = wideRule.Evaluate(input)
	}
}

// Benchmark: Rule with error
func BenchmarkRuleWithError(b *testing.B) {
	rule := New("error check", func(input benchInput) (bool, error) {
		if input.value < 0 {
			return false, errors.New("negative value")
		}
		return input.value > 100, nil
	})

	input := benchInput{value: -1}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = rule.Evaluate(input)
	}
}

// Benchmark: Evaluator.Evaluate (with timing overhead)
func BenchmarkEvaluatorEvaluate(b *testing.B) {
	rule := New("value check", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})

	evaluator := NewEvaluator(rule)
	input := benchInput{value: 150}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.Evaluate(input)
	}
}

// Benchmark: Evaluator.EvaluateDetailed with simple rule
func BenchmarkEvaluatorDetailedSimple(b *testing.B) {
	rule := New("value check", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})

	evaluator := NewEvaluator(rule)
	input := benchInput{value: 150}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailed(input)
	}
}

// Benchmark: Evaluator.EvaluateDetailed with AND rule (3 children)
func BenchmarkEvaluatorDetailedAnd3(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	andRule := And("all checks", rule1, rule2, rule3)
	evaluator := NewEvaluator(andRule)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailed(input)
	}
}

// Benchmark: Evaluator.EvaluateDetailed with deeply nested rules
func BenchmarkEvaluatorDetailedNested(b *testing.B) {
	rule1 := New("value > 0", func(input benchInput) (bool, error) {
		return input.value > 0, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	level1 := And("level1", rule1, rule2)
	level2 := Or("level2", level1, rule3)
	level3 := And("level3", level2, rule1)

	evaluator := NewEvaluator(level3)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailed(input)
	}
}

// Benchmark: Parallel evaluation (simulating concurrent requests)
func BenchmarkParallelEvaluation(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	complexRule := And("complex", rule1, rule2, rule3)
	input := benchInput{value: 150, active: true}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = complexRule.Evaluate(input)
		}
	})
}

// Benchmark: Real-world scenario - order validation
func BenchmarkRealWorldOrderValidation(b *testing.B) {
	// Simulate realistic business rules for order validation
	hasItems := New("has items", func(input benchInput) (bool, error) {
		return input.value > 0, nil
	})

	validQuantity := New("valid quantity", func(input benchInput) (bool, error) {
		return input.value > 0 && input.value <= 1000, nil
	})

	isActive := New("is active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	validStatus := New("valid status", func(input benchInput) (bool, error) {
		return input.status == "pending" || input.status == "confirmed", nil
	})

	// Basic validation
	basicChecks := And("basic validation", hasItems, validQuantity)

	// Status checks
	statusChecks := And("status validation", isActive, validStatus)

	// Combined validation
	orderValidation := And("order validation", basicChecks, statusChecks)

	input := benchInput{value: 5, status: "pending", active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = orderValidation.Evaluate(input)
	}
}

// Benchmark: EvaluateDetailedShortCircuit with AND rule (first fails)
func BenchmarkEvaluatorDetailedShortCircuitAndFirstFails(b *testing.B) {
	rule1 := New("value > 1000", func(input benchInput) (bool, error) {
		return input.value > 1000, nil
	})
	rule2 := New("value < 2000", func(input benchInput) (bool, error) {
		return input.value < 2000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	andRule := And("all checks", rule1, rule2, rule3)
	evaluator := NewEvaluator(andRule)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailedShortCircuit(input)
	}
}

// Benchmark: EvaluatorDetailedShortCircuit with OR rule (first succeeds)
func BenchmarkEvaluatorDetailedShortCircuitOrFirstSucceeds(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	orRule := Or("any check", rule1, rule2, rule3)
	evaluator := NewEvaluator(orRule)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateDetailedShortCircuit(input)
	}
}

// Benchmark: EvaluateFast method
func BenchmarkEvaluateFast(b *testing.B) {
	rule1 := New("value > 100", func(input benchInput) (bool, error) {
		return input.value > 100, nil
	})
	rule2 := New("value < 1000", func(input benchInput) (bool, error) {
		return input.value < 1000, nil
	})
	rule3 := New("active", func(input benchInput) (bool, error) {
		return input.active, nil
	})

	andRule := And("all checks", rule1, rule2, rule3)
	evaluator := NewEvaluator(andRule)
	input := benchInput{value: 150, active: true}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateFast(input)
	}
}
