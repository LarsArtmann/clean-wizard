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

var _ = ginkgo.Describe("Go cleaner", func() {
	var ctx context.Context

	goInstalled := func() bool {
		_, lookErr := exec.LookPath("go")
		return lookErr == nil
	}

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("identity", func() {
		ginkgo.It("exposes the go cleaner name and operation type", func() {
			gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(gc.Name()).To(gomega.Equal("go"))
			gomega.Expect(gc.Type()).To(gomega.Equal(domain.OperationTypeGoPackages))
		})
	})

	ginkgo.Describe("availability", func() {
		ginkgo.It("reports availability that matches the go binary", func() {
			gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(gc.IsAvailable(ctx)).To(gomega.Equal(goInstalled()))
		})
	})

	ginkgo.Describe("cleaning", func() {
		ginkgo.Context("when go is not installed", func() {
			ginkgo.BeforeEach(func() {
				if goInstalled() {
					ginkgo.Skip("go is installed; unavailable-behavior specs require it to be absent")
				}
			})

			ginkgo.It("refuses to clean with an infrastructure error", func() {
				gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				cleanRes := gc.Clean(ctx)
				gomega.Expect(cleanRes.IsErr()).To(gomega.BeTrue())
				gomega.Expect(errorfamily.Classify(cleanRes.Error())).To(gomega.Equal(errorfamily.Infrastructure))
			})
		})

		ginkgo.Context("when go is installed", func() {
			ginkgo.BeforeEach(func() {
				if !goInstalled() {
					ginkgo.Skip("go is not installed")
				}
			})

			ginkgo.It("performs a dry run without deleting caches", func() {
				gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE|cleaner.GoCacheTestCache)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				cleanRes := gc.Clean(ctx)
				gomega.Expect(cleanRes.IsOk()).To(gomega.BeTrue())
			})

			ginkgo.It("scans caches without deleting them", func() {
				gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				scanRes := gc.Scan(ctx)
				gomega.Expect(scanRes.IsOk()).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Describe("settings validation", func() {
		ginkgo.It("accepts a fully specified go_packages section", func() {
			gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			settings := &domain.OperationSettings{ //nolint:exhaustruct
				GoPackages: &domain.GoPackagesSettings{ //nolint:exhaustruct
					CleanCache:      domain.CacheCleanupEnabled,
					CleanTestCache:  domain.CacheCleanupDisabled,
					CleanModCache:   domain.CacheCleanupEnabled,
					CleanBuildCache: domain.CacheCleanupDisabled,
					CleanLintCache:  domain.CacheCleanupDisabled,
				},
			}

			gomega.Expect(gc.ValidateSettings(settings)).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("rejects an invalid cache cleanup mode", func() {
			gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			settings := &domain.OperationSettings{ //nolint:exhaustruct
				GoPackages: &domain.GoPackagesSettings{ //nolint:exhaustruct
					CleanCache: domain.CacheCleanupMode(42),
				},
			}

			gomega.Expect(gc.ValidateSettings(settings)).To(gomega.HaveOccurred())
		})

		ginkgo.It("accepts settings without a go_packages section", func() {
			gc, err := cleaner.NewGoCleaner(true, true, cleaner.GoCacheGOCACHE)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(gc.ValidateSettings(&domain.OperationSettings{})).NotTo(gomega.HaveOccurred()) //nolint:exhaustruct
		})
	})
})
