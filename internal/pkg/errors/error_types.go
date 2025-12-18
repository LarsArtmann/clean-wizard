package errors

import (
	"fmt"
	"time"
)

// ErrorDetails represents strongly-typed error context information
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

// CleanWizardError represents structured error with context
type CleanWizardError struct {
	Code      ErrorCode
	Level     ErrorLevel
	Message   string
	Operation string
	Details   *ErrorDetails `json:"details,omitempty"`
	Timestamp time.Time
	Stack     string
}

// Error implements error interface
func (e *CleanWizardError) Error() string {
	if e.Details == nil {
		return fmt.Sprintf("[%s] %s: %s", e.Level.String(), e.Code.String(), e.Message)
	}

	var details []string
	if e.Details.Field != "" {
		details = append(details, fmt.Sprintf("field=%s", e.Details.Field))
	}
	if e.Details.Value != "" {
		details = append(details, fmt.Sprintf("value=%s", e.Details.Value))
	}
	if e.Details.Expected != "" {
		details = append(details, fmt.Sprintf("expected=%s", e.Details.Expected))
	}
	if e.Details.Actual != "" {
		details = append(details, fmt.Sprintf("actual=%s", e.Details.Actual))
	}
	if e.Details.Operation != "" {
		details = append(details, fmt.Sprintf("operation=%s", e.Details.Operation))
	}
	if e.Details.FilePath != "" {
		details = append(details, fmt.Sprintf("file=%s", e.Details.FilePath))
	}
	if e.Details.LineNumber > 0 {
		details = append(details, fmt.Sprintf("line=%d", e.Details.LineNumber))
	}
	if e.Details.RetryCount > 0 {
		details = append(details, fmt.Sprintf("retry=%d", e.Details.RetryCount))
	}
	if e.Details.Duration != "" {
		details = append(details, fmt.Sprintf("duration=%s", e.Details.Duration))
	}

	// Add metadata
	for key, value := range e.Details.Metadata {
		details = append(details, fmt.Sprintf("%s=%s", key, value))
	}

	detailStr := ""
	if len(details) > 0 {
		detailStr = fmt.Sprintf(" (%s)", fmt.Sprintf("%s", details))
	}

	return fmt.Sprintf("[%s] %s: %s%s", e.Level.String(), e.Code.String(), e.Message, detailStr)
}

// Unwrap returns the underlying error for compatibility with Go 1.13+ error wrapping
func (e *CleanWizardError) Unwrap() error {
	if e.Details != nil && e.Details.Value != "" {
		return fmt.Errorf("wrapped error: %s", e.Details.Value)
	}
	return nil
}
