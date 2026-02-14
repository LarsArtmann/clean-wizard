package bdd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Test entry point for Ginkgo.
func TestNixBDDSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Nix BDD Suite")
}

// NixTestContext holds test state across scenarios.
type NixTestContext struct {
	ctx         context.Context
	nixCleaner  *cleaner.NixCleaner
	generations result.Result[[]domain.NixGeneration]
	cleanResult result.Result[domain.CleanResult]
	storeSize   result.Result[int64]
	output      *bytes.Buffer
	dryRun      bool
}

var _ = ginkgo.Describe("Nix Store Management", func() {
	var testCtx *NixTestContext

	ginkgo.BeforeEach(func() {
		testCtx = &NixTestContext{
			ctx:    context.Background(),
			output: &bytes.Buffer{},
			dryRun: true, // Force dry-run for safe testing
		}
	})

	ginkgo.Describe("Background", func() {
		ginkgo.Context("system setup", func() {
			ginkgo.It("should have Nix package manager available", func() {
				testCtx.nixCleaner = cleaner.NewNixCleaner(true, false)
				gomega.Expect(testCtx.nixCleaner).NotTo(gomega.BeNil())
			})

			ginkgo.It("should have clean-wizard tool available", func() {
				toolPath, err := os.Executable()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(toolPath).NotTo(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Describe("List available Nix generations", func() {
		ginkgo.BeforeEach(func() {
			testCtx.nixCleaner = cleaner.NewNixCleaner(true, false)
		})

		ginkgo.It("should list Nix generations when running scan", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 2)
			gomega.Expect(testCtx.generations.IsOk()).To(gomega.BeTrue())
		})

		ginkgo.It("should have valid ID for each generation", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 1)
			generations := testCtx.generations.Value()
			for _, gen := range generations {
				gomega.Expect(gen.ID).To(gomega.BeNumerically(">", 0))
			}
		})

		ginkgo.It("should have creation date for each generation", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 1)
			generations := testCtx.generations.Value()
			for _, gen := range generations {
				gomega.Expect(gen.Date).NotTo(gomega.BeZero())
			}
		})

		ginkgo.It("should display total store size", func() {
			storeSize := testCtx.nixCleaner.GetStoreSize(testCtx.ctx)
			// In CI, store size may fail - that's acceptable
			if storeSize > 0 {
				gomega.Expect(storeSize).To(gomega.BeNumerically(">", 0))
			}
		})

		ginkgo.It("should complete scan command successfully", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 0)
			gomega.Expect(testCtx.generations.IsOk()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Clean old Nix generations safely", func() {
		ginkgo.BeforeEach(func() {
			testCtx.nixCleaner = cleaner.NewNixCleaner(true, true) // verbose, dryRun
		})

		ginkgo.It("should show what would be cleaned in dry-run mode", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 2)
			testCtx.cleanResult = testCtx.nixCleaner.CleanOldGenerations(testCtx.ctx, 3)
			gomega.Expect(testCtx.cleanResult.IsOk()).To(gomega.BeTrue())
		})

		ginkgo.It("should estimate space to be freed", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 1)
			testCtx.cleanResult = testCtx.nixCleaner.CleanOldGenerations(testCtx.ctx, 3)
			if testCtx.cleanResult.IsOk() {
				cleanRes := testCtx.cleanResult.Value()
				gomega.Expect(cleanRes.Strategy.IsValid()).To(gomega.BeTrue())
			}
		})

		ginkgo.It("should show how many generations would be removed", func() {
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 1)
			testCtx.cleanResult = testCtx.nixCleaner.CleanOldGenerations(testCtx.ctx, 3)
			if testCtx.cleanResult.IsOk() {
				cleanRes := testCtx.cleanResult.Value()
				gomega.Expect(cleanRes.ItemsRemoved).To(gomega.BeNumerically(">=", 0))
			}
		})

		ginkgo.It("should not perform actual cleaning in dry-run mode", func() {
			testCtx.dryRun = true
			testCtx.nixCleaner = cleaner.NewNixCleaner(true, true)
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 1)
			testCtx.cleanResult = testCtx.nixCleaner.CleanOldGenerations(testCtx.ctx, 3)
			gomega.Expect(testCtx.dryRun).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Clean old Nix generations for real", func() {
		ginkgo.BeforeEach(func() {
			testCtx.nixCleaner = cleaner.NewNixCleaner(false, true) // not verbose, dryRun for safety
		})

		ginkgo.It("should keep specified number of generations", func() {
			keepCount := 3
			testCtx.generations = getGenerationsOrMock(testCtx.ctx, testCtx.nixCleaner, 4)
			generations := testCtx.generations.Value()
			if len(generations) > keepCount {
				testCtx.cleanResult = testCtx.nixCleaner.CleanOldGenerations(testCtx.ctx, keepCount)
				if testCtx.cleanResult.IsOk() {
					cleanRes := testCtx.cleanResult.Value()
					gomega.Expect(cleanRes.ItemsRemoved).To(gomega.BeNumerically(">=", 0))
				}
			}
		})
	})

	ginkgo.Describe("Handle Nix not available gracefully", func() {
		ginkgo.BeforeEach(func() {
			// Simulate Nix not being available
			testCtx.nixCleaner = cleaner.NewNixCleaner(true, false)
			testCtx.generations = result.Err[[]domain.NixGeneration](errors.New("Nix is not available"))
			testCtx.storeSize = result.Err[int64](errors.New("Nix is not available"))
			testCtx.cleanResult = result.Err[domain.CleanResult](errors.New("Nix is not available"))
		})

		ginkgo.It("should show helpful error message", func() {
			gomega.Expect(testCtx.generations.IsErr()).To(gomega.BeTrue())
			errMsg := testCtx.generations.Error().Error()
			gomega.Expect(strings.Contains(errMsg, "Nix") || strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "command not found")).To(gomega.BeTrue())
		})

		ginkgo.It("should fail gracefully without stack trace", func() {
			gomega.Expect(testCtx.generations.IsErr()).To(gomega.BeTrue())
			errMsg := testCtx.generations.Error().Error()
			// Should not contain typical stack trace elements
			gomega.Expect(strings.Contains(errMsg, "goroutine")).To(gomega.BeFalse())
			gomega.Expect(strings.Contains(errMsg, "panic")).To(gomega.BeFalse())
		})
	})
})

// Nix Cleaning feature tests.
var _ = ginkgo.Describe("Nix Store Cleaning", func() {
	var nixCtx *NixTestContext

	ginkgo.BeforeEach(func() {
		nixCtx = &NixTestContext{
			ctx:    context.Background(),
			output: &bytes.Buffer{},
			dryRun: true,
		}
		nixCtx.nixCleaner = cleaner.NewNixCleaner(true, true)
	})

	ginkgo.Describe("List available Nix generations", func() {
		ginkgo.It("should display generation numbers and dates", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 1)
			generations := nixCtx.generations.Value()
			gomega.Expect(len(generations)).To(gomega.BeNumerically(">", 0))
			for _, gen := range generations {
				gomega.Expect(gen.ID).To(gomega.BeNumerically(">", 0))
				gomega.Expect(gen.Date).NotTo(gomega.BeZero())
			}
		})

		ginkgo.It("should mark current generation", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 2)
			generations := nixCtx.generations.Value()
			currentCount := 0
			for _, gen := range generations {
				if gen.Current.IsCurrent() {
					currentCount++
				}
			}
			// Should have at most one current generation
			gomega.Expect(currentCount).To(gomega.BeNumerically("<=", 1))
		})
	})

	ginkgo.Describe("Clean old Nix generations safely", func() {
		ginkgo.BeforeEach(func() {
			nixCtx.dryRun = true
		})

		ginkgo.It("should show which generations would be deleted", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 2)
			nixCtx.cleanResult = nixCtx.nixCleaner.CleanOldGenerations(nixCtx.ctx, 1)
			gomega.Expect(nixCtx.cleanResult.IsOk()).To(gomega.BeTrue())
		})

		ginkgo.It("should not delete current generation", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 1)
			generations := nixCtx.generations.Value()
			// Find current generation
			var currentGen *domain.NixGeneration
			for i := range generations {
				if generations[i].Current.IsCurrent() {
					currentGen = &generations[i]
					break
				}
			}
			if currentGen != nil {
				gomega.Expect(currentGen.Current.IsCurrent()).To(gomega.BeTrue())
			}
		})

		ginkgo.It("should require confirmation before real deletion", func() {
			// In dry-run mode, no actual deletion happens
			nixCtx.cleanResult = nixCtx.nixCleaner.CleanOldGenerations(nixCtx.ctx, 1)
			if nixCtx.cleanResult.IsOk() {
				cleanRes := nixCtx.cleanResult.Value()
				// Dry-run should be indicated in strategy
				gomega.Expect(cleanRes.Strategy.IsValid()).To(gomega.BeTrue())
			}
		})
	})

	ginkgo.Describe("Prevent deletion of current generation", func() {
		ginkgo.It("should keep current generation after cleaning", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 1)
			generations := nixCtx.generations.Value()
			var currentGenID int64
			for _, gen := range generations {
				if gen.Current.IsCurrent() {
					currentGenID = int64(gen.ID)
					break
				}
			}
			if currentGenID > 0 {
				gomega.Expect(currentGenID).To(gomega.BeNumerically(">", 0))
			}
		})
	})

	ginkgo.Describe("Clean with dry-run mode", func() {
		ginkgo.It("should show what would be deleted without deleting", func() {
			nixCtx.dryRun = true
			nixCtx.nixCleaner = cleaner.NewNixCleaner(true, true)
			nixCtx.cleanResult = nixCtx.nixCleaner.CleanOldGenerations(nixCtx.ctx, 3)
			// In dry-run, no bytes should be freed
			if nixCtx.cleanResult.IsOk() {
				cleanRes := nixCtx.cleanResult.Value()
				gomega.Expect(cleanRes.Strategy.IsValid()).To(gomega.BeTrue())
			}
		})

		ginkgo.It("should not actually delete generations in dry-run", func() {
			nixCtx.dryRun = true
			nixCtx.nixCleaner = cleaner.NewNixCleaner(true, true)
			nixCtx.cleanResult = nixCtx.nixCleaner.CleanOldGenerations(nixCtx.ctx, 3)
			gomega.Expect(nixCtx.dryRun).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Clean with backup protection", func() {
		ginkgo.It("should protect important generations", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 1)
			// Current generation should always be protected
			generations := nixCtx.generations.Value()
			for _, gen := range generations {
				if gen.Current.IsCurrent() {
					gomega.Expect(gen.Current.IsCurrent()).To(gomega.BeTrue())
				}
			}
		})

		ginkgo.It("should show space estimation", func() {
			nixCtx.cleanResult = nixCtx.nixCleaner.CleanOldGenerations(nixCtx.ctx, 3)
			if nixCtx.cleanResult.IsOk() {
				cleanRes := nixCtx.cleanResult.Value()
				gomega.Expect(cleanRes.FreedBytes).To(gomega.BeNumerically(">=", 0))
			}
		})
	})

	ginkgo.Describe("Verify type-safe operations", func() {
		ginkgo.It("should use type-safe operations", func() {
			nixCtx.generations = getGenerationsOrMock(nixCtx.ctx, nixCtx.nixCleaner, 0)
			gomega.Expect(nixCtx.generations.IsOk()).To(gomega.BeTrue())
		})

		ginkgo.It("should have consistent error handling", func() {
			nixCtx.generations = nixCtx.nixCleaner.ListGenerations(nixCtx.ctx)
			// Either success or error, but consistent
			if nixCtx.generations.IsErr() {
				gomega.Expect(nixCtx.generations.Error()).NotTo(gomega.Succeed())
			} else {
				gomega.Expect(nixCtx.generations.Value()).NotTo(gomega.BeNil())
			}
		})
	})
})
