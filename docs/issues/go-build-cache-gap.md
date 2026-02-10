# Go Build Cache Gap Analysis

**Issue:** Clean Wizard misses macOS-specific Go build cache location
**SystemNix Location:** `/private/var/folders/*/T/go-build*`
**Clean Wizard Coverage:** `/tmp/go-build*` + `$HOME/Library/Caches/go-build*`

---

## The Problem

Go uses `os.TempDir()` for build caches on macOS, which returns `/private/var/folders/*/T` rather than `/tmp`.

### SystemNix Command

```bash
find /private/var/folders/07/y9f_lh8s1zq2kr67_k94w22h0000gn/T -name "go-build*" -type d -print0 | xargs -0 trash
```

### Clean Wizard Current Behavior

**File:** `internal/cleaner/golang_cache_cleaner.go:132-141`

```go
// cleanGoBuildCache removes go-build* folders.
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"
	tempDir := "/tmp"
	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		tempDir = homeDir + "/Library/Caches"
	}

	// Use shell globbing to find build cache folders
	matches, err := filepath.Glob(tempDir + "/" + buildCachePattern)
	// ...
}
```

**Coverage:**

- ✅ `/tmp/go-build*`
- ✅ `$HOME/Library/Caches/go-build*`
- ❌ `/private/var/folders/*/T/go-build*` (MISSING)

---

## macOS Temp Directory Structure

macOS has multiple temp directory locations:

| Location                   | Purpose            | Example                     |
| -------------------------- | ------------------ | --------------------------- |
| `/tmp`                     | Standard Unix temp | Symlinked to `/private/tmp` |
| `/private/var/folders/*/T` | App-specific temp  | Go build cache, etc.        |
| `$HOME/Library/Caches/`    | User caches        | Application caches          |
| `/var/folders/*/T`         | Legacy             | Some apps still use this    |

**Go on macOS:** Uses `os.TempDir()` which returns `/var/folders/.../T` (or `/private/var/folders/.../T` on newer systems).

---

## Impact Assessment

| Scenario                                     | SystemNix      | Clean Wizard   | Gap                  |
| -------------------------------------------- | -------------- | -------------- | -------------------- |
| Go build cache in `/tmp`                     | ✅ Cleans      | ✅ Cleans      | 0                    |
| Go build cache in `~/Library/Caches/`        | ❌ Not cleaned | ✅ Cleans      | SystemNix gap        |
| Go build cache in `/private/var/folders/*/T` | ✅ Cleans      | ❌ Not cleaned | **Clean Wizard gap** |

**Estimated Space at Risk:**

- Go build caches can be **hundreds of MB to several GB**
- Frequent Go builds → larger cache footprint
- Large projects (Kubernetes, etc.) → multi-GB caches

---

## Fix Implementation

### Option A: Use Go's `os.TempDir()` (Recommended)

```go
import "os"

// cleanGoBuildCache removes go-build* folders.
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"

	// Use Go's TempDir which returns the correct macOS temp location
	tempDir := os.TempDir()

	// Also check common cache locations
	locations := []string{
		tempDir,
		"/tmp",
	}

	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		locations = append(locations, homeDir+"/Library/Caches")
	}

	var allMatches []string
	for _, dir := range locations {
		matches, err := filepath.Glob(filepath.Join(dir, buildCachePattern))
		if err != nil {
			continue
		}
		allMatches = append(allMatches, matches...)
	}
	// ... rest of cleanup logic
}
```

**Pros:**

- Uses Go's built-in temp dir detection
- Cross-platform compatible
- Future-proof

**Cons:**

- Requires Go 1.17+ for proper macOS temp handling

---

### Option B: Walk `/private/var/folders/` (Simple)

```go
// cleanGoBuildCache removes go-build* folders.
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	buildCachePattern := "go-build*"

	var locations []string

	// Check macOS-specific temp locations
	if entries, err := os.ReadDir("/private/var/folders"); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == ".." {
				continue
			}
			tPath := filepath.Join("/private/var/folders", entry.Name(), "T")
			if _, err := os.Stat(tPath); err == nil {
				locations = append(locations, tPath)
			}
		}
	}

	// Fallback to common locations
	locations = append(locations, "/tmp")
	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
		locations = append(locations, homeDir+"/Library/Caches")
	}

	// ... rest of cleanup logic
}
```

**Pros:**

- Explicit macOS temp dir handling
- Handles edge cases

**Cons:**

- Platform-specific code
- Slightly more complex

---

### Option C: Run `go env GOCACHE` + Walk Parent (Most Accurate)

```go
// cleanGoBuildCache removes go-build* folders.
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	// Get the actual GOCACHE location
	gocacheDir, err := gcc.helper.getGoEnv(ctx, "GOCACHE")
	if err == nil && gocacheDir != "" {
		// Clean the GOCACHE directory directly
		return gcc.cleanGoCacheDir(gocacheDir)
	}

	// Fallback to pattern matching
	// ... existing logic
}
```

**Pros:**

- Uses Go's configured cache location
- Most accurate

**Cons:**

- Requires Go to be available
- May miss build caches outside GOCACHE

---

## Recommended Fix: Option A (os.TempDir)

This is the cleanest solution that works across platforms and uses Go's built-in logic.

### Changes Required

**File:** `internal/cleaner/golang_cache_cleaner.go`

**Lines 132-141:**

```diff
 // cleanGoBuildCache removes go-build* folders.
 func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
 	buildCachePattern := "go-build*"
-	tempDir := "/tmp"
-	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
-		tempDir = homeDir + "/Library/Caches"
-	}
+
+	// Use Go's TempDir for cross-platform compatibility
+	tempDir := os.TempDir()
+
+	// Check multiple locations for comprehensive coverage
+	locations := []string{tempDir, "/tmp"}
+	if homeDir := gcc.helper.getHomeDir(); homeDir != "" {
+		locations = append(locations, homeDir+"/Library/Caches")
+	}

 	// Use shell globbing to find build cache folders
-	matches, err := filepath.Glob(tempDir + "/" + buildCachePattern)
+	var allMatches []string
+	for _, dir := range locations {
+		matches, err := filepath.Glob(filepath.Join(dir, buildCachePattern))
+		if err != nil {
+			continue
+		}
+		allMatches = append(allMatches, matches...)
+	}
+	matches := allMatches
```

**Import change (line 3-10):**

```diff
 import (
 	"context"
 	"fmt"
+	"os"
 	"os/exec"
 	"path/filepath"
 	"time"
 	// ...
 )
```

---

## Testing

### Test Cases

```go
// TestGoBuildCache_MacOSLocations
// TestGoBuildCache_CleansPrivateVarFolders
// TestGoBuildCache_AllLocations
```

### Manual Verification

```bash
# Before fix - show what's missed
echo "=== Clean Wizard would miss ==="
find /private/var/folders -name "go-build*" -type d 2>/dev/null | head -5
du -sh /private/var/folders -name "go-build*" -type d 2>/dev/null | tail -5

echo ""
echo "=== SystemNix cleans (all locations) ==="
find /tmp -name "go-build*" -type d 2>/dev/null | head -5
du -sh /tmp -name "go-build*" -type d 2>/dev/null | tail -5
```

---

## Verification

After fix, run:

```bash
# 1. Create some test build cache
go build -o /tmp/test-build-cache ./...

# 2. Show current state
clean-wizard clean --mode quick --dry-run

# 3. Clean and verify
clean-wizard clean --mode quick

# 4. Check what was freed
echo "=== Should show Go build cache freed space ==="
```

---

## Additional Consideration: Age-Based Cleanup

SystemNix doesn't have age-based cleanup for Go build cache. Clean Wizard could add this:

```go
// If we wanted age-based cleanup:
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
	// ... existing logic ...

	// Add age-based filtering
	olderThan := 7 * 24 * time.Hour // 7 days
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) < olderThan {
			continue // Skip recent builds
		}

		// Clean old build caches
	}
}
```

**Recommendation:** Keep simple for now, add age-based if requested.

---

## Summary

| Aspect             | Value                                                    |
| ------------------ | -------------------------------------------------------- |
| **Issue**          | Clean Wizard misses `/private/var/folders/*/T/go-build*` |
| **Impact**         | May miss hundreds of MB to GB of cache                   |
| **Fix**            | Use `os.TempDir()` + check common locations              |
| **Effort**         | ~1 hour                                                  |
| **Risk**           | Low                                                      |
| **Files Modified** | `internal/cleaner/golang_cache_cleaner.go`               |

---

_Document created: 2026-02-09_
