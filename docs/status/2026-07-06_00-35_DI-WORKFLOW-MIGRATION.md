# Status Report: DI + Workflow Migration (samber/do v2 + Azure/go-workflow)

**Date:** 2026-07-06 00:35
**Session Scope:** Migrating clean-wizard to the BuildFlow DI + workflow pattern

---

## Executive Summary

Migrated the project from hand-wired dual-registry architecture to a proper **dependency injection container** (`samber/do v2`) and **workflow orchestration engine** (`Azure/go-workflow`), mirroring the pattern proven in `/home/lars/projects/BuildFlow/`.

**197 lines added, 164 lines deleted** across 8 modified files. **932 new lines** across 13 new files in two new packages. All existing tests pass (298+) plus **19 new tests** for the DI and execution packages.

---

## a) FULLY DONE ✓

### 1. Dependencies Added

- `github.com/samber/do/v2 v2.0.0` — dependency injection
- `github.com/Azure/go-workflow v0.1.13` — workflow orchestration
- Transitive deps: `cenkalti/backoff/v4`, `benbjohnson/clock`
- `go mod verify` passes clean

### 2. `internal/di/` Package (6 files, 294 lines)

| File              | Lines | Purpose                                                         |
| ----------------- | ----- | --------------------------------------------------------------- |
| `container.go`    | 39    | `Container` wrapper around `do.Injector`, `New()` + cleanup     |
| `options.go`      | 9     | `RunSettings` struct (Verbose, DryRun)                          |
| `providers.go`    | 50    | `RegisterAllServices`, `CleanerPackage`, lazy registry provider |
| `accessors.go`    | 39    | Typed accessors: `Config()`, `Settings()`, `CleanerRegistry()`  |
| `test_helpers.go` | 18    | `OverrideRegistry()`, `OverrideSettings()` for test doubles     |
| `di_test.go`      | 139   | **8 tests** covering registration, override, error cases        |

**Key patterns replicated from BuildFlow:**

- `do.Provide` for lazy singleton (cleaner registry resolves `RunSettings` via `do.Invoke`)
- `do.ProvideValue` for eager values (config, settings)
- `do.Package` for grouped registrations
- `do.OverrideValue` for test doubles
- Typed accessors wrapping `do.Invoke[T]` with error context

### 3. `internal/execution/` Package (7 files, 638 lines)

| File                | Lines | Purpose                                                                             |
| ------------------- | ----- | ----------------------------------------------------------------------------------- |
| `doc.go`            | 10    | Package doc — declares DI-agnostic design principle                                 |
| `options.go`        | 28    | `RunOption` pattern: `WithMaxConcurrency`, `WithVerbose`                            |
| `results.go`        | 130   | `StepResult`, `WorkflowResult`, `resultCollector`, error classification             |
| `hooks.go`          | 55    | `makeBeforeHook`, `makeAfterHook` — timing, verbose output                          |
| `builder.go`        | 146   | `Builder.BuildClean`, `Builder.BuildScan` — compiles cleaners → `flow.FuncIO` steps |
| `workflow.go`       | 83    | `RunCleaners`, `RunScans` — entry points; configures & runs `workflow.Do(ctx)`      |
| `execution_test.go` | 186   | **11 tests** covering success/failure/skip/empty/unknown-cleaner                    |

**Key patterns replicated from BuildFlow:**

- `flow.FuncIO[I, O]` for typed steps (each cleaner becomes a step)
- `flow.Step(step).BeforeStep(before).AfterStep(after)` for hooks
- `Workflow.MaxConcurrency` and `Workflow.DontPanic` configuration
- Per-step error collection (non-short-circuiting — one failure doesn't stop others)
- `flow.ErrWorkflow` unpacked into per-step results
- Options pattern (`RunOption` functional options)

### 4. Command Layer Rewired

| File                | Change                                                                                                                                         |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `clean.go`          | Creates DI container, resolves registry, calls `execution.RunCleaners`, updated `outputJSON` to use `WorkflowResult`                           |
| `clean_execute.go`  | Rewritten to consume `*execution.WorkflowResult`; removed `cleanResult` struct, `executeCleaners`, `handleCleanerError`, `isNotAvailableError` |
| `cleaner_config.go` | `GetCleanerConfigs` now accepts `*cleaner.Registry` param instead of calling `DefaultRegistry()` internally                                    |
| `scan.go`           | Creates DI container, resolves registry, passes it to `GetCleanerConfigs` and `scanCleanerReal`                                                |

### 5. Removed `cleaner_implementations.go` (357 lines deleted)

Eliminated the entire dual-registry dispatch layer:

- `cleanerRegistry` map (13 entries)
- `runCleaner` dispatcher
- 13 `run*Cleaner` functions
- 6 generic helper functions (`runGenericCleaner`, `runGenericCleanerWithError`, etc.)
- 3 factory functions

### 6. Moved Go Process Safety Check

- `hasOtherGoProcesses()` / `isProcessRunning()` moved from command layer to `internal/cleaner/golang_cleaner.go`
- Check now runs inside `GoCleaner.Clean()` — where it belongs domain-wise
- Added `ErrGoProcessesRunning` sentinel error

### 7. Tests Pass

- `go build ./...` — clean
- `go vet ./internal/di/... ./internal/execution/...` — clean
- `go test ./... -short` — all pass (298+ existing + 19 new)
- `go mod verify` — clean

### 8. Documentation Updated

- `AGENTS.md` updated with new architecture patterns, DI + Workflow section, updated dependency list

---

## b) PARTIALLY DONE

### 1. DI Container is Per-Command, Not Application-Global

**Current:** `clean` and `scan` commands each create their own `di.New()` container.
**BuildFlow's approach:** Single container created at application startup, shared across all commands.
**Why partial:** Works correctly but doesn't fully match the pattern. The other commands (`init`, `profile`, `config`, `githistory`) don't use DI at all yet.

### 2. Cleaner Registry is Still Default-Configured

**Current:** `DefaultRegistryWithConfig(verbose, dryRun)` creates all 13 cleaners with hardcoded defaults.
**BuildFlow's approach:** Each service is individually registered with its own provider function, allowing per-service configuration from the config.
**Why partial:** The DI container wraps the existing `DefaultRegistryWithConfig` rather than registering individual cleaner providers.

### 3. Execution Layer Parallelism

**Current:** `Workflow.MaxConcurrency` defaults to 0 (unlimited), but the `RunCleaners` options support `WithMaxConcurrency(n)`.
**Why partial:** The clean command doesn't pass `WithMaxConcurrency`, so it runs with Go-workflow's default behavior. Sequential ordering is not guaranteed.

---

## c) NOT STARTED

### 1. Remaining Commands Not Migrated to DI

- `githistory` command — still uses direct constructor calls
- `init` command — still uses direct constructor calls
- `profile` command — still uses direct constructor calls
- `config` command — still uses direct constructor calls

### 2. Adapter Registration in DI

- `internal/adapters/` (Nix, HTTP, Cache, RateLimiter, Exec) are still concrete structs imported directly by cleaners
- BuildFlow registers adapters as DI services with interfaces (`do.As[*Concrete, Interface]`)
- No `do.ShutdownerWithError` implementations on adapters

### 3. Lifecycle Management

- No adapters implement `do.ShutdownerWithError`
- No adapters implement `do.HealthcheckerWithContext`
- BuildFlow has compile-time guarantees: `var _ do.ShutdownerWithError = (*cache.Cache)(nil)`

### 4. Domain Interface Consolidation

- `cleaner.Cleaner` vs `domain.OperationHandler` / `GenerationCleaner` / `PackageCleaner` — impedance mismatch still exists
- BuildFlow uses `do.As` aliasing to register concrete → interface

### 5. Removing Dormant `result.FlowBuilder` / `BranchFlow` / `ParallelFlow`

- These were the old flow combinators, now superseded by go-workflow
- They still exist in `internal/result/` but are unused in production paths
- AGENTS.md notes them as dormant

### 6. Config Hot-Reload via DI

- Config is registered as a static `ProvideValue` — no watcher or scope-based override

---

## d) TOTALLY FUCKED UP / RISKS

### 1. No Integration Test of the Actual Workflow Execution

The new `execution.RunCleaners` was tested with mock cleaners only. No test runs the actual clean command end-to-end through the workflow engine with real cleaners. The workflow's parallel execution behavior, hook ordering, and error propagation with real system cleaners is **unverified beyond unit tests**.

### 2. Cleaners Created with Default Config Regardless of User Profile

The DI provider calls `cleaner.DefaultRegistryWithConfig(verbose, dryRun)` which hardcodes all cleaner defaults (e.g., Homebrew mode=all, Docker prune=all, Go cache=all, BuildCache 30d). **User profile settings from the YAML config are NOT passed to individual cleaners.** This was a pre-existing problem, but the migration didn't fix it — the DI layer just wraps the old factory.

### 3. Result Ordering is Non-Deterministic

`go-workflow` runs steps in parallel by default. The `WorkflowResult.Steps` slice order depends on completion order, not registration order. The display functions iterate `CleanResultsMap()` (a map), so output ordering was already non-deterministic — but the results table may show cleaners in different order run-to-run.

### 4. Nix Cleaner Special-Case Lost

The old `runNixCleaner` called `CleanOldGenerations(ctx, 5)` with an explicit keepCount. The new execution layer calls `Clean(ctx)` which delegates to `CleanOldGenerations(ctx, nc.keepCount)`. The keepCount defaults to 5 via `NewNixCleaner`'s variadic parameter, so behavior is preserved — but the explicit `5` is now buried in the constructor default rather than visible at the call site.

### 5. `format.CleanResultsToJSON` Still Takes Separate Maps

The JSON output path reconstructs `skipped` and `failed` maps from `WorkflowResult.Skipped()` and `.Failed()`. This works but is a step backward from the unified `WorkflowResult` — the format package should eventually accept `*WorkflowResult` directly.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Register individual cleaners as DI services** — not one monolithic `DefaultRegistryWithConfig`. Each cleaner should have its own provider function reading config from the DI container, matching BuildFlow's per-service provider pattern.
2. **Make adapters interface-backed** — register `NixAdapter`, `HTTPClient`, `CacheManager` behind interfaces in DI with `do.As` aliasing. Cleaners should depend on interfaces, not concrete structs.
3. **Implement `do.ShutdownerWithError`** on adapters that hold resources (CacheManager, HTTPClient connection pools).
4. **Create a single application-global DI container** in `main.go` or a bootstrap layer, shared by all commands via context.
5. **Consolidate `cleaner.Cleaner` and `domain.OperationHandler`** — eliminate the impedance mismatch. One interface, one role.

### Execution

6. **Add retry support** via `flow.Step(x).Retry(opts...)` — BuildFlow uses `cenkalti/backoff` exponential backoff with error classification. Cleaners that hit transient errors (Nix store lock, Docker daemon busy) would benefit.
7. **Add timeout support** via `flow.Step(x).Timeout(dur)` — per-cleaner timeouts configurable from YAML.
8. **Wire `WithMaxConcurrency`** from config or a `--concurrency` flag.
9. **Add `flow.If` / `flow.Switch` patterns** — e.g., only run Docker cleaner if Docker daemon is running, or branch Nix cleaner based on generation count.
10. **Add step progress reporting** — BuildFlow has a `ProgressBridge` TUI; clean-wizard could show live progress per cleaner.

### Cleanup

11. **Remove `result.FlowBuilder`, `BranchFlow`, `ParallelFlow`** — superseded by go-workflow, dormant code.
12. **Remove `cleaner.ParallelExecutor`** and `Registry.CleanAllParallel` — superseded by execution layer.
13. **Update `format.CleanResultsToJSON`** to accept `*execution.WorkflowResult` directly.
14. **Add `--keep-generations` flag** for Nix cleaner instead of burying the default in the constructor.

### Testing

15. **Integration test** — run the full `clean --dry-run` command through the workflow engine.
16. **BDD tests** for the execution layer (Ginkgo).
17. **Test parallel execution** — verify concurrent cleaners don't race on shared resources.

---

## f) Up to 25 Things to Do Next (Prioritized)

| #   | Task                                                                              | Impact   | Effort |
| --- | --------------------------------------------------------------------------------- | -------- | ------ |
| 1   | **Integration test: run `clean --dry-run` through workflow engine end-to-end**    | Critical | M      |
| 2   | Register individual cleaners as separate DI providers (not monolithic factory)    | High     | L      |
| 3   | Pass user config profile settings to individual cleaner providers                 | High     | L      |
| 4   | Migrate `githistory` command to use DI container                                  | Medium   | S      |
| 5   | Migrate `init`, `profile`, `config` commands to use DI container                  | Medium   | M      |
| 6   | Make adapters interface-backed and register in DI with `do.As`                    | High     | L      |
| 7   | Implement `do.ShutdownerWithError` on `CacheManager`, `HTTPClient`                | Medium   | S      |
| 8   | Add retry support to execution layer (`flow.Retry` with backoff)                  | High     | M      |
| 9   | Add per-cleaner timeout support (`flow.Timeout`)                                  | Medium   | S      |
| 10  | Wire `WithMaxConcurrency` from config or `--concurrency` CLI flag                 | Medium   | S      |
| 11  | Remove dormant `result.FlowBuilder` / `BranchFlow` / `ParallelFlow`               | Medium   | S      |
| 12  | Remove `cleaner.ParallelExecutor` and `Registry.CleanAllParallel`                 | Low      | S      |
| 13  | Update `format.CleanResultsToJSON` to accept `*WorkflowResult`                    | Low      | S      |
| 14  | Add `--keep-generations` flag for Nix cleaner                                     | Low      | S      |
| 15  | Add deterministic sorting of `WorkflowResult.Steps` by registration order         | Medium   | S      |
| 16  | Add `flow.If` conditional for Docker cleaner (check daemon running)               | Medium   | S      |
| 17  | Add BDD tests for execution layer (Ginkgo)                                        | Medium   | M      |
| 18  | Consolidate `cleaner.Cleaner` and `domain.OperationHandler` interface             | High     | L      |
| 19  | Add `do.ExplainInjector` debug output behind `--di-debug` flag                    | Low      | S      |
| 20  | Create `internal/bootstrap/` package for application-global DI setup              | Medium   | M      |
| 21  | Add step progress TUI (like BuildFlow's `ProgressBridge`)                         | Low      | L      |
| 22  | Add resume/checkpoint support via `flow.Workflow` state                           | Low      | L      |
| 23  | Add `flow.Switch` for Nix cleaner generation-count branching                      | Low      | M      |
| 24  | Add audit log of DI service registrations (like BuildFlow's `samber-do-auditlog`) | Low      | S      |
| 25  | **Commit the migration** as a feature branch with proper commit message           | Critical | S      |

---

## g) Top #1 Question I Cannot Answer Myself

**Should we commit this migration now, or first add integration tests that run the full `clean --dry-run` command through the new workflow engine?**

The unit tests pass (19 new + 298 existing), and `go build ./...` is clean. But no test actually executes `execution.RunCleaners` with real cleaner implementations through the go-workflow DAG. If there's a subtle issue with how `flow.FuncIO` wraps the `Clean(ctx)` method — e.g., context cancellation behavior, panic recovery via `DontPanic: true`, or step dependency ordering — it would only surface at runtime.

I cannot determine the user's risk tolerance for committing architectural changes without end-to-end integration verification versus iterating further before committing.
