package factories

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// CreateBenchmarkConfig creates clean config for performance testing
func CreateBenchmarkConfig() *config.Config {
	return &Config{
		Version:      "1.0.0",
		SafetyLevel:  shared.SafetyLevelEnabled,
		MaxDiskUsage: 75,
		Protected:    []string{"/System", "/Applications", "/Library", "/usr", "/etc", "/var"},
		Profiles: map[string]*ConfigProfile{
			"daily":         CreateDailyCleanupProfile(),
			"weekly":        CreateWeeklyCleanupProfile(),
			"comprehensive": CreateComprehensiveCleanupProfile(),
		},
		CurrentProfile: "daily",
		LastClean:      time.Now(),
		Updated:        time.Now(),
	}
}

// CreateComprehensiveCleanupProfile creates a profile for comprehensive cleanup
func CreateComprehensiveCleanupProfile() *ConfigProfile {
	return &ConfigProfile{
		Name:        "Comprehensive Cleanup",
		Description: "Comprehensive cleanup including development tools",
		Status:      shared.StatusDisabled, // Disabled by default for safety
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{NixGenerations: &shared.NixGenerationsSettings{Generations: 2}},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files from user directories",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{TempFiles: &shared.TempFilesSettings{OlderThan: "1d"}},
			},
		},
	}
}
