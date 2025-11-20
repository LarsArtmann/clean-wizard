package commands

import (
	"fmt"
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
func ApplyValidationToConfig(loadedCfg *domain.Config, validationLevel config.ValidationLevel) error {
	if validationLevel > config.ValidationLevelNone {
		fmt.Printf("🔍 Applying validation level: %s\n", validationLevel.String())

		if validationLevel >= config.ValidationLevelBasic {
			// Basic validation
			if len(loadedCfg.Protected) == 0 {
				return fmt.Errorf("basic validation failed: protected paths cannot be empty")
			}
		}

		if validationLevel >= config.ValidationLevelComprehensive {
			// Comprehensive validation
			if err := loadedCfg.Validate(); err != nil {
				return fmt.Errorf("comprehensive validation failed: %w", err)
			}
		}

		if validationLevel >= config.ValidationLevelStrict {
			// Strict validation
			if loadedCfg.SafetyLevel == domain.SafetyLevelDisabled {
				return fmt.Errorf("strict validation failed: safety_level must be enabled")
			}
		}
	}
	return nil
}
