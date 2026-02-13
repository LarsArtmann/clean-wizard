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

// setupViper creates and configures a viper instance with defaults.
func setupViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/clean-wizard")

	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.SetConfigFile(configPath)
	}

	// Set defaults
	v.SetDefault("version", "1.0.0")
	v.SetDefault("safe_mode", true)
	v.SetDefault("max_disk_usage_percent", 50)
	v.SetDefault("protected", []string{"/System", "/Library"})

	return v
}

// readConfigFile attempts to read the config file, returning default config if not found.
func readConfigFile(ctx context.Context, v *viper.Viper) (*domain.Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				return GetDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
		return nil, nil // File read successfully, continue to unmarshal
	}
}

// unmarshalConfig unmarshals viper config into domain.Config and validates it.
func unmarshalConfig(v *viper.Viper) (*domain.Config, error) {
	var config domain.Config

	// Unmarshal basic fields
	config.Version = v.GetString("version")
	config.SafeMode = boolToSafeMode(v.GetBool("safe_mode"))
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal profiles")
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Fix risk levels and settings after unmarshaling
	fixProfileSettings(v, &config)

	// Validate configuration
	if err := validateLoadedConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// fixProfileSettings fixes risk levels and settings after unmarshaling.
func fixProfileSettings(v *viper.Viper, config *domain.Config) {
	for name, profile := range config.Profiles {
		for i := range profile.Operations {
			op := &profile.Operations[i]
			op.RiskLevel = parseRiskLevel(v, name, i)
			unmarshalOperationSettings(v, name, i, op)
		}
	}
}

// validateLoadedConfig validates the loaded configuration.
func validateLoadedConfig(config *domain.Config) error {
	if err := config.Validate(); err != nil {
		return pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	validator := NewConfigValidator()
	if validator == nil {
		return nil
	}

	validationResult := validator.ValidateConfig(config)
	if !validationResult.IsValid {
		for _, err := range validationResult.Errors {
			logrus.WithField("field", err.Field).WithError(fmt.Errorf("%s", err.Message)).Error("Configuration validation error")
		}
		return fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
	}

	return nil
}

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// Load loads the configuration from file or creates default.
func Load() (*domain.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads configuration with context support.
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
	v := setupViper()

	// Try to read configuration file
	config, err := readConfigFile(ctx, v)
	if err != nil || config != nil {
		return config, err
	}

	// Unmarshal and process configuration
	return unmarshalConfig(v)
}

// boolToSafeMode converts boolean to SafeMode enum.
func boolToSafeMode(b bool) domain.SafeMode {
	if b {
		return domain.SafeModeEnabled
	}
	return domain.SafeModeDisabled
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

// parseRiskLevel extracts and converts risk level string from viper to domain enum.
func parseRiskLevel(v *viper.Viper, profileName string, operationIndex int) domain.RiskLevelType {
	var riskLevelStr string
	key := fmt.Sprintf("profiles.%s.operations.%d.risk_level", profileName, operationIndex)
	if err := v.UnmarshalKey(key, &riskLevelStr); err != nil {
		logrus.WithField("profile", profileName).WithField("operation", operationIndex).WithField("error", err).Warn("Failed to unmarshal risk level, defaulting to LOW")
		return domain.RiskLevelType(domain.RiskLevelLowType)
	}

	switch strings.ToUpper(riskLevelStr) {
	case "LOW":
		return domain.RiskLevelType(domain.RiskLevelLowType)
	case "MEDIUM":
		return domain.RiskLevelType(domain.RiskLevelMediumType)
	case "HIGH":
		return domain.RiskLevelType(domain.RiskLevelHighType)
	case "CRITICAL":
		return domain.RiskLevelType(domain.RiskLevelCriticalType)
	default:
		logrus.WithField("risk_level", riskLevelStr).Warn("Invalid risk level, defaulting to LOW")
		return domain.RiskLevelType(domain.RiskLevelLowType)
	}
}

// unmarshalOperationSettings extracts operation settings from viper and populates the operation.
func unmarshalOperationSettings(v *viper.Viper, profileName string, operationIndex int, op *domain.CleanupOperation) {
	settingsKey := fmt.Sprintf("profiles.%s.operations.%d.settings", profileName, operationIndex)
	settingsMap := v.GetStringMap(settingsKey)

	if len(settingsMap) == 0 {
		logrus.Debug("No settings map found")
		return
	}

	if _, exists := settingsMap["nix_generations"]; exists {
		nixGenSettings := &domain.NixGenerationsSettings{}
		nixGenKey := settingsKey + ".nix_generations"
		if err := v.UnmarshalKey(nixGenKey, nixGenSettings); err == nil {
			op.Settings = &domain.OperationSettings{}
			op.Settings.NixGenerations = nixGenSettings
		} else {
			logrus.WithError(err).Error("Failed to unmarshal nix_generations settings")
		}
	} else {
		logrus.Debug("No nix_generations settings found")
	}
}

// newCleanupOperation creates a cleanup operation with the specified parameters.
func newCleanupOperation(name, description string, riskLevel domain.RiskLevelType, opType domain.OperationType) domain.CleanupOperation {
	return domain.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   riskLevel,
		Enabled:     domain.ProfileStatusEnabled,
		Settings:    domain.DefaultSettings(opType),
	}
}

// newProfile creates a cleanup profile with the specified name, description, and operations.
func newProfile(name, description string, operations []domain.CleanupOperation) *domain.Profile {
	return &domain.Profile{
		Name:        name,
		Description: description,
		Operations:  operations,
		Enabled:     domain.ProfileStatusEnabled,
	}
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
			"daily": newProfile("daily", "Quick daily cleanup", []domain.CleanupOperation{
				newCleanupOperation("nix-generations", "Clean old Nix generations", domain.RiskLevelType(domain.RiskLevelLowType), domain.OperationTypeNixGenerations),
				newCleanupOperation("temp-files", "Clean temporary files", domain.RiskLevelType(domain.RiskLevelLowType), domain.OperationTypeTempFiles),
			}),
			"aggressive": newProfile("aggressive", "Deep aggressive cleanup", []domain.CleanupOperation{
				newCleanupOperation("nix-generations", "Clean old Nix generations", domain.RiskLevelType(domain.RiskLevelHighType), domain.OperationTypeNixGenerations),
				newCleanupOperation("homebrew-cleanup", "Clean old Homebrew packages", domain.RiskLevelType(domain.RiskLevelMediumType), domain.OperationTypeHomebrew),
			}),
			"comprehensive": newProfile("comprehensive", "Complete system cleanup", []domain.CleanupOperation{
				newCleanupOperation("nix-generations", "Clean old Nix generations", domain.RiskLevelType(domain.RiskLevelCriticalType), domain.OperationTypeNixGenerations),
				newCleanupOperation("homebrew-cleanup", "Clean old Homebrew packages", domain.RiskLevelType(domain.RiskLevelMediumType), domain.OperationTypeHomebrew),
				newCleanupOperation("system-temp", "Clean system temporary files", domain.RiskLevelType(domain.RiskLevelMediumType), domain.OperationTypeSystemTemp),
			}),
		},
		LastClean: now,
		Updated:   now,
	}
}
