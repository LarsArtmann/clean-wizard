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

1. `IsNotAvailableError` Keyword Fallback Still Exists — **partial** — keyword fallback still in `cleaner.go`; retained as safety net for OS errors (`exec.LookPath`, etc.). Acceptable.
2. `RetryConfig` Hardcoded in Command Layer — **resolved at `edaff33`** — `RetryConfigFromAttempts(n)` shared builder extracted; commands now use it
3. Stale Status Reports Not Cleaned Up — **resolved at 2026-07-13 + 2026-08-10 audits**

---

## c) NOT STARTED

1. Consolidate 4 error packages — **done at `edaff33`** — `go-error-family` fully adopted; `internal/pkg/errors/` deleted (1283 lines)
2. ~~Adopt `go-error-family`~~ done at `edaff33`
3. Add `RetryProfile` type — **done at `132f5f6`**
4. Wire `OperationSettings` from YAML config — still open; tracked as TODO #6
5. Register individual cleaners as DI providers — still open; tracked as TODO #29
6. `flow.If` for Docker daemon check — NOT-DO; handled by `IsAvailable`
7. Migrate `githistory` command to DI — NOT-DO; explicitly deferred
8. Migrate `init`, `profile`, `config` commands to DI — NOT-DO; explicitly deferred
9. Make adapters interface-backed — still open; tracked as TODO #30
10. Implement `do.ShutdownerWithError` — NOT-DO
11. Consolidate `cleaner.Cleaner` vs `domain.OperationHandler` — NOT-DO (ROADMAP non-goal #4)
12. Add BDD tests for execution layer (Ginkgo) — still open; tracked as TODO #7
13. Logger globals → DI-injected logger — still open; tracked as TODO #11
14. Re-enable `t.Parallel()` on logger tests — still open; blocked by #13
15. Add `--keep-generations` flag — still open; tracked as TODO #23
16. Add progress TUI — aspirational (ROADMAP Theme #3)
17. Per-cleaner timeout via `flow.Timeout` — NOT-DO
18. Profile-based filtering for `scan --profile` — still open; tracked as TODO #10
19. ~~Clean up stale status reports~~ done at 2026-07-13 + 2026-08-10 audits
20. Split `internal/domain/` god package — still open; tracked as TODO #27
21. Split `internal/cleaner/` flat structure — still open; tracked as TODO #28

## d) TOTALLY FUCKED UP

1. `--retries 3` Default May Surprise Users — **partially mitigated at `1b96d06`** — `--retries 0` escape hatch documented; smart retry via `errorfamily.IsRetryable()` reduces wasted budget
2. RetryConfig Values Duplicated Across Commands — **resolved at `edaff33`** — `RetryConfigFromAttempts(n)` shared builder
3. Scan `--profile` Warning Is a Band-Aid — **partially resolved** at `1b96d06`; full filtering still tracked as TODO #10
4. `paralleltest` Lint Warnings on Integration Tests — NOT-DO; tests legitimately cannot run in parallel
5. No Test for Smart Retry (NotAvailableError → backoff.Stop) — **resolved at `edaff33`**

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (Low Effort, High Confidence)

1. ~~Add test for smart retry~~ done at `edaff33`
2. ~~Extract shared retry config builder~~ done at `edaff33`
3. ~~Use `DefaultRetryConfig()` in commands instead of inline hardcoded values~~ done at `edaff33`
4. Tighten `IsNotAvailableError` — NOT-DO; safety net for OS errors is correct

### Near-Term (Medium Effort, Real Value)

5. Wire `OperationSettings` from config profiles — still open; tracked as TODO #6
6. Implement `scan --profile` filtering — still open; tracked as TODO #10
7. ~~Adopt `go-error-family`~~ done at `edaff33`
8. ~~Add `RetryProfile` type~~ done at `132f5f6`
9. Logger globals → DI — still open; tracked as TODO #11
10. ~~Consolidate error packages~~ done at `edaff33`

### Strategic (Higher Effort, Architectural)

11. Register individual cleaners as DI providers — still open; tracked as TODO #29
12. Make adapters interface-backed — still open; tracked as TODO #30
13. Consolidate `cleaner.Cleaner` vs `domain.OperationHandler` — NOT-DO (ROADMAP non-goal #4)
14. Add BDD tests — still open; tracked as TODO #7
15. Split `internal/domain/` god package — still open; tracked as TODO #27

---

## f) Top 25 Things to Do Next

| #   | Task                                                                 | Resolution |
| --- | -------------------------------------------------------------------- | ---------- |
| 1   | ~~**Add test: smart retry stops on NotAvailableError**~~             | done at `edaff33` |
| 2   | ~~**Extract shared retry config builder**~~                          | done at `edaff33` |
| 3   | ~~**Use `DefaultRetryConfig()`** in commands~~                       | done at `edaff33` |
| 4   | Log warning when keyword fallback in `IsNotAvailableError` fires     | NOT-DO |
| 5   | ~~Clean up stale status reports~~                                   | done at 2026-07-13 + 2026-08-10 audits |
| 6   | Add `--timeout` per-cleaner flag                                     | NOT-DO |
| 7   | Wire `OperationSettings` from config to cleaner constructors         | still open — TODO #6 |
| 8   | Implement `scan --profile` filtering or remove the flag              | still open — TODO #10 |
| 9   | ~~Adopt `go-error-family`~~                                         | done at `edaff33` |
| 10  | ~~Add `RetryProfile` type~~                                          | done at `132f5f6` |
| 11  | Logger globals → DI-injected logger                                  | still open — TODO #11 |
| 12  | Re-enable `t.Parallel()` on logger tests                             | blocked by #11 |
| 13  | ~~Consolidate 4 error packages~~                                    | done at `edaff33` |
| 14  | Register individual cleaners as separate DI providers                | still open — TODO #29 |
| 15  | Make adapters interface-backed with `do.As`                          | still open — TODO #30 |
| 16  | Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`           | NOT-DO |
| 17  | Add BDD tests for execution layer (Ginkgo)                           | still open — TODO #7 |
| 18  | Migrate `githistory` command to DI                                   | NOT-DO |
| 19  | Add `flow.If` for Docker daemon check                                | NOT-DO |
| 20  | Implement `do.ShutdownerWithError` on resource holders               | NOT-DO |
| 21  | Split `internal/domain/` god package (23 files)                      | still open — TODO #27 |
| 22  | Split `internal/cleaner/` flat structure (50+ files)                 | still open — TODO #28 |
| 23  | Add progress TUI                                                     | aspirational |
| 24  | Add `--keep-generations` flag for Nix cleaner                        | still open — TODO #23 |
| 25  | Add `--dry-run` to scan command (parity with clean)                  | still open — TODO #22 |

## g) Top #1 Question I Cannot Answer Myself

> **Resolved at `edaff33`:** Yes — `go-error-family` was adopted. `bridge` subpackage deliberately **not adopted** (clean-wizard doesn't use `samber/oops`). See the 2026-07-06 go-error-family adoption report for full rationale.

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
