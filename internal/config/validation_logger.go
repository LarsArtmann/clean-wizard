package config

import (
	"fmt"
)

// ValidationLogger interface for validation logging
type ValidationLogger interface {
	LogValidation(result *ValidationResult)
	LogSanitization(result *SanitizationResult)
	LogError(field, operation string, err error)
}

// DefaultValidationLogger provides default logging implementation
type DefaultValidationLogger struct {
	enableDetailedLogging bool
}

// NewDefaultValidationLogger creates a default validation logger
func NewDefaultValidationLogger(enableDetailed bool) *DefaultValidationLogger {
	return &DefaultValidationLogger{
		enableDetailedLogging: enableDetailed,
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
	if l.enableDetailedLogging && len(result.Changes) > 0 {
		fmt.Printf("🔧 Configuration sanitized %d fields\n", len(result.Changes))
	}
}

// LogError logs validation errors
func (l *DefaultValidationLogger) LogError(field, operation string, err error) {
	if l.enableDetailedLogging {
		fmt.Printf("❌ Validation error in %s.%s: %v\n", field, operation, err)
	}
}
