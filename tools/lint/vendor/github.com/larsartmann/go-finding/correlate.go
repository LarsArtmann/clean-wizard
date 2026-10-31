package finding

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

// Correlation heuristics constants.
const (
	minFindingsInFile     = 2     // Minimum findings in a file for correlation analysis
	maxLineDiff           = 5     // Maximum line difference for considering findings related
	correlationScoreScale = 5.0   // For converting lineDiff to score
	minCorrelationScore   = 0.5   // Minimum correlation score for matching
	maxCorrelations       = 10000 // Maximum correlations to prevent O(n²) hangs
)

// reasonOverlappingRanges is the correlation reason for overlapping range findings.
const reasonOverlappingRanges = "overlapping ranges"

// CorrelationScore measures the strength of a correlation between findings.
// Unlike Confidence (which measures certainty of a single finding),
// CorrelationScore measures how strongly two findings are related.
type CorrelationScore float64

// IsValid returns true if the score is in the valid range [0.0, 1.0].
func (s CorrelationScore) IsValid() bool {
	return float64(s) >= 0.0 && float64(s) <= 1.0
}

// String returns a human-readable representation of the correlation score.
func (s CorrelationScore) String() string {
	return fmt.Sprintf("%.2f", float64(s))
}

// Correlation represents a relationship between two or more findings.
type Correlation struct {
	FindingIDs []ID             `json:"findingIds"`
	Reason     string           `json:"reason"` // Why they're correlated
	Score      CorrelationScore `json:"score"`  // 0.0-1.0 correlation strength
}

// Correlate finds potentially related findings across tools.
// Uses two strategies depending on the data:
//   - For findings with Range: uses IntervalIndex for O(n + k) overlap queries.
//   - For point-only findings: uses line-proximity heuristics (same file + nearby lines).
//
// This can be used standalone or enabled in Pipeline via Config.CorrelateFindings.
// When enabled, the pipeline populates PipelineResult.Correlations automatically.
//
// # Complexity
//
// For range-based findings: O(n log n) to build the index, O(n + k) per query.
// For point-based findings: O(n) per file with sorted early-break.
//
// The maxCorrelations constant (10,000) caps total output across both strategies.
func Correlate(findings []Finding) []Correlation {
	capHint := min(len(findings), maxCorrelations)
	correlations := make([]Correlation, 0, capHint)

	byFile := GroupByFile(findings)

	files := slices.Sorted(maps.Keys(byFile))

	for _, file := range files {
		if len(correlations) >= maxCorrelations {
			break
		}

		fileFindings := byFile[file]
		if len(fileFindings) < minFindingsInFile {
			continue
		}

		withRange := make([]Finding, 0, len(fileFindings))
		withoutRange := make([]Finding, 0, len(fileFindings))

		for _, f := range fileFindings {
			if f.HasRange() {
				withRange = append(withRange, f)
			} else {
				withoutRange = append(withoutRange, f)
			}
		}

		if len(withRange) >= minFindingsInFile {
			correlations = correlateByOverlap(withRange, correlations)
		}

		if len(withoutRange) >= minFindingsInFile {
			correlations = correlateByProximity(withoutRange, correlations)
		}

		// Cross-correlate range and point findings.
		if len(withRange) > 0 && len(withoutRange) > 0 {
			correlations = correlateRangesAndPoints(withRange, withoutRange, correlations)
		}
	}

	return correlations
}

// buildLineIntervals converts range-based findings into line intervals for overlap queries.
// End is half-open (End.Line + 1) so adjacent ranges don't falsely overlap.
func buildLineIntervals(findings []Finding) []Interval[int] {
	intervals := make([]Interval[int], len(findings))
	for i, f := range findings {
		intervals[i] = Interval[int]{
			Start: f.Range.Start.Line,
			End:   f.Range.End.Line + 1, // half-open
			Value: i,
		}
	}

	return intervals
}

// correlateByOverlap uses IntervalIndex to find overlapping range-based findings.
func correlateByOverlap(findings []Finding, correlations []Correlation) []Correlation {
	idx := NewIntervalIndex(buildLineIntervals(findings))

	for i, f1 := range findings {
		if f1.Range == nil {
			continue
		}

		start := f1.Range.Start.Line
		end := f1.Range.End.Line + 1

		overlaps := idx.Query(start, end)
		for _, ov := range overlaps {
			j := ov.Value
			if j <= i {
				continue
			}

			f2 := findings[j]
			if f1.ToolName == f2.ToolName {
				continue
			}

			overlapLines := overlapLength(
				f1.Range.Start.Line, f1.Range.End.Line,
				f2.Range.Start.Line, f2.Range.End.Line,
			)
			shortest := min(f1.Range.End.Line-f1.Range.Start.Line, f2.Range.End.Line-f2.Range.Start.Line) + 1

			var score float64
			if shortest > 0 {
				score = float64(overlapLines) / float64(shortest)
			}

			if score > minCorrelationScore {
				correlations = append(correlations, Correlation{
					FindingIDs: []ID{f1.ID, f2.ID},
					Reason:     reasonOverlappingRanges,
					Score:      CorrelationScore(score),
				})
				if len(correlations) >= maxCorrelations {
					return correlations
				}
			}
		}
	}

	return correlations
}

// correlateByProximity uses line-proximity heuristics for point-based findings.
func correlateByProximity(findings []Finding, correlations []Correlation) []Correlation {
	slices.SortFunc(findings, func(a, b Finding) int {
		return cmp.Compare(a.Position.Line, b.Position.Line)
	})

	for i, f1 := range findings {
		if f1.Position.Line == 0 {
			continue
		}

		for _, f2 := range findings[i+1:] {
			if f1.ToolName == f2.ToolName {
				continue
			}

			if f2.Position.Line == 0 {
				continue
			}

			lineDiff := f2.Position.Line - f1.Position.Line
			if lineDiff > maxLineDiff {
				break
			}

			confidence := 1.0 - (float64(lineDiff) / correlationScoreScale)
			if confidence > minCorrelationScore {
				correlations = append(correlations, Correlation{
					FindingIDs: []ID{f1.ID, f2.ID},
					Reason:     "same file, nearby lines",
					Score:      CorrelationScore(confidence),
				})
				if len(correlations) >= maxCorrelations {
					return correlations
				}
			}
		}
	}

	return correlations
}

// correlateRangesAndPoints correlates range-based findings with point-based findings.
func correlateRangesAndPoints(
	rangeFindings []Finding,
	pointFindings []Finding,
	correlations []Correlation,
) []Correlation {
	idx := NewIntervalIndex(buildLineIntervals(rangeFindings))

	for _, pointFinding := range pointFindings {
		line := pointFinding.Position.Line

		overlaps := idx.Query(line, line+1)
		for _, ov := range overlaps {
			rangeFinding := rangeFindings[ov.Value]
			if pointFinding.ToolName == rangeFinding.ToolName {
				continue
			}

			span := rangeFinding.Range.End.Line - rangeFinding.Range.Start.Line + 1

			var score float64
			if span > 0 {
				score = 1.0 - (1.0 / float64(span))
			}

			if score > minCorrelationScore {
				correlations = append(correlations, Correlation{
					FindingIDs: []ID{pointFinding.ID, rangeFinding.ID},
					Reason:     "point within range",
					Score:      CorrelationScore(score),
				})
				if len(correlations) >= maxCorrelations {
					return correlations
				}
			}
		}
	}

	return correlations
}

// overlapLength returns the number of overlapping lines between two ranges.
func overlapLength(s1, e1, s2, e2 int) int {
	start := max(s1, s2)

	end := min(e1, e2)
	if start > end {
		return 0
	}

	return end - start + 1
}
