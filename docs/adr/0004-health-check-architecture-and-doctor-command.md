# ADR-0004: Health-Check Architecture and the Doctor Command

**Date:** 2026-10-31
**Status:** Accepted
**Deciders:** Lars Artmann

---

## Context

The 2026-10-31 samber/do audit scored Self-Health 1/10: zero
`do.Healthchecker*` implementations, zero `injector.HealthCheck()` calls, no
health command. samber/do v2.1.0 ships the whole subsystem unused.

Two design questions had to be answered before wiring anything:

1. What does "healthy" mean for a _cleaner-targeted_ CLI, where
   `Cleaner.IsAvailable(ctx)` already reports tool availability?
2. Which services implement health checks, and how are failures surfaced?

## Decision

### 1. Two probes, two domains (do not merge)

| Probe                         | Domain          | `false`/error means              | CLI effect                         |
| ----------------------------- | --------------- | -------------------------------- | ---------------------------------- |
| `Cleaner.IsAvailable(ctx)`    | scan outcome    | "this cleaner is not applicable" | cleaner _skipped_, exit 0          |
| `adapters.X.HealthCheck(ctx)` | container fault | "the system is broken"           | classified Infrastructure, exit 69 |

Both happen to probe `nix --version` today, but they answer different
questions: availability is a _result_; a failed health check is an _error_.
Documented in `docs/DOMAIN_LANGUAGE.md` (IsAvailable / HealthCheck / Doctor).

### 2. Health checks live on adapters, and lazy services must be materialized

- `NixAdapter` implements `HealthCheck(ctx)` (guarded by a local
  `ContextHealthchecker` mirror so `internal/adapters` stays free of the DI
  library; `internal/di` asserts the do-side contract).
- **Critical samber/do semantics:** `serviceLazy.healthcheck` returns nil when
  the service was never invoked — an unbuilt container reports _everything
  healthy_. `di.HealthReport` therefore materializes every service
  (config, run settings, NixStore alias, registry, all 13 cleaners) before
  running `injector.HealthCheckWithContext`. Construction failures merge into
  the failing service's entry.
- Aliases forward health checks to the target service, so the
  `adapters.NixStore` alias and the `*NixAdapter` concrete service both report.

### 3. The doctor command is the health surface

`clean-wizard doctor`:

- builds the container exactly like clean/scan (config load → `di.New` →
  `RegisterAllServices`),
- reports per-service ✅/❌ with classified error detail (human) or
  `--json` / `--sarif` (machine channels, mutually exclusive),
- exits 0 when all services are healthy, 69 (Infrastructure,
  `errorfamily.ExitCode`) otherwise — CI/script gate friendly.

`clean --check-health` reuses the same report as an opt-in pre-run gate:
unhealthy container → abort with exit 69 before any cleaning starts.

### 4. Naming

`doctor` (not `health`): convention for "inspect my installation and tell me
what's broken" (cf. `brew doctor`, `npm doctor`). The flag for the pre-run
gate is `--check-health` (verb-first reads as an action).

### 5. Failure semantics

Health failures classify as Infrastructure (`doctor.unhealthy`, exit 69,
non-retryable): a missing binary will not fix itself by retrying. The 30s
`DefaultHealthCheckGlobalTimeout` bounds the whole report; a hung probe is
itself a failure.

## Consequences

- Self-Health goes from 1/10 to a wired, tested, documented subsystem.
- New container services with external dependencies should implement
  `HealthCheck(ctx)` and get a guard in `internal/adapters/lifecycle.go`
  (mirror) or `internal/di` (do-side).
- Doctor output is part of the CLI contract: `--json` shape is tested
  (`TestBuildDoctorJSON_ShapeSnapshot`); changes need a schema note.
- Machine channels are mutually exclusive by rejection
  (`doctor.json_sarif_conflict`), mirroring clean's `--report`/`--json` rule.

## References

- Audit: `docs/reviews/2026-10-31_samber-do-healthcheck-architecture-audit.md`
- Code: `internal/di/health.go`, `internal/adapters/nix.go` (HealthCheck),
  `cmd/clean-wizard/commands/doctor.go`
- Boundary glossary: `docs/DOMAIN_LANGUAGE.md`
- Related: ADR-0003 (container truth), ticket #56 (ExecWithTimeout —
  HealthCheck deliberately uses its own exec path)

## Appendix: startup measurement (T27, 2026-10-31)

`internal/di/bench_test.go`, AMD RYZEN AI MAX+ 395, go1.27, `-benchtime 30x`:

| Benchmark                          | ns/op | What it measures                                        |
| ---------------------------------- | ----- | ------------------------------------------------------- |
| `BenchmarkRegisterAllServices`     | ~86k  | container create + register all lazy services           |
| `BenchmarkRegistryMaterialization` | ~275k | registration + resolve all 13 cleaners + registry build |

Both are sub-millisecond: the lazy per-cleaner registry costs nothing
measurable at startup, and eager materialization would save nothing. The
lazy design stays; re-measure only if the registry grows an order of
magnitude.

## Architecture re-score (T27.03, post-remediation)

Baseline (audit §6) → post-remediation:

| Quality          | Before | After | Why                                                                            |
| ---------------- | ------ | ----- | ------------------------------------------------------------------------------ |
| Service-Oriented | 6/10   | 9/10  | Every DI service has a consumer; the container matches documented architecture |
| Composable       | 8/10   | 9/10  | Store seam is real (do.Override + fake), factory takes dependencies explicitly |
| Resilient        | 7/10   | 8/10  | Health checks classified; clean can gate on container health; honest estimates |
| Self-Health      | 1/10   | 8/10  | doctor (human/json/sarif, exit 69), Healthchecker impls, pre-run gate          |

Not 10/10: doctor covers adapter-level health only (per-cleaner deep probes
remain scan-domain), and the CLI framework (cobra) is not yet cmdguard
(TODO 32, deferred with checklist in TODO_LIST).
