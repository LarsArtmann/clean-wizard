package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/errors"
	"github.com/LarsArtmann/clean-wizard/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// Load loads the configuration from file or creates default
func Load() (*types.Config, error) {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Add configuration paths
	configPaths := []string{
		"$HOME",
		"$HOME/.config",
		".",
		"/etc/clean-wizard",
	}

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Set defaults
	setDefaults(v)

	// Try to read configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, return default config
			return getDefaultConfig(), nil
		}
		return nil, errors.ConfigLoadError(err)
	}

	// Unmarshal configuration
	var config types.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString("version")
	config.SafeMode = v.GetBool("safe_mode")
	config.DryRun = v.GetBool("dry_run")
	config.Verbose = v.GetBool("verbose")
	config.Backup = v.GetBool("backup")
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected_paths")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal profiles")
		return nil, errors.ConfigLoadError(err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, errors.ConfigValidateError(err.Error())
	}

	return &config, nil
}

// Save saves the configuration to file
func Save(config *types.Config) error {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)
	v.SetConfigFile(configPath)

	// Set all values from config
	v.Set("version", config.Version)
	v.Set("safe_mode", config.SafeMode)
	v.Set("dry_run", config.DryRun)
	v.Set("verbose", config.Verbose)
	v.Set("backup", config.Backup)
	v.Set("max_disk_usage_percent", config.MaxDiskUsage)
	v.Set("protected_paths", config.Protected)

	// Set profiles
	for name, profile := range config.Profiles {
		v.Set("profiles."+name+".name", profile.Name)
		v.Set("profiles."+name+".description", profile.Description)
		v.Set("profiles."+name+".created_at", profile.CreatedAt)
		v.Set("profiles."+name+".updated_at", profile.UpdatedAt)

		for i, op := range profile.Operations {
			opKey := fmt.Sprintf("profiles.%s.operations.%d", name, i)
			v.Set(opKey+".name", op.Name)
			v.Set(opKey+".description", op.Description)
			v.Set(opKey+".risk_level", op.RiskLevel)
			v.Set(opKey+".enabled", op.Enabled)
			if op.Settings != nil {
				v.Set(opKey+".settings", op.Settings)
			}
		}
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.ConfigSaveError(err)
	}

	// Write configuration file
	if err := v.WriteConfig(); err != nil {
		return errors.ConfigSaveError(err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), configName+"."+configType)
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// General settings
	v.SetDefault("version", "dev")
	v.SetDefault("safe_mode", true)
	v.SetDefault("dry_run", true)
	v.SetDefault("verbose", false)
	v.SetDefault("backup", true)
	v.SetDefault("max_disk_usage_percent", 90)

	// Protected paths
	v.SetDefault("protected_paths", []string{
		"/nix/store",
		"/Users",
		"/System",
		"/Applications",
		"/Library",
	})

	// Default profiles
	v.SetDefault("profiles.daily.name", "daily")
	v.SetDefault("profiles.daily.description", "Quick daily cleanup")
	v.SetDefault("profiles.daily.operations", []map[string]any{
		{
			"name":        "nix-generations",
			"description": "Clean old Nix generations",
			"risk_level":  "low",
			"enabled":     true,
			"settings":    map[string]any{"generations": 1, "optimize": false},
		},
		{
			"name":        "homebrew-cleanup",
			"description": "Clean Homebrew caches",
			"risk_level":  "low",
			"enabled":     true,
			"settings":    map[string]any{"autoremove": true, "prune": "recent"},
		},
	})

	v.SetDefault("profiles.comprehensive.name", "comprehensive")
	v.SetDefault("profiles.comprehensive.description", "Complete system cleanup")
	v.SetDefault("profiles.comprehensive.operations", []map[string]any{
		{
			"name":        "nix-store",
			"description": "Clean Nix store and optimize",
			"risk_level":  "low",
			"enabled":     true,
			"settings":    map[string]any{"generations": 1, "optimize": true},
		},
		{
			"name":        "homebrew-full",
			"description": "Full Homebrew cleanup",
			"risk_level":  "low",
			"enabled":     true,
			"settings":    map[string]any{"autoremove": true, "prune": "all"},
		},
		{
			"name":        "package-caches",
			"description": "Clean package manager caches",
			"risk_level":  "low",
			"enabled":     true,
			"settings":    map[string]any{"go": true, "npm": true, "cargo": true},
		},
	})

	v.SetDefault("profiles.aggressive.name", "aggressive")
	v.SetDefault("profiles.aggressive.description", "Nuclear option - everything")
	v.SetDefault("profiles.aggressive.operations", []map[string]any{
		{
			"name":        "nix-nuclear",
			"description": "Remove all Nix generations",
			"risk_level":  "high",
			"enabled":     true,
			"settings":    map[string]any{"generations": 0, "optimize": true},
		},
		{
			"name":        "language-versions",
			"description": "Clean all language version managers",
			"risk_level":  "high",
			"enabled":     true,
			"settings":    map[string]any{"node": true, "python": true, "ruby": true},
		},
		{
			"name":        "all-caches",
			"description": "Clean all caches",
			"risk_level":  "medium",
			"enabled":     true,
			"settings":    map[string]any{"all": true},
		},
	})
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *types.Config {
	now := time.Now()

	return &types.Config{
		Version:      "dev",
		SafeMode:     true,
		DryRun:       true,
		Verbose:      false,
		Backup:       true,
		MaxDiskUsage: 90,
		Protected: []string{
			"/nix/store",
			"/Users",
			"/System",
			"/Applications",
			"/Library",
		},
		Profiles: map[string]types.Profile{
			"daily": {
				Name:        "daily",
				Description: "Quick daily cleanup",
				Operations: []types.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"generations": 1, "optimize": false},
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean Homebrew caches",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"autoremove": true, "prune": "recent"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Operations: []types.CleanupOperation{
					{
						Name:        "nix-store",
						Description: "Clean Nix store and optimize",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"generations": 1, "optimize": true},
					},
					{
						Name:        "homebrew-full",
						Description: "Full Homebrew cleanup",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"autoremove": true, "prune": "all"},
					},
					{
						Name:        "package-caches",
						Description: "Clean package manager caches",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"go": true, "npm": true, "cargo": true},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Nuclear option - everything",
				Operations: []types.CleanupOperation{
					{
						Name:        "nix-nuclear",
						Description: "Remove all Nix generations",
						RiskLevel:   types.RiskHigh,
						Enabled:     true,
						Settings:    map[string]any{"generations": 0, "optimize": true},
					},
					{
						Name:        "language-versions",
						Description: "Clean all language version managers",
						RiskLevel:   types.RiskHigh,
						Enabled:     true,
						Settings:    map[string]any{"node": true, "python": true, "ruby": true},
					},
					{
						Name:        "all-caches",
						Description: "Clean all caches",
						RiskLevel:   types.RiskMedium,
						Enabled:     true,
						Settings:    map[string]any{"all": true},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

// validate validates the configuration
func validate(config *types.Config) error {
	if config.MaxDiskUsage < 0 || config.MaxDiskUsage > 100 {
		return fmt.Errorf("max_disk_usage_percent must be between 0 and 100")
	}

	if len(config.Protected) == 0 {
		return fmt.Errorf("at least one protected path must be specified")
	}

	if len(config.Profiles) == 0 {
		return fmt.Errorf("at least one profile must be specified")
	}

	for name, profile := range config.Profiles {
		if profile.Name == "" {
			return fmt.Errorf("profile %s must have a name", name)
		}
		if profile.Description == "" {
			return fmt.Errorf("profile %s must have a description", name)
		}
		if len(profile.Operations) == 0 {
			return fmt.Errorf("profile %s must have at least one operation", name)
		}
	}

	return nil
}

// IsProtectedPath checks if a path is protected
func IsProtectedPath(path string, protected []string) bool {
	for _, protectedPath := range protected {
		if path == protectedPath || filepath.HasPrefix(path, protectedPath+"/") {
			return true
		}
	}
	return false
}
