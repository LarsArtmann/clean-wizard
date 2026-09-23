package cleaner

import (
	"context"
)

// AssertNoItemsToCleanResult verifies that a cleaner returns the expected conservative
// result when there are no items to clean. This consolidates duplicate test patterns
// across Ginkgo test files.
func AssertNoItemsToCleanResult(ctx context.Context, cleaner Cleaner, setupEmptyState func()) {
	GinkgoNoItemsToCleanTest(ctx, cleaner, setupEmptyState)
}
