package config_test

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/application/config/factories"
)

// TestIntegration_ValidationSanitizationPipeline tests complete validation and sanitization workflow
func TestIntegration_ValidationSanitizationPipeline(t *testing.T) {
	t.Run("Complete pipeline with complex configuration", func(t *testing.T) {
		// Create complex configuration that exercises all validation paths
		cfg := factories.CreateIntegrationConfig()

		// Step 1: Sanitization (run first to clean data)
		sanitizer := NewConfigSanitizer()
		var sanitizationResult ValidationResult
		sanitizer.SanitizeConfig(cfg, &sanitizationResult)

		if sanitizationResult.Duration <= 0 {
			t.Errorf("Sanitization should take positive time, got: %v", sanitizationResult.Duration)
		}

		// Step 2: Validation (run after sanitization)
		validator := NewConfigValidator()
		validationResult := validator.ValidateConfig(cfg)

		if !validationResult.IsValid {
			t.Errorf("Configuration should be valid after sanitization, got errors: %v", validationResult.Errors)
		}

		// Step 3: Post-sanitization validation
		postValidationResult := validator.ValidateConfig(cfg)

		if !postValidationResult.IsValid {
			t.Errorf("Configuration should remain valid after sanitization, got errors: %v", postValidationResult.Errors)
		}

		// Verify sanitization effects
		originalVersion := " 1.0.0  "
		if cfg.Version == originalVersion {
			t.Errorf("Version should be sanitized from whitespace")
		}

		if cfg.Version != "1.0.0" {
			t.Errorf("Expected sanitized version '1.0.0', got: %s", cfg.Version)
		}

		// Check for duplicate removal in protected paths
		protectedPathCount := 0
		for _, path := range cfg.Protected {
			if path == "/System" {
				protectedPathCount++
			}
		}
		if protectedPathCount > 1 {
			t.Errorf("Duplicate protected paths should be removed")
		}

		// Verify profile name sanitization
		if cfg.Profiles["daily"].Name == "  Daily Cleanup  " {
			t.Errorf("Profile name should be sanitized from whitespace")
		}

		if cfg.Profiles["daily"].Name != "Daily Cleanup" {
			t.Errorf("Expected sanitized profile name 'Daily Cleanup', got: %s", cfg.Profiles["daily"].Name)
		}

		// Verify operation name sanitization
		dailyProfile := cfg.Profiles["daily"]
		if dailyProfile.Operations[0].Name == "nix-generations" {
			// Should remain unchanged (already clean)
		} else {
			t.Errorf("Nix operation name should remain unchanged")
		}

		// Verify settings sanitization
		if tempFilesSettings := dailyProfile.Operations[1].Settings.TempFiles; tempFilesSettings != nil {
			if tempFilesSettings.OlderThan == " 7d  " {
				t.Errorf("Temp files older_than should be sanitized from whitespace")
			}

			if tempFilesSettings.OlderThan != "7d" {
				t.Errorf("Expected sanitized older_than '7d', got: %s", tempFilesSettings.OlderThan)
			}

			// Check for duplicate removal in excludes
			excludeCount := 0
			for _, exclude := range tempFilesSettings.Excludes {
				if exclude == "/tmp/keep" {
					excludeCount++
				}
			}
			if excludeCount > 1 {
				t.Errorf("Duplicate temp files excludes should be removed")
			}
		}

		// Calculate fields modified safely (guard against nil)
		fieldsModified := 0
		if sanitizationResult.Sanitized != nil {
			fieldsModified = len(sanitizationResult.Sanitized.FieldsModified)
		}

		// Log successful integration
		t.Logf("✓ Integration test passed - validation and sanitization pipeline working correctly")
		t.Logf("  - Validation: %d errors, %d warnings", len(validationResult.Errors), len(validationResult.Warnings))
		t.Logf("  - Sanitization: %d fields modified, %d warnings", fieldsModified, len(sanitizationResult.Warnings))
		t.Logf("  - Post-validation: %d errors, %d warnings", len(postValidationResult.Errors), len(postValidationResult.Warnings))
		t.Logf("  - Total duration: %v", validationResult.Duration+sanitizationResult.Duration)
	})
}
