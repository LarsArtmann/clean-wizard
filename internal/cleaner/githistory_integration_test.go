package cleaner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("GitHistory Integration", func() {
	var (
		tempRepoDir string
		ctx         context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		tempRepoDir, err = os.MkdirTemp("", "githistory-integration-test-*")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		if tempRepoDir != "" {
			_ = os.RemoveAll(tempRepoDir)
		}
	})

	ginkgo.Describe("with real git repository", func() {
		ginkgo.BeforeEach(func() {
			// Initialize git repo
			runGitCommand(tempRepoDir, "init")
			runGitCommand(tempRepoDir, "config", "user.email", "test@example.com")
			runGitCommand(tempRepoDir, "config", "user.name", "Test User")
			runGitCommand(tempRepoDir, "config", "commit.gpgsign", "false")
		})

		ginkgo.Context("SafetyChecker", func() {
			ginkgo.It("should detect git repository", func() {
				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.IsGitRepo).To(gomega.BeTrue())
			})

			ginkgo.It("should detect uncommitted changes", func() {
				// Create uncommitted file
				_ = os.WriteFile(filepath.Join(tempRepoDir, "uncommitted.txt"), []byte("test"), 0o644)

				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasUncommittedChanges).To(gomega.BeTrue())
			})

			ginkgo.It("should not detect uncommitted changes when clean", func() {
				// Make initial commit
				_ = os.WriteFile(filepath.Join(tempRepoDir, "README.md"), []byte("# Test"), 0o644)
				runGitCommand(tempRepoDir, "add", ".")
				runGitCommand(tempRepoDir, "commit", "-m", "initial")

				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasUncommittedChanges).To(gomega.BeFalse())
			})

			ginkgo.It("should detect main branch", func() {
				// Make initial commit on main
				_ = os.WriteFile(filepath.Join(tempRepoDir, "README.md"), []byte("# Test"), 0o644)
				runGitCommand(tempRepoDir, "add", ".")
				runGitCommand(tempRepoDir, "commit", "-m", "initial")

				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.CurrentBranch).To(gomega.Equal("master"))
				gomega.Expect(report.IsProtectedBranch).To(gomega.BeTrue())
			})

			ginkgo.It("should detect LFS when .gitattributes has LFS filter", func() {
				// Create .gitattributes with LFS config
				_ = os.WriteFile(filepath.Join(tempRepoDir, ".gitattributes"),
					[]byte("*.bin filter=lfs diff=lfs merge=lfs -text"), 0o644)

				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasLFS).To(gomega.BeTrue())
			})

			ginkgo.It("should not detect LFS when no .gitattributes", func() {
				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasLFS).To(gomega.BeFalse())
			})

			ginkgo.It("should detect submodules when .gitmodules exists", func() {
				// Create .gitmodules
				_ = os.WriteFile(filepath.Join(tempRepoDir, ".gitmodules"),
					[]byte("[submodule \"test\"]\npath = test\nurl = https://example.com/test.git"), 0o644)

				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasSubmodules).To(gomega.BeTrue())
			})

			ginkgo.It("should not detect submodules when no .gitmodules", func() {
				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				gomega.Expect(report.HasSubmodules).To(gomega.BeFalse())
			})

			ginkgo.It("should check disk space", func() {
				checker := NewGitHistorySafetyChecker(tempRepoDir, false)
				report := checker.Check(ctx)
				// On most systems, there should be at least 1GB free
				gomega.Expect(report.HasSufficientDiskSpace).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("Scanner", func() {
			ginkgo.BeforeEach(func() {
				// Make initial commit
				_ = os.WriteFile(filepath.Join(tempRepoDir, "README.md"), []byte("# Test"), 0o644)
				runGitCommand(tempRepoDir, "add", ".")
				runGitCommand(tempRepoDir, "commit", "-m", "initial")
			})

			ginkgo.It("should scan repository with no binary files", func() {
				scanner := NewGitHistoryScanner(tempRepoDir)
				result, err := scanner.Scan(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Files).To(gomega.BeEmpty())
			})

			ginkgo.It("should detect binary file in history", func() {
				// Create and commit a binary file
				binaryData := make([]byte, 2*1024*1024) // 2MB
				for i := range binaryData {
					binaryData[i] = byte(i % 256)
				}

				_ = os.WriteFile(filepath.Join(tempRepoDir, "large.bin"), binaryData, 0o644)
				runGitCommand(tempRepoDir, "add", "large.bin")
				runGitCommand(tempRepoDir, "commit", "-m", "add binary")

				// Remove file but it's still in history
				_ = os.Remove(filepath.Join(tempRepoDir, "large.bin"))
				runGitCommand(tempRepoDir, "add", "-A")
				runGitCommand(tempRepoDir, "commit", "-m", "remove binary")

				scanner := NewGitHistoryScanner(tempRepoDir, WithMinSizeMB(1))
				result, err := scanner.Scan(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Files).NotTo(gomega.BeEmpty())

				// Find the binary file in results
				found := false
				for _, f := range result.Files {
					if f.Path == "large.bin" {
						found = true
						gomega.Expect(f.SizeBytes).To(gomega.BeNumerically(">=", int64(2*1024*1024)))
						break
					}
				}
				gomega.Expect(found).To(gomega.BeTrue(), "large.bin should be found in scan results")
			})
		})

		ginkgo.Context("Cleaner", func() {
			ginkgo.BeforeEach(func() {
				// Make initial commit
				_ = os.WriteFile(filepath.Join(tempRepoDir, "README.md"), []byte("# Test"), 0o644)
				runGitCommand(tempRepoDir, "add", ".")
				runGitCommand(tempRepoDir, "commit", "-m", "initial")
			})

			ginkgo.It("should return empty result when no files to clean", func() {
				cleaner := NewGitHistoryCleaner(
					WithGitHistoryRepoPath(tempRepoDir),
					WithGitHistoryDryRun(true),
				)

				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult, err := result.Unwrap()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
			})

			ginkgo.It("should detect available git repository", func() {
				cleaner := NewGitHistoryCleaner(WithGitHistoryRepoPath(tempRepoDir))
				gomega.Expect(cleaner.IsAvailable(ctx)).To(gomega.BeTrue())
			})

			ginkgo.It("should get safety report", func() {
				cleaner := NewGitHistoryCleaner(WithGitHistoryRepoPath(tempRepoDir))
				report := cleaner.GetSafetyReport(ctx)
				gomega.Expect(report).NotTo(gomega.BeNil())
				gomega.Expect(report.IsGitRepo).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Describe("FindGitRepositories", func() {
		ginkgo.It("should find nested git repositories", func() {
			// Create nested structure
			repo1 := filepath.Join(tempRepoDir, "project1")
			repo2 := filepath.Join(tempRepoDir, "projects", "project2")

			_ = os.MkdirAll(repo1, 0o755)
			_ = os.MkdirAll(repo2, 0o755)

			// Initialize git repos
			runGitCommand(repo1, "init")
			runGitCommand(repo2, "init")

			repos, err := FindGitRepositories(tempRepoDir, 3)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).To(gomega.ContainElements(repo1, repo2))
		})

		ginkgo.It("should respect max depth", func() {
			// Create deeply nested repo
			deepRepo := filepath.Join(tempRepoDir, "a", "b", "c", "d", "repo")
			_ = os.MkdirAll(deepRepo, 0o755)
			runGitCommand(deepRepo, "init")

			// With depth 3, should not find the repo
			repos, err := FindGitRepositories(tempRepoDir, 3)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).NotTo(gomega.ContainElement(deepRepo))

			// With depth 5, should find the repo
			repos, err = FindGitRepositories(tempRepoDir, 5)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).To(gomega.ContainElement(deepRepo))
		})

		ginkgo.It("should not recurse into git repositories", func() {
			// Create repo with nested directory (but not a git repo)
			repo := filepath.Join(tempRepoDir, "main-repo")
			_ = os.MkdirAll(filepath.Join(repo, "subdir"), 0o755)
			runGitCommand(repo, "init")

			// Should find only the main repo, not recurse inside
			repos, err := FindGitRepositories(tempRepoDir, 3)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(repos).To(gomega.HaveLen(1))
			gomega.Expect(repos[0]).To(gomega.Equal(repo))
		})
	})

	ginkgo.Describe("getDefaultBackupPath", func() {
		ginkgo.It("should return backup path with correct suffix", func() {
			backupPath := getDefaultBackupPath(tempRepoDir)
			gomega.Expect(backupPath).To(gomega.ContainSubstring("-backup.git"))
		})

		ginkgo.It("should be in parent directory", func() {
			backupPath := getDefaultBackupPath(tempRepoDir)
			gomega.Expect(filepath.Dir(backupPath)).To(gomega.Equal(filepath.Dir(tempRepoDir)))
		})
	})
})

// runGitCommand executes a git command in the specified directory.
func runGitCommand(dir string, args ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2024-01-01T00:00:00", "GIT_COMMITTER_DATE=2024-01-01T00:00:00")

	output, err := cmd.CombinedOutput()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "git command failed: %s", string(output))
}
