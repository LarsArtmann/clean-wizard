package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSanitizationResult implements the SanitizationResult interface for testing
type MockSanitizationResult struct {
	Changes []struct {
		Path     string
		Original any
		NewValue any
		Reason   string
	}
}

func (m *MockSanitizationResult) AddChange(path string, original, newValue any, reason string) {
	m.Changes = append(m.Changes, struct {
		Path     string
		Original any
		NewValue any
		Reason   string
	}{
		Path:     path,
		Original: original,
		NewValue: newValue,
		Reason:   reason,
	})
}

func TestTrimWhitespaceField(t *testing.T) {
	tests := []struct {
		name          string
		field         TrimmableField
		changes       *MockSanitizationResult
		expectChange  bool
		expectedValue string
	}{
		{
			name: "trims leading and trailing whitespace",
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := "  hello world  "; return &s }(),
			},
			changes:       &MockSanitizationResult{},
			expectChange:  true,
			expectedValue: "hello world",
		},
		{
			name: "does not change already trimmed string",
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := "hello world"; return &s }(),
			},
			changes:       &MockSanitizationResult{},
			expectChange:  false,
			expectedValue: "hello world",
		},
		{
			name: "handles empty string",
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := ""; return &s }(),
			},
			changes:       &MockSanitizationResult{},
			expectChange:  false,
			expectedValue: "",
		},
		{
			name: "handles string with only whitespace",
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := "   \t\n   "; return &s }(),
			},
			changes:       &MockSanitizationResult{},
			expectChange:  true,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := TrimWhitespaceField(tt.field, tt.changes)

			assert.Equal(t, tt.expectChange, changed)
			assert.Equal(t, tt.expectedValue, *tt.field.Value)

			if tt.expectChange {
				require.Len(t, tt.changes.Changes, 1)
				assert.Equal(t, tt.field.Path, tt.changes.Changes[0].Path)
				assert.Equal(t, "trimmed whitespace", tt.changes.Changes[0].Reason)
			} else {
				assert.Empty(t, tt.changes.Changes)
			}
		})
	}
}

func TestTrimWhitespaceField_NilValue(t *testing.T) {
	field := TrimmableField{
		Name:  "test",
		Path:  "test.path",
		Value: nil,
	}
	changes := &MockSanitizationResult{}

	changed := TrimWhitespaceField(field, changes)

	assert.False(t, changed)
	assert.Empty(t, changes.Changes)
}

func TestTrimIfEnabled(t *testing.T) {
	tests := []struct {
		name         string
		trimEnabled  bool
		field        TrimmableField
		changes      *MockSanitizationResult
		expectChange bool
	}{
		{
			name:        "trims when enabled",
			trimEnabled: true,
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := "  hello  "; return &s }(),
			},
			changes:      &MockSanitizationResult{},
			expectChange: true,
		},
		{
			name:        "does not trim when disabled",
			trimEnabled: false,
			field: TrimmableField{
				Name:  "test",
				Path:  "test.path",
				Value: func() *string { s := "  hello  "; return &s }(),
			},
			changes:      &MockSanitizationResult{},
			expectChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := TrimIfEnabled(tt.trimEnabled, tt.field, tt.changes)

			assert.Equal(t, tt.expectChange, changed)

			if tt.trimEnabled && tt.expectChange {
				assert.Equal(t, "hello", *tt.field.Value)
				assert.Len(t, tt.changes.Changes, 1)
			} else if !tt.trimEnabled {
				assert.Equal(t, "  hello  ", *tt.field.Value)
				assert.Empty(t, tt.changes.Changes)
			}
		})
	}
}

func TestTrimmableFieldsBuilder(t *testing.T) {
	builder := NewTrimmableFieldsBuilder()

	value1 := "  test1  "
	value2 := "  test2  "

	fields := builder.
		AddField("field1", "path1", &value1).
		AddField("field2", "path2", &value2).
		Build()

	assert.Len(t, fields, 2)
	assert.Equal(t, "field1", fields[0].Name)
	assert.Equal(t, "path1", fields[0].Path)
	assert.Equal(t, &value1, fields[0].Value)
	assert.Equal(t, "field2", fields[1].Name)
	assert.Equal(t, "path2", fields[1].Path)
	assert.Equal(t, &value2, fields[1].Value)
}

func TestTrimMultipleFields(t *testing.T) {
	value1 := "  test1  "
	value2 := "  test2  "
	value3 := "test3" // Already trimmed

	fields := []TrimmableField{
		{Name: "field1", Path: "path1", Value: &value1},
		{Name: "field2", Path: "path2", Value: &value2},
		{Name: "field3", Path: "path3", Value: &value3},
	}

	changes := &MockSanitizationResult{}

	// Test with trimming enabled
	changedCount := TrimMultipleFields(true, fields, changes)

	assert.Equal(t, 2, changedCount) // Only first two should change
	assert.Equal(t, "test1", value1)
	assert.Equal(t, "test2", value2)
	assert.Equal(t, "test3", value3) // Unchanged
	assert.Len(t, changes.Changes, 2)

	// Test with trimming disabled
	changes.Changes = nil
	changedCount = TrimMultipleFields(false, fields, changes)

	assert.Equal(t, 0, changedCount)
	assert.Empty(t, changes.Changes)
}
