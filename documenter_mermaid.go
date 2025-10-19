package rules

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateMermaid generates Mermaid diagram documentation for all registered rules.
func GenerateMermaid(opts DocumentOptions) (string, error) {
	return GenerateMermaidFromRules(AllRules(), opts)
}

// GenerateMermaidFromRules generates Mermaid diagram documentation from a list of registered rules.
func GenerateMermaidFromRules(rules []RegisteredRule, opts DocumentOptions) (string, error) {
	var sb strings.Builder

	// Filter rules based on options
	filtered := filterRegisteredRules(rules, opts)

	if len(filtered) == 0 {
		return "graph TD\n    Empty[No rules match filters]", nil
	}

	// Start with flowchart header
	sb.WriteString("graph TD\n")

	if opts.GroupByDomain {
		return generateMermaidByDomain(filtered, opts, &sb)
	}

	return generateMermaidFlat(filtered, opts, &sb)
}

// GenerateDomainMermaid generates Mermaid diagram for a specific domain.
func GenerateDomainMermaid(domain Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomain(domain)
	opts.GroupByDomain = true
	return GenerateMermaidFromRules(rules, opts)
}

// GenerateDomainsMermaid generates Mermaid diagram for multiple domains.
func GenerateDomainsMermaid(domains []Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomains(domains...)
	opts.GroupByDomain = true
	return GenerateMermaidFromRules(rules, opts)
}

// GenerateGroupMermaid generates Mermaid diagram for a specific group.
func GenerateGroupMermaid(groupName string, opts DocumentOptions) (string, error) {
	rules := RulesByGroup(groupName)
	return GenerateMermaidFromRules(rules, opts)
}

// generateMermaidByDomain generates domain-grouped Mermaid diagrams.
func generateMermaidByDomain(rules []RegisteredRule, opts DocumentOptions, sb *strings.Builder) (string, error) {
	// Group rules by domain
	domainMap := make(map[Domain][]RegisteredRule)
	for _, rule := range rules {
		if len(rule.Domains) == 0 {
			// Add to "ungrouped" domain
			domainMap[Domain("Ungrouped")] = append(domainMap[Domain("Ungrouped")], rule)
		} else {
			for _, domain := range rule.Domains {
				domainMap[domain] = append(domainMap[domain], rule)
			}
		}
	}

	// Sort domains for consistent output
	domains := make([]Domain, 0, len(domainMap))
	for domain := range domainMap {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool {
		return string(domains[i]) < string(domains[j])
	})

	// Track which rules have been processed to avoid duplicates
	processedRules := make(map[uintptr]bool)

	// Generate subgraphs for each domain
	for _, domain := range domains {
		domainRules := domainMap[domain]

		// Create subgraph for domain
		sb.WriteString(fmt.Sprintf("\n    subgraph %s[\"%s Domain\"]\n",
			sanitizeMermaidID(string(domain)),
			domain))

		// Add rules in this domain
		for _, regRule := range domainRules {
			rulePtr := getRulePointer(regRule.Rule)
			if processedRules[rulePtr] {
				continue
			}
			processedRules[rulePtr] = true

			writeMermaidRule(sb, &regRule, opts, 2)
		}

		sb.WriteString("    end\n")
	}

	// Add connections between rules
	sb.WriteString("\n    %% Rule connections\n")
	processedRules = make(map[uintptr]bool)
	for _, regRule := range rules {
		rulePtr := getRulePointer(regRule.Rule)
		if processedRules[rulePtr] {
			continue
		}
		processedRules[rulePtr] = true

		writeMermaidConnections(sb, &regRule, opts)
	}

	// Add styling
	writeMermaidStyling(sb)

	return sb.String(), nil
}

// generateMermaidFlat generates flat (ungrouped) Mermaid diagrams.
func generateMermaidFlat(rules []RegisteredRule, opts DocumentOptions, sb *strings.Builder) (string, error) {
	// Track processed rules to avoid duplicates
	processedRules := make(map[uintptr]bool)

	// Add all rule nodes
	for _, regRule := range rules {
		rulePtr := getRulePointer(regRule.Rule)
		if processedRules[rulePtr] {
			continue
		}
		processedRules[rulePtr] = true

		writeMermaidRule(sb, &regRule, opts, 1)
	}

	// Add connections
	sb.WriteString("\n    %% Rule connections\n")
	processedRules = make(map[uintptr]bool)
	for _, regRule := range rules {
		rulePtr := getRulePointer(regRule.Rule)
		if processedRules[rulePtr] {
			continue
		}
		processedRules[rulePtr] = true

		writeMermaidConnections(sb, &regRule, opts)
	}

	// Add styling
	writeMermaidStyling(sb)

	return sb.String(), nil
}

// writeMermaidRule writes a single rule node to the Mermaid diagram.
func writeMermaidRule(sb *strings.Builder, regRule *RegisteredRule, opts DocumentOptions, indentLevel int) {
	indent := strings.Repeat("    ", indentLevel)

	// Build the rule tree to get type information
	node := buildRuleTree(regRule.Rule, regRule, 0, opts.MaxDepth)

	// Generate node ID
	nodeID := getMermaidNodeID(regRule.Rule)

	// Generate node label
	label := node.Name
	if opts.IncludeMetadata {
		label = fmt.Sprintf("%s<br/>%s", label, node.Type.String())
	}

	// Choose node shape based on rule type
	nodeShape := getMermaidNodeShape(node.Type)

	// Write node
	sb.WriteString(fmt.Sprintf("%s%s%s\"%s\"%s\n",
		indent,
		nodeID,
		nodeShape.Open,
		escapeMermaidLabel(label),
		nodeShape.Close))
}

// writeMermaidConnections writes connections between parent and child rules.
func writeMermaidConnections(sb *strings.Builder, regRule *RegisteredRule, opts DocumentOptions) {
	// Build rule tree
	node := buildRuleTree(regRule.Rule, regRule, 0, opts.MaxDepth)

	if len(node.Children) == 0 {
		return
	}

	parentID := getMermaidNodeID(regRule.Rule)

	// Add edges to children
	for _, child := range node.Children {
		childID := getMermaidNodeID(child.Rule)

		// Determine arrow style based on parent type
		arrow := getConnectionArrow(node.Type)

		sb.WriteString(fmt.Sprintf("    %s %s %s\n", parentID, arrow, childID))
	}
}

// writeMermaidStyling adds CSS styling classes to the diagram.
func writeMermaidStyling(sb *strings.Builder) {
	sb.WriteString("\n    %% Styling\n")
	sb.WriteString("    classDef simpleRule fill:#3498db,stroke:#2980b9,color:#fff\n")
	sb.WriteString("    classDef andRule fill:#27ae60,stroke:#229954,color:#fff\n")
	sb.WriteString("    classDef orRule fill:#f39c12,stroke:#e67e22,color:#fff\n")
	sb.WriteString("    classDef notRule fill:#e74c3c,stroke:#c0392b,color:#fff\n")
}

// mermaidNodeShape represents the opening and closing characters for a Mermaid node shape.
type mermaidNodeShape struct {
	Open  string
	Close string
}

// getMermaidNodeShape returns the appropriate node shape for a rule type.
func getMermaidNodeShape(ruleType RuleType) mermaidNodeShape {
	switch ruleType {
	case RuleTypeSimple:
		return mermaidNodeShape{Open: "[", Close: "]"} // Rectangle
	case RuleTypeAnd:
		return mermaidNodeShape{Open: "[[", Close: "]]"} // Double rectangle
	case RuleTypeOr:
		return mermaidNodeShape{Open: "{", Close: "}"} // Rhombus
	case RuleTypeNot:
		return mermaidNodeShape{Open: "[(", Close: ")]"} // Stadium
	default:
		return mermaidNodeShape{Open: "[", Close: "]"} // Rectangle
	}
}

// getConnectionArrow returns the arrow style based on rule type.
func getConnectionArrow(parentType RuleType) string {
	switch parentType {
	case RuleTypeAnd:
		return "-->" // Solid arrow for AND
	case RuleTypeOr:
		return "-.->-" // Dotted arrow for OR
	case RuleTypeNot:
		return "==>=" // Thick arrow for NOT
	default:
		return "-->" // Default solid arrow
	}
}

// getMermaidNodeID generates a unique node ID for Mermaid.
func getMermaidNodeID(rule any) string {
	ptr := getRulePointer(rule)
	return fmt.Sprintf("R%X", ptr)
}

// sanitizeMermaidID removes special characters from IDs.
func sanitizeMermaidID(s string) string {
	// Replace spaces and special chars with underscores
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return s
}

// escapeMermaidLabel escapes special characters in Mermaid labels.
func escapeMermaidLabel(s string) string {
	// Escape quotes
	s = strings.ReplaceAll(s, "\"", "&quot;")
	// Note: <br/> is allowed for line breaks
	return s
}
