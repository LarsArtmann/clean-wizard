package humanizelint

import (
	"context"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleFtoa (H006) detects manual float-to-string conversion with trailing-zero
// stripping that should use humanize.Ftoa.
//
// Triggers when a function contains:
//
//	strings.TrimRight(strings.TrimRight(x, "0"), ".")
//
// This nested-TrimRight pattern is the textbook reimplementation of
// humanize.Ftoa's trailing-zero removal.
func RuleFtoa() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH006,
			Name:        "manual-ftoa",
			Description: "Manual float formatting with trailing-zero stripping — use humanize.Ftoa instead of nested strings.TrimRight",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH006, detectFtoa)
		},
	}
}

func detectFtoa(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	if !hasNestedTrimRight(fn, buildImportAliases(file)) {
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH006,
			"manual float trailing-zero stripping (nested TrimRight) — use humanize.Ftoa instead",
			"Replace with humanize.Ftoa(f) or humanize.FtoaWithDigits(f, 1).",
			line, col, filePath, finding.ConfidenceFull,
		),
	}
}
