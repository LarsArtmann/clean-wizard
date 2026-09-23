package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/logger"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	atomicwrite "github.com/larsartmann/go-atomic-write"
	errorfamily "github.com/larsartmann/go-error-family"
	yamlv3 "gopkg.in/yaml.v3"
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
	_ = k.Set("version", CurrentFormatVersion.String())
	_ = k.Set("safe_mode", true)
	_ = k.Set("max_disk_usage_percent", DefaultMaxDiskUsage)
	_ = k.Set("protected", types.DefaultProtectedPaths())

	return k
}

// getConfigPath returns the config file path to use.
func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	return filepath.Join(os.Getenv("HOME"), configName+"."+configType)
}

// DefaultConfigPath returns the configuration file path the loader uses,
// honoring CONFIG_PATH. Commands resolve the file through this so CLI
// behavior and loader behavior can never diverge.
func DefaultConfigPath() string {
	return getConfigPath()
}

// readConfigFileFromPath attempts to read a config file from the given path.
func readConfigFileFromPath(
	ctx context.Context,
	k *koanf.Koanf,
	configPath string,
) (*types.Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err() //nolint:wrapcheck
	default:
		err := k.Load(file.Provider(configPath), yaml.Parser())
		if err != nil {
			if os.IsNotExist(err) {
				return GetDefaultConfig(), nil
			}

			return nil, errorfamily.WrapRejection(err, "config.load", "failed to load configuration from "+configPath)
		}

		return nil, ErrConfigShouldUnmarshal
	}
}

// unmarshalConfig unmarshals koanf config into types.Config, checks the format
// version, and validates it.
func unmarshalConfig(k *koanf.Koanf) (*types.Config, error) {
	config, err := parseConfig(k)
	if err != nil {
		return nil, err
	}

	if err := CheckConfigVersion(config); err != nil {
		return nil, err
	}

	// Validate configuration
	err = validateLoadedConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// parseConfig unmarshals koanf config into types.Config without validating it.
func parseConfig(k *koanf.Koanf) (*types.Config, error) {
	var config types.Config

	// Unmarshal basic fields
	config.Version = k.String("version")
	config.SafeMode = boolToSafeMode(k.Bool("safe_mode"))
	config.MaxDiskUsage = k.Int("max_disk_usage_percent")
	config.Protected = k.Strings("protected")

	// Unmarshal profiles section. koanf's own struct decoder bypasses the
	// enums' UnmarshalYAML hooks (it string-parses int-backed fields), so the
	// raw map is re-encoded as YAML and decoded through yamlv3 instead — the
	// same pattern unmarshalOperationSettings uses for settings. This accepts
	// both symbolic ("enabled") and numeric enum forms.
	profilesKey := "profiles"
	if k.Exists(profilesKey) {
		profilesYAML, encodeErr := yamlv3.Marshal(k.Get(profilesKey))
		if encodeErr != nil {
			logger.Error("Failed to encode profiles", "error", encodeErr)

			return nil, errorfamily.WrapRejection(encodeErr, "config.load", "failed to encode profiles")
		}

		if decodeErr := yamlv3.Unmarshal(profilesYAML, &config.Profiles); decodeErr != nil {
			logger.Error("Failed to unmarshal profiles", "error", decodeErr)

			return nil, errorfamily.WrapRejection(decodeErr, "config.load", "failed to unmarshal profiles")
		}
	}

	// Fix risk levels and settings after unmarshaling
	fixProfileSettings(k, &config)

	return &config, nil
}

// fixProfileSettings fixes risk levels and settings after unmarshaling.
func fixProfileSettings(k *koanf.Koanf, config *types.Config) {
	for name, profile := range config.Profiles {
		for i := range profile.Operations {
			op := &profile.Operations[i]
			op.RiskLevel = parseRiskLevel(k, name, i)
			unmarshalOperationSettings(k, name, i, op)
		}
	}
}

// validateLoadedConfig validates the loaded configuration.
func validateLoadedConfig(config *types.Config) error {
	err := config.Validate()
	if err != nil {
		return errorfamily.WrapRejection(err, "config.load", "configuration validation failed")
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

			if err.Suggestion != "" {
				logger.Error("Configuration validation suggestion",
					"field", err.Field,
					"suggestion", err.Suggestion)
			}
		}

		return fmt.Errorf(
			"configuration validation failed with %d errors",
			len(validationResult.Errors),
		)
	}

	for _, warning := range validationResult.Warnings {
		logger.Warn("Configuration validation warning",
			"field", warning.Field,
			"message", warning.Message,
			"suggestion", warning.Suggestion)
	}

	return nil
}

// Load loads the configuration from file or creates default.
func Load() (*types.Config, error) {
	return LoadWithContext(context.Background())
}

// LoadFromPath loads configuration from a specific file path.
func LoadFromPath(configPath string) (*types.Config, error) {
	return LoadWithContextFromPath(context.Background(), configPath)
}

// LoadWithContext loads configuration with context support.
func LoadWithContext(ctx context.Context) (*types.Config, error) {
	return LoadWithContextFromPath(ctx, getConfigPath())
}

// LoadWithContextFromPath loads configuration from a specific file path with context support.
func LoadWithContextFromPath(ctx context.Context, configPath string) (*types.Config, error) {
	k := setupKoanf()

	config, err := readConfigFileFromPath(ctx, k, configPath)
	if err != nil {
		if errors.Is(err, ErrConfigShouldUnmarshal) {
			return unmarshalConfig(k)
		}

		return nil, err
	}

	return config, nil
}

// boolToSafeMode converts boolean to SafeMode enum.
func boolToSafeMode(b bool) enums.SafeMode {
	return enums.SafeModeFromBool(b)
}

// Save saves the configuration to file.
func Save(config *types.Config) error {
	// Set configuration path
	configPath := filepath.Join(os.Getenv("HOME"), configName+"."+configType)

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)

	err := os.MkdirAll(configDir, ConfigDirPermission)
	if err != nil {
		return errorfamily.WrapRejection(err, "config.save", "failed to create config directory: "+configDir)
	}

	// Marshal to YAML
	yamlData, err := yaml.Parser().Marshal(configYAMLMap(config))
	if err != nil {
		return errorfamily.WrapRejection(err, "config.save", "failed to marshal config to YAML")
	}

	// Write configuration file atomically so a crash never leaves a truncated
	// configuration behind.
	err = atomicwrite.WriteWithPerm(configPath, yamlData, ConfigFilePermission)
	if err != nil {
		return errorfamily.WrapRejection(err, "config.save", "failed to write config file: "+configPath)
	}

	logger.Info("Configuration saved successfully", "config_path", configPath)

	return nil
}

// configYAMLMap renders the configuration as the map written to disk. It is
// shared by Save and the migration engine so both produce identical YAML.
func configYAMLMap(config *types.Config) map[string]any {
	// Build the config map for YAML output
	configMap := map[string]any{
		"version":                config.Version,
		"safe_mode":              config.SafeMode.String(), //nolint:goconst
		"max_disk_usage_percent": config.MaxDiskUsage,
		"protected":              config.Protected,
		"last_clean":             config.LastClean,
		"updated":                config.Updated,
	}

	if config.CurrentProfile != "" {
		configMap["current_profile"] = config.CurrentProfile
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

	return configMap
}

// GetCurrentTime returns current time (helper for testing).
func GetCurrentTime() time.Time {
	return time.Now()
}

// operationRawValue navigates the koanf maps to the raw value of a field inside
// profiles.<name>.operations.<index>.<field>. koanf does not flatten into slices,
// so path-style lookups like "operations.0.settings" never resolve; the traversal
// into the operations array must be done by hand.
func operationRawValue(k *koanf.Koanf, profileName string, operationIndex int, field string) any {
	profiles, ok := k.Get("profiles").(map[string]any)
	if !ok {
		return nil
	}

	profile, ok := profiles[profileName].(map[string]any)
	if !ok {
		return nil
	}

	operations, ok := profile["operations"].([]any)
	if !ok || operationIndex < 0 || operationIndex >= len(operations) {
		return nil
	}

	operation, ok := operations[operationIndex].(map[string]any)
	if !ok {
		return nil
	}

	return operation[field]
}

// parseRiskLevel extracts and converts risk level string from koanf to domain enum.
func parseRiskLevel(k *koanf.Koanf, profileName string, operationIndex int) enums.RiskLevelType {
	switch raw := operationRawValue(k, profileName, operationIndex, "risk_level").(type) {
	case string:
		if raw == "" {
			logger.Warn("No risk level found, defaulting to LOW",
				"profile", profileName,
				"operation", operationIndex)

			return enums.RiskLevelLowType
		}

		switch strings.ToUpper(raw) {
		case "LOW":
			return enums.RiskLevelLowType
		case "MEDIUM":
			return enums.RiskLevelMediumType
		case "HIGH":
			return enums.RiskLevelHighType
		case "CRITICAL":
			return enums.RiskLevelCriticalType
		}
	case int:
		if raw >= 0 && raw <= int(enums.RiskLevelCriticalType) {
			return enums.RiskLevelType(raw)
		}
	}

	logger.Warn("Invalid risk level, defaulting to LOW",
		"profile", profileName,
		"operation", operationIndex)

	return enums.RiskLevelLowType
}

// unmarshalOperationSettings extracts operation settings from koanf and populates the operation.
// The raw settings map is re-encoded as YAML and decoded through the domain types so that
// type-safe enums (which accept both integer and string forms) are parsed by their
// yaml.UnmarshalYAML hooks instead of being silently dropped by koanf's field-name matching.
func unmarshalOperationSettings(
	k *koanf.Koanf,
	profileName string,
	operationIndex int,
	op *types.CleanupOperation,
) {
	raw := operationRawValue(k, profileName, operationIndex, "settings")
	if raw == nil {
		logger.Debug("No settings map found")

		return
	}

	settingsYAML, err := yamlv3.Marshal(raw)
	if err != nil {
		logger.Error("Failed to encode operation settings",
			"error", err,
			"profile", profileName,
			"operation_index", operationIndex)

		return
	}

	settings := &operations.OperationSettings{} //nolint:exhaustruct
	if err := yamlv3.Unmarshal(settingsYAML, settings); err != nil {
		logger.Error("Failed to unmarshal operation settings",
			"error", err,
			"profile", profileName,
			"operation_index", operationIndex)

		return
	}

	op.Settings = settings
}

// newCleanupOperation creates a cleanup operation with the specified parameters.
func newCleanupOperation(
	name, description string, riskLevel enums.RiskLevelType, opType operations.OperationType,
) types.CleanupOperation {
	return types.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   riskLevel,
		Enabled:     enums.ProfileStatusEnabled,
		Settings:    operations.DefaultSettings(opType),
	}
}

// newProfile creates a cleanup profile with the specified name, description, and operations.
func newProfile(name, description string, operations []types.CleanupOperation) *types.Profile {
	return &types.Profile{
		Name:        name,
		Description: description,
		Operations:  operations,
		Enabled:     enums.ProfileStatusEnabled,
	}
}

// GetDefaultConfig returns the default configuration.
func GetDefaultConfig() *types.Config {
	now := GetCurrentTime()

	return &types.Config{ //nolint:exhaustruct
		Version:      CurrentFormatVersion.String(),
		SafeMode:     enums.SafeModeEnabled, // Default to safe mode
		MaxDiskUsage: DefaultMaxDiskUsage,
		Protected: []string{
			"/System",
			"/Applications",
			"/Library",
		},
		Profiles: map[string]*types.Profile{
			"daily": newProfile("daily", "Quick daily cleanup", []types.CleanupOperation{
				newCleanupOperation(
					"nix-generations",
					"Clean old Nix generations",
					enums.RiskLevelLowType,
					operations.OperationTypeNixGenerations,
				),
				newCleanupOperation(
					"temp-files",
					"Clean temporary files",
					enums.RiskLevelLowType,
					operations.OperationTypeTempFiles,
				),
			}),
			"aggressive": newProfile(
				"aggressive",
				"Deep aggressive cleanup",
				[]types.CleanupOperation{
					newCleanupOperation(
						"nix-generations",
						"Clean old Nix generations",
						enums.RiskLevelHighType,
						operations.OperationTypeNixGenerations,
					),
					newCleanupOperation(
						"homebrew-cleanup",
						"Clean old Homebrew packages",
						enums.RiskLevelMediumType,
						operations.OperationTypeHomebrew,
					),
				},
			),
			"comprehensive": newProfile(
				"comprehensive",
				"Complete system cleanup",
				[]types.CleanupOperation{
					newCleanupOperation(
						"nix-generations",
						"Clean old Nix generations",
						enums.RiskLevelCriticalType,
						operations.OperationTypeNixGenerations,
					),
					newCleanupOperation(
						"homebrew-cleanup",
						"Clean old Homebrew packages",
						enums.RiskLevelMediumType,
						operations.OperationTypeHomebrew,
					),
					newCleanupOperation(
						"system-temp",
						"Clean system temporary files",
						enums.RiskLevelMediumType,
						operations.OperationTypeSystemTemp,
					),
				},
			),
		},
		LastClean: now,
		Updated:   now,
	}
}
