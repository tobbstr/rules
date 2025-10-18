package rules

// Builder provides a fluent API for constructing complex business rules.
type Builder[T any] struct {
	rules []Rule[T]
}

// NewBuilder creates a new rule builder.
func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{
		rules: make([]Rule[T], 0),
	}
}

// Add adds a rule to the builder.
func (b *Builder[T]) Add(rule Rule[T]) *Builder[T] {
	b.rules = append(b.rules, rule)
	return b
}

// AddSimple adds a simple rule with a predicate function.
func (b *Builder[T]) AddSimple(
	name string,
	predicate PredicateFunc[T],
) *Builder[T] {
	b.rules = append(b.rules, New(name, predicate))
	return b
}

// AddCondition adds a simple rule with a synchronous boolean predicate.
func (b *Builder[T]) AddCondition(
	name string,
	condition func(T) bool,
) *Builder[T] {
	predicate := func(input T) (bool, error) {
		return condition(input), nil
	}
	b.rules = append(b.rules, New(name, predicate))
	return b
}

// BuildAnd builds a rule that is satisfied only if all added rules are
// satisfied.
func (b *Builder[T]) BuildAnd(name string) Rule[T] {
	return And(name, b.rules...)
}

// BuildOr builds a rule that is satisfied if at least one added rule is
// satisfied.
func (b *Builder[T]) BuildOr(name string) Rule[T] {
	return Or(name, b.rules...)
}

// Clear clears all rules from the builder.
func (b *Builder[T]) Clear() *Builder[T] {
	b.rules = make([]Rule[T], 0)
	return b
}

// Count returns the number of rules in the builder.
func (b *Builder[T]) Count() int {
	return len(b.rules)
}
