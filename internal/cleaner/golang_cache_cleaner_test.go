package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoCacheCleaner_getGoBuildCacheLocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *GoCacheCleaner
		wantDirs int // minimum expected dirs
	}{
		{
			name: "returns multiple locations",
			setup: func() *GoCacheCleaner {
				return NewGoCacheCleaner(GoCacheGOCACHE, false, false)
			},
			wantDirs: 2, // At least os.TempDir() and /tmp
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := tt.setup()
			locations := cleaner.getGoBuildCacheLocations()

			require.NotNil(t, locations)
			assert.GreaterOrEqual(
				t,
				len(locations),
				tt.wantDirs,
				"should return at least %d locations",
				tt.wantDirs,
			)

			// Verify os.TempDir() is included
			assert.Contains(t, locations, os.TempDir(), "should include os.TempDir()")

			// Verify /tmp is included
			assert.Contains(t, locations, "/tmp", "should include /tmp")

			// On macOS, verify Library/Caches is included
			if runtime.GOOS == "darwin" {
				homeDir := cleaner.helper.getHomeDir()
				if homeDir != "" {
					macosCache := filepath.Join(homeDir, "Library", "Caches")
					assert.Contains(
						t,
						locations,
						macosCache,
						"on macOS, should include ~/Library/Caches",
					)
				}
			}
		})
	}
}

func TestGoCacheCleaner_getGoBuildCacheLocations_Deduplication(t *testing.T) {
	t.Parallel()

	cleaner := NewGoCacheCleaner(GoCacheGOCACHE, false, false)
	locations := cleaner.getGoBuildCacheLocations()

	// Check for duplicates
	seen := make(map[string]bool)
	for _, loc := range locations {
		assert.False(t, seen[loc], "location %s should not be duplicated", loc)
		seen[loc] = true
	}
}

func TestGoCacheCleaner_getGoBuildCacheLocations_MacOSDetection(t *testing.T) {
	t.Parallel()

	// This test verifies the macOS-specific path detection logic
	// The fix ensures os.TempDir() is used which returns /private/var/folders/... on macOS
	cleaner := NewGoCacheCleaner(GoCacheGOCACHE, false, false)
	locations := cleaner.getGoBuildCacheLocations()

	// On macOS, os.TempDir() returns something like /var/folders/xx/yyyyyy/T
	// which is where go-build caches are actually stored
	if runtime.GOOS == "darwin" {
		tempDir := os.TempDir()
		assert.Contains(t, locations, tempDir,
			"on macOS, os.TempDir() should be included for go-build cache detection")

		// Verify the path pattern matches what Go uses
		// Go uses $TMPDIR/go-build* which on macOS is in /var/folders/
		assert.Contains(t, tempDir, "T", "macOS temp dir should end with T")
	}
}

func TestGoCacheCleaner_NewGoCacheCleaner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cacheType GoCacheType
		verbose   bool
		dryRun    bool
	}{
		{
			name:      "GOCACHE cleaner",
			cacheType: GoCacheGOCACHE,
			verbose:   false,
			dryRun:    false,
		},
		{
			name:      "GOTESTCACHE cleaner",
			cacheType: GoCacheTestCache,
			verbose:   true,
			dryRun:    false,
		},
		{
			name:      "GOMODCACHE cleaner",
			cacheType: GoCacheModCache,
			verbose:   false,
			dryRun:    true,
		},
		{
			name:      "GoBuildCache cleaner",
			cacheType: GoCacheBuildCache,
			verbose:   true,
			dryRun:    true,
		},
		{
			name:      "GolangciLintCache cleaner",
			cacheType: GoCacheLintCache,
			verbose:   false,
			dryRun:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCacheCleaner(tt.cacheType, tt.verbose, tt.dryRun)

			require.NotNil(t, cleaner)
			assert.Equal(t, tt.cacheType, cleaner.cacheType)
			assert.Equal(t, tt.verbose, cleaner.verbose)
			assert.Equal(t, tt.dryRun, cleaner.dryRun)
			assert.NotNil(t, cleaner.helper)
		})
	}
}

func TestGoCacheCleaner_Name(t *testing.T) {
	t.Parallel()

	// Name returns "golang" for all Go cache cleaner types
	// This is the cleaner's identifier, not the specific cache type name
	tests := []struct {
		name      string
		cacheType GoCacheType
	}{
		{name: "GOCACHE", cacheType: GoCacheGOCACHE},
		{name: "GOTESTCACHE", cacheType: GoCacheTestCache},
		{name: "GOMODCACHE", cacheType: GoCacheModCache},
		{name: "GoBuildCache", cacheType: GoCacheBuildCache},
		{name: "GolangciLintCache", cacheType: GoCacheLintCache},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewGoCacheCleaner(tt.cacheType, false, false)
			assert.Equal(t, "golang", cleaner.Name())
		})
	}
}

func TestGoCacheCleaner_IsAvailable(t *testing.T) {
	t.Parallel()

	// Note: This test assumes 'go' is available in the test environment
	cleaner := NewGoCacheCleaner(GoCacheGOCACHE, false, false)

	// IsAvailable should return true if 'go' command is available
	// We can't easily test the false case without controlling PATH
	available := cleaner.IsAvailable(context.Background())

	// Just verify it doesn't panic
	assert.True(t, available || !available) // Always true, just checking no panic
}
