package factories

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// CreateBenchmarkConfig creates clean config for performance testing
// Original location: internal/config/test_data.go lines 322-340
func CreateBenchmarkConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: 75,
		Protected:    []string{"/System", "/Applications", "/Library", "/usr", "/etc", "/var"},
		Profiles: map[string]*domain.Profile{
			"daily":  CreateDailyCleanupProfile(),
			"weekly": CreateWeeklyCleanupProfile(),
		},
	}
}

// CreateValidationTestConfigs creates test configurations specifically for validation testing
// Original location: internal/config/test_data.go lines 180-200
func CreateValidationTestConfigs() map[string]*domain.Config {
	return map[string]*domain.Config{
		"valid": createBaseConfig("1.0.0", 50, []string{"/System", "/Library", "/Applications"}),
		"invalid_high_disk": createBaseConfig("1.0.0", 150, []string{"/System"}), // Invalid: too high
	}
}

// createBaseConfig creates a base configuration with standard settings
// Original location: internal/config/test_data.go lines 76-90
func createBaseConfig(version string, maxDiskUsage int, protected []string) *domain.Config {
	return &domain.Config{
		Version:      version,
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: maxDiskUsage,
		Protected:    protected,
		Profiles: map[string]*domain.Profile{
			"daily": createStandardProfile(),
		},
	}
}