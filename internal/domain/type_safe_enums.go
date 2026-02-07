package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAMLEnum is a generic helper for unmarshaling YAML node values to enum types.
// Accepts both string (case-insensitive) and integer representations.
func UnmarshalYAMLEnum[T ~int](
	value *yaml.Node,
	target *T,
	valueMap map[string]T,
	errorMsg string,
) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		upperKey := strings.ToUpper(s)
		for key, enumVal := range valueMap {
			if strings.ToUpper(key) == upperKey {
				*target = enumVal
				return nil
			}
		}
		return fmt.Errorf("%s: %s", errorMsg, s)
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		for _, enumVal := range valueMap {
			if int(enumVal) == i {
				*target = enumVal
				return nil
			}
		}
		return fmt.Errorf("%s value: %d", errorMsg, i)
	}

	return fmt.Errorf("cannot parse %s: expected string or int", errorMsg)
}

// UnmarshalYAMLEnumWithDefault is like UnmarshalYAMLEnum but returns a default value for invalid inputs.
func UnmarshalYAMLEnumWithDefault[T ~int](
	value *yaml.Node,
	target *T,
	valueMap map[string]T,
	defaultVal T,
	errorMsg string,
) T {
	var s string
	if err := value.Decode(&s); err == nil {
		upperKey := strings.ToUpper(s)
		for key, enumVal := range valueMap {
			if strings.ToUpper(key) == upperKey {
				return enumVal
			}
		}
		return defaultVal
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		for _, enumVal := range valueMap {
			if int(enumVal) == i {
				return enumVal
			}
		}
		return defaultVal
	}

	return defaultVal
}

// UnmarshalJSONEnum is a generic helper for unmarshaling JSON string values to enum types.
func UnmarshalJSONEnum[T any](
	data []byte,
	target *T,
	valueMap map[string]T,
	errorMsg string,
) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	upperKey := strings.ToUpper(s)
	for key, value := range valueMap {
		if strings.ToUpper(key) == upperKey {
			*target = value
			return nil
		}
	}
	return fmt.Errorf("%s: %s", errorMsg, s)
}

// TypeSafeEnum provides compile-time guaranteed enums with JSON serialization.
type TypeSafeEnum[T any] interface {
	String() string
	IsValid() bool
	Values() []T
}

// RiskLevelType represents the risk level enum with compile-time safety.
type RiskLevelType int

const (
	RiskLevelLowType RiskLevelType = iota
	RiskLevelMediumType
	RiskLevelHighType
	RiskLevelCriticalType
)

// String returns the string representation.
func (rl RiskLevelType) String() string {
	switch rl {
	case RiskLevelLowType:
		return "LOW"
	case RiskLevelMediumType:
		return "MEDIUM"
	case RiskLevelHighType:
		return "HIGH"
	case RiskLevelCriticalType:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if risk level is valid.
func (rl RiskLevelType) IsValid() bool {
	return rl >= RiskLevelLowType && rl <= RiskLevelCriticalType
}

// Values returns all possible values.
func (rl RiskLevelType) Values() []RiskLevelType {
	return []RiskLevelType{
		RiskLevelLowType,
		RiskLevelMediumType,
		RiskLevelHighType,
		RiskLevelCriticalType,
	}
}

// MarshalJSON implements json.Marshaler.
func (rl RiskLevelType) MarshalJSON() ([]byte, error) {
	if !rl.IsValid() {
		return nil, fmt.Errorf("invalid risk level: %d", rl)
	}
	return json.Marshal(rl.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONEnum(data, rl, map[string]RiskLevelType{
		"LOW":      RiskLevelLowType,
		"MEDIUM":   RiskLevelMediumType,
		"HIGH":     RiskLevelHighType,
		"CRITICAL": RiskLevelCriticalType,
	}, "invalid risk level")
}

// Icon returns emoji for risk level.
func (rl RiskLevelType) Icon() string {
	switch rl {
	case RiskLevelLowType:
		return "🟢"
	case RiskLevelMediumType:
		return "🟡"
	case RiskLevelHighType:
		return "🟠"
	case RiskLevelCriticalType:
		return "🔴"
	default:
		return "⚪"
	}
}

// IsHigherThan returns true if this risk level is higher than comparison.
func (rl RiskLevelType) IsHigherThan(other RiskLevelType) bool {
	return rl > other
}

// IsHigherOrEqualThan returns true if this risk level is higher or equal than comparison.
func (rl RiskLevelType) IsHigherOrEqualThan(other RiskLevelType) bool {
	return rl >= other
}

// Convenience constants for backward compatibility are now in types.go

// ValidationLevelType represents validation levels with compile-time safety.
type ValidationLevelType int

const (
	ValidationLevelNoneType ValidationLevelType = iota
	ValidationLevelBasicType
	ValidationLevelComprehensiveType
	ValidationLevelStrictType
)

// String returns the string representation.
func (vl ValidationLevelType) String() string {
	switch vl {
	case ValidationLevelNoneType:
		return "NONE"
	case ValidationLevelBasicType:
		return "BASIC"
	case ValidationLevelComprehensiveType:
		return "COMPREHENSIVE"
	case ValidationLevelStrictType:
		return "STRICT"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if validation level is valid.
func (vl ValidationLevelType) IsValid() bool {
	return vl >= ValidationLevelNoneType && vl <= ValidationLevelStrictType
}

// Values returns all possible values.
func (vl ValidationLevelType) Values() []ValidationLevelType {
	return []ValidationLevelType{
		ValidationLevelNoneType,
		ValidationLevelBasicType,
		ValidationLevelComprehensiveType,
		ValidationLevelStrictType,
	}
}

// MarshalJSON implements json.Marshaler.
func (vl ValidationLevelType) MarshalJSON() ([]byte, error) {
	if !vl.IsValid() {
		return nil, fmt.Errorf("invalid validation level: %d", vl)
	}
	return json.Marshal(vl.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONEnum(data, vl, map[string]ValidationLevelType{
		"NONE":          ValidationLevelNoneType,
		"BASIC":         ValidationLevelBasicType,
		"COMPREHENSIVE": ValidationLevelComprehensiveType,
		"STRICT":        ValidationLevelStrictType,
	}, "invalid validation level")
}

// Convenience constants for backward compatibility are now in types.go

// ChangeOperationType represents change operations with compile-time safety.
type ChangeOperationType int

const (
	ChangeOperationAddedType ChangeOperationType = iota
	ChangeOperationRemovedType
	ChangeOperationModifiedType
)

// String returns the string representation.
func (co ChangeOperationType) String() string {
	switch co {
	case ChangeOperationAddedType:
		return "ADDED"
	case ChangeOperationRemovedType:
		return "REMOVED"
	case ChangeOperationModifiedType:
		return "MODIFIED"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if change operation is valid.
func (co ChangeOperationType) IsValid() bool {
	return co >= ChangeOperationAddedType && co <= ChangeOperationModifiedType
}

// Values returns all possible values.
func (co ChangeOperationType) Values() []ChangeOperationType {
	return []ChangeOperationType{
		ChangeOperationAddedType,
		ChangeOperationRemovedType,
		ChangeOperationModifiedType,
	}
}

// MarshalJSON implements json.Marshaler.
func (co ChangeOperationType) MarshalJSON() ([]byte, error) {
	if !co.IsValid() {
		return nil, fmt.Errorf("invalid change operation: %d", co)
	}
	return json.Marshal(co.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONEnum(data, co, map[string]ChangeOperationType{
		"ADDED":    ChangeOperationAddedType,
		"REMOVED":  ChangeOperationRemovedType,
		"MODIFIED": ChangeOperationModifiedType,
	}, "invalid change operation")
}

// CleanStrategyType represents cleaning strategies with compile-time safety.
type CleanStrategyType int

const (
	StrategyAggressiveType CleanStrategyType = iota
	StrategyConservativeType
	StrategyDryRunType
)

// String returns string representation.
func (cs CleanStrategyType) String() string {
	switch cs {
	case StrategyAggressiveType:
		return "aggressive"
	case StrategyConservativeType:
		return "conservative"
	case StrategyDryRunType:
		return "dry-run"
	default:
		return "unknown"
	}
}

// IsValid checks if clean strategy is valid.
func (cs CleanStrategyType) IsValid() bool {
	return cs >= StrategyAggressiveType && cs <= StrategyDryRunType
}

// Values returns all possible values.
func (cs CleanStrategyType) Values() []CleanStrategyType {
	return []CleanStrategyType{
		StrategyAggressiveType,
		StrategyConservativeType,
		StrategyDryRunType,
	}
}

// MarshalJSON implements json.Marshaler.
func (cs CleanStrategyType) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf("invalid clean strategy: %d", cs)
	}
	return json.Marshal(cs.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONEnum(data, cs, map[string]CleanStrategyType{
		"aggressive":   StrategyAggressiveType,
		"conservative": StrategyConservativeType,
		"dry-run":      StrategyDryRunType,
		"dryrun":       StrategyDryRunType,
	}, "invalid clean strategy")
}

// Icon returns emoji for clean strategy.
func (cs CleanStrategyType) Icon() string {
	switch cs {
	case StrategyAggressiveType:
		return "🔥"
	case StrategyConservativeType:
		return "🛡️"
	case StrategyDryRunType:
		return "🔍"
	default:
		return "❓"
	}
}

// Convenience constants for backward compatibility are now in types.go

// UnmarshalYAML implements yaml.Unmarshaler for RiskLevelType.
func (rl *RiskLevelType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		// If string unmarshaling fails, try integer
		var i int
		if err := value.Decode(&i); err == nil {
			*rl = RiskLevelType(i)
			return nil
		}
		return err
	}

	switch strings.ToUpper(s) {
	case "LOW":
		*rl = RiskLevelLowType
	case "MEDIUM":
		*rl = RiskLevelMediumType
	case "HIGH":
		*rl = RiskLevelHighType
	case "CRITICAL":
		*rl = RiskLevelCriticalType
	default:
		return fmt.Errorf("invalid risk level: %s (must be LOW, MEDIUM, HIGH, or CRITICAL)", s)
	}
	return nil
}

// MarshalYAML implements yaml.Marshaler for RiskLevelType.
func (rl RiskLevelType) MarshalYAML() (any, error) {
	if !rl.IsValid() {
		return nil, fmt.Errorf("invalid risk level: %d", rl)
	}
	return rl.String(), nil
}
