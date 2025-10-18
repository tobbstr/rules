package rules

import (
	"errors"
	"testing"
)

type testInput struct {
	value int
	valid bool
}

func TestSimpleRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		predicate PredicateFunc[testInput]
		input     testInput
		want      bool
		wantErr   bool
	}{
		{
			name: "satisfied",
			predicate: func(input testInput,
			) (bool, error) {
				return input.value > 10, nil
			},
			input: testInput{value: 15},
			want:  true,
		},
		{
			name: "not satisfied",
			predicate: func(input testInput,
			) (bool, error) {
				return input.value > 10, nil
			},
			input: testInput{value: 5},
			want:  false,
		},
		{
			name: "error",
			predicate: func(input testInput,
			) (bool, error) {
				return false, errors.New("test error")
			},
			input:   testInput{value: 5},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := New("test rule", tt.predicate)
			got, err := rule.Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Evaluate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSimpleRuleWithContext removed - context is no longer used in rules

func TestAndRule(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)
	rule2 := New(
		"value < 20",
		func(input testInput) (bool, error) {
			return input.value < 20, nil
		},
	)
	rule3 := New(
		"valid",
		func(input testInput) (bool, error) {
			return input.valid, nil
		},
	)

	tests := []struct {
		name    string
		rules   []Rule[testInput]
		input   testInput
		want    bool
		wantErr bool
	}{
		{
			name:  "all satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 15, valid: true},
			want:  true,
		},
		{
			name:  "one not satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 15, valid: false},
			want:  false,
		},
		{
			name:  "none satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 5, valid: false},
			want:  false,
		},
		{
			name:    "empty rules",
			rules:   []Rule[testInput]{},
			input:   testInput{value: 15},
			want:    false,
			wantErr: true,
		},
		{
			name:    "nil rule",
			rules:   []Rule[testInput]{rule1, nil, rule2},
			input:   testInput{value: 15},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := And("test AND", tt.rules...)
			got, err := rule.Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Evaluate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrRule(t *testing.T) {
	t.Parallel()

	rule1 := New(
		"value > 20",
		func(input testInput) (bool, error) {
			return input.value > 20, nil
		},
	)
	rule2 := New(
		"value < 10",
		func(input testInput) (bool, error) {
			return input.value < 10, nil
		},
	)
	rule3 := New(
		"valid",
		func(input testInput) (bool, error) {
			return input.valid, nil
		},
	)

	tests := []struct {
		name    string
		rules   []Rule[testInput]
		input   testInput
		want    bool
		wantErr bool
	}{
		{
			name:  "first satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 25, valid: false},
			want:  true,
		},
		{
			name:  "last satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 15, valid: true},
			want:  true,
		},
		{
			name:  "none satisfied",
			rules: []Rule[testInput]{rule1, rule2, rule3},
			input: testInput{value: 15, valid: false},
			want:  false,
		},
		{
			name:    "empty rules",
			rules:   []Rule[testInput]{},
			input:   testInput{value: 15},
			want:    false,
			wantErr: true,
		},
		{
			name:    "nil rule",
			rules:   []Rule[testInput]{rule1, nil, rule2},
			input:   testInput{value: 15},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := Or("test OR", tt.rules...)
			got, err := rule.Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Evaluate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotRule(t *testing.T) {
	t.Parallel()

	rule := New(
		"value > 10",
		func(input testInput) (bool, error) {
			return input.value > 10, nil
		},
	)

	tests := []struct {
		name    string
		rule    Rule[testInput]
		input   testInput
		want    bool
		wantErr bool
	}{
		{
			name:  "negation of satisfied",
			rule:  rule,
			input: testInput{value: 15},
			want:  false,
		},
		{
			name:  "negation of not satisfied",
			rule:  rule,
			input: testInput{value: 5},
			want:  true,
		},
		{
			name:    "nil rule",
			rule:    nil,
			input:   testInput{value: 5},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			notRule := Not("test NOT", tt.rule)
			got, err := notRule.Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Evaluate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleNameAndDescription(t *testing.T) {
	t.Parallel()

	t.Run("simple rule with name", func(t *testing.T) {
		t.Parallel()

		rule := New(
			"test rule",
			func(input testInput) (bool, error) {
				return true, nil
			},
		)

		if rule.Name() != "test rule" {
			t.Errorf("Name() = %v, want %v", rule.Name(), "test rule")
		}
	})

	t.Run("simple rule with description", func(t *testing.T) {
		t.Parallel()

		rule := NewWithDescription(
			"test rule",
			"this is a test rule",
			func(input testInput) (bool, error) {
				return true, nil
			},
		)

		if rule.Name() != "test rule" {
			t.Errorf("Name() = %v, want %v", rule.Name(), "test rule")
		}
		if rule.Description() != "this is a test rule" {
			t.Errorf(
				"Description() = %v, want %v",
				rule.Description(),
				"this is a test rule",
			)
		}
	})
}

func TestHierarchicalRules(t *testing.T) {
	t.Parallel()

	// Create a complex hierarchical rule structure
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

	rangeRule := And("value in range", rule1, rule2)
	complexRule := Or("valid or in range", rule3, rangeRule)

	tests := []struct {
		name  string
		input testInput
		want  bool
	}{
		{
			name:  "valid but out of range",
			input: testInput{value: 5, valid: true},
			want:  true,
		},
		{
			name:  "invalid but in range",
			input: testInput{value: 50, valid: false},
			want:  true,
		},
		{
			name:  "valid and in range",
			input: testInput{value: 50, valid: true},
			want:  true,
		},
		{
			name:  "invalid and out of range",
			input: testInput{value: 5, valid: false},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := complexRule.Evaluate(tt.input)
			if err != nil {
				t.Errorf("Evaluate() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
