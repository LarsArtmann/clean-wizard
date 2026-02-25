# Status Report: Git History Nix Integration Issues

**Date:** 2026-02-25 01:44
**Status:** 🔴 CRITICAL BUGS IDENTIFIED
**Component:** `git-history` command, nix integration, git-filter-repo execution

---

## Executive Summary

The `clean-wizard git-history` command has **two critical bugs** preventing it from working with nix:

1. **Execution Failure**: `--protect-blobs-from HEAD` flag doesn't exist in git-filter-repo
2. **Detection Caching**: Nix detection failures are cached permanently per process

---

## Problem Statement

### User Experience

```bash
# First run - nix detected, but execution fails
❌ Error: git-filter-repo: error: unrecognized arguments: --protect-blobs-from HEAD

# Second run - nix NOT detected (cached failure)
❌ git-filter-repo is not installed. Install with: brew install git-filter-repo
```

### Evidence

**Nix IS working correctly:**

```bash
$ which nix
/run/current-system/sw/bin/nix

$ nix eval --raw nixpkgs#git-filter-repo.name
git-filter-repo-2.47.0
```

---

## Root Cause Analysis

### Bug #1: Invalid Flag `--protect-blobs-from`

| Aspect       | Details                                                                     |
| ------------ | --------------------------------------------------------------------------- |
| **Error**    | `git-filter-repo: error: unrecognized arguments: --protect-blobs-from HEAD` |
| **Cause**    | Flag doesn't exist in git-filter-repo 2.47.0                                |
| **Impact**   | Execution always fails when using nix                                       |
| **Location** | Unknown - needs investigation                                               |

### Bug #2: Detection Caching

| Aspect      | Details                                                              |
| ----------- | -------------------------------------------------------------------- |
| **File**    | `internal/cleaner/githistory_filterrepo.go:34-49`                    |
| **Problem** | `sync.Once` caches the result of `detectProvider()` forever          |
| **Impact**  | If detection fails once, it stays failed for entire process lifetime |
| **Code**    | See below                                                            |

```go
// filterRepoDetector caches the detected provider.
type filterRepoDetector struct {
    once     sync.Once
    provider FilterRepoProvider
}

func DetectFilterRepoProvider() FilterRepoProvider {
    detector.once.Do(func() {
        detector.provider = detectProvider() // Cached forever, even on failure
    })
    return detector.provider
}
```

---

## Code Analysis

### Current Architecture

```
githistory_filterrepo.go
├── FilterRepoProvider (enum: None, System, Nix)
├── DetectFilterRepoProvider() → cached detection
├── isNixAvailable() → runs: nix eval --raw nixpkgs#git-filter-repo.name
└── BuildFilterRepoCommand() → builds: nix run nixpkgs#git-filter-repo -- <args>

githistory_safety.go
├── isFilterRepoAvailable() → calls DetectFilterRepoProvider()
└── Check() → adds blocker if FilterRepoNone

githistory_executor.go
├── runFilterRepo() → calls BuildFilterRepoCommand()
└── Adds --protect-blobs-from HEAD (BUG LOCATION UNKNOWN)
```

### Detection Flow

```
1. isSystemInstallAvailable()
   └── git filter-repo --version

2. isNixAvailable()
   ├── exec.LookPath("nix")
   └── nix eval --raw nixpkgs#git-filter-repo.name

3. Return FilterRepoNone if both fail
```

---

## What's Working ✅

| Component       | Status | Notes                                                 |
| --------------- | ------ | ----------------------------------------------------- |
| Nix binary      | ✅     | Available at `/run/current-system/sw/bin/nix`         |
| Nixpkgs access  | ✅     | Can evaluate `nixpkgs#git-filter-repo.name`           |
| Command builder | ✅     | Correctly builds `nix run nixpkgs#git-filter-repo --` |
| Safety checker  | ✅     | Correctly calls `DetectFilterRepoProvider()`          |

---

## What's Broken ❌

| Component                   | Status | Issue                                    |
| --------------------------- | ------ | ---------------------------------------- |
| `--protect-blobs-from` flag | ❌     | Doesn't exist in git-filter-repo         |
| Detection caching           | ❌     | Caches failures permanently              |
| Error messaging             | ⚠️     | Doesn't explain WHY nix detection failed |
| Verbose logging             | ⚠️     | No debug output for detection process    |

---

## Required Fixes

### Priority 1: CRITICAL

| #   | Task                                        | File                       | Effort |
| --- | ------------------------------------------- | -------------------------- | ------ |
| 1   | Find and remove `--protect-blobs-from HEAD` | Unknown                    | 30min  |
| 2   | Fix caching to not cache failures           | `githistory_filterrepo.go` | 15min  |

### Priority 2: HIGH

| #   | Task                                   | File                       | Effort |
| --- | -------------------------------------- | -------------------------- | ------ |
| 3   | Add verbose debug output for detection | `githistory_filterrepo.go` | 15min  |
| 4   | Add error context when detection fails | `githistory_filterrepo.go` | 10min  |

### Priority 3: MEDIUM

| #   | Task                                           | File                       | Effort |
| --- | ---------------------------------------------- | -------------------------- | ------ |
| 5   | Consider `nix shell` for better error handling | `githistory_filterrepo.go` | 30min  |
| 6   | Add retry logic for transient nix failures     | `githistory_filterrepo.go` | 20min  |
| 7   | Increase detection timeout (currently 5s)      | `githistory_filterrepo.go` | 5min   |

---

## Investigation Needed

### Questions to Answer

1. **WHERE is `--protect-blobs-from HEAD` added?**
   - Search: `githistory_executor.go`, `githistory.go`
   - Grep for: `protect-blobs-from`

2. **Why does detection fail intermittently?**
   - Network timeout during `nix eval`?
   - Nix store not ready?
   - First-time download blocking?

---

## Test Cases Needed

```bash
# Test 1: Verify nix detection
clean-wizard git-history --verbose

# Test 2: Verify execution without invalid flag
clean-wizard git-history --dry-run

# Test 3: Verify caching doesn't persist failures
# (Run twice, second should succeed if first failed due to transient issue)
```

---

## Next Actions

1. **Search for `--protect-blobs-from`** in codebase
2. **Remove the invalid flag** or replace with correct git-filter-repo syntax
3. **Fix caching logic** to retry on failure
4. **Add verbose logging** for detection process
5. **Test end-to-end** with nix

---

## Related Files

- `internal/cleaner/githistory_filterrepo.go` - Detection and command building
- `internal/cleaner/githistory_safety.go` - Safety checks
- `internal/cleaner/githistory_executor.go` - Execution (likely where flag is added)
- `internal/cleaner/githistory.go` - Main cleaner logic

---

## References

- git-filter-repo docs: https://github.com/newren/git-filter-repo
- Nix command reference: https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-run.html
