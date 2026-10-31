package humanizelint

import (
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
)

// ---------------------------------------------------------------------------
// ParseBytes pattern helpers (H007)
// ---------------------------------------------------------------------------

// hasByteUnitSuffixChecks reports whether fn contains 2+ calls to
// strings.HasSuffix, strings.CutSuffix, or strings.TrimSuffix where the suffix
// argument is a known byte-unit string (KB, MB, GB, KiB, MiB, etc.).
func hasByteUnitSuffixChecks(fn *ast.FuncDecl, aliases map[string]string) int {
	count := 0

	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !isPackageCall(call, "strings", "HasSuffix", aliases) &&
			!isPackageCall(call, "strings", "CutSuffix", aliases) &&
			!isPackageCall(call, "strings", "TrimSuffix", aliases) {
			return true
		}

		if len(call.Args) < 2 {
			return true
		}

		if val, ok := unquoteString(getBasicLit(call.Args[1])); ok {
			if byteUnitSet[val] {
				count++
			}
		}

		return true
	})

	return count
}

// countByteUnitKeys counts how many string-literal keys in a composite literal
// are recognised byte-unit names (KB, MB, GiB, etc.).
func countByteUnitKeys(cl *ast.CompositeLit) int {
	count := 0

	for _, el := range cl.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		if val, ok := unquoteString(getBasicLit(kv.Key)); ok {
			if byteUnitSet[val] {
				count++
			}
		}
	}

	return count
}

// hasByteUnitMultiplierMap reports whether fn contains a map literal with
// string keys that are byte-unit names (KB, MB, GiB, etc.) and integer values.
func hasByteUnitMultiplierMap(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if countByteUnitKeys(cl) >= 2 {
			hit = true
		}

		return true
	})

	return hit
}

// isByteUnitMultiplierMapLiteral reports whether expr is a map literal with
// integer values and 2+ byte-unit keys. The type check (map[string]int*) is
// essential to avoid false positives on lookup sets like map[string]bool.
func isByteUnitMultiplierMapLiteral(expr ast.Expr) bool {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}

	mapType, ok := cl.Type.(*ast.MapType)
	if !ok {
		return false
	}

	keyIdent, ok := mapType.Key.(*ast.Ident)
	if !ok || keyIdent.Name != stringTypeName {
		return false
	}

	valIdent, ok := mapType.Value.(*ast.Ident)
	if !ok {
		return false
	}

	switch valIdent.Name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		// Numeric multiplier — proceed
	default:
		return false
	}

	return countByteUnitKeys(cl) >= 2
}

// scanFileLevelByteMultiplierMaps scans file-level var declarations for
// byte-unit multiplier maps that should use humanize.ParseBytes. Returns one
// finding per matching var spec.
func scanFileLevelByteMultiplierMaps(
	fset *token.FileSet,
	file *ast.File,
	filePath string,
) []finding.Finding {
	var findings []finding.Finding

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, val := range valueSpec.Values {
				if !isByteUnitMultiplierMapLiteral(val) {
					continue
				}

				line, col := posOf(fset, valueSpec.Pos())

				findings = append(findings, makeFindingWithConfidence(
					RuleIDH007,
					"package-level byte-unit multiplier map — use humanize.ParseBytes instead",
					"Replace with humanize.ParseBytes(s) which handles all SI and IEC suffixes.",
					line, col, filePath, finding.ConfidenceMedium,
				))
			}
		}
	}

	return findings
}
