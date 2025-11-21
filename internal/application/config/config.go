package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/shared/utils/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// Load loads the configuration from file or creates default
func Load() (*config.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads configuration with context support
func LoadWithContext(ctx context.Context) (*config.Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/clean-wizard")

	// Check for CONFIG_PATH environment variable
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.SetConfigFile(configPath)
	}

	// Set defaults using helper function
	setupDefaults(v)

	// Try to read configuration file using helper function
	configExists := true
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := readConfigFile(v); err != nil {
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
		
		// Check if config file was actually found
		if v.ConfigFileUsed() == "" {
			configExists = false
		}
	}

	// If no config exists, create default configuration
	if !configExists {
		log.Info().Msg("Configuration file not found, creating default configuration")
		defaultConfig, err := shared.CreateDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		
		// Save default config for user
		defaultConfigPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)
		if err := SaveConfigToFile(defaultConfig, defaultConfigPath); err != nil {
			log.Warn().Err(err).Msg("Failed to save default configuration")
			// Continue anyway - we have working config in memory
		} else {
			log.Info().Str("path", defaultConfigPath).Msg("Default configuration created successfully")
		}
		
		return defaultConfig, nil
	}

	// Unmarshal profiles section
	var config Config

	// Use common loading and validation logic
	if err := loadAndValidateConfig(&config, v, "LoadWithContext"); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadWithContextAndPath loads configuration with context support and explicit file path
func LoadWithContextAndPath(ctx context.Context, configPath string) (*config.Config, error) {
	v := viper.New()

	// Use the provided path directly
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Fall back to default behavior if no path provided
		return LoadWithContext(ctx)
	}

	// Set defaults
	v.SetDefault(ConfigKeyVersion, "1.0.0")
	v.SetDefault("safe_mode", DefaultSafeMode)
	v.SetDefault(ConfigKeyMaxDiskUsage, DefaultDiskUsagePercent)
	v.SetDefault(ConfigKeyProtected, []string{DefaultProtectedPathSystem, DefaultProtectedPathLibrary}) // Basic protection

	// Try to read configuration file
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, create and return working default config
				log.Info().Msg("Configuration file not found, creating default configuration")
				defaultConfig, err := shared.CreateDefaultConfig()
				if err != nil {
					return nil, fmt.Errorf("failed to create default config: %w", err)
				}
				
				// Save default config for user
				if configPath != "" {
					if err := SaveConfigToFile(defaultConfig, configPath); err != nil {
						log.Warn().Err(err).Msg("Failed to save default configuration")
					} else {
						log.Info().Str("path", configPath).Msg("Default configuration created successfully")
					}
				}
				
				return defaultConfig, nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
		}
	}

	// Unmarshal profiles section
	var config Config

	// Use common loading and validation logic
	if err := loadAndValidateConfig(&config, v, "LoadWithContextAndPath"); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfigToFile saves configuration to a specific file path
func SaveConfigToFile(config *config.Config, configPath string) error {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigFile(configPath)

	// Set configuration values with enum string conversions
	v.Set("version", ConfigVersion)
	v.Set("safety_level", ConfigSafetyLevel.String())
	v.Set("max_disk_usage", ConfigMaxDiskUsage)
	v.Set("protected", ConfigProtected)
	v.Set("current_profile", ConfigCurrentProfile)
	v.Set("last_clean", ConfigLastClean)
	v.Set("updated", ConfigUpdated)
	
	// Convert profiles with proper enum string conversions using interface{} to avoid struct marshaling
	profiles := make(map[string]interface{})
	for name, profile := range ConfigProfiles {
		profileData := make(map[string]interface{})
		profileData["name"] = profile.Name
		profileData["description"] = profile.Description
		profileData["status"] = profile.Status.String()
		
		// Convert operations with proper enum string conversions
		operations := make([]interface{}, len(profile.Operations))
		for i, op := range profile.Operations {
			opData := make(map[string]interface{})
			opData["name"] = op.Name
			opData["description"] = op.Description
			opData["risk_level"] = op.RiskLevel.String()
			opData["status"] = op.Status.String()
			opData["settings"] = op.Settings
			operations[i] = opData
		}
		profileData["operations"] = operations
		profiles[name] = profileData
	}
	v.Set("profiles", profiles)

	// Write configuration file
	return v.WriteConfig()
}

// Save saves the configuration to file
func Save(config *config.Config) error {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)

	// Set configuration values
	v.Set(ConfigKeyVersion, ConfigVersion)
	v.Set("safety_level", ConfigSafetyLevel.String())
	v.Set(ConfigKeyMaxDiskUsage, ConfigMaxDiskUsage)
	v.Set(ConfigKeyProtected, ConfigProtected)
	v.Set(ConfigKeyLastClean, ConfigLastClean)
	v.Set(ConfigKeyUpdated, ConfigUpdated)

	// Set profiles
	for name, profile := range ConfigProfiles {
		v.Set(fmt.Sprintf(ProfileTemplateName, name), profile.Name)
		v.Set(fmt.Sprintf(ProfileTemplateDesc, name), profile.Description)
		v.Set(fmt.Sprintf(ProfileTemplateStatus, name), profile.Status.String())

		for i, op := range profile.Operations {
			v.Set(fmt.Sprintf(OperationTemplateName, name, i), op.Name)
			v.Set(fmt.Sprintf(OperationTemplateDesc, name, i), op.Description)
			v.Set(fmt.Sprintf(OperationTemplateRiskLevel, name, i), op.RiskLevel)
			v.Set(fmt.Sprintf(OperationTemplateStatus, name, i), op.Status.String())
			if op.Settings != nil {
				v.Set(fmt.Sprintf("profiles.%s.operations.%d.settings", name, i), op.Settings)
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

	log.Info().Str("config_path", configPath).Msg("Configuration saved successfully")
	return nil
}

// GetCurrentTime returns current time (helper for testing)
func GetCurrentTime() time.Time {
	return time.Now()
}

// fixConfigTypesAndEnums handles the conversion of string-based config values to proper enums and types

// createDailyProfile creates daily cleanup profile for default config
func createDailyProfile() *shared.Profile {
	return &shared.Profile{
		Name:        "daily",
		Description: "Quick daily cleanup",
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    shared.DefaultSettings(shared.OperationTypeNixGenerations),
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   shared.RiskLow,
				Status:      shared.StatusEnabled,
				Settings:    shared.DefaultSettings(shared.OperationTypeTempFiles),
			},
		},
		Status: shared.StatusEnabled,
	}
}

// createAggressiveProfile creates aggressive cleanup profile for default config
func createAggressiveProfile() *shared.Profile {
	return &shared.Profile{
		Name:        "aggressive",
		Description: "Deep aggressive cleanup",
		Operations: []ConfigCleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   shared.RiskHigh,
				Status:      shared.StatusEnabled,
				Settings:    shared.DefaultSettings(shared.OperationTypeNixGenerations),
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean old Homebrew packages",
				RiskLevel:   shared.RiskMedium,
				Status:      shared.StatusEnabled,
				Settings:    shared.DefaultSettings(shared.OperationTypeHomebrew),
			},
		},
		Status: shared.StatusEnabled,
	}
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *config.Config {
	now := GetCurrentTime()

	return &Config{
		Version:      "1.0.0",
		SafetyLevel:  shared.SafetyLevelEnabled, // Default to safety enabled
		MaxDiskUsage: 50,
		Protected: []string{
			"/System",
			"/Applications",
			"/Library",
		},
		Profiles: map[string]*shared.Profile{
			"daily":      createDailyProfile(),
			"aggressive": createAggressiveProfile(),
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Operations: []ConfigCleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   shared.RiskCritical,
						Status:      shared.StatusEnabled,
						Settings:    shared.DefaultSettings(shared.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   shared.RiskMedium,
						Status:      shared.StatusEnabled,
						Settings:    shared.DefaultSettings(shared.OperationTypeHomebrew),
					},
					{
						Name:        "system-temp",
						Description: "Clean system temporary files",
						RiskLevel:   shared.RiskMedium,
						Status:      shared.StatusEnabled,
						Settings:    shared.DefaultSettings(shared.OperationTypeSystemTemp),
					},
				},
				Status: shared.StatusEnabled,
			},
		},
		LastClean: now,
		Updated:   now,
	}
}

// loadAndValidateConfig performs common config loading, fixing, and validation logic
func loadAndValidateConfig(config *config.Config, v *viper.Viper, functionName string) error {
	// Manually unmarshal fields to avoid YAML tag issues
	ConfigVersion = v.GetString(ConfigKeyVersion)
	// Apply safety configuration using type-safe domain logic with proper dependency inversion
	safetyConfig := shared.ParseSafetyConfig(v)
	ConfigSafetyLevel = safetyConfig.ToSafetyLevel()
	ConfigMaxDiskUsage = v.GetInt(ConfigKeyMaxDiskUsage)
	ConfigProtected = v.GetStringSlice(ConfigKeyProtected)

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &ConfigProfiles); err != nil {
		log.Err(err).Msg("Failed to unmarshal profiles")
		return pkgerrors.HandleConfigError(functionName, err)
	}

	// Fix all configuration issues using consolidated ConfigFixer
	fixer := NewConfigFixer(v)
	fixer.FixAll(config)

	// Apply comprehensive validation with strict enforcement
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(config)
		if !validationResult.IsValid {
			// CRITICAL: Fail fast on validation errors for production safety
			for _, err := range validationResult.Errors {
				log.Error().
					Str("field", err.Field).
					Err(fmt.Errorf("%s", err.Message)).
					Msg("Configuration validation error")
			}
			return fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
		}
	}

	return nil
}
