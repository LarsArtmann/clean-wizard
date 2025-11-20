package errors

import (
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/format"
)

func TestAllImprovements(t *testing.T) {
	// Test 1: Legacy "line" with float64 and "file" handling
	err := errors.NewError(errors.ErrConfigValidation, "Test error").
		WithDetail("line", 42.0). // float64 should be converted to int
		WithDetail("file", "/test/path").
		WithDetail("duration", time.Second*5) // time.Duration should use format.Duration

	if err.Details.LineNumber != 42 {
		t.Errorf("Expected line number 42, got %d", err.Details.LineNumber)
	}

	if err.Details.FilePath != "/test/path" {
		t.Errorf("Expected file path '/test/path', got '%s'", err.Details.FilePath)
	}

	if err.Details.Duration == "" {
		t.Error("Expected duration to be formatted")
	}

	// Verify duration uses proper format (not just String())
	if !strings.Contains(err.Details.Duration, "5s") {
		t.Errorf("Expected duration to contain '5s', got '%s'", err.Details.Duration)
	}

	// Test 2: Deterministic metadata ordering
	err = err.WithDetail("zebra_key", "zebra").
		WithDetail("alpha_key", "alpha").
		WithDetail("beta_key", "beta")

	errorStr := err.Error()

	// Should contain keys in sorted order (alpha, beta, zebra)
	alphaIndex := strings.Index(errorStr, "alpha_key=")
	betaIndex := strings.Index(errorStr, "beta_key=")
	zebraIndex := strings.Index(errorStr, "zebra_key=")

	if !(alphaIndex < betaIndex && betaIndex < zebraIndex) {
		t.Errorf("Metadata keys not sorted: alpha=%d, beta=%d, zebra=%d", alphaIndex, betaIndex, zebraIndex)
	}

	t.Logf("✅ Deterministic output: %s", errorStr)

	// Test 3: Different error levels (log method doesn't crash)
	levels := []errors.ErrorLevel{
		errors.LevelInfo,
		errors.LevelWarn,
		errors.LevelError,
		errors.LevelFatal, // Note: Fatal will exit, so we skip it in tests
	}

	for _, level := range levels[:3] { // Skip Fatal level
		testErr := errors.NewError(errors.ErrConfigValidation, "Test "+level.String()).
			WithDetail("line", float64(100)).
			WithDetail("duration", time.Millisecond*100)
		testErr.Level = level
		
		// This should not panic
		testErr.Log()
		t.Logf("✅ Log() succeeded for level: %s", level.String())
	}

	t.Log("✅ All improvements working correctly!")
}