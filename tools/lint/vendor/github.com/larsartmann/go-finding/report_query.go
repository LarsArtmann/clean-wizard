package finding

import (
	"iter"
	"slices"
	"time"
)

// FindingsSnapshot returns a deep copy of all findings in the report.
// The returned slice is safe for concurrent use without holding any lock,
// making it suitable for v1.0 migration from direct Findings slice access.
// Each Finding is fully cloned via Clone(), so mutations are isolated.
func (r *Report) FindingsSnapshot() []Finding {
	return withReadLock(r, func() []Finding {
		if len(r.findings) == 0 {
			return nil
		}

		snapshot := make([]Finding, len(r.findings))
		for i, f := range r.findings {
			snapshot[i] = f.Clone()
		}

		return snapshot
	})
}

// ActiveFindings returns all non-suppressed findings.
// Uses time.Now() for suppression expiry checks. For deterministic results
// in tests, filter Findings directly with IsSuppressedAt.
// Safe for concurrent use.
func (r *Report) ActiveFindings() []Finding {
	return withReadLock(r, func() []Finding {
		now := time.Now()
		active := make([]Finding, 0, len(r.findings))

		for _, f := range r.findings {
			if !f.IsSuppressedAt(now) {
				active = append(active, f)
			}
		}

		return active
	})
}

// BySeverity returns findings filtered by severity, excluding suppressed.
// For composable filtering, use filter.BySeverity with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) BySeverity(sev Severity) []Finding {
	return Filter(r.ActiveFindings(), BySeverity(sev))
}

// CountBySeverity returns the count of findings for the given severity,
// including suppressed findings. Uses the pre-computed summary.
func (r *Report) CountBySeverity(sev Severity) int {
	return withReadLock(r, func() int {
		return r.Summary.BySeverity[sev]
	})
}

// ByCategory returns findings filtered by category, excluding suppressed.
// For composable filtering, use filter.ByCategory with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) ByCategory(cat Category) []Finding {
	return Filter(r.ActiveFindings(), ByCategory(cat))
}

// ByFixStrategy returns findings filtered by fix strategy, excluding suppressed.
// For composable filtering, use filter.ByFixStrategy with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) ByFixStrategy(fs FixStrategy) []Finding {
	return Filter(r.ActiveFindings(), ByFixStrategy(fs))
}

// GroupFindings returns findings grouped by GroupID, excluding suppressed.
// Findings without a GroupID are omitted. Map iteration order is
// unspecified; when deterministic group order matters (tests, serialization,
// stable CLI output), use [Report.GroupFindingsSorted]. Findings within a
// group preserve report order in both variants.
// Safe for concurrent use.
func (r *Report) GroupFindings() map[GroupID][]Finding {
	groups := make(map[GroupID][]Finding)

	for _, f := range r.ActiveFindings() {
		if f.GroupID == "" {
			continue
		}

		groups[f.GroupID] = append(groups[f.GroupID], f)
	}

	if len(groups) == 0 {
		return nil
	}

	return groups
}

// Group is one GroupID group of findings, as returned by
// [Report.GroupFindingsSorted].
type Group struct {
	// ID is the group's GroupID.
	ID GroupID
	// Findings are the group members in report order.
	Findings []Finding
}

// GroupFindingsSorted returns the GroupID groups in deterministic order
// (sorted by GroupID), excluding suppressed findings. It is the ordered
// variant of [Report.GroupFindings] for tests, serialization, and stable
// output. Returns nil when no grouped findings exist.
// Safe for concurrent use.
func (r *Report) GroupFindingsSorted() []Group {
	groups := r.GroupFindings()
	if len(groups) == 0 {
		return nil
	}

	ids := make([]GroupID, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	sorted := make([]Group, 0, len(ids))
	for _, id := range ids {
		sorted = append(sorted, Group{ID: id, Findings: groups[id]})
	}

	return sorted
}

// FindByID returns the finding with the given ID, or nil if not found.
// The returned Finding is a shallow copy; modifications to value fields do not
// affect the report, but mutations to slice/map fields (Tags, Related, Metadata)
// will be shared. Use Clone() for a deep copy.
// Safe for concurrent use.
func (r *Report) FindByID(id ID) *Finding {
	return withReadLock(r, func() *Finding {
		for i, f := range r.findings {
			if f.ID == id {
				cp := r.findings[i]

				return &cp
			}
		}

		return nil
	})
}

// FindByRule returns all non-suppressed findings matching the given rule name.
// Safe for concurrent use.
func (r *Report) FindByRule(rule RuleName) []Finding {
	return Filter(r.ActiveFindings(), ByRule(rule))
}

// Len returns the number of findings in the report.
// Safe for concurrent use.
func (r *Report) Len() int {
	return withReadLock(r, func() int {
		return len(r.findings)
	})
}

// Filter returns a new report containing only findings that match all predicates.
// Safe for concurrent use.
func (r *Report) Filter(predicates ...FilterFunc) *Report {
	filtered := withReadLock(r, func() []Finding {
		return Filter(r.findings, predicates...)
	})

	result := NewReport(r.Tool)
	result.AddFindings(filtered)

	return result
}

// Map returns a new report with the given function applied to each finding.
// Safe for concurrent use.
func (r *Report) Map(fn func(Finding) Finding) *Report {
	findings := r.readFindings()

	result := NewReport(r.Tool)
	for _, f := range findings {
		result.AddFinding(fn(f))
	}

	return result
}

// All returns all findings in the report (including suppressed).
// The yielded Finding values are shallow copies; modifications to value fields
// do not affect the report, but mutations to slice/map fields (Tags, Related,
// Metadata) will be shared. Use Clone() for a deep copy.
//
// IMPORTANT: The returned iterator holds a read lock for the duration of
// iteration. You MUST exhaust the iterator (e.g., with a break or range)
// to release the lock. If you need a snapshot without holding the lock,
// call ActiveFindings() or use Filter.
func (r *Report) All() iter.Seq[Finding] {
	return func(yield func(Finding) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		for _, f := range r.findingsLocked() {
			if !yield(f) {
				return
			}
		}
	}
}
