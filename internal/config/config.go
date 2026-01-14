package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// boolToSafeMode converts boolean to SafeMode enum.
func boolToSafeMode(b bool) domain.SafeMode {
	if b {
		return domain.SafeModeEnabled
	}
	return domain.SafeModeDisabled
}

// Load loads the configuration from file or creates default.
func Load() (*domain.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads configuration with context support.
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/clean-wizard")

	// Check for CONFIG_PATH environment variable
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.SetConfigFile(configPath)
	}

	// Set defaults
	v.SetDefault("version", "1.0.0")
	v.SetDefault("safe_mode", true)
	v.SetDefault("max_disk_usage_percent", 50)
	v.SetDefault("protected", []string{"/System", "/Library"}) // Basic protection

	// Try to read configuration file
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// Config file not found, return default config
				return GetDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
	}

	// Unmarshal profiles section
	var config domain.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString("version")
	config.SafeMode = boolToSafeMode(v.GetBool("safe_mode"))
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal profiles")
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Fix risk levels and settings after unmarshaling (workaround for custom type unmarshaling)
	for name, profile := range config.Profiles {
		for i := range profile.Operations {
			op := &profile.Operations[i] // Get pointer to actual operation

			// Convert string risk level to RiskLevel enum
			var riskLevelStr string
			v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", name, i), &riskLevelStr)

			switch strings.ToUpper(riskLevelStr) {
			case "LOW":
				op.RiskLevel = domain.RiskLow
			case "MEDIUM":
				op.RiskLevel = domain.RiskMedium
			case "HIGH":
				op.RiskLevel = domain.RiskHigh
			case "CRITICAL":
				op.RiskLevel = domain.RiskCritical
			default:
				logrus.WithField("risk_level", riskLevelStr).Warn("Invalid risk level, defaulting to LOW")
				op.RiskLevel = domain.RiskLow
			}

			// Explicitly unmarshal settings for each operation type
			settingsKey := fmt.Sprintf("profiles.%s.operations.%d.settings", name, i)
			settingsMap := v.GetStringMap(settingsKey)

			if len(settingsMap) > 0 {
				// Check for nix_generations settings
				if _, exists := settingsMap["nix_generations"]; exists {
					nixGenSettings := &domain.NixGenerationsSettings{}
					nixGenKey := settingsKey + ".nix_generations"

					if err := v.UnmarshalKey(nixGenKey, nixGenSettings); err == nil {
						// Use direct field assignment to actual operation
						op.Settings = &domain.OperationSettings{}
						op.Settings.NixGenerations = nixGenSettings
					} else {
						logrus.WithError(err).Error("Failed to unmarshal nix_generations settings")
					}
				} else {
					logrus.Debug("No nix_generations settings found")
				}
			} else {
				logrus.Debug("No settings map found")
			}
		}
	}

	// Removed debug logging for production

	// Enable comprehensive validation - CRITICAL for production safety
	if err := config.Validate(); err != nil {
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Removed debug logging for production

	// Apply comprehensive validation with strict enforcement
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(&config)
		if !validationResult.IsValid {
			// CRITICAL: Fail fast on validation errors for production safety
			for _, err := range validationResult.Errors {
				logrus.WithField("field", err.Field).WithError(fmt.Errorf("%s", err.Message)).Error("Configuration validation error")
			}
			return nil, fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
		}
	}

	// Removed debug logging for production

	return &config, nil
}

// Save saves the configuration to file.
func Save(config *domain.Config) error {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)

	// Set configuration values
	v.Set("version", config.Version)
	v.Set("safe_mode", config.SafeMode.String())
	v.Set("max_disk_usage_percent", config.MaxDiskUsage)
	v.Set("protected", config.Protected)
	v.Set("last_clean", config.LastClean)
	v.Set("updated", config.Updated)

	// Set profiles
	for name, profile := range config.Profiles {
		v.Set("profiles."+name+".name", profile.Name)
		v.Set("profiles."+name+".description", profile.Description)
		v.Set("profiles."+name+".enabled", profile.Enabled.String())

		for i, op := range profile.Operations {
			opKey := fmt.Sprintf("profiles.%s.operations.%d", name, i)
			v.Set(opKey+".name", op.Name)
			v.Set(opKey+".description", op.Description)
			v.Set(opKey+".risk_level", op.RiskLevel.String())
			v.Set(opKey+".enabled", op.Enabled.String())
			if op.Settings != nil {
				v.Set(opKey+".settings", op.Settings)
			}
		}
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	// Write configuration file
	if err := v.WriteConfigAs(configPath); err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	logrus.WithField("config_path", configPath).Info("Configuration saved successfully")
	return nil
}

// GetCurrentTime returns current time (helper for testing).
func GetCurrentTime() time.Time {
	return time.Now()
}

// GetDefaultConfig returns the default configuration.
func GetDefaultConfig() *domain.Config {
	now := GetCurrentTime()

	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     domain.SafeModeEnabled, // Default to safe mode
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
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
					},
				},
				Enabled: domain.ProfileStatusEnabled,
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Deep aggressive cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskHigh,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
				},
				Enabled: domain.ProfileStatusEnabled,
			},
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskCritical,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
					{
						Name:        "system-temp",
						Description: "Clean system temporary files",
						RiskLevel:   domain.RiskMedium,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeSystemTemp),
					},
				},
				Enabled: domain.ProfileStatusEnabled,
			},
		},
		LastClean: now,
		Updated:   now,
	}
}
