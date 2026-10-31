package humanizelint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/go-finding"
)

// ParsedFile holds a parsed Go source file together with its token file set.
type ParsedFile struct {
	Path string
	Fset *token.FileSet
	File *ast.File
}

// WalkError is returned by WalkGoDir and checkFuncDecls when the directory walk
// itself fails. It carries the directory the walker was asked to scan so callers
// can produce actionable error messages without parsing the error string.
type WalkError struct {
	Dir string
	Err error
}

// Error implements the error interface.
func (e *WalkError) Error() string {
	return fmt.Sprintf("walking %s: %v", e.Dir, e.Err)
}

// Unwrap returns the underlying error so errors.Is / errors.AsType work
// through the chain to the original fs.PathError, syscall.Errno, etc.
func (e *WalkError) Unwrap() error {
	return e.Err
}

// skipDirs are directory basenames that WalkGoDir never descends into.
// Hidden directories (leading ".") and backup trees (trailing ".bak") are
// skipped by rule — see skipDirName.
var skipDirs = map[string]bool{ //nolint:gochecknoglobals // package-level lookup table
	"vendor":       true,
	"testdata":     true,
	"node_modules": true,
	"archived":     true,
	"forks":        true,
	"__debug":      true,
}

// skipDirName reports whether WalkGoDir must not descend into the directory
// with this basename:
//   - hidden directories (leading ".") — tooling state, not maintained source
//     (.git, .idea, .bf-jscpd-worktree, ...);
//   - backup trees (trailing ".bak") — frozen copies that would double-report
//     every upstream finding (e.g. BuildFlow.vendor.bak);
//   - exact basenames in skipDirs, including "archived" and "forks" — archived
//     and upstream-owned code yields findings nobody can act on.
func skipDirName(name string) bool {
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".bak") {
		return true
	}

	return skipDirs[name]
}

// WalkGoDir walks dir recursively, parses every non-test .go file, and returns
// them. Files that fail to parse are silently skipped — syntax errors are the
// compiler's job, not the linter's. On walk failure it returns a *WalkError
// wrapping the underlying fs error. Callers should errors.AsType[*WalkError](err)
// to read .Dir.
//
// Each file is read once and the content is reused for both the generated-file
// check (via gogenfilter) and the AST parser. This keeps the walker I/O-free
// beyond a single read per file and gives full two-phase generated detection.
//
// Always returns *WalkError (an error type) — signature stays `error` for
// v0.1.x API compatibility.
func WalkGoDir(dir string) ([]ParsedFile, error) {
	var files []ParsedFile

	fset := token.NewFileSet()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}

		if d.IsDir() {
			// The root entry is exempt from skip rules: `skipDirName(".")`
			// matches the caller-supplied root ".", which would abort the
			// entire walk with zero files scanned (go-humanize-linter .).
			// Whatever directory the caller explicitly names gets scanned.
			if path == dir {
				return nil
			}

			if skipDirName(d.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Read once and reuse the bytes for both the check and the parser.
		// Path comes from filepath.WalkDir on the caller-supplied directory;
		// trust is delegated to the caller (WalkGoDir is a public API).
		content, readErr := os.ReadFile(path) //nolint:gosec // G304
		if readErr != nil {
			return nil //nolint:nilerr // unreadable files are the OS's problem, not the linter's
		}

		if IsGeneratedFile(path, content) {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, content, parser.ParseComments)
		if parseErr != nil {
			return nil //nolint:nilerr // syntax errors are the compiler's job, not the linter's
		}

		files = append(files, ParsedFile{Path: path, Fset: fset, File: file})

		return nil
	})
	if err != nil {
		return nil, &WalkError{Dir: dir, Err: err}
	}

	return files, nil
}

// detectorFunc inspects a single function declaration and returns findings for
// any reimplementation patterns it detects.
type detectorFunc func(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, filePath string) []finding.Finding

// checkFuncDecls walks dir, parses Go files, and applies detect to every
// function declaration. ruleID is used for per-rule //nolint scoping: a
// directive like //nolint:gohumanize:H002 suppresses only H002, not every
// rule. On walk failure it returns a *WalkError wrapping the underlying fs
// error so callers can read Dir via errors.AsType[*WalkError].
func checkFuncDecls(dir, ruleID string, detect detectorFunc) ([]finding.Finding, error) {
	files, err := WalkGoDir(dir)
	if err != nil {
		return nil, &WalkError{Dir: dir, Err: err}
	}

	var all []finding.Finding

	for _, pf := range files {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if isSuppressedRule(funcSuppressions(pf.Fset, pf.File, fn), ruleID) {
				continue
			}

			all = append(all, detect(pf.Fset, pf.File, fn, pf.Path)...)
		}
	}

	return all, nil
}

// posOf extracts the line and column from an AST position within fset.
func posOf(fset *token.FileSet, pos token.Pos) (int, int) {
	p := fset.Position(pos)

	return p.Line, p.Column
}
