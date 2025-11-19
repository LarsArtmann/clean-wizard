package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TypeSafeValidationRules defines validation constraints with compile-time type safety
type TypeSafeValidationRules struct {
	// Numeric Constraints - using proper typed bounds
	MaxDiskUsage      *NumericValidationRule `json:"max_disk_usage"`
	MinProtectedPaths *NumericValidationRule `json:"min_protected_paths"`
	MaxProfiles       *NumericValidationRule `json:"max_profiles"`
	MaxOperations     *NumericValidationRule `json:"max_operations"`

	// String Constraints - using proper typed patterns
	ProfileNamePattern *StringValidationRule `json:"profile_name_pattern"`
	PathPattern        *StringValidationRule `json:"path_pattern"`

	// Array Constraints - using proper typed constraints
	UniquePaths    bool `json:"unique_paths"`
	UniqueProfiles bool `json:"unique_profiles"`

	// Safety Constraints - using typed domain values
	ProtectedSystemPaths  []string `json:"protected_system_paths"`
	DefaultProtectedPaths []string `json:"default_protected_paths"`
	RequireSafeMode       bool     `json:"require_safe_mode"`

	// Risk Constraints - using type-safe enums
	MaxRiskLevel   domain.RiskLevelType `json:"max_risk_level"`
	BackupRequired domain.RiskLevelType `json:"backup_required"`
}

// NumericValidationRule provides type-safe numeric validation
type NumericValidationRule struct {
	Required bool   `json:"required"`
	Min      *int   `json:"min,omitempty"`
	Max      *int   `json:"max,omitempty"`
	Message  string `json:"message,omitempty"`
}

// StringValidationRule provides type-safe string validation
type StringValidationRule struct {
	Required bool   `json:"required"`
	Pattern  string `json:"pattern,omitempty"`
	Message  string `json:"message,omitempty"`
}

// NewTypeSafeValidationRules creates type-safe validation rules with defaults
func NewTypeSafeValidationRules() *TypeSafeValidationRules {
	// Extract constants to variables for reference
	minUsage := 10
	maxUsage := 95
	minPaths := 1
	maxProfiles := MaxProfiles
	maxOps := MaxOperations

	return &TypeSafeValidationRules{
		MaxDiskUsage: &NumericValidationRule{
			Required: true,
			Min:      &minUsage,
			Max:      &maxUsage,
			Message:  "Max disk usage must be between 10% and 95%",
		},
		MinProtectedPaths: &NumericValidationRule{
			Required: true,
			Min:      &minPaths,
			Message:  "At least one protected path is required",
		},
		MaxProfiles: &NumericValidationRule{
			Required: false,
			Max:      &maxProfiles,
			Message:  "Consider limiting profiles to maintain clarity",
		},
		MaxOperations: &NumericValidationRule{
			Required: false,
			Max:      &maxOps,
			Message:  "Consider limiting operations per profile for better maintainability",
		},
		ProfileNamePattern: &StringValidationRule{
			Required: true,
			Pattern:  "^[a-zA-Z0-9_-]+$",
			Message:  "Profile names must be alphanumeric with underscores and hyphens",
		},
		PathPattern: &StringValidationRule{
			Required: true,
			Pattern:  "^/.*",
			Message:  "Paths must be absolute (starting with '/')",
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
		RequireSafeMode: true,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}
}

// GetTypeSafeSchemaRules returns type-safe schema rules for external consumption
func (tsvr *TypeSafeValidationRules) GetTypeSafeSchemaRules() *TypeSafeValidationRules {
	// Deep copy to prevent external modifications
	copied := &TypeSafeValidationRules{
		UniquePaths:     tsvr.UniquePaths,
		UniqueProfiles:  tsvr.UniqueProfiles,
		RequireSafeMode: tsvr.RequireSafeMode,
		MaxRiskLevel:    tsvr.MaxRiskLevel,
		BackupRequired:  tsvr.BackupRequired,
	}

	// Deep copy slices
	if tsvr.ProtectedSystemPaths != nil {
		copied.ProtectedSystemPaths = make([]string, len(tsvr.ProtectedSystemPaths))
		copy(copied.ProtectedSystemPaths, tsvr.ProtectedSystemPaths)
	}

	if tsvr.DefaultProtectedPaths != nil {
		copied.DefaultProtectedPaths = make([]string, len(tsvr.DefaultProtectedPaths))
		copy(copied.DefaultProtectedPaths, tsvr.DefaultProtectedPaths)
	}

	// Deep copy numeric rules
	if tsvr.MaxDiskUsage != nil {
		copied.MaxDiskUsage = tsvr.copyNumericRule(tsvr.MaxDiskUsage)
	}
	if tsvr.MinProtectedPaths != nil {
		copied.MinProtectedPaths = tsvr.copyNumericRule(tsvr.MinProtectedPaths)
	}
	if tsvr.MaxProfiles != nil {
		copied.MaxProfiles = tsvr.copyNumericRule(tsvr.MaxProfiles)
	}
	if tsvr.MaxOperations != nil {
		copied.MaxOperations = tsvr.copyNumericRule(tsvr.MaxOperations)
	}

	// Deep copy string rules
	if tsvr.ProfileNamePattern != nil {
		copied.ProfileNamePattern = tsvr.copyStringRule(tsvr.ProfileNamePattern)
	}
	if tsvr.PathPattern != nil {
		copied.PathPattern = tsvr.copyStringRule(tsvr.PathPattern)
	}

	return copied
}

// Helper methods for deep copying
func (tsvr *TypeSafeValidationRules) copyNumericRule(rule *NumericValidationRule) *NumericValidationRule {
	copied := &NumericValidationRule{
		Required: rule.Required,
		Message:  rule.Message,
	}
	if rule.Min != nil {
		min := *rule.Min
		copied.Min = &min
	}
	if rule.Max != nil {
		max := *rule.Max
		copied.Max = &max
	}
	return copied
}

func (tsvr *TypeSafeValidationRules) copyStringRule(rule *StringValidationRule) *StringValidationRule {
	return &StringValidationRule{
		Required: rule.Required,
		Pattern:  rule.Pattern,
		Message:  rule.Message,
	}
}
