package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"github.com/rs/zerolog/log"
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
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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
	// Apply safety configuration using type-safe domain logic
	safetyConfig := domain.ParseSafetyConfig(v)
	config.SafetyLevel = safetyConfig.ToSafetyLevel()
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		log.Err(err).Msg("Failed to unmarshal profiles")
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Fix risk levels after unmarshaling (workaround for custom type unmarshaling)
	for name, profile := range config.Profiles {
		// Convert boolean enabled to StatusType enum for profiles
		var profileEnabled bool
		if v.IsSet(fmt.Sprintf("profiles.%s.enabled", name)) {
			v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled)
			if profileEnabled {
				profile.Status = domain.StatusEnabled
			} else {
				profile.Status = domain.StatusDisabled
			}
		} else {
			// Fallback to string status parsing for backward compatibility
			var profileStatusStr string
			v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr)
			switch strings.ToUpper(strings.TrimSpace(profileStatusStr)) {
			case "DISABLED":
				profile.Status = domain.StatusDisabled
			case "ENABLED":
				profile.Status = domain.StatusEnabled
			case "INHERITED":
				profile.Status = domain.StatusInherited
			default:
				if profileStatusStr != "" {
					log.Warn().Str("profile", name).Str("status", profileStatusStr).Msg("Invalid profile status, defaulting to ENABLED")
				}
				profile.Status = domain.StatusEnabled
			}
		}

		for i, op := range profile.Operations {
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
				log.Warn().Str("risk_level", riskLevelStr).Msg("Invalid risk level, defaulting to LOW")
				op.RiskLevel = domain.RiskLow
			}

			// Convert boolean enabled to StatusType enum for operations
			operationEnabledKey := fmt.Sprintf("profiles.%s.operations.%d.enabled", name, i)
			if v.IsSet(operationEnabledKey) {
				var operationEnabled bool
				v.UnmarshalKey(operationEnabledKey, &operationEnabled)
				if operationEnabled {
					op.Status = domain.StatusEnabled
				} else {
					op.Status = domain.StatusDisabled
				}
			} else {
				// Fallback to string status parsing for backward compatibility
				var opStatusStr string
				v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.status", name, i), &opStatusStr)
				switch strings.ToUpper(strings.TrimSpace(opStatusStr)) {
				case "DISABLED":
					op.Status = domain.StatusDisabled
				case "ENABLED":
					op.Status = domain.StatusEnabled
				case "INHERITED":
					op.Status = domain.StatusInherited
				default:
					if opStatusStr != "" {
						log.Warn().Str("profile", name).Int("operation", i).Str("status", opStatusStr).Msg("Invalid operation status, defaulting to ENABLED")
					}
					op.Status = domain.StatusEnabled
				}
			}
		}
	}

	// Enable comprehensive validation - CRITICAL for production safety
	if err := config.Validate(); err != nil {
		return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	// Apply comprehensive validation with strict enforcement
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(&config)
		if !validationResult.IsValid {
			// CRITICAL: Fail fast on validation errors for production safety
			for _, err := range validationResult.Errors {
				log.Error().
					Str("field", err.Field).
					Err(fmt.Errorf("%s", err.Message)).
					Msg("Configuration validation error")
			}
			return nil, fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
		}
	}

	return &config, nil
}

// LoadWithContextAndPath loads configuration with context support and explicit file path
func LoadWithContextAndPath(ctx context.Context, configPath string) (*domain.Config, error) {
	v := viper.New()

	// Use the provided path directly
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Fall back to default behavior if no path provided
		return LoadWithContext(ctx)
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
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, return default config
				return GetDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
		}
	}

	// Unmarshal profiles section
	var config domain.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString("version")
	// Apply safety configuration using type-safe domain logic
	safetyConfig := domain.ParseSafetyConfig(v)
	config.SafetyLevel = safetyConfig.ToSafetyLevel()
	config.MaxDiskUsage = v.GetInt("max_disk_usage_percent")
	config.Protected = v.GetStringSlice("protected")

	// Unmarshal profiles section
	if err := v.UnmarshalKey("profiles", &config.Profiles); err != nil {
		log.Err(err).Msg("Failed to unmarshal profiles")
		return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
	}

	// Fix risk levels after unmarshaling (workaround for custom type unmarshaling)
	for name, profile := range config.Profiles {
		// Convert boolean enabled to StatusType enum for profiles
		var profileEnabled bool
		if v.IsSet(fmt.Sprintf("profiles.%s.enabled", name)) {
			v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled)
			if profileEnabled {
				profile.Status = domain.StatusEnabled
			} else {
				profile.Status = domain.StatusDisabled
			}
		} else {
			// Fallback to string status parsing for backward compatibility
			var profileStatusStr string
			v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr)
			switch strings.ToUpper(strings.TrimSpace(profileStatusStr)) {
			case "DISABLED":
				profile.Status = domain.StatusDisabled
			case "ENABLED":
				profile.Status = domain.StatusEnabled
			case "INHERITED":
				profile.Status = domain.StatusInherited
			default:
				if profileStatusStr != "" {
					log.Warn().Str("profile", name).Str("status", profileStatusStr).Msg("Invalid profile status, defaulting to ENABLED")
				}
				profile.Status = domain.StatusEnabled
			}
		}

		for i, op := range profile.Operations {
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
				log.Warn().Str("risk_level", riskLevelStr).Msg("Invalid risk level, defaulting to LOW")
				op.RiskLevel = domain.RiskLow
			}

			// Convert boolean enabled to StatusType enum for operations
			operationEnabledKey := fmt.Sprintf("profiles.%s.operations.%d.enabled", name, i)
			if v.IsSet(operationEnabledKey) {
				var operationEnabled bool
				v.UnmarshalKey(operationEnabledKey, &operationEnabled)
				if operationEnabled {
					op.Status = domain.StatusEnabled
				} else {
					op.Status = domain.StatusDisabled
				}
			} else {
				// Fallback to string status parsing for backward compatibility
				var opStatusStr string
				v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.status", name, i), &opStatusStr)
				switch strings.ToUpper(strings.TrimSpace(opStatusStr)) {
				case "DISABLED":
					op.Status = domain.StatusDisabled
				case "ENABLED":
					op.Status = domain.StatusEnabled
				case "INHERITED":
					op.Status = domain.StatusInherited
				default:
					if opStatusStr != "" {
						log.Warn().Str("profile", name).Int("operation", i).Str("status", opStatusStr).Msg("Invalid operation status, defaulting to ENABLED")
					}
					op.Status = domain.StatusEnabled
				}
			}
		}
	}

	// Enable comprehensive validation - CRITICAL for production safety
	if err := config.Validate(); err != nil {
		return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
	}

	// Apply comprehensive validation with strict enforcement
	if validator := NewConfigValidator(); validator != nil {
		validationResult := validator.ValidateConfig(&config)
		if !validationResult.IsValid {
			// CRITICAL: Fail fast on validation errors for production safety
			for _, err := range validationResult.Errors {
				log.Error().
					Str("field", err.Field).
					Err(fmt.Errorf("%s", err.Message)).
					Msg("Configuration validation error")
			}
			return nil, fmt.Errorf("configuration validation failed with %d errors", len(validationResult.Errors))
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
	v.Set("safety_level", config.SafetyLevel.String())
	v.Set("max_disk_usage_percent", config.MaxDiskUsage)
	v.Set("protected", config.Protected)
	v.Set("last_clean", config.LastClean)
	v.Set("updated", config.Updated)

	// Set profiles
	for name, profile := range config.Profiles {
		v.Set("profiles."+name+".name", profile.Name)
		v.Set("profiles."+name+".description", profile.Description)
		v.Set("profiles."+name+".status", profile.Status.String())

		for i, op := range profile.Operations {
			opKey := fmt.Sprintf("profiles.%s.operations.%d", name, i)
			v.Set(opKey+".name", op.Name)
			v.Set(opKey+".description", op.Description)
			v.Set(opKey+".risk_level", op.RiskLevel)
			v.Set(opKey+".enabled", op.Status.String())
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

	log.Info().Str("config_path", configPath).Msg("Configuration saved successfully")
	return nil
}

// GetCurrentTime returns current time (helper for testing)
func GetCurrentTime() time.Time {
	return time.Now()
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *domain.Config {
	now := GetCurrentTime()

	return &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  domain.SafetyLevelEnabled, // Default to safety enabled
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
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "temp-files",
						Description: "Clean temporary files",
						RiskLevel:   domain.RiskLow,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
					},
				},
				Status: domain.StatusEnabled,
			},
			"aggressive": {
				Name:        "aggressive",
				Description: "Deep aggressive cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskHigh,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
				},
				Status: domain.StatusEnabled,
			},
			"comprehensive": {
				Name:        "comprehensive",
				Description: "Complete system cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskCritical,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
					{
						Name:        "homebrew-cleanup",
						Description: "Clean old Homebrew packages",
						RiskLevel:   domain.RiskMedium,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
					},
					{
						Name:        "system-temp",
						Description: "Clean system temporary files",
						RiskLevel:   domain.RiskMedium,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeSystemTemp),
					},
				},
				Status: domain.StatusEnabled,
			},
		},
		LastClean: now,
		Updated:   now,
	}
}
