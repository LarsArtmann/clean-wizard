package benchmarks

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BenchmarkValidationMiddleware benchmarks validation middleware performance
func BenchmarkValidationMiddleware(b *testing.B) {
	middleware := config.NewValidationMiddleware()
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := middleware.ValidateAndSanitize(ctx, "dummy-path")
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
		if !result.IsValid {
			b.Fatal("Expected valid config")
		}
	}
}

// BenchmarkConfigValidator benchmarks config validator performance
func BenchmarkConfigValidator(b *testing.B) {
	validator := config.NewConfigValidator()
	
	testConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 50,
		Protected:    []string{"/", "/System", "/Library"},
		Profiles:     map[string]*domain.Profile{},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateConfig(testConfig)
		if !result.IsValid {
			b.Fatal("Expected valid config")
		}
	}
}

// BenchmarkConfigSanitizer benchmarks config sanitizer performance
func BenchmarkConfigSanitizer(b *testing.B) {
	sanitizer := config.NewConfigSanitizer()
	
	testConfig := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 50,
		Protected:    []string{"/", "/System", "/Library"},
		Profiles:     map[string]*domain.Profile{},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := sanitizer.SanitizeConfig(testConfig)
		if result == nil {
			b.Fatal("Expected sanitization result")
		}
	}
}
