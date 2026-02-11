package context

import (
	"context"
	"time"
)

// ValidationConfig holds configuration for validation operations.
type ValidationConfig struct {
	ConfigPath      string
	ValidationLevel string
	Profile         string
	Section         string
	MinValue        any
	MaxValue        any
	AllowedValues   []string
	ReferencedField string
	Constraints     map[string]string
	Metadata        map[string]string
}

// NewValidationConfig creates a new ValidationConfig with default values.
func NewValidationConfig() ValidationConfig {
	return ValidationConfig{
		Constraints:   make(map[string]string),
		Metadata:      make(map[string]string),
		AllowedValues: []string{},
	}
}

// WithConfigPath sets the config path.
func (c *ValidationConfig) WithConfigPath(path string) *ValidationConfig {
	c.ConfigPath = path
	return c
}

// WithValidationLevel sets the validation level.
func (c *ValidationConfig) WithValidationLevel(level string) *ValidationConfig {
	c.ValidationLevel = level
	return c
}

// WithProfile sets the profile.
func (c *ValidationConfig) WithProfile(profile string) *ValidationConfig {
	c.Profile = profile
	return c
}

// WithSection sets the section.
func (c *ValidationConfig) WithSection(section string) *ValidationConfig {
	c.Section = section
	return c
}

// WithMinValue sets the minimum value.
func (c *ValidationConfig) WithMinValue(value any) *ValidationConfig {
	c.MinValue = value
	return c
}

// WithMaxValue sets the maximum value.
func (c *ValidationConfig) WithMaxValue(value any) *ValidationConfig {
	c.MaxValue = value
	return c
}

// WithAllowedValues sets the allowed values.
func (c *ValidationConfig) WithAllowedValues(values ...string) *ValidationConfig {
	c.AllowedValues = values
	return c
}

// WithReferencedField sets the referenced field.
func (c *ValidationConfig) WithReferencedField(field string) *ValidationConfig {
	c.ReferencedField = field
	return c
}

// WithConstraints sets the constraints.
func (c *ValidationConfig) WithConstraints(constraints map[string]string) *ValidationConfig {
	c.Constraints = constraints
	return c
}

// WithMetadata adds metadata.
func (c *ValidationConfig) WithMetadata(key, value string) *ValidationConfig {
	c.Metadata[key] = value
	return c
}

// ValidationResult contains the result of a validation operation.
type ValidationResult struct {
	IsValid   bool
	Errors    []ValidationError
	Warnings  []ValidationWarning
	Duration  time.Duration
	Timestamp time.Time
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Rule    string
	Value   string
	Message string
	Context *Context[ValidationConfig]
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Field   string
	Message string
	Context *Context[ValidationConfig]
}

// NewValidationResult creates a new ValidationResult.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}
}

// AddError adds an error to the result.
func (r *ValidationResult) AddError(field, rule, value, message string, ctx *Context[ValidationConfig]) *ValidationResult {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Rule:    rule,
		Value:   value,
		Message: message,
		Context: ctx,
	})
	r.IsValid = false
	return r
}

// AddWarning adds a warning to the result.
func (r *ValidationResult) AddWarning(field, message string, ctx *Context[ValidationConfig]) *ValidationResult {
	r.Warnings = append(r.Warnings, ValidationWarning{
		Field:   field,
		Message: message,
		Context: ctx,
	})
	return r
}

// Execute executes the validation operation.
func (c *Context[ValidationConfig]) Execute(validator func() *ValidationResult) *ValidationResult {
	if validator != nil {
		return validator()
	}
	return NewValidationResult()
}

// LegacyValidationContext provides backward compatibility with the old ValidationContext type.
// DEPRECATED: Use Context[ValidationConfig] for new code.
// This type will be removed in v2.0.
type LegacyValidationContext struct {
	ConfigPath      string
	ValidationLevel string
	Profile         string
	Section         string
	MinValue        any
	MaxValue        any
	AllowedValues   []string
	ReferencedField string
	Constraints     map[string]string
	Metadata        map[string]string
}

// ToLegacyValidationContext converts a Context[ValidationConfig] to the legacy ValidationContext format.
func ToLegacyValidationContext(c *Context[ValidationConfig]) *LegacyValidationContext {
	// Directly access the ValueType field to avoid generic type inference issues
	return &LegacyValidationContext{
		ConfigPath:      c.ValueType.ConfigPath,
		ValidationLevel: c.ValueType.ValidationLevel,
		Profile:         c.ValueType.Profile,
		Section:         c.ValueType.Section,
		MinValue:        c.ValueType.MinValue,
		MaxValue:        c.ValueType.MaxValue,
		AllowedValues:   c.ValueType.AllowedValues,
		ReferencedField: c.ValueType.ReferencedField,
		Constraints:     c.ValueType.Constraints,
		Metadata:        c.Metadata,
	}
}

// FromLegacyValidationContext creates a Context[ValidationConfig] from the legacy ValidationContext format.
func FromLegacyValidationContext(ctx context.Context, legacy *LegacyValidationContext) *Context[ValidationConfig] {
	if legacy == nil {
		return NewContext(ctx, NewValidationConfig())
	}

	config := ValidationConfig{
		ConfigPath:      legacy.ConfigPath,
		ValidationLevel: legacy.ValidationLevel,
		Profile:         legacy.Profile,
		Section:         legacy.Section,
		MinValue:        legacy.MinValue,
		MaxValue:        legacy.MaxValue,
		AllowedValues:   legacy.AllowedValues,
		ReferencedField: legacy.ReferencedField,
		Constraints:     legacy.Constraints,
		Metadata:        legacy.Metadata,
	}

	newCtx := NewContext(ctx, config)
	for k, v := range legacy.Metadata {
		newCtx = newCtx.WithMetadata(k, v)
	}
	return newCtx
}

// NewLegacyValidationContext creates a new LegacyValidationContext with default values.
// DEPRECATED: Use NewContext[ValidationConfig] for new code.
func NewLegacyValidationContext() *LegacyValidationContext {
	return &LegacyValidationContext{
		Constraints:   make(map[string]string),
		Metadata:      make(map[string]string),
		AllowedValues: []string{},
	}
}
