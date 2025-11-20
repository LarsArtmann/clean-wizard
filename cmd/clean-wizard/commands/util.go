package commands

import (
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ParseValidationLevel converts string to ValidationLevel
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

// ApplyValidationToConfig applies validation rules to the loaded configuration
// Uses shared validation logic from validation_helper.go to eliminate duplication
func ApplyValidationToConfig(loadedCfg *domain.Config, validationLevel config.ValidationLevel) error {
	return ApplyValidationToConfigShared(loadedCfg, validationLevel)
}
