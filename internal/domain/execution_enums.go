package domain

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// unmarshalBinaryEnum is a generic helper for unmarshaling binary enums (0/1 values).
// Takes mapping functions for string and integer decoding.
func unmarshalBinaryEnum[T ~int](
	value *yaml.Node,
	result *T,
	stringMap func(string) (T, bool),
	intMap func(int) (T, bool),
	name string,
) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		if v, ok := stringMap(strings.ToUpper(s)); ok {
			*result = v
			return nil
		}
		return fmt.Errorf("invalid %s: %s", name, s)
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		if v, ok := intMap(i); ok {
			*result = v
			return nil
		}
		return fmt.Errorf("invalid %s value: %d", name, i)
	}

	return fmt.Errorf("cannot parse %s: expected string or int", name)
}

// binaryEnumStringMap creates a string mapping function for binary enums.
func binaryEnumStringMap[T ~int](disabled, enabled T) func(string) (T, bool) {
	return func(s string) (T, bool) {
		switch s {
		case "DISABLED", "0", "FALSE":
			return disabled, true
		case "ENABLED", "1", "TRUE":
			return enabled, true
		default:
			return 0, false
		}
	}
}

// binaryEnumIntMap creates an integer mapping function for binary enums.
func binaryEnumIntMap[T ~int](disabled, enabled T) func(int) (T, bool) {
	return func(i int) (T, bool) {
		switch i {
		case 0:
			return disabled, true
		case 1:
			return enabled, true
		default:
			return 0, false
		}
	}
}

// ExecutionMode represents execution behavior as a type-safe enum.
type ExecutionMode int

const (
	// ExecutionModeDryRun represents dry-run execution mode.
	ExecutionModeDryRun ExecutionMode = iota
	// ExecutionModeNormal represents normal execution mode.
	ExecutionModeNormal
	// ExecutionModeForce represents force execution mode.
	ExecutionModeForce
)

// String returns string representation of execution mode.
func (em ExecutionMode) String() string {
	switch em {
	case ExecutionModeDryRun:
		return "DRY_RUN"
	case ExecutionModeNormal:
		return "NORMAL"
	case ExecutionModeForce:
		return "FORCE"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if execution mode is valid.
func (em ExecutionMode) IsValid() bool {
	return em >= ExecutionModeDryRun && em <= ExecutionModeForce
}

// Values returns all possible execution modes.
func (em ExecutionMode) Values() []ExecutionMode {
	return []ExecutionMode{
		ExecutionModeDryRun,
		ExecutionModeNormal,
		ExecutionModeForce,
	}
}

// IsDryRun checks if mode is dry-run.
func (em ExecutionMode) IsDryRun() bool {
	return em == ExecutionModeDryRun
}

// IsForce checks if mode is force.
func (em ExecutionMode) IsForce() bool {
	return em == ExecutionModeForce
}

// MarshalYAML implements yaml.Marshaler interface for ExecutionMode.
func (em ExecutionMode) MarshalYAML() (any, error) {
	return int(em), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for ExecutionMode.
// Accepts both string and integer representations.
func (em *ExecutionMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, em, map[string]ExecutionMode{
		"DRY_RUN": ExecutionModeDryRun,
		"NORMAL":  ExecutionModeNormal,
		"FORCE":   ExecutionModeForce,
	}, "invalid execution mode")
}

// SafeMode represents safety level as a type-safe enum.
type SafeMode int

const (
	// SafeModeDisabled represents disabled safety.
	SafeModeDisabled SafeMode = iota
	// SafeModeEnabled represents enabled safety.
	SafeModeEnabled
	// SafeModeStrict represents strict safety mode.
	SafeModeStrict
)

// String returns string representation of safe mode.
func (sm SafeMode) String() string {
	switch sm {
	case SafeModeDisabled:
		return "DISABLED"
	case SafeModeEnabled:
		return "ENABLED"
	case SafeModeStrict:
		return "STRICT"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if safe mode is valid.
func (sm SafeMode) IsValid() bool {
	return sm >= SafeModeDisabled && sm <= SafeModeStrict
}

// Values returns all possible safe modes.
func (sm SafeMode) Values() []SafeMode {
	return []SafeMode{
		SafeModeDisabled,
		SafeModeEnabled,
		SafeModeStrict,
	}
}

// IsEnabled checks if safety is enabled.
func (sm SafeMode) IsEnabled() bool {
	return sm == SafeModeEnabled || sm == SafeModeStrict
}

// IsStrict checks if safety is strict.
func (sm SafeMode) IsStrict() bool {
	return sm == SafeModeStrict
}

// MarshalYAML implements yaml.Marshaler interface for SafeMode.
func (sm SafeMode) MarshalYAML() (any, error) {
	return int(sm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for SafeMode.
// Accepts both string and integer representations.
func (sm *SafeMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, sm, map[string]SafeMode{
		"DISABLED": SafeModeDisabled,
		"ENABLED":  SafeModeEnabled,
		"STRICT":   SafeModeStrict,
	}, "invalid safe mode")
}

// ProfileStatus represents profile enabled state as type-safe enum.
type ProfileStatus int

const (
	// ProfileStatusDisabled represents disabled profile.
	ProfileStatusDisabled ProfileStatus = iota
	// ProfileStatusEnabled represents enabled profile.
	ProfileStatusEnabled
)

// String returns string representation of profile status.
func (ps ProfileStatus) String() string {
	switch ps {
	case ProfileStatusDisabled:
		return "DISABLED"
	case ProfileStatusEnabled:
		return "ENABLED"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if profile status is valid.
func (ps ProfileStatus) IsValid() bool {
	return ps >= ProfileStatusDisabled && ps <= ProfileStatusEnabled
}

// Values returns all possible profile statuses.
func (ps ProfileStatus) Values() []ProfileStatus {
	return []ProfileStatus{
		ProfileStatusDisabled,
		ProfileStatusEnabled,
	}
}

// IsEnabled checks if profile is enabled.
func (ps ProfileStatus) IsEnabled() bool {
	return ps == ProfileStatusEnabled
}

// MarshalYAML implements yaml.Marshaler interface for ProfileStatus.
func (ps ProfileStatus) MarshalYAML() (any, error) {
	return int(ps), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for ProfileStatus.
// Accepts both string and integer representations.
func (ps *ProfileStatus) UnmarshalYAML(value *yaml.Node) error {
	return unmarshalBinaryEnum(
		value, ps,
		binaryEnumStringMap(ProfileStatusDisabled, ProfileStatusEnabled),
		binaryEnumIntMap(ProfileStatusDisabled, ProfileStatusEnabled),
		"profile status",
	)
}

// OptimizationMode represents optimization preference as type-safe enum.
type OptimizationMode int

const (
	// OptimizationModeDisabled represents disabled optimization.
	OptimizationModeDisabled OptimizationMode = iota
	// OptimizationModeEnabled represents enabled optimization.
	OptimizationModeEnabled
)

// String returns string representation of optimization mode.
func (om OptimizationMode) String() string {
	switch om {
	case OptimizationModeDisabled:
		return "DISABLED"
	case OptimizationModeEnabled:
		return "ENABLED"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if optimization mode is valid.
func (om OptimizationMode) IsValid() bool {
	return om >= OptimizationModeDisabled && om <= OptimizationModeEnabled
}

// Values returns all possible optimization modes.
func (om OptimizationMode) Values() []OptimizationMode {
	return []OptimizationMode{
		OptimizationModeDisabled,
		OptimizationModeEnabled,
	}
}

// IsEnabled checks if optimization is enabled.
func (om OptimizationMode) IsEnabled() bool {
	return om == OptimizationModeEnabled
}

// MarshalYAML implements yaml.Marshaler interface for OptimizationMode.
func (om OptimizationMode) MarshalYAML() (any, error) {
	return int(om), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for OptimizationMode.
// Accepts both string and integer representations.
func (om *OptimizationMode) UnmarshalYAML(value *yaml.Node) error {
	return unmarshalBinaryEnum(
		value, om,
		binaryEnumStringMap(OptimizationModeDisabled, OptimizationModeEnabled),
		binaryEnumIntMap(OptimizationModeDisabled, OptimizationModeEnabled),
		"optimization mode",
	)
}

// HomebrewMode represents homebrew operation mode as type-safe enum.
type HomebrewMode int

const (
	// HomebrewModeAll represents cleaning all packages.
	HomebrewModeAll HomebrewMode = iota
	// HomebrewModeUnusedOnly represents cleaning only unused packages.
	HomebrewModeUnusedOnly
)

// String returns string representation of homebrew mode.
func (hm HomebrewMode) String() string {
	switch hm {
	case HomebrewModeAll:
		return "ALL"
	case HomebrewModeUnusedOnly:
		return "UNUSED_ONLY"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if homebrew mode is valid.
func (hm HomebrewMode) IsValid() bool {
	return hm >= HomebrewModeAll && hm <= HomebrewModeUnusedOnly
}

// Values returns all possible homebrew modes.
func (hm HomebrewMode) Values() []HomebrewMode {
	return []HomebrewMode{
		HomebrewModeAll,
		HomebrewModeUnusedOnly,
	}
}

// IsUnusedOnly checks if mode is unused-only.
func (hm HomebrewMode) IsUnusedOnly() bool {
	return hm == HomebrewModeUnusedOnly
}

// MarshalYAML implements yaml.Marshaler interface for HomebrewMode.
func (hm HomebrewMode) MarshalYAML() (any, error) {
	return int(hm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for HomebrewMode.
// Accepts both string and integer representations.
func (hm *HomebrewMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, hm, map[string]HomebrewMode{
		"ALL":          HomebrewModeAll,
		"UNUSED_ONLY":  HomebrewModeUnusedOnly,
	}, "invalid homebrew mode")
}

// GenerationStatus represents generation status as type-safe enum.
type GenerationStatus int

const (
	// GenerationStatusHistorical represents historical generation.
	GenerationStatusHistorical GenerationStatus = iota
	// GenerationStatusCurrent represents current generation.
	GenerationStatusCurrent
)

// String returns string representation of generation status.
func (gs GenerationStatus) String() string {
	switch gs {
	case GenerationStatusHistorical:
		return "HISTORICAL"
	case GenerationStatusCurrent:
		return "CURRENT"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if generation status is valid.
func (gs GenerationStatus) IsValid() bool {
	return gs >= GenerationStatusHistorical && gs <= GenerationStatusCurrent
}

// Values returns all possible generation statuses.
func (gs GenerationStatus) Values() []GenerationStatus {
	return []GenerationStatus{
		GenerationStatusHistorical,
		GenerationStatusCurrent,
	}
}

// IsCurrent checks if generation is current.
func (gs GenerationStatus) IsCurrent() bool {
	return gs == GenerationStatusCurrent
}

// ScanMode represents scanning behavior as type-safe enum.
type ScanMode int

const (
	// ScanModeNonRecursive represents non-recursive scanning.
	ScanModeNonRecursive ScanMode = iota
	// ScanModeRecursive represents recursive scanning.
	ScanModeRecursive
)

// String returns string representation of scan mode.
func (sm ScanMode) String() string {
	switch sm {
	case ScanModeNonRecursive:
		return "NON_RECURSIVE"
	case ScanModeRecursive:
		return "RECURSIVE"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if scan mode is valid.
func (sm ScanMode) IsValid() bool {
	return sm >= ScanModeNonRecursive && sm <= ScanModeRecursive
}

// Values returns all possible scan modes.
func (sm ScanMode) Values() []ScanMode {
	return []ScanMode{
		ScanModeNonRecursive,
		ScanModeRecursive,
	}
}

// IsRecursive checks if mode is recursive.
func (sm ScanMode) IsRecursive() bool {
	return sm == ScanModeRecursive
}
