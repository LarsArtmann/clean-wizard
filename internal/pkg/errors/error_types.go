package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorDetails represents strongly-typed error context information.
type ErrorDetails struct {
	Field      string            `json:"field,omitempty"`
	Value      string            `json:"value,omitempty"`
	Expected   string            `json:"expected,omitempty"`
	Actual     string            `json:"actual,omitempty"`
	Operation  string            `json:"operation,omitempty"`
	FilePath   string            `json:"file_path,omitempty"`
	LineNumber int               `json:"line_number,omitempty"`
	RetryCount int               `json:"retry_count,omitempty"`
	Duration   string            `json:"duration,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// CleanWizardError represents structured error with context.
type CleanWizardError struct {
	Code      ErrorCode
	Level     ErrorLevel
	Message   string
	Operation string
	Details   *ErrorDetails `json:"details,omitempty"`
	Timestamp time.Time
	Stack     string
}

// formatErrorDetails formats the error details as a string.
func formatErrorDetails(details *ErrorDetails) string {
	if details == nil {
		return ""
	}

	var parts []string

	addDetail := func(name, value string) {
		if value != "" {
			parts = append(parts, name+"="+value)
		}
	}

	addDetail("field", details.Field)
	addDetail("value", details.Value)
	addDetail("expected", details.Expected)
	addDetail("actual", details.Actual)
	addDetail("operation", details.Operation)
	addDetail("file", details.FilePath)

	if details.LineNumber > 0 {
		parts = append(parts, fmt.Sprintf("line=%d", details.LineNumber))
	}
	if details.RetryCount > 0 {
		parts = append(parts, fmt.Sprintf("retry=%d", details.RetryCount))
	}
	addDetail("duration", details.Duration)

	// Add metadata
	for key, value := range details.Metadata {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf(" (%s)", strings.Join(parts, " "))
}

// Error implements error interface.
func (e *CleanWizardError) Error() string {
	base := fmt.Sprintf("[%s] %s: %s", e.Level.String(), e.Code.String(), e.Message)
	return base + formatErrorDetails(e.Details)
}

// Unwrap returns the underlying error for compatibility with Go 1.13+ error wrapping.
func (e *CleanWizardError) Unwrap() error {
	if e.Details != nil && e.Details.Value != "" {
		return fmt.Errorf("wrapped error: %s", e.Details.Value)
	}

	return nil
}
