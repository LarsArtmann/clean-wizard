package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigSanitizer provides configuration sanitization and normalization.
type ConfigSanitizer struct {
	rules *SanitizationRules
}

// SanitizationRules defines how to sanitize different configuration aspects.
type SanitizationRules struct {
	// Path sanitization
	NormalizePaths bool `json:"normalize_paths"`
	ExpandHomeDir  bool `json:"expand_home_dir"`
	ValidateExists bool `json:"validate_exists"`

	// Numeric sanitization
	ClampValues      bool `json:"clamp_values"`
	RoundPercentages bool `json:"round_percentages"`

	// String sanitization
	TrimWhitespace bool `json:"trim_whitespace"`
	NormalizeCase  bool `json:"normalize_case"`

	// Structural sanitization
	SortArrays       bool `json:"sort_arrays"`
	RemoveDuplicates bool `json:"remove_duplicates"`
	AddDefaults      bool `json:"add_defaults"`

	// Safety defaults
	DefaultSafeMode       domain.SafeMode `json:"default_safe_mode"`
	DefaultMaxDiskUsage   int             `json:"default_max_disk_usage"`
	DefaultBackup         time.Duration   `json:"default_backup"`
	DefaultProtectedPaths []string        `json:"default_protected_paths"`
}

// SanitizationChange represents a specific field change with context.
type SanitizationChange struct {
	Original  any       `json:"original"`
	Sanitized any       `json:"sanitized"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// SanitizationResult contains sanitization outcomes.
type SanitizationResult struct {
	SanitizedFields []string                       `json:"sanitized_fields"`
	Warnings        []SanitizationWarning          `json:"warnings"`
	Changes         map[string]*SanitizationChange `json:"changes"`
	Timestamp       time.Time                      `json:"timestamp"`
}

// SanitizationWarning represents a sanitization warning.
type SanitizationWarning struct {
	Field     string `json:"field"`
	Original  any    `json:"original"`
	Sanitized any    `json:"sanitized"`
	Reason    string `json:"reason"`
}

// NewConfigSanitizer creates a configuration sanitizer with default rules.
func NewConfigSanitizer() *ConfigSanitizer {
	return &ConfigSanitizer{
		rules: getDefaultSanitizationRules(),
	}
}

// NewConfigSanitizerWithRules creates a sanitizer with custom rules.
func NewConfigSanitizerWithRules(rules *SanitizationRules) *ConfigSanitizer {
	return &ConfigSanitizer{
		rules: rules,
	}
}

// SanitizeConfig performs comprehensive configuration sanitization.
func (cs *ConfigSanitizer) SanitizeConfig(cfg *domain.Config, validationResult *ValidationResult) {
	start := time.Now()
	defer func() {
		// Update validation result duration if provided
		if validationResult != nil {
			validationResult.Duration = time.Since(start)
		}
	}()

	result := &SanitizationResult{
		SanitizedFields: []string{},
		Warnings:        []SanitizationWarning{},
		Changes:         make(map[string]*SanitizationChange),
		Timestamp:       time.Now(),
	}

	// Sanitize basic fields
	cs.sanitizeBasicFields(cfg, result)

	// Sanitize protected paths
	cs.sanitizeProtectedPaths(cfg, result)

	// Sanitize profiles
	cs.sanitizeProfiles(cfg, result)

	// Apply defaults
	if cs.rules.AddDefaults {
		cs.applyDefaults(cfg, result)
	}

	// Update validation result with sanitization info
	if validationResult != nil {
		validationResult.Sanitized = &ValidationSanitizedData{
			FieldsModified: result.SanitizedFields,
			RulesApplied:   []string{"basic_sanitization", "default_values"},
			Metadata: map[string]string{
				"sanitized_count": strconv.Itoa(len(result.SanitizedFields)),
				"warnings_count":  strconv.Itoa(len(result.Warnings)),
			},
		}

		// Copy sanitization warnings to validation result
		for _, warning := range result.Warnings {
			validationResult.Warnings = append(validationResult.Warnings, ValidationWarning{
				Field:      warning.Field,
				Message:    warning.Reason,
				Suggestion: fmt.Sprintf("Value was changed from %v to %v", warning.Original, warning.Sanitized),
				Context: &ValidationContext{
					Metadata: map[string]string{
						"original":  fmt.Sprintf("%v", warning.Original),
						"sanitized": fmt.Sprintf("%v", warning.Sanitized),
						"field":     warning.Field,
					},
				},
			})
		}
	}
}

// sanitizeBasicFields sanitizes basic configuration fields.
func (cs *ConfigSanitizer) sanitizeBasicFields(cfg *domain.Config, result *SanitizationResult) {
	// Sanitize version
	if cs.rules.TrimWhitespace {
		original := cfg.Version
		cfg.Version = strings.TrimSpace(cfg.Version)
		if original != cfg.Version {
			result.addChange("version", original, cfg.Version, "trimmed whitespace")
		}
	}

	if cs.rules.NormalizeCase {
		original := cfg.Version
		cfg.Version = strings.ToLower(cfg.Version)
		if original != cfg.Version {
			result.addChange("version", original, cfg.Version, "normalized to lowercase")
		}
	}

	// Sanitize max_disk_usage
	if cs.rules.ClampValues {
		original := cfg.MaxDiskUsage
		if cfg.MaxDiskUsage < MinDiskUsagePercent {
			cfg.MaxDiskUsage = MinDiskUsagePercent
			result.addWarning("max_disk_usage", original, cfg.MaxDiskUsage, "clamped to minimum value")
		} else if cfg.MaxDiskUsage > MaxDiskUsagePercent {
			cfg.MaxDiskUsage = MaxDiskUsagePercent
			result.addWarning("max_disk_usage", original, cfg.MaxDiskUsage, "clamped to maximum value")
		}
	}

	if cs.rules.RoundPercentages {
		original := cfg.MaxDiskUsage
		cfg.MaxDiskUsage = int(float64(cfg.MaxDiskUsage+RoundingIncrement/2)/RoundingIncrement) * RoundingIncrement // Round to nearest increment
		if original != cfg.MaxDiskUsage {
			result.addChange("max_disk_usage", original, cfg.MaxDiskUsage, "rounded to nearest 10%")
		}
	}

	// Ensure safe mode defaults
	if cs.rules.DefaultSafeMode.IsEnabled() && !cfg.SafeMode.IsEnabled() {
		original := cfg.SafeMode
		cfg.SafeMode = domain.SafeModeEnabled
		result.addChange("safe_mode", original, cfg.SafeMode, "enabled safe mode for security")
	}
}

// Helper methods

func (r *SanitizationResult) addChange(field string, original, sanitized any, reason string) {
	r.SanitizedFields = append(r.SanitizedFields, field)
	r.Changes[field] = &SanitizationChange{
		Original:  original,
		Sanitized: sanitized,
		Reason:    reason,
		Timestamp: time.Now(),
	}
}

func (r *SanitizationResult) addWarning(field string, original, sanitized any, reason string) {
	r.Warnings = append(r.Warnings, SanitizationWarning{
		Field:     field,
		Original:  original,
		Sanitized: sanitized,
		Reason:    reason,
	})
}

// getDefaultSanitizationRules returns default sanitization configuration.
func getDefaultSanitizationRules() *SanitizationRules {
	return &SanitizationRules{
		NormalizePaths:        true,
		ExpandHomeDir:         true,
		ValidateExists:        false, // Don't validate exists by default as paths may not exist yet
		ClampValues:           true,
		RoundPercentages:      true,
		TrimWhitespace:        true,
		NormalizeCase:         false, // Preserve case for paths and names
		SortArrays:            true,
		RemoveDuplicates:      true,
		AddDefaults:           true,
		DefaultSafeMode:       domain.SafeModeEnabled,
		DefaultMaxDiskUsage:   50,
		DefaultBackup:         24 * time.Hour,
		DefaultProtectedPaths: []string{"/System", "/Applications", "/Library"},
	}
}

// sanitizeOlderThan sanitizes the OlderThan field with whitespace trimming and duration validation.
func (cs *ConfigSanitizer) sanitizeOlderThan(fieldPrefix string, olderThan *string, result *SanitizationResult) {
	if !cs.rules.TrimWhitespace || *olderThan == "" {
		return
	}

	original := *olderThan
	*olderThan = strings.TrimSpace(*olderThan)
	if original != *olderThan {
		result.addChange(fieldPrefix+".older_than", original, *olderThan, "trimmed whitespace")
	}

	// Validate duration format using custom parser
	if _, err := domain.ParseCustomDuration(*olderThan); err != nil {
		result.addWarning(fieldPrefix+".older_than", *olderThan, *olderThan, fmt.Sprintf("invalid duration format: %v", err))
	}
}
