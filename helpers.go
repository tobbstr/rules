package rules

// Always creates a rule that is always satisfied.
func Always[T any](name string) Rule[T] {
	return New(name, func(input T) (bool, error) {
		return true, nil
	})
}

// Never creates a rule that is never satisfied.
func Never[T any](name string) Rule[T] {
	return New(name, func(input T) (bool, error) {
		return false, nil
	})
}

// AllOf is an alias for And that creates a rule satisfied only if all
// provided rules are satisfied.
func AllOf[T any](name string, rules ...Rule[T]) Rule[T] {
	return And(name, rules...)
}

// AnyOf is an alias for Or that creates a rule satisfied if at least one
// of the provided rules is satisfied.
func AnyOf[T any](name string, rules ...Rule[T]) Rule[T] {
	return Or(name, rules...)
}

// NoneOf creates a rule that is satisfied only if none of the provided
// rules are satisfied.
func NoneOf[T any](name string, rules ...Rule[T]) Rule[T] {
	return Not(name, Or(name+" (internal)", rules...))
}

// AtLeast creates a rule that is satisfied if at least n of the provided
// rules are satisfied. Automatically inherits domains from child rules.
func AtLeast[T any](
	name string,
	n int,
	rules ...Rule[T],
) Rule[T] {
	predicate := func(input T) (bool, error) {
		satisfied := 0
		for _, rule := range rules {
			if rule == nil {
				continue
			}
			result, err := rule.Evaluate(input)
			if err != nil {
				return false, err
			}
			if result {
				satisfied++
				if satisfied >= n {
					return true, nil
				}
			}
		}
		return satisfied >= n, nil
	}
	rule := New(name, predicate)

	// Collect and deduplicate domains from children
	domains := collectDomainsFromRules(rules)
	if len(domains) > 0 {
		_ = Register(rule, WithDomains(domains...))
	}

	return rule
}

// Exactly creates a rule that is satisfied if exactly n of the provided
// rules are satisfied. Automatically inherits domains from child rules.
func Exactly[T any](
	name string,
	n int,
	rules ...Rule[T],
) Rule[T] {
	predicate := func(input T) (bool, error) {
		satisfied := 0
		for _, rule := range rules {
			if rule == nil {
				continue
			}
			result, err := rule.Evaluate(input)
			if err != nil {
				return false, err
			}
			if result {
				satisfied++
			}
		}
		return satisfied == n, nil
	}
	rule := New(name, predicate)

	// Collect and deduplicate domains from children
	domains := collectDomainsFromRules(rules)
	if len(domains) > 0 {
		_ = Register(rule, WithDomains(domains...))
	}

	return rule
}

// AtMost creates a rule that is satisfied if at most n of the provided
// rules are satisfied. Automatically inherits domains from child rules.
func AtMost[T any](
	name string,
	n int,
	rules ...Rule[T],
) Rule[T] {
	predicate := func(input T) (bool, error) {
		satisfied := 0
		for _, rule := range rules {
			if rule == nil {
				continue
			}
			result, err := rule.Evaluate(input)
			if err != nil {
				return false, err
			}
			if result {
				satisfied++
				if satisfied > n {
					return false, nil
				}
			}
		}
		return true, nil
	}
	rule := New(name, predicate)

	// Collect and deduplicate domains from children
	domains := collectDomainsFromRules(rules)
	if len(domains) > 0 {
		_ = Register(rule, WithDomains(domains...))
	}

	return rule
}
