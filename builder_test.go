package rules

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	t.Parallel()

	t.Run("add simple rules", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		rule1 := New(
			"value > 10",
			func(input testInput) (bool, error) {
				return input.value > 10, nil
			},
		)
		rule2 := New(
			"valid",
			func(input testInput) (bool, error) {
				return input.valid, nil
			},
		)

		builder.Add(rule1).Add(rule2)

		if builder.Count() != 2 {
			t.Errorf("Expected 2 rules, got %d", builder.Count())
		}
	})

	t.Run("build AND rule", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		builder.
			AddSimple(
				"value > 10",
				func(input testInput,
				) (bool, error) {
					return input.value > 10, nil
				},
			).
			AddSimple(
				"value < 100",
				func(input testInput,
				) (bool, error) {
					return input.value < 100, nil
				},
			)

		rule := builder.BuildAnd("range check")

		satisfied, err := rule.Evaluate(testInput{value: 50})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})

	t.Run("build OR rule", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		builder.
			AddSimple(
				"value < 10",
				func(input testInput,
				) (bool, error) {
					return input.value < 10, nil
				},
			).
			AddSimple(
				"value > 100",
				func(input testInput,
				) (bool, error) {
					return input.value > 100, nil
				},
			)

		rule := builder.BuildOr("out of range")

		satisfied, err := rule.Evaluate(testInput{value: 5})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})

	t.Run("add condition", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		builder.
			AddCondition("value > 10", func(input testInput) bool {
				return input.value > 10
			}).
			AddCondition("valid", func(input testInput) bool {
				return input.valid
			})

		rule := builder.BuildAnd("conditions")

		satisfied, err := rule.Evaluate(testInput{value: 50, valid: true})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})

	t.Run("clear builder", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		builder.AddCondition("value > 10", func(input testInput) bool {
			return input.value > 10
		})

		if builder.Count() != 1 {
			t.Errorf("Expected 1 rule, got %d", builder.Count())
		}

		builder.Clear()

		if builder.Count() != 0 {
			t.Errorf("Expected 0 rules after clear, got %d", builder.Count())
		}
	})

	t.Run("reuse builder", func(t *testing.T) {
		t.Parallel()

		builder := NewBuilder[testInput]()

		builder.AddCondition("value > 10", func(input testInput) bool {
			return input.value > 10
		})

		rule1 := builder.BuildAnd("first rule")

		builder.Clear()

		builder.AddCondition("value < 100", func(input testInput) bool {
			return input.value < 100
		})

		rule2 := builder.BuildAnd("second rule")

		// Test first rule
		satisfied1, err1 := rule1.Evaluate(testInput{value: 50})
		if err1 != nil {
			t.Errorf("Unexpected error: %v", err1)
		}
		if !satisfied1 {
			t.Error("Expected first rule to be satisfied")
		}

		// Test second rule
		satisfied2, err2 := rule2.Evaluate(testInput{value: 50})
		if err2 != nil {
			t.Errorf("Unexpected error: %v", err2)
		}
		if !satisfied2 {
			t.Error("Expected second rule to be satisfied")
		}
	})
}

func TestBuilderComplexHierarchy(t *testing.T) {
	t.Parallel()

	// Build a complex hierarchical rule using the builder
	rangeBuilder := NewBuilder[testInput]()
	rangeBuilder.
		AddCondition("value > 10", func(input testInput) bool {
			return input.value > 10
		}).
		AddCondition("value < 100", func(input testInput) bool {
			return input.value < 100
		})

	rangeRule := rangeBuilder.BuildAnd("in range")

	validBuilder := NewBuilder[testInput]()
	validBuilder.
		Add(rangeRule).
		AddCondition("valid", func(input testInput) bool {
			return input.valid
		})

	finalRule := validBuilder.BuildAnd("valid and in range")

	tests := []struct {
		name  string
		input testInput
		want  bool
	}{
		{
			name:  "all satisfied",
			input: testInput{value: 50, valid: true},
			want:  true,
		},
		{
			name:  "invalid",
			input: testInput{value: 50, valid: false},
			want:  false,
		},
		{
			name:  "out of range",
			input: testInput{value: 5, valid: true},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := finalRule.Evaluate(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
