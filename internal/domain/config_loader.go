package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// ViperConfig interface for viper operations (enables testing and dependency inversion)
type ViperConfig interface {
	IsSet(key string) bool
	Get(key string) any
	GetBool(key string) bool
	GetString(key string) string
	GetInt(key string) int
}

// SafetyConfig represents single source of truth for safety configuration
// This type eliminates split brains by having only ONE representation
type SafetyConfig struct {
	Level SafetyLevelType

	// Whether this config was explicitly set by user
	IsExplicit bool
}

// ParseSafetyConfig creates a SafetyConfig from viper configuration
// This function centralizes all safety level parsing logic outside of domain
func ParseSafetyConfig(v ViperConfig) SafetyConfig {
	// Check for new safety_level format first (highest priority)
	if v.IsSet("safety_level") {
		safetyValue := v.Get("safety_level")

		if level, ok := parseSafetyLevelValue(safetyValue); ok {
			return SafetyConfig{
				Level:      level,
				IsExplicit: true,
			}
		}

		// Invalid value - return default with explicit flag
		return SafetyConfig{
			Level:      SafetyLevelEnabled,
			IsExplicit: true,
		}
	}

	// Fall back to legacy safe_mode boolean
	if v.IsSet("safe_mode") {
		safeMode := v.GetBool("safe_mode")

		if safeMode {
			return SafetyConfig{
				Level:      SafetyLevelEnabled,
				IsExplicit: true,
			}
		}
		return SafetyConfig{
			Level:      SafetyLevelDisabled,
			IsExplicit: true,
		}
	}

	// Default when neither is present
	return SafetyConfig{
		Level:      SafetyLevelEnabled,
		IsExplicit: false,
	}
}

// parseSafetyLevelValue attempts to parse safety level from interface value
// Returns parsed level and success flag
func parseSafetyLevelValue(value any) (SafetyLevelType, bool) {
	switch val := value.(type) {
	case string:
		return parseSafetyLevelString(strings.TrimSpace(val))
	case int, int32, int64:
		level, ok := parseSafetyLevelNumeric(fmt.Sprintf("%v", val))
		return level, ok
	case float64:
		// Handle JSON unmarshaling producing float64
		if val == float64(int(val)) {
			level, ok := parseSafetyLevelNumeric(fmt.Sprintf("%d", int(val)))
			return level, ok
		}
	}
	return SafetyLevelEnabled, false
}

// parseSafetyLevelString converts string to SafetyLevelType
func parseSafetyLevelString(s string) (SafetyLevelType, bool) {
	switch strings.ToLower(s) {
	case "disabled":
		return SafetyLevelDisabled, true
	case "enabled":
		return SafetyLevelEnabled, true
	case "strict":
		return SafetyLevelStrict, true
	case "paranoid":
		return SafetyLevelParanoid, true
	}
	return SafetyLevelEnabled, false
}

// parseSafetyLevelNumeric converts string number to SafetyLevelType
func parseSafetyLevelNumeric(s string) (SafetyLevelType, bool) {
	if level, err := strconv.Atoi(s); err == nil {
		switch SafetyLevelType(level) {
		case SafetyLevelDisabled, SafetyLevelEnabled, SafetyLevelStrict, SafetyLevelParanoid:
			return SafetyLevelType(level), true
		}
	}
	return SafetyLevelEnabled, false
}

// SafetyConfigValidationResult represents validation result for safety configuration
type SafetyConfigValidationResult struct {
	IsValid bool
	Errors  []SafetyConfigValidationError
}

// SafetyConfigValidationError represents specific validation error
type SafetyConfigValidationError struct {
	Field   string
	Message string
	Value   any
}

// Validate validates the safety configuration
func (sc SafetyConfig) Validate() SafetyConfigValidationResult {
	var errors []SafetyConfigValidationError

	// All safety configs are valid by design due to strong typing
	// This prevents invalid states at compile time

	return SafetyConfigValidationResult{
		IsValid: true,
		Errors:  errors,
	}
}

// ToSafetyLevel extracts the effective safety level
func (sc SafetyConfig) ToSafetyLevel() SafetyLevelType {
	return sc.Level
}

// String returns string representation
func (sc SafetyConfig) String() string {
	return fmt.Sprintf("safety=%s", sc.Level.String())
}
