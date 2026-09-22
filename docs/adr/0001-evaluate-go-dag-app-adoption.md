# ADR-0001: Evaluate `go-dag-app` Adoption for clean-wizard

**Date:** 2026-09-23
**Status:** Rejected for now (revisit triggers defined)
**Deciders:** Lars Artmann

---

## Context

[`go-dag-app`](https://github.com/LarsArtmann/go-dag-app) is the shared
DI-container bootstrap for LarsArtmann DAG applications (currently BuildFlow
and library-policy). It is deliberately tiny — three exported functions in a
single package (~90 lines):

| Function                        | Role                                                                                                    |
| ------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `IsDebugDIEnabled(appName)`     | Reads `<APP>_DEBUG_DI`; `"1"` or `"true"` (case-insensitive) enables the debug gate                     |
| `PrintDependencyTree(injector)` | Prints the resolved tree to stdout via `do.ExplainInjector`; nil-safe                                   |
| `NewContainer(appName)`         | Fresh `do.Injector` plus a cleanup that prints the tree (if gated on), then calls `injector.Shutdown()` |

It is explicitly **not** a DAG/workflow engine: dependency-graph compilation,
retry, resume, and audit-log diffing stay in the consuming apps.

clean-wizard is a DAG application by the same definition (samber/do v2 +
Azure/go-workflow) and already mirrors BuildFlow's DI patterns
(`internal/di/providers.go:15`). The question: should clean-wizard join the
fleet and adopt `go-dag-app`?

### Current state of clean-wizard's DI

- `di.New() (*Container, func())` — typed wrapper struct + cleanup that
  discards the shutdown error (`internal/di/container.go:27`)
- `Container.Shutdown()` exposes the full `*do.ShutdownReport`
  (`internal/di/container.go:38-40`, exercised in `internal/di/di_test.go:34`)
- Per-command containers at two call sites
  (`cmd/clean-wizard/commands/clean.go:142`, `cmd/clean-wizard/commands/scan.go:90`)
- No debug-DI env gate, no dependency-tree printer — and that is a
  **documented decision**: ROADMAP non-goal #2
  ("`do.ExplainInjector` debug output — YAGNI"), confirmed NOT-DO in three
  archived status/planning reports (2026-07-06 DI-WORKFLOW-MIGRATION,
  PARETO-HARDENING-FINAL, FULL-SESSION-REVIEW)

## Considered Options

1. **Full adoption** — replace `di.New` with `dagapp.NewContainer`
2. **Leaf-only adoption** — delegate only `IsDebugDIEnabled` +
   `PrintDependencyTree` (the BuildFlow/library-policy adoption shape)
3. **No adoption** — keep the local bootstrap (status quo)

## PRO

| Argument                          | Evidence                                                                                                                                                                       |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Zero new dependency tree          | Both pin `samber/do/v2 v2.1.0` exactly; `go-type-to-string` is already an indirect dep in clean-wizard's go.mod                                                                |
| Fleet consistency                 | clean-wizard already mirrors BuildFlow patterns (`internal/di/providers.go:15`); go-dag-app is the canonical bootstrap for this app family                                     |
| Private-dep path already proven   | clean-wizard consumes private `go-error-family` via module proxy + nix `vendorHash` (flake.nix:38) — a second private larsartmann module is the same mechanism                 |
| Inherit bootstrap plumbing fixes  | e.g. go-dag-app TODO #2 (behavioral test for the cleanup debug branch); dagapp's cleanup logs shutdown errors at Debug, ours discards silently (`internal/di/container.go:27`) |
| `NewContainer` API is still fluid | Per go-dag-app AGENTS.md it has no production caller yet — clean-wizard could shape it (e.g. ShutdownReport propagation) now, cheaply                                          |

## CONTRA

| Argument                                         | Evidence                                                                                                                                                                                     |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| The only new capability is a documented non-goal | ROADMAP.md non-goal #2 (YAGNI); adoption silently reverses a deliberate decision                                                                                                             |
| Zero duplication removed                         | clean-wizard has no env gate or tree printer to delegate — nothing gets deleted by adopting                                                                                                  |
| API regression if `NewContainer` adopted         | Its cleanup discards the `*do.ShutdownReport`; our `Container.Shutdown()` exposes it (`internal/di/di_test.go:34`). We would wrap the wrapper                                                |
| Third consumer hardens a deliberately-fluid API  | go-dag-app AGENTS.md: "NewContainer… signature is cheap to change; the two delegated leaf functions are not". Adding a caller now freezes it before its shape is validated                   |
| Fleet coordination cost                          | go-dag-app go.mod pinned to `go 1.26` deliberately while clean-wizard is on `go 1.27`; its TODO #1 (coordinated fleet bump) gains a party; versioning discipline is "manual" (README Status) |
| Marginal value vs. coupling                      | ~90 trivial lines don't justify cross-repo coupling, release tagging (v0.1.x), and hermetic-build surface for a CLI with 2 container call sites                                              |

## Decision

**No adoption (option 3).** The package solves a problem clean-wizard does
not have and explicitly declined (ROADMAP non-goal #2). The mismatch is
strategic, not technical: on every technical axis — same samber/do version,
older `go` directive in the dependency, proven private-module fetch —
adoption would work. It just buys nothing today and costs coupling.

Leaf-only adoption (option 2) is the correct shape **if and when** the debug
tree is ever wanted; full adoption of `NewContainer` (option 1) is
**rejected until** its API preserves shutdown-report visibility.

### Revisit triggers

1. ROADMAP non-goal #2 is flipped and `CLEAN_WIZARD_DEBUG_DI` is wanted for
   debugging the 13-cleaner registry wiring → adopt **only** the two leaf
   delegates (BuildFlow/library-policy shape), skip `NewContainer`.
2. `NewContainer` gains a ShutdownReport-aware signature → clean-wizard is
   its natural first caller; shape the API upstream before consuming.
3. go-dag-app grows real shared DAG features overlapping
   `internal/execution` (retry/resume/graph compilation) → re-evaluate
   broadly.

## Consequences

- clean-wizard keeps its 3-function local bootstrap
  (`internal/di/container.go`); the known warts (silently discarded shutdown
  error in the cleanup closure) are fixable locally in one line if desired.
- go-dag-app retains only two consumers; its `NewContainer` API stays fluid.
- This ADR is the pointer: do not re-litigate adoption without new evidence
  matching one of the revisit triggers.

## References

- go-dag-app: README.md (API, consumers, status), AGENTS.md (consumer
  coupling, API-freeze policy), container.go, TODO_LIST.md
- clean-wizard: `internal/di/container.go`, `internal/di/providers.go`,
  `cmd/clean-wizard/commands/clean.go:142`, `cmd/clean-wizard/commands/scan.go:90`,
  ROADMAP.md (Explicitly NOT Pursuing #2), flake.nix:38
