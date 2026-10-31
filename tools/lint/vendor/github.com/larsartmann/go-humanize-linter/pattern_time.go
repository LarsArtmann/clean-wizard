package humanizelint

import (
	"go/ast"
	"strings"
)

// ---------------------------------------------------------------------------
// Relative-time pattern helpers (H003)
// ---------------------------------------------------------------------------

// hasTimeSinceOrSub reports whether fn calls time.Since or a .Sub method on a
// time value.
func hasTimeSinceOrSub(fn *ast.FuncDecl, aliases map[string]string) bool {
	hit := false

	ast.Inspect(fn, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// time.Since(x)
		if isPackageCall(call, "time", "Since", aliases) {
			hit = true

			return true
		}

		// x.Sub(y)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Sub" {
				hit = true
			}
		}

		return true
	})

	return hit
}

// timePackagePath is the canonical import path for the time package.
const timePackagePath = "time"

// isDotImportOf reports whether the aliases map contains a dot import for the
// given canonical package path.
func isDotImportOf(aliases map[string]string, path string) bool {
	if aliases == nil {
		return false
	}

	resolved, ok := aliases["."]
	if !ok {
		return false
	}

	return resolved == path || strings.HasSuffix(resolved, "/"+path)
}

// isTimeDurationName reports whether name is one of the time-package duration
// constants that typically appear in relative-time threshold comparisons.
func isTimeDurationName(name string) bool {
	switch name {
	case "Second", "Minute", "Hour", "Day", "Week", "Month", "Year":
		return true
	default:
		return false
	}
}

// isTimeDurationSelector reports whether node is a selector expression that
// resolves to a time-package duration constant (e.g. time.Hour, tm.Hour).
func isTimeDurationSelector(node ast.Node, aliases map[string]string) bool {
	sel, ok := node.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !isTimeDurationName(sel.Sel.Name) {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Name == timePackagePath {
		return true
	}

	if aliases == nil {
		return false
	}

	resolved, ok := aliases[ident.Name]
	if !ok {
		return false
	}

	return resolved == timePackagePath || strings.HasSuffix(resolved, "/"+timePackagePath)
}

// isTimeDurationIdentifier reports whether node is a bare identifier that names a
// time-package duration constant. This only matches when a dot import for the
// time package is active.
func isTimeDurationIdentifier(node ast.Node) bool {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return false
	}

	return isTimeDurationName(ident.Name)
}

// hasTimeThresholdComparison reports whether fn compares an expression against
// a time-package duration constant (time.Minute, time.Hour, time.Second, etc.)
// or a multiplication of one (e.g. 24*time.Hour). Supports named aliases
// (tm "time" → tm.Hour) and dot imports (. "time" → bare Hour).
func hasTimeThresholdComparison(fn *ast.FuncDecl, aliases map[string]string) bool {
	dotTime := aliases != nil && isDotImportOf(aliases, timePackagePath)

	hit := false

	ast.Inspect(fn, func(node ast.Node) bool {
		if isTimeDurationSelector(node, aliases) {
			hit = true

			return false
		}

		if dotTime && isTimeDurationIdentifier(node) {
			hit = true

			return false
		}

		return true
	})

	return hit
}
