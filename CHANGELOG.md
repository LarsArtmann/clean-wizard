# CHANGELOG

**Last Updated:** 2026-08-10

---

## [Unreleased]

### Added

#### 2026-08-10 (continued)

- **Cargo test expectation aligned** (`internal/cleaner/testhelper_test.go`) — `TestBooleanSettingsCleaners/Cargo` now expects 2 items (registry + git cache subdirs) instead of 1, matching the actual `CargoCleaner.Clean` dry-run contract. Closes TODO #24.
- **TODO_LIST.md refresh** — TODO #24 marked DONE; test count updated to 23/23 packages PASS

#### 2026-08-10

- **Retry test regression fixed** (`internal/execution/integration_test.go`) — `TestRunCleaners_Retry` now uses `SizeEstimate.Value()` instead of deprecated `FreedBytes` field; assertion matches actual cleaner contract
- **`recordFinal` in-place mutation fix** (`internal/execution/results.go`) — `for _, v := range slices.Backward(rc.results)` mutated a copy and silently dropped the assignment; replaced with index-based assignment so retry outcomes correctly overwrite previous attempts (without this, retry-succeeded steps were reported as failed because the original error remained)
- **TODO list refresh** (`TODO_LIST.md`) — verified items against code, added new items harvested from 2026-07-06 → 2026-08-05 sessions
- **FEATURES.md status correction** — Projects Management Automation cleaner now correctly described as `FULLY_FUNCTIONAL` (uses typed `*NotAvailableError` for missing-tool signaling); `docker_parsing.go` H007 violation added as known issue

#### 2026-08-05

- **`go-humanize` adopted for size parsing** (`internal/cleaner/golangcilint.go`) — replaced 11-entry `golangciLintSizeMultiplier` map and 8 byte-conversion constants with `humanize.ParseBytes` (commit `b7692ff`)
- **TestParseSize passes 10/10 cases** — binary (3.1KiB, 1.5MiB, 500B, 1GiB, 1TiB), decimal (1KB, 1MB, 1GB, 1TB), invalid
- **Full cleaner suite passes 243/243 Ginkgo specs** — `go test ./internal/cleaner/ -short` clean

#### 2026-07-15

- **Public website launched** (`https://cleanwizard.lars.software`) — Astro 7 + Starlight + Tailwind v4, deployed to Firebase hosting with SSL cert (Let's Encrypt), CI/CD workflow (`.github/workflows/website.yml`), 13 GitHub topics set
- **Landing page copywriting overhaul** — replaced fabricated hero output with real captured `clean-wizard scan`/`clean --dry-run` output (38 GiB scan, 12 GiB freed); rewrote all 6 feature cards in user-facing language; added ProblemSection with real pain-point data; replaced strawman "SystemNix" comparison with honest "Generic Cleaners"

#### 2026-07-14

- **`encoding/json/v2` migration** — all 7 production source files migrated to `encoding/json/v2` + `encoding/json/jsontext` (`MarshalIndent` → `Marshal` + `jsontext.WithIndent`)
- **`go-error-family` bumped v0.6.1 → v0.7.0** — transitively imports `encoding/json/v2`
- **`GOEXPERIMENT=jsonv2`** set in flake.nix `buildGoModule.env`, both `default` and `ci` devShells
- **GitHub description, homepage URL, 13 topics** set for `LarsArtmann/clean-wizard`

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
- **2026-08-10**: `recordFinal` copy-in-loop bug — was iterating with `for _, v := range` (mutated copy, dropped assignment); now index-based `rc.results[i] = ...`
- **2026-08-10**: `TestRunCleaners_Retry` regression assertion — was asserting deprecated `FreedBytes`; now asserts `SizeEstimate.Value()`
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
