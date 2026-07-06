# Status Report: Brutal Self-Review + Hardening Execution

**Date:** 2026-07-06 03:42
**Session Scope:** Brutal self-review of DI+workflow migration, then execution of all Tier 1-3 improvement tasks
**Commits:** `6a539e7` → `c2ce0dc` (4 commits, all pushed to `master`)
**Test Status:** ALL 300+ tests pass in short mode; BuildFlow 27/27 green on every commit
**Build Status:** `go build ./...` clean

---

## Executive Summary

This session started with a brutal self-review of the previous migration sessions (commits `43df609` through `de105b0`). The review identified 11 concrete improvement tasks across correctness, type safety, UX, and documentation. All 11 were executed, verified, committed, and pushed.

**Key wins:** Short-mode test time for the commands package dropped from **34.9s → 0.007s** (integration tests now properly skip). The retry fix (`recordFinal`) now has a regression test asserting exactly 1 step entry after 3 attempts. All 13 cleaners now return `*NotAvailableError` for unavailable conditions — the keyword fallback is truly just a safety net.

---

## a) FULLY DONE ✓

### Commit 1: `6a539e7` — Test Hardening + Dead Code Removal

| #   | Task                               | Detail                                                                                                                                                                                                                                  |
| --- | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Retry test assertions**          | `TestRunCleaners_Retry` now asserts `len(Steps) == 1`, `StepStatusSucceeded`, `FreedBytes == 42`, and `attempts == 3`. This verifies the `recordFinal()` fix prevents duplicate entries — a regression would fail immediately.          |
| 2   | **Dead `record()` method removed** | Old `resultCollector.record()` method deleted from `results.go`. Had zero callers after the `recordFinal()` migration.                                                                                                                  |
| 3   | **Integration test skip guards**   | 3 tests now skip under `testing.Short()`: `TestRunCleaners_RealRegistry_DryRun`, `TestRunScans_RealRegistry_DryRun`, `TestRunCleanCommand_DryRun_JSON`. These invoke real system cleaners and were adding 35s+ to every short test run. |

### Commit 2: `c102e0f` — NotAvailableError Migration Complete

| #   | Cleaner                        | Before                                                                              | After                                                                               |
| --- | ------------------------------ | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| 4   | `projectsmanagementautomation` | `errors.New("projects-management-automation not available")`                        | `&NotAvailableError{CleanerName: "projects-management-automation"}`                 |
| 5   | `systemcache`                  | `errors.New("not available on this platform (requires macOS or Linux)")`            | `&NotAvailableError{CleanerName: "systemcache", Reason: "requires macOS or Linux"}` |
| 6   | `golang_cleaner` (sentinel)    | `ErrGoCacheNotAvailable = &NotAvailableError{CleanerName: "go"}` (pointer sentinel) | Removed sentinel entirely; returns `&NotAvailableError{CleanerName: "go"}` inline   |

**Why the sentinel was removed:** `ErrGoCacheNotAvailable` was never checked via `errors.Is()` — classification always went through `IsNotAvailableError()`'s `errors.As` path. The sentinel abstraction added zero value while creating a latent footgun: if any code created its own `&NotAvailableError{CleanerName: "go"}`, `errors.Is` comparison would fail due to pointer-identity mismatch.

### Commit 3: `1b96d06` — Scan Command Parity + Retry Default

| #   | Feature                          | Detail                                                                                                                                                                                                                                                                               |
| --- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 7   | **`--retries` on scan**          | Scan command now accepts `--retries N` flag, matching clean. When N > 0, passes `RetryConfig` to `execution.RunScans`.                                                                                                                                                               |
| 8   | **`--concurrency`/`-C` on scan** | Scan command now accepts `--concurrency N` flag, matching clean. When N > 0, passes to `execution.WithMaxConcurrency()` and sets `RunSettings.MaxConcurrency`.                                                                                                                       |
| 9   | **`--retries` default = 3**      | Both clean and scan commands now default to 3 retries. Production runs recover from transient failures (Nix store locks, Docker daemon hiccups) by default. `IsNotAvailableError` smart retry ensures non-retryable errors stop immediately with zero delay. `--retries 0` disables. |
| 10  | **`--profile` warning in scan**  | Previously, `scan --profile daily` silently discarded the value (parameter was `_ string`). Now prints: `⚠️ Warning: --profile "daily" is not yet supported for scan; showing all available cleaners`                                                                                |

### Commit 4: `c2ce0dc` — Documentation Update

| #   | Change                                                                                                                                                                                                                                                                                          |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 11  | Updated `AGENTS.md` with: retry-by-default behavior, complete `*NotAvailableError` migration status, specific error package locations in Known Issues, logger globals as a known issue, updated test facts (CLI integration test exists, `testing.Short()` guards), removed stale `err113` note |

### Verification Results

```
go build ./...              → PASS (0 errors)
go test ./... -short        → PASS (all 22 packages, 300+ tests)
BuildFlow pre-commit        → 27/27 green (every commit)
```

---

## b) PARTIALLY DONE

### 1. `IsNotAvailableError` Keyword Fallback Still Exists

The keyword-based fallback in `IsNotAvailableError()` (checking for "not available", "not installed" substrings) still exists in `cleaner/cleaner.go`. All 13 internal cleaners now return typed `*NotAvailableError`, so the fallback is technically dead code for internal use. However, OS-level errors from `exec.LookPath`, `os.Stat`, etc. may still contain "not available" strings, so the fallback serves as a safety net for those cases.

**Status:** Functional but carrying dead weight for internal cleaners. Could be tightened to only check `errors.As` once we're confident no OS errors need keyword matching.

### 2. `RetryConfig` Hardcoded in Command Layer

Both `clean.go` and `scan.go` construct `RetryConfig{MaxAttempts: 3, InitialBackoff: 2s, MaxBackoff: 30s}` inline from the `--retries` flag value. The `DefaultRetryConfig()` function exists in `execution/retry.go` but is not used. The backoff/MaxBackoff values are duplicated across both commands.

**Status:** Works correctly but is a minor DRY violation. A shared helper or config-driven approach would be cleaner.

### 3. Stale Status Reports Not Cleaned Up

`docs/status/2026-07-06_00-35_DI-WORKFLOW-MIGRATION.md` still references deleted `FlowBuilder`/`BranchFlow` as "dormant" — they've been deleted. Multiple status reports from the migration session are now partially stale.

**Status:** Pre-existing issue. Not blocking but creates confusion for anyone reading the history.

---

## c) NOT STARTED

1. **Consolidate 4 error packages** — `internal/pkg/errors/` (ghost, used only by config), `cleaner.NotAvailableError`, `domain.ValidationError`, scattered sentinel `var Err...` across commands
2. **Adopt `go-error-family`** — BuildFlow uses `errorfamily.Transient` for retry classification; cleaner than hand-rolled `IsNotAvailableError`
3. **Add `RetryProfile` type** — Default/Aggressive/Conservative/None profiles matching BuildFlow's pattern
4. **Wire `OperationSettings` from YAML config** — user profile settings don't reach cleaner constructors; hardcoded defaults used instead
5. **Register individual cleaners as DI providers** — large refactor; enables per-cleaner config from YAML
6. **`flow.If` for Docker daemon check** — BuildFlow conditional pattern
7. **Migrate `githistory` command to DI** — still uses direct constructor calls
8. **Migrate `init`, `profile`, `config` commands to DI** — not using container
9. **Make adapters interface-backed** with `do.As` aliasing
10. **Implement `do.ShutdownerWithError`** on resource-holding adapters
11. **Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`** — impedance mismatch
12. **Add BDD tests for execution layer** (Ginkgo)
13. **Logger globals → DI-injected logger** — root cause of test races (we patched symptoms by removing `t.Parallel()`)
14. **Re-enable `t.Parallel()` on logger tests** — after logger globals fix
15. **Add `--keep-generations` flag** for Nix cleaner
16. **Add progress TUI** — live per-cleaner status during workflow execution
17. **Per-cleaner timeout** via `flow.Timeout`
18. **Profile-based filtering for `scan --profile`** — flag now warns but doesn't filter
19. **Clean up stale status reports** — references to deleted types
20. **Split `internal/domain/` god package** — 23 files in one package
21. **Split `internal/cleaner/` flat structure** — 50+ files, no sub-packages

---

## d) TOTALLY FUCKED UP / RISKS

### 1. `--retries 3` Default May Surprise Users

Changing the default from 0 (disabled) to 3 (enabled) means every `clean-wizard clean` invocation now retries transient failures with 2s→4s→8s backoff. A hung Docker daemon or unresponsive Nix store will cause up to ~14s of delay per cleaner before failing. The `IsNotAvailableError` smart retry prevents delay for non-retryable errors, but genuinely transient-but-stuck errors (daemon starting up) will still cause delays.

**Risk:** Users running `clean` interactively may experience unexpected pauses. The `--retries 0` escape hatch exists but isn't obvious.

### 2. RetryConfig Values Duplicated Across Commands

`clean.go` and `scan.go` both hardcode `InitialBackoff: 2 * time.Second, MaxBackoff: 30 * time.Second`. If someone changes one, they might forget the other. `DefaultRetryConfig()` exists but isn't used.

### 3. Scan `--profile` Warning Is a Band-Aid

The warning tells users the flag isn't supported, but `scan --profile daily` still runs ALL cleaners. A user who doesn't read warnings (most users) will assume filtering happened. This is better than silent ignoring, but still a UX lie.

### 4. `paralleltest` Lint Warnings on Integration Tests

The linter flags all integration tests in `execution/integration_test.go` for missing `t.Parallel()`. These tests can't run in parallel (they share the workflow engine and may interfere). The warnings are noise but indicate the tests aren't parallelized.

### 5. No Test for Smart Retry (NotAvailableError → backoff.Stop)

The `NextBackOff` hook in `retry.go` calls `cleaner.IsNotAvailableError(re.Error)` and returns `backoff.Stop`. This was verified manually but has **no automated test**. A regression could cause the retry engine to waste time retrying "cargo not installed" errors.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (Low Effort, High Confidence)

1. **Add test for smart retry** — verify `NotAvailableError` + retries enabled → only 1 attempt (no backoff)
2. **Extract shared retry config builder** — DRY the `RetryConfig` construction between clean and scan
3. **Use `DefaultRetryConfig()`** in commands instead of inline hardcoded values
4. **Tighten `IsNotAvailableError`** — log a warning when keyword fallback fires, so we know if it's still needed

### Near-Term (Medium Effort, Real Value)

5. **Wire `OperationSettings` from config profiles** through to cleaner constructors
6. **Implement `scan --profile` filtering** — or remove the flag entirely
7. **Adopt `go-error-family`** for retry classification — replaces hand-rolled keyword matching
8. **Add `RetryProfile` type** — flexible retry presets
9. **Logger globals → DI** — fix root cause of test races
10. **Consolidate error packages** into one coherent design

### Strategic (Higher Effort, Architectural)

11. **Register individual cleaners as DI providers** — enable per-cleaner config
12. **Make adapters interface-backed** with `do.As` aliasing
13. **Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`**
14. **Add BDD tests** for execution layer (Ginkgo)
15. **Split `internal/domain/` god package** into sub-packages

---

## f) Top 25 Things to Do Next

| #   | Task                                                                 | Impact   | Effort |
| --- | -------------------------------------------------------------------- | -------- | ------ |
| 1   | **Add test: smart retry stops on NotAvailableError**                 | CRITICAL | S      |
| 2   | **Extract shared retry config builder** (DRY clean/scan)             | HIGH     | S      |
| 3   | **Use `DefaultRetryConfig()`** in commands instead of inline         | HIGH     | S      |
| 4   | **Log warning when keyword fallback in `IsNotAvailableError` fires** | MEDIUM   | S      |
| 5   | **Clean up stale status reports** (references to deleted types)      | LOW      | S      |
| 6   | **Add `--timeout` per-cleaner flag** wired to `flow.Timeout`         | MEDIUM   | M      |
| 7   | **Wire `OperationSettings` from config to cleaner constructors**     | HIGH     | L      |
| 8   | **Implement `scan --profile` filtering** or remove the flag          | MEDIUM   | M      |
| 9   | **Adopt `go-error-family`** for retry classification                 | HIGH     | M      |
| 10  | **Add `RetryProfile` type** (Default/Aggressive/Conservative/None)   | MEDIUM   | M      |
| 11  | **Logger globals → DI-injected logger**                              | MEDIUM   | M      |
| 12  | **Re-enable `t.Parallel()` on logger tests**                         | LOW      | S      |
| 13  | **Consolidate 4 error packages** into one coherent design            | HIGH     | L      |
| 14  | **Register individual cleaners as separate DI providers**            | HIGH     | L      |
| 15  | **Make adapters interface-backed with `do.As`**                      | MEDIUM   | L      |
| 16  | **Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`**       | MEDIUM   | L      |
| 17  | **Add BDD tests for execution layer** (Ginkgo)                       | MEDIUM   | M      |
| 18  | **Migrate `githistory` command to DI**                               | LOW      | S      |
| 19  | **Add `flow.If` for Docker daemon check**                            | MEDIUM   | S      |
| 20  | **Implement `do.ShutdownerWithError`** on resource holders           | LOW      | S      |
| 21  | **Split `internal/domain/` god package** (23 files)                  | MEDIUM   | L      |
| 22  | **Split `internal/cleaner/` flat structure** (50+ files)             | MEDIUM   | L      |
| 23  | **Add progress TUI** (live per-cleaner status)                       | LOW      | L      |
| 24  | **Add `--keep-generations` flag** for Nix cleaner                    | LOW      | S      |
| 25  | **Add `--dry-run` to scan command** (parity with clean)              | LOW      | S      |

---

## g) Top #1 Question I Cannot Answer Myself

**Should we adopt `go-error-family` (`github.com/larsartmann/go-error-family`) to replace the hand-rolled `IsNotAvailableError` + keyword fallback?**

Arguments for:

- BuildFlow uses it successfully for `errorfamily.Transient` classification
- Eliminates the fragile keyword-matching fallback entirely
- Provides a proven, typed error family system instead of a single ad-hoc type
- Would allow retry classification to be more granular (Transient vs Permanent vs NotAvailable)

Arguments against:

- It's a LarsArtmann library — I can't verify its maturity, API stability, or community adoption from inside this repo
- The current `*NotAvailableError` + `errors.As` path works correctly for all 13 cleaners now
- Adding a dependency for error classification may be over-engineering if the keyword fallback is rarely hit
- The migration would touch every cleaner's error returns again

I don't know if `go-error-family` is production-ready enough to bet on, or if the current typed error approach is "good enough" to keep. This is an architectural decision that affects every error path in the codebase.
