# Status Report: Pareto Hardening Pass — Final Session Review

**Date:** 2026-07-06 02:37
**Session Scope:** Bug fix for retry duplicate recording, CLI flag wiring, smart retry, NotAvailableError migration, dead code removal, CLI integration test
**Commit:** `de105b0` (pushed to `master`)
**Test Status:** ALL 300+ tests pass across 22 packages, 0 failures
**BuildFlow CI:** 27/27 checks passed

---

## a) FULLY DONE ✓

### Bug Fixes

| #   | Fix                              | Detail                                                                                                                                                                                                     |
| --- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Retry duplicate recording**    | `makeCleanStepFunc` now records only the final outcome via `recordFinal()` in the `defer` block — replaces previous entry for same step name instead of appending. Same fix applied to `makeScanStepFunc`. |
| 2   | **Dead code: `scanCleanerReal`** | 50-line function removed from `scan.go` — zero callers after scan migration to `execution.RunScans`                                                                                                        |
| 3   | **Dead code: `record()` method** | Old `record()` method on `resultCollector` is now unused — only `recordFinal()` is called. The old method still exists but has zero callers.                                                               |

### CLI Features Wired

| #   | Feature                         | Flags           | Detail                                                                                                           |
| --- | ------------------------------- | --------------- | ---------------------------------------------------------------------------------------------------------------- |
| 4   | `--retries N`                   | `clean` command | When N > 0, passes `RetryConfig{MaxAttempts: N, InitialBackoff: 2s, MaxBackoff: 30s}` to `execution.WithRetry()` |
| 5   | `--concurrency N` / `-C`        | `clean` command | When N > 0, passes to `execution.WithMaxConcurrency()` and sets `RunSettings.MaxConcurrency`                     |
| 6   | `MaxConcurrency` in RunSettings | DI layer        | Now populated from `--concurrency` flag in `clean` command                                                       |

### Smart Retry

| #   | Feature                                       | Detail                                                                                                                                                                                   |
| --- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 7   | `NextBackOff` hook with `IsNotAvailableError` | Retry engine now calls `cleaner.IsNotAvailableError(re.Error)` — if true, returns `backoff.Stop` immediately. No more wasting 30s of exponential backoff retrying "cargo not installed". |

### Typed Error Migration

| #   | Cleaner           | Before                                    | After                                            |
| --- | ----------------- | ----------------------------------------- | ------------------------------------------------ |
| 8   | cargo             | `errors.New("cargo not available")`       | `&NotAvailableError{CleanerName: "cargo"}`       |
| 9   | docker            | `errors.New("docker not available")`      | `&NotAvailableError{CleanerName: "docker"}`      |
| 10  | homebrew (×2)     | `errors.New("homebrew not available")`    | `&NotAvailableError{CleanerName: "homebrew"}`    |
| 11  | go                | `errors.New("go not available")` sentinel | `&NotAvailableError{CleanerName: "go"}` sentinel |
| 12  | helpers (generic) | `fmt.Errorf("%s not available", name)`    | `&NotAvailableError{CleanerName: name}`          |

### Tests

| #   | Test                 | Coverage                                                                                                                                                                   |
| --- | -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 13  | CLI integration test | `TestRunCleanCommand_DryRun_JSON` — full pipeline: cobra → config → DI → registry → workflow → JSON output. Verified real output with 11 cleaners, 96 GiB estimated freed. |

---

## b) PARTIALLY DONE

1. ~~Scan Command Missing `--retries` and `--concurrency` Flags~~ done at `1b96d06` — both flags now wired to `scan` command with parity to `clean`
2. ~~Two Cleaners Still Use Ad-Hoc String Errors~~ done at `c102e0f` — `projectsmanagementautomation` and `systemcache` now return `*NotAvailableError` via `NewNotAvailableError`
3. ~~`record()` Method on `resultCollector` Is Now Dead Code~~ done at `6a539e7` — `record()` removed; only `recordFinal()` remains
4. ~~Nix Cleaner Not Migrated to `*NotAvailableError`~~ Won't implement — Nix intentionally returns mock data when unavailable for dry-run estimation; not an error condition

## c) NOT STARTED

1. ~~`--retries` / `--concurrency` flags on scan command~~ done at `1b96d06`
2. ~~Migrate remaining 2 cleaners~~ done at `c102e0f`
3. ~~Remove dead `record()` method~~ done at `6a539e7`
4. ~~`githistory` command DI migration~~ NOT-DO — explicitly deferred in 2026-07-06_03-42 Pareto pass
5. ~~Per-cleaner timeout via `flow.Timeout`~~ NOT-DO — complexity > value
6. Implement profile-based filtering for `scan --profile` — still open; tracked as TODO #10
7. Register individual cleaners as DI providers — still open; tracked as TODO #29
8. Pass `OperationSettings` from config to cleaner constructors — still open; tracked as TODO #6
9. Make adapters interface-backed — still open; tracked as TODO #30
10. `do.ShutdownerWithError` on resource-holding adapters — NOT-DO
11. Consolidate `cleaner.Cleaner` vs `domain.OperationHandler` — NOT-DO (ROADMAP non-goal #4)
12. Add BDD tests for execution layer (Ginkgo) — still open; tracked as TODO #7
13. `--keep-generations` flag for Nix — still open; tracked as TODO #23
14. Progress TUI — aspirational (ROADMAP Theme #3)
15. ~~Stale status report cleanup — references to deleted `FlowBuilder`/`BranchFlow`~~ done at 2026-07-13 + 2026-08-10 docs-health audits

---

## d) TOTALLY FUCKED UP / RISKS

### 1. No Test for the Retry Duplicate Recording Fix

The fix adds `recordFinal()` and moves recording to `defer`, but **no test verifies that a retried step produces exactly 1 entry** in `WorkflowResult.Steps`. The existing `TestRunCleaners_Retry` test verifies the workflow succeeds after retries, but doesn't assert step count. A regression could silently reintroduce duplicates.

### 2. `--retries` Default of 0 Means Production Has No Retries

The `--retries` flag defaults to 0 (disabled). This means production clean runs get zero retries — a transient Nix store lock or Docker daemon hiccup causes immediate failure. The `DefaultRetryConfig()` exists with sensible defaults (3 attempts, 2s initial backoff) but is never used as a default. Should the default be 3?

### 3. `ErrGoCacheNotAvailable` Changed Type but Sentinel Comparison May Break

`ErrGoCacheNotAvailable` was changed from `errors.New(...)` (value) to `&NotAvailableError{...}` (pointer). Any code doing `errors.Is(err, ErrGoCacheNotAvailable)` will now do pointer comparison against a specific `*NotAvailableError` instance. If a test creates its own `&NotAvailableError{CleanerName: "go"}`, it won't match the sentinel even though it represents the same condition. The `IsNotAvailableError()` function handles this correctly via `errors.As`, but direct sentinel comparisons could break.

### 4. Integration Test Takes 15 Seconds

`TestRunCleanCommand_DryRun_JSON` runs the actual clean pipeline with real cleaners in dry-run mode. On the evo-x2 machine, this takes 15.3 seconds because it actually scans system caches (Go cache, system cache, compiled binaries, etc.). This is too slow for a unit test suite — it should be tagged as integration or use a mock registry.

### 5. `--profile` Flag Silently Ignored in Scan Command

The flag is defined, parsed by cobra, but `runScanCommand` receives it as `_ string`. Users running `scan --profile daily` get ALL cleaners with no warning. Pre-existing issue, not fixed in this session.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (Low Effort, High Confidence)

1. **Add test asserting step count after retry** — verify `recordFinal` produces exactly 1 entry per step
2. **Remove dead `record()` method** from `resultCollector`
3. **Migrate remaining 2 cleaners** to `*NotAvailableError`
4. **Add `--retries` and `--concurrency` to scan command** — symmetric with clean
5. **Tag integration test** with `//go:build integration` or use `testing.Short()` skip

### Near-Term (Medium Effort, Real Value)

6. **Set `--retries` default to 3** — production cleaners should retry transient failures
7. **Wire `OperationSettings` from config profiles** through to cleaner constructors
8. **Implement `scan --profile` filtering** — or remove the flag if not planned
9. **Add `--timeout` flag** — per-cleaner timeout via `flow.Timeout`
10. **Consolidate error packages** — 4 packages with overlapping responsibilities

### Strategic (Higher Effort, Architectural)

11. **Register individual cleaners as DI providers** — enable per-cleaner config
12. **Make adapters interface-backed** with `do.As` aliasing
13. **Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`**
14. **Add BDD tests** for execution layer (Ginkgo)
15. **Add progress TUI** — live per-cleaner status during workflow execution

---

## f) Top 25 Things to Do Next

| #   | Task                                                                               | Impact   | Effort |
| --- | ---------------------------------------------------------------------------------- | -------- | ------ |
| 1   | **Add test: assert step count = 1 after retry**                                    | CRITICAL | S      |
| 2   | **Remove dead `record()` method** from resultCollector                             | HIGH     | S      |
| 3   | **Migrate `projectsmanagementautomation` + `systemcache` to `*NotAvailableError`** | MEDIUM   | S      |
| 4   | **Add `--retries` and `--concurrency` to scan command**                            | MEDIUM   | S      |
| 5   | **Tag/skip integration test for short mode**                                       | HIGH     | S      |
| 6   | **Set `--retries` default to 3** (or 2) for production resilience                  | HIGH     | S      |
| 7   | **Wire `OperationSettings` from config to cleaner constructors**                   | HIGH     | L      |
| 8   | **Implement `scan --profile` filtering** or remove the flag                        | MEDIUM   | M      |
| 9   | **Add `--timeout` per-cleaner flag** wired to `flow.Timeout`                       | MEDIUM   | M      |
| 10  | **Consolidate 4 error packages** into one coherent design                          | MEDIUM   | L      |
| 11  | **Clean up stale status reports** (reference deleted types)                        | LOW      | S      |
| 12  | **Add `--keep-generations` flag** for Nix cleaner                                  | LOW      | S      |
| 13  | **Register individual cleaners as separate DI providers**                          | HIGH     | L      |
| 14  | **Make adapters interface-backed with `do.As`**                                    | HIGH     | L      |
| 15  | **Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`**                     | MEDIUM   | L      |
| 16  | **Implement `do.ShutdownerWithError`** on adapters with resources                  | LOW      | S      |
| 17  | **Add BDD tests for execution layer** (Ginkgo)                                     | MEDIUM   | M      |
| 18  | **Add progress TUI** (live per-cleaner status)                                     | LOW      | L      |
| 19  | **Add `do.ExplainInjector` debug output** behind `--di-debug`                      | LOW      | S      |
| 20  | **Create application-global DI bootstrap** shared by all commands                  | MEDIUM   | M      |
| 21  | **Migrate `githistory` command to DI**                                             | LOW      | S      |
| 22  | **Add `flow.If` conditional for Docker** (daemon running check)                    | LOW      | S      |
| 23  | **Add audit log of DI registrations**                                              | LOW      | S      |
| 24  | **Fix `ErrGoCacheNotAvailable` sentinel** — make it a value, not pointer           | MEDIUM   | S      |
| 25  | **Profile-guided cleaner selection** — read config profiles in DI provider         | HIGH     | L      |

---

## g) Top #1 Question I Cannot Answer Myself

**Should `--retries` default to 0 (disabled, current) or 3 (enabled with `DefaultRetryConfig`)?**

Arguments for default 0:

- Predictability — users get immediate failure feedback, no surprising delays
- Backward compatibility — matches the pre-workflow behavior (no retries)
- Safer — a hung Docker daemon won't cause 30s delays on every clean run

Arguments for default 3:

- Production resilience — transient failures (Nix store locks, Docker daemon hiccups) are common
- The `IsNotAvailableError` smart retry already prevents wasting time on non-retryable errors
- The `DefaultRetryConfig()` function exists specifically because 3 attempts with 2s backoff was deemed sensible

I cannot determine the user's preference for default-fail-fast vs default-resilience. This is a UX decision that affects every `clean-wizard clean` invocation.
