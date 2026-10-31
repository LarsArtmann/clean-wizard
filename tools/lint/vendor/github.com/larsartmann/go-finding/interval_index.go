package finding

import (
	"cmp"
	"slices"
)

// Interval represents a half-open range [Start, End) on a single axis.
type Interval[T any] struct {
	Start int
	End   int
	Value T
}

// IntervalIndex supports efficient overlap queries over intervals.
// Create one with [NewIntervalIndex]. The index is immutable after construction.
//
// Query complexity is O(n + k) where n is the number of intervals and k is
// the number of matching results. This is a sorted-slice implementation
// (not an augmented interval tree), which is fast enough for typical finding
// counts. For very large interval sets with adversarial distributions,
// consider implementing a true O(log n + k) interval tree.
//
// Use cases include finding overlapping findings by line range, conflict
// detection, and spatial correlation queries.
type IntervalIndex[T any] struct {
	byStart []Interval[T]
}

// NewIntervalIndex builds an interval index from the given intervals.
// If the input is empty, returns an index that answers all queries with nil.
func NewIntervalIndex[T any](intervals []Interval[T]) *IntervalIndex[T] {
	sorted := make([]Interval[T], len(intervals))
	copy(sorted, intervals)

	slices.SortFunc(sorted, func(a, b Interval[T]) int {
		if c := cmp.Compare(a.Start, b.Start); c != 0 {
			return c
		}

		return cmp.Compare(a.End, b.End)
	})

	return &IntervalIndex[T]{byStart: sorted}
}

// Query returns all intervals that overlap the half-open range [start, end).
// Overlap means: iv.Start < end && start < iv.End.
// Returns nil if no intervals match.
func (idx *IntervalIndex[T]) Query(start, end int) []Interval[T] {
	if len(idx.byStart) == 0 || start >= end {
		return nil
	}

	// Binary search for the first interval where Start >= end.
	// Everything before that could overlap.
	cutoff, _ := slices.BinarySearchFunc(idx.byStart, end, func(iv Interval[T], target int) int {
		return cmp.Compare(iv.Start, target)
	})

	var result []Interval[T]

	for i := range cutoff {
		iv := idx.byStart[i]

		if iv.End > start {
			result = append(result, iv)
		}
	}

	return result
}

// Len returns the number of intervals in the index.
func (idx *IntervalIndex[T]) Len() int {
	return len(idx.byStart)
}
