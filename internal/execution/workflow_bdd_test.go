package execution_test

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Clean wizard workflow execution", func() {
	var ctx context.Context

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("running selected cleaners", func() {
		ginkgo.It("reports succeeded steps with aggregated totals", func() {
			registry := registerFakes(
				newFakeCleaner("alpha"),
				newFakeCleaner("beta"),
			)

			wr, err := execution.RunCleaners(ctx, registry, []string{"alpha", "beta"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(2))
			gomega.Expect(wr.Failed()).To(gomega.BeEmpty())
			gomega.Expect(wr.Skipped()).To(gomega.BeEmpty())
			gomega.Expect(wr.TotalBytesFreed).To(gomega.Equal(uint64(200)))
			gomega.Expect(wr.TotalItemsRemoved).To(gomega.Equal(uint(2)))
		})

		ginkgo.It("reports steps in selection order regardless of completion time", func() {
			slow := newFakeCleaner("slow-cleaner")
			slow.cleanDelay = 50 * time.Millisecond
			fast := newFakeCleaner("fast-cleaner")

			registry := registerFakes(slow, fast)

			wr, err := execution.RunCleaners(ctx, registry, []string{"slow-cleaner", "fast-cleaner"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			names := make([]string, 0, len(wr.Steps))
			for _, step := range wr.Steps {
				names = append(names, step.Name)
			}

			gomega.Expect(names).To(gomega.Equal([]string{"slow-cleaner", "fast-cleaner"}))
		})

		ginkgo.It("classifies unavailable cleaners as skipped and others as failed", func() {
			notAvailableErr := cleaner.NewNotAvailableError("some-tool", "")
			registry := registerFakes(
				newFakeCleaner("healthy"),
				newFakeCleaner("broken", assertError{msg: "disk exploded"}),
				newFakeCleaner("absent", notAvailableErr),
			)

			wr, err := execution.RunCleaners(ctx, registry, []string{"healthy", "broken", "absent"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(1))
			gomega.Expect(wr.Failed()).To(gomega.HaveLen(1))
			gomega.Expect(wr.Skipped()).To(gomega.HaveLen(1))
			gomega.Expect(wr.TotalBytesFreed).To(gomega.Equal(uint64(100)))
		})

		ginkgo.It("records a panicking cleaner as a failed step without aborting the workflow", func() {
			registry := cleaner.NewRegistry()
			registry.Register("panics", &panickingCleaner{name: "panics"})
			registry.Register("healthy", newFakeCleaner("healthy"))

			wr, err := execution.RunCleaners(ctx, registry, []string{"panics", "healthy"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(wr.Failed()).To(gomega.HaveLen(1))
			gomega.Expect(wr.Failed()[0].Name).To(gomega.Equal("panics"))
			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(1))
		})

		ginkgo.It("rejects a selection that references an unknown cleaner", func() {
			registry := registerFakes(newFakeCleaner("known"))

			_, err := execution.RunCleaners(ctx, registry, []string{"does-not-exist"})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("not found in registry"))
			errorfamilytest.AssertFamily(ginkgo.GinkgoTB(), err, errorfamily.Rejection)
		})

		ginkgo.It("completes with no steps when the selection is empty", func() {
			registry := registerFakes(newFakeCleaner("idle"))

			wr, err := execution.RunCleaners(ctx, registry, []string{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(wr.Steps).To(gomega.BeEmpty())
		})
	})

	ginkgo.Describe("scanning selected cleaners", func() {
		ginkgo.It("reports scan totals per cleaner", func() {
			registry := registerFakes(newFakeCleaner("scanner"))

			wr, err := execution.RunScans(ctx, registry, []string{"scanner"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(1))
			gomega.Expect(wr.TotalBytesFreed).To(gomega.Equal(uint64(300)))
			gomega.Expect(wr.TotalItemsRemoved).To(gomega.Equal(uint(2)))
		})
	})

	ginkgo.Describe("concurrency limits", func() {
		ginkgo.It("runs at most one cleaner at a time when capped to one", func() {
			cleaners := make([]*fakeCleaner, 0, 3)
			for _, name := range []string{"one", "two", "three"} {
				c := newFakeCleaner(name)
				c.cleanDelay = 30 * time.Millisecond
				cleaners = append(cleaners, c)
			}

			registry := registerFakes(cleaners...)

			_, err := execution.RunCleaners(ctx, registry,
				[]string{"one", "two", "three"},
				execution.WithMaxConcurrency(1),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			for _, c := range cleaners {
				gomega.Expect(c.peak()).To(gomega.Equal(int32(1)))
			}
		})

		ginkgo.It("runs at most two cleaners at a time when capped to two", func() {
			cleaners := make([]*fakeCleaner, 0, 4)
			for _, name := range []string{"a", "b", "c", "d"} {
				c := newFakeCleaner(name)
				c.cleanDelay = 60 * time.Millisecond
				cleaners = append(cleaners, c)
			}

			registry := registerFakes(cleaners...)

			_, err := execution.RunCleaners(ctx, registry,
				[]string{"a", "b", "c", "d"},
				execution.WithMaxConcurrency(2),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			for _, c := range cleaners {
				gomega.Expect(c.peak()).To(gomega.BeNumerically("<=", 2))
			}

			gomega.Expect(cleaners[0].peak()).To(gomega.Equal(int32(2)))
		})
	})
})

// assertError is a plain error double with a stable message.
type assertError struct{ msg string }

func (e assertError) Error() string { return e.msg }

// panickingCleaner simulates a cleaner whose Clean call panics.
type panickingCleaner struct {
	name string
}

func (p *panickingCleaner) Name() string                      { return p.name }
func (p *panickingCleaner) Type() domain.OperationType        { return domain.OperationTypeCargoPackages }
func (p *panickingCleaner) IsAvailable(_ context.Context) bool { return true }

func (p *panickingCleaner) Clean(_ context.Context) result.Result[domain.CleanResult] {
	panic("cleaner exploded")
}

func (p *panickingCleaner) Scan(_ context.Context) result.Result[[]domain.ScanItem] {
	return result.Ok([]domain.ScanItem{}) //nolint:exhaustruct
}
