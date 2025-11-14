package config

import (
	"time"
)

// Severity represents validation error severity
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// String returns string representation of severity
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Field      string         `json:"field"`
	Rule       string         `json:"rule"`
	Value      any            `json:"value"`
	Message    string         `json:"message"`
	Severity   Severity       `json:"severity"`
	Suggestion string         `json:"suggestion,omitempty"`
	Context    map[string]any `json:"context,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// ValidationWarning represents a specific validation warning
type ValidationWarning struct {
	Field      string         `json:"field"`
	Message    string         `json:"message"`
	Suggestion string         `json:"suggestion,omitempty"`
	Context    map[string]any `json:"context,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// TODO: Create type-safe sanitized data structure - COMPROMISE: Use any for flexibility
type ValidationSanitizedData struct {
	FieldsModified []string          `json:"fields_modified,omitempty"`
	RulesApplied  []string          `json:"rules_applied,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	// COMPROMISE: Preserve flexibility for complex data
	Data          map[string]any    `json:"data,omitempty"`
}

// ValidationResult represents complete validation result with type-safe sanitized data
type ValidationResult struct {
	IsValid   bool                     `json:"is_valid"`
	Errors    []ValidationError           `json:"errors"`
	Warnings  []ValidationWarning         `json:"warnings"`
	Sanitized *ValidationSanitizedData   `json:"sanitized,omitempty"`
	Duration  time.Duration             `json:"duration"`
	Timestamp time.Time                `json:"timestamp"`
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(field, rule string, value any, message, suggestion string, severity Severity) {
	vr.Errors = append(vr.Errors, ValidationError{
		Field:      field,
		Rule:       rule,
		Value:      value,
		Message:    message,
		Severity:   severity,
		Suggestion: suggestion,
		Timestamp:  time.Now(),
	})
	vr.IsValid = false
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(field, message, suggestion string) {
	vr.Warnings = append(vr.Warnings, ValidationWarning{
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
		Timestamp:  time.Now(),
	})
}

// HasErrors returns true if validation has errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if validation has warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// ErrorCount returns the number of errors
func (vr *ValidationResult) ErrorCount() int {
	return len(vr.Errors)
}

// WarningCount returns the number of warnings
func (vr *ValidationResult) WarningCount() int {
	return len(vr.Warnings)
}

// GetErrorsByField returns errors for a specific field
func (vr *ValidationResult) GetErrorsByField(field string) []ValidationError {
	var errors []ValidationError
	for _, err := range vr.Errors {
		if err.Field == field {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetWarningsByField returns warnings for a specific field
func (vr *ValidationResult) GetWarningsByField(field string) []ValidationWarning {
	var warnings []ValidationWarning
	for _, warn := range vr.Warnings {
		if warn.Field == field {
			warnings = append(warnings, warn)
		}
	}
	return warnings
}

// This file moved type definitions to avoid duplicates
// See sanitizer.go for SanitizationResult and SanitizationChange

// HasChanges returns true if sanitization made changes
func (sr *SanitizationResult) HasChanges() bool {
	return len(sr.Changes) > 0
}

// ChangeCount returns the number of changes
func (sr *SanitizationResult) ChangeCount() int {
	return len(sr.Changes)
}
