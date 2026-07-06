# Status Report: DI + Workflow Migration — Full Session Review

**Date:** 2026-07-06 01:38
**Session Scope:** Two commits — initial DI+workflow migration, then hardening pass with bug fixes, dead code removal, retry support, and comprehensive testing
**Commits:** `43df609` → `65290ce` (pushed to `master`)
**Test Status:** ALL 300+ tests pass across 22 packages, 0 failures

---

## a) FULLY DONE ✓

### Commit 1: `43df609` — Initial Migration

| Item                          | Detail                                                                                      |
| ----------------------------- | ------------------------------------------------------------------------------------------- |
| Dependencies                  | `samber/do v2 v2.0.0`, `Azure/go-workflow v0.1.13` added to `go.mod`                        |
| `internal/di/` package        | 6 files, 294 lines — container wrapper, providers, accessors, test helpers                  |
| `internal/execution/` package | 7 files, 638 lines — builder, hooks, options, results, workflow entry points                |
| Command rewiring              | `clean.go`, `clean_execute.go`, `scan.go`, `cleaner_config.go` all use DI + execution layer |
| Deleted                       | `cleaner_implementations.go` (357 lines of dual-registry dispatch)                          |
| Cleaner improvement           | Go process safety check moved into `GoCleaner.Clean()`                                      |
| Tests                         | 19 new tests (8 DI + 11 execution)                                                          |

### Commit 2: `65290ce` — Hardening Pass

#### Critical Bug Fixes (4)

| #   | Bug                                                                      | Fix                                                            | File                        |
| --- | ------------------------------------------------------------------------ | -------------------------------------------------------------- | --------------------------- |
| 1   | Workflow errors silently dropped when steps existed                      | Error now preserved; returned only when zero steps collected   | `execution/workflow.go`     |
| 2   | Panics in cleaners disappeared without trace                             | `recover()` in step functions records panicked steps as failed | `execution/builder.go`      |
| 3   | `isProcessRunning` failed open (returned false) when `pgrep` unavailable | Now fails closed — checks `exec.LookPath("pgrep")` first       | `cleaner/golang_cleaner.go` |
| 4   | Non-deterministic result ordering from parallel execution                | Results sorted by registration order via `orderIndex` map      | `execution/results.go`      |

#### Quality Improvements (4)

| #   | Change                                                                                                                                          |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| 5   | `resultCollector` mutex changed from `*sync.Mutex` (nil-pointer risk) to value `sync.Mutex` with constructor                                    |
| 6   | `TotalItemsFailed` aggregates from ALL steps, not just successful ones                                                                          |
| 7   | Typed `cleaner.NotAvailableError` + `cleaner.IsNotAvailableError()` replaces fragile string matching — uses `errors.As` first, keyword fallback |
| 8   | JSON error serialization verified correct — `format.CleanResultsToJSON` already calls `.Error()` on errors                                      |

#### Dead Code Removed — 1472 lines net

| File                                    | Lines | Why Dead                                                               |
| --------------------------------------- | ----- | ---------------------------------------------------------------------- |
| `result/flow_builder.go`                | 309   | Superseded by go-workflow — `FlowBuilder`, `Pipeline`, `ParallelFlow`  |
| `result/branch_flow.go`                 | 200   | Superseded by go-workflow — `BranchFlow`, `SwitchFlow`                 |
| `result/branch_flow_test.go`            | 697   | Tests for removed types                                                |
| `cleaner/parallel.go`                   | 166   | `ParallelExecutor`, `CleanAllParallel` — superseded by execution layer |
| `cleaner/metrics.go` (partial)          | 24    | `CleanAllParallelWithMetrics` — referenced deleted `parallel.go` types |
| `cleaner/registry_factory.go` (partial) | 10    | `DefaultRegistry()` — zero callers after DI migration                  |

#### New Features (5)

| #   | Feature                                                                     | Files                                                                |
| --- | --------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| 9   | `--config` flag for `scan` command (matching `clean`)                       | `scan.go`                                                            |
| 10  | Scan command wired to `execution.RunScans` for parallel execution           | `scan.go`, `execution/workflow.go`                                   |
| 11  | `RetryConfig` + `flow.Retry` with `cenkalti/backoff/v4` exponential backoff | `execution/retry.go`, `execution/options.go`, `execution/builder.go` |
| 12  | `cenkalti/backoff/v4` promoted from indirect to direct dependency           | `go.mod`                                                             |
| 13  | `MaxConcurrency` field in `RunSettings`                                     | `di/options.go`                                                      |

#### Test Coverage — 23 new test functions

| File                            | Tests | Coverage                                                                                             |
| ------------------------------- | ----- | ---------------------------------------------------------------------------------------------------- |
| `di/di_test.go`                 | 9     | Container creation, service registration, config/settings/registry resolution, override, error cases |
| `execution/execution_test.go`   | 9     | Success/failure/skip/empty/unknown cleaner, ordering, classification, options                        |
| `execution/integration_test.go` | 5     | Real registry dry-run (clean+scan), panic recovery, deterministic ordering, retry behavior           |

---

## b) PARTIALLY DONE

### 1. Retry Support Wired But Not Enabled by Default

`RetryConfig` exists and `WithRetry()` option is available, but the `clean` command does not pass retry config — it defaults to no retries. To enable, someone needs to add `execution.WithRetry(cfg)` to the `RunCleaners` call in `clean.go`, ideally controlled by config or a `--retries` flag.

### 2. `MaxConcurrency` in RunSettings But Not Wired to Command

`RunSettings.MaxConcurrency` field exists but the clean command doesn't read it or pass `WithMaxConcurrency()` to `RunCleaners`. Needs a `--concurrency` flag or config-driven default.

### 3. Scan Profile Flag Still Ignored

`scan.go` accepts `--profile` but the parameter is discarded with `_`. Profile-based filtering for scan was not implemented (pre-existing issue, not introduced by migration).

### 4. Cleaners Still Use Hardcoded Defaults

`DefaultRegistryWithConfig(verbose, dryRun)` creates all 13 cleaners with hardcoded defaults (Homebrew mode=all, Docker prune=all, etc.). User profile settings from YAML config are not passed to individual cleaners. The DI container wraps this factory rather than registering individual cleaner providers.

### 5. `NotAvailableError` Typed but Not Adopted by Cleaners

The typed error exists in `cleaner/cleaner.go` but individual cleaners still return ad-hoc `errors.New("X not available")`. The `IsNotAvailableError()` function handles this via keyword fallback, but the cleaners haven't been migrated to return `*NotAvailableError` directly.

---

## c) NOT STARTED

1. **Remaining commands not migrated to DI** — `githistory`, `init`, `profile`, `config` commands still use direct constructor calls
2. **Adapter registration in DI** — `NixAdapter`, `HTTPClient`, `CacheManager` are concrete structs, not DI services
3. **Lifecycle management** — no `do.ShutdownerWithError` implementations on adapters
4. **Domain interface consolidation** — `cleaner.Cleaner` vs `domain.OperationHandler` / `GenerationCleaner` / `PackageCleaner` impedance mismatch
5. **Config hot-reload via DI scopes** — config is static `ProvideValue`
6. **Per-cleaner config wiring** — `OperationSettings` type-safe sub-structs exist but aren't connected to cleaner instantiation
7. **`flow.If` / `flow.Switch` patterns** — conditional execution (e.g., only run Docker if daemon is running)
8. **Step progress TUI** — BuildFlow has `ProgressBridge`; clean-wizard has no live progress
9. **Audit log of DI registrations** — BuildFlow has `samber-do-auditlog`
10. **`do.ExplainInjector` debug output** — behind a debug flag
11. **Resume/checkpoint support** — `flow.Workflow` state persistence
12. **`--keep-generations` flag** for Nix cleaner
13. **CLI command tests** — still missing entirely

---

## d) TOTALLY FUCKED UP / RISKS

### 1. No End-to-End CLI Integration Test

The integration tests in `execution/integration_test.go` test `RunCleaners` with a real registry, but **no test actually invokes the `clean` or `scan` cobra command end-to-end**. The wiring between `runCleanCommand()` → `di.New()` → `RegisterAllServices` → `RunCleaners` → `displayResults` is untested as a unit. Issues in flag parsing, config loading interaction with DI, or display rendering would only surface at runtime.

### 2. Stale Documentation References

`docs/status/2026-07-06_00-35_DI-WORKFLOW-MIGRATION.md` references `FlowBuilder`/`BranchFlow` as "dormant" — they've since been deleted. The first status report is now partially stale.

### 3. `buildCleanStepFunc` Records Result on Every Attempt

When retry is enabled, `makeCleanStepFunc` calls `collector.record()` on every attempt — including failed ones that get retried. This means a cleaner that fails twice then succeeds will have **3 entries** in `resultCollector.results`. The final `WorkflowResult.Steps` will contain duplicates. This is a **real bug** in the retry path that wasn't caught because the retry integration test doesn't inspect step count.

### 4. Nix `keepCount` Not Configurable

The old command layer hardcoded `CleanOldGenerations(ctx, 5)`. The new execution layer calls `Clean(ctx)` which delegates to `CleanOldGenerations(ctx, nc.keepCount)` with keepCount defaulting to 5 via the constructor. The value is preserved but not configurable through DI or CLI flags.

### 5. `scanCleanerReal` Function Left as Dead Code

After wiring scan to `execution.RunScans`, the old `scanCleanerReal` function in `scan.go` is still present but no longer called from the main path. It's dead code that should have been removed.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Fix retry duplicate recording** — `makeCleanStepFunc` should only record the final outcome, not every attempt
2. **Register individual cleaners as DI services** — each with its own provider reading from `OperationSettings`
3. **Make adapters interface-backed** — register behind interfaces with `do.As` aliasing
4. **Implement `do.ShutdownerWithError`** on resource-holding adapters
5. **Create application-global DI container** shared by all commands
6. **Consolidate `cleaner.Cleaner` and `domain.OperationHandler`**
7. **Migrate cleaners to return `*NotAvailableError`** instead of ad-hoc string errors

### Execution

8. **Add `--retries` CLI flag** wired to `RetryConfig`
9. **Add `--concurrency` CLI flag** wired to `MaxConcurrency`
10. **Wire `flow.If` for Docker** — check daemon running before executing
11. **Add timeout support** via `flow.Step(x).Timeout(dur)` per cleaner
12. **Add progress TUI** — live per-cleaner status

### Testing

13. **CLI integration test** — invoke `clean --dry-run` as a full command
14. **BDD tests for execution layer** using Ginkgo
15. **Test retry duplicate recording** — verify step count after retries

### Cleanup

16. **Remove `scanCleanerReal` dead function**
17. **Remove stale status report references** to deleted types
18. **Implement profile-based filtering for scan** command

---

## f) Top 25 Things to Do Next

| #   | Task                                                                     | Impact   | Effort |
| --- | ------------------------------------------------------------------------ | -------- | ------ |
| 1   | **Fix retry duplicate recording bug** in `makeCleanStepFunc`             | CRITICAL | S      |
| 2   | Remove dead `scanCleanerReal` function from `scan.go`                    | HIGH     | S      |
| 3   | Add `--retries` CLI flag wired to `RetryConfig`                          | HIGH     | S      |
| 4   | Add `--concurrency` CLI flag wired to `MaxConcurrency`                   | HIGH     | S      |
| 5   | CLI integration test: invoke `clean --dry-run` as full command           | HIGH     | M      |
| 6   | Test retry step count (verify no duplicates after retries)               | HIGH     | S      |
| 7   | Migrate cleaners to return `*NotAvailableError` instead of string errors | HIGH     | M      |
| 8   | Register individual cleaners as separate DI providers                    | HIGH     | L      |
| 9   | Pass user config profile settings to individual cleaner providers        | HIGH     | L      |
| 10  | Wire `MaxConcurrency` from `RunSettings` in clean/scan commands          | MEDIUM   | S      |
| 11  | Migrate `githistory` command to use DI container                         | MEDIUM   | S      |
| 12  | Migrate `init`, `profile`, `config` commands to use DI                   | MEDIUM   | M      |
| 13  | Make adapters interface-backed, register in DI with `do.As`              | HIGH     | L      |
| 14  | Implement `do.ShutdownerWithError` on `CacheManager`, `HTTPClient`       | MEDIUM   | S      |
| 15  | Add `flow.If` conditional for Docker cleaner (check daemon)              | MEDIUM   | S      |
| 16  | Add per-cleaner timeout via `flow.Timeout`                               | MEDIUM   | S      |
| 17  | Add BDD tests for execution layer (Ginkgo)                               | MEDIUM   | M      |
| 18  | Consolidate `cleaner.Cleaner` and `domain.OperationHandler`              | HIGH     | L      |
| 19  | Implement profile-based filtering for scan `--profile` flag              | MEDIUM   | M      |
| 20  | Add `do.ExplainInjector` debug output behind `--di-debug` flag           | LOW      | S      |
| 21  | Create `internal/bootstrap/` for application-global DI setup             | MEDIUM   | M      |
| 22  | Add progress TUI (like BuildFlow's `ProgressBridge`)                     | LOW      | L      |
| 23  | Add `--keep-generations` flag for Nix cleaner                            | LOW      | S      |
| 24  | Clean up stale status report references to deleted types                 | LOW      | S      |
| 25  | Add audit log of DI service registrations                                | LOW      | S      |

---

## g) Top #1 Question I Cannot Answer Myself

**Should the retry `NextBackOff` hook use `cleaner.IsNotAvailableError` to stop retrying "not available" errors immediately?**

Currently, if a cleaner returns "cargo not available" and retries are enabled, the workflow will retry 3 times with exponential backoff — even though retrying will never help (cargo won't suddenly become available). The `cleaner.IsNotAvailableError()` function exists and could be wired into a custom `NextBackOff` function to return `backoff.Stop` for non-retryable errors.

But I don't know whether the user wants this optimization or prefers the simpler "retry everything uniformly" approach. BuildFlow has sophisticated error classification (`errorfamily.Classify(err) == errorfamily.Transient`) that determines retryability — clean-wizard has `IsNotAvailableError()` but no `IsTransient()` / `IsRetryable()` classification for non-availability errors.
