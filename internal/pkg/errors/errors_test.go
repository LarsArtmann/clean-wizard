package errors

import (
	"testing"
	"time"
)

func TestErrorDetailsWithLegacyKeys(t *testing.T) {
	err := NewError(ErrConfigValidation, "Test error")

	// Test legacy keys are properly mapped
	err = err.WithDetail("line", 42)
	if err.Details.LineNumber != 42 {
		t.Errorf("Expected line number 42, got %d", err.Details.LineNumber)
	}

	err = err.WithDetail("file", "/test/path")
	if err.Details.FilePath != "/test/path" {
		t.Errorf("Expected file path '/test/path', got %s", err.Details.FilePath)
	}
}

func TestErrorDetailsWithDurationFormatting(t *testing.T) {
	err := NewError(ErrConfigValidation, "Test error")

	// Test duration formatting using format.Duration
	duration := 5 * time.Minute
	err = err.WithDetail("duration", duration)

	// Should use human-friendly format, not time.Duration.String()
	expected := "5.0 m"
	if err.Details.Duration != expected {
		t.Errorf("Expected duration '%s', got '%s'", expected, err.Details.Duration)
	}
}

func TestErrorDetailsWithNumericConversion(t *testing.T) {
	err := NewError(ErrConfigValidation, "Test error")

	// Test float64 to int conversion for line numbers
	err = err.WithDetail("line", float64(42))
	if err.Details.LineNumber != 42 {
		t.Errorf("Expected line number 42 from float64, got %d", err.Details.LineNumber)
	}

	// Test retry count
	err = err.WithDetail("retry_count", 3)
	if err.Details.RetryCount != 3 {
		t.Errorf("Expected retry count 3, got %d", err.Details.RetryCount)
	}
}

func TestErrorOutputDeterministic(t *testing.T) {
	err := NewErrorWithDetails(ErrConfigValidation, "Test error", &ErrorDetails{
		Metadata: map[string]string{
			"zebra_key": "zebra_value",
			"alpha_key": "alpha_value",
			"beta_key":  "beta_value",
		},
	})

	errorStr := err.Error()

	// Should contain keys in sorted order (alpha, beta, zebra)
	alphaIndex := findSubstring(errorStr, "alpha_key=")
	betaIndex := findSubstring(errorStr, "beta_key=")
	zebraIndex := findSubstring(errorStr, "zebra_key=")

	if alphaIndex == -1 || betaIndex == -1 || zebraIndex == -1 {
		t.Error("Not all metadata keys found in error output")
	}

	// Check deterministic ordering
	if !(alphaIndex < betaIndex && betaIndex < zebraIndex) {
		t.Errorf("Metadata keys not sorted: alpha=%d, beta=%d, zebra=%d", alphaIndex, betaIndex, zebraIndex)
	}
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
