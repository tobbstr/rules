package rules

import (
	"testing"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	rule := Always[testInput]("always")

	satisfied, err := rule.Evaluate(testInput{value: 0})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !satisfied {
		t.Error("Expected Always rule to always be satisfied")
	}
}

func TestNever(t *testing.T) {
	t.Parallel()

	rule := Never[testInput]("never")

	satisfied, err := rule.Evaluate(testInput{value: 0})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if satisfied {
		t.Error("Expected Never rule to never be satisfied")
	}
}

func TestAllOf(t *testing.T) {
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

	rule := AllOf("all", rule1, rule2)

	tests := []struct {
		name  string
		input testInput
		want  bool
	}{
		{
			name:  "all satisfied",
			input: testInput{value: 50},
			want:  true,
		},
		{
			name:  "not all satisfied",
			input: testInput{value: 5},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rule.Evaluate(tt.input)
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

func TestAnyOf(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value < 10",
		func(input testInput) (bool, error) {
			return input.value < 10, nil
		},
	)
	rule2 := New(
		"value > 100",
		func(input testInput) (bool, error) {
			return input.value > 100, nil
		},
	)

	rule := AnyOf("any", rule1, rule2)

	tests := []struct {
		name  string
		input testInput
		want  bool
	}{
		{
			name:  "first satisfied",
			input: testInput{value: 5},
			want:  true,
		},
		{
			name:  "second satisfied",
			input: testInput{value: 150},
			want:  true,
		},
		{
			name:  "none satisfied",
			input: testInput{value: 50},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rule.Evaluate(tt.input)
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

func TestNoneOf(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value < 10",
		func(input testInput) (bool, error) {
			return input.value < 10, nil
		},
	)
	rule2 := New(
		"value > 100",
		func(input testInput) (bool, error) {
			return input.value > 100, nil
		},
	)

	rule := NoneOf("none", rule1, rule2)

	tests := []struct {
		name  string
		input testInput
		want  bool
	}{
		{
			name:  "none satisfied",
			input: testInput{value: 50},
			want:  true,
		},
		{
			name:  "one satisfied",
			input: testInput{value: 5},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rule.Evaluate(tt.input)
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

func TestAtLeast(t *testing.T) {
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
		name  string
		n     int
		input testInput
		want  bool
	}{
		{
			name:  "at least 2 of 3 - all satisfied",
			n:     2,
			input: testInput{value: 50, valid: true},
			want:  true,
		},
		{
			name:  "at least 2 of 3 - exactly 2 satisfied",
			n:     2,
			input: testInput{value: 50, valid: false},
			want:  true,
		},
		{
			name:  "at least 2 of 3 - only 1 satisfied",
			n:     2,
			input: testInput{value: 5, valid: false},
			want:  false,
		},
		{
			name:  "at least 3 of 3 - all satisfied",
			n:     3,
			input: testInput{value: 50, valid: true},
			want:  true,
		},
		{
			name:  "at least 3 of 3 - only 2 satisfied",
			n:     3,
			input: testInput{value: 50, valid: false},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := AtLeast("at least", tt.n, rule1, rule2, rule3)

			got, err := rule.Evaluate(tt.input)
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

func TestExactly(t *testing.T) {
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
		name  string
		n     int
		input testInput
		want  bool
	}{
		{
			name:  "exactly 2 of 3 - exactly 2 satisfied",
			n:     2,
			input: testInput{value: 50, valid: false},
			want:  true,
		},
		{
			name:  "exactly 2 of 3 - all satisfied",
			n:     2,
			input: testInput{value: 50, valid: true},
			want:  false,
		},
		{
			name:  "exactly 2 of 3 - only 1 satisfied",
			n:     2,
			input: testInput{value: 5, valid: false},
			want:  false,
		},
		{
			name:  "exactly 3 of 3 - all satisfied",
			n:     3,
			input: testInput{value: 50, valid: true},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := Exactly("exactly", tt.n, rule1, rule2, rule3)

			got, err := rule.Evaluate(tt.input)
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

func TestAtMost(t *testing.T) {
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
		name  string
		n     int
		input testInput
		want  bool
	}{
		{
			name:  "at most 2 of 3 - exactly 2 satisfied",
			n:     2,
			input: testInput{value: 50, valid: false},
			want:  true,
		},
		{
			name:  "at most 2 of 3 - all satisfied",
			n:     2,
			input: testInput{value: 50, valid: true},
			want:  false,
		},
		{
			name:  "at most 2 of 3 - only 1 satisfied",
			n:     2,
			input: testInput{value: 5, valid: false},
			want:  true,
		},
		{
			name:  "at most 3 of 3 - all satisfied",
			n:     3,
			input: testInput{value: 50, valid: true},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := AtMost("at most", tt.n, rule1, rule2, rule3)

			got, err := rule.Evaluate(tt.input)
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

func TestHelpersWithNilRules(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)

	t.Run("AtLeast with nil rule", func(t *testing.T) {
		t.Parallel()

		rule := AtLeast("at least", 1, rule1, nil)

		// Should ignore nil rule and work with just rule1
		satisfied, err := rule.Evaluate(testInput{value: 50})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})

	t.Run("Exactly with nil rule", func(t *testing.T) {
		t.Parallel()

		rule := Exactly("exactly", 1, rule1, nil)

		// Should ignore nil rule and work with just rule1
		satisfied, err := rule.Evaluate(testInput{value: 50})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})

	t.Run("AtMost with nil rule", func(t *testing.T) {
		t.Parallel()

		rule := AtMost("at most", 1, rule1, nil)

		// Should ignore nil rule and work with just rule1
		satisfied, err := rule.Evaluate(testInput{value: 50})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied")
		}
	})
}
