package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStringField(t *testing.T) {
	tests := []struct {
		name        string
		inputValue  any
		expectValue string
		expectExact bool // true for strict, false for formatting
	}{
		{
			name:        "sets string value directly",
			inputValue:  "hello",
			expectValue: "hello",
			expectExact: true,
		},
		{
			name:        "formats int to string",
			inputValue:  42,
			expectValue: "42",
			expectExact: false,
		},
		{
			name:        "formats bool to string",
			inputValue:  true,
			expectValue: "true",
			expectExact: false,
		},
		{
			name:        "formats nil to string",
			inputValue:  nil,
			expectValue: "<nil>",
			expectExact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			wasExact := setStringField(&target, tt.inputValue)

			assert.Equal(t, tt.expectValue, target)
			assert.Equal(t, tt.expectExact, wasExact)
		})
	}
}

func TestSetStringFieldStrict(t *testing.T) {
	tests := []struct {
		name            string
		inputValue      any
		expectValue     string
		expectSet       bool
		expectUnchanged bool
	}{
		{
			name:        "sets string value directly",
			inputValue:  "hello",
			expectValue: "hello",
			expectSet:   true,
		},
		{
			name:            "rejects non-string value",
			inputValue:      42,
			expectValue:     "",
			expectSet:       false,
			expectUnchanged: true,
		},
		{
			name:            "rejects nil value",
			inputValue:      nil,
			expectValue:     "",
			expectSet:       false,
			expectUnchanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			wasSet := setStringFieldStrict(&target, tt.inputValue)

			if tt.expectUnchanged {
				assert.Empty(t, target)
			} else {
				assert.Equal(t, tt.expectValue, target)
			}
			assert.Equal(t, tt.expectSet, wasSet)
		})
	}
}

func TestSetIntField(t *testing.T) {
	tests := []struct {
		name            string
		inputValue      any
		expectValue     int
		expectSet       bool
		expectUnchanged bool
	}{
		{
			name:        "sets int value directly",
			inputValue:  42,
			expectValue: 42,
			expectSet:   true,
		},
		{
			name:            "rejects string value",
			inputValue:      "42",
			expectValue:     0,
			expectSet:       false,
			expectUnchanged: true,
		},
		{
			name:            "rejects float value",
			inputValue:      42.5,
			expectValue:     0,
			expectSet:       false,
			expectUnchanged: true,
		},
		{
			name:            "rejects nil value",
			inputValue:      nil,
			expectValue:     0,
			expectSet:       false,
			expectUnchanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target int
			wasSet := setIntField(&target, tt.inputValue)

			if tt.expectUnchanged {
				assert.Equal(t, 0, target)
			} else {
				assert.Equal(t, tt.expectValue, target)
			}
			assert.Equal(t, tt.expectSet, wasSet)
		})
	}
}

func TestAddToMetadata(t *testing.T) {
	tests := []struct {
		name       string
		metadata   map[string]string
		key        string
		inputValue any
		expectVal  string
	}{
		{
			name:       "adds to existing map",
			metadata:   make(map[string]string),
			key:        "test",
			inputValue: "value",
			expectVal:  "value",
		},
		{
			name:       "handles nil map by creating new one",
			metadata:   nil,
			key:        "test",
			inputValue: 42,
			expectVal:  "42",
		},
		{
			name:       "formats complex value",
			metadata:   make(map[string]string),
			key:        "complex",
			inputValue: struct{ field string }{field: "test"},
			expectVal:  "{test}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addToMetadata(tt.metadata, tt.key, tt.inputValue)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectVal, result[tt.key])

			// If original was nil, result should be a new map
			if tt.metadata == nil {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectVal, result[tt.key])
			}
		})
	}
}

func TestEnsureDetails(t *testing.T) {
	tests := []struct {
		name         string
		inputDetails **ErrorDetails
		expectField  string
	}{
		{
			name:         "initializes nil details",
			inputDetails: func() **ErrorDetails { var d *ErrorDetails; return &d }(),
			expectField:  "",
		},
		{
			name: "keeps existing details",
			inputDetails: func() **ErrorDetails {
				d := &ErrorDetails{Field: "existing"}
				return &d
			}(),
			expectField: "existing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensureDetails(tt.inputDetails)

			assert.NotNil(t, *tt.inputDetails)
			assert.NotNil(t, (*tt.inputDetails).Metadata)
			assert.Equal(t, tt.expectField, (*tt.inputDetails).Field)
		})
	}
}

func TestErrorDetailsBuilder(t *testing.T) {
	t.Run("basic construction", func(t *testing.T) {
		builder := NewErrorDetails()
		if builder == nil {
			t.Fatal("NewErrorDetails returned nil")
		}
		details := builder.Build()
		if details == nil {
			t.Fatal("Build returned nil")
		}
		if details.Metadata == nil {
			t.Error("Metadata should be initialized")
		}
	})

	t.Run("WithField", func(t *testing.T) {
		details := NewErrorDetails().
			WithField("maxRisk").
			Build()
		if details.Field != "maxRisk" {
			t.Errorf("Field = %q, want %q", details.Field, "maxRisk")
		}
	})

	t.Run("WithValue", func(t *testing.T) {
		details := NewErrorDetails().
			WithValue("high").
			Build()
		if details.Value != "high" {
			t.Errorf("Value = %q, want %q", details.Value, "high")
		}
	})

	t.Run("WithExpected", func(t *testing.T) {
		details := NewErrorDetails().
			WithExpected("low").
			Build()
		if details.Expected != "low" {
			t.Errorf("Expected = %q, want %q", details.Expected, "low")
		}
	})

	t.Run("WithActual", func(t *testing.T) {
		details := NewErrorDetails().
			WithActual("medium").
			Build()
		if details.Actual != "medium" {
			t.Errorf("Actual = %q, want %q", details.Actual, "medium")
		}
	})

	t.Run("WithOperation", func(t *testing.T) {
		details := NewErrorDetails().
			WithOperation("config_validation").
			Build()
		if details.Operation != "config_validation" {
			t.Errorf("Operation = %q, want %q", details.Operation, "config_validation")
		}
	})

	t.Run("WithFilePath", func(t *testing.T) {
		details := NewErrorDetails().
			WithFilePath("/etc/clean-wizard/config.yaml").
			Build()
		if details.FilePath != "/etc/clean-wizard/config.yaml" {
			t.Errorf("FilePath = %q, want %q", details.FilePath, "/etc/clean-wizard/config.yaml")
		}
	})

	t.Run("WithLineNumber", func(t *testing.T) {
		details := NewErrorDetails().
			WithLineNumber(42).
			Build()
		if details.LineNumber != 42 {
			t.Errorf("LineNumber = %d, want %d", details.LineNumber, 42)
		}
	})

	t.Run("WithRetryCount", func(t *testing.T) {
		details := NewErrorDetails().
			WithRetryCount(3).
			Build()
		if details.RetryCount != 3 {
			t.Errorf("RetryCount = %d, want %d", details.RetryCount, 3)
		}
	})

	t.Run("WithDuration", func(t *testing.T) {
		details := NewErrorDetails().
			WithDuration("500ms").
			Build()
		if details.Duration != "500ms" {
			t.Errorf("Duration = %q, want %q", details.Duration, "500ms")
		}
	})

	t.Run("WithMetadata", func(t *testing.T) {
		details := NewErrorDetails().
			WithMetadata("key1", "value1").
			WithMetadata("key2", "value2").
			Build()
		if details.Metadata["key1"] != "value1" {
			t.Errorf("Metadata[key1] = %q, want %q", details.Metadata["key1"], "value1")
		}
		if details.Metadata["key2"] != "value2" {
			t.Errorf("Metadata[key2] = %q, want %q", details.Metadata["key2"], "value2")
		}
	})

	t.Run("fluent chaining", func(t *testing.T) {
		details := NewErrorDetails().
			WithField("testField").
			WithValue("testValue").
			WithExpected("testExpected").
			WithActual("testActual").
			WithOperation("testOp").
			WithFilePath("/test/path").
			WithLineNumber(100).
			WithRetryCount(5).
			WithDuration("1s").
			WithMetadata("mk", "mv").
			Build()

		if details.Field != "testField" {
			t.Errorf("Field = %q, want %q", details.Field, "testField")
		}
		if details.Value != "testValue" {
			t.Errorf("Value = %q, want %q", details.Value, "testValue")
		}
		if details.Expected != "testExpected" {
			t.Errorf("Expected = %q, want %q", details.Expected, "testExpected")
		}
		if details.Actual != "testActual" {
			t.Errorf("Actual = %q, want %q", details.Actual, "testActual")
		}
		if details.Operation != "testOp" {
			t.Errorf("Operation = %q, want %q", details.Operation, "testOp")
		}
		if details.FilePath != "/test/path" {
			t.Errorf("FilePath = %q, want %q", details.FilePath, "/test/path")
		}
		if details.LineNumber != 100 {
			t.Errorf("LineNumber = %d, want %d", details.LineNumber, 100)
		}
		if details.RetryCount != 5 {
			t.Errorf("RetryCount = %d, want %d", details.RetryCount, 5)
		}
		if details.Duration != "1s" {
			t.Errorf("Duration = %q, want %q", details.Duration, "1s")
		}
		if details.Metadata["mk"] != "mv" {
			t.Errorf("Metadata[mk] = %q, want %q", details.Metadata["mk"], "mv")
		}
	})

	t.Run("metadata initialized once", func(t *testing.T) {
		details := NewErrorDetails().
			WithMetadata("key", "value").
			Build()
		if details.Metadata == nil {
			t.Error("Metadata should not be nil")
		}
		if len(details.Metadata) != 1 {
			t.Errorf("Metadata length = %d, want %d", len(details.Metadata), 1)
		}
	})
}
