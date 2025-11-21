package commands

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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

// ResolveProfile resolves which profile to use based on configuration and parameters
// Eliminates duplication between clean and scan commands
func ResolveProfile(loadedCfg *domain.Config, profileName string) (*domain.Profile, error) {
	if loadedCfg == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	// Apply profile if specified
	if profileName != "" {
		profile, exists := loadedCfg.Profiles[profileName]
		if !exists {
			return nil, fmt.Errorf("profile '%s' not found in configuration", profileName)
		}

		if profile.Status == domain.StatusDisabled {
			return nil, fmt.Errorf("profile '%s' is disabled", profileName)
		}

		fmt.Printf("🏷️  Using profile: %s (%s)\n", profileName, profile.Description)
		return profile, nil
	}

	// Use current profile from config if available
	if loadedCfg.CurrentProfile != "" {
		profile := loadedCfg.Profiles[loadedCfg.CurrentProfile]
		if profile != nil && profile.Status == domain.StatusEnabled {
			fmt.Printf("🏷️  Using current profile: %s (%s)\n", loadedCfg.CurrentProfile, profile.Description)
			return profile, nil
		}
	}

	return nil, fmt.Errorf("no valid profile available")
}

// ApplyValidationToConfig applies validation rules to the loaded configuration
// Uses shared validation logic from validation_helper.go to eliminate duplication
func ApplyValidationToConfig(loadedCfg *domain.Config, validationLevel config.ValidationLevel) error {
	return ApplyValidationToConfigShared(loadedCfg, validationLevel)
}
