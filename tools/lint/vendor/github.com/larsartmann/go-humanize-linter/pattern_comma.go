package humanizelint

import (
	"go/ast"
	"go/token"
	"slices"
)

// ---------------------------------------------------------------------------
// Comma-grouping pattern helpers (H002)
// ---------------------------------------------------------------------------

// hasModulo3 reports whether fn contains a `% 3` expression — the classic
// comma-grouping signal.
func hasModulo3(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		if be.Op == token.REM {
			if lit, ok := be.Y.(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "3" {
				hit = true
			}
		}

		return true
	})

	return hit
}

// hasCommaOrSeparator reports whether fn writes a comma or other thousands
// separator character via WriteString, WriteByte, WriteRune, or appears in a
// strings.Join call with ",".
func hasCommaOrSeparator(fn *ast.FuncDecl, aliases map[string]string) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isCommaSeparatorCall(call, aliases) {
			hit = true
		}

		return true
	})

	return hit
}

// isCommaSeparatorCall reports whether a call expression writes a comma or
// thousands separator via strings.Join, WriteString, WriteByte, or WriteRune.
//
// Note: strings.Join with a single space (" ") is treated as a non-signal:
// " " is not a thousands separator, and joining CLI args with a space is a
// common, legitimate pattern that shares none of the heuristic weight of ","
// or ".". Including it here would cause false positives for any function that
// builds a shell-style command line (see regression test).
func isCommaSeparatorCall(call *ast.CallExpr, aliases map[string]string) bool {
	if isPackageCall(call, "strings", "Join", aliases) && len(call.Args) >= 2 {
		if isSeparatorLiteral(call.Args[1], ",", ".") {
			return true
		}
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch sel.Sel.Name {
	case "WriteString":
		return len(call.Args) > 0 && isSeparatorLiteral(call.Args[0], ",", ".")

	case "WriteByte", "WriteRune":
		return isSeparatorRune(call.Args)
	}

	return false
}

// isSeparatorLiteral reports whether expr is a string literal matching any of
// the given separator values.
func isSeparatorLiteral(expr ast.Expr, separators ...string) bool {
	val, ok := unquoteString(getBasicLit(expr))
	if !ok {
		return false
	}

	return slices.Contains(separators, val)
}

// isSeparatorRune reports whether args[0] is a rune literal ',' or '.'.
func isSeparatorRune(args []ast.Expr) bool {
	if len(args) == 0 {
		return false
	}

	lit := getBasicLit(args[0])

	return lit != nil && (lit.Value == "','" || lit.Value == "'.'")
}

// hasStepBy3 reports whether fn contains a for-loop incrementing by 3
// (e.g. i += 3 or i = i + 3).
func hasStepBy3(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}

		if forStmt.Post == nil {
			return true
		}

		ast.Inspect(forStmt.Post, func(nn ast.Node) bool {
			if be, ok := nn.(*ast.BinaryExpr); ok {
				if be.Op == token.ADD {
					if lit, ok := be.Y.(*ast.BasicLit); ok && lit.Value == "3" {
						hit = true
					}
				}
			}

			if as, ok := nn.(*ast.AssignStmt); ok && as.Tok == token.ADD_ASSIGN {
				for _, v := range as.Rhs {
					if lit, ok := v.(*ast.BasicLit); ok && lit.Value == "3" {
						hit = true
					}
				}
			}

			return true
		})

		return true
	})

	return hit
}

// hasForLoop reports whether fn contains any for-loop.
func hasForLoop(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		if _, ok := n.(*ast.ForStmt); ok {
			hit = true
		}

		return true
	})

	return hit
}

// hasDigitConversion reports whether fn calls strconv.Itoa, strconv.FormatInt,
// strconv.FormatUint, or strconv.FormatFloat — the integer-to-string
// conversions that comma formatters always use.
func hasDigitConversion(fn *ast.FuncDecl, aliases map[string]string) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isPackageCall(call, "strconv", "Itoa", aliases) ||
			isPackageCall(call, "strconv", "FormatInt", aliases) ||
			isPackageCall(call, "strconv", "FormatUint", aliases) ||
			isPackageCall(call, "strconv", "FormatFloat", aliases) {
			hit = true
		}

		return true
	})

	return hit
}

// mentionsBigInt reports whether fn references math/big integers — a
// *big.Int parameter, big.NewInt/new(big.Int) call, or a big.* selector —
// used to point H002's suggestion at humanize.BigComma.
func mentionsBigInt(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "big" {
				hit = true
			}
		}

		return true
	})

	return hit
}
