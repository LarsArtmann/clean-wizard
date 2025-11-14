package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ValidationMiddleware provides comprehensive validation for configuration operations
type ValidationMiddleware struct {
	validator          *ConfigValidator
	sanitizer          *ConfigSanitizer
	logger             ValidationLogger
	operationValidator *OperationValidator
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator:          NewConfigValidator(),
		sanitizer:          NewConfigSanitizer(),
		logger:             NewDefaultValidationLogger(false),
		operationValidator: NewOperationValidator(false),
	}
}

// NewValidationMiddlewareWithLogger creates middleware with custom logger
func NewValidationMiddlewareWithLogger(logger ValidationLogger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator:          NewConfigValidator(),
		sanitizer:          NewConfigSanitizer(),
		logger:             logger,
		operationValidator: NewOperationValidator(false),
	}
}

// Helper function to eliminate duplication
func (vm *ValidationMiddleware) logValidationSuccess(result *SanitizationResult, duration time.Duration) {
	vm.logger.LogValidation(&ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Sanitized: &ValidationSanitizedData{
			Data: map[string]any{
				"sanitized": result.Sanitized,
				"changes":   result.Changes,
			},
		},
		Duration:  duration,
		Timestamp: time.Now(),
	})
}

// ValidateAndLoadConfig validates and loads configuration
func (vm *ValidationMiddleware) ValidateAndLoadConfig(ctx context.Context) (*domain.Config, error) {
	start := time.Now()

	// Load configuration (implementation depends on your storage)
	config, err := vm.loadConfig(ctx)
	if err != nil {
		vm.logger.LogError("config", "load", err)
		return nil, err
	}

	// Validate configuration
	validationResult := vm.validator.ValidateConfig(config)
	vm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		return nil, errors.ValidationError("configuration validation failed", convertValidationErrors(validationResult.Errors))
	}

	// Apply sanitization
	sanitizationResult := vm.sanitizer.SanitizeConfig(config)
	vm.logger.LogSanitization(sanitizationResult)

	// Log success
	duration := time.Since(start)
	if vm.logger.(*DefaultValidationLogger).enableDetailedLogging {
		vm.logValidationSuccess(sanitizationResult, duration)
	}

	return config, nil
}

// ValidateAndSaveConfig validates and saves configuration
func (vm *ValidationMiddleware) ValidateAndSaveConfig(ctx context.Context, cfg *domain.Config) (*domain.Config, error) {
	start := time.Now()

	// Validate configuration
	validationResult := vm.validator.ValidateConfig(cfg)
	vm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		vm.logger.LogError("config", "save",
			errors.ValidationError("configuration validation failed", convertValidationErrors(validationResult.Errors)))
		return nil, errors.ValidationError("configuration validation failed", convertValidationErrors(validationResult.Errors))
	}

	// Apply sanitization
	sanitizationResult := vm.sanitizer.SanitizeConfig(cfg)
	vm.logger.LogSanitization(sanitizationResult)

	// Save configuration
	if err := vm.saveConfig(ctx, cfg); err != nil {
		vm.logger.LogError("config", "save", err)
		return nil, err
	}

	duration := time.Since(start)
	if vm.logger.(*DefaultValidationLogger).enableDetailedLogging {
		vm.logger.LogValidation(&ValidationResult{
			IsValid:  true,
			Errors:   []ValidationError{},
			Warnings: []ValidationWarning{},
			Sanitized: &ValidationSanitizedData{
				Data: map[string]any{
					"sanitized": sanitizationResult.Sanitized,
					"changes":   sanitizationResult.Changes,
				},
			},
			Duration:  duration,
			Timestamp: time.Now(),
		})
	}

	return cfg, nil
}

// ValidateConfigChange validates configuration changes
func (vm *ValidationMiddleware) ValidateConfigChange(ctx context.Context, current, proposed *domain.Config) *ConfigChangeResult {
	result := &ConfigChangeResult{
		IsValid:   true,
		Changes:   []ConfigChange{},
		Errors:    []string{},
		Warnings:  []string{},
		Risk:      "unknown",
		Timestamp: time.Now(),
	}

	// Analyze changes
	changes := vm.analyzeConfigChanges(current, proposed)
	result.Changes = changes

	// Validate changes
	if err := vm.validateChangeBusinessRules(changes); err != nil {
		result.AddError(err.Error())
	}

	// Validate new configuration
	validationResult := vm.validator.ValidateConfig(proposed)
	if !validationResult.IsValid {
		result.AddError("Proposed configuration has validation errors")
	}

	// Assess overall risk
	result.Risk = vm.assessChangeRisk(changes)

	return result
}

// ValidateProfileOperation validates a specific profile operation with type safety
func (vm *ValidationMiddleware) ValidateProfileOperation(ctx context.Context, profileName, operationName string, settings *domain.OperationSettings) *ProfileOperationResult {
	result := &ProfileOperationResult{
		IsValid:     true,
		Operation:   operationName,
		Profile:     profileName,
		Risk:        "unknown",
		Suggestions: []string{},
		Errors:      []string{},
		Timestamp:   time.Now(),
	}

	// Create temporary operation for validation
	operation := &domain.CleanupOperation{
		Name:        operationName,
		Description: fmt.Sprintf("Operation %s in profile %s", operationName, profileName),
		RiskLevel:   domain.RiskMedium, // Default, can be adjusted
		Enabled:     true,
		Settings:    settings,
	}

	// Validate operation settings
	if err := vm.operationValidator.ValidateOperationSettings(operation); err != nil {
		result.AddError(err.Error())
	}

	// Validate operation-specific settings
	if err := vm.validateOperationSettings(operationName, tempOp); err != nil {
		result.IsValid = false
		result.Error = err
		return result
	}

	return result
}

// Helper methods
func (vm *ValidationMiddleware) loadConfig(ctx context.Context) (*domain.Config, error) {
	// CRITICAL FIX: Actually load configuration from file
	configPath := vm.getConfigPath()
	
	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		vm.logger.LogError("config", "read", err)
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}
	
	// Parse YAML
	config := &domain.Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		vm.logger.LogError("config", "parse", err)
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}
	
	// Successfully loaded and parsed
	return config, nil
}

func (vm *ValidationMiddleware) getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	
	// Default config paths to try
	defaultPaths := []string{
		"clean-wizard.yaml",
		"working-config.yaml", 
		"config.yaml",
		"clean-wizard.yml",
		"config.yml",
	}
	
	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// Fallback to working-config.yaml
	return "working-config.yaml"
}

func (vm *ValidationMiddleware) saveConfig(ctx context.Context, cfg *domain.Config) error {
	// This would save to file, database, etc.
	// For now, just return success
	return nil
}

// analyzeConfigChanges analyzes configuration changes
func (vm *ValidationMiddleware) analyzeConfigChanges(current, proposed *domain.Config) []ConfigChange {
	changes := []ConfigChange{}

	// Analyze basic field changes
	if current.Version != proposed.Version {
		changes = append(changes, ConfigChange{
			Field:     "version",
			Operation: "modified",
			OldValue:  current.Version,
			NewValue:  proposed.Version,
			Risk:      "low",
		})
	}

	if current.SafeMode != proposed.SafeMode {
		changes = append(changes, ConfigChange{
			Field:     "safe_mode",
			Operation: "modified",
			OldValue:  current.SafeMode,
			NewValue:  proposed.SafeMode,
			Risk:      "medium",
		})
	}

	if current.MaxDiskUsage != proposed.MaxDiskUsage {
		changes = append(changes, ConfigChange{
			Field:     "max_disk_usage",
			Operation: "modified",
			OldValue:  current.MaxDiskUsage,
			NewValue:  proposed.MaxDiskUsage,
			Risk:      "low",
		})
	}

	// Analyze protected paths changes
	pathChanges := vm.analyzePathChanges("protected", current.Protected, proposed.Protected)
	changes = append(changes, pathChanges...)

	// Analyze profile changes
	profileChanges := vm.analyzeProfileChanges(current.Profiles, proposed.Profiles)
	changes = append(changes, profileChanges...)

	return changes
}

// analyzePathChanges analyzes path array changes
func (vm *ValidationMiddleware) analyzePathChanges(field string, current, proposed []string) []ConfigChange {
	changes := []ConfigChange{}

	currentSet := vm.makeStringSet(current)
	proposedSet := vm.makeStringSet(proposed)

	// Find removed paths
	for _, path := range current {
		if !proposedSet[path] {
			changes = append(changes, ConfigChange{
				Field:     field,
				Operation: "removed",
				OldValue:  path,
				NewValue:  nil,
				Risk:      "medium",
			})
		}
	}

	// Find added paths
	for _, path := range proposed {
		if !currentSet[path] {
			changes = append(changes, ConfigChange{
				Field:     field,
				Operation: "added",
				OldValue:  nil,
				NewValue:  path,
				Risk:      "low",
			})
		}
	}

	return changes
}

// analyzeProfileChanges analyzes profile map changes
func (vm *ValidationMiddleware) analyzeProfileChanges(current, proposed map[string]*domain.Profile) []ConfigChange {
	changes := []ConfigChange{}

	currentSet := vm.makeProfileSet(current)
	proposedSet := vm.makeProfileSet(proposed)

	// Find removed profiles
	for name, profile := range current {
		if !proposedSet[name] {
			changes = append(changes, ConfigChange{
				Field:     fmt.Sprintf("profiles.%s", name),
				Operation: "removed",
				OldValue:  profile,
				NewValue:  nil,
				Risk:      "medium",
			})
		}
	}

	// Find added profiles
	for name, profile := range proposed {
		if !currentSet[name] {
			changes = append(changes, ConfigChange{
				Field:     fmt.Sprintf("profiles.%s", name),
				Operation: "added",
				OldValue:  nil,
				NewValue:  profile,
				Risk:      "low",
			})
		} else {
			// Analyze profile modifications
			currentProfile := current[name]
			if currentProfile.Name != profile.Name {
				changes = append(changes, ConfigChange{
					Field:     fmt.Sprintf("profiles.%s.name", name),
					Operation: "modified",
					OldValue:  currentProfile.Name,
					NewValue:  profile.Name,
					Risk:      "low",
				})
			}
			if currentProfile.Enabled != profile.Enabled {
				changes = append(changes, ConfigChange{
					Field:     fmt.Sprintf("profiles.%s.enabled", name),
					Operation: "modified",
					OldValue:  currentProfile.Enabled,
					NewValue:  profile.Enabled,
					Risk:      "medium",
				})
			}
		}
	}

	return changes
}

// validateChangeBusinessRules validates business rules for changes
func (vm *ValidationMiddleware) validateChangeBusinessRules(changes []ConfigChange) error {
	for _, change := range changes {
		if change.Field == "safe_mode" && change.NewValue == false {
			return fmt.Errorf("cannot disable safe mode - security policy violation")
		}
		if change.Field == "protected" && change.Operation == "removed" {
			return fmt.Errorf("cannot remove protected paths - security policy violation")
		}
	}
	return nil
}

<<<<<<< HEAD
// assessChangeRisk assesses overall change risk
func (vm *ValidationMiddleware) assessChangeRisk(changes []ConfigChange) string {
	highRiskCount := 0
	mediumRiskCount := 0
=======
// validateOperationSettings validates operation-specific settings with type safety
func (vm *ValidationMiddleware) validateOperationSettings(operationName string, op domain.CleanupOperation) error {
	// Use the already-validated settings from the operation
	if op.Settings == nil {
		return nil // Settings are optional
	}
	
	opType := domain.GetOperationType(operationName)
	return op.Settings.ValidateSettings(opType)
}
>>>>>>> master

	for _, change := range changes {
		switch change.Risk {
		case "high":
			highRiskCount++
		case "medium":
			mediumRiskCount++
		}
	}

	if highRiskCount > 0 {
		return "high"
	}
	if mediumRiskCount > 2 {
		return "medium"
	}
	return "low"
}

// assessOperationRisk assesses operation risk
func (vm *ValidationMiddleware) assessOperationRisk(op *domain.CleanupOperation) string {
	// Use operation's risk level
	return string(op.RiskLevel)
}

// assessProfileRisk assesses profile risk
func (vm *ValidationMiddleware) assessProfileRisk(profile *domain.Profile) string {
	maxRisk := domain.RiskLow
	for _, op := range profile.Operations {
		if op.RiskLevel.IsHigherThan(maxRisk) {
			maxRisk = op.RiskLevel
		}
	}
	return string(maxRisk)
}

// Helper methods for making sets
func (vm *ValidationMiddleware) makeStringSet(slice []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range slice {
		set[item] = true
	}
	return set
}

func (vm *ValidationMiddleware) makeProfileSet(profiles map[string]*domain.Profile) map[string]bool {
	set := make(map[string]bool)
	for name := range profiles {
		set[name] = true
	}
	return set
}

// formatValidationErrors formats validation errors for display
func (vm *ValidationMiddleware) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	result := "Validation errors:\n"
	for _, err := range errors {
		result += fmt.Sprintf("  - %s: %s\n", err.Field, err.Message)
	}
	return result
}

// getChangeOperation determines the change operation
func (vm *ValidationMiddleware) getChangeOperation(old, new any) string {
	if old == nil {
		return "added"
	}
	if new == nil {
		return "removed"
	}
	return "modified"
}

// loadConfigWithValidation loads configuration with validation (needed for enhanced_loader)
func (vm *ValidationMiddleware) loadConfigWithValidation(ctx context.Context) (*domain.Config, error) {
	return vm.ValidateAndLoadConfig(ctx)
}

// saveConfig saves configuration (needed for enhanced_loader)
// convertValidationErrors converts []ValidationError to []any
func convertValidationErrors(validationErrors []ValidationError) []any {
	converted := make([]any, len(validationErrors))
	for i, err := range validationErrors {
		converted[i] = err
	}
	return converted
}
