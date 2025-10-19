package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tobbstr/rules"
)

const (
	defaultOutputDir = "docs"
	defaultTitle     = "Business Rules Documentation"
)

type config struct {
	outputDir       string
	title           string
	description     string
	groupByDomain   bool
	includeMetadata bool
	formats         []string
}

func main() {
	cfg := parseFlags()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Documentation generated successfully in %s/\n", cfg.outputDir)
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.outputDir, "output", defaultOutputDir,
		"Output directory for generated documentation")
	flag.StringVar(&cfg.title, "title", defaultTitle,
		"Documentation title")
	flag.StringVar(&cfg.description, "description", "",
		"Documentation description")
	flag.BoolVar(&cfg.groupByDomain, "group-by-domain", true,
		"Group rules by domain in documentation")
	flag.BoolVar(&cfg.includeMetadata, "include-metadata", true,
		"Include metadata (owner, version, etc.) in documentation")

	var formatsFlag string
	flag.StringVar(&formatsFlag, "formats", "markdown,html,json,mermaid",
		"Comma-separated list of formats to generate (markdown,html,json,mermaid)")

	flag.Parse()

	// Parse formats (simple split by comma)
	if formatsFlag == "all" {
		cfg.formats = []string{"markdown", "html", "json", "mermaid"}
	} else {
		// Simple parsing - in production might want something more robust
		formats := []string{}
		current := ""
		for _, c := range formatsFlag {
			if c == ',' {
				if current != "" {
					formats = append(formats, current)
					current = ""
				}
			} else {
				current += string(c)
			}
		}
		if current != "" {
			formats = append(formats, current)
		}
		cfg.formats = formats
	}

	return cfg
}

func run(cfg *config) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Get all registered rules
	allRules := rules.AllRules()
	if len(allRules) == 0 {
		fmt.Println("⚠ No rules registered - documentation will be empty")
		fmt.Println("  Make sure rules are registered before generating documentation")
	}

	// Document options
	opts := rules.DocumentOptions{
		Title:           cfg.title,
		Description:     cfg.description,
		GroupByDomain:   cfg.groupByDomain,
		IncludeMetadata: cfg.includeMetadata,
	}

	// Generate each requested format
	for _, format := range cfg.formats {
		if err := generateFormat(cfg, opts, format); err != nil {
			return fmt.Errorf("generating %s: %w", format, err)
		}
		fmt.Printf("  ✓ Generated %s\n", format)
	}

	return nil
}

func generateFormat(cfg *config, opts rules.DocumentOptions, format string) error {
	switch format {
	case "markdown", "md":
		return generateMarkdown(cfg, opts)
	case "html":
		return generateHTML(cfg, opts)
	case "json":
		return generateJSON(cfg, opts)
	case "mermaid", "mmd":
		return generateMermaid(cfg, opts)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func generateMarkdown(cfg *config, opts rules.DocumentOptions) error {
	content, err := rules.GenerateMarkdown(opts)
	if err != nil {
		return fmt.Errorf("generating markdown: %w", err)
	}

	filename := filepath.Join(cfg.outputDir, "rules.md")
	if err := writeFile(filename, content); err != nil {
		return fmt.Errorf("writing markdown: %w", err)
	}

	return nil
}

func generateHTML(cfg *config, opts rules.DocumentOptions) error {
	content, err := rules.GenerateHTML(opts)
	if err != nil {
		return fmt.Errorf("generating HTML: %w", err)
	}

	filename := filepath.Join(cfg.outputDir, "rules.html")
	if err := writeFile(filename, content); err != nil {
		return fmt.Errorf("writing HTML: %w", err)
	}

	return nil
}

func generateJSON(cfg *config, opts rules.DocumentOptions) error {
	content, err := rules.GenerateJSON(opts)
	if err != nil {
		return fmt.Errorf("generating JSON: %w", err)
	}

	filename := filepath.Join(cfg.outputDir, "rules.json")
	if err := writeFile(filename, content); err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}

	return nil
}

func generateMermaid(cfg *config, opts rules.DocumentOptions) error {
	content, err := rules.GenerateMermaid(opts)
	if err != nil {
		return fmt.Errorf("generating Mermaid: %w", err)
	}

	filename := filepath.Join(cfg.outputDir, "rules.mmd")
	if err := writeFile(filename, content); err != nil {
		return fmt.Errorf("writing Mermaid: %w", err)
	}

	return nil
}

func writeFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return nil
}
