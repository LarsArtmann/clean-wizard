package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSizeEstimateStatusType_String(t *testing.T) {
	tests := []struct {
		name     string
		status   SizeEstimateStatusType
		expected string
	}{
		{
			name:     "KNOWN returns KNOWN",
			status:   SizeEstimateStatusKnown,
			expected: "KNOWN",
		},
		{
			name:     "UNKNOWN returns UNKNOWN",
			status:   SizeEstimateStatusUnknown,
			expected: "UNKNOWN",
		},
		{
			name:     "invalid returns INVALID",
			status:   SizeEstimateStatusType(99),
			expected: "INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestSizeEstimateStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   SizeEstimateStatusType
		expected bool
	}{
		{
			name:     "KNOWN is valid",
			status:   SizeEstimateStatusKnown,
			expected: true,
		},
		{
			name:     "UNKNOWN is valid",
			status:   SizeEstimateStatusUnknown,
			expected: true,
		},
		{
			name:     "negative is invalid",
			status:   SizeEstimateStatusType(-1),
			expected: false,
		},
		{
			name:     "out of range is invalid",
			status:   SizeEstimateStatusType(99),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestSizeEstimateStatusType_Values(t *testing.T) {
	values := SizeEstimateStatusKnown.Values()

	assert.Len(t, values, 2)
	assert.Contains(t, values, SizeEstimateStatusKnown)
	assert.Contains(t, values, SizeEstimateStatusUnknown)
}

func TestSizeEstimateStatusType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   SizeEstimateStatusType
		expected string
		wantErr  bool
	}{
		{
			name:     "KNOWN marshals correctly",
			status:   SizeEstimateStatusKnown,
			expected: `"KNOWN"`,
			wantErr:  false,
		},
		{
			name:     "UNKNOWN marshals correctly",
			status:   SizeEstimateStatusUnknown,
			expected: `"UNKNOWN"`,
			wantErr:  false,
		},
		{
			name:    "invalid status fails to marshal",
			status:  SizeEstimateStatusType(99),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestSizeEstimateStatusType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		expected  SizeEstimateStatusType
		wantErr   bool
	}{
		{
			name:      "KNOWN unmarshals correctly",
			jsonInput: `"KNOWN"`,
			expected:  SizeEstimateStatusKnown,
			wantErr:   false,
		},
		{
			name:      "UNKNOWN unmarshals correctly",
			jsonInput: `"UNKNOWN"`,
			expected:  SizeEstimateStatusUnknown,
			wantErr:   false,
		},
		{
			name:      "lowercase works",
			jsonInput: `"known"`,
			expected:  SizeEstimateStatusKnown,
			wantErr:   false,
		},
		{
			name:      "invalid value fails",
			jsonInput: `"INVALID"`,
			wantErr:   true,
		},
		{
			name:      "malformed JSON fails",
			jsonInput: `invalid`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status SizeEstimateStatusType
			err := json.Unmarshal([]byte(tt.jsonInput), &status)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestSizeEstimateStatusType_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		status SizeEstimateStatusType
	}{
		{
			name:   "KNOWN roundtrip",
			status: SizeEstimateStatusKnown,
		},
		{
			name:   "UNKNOWN roundtrip",
			status: SizeEstimateStatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.status)
			require.NoError(t, err)

			// Unmarshal
			var result SizeEstimateStatusType
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.status, result)
		})
	}
}
