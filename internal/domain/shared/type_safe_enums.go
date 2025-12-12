package shared

//go:generate stringer -type=RiskLevelType
//go:generate stringer -type=StatusType
//go:generate stringer -type=OperationNameType
//go:generate stringer -type=OptimizationLevelType
//go:generate stringer -type=ExecutionModeType
//go:generate stringer -type=CleanStrategyType
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

// Values returns all possible enum values
func (eh *EnumHelper[T]) Values() []T {
	eh.once.Do(func() {
		eh.valuesCache = eh.allValues()
	})
	return eh.valuesCache
}

// Parse parses string to enum value
func (eh *EnumHelper[T]) Parse(s string) (T, error) {
	var zero T
	// Check if string is empty
	if s == "" {
		return zero, fmt.Errorf("cannot parse empty string")
	}

	// Try exact match first
	for value, str := range eh.stringValues {
		if eh.caseSensitive {
			if str == s {
				return value, nil
			}
		} else {
			if strings.EqualFold(str, s) {
				return value, nil
			}
		}
	}

	return zero, fmt.Errorf("invalid enum value: %q", s)
}

// MarshalJSON implements json.Marshaler for enums
func (eh *EnumHelper[T]) MarshalJSON(value T) ([]byte, error) {
	return json.Marshal(eh.String(value))
}

// UnmarshalJSON implements json.Unmarshaler for enums
func (eh *EnumHelper[T]) UnmarshalJSON(data []byte) (T, error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var zero T
		return zero, err
	}
	return eh.Parse(s)
}

// RiskLevelType represents risk level enum with compile-time safety
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

// IsValid checks if risk level is valid
func (rl RiskLevelType) IsValid() bool {
	return riskLevelHelper.IsValid(rl)
}

// ParseRiskLevel parses string to RiskLevelType
func ParseRiskLevel(s string) (RiskLevelType, error) {
	return riskLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (rl RiskLevelType) MarshalJSON() ([]byte, error) {
	return riskLevelHelper.MarshalJSON(rl)
}

// UnmarshalJSON implements json.Unmarshaler
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error {
	value, err := riskLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*rl = value
	}
	return err
}

// operationNameHelper provides shared functionality for OperationNameType
var operationNameHelper = NewEnumHelper(map[OperationNameType]string{
	OperationNameNixGenerations: "NIX_GENERATIONS",
	OperationNameHomebrew:       "HOMEBREW",
	OperationNamePackageCache:   "PACKAGE_CACHE",
	OperationNameTempFiles:      "TEMP_FILES",
	OperationNameSystemTemp:     "SYSTEM_TEMP",
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
}, true)

// IsValid checks if operation name is valid
func (on OperationNameType) IsValid() bool {
	return operationNameHelper.IsValid(on)
}

// ParseOperationName parses string to OperationNameType
func ParseOperationName(s string) (OperationNameType, error) {
	return operationNameHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (on OperationNameType) MarshalJSON() ([]byte, error) {
	return operationNameHelper.MarshalJSON(on)
}

// UnmarshalJSON implements json.Unmarshaler
func (on *OperationNameType) UnmarshalJSON(data []byte) error {
	value, err := operationNameHelper.UnmarshalJSON(data)
	if err == nil {
		*on = value
	}
	return err
}

// ValidationLevelType represents validation level enum with compile-time safety
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
}, true)

// IsValid checks if validation level is valid
func (vl ValidationLevelType) IsValid() bool {
	return validationLevelHelper.IsValid(vl)
}

// ParseValidationLevel parses string to ValidationLevelType
func ParseValidationLevel(s string) (ValidationLevelType, error) {
	return validationLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (vl ValidationLevelType) MarshalJSON() ([]byte, error) {
	return validationLevelHelper.MarshalJSON(vl)
}

// UnmarshalJSON implements json.Unmarshaler
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	value, err := validationLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*vl = value
	}
	return err
}

// ChangeOperationType represents change operation enum with compile-time safety
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
}, true)

// IsValid checks if change operation is valid
func (co ChangeOperationType) IsValid() bool {
	return changeOperationHelper.IsValid(co)
}

// ParseChangeOperation parses string to ChangeOperationType
func ParseChangeOperation(s string) (ChangeOperationType, error) {
	return changeOperationHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (co ChangeOperationType) MarshalJSON() ([]byte, error) {
	return changeOperationHelper.MarshalJSON(co)
}

// UnmarshalJSON implements json.Unmarshaler
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	value, err := changeOperationHelper.UnmarshalJSON(data)
	if err == nil {
		*co = value
	}
	return err
}

// CleanStrategyType represents clean strategy enum with compile-time safety
type CleanStrategyType int

const (
	StrategyAggressiveType CleanStrategyType = iota
	StrategyConservativeType
	StrategyDryRunType
)

// cleanStrategyHelper provides shared functionality for CleanStrategyType
var cleanStrategyHelper = NewEnumHelper(map[CleanStrategyType]string{
	StrategyAggressiveType:   "AGGRESSIVE",
	StrategyConservativeType: "CONSERVATIVE",
	StrategyDryRunType:       "DRY_RUN",
}, func(cs CleanStrategyType) bool {
	return cs >= StrategyAggressiveType && cs <= StrategyDryRunType
}, func() []CleanStrategyType {
	return []CleanStrategyType{
		StrategyAggressiveType,
		StrategyConservativeType,
		StrategyDryRunType,
	}
}, true)

// IsValid checks if clean strategy is valid
func (cs CleanStrategyType) IsValid() bool {
	return cleanStrategyHelper.IsValid(cs)
}

// ParseCleanStrategy parses string to CleanStrategyType
func ParseCleanStrategy(s string) (CleanStrategyType, error) {
	return cleanStrategyHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (cs CleanStrategyType) MarshalJSON() ([]byte, error) {
	return cleanStrategyHelper.MarshalJSON(cs)
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	value, err := cleanStrategyHelper.UnmarshalJSON(data)
	if err == nil {
		*cs = value
	}
	return err
}

// ScanTypeType represents scan type enum with compile-time safety
type ScanTypeType int

const (
	ScanTypeNixStoreType ScanTypeType = iota
	ScanTypeHomebrewType
	ScanTypePackageCacheType
	ScanTypeTempType
	ScanTypeSystemType
)

// scanTypeHelper provides shared functionality for ScanTypeType
var scanTypeHelper = NewEnumHelper(map[ScanTypeType]string{
	ScanTypeNixStoreType:     "NIX_STORE",
	ScanTypeHomebrewType:     "HOMEBREW",
	ScanTypePackageCacheType: "PACKAGE_CACHE",
	ScanTypeTempType:         "TEMP",
	ScanTypeSystemType:       "SYSTEM",
}, func(st ScanTypeType) bool {
	return st >= ScanTypeNixStoreType && st <= ScanTypeSystemType
}, func() []ScanTypeType {
	return []ScanTypeType{
		ScanTypeNixStoreType,
		ScanTypeHomebrewType,
		ScanTypePackageCacheType,
		ScanTypeTempType,
		ScanTypeSystemType,
	}
}, true)

// IsValid checks if scan type is valid
func (st ScanTypeType) IsValid() bool {
	return scanTypeHelper.IsValid(st)
}

// ParseScanType parses string to ScanTypeType
func ParseScanType(s string) (ScanTypeType, error) {
	return scanTypeHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (st ScanTypeType) MarshalJSON() ([]byte, error) {
	return scanTypeHelper.MarshalJSON(st)
}

// UnmarshalJSON implements json.Unmarshaler
func (st *ScanTypeType) UnmarshalJSON(data []byte) error {
	value, err := scanTypeHelper.UnmarshalJSON(data)
	if err == nil {
		*st = value
	}
	return err
}

// StatusType represents status enum with compile-time safety
type StatusType int

const (
	StatusInactiveType StatusType = iota
	StatusActiveType
	StatusCompletedType
	StatusFailedType
	StatusPendingType
)

// statusHelper provides shared functionality for StatusType
var statusHelper = NewEnumHelper(map[StatusType]string{
	StatusInactiveType:  "INACTIVE",
	StatusActiveType:    "ACTIVE",
	StatusCompletedType: "COMPLETED",
	StatusFailedType:    "FAILED",
	StatusPendingType:   "PENDING",
}, func(s StatusType) bool {
	return s >= StatusInactiveType && s <= StatusPendingType
}, func() []StatusType {
	return []StatusType{
		StatusInactiveType,
		StatusActiveType,
		StatusCompletedType,
		StatusFailedType,
		StatusPendingType,
	}
}, true)

// IsValid checks if status is valid
func (s StatusType) IsValid() bool {
	return statusHelper.IsValid(s)
}

// ParseStatus parses string to StatusType
func ParseStatus(s string) (StatusType, error) {
	return statusHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (s StatusType) MarshalJSON() ([]byte, error) {
	return statusHelper.MarshalJSON(s)
}

// UnmarshalJSON implements json.Unmarshaler
func (s *StatusType) UnmarshalJSON(data []byte) error {
	value, err := statusHelper.UnmarshalJSON(data)
	if err == nil {
		*s = value
	}
	return err
}

// EnforcementLevelType represents enforcement level enum with compile-time safety
type EnforcementLevelType int

const (
	EnforcementLevelNoneType EnforcementLevelType = iota
	EnforcementLevelWarningType
	EnforcementLevelStrictType
	EnforcementLevelCriticalType
)

// enforcementLevelHelper provides shared functionality for EnforcementLevelType
var enforcementLevelHelper = NewEnumHelper(map[EnforcementLevelType]string{
	EnforcementLevelNoneType:     "NONE",
	EnforcementLevelWarningType:  "WARNING",
	EnforcementLevelStrictType:   "STRICT",
	EnforcementLevelCriticalType: "CRITICAL",
}, func(el EnforcementLevelType) bool {
	return el >= EnforcementLevelNoneType && el <= EnforcementLevelCriticalType
}, func() []EnforcementLevelType {
	return []EnforcementLevelType{
		EnforcementLevelNoneType,
		EnforcementLevelWarningType,
		EnforcementLevelStrictType,
		EnforcementLevelCriticalType,
	}
}, true)

// IsValid checks if enforcement level is valid
func (el EnforcementLevelType) IsValid() bool {
	return enforcementLevelHelper.IsValid(el)
}

// ParseEnforcementLevel parses string to EnforcementLevelType
func ParseEnforcementLevel(s string) (EnforcementLevelType, error) {
	return enforcementLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (el EnforcementLevelType) MarshalJSON() ([]byte, error) {
	return enforcementLevelHelper.MarshalJSON(el)
}

// UnmarshalJSON implements json.Unmarshaler
func (el *EnforcementLevelType) UnmarshalJSON(data []byte) error {
	value, err := enforcementLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*el = value
	}
	return err
}

// SelectedStatusType represents selected status enum with compile-time safety
type SelectedStatusType int

const (
	SelectedStatusNotSelectedType SelectedStatusType = iota
	SelectedStatusSelectedType
	SelectedStatusPartialType
	SelectedStatusExcludedType
)

// selectedStatusHelper provides shared functionality for SelectedStatusType
var selectedStatusHelper = NewEnumHelper(map[SelectedStatusType]string{
	SelectedStatusNotSelectedType: "NOT_SELECTED",
	SelectedStatusSelectedType:    "SELECTED",
	SelectedStatusPartialType:     "PARTIAL",
	SelectedStatusExcludedType:    "EXCLUDED",
}, func(ss SelectedStatusType) bool {
	return ss >= SelectedStatusNotSelectedType && ss <= SelectedStatusExcludedType
}, func() []SelectedStatusType {
	return []SelectedStatusType{
		SelectedStatusNotSelectedType,
		SelectedStatusSelectedType,
		SelectedStatusPartialType,
		SelectedStatusExcludedType,
	}
}, true)

// IsValid checks if selected status is valid
func (ss SelectedStatusType) IsValid() bool {
	return selectedStatusHelper.IsValid(ss)
}

// ParseSelectedStatus parses string to SelectedStatusType
func ParseSelectedStatus(s string) (SelectedStatusType, error) {
	return selectedStatusHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (ss SelectedStatusType) MarshalJSON() ([]byte, error) {
	return selectedStatusHelper.MarshalJSON(ss)
}

// UnmarshalJSON implements json.Unmarshaler
func (ss *SelectedStatusType) UnmarshalJSON(data []byte) error {
	value, err := selectedStatusHelper.UnmarshalJSON(data)
	if err == nil {
		*ss = value
	}
	return err
}

// RecursionLevelType represents recursion level enum with compile-time safety
type RecursionLevelType int

const (
	RecursionLevelNoneType RecursionLevelType = iota
	RecursionLevelShallowType
	RecursionLevelMediumType
	RecursionLevelDeepType
	RecursionLevelFullType
)

// recursionLevelHelper provides shared functionality for RecursionLevelType
var recursionLevelHelper = NewEnumHelper(map[RecursionLevelType]string{
	RecursionLevelNoneType:    "NONE",
	RecursionLevelShallowType: "SHALLOW",
	RecursionLevelMediumType:  "MEDIUM",
	RecursionLevelDeepType:    "DEEP",
	RecursionLevelFullType:    "FULL",
}, func(rl RecursionLevelType) bool {
	return rl >= RecursionLevelNoneType && rl <= RecursionLevelFullType
}, func() []RecursionLevelType {
	return []RecursionLevelType{
		RecursionLevelNoneType,
		RecursionLevelShallowType,
		RecursionLevelMediumType,
		RecursionLevelDeepType,
		RecursionLevelFullType,
	}
}, true)

// IsValid checks if recursion level is valid
func (rl RecursionLevelType) IsValid() bool {
	return recursionLevelHelper.IsValid(rl)
}

// ParseRecursionLevel parses string to RecursionLevelType
func ParseRecursionLevel(s string) (RecursionLevelType, error) {
	return recursionLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (rl RecursionLevelType) MarshalJSON() ([]byte, error) {
	return recursionLevelHelper.MarshalJSON(rl)
}

// UnmarshalJSON implements json.Unmarshaler
func (rl *RecursionLevelType) UnmarshalJSON(data []byte) error {
	value, err := recursionLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*rl = value
	}
	return err
}

// OptimizationLevelType represents optimization level enum with compile-time safety
type OptimizationLevelType int

const (
	OptimizationLevelNoneType OptimizationLevelType = iota
	OptimizationLevelBasicType
	OptimizationLevelAggressiveType
	OptimizationLevelMaximumType
)

// optimizationLevelHelper provides shared functionality for OptimizationLevelType
var optimizationLevelHelper = NewEnumHelper(map[OptimizationLevelType]string{
	OptimizationLevelNoneType:       "NONE",
	OptimizationLevelBasicType:      "BASIC",
	OptimizationLevelAggressiveType: "AGGRESSIVE",
	OptimizationLevelMaximumType:    "MAXIMUM",
}, func(ol OptimizationLevelType) bool {
	return ol >= OptimizationLevelNoneType && ol <= OptimizationLevelMaximumType
}, func() []OptimizationLevelType {
	return []OptimizationLevelType{
		OptimizationLevelNoneType,
		OptimizationLevelBasicType,
		OptimizationLevelAggressiveType,
		OptimizationLevelMaximumType,
	}
}, true)

// IsValid checks if optimization level is valid
func (ol OptimizationLevelType) IsValid() bool {
	return optimizationLevelHelper.IsValid(ol)
}

// ParseOptimizationLevel parses string to OptimizationLevelType
func ParseOptimizationLevel(s string) (OptimizationLevelType, error) {
	return optimizationLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (ol OptimizationLevelType) MarshalJSON() ([]byte, error) {
	return optimizationLevelHelper.MarshalJSON(ol)
}

// UnmarshalJSON implements json.Unmarshaler
func (ol *OptimizationLevelType) UnmarshalJSON(data []byte) error {
	value, err := optimizationLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*ol = value
	}
	return err
}

// FileSelectionStrategyType represents file selection strategy enum with compile-time safety
type FileSelectionStrategyType int

const (
	FileSelectionStrategyAllType FileSelectionStrategyType = iota
	FileSelectionStrategySizeType
	FileSelectionStrategyTypeType
	FileSelectionStrategyDateType
	FileSelectionStrategyPatternType
)

// fileSelectionStrategyHelper provides shared functionality for FileSelectionStrategyType
var fileSelectionStrategyHelper = NewEnumHelper(map[FileSelectionStrategyType]string{
	FileSelectionStrategyAllType:     "ALL",
	FileSelectionStrategySizeType:    "SIZE",
	FileSelectionStrategyTypeType:    "TYPE",
	FileSelectionStrategyDateType:    "DATE",
	FileSelectionStrategyPatternType: "PATTERN",
}, func(fss FileSelectionStrategyType) bool {
	return fss >= FileSelectionStrategyAllType && fss <= FileSelectionStrategyPatternType
}, func() []FileSelectionStrategyType {
	return []FileSelectionStrategyType{
		FileSelectionStrategyAllType,
		FileSelectionStrategySizeType,
		FileSelectionStrategyTypeType,
		FileSelectionStrategyDateType,
		FileSelectionStrategyPatternType,
	}
}, true)

// IsValid checks if file selection strategy is valid
func (fss FileSelectionStrategyType) IsValid() bool {
	return fileSelectionStrategyHelper.IsValid(fss)
}

// ParseFileSelectionStrategy parses string to FileSelectionStrategyType
func ParseFileSelectionStrategy(s string) (FileSelectionStrategyType, error) {
	return fileSelectionStrategyHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (fss FileSelectionStrategyType) MarshalJSON() ([]byte, error) {
	return fileSelectionStrategyHelper.MarshalJSON(fss)
}

// UnmarshalJSON implements json.Unmarshaler
func (fss *FileSelectionStrategyType) UnmarshalJSON(data []byte) error {
	value, err := fileSelectionStrategyHelper.UnmarshalJSON(data)
	if err == nil {
		*fss = value
	}
	return err
}

// SafetyLevelType represents safety level enum with compile-time safety
type SafetyLevelType int

const (
	SafetyLevelUnsafeType SafetyLevelType = iota
	SafetyLevelSafeType
	SafetyLevelStrictType
	SafetyLevelParanoidType
)

// safetyLevelHelper provides shared functionality for SafetyLevelType
var safetyLevelHelper = NewEnumHelper(map[SafetyLevelType]string{
	SafetyLevelUnsafeType:   "UNSAFE",
	SafetyLevelSafeType:     "SAFE",
	SafetyLevelStrictType:   "STRICT",
	SafetyLevelParanoidType: "PARANOID",
}, func(sl SafetyLevelType) bool {
	return sl >= SafetyLevelUnsafeType && sl <= SafetyLevelParanoidType
}, func() []SafetyLevelType {
	return []SafetyLevelType{
		SafetyLevelUnsafeType,
		SafetyLevelSafeType,
		SafetyLevelStrictType,
		SafetyLevelParanoidType,
	}
}, true)

// IsValid checks if safety level is valid
func (sl SafetyLevelType) IsValid() bool {
	return safetyLevelHelper.IsValid(sl)
}

// ParseSafetyLevel parses string to SafetyLevelType
func ParseSafetyLevel(s string) (SafetyLevelType, error) {
	return safetyLevelHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (sl SafetyLevelType) MarshalJSON() ([]byte, error) {
	return safetyLevelHelper.MarshalJSON(sl)
}

// UnmarshalJSON implements json.Unmarshaler
func (sl *SafetyLevelType) UnmarshalJSON(data []byte) error {
	value, err := safetyLevelHelper.UnmarshalJSON(data)
	if err == nil {
		*sl = value
	}
	return err
}

// ExecutionModeType represents execution mode enum with compile-time safety
type ExecutionModeType int

const (
	ExecutionModeSequentialType ExecutionModeType = iota
	ExecutionModeParallelType
	ExecutionModeBatchType
	ExecutionModeInteractiveType
)

// executionModeHelper provides shared functionality for ExecutionModeType
var executionModeHelper = NewEnumHelper(map[ExecutionModeType]string{
	ExecutionModeSequentialType:  "SEQUENTIAL",
	ExecutionModeParallelType:    "PARALLEL",
	ExecutionModeBatchType:       "BATCH",
	ExecutionModeInteractiveType: "INTERACTIVE",
}, func(em ExecutionModeType) bool {
	return em >= ExecutionModeSequentialType && em <= ExecutionModeInteractiveType
}, func() []ExecutionModeType {
	return []ExecutionModeType{
		ExecutionModeSequentialType,
		ExecutionModeParallelType,
		ExecutionModeBatchType,
		ExecutionModeInteractiveType,
	}
}, true)

// IsValid checks if execution mode is valid
func (em ExecutionModeType) IsValid() bool {
	return executionModeHelper.IsValid(em)
}

// ParseExecutionMode parses string to ExecutionModeType
func ParseExecutionMode(s string) (ExecutionModeType, error) {
	return executionModeHelper.Parse(s)
}

// MarshalJSON implements json.Marshaler
func (em ExecutionModeType) MarshalJSON() ([]byte, error) {
	return executionModeHelper.MarshalJSON(em)
}

// UnmarshalJSON implements json.Unmarshaler
func (em *ExecutionModeType) UnmarshalJSON(data []byte) error {
	value, err := executionModeHelper.UnmarshalJSON(data)
	if err == nil {
		*em = value
	}
	return err
}
