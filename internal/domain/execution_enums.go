package domain

import "gopkg.in/yaml.v3"

type ExecutionMode int

const (
	ExecutionModeDryRun ExecutionMode = iota
	ExecutionModeNormal
	ExecutionModeForce
)

var executionModeStrings = []string{"DRY_RUN", "NORMAL", "FORCE"}

func (em ExecutionMode) String() string          { return EnumString(em, executionModeStrings) }
func (em ExecutionMode) IsValid() bool           { return EnumIsValid(em, ExecutionModeForce) }
func (em ExecutionMode) Values() []ExecutionMode { return EnumValues(ExecutionModeForce) }
func (em ExecutionMode) IsDryRun() bool          { return em == ExecutionModeDryRun }
func (em ExecutionMode) IsForce() bool           { return em == ExecutionModeForce }

func (em ExecutionMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(em, executionModeStrings)
}

func (em *ExecutionMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(em), executionModeStrings, "execution mode")
}

type SafeMode int

const (
	SafeModeDisabled SafeMode = iota
	SafeModeEnabled
	SafeModeStrict
)

var safeModeStrings = []string{"DISABLED", "ENABLED", "STRICT"}

func (sm SafeMode) String() string     { return EnumString(sm, safeModeStrings) }
func (sm SafeMode) IsValid() bool      { return EnumIsValid(sm, SafeModeStrict) }
func (sm SafeMode) Values() []SafeMode { return EnumValues(SafeModeStrict) }
func (sm SafeMode) IsEnabled() bool    { return sm == SafeModeEnabled || sm == SafeModeStrict }
func (sm SafeMode) IsStrict() bool     { return sm == SafeModeStrict }

func (sm SafeMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(sm, safeModeStrings)
}

func (sm *SafeMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(sm), safeModeStrings, "safe mode")
}

type ProfileStatus int

const (
	ProfileStatusDisabled ProfileStatus = iota
	ProfileStatusEnabled
)

var profileStatusStrings = []string{"DISABLED", "ENABLED"}

func (ps ProfileStatus) String() string          { return EnumString(ps, profileStatusStrings) }
func (ps ProfileStatus) IsValid() bool           { return EnumIsValid(ps, ProfileStatusEnabled) }
func (ps ProfileStatus) Values() []ProfileStatus { return EnumValues(ProfileStatusEnabled) }
func (ps ProfileStatus) IsEnabled() bool         { return ps == ProfileStatusEnabled }

func (ps ProfileStatus) MarshalYAML() (any, error) {
	return EnumMarshalYAML(ps, profileStatusStrings)
}

func (ps *ProfileStatus) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(ps), profileStatusStrings, "profile status")
}

type OptimizationMode int

const (
	OptimizationModeDisabled OptimizationMode = iota
	OptimizationModeEnabled
)

var optimizationModeStrings = []string{"DISABLED", "ENABLED"}

func (om OptimizationMode) String() string { return EnumString(om, optimizationModeStrings) }

func (om OptimizationMode) IsValid() bool { return EnumIsValid(om, OptimizationModeEnabled) }

func (om OptimizationMode) Values() []OptimizationMode { return EnumValues(OptimizationModeEnabled) }
func (om OptimizationMode) IsEnabled() bool            { return om == OptimizationModeEnabled }

func (om OptimizationMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(om, optimizationModeStrings)
}

func (om *OptimizationMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(om), optimizationModeStrings, "optimization mode")
}

type HomebrewMode int

const (
	HomebrewModeAll HomebrewMode = iota
	HomebrewModeUnusedOnly
)

var homebrewModeStrings = []string{"ALL", "UNUSED_ONLY"}

func (hm HomebrewMode) String() string         { return EnumString(hm, homebrewModeStrings) }
func (hm HomebrewMode) IsValid() bool          { return EnumIsValid(hm, HomebrewModeUnusedOnly) }
func (hm HomebrewMode) Values() []HomebrewMode { return EnumValues(HomebrewModeUnusedOnly) }
func (hm HomebrewMode) IsUnusedOnly() bool     { return hm == HomebrewModeUnusedOnly }

func (hm HomebrewMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(hm, homebrewModeStrings)
}

func (hm *HomebrewMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(hm), homebrewModeStrings, "homebrew mode")
}

type GenerationStatus int

const (
	GenerationStatusHistorical GenerationStatus = iota
	GenerationStatusCurrent
)

var generationStatusStrings = []string{"HISTORICAL", "CURRENT"}

func (gs GenerationStatus) String() string { return EnumString(gs, generationStatusStrings) }

func (gs GenerationStatus) IsValid() bool { return EnumIsValid(gs, GenerationStatusCurrent) }

func (gs GenerationStatus) Values() []GenerationStatus { return EnumValues(GenerationStatusCurrent) }
func (gs GenerationStatus) IsCurrent() bool            { return gs == GenerationStatusCurrent }

type ScanMode int

const (
	ScanModeNonRecursive ScanMode = iota
	ScanModeRecursive
)

var scanModeStrings = []string{"NON_RECURSIVE", "RECURSIVE"}

func (sm ScanMode) String() string     { return EnumString(sm, scanModeStrings) }
func (sm ScanMode) IsValid() bool      { return EnumIsValid(sm, ScanModeRecursive) }
func (sm ScanMode) Values() []ScanMode { return EnumValues(ScanModeRecursive) }
func (sm ScanMode) IsRecursive() bool  { return sm == ScanModeRecursive }
