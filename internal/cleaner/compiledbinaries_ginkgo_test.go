package cleaner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// mockBinaryScanner implements BinaryScanner for testing
type mockBinaryScanner struct {
	binaries    []BinaryInfo
	scanErr     error
	scanCallDir string
}

func (m *mockBinaryScanner) ScanDirectory(ctx context.Context, dir string, categories []BinaryCategory, minSize int64) ([]BinaryInfo, error) {
	m.scanCallDir = dir
	return m.binaries, m.scanErr
}

// mockBinaryTrashOperator implements BinaryTrashOperator for testing
type mockBinaryTrashOperator struct {
	trashErr       error
	trashedFiles   []string
	trashCallCount int
	fileSizes      map[string]int64
	fileModTimes   map[string]time.Time
}

func (m *mockBinaryTrashOperator) TrashBinary(ctx context.Context, path string) error {
	m.trashCallCount++
	m.trashedFiles = append(m.trashedFiles, path)
	return m.trashErr
}

func (m *mockBinaryTrashOperator) GetFileSize(path string) int64 {
	if m.fileSizes != nil {
		return m.fileSizes[path]
	}
	return 0
}

func (m *mockBinaryTrashOperator) GetFileModTime(path string) (time.Time, error) {
	if m.fileModTimes != nil {
		if t, ok := m.fileModTimes[path]; ok {
			return t, nil
		}
	}
	return time.Time{}, errors.New("file not found")
}



var _ = ginkgo.Describe("CompiledBinariesCleaner", func() {
	var (
		ctx           context.Context
		mockScanner   *mockBinaryScanner
		mockOperator  *mockBinaryTrashOperator
		cleaner       *CompiledBinariesCleaner
		tempDir       string
	)

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
		mockScanner = &mockBinaryScanner{}
		mockOperator = &mockBinaryTrashOperator{
			fileSizes:    make(map[string]int64),
			fileModTimes: make(map[string]time.Time),
		}
		tempDir, _ = os.MkdirTemp("", "compiled-binaries-test-*")
	})

	ginkgo.AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	// ============================================================================
	// CONSTRUCTOR TESTS
	// ============================================================================
	ginkgo.Describe("NewCompiledBinariesCleaner", func() {
		ginkgo.Context("with default configuration", func() {
			ginkgo.It("should create cleaner with default minimum size", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
				gomega.Expect(cleaner).NotTo(gomega.BeNil())
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(DefaultMinSizeMB))
			})

			ginkgo.It("should create cleaner with default older than", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
				gomega.Expect(cleaner.olderThan).To(gomega.Equal(DefaultOlderThan))
			})

			ginkgo.It("should create cleaner with default base paths", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
				gomega.Expect(cleaner.basePaths).NotTo(gomega.BeEmpty())
			})

			ginkgo.It("should create cleaner with all categories enabled", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
				gomega.Expect(cleaner.includeCategories).To(gomega.ContainElements(
					CategoryTmp, CategoryTest, CategoryBin, CategoryDist, CategoryRoot,
				))
			})
		})

		ginkgo.Context("with custom configuration", func() {
			ginkgo.It("should accept custom minimum size", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 50, "", nil, nil)
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(50))
			})

			ginkgo.It("should accept custom older than", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "7d", nil, nil)
				gomega.Expect(cleaner.olderThan).To(gomega.Equal("7d"))
			})

			ginkgo.It("should accept custom base paths", func() {
				paths := []string{"/custom/path1", "/custom/path2"}
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", paths, nil)
				gomega.Expect(cleaner.basePaths).To(gomega.Equal(paths))
			})

			ginkgo.It("should accept custom exclude patterns", func() {
				patterns := []string{"*.exclude", "specific-binary"}
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, patterns)
				gomega.Expect(cleaner.excludePatterns).To(gomega.Equal(patterns))
			})

			ginkgo.It("should set verbose flag correctly", func() {
				cleaner = NewCompiledBinariesCleaner(true, false, 0, "", nil, nil)
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
			})

			ginkgo.It("should set dryRun flag correctly", func() {
				cleaner = NewCompiledBinariesCleaner(false, true, 0, "", nil, nil)
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("with functional options", func() {
			ginkgo.It("should accept custom BinaryScanner via option", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil,
					WithBinaryScanner(mockScanner),
				)
				gomega.Expect(cleaner.scanner).To(gomega.Equal(mockScanner))
			})

			ginkgo.It("should accept custom BinaryTrashOperator via option", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil,
					WithBinaryTrashOperator(mockOperator),
				)
				gomega.Expect(cleaner.trashOperator).To(gomega.Equal(mockOperator))
			})

			ginkgo.It("should accept custom base paths via option", func() {
				paths := []string{"/option/path"}
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil,
					WithBasePaths(paths),
				)
				gomega.Expect(cleaner.basePaths).To(gomega.Equal(paths))
			})

			ginkgo.It("should accept custom categories via option", func() {
				categories := []BinaryCategory{CategoryTest, CategoryBin}
				cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil,
					WithIncludeCategories(categories),
				)
				gomega.Expect(cleaner.includeCategories).To(gomega.Equal(categories))
			})

			ginkgo.It("should accept multiple options together", func() {
				cleaner = NewCompiledBinariesCleaner(true, true, 20, "30d", []string{"/path"}, []string{},
					WithBinaryScanner(mockScanner),
					WithBinaryTrashOperator(mockOperator),
				)
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
				gomega.Expect(cleaner.minSizeMB).To(gomega.Equal(20))
				gomega.Expect(cleaner.olderThan).To(gomega.Equal("30d"))
				gomega.Expect(cleaner.scanner).To(gomega.Equal(mockScanner))
				gomega.Expect(cleaner.trashOperator).To(gomega.Equal(mockOperator))
			})
		})
	})

	// ============================================================================
	// BASIC METHOD TESTS
	// ============================================================================
	ginkgo.Describe("Name and Type methods", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
		})

		ginkgo.It("should return correct name", func() {
			gomega.Expect(cleaner.Name()).To(gomega.Equal("compiled-binaries"))
		})

		ginkgo.It("should return correct operation type", func() {
			gomega.Expect(cleaner.Type()).To(gomega.Equal(domain.OperationTypeCompiledBinaries))
		})
	})

	// ============================================================================
	// IsAvailable TESTS
	// ============================================================================
	ginkgo.Describe("IsAvailable", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
		})

		ginkgo.It("should return a boolean value", func() {
			result := cleaner.IsAvailable(ctx)
			gomega.Expect(result).To(gomega.BeAssignableToTypeOf(true))
		})

		ginkgo.It("should not panic when checking availability", func() {
			gomega.Expect(func() { cleaner.IsAvailable(ctx) }).NotTo(gomega.Panic())
		})

		ginkgo.It("should handle context parameter", func() {
			result := cleaner.IsAvailable(context.Background())
			_ = result
		})

		ginkgo.It("should work with cancelled context", func() {
			cancelledCtx, cancel := context.WithCancel(context.Background())
			cancel()
			result := cleaner.IsAvailable(cancelledCtx)
			_ = result
		})
	})

	// ============================================================================
	// ValidateSettings TESTS
	// ============================================================================
	ginkgo.Describe("ValidateSettings", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 0, "", nil, nil)
		})

		ginkgo.Context("with nil settings", func() {
			ginkgo.It("should return nil for nil settings", func() {
				err := cleaner.ValidateSettings(nil)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty OperationSettings", func() {
			ginkgo.It("should return nil when CompiledBinaries is nil", func() {
				settings := &domain.OperationSettings{}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid settings", func() {
			ginkgo.It("should return nil for valid empty CompiledBinariesSettings", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid min_size_mb", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						MinSizeMB: 10,
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid older_than with days", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						OlderThan: "7d",
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid older_than with hours", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						OlderThan: "24h",
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid exclude patterns", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						ExcludePatterns: []string{"*.exclude", "specific-*"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid include categories", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						IncludePatterns: []string{"tmp", "test", "bin"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid combined settings", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						MinSizeMB:       20,
						OlderThan:       "30d",
						BasePaths:       []string{"/custom/path"},
						ExcludePatterns: []string{"*.safe"},
						IncludePatterns: []string{"tmp", "test"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid settings", func() {
			ginkgo.It("should return error for negative min_size_mb", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						MinSizeMB: -1,
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("min_size_mb must be >= 0"))
			})

			ginkgo.It("should return error for invalid older_than format", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						OlderThan: "invalid",
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid older_than format"))
			})

			ginkgo.It("should return error for invalid glob pattern", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						ExcludePatterns: []string{"[invalid"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid exclude pattern"))
			})

			ginkgo.It("should return error for invalid include category", func() {
				settings := &domain.OperationSettings{
					CompiledBinaries: &domain.CompiledBinariesSettings{
						IncludePatterns: []string{"invalid-category"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid include category"))
			})
		})
	})

	// ============================================================================
	// Scan METHOD TESTS
	// ============================================================================
	ginkgo.Describe("Scan", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 10, "", []string{tempDir}, nil,
				WithBinaryScanner(mockScanner),
				WithBinaryTrashOperator(mockOperator),
			)
		})

		ginkgo.Context("when scan succeeds", func() {
			ginkgo.It("should return scan items from scanner", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
				}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items).To(gomega.HaveLen(2))
			})

			ginkgo.It("should convert BinaryInfo to ScanItem correctly", func() {
				now := time.Now()
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary", Size: 1024, ModTime: now, Category: CategoryTmp},
				}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items[0].Path).To(gomega.Equal("/path/to/binary"))
				gomega.Expect(items[0].Size).To(gomega.Equal(int64(1024)))
				gomega.Expect(items[0].ScanType).To(gomega.Equal(domain.ScanTypeSystem))
			})

			ginkgo.It("should return empty slice when no binaries found", func() {
				mockScanner.binaries = []BinaryInfo{}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("when scan fails", func() {
			ginkgo.It("should skip directories with errors and continue", func() {
				mockScanner.scanErr = errors.New("scan error")
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("with non-existent paths", func() {
			ginkgo.It("should skip non-existent base paths", func() {
				cleaner = NewCompiledBinariesCleaner(false, false, 10, "", []string{"/non/existent/path"}, nil,
					WithBinaryScanner(mockScanner),
				)
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("with age filter", func() {
			ginkgo.It("should filter files by age when older_than is set", func() {
				oldTime := time.Now().Add(-8 * 24 * time.Hour)
				recentTime := time.Now().Add(-1 * 24 * time.Hour)

				mockScanner.binaries = []BinaryInfo{
					{Path: "/old/binary", Size: 20 * 1024 * 1024, ModTime: oldTime, Category: CategoryTest},
					{Path: "/recent/binary", Size: 20 * 1024 * 1024, ModTime: recentTime, Category: CategoryTest},
				}

				cleaner = NewCompiledBinariesCleaner(false, false, 10, "7d", []string{tempDir}, nil,
					WithBinaryScanner(mockScanner),
				)

				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items).To(gomega.HaveLen(1))
				gomega.Expect(items[0].Path).To(gomega.Equal("/old/binary"))
			})
		})
	})

	// ============================================================================
	// Clean METHOD TESTS
	// ============================================================================
	ginkgo.Describe("Clean", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 10, "", []string{tempDir}, nil,
				WithBinaryScanner(mockScanner),
				WithBinaryTrashOperator(mockOperator),
			)
		})

		ginkgo.Context("with no items to clean", func() {
			ginkgo.It("should return conservative result when no items found", func() {
				mockScanner.binaries = []BinaryInfo{}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyConservative))
			})
		})

		ginkgo.Context("in dry-run mode", func() {
			ginkgo.BeforeEach(func() {
				cleaner = NewCompiledBinariesCleaner(false, true, 10, "", []string{tempDir}, nil,
					WithBinaryScanner(mockScanner),
					WithBinaryTrashOperator(mockOperator),
				)
			})

			ginkgo.It("should return dry-run result without calling trash", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyDryRun))
				gomega.Expect(mockOperator.trashCallCount).To(gomega.Equal(0))
			})

			ginkgo.It("should report correct size estimate in dry-run", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.FreedBytes).To(gomega.Equal(uint64(35 * 1024 * 1024)))
			})
		})

		ginkgo.Context("in normal mode", func() {
			ginkgo.It("should call TrashBinary for each item", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(mockOperator.trashCallCount).To(gomega.Equal(2))
			})

			ginkgo.It("should track items removed count correctly", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
					{Path: "/path/to/binary3", Size: 10 * 1024 * 1024, Category: CategoryTmp},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(3)))
			})

			ginkgo.It("should track bytes freed correctly", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.FreedBytes).To(gomega.Equal(uint64(35 * 1024 * 1024)))
			})

			ginkgo.It("should handle trash failures gracefully", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
					{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
				}
				mockOperator.trashErr = errors.New("trash failed")
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsFailed).To(gomega.Equal(uint(2)))
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
			})

			ginkgo.It("should set aggressive strategy after actual clean", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyAggressive))
			})

			ginkgo.It("should measure clean time", func() {
				mockScanner.binaries = []BinaryInfo{
					{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
				}
				before := time.Now()
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.CleanTime).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(cleanResult.CleanedAt).To(gomega.BeTemporally(">=", before))
			})
		})
	})

	// ============================================================================
	// GetStoreSize TESTS
	// ============================================================================
	ginkgo.Describe("GetStoreSize", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewCompiledBinariesCleaner(false, false, 10, "", []string{tempDir}, nil,
				WithBinaryScanner(mockScanner),
				WithBinaryTrashOperator(mockOperator),
			)
		})

		ginkgo.It("should return 0 when scan returns empty", func() {
			mockScanner.binaries = []BinaryInfo{}
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(0)))
		})

		ginkgo.It("should return total size of all binaries", func() {
			mockScanner.binaries = []BinaryInfo{
				{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
				{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
			}
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(35 * 1024 * 1024)))
		})
	})

	// ============================================================================
	// parseAgeDuration TESTS
	// ============================================================================
	ginkgo.Describe("parseAgeDuration", func() {
		ginkgo.It("should parse days correctly", func() {
			d, err := parseAgeDuration("7d")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(d).To(gomega.Equal(7 * 24 * time.Hour))
		})

		ginkgo.It("should parse hours correctly", func() {
			d, err := parseAgeDuration("24h")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(d).To(gomega.Equal(24 * time.Hour))
		})

		ginkgo.It("should parse weeks correctly", func() {
			d, err := parseAgeDuration("2w")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(d).To(gomega.Equal(14 * 24 * time.Hour))
		})

		ginkgo.It("should parse months correctly", func() {
			d, err := parseAgeDuration("1m")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(d).To(gomega.Equal(30 * 24 * time.Hour))
		})

		ginkgo.It("should parse years correctly", func() {
			d, err := parseAgeDuration("1y")
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(d).To(gomega.Equal(365 * 24 * time.Hour))
		})

		ginkgo.It("should return error for invalid format", func() {
			_, err := parseAgeDuration("invalid")
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return error for too short format", func() {
			_, err := parseAgeDuration("d")
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return error for unknown unit", func() {
			_, err := parseAgeDuration("7x")
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})

// ============================================================================
// DEFAULT SCANNER TESTS
// ============================================================================
var _ = ginkgo.Describe("defaultBinaryScanner", func() {
	var (
		scanner     *defaultBinaryScanner
		tempDir     string
		ctx         context.Context
	)

	ginkgo.BeforeEach(func() {
		scanner = &defaultBinaryScanner{
			includeCategories: []BinaryCategory{CategoryTmp, CategoryTest, CategoryBin, CategoryDist, CategoryRoot},
		}
		tempDir, _ = os.MkdirTemp("", "scanner-test-*")
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	ginkgo.Describe("ScanDirectory", func() {
		ginkgo.Context("categorization", func() {
			ginkgo.It("should categorize *.test files as CategoryTest", func() {
				testFile := filepath.Join(tempDir, "app.test")
				_ = os.WriteFile(testFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryTest}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Category).To(gomega.Equal(CategoryTest))
			})

			ginkgo.It("should categorize tmp/* files as CategoryTmp", func() {
				tmpDir := filepath.Join(tempDir, "tmp")
				_ = os.MkdirAll(tmpDir, 0755)
				binaryFile := filepath.Join(tmpDir, "build-output")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryTmp}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Category).To(gomega.Equal(CategoryTmp))
			})

			ginkgo.It("should categorize bin/* files as CategoryBin", func() {
				binDir := filepath.Join(tempDir, "bin")
				_ = os.MkdirAll(binDir, 0755)
				binaryFile := filepath.Join(binDir, "myapp")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryBin}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Category).To(gomega.Equal(CategoryBin))
			})

			ginkgo.It("should categorize dist/* files as CategoryDist", func() {
				distDir := filepath.Join(tempDir, "dist")
				_ = os.MkdirAll(distDir, 0755)
				binaryFile := filepath.Join(distDir, "release-binary")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryDist}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Category).To(gomega.Equal(CategoryDist))
			})

			ginkgo.It("should categorize root executables as CategoryRoot", func() {
				binaryFile := filepath.Join(tempDir, "myapp")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryRoot}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Category).To(gomega.Equal(CategoryRoot))
			})
		})

		ginkgo.Context("size filtering", func() {
			ginkgo.It("should skip files below minimum size", func() {
				smallFile := filepath.Join(tempDir, "small.test")
				_ = os.WriteFile(smallFile, make([]byte, 5*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryTest}, 10*1024*1024)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should include files at or above minimum size", func() {
				largeFile := filepath.Join(tempDir, "large.test")
				_ = os.WriteFile(largeFile, make([]byte, 15*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryTest}, 10*1024*1024)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
			})
		})

		ginkgo.Context("directory exclusion", func() {
			ginkgo.It("should skip node_modules directory", func() {
				nodeModulesDir := filepath.Join(tempDir, "node_modules", ".bin")
				_ = os.MkdirAll(nodeModulesDir, 0755)
				binaryFile := filepath.Join(nodeModulesDir, "tool")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should skip venv directory", func() {
				venvDir := filepath.Join(tempDir, "venv", "bin")
				_ = os.MkdirAll(venvDir, 0755)
				binaryFile := filepath.Join(venvDir, "python")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should skip .terraform directory", func() {
				terraformDir := filepath.Join(tempDir, ".terraform", "providers")
				_ = os.MkdirAll(terraformDir, 0755)
				binaryFile := filepath.Join(terraformDir, "provider")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should skip .git directory", func() {
				gitDir := filepath.Join(tempDir, ".git", "objects")
				_ = os.MkdirAll(gitDir, 0755)
				binaryFile := filepath.Join(gitDir, "some-file")
				_ = os.WriteFile(binaryFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("binary exclusion", func() {
			ginkgo.It("should skip chromedriver", func() {
				chromedriver := filepath.Join(tempDir, "chromedriver")
				_ = os.WriteFile(chromedriver, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should skip geckodriver", func() {
				geckodriver := filepath.Join(tempDir, "geckodriver")
				_ = os.WriteFile(geckodriver, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("executable check", func() {
			ginkgo.It("should skip non-executable files", func() {
				nonExecFile := filepath.Join(tempDir, "data.txt")
				_ = os.WriteFile(nonExecFile, make([]byte, 20*1024*1024), 0644)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, scanner.includeCategories, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.BeEmpty())
			})

			ginkgo.It("should include executable files", func() {
				execFile := filepath.Join(tempDir, "myapp")
				_ = os.WriteFile(execFile, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryRoot}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
			})
		})

		ginkgo.Context("exclude patterns", func() {
			ginkgo.It("should exclude files matching patterns", func() {
				scanner.excludePatterns = []string{"*.safe"}
				binary1 := filepath.Join(tempDir, "myapp")
				binary2 := filepath.Join(tempDir, "important.safe")
				_ = os.WriteFile(binary1, make([]byte, 20*1024*1024), 0755)
				_ = os.WriteFile(binary2, make([]byte, 20*1024*1024), 0755)

				binaries, err := scanner.ScanDirectory(ctx, tempDir, []BinaryCategory{CategoryRoot}, 0)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(binaries).To(gomega.HaveLen(1))
				gomega.Expect(binaries[0].Path).To(gomega.Equal(binary1))
			})
		})
	})
})

// ============================================================================
// INTEGRATION TESTS
// ============================================================================
var _ = ginkgo.Describe("CompiledBinariesCleaner Integration", func() {
	ginkgo.It("should work with default implementations", func() {
		cleaner := NewCompiledBinariesCleaner(false, false, 10, "", nil, nil)
		ctx := context.Background()

		if !cleaner.IsAvailable(ctx) {
			ginkgo.Skip("Skipping integration test: trash not available")
		}

		scanResult := cleaner.Scan(ctx)
		if scanResult.IsOk() {
			gomega.Expect(scanResult.Value()).To(gomega.BeAssignableToTypeOf([]domain.ScanItem{}))
		}
	})

	ginkgo.It("should handle real filesystem operations", func() {
		cleaner := NewCompiledBinariesCleaner(false, false, 10, "", nil, nil)
		ctx := context.Background()

		if !cleaner.IsAvailable(ctx) {
			ginkgo.Skip("Skipping integration test: trash not available")
		}

		tmpDir, err := os.MkdirTemp("", "integration-test-*")
		gomega.Expect(err).To(gomega.BeNil())
		defer os.RemoveAll(tmpDir)

		execFile := filepath.Join(tmpDir, "test-binary")
		err = os.WriteFile(execFile, make([]byte, 15*1024*1024), 0755)
		gomega.Expect(err).To(gomega.BeNil())

		info, err := os.Stat(execFile)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(info.Mode()&0111).NotTo(gomega.Equal(os.FileMode(0)))
	})
})
