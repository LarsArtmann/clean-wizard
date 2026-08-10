# ROADMAP

**Last Updated:** 2026-08-10
**Focus:** Long-term direction and raw ideas — no timeline commitment

---

## Themes

### 1. Configuration-Driven Cleaning

Today cleaners use hardcoded defaults. The vision: every cleaner reads its settings from
YAML profiles, letting users define per-cleaner behavior (cache types, age thresholds,
keep counts) declaratively. This requires registering individual cleaners as DI providers
and wiring `OperationSettings` through to constructors.

### 2. Extensibility

A plugin system would allow third-party cleaners without modifying the core registry.
The current `Cleaner` interface and registry pattern provide the foundation, but the
loading and discovery mechanism does not exist yet.

### 3. Observability

Live progress TUI during workflow execution (per-cleaner status, real-time freed space).
Structured audit log of DI registrations. Resume/checkpoint support for interrupted runs.

### 4. Honest Size Estimation

Nix cleaner still uses a hardcoded 50MB-per-generation estimate. Other cleaners scan
actual sizes from disk. The next step: replace Nix's estimate with real `nix-store
--query` size queries so dry-run output reflects reality.

### 5. Quality Gates

The project has 300+ tests but no single canonical CI pipeline that ties them together.
Goals: `go build` → `go test -short` → `go test -race` → `nix flake check` → linter
must all pass before merge. The `go-humanize-linter` H007 check should ship inside the
repo and run automatically.

---

## Raw Ideas

| Category             | Idea                              | Notes                                                                      |
| -------------------- | --------------------------------- | -------------------------------------------------------------------------- |
| Plugin Architecture  | Plugin system for cleaners        | Third-party cleaners loaded dynamically; not required for v1               |
| Progress TUI         | Live per-cleaner status display   | Like BuildFlow's ProgressBridge; requires workflow engine hooks            |
| Resume Support       | Checkpoint interrupted clean runs | `flow.Workflow` state persistence; low priority                            |
| RiskLevel Automation | Auto mapstructure decode hook     | Investigated: manual mapstructure processing works; auto needs extra hooks |
| Web UI               | Configuration and monitoring UI   | Long-term; CLI-first for now                                               |
| Cloud Integration    | Remote execution / cloud storage  | Very long-term; no current use case                                        |
| Shared Size Parsing  | `conversions.ParseIECBytes(s)`    | Wrap `humanize.ParseBytes` once for all cleaners — eliminates H007 risk   |
| Linter Distribution  | Move `/tmp/go-humanize-linter` into `tools/lint/` | Makes H007 reproducible; wire into flake.nix `checks`              |
| Cleaner Refactor     | Split `internal/cleaner/` by platform domain | Group `nix/`, `docker/`, `golang/`, `macos/`, `linux/` — supports modular loading |

---

## Explicitly NOT Pursuing (Non-Goals)

1. **Application-global DI bootstrap** — per-command DI containers are sufficient for a CLI tool
2. **`do.ExplainInjector` debug output** — YAGNI; debug flags add complexity for minimal value
3. **`do.ShutdownerWithError` on adapters** — no adapters currently hold resources requiring cleanup
4. **Consolidating `cleaner.Cleaner` vs `domain.OperationHandler`** — risky refactor, low value
5. **Re-introducing `cockroachdb/errors`** — fully eliminated in `edaff33`; `go-error-family` is the sole error library

---

**Note:** Items here are aspirational. No timeline commitments. Refined ideas become
actionable tasks in `TODO_LIST.md`.
