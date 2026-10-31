package humanizelint

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/larsartmann/go-finding"
)

// SuppressionDirective records a //nolint-style directive that is attached
// to a function declaration. It is used by --verify-suppressions to detect
// stale or misspelled suppressions.
type SuppressionDirective struct {
	FilePath     string
	FunctionLine int
	Line         int
	Column       int
	RawText      string
	Suppressed   []string
}

// allRuleIDs returns the stable rule identifiers H001-H009. Kept as a function
// rather than a global so the set cannot drift from the canonical constants.
func allRuleIDs() []string {
	return []string{
		RuleIDH001,
		RuleIDH002,
		RuleIDH003,
		RuleIDH004,
		RuleIDH005,
		RuleIDH006,
		RuleIDH007,
		RuleIDH008,
		RuleIDH009,
	}
}

// suppressesGohumanize reports whether the parsed suppression list mentions our
// linter (gohumanize) or any of our rule IDs.
func suppressesGohumanize(suppressed []string) bool {
	if isSuppressedAll(suppressed) {
		return true
	}

	for _, ruleID := range allRuleIDs() {
		if isSuppressedRule(suppressed, ruleID) {
			return true
		}
	}

	return false
}

// targetsGohumanize reports whether a directive is addressed to this linter:
// it names gohumanize or scopes a rule ID. A bare "all" is deliberately blunt
// and usually exists for another linter's benefit, so its staleness cannot be
// judged from this linter's findings alone — verification skips it.
func targetsGohumanize(suppressed []string) bool {
	for _, s := range suppressed {
		if s == nolintLinterName || looksLikeRuleID(s) {
			return true
		}
	}

	return false
}

// hasUnknownHumanizeLinterName reports whether the suppression list contains a
// linter name that looks like ours but is not exactly "gohumanize". The common
// AI mistake is writing "go-humanize-linter" because that matches the module
// path. Such directives silently do nothing.
func hasUnknownHumanizeLinterName(suppressed []string) bool {
	for _, s := range suppressed {
		lower := strings.ToLower(s)

		if lower == nolintLinterName || lower == nolintAllMarker {
			continue
		}

		if strings.Contains(lower, "humanize") {
			return true
		}
	}

	return false
}

// collectSuppressions walks dir and returns every //nolint-style directive that
// either suppresses our linter or misspells its name.
func collectSuppressions(dir string) ([]SuppressionDirective, error) {
	files, err := WalkGoDir(dir)
	if err != nil {
		return nil, err
	}

	var out []SuppressionDirective

	for _, pf := range files {
		out = append(out, suppressionsInFile(pf.Fset, pf.File, pf.Path)...)
	}

	return out, nil
}

// suppressionsInFile returns every directive in one parsed file that either
// suppresses our linter or misspells its name. Shared by the directory walk
// (CLI) and the pre-parsed-files walk (golangci plugin).
func suppressionsInFile(fset *token.FileSet, file *ast.File, filePath string) []SuppressionDirective {
	if file == nil {
		return nil
	}

	var out []SuppressionDirective

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		out = append(out, extractFunctionSuppressions(fset, file, fn, filePath)...)
	}

	return out
}

func extractFunctionSuppressions(
	fset *token.FileSet,
	file *ast.File,
	fn *ast.FuncDecl,
	filePath string,
) []SuppressionDirective {
	if file == nil || fn == nil {
		return nil
	}

	// fnLine is stored in SuppressionDirective.FunctionLine for use by
	// directiveMatchesFinding, which looks up findings at this line.
	// All detectors currently emit findings at fn.Pos(), so this matches.
	fnLine := fset.Position(fn.Pos()).Line

	var out []SuppressionDirective

	for _, group := range file.Comments {
		isDoc := group == fn.Doc

		for _, comment := range group.List {
			suppressed := suppressedRules(comment.Text)
			if suppressed == nil {
				continue
			}

			cmtLine := fset.Position(comment.Pos()).Line
			if !commentAssociatedWithFunc(fset, fn, isDoc, cmtLine) {
				continue
			}

			pos := fset.Position(comment.Pos())

			out = append(out, SuppressionDirective{
				FilePath:     filePath,
				FunctionLine: fnLine,
				Line:         pos.Line,
				Column:       pos.Column,
				RawText:      comment.Text,
				Suppressed:   suppressed,
			})
		}
	}

	return out
}

// VerifySuppressions reports two classes of suppression problems:
//  1. Unknown linter names that look like "gohumanize" (e.g. "go-humanize-linter").
//  2. //nolint:gohumanize[:Hxxx] directives that did not suppress any finding.
//
// Staleness is judged against a fresh UNSUPPRESSED detection pass over dir,
// not a run report: the detection pipeline drops suppressed findings before
// any report is built, so report-based matching flagged every working
// directive as stale — the findings it suppressed were absent by
// construction. Re-detecting also makes the verdict rule-set independent:
// a directive stays valid when its rule is merely disabled in this run.
//
// The returned findings use the pseudo-rule ID "H0SUP" so they are clearly
// verification diagnostics, not humanize reimplementation findings.
func VerifySuppressions(dir string) ([]finding.Finding, error) {
	directives, err := collectSuppressions(dir)
	if err != nil {
		return nil, err
	}

	index, err := unsuppressedFindingsIndex(dir)
	if err != nil {
		return nil, err
	}

	return verifyDirectives(directives, index), nil
}

// VerifySuppressionsInFiles is the plugin-path variant of VerifySuppressions.
// It collects directives from pre-parsed files (pass.Files) and re-detects
// unsuppressed findings from the same files — see VerifySuppressions for why
// a run report cannot decide staleness.
func VerifySuppressionsInFiles(fset *token.FileSet, files []*ast.File) []finding.Finding {
	var directives []SuppressionDirective

	for _, file := range files {
		if file == nil {
			continue
		}

		filePath := fset.Position(file.Pos()).Filename
		directives = append(directives, suppressionsInFile(fset, file, filePath)...)
	}

	return verifyDirectives(directives, unsuppressedFindingsFromFiles(fset, files))
}

// verifyDirectives is the shared core of VerifySuppressions and
// VerifySuppressionsInFiles. It checks each directive for unknown linter
// names and stale suppressions, returning H0SUP findings for problems.
func verifyDirectives(
	directives []SuppressionDirective,
	findingsByPosition map[string]map[int][]finding.Finding,
) []finding.Finding {
	var out []finding.Finding

	for _, d := range directives {
		if hasUnknownHumanizeLinterName(d.Suppressed) {
			out = append(out, makeSuppressionVerificationFinding(
				d, "unknown linter name in //nolint directive (did you mean gohumanize?)",
			))

			continue
		}

		if !suppressesGohumanize(d.Suppressed) || !targetsGohumanize(d.Suppressed) {
			continue
		}

		if !directiveMatchesFinding(d, d.Suppressed, findingsByPosition) {
			out = append(out, makeSuppressionVerificationFinding(
				d, "//nolint:gohumanize directive suppresses zero findings in this function",
			))
		}
	}

	return out
}

// unsuppressedFindingsIndex walks dir and runs every detector over every
// function, honouring no //nolint directives, then indexes the findings by
// file path and function-declaration line. This is the lookup structure
// used to test whether a suppression directive actually suppressed
// something — detectors emit at fn.Pos(), so the declaration line is the
// shared key on both sides (see ADR 0006).
func unsuppressedFindingsIndex(dir string) (map[string]map[int][]finding.Finding, error) {
	files, err := WalkGoDir(dir)
	if err != nil {
		return nil, err
	}

	index := make(map[string]map[int][]finding.Finding)

	for _, pf := range files {
		indexDetection(pf.Fset, pf.File, pf.Path, index)
	}

	return index, nil
}

// unsuppressedFindingsFromFiles is the plugin-path variant of
// unsuppressedFindingsIndex over pre-parsed files.
func unsuppressedFindingsFromFiles(
	fset *token.FileSet,
	files []*ast.File,
) map[string]map[int][]finding.Finding {
	index := make(map[string]map[int][]finding.Finding)

	for _, file := range files {
		if file == nil {
			continue
		}

		indexDetection(fset, file, fset.Position(file.Pos()).Filename, index)
	}

	return index
}

// indexDetection runs every registered detector over every function of one
// parsed file and accumulates the findings into index, ignoring all
// //nolint directives.
func indexDetection(
	fset *token.FileSet,
	file *ast.File,
	filePath string,
	index map[string]map[int][]finding.Finding,
) {
	if file == nil {
		return
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		fnLine := fset.Position(fn.Pos()).Line

		for _, detector := range allRuleDetectors() {
			for _, f := range detector.detector(fset, file, fn, filePath) {
				if index[filePath] == nil {
					index[filePath] = make(map[int][]finding.Finding)
				}

				index[filePath][fnLine] = append(index[filePath][fnLine], f)
			}
		}
	}
}

// directiveMatchesFinding reports whether any finding in the same function
// (same file and function-start line) is covered by the suppression directive.
//
// Note: This keys on d.FunctionLine (the function declaration line), which
// works because all detectors emit findings at fn.Pos(). If detectors are
// upgraded to per-statement positions (see ADR 0006, Approach A), this lookup
// must be changed to match the finding's line falling within the
// [Lbrace, Rbrace] range instead.
func directiveMatchesFinding(
	d SuppressionDirective,
	suppressed []string,
	findingsByPosition map[string]map[int][]finding.Finding,
) bool {
	fileFindings, ok := findingsByPosition[d.FilePath]
	if !ok {
		return false
	}

	for _, f := range fileFindings[d.FunctionLine] {
		if isSuppressedRule(suppressed, string(f.Rule)) {
			return true
		}
	}

	return false
}

func makeSuppressionVerificationFinding(d SuppressionDirective, message string) finding.Finding {
	return finding.NewBuilder(
		finding.RuleName(RuleIDH0SUP),
		finding.ToolName("go-humanize-linter"),
		message,
		finding.SeverityWarning,
		finding.Pos(finding.FilePath(d.FilePath), d.Line, d.Column),
	).
		WithCategory(finding.CategoryConfiguration).
		WithConfidence(finding.ConfidenceHigh).
		MustBuild()
}

// VerifySuppressionComment is a convenience helper for tests: it parses a
// single comment string and returns whether it suppresses our linter and
// whether it contains a misspelled linter name.
func VerifySuppressionComment(comment string) (bool, bool) {
	suppressed := suppressedRules(comment)
	if suppressed == nil {
		return false, false
	}

	return suppressesGohumanize(suppressed), hasUnknownHumanizeLinterName(suppressed)
}
