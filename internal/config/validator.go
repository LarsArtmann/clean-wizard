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

// ValidationError represents a specific validation error
type ValidationError struct {
	Field      string             `json:"field"`
	Rule       string             `json:"rule"`
	Value      any                `json:"value"`
	Message    string             `json:"message"`
	Severity   ValidationSeverity `json:"severity"`
	Suggestion string             `json:"suggestion,omitempty"`
	Context    map[string]any     `json:"context,omitempty"`
}

// ValidationWarning represents a non-critical validation issue
type ValidationWarning struct {
	Field      string         `json:"field"`
	Message    string         `json:"message"`
	Suggestion string         `json:"suggestion,omitempty"`
	Context    map[string]any `json:"context,omitempty"`
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

	// Level 2: Field-level validation with rules
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
		return cv.validateMaxDiskUsage(value)
	case "protected":
		return cv.validateProtectedPaths(value)
	case "profiles":
		if cfg, ok := value.(*domain.Config); ok {
			return cv.validateProfiles(cfg)
		}
		return fmt.Errorf("profiles validation requires *domain.Config, got %T", value)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
}
