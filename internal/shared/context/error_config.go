package context

import (
	"context"
	"time"
)

// ErrorConfig holds configuration for error handling operations.
type ErrorConfig struct {
	Operation   string
	Field       string
	Value       string
	Expected    string
	Actual      string
	Code        string
	Level       string
	Severity    string
	RetryCount  int
	Metadata    map[string]string
}

// NewErrorConfig creates a new ErrorConfig with default values.
func NewErrorConfig() ErrorConfig {
	return ErrorConfig{
		Metadata: make(map[string]string),
	}
}

// WithOperation sets the operation name.
func (c *ErrorConfig) WithOperation(operation string) *ErrorConfig {
	c.Operation = operation
	return c
}

// WithField sets the field name.
func (c *ErrorConfig) WithField(field string) *ErrorConfig {
	c.Field = field
	return c
}

// WithValue sets the actual value.
func (c *ErrorConfig) WithValue(value string) *ErrorConfig {
	c.Value = value
	return c
}

// WithExpected sets the expected value.
func (c *ErrorConfig) WithExpected(expected string) *ErrorConfig {
	c.Expected = expected
	return c
}

// WithActual sets the actual value (alias for WithValue).
func (c *ErrorConfig) WithActual(actual string) *ErrorConfig {
	c.Actual = actual
	return c
}

// WithCode sets the error code.
func (c *ErrorConfig) WithCode(code string) *ErrorConfig {
	c.Code = code
	return c
}

// WithLevel sets the error level.
func (c *ErrorConfig) WithLevel(level string) *ErrorConfig {
	c.Level = level
	return c
}

// WithSeverity sets the severity.
func (c *ErrorConfig) WithSeverity(severity string) *ErrorConfig {
	c.Severity = severity
	return c
}

// WithRetryCount sets the retry count.
func (c *ErrorConfig) WithRetryCount(count int) *ErrorConfig {
	c.RetryCount = count
	return c
}

// WithMetadata adds metadata.
func (c *ErrorConfig) WithMetadata(key, value string) *ErrorConfig {
	c.Metadata[key] = value
	return c
}

// SanitizationConfig holds configuration for sanitization operations.
type SanitizationConfig struct {
	Operation      string
	Field          string
	Rules          map[string]string
	TrimWhitespace bool
	NormalizeCase  bool
	ClampValues    bool
	Metadata       map[string]string
}

// NewSanitizationConfig creates a new SanitizationConfig with default values.
func NewSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		Rules:    make(map[string]string),
		Metadata: make(map[string]string),
	}
}

// WithOperation sets the operation name.
func (c *SanitizationConfig) WithOperation(operation string) *SanitizationConfig {
	c.Operation = operation
	return c
}

// WithField sets the field name.
func (c *SanitizationConfig) WithField(field string) *SanitizationConfig {
	c.Field = field
	return c
}

// WithRule sets a sanitization rule.
func (c *SanitizationConfig) WithRule(key, value string) *SanitizationConfig {
	c.Rules[key] = value
	return c
}

// WithTrimWhitespace enables whitespace trimming.
func (c *SanitizationConfig) WithTrimWhitespace(enabled bool) *SanitizationConfig {
	c.TrimWhitespace = enabled
	return c
}

// WithNormalizeCase enables case normalization.
func (c *SanitizationConfig) WithNormalizeCase(enabled bool) *SanitizationConfig {
	c.NormalizeCase = enabled
	return c
}

// WithClampValues enables value clamping.
func (c *SanitizationConfig) WithClampValues(enabled bool) *SanitizationConfig {
	c.ClampValues = enabled
	return c
}

// WithMetadata adds metadata.
func (c *SanitizationConfig) WithMetadata(key, value string) *SanitizationConfig {
	c.Metadata[key] = value
	return c
}

// ErrorResult contains the result of an error handling operation.
type ErrorResult struct {
	Handled     bool
	Recovered   bool
	Fatal       bool
	Retryable   bool
	Duration    time.Duration
	Timestamp   time.Time
	Context     *Context[ErrorConfig]
}

// NewErrorResult creates a new ErrorResult.
func NewErrorResult() *ErrorResult {
	return &ErrorResult{
		Timestamp: time.Now(),
	}
}

// SanitizationResultV2 contains the result of a sanitization operation.
type SanitizationResultV2 struct {
	Changed    bool
	Fields     []string
	Warnings   []string
	Duration   time.Duration
	Timestamp  time.Time
	Context    *Context[SanitizationConfig]
}

// NewSanitizationResultV2 creates a new SanitizationResultV2.
func NewSanitizationResultV2() *SanitizationResultV2 {
	return &SanitizationResultV2{
		Timestamp: time.Now(),
	}
}

// LegacyErrorDetails provides backward compatibility with the old ErrorDetails type.
// DEPRECATED: Use Context[ErrorConfig] for new code.
// This type will be removed in v2.0.
type LegacyErrorDetails struct {
	Field      string
	Value      string
	Expected   string
	Actual     string
	Operation  string
	FilePath   string
	LineNumber int
	RetryCount int
	Duration   string
	Metadata   map[string]string
}

// ToLegacyErrorDetails converts a Context[ErrorConfig] to the legacy ErrorDetails format.
func ToLegacyErrorDetails(c *Context[ErrorConfig]) *LegacyErrorDetails {
	return &LegacyErrorDetails{
		Field:      c.ValueType.Field,
		Value:      c.ValueType.Value,
		Expected:   c.ValueType.Expected,
		Actual:     c.ValueType.Actual,
		Operation: c.ValueType.Operation,
		Metadata:  c.Metadata,
	}
}

// FromLegacyErrorDetails creates a Context[ErrorConfig] from the legacy ErrorDetails format.
func FromLegacyErrorDetails(ctx context.Context, legacy *LegacyErrorDetails) *Context[ErrorConfig] {
	if legacy == nil {
		return NewContext(ctx, NewErrorConfig())
	}

	config := ErrorConfig{
		Field:     legacy.Field,
		Value:     legacy.Value,
		Expected:  legacy.Expected,
		Actual:    legacy.Actual,
		Operation: legacy.Operation,
		Metadata:  legacy.Metadata,
	}

	newCtx := NewContext(ctx, config)
	for k, v := range legacy.Metadata {
		newCtx = newCtx.WithMetadata(k, v)
	}
	return newCtx
}

// NewLegacyErrorDetails creates a new LegacyErrorDetails with default values.
// DEPRECATED: Use NewContext[ErrorConfig] for new code.
func NewLegacyErrorDetails() *LegacyErrorDetails {
	return &LegacyErrorDetails{
		Metadata: make(map[string]string),
	}
}