package finding

import (
	"path/filepath"
	"strings"
)

// Relativize returns a copy of findings in which every Position.File that is
// an absolute path under root is rewritten to a root-relative, slash-separated
// path and the finding's ID is regenerated from the new position. The input
// slice is never modified; nil yields nil.
//
// IDs are then checkout-independent: a detector invoked with a relative target
// (CLI lane) and the same detector invoked with an absolute working directory
// (SDK/orchestrator lane) produce the same ID for the same source fact, so
// baselines built on one lane gate identically on the other, and IDs no longer
// embed host-specific paths like /home/....
//
// Findings pass through with position and ID unchanged when their file is
// empty, already relative, or absolute but outside root (the caller may
// legitimately report against files it cannot relativize), or when root cannot
// be resolved to an absolute path. Line-0 file-level findings carry hash IDs
// derived from (tool, rule, file); relativizing their file changes the hash,
// which is the point: the new hash is stable across hosts where the old one
// was not.
//
// Column presence in IDs (line-only vs line:col) is deliberately not
// normalized: that is a per-rule-family shape difference, and changing it
// would invalidate existing baselines without a portability gain.
func Relativize(findings []Finding, root string) []Finding {
	if findings == nil {
		return nil
	}

	out := make([]Finding, len(findings))
	copy(out, findings)

	if root == "" {
		return out
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return out
	}

	absRoot = filepath.Clean(absRoot)

	for i := range out {
		file := string(out[i].Position.File)
		if file == "" || !filepath.IsAbs(file) {
			continue
		}

		rel, err := filepath.Rel(absRoot, filepath.Clean(file))
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}

		out[i].Position.File = FilePath(filepath.ToSlash(rel))
		out[i].ID = GenerateID(out[i].ToolName, out[i].Rule, out[i].Position)
	}

	return out
}
