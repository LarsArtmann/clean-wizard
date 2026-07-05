# Clean Wizard - Project Instructions

**Updated:** 2026-07-06

## Build & Test

```bash
go build ./...
go test ./... -short
```

## Target Machines

- **evo-x2**: NixOS Linux x86_64, Nix 2.34.7, Go 1.26.3, Docker, pnpm, bun, golangci-lint
  - No: cargo, brew, npm, yarn, pip/pip3
  - Major caches: go-build (25GB), goimports (8.2GB), pip (6.5GB), nix (3.1GB), gopls (2.1GB), puppeteer (1.2GB), JetBrains (845MB), pnpm (717MB)
- **macOS**: Primary historical target (Spotlight, Xcode, CocoaPods, Homebrew)

## Project Structure

- `cmd/clean-wizard/` - CLI entry point and commands (Cobra)
- `internal/di/` - Dependency injection container (samber/do v2)
- `internal/execution/` - Workflow orchestration engine (Azure/go-workflow)
- `internal/cleaner/` - 13+ cleaner implementations, registry, factory
- `internal/domain/` - Domain types, 27 CacheType enums, interfaces, settings
- `internal/config/` - Configuration loading (koanf/yaml), validation
- `internal/result/` - Result[T] type for functional error handling
- `internal/adapters/` - External tool adapters (Nix, Exec, HTTP, Cache)
- `internal/middleware/` - Validation middleware
- `internal/conversions/` - Unit conversions
- `internal/format/` - Byte formatting, JSON output
- `tests/bdd/` - Ginkgo-based BDD tests
- `docs/` - Documentation

## Key Files

- `TODO_LIST.md` - Current source of truth for pending work
- `FEATURES.md` - Feature status documentation
- `docs/status/` - Status reports
- `docs/architecture-understanding/` - D2 architecture diagrams

## Architecture Patterns

- **Dependency Injection** - `internal/di/` using samber/do v2; `RegisterAllServices` wires all services into a single container; typed accessors wrap `do.Invoke[T]`
- **Workflow Orchestration** - `internal/execution/` using Azure/go-workflow; `RunCleaners`/`RunScans` compile cleaners into a DAG of `flow.FuncIO` steps with `BeforeStep`/`AfterStep` hooks
- **Registry Pattern** - `cleaner.Registry` for thread-safe cleaner management, resolved from DI container
- **Result Type** - `result.Result[T]` for functional error handling in cleaner methods
- **Type-Safe Enums** - 27 CacheType enums with generic helpers in `enum_macros.go`
- **Adapter Pattern** - External tools wrapped in `internal/adapters/`
- **Platform-Aware Defaults** - `DefaultProtectedPaths()`, `getDefaultSystemCacheTypes()` use `runtime.GOOS`
- **Typed Error Classification** - `cleaner.NotAvailableError` + `cleaner.IsNotAvailableError()` replaces fragile string matching

## DI + Workflow Architecture

The application follows the same pattern as BuildFlow:

1. **CLI parses flags** → loads config → creates DI container per command invocation
2. **`di.RegisterAllServices(injector, cfg, settings)`** registers config, run settings, and cleaner registry as lazy singletons
3. **Command resolves services** from DI via typed accessors (`di.CleanerRegistry(i)`)
4. **`execution.RunCleaners(ctx, registry, names, opts...)`** compiles cleaners into a go-workflow DAG and executes it
5. **Workflow steps** wrap each cleaner's `Clean(ctx)` method with panic recovery, collecting results in a thread-safe `resultCollector`
6. **`WorkflowResult`** aggregates per-step outcomes with succeeded/skipped/failed classification; results sorted by registration order for deterministic output

Key design principles:

- The execution layer is **DI-agnostic** — it receives `*cleaner.Registry` and cleaner names as plain parameters
- Workflow errors are preserved — panics are recovered and recorded as failed steps
- Step results are **deterministically ordered** by registration order regardless of parallel completion

## Execution Layer Capabilities

- **Parallel execution** via go-workflow DAG (`MaxConcurrency` configurable via `RunSettings`)
- **Panic recovery** — `DontPanic: true` + `recover()` in step functions
- **Retry support** — `flow.Retry` with exponential backoff (`cenkalti/backoff/v4`)
- **Step hooks** — `BeforeStep` for timing/logging, `AfterStep` for verbose output
- **Error classification** — `cleaner.IsNotAvailableError()` with typed + string fallback

## Dependencies

- `charm.land/huh/v2` - TUI forms
- `charm.land/lipgloss/v2` - Terminal styling
- `github.com/charmbracelet/fang` - Help command generation
- `github.com/cockroachdb/errors` - Error wrapping
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega` - BDD testing
- `github.com/knadh/koanf/v2` - Configuration
- `github.com/samber/do/v2` - Dependency injection
- `github.com/Azure/go-workflow` - Workflow orchestration engine
- `github.com/cenkalti/backoff/v4` - Exponential backoff for retries
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML handling

## Known Issues

- 4 error packages with overlapping responsibilities (split brain)
- `internal/domain/` is a god package (20+ files)
- `internal/cleaner/` has 50+ files flat (no sub-packages)
- ~40 `err113` lint violations (dynamic errors via fmt.Errorf)
- Cleaners still use hardcoded defaults instead of user profile config

## Test Facts

- 300+ test functions across 63+ test files
- DI package tests: `internal/di/di_test.go` (8 tests)
- Execution package tests: `internal/execution/execution_test.go` + `integration_test.go` (16 tests)
- Ginkgo BDD tests exist for: GitHistory, Nix, CompiledBinaries, ProjectExecutables
- 9 of 13 cleaners have NO BDD tests
- CLI command tests are missing entirely
