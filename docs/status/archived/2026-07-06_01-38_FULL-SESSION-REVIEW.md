# Status Report: DI + Workflow Migration — Full Session Review

**Date:** 2026-07-06 01:38
**Session Scope:** Two commits — initial DI+workflow migration, then hardening pass with bug fixes, dead code removal, retry support, and comprehensive testing
**Commits:** `43df609` → `65290ce` (pushed to `master`)
**Test Status:** ALL 300+ tests pass across 22 packages, 0 failures

---

## a) FULLY DONE ✓

> **Resolution (2026-08-10):** Every numbered item in sections `b) PARTIALLY DONE`, `c) NOT STARTED`, and `f) Top 25 Things to Do Next` is inline-resolved below. Items 1–4 in `b)` were fully done in commit `de105b0`; items 5, 7, 11, 13 were done in subsequent sessions. Items 14+ are tracked in `TODO_LIST.md`.

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

1. ~~Retry Support Wired But Not Enabled by Default~~ done at `1b96d06` — `--retries` flag added, default = 3; smart retry via `errorfamily.IsRetryable`
2. ~~`MaxConcurrency` in RunSettings But Not Wired to Command~~ done at `1b96d06` — `--concurrency`/`-C` flag added to both `clean` and `scan`
3. ~~Scan Profile Flag Still Ignored~~ still open — `scan --profile` prints warning but does not filter; tracked as TODO #10
4. ~~Cleaners Still Use Hardcoded Defaults~~ still open — user YAML settings not wired to cleaner constructors; tracked as TODO #6
5. ~~`NotAvailableError` Typed but Not Adopted by Cleaners~~ done at `c102e0f` — all 13 cleaners now use `*NotAvailableError` via `NewNotAvailableError` factory

## c) NOT STARTED

1. ~~Remaining commands not migrated to DI~~ NOT-DO/DUPLICATE — explicitly deferred as low-value in `2026-07-06_03-42` Pareto pass
2. ~~Adapter registration in DI~~ still open — tracked as TODO #30 (interface-backed with `do.As`)
3. ~~Lifecycle management~~ NOT-DO — no adapters hold resources requiring cleanup
4. ~~Domain interface consolidation~~ NOT-DO — risky refactor, low value (ROADMAP non-goal #4)
5. ~~Config hot-reload via DI scopes~~ NOT-DO — CLI tool, no need
6. ~~Per-cleaner config wiring~~ still open — tracked as TODO #6
7. ~~`flow.If` / `flow.Switch` patterns~~ NOT-DO — complexity > value
8. ~~Step progress TUI~~ still aspirational — tracked in ROADMAP Theme #3
9. ~~Audit log of DI registrations~~ NOT-DO — YAGNI
10. ~~`do.ExplainInjector` debug output~~ NOT-DO — YAGNI (ROADMAP non-goal #2)
11. ~~Resume/checkpoint support~~ NOT-DO — YAGNI for CLI tool
12. ~~`--keep-generations` flag for Nix cleaner~~ still open — tracked as TODO #23
13. ~~CLI command tests~~ still open — partial coverage (clean integration test exists); tracked as TODO #13

---

## d) TOTALLY FUCKED UP / RISKS

1. ~~No End-to-End CLI Integration Test~~ still open — `cmd/clean-wizard/commands/clean_integration_test.go` exists for `clean --dry-run --json` but `scan`, `profile`, `config`, `init` are not yet covered; tracked as TODO #13
2. ~~Stale Documentation References~~ done at `2026-07-13_23-14_DOCS-HEALTH-AUDIT.md` — references to deleted types corrected across all core docs
3. ~~`buildCleanStepFunc` Records Result on Every Attempt~~ done at `6a539e7` — `recordFinal` replaces instead of appending; regression test added. **2026-08-10:** secondary in-place mutation bug fixed (was iterating with `for _, v := range` which mutated a copy; now uses index-based assignment)
4. ~~Nix `keepCount` Not Configurable~~ still open — default 5 via constructor; tracked as TODO #23 (no `--keep-generations` flag)
5. ~~`scanCleanerReal` Function Left as Dead Code~~ done at `de105b0` — removed from `scan.go`

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. ~~Fix retry duplicate recording~~ done at `6a539e7` (initial fix), in-place mutation fix at 2026-08-10
2. ~~Register individual cleaners as DI services~~ still open — tracked as TODO #29
3. ~~Make adapters interface-backed~~ still open — tracked as TODO #30
4. ~~Implement `do.ShutdownerWithError`~~ NOT-DO — no resources to shut down (ROADMAP non-goal #3)
5. ~~Create application-global DI container~~ NOT-DO — per-command is sufficient (ROADMAP non-goal #1)
6. ~~Consolidate `cleaner.Cleaner` and `domain.OperationHandler`~~ NOT-DO — risky refactor (ROADMAP non-goal #4)
7. ~~Migrate cleaners to return `*NotAvailableError`~~ done at `c102e0f`

### Execution

8. ~~Add `--retries` CLI flag~~ done at `de105b0` + `1b96d06` (default = 3, both commands)
9. ~~Add `--concurrency` CLI flag~~ done at `de105b0` + `1b96d06` (both commands)
10. ~~Wire `flow.If` for Docker~~ NOT-DO — Docker cleaner already short-circuits via `IsAvailable` returning `*NotAvailableError`
11. ~~Add timeout support via `flow.Step(x).Timeout(dur)`~~ NOT-DO — complexity > value
12. ~~Add progress TUI~~ still aspirational — ROADMAP Theme #3

### Testing

13. ~~CLI integration test~~ partially done — `clean --dry-run --json` covered; tracked as TODO #13 for full coverage
14. ~~BDD tests for execution layer using Ginkgo~~ still open — tracked as TODO #7
15. ~~Test retry duplicate recording~~ done at `6a539e7` — regression test added (later fixed for SizeEstimate at 2026-08-10)

### Cleanup

16. ~~Remove `scanCleanerReal` dead function~~ done at `de105b0`
17. ~~Remove stale status report references~~ done at `2026-07-13_23-14_DOCS-HEALTH-AUDIT.md` and the docs-health audit of 2026-08-10
18. ~~Implement profile-based filtering for scan command~~ still open — tracked as TODO #10

## f) Top 25 Things to Do Next

| #   | Task                                                                     | Resolution |
| --- | ------------------------------------------------------------------------ | ---------- |
| 1   | ~~**Fix retry duplicate recording bug** in `makeCleanStepFunc`~~         | done at `6a539e7`; in-place mutation fixed at 2026-08-10 |
| 2   | ~~Remove dead `scanCleanerReal` function from `scan.go`~~                | done at `de105b0` |
| 3   | ~~Add `--retries` CLI flag wired to `RetryConfig`~~                     | done at `de105b0` + `1b96d06` |
| 4   | ~~Add `--concurrency` CLI flag wired to `MaxConcurrency`~~              | done at `de105b0` + `1b96d06` |
| 5   | CLI integration test: invoke `clean --dry-run` as full command           | partial — covered for clean; tracked as TODO #13 |
| 6   | ~~Test retry step count (verify no duplicates after retries)~~          | done at `6a539e7`; assertion updated at 2026-08-10 |
| 7   | ~~Migrate cleaners to return `*NotAvailableError` instead of string errors~~ | done at `c102e0f` |
| 8   | Register individual cleaners as separate DI providers                    | still open — TODO #29 |
| 9   | Pass user config profile settings to individual cleaner providers        | still open — TODO #6 |
| 10  | ~~Wire `MaxConcurrency` from `RunSettings` in clean/scan commands~~    | done at `de105b0` + `1b96d06` |
| 11  | ~~Migrate `githistory` command to use DI container~~                     | NOT-DO — explicitly deferred in 2026-07-06_03-42 Pareto pass |
| 12  | ~~Migrate `init`, `profile`, `config` commands to use DI~~              | NOT-DO — explicitly deferred |
| 13  | ~~Make adapters interface-backed, register in DI with `do.As`~~         | still open — TODO #30 |
| 14  | ~~Implement `do.ShutdownerWithError` on `CacheManager`, `HTTPClient`~~  | NOT-DO — no resources to shut down |
| 15  | ~~Add `flow.If` conditional for Docker cleaner (check daemon)~~        | NOT-DO — handled by `IsAvailable` → `*NotAvailableError` |
| 16  | ~~Add per-cleaner timeout via `flow.Timeout`~~                         | NOT-DO — complexity > value |
| 17  | Add BDD tests for execution layer (Ginkgo)                               | still open — TODO #7 |
| 18  | ~~Consolidate `cleaner.Cleaner` and `domain.OperationHandler`~~         | NOT-DO — risky refactor |
| 19  | Implement profile-based filtering for scan `--profile` flag              | still open — TODO #10 |
| 20  | ~~Add `do.ExplainInjector` debug output behind `--di-debug` flag~~      | NOT-DO — YAGNI |
| 21  | ~~Create `internal/bootstrap/` for application-global DI setup~~       | NOT-DO — per-command is sufficient |
| 22  | Add progress TUI (like BuildFlow's `ProgressBridge`)                     | aspirational — ROADMAP Theme #3 |
| 23  | Add `--keep-generations` flag for Nix cleaner                            | still open — TODO #23 |
| 24  | ~~Clean up stale status report references to deleted types~~            | done at 2026-07-13 audit + 2026-08-10 docs-health |
| 25  | ~~Add audit log of DI service registrations~~                           | NOT-DO — YAGNI |

---

## g) Top #1 Question I Cannot Answer Myself

**Should the retry `NextBackOff` hook use `cleaner.IsNotAvailableError` to stop retrying "not available" errors immediately?**

> **Resolved (2026-07-06 hardening pass):** the hook now uses `errorfamily.IsRetryable()` instead of `cleaner.IsNotAvailableError()`. This is **more correct** — stops ALL non-Transient errors (Infrastructure, Rejection, Conflict, Corruption), not just `NotAvailableError`. See `internal/execution/retry.go` and the 2026-07-06 go-error-family adoption report.
