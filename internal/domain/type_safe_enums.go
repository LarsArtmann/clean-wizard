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
			compareValue = strings.ToUpper(s)
			strToCompare = strVal
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

// String returns the string representation
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

// IsValid checks if validation level is valid
func (vl ValidationLevelType) IsValid() bool {
	return vl >= ValidationLevelNoneType && vl <= ValidationLevelStrictType
}

// Values returns all possible values
func (vl ValidationLevelType) Values() []ValidationLevelType {
	return []ValidationLevelType{
		ValidationLevelNoneType,
		ValidationLevelBasicType,
		ValidationLevelComprehensiveType,
		ValidationLevelStrictType,
	}
}

// MarshalJSON implements json.Marshaler
func (vl ValidationLevelType) MarshalJSON() ([]byte, error) {
	if !vl.IsValid() {
		return nil, fmt.Errorf("invalid validation level: %d", vl)
	}
	return json.Marshal(vl.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToUpper(s) {
	case "NONE":
		*vl = ValidationLevelNoneType
	case "BASIC":
		*vl = ValidationLevelBasicType
	case "COMPREHENSIVE":
		*vl = ValidationLevelComprehensiveType
	case "STRICT":
		*vl = ValidationLevelStrictType
	default:
		return fmt.Errorf("invalid validation level: %s", s)
	}
	return nil
}

// Convenience constants for backward compatibility are now in types.go

// ChangeOperationType represents change operations with compile-time safety
type ChangeOperationType int

const (
	ChangeOperationAddedType ChangeOperationType = iota
	ChangeOperationRemovedType
	ChangeOperationModifiedType
)

// String returns the string representation
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

// IsValid checks if change operation is valid
func (co ChangeOperationType) IsValid() bool {
	return co >= ChangeOperationAddedType && co <= ChangeOperationModifiedType
}

// Values returns all possible values
func (co ChangeOperationType) Values() []ChangeOperationType {
	return []ChangeOperationType{
		ChangeOperationAddedType,
		ChangeOperationRemovedType,
		ChangeOperationModifiedType,
	}
}

// MarshalJSON implements json.Marshaler
func (co ChangeOperationType) MarshalJSON() ([]byte, error) {
	if !co.IsValid() {
		return nil, fmt.Errorf("invalid change operation: %d", co)
	}
	return json.Marshal(co.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToUpper(s) {
	case "ADDED":
		*co = ChangeOperationAddedType
	case "REMOVED":
		*co = ChangeOperationRemovedType
	case "MODIFIED":
		*co = ChangeOperationModifiedType
	default:
		return fmt.Errorf("invalid change operation: %s", s)
	}
	return nil
}

// CleanStrategyType represents cleaning strategies with compile-time safety
type CleanStrategyType int

const (
	StrategyAggressiveType CleanStrategyType = iota
	StrategyConservativeType
	StrategyDryRunType
)

// String returns string representation
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

// IsValid checks if clean strategy is valid
func (cs CleanStrategyType) IsValid() bool {
	return cs >= StrategyAggressiveType && cs <= StrategyDryRunType
}

// Values returns all possible values
func (cs CleanStrategyType) Values() []CleanStrategyType {
	return []CleanStrategyType{
		StrategyAggressiveType,
		StrategyConservativeType,
		StrategyDryRunType,
	}
}

// MarshalJSON implements json.Marshaler
func (cs CleanStrategyType) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf("invalid clean strategy: %d", cs)
	}
	return json.Marshal(cs.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "aggressive":
		*cs = StrategyAggressiveType
	case "conservative":
		*cs = StrategyConservativeType
	case "dry-run", "dryrun":
		*cs = StrategyDryRunType
	default:
		return fmt.Errorf("invalid clean strategy: %s", s)
	}
	return nil
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

// String returns string representation
func (st ScanTypeType) String() string {
	switch st {
	case ScanTypeNixStoreType:
		return "nix_store"
	case ScanTypeHomebrewType:
		return "homebrew"
	case ScanTypeSystemType:
		return "system"
	case ScanTypeTempType:
		return "temp_files"
	default:
		return "unknown"
	}
}

// IsValid checks if scan type is valid
func (st ScanTypeType) IsValid() bool {
	return st >= ScanTypeNixStoreType && st <= ScanTypeTempType
}

// Values returns all possible values
func (st ScanTypeType) Values() []ScanTypeType {
	return []ScanTypeType{
		ScanTypeNixStoreType,
		ScanTypeHomebrewType,
		ScanTypeSystemType,
		ScanTypeTempType,
	}
}

// MarshalJSON implements json.Marshaler
func (st ScanTypeType) MarshalJSON() ([]byte, error) {
	if !st.IsValid() {
		return nil, fmt.Errorf("invalid scan type: %d", st)
	}
	return json.Marshal(st.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (st *ScanTypeType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "nix_store", "nix-store":
		*st = ScanTypeNixStoreType
	case "homebrew":
		*st = ScanTypeHomebrewType
	case "system":
		*st = ScanTypeSystemType
	case "temp_files", "temp-files", "temp":
		*st = ScanTypeTempType
	default:
		return fmt.Errorf("invalid scan type: %s", s)
	}
	return nil
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
