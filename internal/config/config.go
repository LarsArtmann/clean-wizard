package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 50,
		Protected:    []string{"/", "/System", "/Library", "/usr", "/etc"},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:         "daily",
				Description:  "Daily cleanup profile",
				MaxRiskLevel: domain.RiskMedium,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
					},
				},
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Deep aggressive cleanup",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Deep Nix generations cleanup",
						RiskLevel:   domain.RiskHigh,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
				},
			},
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskCritical,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
					{
						Name:        "system-temp",
						Description: "Clean system temporary files",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeSystemTemp),
					},
				},
			},
		},
	}
}

// GetTestConfig returns a configuration for testing
func GetTestConfig() *domain.Config {
	return &domain.Config{
		Version:      "test",
		SafeMode:     false,
		MaxDiskUsage: 75,
		Protected:    []string{"/", "/System"},
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:         "test",
				Description:  "Test profile",
				MaxRiskLevel: domain.RiskCritical,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "test-operation",
						Description: "Test operation",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
					},
				},
			},
		},
	}
}

// ValidateConfigStructure validates basic configuration structure
func ValidateConfigStructure(cfg *domain.Config) []string {
	errors := []string{}

	if cfg.Version == "" {
		errors = append(errors, "version is required")
	}

	if len(cfg.Protected) == 0 {
		errors = append(errors, "protected paths cannot be empty")
	}

	if len(cfg.Profiles) == 0 {
		errors = append(errors, "at least one profile is required")
	}

	for name, profile := range cfg.Profiles {
		if profile.Name == "" {
			errors = append(errors, "profile name is required")
		}
		if len(profile.Operations) == 0 {
			errors = append(errors, "profile '"+name+"' must have at least one operation")
		}
		for _, op := range profile.Operations {
			if op.Name == "" {
				errors = append(errors, "operation name is required")
			}
			if op.Settings == nil {
				errors = append(errors, "operation settings are required")
			}
		}
	}

	return errors
}