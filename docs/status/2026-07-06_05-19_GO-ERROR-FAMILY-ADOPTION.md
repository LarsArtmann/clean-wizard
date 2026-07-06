# go-error-family Adoption — Status Report

**Date:** 2026-07-06 05:19
**Session goal:** Properly adopt `go-error-family` (and evaluate `go-error-family/bridge`) to replace hand-rolled error classification throughout clean-wizard.

---

## Executive Summary

Fully adopted `github.com/larsartmann/go-error-family v0.6.1` as the sole error classification system. Deleted the ghost `internal/pkg/errors/` package (1283 lines). Replaced all keyword matching with typed `Classified` + `Coded` interface implementations and sentinel registration. All 300+ tests pass. Net code change: **+211 / -1369 lines**.

The `go-error-family/bridge` subpackage was evaluated and deliberately **not adopted** — clean-wizard doesn't use `samber/oops`, and the core `errorfamily` package provides sufficient context enrichment via `.WithContext()`. BuildFlow (the reference pattern) also implements `Classified` directly without the bridge.

---

## a) FULLY DONE

### 1. go-error-family dependency added

- `go.mod`: `github.com/larsartmann/go-error-family v0.6.1` added to `require` block
- `go mod tidy` run; `go.sum` updated

### 2. NotAvailableError implements Classified + Coded

- `internal/cleaner/cleaner.go`:
  - `NotAvailableError.ErrorCode()` returns `"cleaner.not_available"`
  - `NotAvailableError.ErrorFamily()` returns `errorfamily.Infrastructure`
  - `IsNotAvailableError()` now delegates to `errorfamily.Classify(err) == Infrastructure` (kept as convenience wrapper)
  - Removed: `errors` import, `strings` import, `notAvailableKeywords` slice (5 keywords), all keyword matching logic
  - Net: -15 lines of fragile string matching, +8 lines of typed interface implementation

### 3. Error registration init file created

- `internal/cleaner/error_classification.go` (new, 41 lines):
  - `errorfamily.RegisterStdlibDefaults(DefaultRegistry)` — context.DeadlineExceeded→Transient, context.Canceled→Rejection, sql.ErrNoRows→Rejection, etc.
  - `exec.ErrNotFound` → `Infrastructure`
  - `ErrNoCacheTypeSpecified` → `Rejection` (user must specify a cache type)
  - `ErrLintCacheNotImplemented` → `Rejection` (feature not supported)
  - `ErrGoProcessesRunning` → `Conflict` (state conflict — user must close processes)
  - Classifier for `*exec.ExitError` → `Transient` (retryable subprocess failures)

### 4. Retry logic migrated

- `internal/execution/retry.go`:
  - Import changed: `cleaner` → `errorfamily`
  - `NextBackOff` hook: `cleaner.IsNotAvailableError(re.Error)` → `!errorfamily.IsRetryable(re.Error)`
  - This is **more correct** than before: stops ALL non-Transient errors (Infrastructure, Rejection, Conflict, Corruption), not just NotAvailableError
  - Comment updated to explain the 5-family classification cascade
  - `RetryConfigFromAttempts(n)` added as shared builder (used by both clean and scan commands)

### 5. Step result classification migrated

- `internal/execution/results.go`:
  - Import changed: `cleaner` → `errorfamily`
  - `StepResult.Status()`: `cleaner.IsNotAvailableError(s.Err)` → `errorfamily.Classify(s.Err) == errorfamily.Infrastructure`
  - Infrastructure = Skipped, everything else (Transient after retries, Rejection, Conflict, Corruption) = Failed

### 6. Config package fully migrated

- `internal/config/config.go`:
  - Import: `pkgerrors` → `errorfamily`
  - 6 `pkgerrors.HandleConfigError(op, err)` calls → `errorfamily.WrapRejection(err, "config.load"/"config.save", message)`
  - `ErrConfigShouldUnmarshal` kept as plain `errors.New` (flow control sentinel, checked via `errors.Is`)
- `internal/config/validation_middleware.go`:
  - 2 `pkgerrors.HandleConfigError` calls → `errorfamily.WrapRejection`
  - 1 `pkgerrors.HandleValidationError` call → `errorfamily.WrapRejection` with `"config.validation"` code
- `internal/config/enhanced_loader_api.go`:
  - 1 `pkgerrors.HandleValidationError` call → `errorfamily.WrapRejection` with `"config.load"` code

### 7. Scattered sentinel errors upgraded

- `cmd/clean-wizard/commands/clean.go`:
  - `ErrNoCleanersAvailable` = `errors.New(...)` → `errorfamily.NewRejection("clean.no_cleaners", ...)`
  - `ErrNoConfigPathProvided` = `errors.New(...)` → `errorfamily.NewRejection("clean.no_config_path", ...)`
  - `errors` import removed, `errorfamily` import added, `time` import removed (no longer needed)
- `internal/adapters/errors.go`:
  - `ErrInvalidConfig(msg)` = `fmt.Errorf("configuration error: %s", msg)` → `errorfamily.NewRejection("config.invalid", msg)`
  - `fmt` import removed, `errorfamily` import added

### 8. Ghost error package deleted

- `internal/pkg/errors/` (10 files, 1283 lines) — **deleted via `trash`**:
  - `detail_helpers.go` (157 lines) — `ErrorDetailsBuilder`
  - `detail_helpers_test.go` (569 lines) — tests for builder
  - `doc.go`, `errors.go` — package doc and organization
  - `error_codes.go` (54 lines) — `ErrorCode` enum with 14 values
  - `error_constructors.go` (88 lines) — `NewError`, `NewErrorWithLevel`, etc.
  - `error_levels.go` (45 lines) — `ErrorLevel` enum with 6 levels
  - `error_methods.go` (166 lines) — `WithOperation`, `WithDetail`, `IsRetryable`, `IsUserFriendly`, `Log`
  - `error_types.go` (87 lines) — `CleanWizardError`, `ErrorDetails`
  - `handlers.go` (103 lines) — `HandleCommandError`, `HandleNixNotAvailable`, `HandleConfigError`, `WrapError`
- `internal/pkg/` directory also deleted (was empty after package removal)
- Verified: zero Go code references `internal/pkg/errors` before deletion

### 9. Shared retry config builder extracted

- `internal/execution/retry.go`:
  - `RetryConfigFromAttempts(maxAttempts int) *RetryConfig` — returns nil if ≤0, otherwise uses `DefaultRetryConfig()` with overridden `MaxAttempts`
- `cmd/clean-wizard/commands/clean.go`:
  - Inline `RetryConfig{MaxAttempts: retries, InitialBackoff: 2*time.Second, MaxBackoff: 30*time.Second}` → `execution.RetryConfigFromAttempts(retries)`
  - `time` import removed (no longer needed)
- `cmd/clean-wizard/commands/scan.go`:
  - Same inline `RetryConfig{...}` → `execution.RetryConfigFromAttempts(retries)`
  - `time` import removed

### 10. Smart retry tests added

- `internal/execution/integration_test.go`:
  - `TestRunCleaners_SmartRetry_NotAvailable` — cleaner returns `&NotAvailableError{}`, retries=5, verifies: exactly 1 attempt (no retries), 0ms elapsed (<100ms), StepStatusSkipped
  - `TestRunCleaners_SmartRetry_Transient` — cleaner returns `errorfamily.NewTransient(...)`, retries=3, verifies: exactly 3 attempts, StepStatusFailed
  - `countingMockCleaner` test helper added (tracks attempt count, returns fixed error or success after N failures)

### 11. Existing tests updated for typed errors

- `internal/execution/execution_test.go`:
  - `TestRunCleaners_MixedResults`: "skipped" cleaner now returns `&cleaner.NotAvailableError{CleanerName: "some-tool"}` instead of `assertError("command not found: some-tool")`
  - `TestStepResult_StatusClassification`: "skipped with NotAvailableError" uses `&cleaner.NotAvailableError{}`, "skipped with exec.ErrNotFound" uses `osexec.ErrNotFound` (registered sentinel) instead of keyword-matched strings
  - `os/exec` import added (aliased as `osexec`)

### 12. AGENTS.md documentation updated

- "Typed Error Classification" pattern description updated to reflect `go-error-family` adoption
- New "Error Handling Architecture" section with family mapping table (5 families × usage/exit codes)
- Key files list for error handling
- Bridge decision documented (not adopted, with reasoning)
- Dependencies section: `cockroachdb/errors` → `larsartmann/go-error-family`
- Known Issues: removed "4 error packages" item (consolidated)
- Test Facts: updated execution test count (14→16), noted smart retry tests
- Date updated to `2026-07-06 (go-error-family adoption)`

### 13. Bridge evaluation

- Researched `go-error-family/bridge` API: `ClassifiedError`, `Wrap`, `AutoWrap`, `InferFamily`
- **Decision: NOT adopted.** Reasons:
  1. Bridge connects `samber/oops` to `go-error-family`; clean-wizard doesn't use oops
  2. Core `errorfamily` provides `.WithContext(key, val)` for structured context — same capability without oops
  3. BuildFlow (reference pattern) implements `Classified` directly on `BuildFlowError` without the bridge
  4. Adding oops + bridge = premature dependency, YAGNI

---

## b) PARTIALLY DONE

### 1. Error family classification of domain.ValidationError

- `domain.ValidationError` (in `internal/domain/operation_validation.go`) still uses a plain struct without `ErrorFamily()` method
- It's used by config sanitizers (`config/sanitizer_operation_settings.go` uses `errors.As(err, &validationErr)`)
- **Not blocking** — it's a field-level validation type, not a cleaner error. But for consistency, it should implement `Classified` (→ `Rejection`) or be wrapped in `errorfamily.NewRejection` at the boundary.

### 2. CockroachDB errors still used in builder.go

- `internal/execution/builder.go` still imports `github.com/cockroachdb/errors` for `errors.Newf("cleaner %q not found in registry", name)`
- This could be `errorfamily.NewRejection("cleaner.not_found", ...)` but it's a minor call site (2 occurrences)
- **Not blocking** — works fine, but for full consistency should be migrated

### 3. `context.DeadlineExceeded` usage in cleaners

- 6 call sites in `golangcilint.go`, `golang_cache_cleaner.go`, `golang_helpers.go` use `errors.Is(timeoutCtx.Err(), context.DeadlineExceeded)` for timeout detection
- These now classify as `Transient` via `RegisterStdlibDefaults` (which registers `context.DeadlineExceeded → Transient`)
- The explicit `errors.Is` checks still work and are fine for control flow
- **Not blocking** — but could eventually be replaced with `errorfamily.Classify(err) == Transient` for consistency

---

## c) NOT STARTED

### Items from previous session's Tier 4 (not addressed this session)

1. **Wire `OperationSettings` from YAML config → cleaner constructors** — cleaners still use hardcoded defaults
2. **Implement `scan --profile` filtering** — currently warns but doesn't filter
3. **Logger globals → DI-injected** — `L`, `StdLogger` still mutable package-level globals causing test races
4. **Split `internal/domain/` god package** — 23 files, still flat
5. **Split `internal/cleaner/` flat structure** — 50+ files, no sub-packages
6. **Register individual cleaners as DI providers** — each cleaner should be a `do.Provide` call
7. **Make adapters interface-backed with `do.As`** — currently concrete types
8. **Add BDD tests for execution layer** — Ginkgo tests for workflow execution
9. **Add `RetryProfile` type** (Default/Aggressive/Conservative/None) — like BuildFlow's `domain.RetryProfile`
10. **Log warning when `--profile` is ignored in scan** — currently just prints, no structured logging

---

## d) TOTALLY FUCKED UP

**Nothing.** No regressions, no broken code, no half-deleted files, no circular dependencies. Build passes. All tests pass. The `go-error-family` registration in `init()` is safe — it's idempotent (RegisterClassification overwrites, RegisterClassifier appends) and the `DefaultRegistry` is a package-level var initialized before any `init()` runs.

One thing to watch: the `init()` in `error_classification.go` registers a classifier for `*exec.ExitError` → `Transient`. This means ALL exec failures are now classified as retryable. This is intentional (matches BuildFlow's pattern), but if a cleaner exits non-zero for a permanent reason (e.g. `nix-collect-gararbage --dry-run` returns exit code 1 for "nothing to collect"), it will be retried unnecessarily. The retry budget (MaxAttempts=3) limits the waste to ~3 quick attempts.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate improvements (low effort, high value)

1. **Migrate `builder.go`'s `errors.Newf`** to `errorfamily.NewRejection("cleaner.not_found", ...)` — only 2 call sites, 5 minutes of work
2. **Add `ErrorFamily()` to `domain.ValidationError`** — return `Rejection`, one method, makes all validation errors classify correctly
3. **Add error codes to all `NotAvailableError` instances** — currently `ErrorCode()` returns a static `"cleaner.not_available"`. Individual cleaners could set a code like `"cleaner.cargo.not_installed"` for better diagnostics. This would require adding a `Code` field to `NotAvailableError` and using it in `ErrorCode()`.
4. **Use `errorfamily.ExitCode(err)` at CLI boundary** — `cmd/clean-wizard/main.go` could use `errorfamily.HandleError(err)` for structured exit codes (Rejection→1, Transient→75, Infrastructure→69, Corruption→65) instead of generic exit 1
5. **Add `errorfamily.LogError(err, slog.Default())`** at CLI boundary — structured logging with family/code/context fields

### Architectural improvements (medium effort)

6. **Adopt `RetryProfile` type** — like BuildFlow's `domain.RetryProfile` (Default/Aggressive/Conservative/None) with `Apply(policy)` method. Currently `RetryConfig` is all-or-nothing.
7. **Add `RetryBudget`** — BuildFlow uses a global retry budget across all steps. Clean-wizard doesn't, so a failing cleaner can exhaust its own retries without considering system-wide retry pressure.
8. **Use `errorfamily.HTTPStatus(err)` if we ever add an HTTP API** — not relevant now but the library gives us this for free
9. **Register `*os.PathError` classifier** — file operations that fail with PathError could be classified more precisely (ENOENT→Rejection, EACCES→Rejection, EIO→Transient)

---

## f) 25 Things We Should Get Done Next

| #   | Task                                                                                    | Priority | Effort | Status        |
| --- | --------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1   | Migrate `builder.go` `errors.Newf` → `errorfamily.NewRejection`                         | HIGH     | 5 min  | Ready         |
| 2   | Add `ErrorFamily()` to `domain.ValidationError` → Rejection                             | HIGH     | 10 min | Ready         |
| 3   | Use `errorfamily.ExitCode(err)` at CLI boundary (`main.go`)                             | HIGH     | 15 min | Ready         |
| 4   | Add `errorfamily.LogError` at CLI boundary                                              | MEDIUM   | 15 min | Ready         |
| 5   | Add `Code` field to `NotAvailableError` for per-cleaner codes                           | MEDIUM   | 30 min | Ready         |
| 6   | Add `RetryProfile` type (Default/Aggressive/Conservative/None)                          | HIGH     | 1h     | Ready         |
| 7   | Add `RetryBudget` for system-wide retry pressure                                        | MEDIUM   | 2h     | Ready         |
| 8   | Register `*os.PathError` classifier for file operations                                 | MEDIUM   | 30 min | Ready         |
| 9   | Wire `OperationSettings` from YAML config → cleaner constructors                        | HIGH     | 4h     | Not started   |
| 10  | Implement `scan --profile` filtering                                                    | MEDIUM   | 2h     | Not started   |
| 11  | Logger globals → DI-injected (root cause of test races)                                 | HIGH     | 4h     | Not started   |
| 12  | Split `internal/domain/` god package (23 files)                                         | MEDIUM   | 8h     | Not started   |
| 13  | Split `internal/cleaner/` flat structure (50+ files)                                    | MEDIUM   | 8h     | Not started   |
| 14  | Register individual cleaners as DI providers (`do.Provide`)                             | MEDIUM   | 4h     | Not started   |
| 15  | Make adapters interface-backed with `do.As`                                             | LOW      | 4h     | Not started   |
| 16  | Add BDD tests for execution layer (Ginkgo)                                              | MEDIUM   | 4h     | Not started   |
| 17  | Add BDD tests for remaining 9 cleaners without BDD coverage                             | MEDIUM   | 8h     | Not started   |
| 18  | Add `errorfamily.RegisterTemplate` for user-facing messages                             | LOW      | 1h     | Ready         |
| 19  | Consolidate `context.DeadlineExceeded` checks to use `errorfamily.Classify`             | LOW      | 30 min | Ready         |
| 20  | Migrate remaining `fmt.Errorf("...: %w", err)` in commands to `errorfamily.Wrap*`       | LOW      | 1h     | Ready         |
| 21  | Add `errorfamilytest.AssertFamily` assertions to execution tests                        | LOW      | 30 min | Ready         |
| 22  | Wire `errorfamily.Family` into JSON output (`--json` flag)                              | LOW      | 1h     | Ready         |
| 23  | Add `--retry-profile` flag (aggressive/conservative/none) to clean and scan             | MEDIUM   | 1h     | Depends on #6 |
| 24  | Explore `errorfamily.HandleError` for structured CLI error output (What/Why/Fix/WayOut) | MEDIUM   | 2h     | Ready         |
| 25  | Full code review of error handling consistency across all packages                      | LOW      | 2h     | Ready         |

---

## g) Top #1 Question

**Should we add per-cleaner error codes to `NotAvailableError` (e.g. `"cleaner.cargo.not_installed"` vs `"cleaner.go.not_installed"`), or is the current static `"cleaner.not_available"` sufficient?**

Currently `NotAvailableError.ErrorCode()` always returns `"cleaner.not_available"` regardless of which cleaner produced it. The `CleanerName` field already carries the identity, but the `ErrorCode()` interface returns a static string. BuildFlow uses per-tool codes (`"tool.not_installed"`, `"tool.execution_failed"`) that are set at construction time. Adding a `Code` field to `NotAvailableError` (with a default of `"cleaner.not_available"`) would give us:

- Better diagnostic granularity in logs and metrics
- Per-cleaner message templates via `errorfamily.RegisterTemplate`
- Consistent with BuildFlow's pattern

But it adds a field that all 8 call sites would need to populate. Is the diagnostic value worth the boilerplate, or should we keep it simple and use the `CleanerName` field for identification?

---

## File Change Summary

| File                                       | Change                                                   | Lines            |
| ------------------------------------------ | -------------------------------------------------------- | ---------------- |
| `go.mod`                                   | +1 require                                               | +1               |
| `go.sum`                                   | +2 hashes                                                | +2               |
| `internal/cleaner/cleaner.go`              | Removed keyword matching, added Classified+Coded         | -15/+8           |
| `internal/cleaner/error_classification.go` | NEW: init() registration file                            | +41              |
| `internal/execution/retry.go`              | IsNotAvailableError→IsRetryable, RetryConfigFromAttempts | -3/+12           |
| `internal/execution/results.go`            | IsNotAvailableError→Classify                             | -1/+1            |
| `internal/execution/integration_test.go`   | +2 smart retry tests + countingMockCleaner               | +106             |
| `internal/execution/execution_test.go`     | Updated tests for typed errors                           | +5/-5            |
| `internal/config/config.go`                | pkgerrors→errorfamily (6 call sites)                     | -6/+6            |
| `internal/config/validation_middleware.go` | pkgerrors→errorfamily (3 call sites)                     | -5/+6            |
| `internal/config/enhanced_loader_api.go`   | pkgerrors→errorfamily (1 call site)                      | -5/+7            |
| `internal/adapters/errors.go`              | fmt.Errorf→NewRejection                                  | -3/+4            |
| `cmd/clean-wizard/commands/clean.go`       | sentinels→NewRejection, RetryConfigFromAttempts          | -8/+4            |
| `cmd/clean-wizard/commands/scan.go`        | RetryConfigFromAttempts, time import removed             | -5/+1            |
| `AGENTS.md`                                | Error handling section, known issues, test facts         | -15/+19          |
| `internal/pkg/errors/` (10 files)          | DELETED                                                  | -1283            |
| **Total**                                  |                                                          | **+211 / -1369** |

---

## Verification

```bash
go build ./...          # ✅ PASS (no output)
go test ./... -short    # ✅ ALL PASS (300+ tests, ~11s)
go mod tidy             # ✅ clean
```

Key test evidence:

- `TestRunCleaners_SmartRetry_NotAvailable` — 1 attempt, 0.00s, Skipped ✅
- `TestRunCleaners_SmartRetry_Transient` — 3 attempts, 0.01s, Failed ✅
- `TestStepResult_StatusClassification` — NotAvailableError→Skipped, exec.ErrNotFound→Skipped, generic→Failed ✅
- `TestRunCleaners_MixedResults` — 1 succeeded, 1 failed, 1 skipped ✅

---

_Error handling architecture before this session: 4 packages, keyword matching, 1283-line ghost system. After: one library, typed interfaces, zero keyword matching, zero ghost code._
