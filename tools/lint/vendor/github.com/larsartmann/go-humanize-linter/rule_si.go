package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleSI (H005) detects manual SI-prefix formatting ("1.2K", "3.4M") that
// should use humanize.SI or humanize.SIWithDigits.
//
// Triggers when a function contains BOTH:
//   - Division by 1000 or 1000000 (or 1000000000)
//   - A standalone "K" or "M" string literal used as a suffix
//
// This is distinct from H001 (byte formatting) because SI uses 1000-base and
// does not append "B".
func RuleSI() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH005,
			Name:        "manual-si-format",
			Description: "Manual SI-prefix formatting (K/M) — use humanize.SI or humanize.SIWithDigits instead",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH005, detectSIFormat)
		},
	}
}

func detectSIFormat(fset *token.FileSet, _ *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	// Exclude byte formatting (H001 handles that).
	if len(countByteUnits(fn)) > 0 {
		return nil
	}

	div1000 := hasDivisionBy1000(fn)
	kmSuffix := hasKMSuffix(fn)

	if !div1000 || !kmSuffix {
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH005,
			fmt.Sprintf(
				"manual SI-prefix formatting (div1000=%v, kmSuffix=%v) — use humanize.SI instead",
				div1000, kmSuffix,
			),
			"Replace with humanize.SI(float64(n), \"\") or humanize.SIWithDigits(float64(n), 1, \"\").",
			line, col, filePath, finding.ConfidenceHigh,
		),
	}
}
