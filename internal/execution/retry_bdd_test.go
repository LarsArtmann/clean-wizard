package execution_test

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Workflow step retries", func() {
	var ctx context.Context

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("when retries are enabled", func() {
		ginkgo.It("retries transient failures until the step succeeds", func() {
			transient := assertError{msg: "temporary I/O hiccup"}
			flaky := newFakeCleaner("flaky", transient, transient, nil)

			registry := registerFakes(flaky)

			retry := execution.RetryConfigFromAttempts(3)
			wr, err := execution.RunCleaners(ctx, registry, []string{"flaky"}, execution.WithRetry(retry))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(flaky.callCount()).To(gomega.Equal(int32(3)))
			gomega.Expect(wr.Succeeded()).To(gomega.HaveLen(1))
			gomega.Expect(wr.Steps).To(gomega.HaveLen(1))
		})

		ginkgo.It("fails the step when every attempt fails", func() {
			transient := assertError{msg: "still broken"}
			hopeless := newFakeCleaner("hopeless", transient)

			registry := registerFakes(hopeless)

			retry := execution.RetryConfigFromAttempts(3)
			wr, err := execution.RunCleaners(ctx, registry, []string{"hopeless"}, execution.WithRetry(retry))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(hopeless.callCount()).To(gomega.Equal(int32(3)))
			gomega.Expect(wr.Failed()).To(gomega.HaveLen(1))
			gomega.Expect(wr.Steps).To(gomega.HaveLen(1))
		})

		ginkgo.It("never retries an unavailable cleaner", func() {
			notAvailableErr := cleaner.NewNotAvailableError("some-tool", "")
			absent := newFakeCleaner("absent", notAvailableErr)

			registry := registerFakes(absent)

			retry := execution.RetryConfigFromAttempts(3)
			wr, err := execution.RunCleaners(ctx, registry, []string{"absent"}, execution.WithRetry(retry))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(absent.callCount()).To(gomega.Equal(int32(1)))
			gomega.Expect(wr.Skipped()).To(gomega.HaveLen(1))
			errorfamilytest.AssertFamily(ginkgo.GinkgoTB(), wr.Skipped()[0].Err, errorfamily.Infrastructure)
		})
	})

	ginkgo.Describe("when retries are disabled", func() {
		ginkgo.It("attempts each step exactly once", func() {
			transient := assertError{msg: "one shot only"}
			flaky := newFakeCleaner("flaky", transient)

			registry := registerFakes(flaky)

			wr, err := execution.RunCleaners(ctx, registry, []string{"flaky"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(flaky.callCount()).To(gomega.Equal(int32(1)))
			gomega.Expect(wr.Failed()).To(gomega.HaveLen(1))
		})
	})
})
