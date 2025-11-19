package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TypeSafeEnum provides compile-time guaranteed enums with JSON serialization
type TypeSafeEnum[T any] interface {
	String() string
	IsValid() bool
	Values() []T
}

// EnumHelper provides generic enum functionality to reduce code duplication
type EnumHelper[T ~int] struct {
	stringValues map[T]string
	validRange   func(T) bool
	allValues    func() []T
	caseSensitive bool
}

// NewEnumHelper creates a new helper for enum type-safe operations
func NewEnumHelper[T ~int](stringValues map[T]string, validRange func(T) bool, allValues func() []T, caseSensitive bool) *EnumHelper[T] {
	return &EnumHelper[T]{
		stringValues:   stringValues,
		validRange:     validRange,
		allValues:      allValues,
		caseSensitive: caseSensitive,
	}
}

// String returns string representation for enum value
func (eh *EnumHelper[T]) String(value T) string {
	if str, exists := eh.stringValues[value]; exists {
		return str
	}
	return "UNKNOWN"
}

// IsValid checks if enum value is valid
func (eh *EnumHelper[T]) IsValid(value T) bool {
	return eh.validRange(value)
}

// Values returns all possible enum values
func (eh *EnumHelper[T]) Values() []T {
	return eh.allValues()
}

// MarshalJSON converts enum to JSON string
func (eh *EnumHelper[T]) MarshalJSON(value T) ([]byte, error) {
	if !eh.IsValid(value) {
		return nil, fmt.Errorf("invalid enum value: %v", value)
	}
	return json.Marshal(eh.String(value))
}

// UnmarshalJSON converts JSON string to enum value
func (eh *EnumHelper[T]) UnmarshalJSON(data []byte, valueSetter func(T)) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	// Find matching enum value by string
	for enumVal, strVal := range eh.stringValues {
		compareValue := s
		strToCompare := strVal
		
		if !eh.caseSensitive {
			compareValue = strings.ToLower(s)
			strToCompare = strings.ToLower(strVal)
		}
		
		if compareValue == strToCompare {
			valueSetter(enumVal)
			return nil
		}
	}
	
	return fmt.Errorf("invalid enum value: %s", s)
}

// RiskLevelType represents the risk level enum with compile-time safety
type RiskLevelType int

const (
	RiskLevelLowType RiskLevelType = iota
	RiskLevelMediumType
	RiskLevelHighType
	RiskLevelCriticalType
)

// riskLevelHelper provides shared functionality for RiskLevelType
var riskLevelHelper = NewEnumHelper(map[RiskLevelType]string{
	RiskLevelLowType:      "LOW",
	RiskLevelMediumType:    "MEDIUM",
	RiskLevelHighType:      "HIGH",
	RiskLevelCriticalType:  "CRITICAL",
}, func(rl RiskLevelType) bool {
	return rl >= RiskLevelLowType && rl <= RiskLevelCriticalType
}, func() []RiskLevelType {
	return []RiskLevelType{
		RiskLevelLowType,
		RiskLevelMediumType,
		RiskLevelHighType,
		RiskLevelCriticalType,
	}
}, true) // case sensitive for risk levels

// String returns the string representation
func (rl RiskLevelType) String() string {
	return riskLevelHelper.String(rl)
}

// IsValid checks if risk level is valid
func (rl RiskLevelType) IsValid() bool {
	return riskLevelHelper.IsValid(rl)
}

// Values returns all possible values
func (rl RiskLevelType) Values() []RiskLevelType {
	return riskLevelHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (rl RiskLevelType) MarshalJSON() ([]byte, error) {
	return riskLevelHelper.MarshalJSON(rl)
}

// UnmarshalJSON implements json.Unmarshaler
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error {
	return riskLevelHelper.UnmarshalJSON(data, func(val RiskLevelType) {
		*rl = val
	})
}

// Icon returns emoji for risk level (UI CONCERN - SHOULD BE MOVED TO ADAPTER LAYER)
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

// IsHigherThan returns true if this risk level is higher than comparison
func (rl RiskLevelType) IsHigherThan(other RiskLevelType) bool {
	return rl > other
}

// IsHigherOrEqualThan returns true if this risk level is higher or equal than comparison
func (rl RiskLevelType) IsHigherOrEqualThan(other RiskLevelType) bool {
	return rl >= other
}

// Convenience constants for backward compatibility are now in types.go

// ValidationLevelType represents validation levels with compile-time safety
type ValidationLevelType int

const (
	ValidationLevelNoneType ValidationLevelType = iota
	ValidationLevelBasicType
	ValidationLevelComprehensiveType
	ValidationLevelStrictType
)

// validationLevelHelper provides shared functionality for ValidationLevelType
var validationLevelHelper = NewEnumHelper(map[ValidationLevelType]string{
	ValidationLevelNoneType:          "NONE",
	ValidationLevelBasicType:         "BASIC",
	ValidationLevelComprehensiveType: "COMPREHENSIVE",
	ValidationLevelStrictType:        "STRICT",
}, func(vl ValidationLevelType) bool {
	return vl >= ValidationLevelNoneType && vl <= ValidationLevelStrictType
}, func() []ValidationLevelType {
	return []ValidationLevelType{
		ValidationLevelNoneType,
		ValidationLevelBasicType,
		ValidationLevelComprehensiveType,
		ValidationLevelStrictType,
	}
}, true) // case sensitive for validation levels

// String returns the string representation
func (vl ValidationLevelType) String() string {
	return validationLevelHelper.String(vl)
}

// IsValid checks if validation level is valid
func (vl ValidationLevelType) IsValid() bool {
	return validationLevelHelper.IsValid(vl)
}

// Values returns all possible values
func (vl ValidationLevelType) Values() []ValidationLevelType {
	return validationLevelHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (vl ValidationLevelType) MarshalJSON() ([]byte, error) {
	return validationLevelHelper.MarshalJSON(vl)
}

// UnmarshalJSON implements json.Unmarshaler
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	return validationLevelHelper.UnmarshalJSON(data, func(val ValidationLevelType) {
		*vl = val
	})
}

// Convenience constants for backward compatibility are now in types.go

// ChangeOperationType represents change operations with compile-time safety
type ChangeOperationType int

const (
	ChangeOperationAddedType ChangeOperationType = iota
	ChangeOperationRemovedType
	ChangeOperationModifiedType
)

// changeOperationHelper provides shared functionality for ChangeOperationType
var changeOperationHelper = NewEnumHelper(map[ChangeOperationType]string{
	ChangeOperationAddedType:    "ADDED",
	ChangeOperationRemovedType:  "REMOVED",
	ChangeOperationModifiedType: "MODIFIED",
}, func(co ChangeOperationType) bool {
	return co >= ChangeOperationAddedType && co <= ChangeOperationModifiedType
}, func() []ChangeOperationType {
	return []ChangeOperationType{
		ChangeOperationAddedType,
		ChangeOperationRemovedType,
		ChangeOperationModifiedType,
	}
}, true) // case sensitive for change operations

// String returns the string representation
func (co ChangeOperationType) String() string {
	return changeOperationHelper.String(co)
}

// IsValid checks if change operation is valid
func (co ChangeOperationType) IsValid() bool {
	return changeOperationHelper.IsValid(co)
}

// Values returns all possible values
func (co ChangeOperationType) Values() []ChangeOperationType {
	return changeOperationHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (co ChangeOperationType) MarshalJSON() ([]byte, error) {
	return changeOperationHelper.MarshalJSON(co)
}

// UnmarshalJSON implements json.Unmarshaler
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	return changeOperationHelper.UnmarshalJSON(data, func(val ChangeOperationType) {
		*co = val
	})
}

// CleanStrategyType represents cleaning strategies with compile-time safety
type CleanStrategyType int

const (
	StrategyAggressiveType CleanStrategyType = iota
	StrategyConservativeType
	StrategyDryRunType
)

// cleanStrategyHelper provides shared functionality for CleanStrategyType
var cleanStrategyHelper = NewEnumHelper(map[CleanStrategyType]string{
	StrategyAggressiveType:   "aggressive",
	StrategyConservativeType: "conservative", 
	StrategyDryRunType:       "dry-run",
}, func(cs CleanStrategyType) bool {
	return cs >= StrategyAggressiveType && cs <= StrategyDryRunType
}, func() []CleanStrategyType {
	return []CleanStrategyType{
		StrategyAggressiveType,
		StrategyConservativeType,
		StrategyDryRunType,
	}
}, false) // case insensitive for strategies (accept "dryrun", "dry-run")

// String returns string representation
func (cs CleanStrategyType) String() string {
	return cleanStrategyHelper.String(cs)
}

// IsValid checks if clean strategy is valid
func (cs CleanStrategyType) IsValid() bool {
	return cleanStrategyHelper.IsValid(cs)
}

// Values returns all possible values
func (cs CleanStrategyType) Values() []CleanStrategyType {
	return cleanStrategyHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (cs CleanStrategyType) MarshalJSON() ([]byte, error) {
	return cleanStrategyHelper.MarshalJSON(cs)
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	return cleanStrategyHelper.UnmarshalJSON(data, func(val CleanStrategyType) {
		*cs = val
	})
}

// Icon returns emoji for clean strategy (UI CONCERN - SHOULD BE MOVED TO ADAPTER LAYER)
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

// ScanTypeType represents scanning domains with compile-time safety
type ScanTypeType int

const (
	ScanTypeNixStoreType ScanTypeType = iota
	ScanTypeHomebrewType
	ScanTypeSystemType
	ScanTypeTempType
)

// scanTypeHelper provides shared functionality for ScanTypeType
var scanTypeHelper = NewEnumHelper(map[ScanTypeType]string{
	ScanTypeNixStoreType: "nix_store",
	ScanTypeHomebrewType: "homebrew",
	ScanTypeSystemType:   "system",
	ScanTypeTempType:     "temp_files",
}, func(st ScanTypeType) bool {
	return st >= ScanTypeNixStoreType && st <= ScanTypeTempType
}, func() []ScanTypeType {
	return []ScanTypeType{
		ScanTypeNixStoreType,
		ScanTypeHomebrewType,
		ScanTypeSystemType,
		ScanTypeTempType,
	}
}, false) // case insensitive for scan types (accept "nix-store", "nix_store", "temp", etc.)

// String returns string representation
func (st ScanTypeType) String() string {
	return scanTypeHelper.String(st)
}

// IsValid checks if scan type is valid
func (st ScanTypeType) IsValid() bool {
	return scanTypeHelper.IsValid(st)
}

// Values returns all possible values
func (st ScanTypeType) Values() []ScanTypeType {
	return scanTypeHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (st ScanTypeType) MarshalJSON() ([]byte, error) {
	return scanTypeHelper.MarshalJSON(st)
}

// UnmarshalJSON implements json.Unmarshaler
func (st *ScanTypeType) UnmarshalJSON(data []byte) error {
	return scanTypeHelper.UnmarshalJSON(data, func(val ScanTypeType) {
		*st = val
	})
}

// ScanTypeIcon returns emoji for scan type (UI CONCERN - SHOULD BE MOVED TO ADAPTER LAYER)
func (st ScanTypeType) Icon() string {
	switch st {
	case ScanTypeNixStoreType:
		return "📦"
	case ScanTypeHomebrewType:
		return "🍺"
	case ScanTypeSystemType:
		return "💻"
	case ScanTypeTempType:
		return "🗑️"
	default:
		return "❓"
	}
}

// Convenience constants for backward compatibility are now in types.go
