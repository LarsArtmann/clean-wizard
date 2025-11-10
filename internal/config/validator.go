package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidator provides comprehensive type-safe configuration validation
type ConfigValidator struct {
	rules     *ConfigValidationRules
	sanitizer *ConfigSanitizer
}

// ConfigValidationRules defines all validation constraints
type ConfigValidationRules struct {
	// Numeric Constraints
	MaxDiskUsage      *ValidationRule[int] `json:"max_disk_usage"`
	MinProtectedPaths *ValidationRule[int] `json:"min_protected_paths"`
	MaxProfiles       *ValidationRule[int] `json:"max_profiles"`
	MaxOperations     *ValidationRule[int] `json:"max_operations"`

	// String Constraints
	ProfileNamePattern *ValidationRule[string] `json:"profile_name_pattern"`
	PathPattern        *ValidationRule[string] `json:"path_pattern"`

	// Array Constraints
	UniquePaths    bool `json:"unique_paths"`
	UniqueProfiles bool `json:"unique_profiles"`

	// Safety Constraints
	ProtectedSystemPaths []string `json:"protected_system_paths"`
	RequireSafeMode      bool     `json:"require_safe_mode"`

	// Risk Constraints
	MaxRiskLevel   domain.RiskLevel `json:"max_risk_level"`
	BackupRequired domain.RiskLevel `json:"backup_required"`
}

// ValidationRule represents a validation constraint for a specific type
type ValidationRule[T comparable] struct {
	Required bool   `json:"required"`
	Min      *T     `json:"min,omitempty"`
	Max      *T     `json:"max,omitempty"`
	Enum     []T    `json:"enum,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ValidationResult contains validation results with detailed error information
type ValidationResult struct {
	IsValid   bool                   `json:"is_valid"`
	Errors    []ValidationError      `json:"errors,omitempty"`
	Warnings  []ValidationWarning    `json:"warnings,omitempty"`
	Sanitized map[string]interface{} `json:"sanitized,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Field      string                 `json:"field"`
	Rule       string                 `json:"rule"`
	Value      interface{}            `json:"value"`
	Message    string                 `json:"message"`
	Severity   ValidationSeverity     `json:"severity"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// ValidationWarning represents a non-critical validation issue
type ValidationWarning struct {
	Field      string                 `json:"field"`
	Message    string                 `json:"message"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// ValidationSeverity represents error severity levels
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

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
		Sanitized: make(map[string]interface{}),
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

	// Sanitize configuration
	if cv.sanitizer != nil {
		cv.sanitizer.SanitizeConfig(cfg, result)
	}

	result.Duration = time.Since(start)
	result.IsValid = len(result.Errors) == 0

	return result
}

// ValidateField validates a specific configuration field
func (cv *ConfigValidator) ValidateField(field string, value interface{}) error {
	switch field {
	case "max_disk_usage":
		return cv.validateMaxDiskUsage(value)
	case "protected":
		return cv.validateProtectedPaths(value)
	case "profiles":
		return cv.validateProfiles(value)
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
}

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
}

// validateFieldConstraints validates individual fields against rules
func (cv *ConfigValidator) validateFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// MaxDiskUsage validation
	if err := cv.validateMaxDiskUsage(cfg.MaxDiskUsage); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "max_disk_usage",
			Rule:       "range",
			Value:      cfg.MaxDiskUsage,
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: "Set max_disk_usage between 10 and 95",
			Context: map[string]interface{}{
				"min": 10,
				"max": 95,
			},
		})
	}

	// Protected paths validation
	if err := cv.validateProtectedPaths(cfg.Protected); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:      "protected",
			Rule:       "format",
			Value:      cfg.Protected,
			Message:    err.Error(),
			Severity:   SeverityError,
			Suggestion: "Ensure all paths are valid absolute paths",
		})
	}

	// Check for duplicate paths
	if cv.rules.UniquePaths {
		duplicates := cv.findDuplicatePaths(cfg.Protected)
		if len(duplicates) > 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "protected",
				Message:    fmt.Sprintf("Duplicate protected paths found: %v", duplicates),
				Suggestion: "Remove duplicate paths from protected list",
			})
		}
	}
}

// validateCrossFieldConstraints validates relationships between fields
func (cv *ConfigValidator) validateCrossFieldConstraints(cfg *domain.Config, result *ValidationResult) {
	// Safe mode vs risk level consistency
	if !cfg.SafeMode {
		maxRisk := cv.findMaxRiskLevel(cfg)
		if maxRisk == domain.RiskCritical {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "safe_mode",
				Message:    "Safe mode is disabled but critical risk operations exist",
				Suggestion: "Enable safe mode or remove critical risk operations",
				Context: map[string]interface{}{
					"max_risk": string(maxRisk),
				},
			})
		}
	}

	// Check profile count limits
	if cv.rules.MaxProfiles != nil && len(cfg.Profiles) > *cv.rules.MaxProfiles.Min {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      "profiles",
			Message:    fmt.Sprintf("Profile count (%d) exceeds recommended limit (%d)", len(cfg.Profiles), *cv.rules.MaxProfiles.Min),
			Suggestion: "Consider consolidating profiles to improve maintainability",
		})
	}
}

// validateBusinessLogic validates business logic constraints
func (cv *ConfigValidator) validateBusinessLogic(cfg *domain.Config, result *ValidationResult) {
	for name, profile := range cfg.Profiles {
		// Validate profile business logic
		if len(profile.Operations) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Rule:       "business_logic",
				Value:      len(profile.Operations),
				Message:    fmt.Sprintf("Profile '%s' must have at least one operation", name),
				Severity:   SeverityError,
				Suggestion: "Add at least one valid operation to the profile",
			})
		}

		// Check operation count limits
		if cv.rules.MaxOperations != nil && len(profile.Operations) > *cv.rules.MaxOperations.Min {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      fmt.Sprintf("profiles.%s.operations", name),
				Message:    fmt.Sprintf("Profile '%s' has many operations (%d), consider simplifying", name, len(profile.Operations)),
				Suggestion: "Split complex profiles into smaller, focused profiles",
			})
		}
	}
}

// validateSecurityConstraints validates security-related constraints
func (cv *ConfigValidator) validateSecurityConstraints(cfg *domain.Config, result *ValidationResult) {
	// Ensure critical system paths are protected
	for _, systemPath := range cv.rules.ProtectedSystemPaths {
		if !cv.isPathProtected(cfg.Protected, systemPath) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "protected",
				Rule:       "security",
				Value:      cfg.Protected,
				Message:    fmt.Sprintf("Critical system path not protected: %s", systemPath),
				Severity:   SeverityError,
				Suggestion: fmt.Sprintf("Add '%s' to protected paths", systemPath),
				Context: map[string]interface{}{
					"system_path": systemPath,
				},
			})
		}
	}

	// Validate path formats
	for _, path := range cfg.Protected {
		if !filepath.IsAbs(path) {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "protected",
				Rule:       "path_format",
				Value:      path,
				Message:    fmt.Sprintf("Protected path must be absolute: %s", path),
				Severity:   SeverityError,
				Suggestion: fmt.Sprintf("Convert to absolute path: %s", filepath.Join("/", path)),
			})
		}
	}
}

// validateMaxDiskUsage validates max disk usage percentage
func (cv *ConfigValidator) validateMaxDiskUsage(value interface{}) error {
	usage, ok := value.(int)
	if !ok {
		return fmt.Errorf("max_disk_usage must be an integer, got %T", value)
	}

	if cv.rules.MaxDiskUsage != nil {
		if cv.rules.MaxDiskUsage.Min != nil && usage < *cv.rules.MaxDiskUsage.Min {
			return fmt.Errorf("max_disk_usage (%d) below minimum (%d)", usage, *cv.rules.MaxDiskUsage.Min)
		}
		if cv.rules.MaxDiskUsage.Max != nil && usage > *cv.rules.MaxDiskUsage.Max {
			return fmt.Errorf("max_disk_usage (%d) above maximum (%d)", usage, *cv.rules.MaxDiskUsage.Max)
		}
	}

	return nil
}

// validateProtectedPaths validates protected paths array
func (cv *ConfigValidator) validateProtectedPaths(value interface{}) error {
	paths, ok := value.([]string)
	if !ok {
		return fmt.Errorf("protected must be a string array, got %T", value)
	}

	if cv.rules.MinProtectedPaths != nil && len(paths) < *cv.rules.MinProtectedPaths.Min {
		return fmt.Errorf("protected paths count (%d) below minimum (%d)", len(paths), *cv.rules.MinProtectedPaths.Min)
	}

	for i, path := range paths {
		if path == "" {
			return fmt.Errorf("protected path %d cannot be empty", i)
		}
	}

	return nil
}

// validateProfiles validates profiles map
func (cv *ConfigValidator) validateProfiles(value interface{}) error {
	profiles, ok := value.(map[string]*domain.Profile)
	if !ok {
		return fmt.Errorf("profiles must be a map of profiles, got %T", value)
	}

	for name, profile := range profiles {
		if err := cv.validateProfileName(name); err != nil {
			return err
		}
		if err := profile.Validate(name); err != nil {
			return fmt.Errorf("profile %s: %w", name, err)
		}
	}

	return nil
}

// validateProfileName validates profile name format
func (cv *ConfigValidator) validateProfileName(name string) error {
	if cv.rules.ProfileNamePattern != nil && cv.rules.ProfileNamePattern.Pattern != "" {
		// Simple alphanumeric with underscores validation
		for _, char := range name {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-') {
				return fmt.Errorf("profile name '%s' contains invalid character: %c", name, char)
			}
		}
	}
	return nil
}

// Helper methods

func (cv *ConfigValidator) findDuplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	duplicates := []string{}

	for _, path := range paths {
		if seen[path] {
			duplicates = append(duplicates, path)
		} else {
			seen[path] = true
		}
	}

	return duplicates
}

func (cv *ConfigValidator) findMaxRiskLevel(cfg *domain.Config) domain.RiskLevel {
	maxRisk := domain.RiskLow
	for _, profile := range cfg.Profiles {
		for _, op := range profile.Operations {
			if op.RiskLevel == domain.RiskCritical {
				return domain.RiskCritical
			}
			if op.RiskLevel == domain.RiskHigh {
				maxRisk = domain.RiskHigh
			} else if op.RiskLevel == domain.RiskMedium && maxRisk == domain.RiskLow {
				maxRisk = domain.RiskMedium
			}
		}
	}
	return maxRisk
}

func (cv *ConfigValidator) isPathProtected(protected []string, targetPath string) bool {
	for _, path := range protected {
		if strings.HasPrefix(path, targetPath) || strings.HasPrefix(targetPath, path) {
			return true
		}
	}
	return false
}

// getDefaultValidationRules returns default validation constraints
func getDefaultValidationRules() *ConfigValidationRules {
	minUsage := 10
	maxUsage := 95
	minPaths := 1
	maxProfiles := 10
	maxOps := 20

	return &ConfigValidationRules{
		MaxDiskUsage: &ValidationRule[int]{
			Required: true,
			Min:      &minUsage,
			Max:      &maxUsage,
			Message:  "Max disk usage must be between 10% and 95%",
		},
		MinProtectedPaths: &ValidationRule[int]{
			Required: true,
			Min:      &minPaths,
			Message:  "At least one protected path is required",
		},
		MaxProfiles: &ValidationRule[int]{
			Required: false,
			Min:      &maxProfiles,
			Message:  "Consider limiting profiles to maintain clarity",
		},
		MaxOperations: &ValidationRule[int]{
			Required: false,
			Min:      &maxOps,
			Message:  "Consider limiting operations per profile for better maintainability",
		},
		ProfileNamePattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^[a-zA-Z0-9_-]+$",
			Message:  "Profile names must be alphanumeric with underscores and hyphens",
		},
		UniquePaths:    true,
		UniqueProfiles: true,
		ProtectedSystemPaths: []string{
			"/",
			"/System",
			"/Applications",
			"/Library",
			"/usr",
			"/etc",
			"/var",
			"/bin",
			"/sbin",
		},
		RequireSafeMode: true,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}
}
