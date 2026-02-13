# Size Reporting Fixes - Execution Plan

> **Date:** 2026-02-11 21:00
> **Status:** READY TO EXECUTE
> **Total Effort:** ~6 hours
> **Strategic Direction:** Technical Debt Reduction (Option B)

---

## Overview

This plan fixes inaccurate size reporting across 5 cleaners to provide users with accurate dry-run estimates and actual bytes freed data.

**Impact:** HIGH - Immediate user trust and UX improvement
**Risk:** LOW - Localized changes, well-tested

---

## Phase 1: Docker Size Reporting Fix (2 hours)

### Task 1.1: Create Size Parsing Utility (45 min)

**File:** `internal/cleaner/docker_parsing.go` (new)

Create functions to parse Docker prune output:

```go
// ParseDockerReclaimedSpace extracts "Total reclaimed space: X" from docker prune output
func ParseDockerReclaimedSpace(output string) (int64, error) {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Total reclaimed space:") {
            // Extract: "Total reclaimed space: 2.5GB"
            parts := strings.Split(line, ":")
            if len(parts) != 2 {
                continue
            }
            sizeStr := strings.TrimSpace(parts[1])
            // Parse size string (e.g., "2.5GB", "100MB", "1.84kB")
            return ParseDockerSize(sizeStr)
        }
    }
    return 0, nil // No space found (0 is valid)
}

// ParseDockerSize converts Docker size string to bytes
func ParseDockerSize(sizeStr string) (int64, error) {
    // Remove any whitespace
    sizeStr = strings.TrimSpace(sizeStr)

    // Handle "0B" case
    if sizeStr == "0B" || sizeStr == "0" {
        return 0, nil
    }

    // Parse number and unit
    var number float64
    var unit string
    _, err := fmt.Sscanf(sizeStr, "%f%s", &number, &unit)
    if err != nil {
        return 0, fmt.Errorf("invalid size format: %s", sizeStr)
    }

    // Convert to bytes
    switch strings.ToLower(unit) {
    case "b":
        return int64(number), nil
    case "kb":
        return int64(number * 1024), nil
    case "mb":
        return int64(number * 1024 * 1024), nil
    case "gb":
        return int64(number * 1024 * 1024 * 1024), nil
    case "tb":
        return int64(number * 1024 * 1024 * 1024 * 1024), nil
    default:
        return 0, fmt.Errorf("unknown size unit: %s", unit)
    }
}
```

### Task 1.2: Update pruneDocker to Use Parser (30 min)

**File:** `internal/cleaner/docker.go` (line 293-300)

Change from hardcoded `FreedBytes: 0` to parsed value:

```go
// Before:
return result.Ok(domain.CleanResult{
    FreedBytes:   0, // Bytes freed unknown from prune output
    ...
})

// After:
bytesFreed, err := ParseDockerReclaimedSpace(string(output))
if err != nil && dc.verbose {
    fmt.Printf("  Warning: failed to parse reclaimed space: %v\n", err)
}
return result.Ok(domain.CleanResult{
    FreedBytes:   uint64(bytesFreed),
    ...
})
```

### Task 1.3: Add Tests for Size Parsing (30 min)

**File:** `internal/cleaner/docker_test.go` (add tests)

Test cases:

- Valid sizes: "1.84kB", "13.5 MB", "2.5GB", "0B"
- Invalid sizes: "", "invalid", "1.5XB"
- Full prune output parsing

### Task 1.4: Run Tests and Verify (15 min)

- Run `go test ./internal/cleaner -v`
- Test with actual docker if available
- Verify parsing works with real output

---

## Phase 2: Cargo Size Reporting Fix (1 hour)

### Task 2.1: Calculate Actual Bytes Freed (30 min)

**File:** `internal/cleaner/cargo.go` (lines 178-224)

Modify `executeCargoCleanCommand` to calculate bytes:

```go
// Before:
return result.Ok(domain.CleanResult{
    FreedBytes:   0,
    ...
})

// After:
// Get cache directory size before and after
var bytesFreed int64
cacheDir := cc.getCargoCacheDir()
if cacheDir != "" {
    beforeSize := GetDirSize(cacheDir)
    // Execute clean command
    cmd := cc.execWithTimeout(ctx, cmdName, args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return result.Err[domain.CleanResult](fmt.Errorf(errorFormat, err, string(output)))
    }
    afterSize := GetDirSize(cacheDir)
    bytesFreed = beforeSize - afterSize
    if bytesFreed < 0 {
        bytesFreed = 0 // Ensure non-negative
    }
}
```

### Task 2.2: Add getCargoCacheDir Helper (15 min)

```go
func (cc *CargoCleaner) getCargoCacheDir() string {
    cargoHome := os.Getenv("CARGO_HOME")
    if cargoHome == "" {
        if homeDir, err := GetHomeDir(); err == nil && homeDir != "" {
            cargoHome = homeDir + "/.cargo"
        }
    }
    return cargoHome
}
```

### Task 2.3: Update Clean Method (15 min)

Update `executeCargoCleanCommand` to use `uint64(bytesFreed)`.

---

## Phase 3: Nix Size Reporting Fix (1 hour)

### Task 3.1: Use Real Store Size Data (30 min)

**File:** `internal/cleaner/nix.go` (line 132)

Replace hardcoded 50MB estimate:

```go
// Before:
estimatedBytes := int64(toRemove * 50 * 1024 * 1024)

// After:
storeSize := nc.GetStoreSize(ctx)
generationsResult := nc.ListGenerations(ctx)
if generationsResult.IsErr() {
    return conversions.ToCleanResultFromError(generationsResult.Error())
}
generations := generationsResult.Value()
// Calculate average generation size
avgSize := storeSize / int64(len(generations))
estimatedBytes := avgSize * int64(toRemove)
```

### Task 3.2: Add Test for Real Size Calculation (30 min)

Test that average size calculation works correctly.

---

## Phase 4: Go Size Reporting Fix (1 hour)

### Task 4.1: Parse golangci-lint Output (30 min)

**File:** `internal/cleaner/golang_cleaner.go`

Find where golangci-lint cache is cleaned and parse output:

```go
// Example output: "golangci-lint cache cleaned"
// Need to calculate directory size before/after
lintCachePath := getGolangciLintCachePath()
beforeSize := GetDirSize(lintCachePath)
// Execute clean
afterSize := GetDirSize(lintCachePath)
bytesFreed := beforeSize - afterSize
```

### Task 4.2: Parse Go Build Cache Output (30 min)

**File:** `internal/cleaner/golang_cache_cleaner.go`

Parse `go clean -cache` output or calculate directory size before/after.

---

## Phase 5: Node Packages Size Reporting Fix (1 hour)

### Task 5.1: Parse npm Cache Clean Output (20 min)

**File:** `internal/cleaner/nodepackages.go`

Calculate npm cache directory size before/after clean.

### Task 5.2: Parse yarn Cache Clean Output (20 min)

Calculate yarn cache directory size before/after clean.

### Task 5.3: Parse pnpm Cache Clean Output (20 min)

Calculate pnpm cache directory size before/after clean.

---

## Execution Checklist

### Before Starting

- [ ] All tests passing (`just test`)
- [ ] Code compiles (`go build ./...`)
- [ ] Current state committed

### Phase 1: Docker

- [ ] Create `docker_parsing.go` with size parser
- [ ] Update `pruneDocker` to use parser
- [ ] Add comprehensive tests
- [ ] Verify tests pass
- [ ] Commit "feat(docker): add accurate size reporting"

### Phase 2: Cargo

- [ ] Implement `getCargoCacheDir` helper
- [ ] Update `executeCargoCleanCommand` to calculate bytes
- [ ] Test with mock and real cargo if available
- [ ] Verify tests pass
- [ ] Commit "feat(cargo): add accurate size reporting"

### Phase 3: Nix

- [ ] Update dry-run to use real store size
- [ ] Calculate average generation size
- [ ] Add tests
- [ ] Verify tests pass
- [ ] Commit "feat(nix): use real store size for estimates"

### Phase 4: Go

- [ ] Parse golangci-lint output
- [ ] Parse go build cache output
- [ ] Calculate real bytes freed
- [ ] Add tests
- [ ] Verify tests pass
- [ ] Commit "feat(golang): add accurate size reporting"

### Phase 5: Node

- [ ] Calculate npm cache size
- [ ] Calculate yarn cache size
- [ ] Calculate pnpm cache size
- [ ] Add tests
- [ ] Verify tests pass
- [ ] Commit "feat(nodepackages): add accurate size reporting"

### Final Verification

- [ ] All tests pass (`just test`)
- [ ] Code compiles (`go build ./...`)
- [ ] Verify size reporting in dry-run
- [ ] Create summary document
- [ ] Commit final summary

---

## Success Criteria

1. **All cleaners report accurate bytes freed** (no more hardcoded estimates)
2. **Docker parses "Total reclaimed space"** correctly
3. **All tests pass** (100% test success rate)
4. **No regressions** (existing functionality intact)
5. **User-visible improvement** (dry-run shows real estimates)

---

## Rollback Plan

If any phase causes issues:

1. Revert the specific commit
2. Run `just test` to verify rollback
3. Report issue and continue with other phases

---

## Notes

- Use existing `GetDirSize` function from `internal/cleaner/fsutil.go`
- Handle cases where cache directories don't exist gracefully
- Ensure bytes freed is always non-negative
- Add verbose logging for debugging
- Keep dry-run estimates separate from actual clean values
