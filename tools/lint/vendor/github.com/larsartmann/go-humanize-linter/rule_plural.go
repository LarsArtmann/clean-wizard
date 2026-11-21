package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RulePlural (H004) detects manual English pluralization that should use
// github.com/dustin/go-humanize/english.Plural or english.PluralWord.
//
// Triggers when a function:
//   - Has parameters named "singular" and "plural" (explicit reimplementation)
//   - OR contains an `if x == 1` / `if x != 1` conditional where the branches
//     return different string values (the classic plural switch)
//
// A function whose body already delegates to english.Plural or
// english.PluralWord is a legitimate wrapper (zero-suppression, type
// conversion at the boundary), not a reimplementation — the signature-only
// heuristic must not flag it.
func RulePlural() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH004,
			Name:        "manual-plural",
			Description: "Manual pluralization — use github.com/dustin/go-humanize/english.Plural or english.PluralWord instead of if-n==1 switches",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityError,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH004, detectPlural)
		},
	}
}

func detectPlural(fset *token.FileSet, _ *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	namedParams := hasPluralNamedParams(fn)
	equalsOne := hasEqualsOneBranch(fn)

	// For the equalsOne path, require the function to return a string type.
	// Without this, every `if result != 1 { return fmt.Errorf(...) }` in SQL
	// health checks, validation, and code generation triggers a false positive.
	if equalsOne && !namedParams && !funcReturnsString(fn) {
		equalsOne = false
	}

	if namedParams && !equalsOne && delegatesToEnglishPlural(fn) {
		return nil
	}

	if !namedParams && !equalsOne {
		return nil
	}

	confidence := finding.ConfidenceHigh
	if namedParams && equalsOne {
		confidence = finding.ConfidenceFull
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH004,
			fmt.Sprintf(
				"manual pluralization (namedParams=%v, equalsOne=%v) — use github.com/dustin/go-humanize/english instead",
				namedParams,
				equalsOne,
			),
			"Replace with english.PluralWord(n, singular, plural) for just the word, or english.Plural(n, singular, plural) for \"count word\".",
			line,
			col,
			filePath,
			confidence,
		),
	}
}

// delegatesToEnglishPlural reports whether the function body already calls
// english.Plural or english.PluralWord: such a function delegates instead of
// reimplementing, so the signature-only heuristic must stay silent.
func delegatesToEnglishPlural(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	delegates := false

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if pkg.Name == "english" && (sel.Sel.Name == "Plural" || sel.Sel.Name == "PluralWord") {
			delegates = true

			return false
		}

		return true
	})

	return delegates
}
