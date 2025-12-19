package strings

import (
	"strings"
)

// TrimmableField represents a string field that can be trimmed
type TrimmableField struct {
	Name     string  // Field name for identification
	Value    *string // Pointer to string value to modify
	Path     string  // Full path for change tracking
	Original string  // Original value (if needed separately)
}

// SanitizationResult interface for change tracking
type SanitizationResult interface {
	AddChange(path string, original, newValue any, reason string)
}

// TrimWhitespaceField trims whitespace from a string field and tracks changes
// Returns true if the value was changed, false otherwise
func TrimWhitespaceField(field TrimmableField, changes SanitizationResult) bool {
	if field.Value == nil {
		return false
	}

	original := *field.Value
	trimmed := strings.TrimSpace(original)

	if original != trimmed {
		*field.Value = trimmed
		if changes != nil {
			changes.AddChange(field.Path, original, trimmed, "trimmed whitespace")
		}
		return true
	}

	return false
}

// TrimIfEnabled conditionally trims a field if trimEnabled is true
// Returns true if the value was changed, false otherwise
func TrimIfEnabled(trimEnabled bool, field TrimmableField, changes SanitizationResult) bool {
	if !trimEnabled {
		return false
	}

	return TrimWhitespaceField(field, changes)
}

// TrimmableFieldsBuilder helps build trimmable field configurations
type TrimmableFieldsBuilder struct {
	fields []TrimmableField
}

// NewTrimmableFieldsBuilder creates a new builder for trimmable fields
func NewTrimmableFieldsBuilder() *TrimmableFieldsBuilder {
	return &TrimmableFieldsBuilder{
		fields: make([]TrimmableField, 0),
	}
}

// AddField adds a trimmable field to the builder
func (b *TrimmableFieldsBuilder) AddField(name, path string, value *string) *TrimmableFieldsBuilder {
	b.fields = append(b.fields, TrimmableField{
		Name:  name,
		Value: value,
		Path:  path,
	})
	return b
}

// Build returns the configured trimmable fields
func (b *TrimmableFieldsBuilder) Build() []TrimmableField {
	return b.fields
}

// TrimMultipleFields trims multiple fields with the same configuration
// Returns the count of fields that were actually changed
func TrimMultipleFields(trimEnabled bool, fields []TrimmableField, changes SanitizationResult) int {
	if !trimEnabled {
		return 0
	}

	changed := 0
	for _, field := range fields {
		if TrimWhitespaceField(field, changes) {
			changed++
		}
	}

	return changed
}
