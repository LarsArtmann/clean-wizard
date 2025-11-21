package config

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	pkgerrors "github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
)

// ChangeOperation represents type of configuration change with type safety
type ChangeOperation string

const (
	OperationAdded    ChangeOperation = "added"
	OperationRemoved  ChangeOperation = "removed"
	OperationModified ChangeOperation = "modified"
)

// IsValid checks if change operation is valid
func (co ChangeOperation) IsValid() bool {
	switch co {
	case OperationAdded, OperationRemoved, OperationModified:
		return true
	default:
		return false
	}
}

// String returns string representation
func (co ChangeOperation) String() string {
	return string(co)
}

// ValidationMiddleware provides comprehensive validation for configuration operations
type ValidationMiddleware struct {
	validator *ConfigValidator
	sanitizer *ConfigSanitizer
	logger    ValidationLogger
	options   *ValidationMiddlewareOptions
}

// ValidationLogger interface for validation logging
type ValidationLogger interface {
	LogValidation(result *ValidationResult)
	LogSanitization(result *SanitizationResult)
	LogError(field, operation string, err error)
}

// ValidationMiddlewareOptions provides configuration for validation middleware
type ValidationMiddlewareOptions struct {
	RequireSafeModeConfirmation bool   `json:"require_safe_mode_confirmation"`
	EnableDetailedLogging       bool   `json:"enable_detailed_logging"`
	Environment                 string `json:"environment"` // "development", "production", etc.
}

// DefaultValidationLogger provides default logging implementation
type DefaultValidationLogger struct {
	enableDetailedLogging bool
}

// ConfigChangeResult represents configuration change validation result
type ConfigChangeResult struct {
	IsValid   bool                `json:"is_valid"`
	Changes   []ConfigChange      `json:"changes"`
	Errors    []ValidationError   `json:"errors,omitempty"`
	Warnings  []ValidationWarning `json:"warnings,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
}

// ConfigChange represents a single configuration change
type ConfigChange struct {
	Field     string           `json:"field"`
	OldValue  any              `json:"old_value"`
	NewValue  any              `json:"new_value"`
	Operation ChangeOperation  `json:"operation"`
	Risk      domain.RiskLevel `json:"risk"`
}

// ProfileOperationResult represents profile operation validation result
type ProfileOperationResult struct {
	IsValid   bool                     `json:"is_valid"`
	Operation *domain.CleanupOperation `json:"operation,omitempty"`
	Error     error                    `json:"error,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
}

// NewDefaultValidationLogger creates a default validation logger
func NewDefaultValidationLogger(enableDetailed bool) *DefaultValidationLogger {
	return &DefaultValidationLogger{
		enableDetailedLogging: enableDetailed,
	}
}

// WithRequireSafeModeConfirmation enables safe mode confirmation requirement
func WithRequireSafeModeConfirmation(require bool) func(*ValidationMiddlewareOptions) {
	return func(opts *ValidationMiddlewareOptions) {
		opts.RequireSafeModeConfirmation = require
	}
}

// WithEnvironment sets the environment for validation middleware
func WithEnvironment(env string) func(*ValidationMiddlewareOptions) {
	return func(opts *ValidationMiddlewareOptions) {
		opts.Environment = env
	}
}

// WithDetailedLogging enables detailed logging in validation middleware
func WithDetailedLogging(enable bool) func(*ValidationMiddlewareOptions) {
	return func(opts *ValidationMiddlewareOptions) {
		opts.EnableDetailedLogging = enable
	}
}

// LogValidation logs validation results
func (l *DefaultValidationLogger) LogValidation(result *ValidationResult) {
	if l.enableDetailedLogging {
		if result.IsValid {
			fmt.Printf("✅ Configuration validation passed in %v\n", result.Duration)
		} else {
			fmt.Printf("❌ Configuration validation failed with %d errors\n", len(result.Errors))
			for _, err := range result.Errors {
				fmt.Printf("  - %s: %s\n", err.Field, err.Message)
			}
		}
	}
}

// LogSanitization logs sanitization results
func (l *DefaultValidationLogger) LogSanitization(result *SanitizationResult) {
	if l.enableDetailedLogging && len(result.SanitizedFields) > 0 {
		fmt.Printf("🔧 Configuration sanitized %d fields\n", len(result.SanitizedFields))
	}
}

// LogError logs validation errors
func (l *DefaultValidationLogger) LogError(field, operation string, err error) {
	fmt.Printf("⚠️  Validation error in %s.%s: %v\n", operation, field, err)
}

// NewValidationMiddleware creates comprehensive validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return NewValidationMiddlewareWithOptions()
}

// NewValidationMiddlewareWithOptions creates middleware with custom options
func NewValidationMiddlewareWithOptions(options ...func(*ValidationMiddlewareOptions)) *ValidationMiddleware {
	opts := &ValidationMiddlewareOptions{
		RequireSafeModeConfirmation: false,
		EnableDetailedLogging:       true,
		Environment:                 "development",
	}

	for _, opt := range options {
		opt(opts)
	}

	validator := NewConfigValidator()
	logger := NewDefaultValidationLogger(opts.EnableDetailedLogging)

	return &ValidationMiddleware{
		validator: validator,
		sanitizer: NewConfigSanitizer(),
		logger:    logger,
		options:   opts,
	}
}

// NewValidationMiddlewareWithLogger creates middleware with custom logger
func NewValidationMiddlewareWithLogger(logger ValidationLogger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: NewConfigValidator(),
		sanitizer: NewConfigSanitizer(),
		logger:    logger,
		options: &ValidationMiddlewareOptions{
			RequireSafeModeConfirmation: false,
			EnableDetailedLogging:       false,
			Environment:                 "development",
		},
	}
}

// ValidateAndLoadConfig loads and validates configuration with comprehensive checks
func (vm *ValidationMiddleware) ValidateAndLoadConfig(ctx context.Context) (*domain.Config, error) {
	// Load configuration using existing loader
	cfg, err := LoadWithContext(ctx)
	if err != nil {
		return nil, pkgerrors.HandleConfigError("ValidateAndLoadConfig", err)
	}

	// Perform comprehensive validation
	validationResult := vm.validator.ValidateConfig(cfg)
	vm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("ValidateAndLoadConfig",
			fmt.Errorf("validation failed: %s", vm.formatValidationErrors(validationResult.Errors)))
	}

	// Configuration is valid
	return cfg, nil
}

// ValidateAndSaveConfig validates and saves configuration
func (vm *ValidationMiddleware) ValidateAndSaveConfig(ctx context.Context, cfg *domain.Config) (*domain.Config, error) {
	// Perform comprehensive validation first
	validationResult := vm.validator.ValidateConfig(cfg)
	vm.logger.LogValidation(validationResult)

	if !validationResult.IsValid {
		return nil, pkgerrors.HandleValidationError("ValidateAndSaveConfig",
			fmt.Errorf("validation failed: %s", vm.formatValidationErrors(validationResult.Errors)))
	}

	// Save configuration
	if err := Save(cfg); err != nil {
		return nil, pkgerrors.HandleConfigError("ValidateAndSaveConfig", err)
	}

	// Return saved configuration
	return cfg, nil
}

// ValidateConfigChange validates a specific configuration change
func (vm *ValidationMiddleware) ValidateConfigChange(ctx context.Context, current, proposed *domain.Config) *ConfigChangeResult {
	changeResult := &ConfigChangeResult{
		IsValid:   true,
		Changes:   []ConfigChange{},
		Warnings:  []ValidationWarning{},
		Timestamp: time.Now(),
	}

	// Validate proposed configuration
	validationResult := vm.validator.ValidateConfig(proposed)
	changeResult.IsValid = validationResult.IsValid
	changeResult.Warnings = validationResult.Warnings

	if !validationResult.IsValid {
		changeResult.Errors = validationResult.Errors
		return changeResult
	}

	// Analyze changes
	changes := vm.analyzeConfigChanges(current, proposed)
	changeResult.Changes = changes

	// Validate changes against business rules
	if err := vm.validateChangeBusinessRules(changes); err != nil {
		changeResult.IsValid = false
		changeResult.Errors = append(changeResult.Errors, ValidationError{
			Field:    "business_rules",
			Rule:     "validation",
			Value:    err.Error(),
			Message:  err.Error(),
			Severity: SeverityError,
		})
		return changeResult
	}

	return changeResult
}

// ValidateProfileOperation validates a specific profile operation with type safety
func (vm *ValidationMiddleware) ValidateProfileOperation(ctx context.Context, profileName, operationName string, settings *domain.OperationSettings) *ProfileOperationResult {
	result := &ProfileOperationResult{
		IsValid:   true,
		Timestamp: time.Now(),
	}

	// Validate profile name
	if err := vm.validator.validateProfileName(profileName); err != nil {
		result.IsValid = false
		result.Error = err
		return result
	}

	// Create temporary operation for validation
	tempOp := domain.CleanupOperation{
		Name:        operationName,
		Description: "Validation operation",
		RiskLevel:   domain.RiskLow, // Default to low for validation
		Status:      domain.StatusEnabled,
		Settings:    settings,
	}

	// Validate operation
	if err := tempOp.Validate(); err != nil {
		result.IsValid = false
		result.Error = err
		return result
	}

	// Validate operation-specific settings
	if err := vm.validateOperationSettings(operationName, tempOp); err != nil {
		result.IsValid = false
		result.Error = err
		return result
	}

	result.Operation = &tempOp
	return result
}

// formatValidationErrors formats validation errors into a readable string
func (vm *ValidationMiddleware) formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	message := fmt.Sprintf("Validation failed (%d errors):", len(errors))
	for i, err := range errors {
		message += fmt.Sprintf("\n%d. %s: %s", i+1, err.Field, err.Message)
	}
	return message
}
