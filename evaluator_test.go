package rules

import (
	"testing"
)

func TestEvaluator(t *testing.T) {
	t.Parallel()

	rule := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)

	evaluator := NewEvaluator(rule)

	t.Run("simple evaluation", func(t *testing.T) {
		result := evaluator.Evaluate(testInput{value: 15})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}

		if result.RuleName != "value > 10" {
			t.Errorf(
				"Expected rule name %q, got %q",
				"value > 10",
				result.RuleName,
			)
		}

		if result.Duration <= 0 {
			t.Error("Expected positive duration")
		}
	})

	t.Run("not satisfied", func(t *testing.T) {
		result := evaluator.Evaluate(testInput{value: 5})

		if result.Satisfied {
			t.Error("Expected rule to not be satisfied")
		}

		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}
	})
}

func TestEvaluatorDetailed(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)
	rule2 := New(
		"value < 100",
		func(input testInput) (bool, error) {
			return input.value < 100, nil
		},
	)
	rule3 := New(
		"valid",
		func(input testInput) (bool, error) {
			return input.valid, nil
		},
	)

	complexRule := And("complex rule", rule1, rule2, rule3)
	evaluator := NewEvaluator(complexRule)

	t.Run("all satisfied", func(t *testing.T) {
		result := evaluator.EvaluateDetailed(testInput{value: 50, valid: true})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		if len(result.Children) != 3 {
			t.Errorf("Expected 3 children, got %d", len(result.Children))
		}

		for _, child := range result.Children {
			if !child.Satisfied {
				t.Errorf(
					"Expected child %q to be satisfied",
					child.RuleName,
				)
			}
		}
	})

	t.Run("one not satisfied", func(t *testing.T) {
		result := evaluator.EvaluateDetailed(testInput{value: 50, valid: false})

		if result.Satisfied {
			t.Error("Expected rule to not be satisfied")
		}

		if len(result.Children) != 3 {
			t.Errorf("Expected 3 children, got %d", len(result.Children))
		}

		satisfiedCount := 0
		for _, child := range result.Children {
			if child.Satisfied {
				satisfiedCount++
			}
		}

		if satisfiedCount != 2 {
			t.Errorf(
				"Expected 2 satisfied children, got %d",
				satisfiedCount,
			)
		}
	})

	t.Run("nested rules", func(t *testing.T) {
		rangeRule := And("range", rule1, rule2)
		nestedRule := Or("nested", rangeRule, rule3)

		evaluator := NewEvaluator(nestedRule)
		result := evaluator.EvaluateDetailed(testInput{value: 50, valid: false})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		if len(result.Children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(result.Children))
		}

		// Check that the first child (rangeRule) has 2 children
		if len(result.Children[0].Children) != 2 {
			t.Errorf(
				"Expected first child to have 2 children, got %d",
				len(result.Children[0].Children),
			)
		}
	})
}

func TestResultString(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"rule1",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)
	rule2 := New(
		"rule2",
		func(input testInput) (bool, error) {
			return input.value < 100, nil
		},
	)

	complexRule := And("complex", rule1, rule2)
	evaluator := NewEvaluator(complexRule)

	result := evaluator.EvaluateDetailed(testInput{value: 50})

	str := result.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Check that the string contains rule names
	if !contains(str, "complex") {
		t.Error("Expected string to contain 'complex'")
	}
	if !contains(str, "rule1") {
		t.Error("Expected string to contain 'rule1'")
	}
	if !contains(str, "rule2") {
		t.Error("Expected string to contain 'rule2'")
	}
}

func TestResultIsSuccessful(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name: "satisfied without error",
			result: Result{
				Satisfied: true,
				Error:     nil,
			},
			want: true,
		},
		{
			name: "satisfied with error",
			result: Result{
				Satisfied: true,
				Error:     ErrEvaluationFailed,
			},
			want: false,
		},
		{
			name: "not satisfied without error",
			result: Result{
				Satisfied: false,
				Error:     nil,
			},
			want: false,
		},
		{
			name: "not satisfied with error",
			result: Result{
				Satisfied: false,
				Error:     ErrEvaluationFailed,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.result.IsSuccessful()
			if got != tt.want {
				t.Errorf("IsSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResultHasError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name: "with error",
			result: Result{
				Error: ErrEvaluationFailed,
			},
			want: true,
		},
		{
			name: "without error",
			result: Result{
				Error: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.result.HasError()
			if got != tt.want {
				t.Errorf("HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluatorFast(t *testing.T) {
	t.Parallel()

	rule := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)

	evaluator := NewEvaluator(rule)

	t.Run("satisfied", func(t *testing.T) {
		satisfied, err := evaluator.EvaluateFast(testInput{value: 15})

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("not satisfied", func(t *testing.T) {
		satisfied, err := evaluator.EvaluateFast(testInput{value: 5})

		if satisfied {
			t.Error("Expected rule to not be satisfied")
		}

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestEvaluatorDetailedShortCircuit(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)
	rule2 := New(
		"value < 100",
		func(input testInput) (bool, error) {
			return input.value < 100, nil
		},
	)
	rule3 := New(
		"valid",
		func(input testInput) (bool, error) {
			return input.valid, nil
		},
	)

	t.Run("AND stops on first failure", func(t *testing.T) {
		andRule := And("all checks", rule1, rule2, rule3)
		evaluator := NewEvaluator(andRule)

		// First rule fails, should short-circuit
		result := evaluator.EvaluateDetailedShortCircuit(testInput{value: 5, valid: true})

		if result.Satisfied {
			t.Error("Expected rule to not be satisfied")
		}

		// Should only have 1 child result (the failing one)
		if len(result.Children) != 1 {
			t.Errorf("Expected 1 child result, got %d", len(result.Children))
		}

		if result.Children[0].RuleName != "value > 10" {
			t.Errorf("Expected first child, got %q", result.Children[0].RuleName)
		}
	})

	t.Run("OR stops on first success", func(t *testing.T) {
		orRule := Or("any check", rule1, rule2, rule3)
		evaluator := NewEvaluator(orRule)

		// First rule succeeds, should short-circuit
		result := evaluator.EvaluateDetailedShortCircuit(testInput{value: 15, valid: false})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		// Should only have 1 child result (the succeeding one)
		if len(result.Children) != 1 {
			t.Errorf("Expected 1 child result, got %d", len(result.Children))
		}

		if result.Children[0].RuleName != "value > 10" {
			t.Errorf("Expected first child, got %q", result.Children[0].RuleName)
		}
	})

	t.Run("AND evaluates all when all pass", func(t *testing.T) {
		andRule := And("all checks", rule1, rule2)
		evaluator := NewEvaluator(andRule)

		result := evaluator.EvaluateDetailedShortCircuit(testInput{value: 50})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		// Should have all child results
		if len(result.Children) != 2 {
			t.Errorf("Expected 2 child results, got %d", len(result.Children))
		}
	})

	t.Run("nested rules with short-circuit", func(t *testing.T) {
		innerAnd := And("inner", rule1, rule2)
		outerOr := Or("outer", innerAnd, rule3)

		evaluator := NewEvaluator(outerOr)

		// Inner AND succeeds, outer OR should short-circuit
		result := evaluator.EvaluateDetailedShortCircuit(testInput{value: 50, valid: false})

		if !result.Satisfied {
			t.Error("Expected rule to be satisfied")
		}

		// Should only have 1 child (the succeeding innerAnd)
		if len(result.Children) != 1 {
			t.Errorf("Expected 1 child result, got %d", len(result.Children))
		}

		// The inner AND should have evaluated all its children
		if len(result.Children[0].Children) != 2 {
			t.Errorf(
				"Expected inner AND to have 2 children, got %d",
				len(result.Children[0].Children),
			)
		}
	})
}

func TestResultUnsatisfiedRules(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)
	rule2 := New(
		"value < 100",
		func(input testInput) (bool, error) {
			return input.value < 100, nil
		},
	)
	rule3 := New(
		"valid",
		func(input testInput) (bool, error) {
			return input.valid, nil
		},
	)

	tests := []struct {
		name     string
		rule     Rule[testInput]
		input    testInput
		expected []string
	}{
		{
			name:     "simple rule satisfied",
			rule:     rule1,
			input:    testInput{value: 15},
			expected: []string{},
		},
		{
			name:     "simple rule not satisfied",
			rule:     rule1,
			input:    testInput{value: 5},
			expected: []string{"value > 10"},
		},
		{
			name:     "AND all satisfied",
			rule:     And("all checks", rule1, rule2, rule3),
			input:    testInput{value: 50, valid: true},
			expected: []string{},
		},
		{
			name:     "AND one not satisfied",
			rule:     And("all checks", rule1, rule2, rule3),
			input:    testInput{value: 50, valid: false},
			expected: []string{"all checks", "valid"},
		},
		{
			name:     "AND multiple not satisfied",
			rule:     And("all checks", rule1, rule2, rule3),
			input:    testInput{value: 5, valid: false},
			expected: []string{"all checks", "value > 10", "valid"},
		},
		{
			name:     "OR some children unsatisfied but OR satisfied",
			rule:     Or("any check", rule1, rule2, rule3),
			input:    testInput{value: 5, valid: false},
			expected: []string{"value > 10", "valid"},
		},
		{
			name:     "OR one child satisfied",
			rule:     Or("any check", rule1, rule2, rule3),
			input:    testInput{value: 50, valid: false},
			expected: []string{"valid"},
		},
		{
			name:     "OR all children unsatisfied",
			rule:     Or("any check", rule1, rule3),
			input:    testInput{value: 5, valid: false},
			expected: []string{"any check", "value > 10", "valid"},
		},
		{
			name:  "nested rules with failures",
			rule:  And("outer", And("inner", rule1, rule2), rule3),
			input: testInput{value: 5, valid: false},
			expected: []string{
				"outer",
				"inner",
				"value > 10",
				"valid",
			},
		},
		{
			name:     "NOT satisfied (child unsatisfied)",
			rule:     Not("not check", rule1),
			input:    testInput{value: 5},
			expected: []string{"value > 10"},
		},
		{
			name:     "NOT not satisfied (child satisfied)",
			rule:     Not("not check", rule1),
			input:    testInput{value: 50},
			expected: []string{"not check"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evaluator := NewEvaluator(tt.rule)
			result := evaluator.EvaluateDetailed(tt.input)

			unsatisfied := result.UnsatisfiedRules()

			if len(unsatisfied) != len(tt.expected) {
				t.Errorf(
					"Expected %d unsatisfied rules, got %d\nExpected: %v\nGot: %v",
					len(tt.expected),
					len(unsatisfied),
					tt.expected,
					unsatisfied,
				)
				return
			}

			// Check each expected rule is in the result
			for _, expectedRule := range tt.expected {
				found := false
				for _, actualRule := range unsatisfied {
					if actualRule == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"Expected rule %q not found in unsatisfied rules: %v",
						expectedRule,
						unsatisfied,
					)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
