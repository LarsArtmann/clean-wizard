package finding

import (
	"context"
	"slices"
)

// DetectResult carries everything one detector run produced: its findings
// plus the coverage evidence a host needs to answer "did the tool actually
// see every file?". Summary.FilesScanned and Summary.SkippedModules exist
// since v1.14.0, but the Detector contract returns []Finding only, so a
// detector had no way to populate them — consumers rendered filesScanned: 0
// for provider-built reports with no way to tell a real zero from a missing
// measurement.
type DetectResult struct {
	// Findings are the issues the run surfaced. The same findings a plain
	// Detector.Detect call would return.
	Findings []Finding

	// FilesScanned is the number of files the detector actually examined,
	// including clean ones. Zero is legal evidence (an empty target tree),
	// which is exactly why the field rides on the run result instead of
	// being inferred from the findings.
	FilesScanned int

	// SkippedModules lists nested module roots the detector deliberately
	// did not descend into (e.g. vendored or sandboxed nested go.mod trees).
	// Absent modules are not scanned files; reporting both sides makes the
	// coverage arithmetic checkable.
	SkippedModules []string
}

// DetectorV2 is the OPTIONAL richer detect contract. Detectors that can
// measure their own coverage implement DetectV2 alongside the unchanged
// Detector methods; hosts that only know Detector keep working untouched
// (RunDetect falls back). The v1 method stays part of the interface so a
// DetectorV2 value is usable anywhere a Detector is — DetectV2 and Detect
// should agree on the findings they return for the same run.
type DetectorV2 interface {
	Detector

	// DetectV2 runs the detector and returns findings plus coverage.
	DetectV2(ctx context.Context) (DetectResult, error)
}

// RunDetect executes a detector through the richest contract it implements:
// DetectorV2's DetectV2 when available, plain Detect otherwise. V1
// detectors produce a result with coverage zeroed — that is the documented
// contract, not an error (the host cannot know what a v1 detector scanned).
// The detector's error is passed through unwrapped: callers layer their own
// context (the pipeline adds "detector %s:"), and a wrapper here would only
// duplicate it.
func RunDetect(ctx context.Context, d Detector) (DetectResult, error) {
	if v2, ok := d.(DetectorV2); ok {
		return v2.DetectV2(ctx) //nolint:wrapcheck // deliberate pass-through: the caller owns error context
	}

	findings, err := d.Detect(ctx)
	if err != nil {
		return DetectResult{}, err //nolint:wrapcheck // deliberate pass-through: the caller owns error context
	}

	return DetectResult{Findings: findings}, nil
}

// AddDetectResult adds a run's findings to the report and records its
// coverage. FilesScanned accumulates across calls (a report fed by several
// detectors sums their scan counts) and SkippedModules is unioned in first
// seen order. ComputeSummary does not touch either field, so coverage
// survives summary recomputation.
// Safe for concurrent use.
func (r *Report) AddDetectResult(res DetectResult) {
	withLock(r, func() struct{} {
		r.findings = append(r.findings, res.Findings...)
		r.Summary.FilesScanned += res.FilesScanned

		for _, module := range res.SkippedModules {
			if !r.skippedModuleLocked(module) {
				r.Summary.SkippedModules = append(r.Summary.SkippedModules, module)
			}
		}

		return struct{}{}
	})
}

// skippedModuleLocked reports whether module is already recorded.
// Caller MUST hold r.mu.
func (r *Report) skippedModuleLocked(module string) bool {
	return slices.Contains(r.Summary.SkippedModules, module)
}
