package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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
	Level shared.SafetyLevelType

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
			Level:      shared.SafetyLevelSafeType,
			IsExplicit: true,
		}
	}

	// Fall back to legacy safe_mode boolean
	if v.IsSet("safe_mode") {
		safeMode := v.GetBool("safe_mode")

		if safeMode {
			return SafetyConfig{
				Level:      shared.SafetyLevelSafeType,
				IsExplicit: true,
			}
		}
		return SafetyConfig{
			Level:      shared.SafetyLevelUnsafeType,
			IsExplicit: true,
		}
	}

	// Default when neither is present
	return SafetyConfig{
		Level:      shared.SafetyLevelSafeType,
		IsExplicit: false,
	}
}

// parseSafetyLevelValue attempts to parse safety level from interface value
// Returns parsed level and success flag
func parseSafetyLevelValue(value any) (shared.SafetyLevelType, bool) {
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
	return shared.SafetyLevelSafeType, false
}

// parseSafetyLevelString converts string to shared.SafetyLevelType
func parseSafetyLevelString(s string) (shared.SafetyLevelType, bool) {
	switch strings.ToLower(s) {
	case "disabled":
		return shared.SafetyLevelUnsafeType, true
	case "enabled":
		return shared.SafetyLevelSafeType, true
	case "strict":
		return shared.SafetyLevelStrictType, true
	case "paranoid":
		return shared.SafetyLevelParanoidType, true
	}
	return shared.SafetyLevelSafeType, false
}

// parseSafetyLevelNumeric converts string number to shared.SafetyLevelType
func parseSafetyLevelNumeric(s string) (shared.SafetyLevelType, bool) {
	if level, err := strconv.Atoi(s); err == nil {
		switch shared.SafetyLevelType(level) {
		case shared.SafetyLevelUnsafeType, shared.SafetyLevelSafeType, shared.SafetyLevelStrictType, shared.SafetyLevelParanoidType:
			return shared.SafetyLevelType(level), true
		}
	}
	return shared.SafetyLevelSafeType, false
}

// Validate validates the safety configuration
func (sc SafetyConfig) Validate() error {
	// SafetyConfig uses strong typing which prevents invalid states at compile time
	// The shared.SafetyLevelType enum ensures only valid values can be assigned
	// Currently, no runtime validation is needed for the existing fields
	return nil
}

// ToSafetyLevel extracts the effective safety level
func (sc SafetyConfig) ToSafetyLevel() shared.SafetyLevelType {
	return sc.Level
}

// String returns string representation
func (sc SafetyConfig) String() string {
	return fmt.Sprintf("safety=%s", sc.Level.String())
}
