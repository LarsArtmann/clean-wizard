# CHANGELOG

**Last Updated:** 2026-07-13

---

## [Unreleased]

### Added

#### 2026-07-06

- **DI container** (`samber/do v2`) — dependency injection with lazy singleton registry, typed accessors, and test override helpers (`internal/di/`)
- **Workflow orchestration engine** (`Azure/go-workflow`) — DAG-based parallel execution with panic recovery, step hooks, and deterministic result ordering (`internal/execution/`)
- **Retry support** — `RetryConfig` with exponential backoff (`cenkalti/backoff/v4`), `--retries` flag (default 3), smart retry that stops immediately on non-retryable errors
- **RetryProfile presets** — `--retry-profile` flag (default/aggressive/conservative/none) on both clean and scan commands
- **`--concurrency`/`-C` flag** — max concurrent cleaners, wired to `MaxConcurrency` in `RunSettings`
- **Error classification** (`go-error-family v0.6.1`) — 5-family behavioral classification (Rejection, Conflict, Transient, Corruption, Infrastructure) replacing all hand-rolled keyword matching
- **Per-cleaner error codes** — `NewNotAvailableError` factory derives diagnostic codes (`cleaner.<name>.not_available`)
- **CLI exit codes** — BSD sysexits mapping via `errorfamily.ExitCode()` (Rejection=1, Transient=75, Infrastructure=69, Corruption=65)
- **JSON output enrichment** — `family`/`code`/`retryable` fields in clean JSON output with deterministic alphabetical ordering
- **`*os.PathError` classifier** — permanent errno values (ENOSPC, EROFS, ELOOP) classified as Rejection

#### 2026-04-03

- Enum Consolidation Refactor: Consolidated all 19 iota-based enum types across 4 files onto unified `enum_macros.go` helpers (52% line reduction)
- All enums now use `EnumString`, `EnumIsValid`, `EnumValues`, `EnumMarshalJSON`, `EnumUnmarshalJSON`, `EnumMarshalYAML`, `EnumUnmarshalYAML`
- YAML marshaling now returns strings instead of ints

#### 2026-04-02

- Unit tests for cleanerMetadata (`cleaner_types_test.go` - 4 tests)
- Init() validation for operationTypeToCleanerType entries

### Changed

- Execution model migrated from sequential dispatch to DAG-based parallel workflow engine
- All error classification migrated from hand-rolled keyword matching to `go-error-family` behavioral classification
- `--retries` default changed from 0 (disabled) to 3 (enabled with smart retry)
- Scan command now accepts `--retries`, `--concurrency`, `--retry-profile` (parity with clean)
- Scan `--profile` flag now warns when unsupported instead of silently ignoring
- `ValidationError` now implements `Classified` (→ Rejection) + `Coded` (`validation.rejected`)
- Error messages simplified to consistent format
- Git History dry-run default changed from true to false

### Removed

- `cmd/clean-wizard/commands/cleaner_implementations.go` (357 lines of dual-registry dispatch)
- `internal/pkg/errors/` package (1283 lines — ghost error package replaced by `go-error-family`)
- `internal/result/flow_builder.go`, `branch_flow.go`, `branch_flow_test.go` (~1472 lines — superseded by go-workflow)
- `internal/cleaner/parallel.go` (`ParallelExecutor` — superseded by execution layer)
- `cockroachdb/errors` dependency (fully eliminated from go.mod/go.sum)
- `DefaultRegistry()` function (replaced by `DefaultRegistryWithConfig(verbose, dryRun)`)
- `ErrGoCacheNotAvailable` sentinel (replaced by inline `NewNotAvailableError` factory)
- Dead `UnmarshalYAMLEnum`, `UnmarshalJSONEnum`, `UnmarshalYAMLEnumWithDefault` helpers
- `TypeSafeEnum` interface
- Langversion cleaner stub (CleanerTypeLangVersionMgr)
- 49 deprecation warnings across 45+ files

### Fixed

- Retry duplicate recording — `recordFinal()` replaces `record()` so retried steps produce exactly 1 entry
- Workflow errors no longer silently dropped when steps exist
- Panics in cleaners now recovered and recorded as failed steps
- `isProcessRunning` fails closed when `pgrep` is unavailable (was failing open)
- Results sorted by registration order for deterministic output (was non-deterministic from parallel execution)
- Latent `:=` vs `=` bug in `enum_macros.go:108`
- Docker size reporting (was returning 0)
- Cargo size reporting
- Git History form field overwriting bug
- Git History Scanner: eliminated 40+ tree object warnings, optimized batch API

---

## [0.1.0] - 2026-03-22

### Added

#### Core Infrastructure

- CleanerRegistry Integration (`internal/cleaner/registry.go` - 231 lines, 12 tests)
- Generic Context System (unified ValidationContext, ErrorDetails, SanitizationChange into Context[T])
- Domain Model Enhancement (Config struct has Validate(), Sanitize(), ApplyProfile())
- 13 cleaners implementing Clean(), IsAvailable(), Name()
- 5 CLI commands: clean, scan, init, profile, config

#### Utilities

- Generic Validation Interface (`internal/shared/utils/validation/validation.go`)
- Config Loading Utility (`internal/shared/utils/config/config.go`)
- String Trimming Utility (`internal/shared/utils/strings/trimming.go`)
- Error Details Utility (now replaced by `go-error-family`)
- Schema Min/Max Utility (`internal/shared/utils/schema/minmax.go`)

#### Cleaners

- CompiledBinariesCleaner (576 lines, 918 tests)
- Git History Cleaner with interactive binary cleaning (900+ tests)
- Timeout protection on all exec commands

#### Documentation

- ARCHITECTURE.md
- CLEANER_REGISTRY.md
- ENUM_QUICK_REFERENCE.md

### Changed

- NodePackages refactored to use domain.PackageManagerType
- BuildCache keeps local JVMBuildToolType (JVM-specific)
- Dry-run estimates now use real sizes with fallbacks
- Linux SystemCache support expanded (XdgCache, Thumbnails, Pip, Npm, Yarn, Ccache)

### Removed

- Language Version Manager NO-OP cleaner
- 69 lines of duplicate enum code

### Fixed

- All enum types: RiskLevel, Enabled, DockerPruneMode now have IsValid(), Values(), String()
- Result type enhanced with: Validate, ValidateWithError, AndThen, FlatMap, OrElse, Map, Tap
- Context propagation in error messages
