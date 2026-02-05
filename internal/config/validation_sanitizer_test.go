package config

import (
	"testing"
)

func TestConfigSanitizer_SanitizeConfig(t *testing.T) {
	sanitizer := NewConfigSanitizer()

	tests := GetSanitizationTestCases()

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
