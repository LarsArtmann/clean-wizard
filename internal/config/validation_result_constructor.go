package config

import (
	"time"
)

// NewValidationResult creates a new validation result with safe defaults
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: nil,
		Timestamp: time.Now(),
	}
}