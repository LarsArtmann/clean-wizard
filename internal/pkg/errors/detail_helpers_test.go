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
				assert.Equal(t, "", target)
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
