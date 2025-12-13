package validation_test

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/application/config/factories"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// CreateBenchmarkConfig alias for test compatibility
var CreateBenchmarkConfig = factories.CreateBenchmarkConfig

// BenchmarkValidation_ConfigValidation tests full configuration validation performance
func BenchmarkValidation_ConfigValidation(b *testing.B) {
	// Create complex configuration for realistic testing
	cfg := CreateBenchmarkConfig()

	validator := NewConfigValidator()

	for b.Loop() {
		validator.ValidateConfig(cfg)
	}
}

// BenchmarkValidation_ProfileNameValidation tests profile name validation performance
func BenchmarkValidation_ProfileNameValidation(b *testing.B) {
	validator := NewConfigValidator()
	profileNames := []string{
		"daily-cleanup",
		"weekly_maintenance",
		"monthly-deep-clean",
		"system-optimization",
		"cache-cleanup",
		"log-rotation",
		"temp-cleanup",
		"backup-cleanup",
		"security-scan",
		"performance-optimization",
	}

	for b.Loop() {
		for _, name := range profileNames {
			_ = validator.validateProfileName(name)
		}
	}
}

// BenchmarkValidation_OperationSettingsValidation tests operation settings validation performance
func BenchmarkValidation_OperationSettingsValidation(b *testing.B) {
	settings := &shared.OperationSettings{
		NixGenerations: &shared.NixGenerationsSettings{
			Generations:  3,
			Optimization: shared.OptimizationLevelConservative,
		},
		TempFiles: &shared.TempFilesSettings{
			OlderThan: "7d",
			Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
		},
		Homebrew: &shared.HomebrewSettings{
			FileSelectionStrategy: shared.FileSelectionStrategyUnusedOnly,
			Prune:                 "30d",
		},
		SystemTemp: &shared.SystemTempSettings{
			Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
			OlderThan: "14d",
		},
	}

	operationTypes := []shared.OperationType{
		shared.OperationTypeNixGenerations,
		shared.OperationTypeTempFiles,
		shared.OperationTypeHomebrew,
		shared.OperationTypeSystemTemp,
	}

	for b.Loop() {
		for _, opType := range operationTypes {
			_ = settings.ValidateSettings(opType)
		}
	}
}

// BenchmarkValidation_MaxDiskUsageValidation tests max disk usage validation performance
func BenchmarkValidation_MaxDiskUsageValidation(b *testing.B) {
	validator := NewConfigValidator()
	testValues := []int{10, 25, 50, 75, 95}

	for b.Loop() {
		for _, value := range testValues {
			result := NewValidationResult() // Fresh result for each validation
			cfg := &shared.Config{}         // Fresh config for each test
			cfg.MaxDiskUsage = value
			validator.validateFieldConstraints(cfg, result)
		}
	}
}

// BenchmarkValidation_RegexCompilation tests regex pattern compilation performance
func BenchmarkValidation_RegexCompilation(b *testing.B) {
	pattern := "^[a-zA-Z0-9_-]+$"
	testNames := []string{
		"valid-name",
		"another_valid_name",
		"test-123",
		"FINAL_test",
	}

	// Create temporary rule to test compilation
	rule := &ValidationRule[string]{
		Required: true,
		Pattern:  pattern,
		Message:  "Invalid profile name format",
	}

	for b.Loop() {
		for _, name := range testNames {
			if compiledRegex := rule.GetCompiledRegex(); compiledRegex != nil {
				_ = compiledRegex.MatchString(name)
			}
		}
	}
}
