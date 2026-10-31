// rule_commaf.go — H009 Commaf detector (humanize.Commaf replacement).

package humanizelint

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleCommaf (H009) detects manual float-with-thousands-separator
// formatting that should use humanize.Commaf.
//
// Triggers when a function contains BOTH a "%.Nf" fmt.Sprintf call AND
// manual comma/separator writing. The combination is a near-certain
// reimplementation of humanize.Commaf.
func RuleCommaf() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH009,
			Name:        "manual-commaf",
			Description: "Manual float-with-comma formatting (%.Nf + manual group separator) — use humanize.Commaf instead",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH009, detectCommaf)
		},
	}
}

func detectCommaf(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	if !hasCommafPattern(fn, buildImportAliases(file)) {
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH009,
			"manual float-with-comma formatting (%.Nf + manual separator) — use humanize.Commaf instead",
			"Replace with humanize.Commaf(f) — or humanize.CommafWithDigits(f, n) for custom "+
				"precision — which produces e.g. 1,234.56 directly.",
			line, col, filePath, finding.ConfidenceMedium,
		),
	}
}
