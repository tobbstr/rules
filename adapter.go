package rules

import (
	"fmt"
)

// MapperFunc is a function that transforms an input of type TSource to TTarget.
type MapperFunc[TSource, TTarget any] func(TSource) TTarget

// Map transforms a Rule[TTarget] into a Rule[TSource] by applying a mapping
// function to convert TSource to TTarget before evaluation.
//
// This allows combining rules that operate on different types by providing
// appropriate mapping functions.
//
// Example:
//
//	// Rule that operates on User
//	userRule := rules.New("user is admin",
//	    func(user User) (bool, error) {
//	        return user.Role == "admin", nil
//	    },
//	)
//
//	// Transform to operate on Request (which contains User)
//	requestRule := rules.Map(
//	    "request from admin",
//	    userRule,
//	    func(req Request) User { return req.User },
//	)
func Map[TSource, TTarget any](
	name string,
	rule Rule[TTarget],
	mapper MapperFunc[TSource, TTarget],
) Rule[TSource] {
	return &mappedRule[TSource, TTarget]{
		name:   name,
		rule:   rule,
		mapper: mapper,
	}
}

// mappedRule wraps a rule and applies a mapping function to the input.
type mappedRule[TSource, TTarget any] struct {
	name   string
	rule   Rule[TTarget]
	mapper MapperFunc[TSource, TTarget]
}

func (r *mappedRule[TSource, TTarget]) Evaluate(
	input TSource,
) (bool, error) {
	if r.rule == nil {
		return false, fmt.Errorf(
			"evaluating mapped rule %q: %w",
			r.name,
			ErrNilRule,
		)
	}

	// Apply mapping function
	target := r.mapper(input)

	// Evaluate the underlying rule
	satisfied, err := r.rule.Evaluate(target)
	if err != nil {
		return false, fmt.Errorf(
			"evaluating mapped rule %q: %w",
			r.name,
			err,
		)
	}

	return satisfied, nil
}

func (r *mappedRule[TSource, TTarget]) Name() string {
	return r.name
}

func (r *mappedRule[TSource, TTarget]) Description() string {
	return fmt.Sprintf(
		"%s: MAPPED(%s)",
		r.name,
		r.rule.Name(),
	)
}

// Combine allows combining rules from different type hierarchies by providing
// a combined input type and extractors for each rule's input type.
//
// Example:
//
//	type OrderRequest struct {
//	    Order Order
//	    User  User
//	}
//
//	combined := rules.And("order request validation",
//	    rules.Map("user check", userRule,
//	        func(req OrderRequest) User { return req.User }),
//	    rules.Map("order check", orderRule,
//	        func(req OrderRequest) Order { return req.Order }),
//	)
//
// This is a convenience function that makes the pattern more explicit.
func Combine[TCombined, T1, T2 any](
	name string,
	rule1 Rule[T1],
	extractor1 MapperFunc[TCombined, T1],
	rule2 Rule[T2],
	extractor2 MapperFunc[TCombined, T2],
) Rule[TCombined] {
	return And(
		name,
		Map("extracted-1", rule1, extractor1),
		Map("extracted-2", rule2, extractor2),
	)
}

// Combine3 combines three rules from different types into a single rule.
func Combine3[TCombined, T1, T2, T3 any](
	name string,
	rule1 Rule[T1],
	extractor1 MapperFunc[TCombined, T1],
	rule2 Rule[T2],
	extractor2 MapperFunc[TCombined, T2],
	rule3 Rule[T3],
	extractor3 MapperFunc[TCombined, T3],
) Rule[TCombined] {
	return And(
		name,
		Map("extracted-1", rule1, extractor1),
		Map("extracted-2", rule2, extractor2),
		Map("extracted-3", rule3, extractor3),
	)
}

// CombineMany combines multiple mapped rules into a single AND rule.
// This is useful when you have many rules operating on different types.
func CombineMany[TCombined any](
	name string,
	mappedRules ...Rule[TCombined],
) Rule[TCombined] {
	return And(name, mappedRules...)
}
