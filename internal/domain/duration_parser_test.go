package domain

import (
	"testing"
	"time"
)

// TestParseCustomDuration provides comprehensive testing for custom duration parsing.
func TestParseCustomDuration(t *testing.T) {
	t.Run("valid durations", testValidDurations)
	t.Run("invalid durations", testInvalidDurations)
}

func testValidDurations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"Go duration - seconds", "30s", 30 * time.Second},
		{"Go duration - minutes", "15m", 15 * time.Minute},
		{"Go duration - hours", "2h", 2 * time.Hour},
		{"Go duration - complex", "1h30m", 1*time.Hour + 30*time.Minute},
		{"custom - single day", "1d", 24 * time.Hour},
		{"custom - multiple days", "7d", 7 * 24 * time.Hour},
		{"custom - fractional days", "1.5d", 36 * time.Hour},
		{"custom - many days", "30d", 30 * 24 * time.Hour},
		{"whitespace handling", " 7d ", 7 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCustomDuration(tt.input)
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v for '%s', got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func testInvalidDurations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"unsupported unit", "1w"},
		{"malformed days", "1.xd"},
		{"negative days", "-1d"},
		{"just unit", "d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCustomDuration(tt.input)
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", tt.input)
			}
		})
	}
}

// TestValidateCustomDuration tests the validation function.
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

// BenchmarkParseCustomDuration benchmarks custom duration parser performance.
func BenchmarkParseCustomDuration(b *testing.B) {
	testInputs := []string{"1d", "7d", "24h", "30m", "1h30m", "15s"}

	for b.Loop() {
		for _, input := range testInputs {
			_, _ = ParseCustomDuration(input)
		}
	}
}

// BenchmarkParseGoDuration benchmarks Go's native time.ParseDuration.
func BenchmarkParseGoDuration(b *testing.B) {
	testInputs := []string{"24h", "30m", "1h30m", "15s"}

	for b.Loop() {
		for _, input := range testInputs {
			_, _ = time.ParseDuration(input)
		}
	}
}
