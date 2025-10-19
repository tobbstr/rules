package rules

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// GenerateMarkdown generates Markdown documentation for all registered rules.
func GenerateMarkdown(opts DocumentOptions) (string, error) {
	return GenerateMarkdownFromRules(AllRules(), opts)
}

// GenerateMarkdownFromRules generates Markdown documentation from a list of registered rules.
func GenerateMarkdownFromRules(rules []RegisteredRule, opts DocumentOptions) (string, error) {
	var sb strings.Builder

	// Write header
	if opts.Title != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", opts.Title))
	} else {
		sb.WriteString("# Business Rules Documentation\n\n")
	}

	// Write description
	if opts.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", opts.Description))
	}

	// Write generation timestamp
	sb.WriteString(fmt.Sprintf("*Generated: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Filter rules based on domain filters
	filtered := filterRegisteredRules(rules, opts)

	if len(filtered) == 0 {
		sb.WriteString("*No rules match the specified filters.*\n")
		return sb.String(), nil
	}

	// Group by domain or generate flat structure
	if opts.GroupByDomain {
		return generateMarkdownByDomain(filtered, opts, sb.String())
	}

	return generateMarkdownFlat(filtered, opts, sb.String())
}

// GenerateDomainMarkdown generates Markdown documentation for a specific domain.
func GenerateDomainMarkdown(domain Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomain(domain)
	opts.Title = fmt.Sprintf("%s Domain Rules", domain)
	return GenerateMarkdownFromRules(rules, opts)
}

// GenerateDomainsMarkdown generates Markdown documentation for multiple domains.
func GenerateDomainsMarkdown(domains []Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomains(domains...)
	if opts.Title == "" {
		domainNames := make([]string, len(domains))
		for i, d := range domains {
			domainNames[i] = string(d)
		}
		opts.Title = fmt.Sprintf("Rules for Domains: %s", strings.Join(domainNames, ", "))
	}
	return GenerateMarkdownFromRules(rules, opts)
}

// GenerateGroupMarkdown generates Markdown documentation for a specific group.
func GenerateGroupMarkdown(groupName string, opts DocumentOptions) (string, error) {
	rules := RulesByGroup(groupName)
	opts.Title = fmt.Sprintf("%s Rules", groupName)
	return GenerateMarkdownFromRules(rules, opts)
}

// filterRegisteredRules filters rules based on document options.
func filterRegisteredRules(rules []RegisteredRule, opts DocumentOptions) []RegisteredRule {
	var filtered []RegisteredRule

	for _, rule := range rules {
		// Check exclude list
		excluded := false
		for _, exclude := range opts.ExcludeDomains {
			for _, domain := range rule.Domains {
				if domain == exclude {
					excluded = true
					break
				}
			}
			if excluded {
				break
			}
		}
		if excluded {
			continue
		}

		// Check include list
		if len(opts.IncludeDomains) > 0 {
			included := false
			for _, include := range opts.IncludeDomains {
				for _, domain := range rule.Domains {
					if domain == include {
						included = true
						break
					}
				}
				if included {
					break
				}
			}
			if !included {
				continue
			}
		}

		filtered = append(filtered, rule)
	}

	return filtered
}

// generateMarkdownByDomain generates domain-grouped Markdown documentation.
func generateMarkdownByDomain(rules []RegisteredRule, opts DocumentOptions, header string) (string, error) {
	var sb strings.Builder
	sb.WriteString(header)

	// Group rules by domain and by group
	grouped := groupRulesByGroup(rules)

	// Sort group names for consistent output
	groupNames := make([]string, 0, len(grouped))
	for name := range grouped {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Generate table of contents
	sb.WriteString("## Table of Contents\n\n")
	for _, groupName := range groupNames {
		anchor := strings.ToLower(strings.ReplaceAll(groupName, " ", "-"))
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", groupName, anchor))
	}
	sb.WriteString("\n---\n\n")

	// Generate sections for each group
	for _, groupName := range groupNames {
		groupRules := grouped[groupName]
		sb.WriteString(fmt.Sprintf("## %s\n\n", groupName))

		// Show domains for this group if cross-domain
		domains := collectDomainsFromRegisteredRules(groupRules)
		if len(domains) > 1 {
			domainStrs := make([]string, len(domains))
			for i, d := range domains {
				domainStrs[i] = string(d)
			}
			sb.WriteString(fmt.Sprintf("**Domains**: %s\n\n", strings.Join(domainStrs, ", ")))
		} else if len(domains) == 1 {
			sb.WriteString(fmt.Sprintf("**Domain**: %s\n\n", domains[0]))
		}

		// Document each rule in the group
		for i, regRule := range groupRules {
			if i > 0 {
				sb.WriteString("\n---\n\n")
			}
			generateRuleMarkdown(&sb, regRule, opts, 3)
		}

		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// generateMarkdownFlat generates flat (ungrouped) Markdown documentation.
func generateMarkdownFlat(rules []RegisteredRule, opts DocumentOptions, header string) (string, error) {
	var sb strings.Builder
	sb.WriteString(header)

	sb.WriteString("## Rules\n\n")

	for i, regRule := range rules {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		generateRuleMarkdown(&sb, regRule, opts, 3)
	}

	return sb.String(), nil
}

// generateRuleMarkdown generates Markdown documentation for a single rule.
func generateRuleMarkdown(sb *strings.Builder, regRule RegisteredRule, opts DocumentOptions, headerLevel int) {
	// Build the rule tree
	node := buildRuleTree(regRule.Rule, &regRule, 0, opts.MaxDepth)

	// Write rule header
	headerPrefix := strings.Repeat("#", headerLevel)
	sb.WriteString(fmt.Sprintf("%s %s (%s)\n\n", headerPrefix, node.Name, node.Type))

	// Write description
	if node.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", node.Description))
	}

	// Write basic info
	sb.WriteString(fmt.Sprintf("**Type**: %s\n\n", node.Type))

	// Write domains
	if len(node.Domains) > 0 {
		if len(node.Domains) == 1 {
			sb.WriteString(fmt.Sprintf("**Domain**: %s\n\n", node.Domains[0]))
		} else {
			domainStrs := make([]string, len(node.Domains))
			for i, d := range node.Domains {
				domainStrs[i] = string(d)
			}
			sb.WriteString(fmt.Sprintf("**Domains**: %s\n\n", strings.Join(domainStrs, ", ")))
		}
	}

	// Write group
	if node.Group != "" {
		sb.WriteString(fmt.Sprintf("**Group**: %s\n\n", node.Group))
	}

	// Write metadata if requested
	if opts.IncludeMetadata && node.Metadata != nil {
		writeMetadata(sb, node.Metadata)
	}

	// Write children
	if len(node.Children) > 0 {
		writeChildren(sb, node, opts, headerLevel+1)
	}
}

// writeMetadata writes rule metadata to the string builder.
func writeMetadata(sb *strings.Builder, metadata *RuleMetadata) {
	if metadata.RequirementID != "" {
		sb.WriteString(fmt.Sprintf("**Requirement ID**: %s\n\n", metadata.RequirementID))
	}
	if metadata.BusinessDescription != "" {
		sb.WriteString(fmt.Sprintf("**Business Requirement**: %s\n\n", metadata.BusinessDescription))
	}
	if metadata.Owner != "" {
		sb.WriteString(fmt.Sprintf("**Owner**: %s\n\n", metadata.Owner))
	}
	if metadata.Version != "" {
		sb.WriteString(fmt.Sprintf("**Version**: %s\n\n", metadata.Version))
	}
	if len(metadata.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags**: %s\n\n", strings.Join(metadata.Tags, ", ")))
	}
	if !metadata.CreatedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("**Created**: %s\n\n", metadata.CreatedAt.Format("2006-01-02")))
	}
	if !metadata.UpdatedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("**Updated**: %s\n\n", metadata.UpdatedAt.Format("2006-01-02")))
	}
	if len(metadata.Dependencies) > 0 {
		deps := make([]string, len(metadata.Dependencies))
		for i, d := range metadata.Dependencies {
			deps[i] = string(d)
		}
		sb.WriteString(fmt.Sprintf("**Dependencies**: %s\n\n", strings.Join(deps, ", ")))
	}
	if len(metadata.RelatedRules) > 0 {
		sb.WriteString(fmt.Sprintf("**Related Rules**: %s\n\n", strings.Join(metadata.RelatedRules, ", ")))
	}
}

// writeChildren writes child rules to the string builder.
func writeChildren(sb *strings.Builder, node *ruleNode, opts DocumentOptions, headerLevel int) {
	// Write appropriate header based on rule type
	switch node.Type {
	case RuleTypeAnd:
		sb.WriteString("**All of these conditions must be satisfied:**\n\n")
	case RuleTypeOr:
		sb.WriteString("**At least one of these conditions must be satisfied:**\n\n")
	case RuleTypeNot:
		sb.WriteString("**Negation of:**\n\n")
	default:
		sb.WriteString("**Child rules:**\n\n")
	}

	// Write each child
	for _, child := range node.Children {
		writeChildRule(sb, child, opts, headerLevel)
	}
}

// writeChildRule writes a single child rule.
func writeChildRule(sb *strings.Builder, child *ruleNode, opts DocumentOptions, headerLevel int) {
	headerPrefix := strings.Repeat("#", headerLevel)
	sb.WriteString(fmt.Sprintf("%s %s (%s)\n\n", headerPrefix, child.Name, child.Type))

	if child.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", child.Description))
	}

	// Show domains for child if different from parent
	if len(child.Domains) > 0 {
		if len(child.Domains) == 1 {
			sb.WriteString(fmt.Sprintf("**Domain**: %s\n\n", child.Domains[0]))
		} else {
			domainStrs := make([]string, len(child.Domains))
			for i, d := range child.Domains {
				domainStrs[i] = string(d)
			}
			sb.WriteString(fmt.Sprintf("**Domains**: %s\n\n", strings.Join(domainStrs, ", ")))
		}
	}

	// Recursively write children
	if len(child.Children) > 0 {
		writeChildren(sb, child, opts, headerLevel+1)
	}
}

// collectDomainsFromRegisteredRules collects unique domains from a list of registered rules.
func collectDomainsFromRegisteredRules(rules []RegisteredRule) []Domain {
	domainSet := make(map[Domain]bool)
	for _, rule := range rules {
		for _, domain := range rule.Domains {
			domainSet[domain] = true
		}
	}

	domains := make([]Domain, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	// Sort for consistent output
	sort.Slice(domains, func(i, j int) bool {
		return string(domains[i]) < string(domains[j])
	})

	return domains
}
