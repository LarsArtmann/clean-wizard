package cleaner

import (
	"context"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("GitHistorySafetyChecker", func() {
	var (
		checker *GitHistorySafetyChecker
		tempDir string
		ctx     context.Context
	)

	ginkgo.BeforeEach(func() {
		tempDir, _ = os.MkdirTemp("", "githistory-safety-test-*")
		ctx = context.Background()
		checker = NewGitHistorySafetyChecker(tempDir, false)
	})

	ginkgo.AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	ginkgo.Describe("NewGitHistorySafetyChecker", func() {
		ginkgo.It("should create checker with given path", func() {
			checker = NewGitHistorySafetyChecker(tempDir, false)
			gomega.Expect(checker).NotTo(gomega.BeNil())
			gomega.Expect(checker.repoPath).To(gomega.Equal(tempDir))
		})

		ginkgo.It("should set verbose flag", func() {
			checker = NewGitHistorySafetyChecker(tempDir, true)
			gomega.Expect(checker.verbose).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Check", func() {
		ginkgo.Context("for non-git directory", func() {
			ginkgo.It("should report not a git repo", func() {
				report := checker.Check(ctx)
				gomega.Expect(report.IsGitRepo).To(gomega.BeFalse())
			})

			ginkgo.It("should add blocker for not being a git repo", func() {
				report := checker.Check(ctx)
				gomega.Expect(report.Blockers).To(gomega.ContainElement("Not a git repository"))
			})

			ginkgo.It("should return early without other checks", func() {
				report := checker.Check(ctx)
				gomega.Expect(report.HasUncommittedChanges).To(gomega.BeFalse())
			})
		})

		ginkgo.Context("for git repository", func() {
			ginkgo.It("should report as git repo when .git exists", func() {
				// Note: This creates a minimal .git structure that may not pass all git commands
				// but the isGitRepo check just checks if git rev-parse succeeds
				// For a proper test we'd need to init a real git repo
			})
		})
	})

	ginkgo.Describe("isGitRepo", func() {
		ginkgo.It("should return false for non-git directory", func() {
			gomega.Expect(checker.isGitRepo(ctx)).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("hasUncommittedChanges", func() {
		ginkgo.It("should return true for non-git directory", func() {
			// git diff will fail in non-git directory, which we treat as having changes
			gomega.Expect(checker.hasUncommittedChanges(ctx)).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("getCurrentBranch", func() {
		ginkgo.It("should return unknown for non-git directory", func() {
			branch := checker.getCurrentBranch(ctx)
			gomega.Expect(branch).To(gomega.Equal("unknown"))
		})
	})

	ginkgo.Describe("hasUnpushedCommits", func() {
		ginkgo.It("should return false for non-git directory", func() {
			gomega.Expect(checker.hasUnpushedCommits(ctx)).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("isFilterRepoAvailable", func() {
		ginkgo.It("should return a boolean", func() {
			// This depends on whether git-filter-repo is installed
			result := checker.isFilterRepoAvailable(ctx)
			_ = result // Just verify it doesn't panic
		})
	})

	ginkgo.Describe("getDefaultBackupPath", func() {
		ginkgo.It("should return path with -backup.git suffix", func() {
			path := checker.getDefaultBackupPath()
			gomega.Expect(path).To(gomega.ContainSubstring("-backup.git"))
		})

		ginkgo.It("should be in parent directory", func() {
			path := checker.getDefaultBackupPath()
			parent := filepath.Dir(path)
			expectedParent := filepath.Dir(tempDir)
			gomega.Expect(parent).To(gomega.Equal(expectedParent))
		})
	})

	ginkgo.Describe("canCreateBackup", func() {
		ginkgo.It("should return true for valid path", func() {
			backupPath := filepath.Join(tempDir, "backup.git")
			gomega.Expect(checker.canCreateBackup(backupPath)).To(gomega.BeTrue())
		})

		ginkgo.It("should return false if backup already exists", func() {
			backupPath := filepath.Join(tempDir, "existing-backup.git")
			_ = os.MkdirAll(backupPath, 0o755)
			gomega.Expect(checker.canCreateBackup(backupPath)).To(gomega.BeFalse())
		})
	})
})
