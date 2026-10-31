// pattern_commaf.go — H009 manual-commaf detection.
//
// Commaf detection: float formatting + manual separator grouping. Triggered
// when a function calls fmt.Sprintf("%.Nf", x) and writes the result to a
// string with manual comma/separator insertion, OR uses strconv.FormatFloat
// with a fixed positive precision inside a comma-grouping loop.

package humanizelint

import (
	"go/ast"
	"go/token"
	"strings"
)

// Format-string parsing helpers. Kept as named constants so the format-string
// walker in hasPercentNF stays readable and so the magic-number linter stops
// flagging index arithmetic as "noise".
const (
	percentChar        = '%'
	percentDot         = '.'
	digitZero          = '0'
	digitNine          = '9'
	floatVerb          = 'f'
	minFormatWalkRange = 2 // smallest "%<verb>" worth scanning ahead
	minPercentWidth    = 3 // start index for the %<digit>f scan (after % and digit)
	dotPrecisionOffset = 2 // start index for the %.<digit>f scan (after % and .)
)

// hasCommafPattern reports whether fn contains the "%.Nf + manual comma
// grouping" anti-pattern. The signature is:
//
//  1. A call to fmt.Sprintf with a "%.Nf" format string (or
//     strconv.FormatFloat with a non-zero precision), AND
//  2. A manual separator group loop in the same function.

func hasCommafPattern(fn *ast.FuncDecl, aliases map[string]string) bool {
	hasFloatFormat := false
	hasSeparatorLoop := false

	ast.Inspect(fn, func(n ast.Node) bool {
		if hasFloatFormat && hasSeparatorLoop {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isPackageCall(call, "fmt", "Sprintf", aliases) && hasPercentNF(call) {
			hasFloatFormat = true
		}

		if isPackageCall(call, "strconv", "FormatFloat", aliases) && hasFormatFloatPrecision(call) {
			hasFloatFormat = true
		}

		if isCommaSeparatorCall(call, aliases) {
			hasSeparatorLoop = true
		}

		return true
	})

	return hasFloatFormat && hasSeparatorLoop
}

// hasFormatFloatPrecision reports whether a strconv.FormatFloat call is
// invoked with a fixed positive precision — the same shape that
// humanize.Commaf produces (e.g. "%.2f" equivalent). Literal verbs other
// than 'f' ('g', 'e', 'b', 'x') never render comma-groupable decimal
// strings, mirroring walkFormatFloatVerbs' rejection of %.Ng; a non-literal
// verb is inconclusive and defers to the precision check. Negative
// precision (-1 = shortest round-trip form, the CSV/JSON writer's choice)
// is excluded: source-level "-1" parses as ast.UnaryExpr, and the sign must
// be resolved before literal unwrapping — getBasicLit strips the minus,
// which turned the old "-1" exclusion into dead code and tripped H009 on
// CSV writers (real-world false positive, seen in go-health-dashboard).
func hasFormatFloatPrecision(call *ast.CallExpr) bool {
	if len(call.Args) < minFormatWalkRange {
		return false
	}

	if !hasFixedFloatVerb(call) {
		return false
	}

	digits, negative, ok := signedIntLiteral(call.Args[2])
	if !ok || negative {
		return false
	}

	return digits != "0"
}

// hasFixedFloatVerb reports whether the call's format verb (strconv
// FormatFloat arg 1) is a literal that can produce fixed-decimal output.
// Only 'f' can; 'g'/'e' render scientific notation and 'b'/'x' binary/hex.
// A non-literal verb (variable, named constant) is inconclusive and returns
// true so the precision check decides.
func hasFixedFloatVerb(call *ast.CallExpr) bool {
	if len(call.Args) < minFormatWalkRange {
		return true
	}

	lit := getBasicLit(call.Args[1])

	if lit == nil {
		return true
	}

	return lit.Value == "'f'"
}

// signedIntLiteral resolves expr to its integer digits and sign. Handles
// parser-shaped negation (ast.UnaryExpr over ast.BasicLit, optionally
// parenthesized) and, defensively, a leading minus inside the literal
// value itself. ok is false for anything else (variables, strings, ...).
func signedIntLiteral(expr ast.Expr) (string, bool, bool) {
	negative := false

	for {
		switch e := expr.(type) {
		case *ast.ParenExpr:
			expr = e.X
		case *ast.UnaryExpr:
			//nolint:exhaustive // only + and - are sign-relevant unary operators
			switch e.Op {
			case token.SUB:
				negative = !negative
			case token.ADD:
			default:
				return "", false, false
			}

			expr = e.X
		case *ast.BasicLit:
			if e.Kind != token.INT {
				return "", false, false
			}

			value := e.Value

			if strings.HasPrefix(value, "-") {
				negative = !negative
				value = strings.TrimPrefix(value, "-")
			}

			return value, negative, true
		default:
			return "", false, false
		}
	}
}

// hasPercentNF reports whether a fmt.Sprintf call uses a "%.Nf" format
// (e.g. "%.1f", "%.2f") — a non-zero precision float.
func hasPercentNF(call *ast.CallExpr) bool {
	if len(call.Args) < 1 {
		return false
	}

	lit := getBasicLit(call.Args[0])
	if lit == nil || lit.Kind != token.STRING {
		return false
	}

	val, ok := unquoteString(lit)
	if !ok {
		return false
	}

	return walkFormatFloatVerbs(val)
}

// walkFormatFloatVerbs scans val for any "%.<digit>f" or "%<digit>f" verb and
// returns true on the first match. Pure string walking — no AST access, so
// easy to unit-test in isolation.
func walkFormatFloatVerbs(val string) bool {
	for i := 0; i+minFormatWalkRange < len(val); i++ {
		if val[i] != percentChar {
			continue
		}

		switch {
		case val[i+1] == percentDot:
			if !scanDottedPercentFloat(val, i) {
				continue
			}

			return true
		case val[i+1] >= digitZero && val[i+1] <= digitNine:
			if !scanBarePercentFloat(val, i) {
				continue
			}

			return true
		}
	}

	return false
}

// scanDottedPercentFloat returns true when val[i:] starts with "%.<digit>f"
// (a "%." + digit + "f" verb, optionally with width). i must point at the
// '%' character.
func scanDottedPercentFloat(val string, i int) bool {
	if i+dotPrecisionOffset >= len(val) {
		return false
	}

	digit := val[i+dotPrecisionOffset]
	if digit < digitZero || digit > digitNine {
		return false
	}

	for j := i + minPercentWidth; j < len(val); j++ {
		c := val[j]
		if c == floatVerb {
			return true
		}

		if c < digitZero || c > digitNine {
			return false
		}
	}

	return false
}

// scanBarePercentFloat returns true when val[i:] starts with "%<digit>f"
// (a "%" + digit + "f" verb, optionally with width). i must point at the
// '%' character.
func scanBarePercentFloat(val string, i int) bool {
	digit := val[i+1]
	if digit < digitZero || digit > digitNine {
		return false
	}

	for j := i + minFormatWalkRange; j < len(val); j++ {
		c := val[j]
		if c == floatVerb {
			return true
		}

		if c < digitZero || c > digitNine {
			return false
		}
	}

	return false
}
