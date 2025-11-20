package domain

import (
	"testing"
	"time"
)

// TestParseCustomDuration provides comprehensive testing for custom duration parsing
func TestParseCustomDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectValid bool
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "Valid Go duration - seconds",
			input:       "30s",
			expectValid: true,
			expected:    30 * time.Second,
			expectError: false,
		},
		{
			name:        "Valid Go duration - minutes",
			input:       "15m",
			expectValid: true,
			expected:    15 * time.Minute,
			expectError: false,
		},
		{
			name:        "Valid Go duration - hours",
			input:       "2h",
			expectValid: true,
			expected:    2 * time.Hour,
			expectError: false,
		},
		{
			name:        "Valid Go duration - complex",
			input:       "1h30m",
			expectValid: true,
			expected:    1*time.Hour + 30*time.Minute,
			expectError: false,
		},
		{
			name:        "Valid custom duration - single day",
			input:       "1d",
			expectValid: true,
			expected:    24 * time.Hour,
			expectError: false,
		},
		{
			name:        "Valid custom duration - multiple days",
			input:       "7d",
			expectValid: true,
			expected:    7 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "Valid custom duration - fractional days",
			input:       "1.5d",
			expectValid: true,
			expected:    36 * time.Hour,
			expectError: false,
		},
		{
			name:        "Valid custom duration - many days",
			input:       "30d",
			expectValid: true,
			expected:    30 * 24 * time.Hour,
			expectError: false,
		},
		{
			name:        "Invalid duration - empty string",
			input:       "",
			expectValid: false,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid duration - unsupported unit",
			input:       "1w", // weeks not supported
			expectValid: false,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid duration - malformed days",
			input:       "1.xd",
			expectValid: false,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid duration - negative days",
			input:       "-1d",
			expectValid: false,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid duration - just unit",
			input:       "d",
			expectValid: false,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Whitespace handling - valid with spaces",
			input:       " 7d ",
			expectValid: true,
			expected:    7 * 24 * time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCustomDuration(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
					return
				}
				t.Logf("✓ Expected error for input '%s': %v", tt.input, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
					return
				}

				if result != tt.expected {
					t.Errorf("Expected duration %v for input '%s', but got %v", tt.expected, tt.input, result)
					return
				}

				t.Logf("✓ Parsed '%s' to %v", tt.input, result)
			}
		})
	}
}

// TestValidateCustomDuration tests the validation function
func TestValidateCustomDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid duration",
			input:       "7d",
			expectError: false,
		},
		{
			name:        "Valid Go duration",
			input:       "24h",
			expectError: false,
		},
		{
			name:        "Empty duration",
			input:       "",
			expectError: true,
		},
		{
			name:        "Invalid duration",
			input:       "1w",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCustomDuration(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tt.input)
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if tt.expectError {
				t.Logf("✓ Expected error for input '%s': %v", tt.input, err)
			} else {
				t.Logf("✓ Validated duration '%s'", tt.input)
			}
		})
	}
}

// BenchmarkParseCustomDuration benchmarks custom duration parser performance
func BenchmarkParseCustomDuration(b *testing.B) {
	testInputs := []string{"1d", "7d", "24h", "30m", "1h30m", "15s"}

	for b.Loop() {
		for _, input := range testInputs {
			_, _ = ParseCustomDuration(input)
		}
	}
}

// TestFormatDuration tests the duration formatting function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		// Edge cases
		{
			name:     "Zero duration",
			input:    0,
			expected: "0",
		},
		{
			name:     "Very small duration",
			input:    500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "Duration less than 1 minute",
			input:    30 * time.Second,
			expected: "30s",
		},

		// Exact minutes - should format as "Xm"
		{
			name:     "Exact 1 minute",
			input:    1 * time.Minute,
			expected: "1m",
		},
		{
			name:     "Exact 30 minutes",
			input:    30 * time.Minute,
			expected: "30m",
		},
		{
			name:     "Exact 59 minutes",
			input:    59 * time.Minute,
			expected: "59m",
		},

		// Non-exact minutes - should return d.String()
		{
			name:     "1 minute 30 seconds",
			input:    1*time.Minute + 30*time.Second,
			expected: "1m30s",
		},
		{
			name:     "30 minutes 30 seconds",
			input:    30*time.Minute + 30*time.Second,
			expected: "30m30s",
		},
		{
			name:     "59 minutes 30 seconds",
			input:    59*time.Minute + 30*time.Second,
			expected: "59m30s",
		},

		// Exact hours - should format as "Xh"
		{
			name:     "Exact 1 hour",
			input:    1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "Exact 12 hours",
			input:    12 * time.Hour,
			expected: "12h",
		},
		{
			name:     "Exact 23 hours",
			input:    23 * time.Hour,
			expected: "23h",
		},

		// Non-exact hours - should return d.String()
		{
			name:     "1 hour 30 minutes",
			input:    1*time.Hour + 30*time.Minute,
			expected: "1h30m0s",
		},
		{
			name:     "12 hours 30 minutes",
			input:    12*time.Hour + 30*time.Minute,
			expected: "12h30m0s",
		},
		{
			name:     "23 hours 30 minutes",
			input:    23*time.Hour + 30*time.Minute,
			expected: "23h30m0s",
		},
		{
			name:     "1 hour 30 minutes 30 seconds",
			input:    1*time.Hour + 30*time.Minute + 30*time.Second,
			expected: "1h30m30s",
		},

		// Exact days - should format as "Xd"
		{
			name:     "Exact 1 day",
			input:    24 * time.Hour,
			expected: "1d",
		},
		{
			name:     "Exact 7 days",
			input:    7 * 24 * time.Hour,
			expected: "7d",
		},
		{
			name:     "Exact 30 days",
			input:    30 * 24 * time.Hour,
			expected: "30d",
		},

		// Non-exact days - should return d.String()
		{
			name:     "1 day 1 hour",
			input:    24*time.Hour + 1*time.Hour,
			expected: "25h0m0s",
		},
		{
			name:     "7 days 12 hours",
			input:    7*24*time.Hour + 12*time.Hour,
			expected: "180h0m0s",
		},
		{
			name:     "30 days 1 hour 30 minutes",
			input:    30*24*time.Hour + 1*time.Hour + 30*time.Minute,
			expected: "721h30m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s' for input %v, but got '%s'", tt.expected, tt.input, result)
			} else {
				t.Logf("✓ FormatDuration(%v) = '%s'", tt.input, result)
			}
		})
	}
}

// BenchmarkParseGoDuration benchmarks Go's native time.ParseDuration
func BenchmarkParseGoDuration(b *testing.B) {
	testInputs := []string{"24h", "30m", "1h30m", "15s"}

	for b.Loop() {
		for _, input := range testInputs {
			_, _ = time.ParseDuration(input)
		}
	}
}
