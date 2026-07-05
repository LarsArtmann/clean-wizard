# Clean Wizard - Project Instructions

**Updated:** 2026-07-05

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
- `internal/result/` - Result[T], FlowBuilder[T], ParallelFlow[T], BranchFlow[T]
- `internal/adapters/` - External tool adapters (Nix, Exec, HTTP, Cache)
- `internal/middleware/` - Validation middleware
- `internal/conversions/` - Unit conversions
- `internal/format/` - Byte formatting
- `tests/bdd/` - Ginkgo-based BDD tests
- `docs/` - Documentation

## Key Files

- `TODO_LIST.md` - Current source of truth for pending work (30 items)
- `FEATURES.md` - Feature status documentation
- `BDD_TESTS_REVIEW.md` - BDD test coverage analysis
- `docs/status/2026-05-03_08-50-CODE-QUALITY-SCAN.md` - Latest quality scan
- `docs/status/2026-05-03_08-50-ARCHITECTURE-REVIEW.md` - Latest architecture review
- `docs/architecture-understanding/` - D2 architecture diagrams

## Architecture Patterns

- **Dependency Injection** - `internal/di/` using samber/do v2; `RegisterAllServices` wires all services into a single container; typed accessors wrap `do.Invoke[T]`
- **Workflow Orchestration** - `internal/execution/` using Azure/go-workflow; `RunCleaners`/`RunScans` compile cleaners into a DAG of `flow.FuncIO` steps with `BeforeStep`/`AfterStep` hooks
- **Registry Pattern** - `cleaner.Registry` for thread-safe cleaner management, resolved from DI container
- **Result Type** - `result.Result[T]` for functional error handling in cleaner methods
- **Type-Safe Enums** - 27 CacheType enums with generic helpers in `enum_macros.go`
- **Adapter Pattern** - External tools wrapped in `internal/adapters/`
- **Platform-Aware Defaults** - `DefaultProtectedPaths()`, `getDefaultSystemCacheTypes()` use `runtime.GOOS`

## DI + Workflow Architecture

The application follows the same pattern as BuildFlow:

1. **CLI parses flags** → loads config → creates DI container per command invocation
2. **`di.RegisterAllServices(injector, cfg, settings)`** registers config, run settings, and cleaner registry as lazy singletons
3. **Command resolves services** from DI via typed accessors (`di.CleanerRegistry(i)`)
4. **`execution.RunCleaners(ctx, registry, names, opts...)`** compiles cleaners into a go-workflow DAG and executes it
5. **Workflow steps** wrap each cleaner's `Clean(ctx)` method, collecting results in a thread-safe `resultCollector`
6. **`WorkflowResult`** aggregates per-step outcomes with succeeded/skipped/failed classification

Key design principle: the execution layer is **DI-agnostic** — it receives `*cleaner.Registry` and cleaner names as plain parameters. The DI container provides the service graph; the CLI extracts services and hands them to the workflow engine.

## Dependencies

- `charm.land/huh/v2` - TUI forms
- `charm.land/lipgloss/v2` - Terminal styling
- `github.com/charmbracelet/fang` - Help command generation
- `github.com/cockroachdb/errors` - Error wrapping
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega` - BDD testing
- `github.com/knadh/koanf/v2` - Configuration
- `github.com/samber/do/v2` - Dependency injection
- `github.com/Azure/go-workflow` - Workflow orchestration engine
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML handling

## Known Issues

- 4 error packages with overlapping responsibilities (split brain)
- `internal/domain/` is a god package (20+ files)
- `internal/cleaner/` has 50+ files flat (no sub-packages)
- ~40 `err113` lint violations (dynamic errors via fmt.Errorf)
- 15 source files over 350 lines
- `result.FlowBuilder`/`BranchFlow`/`ParallelFlow` are dormant (replaced by go-workflow but not yet removed)

## Test Facts

- 298+ test functions across 63+ test files
- DI package tests: `internal/di/di_test.go`
- Execution package tests: `internal/execution/execution_test.go`
- Ginkgo BDD tests exist for: GitHistory, Nix, CompiledBinaries, ProjectExecutables
- 9 of 13 cleaners have NO BDD tests
- CLI command tests are missing entirely
