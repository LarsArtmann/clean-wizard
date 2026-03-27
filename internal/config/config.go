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
	"github.com/LarsArtmann/clean-wizard/internal/logger"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ErrConfigShouldUnmarshal is returned when the config file was read successfully
// and should be unmarshaled from koanf.
var ErrConfigShouldUnmarshal = errors.New("config file read successfully, proceed to unmarshal")

const (
	configName = ".clean-wizard"
	configType = "yaml"
)

// setupKoanf creates and configures a koanf instance with defaults.
func setupKoanf() *koanf.Koanf {
	k := koanf.New(".")

	// Set defaults
	_ = k.Set("version", "1.0.0")
	_ = k.Set("safe_mode", true)
	_ = k.Set("max_disk_usage_percent", DefaultMaxDiskUsage)
	_ = k.Set("protected", []string{domain.PathSystem, domain.PathLibrary})

	return k
}

// getConfigPath returns the config file path to use.
func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	return filepath.Join(os.Getenv("HOME"), configName+"."+configType)
}

// readConfigFile attempts to read the config file, returning default config if not found.
func readConfigFile(ctx context.Context, k *koanf.Koanf) (*domain.Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		configPath := getConfigPath()

		err := k.Load(file.Provider(configPath), yaml.Parser())
		if err != nil {
			if os.IsNotExist(err) {
				return GetDefaultConfig(), nil
			}
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}

		return nil, ErrConfigShouldUnmarshal
	}
}

// unmarshalConfig unmarshals koanf config into domain.Config and validates it.
func unmarshalConfig(k *koanf.Koanf) (*domain.Config, error) {
	var config domain.Config

	// Unmarshal basic fields
	config.Version = k.String("version")
	config.SafeMode = boolToSafeMode(k.Bool("safe_mode"))
	config.MaxDiskUsage = k.Int("max_disk_usage_percent")
	config.Protected = k.Strings("protected")

	// Unmarshal profiles section
	profilesKey := "profiles"
	if k.Exists(profilesKey) {
		err := k.Unmarshal(profilesKey, &config.Profiles)
		if err != nil {
			logger.Error("Failed to unmarshal profiles", "error", err)
			return nil, pkgerrors.HandleConfigError("LoadWithContext", err)
		}
	}

	// Fix risk levels and settings after unmarshaling
	fixProfileSettings(k, &config)

	// Validate configuration
	err := validateLoadedConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// fixProfileSettings fixes risk levels and settings after unmarshaling.
func fixProfileSettings(k *koanf.Koanf, config *domain.Config) {
	for name, profile := range config.Profiles {
		for i := range profile.Operations {
			op := &profile.Operations[i]
			op.RiskLevel = parseRiskLevel(k, name, i)
			unmarshalOperationSettings(k, name, i, op)
		}
	}
}

// validateLoadedConfig validates the loaded configuration.
func validateLoadedConfig(config *domain.Config) error {
	err := config.Validate()
	if err != nil {
		return pkgerrors.HandleConfigError("LoadWithContext", err)
	}

	validator := NewConfigValidator()
	if validator == nil {
		return nil
	}

	validationResult := validator.ValidateConfig(config)
	if !validationResult.IsValid {
		for _, err := range validationResult.Errors {
			logger.Error("Configuration validation error",
				"field", err.Field,
				"error", err.Message)
		}

		return fmt.Errorf(
			"configuration validation failed with %d errors",
			len(validationResult.Errors),
		)
	}

	return nil
}

// Load loads the configuration from file or creates default.
func Load() (*domain.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadWithContext loads configuration with context support.
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
	k := setupKoanf()

	// Try to read configuration file
	config, err := readConfigFile(ctx, k)
	if err != nil {
		if errors.Is(err, ErrConfigShouldUnmarshal) {
			// File read successfully, unmarshal and process configuration
			return unmarshalConfig(k)
		}

		return nil, err
	}

	return config, nil
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
	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)

	// Build the config map for YAML output
	configMap := map[string]any{
		"version":                config.Version,
		"safe_mode":              config.SafeMode.String(),
		"max_disk_usage_percent": config.MaxDiskUsage,
		"protected":              config.Protected,
		"last_clean":             config.LastClean,
		"updated":                config.Updated,
	}

	// Build profiles map
	profilesMap := make(map[string]any)
	for name, profile := range config.Profiles {
		profileMap := map[string]any{
			"name":        profile.Name,
			"description": profile.Description,
			"enabled":     profile.Enabled.String(),
		}

		operations := make([]any, len(profile.Operations))
		for i, op := range profile.Operations {
			opMap := map[string]any{
				"name":        op.Name,
				"description": op.Description,
				"risk_level":  op.RiskLevel.String(),
				"enabled":     op.Enabled.String(),
			}
			if op.Settings != nil {
				opMap["settings"] = op.Settings
			}
			operations[i] = opMap
		}
		profileMap["operations"] = operations
		profilesMap[name] = profileMap
	}
	configMap["profiles"] = profilesMap

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)

	err := os.MkdirAll(configDir, ConfigDirPermission)
	if err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	// Marshal to YAML
	yamlData, err := yaml.Parser().Marshal(configMap)
	if err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	// Write configuration file
	err = os.WriteFile(configPath, yamlData, ConfigFilePermission)
	if err != nil {
		return pkgerrors.HandleConfigError("Save", err)
	}

	logger.Info("Configuration saved successfully", "config_path", configPath)

	return nil
}

// GetCurrentTime returns current time (helper for testing).
func GetCurrentTime() time.Time {
	return time.Now()
}

// parseRiskLevel extracts and converts risk level string from koanf to domain enum.
func parseRiskLevel(k *koanf.Koanf, profileName string, operationIndex int) domain.RiskLevelType {
	key := fmt.Sprintf("profiles.%s.operations.%d.risk_level", profileName, operationIndex)

	riskLevelStr := k.String(key)
	if riskLevelStr == "" {
		logger.Warn("No risk level found, defaulting to LOW",
			"profile", profileName,
			"operation", operationIndex)
		return domain.RiskLevelLowType
	}

	switch strings.ToUpper(riskLevelStr) {
	case "LOW":
		return domain.RiskLevelLowType
	case "MEDIUM":
		return domain.RiskLevelMediumType
	case "HIGH":
		return domain.RiskLevelHighType
	case "CRITICAL":
		return domain.RiskLevelCriticalType
	default:
		logger.Warn("Invalid risk level, defaulting to LOW", "risk_level", riskLevelStr)
		return domain.RiskLevelLowType
	}
}

// unmarshalOperationSettings extracts operation settings from koanf and populates the operation.
func unmarshalOperationSettings(
	k *koanf.Koanf,
	profileName string,
	operationIndex int,
	op *domain.CleanupOperation,
) {
	settingsKey := fmt.Sprintf("profiles.%s.operations.%d.settings", profileName, operationIndex)

	if !k.Exists(settingsKey) {
		logger.Debug("No settings map found")
		return
	}

	// Check if nix_generations settings exist
	nixGenKey := settingsKey + ".nix_generations"
	if k.Exists(nixGenKey) {
		nixGenSettings := &domain.NixGenerationsSettings{}
		err := k.Unmarshal(nixGenKey, nixGenSettings)
		if err == nil {
			op.Settings = &domain.OperationSettings{}
			op.Settings.NixGenerations = nixGenSettings
		} else {
			logger.Error("Failed to unmarshal nix_generations settings", "error", err)
		}
	} else {
		logger.Debug("No nix_generations settings found")
	}
}

// newCleanupOperation creates a cleanup operation with the specified parameters.
func newCleanupOperation(
	name, description string, riskLevel domain.RiskLevelType, opType domain.OperationType,
) domain.CleanupOperation {
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
		MaxDiskUsage: DefaultMaxDiskUsage,
		Protected: []string{
			"/System",
			"/Applications",
			"/Library",
		},
		Profiles: map[string]*domain.Profile{
			"daily": newProfile("daily", "Quick daily cleanup", []domain.CleanupOperation{
				newCleanupOperation(
					"nix-generations",
					"Clean old Nix generations",
					domain.RiskLevelLowType,
					domain.OperationTypeNixGenerations,
				),
				newCleanupOperation(
					"temp-files",
					"Clean temporary files",
					domain.RiskLevelLowType,
					domain.OperationTypeTempFiles,
				),
			}),
			"aggressive": newProfile(
				"aggressive",
				"Deep aggressive cleanup",
				[]domain.CleanupOperation{
					newCleanupOperation(
						"nix-generations",
						"Clean old Nix generations",
						domain.RiskLevelHighType,
						domain.OperationTypeNixGenerations,
					),
					newCleanupOperation(
						"homebrew-cleanup",
						"Clean old Homebrew packages",
						domain.RiskLevelMediumType,
						domain.OperationTypeHomebrew,
					),
				},
			),
			"comprehensive": newProfile(
				"comprehensive",
				"Complete system cleanup",
				[]domain.CleanupOperation{
					newCleanupOperation(
						"nix-generations",
						"Clean old Nix generations",
						domain.RiskLevelCriticalType,
						domain.OperationTypeNixGenerations,
					),
					newCleanupOperation(
						"homebrew-cleanup",
						"Clean old Homebrew packages",
						domain.RiskLevelMediumType,
						domain.OperationTypeHomebrew,
					),
					newCleanupOperation(
						"system-temp",
						"Clean system temporary files",
						domain.RiskLevelMediumType,
						domain.OperationTypeSystemTemp,
					),
				},
			),
		},
		LastClean: now,
		Updated:   now,
	}
}
