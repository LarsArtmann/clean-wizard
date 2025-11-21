package validation_test

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// createWhitespacedConfigForSanitizer creates a test configuration with whitespace that needs cleaning
func TestConfigSanitizer_SanitizeConfig(t *testing.T) {
	sanitizer := NewConfigSanitizer()

	tests := []struct {
		name             string
		config           *domain.Config
		expectedChanges  []string
		expectedWarnings int
	}{}

	// Add standard test cases from shared test data
	standardCases := GetStandardTestCases()
	for _, tc := range standardCases {
		tests = append(tests, struct {
			name             string
			config           *domain.Config
			expectedChanges  []string
			expectedWarnings int
		}{
			name:             tc.name,
			config:           tc.config,
			expectedChanges:  tc.expectedChanges,
			expectedWarnings: tc.expectedWarnings,
		})
	}

	// Add unique test case for this file
	tests = append(tests, struct {
		name             string
		config           *domain.Config
		expectedChanges  []string
		expectedWarnings int
	}{
		name: "duplicate paths",
		config: &domain.Config{
			Version:      "1.0.0",
			SafetyLevel:  domain.SafetyLevelEnabled,
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
							Status:      domain.StatusEnabled,
						},
					},
					Status: domain.StatusEnabled,
				},
			},
		},
		expectedChanges:  []string{"profiles.daily.operations[0].settings"}, // Settings sanitized
		expectedWarnings: 0,
	})

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
