# Modularization Proposal — Clean Wizard

**Date:** 2026-05-14 | **Status:** Draft | **Module:** `github.com/LarsArtmann/clean-wizard`

---

## 1. Executive Summary

Clean Wizard is a Go CLI tool for system cleanup with 29 packages in a single module. The codebase has three critical structural problems:

1. **God package** `internal/cleaner` — 60 files, ~16,000 lines, 14 concrete cleaners all in one flat package
2. **God package** `internal/domain` — 22 files, ~2,400 lines, mixing core types with cleaner-specific types
3. **Split brain** — 3 error packages (`internal/errors`, `internal/pkg/errors`, `pkg/errors`) with overlapping responsibilities

This proposal splits the monolith into **5 sub-modules** coordinated via `go.work`, with package-level decomposition of the two god-packages. The result: each cleaner becomes independently testable, the core domain is a focused leaf module with zero external dependencies, and the CLI application wires everything together at the top.

### Expected Benefits

- **Compile-time enforced boundaries** between core domain, cleaner implementations, and CLI
- **Independent testing** — cleaner modules can run tests without pulling in TUI dependencies
- **Faster CI** — only changed modules need rebuilding
- **Clearer ownership** — each module has a single, well-defined purpose
- **Dependency isolation** — TUI libs (huh, lipgloss, fang) isolated to CLI module only

---

## 2. Current State Analysis

### Starting State: Monolith

| Metric                 | Value            |
| ---------------------- | ---------------- |
| `go.mod` files         | 1                |
| `go.work`              | None             |
| Packages               | 29               |
| Non-test lines         | ~31,000          |
| External deps (direct) | 24               |
| Circular dependencies  | None (valid DAG) |

### Current Dependency Layers

```
Layer 0 (Leaf):     domain, result, errors*, version, testing, shared/utils/*
Layer 1:            format, logger, shared/utils/validation, shared/utils/config
Layer 2:            conversions, middleware, config, api
Layer 3:            adapters
Layer 4:            cleaner, testhelper
Layer 5 (App):      cmd/clean-wizard/commands, cmd/clean-wizard
Layer 6 (Tests):    tests/bdd, tests/benchmark, test
```

### Coupling Hotspots

1. **`internal/domain`** — every package imports it; contains both shared types and cleaner-specific types
2. **`internal/cleaner`** — 60 files flat, 12 files over 300 lines, mixes production code with test helpers
3. **3 error packages** — overlapping constructors and types
4. **Test deps in production code** — `stretchr/testify`, `onsi/ginkgo`, `onsi/gomega` imported in non-`_test.go` files
5. **`cleaner` → `adapters` coupling** — adapters change forces all cleaners to recompile

See [DEPENDENCY_GRAPH.md](./DEPENDENCY_GRAPH.md) for full analysis.

---

## 3. Proposed Module Structure

### Module Map

| Module       | Path                 | Purpose                                                        | External Deps                                                     |
| ------------ | -------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------- |
| **core**     | `/core`              | Domain types, interfaces, enums, result type, shared utilities | `gopkg.in/yaml.v3` only                                           |
| **cleaners** | `/cleaners`          | All 14 cleaner implementations, registry, factory, metrics     | `cockroachdb/errors`, `golang.org/x/sys`                          |
| **adapters** | `/adapters`          | External tool wrappers (Nix, Exec, HTTP, Cache, Environment)   | `caarlos0/env`, `go-resty`, `maypok86/otter`, `golang.org/x/time` |
| **config**   | `/config`            | Configuration loading, validation, sanitization, profiles      | `knadh/koanf/v2`, `knadh/koanf/parsers/yaml`                      |
| **app**      | `/app` (root module) | CLI commands, TUI forms, formatting, wiring                    | `charm.land/*`, `spf13/cobra`                                     |

### Module Dependency DAG

```
app  ──→  config
 │        │
 │        ↓
 ├─────→ cleaners ──→ adapters
 │        │            │
 │        ↓            ↓
 └─────→  core  ←─────┘
```

**Key invariant:** `core` has zero internal dependencies. Everything depends on `core`. Nothing depends on `app`.

---

## 4. Module Definitions

### 4.1 Core Module — `/core`

**Module path:** `github.com/LarsArtmann/clean-wizard/core`

| Field               | Content                                                      |
| ------------------- | ------------------------------------------------------------ |
| Purpose             | Domain types, interfaces, enums, Result[T], shared utilities |
| Dependencies (prod) | None (zero internal deps)                                    |
| Dependencies (test) | `stretchr/testify` (for enum/type tests)                     |
| External deps       | `gopkg.in/yaml.v3`                                           |

**Package structure:**

```
core/
├── domain/                  # Pure domain types and interfaces
│   ├── types.go             # CleanResult, ScanResult, ScanItem, CleanRequest, ScanRequest
│   ├── interfaces.go        # OperationHandler, Scanner, PackageCleaner, GenerationCleaner
│   ├── enums.go             # All 19 type-safe enums (consolidated)
│   ├── enum_macros.go       # Generic enum helpers
│   ├── config.go            # Config, Profile types
│   ├── operation_types.go   # OperationType, OperationSettings
│   ├── operation_defaults.go # Default settings per operation
│   ├── operation_validation.go # Shared validation logic
│   ├── duration_parser.go   # Custom duration parsing
│   ├── system_paths.go      # Protected system paths
│   ├── sanitization_types.go # Sanitization types
│   └── githistory_types.go  # GitHistory-specific types (shared with git-history cleaner)
├── result/                  # Result[T], FlowBuilder[T], ParallelFlow[T], BranchFlow[T]
├── errors/                  # Unified error package (consolidate 3→1)
│   ├── types.go             # CleanWizardError, ErrorCode, ErrorLevel
│   ├── constructors.go      # CleaningError, ConfigLoadError, ValidationError, etc.
│   ├── builder.go           # ErrorDetailsBuilder
│   └── handlers.go          # HandleCommandError, HandleConfigError, etc.
├── validation/              # Shared validation utilities
├── format/                  # Byte formatting, date formatting (used by cleaners + CLI)
└── testing/                 # TestCase[T], RunTestCases (test helpers)
```

**Key decisions:**

1. **Nix-specific types stay in core** — `NixGeneration`, `NixGenerationsSettings` are used by both domain interfaces (`GenerationCleaner`) and the Nix adapter/cleaner. They belong in core.
2. **Git-history types stay in core** — `GitHistoryFile`, `GitHistorySettings` are referenced in domain types and config.
3. **Docker-specific types** — `DockerSettings`, `DockerPruneMode` stay in core because config references them directly.
4. **Error consolidation** — Merge all 3 error packages into `core/errors`. The structured `CleanWizardError` from `internal/pkg/errors` becomes the canonical type.

### 4.2 Cleaners Module — `/cleaners`

**Module path:** `github.com/LarsArtmann/clean-wizard/cleaners`

| Field               | Content                                                                     |
| ------------------- | --------------------------------------------------------------------------- |
| Purpose             | All cleaner implementations, registry, factory, parallel execution, metrics |
| Dependencies (prod) | `core`, `adapters`                                                          |
| Dependencies (test) | `core` (via test helpers)                                                   |
| External deps       | `cockroachdb/errors`, `golang.org/x/sys/unix`                               |

**Package structure:**

```
cleaners/
├── cleaner.go               # Cleaner interface, CleanerBase, AgeBasedCleaner
├── registry.go              # Registry (thread-safe cleaner map)
├── factory.go               # DefaultRegistry, DefaultRegistryWithConfig
├── helpers.go               # Generic clean/scan iterators, TrashPath
├── fsutil.go                # Filesystem utilities (dir size, scan, disk usage)
├── validate.go              # Settings validation helpers
├── conversions.go           # Result conversion functions (moved from internal/conversions)
├── metrics.go               # MetricsCollector, TrackedCleaner, MetricsEnabledRegistry
├── parallel.go              # ParallelExecutor
├── nix.go                   # NixCleaner
├── homebrew.go              # HomebrewCleaner
├── docker.go                # DockerCleaner
├── docker_parsing.go        # Docker output parsing
├── cargo.go                 # CargoCleaner
├── golang.go                # GoCleaner (orchestrator)
├── golang_cache.go          # GoCacheCleaner (per-cache-type)
├── golang_scanner.go        # Go cache scanner
├── golang_helpers.go        # Go env/path helpers
├── golang_types.go          # GoCacheType bitflags
├── nodepackages.go          # NodePackageManagerCleaner
├── buildcache.go            # BuildCacheCleaner
├── systemcache.go           # SystemCacheCleaner
├── tempfiles.go             # TempFilesCleaner
├── golangcilint.go          # GolangciLintCacheCleaner
├── projectexecutables.go    # ProjectExecutablesCleaner
├── compiledbinaries.go      # CompiledBinariesCleaner
├── projectsmanagement.go    # ProjectsManagementAutomationCleaner
├── githistory.go            # GitHistoryCleaner
├── githistory_scanner.go    # Git history binary scanner
├── githistory_executor.go   # git-filter-repo execution
├── githistory_filterrepo.go # FilterRepo provider detection
├── githistory_safety.go     # Safety checks
├── testhelpers/             # Test factories, assertions, Ginkgo helpers
│   ├── interfaces.go
│   ├── factories.go
│   ├── assertions.go
│   ├── testing_helpers.go
│   └── ginkgo_helpers.go
└── internal/                # Internal shared utilities (not exported)
    └── testutil/            # Inlined test utilities
```

**Key decisions:**

1. **Not splitting into per-cleaner sub-packages** — The 14 cleaners share `CleanerBase`, `helpers.go`, `fsutil.go`, and the registry. Per-cleaner packages would create import overhead without clear benefit. Flat package with clear file naming is sufficient.
2. **Test helpers in `testhelpers/` sub-package** — Moves non-`_test.go` helper files out of the production package, eliminating `testify`/`ginkgo` production imports.
3. **`Cleaner` interface lives in cleaners module** — It's cleaner-specific, not a domain primitive. `domain.OperationHandler` is the domain-level interface; `Cleaner` is the implementation contract.

### 4.3 Adapters Module — `/adapters`

**Module path:** `github.com/LarsArtmann/clean-wizard/adapters`

| Field               | Content                                                                               |
| ------------------- | ------------------------------------------------------------------------------------- |
| Purpose             | External tool wrappers (Nix, Exec, HTTP, Cache, Environment config)                   |
| Dependencies (prod) | `core`                                                                                |
| Dependencies (test) | `core`                                                                                |
| External deps       | `caarlos0/env/v6`, `go-resty/resty/v2`, `maypok86/otter/v2`, `golang.org/x/time/rate` |

**Package structure:**

```
adapters/
├── nix.go           # NixAdapter — generation listing, store sizing, GC
├── exec.go          # ExecWithTimeout, ExecWithDefaultTimeout
├── http_client.go   # HTTPClient with retry, auth, rate limiting
├── cache_manager.go # CacheManager (Otter-based)
├── environment.go   # EnvironmentConfig loading and validation
├── rate_limiter.go  # RateLimiter
└── errors.go        # Adapter-specific error constructors
```

**Key decisions:**

1. **Adapters depend only on core** — No dependency on cleaners or config. Cleaners import adapters, never the reverse.
2. **Error constructors stay** — `ErrCacheMiss`, `ErrNotFound`, etc. are adapter-specific and don't duplicate core errors.

### 4.4 Config Module — `/config`

**Module path:** `github.com/LarsArtmann/clean-wizard/config`

| Field               | Content                                                                       |
| ------------------- | ----------------------------------------------------------------------------- |
| Purpose             | YAML config loading, validation, sanitization, profile management, middleware |
| Dependencies (prod) | `core`                                                                        |
| Dependencies (test) | `core`                                                                        |
| External deps       | `knadh/koanf/v2`, `knadh/koanf/parsers/yaml`, `knadh/koanf/providers/file`    |

**Package structure:**

```
config/
├── config.go               # Load, Save, GetDefaultConfig
├── enhanced_loader.go      # EnhancedConfigLoader
├── enhanced_loader_api.go  # Public API methods
├── enhanced_loader_types.go
├── enhanced_loader_private.go
├── enhanced_loader_cache.go
├── enhanced_loader_defaults.go
├── safe.go                 # SafeConfig, SafeProfile builders
├── sanitizer.go            # ConfigSanitizer
├── sanitizer_*.go          # Per-operation sanitizers
├── validator.go            # ConfigValidator
├── validator_*.go          # Per-domain validators
├── validation_middleware.go # ValidationMiddleware
├── validation_middleware_*.go
├── risk_util.go
├── errors.go               # Config-specific errors
└── bdd_*.go                # BDD test helpers
```

**Key decisions:**

1. **Config depends only on core** — No dependency on cleaners or adapters. Validation uses domain types only.
2. **Logger dependency removed** — `internal/logger` currently wraps `charm.land/log/v2`. Config uses it minimally. Replace with `log/slog` from stdlib (already available in Go 1.26).

### 4.5 App Module — `/` (root)

**Module path:** `github.com/LarsArtmann/clean-wizard` (root)

| Field               | Content                                                                                                            |
| ------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Purpose             | CLI entry point, Cobra commands, TUI forms, output formatting, wiring                                              |
| Dependencies (prod) | `core`, `cleaners`, `adapters`, `config`                                                                           |
| Dependencies (test) | All sub-modules                                                                                                    |
| External deps       | `charm.land/huh/v2`, `charm.land/lipgloss/v2`, `charm.land/log/v2`, `github.com/charmbracelet/fang`, `spf13/cobra` |

**Package structure:**

```
cmd/clean-wizard/            # CLI entry point
cmd/clean-wizard/commands/   # Cobra commands (clean, scan, config, profile)
internal/logger/             # Logging wrapper (charm.land/log/v2)
internal/testhelper/         # Integration test helpers
internal/version/            # Version info (build-time injected)
internal/shared/utils/       # Remaining utilities (gitfilterrepo, strings, schema, fileutil, config)
tests/bdd/                   # BDD integration tests
tests/benchmark/             # Benchmark tests
```

**Key decisions:**

1. **Root module stays as `app`** — It's the composition root. Wires cleaners, config, and adapters together.
2. **TUI deps isolated** — `huh`, `lipgloss`, `fang` only in root module. Cleaners never see TUI code.
3. **`internal/api` deleted** — Dead code (never imported). Not migrated.
4. **`internal/shared/utils/config` stays in root** — It depends on the `config` module; can't be in `core`.

---

## 5. DAG Verification

```
core          → (no internal deps) ✅
adapters      → core ✅
config        → core ✅
cleaners      → core, adapters ✅
app (root)    → core, cleaners, adapters, config ✅
```

No cycles. No upward dependencies from sub-modules to root. Valid DAG.

**Test dependency graph:**

```
core (tests)        → (self only) ✅
adapters (tests)    → core, adapters ✅
config (tests)      → core, config ✅
cleaners (tests)    → core, adapters, cleaners ✅
app (tests)         → all modules ✅ (composition root)
```

No bidirectional test dependencies in production code. ✅

---

## 6. Replace / Workspace Strategy

**Recommendation: `go.work` at repo root.**

```
go 1.26.2

use (
    .
    ./core
    ./cleaners
    ./adapters
    ./config
)
```

**Rationale:**

- Cleaner than per-module `replace` directives (one file vs 4×4 replace blocks)
- Go tooling handles `go.work` natively — `go build ./...`, `go test ./...` work at root
- `go.work` is automatically ignored when consumers import published modules
- No `replace` directives in any `go.mod` — clean for publishing

**Per-module go.mod files:**

- No `replace` directives in any module
- `go.work` handles local development
- When publishing: tag root module, consumers use normal `go get`

---

## 7. Test Dependency Isolation

| Module   | Production Deps                                                 | Test-Only Deps                                      |
| -------- | --------------------------------------------------------------- | --------------------------------------------------- |
| core     | `gopkg.in/yaml.v3`                                              | `stretchr/testify`                                  |
| adapters | core + `caarlos0/env`, `go-resty`, `otter`, `golang.org/x/time` | `stretchr/testify`                                  |
| config   | core + `knadh/koanf/v2`                                         | `stretchr/testify`                                  |
| cleaners | core, adapters + `cockroachdb/errors`, `golang.org/x/sys`       | `onsi/ginkgo/v2`, `onsi/gomega`, `stretchr/testify` |
| app      | all modules + `charm.land/*`, `spf13/cobra`                     | all test frameworks                                 |

**Critical fix:** Move all test helper files from `internal/cleaner/test_*.go` into `cleaners/testhelpers/` sub-package. This eliminates `testify`, `ginkgo`, and `gomega` from the cleaners production `go.mod`.

---

## 8. Interface Extraction Plan

### Core → Cleaners boundary

**In core:** `domain.OperationHandler`, `domain.Scanner` — abstract interfaces for any operation handler.

**In cleaners:** `cleaner.Cleaner` — concrete interface for cleaner implementations. Adds `Clean()`, `Scan()`, `IsAvailable()`, `Name()`, `Type()`.

The `Cleaner` interface lives in the cleaners module because it's specific to the cleaner subsystem. The core module only defines the abstract operation interfaces.

### Adapters → Cleaners boundary

**In adapters:** `NixAdapter`, `HTTPClient`, `CacheManager` — concrete implementations.

**In cleaners:** Cleaners depend on adapter interfaces via constructor injection. No direct struct coupling.

### Config → App boundary

**In config:** `Load()`, `Save()`, validation middleware — pure config operations.

**In app:** Commands call config functions. No config code imports CLI code.

---

## 9. Versioning Strategy

**Recommendation: Shared version (single git tag).**

| Strategy             | Chosen? | Rationale                                             |
| -------------------- | ------- | ----------------------------------------------------- |
| Shared version       | ✅ Yes  | Single CLI tool, no external consumers of sub-modules |
| Independent semver   | ❌ No   | Overhead without benefit — no external consumers      |
| Root-only versioning | ❌ No   | Subset of shared version                              |

All modules bump together under a single `vX.Y.Z` tag. The sub-modules are internal implementation detail — no one imports them independently.

---

## 10. Migration Strategy

Ordered steps, each independently executable. See [EXECUTION_PLAN.md](./EXECUTION_PLAN.md) for full details.

### Phase A: Pre-modularization Cleanup (1% → 51% impact)

1. Consolidate error packages (3 → 1)
2. Extract test helpers from `internal/cleaner/` into `internal/cleaner/testhelpers/`
3. Remove `stretchr/testify` production imports from `internal/domain`

### Phase B: Extract Core Module (4% → 64% impact)

4. Create `/core/go.mod`
5. Move domain, result, errors, format, conversions, validation, testing to core
6. Update all import paths
7. Create `go.work`

### Phase C: Extract Adapters Module (4% → 64% impact)

8. Create `/adapters/go.mod`
9. Move adapters to `/adapters/`
10. Update import paths

### Phase D: Extract Config Module (4% → 64% impact)

11. Create `/config/go.mod`
12. Move config to `/config/`
13. Remove logger dependency (use `log/slog`)

### Phase E: Extract Cleaners Module (4% → 64% impact)

14. Create `/cleaners/go.mod`
15. Move cleaners to `/cleaners/`
16. Update import paths

### Phase F: Finalize (20% → 80% impact)

17. Update root `go.mod` to depend on sub-modules
18. Update `flake.nix` for multi-module build
19. Verify full build + test suite
20. Update documentation

---

## 11. Risk Assessment

| Risk                                  | Likelihood | Impact | Mitigation                                                        |
| ------------------------------------- | ---------- | ------ | ----------------------------------------------------------------- |
| Import path breakage during migration | High       | High   | Automate with `sed`/`gofmt`; verify build after each step         |
| Test failures from moved test helpers | Medium     | Medium | Move test helpers first, verify all tests pass before proceeding  |
| Circular dependency discovery         | Low        | High   | DAG verified in proposal; `go vet` catches cycles at compile time |
| `go.work` confusion in CI             | Low        | Medium | Add `GOWORK=off` flag to CI release builds                        |
| Increased repo complexity             | Medium     | Low    | Clear README documenting module structure                         |
| `gopkg.in/yaml.v3` in core module     | Low        | Low    | Single dependency; acceptable for domain types with YAML tags     |

---

## 12. Build System Impact

### flake.nix

- Root build compiles all modules via `go.work`
- Per-module test targets: `nix run .#test-core`, `nix run .#test-cleaners`, etc.
- Lint runs across all modules
- CI can parallelize per-module builds

### CI/CD

- Add `go work sync` to CI pipeline
- Release builds use `GOWORK=off` + `replace` directives (or `go mod tidy` in each module)
- Test caching improves — unchanged modules skip test runs

---

## Appendix: Dead Code Identified

| Package        | Files                                   | Lines | Recommendation                     |
| -------------- | --------------------------------------- | ----- | ---------------------------------- |
| `internal/api` | 3 (mapper.go, types.go, mapper_test.go) | ~650  | **DELETE** — not imported anywhere |

## Appendix: Self-Review Findings

### Critical Corrections

1. **`internal/api` is dead code** — Not imported by any file (production or test). Must be deleted, not migrated. Reduces scope.

2. **`internal/conversions` is heavily coupled to cleaners** — 16 cleaner files import it. It's not a thin utility; it's a core part of the cleaner pipeline. It should move to the `cleaners` module, not `core`.

3. **`internal/format` is used by cleaners AND CLI** — `docker.go`, `systemcache.go` import format. It can't go in core alone. Options: (a) put in core (cleanest), or (b) duplicate the small formatting functions. **Decision: put in core** — it's tiny (67 lines non-test) and has only one external dep (`go-humanize`).

4. **Config's logger dependency is minimal** — Only `config.go` imports `logger`. Replacing with `slog` is straightforward.

5. **Only 3 adapter symbols are used by cleaners** — `ExecWithTimeout`, `NewNixAdapter`, `NixAdapter`. The adapter boundary is clean.

6. **`internal/testhelper` is only used by `test/` directory** — It can stay in root module. No need to put it in a sub-module.

7. **`internal/shared/utils/config` imports both `config` and `domain`** — This creates a circular dependency risk: if `config` module depends on `core`, and `shared/utils/config` depends on `config`, then `shared/utils/config` must be in `config` or `app` module, not in `core`.

### Revised Package Placement

Based on findings:

| Package                        | Proposed Module | Revised Module | Reason                                                |
| ------------------------------ | --------------- | -------------- | ----------------------------------------------------- |
| `internal/api`                 | app             | **DELETE**     | Dead code                                             |
| `internal/conversions`         | core            | **cleaners**   | 16 cleaner files depend on it; it's a cleaner concern |
| `internal/shared/utils/config` | core            | **app**        | Depends on `config` module; can't be in core          |
| `internal/testhelper`          | app             | **app**        | Only used by integration tests; stays in root         |
| `internal/format`              | core            | **core**       | Confirmed: small, shared by cleaners and CLI          |
| `internal/logger`              | app             | **app**        | Wraps charm.land/log; TUI concern                     |

### Updated Module Dependencies

```
core        → (no internal deps) ✅
adapters    → core ✅
config      → core ✅ (logger dep removed, use slog)
cleaners    → core, adapters ✅ (includes conversions)
app (root)  → core, cleaners, adapters, config ✅
```

Per `how-to-golang` policy:

| Current Dependency       | Status           | Action                                                        |
| ------------------------ | ---------------- | ------------------------------------------------------------- |
| `stretchr/testify`       | BANNED           | Keep for existing tests; migrate to `ginkgo/gomega` over time |
| `gopkg.in/yaml.v3`       | BANNED           | Replace with `go-faster/yaml` (defer to separate task)        |
| `charm.land/huh/v2`      | OK (v2)          | No action                                                     |
| `charm.land/lipgloss/v2` | OK (v2)          | No action                                                     |
| `charm.land/log/v2`      | OK (v2)          | No action                                                     |
| `cockroachdb/errors`     | OK               | No action                                                     |
| `knadh/koanf/v2`         | OK               | No action                                                     |
| `spf13/cobra`            | OK               | No action                                                     |
| `maypok86/otter/v2`      | OK (recommended) | No action                                                     |
| `dustin/go-humanize`     | OK               | No action                                                     |

**IMPORTANT:** `internal/api` is dead code — not imported by any file (production or test). It should be deleted, not migrated. Removing it eliminates 2 packages and ~650 lines.
