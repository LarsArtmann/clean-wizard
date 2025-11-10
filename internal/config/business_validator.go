package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BusinessValidator handles business logic validation
type BusinessValidator struct {
	rules *ConfigValidationRules
}

// NewBusinessValidator creates a new business validator
func NewBusinessValidator(rules *ConfigValidationRules) *BusinessValidator {
	return &BusinessValidator{rules: rules}
}

// ValidateBusinessLogic validates business logic constraints
func (bv *BusinessValidator) ValidateBusinessLogic(cfg *domain.Config) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	for name, profile := range cfg.Profiles {
		// Validate profile business logic
		if len(profile.Operations) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Rule:       "business_logic",
				Value:      len(profile.Operations),
				Message:    fmt.Sprintf("Profile '%s' must have at least one operation", name),
				Severity:   SeverityError,
				Suggestion: "Add at least one valid operation to profile",
			})
		}

		// Check operation count limits
		if bv.rules.MaxOperations != nil && len(profile.Operations) > *bv.rules.MaxOperations.Min {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Message:    fmt.Sprintf("Profile '%s' has many operations (%d), consider simplifying", name, len(profile.Operations)),
				Suggestion: "Split complex profiles into smaller, focused profiles",
			})
		}

		// Validate enabled profiles have valid operations
		if profile.Enabled {
			if err := bv.validateProfileOperations(name, profile.Operations); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:      fmt.Sprintf("profiles.%s.operations", name),
					Rule:       "business_logic",
					Value:      profile.Operations,
					Message:    err.Error(),
					Severity:   SeverityError,
					Suggestion: "Ensure all operations in enabled profile are valid",
				})
			}
		}
	}

	// Check profile count limits
	if bv.rules.MaxProfiles != nil && len(cfg.Profiles) > *bv.rules.MaxProfiles.Min {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    fmt.Sprintf("Profile count (%d) exceeds recommended limit (%d)", len(cfg.Profiles), *bv.rules.MaxProfiles.Min),
			Suggestion: "Consider consolidating profiles to improve maintainability",
		})
	}

	return result
}

// validateProfileOperations validates operations within a profile
func (bv *BusinessValidator) validateProfileOperations(profileName string, operations []domain.CleanupOperation) error {
	if len(operations) == 0 {
		return fmt.Errorf("profile '%s' cannot be enabled with no operations", profileName)
	}

	// Check for duplicate operations
	seen := make(map[string]bool)
	for _, op := range operations {
		if seen[op.Name] {
			return fmt.Errorf("duplicate operation '%s' in profile '%s'", op.Name, profileName)
		}
		seen[op.Name] = true

		// Validate operation has required settings
		if op.Settings == nil || !op.Settings.IsValid() {
			return fmt.Errorf("operation '%s' in profile '%s' has invalid settings", op.Name, profileName)
		}
	}

	return nil
}