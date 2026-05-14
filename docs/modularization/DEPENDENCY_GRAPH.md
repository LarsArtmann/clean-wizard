# Dependency Graph Analysis

**Generated:** 2026-05-14 | **Module:** `github.com/LarsArtmann/clean-wizard`

## Current State: Single Monolith

One `go.mod` at project root. No `go.work`. No sub-modules. All 29 packages in a single module.

---

## Internal Package Dependency Graph

### Production Code Dependencies (non-test imports)

```
cmd/clean-wizard
  └── cmd/clean-wizard/commands
        ├── internal/cleaner
        ├── internal/config
        ├── internal/domain
        └── internal/format

internal/domain           ← LEAF (zero internal deps)
internal/result           ← LEAF (zero internal deps)
internal/errors           ← LEAF (zero internal deps)
internal/pkg/errors       ← LEAF (zero internal deps)
internal/version          ← LEAF (zero internal deps)
internal/testing          ← LEAF (zero internal deps)
internal/format
  └── internal/domain

internal/logger           ← LEAF (self-reference only)

internal/shared/utils/
  ├── fileutil            ← LEAF
  ├── gitfilterrepo       ← LEAF
  ├── schema              ← LEAF
  ├── strings             ← LEAF
  ├── testutil            ← LEAF
  ├── validation
  │     └── internal/result
  └── config
        ├── internal/config
        └── internal/domain

internal/conversions
  ├── internal/domain
  ├── internal/result
  └── internal/shared/utils/validation

internal/adapters
  ├── internal/domain
  ├── internal/result
  └── internal/conversions

internal/api
  ├── internal/domain
  └── internal/result

internal/middleware
  ├── internal/domain
  ├── internal/result
  └── internal/shared/utils/validation

internal/config
  ├── internal/domain
  ├── internal/logger
  ├── internal/pkg/errors
  └── internal/shared/utils/strings

internal/cleaner          ← LARGEST CONSUMER (imports 8 internal packages)
  ├── internal/adapters
  ├── internal/conversions
  ├── internal/domain
  ├── internal/format
  ├── internal/result
  ├── internal/shared/utils/fileutil
  ├── internal/shared/utils/gitfilterrepo
  ├── internal/shared/utils/testutil
  └── internal/testing

internal/testhelper
  ├── internal/cleaner
  └── internal/format
```

### Test-Only Dependencies (imports only in `_test.go`)

```
internal/adapters (tests)     → internal/adapters (self)
internal/cleaner (tests)      → internal/adapters, domain, result, testutil
internal/config (tests)       → internal/domain
internal/conversions (tests)  → internal/domain, result
internal/middleware (tests)   → internal/domain
internal/shared/utils/config (tests) → internal/domain
tests/bdd                     → internal/cleaner, domain, result
tests/benchmark               → internal/domain, result
test                          → internal/testhelper
```

---

## External Dependency Map Per Package

| Package | Key External Dependencies |
|---|---|
| `cmd/clean-wizard` | `charmbracelet/fang` |
| `cmd/clean-wizard/commands` | `charm.land/huh/v2`, `charm.land/lipgloss/v2`, `spf13/cobra` |
| `internal/cleaner` | `cockroachdb/errors`, `golang.org/x/sys/unix`, `onsi/ginkgo`, `onsi/gomega`, `stretchr/testify` |
| `internal/config` | `knadh/koanf/v2`, `knadh/koanf/parsers/yaml`, `knadh/koanf/providers/file` |
| `internal/domain` | `gopkg.in/yaml.v3` |
| `internal/adapters` | `caarlos0/env/v6`, `go-resty/resty/v2`, `maypok86/otter/v2`, `golang.org/x/time/rate`, `stretchr/testify` |
| `internal/format` | `dustin/go-humanize` |
| `internal/logger` | `charm.land/log/v2` |
| `internal/shared/utils/config` | `charm.land/log/v2` |
| `tests/bdd` | `onsi/ginkgo/v2`, `onsi/gomega` |

---

## Coupling Hotspots

### 1. `internal/domain` — God Package (22 non-test files, 2,442 lines)

Every other internal package imports `domain`. It contains:
- 19 type-safe enums (iota-based)
- Core domain types: `CleanResult`, `ScanResult`, `ScanItem`, `CleanRequest`, `ScanRequest`
- Nix-specific types: `NixGeneration`, `NixGenerationsSettings`
- Docker-specific types: `DockerSettings`, `DockerPruneMode`
- Git-history types: `GitHistorySettings`, `GitHistoryFile`, `GitHistoryScanResult`
- All operation settings structs (15+ settings types)
- Interfaces: `OperationHandler`, `GenerationCleaner`, `PackageCleaner`, `Scanner`
- Enum macros / generic helpers
- Duration parser
- System paths

**Problem:** Domain mixes truly shared types (`CleanResult`) with cleaner-specific types (`NixGenerationsSettings`). Any change to any cleaner's types forces recompilation of every package.

### 2. `internal/cleaner` — God Package (60 files, ~16,000 lines)

Contains 14 concrete cleaner implementations, shared infrastructure, and test helpers — all in one flat package. Files exceed 300-line threshold (12 files).

**Problem:** Impossible to version, test, or reason about individual cleaners independently.

### 3. Three Error Packages (Split Brain)

| Package | Purpose |
|---|---|
| `internal/errors` | 7 error constructors (`CleanOperationError`, `ConfigLoadError`, etc.) |
| `internal/pkg/errors` | Structured error types with `CleanWizardError`, `ErrorCode`, `ErrorLevel`, builder pattern |
| `pkg/errors` | Minimal `BaseError` with `New()` constructor |

**Problem:** Overlapping responsibilities. No clear ownership. Consumers use different packages for similar needs.

### 4. `internal/cleaner` → `internal/adapters` Coupling

Cleaners import adapters (NixAdapter, ExecWithTimeout). Adapters import conversions + domain. This means changing adapter implementations forces cleaner recompilation.

### 5. Test Dependencies in Production Code

- `internal/cleaner` imports `stretchr/testify` in production files (test helpers that aren't in `_test.go`)
- `internal/adapters` imports `stretchr/testify` in production code
- `internal/domain` imports `stretchr/testify` in production code
- `internal/cleaner` imports `onsi/ginkgo` + `onsi/gomega` in non-test helper files

---

## Dependency Layer Analysis

```
Layer 0 (Leaf — zero deps):
  internal/domain, internal/result, internal/errors,
  internal/pkg/errors, internal/version, internal/testing,
  internal/shared/utils/{fileutil, gitfilterrepo, schema, strings, testutil}

Layer 1 (Depends on Layer 0 only):
  internal/format, internal/logger,
  internal/shared/utils/validation, internal/shared/utils/config

Layer 2 (Depends on Layer 0-1):
  internal/conversions, internal/middleware, internal/config, internal/api

Layer 3 (Depends on Layer 0-2):
  internal/adapters

Layer 4 (Depends on everything):
  internal/cleaner, internal/testhelper

Layer 5 (Application):
  cmd/clean-wizard/commands, cmd/clean-wizard

Layer 6 (Tests):
  tests/bdd, tests/benchmark, test
```

The graph is a valid DAG — no circular dependencies detected.

---

## Package Size Distribution

| Package | Files (non-test) | Lines (non-test) | Status |
|---|---|---|---|
| `internal/cleaner` | 37 | ~16,000 | CRITICAL — god package |
| `internal/domain` | 15 | ~2,442 | WARNING — mixed concerns |
| `internal/config` | 25 | ~3,400 | WARNING — large |
| `internal/adapters` | 7 | ~1,046 | OK |
| `internal/conversions` | 1 | ~400 | OK |
| `internal/api` | 2 | ~351 | OK |
| `internal/result` | ~3 | ~500 | OK |
| `internal/shared/utils/*` | 6 | ~305 | OK |
| All others | ~10 | ~300 | OK |
