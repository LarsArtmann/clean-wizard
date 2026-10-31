package finding

import (
	"cmp"
	"fmt"
	"slices"
)

// ModifiedPair holds both versions of a modified finding.
type ModifiedPair struct {
	Before Finding
	After  Finding
}

// DiffResult holds the difference between two finding sets.
type DiffResult struct {
	Added     []Finding      // Present in "after" but not "before"
	Removed   []Finding      // Present in "before" but not "after"
	Modified  []ModifiedPair // Present in both but with different content
	Unchanged []Finding      // Present in both with identical content
}

// SortFindingsByID sorts a slice of findings by ID.
func SortFindingsByID(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int { return cmp.Compare(a.ID, b.ID) })
}

// Diff compares two finding sets by ID and categorizes them as added, removed, modified, or unchanged.
// Two findings with the same ID are considered "modified" if their content differs (per Equal()).
//
// Note: Findings are keyed by ID. If the input contains duplicate IDs, only the
// last occurrence per ID is used (standard Go map semantics). Callers should
// deduplicate by ID before diffing if duplicates are expected.
// All result slices are sorted by ID.
func Diff(before, after []Finding) DiffResult {
	beforeSet := make(map[ID]Finding, len(before))
	for _, f := range before {
		beforeSet[f.ID] = f
	}

	afterSet := make(map[ID]Finding, len(after))
	for _, f := range after {
		afterSet[f.ID] = f
	}

	var (
		added     = make([]Finding, 0, len(after))
		removed   = make([]Finding, 0, len(before))
		modified  = make([]ModifiedPair, 0, min(len(before), len(after)))
		unchanged = make([]Finding, 0, min(len(before), len(after)))
	)

	for id, beforeF := range beforeSet {
		afterF, exists := afterSet[id]
		if !exists {
			removed = append(removed, beforeF)
		} else if !beforeF.Equal(afterF) {
			modified = append(modified, ModifiedPair{Before: beforeF, After: afterF})
		} else {
			unchanged = append(unchanged, beforeF)
		}
	}

	for id, f := range afterSet {
		if _, exists := beforeSet[id]; !exists {
			added = append(added, f)
		}
	}

	SortFindingsByID(added)
	SortFindingsByID(removed)
	slices.SortFunc(modified, func(a, b ModifiedPair) int {
		return cmp.Compare(a.Before.ID, b.Before.ID)
	})
	SortFindingsByID(unchanged)

	return DiffResult{Added: added, Removed: removed, Modified: modified, Unchanged: unchanged}
}

// HasChanges reports whether the diff contains any additions, removals, or modifications.
func (d DiffResult) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Modified) > 0
}

// Stats returns a human-readable summary of the diff counts.
func (d DiffResult) Stats() string {
	return fmt.Sprintf(
		"+%d -%d ~%d =%d",
		len(d.Added), len(d.Removed), len(d.Modified), len(d.Unchanged),
	)
}
