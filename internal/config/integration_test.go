package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestIntegration_ValidationSanitizationPipeline tests complete validation and sanitization workflow.
func TestIntegration_ValidationSanitizationPipeline(t *testing.T) {
	t.Run("Complete pipeline with complex configuration", func(t *testing.T) {
		cfg := CreateIntegrationTestConfig()

		validationResult, _, sanitizationResult := runValidationPipeline(t, cfg)

		verifyValidationPassed(t, validationResult)
		verifySanitizationEffects(t, cfg, sanitizationResult)
		verifyPostValidation(t, cfg, validationResult, sanitizationResult)
	})
}

func runValidationPipeline(
	t *testing.T,
	cfg *domain.Config,
) (*ValidationResult, *ConfigSanitizer, *ValidationResult) {
	t.Helper()

	validator := NewConfigValidator()
	validationResult := validator.ValidateConfig(cfg)

	sanitizer := NewConfigSanitizer()

	var sanitizationResult ValidationResult
	sanitizer.SanitizeConfig(cfg, &sanitizationResult)

	return validationResult, sanitizer, &sanitizationResult
}

func verifyValidationPassed(t *testing.T, result *ValidationResult) {
	t.Helper()

	if !result.IsValid {
		t.Errorf("Configuration should be valid, got errors: %v", result.Errors)
	}
}

func verifySanitizationEffects(t *testing.T, cfg *domain.Config, result *ValidationResult) {
	t.Helper()

	if result.Duration <= 0 {
		t.Errorf("Sanitization should take positive time, got: %v", result.Duration)
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("Expected sanitized version '1.0.0', got: %s", cfg.Version)
	}

	protectedPathCount := 0
	for _, path := range cfg.Protected {
		if path == "/System" {
			protectedPathCount++
		}
	}

	if protectedPathCount > 1 {
		t.Errorf("Duplicate protected paths should be removed")
	}

	if cfg.Profiles["daily"].Name != "Daily Cleanup" {
		t.Errorf(
			"Expected sanitized profile name 'Daily Cleanup', got: %s",
			cfg.Profiles["daily"].Name,
		)
	}
}

func verifyPostValidation(
	t *testing.T,
	cfg *domain.Config,
	validationResult *ValidationResult,
	sanitizationResult *ValidationResult,
) {
	t.Helper()

	validator := NewConfigValidator()
	postValidationResult := validator.ValidateConfig(cfg)

	if !postValidationResult.IsValid {
		t.Errorf(
			"Configuration should remain valid after sanitization, got errors: %v",
			postValidationResult.Errors,
		)
	}

	fieldsModified := 0
	if sanitizationResult.Sanitized != nil {
		fieldsModified = len(sanitizationResult.Sanitized.FieldsModified)
	}

	t.Logf("✓ Integration test passed - validation and sanitization pipeline working correctly")
	t.Logf(
		"  - Validation: %d errors, %d warnings",
		len(validationResult.Errors),
		len(validationResult.Warnings),
	)
	t.Logf(
		"  - Sanitization: %d fields modified, %d warnings",
		fieldsModified,
		len(sanitizationResult.Warnings),
	)
	t.Logf(
		"  - Post-validation: %d errors, %d warnings",
		len(postValidationResult.Errors),
		len(postValidationResult.Warnings),
	)
	t.Logf("  - Total duration: %v", validationResult.Duration+sanitizationResult.Duration)
}
