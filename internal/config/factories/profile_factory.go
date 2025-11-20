package factories

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// createStandardProfile creates a standard daily cleanup profile
// Exported for test compatibility
func CreateStandardProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "daily",
		Description: "Daily cleanup",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
			},
		},
		Status: domain.StatusEnabled,
	}
}

// CreateDailyCleanupProfile creates a standard daily cleanup profile with all operations
// Original location: internal/config/test_data.go lines 230-295
func CreateDailyCleanupProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "Daily Cleanup",
		Description: "Daily system cleanup",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					NixGenerations: &domain.NixGenerationsSettings{
						Generations:  3,
						Optimization: domain.OptimizationLevelConservative,
					},
				},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
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
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					Homebrew: &domain.HomebrewSettings{
						FileSelectionStrategy: domain.FileSelectionStrategyUnusedOnly,
						Prune:                 "30d",
					},
				},
			},
			{
				Name:        "system-temp",
				Description: "Clean system temp",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					SystemTemp: &domain.SystemTempSettings{
						Paths:     []string{"/tmp", "/var/tmp", "/tmp/.font-unix"},
						OlderThan: "14d",
					},
				},
			},
		},
		Status: domain.StatusEnabled,
	}
}

// CreateWeeklyCleanupProfile creates a standard weekly cleanup profile
// Original location: internal/config/test_data.go lines 300-320
func CreateWeeklyCleanupProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "Weekly Cleanup",
		Description: "Weekly deep cleanup",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Deep Nix cleanup",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					NixGenerations: &domain.NixGenerationsSettings{
						Generations:  5,
						Optimization: domain.OptimizationLevelConservative,
					},
				},
			},
		},
		Status: domain.StatusEnabled,
	}
}