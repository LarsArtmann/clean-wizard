package config

import (
	"fmt"
	"time"

	businessrules "github.com/LarsArtmann/go-business-rules/v2"

	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// ConfigValidator provides comprehensive type-safe configuration validation.
type ConfigValidator struct {
	rules     *ConfigValidationRules
	sanitizer *ConfigSanitizer
}

// ValidationResult contains validation results with detailed error information.
type ValidationResult struct {
	IsValid   bool                     `json:"is_valid"`
	Errors    []ValidationError        `json:"errors,omitempty"`
	Warnings  []ValidationWarning      `json:"warnings,omitempty"`
	Sanitized *ValidationSanitizedData `json:"sanitized,omitempty"`
	Duration  time.Duration            `json:"duration"`
	Timestamp time.Time                `json:"timestamp"`
}

// ValidationSanitizedData provides type-safe configuration data
// FIXED: Removed map[string]any to improve type safety.
type ValidationSanitizedData struct {
	FieldsModified []string          `json:"fields_modified,omitempty"`
	RulesApplied   []string          `json:"rules_applied,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	// Type-safe configuration instead of dynamic data
	ConfigVersion   string   `json:"config_version,omitempty"`
	ValidationLevel string   `json:"validation_level,omitempty"`
	AppliedProfiles []string `json:"applied_profiles,omitempty"`
}

// ValidationContext provides strongly-typed validation context information.
type ValidationContext = operations.ValidationContext

// ValidationError represents a specific validation error.
type ValidationError = operations.ValidationError

// ValidationWarning represents a non-critical validation issue.
type ValidationWarning struct {
	Field      string             `json:"field"`
	Message    string             `json:"message"`
	Suggestion string             `json:"suggestion,omitempty"`
	Context    *ValidationContext `json:"context,omitempty"`
}

// NewConfigValidator creates a comprehensive configuration validator.
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules:     getDefaultValidationRules(),
		sanitizer: NewConfigSanitizer(),
	}
}

// NewConfigValidatorWithRules creates a validator with custom rules.
func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator {
	return &ConfigValidator{
		rules:     rules,
		sanitizer: NewConfigSanitizer(),
	}
}

// ValidateConfig performs comprehensive configuration validation. Every
// check is expressed as a businessrules rule (severity-aware, tagged by
// validation level) and the outcome is bridged back into ValidationResult.
func (cv *ConfigValidator) ValidateConfig(cfg *types.Config) *ValidationResult {
	start := time.Now()

	ruleSet := cv.buildConfigRules(cfg)
	outcome := businessrules.NewValidator().AddRules(ruleSet.rules...).Build()

	result := mapViolations(ruleSet, outcome)
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()

	return result
}

// ValidateField validates a specific configuration field.
func (cv *ConfigValidator) ValidateField(field string, value any) error {
	switch field {
	case "max_disk_usage":
		return cv.validateMaxDiskUsage(value)
	case "protected": //nolint:goconst
		return cv.validateProtectedPaths(value)
	case "profiles": //nolint:goconst
		if cfg, ok := value.(*types.Config); ok {
			return cv.validateProfiles(cfg)
		}

		return fmt.Errorf("profiles validation requires *types.Config, got %T", value)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
}
