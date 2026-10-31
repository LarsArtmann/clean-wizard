package finding

import (
	"slices"
)

// FilterFunc is a predicate for filtering findings.
type FilterFunc func(Finding) bool

// Filter returns findings that match all predicates.
// If no predicates are provided, returns a copy of all findings.
func Filter(findings []Finding, predicates ...FilterFunc) []Finding {
	if len(predicates) == 0 {
		return slices.Clone(findings)
	}

	result := make([]Finding, 0, len(findings))
	for _, f := range findings {
		if matchesAll(f, predicates) {
			result = append(result, f)
		}
	}

	return result
}

// FilterInPlace filters findings in place, modifying the input slice.
// Returns the filtered slice (which may be a sub-slice of the input).
func FilterInPlace(findings []Finding, predicates ...FilterFunc) []Finding {
	if len(predicates) == 0 {
		return findings
	}

	n := 0

	for _, f := range findings {
		if matchesAll(f, predicates) {
			findings[n] = f
			n++
		}
	}

	// Zero out tail to allow GC of referenced structs (Range, Suppression, etc.)
	for i := n; i < len(findings); i++ {
		findings[i] = Finding{}
	}

	return findings[:n]
}

// matchesAll returns true if the finding matches all predicates.
func matchesAll(f Finding, predicates []FilterFunc) bool {
	for _, p := range predicates {
		if !p(f) {
			return false
		}
	}

	return true
}

// BySeverity returns a filter for the given severity.
func BySeverity(sev Severity) FilterFunc {
	return func(f Finding) bool {
		return f.Severity == sev
	}
}

// BySeverityAtLeast returns a filter for severity >= the given level.
// Findings with invalid severity are excluded (return false).
func BySeverityAtLeast(sev Severity) FilterFunc {
	return func(f Finding) bool {
		return f.Severity.GreaterThanOrEqual(sev)
	}
}

// ByCategory returns a filter for the given category.
func ByCategory(cat Category) FilterFunc {
	return func(f Finding) bool {
		return f.Category == cat
	}
}

// ByFixStrategy returns a filter for the given fix strategy.
func ByFixStrategy(fs FixStrategy) FilterFunc {
	return func(f Finding) bool {
		return f.FixStrategy == fs
	}
}

// ByTool returns a filter for the given tool name.
func ByTool(tool ToolName) FilterFunc {
	return func(f Finding) bool {
		return f.ToolName == tool
	}
}

// ByRule returns a filter for the given rule.
func ByRule(rule RuleName) FilterFunc {
	return func(f Finding) bool {
		return f.Rule == rule
	}
}

// ByFile returns a filter for findings in the given file.
func ByFile(file FilePath) FilterFunc {
	return func(f Finding) bool {
		return f.Position.File == file
	}
}

// ByConfidence returns a filter for the exact confidence level.
func ByConfidence(c Confidence) FilterFunc {
	return func(f Finding) bool {
		return f.Confidence == c
	}
}

// ByConfidenceAtLeast returns a filter for confidence >= the given level.
func ByConfidenceAtLeast(c Confidence) FilterFunc {
	return func(f Finding) bool {
		return f.Confidence >= c
	}
}

// NotSuppressed returns a filter for non-suppressed findings.
func NotSuppressed(f Finding) bool {
	return !f.IsSuppressed()
}

// Negate inverts a filter: returns findings that do NOT match the given predicate.
func Negate(predicate FilterFunc) FilterFunc {
	return func(f Finding) bool {
		return !predicate(f)
	}
}

// AnyOf returns a filter that matches if ANY of the given predicates match.
// This is the complement of Filter, which requires ALL predicates to match.
func AnyOf(predicates ...FilterFunc) FilterFunc {
	return func(f Finding) bool {
		for _, p := range predicates {
			if p(f) {
				return true
			}
		}

		return len(predicates) == 0
	}
}

// WithFix returns a filter for findings with fixes.
func WithFix(f Finding) bool {
	return f.HasFix()
}

// WithSuggestion returns a filter for findings with suggestions.
func WithSuggestion(f Finding) bool {
	return f.HasSuggestion()
}

// GroupBy groups findings by a key extractor function.
func GroupBy(findings []Finding, keyFn func(Finding) string) map[string][]Finding {
	groups := make(map[string][]Finding, len(findings))

	for _, finding := range findings {
		key := keyFn(finding)
		groups[key] = append(groups[key], finding)
	}

	return groups
}

// GroupByFile groups findings by file path.
func GroupByFile(findings []Finding) map[FilePath][]Finding {
	groups := make(map[FilePath][]Finding, len(findings))
	for _, f := range findings {
		groups[f.Position.File] = append(groups[f.Position.File], f)
	}

	return groups
}

// GroupBySeverity groups findings by severity.
func GroupBySeverity(findings []Finding) map[Severity][]Finding {
	groups := make(map[Severity][]Finding, len(findings))
	for _, finding := range findings {
		groups[finding.Severity] = append(groups[finding.Severity], finding)
	}

	return groups
}

// GroupByCategory groups findings by category.
func GroupByCategory(findings []Finding) map[Category][]Finding {
	groups := make(map[Category][]Finding, len(findings))

	for _, f := range findings {
		groups[f.Category] = append(groups[f.Category], f)
	}

	return groups
}

// SortByPosition sorts findings by file path, then line, then column.
func SortByPosition(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int {
		return a.Position.Compare(b.Position)
	})
}

// SortBySeverity sorts findings by severity (most severe first).
func SortBySeverity(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int {
		return -a.Severity.Compare(b.Severity)
	})
}
