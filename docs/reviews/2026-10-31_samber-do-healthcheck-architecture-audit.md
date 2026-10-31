# samber/do v2 + Health Checks — Architecture Audit

**Date:** 2026-10-31
**Scope:** `internal/di/`, `internal/adapters/`, `internal/cleaner/factory/`, `cmd/clean-wizard/commands/`
**Library:** `github.com/samber/do/v2 v2.1.0` (go.mod:32)
**Method:** full read of all 9 `do`-importing files, all 6 command files, adapter implementations, factory table, and the samber/do v2.1.0 lifecycle API (module source); DO-1→DO-6 anti-pattern checklist; build + full short test suite re-run (exit 0).

---

## 1. Executive Verdict

**Are we using samber/do v2 + health checks superbly? No — the DI core is genuinely good, but health checks do not exist, and a third of the registered services are dead code.**

| Quality          | Score | One-line verdict                                                                    |
| ---------------- | ----- | ----------------------------------------------------------------------------------- |
| Service-Oriented | 6/10  | Right patterns, but only 2 of 6 commands use DI and 3 of 4 adapter services are dead |
| Composable       | 8/10  | Textbook: `do.Package`, named services, `do.MustAs` aliasing, shared factory table  |
| Resilient        | 7/10  | Classified errors + retries at execution layer; but zero `Shutdowner` implementations |
| Self-Health      | 1/10  | Zero `Healthchecker` impls, zero `injector.HealthCheck()` calls, no doctor command  |

**Headline answers:**

1. **Health checks: 0% wired.** samber/do v2 ships an entire health subsystem (`do.Healthchecker`, `do.HealthcheckerWithContext`, `Scope.HealthCheck[WithContext]`, `do.HealthCheck[T]`, `HealthCheckGlobalTimeout` option) — none of it is referenced anywhere in this repo (verified by project-wide grep: zero matches for `HealthCheck`/`Healthcheck` outside the library). The domain already has the raw material (`Cleaner.IsAvailable(ctx)` at `internal/cleaner/cleaner.go:49-50`) but it is disconnected from the container. There is no `doctor`/`health` command.
2. **The "Adapter Pattern via DI" is aspirational, not real.** `AdaptersPackage` registers and aliases 4 services (`internal/di/adapters.go:20-40`), but the production Nix cleaner constructs its adapter inline, bypassing the DI service (`internal/cleaner/nix/nix.go:68-70`), and `KeyValueCache`/`HTTPRequester`/`Limiter` have **zero consumers** anywhere in the codebase.
3. **What IS superb:** the container lifecycle (`di.New()` → cleanup fn → deferred in both DI commands, DAG-wave shutdown, shutdown errors logged at debug, both paths unit-tested), the DO-1/2/4/6 compliance, typed accessors with classified errors, and the DI-agnostic execution layer.

---

## 2. DI Usage Inventory

All `do` imports live in `internal/di/` (9 files). Production surface:

| File               | Role                                                                       |
| ------------------ | -------------------------------------------------------------------------- |
| `container.go`     | `Container` wrapper; `New() (*Container, func())`; `Shutdown()`             |
| `providers.go`     | `RegisterAllServices` + `CleanerPackage` var + profile-settings resolution |
| `adapters.go`      | `AdaptersPackage` — 4 providers + `do.MustAs` aliases + 2 accessors        |
| `cleaners.go`      | Per-cleaner named services (`cleaner.<name>`) + registry-from-providers     |
| `accessors.go`     | Typed accessors: `Config`, `Settings`, `CleanerRegistry`, `Cleaner`        |
| `options.go`       | `RunSettings` struct (eager value)                                         |
| `test_helpers.go`  | `OverrideRegistry`/`OverrideSettings` (⚠ see F-6)                          |

**Consumers:** exactly two commands build a container — `clean` (`cmd/clean-wizard/commands/clean.go:172-183`) and `scan` (`scan.go:151-162`). The other four commands (`init`, `profile`, `config`, `githistory`) never touch DI; `githistory` constructs its cleaner directly (`githistory.go:283`, `githistory.go:587`).

**Registration order:** `RegisterAllServices` → eager `Config` + `RunSettings` values → `AdaptersPackage` → `CleanerPackage` (providers, then registry that invokes every per-cleaner named service). Everything is a lazy singleton except the two eager values. Correct.

---

## 3. DO-1 → DO-6 Anti-Pattern Scorecard

| ID    | Rule                                             | Verdict    | Evidence                                                                                                                                                   |
| ----- | ------------------------------------------------ | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DO-1  | No `Must*` in runtime paths                      | ✅ PASS    | Production accessors use `do.Invoke`/`do.InvokeNamed` with error propagation (`accessors.go`, `cleaners.go`). `do.MustAs` fires only at registration setup (`adapters.go:24,29,34,39`) — setup, not runtime. `do.MustInvoke` appears only in `container_shutdown_test.go:43`. |
| DO-2  | Every `do.New()` has a reachable `Shutdown()`    | ✅ PASS    | Single production `do.New()` (`container.go:29`), wrapped in the returned cleanup fn; both commands `defer cleanup()` (`clean.go:173`, `scan.go:152`). Test injectors either defer `Shutdown` (`providers_profile_test.go`) or hold no instantiated services (`di_test.go:146,155` — error-path tests). |
| DO-3  | No `Override*` outside tests                     | ⚠️ MINOR   | `OverrideRegistry`/`OverrideSettings` (`test_helpers.go:11,17`) wrap `do.OverrideValue` in a **non-test file**, and have no consumers outside `internal/di`'s own tests. Move to `_test.go` or delete (F-6). |
| DO-4  | No package-level injector                        | ✅ PASS    | No global `do.Injector` anywhere; a fresh container per command invocation, passed by parameter.                                                            |
| DO-5  | No `Invoke` inside loops                         | ✅ BY DESIGN | `cleaners.go:54` invokes named services in a loop — but each iteration resolves a *distinct* named service; hoisting is impossible and this is the intended aggregation pattern. Documented here so a future linter run doesn't "fix" it. |
| DO-6  | Shutdown is self-contained, no cross-service access | ✅ PASS (trivially) | No service implements a `Shutdowner*` interface yet, so nothing can violate it. The contract is already established and tested (`container_shutdown_test.go` failingShutdowner path) — future implementations must keep shutdown local. |

**Structural smells:** none. No service-locator passing (injector stops at the command boundary), no injector stored inside services.

---

## 4. Health Check Audit — the missing subsystem

### 4.1 What samber/do v2.1.0 offers (verified from module source)

- `do.Healthchecker` — `HealthCheck() error` (`di_lifecycle.go:21`)
- `do.HealthcheckerWithContext` — `HealthCheck(context.Context) error` (`di_lifecycle.go:42`)
- `injector.HealthCheck()` / `HealthCheckWithContext(ctx)` — parallel health check of **all** services in scope + ancestors, returns `map[string]error`, honors `HealthCheckGlobalTimeout` root option (`scope.go:294-326`)
- `do.HealthCheck[T](i)` / `HealthCheckNamed[WithContext]` — per-service checks
- `do.Shutdowner{,WithError,WithContext,WithContextAndError}` — lifecycle exit

### 4.2 What clean-wizard uses

**Nothing.** Zero `Healthchecker` implementations, zero `injector.HealthCheck()` calls, no `--health` flag, no `doctor` command (`main.go:18-23` registers exactly: clean, scan, init, profile, config, githistory).

### 4.3 What the domain already has (and doesn't connect)

- `Cleaner.IsAvailable(ctx) bool` (`internal/cleaner/cleaner.go:49-50`) — the domain-native availability check, consumed at scan-time by `getAvailableConfigs`.
- `NotAvailableError` + `errorfamily.Classify` → Infrastructure → "skipped" classification (`internal/execution/results.go`).
- `adapters.NixStore` knows how to detect the nix binary.

**Gap analysis:** availability is discovered *lazily during scan*, per-cleaner, as a bool (no failure *reason* surfaced to the container), and never as a container-level, parallelizable, timeout-bounded health report. samber/do would give that for free: implement `HealthcheckerWithContext` on a small adapter wrapper, call `injector.HealthCheckWithContext(ctx)` once, and render the `map[string]error` — with per-service errors instead of bools, a global timeout, and exit-code integration via `errorfamily`.

### 4.4 Concrete design (recommended)

```go
// internal/di/health.go (new)
func HealthReport(ctx context.Context, i do.Injector) map[string]error {
    return i.HealthCheckWithContext(ctx) // scope + ancestors, parallel, timeout-bounded
}
```

Adapters implement the interface:

```go
// internal/adapters/nix.go
func (a *NixAdapter) HealthCheck(ctx context.Context) error {
    if _, err := a.exec.CommandContext(ctx, "nix", "--version"); err != nil { /* wrap NotAvailableError */ }
    return nil
}
```

New command `clean-wizard doctor`: builds the container exactly like `clean`/`scan`, runs `HealthReport`, prints ✅/❌ per service, maps failures through `errorfamily.ExitCode` (Infrastructure → 69). JSON/SARIF output can reuse `format.FindingsToSARIF` with `category: health`.

**Complementary, not redundant:** keep `Cleaner.IsAvailable` for scan semantics (a cleaner being unavailable is a *normal scan outcome*, not a system fault), and use DI health checks for *infrastructure* truth (binary present, HTTP reachability, config resolvable). The doctor command surfaces both: container health per service, plus per-cleaner availability with reasons.

---

## 5. Finding: the DI adapter services are dead code

This is the most important structural finding. The wiring *looks* like the documented architecture ("the nix cleaner consumes `adapters.NixStore`" — AGENTS.md), but the production path never resolves any of the four registered services:

1. **Nix:** `factory.Nix` (`internal/cleaner/factory/factory.go:116-118`) calls `nix.NewNixCleaner(verbose, dryRun, ...)` → `NewNixCleanerWithStore(..., nil, ...)` → `store == nil` → constructs `adapters.NewNixAdapter(0, 0)` **inline** (`internal/cleaner/nix/nix.go:68-70`). The DI-registered `NixStore` service (`adapters.go:21-24`) is never invoked by any production code.
2. **CacheManager, HTTPClient, RateLimiter:** registered + aliased (`adapters.go:26-39`), consumed by **nothing** — project-wide grep for `KeyValueCache`/`HTTPRequester`/`adapters.Limiter` outside `internal/adapters` + `internal/di` returns zero matches. The only references are `internal/di`'s own tests asserting the services exist.
3. **Doc drift:** AGENTS.md says "the nix cleaner consumes `adapters.NixStore`, not `*NixAdapter`". Literally true (it consumes the *interface type*) but misleading — the *DI service* is bypassed via the nil-store fallback. The interface seam exists for tests (`NewNixCleanerWithStore` with a mock), not for DI.

**Why this matters more than usual:** dead DI services actively lie about the architecture. A reader (or future agent) assumes adapters are injected and swappable through the container, refactors accordingly, and gets burned. Either consume the services or delete the registrations.

**Fix sketch (consume, don't delete — the seam is the point):**

```go
// internal/di/cleaners.go — nix provider resolves the DI-registered store
if rc.Name == cleaner.CleanerNix {
    store, err := do.Invoke[adapters.NixStore](i)
    ...
    return nix.NewNixCleanerWithStore(settings.Verbose, settings.DryRun, store, keepCount...), nil
}
```

Or cleaner: extend `factory.CleanerConstructorFunc` to accept resolved dependencies, with the DI layer passing them in and the non-DI githistory path passing nil-defaults. Either way, one owner for adapter construction.

---

## 6. Architecture Quality Assessment

### 6.1 Service-Oriented — 6/10

- ✅ Every cleaner is an individually addressable named service (`cleaner.<name>`, `cleaners.go:25`) — per-cleaner invocation without materializing the registry is exactly right.
- ✅ Interfaces at the right seams: `adapters.NixStore`, `KeyValueCache`, `HTTPRequester`, `Limiter` (`internal/adapters/interfaces.go`); consumers depend on interfaces, `do.MustAs` validates aliasing eagerly at registration.
- ❌ Only 2/6 commands go through DI. `githistory` hand-wires its cleaner; `init`/`profile`/`config` don't need cleaners, so the inconsistency is tolerable — but `githistory` is a cleaner command and belongs in the registry/DI path.
- ❌ 3 of 4 adapter services are dead (§5).

### 6.2 Composable — 8/10

- ✅ `do.Package` groups providers (`AdaptersPackage`, `CleanerPackage`), mirroring BuildFlow's composition.
- ✅ Single construction truth: `factory.CleanerRegistrations()` feeds both the DI providers and (via `DefaultRegistryWithConfig`) the non-DI path — no split-brain cleaner table.
- ✅ Execution layer is DI-agnostic (`execution.RunCleaners(ctx, *cleaner.Registry, ...)`) — the container stops at the command boundary; workflow code stays container-free.
- ✅ Profile settings resolution is a pure function over the container (`resolveProfileOperationSettings`), composable into every cleaner provider.
- ⚠️ `RegisterAllServices` always returns `nil` — its `error` signature is a lie today. Either validate composition (e.g. invoke config once, validate profile exists) or drop the error (F-7).
- ⚠️ `Container.Injector()` re-opens the container to raw `do.*` calls at every call site — acceptable for a small CLI, but the accessor layer (`Config`, `Settings`, `CleanerRegistry`, `Cleaner`, `NixStore`, `KeyValueCache`) is the better habit; every new resolution should go through an accessor.

### 6.3 Resilient — 7/10

- ✅ Every resolution failure is wrapped with `errorfamily` (Rejection) with `di.*` codes (`accessors.go`, `cleaners.go`, `providers.go:48`) — DI failures get classified exit codes at the CLI boundary like every other error.
- ✅ Retry + backoff live at the execution layer (`internal/execution/retry.go`), not the DI layer — correct placement.
- ✅ Shutdown failure handling is thoughtful: logged at Debug, never masks command outcome, silent on happy path (`container.go:31-43`, both paths tested in `container_shutdown_test.go`).
- ❌ No service implements any `Shutdowner*` — yet some hold resources: `HTTPClient` wraps a resty client (connection pool), `CacheManager` wraps an otter cache. For short-lived CLI processes this is benign, but the resilience story is incomplete: the container *can* shut things down gracefully and *nothing uses it* except tests.
- ❌ Availability failures are discovered at scan time with bool granularity; no container-level health surface to fail fast before a run (§4).

### 6.4 Self-Health — 1/10

The subsystem exists in the library, the domain has availability primitives, the error families and exit codes are ready — and nothing is connected. This is the gap between "we have a DI container" and "we have a self-describing service architecture". See §4.4 for the concrete wiring; it is a half-day of work for the whole payoff.

---

## 7. Minor Findings

| ID   | Severity | Finding                                                                                                                                                                |
| ---- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| F-1  | Info     | `CacheManager.cleanupInterval` (`cache_manager.go:32,49`) is stored and never used; `FlushExpired()` (`cache_manager.go:138-141`) is an empty body. Dead fields on a dead (in DI terms) service. |
| F-2  | Info     | Lint `ireturn` warnings on accessors returning interfaces are **intentional and correct** for DI accessors (return the seam, not the impl). Accepted tradeoff — suppress with `//nolint:ireturn` + reason or configure exclusions. |
| F-3  | Info     | `resolveProfileOperationSettings` returns `(nil, nil)` when no profile is selected (`providers.go:42-44`) — flagged `nilnil` by lint, but it is deliberate domain semantics (no profile → factory defaults). Document with a comment, don't "fix". |
| F-4  | Low      | `do.MustAs` panics on misconfiguration. At registration time this is acceptable (DO-1 exception: setup paths), but `do.As` (error-returning) + a classified Rejection would make `RegisterAllServices`' error signature honest. |
| F-5  | Low      | AGENTS.md adapter-pattern bullet needs a correction (§5.3): interface seam exists for tests, DI service bypassed in production.                                                            |
| F-6  | Medium   | `test_helpers.go` ships `do.OverrideValue` wrappers in a production file with zero external consumers (DO-3 hygiene). Move into `di_test.go` or delete.                            |
| F-7  | Low      | `RegisterAllServices` error return is constant `nil` — make it validate or simplify the signature.                                                                               |

---

## 8. Prioritized Recommendations

| Priority | Action                                                                                                                                                             | Effort  | Payoff                                                          |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------- | --------------------------------------------------------------- |
| **P0**   | Wire the adapter seam for real: nix provider resolves `adapters.NixStore` from DI (§5 fix sketch); decide consume-or-delete for CacheManager/HTTPClient/RateLimiter   | ~2h     | Kills the dead-service lie; makes the documented architecture true |
| **P0**   | Implement `do.HealthcheckerWithContext` on `NixAdapter` (→ `nix --version` exec) and optionally `HTTPClient`; add `clean-wizard doctor` running `injector.HealthCheckWithContext` + `errorfamily.ExitCode` (§4.4) | ~half-day | The entire self-health story; fail-fast before runs; scriptable `doctor --json` |
| **P1**   | Move `githistory` onto the DI/registry path                                                                                                                         | ~2h     | Service orientation becomes consistent across cleaner commands   |
| **P1**   | Relocate/delete `OverrideRegistry`/`OverrideSettings` (F-6)                                                                                                         | 10min   | DO-3 hygiene                                                     |
| **P2**   | Set `do.RootOptions.HealthCheckGlobalTimeout` when doctor lands; consider per-run health gate in `clean` (`--check-health` flag)                                      | ~1h     | Bounded, predictable health checks                               |
| **P2**   | Implement `ShutdownerWithError` on `CacheManager`/`HTTPClient` (flush + close idle conns) so the tested shutdown contract has real users                             | ~1h     | Lifecycle completeness                                           |
| **P2**   | Fix `RegisterAllServices` signature (F-7) + AGENTS.md doc drift (F-5)                                                                                               | 30min   | Honesty of the composition root                                  |

---

## 9. Verification

- Build: `GOEXPERIMENT=jsonv2 go build ./...` — pass
- Tests: `GOEXPERIMENT=jsonv2 go test ./... -short -count=1` — pass (exit 0), including `internal/di` (11 tests: registration, overrides, error paths, shutdown logging both paths)
- Every claim above carries a `file:line` reference re-verified against the working tree at audit time
- samber/do v2.1.0 API claims verified against module source (`di_lifecycle.go`, `scope.go`), not documentation
