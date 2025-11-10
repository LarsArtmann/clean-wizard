package config

import (
	"time"

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
						Name:        "nix-store-cleanup",
						Description: "Clean Nix store",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							Type: domain.OperationTypeNixStore,
							NixStore: &domain.NixStoreSettings{
								KeepGenerations: 3,
								MinAge:          24 * time.Hour,
								IncludeProfiles: false,
								DryRun:          false,
							},
						},
					},
					{
						Name:        "temp-files-cleanup",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							Type: domain.OperationTypeTempFiles,
							TempFiles: &domain.TempFilesSettings{
								MaxAge:       7 * 24 * time.Hour,
								Paths:        []string{"/tmp", "/var/tmp"},
								ExcludePaths: []string{"/tmp/keep"},
								DryRun:       false,
							},
						},
					},
				},
			},
			"aggressive": {
				Name:         "aggressive",
				Description:  "Deep aggressive cleanup",
				MaxRiskLevel: domain.RiskHigh,
				Enabled:      false,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-store-deep-cleanup",
						Description: "Deep Nix store cleanup",
						RiskLevel:   domain.RiskHigh,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							Type: domain.OperationTypeNixStore,
							NixStore: &domain.NixStoreSettings{
								KeepGenerations: 1,
								MinAge:          12 * time.Hour,
								IncludeProfiles: true,
								DryRun:          false,
							},
						},
					},
					{
						Name:        "package-cache-cleanup",
						Description: "Clean package caches",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							Type: domain.OperationTypePackageCache,
							Package: &domain.PackageCacheSettings{
								MaxAge:       24 * time.Hour,
								MaxSize:      1024 * 1024 * 1024, // 1GB
								IncludeTypes: []string{"brew", "npm", "cargo"},
								DryRun:       false,
							},
						},
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
						Settings: &domain.OperationSettings{
							Type: domain.OperationTypeTempFiles,
							TempFiles: &domain.TempFilesSettings{
								MaxAge: 1 * time.Hour,
								Paths:  []string{"/tmp/test"},
								DryRun: true,
							},
						},
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
