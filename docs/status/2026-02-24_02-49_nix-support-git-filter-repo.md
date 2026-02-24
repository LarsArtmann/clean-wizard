# Status Report: Nix Support for git-filter-repo

**Date:** 2026-02-24 02:49
**Author:** Crush (AI Assistant)

---

## Summary

Added automatic Nix detection support for `git-filter-repo`, eliminating the Python dependency requirement for users with Nix installed. The system now auto-detects the best available provider for running git-filter-repo.

---

## Changes Made

### New Files

| File | Purpose |
|------|---------|
| `internal/cleaner/githistory_filterrepo.go` | Provider detection logic (system → nix → none) |
| `internal/cleaner/githistory_filterrepo_test.go` | Unit tests for detection |

### Modified Files

| File | Changes |
|------|---------|
| `internal/cleaner/githistory_safety.go` | Uses new `DetectFilterRepoProvider()`, reports provider type in safety report |
| `internal/cleaner/githistory_executor.go` | Uses `BuildFilterRepoCommand()` for provider-aware execution |
| `internal/domain/githistory_types.go` | Added `FilterRepoProvider` field to `GitHistorySafetyReport` |
| `FEATURES.md` | Updated documentation to mention Nix support |

---

## Technical Details

### Provider Detection Priority

1. **System Install** - `git filter-repo --version` succeeds
2. **Nix** - `nix` command available AND `nixpkgs#git-filter-repo` accessible
3. **None** - No provider available (blocks git-history cleaner)

### Execution Paths

| Provider | Command |
|----------|---------|
| System | `git filter-repo <args>` |
| Nix | `nix run nixpkgs#git-filter-repo -- <args>` |

### API Changes

```go
// New type for provider identification
type FilterRepoProvider int

const (
    FilterRepoNone FilterRepoProvider = iota
    FilterRepoSystem
    FilterRepoNix
)

// Detection function (cached after first call)
func DetectFilterRepoProvider() FilterRepoProvider

// Command builder for provider-aware execution
func BuildFilterRepoCommand(ctx context.Context, args []string) *exec.Cmd

// Helper for install hints
func GetInstallHint() string
```

### Safety Report Enhancement

```go
type GitHistorySafetyReport struct {
    // ... existing fields ...
    FilterRepoProvider string `json:"filter_repo_provider,omitempty"` // NEW
}
```

---

## Test Results

```
=== RUN   TestDetectFilterRepoProvider
--- PASS: TestDetectFilterRepoProvider (0.97s)
=== RUN   TestFilterRepoProvider_String
--- PASS: TestFilterRepoProvider_String (0.00s)
=== RUN   TestBuildFilterRepoCommand
--- PASS: TestBuildFilterRepoCommand (0.48s)
=== RUN   TestGetInstallHint
--- PASS: TestGetInstallHint (0.58s)
PASS
```

---

## User Experience Impact

### Before
```
❌ git-filter-repo is not installed. Install it with: pip install git-filter-repo or brew install git-filter-repo
```

### After (with Nix available)
```
✅ Using Nix to run git-filter-repo automatically
```

### After (no provider)
```
❌ git-filter-repo is not installed. Install with: brew install git-filter-repo, or ensure nix is available to use it automatically
```

---

## Why This Matters

- **No Python dependency** for Nix users
- **Zero-configuration** - auto-detected at runtime
- **Backwards compatible** - system install still preferred
- **Better error messages** - clear guidance on installation options

---

## Future Considerations

| Item | Priority | Notes |
|------|----------|-------|
| Add `git-filter-branch` fallback | Low | Deprecated but would provide last-resort option |
| Cache Nix store builds | Low | First run may download git-filter-repo |
| Support BFG Repo-Cleaner via Nix | Low | Alternative tool for very large repos |

---

## Verification

```bash
# Build and test
go build ./...
go test ./internal/cleaner/... -run "FilterRepo" -v

# Manual test
./clean-wizard git-history
# Should no longer show "git-filter-repo is not installed" if Nix is available
```

---

## Related

- Commit: `3334395 feat(cleaner): implement Git history cleaning functionality`
- Plan: `docs/planning/PLAN_GIT_HISTORY_CLEANER.md`
- Issue context: User frustration about Python dependency
