package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// createCleanupOperation creates a standardized cleanup operation
func createCleanupOperation(name, description string, riskLevel domain.RiskLevel, opType domain.OperationType) domain.CleanupOperation {
	return domain.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   riskLevel,
		Enabled:     true,
		Settings:    domain.DefaultSettings(opType),
	}
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafetyLevelEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/", "/System", "/Library", "/usr", "/etc"},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:         "daily",
				Description:  "Daily cleanup profile",
				MaxRiskLevel: domain.RiskMedium,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					createCleanupOperation("nix-generations", "Clean old Nix generations", domain.RiskLow, domain.OperationTypeNixGenerations),
					createCleanupOperation("temp-files", "Clean temporary files", domain.RiskLow, domain.OperationTypeTempFiles),
				},
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Deep aggressive cleanup",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					createCleanupOperation("nix-generations", "Deep Nix generations cleanup", domain.RiskHigh, domain.OperationTypeNixGenerations),
					createCleanupOperation("homebrew-cleanup", "Clean Homebrew packages", domain.RiskMedium, domain.OperationTypeHomebrew),
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
					createCleanupOperation("system-temp", "Clean system temporary files", domain.RiskMedium, domain.OperationTypeSystemTemp),
				},
			},
		},
	}
}

// GetTestConfig returns a configuration for testing
func GetTestConfig() *domain.Config {
	return &domain.Config{
		Version:      "test",
		SafeMode:     domain.SafetyLevelDisabled,
		MaxDiskUsage: 75,
		Protected:    []string{"/", "/System"},
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:         "test",
				Description:  "Test profile",
				MaxRiskLevel: domain.RiskCritical,
				Enabled:      true,
				Operations: []domain.CleanupOperation{
					createCleanupOperation("test-operation", "Test operation", domain.RiskLow, domain.OperationTypeTempFiles),
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
