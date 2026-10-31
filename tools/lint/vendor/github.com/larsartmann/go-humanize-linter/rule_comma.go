package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleComma (H002) detects manual comma (thousands separator) insertion that
// should use humanize.Comma or humanize.Commaf.
//
// Triggers via two detection paths:
//
//  1. Strong: modulo-3 or step-by-3 grouping + comma writing (high/full confidence)
//  2. Fallback: for-loop + comma writing + digit conversion (medium confidence)
//     — catches cases where the step size is a named constant like digitsPerGroup
func RuleComma() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH002,
			Name:        "manual-comma-format",
			Description: "Manual comma/thousands-separator insertion — use humanize.Comma or humanize.Commaf instead of looping over digits",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH002, detectCommaFormat)
		},
	}
}

func detectCommaFormat(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	aliases := buildImportAliases(file)

	// Suppress H002 when H009 (commaf) would fire on the same function.
	// H009 is the more specific diagnosis: float formatting + comma grouping.
	// Without this, a "%.Nf" + comma-loop function gets two diagnostics.
	if hasCommafPattern(fn, aliases) {
		return nil
	}

	mod3 := hasModulo3(fn)
	step3 := hasStepBy3(fn)
	sep := hasCommaOrSeparator(fn, aliases)

	var (
		confidence finding.Confidence
		signals    string
	)

	switch {
	case (mod3 || step3) && sep:
		if mod3 && step3 {
			confidence = finding.ConfidenceFull
		} else {
			confidence = finding.ConfidenceHigh
		}

		signals = fmt.Sprintf("mod3=%v, step3=%v", mod3, step3)

	case hasForLoop(fn) && sep && hasDigitConversion(fn, aliases):
		// Fallback: for-loop + comma writing + digit conversion.
		// Catches cases using named constants like `digitsPerGroup`.
		confidence = finding.ConfidenceMedium
		signals = "for-loop + comma + digit-conversion (no literal 3 detected)"

	default:
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	suggestion := "Replace with humanize.Comma(int64(n)) for integers or humanize.Commaf(f) for floats."

	if mentionsBigInt(fn) {
		suggestion += " For *big.Int values use humanize.BigComma instead."
	}

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH002,
			fmt.Sprintf("manual comma formatting (%s) — use humanize.Comma instead", signals),
			suggestion,
			line, col, filePath, confidence,
		),
	}
}
