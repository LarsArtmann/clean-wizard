package testhelper

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

var GoCacheFlags = cleaner.GoCacheGOCACHE | cleaner.GoCacheTestCache | cleaner.GoCacheBuildCache

func GoCleanerTest(ctx context.Context, title string) error {
	fmt.Printf("=== %s ===\n\n", title)

	// Test 1: Check Go is available
	goCleaner, err := cleaner.NewGoCleaner(true, false, GoCacheFlags)
	if err != nil {
		return fmt.Errorf("failed to create Go cleaner: %w", err)
	}
	if !goCleaner.IsAvailable(ctx) {
		return fmt.Errorf("Go is not available")
	}
	fmt.Println("✅ Go is available\n")

	// Test 2: Scan for Go caches
	fmt.Println("🔍 Scanning for Go caches...")
	scanResult := goCleaner.Scan(ctx)
	if scanResult.IsErr() {
		return fmt.Errorf("scan failed: %w", scanResult.Error())
	}

	items := scanResult.Value()
	if len(items) == 0 {
		fmt.Println("ℹ️  No Go caches found")
	} else {
		fmt.Printf("✅ Found %d cache location(s):\n", len(items))
		for i, item := range items {
			fmt.Printf("  %d. %s\n", i+1, item.Path)
			fmt.Printf("     Size: %s\n", format.Bytes(int64(item.Size)))
		}
	}
	fmt.Println()

	// Test 3: Dry-run clean
	fmt.Println("🧹 Testing dry-run clean...")
	dryRunCleaner, err := cleaner.NewGoCleaner(true, true, GoCacheFlags)
	if err != nil {
		return fmt.Errorf("failed to create Go cleaner: %w", err)
	}
	cleanResult := dryRunCleaner.Clean(ctx)
	if cleanResult.IsErr() {
		return fmt.Errorf("clean failed: %w", cleanResult.Error())
	}

	result := cleanResult.Value()
	fmt.Printf("✅ Dry-run complete:\n")
	fmt.Printf("   Items would be cleaned: %d\n", result.ItemsRemoved)
	fmt.Printf("   Strategy: %s\n", result.Strategy)
	fmt.Println()

	fmt.Println("✅ All tests passed!")
	fmt.Println("\n📋 Supported Go cache types:")
	fmt.Println("   ✓ go clean -cache     (GOCACHE)")
	fmt.Println("   ✓ go clean -testcache  (GOTESTCACHE)")
	fmt.Println("   ✓ go clean -modcache  (GOMODCACHE)")
	fmt.Println("   ✓ go-build* folders   (Build cache)")
	fmt.Println("\n❌ NOT supported:")
	fmt.Println("   ✗ go clean -fuzzcache (GOFUZZCACHE) - Missing implementation")

	return nil
}