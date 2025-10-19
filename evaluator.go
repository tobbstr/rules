package rules

import (
	"fmt"
	"time"
)

// Result represents the result of evaluating a rule.
type Result struct {
	// Satisfied indicates whether the rule was satisfied.
	Satisfied bool
	// RuleName is the name of the evaluated rule.
	RuleName string
	// Duration is the time taken to evaluate the rule.
	Duration time.Duration
	// Error is any error that occurred during evaluation.
	Error error
	// Children contains results for child rules (for hierarchical rules).
	Children []Result
}

// Evaluator provides detailed evaluation of rules with result tracking.
type Evaluator[T any] struct {
	rule Rule[T]
}

// NewEvaluator creates a new evaluator for the given rule.
func NewEvaluator[T any](rule Rule[T]) *Evaluator[T] {
	return &Evaluator[T]{rule: rule}
}

// Evaluate evaluates the rule and returns a detailed result with timing information.
func (e *Evaluator[T]) Evaluate(input T) Result {
	start := time.Now()
	satisfied, err := e.rule.Evaluate(input)
	duration := time.Since(start)

	return Result{
		Satisfied: satisfied,
		RuleName:  e.rule.Name(),
		Duration:  duration,
		Error:     err,
	}
}

// EvaluateFast evaluates the rule without timing overhead for maximum performance.
// Use this when you don't need timing information in the result.
func (e *Evaluator[T]) EvaluateFast(input T) (bool, error) {
	return e.rule.Evaluate(input)
}

// EvaluateDetailed evaluates the rule and returns a detailed result
// including child rule results for hierarchical rules.
// This evaluates all children to provide a complete view.
func (e *Evaluator[T]) EvaluateDetailed(input T) Result {
	return e.evaluateRuleDetailed(e.rule, input, false)
}

// EvaluateDetailedShortCircuit evaluates the rule and returns a detailed result
// with short-circuit optimization. For AND rules, stops on first failure.
// For OR rules, stops on first success. This is faster but provides incomplete child results.
func (e *Evaluator[T]) EvaluateDetailedShortCircuit(input T) Result {
	return e.evaluateRuleDetailed(e.rule, input, true)
}

func (e *Evaluator[T]) evaluateRuleDetailed(
	rule Rule[T],
	input T,
	shortCircuit bool,
) Result {
	start := time.Now()

	var children []Result
	var satisfied bool
	var err error

	// Check if rule is hierarchical and evaluate children
	// Compute result directly from children to avoid double evaluation
	switch r := rule.(type) {
	case *andRule[T]:
		if len(r.rules) == 0 {
			err = ErrEmptyRules
			satisfied = false
		} else {
			satisfied = true
			children = make([]Result, 0, len(r.rules))
			for _, childRule := range r.rules {
				if childRule == nil {
					err = ErrNilRule
					satisfied = false
					break
				}
				childResult := e.evaluateRuleDetailed(childRule, input, shortCircuit)
				children = append(children, childResult)
				if childResult.Error != nil {
					err = childResult.Error
					satisfied = false
					break
				}
				if !childResult.Satisfied {
					satisfied = false
					if shortCircuit {
						break
					}
					// Continue evaluating remaining children for complete detailed view
				}
			}
		}
	case *orRule[T]:
		if len(r.rules) == 0 {
			err = ErrEmptyRules
			satisfied = false
		} else {
			satisfied = false
			children = make([]Result, 0, len(r.rules))
			for _, childRule := range r.rules {
				if childRule == nil {
					err = ErrNilRule
					satisfied = false
					break
				}
				childResult := e.evaluateRuleDetailed(childRule, input, shortCircuit)
				children = append(children, childResult)
				if childResult.Error != nil {
					err = childResult.Error
					satisfied = false
					break
				}
				if childResult.Satisfied {
					satisfied = true
					if shortCircuit {
						break
					}
					// Continue evaluating remaining children for complete detailed view
				}
			}
		}
	case *notRule[T]:
		if r.rule == nil {
			err = ErrNilRule
			satisfied = false
		} else {
			children = make([]Result, 0, 1)
			childResult := e.evaluateRuleDetailed(r.rule, input, shortCircuit)
			children = append(children, childResult)
			if childResult.Error != nil {
				err = childResult.Error
				satisfied = false
			} else {
				satisfied = !childResult.Satisfied
			}
		}
	default:
		// For simple rules, evaluate directly
		satisfied, err = rule.Evaluate(input)
	}

	duration := time.Since(start)

	return Result{
		Satisfied: satisfied,
		RuleName:  rule.Name(),
		Duration:  duration,
		Error:     err,
		Children:  children,
	}
}

// String returns a string representation of the result.
func (r Result) String() string {
	return r.stringWithIndent(0)
}

func (r Result) stringWithIndent(indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	status := "✓"
	if !r.Satisfied {
		status = "✗"
	}
	if r.Error != nil {
		status = "⚠"
	}

	result := fmt.Sprintf(
		"%s%s %s (took %v)",
		prefix,
		status,
		r.RuleName,
		r.Duration,
	)

	if r.Error != nil {
		result += fmt.Sprintf(" - Error: %v", r.Error)
	}

	for _, child := range r.Children {
		result += "\n" + child.stringWithIndent(indent+1)
	}

	return result
}

// IsSuccessful returns true if the rule was satisfied and no error occurred.
func (r Result) IsSuccessful() bool {
	return r.Satisfied && r.Error == nil
}

// HasError returns true if an error occurred during evaluation.
func (r Result) HasError() bool {
	return r.Error != nil
}
