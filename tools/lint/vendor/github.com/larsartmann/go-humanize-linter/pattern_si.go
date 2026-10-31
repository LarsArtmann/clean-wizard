package humanizelint

import (
	"go/ast"
	"go/token"
	"strings"
)

// ---------------------------------------------------------------------------
// SI / K-M pattern helpers (H005)
// ---------------------------------------------------------------------------

// hasDivisionBy1000 reports whether fn contains a division by the literal
// 1000, 1000000, 1000000000, or the scientific-notation equivalents
// (1e3, 1e6, 1e9, 1e12).
func hasDivisionBy1000(fn *ast.FuncDecl) bool {
	powersOf1000 := map[string]bool{
		"1000": true, "1000000": true, "1000000000": true, "1000000000000": true,
		"1e3": true, "1e6": true, "1e9": true, "1e12": true,
	}

	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		if be.Op == token.QUO {
			if lit, ok := be.Y.(*ast.BasicLit); ok && powersOf1000[normLit(lit.Value)] {
				hit = true
			}
		}

		return true
	})

	return hit
}

// hasKMSuffix reports whether fn contains SI-prefix suffixes — either as
// standalone string literals ("K", "M") or embedded in format strings
// ("%.2fK", "%.1fM"). Excludes byte-unit strings to avoid overlap with H001.
func hasKMSuffix(fn *ast.FuncDecl) bool {
	for _, s := range allStringLiterals(fn) {
		// Standalone single-letter SI prefixes.
		switch s {
		case "K", "M", "G", "T", "k", "m":
			return true
		}

		// Embedded suffixes like "%.2fK" or "%.1fM" — a FORMAT string whose
		// last char is an SI prefix and which is NOT a byte-unit pattern
		// (e.g., "%.1f MB"). Requiring a % verb excludes prose that merely
		// ends in an SI-prefix letter, such as a "THROUGHPUT" table header
		// paired with a /1e3 millisecond conversion (real-world corpus FP).
		if len(s) >= 2 && strings.Contains(s, "%") && !byteUnitRegex.MatchString(s) {
			switch s[len(s)-1] {
			case 'K', 'M', 'G', 'T', 'k':
				return true
			}
		}
	}

	return false
}
