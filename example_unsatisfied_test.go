package rules_test

import (
	"fmt"

	"github.com/tobbstr/rules"
)

func ExampleResult_UnsatisfiedRules() {
	// Define validation rules
	type FormData struct {
		Email   string
		Age     int
		Country string
		Agreed  bool
	}

	hasEmail := rules.New("has email", func(data FormData) (bool, error) {
		return data.Email != "", nil
	})

	validAge := rules.New("valid age", func(data FormData) (bool, error) {
		return data.Age >= 18, nil
	})

	validCountry := rules.New("valid country", func(data FormData) (bool, error) {
		return data.Country == "US" || data.Country == "CA", nil
	})

	agreedToTerms := rules.New("agreed to terms", func(data FormData) (bool, error) {
		return data.Agreed, nil
	})

	// Combine into hierarchical rules
	basicChecks := rules.And("basic checks", hasEmail, validAge)
	complianceChecks := rules.And("compliance checks", validCountry, agreedToTerms)
	formValidation := rules.And("form validation", basicChecks, complianceChecks)

	// Evaluate invalid form data
	invalidData := FormData{
		Email:   "",    // Missing email
		Age:     16,    // Too young
		Country: "UK",  // Invalid country
		Agreed:  false, // Not agreed
	}

	evaluator := rules.NewEvaluator(formValidation)
	result := evaluator.EvaluateDetailed(invalidData)

	// Get all unsatisfied rule names
	unsatisfied := result.UnsatisfiedRules()

	fmt.Println("Validation failed!")
	fmt.Println("Unsatisfied rules:")
	for _, ruleName := range unsatisfied {
		fmt.Printf("  - %s\n", ruleName)
	}

	// Output:
	// Validation failed!
	// Unsatisfied rules:
	//   - form validation
	//   - basic checks
	//   - has email
	//   - valid age
	//   - compliance checks
	//   - valid country
	//   - agreed to terms
}
