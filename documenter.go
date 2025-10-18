package rules

import (
	"fmt"
	"reflect"
	"strings"
)

// DocumentFormat represents the output format for documentation.
type DocumentFormat int

const (
	// FormatMarkdown generates documentation in Markdown format.
	FormatMarkdown DocumentFormat = iota
	// FormatJSON generates documentation in JSON format.
	FormatJSON
	// FormatHTML generates documentation in HTML format.
	FormatHTML
	// FormatMermaid generates documentation as Mermaid diagrams.
	FormatMermaid
)

// DocumentOptions configures documentation generation.
type DocumentOptions struct {
	// Format specifies the output format
	Format DocumentFormat

	// IncludeExamples includes example inputs/outputs if available
	IncludeExamples bool

	// MaxDepth limits the depth of rule hierarchy to document (0 = unlimited)
	MaxDepth int

	// IncludeMetadata includes additional metadata in output
	IncludeMetadata bool

	// Title sets the document title
	Title string

	// Description sets the document description
	Description string

	// GroupByDomain organizes output by domain
	GroupByDomain bool

	// IncludeDomains filters to specific domains (empty = all)
	IncludeDomains []Domain

	// ExcludeDomains excludes specific domains
	ExcludeDomains []Domain

	// ShowCrossDomainLinks highlights cross-domain dependencies
	ShowCrossDomainLinks bool
}

// RuleType represents the type of a rule.
type RuleType int

const (
	// RuleTypeSimple represents a simple predicate-based rule.
	RuleTypeSimple RuleType = iota
	// RuleTypeAnd represents a logical AND rule.
	RuleTypeAnd
	// RuleTypeOr represents a logical OR rule.
	RuleTypeOr
	// RuleTypeNot represents a logical NOT rule.
	RuleTypeNot
	// RuleTypeUnknown represents an unknown rule type.
	RuleTypeUnknown
)

// String returns the string representation of a RuleType.
func (rt RuleType) String() string {
	switch rt {
	case RuleTypeSimple:
		return "SIMPLE"
	case RuleTypeAnd:
		return "AND"
	case RuleTypeOr:
		return "OR"
	case RuleTypeNot:
		return "NOT"
	default:
		return "UNKNOWN"
	}
}

// ruleNode represents a node in the rule hierarchy tree.
type ruleNode struct {
	Rule        any
	Name        string
	Type        RuleType
	Description string
	Domains     []Domain
	Group       string
	Metadata    *RuleMetadata
	Children    []*ruleNode
	Depth       int
}

// getRuleType detects the type of a rule through reflection.
func getRuleType(rule any) RuleType {
	// Use reflection to check the underlying type name
	typeName := fmt.Sprintf("%T", rule)

	if strings.Contains(typeName, "andRule") {
		return RuleTypeAnd
	}
	if strings.Contains(typeName, "orRule") {
		return RuleTypeOr
	}
	if strings.Contains(typeName, "notRule") {
		return RuleTypeNot
	}
	if strings.Contains(typeName, "simpleRule") {
		return RuleTypeSimple
	}

	return RuleTypeUnknown
}

// getChildren extracts child rules from a hierarchical rule using reflection.
func getChildren(rule any) []any {
	// Use reflection to call the Children() or Child() method
	v := reflect.ValueOf(rule)

	// Try Children() method for And/Or rules
	childrenMethod := v.MethodByName("Children")
	if childrenMethod.IsValid() {
		results := childrenMethod.Call(nil)
		if len(results) == 1 {
			childrenVal := results[0]
			if childrenVal.Kind() == reflect.Slice {
				result := make([]any, childrenVal.Len())
				for i := 0; i < childrenVal.Len(); i++ {
					result[i] = childrenVal.Index(i).Interface()
				}
				return result
			}
		}
	}

	// Try Child() method for Not rules
	childMethod := v.MethodByName("Child")
	if childMethod.IsValid() {
		results := childMethod.Call(nil)
		if len(results) == 1 {
			return []any{results[0].Interface()}
		}
	}

	return nil
}

// getRuleName extracts the name from a rule.
func getRuleName(rule any) string {
	// Try to cast to common rule interfaces
	type namedRule interface {
		Name() string
	}

	if nr, ok := rule.(namedRule); ok {
		return nr.Name()
	}

	return "unnamed"
}

// buildRuleTree constructs a tree representation of a rule hierarchy.
func buildRuleTree(rule any, registered *RegisteredRule, depth int, maxDepth int) *ruleNode {
	if maxDepth > 0 && depth >= maxDepth {
		return nil
	}

	node := &ruleNode{
		Rule:  rule,
		Name:  getRuleName(rule),
		Type:  getRuleType(rule),
		Depth: depth,
	}

	// Add metadata from registry if available
	if registered != nil {
		node.Description = registered.Description
		node.Domains = registered.Domains
		node.Group = registered.Group
		node.Metadata = registered.Metadata
	}

	// Build children
	children := getChildren(rule)
	for _, child := range children {
		if child == nil {
			continue
		}

		// Look up child in registry
		var childRegistered *RegisteredRule
		allRules := AllRules()
		childPtr := getRulePointer(child)
		for _, r := range allRules {
			if getRulePointer(r.Rule) == childPtr {
				childRegistered = &r
				break
			}
		}

		childNode := buildRuleTree(child, childRegistered, depth+1, maxDepth)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// groupRulesByGroup groups rules by their group name.
func groupRulesByGroup(rules []RegisteredRule) map[string][]RegisteredRule {
	grouped := make(map[string][]RegisteredRule)

	for _, rule := range rules {
		if rule.Group == "" {
			// Rules without group go by domain or ungrouped
			if len(rule.Domains) == 0 {
				grouped["Ungrouped"] = append(grouped["Ungrouped"], rule)
			} else {
				// Use primary domain as group
				grouped[string(rule.Domains[0])] = append(grouped[string(rule.Domains[0])], rule)
			}
		} else {
			grouped[rule.Group] = append(grouped[rule.Group], rule)
		}
	}

	return grouped
}
