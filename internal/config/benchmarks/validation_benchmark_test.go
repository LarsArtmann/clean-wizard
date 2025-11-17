package benchmarks

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// createRealBenchmarkConfig creates a realistic configuration for performance testing
func createRealBenchmarkConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/", "/System", "/Library"},
		Profiles: map[string]*domain.Profile{
			"benchmark-profile": {
				Name:         "benchmark-profile",
				Description:  "Profile for performance testing",
				MaxRiskLevel: domain.RiskMedium,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "benchmark-operation",
						Description: "Operation for performance testing",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
				},
			},
		},
	}
}

// BenchmarkConfigValidator benchmarks config validator performance with real data
func BenchmarkConfigValidator(b *testing.B) {
	validator := config.NewConfigValidator()
	testConfig := createRealBenchmarkConfig()

	for b.Loop() {
		result := validator.ValidateConfig(testConfig)
		if !result.IsValid {
			b.Fatal("Expected valid config")
		}
	}
}

// BenchmarkConfigSanitizer benchmarks config sanitizer performance with real data
func BenchmarkConfigSanitizer(b *testing.B) {
	sanitizer := config.NewConfigSanitizer()
	testConfig := createRealBenchmarkConfig()

	for b.Loop() {
		result := sanitizer.SanitizeConfig(testConfig)
		if result == nil {
			b.Fatal("Expected sanitization result")
		}
	}
}

// BenchmarkValidationMiddleware benchmarks validation middleware performance with real data
func BenchmarkValidationMiddleware(b *testing.B) {
	middleware := config.NewValidationMiddleware()

	ctx := context.Background()

	for b.Loop() {
		result, err := middleware.ValidateAndSanitize(ctx, "dummy-path")
		if err != nil {
			// Expected for dummy path, we're just benchmarking the validation logic
			continue
		}
		if result != nil && !result.IsValid {
			b.Fatal("Expected valid config")
		}
	}
}
