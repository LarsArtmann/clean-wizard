package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidator provides comprehensive type-safe configuration validation
type ConfigValidator struct {
	rules     *ConfigValidationRules
	sanitizer *ConfigSanitizer
}

// ValidationResult contains validation results with detailed error information
type ValidationResult struct {
	IsValid   bool                     `json:"is_valid"`
	Errors    []ValidationError        `json:"errors,omitempty"`
	Warnings  []ValidationWarning      `json:"warnings,omitempty"`
	Sanitized *ValidationSanitizedData `json:"sanitized,omitempty"`
	Duration  time.Duration            `json:"duration"`
	Timestamp time.Time                `json:"timestamp"`
}

// ValidationSanitizedData provides type-safe configuration data
// FIXED: Removed map[string]any to improve type safety
type ValidationSanitizedData struct {
	FieldsModified []string          `json:"fields_modified,omitempty"`
	RulesApplied   []string          `json:"rules_applied,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	// Type-safe configuration instead of dynamic data
	ConfigVersion   string   `json:"config_version,omitempty"`
	ValidationLevel string   `json:"validation_level,omitempty"`
	AppliedProfiles []string `json:"applied_profiles,omitempty"`
}

// ValidationContext provides strongly-typed validation context information
type ValidationContext struct {
	ConfigPath      string            `json:"config_path,omitempty"`
	ValidationLevel string            `json:"validation_level,omitempty"`
	Profile         string            `json:"profile,omitempty"`
	Section         string            `json:"section,omitempty"`
	MinValue        any               `json:"min_value,omitempty"`
	MaxValue        any               `json:"max_value,omitempty"`
	AllowedValues   []string          `json:"allowed_values,omitempty"`
	ReferencedField string            `json:"referenced_field,omitempty"`
	Constraints     map[string]string `json:"constraints,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Field      string             `json:"field"`
	Rule       string             `json:"rule"`
	Value      any                `json:"value"`
	Message    string             `json:"message"`
	Severity   ValidationSeverity `json:"severity"`
	Suggestion string             `json:"suggestion,omitempty"`
	Context    *ValidationContext `json:"context,omitempty"`
}

// ValidationWarning represents a non-critical validation issue
type ValidationWarning struct {
	Field      string             `json:"field"`
	Message    string             `json:"message"`
	Suggestion string             `json:"suggestion,omitempty"`
	Context    *ValidationContext `json:"context,omitempty"`
}

// NewConfigValidator creates a comprehensive configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules:     getDefaultValidationRules(),
		sanitizer: NewConfigSanitizer(),
	}
}

// NewConfigValidatorWithRules creates a validator with custom rules
func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator {
	return &ConfigValidator{
		rules:     rules,
		sanitizer: NewConfigSanitizer(),
	}
}

// ValidateConfig performs comprehensive configuration validation
func (cv *ConfigValidator) ValidateConfig(cfg *domain.Config) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: nil, // Initialize as nil, will be set by sanitizer if needed
		Timestamp: time.Now(),
	}

	// Level 1: Basic structure validation
	cv.validateBasicStructure(cfg, result)

	// Level 2: Field-level validation
	cv.validateFieldConstraints(cfg, result)

	// Level 3: Cross-field validation
	cv.validateCrossFieldConstraints(cfg, result)

	// Level 4: Business logic validation
	cv.validateBusinessLogic(cfg, result)

	// Level 5: Security validation
	cv.validateSecurityConstraints(cfg, result)

	// NOTE: Sanitization is NOT applied here to preserve original values
	// Sanitization should be applied separately after validation succeeds
	// This prevents state mutation during verification

	result.Duration = time.Since(start)
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateField validates a specific configuration field
func (cv *ConfigValidator) ValidateField(field string, value any) error {
	switch field {
	case "max_disk_usage":
		if intVal, ok := value.(int); ok {
			rules := cv.rules.MaxDiskUsage
			if rules != nil {
				if rules.Min != nil && intVal < *rules.Min {
					return fmt.Errorf("max_disk_usage must be at least %d, got %d", *rules.Min, intVal)
				}
				if rules.Max != nil && intVal > *rules.Max {
					return fmt.Errorf("max_disk_usage must be at most %d, got %d", *rules.Max, intVal)
				}
			}
		} else {
			return fmt.Errorf("max_disk_usage must be an integer, got %T", value)
		}
	case "protected":
		if sliceVal, ok := value.([]string); ok {
			rules := cv.rules.MinProtectedPaths
			if rules != nil && rules.Min != nil && len(sliceVal) < *rules.Min {
				return fmt.Errorf("protected paths must have at least %d items, got %d", *rules.Min, len(sliceVal))
			}
		} else {
			return fmt.Errorf("protected must be a slice of strings, got %T", value)
		}
	case "profiles":
		// Complex validation handled by validateBusinessLogic
		return nil
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
	return nil
}
