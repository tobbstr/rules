package rules

import (
	"testing"
)

// TestDescriptions tests the Description methods for hierarchical rules
func TestDescriptions(t *testing.T) {
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

	t.Run("AND rule description", func(t *testing.T) {
		t.Parallel()

		andRule := And("test AND", rule1, rule2)
		desc := andRule.Description()

		if desc == "" {
			t.Error("Expected non-empty description")
		}

		// Check that it contains rule names
		expectedSubstrings := []string{"test AND", "ALL OF", "rule1", "rule2"}
		for _, substr := range expectedSubstrings {
			if !contains(desc, substr) {
				t.Errorf(
					"Expected description to contain %q, got %q",
					substr,
					desc,
				)
			}
		}
	})

	t.Run("OR rule description", func(t *testing.T) {
		t.Parallel()

		orRule := Or("test OR", rule1, rule2)
		desc := orRule.Description()

		if desc == "" {
			t.Error("Expected non-empty description")
		}

		// Check that it contains rule names
		expectedSubstrings := []string{"test OR", "ANY OF", "rule1", "rule2"}
		for _, substr := range expectedSubstrings {
			if !contains(desc, substr) {
				t.Errorf(
					"Expected description to contain %q, got %q",
					substr,
					desc,
				)
			}
		}
	})

	t.Run("NOT rule description", func(t *testing.T) {
		t.Parallel()

		notRule := Not("test NOT", rule1)
		desc := notRule.Description()

		if desc == "" {
			t.Error("Expected non-empty description")
		}

		// Check that it contains rule name
		expectedSubstrings := []string{"test NOT", "NOT", "rule1"}
		for _, substr := range expectedSubstrings {
			if !contains(desc, substr) {
				t.Errorf(
					"Expected description to contain %q, got %q",
					substr,
					desc,
				)
			}
		}
	})

	t.Run("NOT rule name", func(t *testing.T) {
		t.Parallel()

		notRule := Not("test NOT", rule1)
		name := notRule.Name()

		if name != "test NOT" {
			t.Errorf("Expected name %q, got %q", "test NOT", name)
		}
	})
}

// TestContextCancellation removed - context no longer used

// TestNestedErrorPropagation tests that errors propagate correctly through
// hierarchical rules
func TestNestedErrorPropagation(t *testing.T) {
	t.Parallel()

	errorRule := New(
		"error rule",
		func(input testInput) (bool, error) {
			return false, ErrEvaluationFailed
		},
	)

	successRule := New(
		"success rule",
		func(input testInput) (bool, error) {
			return true, nil
		},
	)

	t.Run("error in AND rule", func(t *testing.T) {
		t.Parallel()

		andRule := And("test AND", successRule, errorRule)

		_, err := andRule.Evaluate(testInput{value: 15})
		if err == nil {
			t.Error("Expected error to propagate")
		}
	})

	t.Run("error in OR rule", func(t *testing.T) {
		t.Parallel()

		orRule := Or("test OR", errorRule, successRule)

		_, err := orRule.Evaluate(testInput{value: 15})
		if err == nil {
			t.Error("Expected error to propagate")
		}
	})

	t.Run("error in NOT rule", func(t *testing.T) {
		t.Parallel()

		notRule := Not("test NOT", errorRule)

		_, err := notRule.Evaluate(testInput{value: 15})
		if err == nil {
			t.Error("Expected error to propagate")
		}
	})
}

// TestEdgeCases tests edge cases not covered by other tests
func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("AtLeast with all nil rules", func(t *testing.T) {
		t.Parallel()

		rule := AtLeast[testInput]("at least", 0, nil, nil)

		satisfied, err := rule.Evaluate(testInput{value: 15})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied (0 of any is satisfied)")
		}
	})

	t.Run("Exactly with all nil rules", func(t *testing.T) {
		t.Parallel()

		rule := Exactly[testInput]("exactly", 0, nil, nil)

		satisfied, err := rule.Evaluate(testInput{value: 15})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !satisfied {
			t.Error("Expected rule to be satisfied (exactly 0 nil rules)")
		}
	})
}
