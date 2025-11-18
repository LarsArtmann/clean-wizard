package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateBasicStructure validates basic configuration structure
func (cv *ConfigValidator) validateBasicStructure(cfg *domain.Config, result *ValidationResult) {
	// Version validation
	if cfg.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "version",
			Rule:       "required",
			Value:      cfg.Version,
			Message:    "Configuration version is required",
			Severity:   SeverityError,
			Suggestion: "Add version field with semantic version (e.g., '1.0.0')",
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
}