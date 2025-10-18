// Package rules provides a type-safe, expressive library for defining and
// evaluating hierarchical business rules.
package rules

import (
	"errors"
	"fmt"
)

var (
	// ErrNilRule is returned when a nil rule is provided.
	ErrNilRule = errors.New("nil rule")
	// ErrEmptyRules is returned when an empty rules list is provided.
	ErrEmptyRules = errors.New("empty rules list")
	// ErrEvaluationFailed is returned when rule evaluation fails.
	ErrEvaluationFailed = errors.New("evaluation failed")
)

// Rule represents a business rule that can be evaluated against an input
// of type T.
// Rules can be composed hierarchically to form complex business logic.
type Rule[T any] interface {
	// Evaluate checks if the rule is satisfied by the given input.
	// It returns true if the rule is met, false otherwise, and any error
	// that occurred during evaluation.
	Evaluate(input T) (bool, error)

	// Name returns a human-readable name for the rule.
	Name() string
}

// PredicateFunc is a function that evaluates a condition against an input.
type PredicateFunc[T any] func(input T) (bool, error)

// simpleRule is a basic rule implementation that wraps a predicate function.
type simpleRule[T any] struct {
	name        string
	description string
	predicate   PredicateFunc[T]
}

// New creates a new simple rule with the given name and predicate function.
func New[T any](name string, predicate PredicateFunc[T]) Rule[T] {
	return &simpleRule[T]{
		name:        name,
		description: name,
		predicate:   predicate,
	}
}

// NewWithDescription creates a new simple rule with a name, description,
// and predicate function.
func NewWithDescription[T any](
	name, description string,
	predicate PredicateFunc[T],
) Rule[T] {
	return &simpleRule[T]{
		name:        name,
		description: description,
		predicate:   predicate,
	}
}

// NewWithDomain creates and automatically registers a rule with a single domain.
func NewWithDomain[T any](
	name string,
	domain Domain,
	predicate PredicateFunc[T],
) Rule[T] {
	rule := New(name, predicate)
	_ = Register(rule, WithDomain(domain))
	return rule
}

// NewWithGroup creates and automatically registers a cross-domain rule with a group name.
func NewWithGroup[T any](
	name string,
	groupName string,
	domains []Domain,
	predicate PredicateFunc[T],
) Rule[T] {
	rule := New(name, predicate)
	_ = Register(rule, WithGroup(groupName, domains...))
	return rule
}

// WithDescription updates the registry entry with a description and returns the same rule.
// This uses pointer-equality lookup in the registry.
func WithDescription[T any](rule Rule[T], description string) Rule[T] {
	_ = UpdateDescription(rule, description)
	return rule
}

func (r *simpleRule[T]) Evaluate(input T) (bool, error) {
	result, err := r.predicate(input)
	if err != nil {
		return false, fmt.Errorf(
			"evaluating rule %q: %w",
			r.name,
			err,
		)
	}

	return result, nil
}

func (r *simpleRule[T]) Name() string {
	return r.name
}

// andRule represents a logical AND of multiple rules.
type andRule[T any] struct {
	name  string
	rules []Rule[T]
}

// And creates a rule that is satisfied only if all provided rules are
// satisfied. Automatically inherits domains from child rules.
func And[T any](name string, rules ...Rule[T]) Rule[T] {
	rule := &andRule[T]{
		name:  name,
		rules: rules,
	}

	// Collect and deduplicate domains from children
	domains := collectDomainsFromRules(rules)
	if len(domains) > 0 {
		_ = Register(rule, WithDomains(domains...))
	}

	return rule
}

func (r *andRule[T]) Evaluate(input T) (bool, error) {
	if len(r.rules) == 0 {
		return false, fmt.Errorf(
			"evaluating AND rule %q: %w",
			r.name,
			ErrEmptyRules,
		)
	}

	for _, rule := range r.rules {
		if rule == nil {
			return false, fmt.Errorf(
				"evaluating AND rule %q: %w",
				r.name,
				ErrNilRule,
			)
		}

		satisfied, err := rule.Evaluate(input)
		if err != nil {
			return false, fmt.Errorf(
				"evaluating AND rule %q: %w",
				r.name,
				err,
			)
		}

		if !satisfied {
			return false, nil
		}
	}

	return true, nil
}

func (r *andRule[T]) Name() string {
	return r.name
}

// orRule represents a logical OR of multiple rules.
type orRule[T any] struct {
	name  string
	rules []Rule[T]
}

// Or creates a rule that is satisfied if at least one of the provided rules
// is satisfied. Automatically inherits domains from child rules.
func Or[T any](name string, rules ...Rule[T]) Rule[T] {
	rule := &orRule[T]{
		name:  name,
		rules: rules,
	}

	// Collect and deduplicate domains from children
	domains := collectDomainsFromRules(rules)
	if len(domains) > 0 {
		_ = Register(rule, WithDomains(domains...))
	}

	return rule
}

func (r *orRule[T]) Evaluate(input T) (bool, error) {
	if len(r.rules) == 0 {
		return false, fmt.Errorf(
			"evaluating OR rule %q: %w",
			r.name,
			ErrEmptyRules,
		)
	}

	for _, rule := range r.rules {
		if rule == nil {
			return false, fmt.Errorf(
				"evaluating OR rule %q: %w",
				r.name,
				ErrNilRule,
			)
		}

		satisfied, err := rule.Evaluate(input)
		if err != nil {
			return false, fmt.Errorf(
				"evaluating OR rule %q: %w",
				r.name,
				err,
			)
		}

		if satisfied {
			return true, nil
		}
	}

	return false, nil
}

func (r *orRule[T]) Name() string {
	return r.name
}

// notRule represents a logical NOT of a rule.
type notRule[T any] struct {
	name string
	rule Rule[T]
}

// Not creates a rule that is satisfied only if the provided rule is not
// satisfied. Automatically inherits domains from the child rule.
func Not[T any](name string, rule Rule[T]) Rule[T] {
	notRule := &notRule[T]{
		name: name,
		rule: rule,
	}

	// Collect domains from child
	domains := collectDomainsFromRules([]Rule[T]{rule})
	if len(domains) > 0 {
		_ = Register(notRule, WithDomains(domains...))
	}

	return notRule
}

func (r *notRule[T]) Evaluate(input T) (bool, error) {
	if r.rule == nil {
		return false, fmt.Errorf(
			"evaluating NOT rule %q: %w",
			r.name,
			ErrNilRule,
		)
	}

	satisfied, err := r.rule.Evaluate(input)
	if err != nil {
		return false, fmt.Errorf(
			"evaluating NOT rule %q: %w",
			r.name,
			err,
		)
	}

	return !satisfied, nil
}

func (r *notRule[T]) Name() string {
	return r.name
}

// collectDomainsFromRules collects and deduplicates domains from child rules.
func collectDomainsFromRules[T any](rules []Rule[T]) []Domain {
	domainSet := make(map[Domain]bool)

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		// Look up the rule in the registry to get its domains
		allRules := AllRules()
		ptr := getRulePointer(rule)

		for _, registered := range allRules {
			if getRulePointer(registered.Rule) == ptr {
				for _, domain := range registered.Domains {
					domainSet[domain] = true
				}
				break
			}
		}
	}

	// Convert set to slice
	result := make([]Domain, 0, len(domainSet))
	for domain := range domainSet {
		result = append(result, domain)
	}

	return result
}
