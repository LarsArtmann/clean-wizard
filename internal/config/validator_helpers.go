package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BasicValidator handles basic structure validation
type BasicValidator struct{}

// NewBasicValidator creates a basic validator
func NewBasicValidator() *BasicValidator {
	return &BasicValidator{}
}

// ValidateBasicStructure validates basic configuration structure
func (bv *BasicValidator) ValidateBasicStructure(cfg *domain.Config, result *ValidationResult) {
	if cfg == nil {
		result.AddError("config", "required", nil, "Configuration cannot be nil", "Provide valid configuration", SeverityCritical)
		return
	}

	// Validate version
	if cfg.Version == "" {
		result.AddError("version", "required", "", "Version is required", "Set configuration version", SeverityError)
	}

	// Validate protected paths
	if len(cfg.Protected) == 0 {
		result.AddError("protected", "required", "", "Protected paths cannot be empty", "Add system paths to protect", SeverityError)
	}

	// Validate profiles
	if len(cfg.Profiles) == 0 {
		result.AddError("profiles", "required", "", "At least one profile is required", "Add cleanup profiles", SeverityError)
	}
}

// FieldValidator handles field-level validation
type FieldValidator struct{}

// NewFieldValidator creates a field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{}
}

// ValidateFields validates all configuration fields
func (fv *FieldValidator) ValidateFields(cfg *domain.Config, rules *ConfigValidationRules, result *ValidationResult) {
	// Validate max disk usage
	if rules.MaxDiskUsage != nil {
		if rules.MaxDiskUsage.Required && cfg.MaxDiskUsage == 0 {
			result.AddError("max_disk_usage", "required", cfg.MaxDiskUsage, "Max disk usage is required", "Set maximum disk usage percentage", SeverityError)
		} else if rules.MaxDiskUsage.Min != nil && cfg.MaxDiskUsage < *rules.MaxDiskUsage.Min {
			result.AddError("max_disk_usage", "min_value", cfg.MaxDiskUsage, fmt.Sprintf("Max disk usage must be at least %d", *rules.MaxDiskUsage.Min), "Increase max disk usage", SeverityError)
		} else if rules.MaxDiskUsage.Max != nil && cfg.MaxDiskUsage > *rules.MaxDiskUsage.Max {
			result.AddError("max_disk_usage", "max_value", cfg.MaxDiskUsage, fmt.Sprintf("Max disk usage cannot exceed %d", *rules.MaxDiskUsage.Max), "Decrease max disk usage", SeverityError)
		}
	}

	// Validate protected paths
	if rules.MinProtectedPaths != nil && len(cfg.Protected) < *rules.MinProtectedPaths {
		result.AddError("protected", "min_count", len(cfg.Protected), fmt.Sprintf("At least %d protected paths required", *rules.MinProtectedPaths), "Add more protected paths", SeverityError)
	}

	// Validate paths uniqueness
	if rules.UniquePaths && fv.hasDuplicates(cfg.Protected) {
		result.AddError("protected", "unique", cfg.Protected, "Protected paths must be unique", "Remove duplicate paths", SeverityError)
	}

	// Validate path patterns
	for i, path := range cfg.Protected {
		if rules.PathPattern != nil && !fv.matchesPattern(path, rules.PathPattern.Pattern) {
			result.AddError(fmt.Sprintf("protected[%d]", i), "pattern", path, "Invalid path format", "Use absolute paths", SeverityError)
		}
	}

	// Validate profiles
	if rules.MaxProfiles != nil && len(cfg.Profiles) > *rules.MaxProfiles {
		result.AddError("profiles", "max_count", len(cfg.Profiles), fmt.Sprintf("Too many profiles (max: %d)", *rules.MaxProfiles), "Remove some profiles", SeverityWarning)
	}

	// Validate profile uniqueness
	if rules.UniqueProfiles {
		profileNames := make(map[string]bool)
		for name := range cfg.Profiles {
			if profileNames[name] {
				result.AddError("profiles", "unique", name, "Duplicate profile name", "Rename profile", SeverityError)
			}
			profileNames[name] = true
		}
	}
}

// BusinessValidator handles business logic validation
type BusinessValidator struct{}

// NewBusinessValidator creates a business validator
func NewBusinessValidator() *BusinessValidator {
	return &BusinessValidator{}
}

// ValidateBusinessRules validates business rules
func (bv *BusinessValidator) ValidateBusinessRules(cfg *domain.Config, rules *ConfigValidationRules, result *ValidationResult) {
	// Validate each profile
	for name, profile := range cfg.Profiles {
		bv.validateProfile(name, profile, rules, result)
	}
}

// validateProfile validates a single profile
func (bv *BusinessValidator) validateProfile(name string, profile *domain.Profile, rules *ConfigValidationRules, result *ValidationResult) {
	// Validate profile name pattern
	if rules.ProfileNamePattern != nil && !bv.matchesPattern(profile.Name, rules.ProfileNamePattern.Pattern) {
		result.AddError(fmt.Sprintf("profiles.%s.name", name), "pattern", profile.Name, "Invalid profile name format", "Use lowercase letters, numbers, hyphens", SeverityError)
	}

	// Validate operations count
	if rules.MaxOperations != nil && len(profile.Operations) > *rules.MaxOperations {
		result.AddError(fmt.Sprintf("profiles.%s.operations", name), "max_count", len(profile.Operations), fmt.Sprintf("Too many operations (max: %d)", *rules.MaxOperations), "Remove some operations", SeverityWarning)
	}

	// Validate max risk level
	if rules.MaxRiskLevel != "" {
		for i, op := range profile.Operations {
			if op.RiskLevel.IsHigherOrEqualThan(rules.MaxRiskLevel) {
				result.AddWarning(fmt.Sprintf("profiles.%s.operations[%d].risk_level", name, i), fmt.Sprintf("High risk operation: %s", op.RiskLevel))
			}
		}
	}
}

// SecurityValidator handles security validation
type SecurityValidator struct{}

// NewSecurityValidator creates a security validator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{}
}

// ValidateSecurity validates security constraints
func (sv *SecurityValidator) ValidateSecurity(cfg *domain.Config, rules *ConfigValidationRules, result *ValidationResult) {
	// Validate safe mode requirement
	if rules.RequireSafeMode && !cfg.SafeMode {
		result.AddError("safe_mode", "required", cfg.SafeMode, "Safe mode is required for security", "Enable safe mode", SeverityError)
	}

	// Validate protected system paths
	if len(rules.ProtectedSystemPaths) > 0 {
		for _, sysPath := range rules.ProtectedSystemPaths {
			if !sv.isPathProtected(cfg.Protected, sysPath) {
				result.AddError("protected", "system_path", sysPath, "System path not protected", "Add to protected paths", SeverityError)
			}
		}
	}
}

// Helper methods

func (fv *FieldValidator) hasDuplicates(slice []string) bool {
	seen := make(map[string]bool)
	for _, item := range slice {
		if seen[item] {
			return true
		}
		seen[item] = true
	}
	return false
}

func (fv *FieldValidator) matchesPattern(value string, pattern string) bool {
	// Simple pattern matching - in production, use regex
	return strings.HasPrefix(value, "/")
}

func (bv *BusinessValidator) matchesPattern(value string, pattern string) bool {
	// Simple pattern matching - in production, use regex
	return strings.ToLower(value) == value
}

func (sv *SecurityValidator) isPathProtected(protected []string, path string) bool {
	for _, p := range protected {
		if p == path {
			return true
		}
	}
	return false
}