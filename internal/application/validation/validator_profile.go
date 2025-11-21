package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateProfiles validates all profiles in configuration
func (cv *ConfigValidator) validateProfiles(cfg *domain.Config) error {
	for name, profile := range cfg.Profiles {
		// Validate profile name
		if err := cv.validateProfileName(name); err != nil {
			return fmt.Errorf("profile %s: %w", name, err)
		}

		// Check for nil profile to prevent panic
		if profile == nil {
			return fmt.Errorf("profile %s: nil profile", name)
		}

		// Validate profile struct
		if err := profile.Validate(name); err != nil {
			return fmt.Errorf("profile %s: %w", name, err)
		}
	}

	return nil
}

// validateProfileName validates profile name format
func (cv *ConfigValidator) validateProfileName(name string) error {
	// Explicitly reject empty names
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Use configured regex pattern if available
	if cv.rules.ProfileNamePattern != nil {
		if cv.rules.ProfileNamePattern.Pattern != "" {
			if compiledRegex := cv.rules.ProfileNamePattern.GetCompiledRegex(); compiledRegex != nil {
				if !compiledRegex.MatchString(name) {
					message := cv.rules.ProfileNamePattern.Message
					if message == "" {
						message = fmt.Sprintf("Profile name '%s' does not match pattern: %s", name, cv.rules.ProfileNamePattern.Pattern)
					}
					return fmt.Errorf("%s", message)
				}
				return nil // Pattern matched successfully
			} else {
				// Regex compilation failed, fall back to pattern in error message
				return fmt.Errorf("profile name pattern '%s' is invalid and cannot be compiled", cv.rules.ProfileNamePattern.Pattern)
			}
		}
		// Pattern field exists but is empty - use default validation
		return fmt.Errorf("profile name pattern is configured but empty")
	}

	// No pattern rule configured - use reasonable default validation
	// Only allow alphanumeric characters, underscores, and hyphens
	for _, char := range name {
		isValid := (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-'
		if !isValid {
			return fmt.Errorf("profile name '%s' contains invalid character: %c (allowed: alphanumeric, underscore, hyphen)", name, char)
		}
	}
	return nil
}
