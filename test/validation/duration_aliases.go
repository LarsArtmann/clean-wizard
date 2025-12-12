package validation_test

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/validation"
)

// ParseCustomDuration alias for test compatibility
var ParseCustomDuration = validation.ParseDuration

// ValidateCustomDuration alias for test compatibility
var ValidateCustomDuration = validation.ParseDuration

// FormatDuration alias for test compatibility
var FormatDuration = func(d time.Duration) string {
	return d.String()
}
