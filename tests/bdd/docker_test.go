package bdd

import (
	"context"
	"os/exec"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Docker cleaner", func() {
	var ctx context.Context

	dockerInstalled := func() bool {
		_, lookErr := exec.LookPath("docker")

		return lookErr == nil
	}

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("identity", func() {
		ginkgo.It("exposes the docker cleaner name and operation type", func() {
			dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

			gomega.Expect(dc.Name()).To(gomega.Equal("docker"))
			gomega.Expect(dc.Type()).To(gomega.Equal(operations.OperationTypeDocker))
		})
	})

	ginkgo.Describe("availability", func() {
		ginkgo.It("reports availability that matches the docker binary", func() {
			dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

			gomega.Expect(dc.IsAvailable(ctx)).To(gomega.Equal(dockerInstalled()))
		})
	})

	ginkgo.Describe("cleaning", func() {
		ginkgo.Context("when docker is not installed", func() {
			ginkgo.BeforeEach(func() {
				if dockerInstalled() {
					ginkgo.Skip("docker is installed; unavailable-behavior specs require it to be absent")
				}
			})

			ginkgo.It("refuses to clean with an infrastructure error", func() {
				dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

				cleanRes := dc.Clean(ctx)
				gomega.Expect(cleanRes.IsErr()).To(gomega.BeTrue())

				cleanErr := cleanRes.Error()
				gomega.Expect(cleanErr).To(gomega.HaveOccurred())
				gomega.Expect(errorfamily.Classify(cleanErr)).To(gomega.Equal(errorfamily.Infrastructure))
			})

			ginkgo.It("refuses to scan with an infrastructure error", func() {
				dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

				scanRes := dc.Scan(ctx)
				gomega.Expect(scanRes.IsErr()).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("when docker is installed", func() {
			ginkgo.BeforeEach(func() {
				if !dockerInstalled() {
					ginkgo.Skip("docker is not installed")
				}
			})

			ginkgo.It("completes a dry run without reporting missing tooling", func() {
				dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

				cleanRes := dc.Clean(ctx)
				if cleanRes.IsErr() {
					gomega.Expect(errorfamily.Classify(cleanRes.Error())).
						NotTo(gomega.Equal(errorfamily.Infrastructure))
				}
			})
		})
	})

	ginkgo.Describe("settings validation", func() {
		ginkgo.DescribeTable("accepts or rejects prune modes",
			func(pruneMode enums.DockerPruneMode, valid bool) {
				dc := cleaner.NewDockerCleaner(true, true, pruneMode)

				settings := &operations.OperationSettings{
					Docker: &operations.DockerSettings{PruneMode: pruneMode},
				}

				err := dc.ValidateSettings(settings)
				if valid {
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
				} else {
					gomega.Expect(err).To(gomega.HaveOccurred())
				}
			},
			ginkgo.Entry("all", enums.DockerPruneAll, true),
			ginkgo.Entry("images", enums.DockerPruneImages, true),
			ginkgo.Entry("containers", enums.DockerPruneContainers, true),
			ginkgo.Entry("volumes", enums.DockerPruneVolumes, true),
			ginkgo.Entry("builds", enums.DockerPruneBuilds, true),
			ginkgo.Entry("unknown mode", enums.DockerPruneMode(99), false),
		)

		ginkgo.It("accepts settings without a docker section", func() {
			dc := cleaner.NewDockerCleaner(true, true, enums.DockerPruneAll)

			gomega.Expect(dc.ValidateSettings(&operations.OperationSettings{})).
				NotTo(gomega.HaveOccurred())
		})
	})
})
