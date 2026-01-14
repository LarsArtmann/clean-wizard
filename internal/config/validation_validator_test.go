package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestConfigValidator_ValidateConfig(t *testing.T) {
	// Use less strict rules for testing
	testRules := &ConfigValidationRules{
		MaxDiskUsage: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 10; return &i }(),
			Max:      func() *int { i := 95; return &i }(),
			Message:  "Max disk usage must be between 10% and 95%",
		},
		MinProtectedPaths: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 1; return &i }(),
			Message:  "At least one protected path is required",
		},
		ProfileNamePattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^[a-zA-Z0-9_-]+$",
			Message:  "Profile names must be alphanumeric with underscores and hyphens",
		},
		UniquePaths:    true,
		UniqueProfiles: true,
		ProtectedSystemPaths: []string{ // Reduced for tests
			"/",
			"/System",
			"/Library",
		},
		RequireSafeMode: true,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}
	validator := NewConfigValidatorWithRules(testRules)

	tests := []struct {
		name        string
		config      *domain.Config
		expectValid bool
		expectError string
	}{
		{
			name: "valid config",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     domain.ProfileStatusEnabled,
							},
						},
						Enabled: domain.ProfileStatusEnabled,
					},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid max disk usage",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 150, // Invalid: > 95
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     domain.ProfileStatusEnabled,
							},
						},
						Enabled: domain.ProfileStatusEnabled,
					},
				},
			},
			expectValid: false,
			expectError: "max_disk_usage",
		},
		{
			name: "missing version",
			config: &domain.Config{
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     domain.ProfileStatusEnabled,
							},
						},
						Enabled: domain.ProfileStatusEnabled,
					},
				},
			},
			expectValid: false,
			expectError: "version",
		},
		{
			name: "empty protected paths",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     domain.ProfileStatusEnabled,
							},
						},
						Enabled: domain.ProfileStatusEnabled,
					},
				},
			},
			expectValid: false,
			expectError: "protected",
		},
		{
			name: "no profiles",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles:     map[string]*domain.Profile{},
			},
			expectValid: false,
			expectError: "profiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.config)

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected validity: %v, got: %v", tt.expectValid, result.IsValid)
				if !result.IsValid {
					for _, err := range result.Errors {
						t.Logf("Error: %s - %s", err.Field, err.Message)
					}
				}
			}

			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Field == tt.expectError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error in field: %s", tt.expectError)
				}
			}
		})
	}
}

func TestConfigValidator_ValidateField(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		field       string
		value       any
		expectError bool
	}{
		{
			name:        "valid max disk usage",
			field:       "max_disk_usage",
			value:       50,
			expectError: false,
		},
		{
			name:        "invalid max disk usage - too high",
			field:       "max_disk_usage",
			value:       150,
			expectError: true,
		},
		{
			name:        "invalid max disk usage - too low",
			field:       "max_disk_usage",
			value:       5,
			expectError: true,
		},
		{
			name:        "invalid max disk usage - wrong type",
			field:       "max_disk_usage",
			value:       "50",
			expectError: true,
		},
		{
			name:        "valid protected paths",
			field:       "protected",
			value:       []string{"/System", "/Library"},
			expectError: false,
		},
		{
			name:        "invalid protected paths - empty",
			field:       "protected",
			value:       []string{},
			expectError: true,
		},
		{
			name:        "invalid protected paths - wrong type",
			field:       "protected",
			value:       "/System",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.field, tt.value)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for field: %s", tt.field)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for field %s: %v", tt.field, err)
			}
		})
	}
}
