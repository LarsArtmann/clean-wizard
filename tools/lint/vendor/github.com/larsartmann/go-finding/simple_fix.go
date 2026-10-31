package finding

import (
	"fmt"
	"os"
	"strings"
)

// SimpleFixResult records the outcome of applying a single BeforeCode→AfterCode fix.
type SimpleFixResult struct {
	FindingID ID     // The finding that was processed
	Applied   bool   // Whether the replacement was applied
	Reason    string // Why it was skipped (when Applied is false)
}

// ApplySimpleFixes applies BeforeCode→AfterCode string replacements for all
// findings with FixStrategyDirect. Files are read, modified in memory, and
// written back. Findings without BeforeCode or AfterCode are skipped.
// This is the 80% case for consumers that don't need the full pipeline FixEngine.
//
// Findings carrying a multi-edit list (more than one TextEdit) are refused,
// not half-applied: this helper applies a single string replacement per
// finding. Use the pipeline module's FixApplier for edit lists.
//
// Returns a map of file path to per-finding results. If a file cannot be read,
// all its findings are marked as not applied with the error reason.
func ApplySimpleFixes(findings []Finding) map[FilePath][]SimpleFixResult {
	results := make(map[FilePath][]SimpleFixResult)

	for file, fileFindings := range GroupByFile(findings) {
		results[file] = applyFixesToFile(file, fileFindings)
	}

	return results
}

func applyFixesToFile(file FilePath, findings []Finding) []SimpleFixResult {
	content, err := os.ReadFile(string(file))
	if err != nil {
		return skippedResults(findings, fmt.Sprintf("read file: %v", err))
	}

	modified := string(content)
	fileResults := make([]SimpleFixResult, 0, len(findings))
	anyApplied := false

	for _, f := range findings {
		if len(f.Edits) > 1 {
			fileResults = append(fileResults, SimpleFixResult{
				FindingID: f.ID,
				Applied:   false,
				Reason:    "finding has a multi-edit fix list; use pipeline.FixApplier",
			})

			continue
		}

		if f.BeforeCode == "" || f.AfterCode == "" {
			fileResults = append(fileResults, SimpleFixResult{
				FindingID: f.ID,
				Applied:   false,
				Reason:    "missing BeforeCode or AfterCode",
			})

			continue
		}

		if !strings.Contains(modified, f.BeforeCode) {
			fileResults = append(fileResults, SimpleFixResult{
				FindingID: f.ID,
				Applied:   false,
				Reason:    "BeforeCode not found in file",
			})

			continue
		}

		modified = strings.Replace(modified, f.BeforeCode, f.AfterCode, 1)
		fileResults = append(fileResults, SimpleFixResult{
			FindingID: f.ID,
			Applied:   true,
		})
		anyApplied = true
	}

	if anyApplied {
		//nolint:gosec // G703: file path is validated by caller
		if writeErr := os.WriteFile(string(file), []byte(modified), 0o644); writeErr != nil {
			return skippedResults(findings, fmt.Sprintf("write file: %v", writeErr))
		}
	}

	return fileResults
}

func skippedResults(findings []Finding, reason string) []SimpleFixResult {
	results := make([]SimpleFixResult, len(findings))
	for i, f := range findings {
		results[i] = SimpleFixResult{
			FindingID: f.ID,
			Applied:   false,
			Reason:    reason,
		}
	}

	return results
}
