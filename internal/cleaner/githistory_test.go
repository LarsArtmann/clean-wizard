package cleaner

import (
	"context"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("GitHistoryCleaner", func() {
	var (
		cleaner *GitHistoryCleaner
		tempDir string
		ctx     context.Context
	)

	ginkgo.BeforeEach(func() {
		tempDir, _ = os.MkdirTemp("", "githistory-cleaner-test-*")
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	ginkgo.Describe("NewGitHistoryCleaner", func() {
		ginkgo.Context("with default options", func() {
			ginkgo.It("should create cleaner with default settings", func() {
				cleaner = NewGitHistoryCleaner()
				gomega.Expect(cleaner).NotTo(gomega.BeNil())
				gomega.Expect(cleaner.repoPath).To(gomega.Equal("."))
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(1))
				gomega.Expect(cleaner.maxFiles).To(gomega.Equal(100))
				gomega.Expect(cleaner.createBackup).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("with functional options", func() {
			ginkgo.It("should accept repo path option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryRepoPath(tempDir))
				gomega.Expect(cleaner.repoPath).To(gomega.Equal(tempDir))
			})

			ginkgo.It("should accept min size option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryMinSizeMB(10))
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(10))
			})

			ginkgo.It("should accept max files option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryMaxFiles(50))
				gomega.Expect(cleaner.maxFiles).To(gomega.Equal(50))
			})

			ginkgo.It("should accept create backup option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryCreateBackup(false))
				gomega.Expect(cleaner.createBackup).To(gomega.BeFalse())
			})

			ginkgo.It("should accept verbose option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryVerbose(true))
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
			})

			ginkgo.It("should accept dry run option", func() {
				cleaner = NewGitHistoryCleaner(WithGitHistoryDryRun(true))
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
			})

			ginkgo.It("should accept exclude extensions option", func() {
				exts := []string{".pdf", ".doc"}
				cleaner = NewGitHistoryCleaner(WithGitHistoryExcludeExtensions(exts))
				gomega.Expect(cleaner.excludeExts).To(gomega.Equal(exts))
			})

			ginkgo.It("should accept include extensions option", func() {
				exts := []string{".exe", ".dll"}
				cleaner = NewGitHistoryCleaner(WithGitHistoryIncludeExtensions(exts))
				gomega.Expect(cleaner.includeExts).To(gomega.Equal(exts))
			})

			ginkgo.It("should accept exclude paths option", func() {
				paths := []string{"vendor/", "third_party/"}
				cleaner = NewGitHistoryCleaner(WithGitHistoryExcludePaths(paths))
				gomega.Expect(cleaner.excludePaths).To(gomega.Equal(paths))
			})

			ginkgo.It("should accept selected files option", func() {
				files := []domain.GitHistoryFile{
					{Path: "binary.exe", SizeBytes: 5 * 1024 * 1024},
				}
				cleaner = NewGitHistoryCleaner(WithGitHistorySelectedFiles(files))
				gomega.Expect(cleaner.selectedFiles).To(gomega.HaveLen(1))
			})

			ginkgo.It("should accept multiple options", func() {
				cleaner = NewGitHistoryCleaner(
					WithGitHistoryRepoPath(tempDir),
					WithGitHistoryMinSizeMB(5),
					WithGitHistoryVerbose(true),
					WithGitHistoryDryRun(true),
				)
				gomega.Expect(cleaner.repoPath).To(gomega.Equal(tempDir))
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(5))
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Describe("Name and Type methods", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner()
		})

		ginkgo.It("should return correct name and type", func() {
			GinkgoAssertNameAndType(cleaner, "git-history", domain.OperationTypeGitHistory)
		})
	})

	ginkgo.Describe("IsAvailable", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner(WithGitHistoryRepoPath(tempDir))
		})

		ginkgo.It("should return false for non-git directory", func() {
			gomega.Expect(cleaner.IsAvailable(ctx)).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("ValidateSettings", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner()
		})

		ginkgo.Context("with nil settings", func() {
			ginkgo.It("should return nil", func() {
				err := cleaner.ValidateSettings(nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		GinkgoValidateEmptySettingsContext(cleaner, "should return nil when GitHistory is nil")

		ginkgo.Context("with valid settings", func() {
			ginkgo.It("should return nil for valid empty GitHistorySettings", func() {
				settings := &domain.OperationSettings{
					GitHistory: &domain.GitHistorySettings{},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return nil for valid min_size_mb", func() {
				settings := &domain.OperationSettings{
					GitHistory: &domain.GitHistorySettings{
						MinSizeMB: 10,
					},
				}
				GinkgoValidateValidSettingsTest(cleaner, settings)
			})

			ginkgo.It("should return nil for valid max_files", func() {
				settings := &domain.OperationSettings{
					GitHistory: &domain.GitHistorySettings{
						MaxFiles: 50,
					},
				}
				GinkgoValidateValidSettingsTest(cleaner, settings)
			})
		})

		ginkgo.Context("with invalid settings", func() {
			ginkgo.It("should return error for negative min_size_mb", func() {
				settings := &domain.OperationSettings{
					GitHistory: &domain.GitHistorySettings{
						MinSizeMB: -1,
					},
				}
				assertValidationError(cleaner, settings, "min_size_mb must be >= 0")
			})

			ginkgo.It("should return error for negative max_files", func() {
				settings := &domain.OperationSettings{
					GitHistory: &domain.GitHistorySettings{
						MaxFiles: -1,
					},
				}
				assertValidationError(cleaner, settings, "max_files must be >= 0")
			})
		})
	})

	ginkgo.Describe("Scan", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner(WithGitHistoryRepoPath(tempDir))
		})

		ginkgo.It("should return error for non-git directory", func() {
			result := cleaner.Scan(ctx)
			gomega.Expect(result.IsErr()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Clean", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner(
				WithGitHistoryRepoPath(tempDir),
				WithGitHistoryDryRun(true),
			)
		})

		ginkgo.Context("with no files selected", func() {
			ginkgo.It("should scan first and return error for non-git directory", func() {
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsErr()).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Describe("GetStoreSize", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner(WithGitHistoryRepoPath(tempDir))
		})

		ginkgo.It("should return 0 for non-git directory", func() {
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(0)))
		})
	})

	ginkgo.Describe("SetSelectedFiles", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewGitHistoryCleaner()
		})

		ginkgo.It("should set selected files", func() {
			files := []domain.GitHistoryFile{
				{Path: "binary.exe", SizeBytes: 5 * 1024 * 1024},
			}
			cleaner.SetSelectedFiles(files)
			gomega.Expect(cleaner.selectedFiles).To(gomega.HaveLen(1))
		})
	})

	ginkgo.Describe("FindGitRepositories", func() {
		ginkgo.It("should return empty slice for empty directory", func() {
			emptyDir := tempDir + "/empty"
			_ = os.MkdirAll(emptyDir, 0o755)
			repos, err := FindGitRepositories(emptyDir, 3)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).To(gomega.BeEmpty())
		})

		ginkgo.It("should return error for non-existent directory", func() {
			_, err := FindGitRepositories("/non/existent/path", 3)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.It("should skip non-directory entries", func() {
			// Create a file in the temp dir
			_ = os.WriteFile(tempDir+"/somefile.txt", []byte("test"), 0o644)
			repos, err := FindGitRepositories(tempDir, 1)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).To(gomega.BeEmpty())
		})
	})
})
