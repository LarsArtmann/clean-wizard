package humanizelint

import (
	"go/ast"
	"go/token"
	"regexp"
	"slices"
	"strings"
)

// ---------------------------------------------------------------------------
// Byte-unit pattern helpers (H001)
// ---------------------------------------------------------------------------

// byteUnitSet contains known byte-size unit suffixes (case-insensitive). The
// detection logic counts distinct units found in string literals within a
// function.
var byteUnitSet = map[string]bool{ //nolint:gochecknoglobals // package-level lookup table
	"KiB": true, "MiB": true, "GiB": true, "TiB": true, "PiB": true, "EiB": true,
	"KB": true, "MB": true, "GB": true, "TB": true, "PB": true, "EB": true,
	"kB": true,
}

// siPrefixStrings are the exact string literals used as byte-prefix arrays.
// "KMGTPE" is the canonical 6-character prefix string; "kMGTPE" is the
// lowercase-k variant.
var siPrefixStrings = map[string]bool{ //nolint:gochecknoglobals // package-level lookup table
	"KMGTPE": true, "kMGTPE": true,
	"KMGTPEZY": true, "kMGTPEZY": true,
}

// byteUnitRegex matches known byte-size unit suffixes as whole words.
// The \b boundaries prevent false positives like "MEDIUMBLOB" → "MB".
var byteUnitRegex = regexp.MustCompile(`(?i)\b(KiB|MiB|GiB|TiB|PiB|EiB|KB|MB|GB|TB|PB|EB|kB)\b`)

// countByteUnits scans all string literals in fn and returns the set of
// distinct byte-unit suffixes found (using word-boundary matching to avoid
// false positives like "MEDIUMBLOB" containing "MB").
func countByteUnits(fn *ast.FuncDecl) map[string]bool {
	found := map[string]bool{}

	for _, s := range allStringLiterals(fn) {
		for _, m := range byteUnitRegex.FindAllString(s, -1) {
			found[strings.ToUpper(m)] = true
		}
	}

	return found
}

// hasKMGTPEIndex reports whether fn contains an index expression on a string
// literal like "KMGTPE" — the classic byte-prefix trick.
func hasKMGTPEIndex(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		idx, ok := n.(*ast.IndexExpr)
		if !ok {
			return true
		}

		if val, ok := unquoteString(getBasicLit(idx.X)); ok {
			if siPrefixStrings[val] {
				hit = true
			}
		}

		return true
	})

	return hit
}

// hasByteUnitSlice reports whether fn contains a []string composite literal
// whose elements include 2+ known byte-unit strings.
func hasByteUnitSlice(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		count := 0

		for _, el := range cl.Elts {
			if val, ok := unquoteString(getBasicLit(el)); ok {
				if byteUnitSet[val] {
					count++
				}
			}
		}

		if count >= 2 {
			hit = true
		}

		return true
	})

	return hit
}

// hasSwitchStatement reports whether fn contains any switch statement.
// Used as an H001 false-positive filter: a switch full of byte-unit labels
// with no division by 1024 is usually a size-bucket lookup table, not a
// byte-size formatter.
func hasSwitchStatement(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		if _, ok := n.(*ast.SwitchStmt); ok {
			hit = true
		}

		if _, ok := n.(*ast.TypeSwitchStmt); ok {
			hit = true
		}

		return true
	})

	return hit
}

// hasDivisionByPowerOf1024 reports whether fn contains a binary division whose
// divisor is the literal 1024, 1048576 (1024²), 1073741824 (1024³), or any
// expression multiplying by 1024.
func hasDivisionByPowerOf1024(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		if be.Op == token.QUO && isLiteral1024ish(be.Y) {
			hit = true
		}

		return true
	})

	return hit
}

// isLiteral1024ish reports whether expr is the literal 1024, 1048576,
// 1073741824, or a multiplication chain involving 1024 (e.g. 1024*1024).
func isLiteral1024ish(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch normLit(e.Value) {
		case "1024", "1048576", "1073741824", "1099511627776":
			return true
		}

	case *ast.BinaryExpr:
		if e.Op == token.MUL {
			return isLiteral1024ish(e.X) || isLiteral1024ish(e.Y)
		}
	}

	return false
}

// hasConst1024 reports whether fn or its enclosing file defines a constant
// whose value is 1024 or a power of 1024. This catches both in-function
// constants (`const unit = 1024`) and file-level constants (`const bytesKB = 1024`).
func hasConst1024(file *ast.File, fn *ast.FuncDecl) bool {
	hit := false

	// Check constants inside the function body.
	ast.Inspect(fn, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			return true
		}

		for _, spec := range gd.Specs {
			if specHas1024Value(spec) {
				hit = true
			}
		}

		return true
	})

	// Check file-level constant declarations.
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}

		for _, spec := range gd.Specs {
			if specHas1024Value(spec) {
				hit = true
			}
		}
	}

	return hit
}

// specHas1024Value reports whether a ValueSpec contains a literal 1024-ish value.
func specHas1024Value(spec ast.Spec) bool {
	vs, ok := spec.(*ast.ValueSpec)
	if !ok {
		return false
	}

	return slices.ContainsFunc(vs.Values, isLiteral1024ish)
}
