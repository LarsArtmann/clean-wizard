package config

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// validateProfiles validates all profiles in configuration.
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

// validateProfileName validates profile name format.
func (cv *ConfigValidator) validateProfileName(name string) error {
	// Reject empty names
	if name == "" {
		return errors.New("profile name cannot be empty")
	}

	// Use configured regex pattern if available
	if cv.rules.ProfileNamePattern != nil && cv.rules.ProfileNamePattern.Pattern != "" {
		return cv.validateProfileNameWithPattern(name)
	}

	// Use default validation
	return cv.validateProfileNameWithDefault(name)
}

// validateProfileNameWithPattern validates profile name using configured regex pattern.
func (cv *ConfigValidator) validateProfileNameWithPattern(name string) error {
	compiledRegex := cv.rules.ProfileNamePattern.GetCompiledRegex()
	if compiledRegex == nil {
		return fmt.Errorf("profile name pattern '%s' is invalid and cannot be compiled", cv.rules.ProfileNamePattern.Pattern)
	}

	if !compiledRegex.MatchString(name) {
		message := cv.rules.ProfileNamePattern.Message
		if message == "" {
			message = fmt.Sprintf("Profile name '%s' does not match pattern: %s", name, cv.rules.ProfileNamePattern.Pattern)
		}
		return fmt.Errorf("%s", message)
	}
	return nil
}

// validateProfileNameWithDefault validates profile name using default character validation.
func (cv *ConfigValidator) validateProfileNameWithDefault(name string) error {
	for _, char := range name {
		if !isValidProfileNameChar(char) {
			return fmt.Errorf("profile name '%s' contains invalid character: %c (allowed: alphanumeric, underscore, hyphen)", name, char)
		}
	}
	return nil
}

// isValidProfileNameChar checks if a character is valid for a profile name.
func isValidProfileNameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-'
}
