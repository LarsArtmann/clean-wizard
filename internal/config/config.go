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

	// Set defaults using helper function
	setupDefaults(v)

	// Try to read configuration file using helper function
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if err := readConfigFile(v); err != nil {
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
	}

	// Unmarshal profiles section
	var config domain.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString(ConfigKeyVersion)
	// Apply safety configuration using type-safe domain logic with proper dependency inversion
	safetyConfig := domain.ParseSafetyConfig(v)
	config.SafetyLevel = safetyConfig.ToSafetyLevel()
	config.MaxDiskUsage = v.GetInt(ConfigKeyMaxDiskUsage)
	config.Protected = v.GetStringSlice(ConfigKeyProtected)

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
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled); err != nil {
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile enabled flag")
			}
			if profileEnabled {
				profile.Status = domain.StatusEnabled
			} else {
				profile.Status = domain.StatusDisabled
			}
		} else {
			// Fallback to string status parsing for backward compatibility
			var profileStatusStr string
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr); err != nil {
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile status")
			}
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
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", name, i), &riskLevelStr); err != nil {
				log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal risk level")
			}

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
				if err := v.UnmarshalKey(operationEnabledKey, &operationEnabled); err != nil {
					log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal operation enabled flag")
				}
				if operationEnabled {
					op.Status = domain.StatusEnabled
				} else {
					op.Status = domain.StatusDisabled
				}
			} else {
				// Fallback to string status parsing for backward compatibility
				var opStatusStr string
				if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.status", name, i), &opStatusStr); err != nil {
					log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal operation status")
				}
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
				// Config file not found, return default config
				return GetDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContextAndPath", err)
		}
	}

	// Unmarshal profiles section
	var config domain.Config

	// Manually unmarshal fields to avoid YAML tag issues
	config.Version = v.GetString(ConfigKeyVersion)
	// Apply safety configuration using type-safe domain logic with proper dependency inversion
	safetyConfig := domain.ParseSafetyConfig(v)
	config.SafetyLevel = safetyConfig.ToSafetyLevel()
	config.MaxDiskUsage = v.GetInt(ConfigKeyMaxDiskUsage)
	config.Protected = v.GetStringSlice(ConfigKeyProtected)

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
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled); err != nil {
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile enabled flag")
			}
			if profileEnabled {
				profile.Status = domain.StatusEnabled
			} else {
				profile.Status = domain.StatusDisabled
			}
		} else {
			// Fallback to string status parsing for backward compatibility
			var profileStatusStr string
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr); err != nil {
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile status")
			}
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
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", name, i), &riskLevelStr); err != nil {
				log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal risk level")
			}

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
				if err := v.UnmarshalKey(operationEnabledKey, &operationEnabled); err != nil {
					log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal operation enabled flag")
				}
				if operationEnabled {
					op.Status = domain.StatusEnabled
				} else {
					op.Status = domain.StatusDisabled
				}
			} else {
				// Fallback to string status parsing for backward compatibility
				var opStatusStr string
				if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.status", name, i), &opStatusStr); err != nil {
					log.Warn().Err(err).Str("profile", name).Int("operation", i).Msg("Failed to unmarshal operation status")
				}
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
	v.Set(ConfigKeyVersion, config.Version)
	v.Set("safety_level", config.SafetyLevel.String())
	v.Set(ConfigKeyMaxDiskUsage, config.MaxDiskUsage)
	v.Set(ConfigKeyProtected, config.Protected)
	v.Set(ConfigKeyLastClean, config.LastClean)
	v.Set(ConfigKeyUpdated, config.Updated)

	// Set profiles
	for name, profile := range config.Profiles {
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
