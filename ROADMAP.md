# ROADMAP

**Last Updated:** 2026-07-13
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

---

## Explicitly NOT Pursuing (Non-Goals)

1. **Application-global DI bootstrap** — per-command DI containers are sufficient for a CLI tool
2. **`do.ExplainInjector` debug output** — YAGNI; debug flags add complexity for minimal value
3. **`do.ShutdownerWithError` on adapters** — no adapters currently hold resources requiring cleanup
4. **Consolidating `cleaner.Cleaner` vs `domain.OperationHandler`** — risky refactor, low value

---

**Note:** Items here are aspirational. No timeline commitments. Refined ideas become
actionable tasks in `TODO_LIST.md`.
