package factories

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// createStandardProfile creates a standard daily cleanup profile
// Exported for test compatibility
func CreateStandardProfile() *ConfigProfile {
	return &ConfigProfile{
		Name:        "Standard Cleanup",
		Description: "Standard profile for daily cleanup",
		Status:      shared.StatusEnabled,
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{NixGenerations: &shared.NixGenerationsSettings{Generations: 5}},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files from user directories",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{TempFiles: &shared.TempFilesSettings{OlderThan: "7d"}},
			},
		},
	}
}

// CreateDailyCleanupProfile creates a profile for daily safe cleanup
func CreateDailyCleanupProfile() *ConfigProfile {
	return &ConfigProfile{
		Name:        "Daily Safe Cleanup",
		Description: "Safe daily cleanup operations only",
		Status:      shared.StatusEnabled,
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{NixGenerations: &shared.NixGenerationsSettings{Generations: 5}},
			},
			{
				Name:        "temp-files", 
				Description: "Clean temporary files from user directories",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{TempFiles: &shared.TempFilesSettings{OlderThan: "7d"}},
			},
		},
	}
}

// CreateWeeklyCleanupProfile creates a profile for weekly more aggressive cleanup
func CreateWeeklyCleanupProfile() *ConfigProfile {
	return &ConfigProfile{
		Name:        "Weekly Cleanup",
		Description: "Weekly cleanup with moderate risk operations",
		Status:      shared.StatusEnabled,
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{NixGenerations: &shared.NixGenerationsSettings{Generations: 3}},
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files from user directories", 
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    &shared.OperationSettings{TempFiles: &shared.TempFilesSettings{OlderThan: "3d"}},
			},
		},
	}
}
