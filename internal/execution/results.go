package execution

import (
	"sort"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// StepResult holds the outcome of a single cleaner step within the workflow.
type StepResult struct {
	Name     string
	Clean    domain.CleanResult
	Err      error
	Duration time.Duration
}

// Status classifies a step result as succeeded, skipped, or failed.
type StepStatus string

const (
	StepStatusSucceeded StepStatus = "succeeded"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusFailed    StepStatus = "failed"
)

// Status returns the classification for this step result.
func (s StepResult) Status() StepStatus {
	if s.Err != nil {
		if cleaner.IsNotAvailableError(s.Err) {
			return StepStatusSkipped
		}
		return StepStatusFailed
	}
	return StepStatusSucceeded
}

// WorkflowResult aggregates results from all cleaner steps in a workflow run.
type WorkflowResult struct {
	Steps             []StepResult
	TotalBytesFreed   uint64
	TotalItemsRemoved uint
	TotalItemsFailed  uint
	Duration          time.Duration
}

// Succeeded returns only steps that completed successfully.
func (wr *WorkflowResult) Succeeded() []StepResult {
	var out []StepResult
	for _, s := range wr.Steps {
		if s.Status() == StepStatusSucceeded {
			out = append(out, s)
		}
	}
	return out
}

// Skipped returns only steps that were skipped (cleaner not available).
func (wr *WorkflowResult) Skipped() []StepResult {
	var out []StepResult
	for _, s := range wr.Steps {
		if s.Status() == StepStatusSkipped {
			out = append(out, s)
		}
	}
	return out
}

// Failed returns only steps that failed with a non-availability error.
func (wr *WorkflowResult) Failed() []StepResult {
	var out []StepResult
	for _, s := range wr.Steps {
		if s.Status() == StepStatusFailed {
			out = append(out, s)
		}
	}
	return out
}

// CleanResultsMap builds a name→CleanResult map for successful steps,
// matching the shape expected by the existing display functions.
func (wr *WorkflowResult) CleanResultsMap() map[string]domain.CleanResult {
	m := make(map[string]domain.CleanResult)
	for _, s := range wr.Steps {
		if s.Status() == StepStatusSucceeded {
			m[s.Name] = s.Clean
		}
	}
	return m
}

// resultCollector accumulates StepResults as the workflow executes.
// It is thread-safe because go-workflow may run steps concurrently.
// Results are tracked with a registration index to preserve deterministic
// ordering matching the input cleaner selection, regardless of completion order.
type resultCollector struct {
	mu         sync.Mutex
	results    []StepResult
	orderIndex map[string]int
}

func newResultCollector() *resultCollector {
	return &resultCollector{orderIndex: make(map[string]int)}
}

func (rc *resultCollector) register(name string, index int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.orderIndex[name] = index
}

func (rc *resultCollector) record(name string, clean domain.CleanResult, err error, duration time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.results = append(rc.results, StepResult{
		Name:     name,
		Clean:    clean,
		Err:      err,
		Duration: duration,
	})
}

// sortedByRegistration returns results ordered by their original registration
// index, ensuring deterministic output regardless of parallel completion order.
func (rc *resultCollector) sortedByRegistration() []StepResult {
	sorted := make([]StepResult, len(rc.results))
	copy(sorted, rc.results)
	sort.SliceStable(sorted, func(i, j int) bool {
		ci, oki := rc.orderIndex[sorted[i].Name]
		cj, okj := rc.orderIndex[sorted[j].Name]
		if !oki || !okj {
			return sorted[i].Name < sorted[j].Name
		}
		return ci < cj
	})
	return sorted
}
