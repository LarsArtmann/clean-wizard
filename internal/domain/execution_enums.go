package domain

// ExecutionMode represents execution behavior as a type-safe enum
type ExecutionMode int

const (
	// ExecutionModeDryRun represents dry-run execution mode
	ExecutionModeDryRun ExecutionMode = iota
	// ExecutionModeNormal represents normal execution mode
	ExecutionModeNormal
	// ExecutionModeForce represents force execution mode
	ExecutionModeForce
)

// String returns string representation of execution mode
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

// IsValid checks if execution mode is valid
func (em ExecutionMode) IsValid() bool {
	return em >= ExecutionModeDryRun && em <= ExecutionModeForce
}

// Values returns all possible execution modes
func (em ExecutionMode) Values() []ExecutionMode {
	return []ExecutionMode{
		ExecutionModeDryRun,
		ExecutionModeNormal,
		ExecutionModeForce,
	}
}

// IsDryRun checks if mode is dry-run
func (em ExecutionMode) IsDryRun() bool {
	return em == ExecutionModeDryRun
}

// IsForce checks if mode is force
func (em ExecutionMode) IsForce() bool {
	return em == ExecutionModeForce
}

// SafeMode represents safety level as a type-safe enum
type SafeMode int

const (
	// SafeModeDisabled represents disabled safety
	SafeModeDisabled SafeMode = iota
	// SafeModeEnabled represents enabled safety
	SafeModeEnabled
	// SafeModeStrict represents strict safety mode
	SafeModeStrict
)

// String returns string representation of safe mode
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

// IsValid checks if safe mode is valid
func (sm SafeMode) IsValid() bool {
	return sm >= SafeModeDisabled && sm <= SafeModeStrict
}

// Values returns all possible safe modes
func (sm SafeMode) Values() []SafeMode {
	return []SafeMode{
		SafeModeDisabled,
		SafeModeEnabled,
		SafeModeStrict,
	}
}

// IsEnabled checks if safety is enabled
func (sm SafeMode) IsEnabled() bool {
	return sm == SafeModeEnabled || sm == SafeModeStrict
}

// IsStrict checks if safety is strict
func (sm SafeMode) IsStrict() bool {
	return sm == SafeModeStrict
}

// ProfileStatus represents profile enabled state as type-safe enum
type ProfileStatus int

const (
	// ProfileStatusDisabled represents disabled profile
	ProfileStatusDisabled ProfileStatus = iota
	// ProfileStatusEnabled represents enabled profile
	ProfileStatusEnabled
)

// String returns string representation of profile status
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

// IsValid checks if profile status is valid
func (ps ProfileStatus) IsValid() bool {
	return ps >= ProfileStatusDisabled && ps <= ProfileStatusEnabled
}

// Values returns all possible profile statuses
func (ps ProfileStatus) Values() []ProfileStatus {
	return []ProfileStatus{
		ProfileStatusDisabled,
		ProfileStatusEnabled,
	}
}

// IsEnabled checks if profile is enabled
func (ps ProfileStatus) IsEnabled() bool {
	return ps == ProfileStatusEnabled
}

// OptimizationMode represents optimization preference as type-safe enum
type OptimizationMode int

const (
	// OptimizationModeDisabled represents disabled optimization
	OptimizationModeDisabled OptimizationMode = iota
	// OptimizationModeEnabled represents enabled optimization
	OptimizationModeEnabled
)

// String returns string representation of optimization mode
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

// IsValid checks if optimization mode is valid
func (om OptimizationMode) IsValid() bool {
	return om >= OptimizationModeDisabled && om <= OptimizationModeEnabled
}

// Values returns all possible optimization modes
func (om OptimizationMode) Values() []OptimizationMode {
	return []OptimizationMode{
		OptimizationModeDisabled,
		OptimizationModeEnabled,
	}
}

// IsEnabled checks if optimization is enabled
func (om OptimizationMode) IsEnabled() bool {
	return om == OptimizationModeEnabled
}

// HomebrewMode represents homebrew operation mode as type-safe enum
type HomebrewMode int

const (
	// HomebrewModeAll represents cleaning all packages
	HomebrewModeAll HomebrewMode = iota
	// HomebrewModeUnusedOnly represents cleaning only unused packages
	HomebrewModeUnusedOnly
)

// String returns string representation of homebrew mode
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

// IsValid checks if homebrew mode is valid
func (hm HomebrewMode) IsValid() bool {
	return hm >= HomebrewModeAll && hm <= HomebrewModeUnusedOnly
}

// Values returns all possible homebrew modes
func (hm HomebrewMode) Values() []HomebrewMode {
	return []HomebrewMode{
		HomebrewModeAll,
		HomebrewModeUnusedOnly,
	}
}

// IsUnusedOnly checks if mode is unused-only
func (hm HomebrewMode) IsUnusedOnly() bool {
	return hm == HomebrewModeUnusedOnly
}

// GenerationStatus represents generation status as type-safe enum
type GenerationStatus int

const (
	// GenerationStatusHistorical represents historical generation
	GenerationStatusHistorical GenerationStatus = iota
	// GenerationStatusCurrent represents current generation
	GenerationStatusCurrent
)

// String returns string representation of generation status
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

// IsValid checks if generation status is valid
func (gs GenerationStatus) IsValid() bool {
	return gs >= GenerationStatusHistorical && gs <= GenerationStatusCurrent
}

// Values returns all possible generation statuses
func (gs GenerationStatus) Values() []GenerationStatus {
	return []GenerationStatus{
		GenerationStatusHistorical,
		GenerationStatusCurrent,
	}
}

// IsCurrent checks if generation is current
func (gs GenerationStatus) IsCurrent() bool {
	return gs == GenerationStatusCurrent
}

// ScanMode represents scanning behavior as type-safe enum
type ScanMode int

const (
	// ScanModeNonRecursive represents non-recursive scanning
	ScanModeNonRecursive ScanMode = iota
	// ScanModeRecursive represents recursive scanning
	ScanModeRecursive
)

// String returns string representation of scan mode
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

// IsValid checks if scan mode is valid
func (sm ScanMode) IsValid() bool {
	return sm >= ScanModeNonRecursive && sm <= ScanModeRecursive
}

// Values returns all possible scan modes
func (sm ScanMode) Values() []ScanMode {
	return []ScanMode{
		ScanModeNonRecursive,
		ScanModeRecursive,
	}
}

// IsRecursive checks if mode is recursive
func (sm ScanMode) IsRecursive() bool {
	return sm == ScanModeRecursive
}
