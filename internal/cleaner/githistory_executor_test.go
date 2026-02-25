package cleaner

import (
	"context"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("GitHistoryExecutor", func() {
	var (
		executor *GitHistoryExecutor
		tempDir  string
		ctx      context.Context
	)

	ginkgo.BeforeEach(func() {
		tempDir, _ = os.MkdirTemp("", "githistory-executor-test-*")
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	ginkgo.Describe("NewGitHistoryExecutor", func() {
		ginkgo.It("should create executor with given path", func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
			gomega.Expect(executor).NotTo(gomega.BeNil())
			gomega.Expect(executor.repoPath).To(gomega.Equal(tempDir))
		})

		ginkgo.It("should set verbose flag", func() {
			executor = NewGitHistoryExecutor(tempDir, true, false)
			gomega.Expect(executor.verbose).To(gomega.BeTrue())
		})

		ginkgo.It("should set dry run flag", func() {
			executor = NewGitHistoryExecutor(tempDir, false, true)
			gomega.Expect(executor.dryRun).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Execute", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, true) // dry run mode
		})

		ginkgo.Context("with no files to remove", func() {
			ginkgo.It("should return error", func() {
				_, err := executor.Execute(
					ctx,
					ExecuteOptions{FilesToRemove: []domain.GitHistoryFile{}},
				)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("no files to remove"))
			})
		})

		ginkgo.Context("in dry run mode", func() {
			ginkgo.It("should return result without executing", func() {
				files := []domain.GitHistoryFile{
					{Path: "binary.exe", SizeBytes: 5 * 1024 * 1024},
				}
				result, err := executor.Execute(ctx, ExecuteOptions{
					FilesToRemove: files,
					CreateBackup:  false,
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.FilesRemoved).To(gomega.HaveLen(1))
				gomega.Expect(result.BackupCreated).To(gomega.BeFalse())
			})

			ginkgo.It("should calculate bytes removed correctly", func() {
				files := []domain.GitHistoryFile{
					{Path: "binary1.exe", SizeBytes: 5 * 1024 * 1024},
					{Path: "binary2.dll", SizeBytes: 3 * 1024 * 1024},
				}
				result, err := executor.Execute(ctx, ExecuteOptions{
					FilesToRemove: files,
					CreateBackup:  false,
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.BytesRemoved).To(gomega.Equal(int64(8 * 1024 * 1024)))
			})
		})
	})

	ginkgo.Describe("calculateTotalSize", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
		})

		ginkgo.It("should sum file sizes correctly", func() {
			files := []domain.GitHistoryFile{
				{Path: "a.exe", SizeBytes: 1000},
				{Path: "b.dll", SizeBytes: 2000},
				{Path: "c.so", SizeBytes: 3000},
			}
			total := executor.calculateTotalSize(files)
			gomega.Expect(total).To(gomega.Equal(int64(6000)))
		})

		ginkgo.It("should return 0 for empty slice", func() {
			files := []domain.GitHistoryFile{}
			total := executor.calculateTotalSize(files)
			gomega.Expect(total).To(gomega.Equal(int64(0)))
		})
	})

	ginkgo.Describe("getDefaultBackupPath", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
		})

		ginkgo.It("should return path with -backup.git suffix", func() {
			path := executor.getDefaultBackupPath()
			gomega.Expect(path).To(gomega.ContainSubstring("-backup.git"))
		})
	})

	ginkgo.Describe("getRepoSize", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
		})

		ginkgo.It("should return 0 for non-existent .git directory", func() {
			size, err := executor.getRepoSize()
			// filepath.Walk returns nil error for non-existent paths, just walks nothing
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(size).To(gomega.Equal(int64(0)))
		})

		ginkgo.It("should calculate size of .git directory", func() {
			// Create a mock .git directory with a file
			gitDir := filepath.Join(tempDir, ".git")
			_ = os.MkdirAll(gitDir, 0o755)
			_ = os.WriteFile(filepath.Join(gitDir, "config"), make([]byte, 100), 0o644)
			_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), make([]byte, 50), 0o644)

			size, err := executor.getRepoSize()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(size).To(gomega.Equal(int64(150)))
		})
	})

	ginkgo.Describe("parseCommitCount", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
		})

		ginkgo.It("should parse commit count from output", func() {
			output := "Processed 150 commits in 2.3 seconds"
			count := executor.parseCommitCount(output)
			gomega.Expect(count).To(gomega.Equal(150))
		})

		ginkgo.It("should return 0 for no match", func() {
			output := "No commits processed"
			count := executor.parseCommitCount(output)
			gomega.Expect(count).To(gomega.Equal(0))
		})

		ginkgo.It("should handle empty output", func() {
			count := executor.parseCommitCount("")
			gomega.Expect(count).To(gomega.Equal(0))
		})
	})

	ginkgo.Describe("EstimateImpact", func() {
		ginkgo.BeforeEach(func() {
			executor = NewGitHistoryExecutor(tempDir, false, false)
		})

		ginkgo.It("should work with empty .git directory", func() {
			files := []domain.GitHistoryFile{
				{Path: "binary.exe", SizeBytes: 5 * 1024 * 1024},
			}
			// getRepoSize returns 0, nil for non-existent .git directory
			estimate, err := executor.EstimateImpact(ctx, files)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(estimate).NotTo(gomega.BeNil())
			gomega.Expect(estimate.CurrentRepoSizeMB).To(gomega.Equal(float64(0)))
		})
	})

	ginkgo.Describe("GetFilesToRemoveFromSelection", func() {
		ginkgo.It("should return selected files", func() {
			allFiles := []domain.GitHistoryFile{
				{Path: "a.exe"},
				{Path: "b.dll"},
				{Path: "c.so"},
			}
			selected := GetFilesToRemoveFromSelection(allFiles, []int{0, 2})
			gomega.Expect(selected).To(gomega.HaveLen(2))
			gomega.Expect(selected[0].Path).To(gomega.Equal("a.exe"))
			gomega.Expect(selected[1].Path).To(gomega.Equal("c.so"))
		})

		ginkgo.It("should handle empty selection", func() {
			allFiles := []domain.GitHistoryFile{
				{Path: "a.exe"},
			}
			selected := GetFilesToRemoveFromSelection(allFiles, []int{})
			gomega.Expect(selected).To(gomega.BeEmpty())
		})

		ginkgo.It("should handle invalid indices", func() {
			allFiles := []domain.GitHistoryFile{
				{Path: "a.exe"},
			}
			selected := GetFilesToRemoveFromSelection(allFiles, []int{-1, 5, 100})
			gomega.Expect(selected).To(gomega.BeEmpty())
		})
	})

	ginkgo.Describe("GetUniquePaths", func() {
		ginkgo.It("should return unique paths sorted", func() {
			files := []domain.GitHistoryFile{
				{Path: "b.exe"},
				{Path: "a.dll"},
				{Path: "b.exe"}, // duplicate
				{Path: "c.so"},
			}
			paths := GetUniquePaths(files)
			gomega.Expect(paths).To(gomega.Equal([]string{"a.dll", "b.exe", "c.so"}))
		})

		ginkgo.It("should handle empty slice", func() {
			paths := GetUniquePaths([]domain.GitHistoryFile{})
			gomega.Expect(paths).To(gomega.BeEmpty())
		})
	})
})
