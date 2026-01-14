package domain

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)
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
func (em ExecutionMode) MarshalYAML() (interface{}, error) {
	return int(em), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for ExecutionMode.
// Accepts both string and integer representations.
func (em *ExecutionMode) UnmarshalYAML(value *yaml.Node) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToUpper(s) {
		case "DRY_RUN":
			*em = ExecutionModeDryRun
		case "NORMAL":
			*em = ExecutionModeNormal
		case "FORCE":
			*em = ExecutionModeForce
		default:
			return fmt.Errorf("invalid execution mode: %s", s)
		}
		return nil
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		switch i {
		case 0:
			*em = ExecutionModeDryRun
		case 1:
			*em = ExecutionModeNormal
		case 2:
			*em = ExecutionModeForce
		default:
			return fmt.Errorf("invalid execution mode value: %d (must be 0, 1, or 2)", i)
		}
		return nil
	}

	return fmt.Errorf("cannot parse execution mode: expected string or int")
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
func (sm SafeMode) MarshalYAML() (interface{}, error) {
	return int(sm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for SafeMode.
// Accepts both string and integer representations.
func (sm *SafeMode) UnmarshalYAML(value *yaml.Node) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToUpper(s) {
		case "DISABLED", "0", "FALSE":
			*sm = SafeModeDisabled
		case "ENABLED", "1", "TRUE":
			*sm = SafeModeEnabled
		case "STRICT", "2":
			*sm = SafeModeStrict
		default:
			return fmt.Errorf("invalid safe mode: %s", s)
		}
		return nil
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		switch i {
		case 0:
			*sm = SafeModeDisabled
		case 1:
			*sm = SafeModeEnabled
		case 2:
			*sm = SafeModeStrict
		default:
			return fmt.Errorf("invalid safe mode value: %d (must be 0, 1, or 2)", i)
		}
		return nil
	}

	return fmt.Errorf("cannot parse safe mode: expected string or int")
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
func (ps ProfileStatus) MarshalYAML() (interface{}, error) {
	return int(ps), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for ProfileStatus.
// Accepts both string and integer representations.
func (ps *ProfileStatus) UnmarshalYAML(value *yaml.Node) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToUpper(s) {
		case "DISABLED", "0", "FALSE":
			*ps = ProfileStatusDisabled
		case "ENABLED", "1", "TRUE":
			*ps = ProfileStatusEnabled
		default:
			return fmt.Errorf("invalid profile status: %s", s)
		}
		return nil
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		switch i {
		case 0:
			*ps = ProfileStatusDisabled
		case 1:
			*ps = ProfileStatusEnabled
		default:
			return fmt.Errorf("invalid profile status value: %d (must be 0 or 1)", i)
		}
		return nil
	}

	return fmt.Errorf("cannot parse profile status: expected string or int")
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
func (om OptimizationMode) MarshalYAML() (interface{}, error) {
	return int(om), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for OptimizationMode.
// Accepts both string and integer representations.
func (om *OptimizationMode) UnmarshalYAML(value *yaml.Node) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToUpper(s) {
		case "DISABLED", "0", "FALSE":
			*om = OptimizationModeDisabled
		case "ENABLED", "1", "TRUE":
			*om = OptimizationModeEnabled
		default:
			return fmt.Errorf("invalid optimization mode: %s", s)
		}
		return nil
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		switch i {
		case 0:
			*om = OptimizationModeDisabled
		case 1:
			*om = OptimizationModeEnabled
		default:
			return fmt.Errorf("invalid optimization mode value: %d (must be 0 or 1)", i)
		}
		return nil
	}

	return fmt.Errorf("cannot parse optimization mode: expected string or int")
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
func (hm HomebrewMode) MarshalYAML() (interface{}, error) {
	return int(hm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for HomebrewMode.
// Accepts both string and integer representations.
func (hm *HomebrewMode) UnmarshalYAML(value *yaml.Node) error {
	// Try as string first
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToUpper(s) {
		case "ALL":
			*hm = HomebrewModeAll
		case "UNUSED_ONLY":
			*hm = HomebrewModeUnusedOnly
		default:
			return fmt.Errorf("invalid homebrew mode: %s", s)
		}
		return nil
	}

	// Try as integer
	var i int
	if err := value.Decode(&i); err == nil {
		switch i {
		case 0:
			*hm = HomebrewModeAll
		case 1:
			*hm = HomebrewModeUnusedOnly
		default:
			return fmt.Errorf("invalid homebrew mode value: %d (must be 0 or 1)", i)
		}
		return nil
	}

	return fmt.Errorf("cannot parse homebrew mode: expected string or int")
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
