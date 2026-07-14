# Domain Language

A **Unified Language** for Clean Wizard — shared across users, developers, and AI agents.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.

## Glossary

| Term              | Definition                                                                                            | Context                                   |
| ----------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| Cleaner           | A unit that cleans a specific target (Nix store, Homebrew cache, Go build cache, etc.)                | Core abstraction; 13 registered           |
| Registry          | Thread-safe map of cleaner names to instances; resolved from the DI container                         | `internal/cleaner/registry.go`            |
| Workflow          | A DAG of cleaning steps compiled from selected cleaners and executed in parallel by go-workflow       | `internal/execution/workflow.go`          |
| Step              | A single cleaner's execution within a workflow; produces a `StepResult`                               | `internal/execution/results.go`           |
| Scan              | Non-destructive estimation pass: discovers available cleaners and estimates reclaimable space         | `scan` command                            |
| Clean             | Destructive pass: executes selected cleaners to reclaim space                                         | `clean` command                           |
| Dry-Run           | Preview mode: no files are deleted; cleaners report what they _would_ clean                           | `--dry-run` flag                          |
| CacheType         | Enum identifying a specific cache kind (27 values: Spotlight, Xcode, GOCACHE, Pip, NixCache, etc.)    | `internal/domain/operation_settings.go`   |
| OperationType     | Enum identifying a cleaner category (Nix, Homebrew, Docker, Go, Cargo, etc.)                          | `internal/domain/operation_types.go`      |
| OperationSettings | Per-cleaner configuration struct (cache types, age thresholds, keep counts, paths)                    | `internal/domain/operation_settings.go`   |
| RunSettings       | Per-invocation settings (verbose, dry-run, max concurrency) injected via DI                           | `internal/di/options.go`                  |
| RetryProfile      | Preset retry configuration (Default, Aggressive, Conservative, None) controlling backoff and attempts | `internal/execution/retry.go`             |
| ErrorFamily       | Behavioral error classification (Rejection, Conflict, Transient, Corruption, Infrastructure)          | `go-error-family` library                 |
| NotAvailableError | Typed error indicating a cleaner's target tool is not installed; classifies as Infrastructure         | `internal/cleaner/cleaner.go`             |
| ValidationError   | Typed error for invalid configuration or input; classifies as Rejection                               | `internal/domain/operation_validation.go` |

## Entities

Objects with identity and lifecycle.

| Term           | Definition                                                                 | Context                         |
| -------------- | -------------------------------------------------------------------------- | ------------------------------- |
| Registry       | Holds cleaner instances; created per command invocation via DI             | `internal/cleaner/registry.go`  |
| DI Container   | Dependency injection container wrapping `do.Injector`; created per command | `internal/di/container.go`      |
| WorkflowResult | Aggregated outcome of a workflow run: succeeded/skipped/failed steps       | `internal/execution/results.go` |

## Value Objects

Immutable objects defined by attributes.

| Term         | Definition                                                               | Context                         |
| ------------ | ------------------------------------------------------------------------ | ------------------------------- |
| StepResult   | Outcome of a single workflow step (status, freed bytes, error, duration) | `internal/execution/results.go` |
| SizeEstimate | Estimated bytes reclaimable by a cleaner (known/approximate/mocked)      | `internal/domain/types.go`      |
| CleanResult  | Result of a clean operation (freed bytes, items processed, details)      | `internal/domain/types.go`      |
| Result[T]    | Generic Rust-style result type wrapping success value or error           | `internal/result/type.go`       |

## Commands

Actions the system can perform.

| Term       | Definition                                                 | Context              |
| ---------- | ---------------------------------------------------------- | -------------------- |
| clean      | Execute selected cleaners to reclaim disk space            | `clean` command      |
| scan       | Discover available cleaners and estimate reclaimable space | `scan` command       |
| init       | Generate a starter YAML configuration file                 | `init` command       |
| profile    | Manage configuration profiles                              | `profile` command    |
| config     | Validate or display configuration                          | `config` command     |
| githistory | Clean large files from git history using git-filter-repo   | `githistory` command |

## Bounded Contexts

| Context              | Description                                                                      |
| -------------------- | -------------------------------------------------------------------------------- |
| Cleaning             | Core cleaning domain: cleaners, scans, cache types, size estimation              |
| Execution            | Workflow orchestration: DAG compilation, parallel execution, retries, step hooks |
| Config               | YAML configuration loading, validation, sanitization, profile management         |
| Git History          | Git repository history rewriting (separate from cache cleaning)                  |
| Error Classification | Behavioral error families driving retry, skip, and exit code decisions           |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
