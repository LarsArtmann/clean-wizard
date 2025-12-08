package shared

//go:generate stringer -type=RiskLevelType
//go:generate stringer -type=StatusType
//go:generate stringer -type=OperationNameType
//go:generate stringer -type=OptimizationLevelType
//go:generate stringer -type=ExecutionModeType
//go:generate stringer -type=StrategyType
//go:generate stringer -type=ChangeOperationType
//go:generate stringer -type=ScanTypeType
//go:generate stringer -type=ValidationLevelType
//go:generate stringer -type=EnforcementLevelType
//go:generate stringer -type=SelectedStatusType
//go:generate stringer -type=RecursionLevelType
//go:generate stringer -type=FileSelectionStrategyType
//go:generate stringer -type=SafetyLevelType

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// EnumHelper provides generic enum functionality to reduce code duplication
type EnumHelper[T ~int] struct {
	stringValues  map[T]string
	validRange    func(T) bool
	allValues     func() []T
	caseSensitive bool
	valuesCache   []T       // Cache for Values()
	once          sync.Once // For one-time initialization
}

// NewEnumHelper creates a new helper for enum type-safe operations
func NewEnumHelper[T ~int](stringValues map[T]string, validRange func(T) bool, allValues func() []T, caseSensitive bool) *EnumHelper[T] {
	return &EnumHelper[T]{
		stringValues:  stringValues,
		validRange:    validRange,
		allValues:     allValues,
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

// Values returns all possible enum values (cached for performance)
func (eh *EnumHelper[T]) Values() []T {
	eh.once.Do(func() {
		eh.valuesCache = append([]T(nil), eh.allValues()...)
	})
	return append([]T(nil), eh.valuesCache...)
}

// MarshalJSON converts enum to JSON string - REMOVED: Use individual type implementations

// MarshalJSONImpl converts enum to JSON string - internal implementation
func (eh *EnumHelper[T]) MarshalJSONImpl(value T) ([]byte, error) {
	if !eh.IsValid(value) {
		return nil, fmt.Errorf("invalid enum value: %v", value)
	}
	return json.Marshal(eh.String(value))
}

// UnmarshalJSONImpl converts JSON string to enum value - internal implementation
func (eh *EnumHelper[T]) UnmarshalJSONImpl(data []byte) (T, error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var zero T
		return zero, err
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
			return enumVal, nil
		}
	}

	var zero T
	return zero, fmt.Errorf("invalid enum value: %s", s)
}

// RiskLevelType represents the risk level enum with compile-time safety
type RiskLevelType int

const (
	RiskLevelLowType RiskLevelType = iota
	RiskLevelMediumType
	RiskLevelHighType
	RiskLevelCriticalType
)

// OperationNameType represents operation names with compile-time safety
type OperationNameType int

const (
	OperationNameNixGenerations OperationNameType = iota
	OperationNameHomebrew
	OperationNamePackageCache
	OperationNameTempFiles
	OperationNameSystemTemp
)

// riskLevelHelper provides shared functionality for RiskLevelType
var riskLevelHelper = NewEnumHelper(map[RiskLevelType]string{
	RiskLevelLowType:      "LOW",
	RiskLevelMediumType:   "MEDIUM",
	RiskLevelHighType:     "HIGH",
	RiskLevelCriticalType: "CRITICAL",
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
	return riskLevelHelper.MarshalJSONImpl(rl)
}

// UnmarshalJSON implements json.Unmarshaler
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error {
	val, err := riskLevelHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*rl = val
	return nil
}

// operationNameHelper provides shared functionality for OperationNameType
var operationNameHelper = NewEnumHelper(map[OperationNameType]string{
	OperationNameNixGenerations: "nix-generations",
	OperationNameHomebrew:       "homebrew",
	OperationNamePackageCache:   "package-cache",
	OperationNameTempFiles:      "temp-files",
	OperationNameSystemTemp:     "system-temp",
}, func(on OperationNameType) bool {
	return on >= OperationNameNixGenerations && on <= OperationNameSystemTemp
}, func() []OperationNameType {
	return []OperationNameType{
		OperationNameNixGenerations,
		OperationNameHomebrew,
		OperationNamePackageCache,
		OperationNameTempFiles,
		OperationNameSystemTemp,
	}
}, true) // case sensitive for operation names

// String returns string representation
func (on OperationNameType) String() string {
	return operationNameHelper.String(on)
}

// IsValid checks if operation name is valid
func (on OperationNameType) IsValid() bool {
	return operationNameHelper.IsValid(on)
}

// Values returns all possible values
func (on OperationNameType) Values() []OperationNameType {
	return operationNameHelper.Values()
}

// MarshalJSON implements json.Marshaler
func (on OperationNameType) MarshalJSON() ([]byte, error) {
	return operationNameHelper.MarshalJSONImpl(on)
}

// UnmarshalJSON implements json.Unmarshaler
func (on *OperationNameType) UnmarshalJSON(data []byte) error {
	val, err := operationNameHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*on = val
	return nil
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
	return validationLevelHelper.MarshalJSONImpl(vl)
}

// UnmarshalJSON implements json.Unmarshaler
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	val, err := validationLevelHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*vl = val
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
	return changeOperationHelper.MarshalJSONImpl(co)
}

// UnmarshalJSON implements json.Unmarshaler
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	val, err := changeOperationHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*co = val
	return nil
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
	return cleanStrategyHelper.MarshalJSONImpl(cs)
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	val, err := cleanStrategyHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*cs = val
	return nil
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

// StatusType represents enabled status of profiles and operations
// Replaces boolean Enabled to eliminate invalid states
type StatusType int

const (
	StatusDisabled StatusType = iota
	StatusEnabled
	StatusInherited
)

// statusTypeHelper provides shared functionality for StatusType
var statusTypeHelper = NewEnumHelper(map[StatusType]string{
	StatusDisabled:  "disabled",
	StatusEnabled:   "enabled",
	StatusInherited: "inherited",
}, func(s StatusType) bool {
	return s >= StatusDisabled && s <= StatusInherited
}, func() []StatusType {
	return []StatusType{StatusDisabled, StatusEnabled, StatusInherited}
}, false)

// String returns string representation
func (s StatusType) String() string {
	return statusTypeHelper.String(s)
}

// IsValid checks if status is valid
func (s StatusType) IsValid() bool {
	return statusTypeHelper.IsValid(s)
}

// Values returns all possible status values
func (s StatusType) Values() []StatusType {
	return statusTypeHelper.Values()
}

// MarshalJSON converts status to JSON string
func (s StatusType) MarshalJSON() ([]byte, error) {
	return statusTypeHelper.MarshalJSONImpl(s)
}

// UnmarshalJSON converts JSON string to status
func (s *StatusType) UnmarshalJSON(data []byte) error {
	val, err := statusTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*s = val
	return nil
}

// EnforcementLevelType represents validation strictness levels
// Replaces boolean RequireSafeMode to eliminate invalid states
type EnforcementLevelType int

const (
	EnforcementLevelNone EnforcementLevelType = iota
	EnforcementLevelWarning
	EnforcementLevelError
	EnforcementLevelStrict
)

// enforcementLevelTypeHelper provides shared functionality for EnforcementLevelType
var enforcementLevelTypeHelper = NewEnumHelper(map[EnforcementLevelType]string{
	EnforcementLevelNone:    "none",
	EnforcementLevelWarning: "warning",
	EnforcementLevelError:   "error",
	EnforcementLevelStrict:  "strict",
}, func(el EnforcementLevelType) bool {
	return el >= EnforcementLevelNone && el <= EnforcementLevelStrict
}, func() []EnforcementLevelType {
	return []EnforcementLevelType{EnforcementLevelNone, EnforcementLevelWarning, EnforcementLevelError, EnforcementLevelStrict}
}, false)

// String returns string representation
func (el EnforcementLevelType) String() string {
	return enforcementLevelTypeHelper.String(el)
}

// IsValid checks if enforcement level is valid
func (el EnforcementLevelType) IsValid() bool {
	return enforcementLevelTypeHelper.IsValid(el)
}

// Values returns all possible enforcement level values
func (el EnforcementLevelType) Values() []EnforcementLevelType {
	return enforcementLevelTypeHelper.Values()
}

// MarshalJSON converts enforcement level to JSON string
func (el EnforcementLevelType) MarshalJSON() ([]byte, error) {
	return enforcementLevelTypeHelper.MarshalJSONImpl(el)
}

// UnmarshalJSON converts JSON string to enforcement level
func (el *EnforcementLevelType) UnmarshalJSON(data []byte) error {
	val, err := enforcementLevelTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*el = val
	return nil
}

// SelectedStatusType represents selection status of operations
// Replaces boolean Current to eliminate invalid states
type SelectedStatusType int

const (
	SelectedStatusNotSelected SelectedStatusType = iota
	SelectedStatusSelected
	SelectedStatusDefault
)

// selectedStatusTypeHelper provides shared functionality for SelectedStatusType
var selectedStatusTypeHelper = NewEnumHelper(map[SelectedStatusType]string{
	SelectedStatusNotSelected: "not_selected",
	SelectedStatusSelected:    "selected",
	SelectedStatusDefault:     "default",
}, func(ss SelectedStatusType) bool {
	return ss >= SelectedStatusNotSelected && ss <= SelectedStatusDefault
}, func() []SelectedStatusType {
	return []SelectedStatusType{SelectedStatusNotSelected, SelectedStatusSelected, SelectedStatusDefault}
}, false)

// String returns string representation
func (ss SelectedStatusType) String() string {
	return selectedStatusTypeHelper.String(ss)
}

// IsValid checks if selected status is valid
func (ss SelectedStatusType) IsValid() bool {
	return selectedStatusTypeHelper.IsValid(ss)
}

// Values returns all possible selected status values
func (ss SelectedStatusType) Values() []SelectedStatusType {
	return selectedStatusTypeHelper.Values()
}

// FromBool converts a boolean to SelectedStatusType
// true -> SelectedStatusSelected, false -> SelectedStatusNotSelected
func FromBool(current bool) SelectedStatusType {
	if current {
		return SelectedStatusSelected
	}
	return SelectedStatusNotSelected
}

// MarshalJSON converts selected status to JSON string
func (ss SelectedStatusType) MarshalJSON() ([]byte, error) {
	return selectedStatusTypeHelper.MarshalJSONImpl(ss)
}

// UnmarshalJSON converts JSON string to selected status
func (ss *SelectedStatusType) UnmarshalJSON(data []byte) error {
	val, err := selectedStatusTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*ss = val
	return nil
}

// RecursionLevelType represents recursion levels for scanning
// Replaces boolean Recursive to eliminate invalid states
type RecursionLevelType int

const (
	RecursionLevelNone RecursionLevelType = iota
	RecursionLevelDirect
	RecursionLevelFull
	RecursionLevelInfinite
)

// recursionLevelTypeHelper provides shared functionality for RecursionLevelType
var recursionLevelTypeHelper = NewEnumHelper(map[RecursionLevelType]string{
	RecursionLevelNone:     "none",
	RecursionLevelDirect:   "direct",
	RecursionLevelFull:     "full",
	RecursionLevelInfinite: "infinite",
}, func(rl RecursionLevelType) bool {
	return rl >= RecursionLevelNone && rl <= RecursionLevelInfinite
}, func() []RecursionLevelType {
	return []RecursionLevelType{RecursionLevelNone, RecursionLevelDirect, RecursionLevelFull, RecursionLevelInfinite}
}, false)

// String returns string representation
func (rl RecursionLevelType) String() string {
	return recursionLevelTypeHelper.String(rl)
}

// IsValid checks if recursion level is valid
func (rl RecursionLevelType) IsValid() bool {
	return recursionLevelTypeHelper.IsValid(rl)
}

// Values returns all possible recursion level values
func (rl RecursionLevelType) Values() []RecursionLevelType {
	return recursionLevelTypeHelper.Values()
}

// MarshalJSON converts recursion level to JSON string
func (rl RecursionLevelType) MarshalJSON() ([]byte, error) {
	return recursionLevelTypeHelper.MarshalJSONImpl(rl)
}

// UnmarshalJSON converts JSON string to recursion level
func (rl *RecursionLevelType) UnmarshalJSON(data []byte) error {
	val, err := recursionLevelTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*rl = val
	return nil
}

// OptimizationLevelType represents optimization levels for operations
// Replaces boolean Optimize to eliminate invalid states
type OptimizationLevelType int

const (
	OptimizationLevelNone OptimizationLevelType = iota
	OptimizationLevelConservative
	OptimizationLevelAggressive
)

// optimizationLevelTypeHelper provides shared functionality for OptimizationLevelType
var optimizationLevelTypeHelper = NewEnumHelper(map[OptimizationLevelType]string{
	OptimizationLevelNone:         "none",
	OptimizationLevelConservative: "conservative",
	OptimizationLevelAggressive:   "aggressive",
}, func(ol OptimizationLevelType) bool {
	return ol >= OptimizationLevelNone && ol <= OptimizationLevelAggressive
}, func() []OptimizationLevelType {
	return []OptimizationLevelType{OptimizationLevelNone, OptimizationLevelConservative, OptimizationLevelAggressive}
}, false)

// String returns string representation
func (ol OptimizationLevelType) String() string {
	return optimizationLevelTypeHelper.String(ol)
}

// IsValid checks if optimization level is valid
func (ol OptimizationLevelType) IsValid() bool {
	return optimizationLevelTypeHelper.IsValid(ol)
}

// Values returns all possible optimization level values
func (ol OptimizationLevelType) Values() []OptimizationLevelType {
	return optimizationLevelTypeHelper.Values()
}

// MarshalJSON converts optimization level to JSON string
func (ol OptimizationLevelType) MarshalJSON() ([]byte, error) {
	return optimizationLevelTypeHelper.MarshalJSONImpl(ol)
}

// UnmarshalJSON converts JSON string to optimization level
func (ol *OptimizationLevelType) UnmarshalJSON(data []byte) error {
	val, err := optimizationLevelTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*ol = val
	return nil
}

// FileSelectionStrategyType represents file selection strategies for cleanup
// Replaces boolean UnusedOnly to eliminate invalid states
type FileSelectionStrategyType int

const (
	FileSelectionStrategyAll FileSelectionStrategyType = iota
	FileSelectionStrategyUnusedOnly
	FileSelectionStrategyManual
)

// fileSelectionStrategyTypeHelper provides shared functionality for FileSelectionStrategyType
var fileSelectionStrategyTypeHelper = NewEnumHelper(map[FileSelectionStrategyType]string{
	FileSelectionStrategyAll:        "all",
	FileSelectionStrategyUnusedOnly: "unused_only",
	FileSelectionStrategyManual:     "manual",
}, func(fss FileSelectionStrategyType) bool {
	return fss >= FileSelectionStrategyAll && fss <= FileSelectionStrategyManual
}, func() []FileSelectionStrategyType {
	return []FileSelectionStrategyType{FileSelectionStrategyAll, FileSelectionStrategyUnusedOnly, FileSelectionStrategyManual}
}, false)

// String returns string representation
func (fss FileSelectionStrategyType) String() string {
	return fileSelectionStrategyTypeHelper.String(fss)
}

// IsValid checks if file selection strategy is valid
func (fss FileSelectionStrategyType) IsValid() bool {
	return fileSelectionStrategyTypeHelper.IsValid(fss)
}

// Values returns all possible file selection strategy values
func (fss FileSelectionStrategyType) Values() []FileSelectionStrategyType {
	return fileSelectionStrategyTypeHelper.Values()
}

// MarshalJSON converts file selection strategy to JSON string
func (fss FileSelectionStrategyType) MarshalJSON() ([]byte, error) {
	return fileSelectionStrategyTypeHelper.MarshalJSONImpl(fss)
}

// UnmarshalJSON converts JSON string to file selection strategy
func (fss *FileSelectionStrategyType) UnmarshalJSON(data []byte) error {
	val, err := fileSelectionStrategyTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*fss = val
	return nil
}

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
	return scanTypeHelper.MarshalJSONImpl(st)
}

// UnmarshalJSON implements json.Unmarshaler
func (st *ScanTypeType) UnmarshalJSON(data []byte) error {
	val, err := scanTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*st = val
	return nil
}

// SafetyLevelType represents safety enforcement levels for configuration
// Replaces boolean SafeMode to eliminate invalid states
type SafetyLevelType int

const (
	SafetyLevelDisabled SafetyLevelType = iota
	SafetyLevelEnabled
	SafetyLevelStrict
	SafetyLevelParanoid
)

// safetyLevelTypeHelper provides shared functionality for SafetyLevelType
var safetyLevelTypeHelper = NewEnumHelper(map[SafetyLevelType]string{
	SafetyLevelDisabled: "disabled",
	SafetyLevelEnabled:  "enabled",
	SafetyLevelStrict:   "strict",
	SafetyLevelParanoid: "paranoid",
}, func(sl SafetyLevelType) bool {
	return sl >= SafetyLevelDisabled && sl <= SafetyLevelParanoid
}, func() []SafetyLevelType {
	return []SafetyLevelType{SafetyLevelDisabled, SafetyLevelEnabled, SafetyLevelStrict, SafetyLevelParanoid}
}, false)

// String returns string representation
func (sl SafetyLevelType) String() string {
	return safetyLevelTypeHelper.String(sl)
}

// IsValid checks if safety level is valid
func (sl SafetyLevelType) IsValid() bool {
	return safetyLevelTypeHelper.IsValid(sl)
}

// Values returns all possible safety level values
func (sl SafetyLevelType) Values() []SafetyLevelType {
	return safetyLevelTypeHelper.Values()
}

// MarshalJSON converts safety level to JSON string
func (sl SafetyLevelType) MarshalJSON() ([]byte, error) {
	return safetyLevelTypeHelper.MarshalJSONImpl(sl)
}

// UnmarshalJSON converts JSON string to safety level
func (sl *SafetyLevelType) UnmarshalJSON(data []byte) error {
	val, err := safetyLevelTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*sl = val
	return nil
}

// ExecutionModeType represents operation execution modes with compile-time safety
// Replaces boolean DryRun to eliminate invalid states
type ExecutionModeType int

const (
	ExecutionModeDryRun ExecutionModeType = iota
	ExecutionModeSimulate
	ExecutionModeExecute
	ExecutionModeForce
)

// executionModeTypeHelper provides shared functionality for ExecutionModeType
var executionModeTypeHelper = NewEnumHelper(map[ExecutionModeType]string{
	ExecutionModeDryRun:   "dry-run",
	ExecutionModeSimulate: "simulate",
	ExecutionModeExecute:  "execute",
	ExecutionModeForce:    "force",
}, func(em ExecutionModeType) bool {
	return em >= ExecutionModeDryRun && em <= ExecutionModeForce
}, func() []ExecutionModeType {
	return []ExecutionModeType{ExecutionModeDryRun, ExecutionModeSimulate, ExecutionModeExecute, ExecutionModeForce}
}, false)

// String returns string representation
func (em ExecutionModeType) String() string {
	return executionModeTypeHelper.String(em)
}

// IsValid checks if execution mode is valid
func (em ExecutionModeType) IsValid() bool {
	return executionModeTypeHelper.IsValid(em)
}

// Values returns all possible execution mode values
func (em ExecutionModeType) Values() []ExecutionModeType {
	return executionModeTypeHelper.Values()
}

// MarshalJSON converts execution mode to JSON string
func (em ExecutionModeType) MarshalJSON() ([]byte, error) {
	return executionModeTypeHelper.MarshalJSONImpl(em)
}

// UnmarshalJSON converts JSON string to execution mode
func (em *ExecutionModeType) UnmarshalJSON(data []byte) error {
	val, err := executionModeTypeHelper.UnmarshalJSONImpl(data)
	if err != nil {
		return err
	}
	*em = val
	return nil
}

// Convenience constants for backward compatibility are now in types.go
