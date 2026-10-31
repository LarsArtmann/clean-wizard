package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleRelTime (H003) detects manual relative-time formatting ("3 hours ago",
// "in 2 days") that should use humanize.RelTime or humanize.Time.
//
// Triggers when a function contains BOTH:
//   - A time-difference computation (time.Since or .Sub call)
//   - A string literal containing "ago", "from now", or "just now"
//
// AND at least one time-threshold comparison (time.Minute, time.Hour, etc.)
// to suppress false positives where "ago" appears in unrelated contexts.
func RuleRelTime() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH003,
			Name:        "manual-reltime-format",
			Description: "Manual relative-time formatting — use humanize.RelTime or humanize.Time instead of switch/case on duration thresholds",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH003, detectRelTimeFormat)
		},
	}
}

func detectRelTimeFormat(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	aliases := buildImportAliases(file)
	timeDiff := hasTimeSinceOrSub(fn, aliases)
	timeThreshold := hasTimeThresholdComparison(fn, aliases)
	hasAgo := hasAnyStringLiteral(fn, "ago", "from now", "just now")

	// Require time-diff + ago string; threshold comparison as corroboration.
	if !timeDiff || !hasAgo {
		return nil
	}

	confidence := finding.ConfidenceHigh
	if timeThreshold {
		confidence = finding.ConfidenceFull
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH003,
			fmt.Sprintf(
				"manual relative-time formatting (timeSince=%v, thresholds=%v) — use humanize.RelTime instead",
				timeDiff, timeThreshold,
			),
			"Replace with humanize.RelTime(a, b, \"ago\", \"from now\") or humanize.Time(t) for time-since-now.",
			line, col, filePath, confidence,
		),
	}
}
