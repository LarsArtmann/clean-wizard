package finding

import (
	"fmt"
	"maps"
	"strings"
	"time"
)

// Normalized returns a copy of the finding with FixStrategy normalized
// (empty string converted to FixStrategyNone). Use this when you need
// the canonical form after direct struct construction, since Validate()
// has a value receiver and cannot mutate the original.
func (f Finding) Normalized() Finding {
	result := f
	result.FixStrategy = NormalizeFixStrategy(f.FixStrategy)

	return result
}

// Clone returns a deep copy of the finding.
func (f Finding) Clone() Finding {
	clone := f

	if f.Range != nil {
		r := *f.Range
		clone.Range = &r
	}

	if len(f.Related) > 0 {
		clone.Related = make([]RelatedRef, len(f.Related))
		for i, r := range f.Related {
			clone.Related[i] = r
			if r.Range != nil {
				rr := *r.Range
				clone.Related[i].Range = &rr
			}
		}
	}

	if f.Suppression != nil {
		s := *f.Suppression
		if s.ExpiresAt != nil {
			t := *s.ExpiresAt
			s.ExpiresAt = &t
		}

		clone.Suppression = &s
	}

	if len(f.Tags) > 0 {
		clone.Tags = make([]Tag, len(f.Tags))
		copy(clone.Tags, f.Tags)
	}

	if len(f.Metadata) > 0 {
		clone.Metadata = maps.Clone(f.Metadata)
	}

	return clone
}

// IsSuppressed returns true if this finding is suppressed at the current time.
func (f Finding) IsSuppressed() bool {
	return f.IsSuppressedAt(time.Now())
}

// IsSuppressedAt returns true if this finding is suppressed at the given time.
// Use this in tests for deterministic suppression checks.
// A finding is suppressed only if its Suppression is valid (correct Kind and
// non-empty Rule) and not expired, consistent with [Suppression.IsActive].
func (f Finding) IsSuppressedAt(now time.Time) bool {
	return f.Suppression != nil && f.Suppression.IsActive(now)
}

// HasFix reports whether this finding has any fix available. This is the
// canonical "is fixable?" check.
//
// Fixability lattice:
//   - IsAutoFixable() ⟹ HasFix()  (strict subset)
//   - HasFix() requires a valid FixStrategy AND code data where applicable
//
// Returns true for FixStrategyDirect when BeforeCode, AfterCode, or an edit
// list is present.
// Returns true for FixStrategySuggest/FixStrategyAI when AfterCode is present.
// Returns false for FixStrategyNone, empty strategy, or any strategy missing
// required code data.
//
// Note: HasFix() respects Validate() semantics — a FixStrategyDirect finding
// without code data returns false (it would fail validation).
func (f Finding) HasFix() bool {
	strategy := NormalizeFixStrategy(f.FixStrategy)

	switch strategy {
	case FixStrategyNone:
		return false
	case FixStrategyDirect:
		return f.BeforeCode != "" || f.AfterCode != "" || f.HasEditList()
	case FixStrategySuggest, FixStrategyAI:
		return f.AfterCode != ""
	default:
		return false
	}
}

// IsAutoFixable reports whether the pipeline can automatically apply this fix.
// Stricter than HasFix: requires FixStrategyDirect AND (BeforeCode, AfterCode,
// or an edit list).
// Use HasFix() to check if any fix exists; use IsAutoFixable() to check if
// the pipeline will attempt auto-application.
func (f Finding) IsAutoFixable() bool {
	return f.Normalized().FixStrategy == FixStrategyDirect &&
		(f.BeforeCode != "" || f.AfterCode != "" || f.HasEditList())
}

// HasSuggestion returns true if this finding has a human-readable suggestion.
func (f Finding) HasSuggestion() bool {
	return f.Suggestion != "" || (f.BeforeCode != "" && f.AfterCode != "")
}

// HasCategory returns true if this finding has a category set.
func (f Finding) HasCategory() bool {
	return f.Category != ""
}

// NormalizedConfidence returns the confidence clamped to [0.0, 1.0].
func (f Finding) NormalizedConfidence() Confidence {
	return f.Confidence.Clamp()
}

// String returns a human-readable summary of the finding.
func (f Finding) String() string {
	var b strings.Builder
	b.WriteString(string(f.Severity))
	b.WriteByte(' ')
	b.WriteString(string(f.ToolName))
	b.WriteString(" [")
	b.WriteString(string(f.Rule))
	b.WriteString("] ")
	b.WriteString(f.Position.String())
	b.WriteString(": ")
	b.WriteString(f.Message)

	if f.Category != "" {
		b.WriteString(" (")
		b.WriteString(string(f.Category))
		b.WriteByte(')')
	}

	return b.String()
}

// HasCodeChange reports whether the finding carries any code change data
// (BeforeCode, AfterCode, or an edit list). This is a low-level check used by
// the FixEngine; prefer HasFix() or IsAutoFixable() for higher-level fixability
// decisions.
func (f Finding) HasCodeChange() bool {
	return f.BeforeCode != "" || f.AfterCode != "" || f.HasEditList()
}

// HasEditList reports whether the finding carries a typed edit list
// (Finding.Edits), the authoritative multi-edit fix representation.
func (f Finding) HasEditList() bool {
	return len(f.Edits) > 0
}

// HasRange reports whether the finding has a valid range set.
func (f Finding) HasRange() bool {
	return f.Range != nil && f.Range.IsValid()
}

// Preview returns a unified-diff-style preview of the fix, or empty string if
// the finding has no fixable code change (BeforeCode and AfterCode both empty
// and no Edits list).
//
// Findings with a typed Edits list render one "-/+" pair per edit, prefixed
// by the edit's start position: spans show removals as a location (TextEdit
// does not carry the removed text), insertions show only the added text.
// BeforeCode/AfterCode findings render as a single "- old" / "+ new" pair.
func (f Finding) Preview() string {
	if !f.HasCodeChange() {
		return ""
	}

	var b strings.Builder

	if len(f.Edits) > 0 {
		for _, e := range f.Edits {
			loc := e.Start.String()

			if e.HasSpan() {
				if e.Start.Offset >= 0 && e.End.Offset > e.Start.Offset {
					fmt.Fprintf(&b, "- %s: %d bytes\n", loc, e.End.Offset-e.Start.Offset)
				} else {
					fmt.Fprintf(&b, "- %s: span\n", loc)
				}
			}

			if e.NewText != "" {
				fmt.Fprintf(&b, "+ %s: %s\n", loc, e.NewText)
			}
		}

		return b.String()
	}

	if f.BeforeCode != "" {
		b.WriteString("- ")
		b.WriteString(f.BeforeCode)
		b.WriteByte('\n')
	}

	if f.AfterCode != "" {
		b.WriteString("+ ")
		b.WriteString(f.AfterCode)
		b.WriteByte('\n')
	}

	return b.String()
}

// Validate performs comprehensive validation and returns an error if the
// finding is invalid. It checks all fields that IsValid checks plus
// additional constraints: FixStrategy validity, Confidence range,
// and structural consistency.
