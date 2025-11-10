package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BasicValidator handles basic configuration structure validation
type BasicValidator struct {
	rules *ConfigValidationRules
}

// NewBasicValidator creates a new basic validator
func NewBasicValidator(rules *ConfigValidationRules) *BasicValidator {
	return &BasicValidator{rules: rules}
}

// ValidateStructure validates basic configuration structure
func (bv *BasicValidator) ValidateStructure(cfg *domain.Config) *ValidationResult {
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: make(map[string]any),
		Timestamp: time.Now(),
	}

	// Version validation
	if cfg.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "version",
			Rule:       "required",
			Value:      cfg.Version,
			Message:    "Configuration version is required",
			Severity:   SeverityError,
			Suggestion: "Set version to \"1.0.0\" or higher",
		})
	}

	// Profiles validation
	if len(cfg.Profiles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "profiles",
			Rule:       "required",
			Value:      cfg.Profiles,
			Message:    "At least one profile is required",
			Severity:   SeverityError,
			Suggestion: "Add a profile with at least one operation",
		})
	}

	// Protected paths validation
	if len(cfg.Protected) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "protected",
			Rule:       "required",
			Value:      cfg.Protected,
			Message:    "Protected paths cannot be empty",
			Severity:   SeverityError,
			Suggestion: "Add system paths like /System, /Applications, /Library",
		})
	}

	return result
}
