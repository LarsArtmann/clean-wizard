package finding

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Severity represents the severity level of a finding.
type Severity string

// Severity levels for findings, ordered by urgency.
const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// severityPriorities maps each canonical severity to its priority label.
var severityPriorities = map[Severity]string{
	SeverityCritical: "critical",
	SeverityError:    "high",
	SeverityWarning:  "medium",
	SeverityInfo:     "low",
}

// IsValid returns true if the severity is a valid value.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	}

	return false
}

// comparisonOp represents a severity comparison operation.
type comparisonOp int

const (
	cmpGreaterThan comparisonOp = iota
	cmpLessThan
	cmpGreaterThanOrEqual
	cmpLessThanOrEqual
)

// GreaterThan returns true if this severity is greater than the other.
// Order: info < warning < error < critical.
func (s Severity) GreaterThan(other Severity) bool {
	return s.compareOp(other, cmpGreaterThan)
}

// LessThan returns true if this severity is less than the other.
func (s Severity) LessThan(other Severity) bool {
	return s.compareOp(other, cmpLessThan)
}

// GreaterThanOrEqual returns true if this severity is greater than or equal to the other.
func (s Severity) GreaterThanOrEqual(other Severity) bool {
	return s.compareOp(other, cmpGreaterThanOrEqual)
}

// LessThanOrEqual returns true if this severity is less than or equal to the other.
func (s Severity) LessThanOrEqual(other Severity) bool {
	return s.compareOp(other, cmpLessThanOrEqual)
}

// compareOp is an internal helper that performs comparison based on the given operation.
func (s Severity) compareOp(other Severity, op comparisonOp) bool {
	if !s.isValidWith(other) {
		return false
	}

	c := s.Compare(other)

	switch op {
	case cmpGreaterThan:
		return c > 0
	case cmpLessThan:
		return c < 0
	case cmpGreaterThanOrEqual:
		return c >= 0
	case cmpLessThanOrEqual:
		return c <= 0
	default:
		return false
	}
}

// String returns the string representation of the severity.
func (s Severity) String() string {
	return string(s)
}

// Badge returns a human-readable severity badge with emoji (e.g., "🔴 CRITICAL").
// Derived from [Severity.Emoji] and the uppercased severity name.
// Returns the string value unchanged for unknown severities.
func (s Severity) Badge() string {
	emoji := s.Emoji()
	if emoji == "" {
		return string(s)
	}

	return emoji + " " + strings.ToUpper(string(s))
}

// Emoji returns the emoji representing the severity level.
// Returns an empty string for unknown severities.
func (s Severity) Emoji() string {
	switch s {
	case SeverityCritical:
		return "🔴"
	case SeverityError:
		return "🟠"
	case SeverityWarning:
		return "🟡"
	case SeverityInfo:
		return "🟢"
	}

	return ""
}

// Compare returns -1, 0, or 1 depending on whether s is less than, equal to,
// or greater than other. Invalid severities rank below all valid ones.
// Two different invalid severities are ordered lexicographically to ensure
// a total ordering.
func (s Severity) Compare(other Severity) int {
	rankS, rankOther := severityRank(s), severityRank(other)
	if rankS != rankOther {
		return cmp.Compare(rankS, rankOther)
	}

	if rankS < 0 {
		// Both invalid — use string comparison as tiebreaker.
		return cmp.Compare(string(s), string(other))
	}

	return 0
}

func severityRank(s Severity) int {
	switch s {
	case SeverityInfo:
		return 0
	case SeverityWarning:
		return 1
	case SeverityError:
		return 2
	case SeverityCritical:
		return 3
	}

	return -1
}

// isValidWith returns true if both severities are valid.
func (s Severity) isValidWith(other Severity) bool {
	return s.IsValid() && other.IsValid()
}

// errInvalidSeverity is returned when parsing an invalid severity string.
var errInvalidSeverity = errors.New("invalid severity")

// severityAliases maps common alternative severity strings to canonical Severity values.
// Guarded by severityAliasesMu for concurrent read/write safety.
var (
	severityAliases = map[string]Severity{
		"warn":       SeverityWarning,
		"high":       SeverityError,
		"medium":     SeverityWarning,
		"low":        SeverityInfo,
		"fatal":      SeverityCritical,
		"critical":   SeverityCritical,
		"crit":       SeverityCritical,
		"note":       SeverityInfo,
		"advice":     SeverityInfo,
		"optional":   SeverityInfo,
		"suggestion": SeverityInfo,
	}
	severityAliasesMu sync.RWMutex
)

// RegisterSeverityAlias adds a custom severity alias for use by ParseSeverity.
// This is safe to call concurrently. If the alias already exists, it is overwritten.
func RegisterSeverityAlias(name string, sev Severity) {
	severityAliasesMu.Lock()
	defer severityAliasesMu.Unlock()

	severityAliases[name] = sev
}

// LookupSeverityAlias returns the canonical Severity for the given alias, if it exists.
// This is safe to call concurrently.
func LookupSeverityAlias(name string) (Severity, bool) {
	severityAliasesMu.RLock()
	defer severityAliasesMu.RUnlock()

	sev, ok := severityAliases[name]

	return sev, ok
}

// ParseSeverity parses a string into a Severity.
// It accepts the canonical names (info, warning, error, critical) and common aliases
// (warn, high, medium, low, fatal, note, advice, suggestion) registered via
// RegisterSeverityAlias.
// Returns an error if the string is not a valid severity level or alias.
func ParseSeverity(s string) (Severity, error) {
	sev := Severity(s)
	if sev.IsValid() {
		return sev, nil
	}

	if alias, ok := LookupSeverityAlias(s); ok {
		return alias, nil
	}

	return "", fmt.Errorf("%w: %q (valid: %s, %s, %s, %s)",
		errInvalidSeverity, s, SeverityInfo, SeverityWarning, SeverityError, SeverityCritical)
}

// MustParseSeverity parses a string into a Severity, panicking on invalid input.
func MustParseSeverity(s string) Severity {
	return must(ParseSeverity(s))
}

// SeverityFromLevel maps a severity level string to a canonical Severity.
// It tries canonical names first, then registered aliases, and falls back
// to the provided fallback if the level is unrecognized. This eliminates the
// severity-mapping switch statements that every consumer independently writes.
func SeverityFromLevel(level string, fallback Severity) Severity {
	sev := Severity(level)
	if sev.IsValid() {
		return sev
	}

	if alias, ok := LookupSeverityAlias(level); ok {
		return alias
	}

	return fallback
}

// PriorityString returns a priority label for the severity, which is the
// reverse mapping of severity levels to common priority terms.
// Critical→"critical", Error→"high", Warning→"medium", Info→"low".
// Returns the string value unchanged for unknown severities.
func (s Severity) PriorityString() string {
	if p, ok := severityPriorities[s]; ok {
		return p
	}

	return string(s)
}
