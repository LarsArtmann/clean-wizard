package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

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

// IsSatisfiedBy checks if a value satisfies this rule
func (vr *ValidationRule[T]) IsSatisfiedBy(value T) bool {
	// Check required constraint
	if vr.Required {
		// This is basic - more sophisticated checks can be added
	}

	// For now, skip range constraints due to type limitations
	// In a real implementation, we'd use reflection or type-specific methods
	
	// Check enum constraints
	if len(vr.Enum) > 0 {
		found := false
		for _, enum := range vr.Enum {
			if value == enum {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetDefaultValidationRules returns default validation rules
func GetDefaultValidationRules() *ConfigValidationRules {
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

// GetStrictValidationRules returns stricter validation rules for production
func GetStrictValidationRules() *ConfigValidationRules {
	rules := GetDefaultValidationRules()
	
	// Make rules stricter
	minUsage := 20
	maxUsage := 80
	minPaths := 3
	maxProfiles := 5
	maxOps := 10

	rules.MaxDiskUsage.Min = &minUsage
	rules.MaxDiskUsage.Max = &maxUsage
	rules.MinProtectedPaths.Min = &minPaths
	rules.MaxProfiles.Min = &maxProfiles
	rules.MaxOperations.Min = &maxOps
	rules.RequireSafeMode = true
	rules.MaxRiskLevel = domain.RiskMedium

	return rules
}

// GetDevelopmentValidationRules returns relaxed rules for development
func GetDevelopmentValidationRules() *ConfigValidationRules {
	rules := GetDefaultValidationRules()
	
	// Relax rules for development
	rules.RequireSafeMode = false
	rules.UniqueProfiles = false
	rules.MaxRiskLevel = domain.RiskCritical

	return rules
}