# Clean Wizard - Project Instructions

**Last Reviewed:** 2026-08-10 (docs-health audit)

## Build & Test

**Requires `GOEXPERIMENT=jsonv2`** (set in flake.nix devShell/build or `export GOEXPERIMENT=jsonv2` for standalone):

```bash
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./... -short
```

Or use the Nix devShell (`nix develop`) which sets it automatically. The devShell and CI shell install `go_1_27` (matching the `pkgs.go_1_27` used by `buildGoModule`) — go.mod requires Go 1.27, so a host `go` 1.26 binary fails with `GOTOOLCHAIN=local`.

**Website CI gotcha:** pnpm 11.20 enforces a default 24h `minimumReleaseAge` supply-chain check in `pnpm install`. Dependency bumps whose regenerated lockfile pulls freshly-published transitive deps (e.g. rolldown for astro) fail CI with `ERR_PNPM_MINIMUM_RELEASE_AGE_VIOLATION` — not a bug; re-run the jobs ~24h later. Related: build-script approvals live in `website/pnpm-workspace.yaml` under `allowBuilds:` (`esbuild: true`) — pnpm v11 ignores `pnpm.*` in `package.json`; a missing or placeholder entry makes `astro build` fail on a missing esbuild binary (cmdguard incident, fixed 2026-09-19).

**flake check gotchas (pre-existing, ticketed TODO #29):** `nix flake check` fails on (a) `treefmt-check`: the sandboxed goimports formatter tries to download the go1.27 toolchain and the offline sandbox refuses DNS — `nix fmt` itself is green; (b) cold-cache `go-modules` builds need network. Formatting ground truth: run `nix fmt` (0 changed = clean) or `buildflow format` inside the devShell.

## Target Machines

- **evo-x2**: NixOS Linux x86_64, Nix 2.34.7, Go 1.26.3, Docker, pnpm, bun, golangci-lint
  - No: cargo, brew, npm, yarn, pip/pip3
  - Major caches: go-build (25GB), goimports (8.2GB), pip (6.5GB), nix (3.1GB), gopls (2.1GB), puppeteer (1.2GB), JetBrains (845MB), pnpm (717MB)
- **macOS**: Primary historical target (Spotlight, Xcode, CocoaPods, Homebrew)

## Project Structure

- `cmd/clean-wizard/` - CLI entry point and commands (Cobra)
- `internal/di/` - Dependency injection container (samber/do v2): per-cleaner providers, adapter aliases, accessors
- `internal/execution/` - Workflow orchestration engine (Azure/go-workflow)
- `internal/cleaner/` - Shared cleaner core (interfaces, registry, helpers, error classification, test factories) plus one sub-package per cleaner (`nix/`, `homebrew/`, `docker/`, `cargo/`, `golang/`, `golangcilint/`, `nodepackages/`, `buildcache/`, `systemcache/`, `tempfiles/`, `projectsmanagementautomation/`, `projectexecutables/`, `compiledbinaries/`, `githistory/`) and `factory/` (per-cleaner constructors + registry assembly)
- `internal/domain/` - Split into `enums/` (type-safe enums + generic marshal helpers), `operations/` (OperationType/OperationSettings/validation/defaults, duration parsing, git-history settings), `types/` (scan/clean results, Config/Profile, cleaner ports, system paths); dependency DAG: enums ← operations ← types
- `internal/config/` - Configuration loading (koanf/yaml), validation
- `internal/result/` - Result[T] type for functional error handling
- `internal/adapters/` - External tool adapters (Nix, Exec, HTTP, Cache)
- `internal/middleware/` - Validation middleware
- `internal/conversions/` - Unit conversions
- `internal/format/` - Byte formatting, JSON output, SARIF output (`sarif.go` converts scan outcomes to findings via go-finding: bytes>0 → info/`unused` finding with clean suggestion, failed scan → error finding with family/code/retryable metadata, unavailable → no finding; locations are `cleaner://<name>` URIs)
- `tests/bdd/` - Ginkgo-based BDD tests
- `docs/` - Documentation

## Key Files

- `TODO_LIST.md` - Current source of truth for pending work
- `FEATURES.md` - Feature status documentation
- `docs/status/` - Status reports
- `docs/architecture-understanding/` - D2 architecture diagrams

## Architecture Patterns

- **Dependency Injection** - `internal/di/` using samber/do v2; `RegisterAllServices` wires config, run settings, adapters, and cleaners; typed accessors wrap `do.Invoke[T]`; per-cleaner named services (`cleaner.<name>`) resolved via `di.Cleaner(i, name)`
- **Workflow Orchestration** - `internal/execution/` using Azure/go-workflow; `RunCleaners`/`RunScans` compile cleaners into a DAG of `flow.FuncIO` steps with `BeforeStep`/`AfterStep` hooks
- **Registry Pattern** - `cleaner.Registry` for thread-safe cleaner management; the DI registry provider aggregates all per-cleaner services in canonical order
- **Result Type** - `result.Result[T]` for functional error handling in cleaner methods
- **Type-Safe Enums** - 27 CacheType enums with generic helpers in `internal/domain/enums/enum_macros.go`
- **Adapter Pattern** - External tools in `internal/adapters/` behind interfaces (`NixStore`, `HTTPRequester`, `Limiter`, `KeyValueCache`); `AdaptersPackage` registers concrete adapters and aliases them to interfaces via `do.MustAs`; the nix cleaner consumes `adapters.NixStore`, not `*NixAdapter`
- **Platform-Aware Defaults** - `DefaultProtectedPaths()`, `getDefaultSystemCacheTypes()` use `runtime.GOOS`
- **ValidateOptionalSettings helper** - Generic helper in `internal/cleaner/helpers.go` that consolidates the `if settings == nil || settings.X == nil { return nil }` boilerplate shared by every cleaner's `ValidateSettings` method
- **CleanerConstructor[T] generic** - `internal/cleaner/test_interfaces.go` defines `type CleanerConstructor[T any] func(verbose, dryRun bool) T` used as alias for `CleanerConstructorWithSettings` and `SimpleCleanerConstructor`
- **CleanerCore base interface** - Minimum cleaner interface (`IsAvailable` + `Clean`) shared by `CleanerWithSettings` and `SimpleCleaner` to avoid duplicate interface declarations
- **Error Classification** — `go-error-family` (`github.com/larsartmann/go-error-family`) is the sole error library (cockroachdb/errors fully removed). All errors classify into 5 families (Rejection, Conflict, Transient, Corruption, Infrastructure). `NotAvailableError` implements `Classified` + `Coded` with per-cleaner codes (`cleaner.<name>.not_available`) via the `NewNotAvailableError` factory. `domain.ValidationError` implements `Classified` (→ Rejection). `errorfamily.IsRetryable()` drives retry decisions; `errorfamily.Classify()` drives skip/failed classification; `errorfamily.ExitCode()` drives sysexits exit codes at the CLI boundary.

## DI + Workflow Architecture

The application follows the same pattern as BuildFlow:

1. **CLI parses flags** → loads config → creates DI container per command invocation
2. **`di.RegisterAllServices(injector, cfg, settings)`** registers config, run settings, adapter services (aliased to interfaces via `do.MustAs`), one named provider per cleaner (`cleaner.<name>`), and the registry that aggregates them
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
- **Error classification** — `errorfamily.Classify(err)` returns `Infrastructure` for unavailable cleaners (skipped), `Transient` for retryable failures (retried with backoff), `Rejection`/`Conflict`/`Corruption` for permanent failures. No keyword matching.
- **Retry by default** — `--retries 3` on both clean and scan commands; `errorfamily.IsRetryable()` returns false for non-Transient errors → `backoff.Stop` (zero delay); `--retries 0` disables
- **RetryProfile presets** — `--retry-profile` flag (default/aggressive/conservative/none) on both clean and scan; overrides `--retries` with pre-tuned backoff/attempt combinations
- **CLI exit codes** — `errorfamily.ExitCode(err)` in `main.go` maps error families to BSD sysexits codes (Rejection=1, Transient=75, Infrastructure=69, Corruption=65); `errorfamily.LogError` adds structured slog output with family/code/retryable fields
- **Profile settings wiring** — `--profile` flag → `di.RunSettings.Profile` → `domain.Config.SettingsForProfile(name)` merges each operation's settings block into one `OperationSettings` (first section wins) → `cleaner.DefaultRegistryWithConfig(verbose, dryRun, settings)`. Resolution helpers in `internal/cleaner/registry_settings.go` translate sections to constructor params: a configured section overrides factory defaults field-by-field (empty string/0/nil field → default); fully-disabled go_packages falls back to default cache flags (`GoCacheNone` is an invalid constructor state). Every registered `CleanerWithSettings` is also validated against the settings at registry creation (Rejection on failure)
- **koanf array-path gotcha** — koanf does NOT flatten into slices, so `k.Get("profiles.x.operations.0.settings")` always returns nil. All per-operation reads go through `operationRawValue` (manual map/slice navigation) in `internal/config/config.go`; settings are then re-encoded as YAML and decoded through domain types so enum `UnmarshalYAML` hooks handle both int and string forms

## Config Validation Engine (go-business-rules, adopted 2026-09-23)

- **Declarative rules, not imperative appends** — every config validation check (structure / field / cross-field / business logic / security) is a `businessrules.Rule` built in `internal/config/validation_rules.go` and executed via `businessrules.NewValidator().AddRules(...).Build()`. `ConfigValidator.ValidateConfig` bridges the outcome back into the project's `ValidationResult` (Critical+Error → `Errors`, Warning+Info → `Warnings`); JSON shape is unchanged
- **4 severity levels** — `operations.ValidationSeverity` now has `SeverityCritical` (`"critical"`); a `..` parent reference in a protected path escalates to critical. Rule tags carry `[level, kind]` (kind = old rule strings like `required`/`range`/`security`); `WithDescription` carries the fix suggestion; per-rule structured data (value, `ValidationContext`) is restored from maps keyed by rule name because `ViolationError` only keeps the check error as text
- **Warnings surface at the load boundary** — `validateLoadedConfig` logs every validation warning (`logger.Warn`, field/message/suggestion); previously warnings were computed and silently dropped. Error suggestions are logged too
- **Deterministic violation order** — profile iteration is sorted; rules execute sequentially in registration order (do not switch to `Stream`, it reorders violations)
- **MinProtectedPaths is now enforced** (declared-but-dead before); structure-level, guarded to not double-fire on empty paths
- Removed with the rewrite: `validator_structure.go`, `validator_crossfield.go`, `validateBusinessLogic`/`validateSecurityConstraints`/`addCriticalRiskError`; conflict-check predicates (`validateProtectedPathsConflict`, `checkTempFilesConflict`, `checkNixConflict`, `findMaxRiskLevel`) remain as rule check functions in `validator_business.go`
- Assessment + non-adopted capabilities (Stream, otel listener, composites, domain-layer rewrite): `docs/planning/2026-09-23_go-business-rules-integration.md`

## Dependencies

- `charm.land/huh/v2` - TUI forms
- `charm.land/lipgloss/v2` - Terminal styling
- `github.com/charmbracelet/fang` - Help command generation
- `github.com/larsartmann/go-error-family` - Error classification (5 families: Rejection, Conflict, Transient, Corruption, Infrastructure)
- `github.com/larsartmann/go-finding` - SARIF 2.1.0 export for scan output (Finding model, Builder API; core module only, no pipeline)
- `github.com/LarsArtmann/go-business-rules/v2` - Severity-aware validation (Info/Warning/Error/Critical): config validation rules, validator, violation bridging
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega` - BDD testing
- `github.com/knadh/koanf/v2` - Configuration
- `github.com/samber/do/v2` - Dependency injection
- `github.com/Azure/go-workflow` - Workflow orchestration engine
- `github.com/cenkalti/backoff/v4` - Exponential backoff for retries
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML handling
- `encoding/json/v2` + `encoding/json/jsontext` - JSON marshaling (requires `GOEXPERIMENT=jsonv2`)

## Error Handling Architecture

All errors use `github.com/larsartmann/go-error-family` for behavioral classification:

| Family             | Retryable | Usage in clean-wizard                                          | Exit Code |
| ------------------ | --------- | -------------------------------------------------------------- | --------- |
| **Infrastructure** | no        | `NotAvailableError` (binary not installed), `exec.ErrNotFound` | 69        |
| **Transient**      | yes       | Exec failures, timeouts, I/O errors (default for unknown)      | 75        |
| **Rejection**      | no        | Bad config, invalid input, missing cache type                  | 1         |
| **Conflict**       | no        | `ErrGoProcessesRunning` (state conflict)                       | 1         |
| **Corruption**     | no        | Nix store corruption (not yet wired)                           | 65        |

Key files:

- `internal/cleaner/cleaner.go` — `NotAvailableError` implements `Classified` + `Coded` (Infrastructure); `NewNotAvailableError(name, reason)` factory derives per-cleaner error codes
- `internal/cleaner/error_classification.go` — `init()` registers `exec.ErrNotFound`, stdlib defaults, cleaner sentinels, `*exec.ExitError`→Transient, `*os.PathError` permanent errno→Rejection (ENOSPC/EROFS/ELOOP), and user-facing message templates
- `internal/domain/operation_validation.go` — `ValidationError` implements `Classified` (Rejection) + `Coded` (`validation.rejected`)
- `internal/execution/retry.go` — `RetryConfig`, `RetryConfigFromAttempts(n)`, `RetryProfile` type (Default/Aggressive/Conservative/None); `NextBackOff` hook uses `errorfamily.IsRetryable()`
- `internal/execution/results.go` — `StepResult.Status()` uses `errorfamily.Classify()` → Infrastructure=skipped, else=failed
- `internal/format/json.go` — JSON output includes `family`/`code`/`retryable` fields for skipped/failed cleaners; cleaners sorted by name for deterministic output
- `cmd/clean-wizard/main.go` — `errorfamily.ExitCode(err)` + `errorfamily.LogError(err, slog.Default())` at CLI boundary

**Bridge not adopted**: `go-error-family/bridge` connects `samber/oops` to `go-error-family`. Clean-wizard doesn't use oops; core `errorfamily` provides `.WithContext()` for structured context. BuildFlow also implements `Classified` directly without the bridge.

### CLI Error Classification Convention (all 7 command files migrated 2026-09-23)

- **Classify at the site that knows the nature**: wrap unclassified causes (huh form errors, config I/O, marshal failures) with `errorfamily.Wrap{Family}[f](err, "<command>.<where>", msg)`. Codes follow `<command>.<where>` (e.g. `init.config_save`, `scan.json_output`). Config load/save failures are Rejection (matches the `clean.config_load` precedent).
- **Context-only wraps stay plain `fmt.Errorf`**: `Registry.Classify` walks the unwrap chain via `errors.AsType[Classified]`, so `fmt.Errorf("ctx: %w", someClassifiedErr)` inherits the sentinel's family. Wrapping those with an explicit family would _override_ the true classification — don't.
- **Sentinels are classified at the source**: `ErrGitNotAvailable` = Infrastructure, `ErrSafetyChecksFailed` = Conflict, `ErrNoGitRepositoriesFound`/`ErrNotAGitRepository`/`ErrProfileNotFound`/`ErrProfileNoCleaners` = Rejection; dynamic detail wraps them with `%w`.
- **Usage errors are Rejection (exit 1)**: root command sets `SetFlagErrorFunc` (code `cli.flag`), root `RunE` rejects unknown commands (`cli.unknown_command`) while bare invocation prints help, and `exactArgsClassified(n)` (code `cli.args`) wraps positional-args validators. Without this, cobra errors hit the boundary unclassified → Transient → exit 75. Tradeoff: fang's "Try --help" hint disappears for wrapped usage errors (its `HasPrefix` check sees the classification prefix).
- **Scan JSON** mirrors clean JSON enrichment: per-result `error`/`family`/`code`/`retryable` populated from workflow step errors via `ScanResult.Err`; `outputScanJSON` returns a classified `Corruption` error on marshal failure instead of printing `{"error": ...}`. A missing config file falls back to defaults by design (scan/clean exit 0).
- **fang owns human error rendering** (`SilenceErrors=true` + styled handler). Do NOT wire `errorfamily.HandleError` in main.go — it would duplicate fang's output. `LogError` + `ExitCode` remain the CLI boundary.

## Known Issues

- ~~`internal/domain/` is a god package (23 files)~~ RESOLVED 2026-09-23: split into `enums/`, `operations/`, `types/` (task 27)
- ~~`internal/cleaner/` has 50+ files flat (no sub-packages)~~ RESOLVED 2026-09-23: split into 14 per-domain sub-packages + factory (task 28)
- Settings only flow when a `--profile` is selected; preset/interactive runs (and the nix/cargo/projects/git-history/golangci-lint cleaners whose constructors take no settings params) still use factory defaults. `NixGenerationsSettings.DryRun`/`Optimize` and `BuildCacheSettings.ToolTypes` have no constructor consumption yet
- Logger uses mutable package-level globals (`L`, `StdLogger`) — causes test race conditions

## Test Facts

- 300+ test functions across 65+ test files
- DI package tests: `internal/di/di_test.go` (9 tests)
- Execution package tests: `execution_test.go` + `integration_test.go` + `retry_profile_test.go` (19 tests, including smart retry tests, RetryProfile tests, errorfamilytest.AssertFamily assertions)
- Cleaner classification tests: `internal/cleaner/error_classification_test.go` (PathError classification matrix, NotAvailableError per-cleaner codes, exec.ErrNotFound)
- Domain classification test: `internal/domain/operation_validation_test.go` (ValidationError → Rejection)
- JSON output tests: `internal/format/json_test.go` (family/code fields, deterministic ordering)
- CLI integration test: `cmd/clean-wizard/commands/clean_integration_test.go` (dry-run JSON pipeline)
- Integration tests use `testing.Short()` skip guards for real-system tests
- Ginkgo BDD tests exist for: GitHistory, Nix, CompiledBinaries, ProjectExecutables
- 9 of 13 cleaners have NO BDD tests
