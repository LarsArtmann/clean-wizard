package humanizelint

import (
	"go/ast"
	"go/token"
)

// ---------------------------------------------------------------------------
// ParseComma pattern helpers (H010)
// ---------------------------------------------------------------------------

// commaParseEvidence holds the corroborating signals H010 collects per
// function. The rule fires only when a comma-stripping signal AND a
// strconv number-parse signal co-occur; see detectParseComma.
type commaParseEvidence struct {
	// stripReplace: strings.ReplaceAll(x, ",", "") or strings.Replace(x, ",", "", n)
	stripReplace bool
	// splitJoin: strings.Split(x, ",") and strings.Join(y, "") both present
	splitJoin bool
	// runeFilterLoop: comparison against ',' combined with a string rebuild
	// (builtin append or a Builder Write*) in the same function
	runeFilterLoop bool
	// nestedInParse: a comma-strip call expression is a direct argument of
	// a strconv number-parse call (single-expression reimplementation)
	nestedInParse bool
	// parsesNumber: strconv.Atoi/ParseInt/ParseUint/ParseFloat/ParseComplex call
	parsesNumber bool
}

// strconvParseFuncs are the strconv entry points whose hand-rolled
// comma-stripping prefixes ParseComma/ParseCommaf replace.
var strconvParseFuncs = []string{ //nolint:gochecknoglobals // package-level lookup table
	"Atoi", "ParseInt", "ParseUint", "ParseFloat", "ParseComplex",
}

// arg counts of the strings.ReplaceAll / strings.Replace calls matched by
// isCommaToEmptyReplace: (s, old, new) and (s, old, new, n).
const (
	replaceAllCommaArgs = 3
	replaceCommaArgs    = 4
)

// isCommaToEmptyReplace reports whether call is
// strings.ReplaceAll(x, ",", "") or strings.Replace(x, ",", "", n).
func isCommaToEmptyReplace(call *ast.CallExpr, aliases map[string]string) bool {
	if isPackageCall(call, "strings", "ReplaceAll", aliases) {
		return isReplaceArgTriple(call, replaceAllCommaArgs)
	}

	if isPackageCall(call, "strings", "Replace", aliases) {
		return isReplaceArgTriple(call, replaceCommaArgs)
	}

	return false
}

// isReplaceArgTriple checks the (old, new) argument pair of a Replace-style
// call: old must unquote to "," and new to "". wantArgs is the minimum
// argument count for the concrete function (3 for ReplaceAll, 4 for Replace).
func isReplaceArgTriple(call *ast.CallExpr, wantArgs int) bool {
	if len(call.Args) < wantArgs {
		return false
	}

	old, ok := unquoteString(getBasicLit(call.Args[1]))
	if !ok || old != "," {
		return false
	}

	replacement, ok := unquoteString(getBasicLit(call.Args[2]))

	return ok && replacement == ""
}

// isStrconvNumberParse reports whether call is one of strconv's number
// parsing functions (Atoi, ParseInt, ParseUint, ParseFloat, ParseComplex).
func isStrconvNumberParse(call *ast.CallExpr, aliases map[string]string) bool {
	for _, name := range strconvParseFuncs {
		if isPackageCall(call, "strconv", name, aliases) {
			return true
		}
	}

	return false
}

// isCommaCharLiteral reports whether expr is the rune literal ','.
func isCommaCharLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)

	return ok && lit.Kind == token.CHAR && lit.Value == "','"
}

// isBuiltinAppend reports whether call is the builtin append(x, y, ...).
func isBuiltinAppend(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)

	return ok && ident.Name == "append"
}

// isBuilderWrite reports whether call is a strings.Builder (or similar)
// WriteString/WriteByte/WriteRune method call — a string-rebuild signal.
func isBuilderWrite(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch sel.Sel.Name {
	case "WriteString", "WriteByte", "WriteRune":
		return true
	default:
		return false
	}
}

// commaParseCollector accumulates H010 signals while walking a function.
type commaParseCollector struct {
	aliases       map[string]string
	evidence      commaParseEvidence
	stripCalls    []*ast.CallExpr
	parseCalls    []*ast.CallExpr
	sawSplitComma bool
	sawJoinEmpty  bool
	hasRebuild    bool
}

// visit is the ast.Inspect callback for collectCommaParseEvidence.
func (c *commaParseCollector) visit(n ast.Node) bool {
	switch node := n.(type) {
	case *ast.CallExpr:
		c.noteCall(node)
	case *ast.BinaryExpr:
		if isCommaCharLiteral(node.X) || isCommaCharLiteral(node.Y) {
			c.evidence.runeFilterLoop = true
		}
	}

	return true
}

// noteCall classifies a single call expression into H010 signals.
func (c *commaParseCollector) noteCall(call *ast.CallExpr) {
	switch {
	case isCommaToEmptyReplace(call, c.aliases):
		c.evidence.stripReplace = true

		c.stripCalls = append(c.stripCalls, call)
	case isStrconvNumberParse(call, c.aliases):
		c.evidence.parsesNumber = true

		c.parseCalls = append(c.parseCalls, call)
	case isBuiltinAppend(call) || isBuilderWrite(call):
		c.hasRebuild = true
	case c.noteStringsSep(call):
	}
}

// noteStringsSep records strings.Split(x, ",") and strings.Join(y, "") calls,
// returning true when the call was one of the two.
func (c *commaParseCollector) noteStringsSep(call *ast.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}

	sep, ok := unquoteString(getBasicLit(call.Args[1]))
	if !ok {
		return false
	}

	switch {
	case isPackageCall(call, "strings", "Split", c.aliases) && sep == ",":
		c.sawSplitComma = true

		return true
	case isPackageCall(call, "strings", "Join", c.aliases) && sep == "":
		c.sawJoinEmpty = true

		return true
	default:
		return false
	}
}

// finalize derives the combined signals from the collected raw evidence.
func (c *commaParseCollector) finalize() commaParseEvidence {
	c.evidence.splitJoin = c.sawSplitComma && c.sawJoinEmpty

	if c.evidence.runeFilterLoop && !c.hasRebuild {
		// A ',' comparison without a string rebuild nearby is more likely a
		// validator ("commas not allowed") than a comma-stripping parser.
		c.evidence.runeFilterLoop = false
	}

	// A strip call nested directly inside a parse call argument list is the
	// single-expression reimplementation of ParseComma/ParseCommaf.
	for _, parse := range c.parseCalls {
		for _, strip := range c.stripCalls {
			if strip.Pos() > parse.Pos() && strip.End() < parse.End() {
				c.evidence.nestedInParse = true
			}
		}
	}

	return c.evidence
}

// collectCommaParseEvidence walks fn and gathers every H010 signal.
func collectCommaParseEvidence(fn *ast.FuncDecl, aliases map[string]string) commaParseEvidence {
	collector := &commaParseCollector{aliases: aliases}

	ast.Inspect(fn, collector.visit)

	return collector.finalize()
}
