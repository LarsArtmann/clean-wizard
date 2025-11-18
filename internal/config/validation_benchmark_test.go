package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BenchmarkValidation_ConfigValidation tests full configuration validation performance
func BenchmarkValidation_ConfigValidation(b *testing.B) {
	// Create complex configuration for realistic testing
	cfg := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 75,
		Protected:    []string{"/System", "/Applications", "/Library", "/usr", "/etc", "/var"},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        "Daily Cleanup",
				Description: "Daily system cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							NixGenerations: &domain.NixGenerationsSettings{
								Generations: 3,
								Optimize:    true,
							},
						},
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							TempFiles: &domain.TempFilesSettings{
								OlderThan: "7d",
								Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
							},
						},
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean Homebrew",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							Homebrew: &domain.HomebrewSettings{
								UnusedOnly: true,
								Prune:      "30d",
							},
						},
					},
					{
						Name:        "system-temp",
						Description: "Clean system temp",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							SystemTemp: &domain.SystemTempSettings{
								Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
								OlderThan: "14d",
							},
						},
					},
				},
				Enabled: true,
			},
			"weekly": {
				Name:        "Weekly Cleanup",
				Description: "Weekly deep cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Deep Nix cleanup",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							NixGenerations: &domain.NixGenerationsSettings{
								Generations: 5,
								Optimize:    true,
							},
						},
					},
				},
				Enabled: true,
			},
		},
	}

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
	settings := &domain.OperationSettings{
		NixGenerations: &domain.NixGenerationsSettings{
			Generations: 3,
			Optimize:    true,
		},
		TempFiles: &domain.TempFilesSettings{
			OlderThan: "7d",
			Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
		},
		Homebrew: &domain.HomebrewSettings{
			UnusedOnly: true,
			Prune:      "30d",
		},
		SystemTemp: &domain.SystemTempSettings{
			Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
			OlderThan: "14d",
		},
	}

	operationTypes := []domain.OperationType{
		domain.OperationTypeNixGenerations,
		domain.OperationTypeTempFiles,
		domain.OperationTypeHomebrew,
		domain.OperationTypeSystemTemp,
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
			_ = validator.validateMaxDiskUsage(value)
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
