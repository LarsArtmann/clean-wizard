package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestConfigSanitizer_SanitizeConfig(t *testing.T) {
	sanitizer := NewConfigSanitizer()

	tests := []struct {
		name             string
		config           *domain.Config
		expectedChanges  []string
		expectedWarnings int
	}{
		{
			name: "whitespace cleanup",
			config: &domain.Config{
				Version:      "  1.0.0  ",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "  daily  ",
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
			expectedChanges:  []string{"version", "profiles.daily.name"},
			expectedWarnings: 0,
		},
		{
			name: "max disk usage clamping",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 150, // Will be clamped to 95
				Protected:    []string{"/System", "/Library"},
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
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1,
		},
		{
			name: "duplicate paths",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     domain.SafeModeEnabled,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library", "/System"}, // Duplicate /System
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
			expectedChanges:  []string{"profiles.daily.operations[0].settings"}, // Settings sanitized
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validation result to capture sanitization
			validationResult := &ValidationResult{
				IsValid:   true,
				Errors:    []ValidationError{},
				Warnings:  []ValidationWarning{},
				Sanitized: nil,
			}

			sanitizer.SanitizeConfig(tt.config, validationResult)

			// Check expected changes
			if validationResult.Sanitized != nil {
				sanitizedFields := validationResult.Sanitized.FieldsModified
				if len(sanitizedFields) == 0 && len(tt.expectedChanges) > 0 {
					t.Errorf("Expected %d sanitized fields, got none", len(tt.expectedChanges))
				}

				for _, expectedChange := range tt.expectedChanges {
					found := false
					for _, field := range sanitizedFields {
						if field == expectedChange || len(field) > len(expectedChange) && field[len(field)-len(expectedChange):] == expectedChange {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected change for field '%s', not found in: %v", expectedChange, sanitizedFields)
					}
				}
			}

			// Check warnings count using ValidationResult.Warnings
			if len(validationResult.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectedWarnings, len(validationResult.Warnings), validationResult.Warnings)
			}
		})
	}
}
