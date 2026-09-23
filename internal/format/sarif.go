package format

import (
	"fmt"
	"strconv"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-finding"
)

// sarifToolName is the tool identity emitted in SARIF driver metadata.
const sarifToolName = "clean-wizard"

// ScanOutcome is the neutral per-cleaner scan outcome consumed by the SARIF
// exporter. It mirrors what the scan workflow knows per cleaner, decoupled
// from command-layer types.
type ScanOutcome struct {
	Name           string
	RegistryName   string
	Description    string
	ItemsCount     uint
	BytesCleanable uint64
	Err            error
}

// ScanOutcomesToSARIF renders scan outcomes as a SARIF 2.1.0 document.
//
// Mapping semantics:
//   - reclaimable space (BytesCleanable > 0) becomes an info finding
//     (category "unused") suggesting 'clean-wizard clean --dry-run'
//   - failed scans (Err, non-Infrastructure) become error findings enriched
//     with family/code/retryable metadata
//   - unavailable cleaners and zero-byte outcomes produce no finding:
//     absence of a cleaner is not a finding
//
// The result is deterministic for a given outcome order.
func ScanOutcomesToSARIF(outcomes []ScanOutcome) ([]byte, error) {
	findings, err := scanOutcomesToFindings(outcomes)
	if err != nil {
		return nil, err
	}

	report := finding.NewReportFromFindings(finding.ToolInfo{Name: sarifToolName}, findings)

	data, err := report.ToSARIF()
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "scan.sarif_output", "failed to generate SARIF output")
	}

	return data, nil
}

// scanOutcomesToFindings converts outcomes to validated findings.
func scanOutcomesToFindings(outcomes []ScanOutcome) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0, len(outcomes))

	for _, outcome := range outcomes {
		f, err := scanOutcomeToFinding(outcome)
		if err != nil {
			return nil, errorfamily.WrapCorruption(
				err, "scan.sarif_output",
				fmt.Sprintf("failed to build finding for cleaner %q", outcome.Name),
			)
		}

		if f != nil {
			findings = append(findings, *f)
		}
	}

	return findings, nil
}

// scanOutcomeToFinding maps one outcome; nil means "no finding".
func scanOutcomeToFinding(outcome ScanOutcome) (*finding.Finding, error) {
	if outcome.Err != nil {
		family := errorfamily.Classify(outcome.Err)
		if family == errorfamily.Infrastructure {
			return nil, nil
		}

		return scanFailureFinding(outcome, family)
	}

	if outcome.BytesCleanable == 0 {
		return nil, nil
	}

	return cleanableFinding(outcome)
}

// cleanableFinding reports reclaimable space as an actionable info finding.
func cleanableFinding(outcome ScanOutcome) (*finding.Finding, error) {
	metadata := map[string]string{
		"items": strconv.FormatUint(uint64(outcome.ItemsCount), 10),
		"bytes": strconv.FormatUint(outcome.BytesCleanable, 10),
	}
	if outcome.Description != "" {
		metadata["description"] = outcome.Description
	}

	f, err := finding.NewBuilder(
		finding.RuleName(scanRuleID(outcome)),
		finding.ToolName(sarifToolName),
		fmt.Sprintf("%d items, %s reclaimable", outcome.ItemsCount, Bytes(int64(outcome.BytesCleanable))),
		finding.SeverityInfo,
		finding.Pos(scanLocationURI(outcome), 1, 1),
	).
		WithCategory(finding.CategoryUnused).
		WithConfidence(finding.ConfidenceFull).
		WithFixStrategy(finding.FixStrategySuggest).
		WithSuggestion("Run 'clean-wizard clean --dry-run' to preview reclaiming this space").
		WithMetadata(metadata).
		Build()
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// scanFailureFinding reports a failed cleaner scan as an error finding.
func scanFailureFinding(outcome ScanOutcome, family errorfamily.Family) (*finding.Finding, error) {
	f, err := finding.NewBuilder(
		finding.RuleName(scanRuleID(outcome)),
		finding.ToolName(sarifToolName),
		outcome.Err.Error(),
		finding.SeverityError,
		finding.Pos(scanLocationURI(outcome), 1, 1),
	).
		WithTags(finding.Tag("scan-failure")).
		WithConfidence(finding.ConfidenceFull).
		WithSuggestion("Resolve the cause and re-run 'clean-wizard scan -v'").
		WithMetadata(map[string]string{
			"family":    family.String(),
			"code":      errorfamily.Code(outcome.Err),
			"retryable": strconv.FormatBool(family.IsRetryable()),
		}).
		Build()
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// scanRuleID returns the stable rule identifier for a cleaner outcome.
func scanRuleID(outcome ScanOutcome) string {
	if outcome.RegistryName != "" {
		return outcome.RegistryName
	}

	return outcome.Name
}

// scanLocationURI returns the logical location of a cleaner outcome. Cleaners
// have no single on-disk artifact, so the URI names the cleaner itself.
func scanLocationURI(outcome ScanOutcome) finding.FilePath {
	return finding.FilePath("cleaner://" + scanRuleID(outcome))
}
