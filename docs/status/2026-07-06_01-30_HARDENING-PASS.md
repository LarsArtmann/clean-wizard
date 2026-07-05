# Status Report: DI + Workflow Migration — Hardening & Cleanup

**Date:** 2026-07-06
**Session Scope:** Bug fixes, dead code removal, retry support, test hardening

---

## Executive Summary

Completed the DI + workflow migration hardening pass. Fixed 4 critical bugs, removed 675 lines of dead code, added retry support with exponential backoff, wired scan command to execution layer, and added integration + edge case tests. All 300+ tests pass.

---

## Completed Work

### Critical Bug Fixes

| #   | Fix                                                             | File                        |
| --- | --------------------------------------------------------------- | --------------------------- |
| 1   | Workflow errors no longer silently dropped when steps exist     | `execution/workflow.go`     |
| 2   | Panics in cleaners now recovered and recorded as failed steps   | `execution/builder.go`      |
| 3   | `isProcessRunning` now fails closed when `pgrep` is unavailable | `cleaner/golang_cleaner.go` |
| 4   | Results sorted by registration order for deterministic output   | `execution/results.go`      |

### Quality Improvements

| #   | Change                                                                                 | Impact                            |
| --- | -------------------------------------------------------------------------------------- | --------------------------------- |
| 5   | `resultCollector` mutex changed from `*sync.Mutex` to value `sync.Mutex` + constructor | Eliminates nil-pointer risk       |
| 6   | `TotalItemsFailed` now aggregates from ALL steps, not just successful ones             | Correct accounting                |
| 7   | Typed `NotAvailableError` + `IsNotAvailableError()` replaces fragile string matching   | Future-proof error classification |
| 8   | JSON error serialization verified correct (false positive from review)                 | No action needed                  |

### Dead Code Removed (675 lines)

| File                                             | Lines | Replaced By                           |
| ------------------------------------------------ | ----- | ------------------------------------- |
| `result/flow_builder.go`                         | 309   | `execution/builder.go` (go-workflow)  |
| `result/branch_flow.go`                          | 200   | `execution/builder.go` (go-workflow)  |
| `result/branch_flow_test.go`                     | ~150  | N/A                                   |
| `cleaner/parallel.go`                            | 166   | `execution/workflow.go` (go-workflow) |
| `cleaner/metrics.go:CleanAllParallelWithMetrics` | 24    | Removed                               |
| `cleaner/registry_factory.go:DefaultRegistry()`  | 10    | `DefaultRegistryWithConfig()` only    |

### New Features

| #   | Feature                                                         | Files                                        |
| --- | --------------------------------------------------------------- | -------------------------------------------- |
| 9   | `--config` flag for scan command                                | `scan.go`                                    |
| 10  | Scan command wired to `execution.RunScans` (parallel execution) | `scan.go`                                    |
| 11  | `RetryConfig` + `flow.Retry` with exponential backoff           | `execution/retry.go`, `execution/options.go` |
| 12  | `cenkalti/backoff/v4` promoted to direct dependency             | `go.mod`                                     |
| 13  | `MaxConcurrency` field in `RunSettings`                         | `di/options.go`                              |

### Test Coverage Added (16 new tests)

| File                            | Tests | Coverage                                                                   |
| ------------------------------- | ----- | -------------------------------------------------------------------------- |
| `di/di_test.go`                 | 8     | Container creation, service registration, override, error cases            |
| `execution/execution_test.go`   | 11    | Success/failure/skip/empty/unknown, ordering, classification               |
| `execution/integration_test.go` | 5     | Real registry dry-run, panic recovery, deterministic ordering, retry, scan |

---

## Architecture State

```
CLI Command
  └── di.New() → RegisterAllServices(injector, cfg, settings)
       └── di.CleanerRegistry(injector) → *cleaner.Registry (lazy singleton)
            └── execution.RunCleaners(ctx, registry, names, opts...)
                 └── Builder.BuildClean() → CompiledWorkflow
                      ├── flow.FuncIO per cleaner (with panic recovery)
                      ├── BeforeStep/AfterStep hooks (timing, verbose)
                      ├── flow.Retry (optional, with exponential backoff)
                      └── Workflow.Do(ctx) → resultCollector
                           └── sortedByRegistration() → *WorkflowResult
                                └── Succeeded() / Skipped() / Failed()
```
