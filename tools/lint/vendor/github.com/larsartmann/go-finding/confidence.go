package finding

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// Confidence represents the certainty level of a finding on a 0.0–1.0 scale.
// Use named constants (ConfidenceLow, ConfidenceMedium, ConfidenceHigh) for
// common values, or Confidence(f) for custom levels.
// The zero value is valid and represents no confidence information.
//
// Be aware: Direct construction with Confidence values outside [0.0, 1.0] is
// possible (e.g., Finding{Confidence: 1.5}). The Validate() method catches
// this. For guaranteed-valid values, use the Builder API (WithConfidence)
// or NewFinding (both clamp automatically).
type Confidence float64

// Standard confidence levels.
const (
	ConfidenceNone   Confidence = 0.0
	ConfidenceLow    Confidence = 0.25
	ConfidenceMedium Confidence = 0.5
	ConfidenceHigh   Confidence = 0.75
	ConfidenceFull   Confidence = 1.0
)

// IsValid returns true if the confidence is within [0.0, 1.0].
func (c Confidence) IsValid() bool {
	return c >= 0 && c <= 1
}

// Clamp returns the confidence clamped to [0.0, 1.0].
func (c Confidence) Clamp() Confidence {
	if c < 0 {
		return 0
	}

	if c > 1 {
		return 1
	}

	return c
}

// Compare returns -1, 0, or +1 depending on whether c is less than, equal to,
// or greater than other.
func (c Confidence) Compare(other Confidence) int {
	return cmp.Compare(float64(c), float64(other))
}

// String returns the confidence as a human-readable string.
// Named levels return their label (e.g., "medium"), custom values return a decimal.
func (c Confidence) String() string {
	switch c {
	case ConfidenceNone:
		return "none"
	case ConfidenceLow:
		return "low"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceHigh:
		return "high"
	case ConfidenceFull:
		return "full"
	default:
		return fmt.Sprintf("%.2f", float64(c))
	}
}

// ErrInvalidConfidence is returned by ParseConfidence when the input string is
// not a recognized confidence level name or a valid decimal number.
var ErrInvalidConfidence = fmt.Errorf("invalid confidence level: use %s, %s, %s, %s, %s, or a decimal in [0.0, 1.0]",
	ConfidenceNone, ConfidenceLow, ConfidenceMedium, ConfidenceHigh, ConfidenceFull)

// ParseConfidence converts a confidence level string to a Confidence value.
// It is the inverse of [Confidence.String]: every named level ("none", "low",
// "medium", "high", "full") maps back to its constant. Decimal strings (e.g.
// "0.42") are also accepted and clamped to [0.0, 1.0]. An empty string defaults
// to ConfidenceLow.
//
// This eliminates the per-consumer switch statements that every CLI linter
// reinvents for its --min-confidence flag.
func ParseConfidence(s string) (Confidence, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "low":
		return ConfidenceLow, nil
	case "none":
		return ConfidenceNone, nil
	case "medium":
		return ConfidenceMedium, nil
	case "high":
		return ConfidenceHigh, nil
	case "full":
		return ConfidenceFull, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ConfidenceNone, fmt.Errorf("%w: %q", ErrInvalidConfidence, s)
	}

	c := Confidence(f)
	if !c.IsValid() {
		return ConfidenceNone, fmt.Errorf("%w: %q", ErrInvalidConfidence, s)
	}

	return c, nil
}
