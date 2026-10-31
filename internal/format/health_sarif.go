package format

import (
	"strconv"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-finding"
)

// HealthOutcome is the neutral per-service health-check outcome consumed by
// the SARIF exporter, decoupled from command-layer types.
type HealthOutcome struct {
	Service string
	Err     error
}

// HealthOutcomesToSARIF renders health-check outcomes as a SARIF 2.1.0
// document.
//
// Mapping semantics:
//   - failed services become error findings (category "health") enriched with
//     family/code/retryable metadata
//   - healthy services produce no finding: health is the absence of findings
//
// The result is deterministic for a given outcome order.
func HealthOutcomesToSARIF(outcomes []HealthOutcome) ([]byte, error) {
	findings := make([]finding.Finding, 0, len(outcomes))

	for _, outcome := range outcomes {
		if outcome.Err == nil {
			continue
		}

		family := errorfamily.Classify(outcome.Err)

		f, err := finding.NewBuilder(
			finding.RuleName(outcome.Service+".unhealthy"),
			finding.ToolName(sarifToolName),
			outcome.Err.Error(),
			finding.SeverityError,
			finding.Pos(healthLocationURI(outcome.Service), 1, 1),
		).
			WithCategory(finding.Category("health")).
			WithTags(finding.Tag("health-check")).
			WithConfidence(finding.ConfidenceFull).
			WithSuggestion("Resolve the service failure and re-run 'clean-wizard doctor'").
			WithMetadata(map[string]string{
				"family":    family.String(),
				"code":      errorfamily.Code(outcome.Err),
				"retryable": strconv.FormatBool(family.IsRetryable()),
			}).
			Build()
		if err != nil {
			return nil, errorfamily.WrapCorruptionf(
				err,
				"doctor.sarif_output",
				"failed to build health finding for service %q",
				outcome.Service,
			)
		}

		findings = append(findings, f)
	}

	return FindingsToSARIF(findings)
}

// healthLocationURI names the service itself as the finding location; DI
// services have no single on-disk artifact.
func healthLocationURI(serviceName string) finding.FilePath {
	return finding.FilePath("service://" + serviceName)
}
