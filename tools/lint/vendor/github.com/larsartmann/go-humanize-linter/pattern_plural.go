package humanizelint

import (
	"go/ast"
	"go/token"
	"strings"
)

// ---------------------------------------------------------------------------
// Pluralization pattern helpers (H004)
// ---------------------------------------------------------------------------

// hasEqualsOneBranch reports whether fn contains an `if x == 1` or `if x != 1`
// conditional where x is a simple identifier (not a function call like len(x)),
// AND the branch body contains at least one string literal or string-producing
// statement (return/assignment with a string expression). The string-in-branch
// requirement eliminates false positives like `if n == 1 { return sorted[0] }`
// (percentile calculation), exit-code extraction, and code generation — none of
// which produce different strings based on the count.
func hasEqualsOneBranch(fn *ast.FuncDecl) bool {
	hit := false

	ast.Inspect(fn, func(n ast.Node) bool {
		if hit {
			return false
		}

		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		be, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		if be.Op != token.EQL && be.Op != token.NEQ {
			return true
		}

		var otherSide ast.Expr

		if isLiteralInt(be.Y, 1) {
			otherSide = be.X
		} else if isLiteralInt(be.X, 1) {
			otherSide = be.Y
		} else {
			return true
		}

		if _, ok := otherSide.(*ast.Ident); !ok {
			return true
		}

		if branchContainsString(ifStmt.Body) {
			hit = true
		}

		return true
	})

	return hit
}

// branchContainsString reports whether a block statement contains any string
// literal or string concatenation (BinaryExpr with +). This confirms that an
// `if x == 1` branch is actually producing different text output.
func branchContainsString(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}

	found := false

	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}

		switch e := n.(type) {
		case *ast.BasicLit:
			if _, ok := unquoteString(e); ok {
				found = true
			}

		case *ast.BinaryExpr:
			if e.Op == token.ADD {
				if exprIsStringy(e.X) || exprIsStringy(e.Y) {
					found = true
				}
			}
		}

		return true
	})

	return found
}

// exprIsStringy is a lightweight heuristic: reports whether expr is a basic
// string literal or an identifier (which might be a string variable).
func exprIsStringy(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		_, ok := unquoteString(e)

		return ok
	case *ast.Ident:
		return true
	}

	return false
}

// hasPluralNamedParams reports whether fn has parameters whose names include
// "singular" and "plural".
func hasPluralNamedParams(fn *ast.FuncDecl) bool {
	var hasSingular, hasPlural bool

	if fn.Type != nil && fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				lower := strings.ToLower(name.Name)
				if strings.Contains(lower, "singular") {
					hasSingular = true
				}

				if strings.Contains(lower, "plural") {
					hasPlural = true
				}
			}
		}
	}

	return hasSingular && hasPlural
}
