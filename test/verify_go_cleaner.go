package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Go Cache Cleaner Verification ===\n")

	// Test 1: Check Go is available
	goCleaner := cleaner.NewGoCleaner(true, false, true, true, true, true)
	if !goCleaner.IsAvailable(ctx) {
		fmt.Println("❌ Go is not available")
		os.Exit(1)
	}
	fmt.Println("✅ Go is available\n")

	// Test 2: Scan for Go caches
	fmt.Println("🔍 Scanning for Go caches...")
	scanResult := goCleaner.Scan(ctx)
	if scanResult.IsErr() {
		fmt.Printf("❌ Scan failed: %v\n", scanResult.Error())
		os.Exit(1)
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
	dryRunCleaner := cleaner.NewGoCleaner(true, true, true, true, true, true)
	cleanResult := dryRunCleaner.Clean(ctx)
	if cleanResult.IsErr() {
		fmt.Printf("❌ Clean failed: %v\n", cleanResult.Error())
		os.Exit(1)
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
}
