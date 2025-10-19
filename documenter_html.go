package rules

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

// GenerateHTML generates HTML documentation for all registered rules.
func GenerateHTML(opts DocumentOptions) (string, error) {
	return GenerateHTMLFromRules(AllRules(), opts)
}

// GenerateHTMLFromRules generates HTML documentation from a list of registered rules.
func GenerateHTMLFromRules(rules []RegisteredRule, opts DocumentOptions) (string, error) {
	// Filter rules based on options
	filtered := filterRegisteredRules(rules, opts)

	// Collect domains and groups for sidebar
	domains := collectDomainsFromRegisteredRules(filtered)
	grouped := groupRulesByGroup(filtered)

	groupNames := make([]string, 0, len(grouped))
	for name := range grouped {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	// Build HTML
	var sb strings.Builder

	// Write HTML header
	writeHTMLHeader(&sb, opts)

	// Write sidebar
	writeHTMLSidebar(&sb, domains, groupNames)

	// Write main content area
	sb.WriteString(`    <main class="main-content">`)
	sb.WriteString("\n")

	// Write title and description
	writeHTMLTitleSection(&sb, opts)

	// Write rules
	if opts.GroupByDomain {
		writeHTMLGroupedRules(&sb, grouped, opts)
	} else {
		writeHTMLFlatRules(&sb, filtered, opts)
	}

	sb.WriteString(`    </main>`)
	sb.WriteString("\n")

	// Write footer
	writeHTMLFooter(&sb)

	return sb.String(), nil
}

// GenerateDomainHTML generates HTML documentation for a specific domain.
func GenerateDomainHTML(domain Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomain(domain)
	opts.Title = fmt.Sprintf("%s Domain Rules", domain)
	return GenerateHTMLFromRules(rules, opts)
}

// GenerateDomainsHTML generates HTML documentation for multiple domains.
func GenerateDomainsHTML(domains []Domain, opts DocumentOptions) (string, error) {
	rules := RulesByDomains(domains...)
	if opts.Title == "" {
		domainNames := make([]string, len(domains))
		for i, d := range domains {
			domainNames[i] = string(d)
		}
		opts.Title = fmt.Sprintf("Rules for Domains: %s", strings.Join(domainNames, ", "))
	}
	return GenerateHTMLFromRules(rules, opts)
}

// GenerateGroupHTML generates HTML documentation for a specific group.
func GenerateGroupHTML(groupName string, opts DocumentOptions) (string, error) {
	rules := RulesByGroup(groupName)
	opts.Title = fmt.Sprintf("%s Rules", groupName)
	return GenerateHTMLFromRules(rules, opts)
}

// writeHTMLHeader writes the HTML document header with embedded CSS and JavaScript.
func writeHTMLHeader(sb *strings.Builder, opts DocumentOptions) {
	title := opts.Title
	if title == "" {
		title = "Business Rules Documentation"
	}

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>`)
	sb.WriteString(html.EscapeString(title))
	sb.WriteString(`</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }

        .container {
            display: flex;
            min-height: 100vh;
        }

        .sidebar {
            width: 280px;
            background: #2c3e50;
            color: #ecf0f1;
            padding: 20px;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            box-shadow: 2px 0 5px rgba(0,0,0,0.1);
        }

        .sidebar h2 {
            font-size: 1.2rem;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #34495e;
        }

        .sidebar .search-box {
            width: 100%;
            padding: 10px;
            margin-bottom: 20px;
            border: none;
            border-radius: 4px;
            background: #34495e;
            color: #ecf0f1;
            font-size: 14px;
        }

        .sidebar .search-box:focus {
            outline: none;
            background: #3d5468;
        }

        .sidebar ul {
            list-style: none;
        }

        .sidebar li {
            margin-bottom: 8px;
        }

        .sidebar a {
            color: #3498db;
            text-decoration: none;
            transition: color 0.2s;
        }

        .sidebar a:hover {
            color: #5dade2;
        }

        .sidebar .section {
            margin-bottom: 25px;
        }

        .main-content {
            margin-left: 280px;
            flex: 1;
            padding: 40px;
            background: white;
            min-height: 100vh;
        }

        .header-section {
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 3px solid #3498db;
        }

        .header-section h1 {
            color: #2c3e50;
            font-size: 2.5rem;
            margin-bottom: 10px;
        }

        .header-section .description {
            color: #7f8c8d;
            font-size: 1.1rem;
            margin-top: 10px;
        }

        .header-section .meta {
            color: #95a5a6;
            font-size: 0.9rem;
            margin-top: 15px;
        }

        .rule-group {
            margin-bottom: 50px;
        }

        .rule-group h2 {
            color: #2c3e50;
            font-size: 2rem;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #ecf0f1;
        }

        .rule-card {
            background: #fff;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 25px;
            margin-bottom: 25px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            transition: box-shadow 0.2s;
        }

        .rule-card:hover {
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }

        .rule-card h3 {
            color: #2c3e50;
            font-size: 1.5rem;
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            cursor: pointer;
        }

        .rule-card h3 .toggle-icon {
            margin-right: 10px;
            font-size: 1rem;
            transition: transform 0.2s;
        }

        .rule-card h3.collapsed .toggle-icon {
            transform: rotate(-90deg);
        }

        .rule-card .type-badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: bold;
            margin-left: 10px;
            text-transform: uppercase;
        }

        .rule-card .type-simple { background: #3498db; color: white; }
        .rule-card .type-and { background: #27ae60; color: white; }
        .rule-card .type-or { background: #f39c12; color: white; }
        .rule-card .type-not { background: #e74c3c; color: white; }

        .rule-card .description {
            color: #555;
            margin-bottom: 15px;
            font-size: 1rem;
        }

        .rule-card .metadata {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }

        .rule-card .metadata-item {
            padding: 10px;
            background: #f8f9fa;
            border-radius: 4px;
        }

        .rule-card .metadata-item strong {
            color: #2c3e50;
            display: block;
            margin-bottom: 5px;
            font-size: 0.85rem;
        }

        .rule-card .metadata-item span {
            color: #555;
            font-size: 0.9rem;
        }

        .rule-card .tag {
            display: inline-block;
            padding: 3px 8px;
            background: #ecf0f1;
            color: #2c3e50;
            border-radius: 3px;
            font-size: 0.8rem;
            margin-right: 5px;
            margin-bottom: 5px;
        }

        .rule-card .domain-badge {
            display: inline-block;
            padding: 4px 10px;
            background: #9b59b6;
            color: white;
            border-radius: 4px;
            font-size: 0.8rem;
            margin-right: 5px;
            margin-bottom: 5px;
        }

        .rule-children {
            margin-top: 20px;
            padding-left: 20px;
            border-left: 3px solid #3498db;
        }

        .rule-children .child-rule {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 15px;
        }

        .rule-children .child-rule h4 {
            color: #2c3e50;
            font-size: 1.2rem;
            margin-bottom: 10px;
        }

        .rule-children .children-header {
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 1.1rem;
        }

        .collapsible-content {
            max-height: 10000px;
            overflow: hidden;
            transition: max-height 0.3s ease;
        }

        .collapsible-content.collapsed {
            max-height: 0;
        }

        @media (max-width: 768px) {
            .sidebar {
                width: 100%;
                position: relative;
                height: auto;
            }

            .main-content {
                margin-left: 0;
            }

            .container {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <div class="container">
`)
}

// writeHTMLSidebar writes the navigation sidebar.
func writeHTMLSidebar(sb *strings.Builder, domains []Domain, groups []string) {
	sb.WriteString(`        <nav class="sidebar">
            <h2>📚 Documentation</h2>
            <input type="text" class="search-box" id="searchBox" placeholder="Search rules...">
`)

	// Write domains section
	if len(domains) > 0 {
		sb.WriteString(`            <div class="section">
                <h2>Domains</h2>
                <ul>
`)
		for _, domain := range domains {
			anchor := strings.ToLower(strings.ReplaceAll(string(domain), " ", "-"))
			sb.WriteString(fmt.Sprintf(`                    <li><a href="#domain-%s">%s</a></li>`,
				html.EscapeString(anchor),
				html.EscapeString(string(domain))))
			sb.WriteString("\n")
		}
		sb.WriteString(`                </ul>
            </div>
`)
	}

	// Write groups section
	if len(groups) > 0 {
		sb.WriteString(`            <div class="section">
                <h2>Groups</h2>
                <ul>
`)
		for _, group := range groups {
			anchor := strings.ToLower(strings.ReplaceAll(group, " ", "-"))
			sb.WriteString(fmt.Sprintf(`                    <li><a href="#group-%s">%s</a></li>`,
				html.EscapeString(anchor),
				html.EscapeString(group)))
			sb.WriteString("\n")
		}
		sb.WriteString(`                </ul>
            </div>
`)
	}

	sb.WriteString(`        </nav>
`)
}

// writeHTMLTitleSection writes the title and description section.
func writeHTMLTitleSection(sb *strings.Builder, opts DocumentOptions) {
	title := opts.Title
	if title == "" {
		title = "Business Rules Documentation"
	}

	sb.WriteString(`            <div class="header-section">
                <h1>`)
	sb.WriteString(html.EscapeString(title))
	sb.WriteString(`</h1>
`)

	if opts.Description != "" {
		sb.WriteString(`                <div class="description">`)
		sb.WriteString(html.EscapeString(opts.Description))
		sb.WriteString(`</div>
`)
	}

	sb.WriteString(`                <div class="meta">Generated: `)
	sb.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	sb.WriteString(`</div>
            </div>
`)
}

// writeHTMLGroupedRules writes rules grouped by group name.
func writeHTMLGroupedRules(sb *strings.Builder, grouped map[string][]RegisteredRule, opts DocumentOptions) {
	// Sort group names
	groupNames := make([]string, 0, len(grouped))
	for name := range grouped {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		groupRules := grouped[groupName]
		anchor := strings.ToLower(strings.ReplaceAll(groupName, " ", "-"))

		sb.WriteString(fmt.Sprintf(`            <div class="rule-group" id="group-%s">
                <h2>%s</h2>
`, html.EscapeString(anchor), html.EscapeString(groupName)))

		// Show domains for this group
		domains := collectDomainsFromRegisteredRules(groupRules)
		if len(domains) > 0 {
			sb.WriteString(`                <div style="margin-bottom: 20px;">
`)
			for _, domain := range domains {
				sb.WriteString(fmt.Sprintf(`                    <span class="domain-badge">%s</span>`,
					html.EscapeString(string(domain))))
				sb.WriteString("\n")
			}
			sb.WriteString(`                </div>
`)
		}

		// Write rules in this group
		for _, regRule := range groupRules {
			writeHTMLRule(sb, regRule, opts)
		}

		sb.WriteString(`            </div>
`)
	}
}

// writeHTMLFlatRules writes rules in a flat structure.
func writeHTMLFlatRules(sb *strings.Builder, rules []RegisteredRule, opts DocumentOptions) {
	sb.WriteString(`            <div class="rule-group">
                <h2>Rules</h2>
`)

	for _, regRule := range rules {
		writeHTMLRule(sb, regRule, opts)
	}

	sb.WriteString(`            </div>
`)
}

// writeHTMLRule writes a single rule card.
func writeHTMLRule(sb *strings.Builder, regRule RegisteredRule, opts DocumentOptions) {
	// Build the rule tree
	node := buildRuleTree(regRule.Rule, &regRule, 0, opts.MaxDepth)

	ruleID := fmt.Sprintf("rule-%p", regRule.Rule)

	sb.WriteString(`                <div class="rule-card">
`)

	// Rule header
	sb.WriteString(fmt.Sprintf(`                    <h3 class="rule-header" onclick="toggleRule('%s')">
                        <span class="toggle-icon">▼</span>
                        <span>%s</span>
                        <span class="type-badge type-%s">%s</span>
                    </h3>
`,
		ruleID,
		html.EscapeString(node.Name),
		strings.ToLower(node.Type.String()),
		html.EscapeString(node.Type.String())))

	// Collapsible content
	sb.WriteString(fmt.Sprintf(`                    <div class="collapsible-content" id="%s">
`, ruleID))

	// Description
	if node.Description != "" {
		sb.WriteString(`                        <div class="description">`)
		sb.WriteString(html.EscapeString(node.Description))
		sb.WriteString(`</div>
`)
	}

	// Metadata grid
	sb.WriteString(`                        <div class="metadata">
`)

	// Domains
	if len(node.Domains) > 0 {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Domains</strong>
                                <div>
`)
		for _, domain := range node.Domains {
			domainAnchor := strings.ToLower(strings.ReplaceAll(string(domain), " ", "-"))
			sb.WriteString(fmt.Sprintf(`                                    <a href="#domain-%s" class="domain-badge">%s</a>`,
				html.EscapeString(domainAnchor),
				html.EscapeString(string(domain))))
			sb.WriteString("\n")
		}
		sb.WriteString(`                                </div>
                            </div>
`)
	}

	// Group
	if node.Group != "" {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Group</strong>
                                <span>`)
		sb.WriteString(html.EscapeString(node.Group))
		sb.WriteString(`</span>
                            </div>
`)
	}

	// Additional metadata if included
	if opts.IncludeMetadata && node.Metadata != nil {
		writeHTMLMetadata(sb, node.Metadata)
	}

	sb.WriteString(`                        </div>
`)

	// Children
	if len(node.Children) > 0 {
		writeHTMLChildren(sb, node, opts)
	}

	sb.WriteString(`                    </div>
                </div>
`)
}

// writeHTMLMetadata writes additional metadata fields.
func writeHTMLMetadata(sb *strings.Builder, metadata *RuleMetadata) {
	if metadata.RequirementID != "" {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Requirement ID</strong>
                                <span>`)
		sb.WriteString(html.EscapeString(metadata.RequirementID))
		sb.WriteString(`</span>
                            </div>
`)
	}

	if metadata.BusinessDescription != "" {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Business Requirement</strong>
                                <span>`)
		sb.WriteString(html.EscapeString(metadata.BusinessDescription))
		sb.WriteString(`</span>
                            </div>
`)
	}

	if metadata.Owner != "" {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Owner</strong>
                                <span>`)
		sb.WriteString(html.EscapeString(metadata.Owner))
		sb.WriteString(`</span>
                            </div>
`)
	}

	if metadata.Version != "" {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Version</strong>
                                <span>`)
		sb.WriteString(html.EscapeString(metadata.Version))
		sb.WriteString(`</span>
                            </div>
`)
	}

	if len(metadata.Tags) > 0 {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Tags</strong>
                                <div>
`)
		for _, tag := range metadata.Tags {
			sb.WriteString(fmt.Sprintf(`                                    <span class="tag">%s</span>`,
				html.EscapeString(tag)))
			sb.WriteString("\n")
		}
		sb.WriteString(`                                </div>
                            </div>
`)
	}

	if !metadata.CreatedAt.IsZero() {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Created</strong>
                                <span>`)
		sb.WriteString(metadata.CreatedAt.Format("2006-01-02"))
		sb.WriteString(`</span>
                            </div>
`)
	}

	if !metadata.UpdatedAt.IsZero() {
		sb.WriteString(`                            <div class="metadata-item">
                                <strong>Updated</strong>
                                <span>`)
		sb.WriteString(metadata.UpdatedAt.Format("2006-01-02"))
		sb.WriteString(`</span>
                            </div>
`)
	}
}

// writeHTMLChildren writes child rules.
func writeHTMLChildren(sb *strings.Builder, node *ruleNode, opts DocumentOptions) {
	sb.WriteString(`                        <div class="rule-children">
`)

	// Write header based on rule type
	switch node.Type {
	case RuleTypeAnd:
		sb.WriteString(`                            <div class="children-header">All of these conditions must be satisfied:</div>
`)
	case RuleTypeOr:
		sb.WriteString(`                            <div class="children-header">At least one of these conditions must be satisfied:</div>
`)
	case RuleTypeNot:
		sb.WriteString(`                            <div class="children-header">Negation of:</div>
`)
	default:
		sb.WriteString(`                            <div class="children-header">Child rules:</div>
`)
	}

	// Write each child
	for _, child := range node.Children {
		writeHTMLChildRule(sb, child, opts)
	}

	sb.WriteString(`                        </div>
`)
}

// writeHTMLChildRule writes a single child rule.
func writeHTMLChildRule(sb *strings.Builder, child *ruleNode, opts DocumentOptions) {
	sb.WriteString(`                            <div class="child-rule">
                                <h4>`)
	sb.WriteString(html.EscapeString(child.Name))
	sb.WriteString(fmt.Sprintf(` <span class="type-badge type-%s">%s</span></h4>
`,
		strings.ToLower(child.Type.String()),
		html.EscapeString(child.Type.String())))

	if child.Description != "" {
		sb.WriteString(`                                <div class="description">`)
		sb.WriteString(html.EscapeString(child.Description))
		sb.WriteString(`</div>
`)
	}

	// Show domains if present
	if len(child.Domains) > 0 {
		sb.WriteString(`                                <div style="margin-top: 10px;">
`)
		for _, domain := range child.Domains {
			sb.WriteString(fmt.Sprintf(`                                    <span class="domain-badge">%s</span>`,
				html.EscapeString(string(domain))))
			sb.WriteString("\n")
		}
		sb.WriteString(`                                </div>
`)
	}

	// Recursively write children
	if len(child.Children) > 0 {
		writeHTMLChildren(sb, child, opts)
	}

	sb.WriteString(`                            </div>
`)
}

// writeHTMLFooter writes the HTML footer with JavaScript.
func writeHTMLFooter(sb *strings.Builder) {
	sb.WriteString(`    </div>

    <script>
        // Toggle rule collapsible content
        function toggleRule(ruleId) {
            const content = document.getElementById(ruleId);
            const header = document.querySelector('[onclick*="' + ruleId + '"]');
            
            if (content.classList.contains('collapsed')) {
                content.classList.remove('collapsed');
                header.classList.remove('collapsed');
            } else {
                content.classList.add('collapsed');
                header.classList.add('collapsed');
            }
        }

        // Search functionality
        document.getElementById('searchBox').addEventListener('input', function(e) {
            const searchTerm = e.target.value.toLowerCase();
            const ruleCards = document.querySelectorAll('.rule-card');
            
            ruleCards.forEach(function(card) {
                const text = card.textContent.toLowerCase();
                if (text.includes(searchTerm)) {
                    card.style.display = 'block';
                } else {
                    card.style.display = 'none';
                }
            });
        });

        // Smooth scrolling for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    </script>
</body>
</html>
`)
}
