package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// RiskLevelType represents the risk level enum with compile-time safety.
type RiskLevelType int

const (
	RiskLevelLowType RiskLevelType = iota
	RiskLevelMediumType
	RiskLevelHighType
	RiskLevelCriticalType
)

var riskLevelTypeStrings = []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}

func (rl RiskLevelType) String() string { return EnumString(rl, riskLevelTypeStrings) }
func (rl RiskLevelType) IsValid() bool  { return EnumIsValid(rl, RiskLevelCriticalType) }
func (rl RiskLevelType) Values() []RiskLevelType {
	return EnumValues[RiskLevelType](RiskLevelCriticalType)
}
func (rl RiskLevelType) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(rl, riskLevelTypeStrings)
}
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(data, (*int)(rl), riskLevelTypeStrings, "risk level")
}

func (rl RiskLevelType) MarshalYAML() (any, error) { return EnumMarshalYAML(rl, riskLevelTypeStrings) }
func (rl *RiskLevelType) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(rl), riskLevelTypeStrings, "risk level")
}

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

func (rl RiskLevelType) IsHigherThan(other RiskLevelType) bool        { return rl > other }
func (rl RiskLevelType) IsHigherOrEqualThan(other RiskLevelType) bool { return rl >= other }

// ValidationLevelType represents validation levels with compile-time safety.
type ValidationLevelType int

const (
	ValidationLevelNoneType ValidationLevelType = iota
	ValidationLevelBasicType
	ValidationLevelComprehensiveType
	ValidationLevelStrictType
)

var validationLevelTypeStrings = []string{"NONE", "BASIC", "COMPREHENSIVE", "STRICT"}

func (vl ValidationLevelType) String() string { return EnumString(vl, validationLevelTypeStrings) }
func (vl ValidationLevelType) IsValid() bool  { return EnumIsValid(vl, ValidationLevelStrictType) }
func (vl ValidationLevelType) Values() []ValidationLevelType {
	return EnumValues[ValidationLevelType](ValidationLevelStrictType)
}
func (vl ValidationLevelType) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(vl, validationLevelTypeStrings)
}
func (vl *ValidationLevelType) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(data, (*int)(vl), validationLevelTypeStrings, "validation level")
}

// ChangeOperationType represents change operations with compile-time safety.
type ChangeOperationType int

const (
	ChangeOperationAddedType ChangeOperationType = iota
	ChangeOperationRemovedType
	ChangeOperationModifiedType
)

var changeOperationTypeStrings = []string{"ADDED", "REMOVED", "MODIFIED"}

func (co ChangeOperationType) String() string { return EnumString(co, changeOperationTypeStrings) }

func (co ChangeOperationType) IsValid() bool { return EnumIsValid(co, ChangeOperationModifiedType) }
func (co ChangeOperationType) Values() []ChangeOperationType {
	return EnumValues[ChangeOperationType](ChangeOperationModifiedType)
}
func (co ChangeOperationType) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(co, changeOperationTypeStrings)
}
func (co *ChangeOperationType) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(data, (*int)(co), changeOperationTypeStrings, "change operation")
}

// CleanStrategyType represents cleaning strategies with compile-time safety.
type CleanStrategyType int

const (
	StrategyAggressiveType CleanStrategyType = iota
	StrategyConservativeType
	StrategyDryRunType
)

var cleanStrategyTypeStrings = []string{"aggressive", "conservative", "dry-run"}

func (cs CleanStrategyType) String() string { return EnumString(cs, cleanStrategyTypeStrings) }
func (cs CleanStrategyType) IsValid() bool  { return EnumIsValid(cs, StrategyDryRunType) }
func (cs CleanStrategyType) Values() []CleanStrategyType {
	return EnumValues[CleanStrategyType](StrategyDryRunType)
}
func (cs CleanStrategyType) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(cs, cleanStrategyTypeStrings)
}
func (cs *CleanStrategyType) UnmarshalJSON(data []byte) error {
	err := EnumUnmarshalJSON(data, (*int)(cs), cleanStrategyTypeStrings, "clean strategy")
	if err != nil {
		var s string
		jsonErr := json.Unmarshal(data, &s)
		if jsonErr == nil && strings.EqualFold(s, "dryrun") {
			*cs = StrategyDryRunType

			return nil
		}
	}

	return err
}

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

// SizeEstimateStatusType represents the status of a size estimate with type safety.
// This replaces the boolean Unknown field, making impossible states unrepresentable.
type SizeEstimateStatusType int

const (
	// SizeEstimateStatusKnown represents a known/calculated size.
	SizeEstimateStatusKnown SizeEstimateStatusType = iota
	// SizeEstimateStatusUnknown represents an unknown/estimated size.
	SizeEstimateStatusUnknown
)

var sizeEstimateStatusTypeStrings = []string{"KNOWN", "UNKNOWN"}

func (ses SizeEstimateStatusType) String() string {
	if !ses.IsValid() {
		return "INVALID"
	}

	return sizeEstimateStatusTypeStrings[ses]
}

func (ses SizeEstimateStatusType) IsValid() bool { return EnumIsValid(ses, SizeEstimateStatusUnknown) }
func (ses SizeEstimateStatusType) Values() []SizeEstimateStatusType {
	return EnumValues[SizeEstimateStatusType](SizeEstimateStatusUnknown)
}

func (ses SizeEstimateStatusType) MarshalJSON() ([]byte, error) {
	if !ses.IsValid() {
		return nil, fmt.Errorf("invalid size estimate status: %d", ses)
	}

	return EnumMarshalJSON(ses, sizeEstimateStatusTypeStrings)
}

func (ses *SizeEstimateStatusType) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(
		data,
		(*int)(ses),
		sizeEstimateStatusTypeStrings,
		"size estimate status",
	)
}
