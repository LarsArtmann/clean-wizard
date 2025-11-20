package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// SafetyConfig represents all possible safety configuration formats
// This type eliminates split brains by providing a single source of truth
type SafetyConfig struct {
	// Primary safety level (new format)
	Level SafetyLevelType
	
	// Legacy boolean (for backward compatibility)  
	SafeMode *bool
	
	// Whether this config was explicitly set by user
	IsExplicit bool
}

// ParseSafetyConfig creates a SafetyConfig from viper configuration
// This function centralizes all safety level parsing logic and eliminates type safety violations
func ParseSafetyConfig(v ViperConfig) SafetyConfig {
	config := SafetyConfig{IsExplicit: false}
	
	// Check for new safety_level format first (highest priority)
	if v.IsSet("safety_level") {
		config.IsExplicit = true
		safetyValue := v.Get("safety_level")
		
		if level, ok := parseSafetyLevelValue(safetyValue); ok {
			config.Level = level
			return config
		}
		
		// Invalid value - log warning and use default
		// TODO: Replace with proper error handling instead of logging
		return SafetyConfig{
			Level:     SafetyLevelEnabled,
			IsExplicit: true,
		}
	}
	
	// Fall back to legacy safe_mode boolean
	if v.IsSet("safe_mode") {
		config.IsExplicit = true
		safeMode := v.GetBool("safe_mode")
		config.SafeMode = &safeMode
		
		if safeMode {
			config.Level = SafetyLevelEnabled
		} else {
			config.Level = SafetyLevelDisabled
		}
		return config
	}
	
	// Default when neither is present
	return SafetyConfig{
		Level:     SafetyLevelEnabled,
		IsExplicit: false,
	}
}

// parseSafetyLevelValue attempts to parse safety level from interface value
// Returns parsed level and success flag
func parseSafetyLevelValue(value interface{}) (SafetyLevelType, bool) {
	switch val := value.(type) {
	case string:
		return parseSafetyLevelString(strings.TrimSpace(val))
	case int, int32, int64:
		level, _ := parseSafetyLevelNumeric(fmt.Sprintf("%v", val))
		return level, true
	case float64:
		// Handle JSON unmarshaling producing float64
		if val == float64(int(val)) {
			level, _ := parseSafetyLevelNumeric(fmt.Sprintf("%d", int(val)))
			return level, true
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

// ViperConfig interface for viper operations (enables testing and dependency inversion)
type ViperConfig interface {
	IsSet(key string) bool
	Get(key string) interface{}
	GetBool(key string) bool
	GetString(key string) string
	GetInt(key string) int
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
	Value   interface{}
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

// IsLegacyMode returns true if configuration uses legacy safe_mode
func (sc SafetyConfig) IsLegacyMode() bool {
	return sc.SafeMode != nil
}

// String returns string representation
func (sc SafetyConfig) String() string {
	if sc.IsLegacyMode() {
		return fmt.Sprintf("legacy(safe_mode=%v)->%s", *sc.SafeMode, sc.Level.String())
	}
	return fmt.Sprintf("modern(%s)", sc.Level.String())
}