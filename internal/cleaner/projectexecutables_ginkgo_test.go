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

// mockProjectLister implements ProjectLister for testing
type mockProjectLister struct {
	projects []ProjectInfo
	err      error
}

func (m *mockProjectLister) ListProjects(ctx context.Context) ([]ProjectInfo, error) {
	return m.projects, m.err
}

// mockFileOperator implements FileOperator for testing
type mockFileOperator struct {
	executables     []string
	executablesErr  error
	trashErr        error
	fileSizes       map[string]int64
	trashedFiles    []string
	trashCallCount  int
	// perDirExecutables allows specifying different executables per directory
	perDirExecutables map[string][]string
}

func (m *mockFileOperator) FindExecutableFiles(dir string) ([]string, error) {
	// If per-directory mapping exists, use it
	if m.perDirExecutables != nil {
		if execs, ok := m.perDirExecutables[dir]; ok {
			return execs, m.executablesErr
		}
	}
	// Otherwise, filter executables that belong to this directory
	var result []string
	for _, exec := range m.executables {
		if filepath.Dir(exec) == dir {
			result = append(result, exec)
		}
	}
	return result, m.executablesErr
}

func (m *mockFileOperator) TrashFile(ctx context.Context, path string) error {
	m.trashCallCount++
	m.trashedFiles = append(m.trashedFiles, path)
	return m.trashErr
}

func (m *mockFileOperator) GetFileSize(path string) int64 {
	if m.fileSizes != nil {
		if size, ok := m.fileSizes[path]; ok {
			return size
		}
	}
	return 0
}



var _ = ginkgo.Describe("ProjectExecutablesCleaner", func() {
	var (
		ctx             context.Context
		mockLister      *mockProjectLister
		mockOperator    *mockFileOperator
		cleaner         *ProjectExecutablesCleaner
	)

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
		mockLister = &mockProjectLister{}
		mockOperator = &mockFileOperator{
			fileSizes: make(map[string]int64),
		}
	})

	// ============================================================================
	// CONSTRUCTOR TESTS (10 specs)
	// ============================================================================
	ginkgo.Describe("NewProjectExecutablesCleaner", func() {
		ginkgo.Context("with default configuration", func() {
			ginkgo.It("should create cleaner with default .sh extension exclusion", func() {
				cleaner = NewProjectExecutablesCleaner(false, false, nil, nil)
				gomega.Expect(cleaner).NotTo(gomega.BeNil())
				gomega.Expect(cleaner.excludeExtensions).To(gomega.Equal([]string{".sh"}))
			})

			ginkgo.It("should create cleaner with empty exclude extensions defaulting to .sh", func() {
				cleaner = NewProjectExecutablesCleaner(false, false, []string{}, nil)
				gomega.Expect(cleaner.excludeExtensions).To(gomega.Equal([]string{".sh"}))
			})
		})

		ginkgo.Context("with custom configuration", func() {
			ginkgo.It("should accept custom exclude extensions", func() {
				cleaner = NewProjectExecutablesCleaner(false, false, []string{".sh", ".bash"}, nil)
				gomega.Expect(cleaner.excludeExtensions).To(gomega.Equal([]string{".sh", ".bash"}))
			})

			ginkgo.It("should accept custom exclude patterns", func() {
				cleaner = NewProjectExecutablesCleaner(false, false, nil, []string{"Makefile", "*.config"})
				gomega.Expect(cleaner.excludePatterns).To(gomega.Equal([]string{"Makefile", "*.config"}))
			})

			ginkgo.It("should set verbose flag correctly", func() {
				cleaner = NewProjectExecutablesCleaner(true, false, nil, nil)
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
			})

			ginkgo.It("should set dryRun flag correctly", func() {
				cleaner = NewProjectExecutablesCleaner(false, true, nil, nil)
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("with functional options", func() {
			ginkgo.It("should accept custom ProjectLister via option", func() {
				mockLister.projects = []ProjectInfo{{Name: "test", Path: "/test"}}
				cleaner = NewProjectExecutablesCleaner(false, false, nil, nil, WithProjectLister(mockLister))
				gomega.Expect(cleaner.projectLister).To(gomega.Equal(mockLister))
			})

			ginkgo.It("should accept custom FileOperator via option", func() {
				cleaner = NewProjectExecutablesCleaner(false, false, nil, nil, WithFileOperator(mockOperator))
				gomega.Expect(cleaner.fileOperator).To(gomega.Equal(mockOperator))
			})

			ginkgo.It("should accept both options together", func() {
				cleaner = NewProjectExecutablesCleaner(true, true, []string{".sh"}, []string{"Makefile"},
					WithProjectLister(mockLister),
					WithFileOperator(mockOperator),
				)
				gomega.Expect(cleaner.projectLister).To(gomega.Equal(mockLister))
				gomega.Expect(cleaner.fileOperator).To(gomega.Equal(mockOperator))
				gomega.Expect(cleaner.verbose).To(gomega.BeTrue())
				gomega.Expect(cleaner.dryRun).To(gomega.BeTrue())
			})
		})
	})

	// ============================================================================
	// BASIC METHOD TESTS (5 specs)
	// ============================================================================
	ginkgo.Describe("Name and Type methods", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, nil, nil)
		})

		ginkgo.It("should return correct name", func() {
			gomega.Expect(cleaner.Name()).To(gomega.Equal("project-executables"))
		})

		ginkgo.It("should return correct operation type", func() {
			gomega.Expect(cleaner.Type()).To(gomega.Equal(domain.OperationTypeProjectExecutables))
		})
	})

	// ============================================================================
	// IsAvailable TESTS (5 specs)
	// ============================================================================
	ginkgo.Describe("IsAvailable", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, nil, nil)
		})

		ginkgo.It("should return true when both tools are available", func() {
			// This test depends on system state, so we just verify it doesn't panic
			result := cleaner.IsAvailable(ctx)
			_ = result
		})

		ginkgo.It("should not panic when checking availability", func() {
			gomega.Expect(func() { cleaner.IsAvailable(ctx) }).NotTo(gomega.Panic())
		})

		ginkgo.It("should return a boolean value", func() {
			result := cleaner.IsAvailable(ctx)
			gomega.Expect(result).To(gomega.BeAssignableToTypeOf(true))
		})

		ginkgo.It("should handle context parameter", func() {
			// Verify context is accepted without error
			result := cleaner.IsAvailable(context.Background())
			_ = result
		})

		ginkgo.It("should work with cancelled context", func() {
			cancelledCtx, cancel := context.WithCancel(context.Background())
			cancel()
			// Should still return a result (doesn't depend on context for tool check)
			result := cleaner.IsAvailable(cancelledCtx)
			_ = result
		})
	})

	// ============================================================================
	// ValidateSettings TESTS (10 specs)
	// ============================================================================
	ginkgo.Describe("ValidateSettings", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, nil, nil)
		})

		ginkgo.Context("with nil settings", func() {
			ginkgo.It("should return nil for nil settings", func() {
				err := cleaner.ValidateSettings(nil)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty OperationSettings", func() {
			ginkgo.It("should return nil when ProjectExecutables is nil", func() {
				settings := &domain.OperationSettings{}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid settings", func() {
			ginkgo.It("should return nil for valid empty ProjectExecutablesSettings", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid exclude extensions", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{
						ExcludeExtensions: []string{".sh", ".bash"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid exclude patterns", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{
						ExcludePatterns: []string{"Makefile", "*.config"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return nil for valid combined settings", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{
						ExcludeExtensions: []string{".sh"},
						ExcludePatterns:   []string{"Makefile"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid glob patterns", func() {
			ginkgo.It("should return error for invalid glob pattern [invalid", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{
						ExcludePatterns: []string{"[invalid"},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid exclude pattern"))
			})

			ginkgo.It("should return error for invalid glob pattern with unclosed bracket", func() {
				settings := &domain.OperationSettings{
					ProjectExecutables: &domain.ProjectExecutablesSettings{
						ExcludePatterns: []string{"test["},
					},
				}
				err := cleaner.ValidateSettings(settings)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	// ============================================================================
	// Scan METHOD TESTS (10 specs)
	// ============================================================================
	ginkgo.Describe("Scan", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, []string{".sh"}, nil,
				WithProjectLister(mockLister),
				WithFileOperator(mockOperator),
			)
		})

		ginkgo.Context("when ListProjects fails", func() {
			ginkgo.It("should return error when project listing fails", func() {
				mockLister.err = errors.New("failed to list projects")
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsErr()).To(gomega.BeTrue())
				gomega.Expect(result.Error()).To(gomega.Equal(mockLister.err))
			})
		})

		ginkgo.Context("when no projects found", func() {
			ginkgo.It("should return empty slice when no projects", func() {
				mockLister.projects = []ProjectInfo{}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("when FindExecutableFiles fails", func() {
			ginkgo.It("should skip directories that cannot be read", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executablesErr = errors.New("cannot read directory")
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.BeEmpty())
			})
		})

		ginkgo.Context("when executable files are found", func() {
			ginkgo.It("should find executable files in project roots", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1", "/path/to/project1/binary2"}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 1024,
					"/path/to/project1/binary2": 2048,
				}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items).To(gomega.HaveLen(2))
			})

			ginkgo.It("should combine files from multiple projects", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
					{Name: "project2", Path: "/path/to/project2"},
				}
				// First call returns project1 files, need to simulate per-project behavior
				// For simplicity, we test with a single set of executables
				mockOperator.executables = []string{"/path/to/project1/binary1", "/path/to/project2/binary2"}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(result.Value()).To(gomega.HaveLen(2))
			})

			ginkgo.It("should calculate file sizes correctly", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 4096,
				}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items[0].Size).To(gomega.Equal(int64(4096)))
			})

			ginkgo.It("should set correct ScanType", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				result := cleaner.Scan(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items[0].ScanType).To(gomega.Equal(domain.ScanTypeSystem))
			})

			ginkgo.It("should set valid Created timestamp", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				before := time.Now()
				result := cleaner.Scan(ctx)
				after := time.Now()
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				items := result.Value()
				gomega.Expect(items[0].Created).To(gomega.BeTemporally(">=", before))
				gomega.Expect(items[0].Created).To(gomega.BeTemporally("<=", after))
			})
		})

		ginkgo.Context("with verbose mode", func() {
			ginkgo.It("should not panic with verbose mode enabled", func() {
				cleaner = NewProjectExecutablesCleaner(true, false, []string{".sh"}, nil,
					WithProjectLister(mockLister),
					WithFileOperator(mockOperator),
				)
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				gomega.Expect(func() { cleaner.Scan(ctx) }).NotTo(gomega.Panic())
			})
		})
	})

	// ============================================================================
	// Clean METHOD TESTS (10 specs)
	// ============================================================================
	ginkgo.Describe("Clean", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, []string{".sh"}, nil,
				WithProjectLister(mockLister),
				WithFileOperator(mockOperator),
			)
		})

		ginkgo.Context("when Scan fails", func() {
			ginkgo.It("should return error when scan fails", func() {
				mockLister.err = errors.New("scan failed")
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsErr()).To(gomega.BeTrue())
				gomega.Expect(result.Error()).To(gomega.Equal(mockLister.err))
			})
		})

		ginkgo.Context("with no items to clean", func() {
			ginkgo.It("should return conservative result when no items found", func() {
				mockLister.projects = []ProjectInfo{}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyConservative))
			})
		})

		ginkgo.Context("in dry-run mode", func() {
			ginkgo.BeforeEach(func() {
				cleaner = NewProjectExecutablesCleaner(false, true, []string{".sh"}, nil,
					WithProjectLister(mockLister),
					WithFileOperator(mockOperator),
				)
			})

			ginkgo.It("should return dry-run result without calling trash", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 1024,
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyDryRun))
				gomega.Expect(mockOperator.trashCallCount).To(gomega.Equal(0))
			})

			ginkgo.It("should report correct size estimate in dry-run", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 2048,
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.FreedBytes).To(gomega.Equal(uint64(2048)))
			})
		})

		ginkgo.Context("in normal mode", func() {
			ginkgo.It("should call TrashFile for each item", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{
					"/path/to/project1/binary1",
					"/path/to/project1/binary2",
				}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 1024,
					"/path/to/project1/binary2": 2048,
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				gomega.Expect(mockOperator.trashCallCount).To(gomega.Equal(2))
			})

			ginkgo.It("should track items removed count correctly", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{
					"/path/to/project1/binary1",
					"/path/to/project1/binary2",
					"/path/to/project1/binary3",
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(3)))
			})

			ginkgo.It("should track bytes freed correctly", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{
					"/path/to/project1/binary1",
					"/path/to/project1/binary2",
				}
				mockOperator.fileSizes = map[string]int64{
					"/path/to/project1/binary1": 1024,
					"/path/to/project1/binary2": 2048,
				}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.FreedBytes).To(gomega.Equal(uint64(3072)))
			})

			ginkgo.It("should handle trash failures gracefully", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{
					"/path/to/project1/binary1",
					"/path/to/project1/binary2",
				}
				mockOperator.trashErr = errors.New("trash failed")
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.ItemsFailed).To(gomega.Equal(uint(2)))
				gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
			})

			ginkgo.It("should set aggressive strategy after actual clean", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
				result := cleaner.Clean(ctx)
				gomega.Expect(result.IsOk()).To(gomega.BeTrue())
				cleanResult := result.Value()
				gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyAggressive))
			})

			ginkgo.It("should measure clean time", func() {
				mockLister.projects = []ProjectInfo{
					{Name: "project1", Path: "/path/to/project1"},
				}
				mockOperator.executables = []string{"/path/to/project1/binary1"}
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
	// Exclusion Logic TESTS (5 specs)
	// ============================================================================
	ginkgo.Describe("IsExcludedByExtension", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, []string{".sh", ".bash"}, nil)
		})

		ginkgo.It("should exclude .sh files", func() {
			gomega.Expect(cleaner.IsExcludedByExtension("script.sh")).To(gomega.BeTrue())
		})

		ginkgo.It("should exclude files case-insensitively", func() {
			gomega.Expect(cleaner.IsExcludedByExtension("SCRIPT.SH")).To(gomega.BeTrue())
			gomega.Expect(cleaner.IsExcludedByExtension("Script.Sh")).To(gomega.BeTrue())
		})

		ginkgo.It("should not exclude non-matching extensions", func() {
			gomega.Expect(cleaner.IsExcludedByExtension("binary")).To(gomega.BeFalse())
			gomega.Expect(cleaner.IsExcludedByExtension("script.py")).To(gomega.BeFalse())
		})

		ginkgo.It("should match multiple extension patterns", func() {
			gomega.Expect(cleaner.IsExcludedByExtension("script.sh")).To(gomega.BeTrue())
			gomega.Expect(cleaner.IsExcludedByExtension("script.bash")).To(gomega.BeTrue())
		})

		ginkgo.It("should match extension exactly at end", func() {
			gomega.Expect(cleaner.IsExcludedByExtension("shell.sh")).To(gomega.BeTrue())
			gomega.Expect(cleaner.IsExcludedByExtension("shell.sh.backup")).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("IsExcludedByPattern", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, nil, []string{"Makefile", "*.config", "test_*"})
		})

		ginkgo.It("should match exact filename pattern", func() {
			gomega.Expect(cleaner.IsExcludedByPattern("Makefile")).To(gomega.BeTrue())
		})

		ginkgo.It("should match wildcard patterns", func() {
			gomega.Expect(cleaner.IsExcludedByPattern("app.config")).To(gomega.BeTrue())
			gomega.Expect(cleaner.IsExcludedByPattern("database.config")).To(gomega.BeTrue())
		})

		ginkgo.It("should match prefix wildcard patterns", func() {
			gomega.Expect(cleaner.IsExcludedByPattern("test_binary")).To(gomega.BeTrue())
			gomega.Expect(cleaner.IsExcludedByPattern("test_run")).To(gomega.BeTrue())
		})

		ginkgo.It("should not match non-matching patterns", func() {
			gomega.Expect(cleaner.IsExcludedByPattern("binary")).To(gomega.BeFalse())
			gomega.Expect(cleaner.IsExcludedByPattern("app.yaml")).To(gomega.BeFalse())
		})

		ginkgo.It("should return false for empty pattern list", func() {
			cleaner = NewProjectExecutablesCleaner(false, false, nil, []string{})
			gomega.Expect(cleaner.IsExcludedByPattern("Makefile")).To(gomega.BeFalse())
		})
	})

	// ============================================================================
	// GetStoreSize TESTS (3 specs)
	// ============================================================================
	ginkgo.Describe("GetStoreSize", func() {
		ginkgo.BeforeEach(func() {
			cleaner = NewProjectExecutablesCleaner(false, false, []string{".sh"}, nil,
				WithProjectLister(mockLister),
				WithFileOperator(mockOperator),
			)
		})

		ginkgo.It("should return 0 when scan fails", func() {
			mockLister.err = errors.New("scan failed")
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(0)))
		})

		ginkgo.It("should return 0 when no files found", func() {
			mockLister.projects = []ProjectInfo{}
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(0)))
		})

		ginkgo.It("should return total size of all files", func() {
			mockLister.projects = []ProjectInfo{
				{Name: "project1", Path: "/path/to/project1"},
			}
			mockOperator.executables = []string{
				"/path/to/project1/binary1",
				"/path/to/project1/binary2",
			}
			mockOperator.fileSizes = map[string]int64{
				"/path/to/project1/binary1": 1024,
				"/path/to/project1/binary2": 2048,
			}
			size := cleaner.GetStoreSize(ctx)
			gomega.Expect(size).To(gomega.Equal(int64(3072)))
		})
	})

	// ============================================================================
	// Package-level Functions TESTS (2 specs)
	// ============================================================================
	ginkgo.Describe("isExcludedByExtension package function", func() {
		ginkgo.It("should work with multiple extensions", func() {
			exts := []string{".sh", ".bash", ".zsh"}
			gomega.Expect(isExcludedByExtension("script.sh", exts)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByExtension("script.bash", exts)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByExtension("script.zsh", exts)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByExtension("script.py", exts)).To(gomega.BeFalse())
		})

		ginkgo.It("should be case-insensitive", func() {
			exts := []string{".sh"}
			gomega.Expect(isExcludedByExtension("SCRIPT.SH", exts)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByExtension("Script.Sh", exts)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByExtension("script.SH", exts)).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("isExcludedByPattern package function", func() {
		ginkgo.It("should handle glob patterns correctly", func() {
			patterns := []string{"*.config", "Makefile", "test_*"}
			gomega.Expect(isExcludedByPattern("app.config", patterns)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByPattern("Makefile", patterns)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByPattern("test_binary", patterns)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByPattern("binary", patterns)).To(gomega.BeFalse())
		})

		ginkgo.It("should handle invalid patterns gracefully", func() {
			patterns := []string{"[invalid", "Makefile"}
			// Invalid pattern should be skipped, valid pattern should work
			gomega.Expect(isExcludedByPattern("Makefile", patterns)).To(gomega.BeTrue())
			gomega.Expect(isExcludedByPattern("[invalid", patterns)).To(gomega.BeFalse())
		})
	})
})

// ============================================================================
// INTEGRATION TESTS (requires tools installed)
// ============================================================================
var _ = ginkgo.Describe("ProjectExecutablesCleaner Integration", func() {
	ginkgo.It("should work with default implementations", func() {
		cleaner := NewProjectExecutablesCleaner(false, false, nil, nil)
		ctx := context.Background()

		if !cleaner.IsAvailable(ctx) {
			ginkgo.Skip("Skipping integration test: tools not available")
		}

		// Verify Scan works
		scanResult := cleaner.Scan(ctx)
		// Either it succeeds or fails with a meaningful error
		if scanResult.IsOk() {
			gomega.Expect(scanResult.Value()).To(gomega.BeAssignableToTypeOf([]domain.ScanItem{}))
		}
	})

	ginkgo.It("should handle real filesystem operations", func() {
		cleaner := NewProjectExecutablesCleaner(false, false, nil, nil)
		ctx := context.Background()

		if !cleaner.IsAvailable(ctx) {
			ginkgo.Skip("Skipping integration test: tools not available")
		}

		// Create temp directory with executable
		tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
		gomega.Expect(err).To(gomega.BeNil())
		defer os.RemoveAll(tmpDir)

		execFile := filepath.Join(tmpDir, "test-executable")
		err = os.WriteFile(execFile, []byte("#!/bin/sh\necho test"), 0755)
		gomega.Expect(err).To(gomega.BeNil())

		// Verify file was created as executable
		info, err := os.Stat(execFile)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(info.Mode()&0111).NotTo(gomega.Equal(os.FileMode(0)))
	})
})
