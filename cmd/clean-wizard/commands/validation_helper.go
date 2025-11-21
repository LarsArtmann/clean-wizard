package commands

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// ApplyValidationToConfigShared applies validation rules to a loaded configuration
// This is the shared implementation that eliminates duplication
func ApplyValidationToConfigShared(loadedCfg *domain.Config, validationLevel config.ValidationLevel) error {
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
