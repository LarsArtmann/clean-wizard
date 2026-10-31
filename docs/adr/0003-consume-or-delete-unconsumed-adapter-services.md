# ADR-0003: Consume-or-Delete for Unconsumed Adapter Services

**Date:** 2026-10-31
**Status:** Accepted
**Deciders:** Lars Artmann

---

## Context

The 2026-10-31 samber/do health-check architecture audit
(`docs/reviews/2026-10-31_samber-do-healthcheck-architecture-audit.md`,
finding F-2) found that 3 of the 4 DI-registered adapter services had zero
consumers:

| Service                    | Interface alias       | Production consumers |
| -------------------------- | --------------------- | -------------------- |
| `*adapters.NixAdapter`     | `adapters.NixStore`   | 0 (worse: the nix cleaner bypassed DI and constructed its adapter inline, `internal/cleaner/nix/nix.go:69`) |
| `*adapters.CacheManager`   | `adapters.KeyValueCache` | 0 |
| `*adapters.HTTPClient`     | `adapters.HTTPRequester` | 0 |
| `*adapters.RateLimiter`    | `adapters.Limiter`    | 0 |

A DI container that registers services nobody resolves is not an
architecture — it is a claim about an architecture. Docs said
"consumers resolve interfaces"; the code did not.

## Decision

Two-part fix, both accepted:

1. **Consume the one adapter with a real consumer.** The nix cleaner now
   receives `adapters.NixStore` from the container: `internal/di/cleaners.go`
   resolves the interface via `do.Invoke` and constructs through
   `factory.NixWithStore` (added 2026-10-31). Tests override the store with
   `do.Override` and a fake (`internal/di/di_test.go`,
   `TestNixCleaner_UsesContainerInjectedStore`).

2. **Delete the three unconsumed registrations** (`CacheManager`,
   `HTTPClient`, `RateLimiter`): their `do.Provide`/`do.MustAs` entries and
   the `di.KeyValueCache` accessor are removed from `internal/di/adapters.go`.

### Registration-vs-implementation scope

Only the **DI registrations** are deleted. The adapter implementations
(`internal/adapters/cache_manager.go`, `http_client.go`,
`rate_limiter.go`) and their interfaces stay, with compile-time guards. A
library type with tests and no caller is inventory; a DI service with no
consumer is a lie. The implementations return to the container the moment a
real consumer exists — re-adding one service is a 3-line change in
`registerAdapters`.

## Consequences

- Every service in the container now has at least one production consumer or
  is resolved as part of the cleaner registry — the container matches the
  documented architecture.
- The audit's "dead adapters" finding is closed; the AGENTS.md adapter bullet
  no longer carries the bypass caveat.
- Health checks (planned doctor command) will not report healthy-but-unused
  services, which would have inflated confidence in untested paths.
- `internal/adapters` carries unused-but-tested exported types until a
  consumer appears; acceptable for an internal library package.

## References

- Audit: `docs/reviews/2026-10-31_samber-do-healthcheck-architecture-audit.md` (F-2)
- Plan: `docs/planning/2026-10-31_11-50_samber-do-healthcheck-pareto-execution-plan.md` (T01.05/T01.06)
- Code: `internal/di/adapters.go`, `internal/di/cleaners.go`,
  `internal/cleaner/factory/factory.go` (`NixWithStore`)
