# Documentation Generator

Command-line tool to generate documentation for your business rules.

## Usage

### Basic Usage

Generate all formats with default settings:

```bash
go run ./cmd/gendocs/main.go
```

This creates:
- `docs/rules.md` - Markdown documentation
- `docs/rules.html` - Interactive HTML documentation
- `docs/rules.json` - Machine-readable JSON
- `docs/rules.mmd` - Mermaid diagram

### Custom Output Directory

```bash
go run ./cmd/gendocs/main.go -output ./documentation
```

### Specific Formats

Generate only specific formats:

```bash
# Only Markdown and HTML
go run ./cmd/gendocs/main.go -formats markdown,html

# Only JSON
go run ./cmd/gendocs/main.go -formats json
```

### Custom Title and Description

```bash
go run ./cmd/gendocs/main.go \
  -title "Order Processing Rules" \
  -description "Business rules for our e-commerce order processing system"
```

### Without Metadata

Exclude metadata (owner, version, etc.):

```bash
go run ./cmd/gendocs/main.go -include-metadata=false
```

### Flat Structure (No Domain Grouping)

```bash
go run ./cmd/gendocs/main.go -group-by-domain=false
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-output` | `docs` | Output directory for generated documentation |
| `-title` | `"Business Rules Documentation"` | Documentation title |
| `-description` | `""` | Documentation description |
| `-formats` | `markdown,html,json,mermaid` | Formats to generate |
| `-group-by-domain` | `true` | Group rules by domain |
| `-include-metadata` | `true` | Include metadata in documentation |

## Integration with Your Project

### Option 1: Import and Register Rules

Create a file that imports your rules to trigger registration:

```go
// cmd/gendocs/rules.go
package main

import (
    _ "your-project/internal/orders/rules"
    _ "your-project/internal/shipping/rules"
    _ "your-project/internal/users/rules"
)
```

### Option 2: Separate Command in Your Project

Create your own documentation generator that imports your rules:

```go
// cmd/generate-docs/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    
    "github.com/tobbstr/rules"
    
    // Import your rule packages
    _ "your-project/internal/orders/rules"
    _ "your-project/internal/shipping/rules"
)

func main() {
    output := flag.String("output", "docs", "Output directory")
    flag.Parse()
    
    // Generate documentation
    md, err := rules.GenerateMarkdown(rules.DocumentOptions{
        Title:           "Our Business Rules",
        GroupByDomain:   true,
        IncludeMetadata: true,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Write to file
    if err := os.WriteFile(*output+"/rules.md", []byte(md), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✓ Documentation generated")
}
```

## Example Output

When run, you'll see:

```
⚠ No rules registered - documentation will be empty
  Make sure rules are registered before generating documentation
  ✓ Generated markdown
  ✓ Generated html
  ✓ Generated json
  ✓ Generated mermaid
✓ Documentation generated successfully in docs/
```

## Pre-commit Hook Integration

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Generating documentation..."
go run ./cmd/gendocs/main.go || exit 1
git add docs/
echo "✓ Documentation updated"
```

## Makefile Integration

```makefile
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@go run ./cmd/gendocs/main.go
	@echo "✓ Documentation generated"

.PHONY: docs-watch
docs-watch:
	@while true; do \
		echo "Watching for changes..."; \
		fswatch -1 *.go internal/**/*.go; \
		make docs; \
	done
```

## CI/CD Validation

Check if documentation is up-to-date in CI:

```yaml
# .github/workflows/ci.yml
- name: Check documentation is current
  run: |
    go run ./cmd/gendocs/main.go
    if ! git diff --exit-code docs/; then
      echo "❌ Documentation is out of date"
      echo "Run 'make docs' and commit the changes"
      exit 1
    fi
    echo "✓ Documentation is up-to-date"
```

