package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// FieldValidator handles field-level validation with rules
type FieldValidator struct {
	rules *ConfigValidationRules
}

// NewFieldValidator creates a new field validator
func NewFieldValidator(rules *ConfigValidationRules) *FieldValidator {
	return &FieldValidator{rules: rules}
}

// ValidateFieldConstraints validates individual fields against rules
func (fv *FieldValidator) ValidateFieldConstraints(cfg *domain.Config) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// MaxDiskUsage validation
	if err := fv.validateMaxDiskUsage(cfg.MaxDiskUsage); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "max_disk_usage",
			Rule:       "range",
			Value:      cfg.MaxDiskUsage,
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: "Set max disk usage between 10% and 95%",
		})
	}

	// Protected paths validation
	if err := fv.validateProtectedPaths(cfg.Protected); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "protected",
			Rule:       "required",
			Value:      len(cfg.Protected),
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: "Add system paths like /System, /Applications, /Library",
		})
	}

	return result
}

// ValidateField validates a specific configuration field
func (fv *FieldValidator) ValidateField(field string, value any) error {
	switch field {
	case "max_disk_usage":
		return fv.validateMaxDiskUsage(value)
	case "protected":
		return fv.validateProtectedPaths(value)
	case "profiles":
		return fv.validateProfiles(value)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
}

// validateMaxDiskUsage validates max disk usage constraint
func (fv *FieldValidator) validateMaxDiskUsage(value any) error {
	maxUsage, ok := value.(int)
	if !ok {
		return fmt.Errorf("max_disk_usage must be an integer")
	}

	if fv.rules.MaxDiskUsage != nil {
		if fv.rules.MaxDiskUsage.Min != nil && maxUsage < *fv.rules.MaxDiskUsage.Min {
			return fmt.Errorf("max_disk_usage must be at least %d", *fv.rules.MaxDiskUsage.Min)
		}
		if fv.rules.MaxDiskUsage.Max != nil && maxUsage > *fv.rules.MaxDiskUsage.Max {
			return fmt.Errorf("max_disk_usage must be at most %d", *fv.rules.MaxDiskUsage.Max)
		}
	}

	if maxUsage < 0 || maxUsage > 100 {
		return fmt.Errorf("max_disk_usage must be between 0 and 100")
	}

	return nil
}

// validateProtectedPaths validates protected paths constraint
func (fv *FieldValidator) validateProtectedPaths(value any) error {
	paths, ok := value.([]string)
	if !ok {
		return fmt.Errorf("protected must be a string array")
	}

	if fv.rules.MinProtectedPaths != nil && len(paths) < *fv.rules.MinProtectedPaths.Min {
		return fmt.Errorf("at least %d protected paths required", *fv.rules.MinProtectedPaths.Min)
	}

	return nil
}

// validateProfiles validates profiles constraint
func (fv *FieldValidator) validateProfiles(value any) error {
	_, ok := value.(map[string]*domain.Profile)
	if !ok {
		return fmt.Errorf("profiles must be a map of profiles")
	}

	return nil // More validation can be added here
}
