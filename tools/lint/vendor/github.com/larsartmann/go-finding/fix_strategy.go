package finding

// FixStrategy indicates how a finding can be remediated.
type FixStrategy string

const (
	// FixStrategyNone indicates no fix is available.
	FixStrategyNone FixStrategy = "none"
	// FixStrategySuggest provides a human-readable suggestion.
	FixStrategySuggest FixStrategy = "suggest"
	// FixStrategyDirect can be automatically applied.
	FixStrategyDirect FixStrategy = "direct"
	// FixStrategyAI requires AI assistance.
	// Pipeline triage groups this with FixStrategySuggest (no auto-apply).
	// NeedsAI() is defined but no AI backend exists yet. Reserve this value
	// for future AI-powered remediation — do not remove.
	FixStrategyAI FixStrategy = "ai"
)

// IsValid returns true if the fix strategy is a valid value.
func (f FixStrategy) IsValid() bool {
	switch f {
	case FixStrategyNone, FixStrategySuggest, FixStrategyDirect, FixStrategyAI:
		return true
	}

	return false
}

// CanAutoApply returns true if this fix strategy can be automatically applied.
func (f FixStrategy) CanAutoApply() bool {
	return f == FixStrategyDirect
}

// NeedsAI returns true if this fix strategy requires AI assistance.
func (f FixStrategy) NeedsAI() bool {
	return f == FixStrategyAI
}

// String returns the string representation of the fix strategy.
func (f FixStrategy) String() string {
	return string(f)
}

// NormalizeFixStrategy converts the empty string (the zero value) to
// FixStrategyNone. This ensures there is exactly one canonical "no fix"
// state, eliminating the split-brain where "" and "none" both meant "no fix"
// but compared unequal. All public entry points (Validate, Builder.Build,
// FromLSP, SARIF import) should call this.
func NormalizeFixStrategy(f FixStrategy) FixStrategy {
	if f == "" {
		return FixStrategyNone
	}

	return f
}
