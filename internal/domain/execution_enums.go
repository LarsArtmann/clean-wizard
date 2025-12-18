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