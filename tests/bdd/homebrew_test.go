package bdd

import (
	"context"
	"os/exec"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Homebrew cleaner", func() {
	var ctx context.Context

	brewInstalled := func() bool {
		_, lookErr := exec.LookPath("brew")
		return lookErr == nil
	}

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("identity", func() {
		ginkgo.It("exposes the homebrew cleaner name and operation type", func() {
			hbc := cleaner.NewHomebrewCleaner(true, true, domain.HomebrewModeAll)

			gomega.Expect(hbc.Name()).To(gomega.Equal("homebrew"))
			gomega.Expect(hbc.Type()).To(gomega.Equal(domain.OperationTypeHomebrew))
		})
	})

	ginkgo.Describe("availability", func() {
		ginkgo.It("reports availability that matches the brew binary", func() {
			hbc := cleaner.NewHomebrewCleaner(true, true, domain.HomebrewModeAll)

			gomega.Expect(hbc.IsAvailable(ctx)).To(gomega.Equal(brewInstalled()))
		})
	})

	ginkgo.Describe("cleaning", func() {
		ginkgo.Context("when brew is not installed", func() {
			ginkgo.BeforeEach(func() {
				if brewInstalled() {
					ginkgo.Skip("brew is installed; unavailable-behavior specs require it to be absent")
				}
			})

			ginkgo.It("refuses to clean with an infrastructure error", func() {
				hbc := cleaner.NewHomebrewCleaner(true, true, domain.HomebrewModeAll)

				cleanRes := hbc.Clean(ctx)
				gomega.Expect(cleanRes.IsErr()).To(gomega.BeTrue())
				gomega.Expect(errorfamily.Classify(cleanRes.Error())).To(gomega.Equal(errorfamily.Infrastructure))
			})
		})

		ginkgo.Context("when brew is installed", func() {
			ginkgo.BeforeEach(func() {
				if !brewInstalled() {
					ginkgo.Skip("brew is not installed")
				}
			})

			ginkgo.It("completes a dry run without reporting missing tooling", func() {
				hbc := cleaner.NewHomebrewCleaner(true, true, domain.HomebrewModeAll)

				cleanRes := hbc.Clean(ctx)
				if cleanRes.IsErr() {
					gomega.Expect(errorfamily.Classify(cleanRes.Error())).NotTo(gomega.Equal(errorfamily.Infrastructure))
				}
			})
		})
	})

	ginkgo.Describe("settings validation", func() {
		ginkgo.DescribeTable("accepts or rejects homebrew modes",
			func(unusedOnly domain.HomebrewMode, valid bool) {
				hbc := cleaner.NewHomebrewCleaner(true, true, unusedOnly)

				settings := &domain.OperationSettings{ //nolint:exhaustruct
					Homebrew: &domain.HomebrewSettings{UnusedOnly: unusedOnly}, //nolint:exhaustruct
				}

				err := hbc.ValidateSettings(settings)
				if valid {
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
				} else {
					gomega.Expect(err).To(gomega.HaveOccurred())
				}
			},
			ginkgo.Entry("all packages", domain.HomebrewModeAll, true),
			ginkgo.Entry("unused only", domain.HomebrewModeUnusedOnly, true),
			ginkgo.Entry("unknown mode", domain.HomebrewMode(42), false),
		)
	})
})
