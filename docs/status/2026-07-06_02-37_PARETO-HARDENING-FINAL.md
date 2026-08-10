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

1. ~~No Test for the Retry Duplicate Recording Fix~~ done at `6a539e7` — `TestRunCleaners_Retry` now asserts `len(Steps) == 1` after 3 attempts. **2026-08-10:** assertion migrated from deprecated `FreedBytes` to `SizeEstimate.Value()`; in-place mutation bug in `recordFinal` fixed.
2. `--retries` Default of 0 Means Production Has No Retries — **resolved at `1b96d06`** — default raised to 3 with smart retry via `errorfamily.IsRetryable()`
3. `ErrGoCacheNotAvailable` Changed Type but Sentinel Comparison May Break — **resolved at `c102e0f`** — sentinel removed entirely; comparison goes through `IsNotAvailableError()` → `errors.As`
4. Integration Test Takes 15 Seconds — **partial fix at `6a539e7`** — `testing.Short()` skip guard added; test still takes 15s when run explicitly
5. `--profile` Flag Silently Ignored in Scan Command — **partially resolved at `1b96d06`** — flag now prints warning; full filtering still tracked as TODO #10

## e) WHAT WE SHOULD IMPROVE

### Immediate (Low Effort, High Confidence)

1. ~~Add test asserting step count after retry~~ done at `6a539e7` + 2026-08-10
2. ~~Remove dead `record()` method~~ done at `6a539e7`
3. ~~Migrate remaining 2 cleaners to `*NotAvailableError`~~ done at `c102e0f`
4. ~~Add `--retries` and `--concurrency` to scan command~~ done at `1b96d06`
5. ~~Tag integration test with `//go:build integration` or use `testing.Short()` skip~~ done at `6a539e7`

### Near-Term (Medium Effort, Real Value)

6. ~~Set `--retries` default to 3 (or 2) for production resilience~~ done at `1b96d06`
7. Wire `OperationSettings` from config profiles — still open; tracked as TODO #6
8. Implement `scan --profile` filtering — still open; tracked as TODO #10
9. Add `--timeout` flag — NOT-DO — complexity > value
10. Consolidate 4 error packages into one coherent design — done at `edaff33` (`go-error-family` adoption); `internal/pkg/errors/` ghost package removed

### Strategic (Higher Effort, Architectural)

11. Register individual cleaners as DI providers — still open; tracked as TODO #29
12. Make adapters interface-backed with `do.As` aliasing — still open; tracked as TODO #30
13. Consolidate `cleaner.Cleaner` vs `domain.OperationHandler` — NOT-DO (ROADMAP non-goal #4)
14. Add BDD tests for execution layer (Ginkgo) — still open; tracked as TODO #7
15. Add progress TUI — aspirational (ROADMAP Theme #3)

## f) Top 25 Things to Do Next

| #   | Task                                                                               | Resolution |
| --- | ---------------------------------------------------------------------------------- | ---------- |
| 1   | ~~**Add test: assert step count = 1 after retry**~~                               | done at `6a539e7` + 2026-08-10 |
| 2   | ~~**Remove dead `record()` method** from resultCollector~~                       | done at `6a539e7` |
| 3   | ~~**Migrate `projectsmanagementautomation` + `systemcache` to `*NotAvailableError`**~~ | done at `c102e0f` |
| 4   | ~~**Add `--retries` and `--concurrency` to scan command**~~                        | done at `1b96d06` |
| 5   | ~~**Tag/skip integration test for short mode**~~                                  | done at `6a539e7` |
| 6   | ~~**Set `--retries` default to 3** (or 2) for production resilience~~              | done at `1b96d06` |
| 7   | Wire `OperationSettings` from config to cleaner constructors                       | still open — TODO #6 |
| 8   | Implement `scan --profile` filtering or remove the flag                            | still open — TODO #10 |
| 9   | Add `--timeout` per-cleaner flag                                                    | NOT-DO |
| 10  | ~~**Consolidate 4 error packages into one coherent design**~~                     | done at `edaff33` |
| 11  | ~~Clean up stale status reports (reference deleted types)~~                       | done at 2026-07-13 + 2026-08-10 audits |
| 12  | Add `--keep-generations` flag for Nix cleaner                                      | still open — TODO #23 |
| 13  | Register individual cleaners as separate DI providers                              | still open — TODO #29 |
| 14  | Make adapters interface-backed with `do.As`                                        | still open — TODO #30 |
| 15  | Consolidate `cleaner.Cleaner` vs `domain.OperationHandler`                         | NOT-DO |
| 16  | Implement `do.ShutdownerWithError` on adapters with resources                      | NOT-DO |
| 17  | Add BDD tests for execution layer (Ginkgo)                                         | still open — TODO #7 |
| 18  | Add progress TUI                                                                   | aspirational |
| 19  | Add `do.ExplainInjector` debug output behind `--di-debug`                          | NOT-DO |
| 20  | Create application-global DI bootstrap shared by all commands                      | NOT-DO |
| 21  | Migrate `githistory` command to DI                                                 | NOT-DO |
| 22  | Add `flow.If` conditional for Docker                                               | NOT-DO |
| 23  | Add audit log of DI registrations                                                  | NOT-DO |
| 24  | ~~Fix `ErrGoCacheNotAvailable` sentinel — make it a value, not pointer~~           | done at `c102e0f` (sentinel removed entirely) |
| 25  | Profile-guided cleaner selection                                                   | still open — TODO #6 (subsumed) |

## g) Top #1 Question I Cannot Answer Myself

**Should `--retries` default to 0 (disabled, current) or 3 (enabled with `DefaultRetryConfig`)?**

> **Resolved at `1b96d06`:** default is **3**. Smart retry via `errorfamily.IsRetryable()` prevents wasting budget on non-retryable errors. `--retries 0` remains available as escape hatch.
