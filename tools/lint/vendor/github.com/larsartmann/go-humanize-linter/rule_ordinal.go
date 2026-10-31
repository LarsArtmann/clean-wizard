// rule_ordinal.go — H008 ordinal detector (humanize.Ordinal replacement).

package humanizelint

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleOrdinal (H008) detects hand-rolled English ordinal-suffix formatting
// (1st, 2nd, 3rd, 4th) that should use humanize.Ordinal.
//
// Triggers when a function contains a switch on n%10 or n%100 whose case
// branches return at least 3 of the 4 ordinal suffixes ("st", "nd", "rd",
// "th"). Three distinct suffixes is enough to flag — the fourth is almost
// always the "default" case.
func RuleOrdinal() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH008,
			Name:        "manual-ordinal",
			Description: "Manual ordinal-suffix formatting (1st, 2nd, 3rd, 4th) — use humanize.Ordinal instead of a switch on n%10",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH008, detectOrdinal)
		},
	}
}

func detectOrdinal(fset *token.FileSet, _ *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	if !hasOrdinalSwitch(fn) {
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH008,
			"manual ordinal-suffix formatting (switch n%10/100 with st/nd/rd/th cases) — use humanize.Ordinal instead",
			"Replace with humanize.Ordinal(n) which handles every numeric case (1st, 11th, 21st, 101st, ...) correctly.",
			line,
			col,
			filePath,
			finding.ConfidenceHigh,
		),
	}
}
