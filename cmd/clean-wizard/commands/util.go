package commands

import (
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/config"
)

// ParseValidationLevel converts string to ValidationLevel.
func ParseValidationLevel(level string) config.ValidationLevel {
	switch strings.ToLower(level) {
	case "none":
		return config.ValidationLevelNone
	case "basic":
		return config.ValidationLevelBasic
	case "comprehensive":
		return config.ValidationLevelComprehensive
	case "strict":
		return config.ValidationLevelStrict
	default:
		return config.ValidationLevelBasic // Safe default
	}
}
