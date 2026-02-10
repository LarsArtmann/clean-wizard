package context

import (
	"context"
	"testing"
)

func TestErrorConfigBuilder(t *testing.T) {
	ctx := context.Background()

	// Test with struct literal - the expected pattern
	config := ErrorConfig{
		Operation:   "test_operation",
		Field:       "test_field",
		Value:       "test_value",
		Expected:    "expected_value",
		Actual:      "actual_value",
		Code:        "ERR001",
		Level:       "error",
		Severity:    "high",
		RetryCount:  3,
		Metadata:    map[string]string{"key1": "value1", "key2": "value2"},
	}

	c := NewContext(ctx, config)

	if c.ValueType.Operation != "test_operation" {
		t.Errorf("Expected Operation 'test_operation', got '%s'", c.ValueType.Operation)
	}

	if c.ValueType.Field != "test_field" {
		t.Errorf("Expected Field 'test_field', got '%s'", c.ValueType.Field)
	}

	if c.ValueType.Value != "test_value" {
		t.Errorf("Expected Value 'test_value', got '%s'", c.ValueType.Value)
	}

	if c.ValueType.Code != "ERR001" {
		t.Errorf("Expected Code 'ERR001', got '%s'", c.ValueType.Code)
	}

	// Check config's metadata (not context's metadata)
	if len(c.ValueType.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries in config, got %d", len(c.ValueType.Metadata))
	}
}

func TestSanitizationConfigBuilder(t *testing.T) {
	ctx := context.Background()

	// Test with struct literal - the expected pattern
	config := SanitizationConfig{
		Operation:      "trim_whitespace",
		Field:          "version",
		Rules:          map[string]string{"max_length": "100", "min_length": "1"},
		TrimWhitespace: true,
		NormalizeCase:  false,
		ClampValues:    true,
		Metadata:       map[string]string{"policy": "strict"},
	}

	c := NewContext(ctx, config)

	if c.ValueType.Operation != "trim_whitespace" {
		t.Errorf("Expected Operation 'trim_whitespace', got '%s'", c.ValueType.Operation)
	}

	if c.ValueType.Field != "version" {
		t.Errorf("Expected Field 'version', got '%s'", c.ValueType.Field)
	}

	if c.ValueType.Rules["max_length"] != "100" {
		t.Errorf("Expected Rules[max_length]='100', got '%s'", c.ValueType.Rules["max_length"])
	}

	if !c.ValueType.TrimWhitespace {
		t.Error("Expected TrimWhitespace to be true")
	}

	// Check config's metadata (not context's metadata)
	if c.ValueType.Metadata["policy"] != "strict" {
		t.Errorf("Expected config metadata policy='strict', got '%s'", c.ValueType.Metadata["policy"])
	}
}

func TestLegacyErrorDetailsConversion(t *testing.T) {
	ctx := context.Background()
	legacy := NewLegacyErrorDetails()
	legacy.Field = "test_field"
	legacy.Value = "test_value"
	legacy.Expected = "expected"
	legacy.Actual = "actual"
	legacy.Operation = "test_op"
	legacy.Metadata["key1"] = "value1"

	// Convert from legacy to context
	convertedCtx := FromLegacyErrorDetails(ctx, legacy)

	if convertedCtx.ValueType.Field != "test_field" {
		t.Errorf("Expected Field 'test_field', got '%s'", convertedCtx.ValueType.Field)
	}

	if convertedCtx.ValueType.Value != "test_value" {
		t.Errorf("Expected Value 'test_value', got '%s'", convertedCtx.ValueType.Value)
	}

	if convertedCtx.Metadata["key1"] != "value1" {
		t.Errorf("Expected context metadata key1='value1', got '%s'", convertedCtx.Metadata["key1"])
	}

	// Convert from context to legacy
	backToLegacy := ToLegacyErrorDetails(convertedCtx)

	if backToLegacy.Field != "test_field" {
		t.Errorf("Expected Field 'test_field', got '%s'", backToLegacy.Field)
	}

	if backToLegacy.Value != "test_value" {
		t.Errorf("Expected Value 'test_value', got '%s'", backToLegacy.Value)
	}
}

func TestLegacyErrorDetailsNil(t *testing.T) {
	ctx := context.Background()

	// Test nil legacy
	convertedCtx := FromLegacyErrorDetails(ctx, nil)

	if convertedCtx.ValueType.Field != "" {
		t.Errorf("Expected empty Field, got '%s'", convertedCtx.ValueType.Field)
	}

	if convertedCtx.ValueType.Operation != "" {
		t.Errorf("Expected empty Operation, got '%s'", convertedCtx.ValueType.Operation)
	}
}

func TestErrorResultTypes(t *testing.T) {
	result := NewErrorResult()

	if result.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}

	result.Handled = true
	result.Recovered = true
	result.Fatal = false
	result.Retryable = true

	if !result.Handled {
		t.Error("Expected Handled to be true")
	}

	if !result.Recovered {
		t.Error("Expected Recovered to be true")
	}

	if result.Fatal {
		t.Error("Expected Fatal to be false")
	}

	if !result.Retryable {
		t.Error("Expected Retryable to be true")
	}
}

func TestSanitizationResultTypes(t *testing.T) {
	result := NewSanitizationResultV2()

	if result.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}

	result.Changed = true
	result.Fields = []string{"field1", "field2"}
	result.Warnings = []string{"warning1"}

	if !result.Changed {
		t.Error("Expected Changed to be true")
	}

	if len(result.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result.Fields))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}