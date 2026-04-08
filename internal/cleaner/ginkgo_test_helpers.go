package cleaner

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// assertResultOk verifies that a result is OK and panics if not.
// This eliminates duplicate assertion code across multiple Ginkgo test files.
//
// Parameters:
//   - result: The result to check (must have IsOk() method)
//
// Usage:
//
//	result := cleaner.Scan(ctx)
//	assertResultOk(result)
func assertResultOk[T any](result result.Result[T]) {
	gomega.Expect(result.IsOk()).To(gomega.BeTrue())
}

// GinkgoNoItemsToCleanTest tests the "no items to clean" scenario for Ginkgo-based cleaner tests.
// This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - ctx: The test context
//   - cleaner: The cleaner instance to test (must have Clean method returning result.Result[domain.CleanResult])
//   - setupEmptyState: Function that sets up the mock to return empty results
//
// Usage:
//
//	ginkgo.Context("with no items to clean", func() {
//		ginkgo.It("should return conservative result when no items found", func() {
//			GinkgoNoItemsToCleanTest(ctx, cleaner, func() {
//				mockScanner.binaries = []BinaryInfo{}
//			})
//		})
//	})
func GinkgoNoItemsToCleanTest(ctx context.Context, cleaner interface {
	Clean(context.Context) result.Result[domain.CleanResult]
}, setupEmptyState func(),
) {
	setupEmptyState()

	result := cleaner.Clean(ctx)
	assertResultOk(result)
	cleanResult := result.Value()
	gomega.Expect(cleanResult.ItemsRemoved).To(gomega.Equal(uint(0)))
	gomega.Expect(cleanResult.Strategy).To(gomega.Equal(domain.StrategyConservative))
}

// GinkgoValidateValidSettingsTest tests that ValidateSettings returns no error
// for valid settings. This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - cleaner: The cleaner instance with ValidateSettings method
//   - settings: The valid OperationSettings to validate
//
// Usage:
//
//	ginkgo.It("should return nil for valid settings", func() {
//		settings := &domain.OperationSettings{
//			GitHistory: &domain.GitHistorySettings{
//				MaxFiles: 50,
//			},
//		}
//		GinkgoValidateValidSettingsTest(cleaner, settings)
//	})
func GinkgoValidateValidSettingsTest(
	cleaner CleanerWithSettings,
	settings *domain.OperationSettings,
) {
	err := cleaner.ValidateSettings(settings)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// GinkgoErrorPropagationTest tests that errors from dependencies are properly propagated
// through Scan or Clean methods. This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - setupError: Function that sets up the mock to return an error
//   - operation: Function that performs the operation (Scan or Clean) and returns a result
//
// Usage:
//
//	ginkgo.Context("when ListProjects fails", func() {
//		ginkgo.It("should return error when project listing fails", func() {
//			GinkgoErrorPropagationTest(
//				func() { mockLister.err = errors.New("failed to list projects") },
//				func() result.Result[[]domain.ScanItem] { return cleaner.Scan(ctx) },
//			)
//		})
//	})
func GinkgoErrorPropagationTest[T any](setupError func(), operation func() result.Result[T]) {
	setupError()

	res := operation()
	gomega.Expect(res.IsErr()).To(gomega.BeTrue())
	gomega.Expect(res.Error()).To(gomega.HaveOccurred())
}

// GinkgoErrorPropagationContext creates a ginkgo.Context and ginkgo.It block for testing
// error propagation scenarios. This eliminates the duplicate Context/It wrapper pattern
// across multiple Ginkgo test files.
//
// Parameters:
//   - contextName: The name for the ginkgo.Context (e.g., "when ListProjects fails")
//   - itName: The name for the ginkgo.It (e.g., "should return error when project listing fails")
//   - setupError: Function that sets up the mock to return an error
//   - operation: Function that performs the operation and returns a result
//
// Usage:
//
//	GinkgoErrorPropagationContext(
//		"when ListProjects fails",
//		"should return error when project listing fails",
//		func() { mockLister.err = errors.New("failed to list projects") },
//		func() result.Result[[]domain.ScanItem] { return cleaner.Scan(ctx) },
//	)
func GinkgoErrorPropagationContext[T any](
	contextName string,
	itName string,
	setupError func(),
	operation func() result.Result[T],
) {
	ginkgo.Context(contextName, func() {
		ginkgo.It(itName, func() {
			GinkgoErrorPropagationTest(setupError, operation)
		})
	})
}

// GinkgoNoItemsToScanTest tests the "no items to scan" scenario for Ginkgo-based cleaner tests.
// This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - ctx: The test context
//   - cleaner: The cleaner instance to test (must have Scan method returning result.Result[[]domain.ScanItem])
//   - setupEmptyState: Function that sets up the mock to return empty results
//
// Usage:
//
//	ginkgo.Context("when no projects found", func() {
//		ginkgo.It("should return empty slice when no projects", func() {
//			GinkgoNoItemsToScanTest(ctx, cleaner, func() {
//				mockLister.projects = []ProjectInfo{}
//			})
//		})
//	})
func GinkgoNoItemsToScanTest(ctx context.Context, cleaner interface {
	Scan(context.Context) result.Result[[]domain.ScanItem]
}, setupEmptyState func(),
) {
	setupEmptyState()

	result := cleaner.Scan(ctx)
	gomega.Expect(result.IsOk()).To(gomega.BeTrue())
	gomega.Expect(result.Value()).To(gomega.BeEmpty())
}

// GinkgoValidateEmptySettingsTest tests that ValidateSettings returns no error
// for empty OperationSettings (no module-specific settings configured).
// This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - cleaner: The cleaner instance with ValidateSettings method
//   - itName: The name for the ginkgo.It block (e.g., "should return nil when CompiledBinaries is nil")
//
// Usage:
//
//	ginkgo.Context("with empty OperationSettings", func() {
//		GinkgoValidateEmptySettingsTest(cleaner, "should return nil when CompiledBinaries is nil")
//	})
func GinkgoValidateEmptySettingsTest(cleaner CleanerWithSettings, itName string) {
	ginkgo.It(itName, func() {
		settings := &domain.OperationSettings{}
		err := cleaner.ValidateSettings(settings)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})
}

// GinkgoValidateEmptySettingsContext creates a ginkgo.Context and ginkgo.It block
// for testing that ValidateSettings returns no error for empty OperationSettings.
// This eliminates the duplicate Context wrapper pattern across multiple Ginkgo test files.
//
// Parameters:
//   - cleaner: The cleaner instance with ValidateSettings method
//   - itName: The name for the ginkgo.It block (e.g., "should return nil when CompiledBinaries is nil")
//
// Usage:
//
//	GinkgoValidateEmptySettingsContext(cleaner, "should return nil when CompiledBinaries is nil")
func GinkgoValidateEmptySettingsContext(cleaner CleanerWithSettings, itName string) {
	ginkgo.Context("with empty OperationSettings", func() {
		GinkgoValidateEmptySettingsTest(cleaner, itName)
	})
}

// GinkgoValidateNilSettingsTest tests that ValidateSettings returns no error for nil settings.
// This eliminates duplicate test code across multiple Ginkgo test files.
//
// Parameters:
//   - cleaner: The cleaner instance with ValidateSettings method
//
// Usage:
//
//	ginkgo.It("should return nil for nil settings", func() {
//		GinkgoValidateNilSettingsTest(cleaner)
//	})
func GinkgoValidateNilSettingsTest(cleaner CleanerWithSettings) {
	err := cleaner.ValidateSettings(nil)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

// StandardTestBinaries returns a slice of BinaryInfo for testing.
// This eliminates duplicate binary setup code across compiled binaries tests.
//
// Returns two test binaries with sizes 20MB and 15MB respectively.
func StandardTestBinaries() []BinaryInfo {
	return []BinaryInfo{
		{Path: "/path/to/binary1", Size: 20 * 1024 * 1024, Category: CategoryTest},
		{Path: "/path/to/binary2", Size: 15 * 1024 * 1024, Category: CategoryBin},
	}
}

// StandardTestBinariesTotalSize returns the total size of StandardTestBinaries.
func StandardTestBinariesTotalSize() int64 {
	return 35 * 1024 * 1024
}
