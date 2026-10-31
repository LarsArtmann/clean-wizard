package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// RuleParseBytes (H007) detects manual byte-size string parsing ("10MB" →
// bytes) that should use humanize.ParseBytes.
//
// Triggers when a function contains:
//   - 2+ strings.HasSuffix/CutSuffix/TrimSuffix calls checking byte-unit suffixes
//   - OR a map[string]int64 literal with 2+ byte-unit keys used as multipliers
//
// Also detects package-level var declarations with byte-unit multiplier maps.
func RuleParseBytes() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH007,
			Name:        "manual-parse-bytes",
			Description: "Manual byte-size string parsing — use humanize.ParseBytes instead of suffix matching and multiplication",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			funcFindings, err := checkFuncDecls(dir, RuleIDH007, detectParseBytes)
			if err != nil {
				return nil, err
			}

			files, err := WalkGoDir(dir)
			if err != nil {
				return nil, &WalkError{Dir: dir, Err: err}
			}

			var varFindings []finding.Finding
			for _, pf := range files {
				varFindings = append(varFindings,
					scanFileLevelByteMultiplierMaps(pf.Fset, pf.File, pf.Path)...)
			}

			return append(funcFindings, varFindings...), nil
		},
	}
}

func detectParseBytes(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	aliases := buildImportAliases(file)
	suffixChecks := hasByteUnitSuffixChecks(fn, aliases)
	multMap := hasByteUnitMultiplierMap(fn)

	if suffixChecks < 2 && !multMap {
		return nil
	}

	var confidence finding.Confidence

	switch {
	case suffixChecks >= 3 || multMap:
		confidence = finding.ConfidenceHigh
	case suffixChecks >= 2:
		confidence = finding.ConfidenceMedium
	default:
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH007,
			fmt.Sprintf(
				"manual byte-size parsing (suffixChecks=%d, multMap=%v) — use humanize.ParseBytes instead",
				suffixChecks, multMap,
			),
			"Replace with humanize.ParseBytes(s) which handles all SI and IEC suffixes.",
			line, col, filePath, confidence,
		),
	}
}
