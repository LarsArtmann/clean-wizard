package finding

import (
	"fmt"
	"iter"
	"strconv"
	"strings"
)

// mergedToolName is the ToolInfo.Name used for reports produced by Combine.
const mergedToolName = "merged"

// emptyToolName is the ToolInfo.Name used for empty reports from Combine.
const emptyToolName = "empty"

// Combine merges multiple reports into a new report with optional deduplication.
// The resulting report has:
//   - Tool.Name = mergedToolName (unless there's only one report)
//   - Findings from all reports
//   - Summary computed from all findings
//
// Use Report.MergeInto(other) to combine two reports without mutation.
func Combine(reports []*Report, opts ...MergeOption) *Report {
	if len(reports) == 0 {
		return NewReport(ToolInfo{Name: emptyToolName})
	}

	if len(reports) == 1 {
		r := reports[0]

		result := &Report{
			Tool:     r.Tool,
			findings: cloneFindings(r.readFindings()),
		}
		result.ComputeSummary()

		return result
	}

	total := 0

	for _, report := range reports {
		if report != nil {
			total += report.Len()
		}
	}

	// Delegate to MergeIter for dedup + clone logic (DRY).
	findings := make([]Finding, 0, total)

	for f := range MergeIter(reports, opts...) {
		findings = append(findings, f)
	}

	merged := &Report{
		Tool: ToolInfo{
			Name:    mergedToolName,
			Version: "",
		},
		findings: findings,
	}
	merged.ComputeSummary()

	return merged
}

func cloneFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return nil
	}

	cloned := make([]Finding, len(findings))
	for i, f := range findings {
		cloned[i] = f.Clone()
	}

	return cloned
}

// MergeOptions controls how reports are merged.
type MergeOptions struct {
	Deduplicate   bool
	DeduplicateBy DeduplicateBy
}

// MergeOption is a functional option for configuring merge behavior.
type MergeOption func(*MergeOptions)

// DeduplicateBy specifies what fields to use for deduplication.
type DeduplicateBy int

// Deduplication strategies control how findings are matched during merge.
const (
	DeduplicateByID       DeduplicateBy = iota // Exact ID matches.
	DeduplicateByPosition                      // File:line:column matching.
	DeduplicateByRule                          // Rule + position matching.
)

// String returns a human-readable name for the deduplication strategy.
func (d DeduplicateBy) String() string {
	switch d {
	case DeduplicateByID:
		return "id"
	case DeduplicateByPosition:
		return "position"
	case DeduplicateByRule:
		return "rule"
	default:
		return fmt.Sprintf("unknown(%d)", d)
	}
}

func defaultMergeOptions() MergeOptions {
	return MergeOptions{
		Deduplicate:   true,
		DeduplicateBy: DeduplicateByID,
	}
}

// WithDeduplication enables/disables deduplication.
func WithDeduplication(enabled bool) MergeOption {
	return func(o *MergeOptions) {
		o.Deduplicate = enabled
	}
}

// WithDeduplicateBy sets the deduplication strategy.
func WithDeduplicateBy(by DeduplicateBy) MergeOption {
	return func(o *MergeOptions) {
		o.DeduplicateBy = by
	}
}

// dedupKeyOverhead accounts for 3 colons + two decimal integers (up to 20 chars).
const dedupKeyOverhead = 23

// positionDedupKey builds a deterministic key from a tool/rule prefix and the
// finding's position (file:line:column). The caller must ensure Position.File is set.
func positionDedupKey(prefix string, f Finding) string {
	var b strings.Builder
	b.Grow(len(prefix) + len(f.Position.File) + dedupKeyOverhead)
	b.WriteString(prefix)
	b.WriteByte(':')
	b.WriteString(string(f.Position.File))
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(f.Position.Line))
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(f.Position.Column))

	return b.String()
}

func dedupKey(finding Finding, opts MergeOptions) (string, bool) {
	switch opts.DeduplicateBy {
	case DeduplicateByID:
		if finding.ID == "" {
			return "", false
		}

		return string(finding.ID), true
	case DeduplicateByPosition:
		if finding.Position.File == "" {
			return "", false
		}

		return positionDedupKey(string(finding.ToolName), finding), true
	case DeduplicateByRule:
		if finding.Position.File == "" {
			return "", false
		}

		return positionDedupKey(string(finding.Rule), finding), true
	default:
		return string(finding.ID), true
	}
}

// MergeIter returns an iterator that yields findings from multiple reports
// in streaming fashion, without loading all findings into memory at once.
// Supports optional deduplication via the same MergeOption as [Combine].
//
// Each report's findings are read under RLock; the iterator does not
// hold locks across reports.
func MergeIter(reports []*Report, opts ...MergeOption) iter.Seq[Finding] {
	return func(yield func(Finding) bool) {
		if len(reports) == 0 {
			return
		}

		options := defaultMergeOptions()
		for _, opt := range opts {
			opt(&options)
		}

		seen := initSeenMap(reports, options.Deduplicate)

		for _, report := range reports {
			if report == nil {
				continue
			}

			if !yieldReportFindings(report, seen, options, yield) {
				return
			}
		}
	}
}

// initSeenMap creates a pre-sized dedup map, or nil if dedup is disabled.
func initSeenMap(reports []*Report, deduplicate bool) map[string]struct{} {
	if !deduplicate {
		return nil
	}

	total := 0

	for _, report := range reports {
		if report != nil {
			total += report.Len()
		}
	}

	return make(map[string]struct{}, total)
}

// yieldReportFindings yields cloned findings from a report, skipping duplicates.
// Returns false if the consumer stopped early.
func yieldReportFindings(
	report *Report,
	seen map[string]struct{},
	options MergeOptions,
	yield func(Finding) bool,
) bool {
	for _, f := range report.readFindings() {
		if options.Deduplicate {
			key, ok := dedupKey(f, options)
			if ok {
				if _, exists := seen[key]; exists {
					continue
				}

				seen[key] = struct{}{}
			}
		}

		if !yield(f.Clone()) {
			return false
		}
	}

	return true
}
