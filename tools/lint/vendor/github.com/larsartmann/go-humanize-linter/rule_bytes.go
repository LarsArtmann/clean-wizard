package humanizelint

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// minUnitCountStrongSignal is the minimum number of distinct byte-unit strings
// that constitutes a strong signal even without explicit division by 1024.
const minUnitCountStrongSignal = 3

// RuleBytes (H001) detects manual byte-size formatting that should use
// humanize.Bytes or humanize.IBytes.
//
// Triggers when a function contains any of these signal clusters:
//
//   - "KMGTPE"[exp] index expression (the classic prefix trick)
//   - A []string literal with 2+ byte-unit strings (KB, MB, KiB, MiB, etc.)
//   - 2+ distinct byte-unit strings in format strings + division by a power of 1024
//   - 3+ distinct byte-unit strings (very strong signal even without explicit division)
func RuleBytes() linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          RuleIDH001,
			Name:        "manual-bytes-format",
			Description: "Manual byte-size formatting — use humanize.Bytes or humanize.IBytes instead of dividing by 1024 and formatting unit strings",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return checkFuncDecls(dir, RuleIDH001, detectBytesFormat)
		},
	}
}

//nolint:gohumanize // suggestion text contains byte-unit strings (KB/MB/KiB) that self-trigger the detector
func detectBytesFormat(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding {
	kmgtp := hasKMGTPEIndex(fn)
	unitSlice := hasByteUnitSlice(fn)
	units := countByteUnits(fn)
	div1024 := hasDivisionByPowerOf1024(fn) || hasConst1024(file, fn)
	switchOnly := hasSwitchStatement(fn) && !div1024
	unitCount := len(units)

	// Size-bucket lookup tables: a switch or []string of byte-unit labels
	// without any division by 1024 is almost never a byte-size formatter.
	if switchOnly || (unitSlice && !div1024) {
		return nil
	}

	msg, confidence, found := bytesFindingResult(kmgtp, unitSlice, unitCount, div1024)
	if !found {
		return nil
	}

	line, col := posOf(fset, fn.Pos())

	return []finding.Finding{
		makeFindingWithConfidence(
			RuleIDH001,
			msg,
			"Replace with humanize.Bytes(uint64(n)) for SI (KB/MB) or humanize.IBytes(uint64(n)) for IEC (KiB/MiB).",
			line,
			col,
			filePath,
			confidence,
		),
	}
}

// bytesFindingResult maps detection signals to a human-readable message and
// confidence level. Returns found=false when no pattern is detected.
func bytesFindingResult(
	kmgtp, unitSlice bool,
	unitCount int,
	div1024 bool,
) (string, finding.Confidence, bool) {
	switch {
	case kmgtp:
		return "manual byte-size formatting (KMGTPE index trick) — use humanize.Bytes or humanize.IBytes instead",
			finding.ConfidenceFull, true
	case unitSlice:
		conf := finding.ConfidenceFull
		if !div1024 {
			// A unit slice without division-by-1024 is often a bucket lookup
			// table (e.g. []string{"B","KB","MB","GB"}) rather than a
			// formatter. Keep the finding but lower the confidence.
			conf = finding.ConfidenceMedium
		}

		return fmt.Sprintf(
				"manual byte-size formatting (unit string slice, %d unit strings) — use humanize.Bytes or humanize.IBytes instead",
				unitCount,
			),
			conf, true
	case unitCount >= minUnitCountStrongSignal:
		return fmt.Sprintf(
				"manual byte-size formatting (%d unit strings, div1024=%v) — use humanize.Bytes or humanize.IBytes instead",
				unitCount,
				div1024,
			),
			finding.ConfidenceHigh, true
	case unitCount >= 2 && div1024:
		return fmt.Sprintf(
				"manual byte-size formatting (%d unit strings, div1024=%v) — use humanize.Bytes or humanize.IBytes instead",
				unitCount,
				div1024,
			),
			finding.ConfidenceHigh, true
	case unitCount >= 2:
		return fmt.Sprintf(
				"manual byte-size formatting (%d unit strings) — use humanize.Bytes or humanize.IBytes instead",
				unitCount,
			),
			finding.ConfidenceMedium, true
	default:
		return "", finding.ConfidenceLow, false
	}
}
