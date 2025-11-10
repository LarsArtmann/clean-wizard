package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// Load loads the configuration from file or creates default
func Load() (*domain.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads configuration with context support
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/clean-wizard")

	// Set defaults
	v.SetDefault("version", "1.0.0")
	v.SetDefault("safe_mode", false)
	v.SetDefault("max_disk_usage_percent", 50)

	// Try to read configuration file
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, return default config
				return getDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
	}

	// Unmarshal profiles section
	var config domain.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString("version")
	config.SafeMode = v.GetBool("safe_mode")
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected_paths")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal profiles")
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Apply comprehensive validation if available
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(&config)
		if !validationResult.IsValid {
			// Log validation errors but don't fail for backwards compatibility
			for _, err := range validationResult.Errors {
				logrus.WithField("field", err.Field).WithError(fmt.Errorf("%s", err.Message)).Error("Configuration validation warning")
			}
		}
	}

	return &config, nil
}

// Save saves the configuration to file
func Save(config *domain.Config) error {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)

	// Set configuration values
	v.Set("version", config.Version)
	v.Set("safe_mode", config.SafeMode)
	v.Set("max_disk_usage_percent", config.MaxDiskUsage)
	v.Set("protected_paths", config.Protected)
	v.Set("last_clean", config.LastClean)
	v.Set("updated", config.Updated)

	// Set profiles
	for name, profile := range config.Profiles {
		v.Set("profiles."+name+".name", profile.Name)
		v.Set("profiles."+name+".description", profile.Description)
		v.Set("profiles."+name+".enabled", profile.Enabled)

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
		return pkgerrors.HandleConfigError("Save", err)
	}

	// Write configuration file
	if err := v.WriteConfig(); err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	logrus.WithField("config_path", configPath).Info("Configuration saved successfully")
	return nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *domain.Config {
	now := time.Now()

	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true, // Default to safe mode
		MaxDiskUsage: 50,
		Protected: []string{
			"/System",
			"/Applications",
			"/Library",
		},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        "daily",
				Description: "Quick daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"generations": 1, "optimize": false},
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings:    map[string]any{"older_than": "7d", "excludes": []string{"/tmp/keep"}},
					},
				},
				Enabled: true,
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Deep aggressive cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskHigh,
						Enabled:     true,
						Settings:    map[string]any{"generations": 1, "optimize": true},
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    map[string]any{"unused_only": true},
					},
				},
				Enabled: true,
			},
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskCritical,
						Enabled:     true,
						Settings:    map[string]any{"generations": 1, "optimize": true},
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    map[string]any{"unused_only": true, "prune": "old"},
					},
					{
						Name:        "system-temp",
						Description: "Clean system temporary files",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings:    map[string]any{"paths": []string{"/tmp", "/var/tmp"}, "older_than": "30d"},
					},
				},
				Enabled: true,
			},
		},
		LastClean: now,
		Updated:   now,
	}
}
