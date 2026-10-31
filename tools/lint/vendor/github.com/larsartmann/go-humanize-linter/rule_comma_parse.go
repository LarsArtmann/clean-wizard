package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleParseComma (H010) detects manual comma-grouped number parsing
// ("1,234,567" → int64/float64) that should use humanize.ParseComma /
// humanize.ParseCommaf (go-humanize v1.1.0+).
//
// Triggers when a function contains BOTH:
//   - a comma-stripping signal: strings.ReplaceAll(x, ",", ""),
//     strings.Replace(x, ",", "", n), strings.Split(x, ",") + strings.Join(y, ""),
//     or a rune loop filtering ',' while rebuilding the string
//   - a corroborating strconv number parse: Atoi, ParseInt, ParseUint,
//     ParseFloat, or ParseComplex
//
// Either signal alone matches thousands of innocent functions (CSV splitting,
// display cleanup, number validation); together they match hand-rolled
// reimplementations of humanize.ParseComma/ParseCommaf.
func RuleParseComma() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH010,
			Name:        "manual-comma-parse",
			Description: "Manual comma-grouped number parsing — use humanize.ParseComma / humanize.ParseCommaf instead of comma-stripping before strconv parsing",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			funcFindings, err := checkFuncDecls(dir, RuleIDH010, detectParseComma)
			if err != nil {
				return nil, err
			}

			return funcFindings, nil
		},
	}
}

func detectParseComma(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	aliases := buildImportAliases(file)
	evidence := collectCommaParseEvidence(fn, aliases)

	if !evidence.parsesNumber {
		return nil
	}

	var confidence finding.Confidence

	var form string

	switch {
	case evidence.nestedInParse:
		confidence = finding.ConfidenceFull
		form = "strip inlined in strconv call"
	case evidence.stripReplace:
		confidence = finding.ConfidenceHigh
		form = "comma strip + strconv parse"
	case evidence.splitJoin || evidence.runeFilterLoop:
		confidence = finding.ConfidenceMedium
		form = "comma strip (indirect) + strconv parse"
	default:
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH010,
			fmt.Sprintf(
				"manual comma-grouped number parsing (%s) — use humanize.ParseComma / humanize.ParseCommaf instead",
				form,
			),
			"Replace with humanize.ParseComma(s) or humanize.ParseCommaf(s) "+
				"(go-humanize v1.1.0+), which strip thousands separators and parse in one call.",
			line, col, filePath, confidence,
		),
	}
}
