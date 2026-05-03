# Clean Wizard - Project Instructions

**Updated:** 2026-05-03

## Build & Test

```bash
go build ./...
go test ./... -short
```

## Project Structure

- `cmd/clean-wizard/` - CLI entry point and commands (Cobra)
- `internal/cleaner/` - 13+ cleaner implementations, registry, factory
- `internal/domain/` - Domain types, 19 type-safe enums, interfaces, settings
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

- **Registry Pattern** - `cleaner.Registry` for thread-safe cleaner management
- **Result Type** - `result.Result[T]` for functional error handling
- **Type-Safe Enums** - 19 iota-based enums with generic helpers in `enum_macros.go`
- **Adapter Pattern** - External tools wrapped in `internal/adapters/`
- **Flow Composition** - `FlowBuilder[T]`, `ParallelFlow[T]`, `BranchFlow[T]`

## Dependencies

- `charm.land/huh/v2` - TUI forms
- `charm.land/lipgloss/v2` - Terminal styling
- `github.com/charmbracelet/fang` - Help command generation
- `github.com/cockroachdb/errors` - Error wrapping
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega` - BDD testing
- `github.com/knadh/koanf/v2` - Configuration
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML handling

## Known Issues

- 4 error packages with overlapping responsibilities (split brain)
- `internal/domain/` is a god package (20+ files)
- `internal/cleaner/` has 50+ files flat (no sub-packages)
- `DefaultSettings(OperationTypeSystemCache)` returns macOS-only cache types
- ~40 `err113` lint violations (dynamic errors via fmt.Errorf)
- 15 source files over 350 lines

## Test Facts

- 298 test functions across 63 test files
- Ginkgo BDD tests exist for: GitHistory, Nix, CompiledBinaries, ProjectExecutables
- 9 of 13 cleaners have NO BDD tests
- CLI command tests are missing entirely
