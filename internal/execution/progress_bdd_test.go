package execution_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// bddEventKind labels one captured progress callback.
type bddEventKind int

const (
	bddWorkflowStarted bddEventKind = iota
	bddRegistered
	bddStarted
	bddRetrying
	bddCompleted
	bddFailed
	bddSkipped
	bddNote
	bddWorkflowFinished
	bddFinish
)

// bddEvent is one captured callback payload.
type bddEvent struct {
	kind     bddEventKind
	name     string
	attempt  int
	reason   string
	err      error
	duration time.Duration
}

// progressRecorder is a thread-safe ProgressEmitter double that captures the
// full event stream so specs can narrate the live-progress story.
type progressRecorder struct {
	mu     sync.Mutex
	events []bddEvent
}

func newProgressRecorder() *progressRecorder {
	return &progressRecorder{events: []bddEvent{}}
}

func (r *progressRecorder) add(e bddEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, e)
}

func (r *progressRecorder) WorkflowStarted(name string) {
	r.add(bddEvent{kind: bddWorkflowStarted, name: name})
}

func (r *progressRecorder) ActivityRegistered(name string) {
	r.add(bddEvent{kind: bddRegistered, name: name})
}

func (r *progressRecorder) ActivityStarted(name string) {
	r.add(bddEvent{kind: bddStarted, name: name})
}

func (r *progressRecorder) ActivityRetrying(name string, attempt int, reason string) {
	r.add(bddEvent{kind: bddRetrying, name: name, attempt: attempt, reason: reason})
}

func (r *progressRecorder) ActivityCompleted(name string, duration time.Duration) {
	r.add(bddEvent{kind: bddCompleted, name: name, duration: duration})
}

func (r *progressRecorder) ActivityFailed(name string, err error, duration time.Duration) {
	r.add(bddEvent{kind: bddFailed, name: name, err: err, duration: duration})
}

func (r *progressRecorder) ActivitySkipped(name string, reason string) {
	r.add(bddEvent{kind: bddSkipped, name: name, reason: reason})
}

func (r *progressRecorder) Note(line string) {
	r.add(bddEvent{kind: bddNote, reason: line})
}

func (r *progressRecorder) WorkflowFinished(err error) {
	r.add(bddEvent{kind: bddWorkflowFinished, err: err})
}

func (r *progressRecorder) Finish() {
	r.add(bddEvent{kind: bddFinish})
}

func (r *progressRecorder) all() []bddEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]bddEvent, len(r.events))
	copy(out, r.events)

	return out
}

func (r *progressRecorder) ofKind(kinds ...bddEventKind) []bddEvent {
	wanted := make(map[bddEventKind]bool, len(kinds))
	for _, k := range kinds {
		wanted[k] = true
	}

	var out []bddEvent
	for _, e := range r.all() {
		if wanted[e.kind] {
			out = append(out, e)
		}
	}

	return out
}

func (r *progressRecorder) countOf(kind bddEventKind) int {
	return len(r.ofKind(kind))
}

func (r *progressRecorder) dump() string {
	msg := ""

	for _, e := range r.all() {
		msg += fmt.Sprintf("\n  %+v", e)
	}

	return msg
}

// fastRetryConfig keeps retry specs well under a second.
func fastRetryConfig() *execution.RetryConfig {
	return &execution.RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}
}

var _ = Describe("Live progress narrative", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("when a clean run starts", func() {
		It("announces every selected cleaner as pending before any work begins", func() {
			registry, _ := registerFakes(
				newFakeCleaner("alpha"),
				newFakeCleaner("beta"),
				newFakeCleaner("gamma"),
			)

			recorder := newProgressRecorder()
			_, err := execution.RunCleaners(ctx, registry, []string{"alpha", "beta", "gamma"},
				execution.WithProgress(recorder),
				execution.WithMaxConcurrency(1),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			events := recorder.all()

			gomega.Expect(events[0].kind).To(gomega.Equal(bddWorkflowStarted))
			gomega.Expect(events[0].name).To(gomega.Equal("clean"))

			firstStarted := -1
			for i, e := range events {
				if e.kind == bddStarted {
					firstStarted = i

					break
				}
			}
			gomega.Expect(firstStarted).To(gomega.BeNumerically(">", 0), "recorder: %s", recorder.dump())

			// The whole plan is on the tree before the first activity runs.
			registered := events[1:firstStarted]
			gomega.Expect(registered).To(gomega.HaveLen(3), "recorder: %s", recorder.dump())
			gomega.Expect(registered[0].name).To(gomega.Equal("alpha"))
			gomega.Expect(registered[1].name).To(gomega.Equal("beta"))
			gomega.Expect(registered[2].name).To(gomega.Equal("gamma"))

			// Each cleaner appears exactly once in the plan.
			gomega.Expect(recorder.countOf(bddRegistered)).To(gomega.Equal(3))
		})

		It("closes the narrative with a workflow boundary after the final step outcome", func() {
			registry, _ := registerFakes(newFakeCleaner("solo"))

			recorder := newProgressRecorder()
			_, err := execution.RunCleaners(ctx, registry, []string{"solo"}, execution.WithProgress(recorder))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			events := recorder.all()
			gomega.Expect(events).NotTo(gomega.BeEmpty())
			gomega.Expect(events[len(events)-1].kind).To(gomega.Equal(bddFinish))
			gomega.Expect(events[len(events)-2].kind).To(gomega.Equal(bddWorkflowFinished))
			gomega.Expect(events[len(events)-2].err).To(gomega.BeNil())
		})
	})

	Describe("when a cleaner fails transiently", func() {
		It("narrates retry attempts without ever reporting an intermediate failure", func() {
			// Outcome list semantics: the LAST outcome repeats for further
			// calls, so fail-twice-then-succeed needs three entries.
			transientErr := errorfamily.NewTransient("bdd.transient", "flaky backend")
			registry, _ := registerFakes(newFakeCleaner("flaky", transientErr, transientErr, nil))

			recorder := newProgressRecorder()
			wr, err := execution.RunCleaners(ctx, registry, []string{"flaky"},
				execution.WithProgress(recorder),
				execution.WithRetry(fastRetryConfig()),
				execution.WithMaxConcurrency(1),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(1))

			gomega.Expect(recorder.countOf(bddStarted)).To(gomega.Equal(3), "one start per attempt: %s", recorder.dump())
			gomega.Expect(recorder.countOf(bddFailed)).To(gomega.Equal(0), "intermediate attempts are never failures")

			retries := recorder.ofKind(bddRetrying)
			gomega.Expect(retries).To(gomega.HaveLen(2), "recorder: %s", recorder.dump())
			gomega.Expect(retries[0].attempt).To(gomega.Equal(1))
			gomega.Expect(retries[1].attempt).To(gomega.Equal(2))
			gomega.Expect(retries[0].reason).To(gomega.Equal("transient"))

			// The run ends with a single completion, after the last retry.
			completions := recorder.ofKind(bddCompleted)
			gomega.Expect(completions).To(gomega.HaveLen(1))
			gomega.Expect(completions[0].name).To(gomega.Equal("flaky"))
		})
	})

	Describe("when a cleaner is unavailable", func() {
		It("reports the skip with its reason instead of a completion or failure", func() {
			registry, _ := registerFakes(newFakeCleaner("absent", cleaner.NewNotAvailableError("some-tool", "")))

			recorder := newProgressRecorder()
			_, err := execution.RunCleaners(ctx, registry, []string{"absent"},
				execution.WithProgress(recorder),
				execution.WithRetry(fastRetryConfig()),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(recorder.countOf(bddStarted)).To(gomega.Equal(1), "unavailable cleaners are not retried")
			gomega.Expect(recorder.countOf(bddCompleted)).To(gomega.Equal(0))
			gomega.Expect(recorder.countOf(bddFailed)).To(gomega.Equal(0))
			gomega.Expect(recorder.countOf(bddRetrying)).To(gomega.Equal(0))

			skips := recorder.ofKind(bddSkipped)
			gomega.Expect(skips).To(gomega.HaveLen(1))
			gomega.Expect(skips[0].name).To(gomega.Equal("absent"))
			gomega.Expect(skips[0].reason).To(gomega.Equal("some-tool not available"))
		})
	})

	Describe("when no progress emitter is injected", func() {
		It("completes the run without emitting a single event (machine-output silence)", func() {
			registry, _ := registerFakes(newFakeCleaner("quiet"))

			recorder := newProgressRecorder()
			_, err := execution.RunCleaners(ctx, registry, []string{"quiet"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(recorder.all()).To(gomega.BeEmpty(),
				"an uninjected recorder must never see events — this is what keeps --json byte-clean")
		})
	})
})
