package factories

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
)

// createStandardProfile creates a standard daily cleanup profile
// Exported for test compatibility
func CreateStandardProfile() *config.Profile {
	return &config.Profile{
		Name:        "daily",
		Description: "Daily cleanup",
		Operations: []config.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   config.RiskLow,
				Status:      config.StatusEnabled,
			},
		},
		Status: config.StatusEnabled,
	}
}

// CreateDailyCleanupProfile creates a standard daily cleanup profile with all operations
// Original location: internal/config/test_data.go lines 230-295
func CreateDailyCleanupProfile() *config.Profile {
	return &config.Profile{
		Name:        "Daily Cleanup",
		Description: "Daily system cleanup",
		Operations: []config.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   config.RiskLow,
				Status:      config.StatusEnabled,
				Settings: &config.OperationSettings{
					NixGenerations: &config.NixGenerationsSettings{
						Generations:  3,
						Optimization: config.OptimizationLevelConservative,
					},
				},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   config.RiskMedium,
				Status:      config.StatusEnabled,
				Settings: &config.OperationSettings{
					TempFiles: &config.TempFilesSettings{
						OlderThan: "7d",
						Excludes:  []string{"/tmp/keep", "/var/tmp/preserve"},
					},
				},
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean Homebrew",
				RiskLevel:   config.RiskLow,
				Status:      config.StatusEnabled,
				Settings: &config.OperationSettings{
					Homebrew: &config.HomebrewSettings{
						FileSelectionStrategy: config.FileSelectionStrategyUnusedOnly,
						Prune:                 "30d",
					},
				},
			},
			{
				Name:        "system-temp",
				Description: "Clean system temp",
				RiskLevel:   config.RiskMedium,
				Status:      config.StatusEnabled,
				Settings: &config.OperationSettings{
					SystemTemp: &config.SystemTempSettings{
						Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
						OlderThan: "14d",
					},
				},
			},
		},
		Status: config.StatusEnabled,
	}
}

// CreateWeeklyCleanupProfile creates a standard weekly cleanup profile
// Original location: internal/config/test_data.go lines 300-320
func CreateWeeklyCleanupProfile() *config.Profile {
	return &config.Profile{
		Name:        "Weekly Cleanup",
		Description: "Weekly deep cleanup",
		Operations: []config.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Deep Nix cleanup",
				RiskLevel:   config.RiskMedium,
				Status:      config.StatusEnabled,
				Settings: &config.OperationSettings{
					NixGenerations: &config.NixGenerationsSettings{
						Generations:  5,
						Optimization: config.OptimizationLevelConservative,
					},
				},
			},
		},
		Status: config.StatusEnabled,
	}
}
