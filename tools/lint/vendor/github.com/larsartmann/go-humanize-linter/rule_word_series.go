package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleWordSeries (H012) detects hand-rolled word-series joining
// ("a, b, and c") that should use english.WordSeries / english.OxfordWordSeries
// (github.com/dustin/go-humanize/english, v1.0.1+).
//
// Triggers when a function contains a strings.Join with a comma separator
// AND a conjunction string literal (" and ", ", and ", " or ", ", or ").
// Confidence rises with more of the full clone present:
//
//   - Full: prefix join (Join(x[:len(x)-1], ", ")) + conjunction + last-element
//     access (x[len(x)-1])
//   - High: comma join + conjunction + len(x)==1/len(x)==2 special-casing
//   - Medium: comma join + conjunction only
//
// A plain strings.Join(x, ", ") without a conjunction literal stays clean —
// that is a simple list join, not a word series.
func RuleWordSeries() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH012,
			Name:        "manual-word-series",
			Description: "Manual word-series joining — use english.WordSeries / english.OxfordWordSeries instead of hand-rolled 'a, b, and c' assembly",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			funcFindings, err := checkFuncDecls(dir, RuleIDH012, detectWordSeries)
			if err != nil {
				return nil, err
			}

			return funcFindings, nil
		},
	}
}

func detectWordSeries(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	evidence := collectWordSeriesEvidence(fn, buildImportAliases(file))

	if !evidence.anyCommaJoin || !evidence.conjunctionLiteral {
		return nil
	}

	var (
		confidence finding.Confidence
		form       string
	)

	switch {
	case evidence.prefixJoin && evidence.lastElementIndex:
		confidence = finding.ConfidenceFull
		form = "prefix join + conjunction + last element"
	case evidence.lengthBranches:
		confidence = finding.ConfidenceHigh
		form = "comma join + conjunction + length branches"
	default:
		confidence = finding.ConfidenceMedium
		form = "comma join + conjunction"
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH012,
			fmt.Sprintf(
				"manual word-series joining (%s) — use english.WordSeries instead",
				form,
			),
			"Replace with english.WordSeries(words, \"and\") — or english.OxfordWordSeries "+
				"for the Oxford comma — which handles 0/1/2-element edge cases.",
			line, col, filePath, confidence,
		),
	}
}
