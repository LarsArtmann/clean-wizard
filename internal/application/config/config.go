package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/shared/utils/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// Load loads configuration from file or creates default
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
		defaultConfig, err := config.CreateDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		
		// Save default config for user
		defaultConfigPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)
		if err := SaveConfigToFile(defaultConfig, defaultConfigPath); err != nil {
			log.Warn().Err(err).Msg("Failed to save default configuration")
		}
		
		return defaultConfig, nil
	}

	// Load and validate the configuration
	if err := loadAndValidateConfig(v); err != nil {
		return nil, fmt.Errorf("failed to load and validate config: %w", err)
	}

	// Map viper data to domain config
	domainConfig, err := mapViperToDomainConfig(v)
	if err != nil {
		return nil, fmt.Errorf("failed to map config: %w", err)
	}

	return domainConfig, nil
}

// LoadWithContextAndPath loads configuration with context support and explicit file path
func LoadWithContextAndPath(ctx context.Context, configPath string) (*config.Config, error) {
	v := viper.New()

	// Use provided path directly
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Fallback to default behavior
		return LoadWithContext(ctx)
	}

	// Set defaults
	setupDefaults(v)

	// Try to read the configuration file
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := v.ReadInConfig(); err != nil {
			return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
		}
	}

	// Load and validate the configuration
	if err := loadAndValidateConfig(v); err != nil {
		return nil, fmt.Errorf("failed to load and validate config: %w", err)
	}

	// Map viper data to domain config
	domainConfig, err := mapViperToDomainConfig(v)
	if err != nil {
		return nil, fmt.Errorf("failed to map config: %w", err)
	}

	return domainConfig, nil
}

// SaveConfigToFile saves configuration to specified file path
func SaveConfigToFile(cfg *config.Config, configPath string) error {
	if cfg == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Use viper to write YAML
	v := viper.New()
	
	// Map domain config to viper
	if err := mapDomainConfigToViper(cfg, v); err != nil {
		return fmt.Errorf("failed to map config for saving: %w", err)
	}

	v.SetConfigFile(configPath)
	return v.WriteConfig()
}

// Save saves configuration to default location
func Save(cfg *config.Config) error {
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)
	return SaveConfigToFile(cfg, configPath)
}

// GetDefaultConfig returns a working default configuration
func GetDefaultConfig() *config.Config {
	defaultConfig, err := config.CreateDefaultConfig()
	if err != nil {
		log.Err(err).Msg("Failed to create default configuration, using minimal config")
		return &config.Config{
			Version:     "1.0.0",
			SafetyLevel: shared.SafetyLevelEnabled,
			Profiles:    make(map[string]*config.Profile),
		}
	}
	return defaultConfig
}

// Helper functions

func setupDefaults(v *viper.Viper) {
	v.SetDefault("version", "1.0.0")
	v.SetDefault("safety_level", "enabled")
	v.SetDefault("max_disk_usage", 50)
	v.SetDefault("protected", []string{
		"/System",
		"/Applications",
		"/Library",
		"/usr/local",
		"/Users/*/Documents",
		"/Users/*/Desktop",
	})
	v.SetDefault("current_profile", "daily")
}

func readConfigFile(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil // Config file not found is not an error
		}
		return err
	}
	return nil
}

func loadAndValidateConfig(v *viper.Viper) error {
	// Basic validation of required fields
	if v.GetString("version") == "" {
		return fmt.Errorf("version is required")
	}

	if v.GetInt("max_disk_usage") < 0 || v.GetInt("max_disk_usage") > 100 {
		return fmt.Errorf("max_disk_usage must be between 0 and 100")
	}

	protected := v.GetStringSlice("protected")
	if len(protected) == 0 {
		return fmt.Errorf("protected paths cannot be empty")
	}

	profiles := v.GetStringMapString("profiles")
	if len(profiles) == 0 {
		return fmt.Errorf("at least one profile is required")
	}

	return nil
}

func mapViperToDomainConfig(v *viper.Viper) (*config.Config, error) {
	domainCfg := &config.Config{
		Version:        v.GetString("version"),
		SafetyLevel:    parseSafetyLevel(v.GetString("safety_level")),
		MaxDiskUsage:   v.GetInt("max_disk_usage"),
		Protected:      v.GetStringSlice("protected"),
		CurrentProfile: v.GetString("current_profile"),
		LastClean:      v.GetTime("last_clean"),
		Updated:        v.GetTime("updated"),
		Profiles:       make(map[string]*config.Profile),
	}

	// Map profiles
	profiles := v.GetStringMap("profiles")
	for name, profileData := range profiles {
		profileMap, ok := profileData.(map[string]interface{})
		if !ok {
			continue
		}

		profile := &config.Profile{
			Name:        name,
			Description: getString(profileMap, "description"),
			Status:      parseStatus(getString(profileMap, "status")),
			Operations:  []config.CleanupOperation{},
		}

		// Map operations
		if opsData, exists := profileMap["operations"]; exists {
			if opsList, ok := opsData.([]interface{}); ok {
				for _, opData := range opsList {
					if opMap, ok := opData.(map[string]interface{}); ok {
						op := config.CleanupOperation{
							Name:        getString(opMap, "name"),
							Description: getString(opMap, "description"),
							RiskLevel:   parseRiskLevel(getString(opMap, "risk_level")),
							Status:      parseStatus(getString(opMap, "status")),
						}
						profile.Operations = append(profile.Operations, op)
					}
				}
			}
		}

		domainCfg.Profiles[name] = profile
	}

	return domainCfg, nil
}

func mapDomainConfigToViper(cfg *config.Config, v *viper.Viper) error {
	v.Set("version", cfg.Version)
	v.Set("safety_level", cfg.SafetyLevel.String())
	v.Set("max_disk_usage", cfg.MaxDiskUsage)
	v.Set("protected", cfg.Protected)
	v.Set("current_profile", cfg.CurrentProfile)
	v.Set("last_clean", cfg.LastClean)
	v.Set("updated", cfg.Updated)

	// Map profiles
	v.Set("profiles", cfg.Profiles)

	return nil
}

// Helper parsing functions
func parseSafetyLevel(level string) shared.SafetyLevelType {
	switch level {
	case "enabled":
		return shared.SafetyLevelEnabled
	case "disabled":
		return shared.SafetyLevelDisabled
	default:
		return shared.SafetyLevelEnabled
	}
}

func parseStatus(status string) shared.StatusType {
	switch status {
	case "enabled":
		return shared.StatusEnabled
	case "disabled":
		return shared.StatusDisabled
	default:
		return shared.StatusEnabled
	}
}

func parseRiskLevel(risk string) shared.RiskLevelType {
	switch risk {
	case "low":
		return shared.RiskLevelLowType
	case "medium":
		return shared.RiskLevelMediumType
	case "high":
		return shared.RiskLevelHighType
	case "critical":
		return shared.RiskLevelCriticalType
	default:
		return shared.RiskLevelLowType
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}