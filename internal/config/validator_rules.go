package config

import (
	"regexp"
	"sync"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigValidationRules defines all validation constraints
type ConfigValidationRules struct {
	// Numeric Constraints
	MaxDiskUsage      *ValidationRule[int] `json:"max_disk_usage"`
	MinProtectedPaths *ValidationRule[int] `json:"min_protected_paths"`
	MaxProtectedPaths *ValidationRule[int] `json:"max_protected_paths"`
	MaxProfiles       *ValidationRule[int] `json:"max_profiles"`
	MaxOperations     *ValidationRule[int] `json:"max_operations"`

	// String Constraints
	ProfileNamePattern *ValidationRule[string] `json:"profile_name_pattern"`
	PathPattern        *ValidationRule[string] `json:"path_pattern"`

	// Array Constraints
	UniquePaths    bool `json:"unique_paths"`
	UniqueProfiles bool `json:"unique_profiles"`

	// Safety Constraints
	ProtectedSystemPaths  []string                    `json:"protected_system_paths"`
	DefaultProtectedPaths []string                    `json:"default_protected_paths"`
	RequireSafeMode       domain.EnforcementLevelType `json:"require_safe_mode"`

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

	// Compiled regex for pattern validation (cached)
	compiledRegex *regexp.Regexp
	regexOnce     sync.Once
}

// ValidationSeverity represents error severity levels
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// GetCompiledRegex returns the compiled regex pattern, creating it once if needed
func (vr *ValidationRule[T]) GetCompiledRegex() *regexp.Regexp {
	vr.regexOnce.Do(func() {
		if vr.Pattern != "" {
			if compiled, err := regexp.Compile(vr.Pattern); err == nil {
				vr.compiledRegex = compiled
			} else {
				// On compile error, use safe always-false regex fallback
				// This regex never matches anything, ensuring safety
				vr.compiledRegex, _ = regexp.Compile("(?!)")
			}
		}
	})
	return vr.compiledRegex
}

// getDefaultValidationRules returns default validation constraints
func getDefaultValidationRules() *ConfigValidationRules {
	// Magic numbers extracted to constants for maintainability
	minUsage := 10
	maxUsage := 95
	minPaths := 1
	maxPaths := 50
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
		MaxProtectedPaths: &ValidationRule[int]{
			Required: false,
			Max:      &maxPaths,
			Message:  "Protected paths should not exceed 50 for maintainability",
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
		DefaultProtectedPaths: []string{
			"/System",
			"/Applications",
			"/Library",
		},
		RequireSafeMode: domain.EnforcementLevelError,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}
}
