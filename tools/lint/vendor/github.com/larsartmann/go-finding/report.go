package finding

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-finding/lockutil"
)

// Report is the top-level container for a tool run.
// The zero value is safe for concurrent use. Use [NewReport] to create
// a Report with pre-allocated findings.
// All methods are safe for concurrent use. Read methods (FindByID, Len,
// ActiveFindings, etc.) acquire a read lock; write methods (AddFinding,
// AddFindings, MergeInto) acquire a write lock.
type Report struct {
	mu       sync.RWMutex
	Tool     ToolInfo `json:"tool"` // Tool metadata
	findings []Finding
	Summary  Summary `json:"summary"` // Aggregated statistics
}

// Validate returns an error if the Report is invalid.
// It checks Tool info and validates each finding, returning joined errors.
// Safe for concurrent use.
func (r *Report) Validate() error {
	return withReadLock(r, func() error {
		var errs []error

		err := r.Tool.Validate()
		if err != nil {
			errs = append(errs, err)
		}

		for i, f := range r.findings {
			err := f.Validate()
			if err != nil {
				errs = append(errs, fmt.Errorf("findings[%d]: %w", i, err))
			}
		}

		return errors.Join(errs...)
	})
}

// ToolInfo contains metadata about the tool that generated the report.
type ToolInfo struct {
	Name    string `json:"name"`              // Tool name
	Version string `json:"version,omitempty"` // Tool version
}

// Validate returns an error if the ToolInfo is invalid.
// A valid ToolInfo requires a non-empty Name.
func (t ToolInfo) Validate() error {
	if t.Name == "" {
		return NewValidationError("ToolInfo.Name is required", nil)
	}

	return nil
}

// Summary contains aggregated statistics for a report.
type Summary struct {
	Total         int                 `json:"total"`                   // Total findings
	BySeverity    map[Severity]int    `json:"bySeverity"`              // Count by severity
	ByCategory    map[Category]int    `json:"byCategory,omitempty"`    // Count by category
	ByFixStrategy map[FixStrategy]int `json:"byFixStrategy,omitempty"` // Count by fix strategy
	FilesAffected int                 `json:"filesAffected,omitempty"` // Unique files with findings
	FilesScanned  int                 `json:"filesScanned"`            // Total files scanned (including clean files); always emitted — "scanned 0" is coverage evidence, not an omission
	Suppressed    int                 `json:"suppressed,omitempty"`    // Count of suppressed findings

	// SkippedModules lists nested go.mod modules under the analysis target
	// that the tool did NOT analyze (go tooling treats nested go.mod files as
	// hard module boundaries). Tools whose analysis is module-agnostic leave
	// it nil, which omits the field.
	SkippedModules []string `json:"skippedModules,omitempty"`
}

// NewReport creates a new report with the given tool info.
func NewReport(tool ToolInfo) *Report {
	r := &Report{
		Tool:     tool,
		findings: make([]Finding, 0),
		Summary:  Summary{},
	}
	r.ComputeSummary()

	return r
}

// NewReportFromFindings creates a report from tool info and a findings slice,
// calling AddFindings and ComputeSummary in one step.
// This eliminates the repetitive 4-line NewReport + AddFindings + ComputeSummary
// boilerplate duplicated across consumers.
func NewReportFromFindings(tool ToolInfo, findings []Finding) *Report {
	r := NewReport(tool)
	r.AddFindings(findings)
	r.ComputeSummary()

	return r
}

// AddFinding adds a finding to the report.
// Safe for concurrent use.
func (r *Report) AddFinding(f Finding) {
	withLock(r, func() struct{} {
		r.findings = append(r.findings, f)

		return struct{}{}
	})
}

// WithFinding adds a finding and returns the report for chaining.
// Example: report.WithFinding(f1).WithFinding(f2).
func (r *Report) WithFinding(f Finding) *Report {
	r.AddFinding(f)

	return r
}

// AddFindings adds multiple findings to the report.
// Safe for concurrent use.
func (r *Report) AddFindings(findings []Finding) {
	withLock(r, func() struct{} {
		r.findings = append(r.findings, findings...)

		return struct{}{}
	})
}

// MergeInto returns a new Report containing findings from both r and other.
// Neither receiver nor other is modified. The new report uses r's ToolInfo.
func (r *Report) MergeInto(other *Report) *Report {
	r.mu.RLock()
	rFindings := make([]Finding, len(r.findings))
	copy(rFindings, r.findings)
	rTool := r.Tool
	r.mu.RUnlock()

	other.mu.RLock()
	oFindings := make([]Finding, len(other.findings))
	copy(oFindings, other.findings)
	other.mu.RUnlock()

	merged := &Report{
		Tool:     rTool,
		findings: make([]Finding, 0, len(rFindings)+len(oFindings)),
	}
	merged.findings = append(merged.findings, rFindings...)
	merged.findings = append(merged.findings, oFindings...)
	merged.ComputeSummary()

	return merged
}

// readFindings returns a shallow copy of findings under RLock.
// Safe for concurrent use. The caller receives a snapshot that won't
// be affected by subsequent AddFinding/AddFindings calls.
func (r *Report) readFindings() []Finding {
	return withReadLock(r, func() []Finding {
		findings := make([]Finding, len(r.findings))
		copy(findings, r.findings)

		return findings
	})
}

// withReadLock runs fn while holding r.mu.RLock and returns its result.
// Consolidates the r.mu.RLock()/defer r.mu.RUnlock() boilerplate around
// short read-side computations on Report.
func withReadLock[T any](r *Report, fn func() T) T {
	return lockutil.RLocked(&r.mu, fn)
}

// withReadLockErr is the (T, error)-returning counterpart to withReadLock.
// Use it for read paths that may fail mid-flight (e.g., serialization).
// Implemented via withReadLock + a closure that returns a (T, error)-shaped
// tuple to avoid duplicating the lock boilerplate.
func withReadLockErr[T any](r *Report, fn func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}

	res := withReadLock(r, func() result {
		v, err := fn()

		return result{value: v, err: err}
	})

	return res.value, res.err
}

// withLock runs fn while holding r.mu.Lock and returns its result.
// Consolidates the r.mu.Lock()/defer r.mu.Unlock() boilerplate around
// short write-side mutations on Report.
func withLock[T any](r *Report, fn func() T) T {
	return lockutil.Locked(&r.mu, fn)
}

// findingsLocked returns the findings slice without acquiring the lock.
// Caller MUST hold r.mu (RLock or Lock).
func (r *Report) findingsLocked() []Finding {
	return r.findings
}

// ComputeSummary recalculates the summary from the current findings.
// Uses time.Now() for suppression expiry checks. For deterministic results
// in tests, use ComputeSummaryAt.
// Safe for concurrent use with AddFinding/AddFindings.
func (r *Report) ComputeSummary() {
	r.computeSummaryAt(time.Now())
}

// ComputeSummaryAt recalculates the summary using the given time for
// suppression expiry checks. Use this in tests for deterministic results.
func (r *Report) ComputeSummaryAt(now time.Time) {
	r.computeSummaryAt(now)
}

func (r *Report) computeSummaryAt(now time.Time) {
	withLock(r, func() struct{} {
		r.Summary.Total = len(r.findings)
		r.Summary.BySeverity = make(map[Severity]int)
		r.Summary.ByCategory = make(map[Category]int)
		r.Summary.ByFixStrategy = make(map[FixStrategy]int)

		files := make(map[string]struct{})
		suppressed := 0

		for _, f := range r.findings {
			r.Summary.BySeverity[f.Severity]++

			r.Summary.ByFixStrategy[f.FixStrategy]++
			if f.Category != "" {
				r.Summary.ByCategory[f.Category]++
			}

			if f.Position.File != "" {
				files[string(f.Position.File)] = struct{}{}
			}

			if f.IsSuppressedAt(now) {
				suppressed++
			}
		}

		r.Summary.FilesAffected = len(files)
		r.Summary.Suppressed = suppressed

		return struct{}{}
	})
}
