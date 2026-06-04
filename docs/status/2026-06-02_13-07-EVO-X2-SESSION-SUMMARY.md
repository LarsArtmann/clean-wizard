# Status Report: evo-x2 NixOS Support — Session Summary

**Date:** 2026-06-02 13:07
**Machine:** evo-x2 (NixOS Linux x86_64, Nix 2.34.7, Go 1.26.3)
**Branch:** master (clean, pushed to origin)
**Commits this session:** 5

---

## a) FULLY DONE ✅

### 1. 12 New CacheType Enums (15→27)

**Commit:** `2f85b44`

- Added: Gopls, Goimports, JetBrains, BunCache, Playwright, Mozilla, NixCache, Zig, Uv, Tinygo, MesaShader, Comgr
- All have `cacheTypeConfig` entries with real paths, display names, scan types
- `cacheTypeStrings`, `IsValid()`, `Values()` all updated
- **Potential reclaimable:** ~17.3GB on evo-x2

### 2. Platform-Aware Defaults

**Commit:** `2f85b44`

- `getDefaultSystemCacheTypes()`: Linux now returns XdgCache, Thumbnails, Pip, Npm, Yarn, Ccache (was only XdgCache + Thumbnails + Homebrew)
- `DefaultProtectedPaths()`: macOS→`/System,/Applications,/Library`; Linux→`/nix/store,/nix/var`
- `CriticalSystemPaths()`: universal (`/`, `/usr`, `/etc`)
- `AllProtectedSystemPaths()`: platform-conditional
- Config defaults (`config.go`, `enhanced_loader.go`, `validator_structure.go`) use `DefaultProtectedPaths()`

### 3. System Cache Cleaner Updated

**Commit:** `ba26a69`

- `AvailableSystemCacheTypes()`: Linux=23 types, macOS=19 types
- Homebrew removed from Linux availability list
- All 12 new cache configs added to `systemCacheConfigs` map

### 4. Nix Store GC Modernized

**Commit:** `01c2f04`

- Replaced legacy `nix-collect-garbage -d` with modern `nix store gc`
- Semantically correct: generations already deleted individually by caller before GC runs

### 5. Tests & Schema Updated

**Commit:** `c74ac2a`

- `enum_yaml_test.go`: CacheType test data expanded to 27 values
- `config.schema.json`: all 27 cache type values documented
- **290 tests pass, 0 failures**

### 6. Documentation Updated

**Commit:** `1831ab6`

- `docs/YAML_ENUM_FORMATS.md`: full 27-value CacheType table with platform tags
- `FEATURES.md`: updated cache type counts and status
- `AGENTS.md`: added evo-x2 machine context, target machines section

---

## b) PARTIALLY DONE ⚠️

### Nix Profile Support

- `nix-env --list-generations` works for system profile (generation 32)
- Per-user profile at `~/.local/state/nix/profiles/profile` has 32 generations managed by `nix profile`
- `nix profile` uses different API (`history`, `rollback`) vs `nix-env` (`--list-generations`, `--delete-generations`)
- **Impact:** Per-user profile generations are NOT cleaned

### Cache Type Test Coverage

- `enumTypeDefinitions` in yaml test has all 27 values for generated unmarshal tests ✅
- Explicit marshaling test cases only cover 8/27 CacheType values (Spotlight through Ccache) ❌

---

## c) NOT STARTED 📝

| #   | Task                                                                        | Impact | Effort |
| --- | --------------------------------------------------------------------------- | ------ | ------ |
| 1   | `nix profile` per-user generation management                                | HIGH   | MED    |
| 2   | `nix store optimise` hard-link deduplication                                | MED    | LOW    |
| 3   | Consolidate 4 error packages into 1                                         | HIGH   | MED    |
| 4   | BDD tests for Docker, Homebrew, Go cleaners                                 | HIGH   | HIGH   |
| 5   | Fix ~40 `err113` lint violations                                            | MED    | MED    |
| 6   | CI pipeline (`go build` + `go test`)                                        | HIGH   | MED    |
| 7   | Split `internal/domain/` into sub-packages                                  | HIGH   | HIGH   |
| 8   | Add `platform` field to `cacheTypeConfig`                                   | MED    | LOW    |
| 9   | Centralize `runtime.GOOS` into platform detector                            | MED    | MED    |
| 10  | Split files >350 lines (compiledbinaries 592, docker 540, nodepackages 540) | MED    | MED    |
| 11  | CLI command tests (scan, clean, profile, config)                            | MED    | HIGH   |
| 12  | Fix mixed receiver warnings (10 enum types)                                 | LOW    | LOW    |
| 13  | Reduce `GetOperationType` complexity (17→<10)                               | LOW    | LOW    |
| 14  | Investigate `net.imput.helium` cache (2.4GB)                                | MED    | LOW    |
| 15  | Investigate Go `~/.cache/go` dir overlap with GoCacheCleaner                | MED    | LOW    |
| 16  | Improve Nix size estimation (hardcoded 50MB/gen)                            | MED    | MED    |
| 17  | Extract test utilities from `internal/cleaner/test_*.go`                    | MED    | MED    |
| 18  | Add service layer `internal/service/` between CLI and domain                | MED    | HIGH   |
| 19  | Split `internal/cleaner/` into per-domain sub-packages                      | HIGH   | HIGH   |
| 20  | Homebrew dry-run support or documented limitation                           | LOW    | MED    |

---

## d) TOTALLY FUCKED UP 💥

**None.** Zero regressions. Build passes, all 290 tests pass, working tree clean.

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Architecture

1. **CacheType enum explosion (27 values)** — Consider a registry pattern where cleaners register their own cache types instead of one monolithic enum
2. **`systemCacheConfigs` global map** — Should be a registry that cleaners contribute to
3. **GoCacheCleaner has its own GoCacheType enum** separate from CacheType — goimports/gopls overlap conceptually
4. **NixAdapter hardcodes profile paths** — Should discover dynamically
5. **`internal/cleaner/` has 60 files flat** — No sub-packages; should split by domain
6. **`internal/domain/` has 22 files** — God package; should split into `enums/`, `operations/`, `types/`
7. **4 error packages** — `internal/errors/`, `internal/pkg/errors/`, `adapters/errors.go`, `pkg/errors/`

### Type Model

8. **`cacheTypeConfig` has no `platform` field** — Platform filtering done externally; easy to accidentally use wrong types
9. **`NixGeneration` doesn't distinguish** `nix-env` vs `nix profile` generations
10. **`runtime.GOOS` checks scattered** across 4+ files — Centralize into platform detector

### Library Usage

11. **`charm.land/huh/v2`** — Could use for interactive cache type selection showing only platform-appropriate options
12. **`koanf/v2`** — Could add provider that auto-detects platform cache types from filesystem

---

## f) Top 25 Next Actions (sorted by impact/effort)

| #   | Task                                                      | Impact | Effort | Category      |
| --- | --------------------------------------------------------- | ------ | ------ | ------------- |
| 1   | Add `nix store optimise` to NixCleaner                    | HIGH   | LOW    | Feature       |
| 2   | Add marshaling tests for all 27 CacheType values          | MED    | LOW    | Testing       |
| 3   | Fix systemcache_test.go mixed test (uses macOS-only type) | LOW    | LOW    | Testing       |
| 4   | Fix mixed receiver warnings (10 enum types)               | LOW    | LOW    | Lint          |
| 5   | Reduce `GetOperationType` complexity (17→<10)             | LOW    | LOW    | Lint          |
| 6   | Add `platform` field to `cacheTypeConfig`                 | MED    | LOW    | Type Model    |
| 7   | Investigate `net.imput.helium` cache (2.4GB)              | MED    | LOW    | Investigation |
| 8   | Investigate `~/.cache/go` overlap with GoCacheCleaner     | MED    | LOW    | Investigation |
| 9   | Add `nix profile` per-user generation management          | HIGH   | MED    | Feature       |
| 10  | Consolidate 4 error packages into 1                       | HIGH   | MED    | Architecture  |
| 11  | Set up CI pipeline                                        | HIGH   | MED    | Infra         |
| 12  | Split files >350 lines                                    | MED    | MED    | Code Quality  |
| 13  | Fix ~40 `err113` lint violations                          | MED    | MED    | Lint          |
| 14  | Centralize `runtime.GOOS` into platform detector          | MED    | MED    | Architecture  |
| 15  | Improve Nix size estimation (hardcoded 50MB/gen)          | MED    | MED    | Feature       |
| 16  | Extract test utilities from cleaner/test\_\*.go           | MED    | MED    | Architecture  |
| 17  | Add profile command tests                                 | MED    | MED    | Testing       |
| 18  | BDD tests for Docker, Homebrew, Go cleaners               | HIGH   | HIGH   | Testing       |
| 19  | CLI command tests (scan, clean, profile, config)          | MED    | HIGH   | Testing       |
| 20  | Split `internal/domain/` into sub-packages                | HIGH   | HIGH   | Architecture  |
| 21  | Add service layer `internal/service/`                     | MED    | HIGH   | Architecture  |
| 22  | Split `internal/cleaner/` into per-domain sub-packages    | HIGH   | HIGH   | Architecture  |
| 23  | Make NixAdapter discover profile paths dynamically        | MED    | MED    | Architecture  |
| 24  | Homebrew dry-run support or documented limitation         | LOW    | MED    | Feature       |
| 25  | Language Version Manager cleaner (currently NO-OP)        | MED    | MED    | Feature       |

---

## g) My #1 Unresolvable Question

**How should `nix profile` generations be handled architecturally?**

The `NixAdapter` has a clean `ListGenerations() / RemoveGeneration() / CollectGarbage()` flow tied to `nix-env`. But `nix profile` uses a fundamentally different API:

| Operation | `nix-env` (current)                   | `nix profile` (needed)                           |
| --------- | ------------------------------------- | ------------------------------------------------ |
| List      | `nix-env --list-generations`          | `nix profile history` (different format)         |
| Delete    | `nix-env --delete-generations <id>`   | No direct equivalent; `nix profile remove <idx>` |
| Profiles  | System-level `/nix/var/nix/profiles/` | Per-user `~/.local/state/nix/profiles/profile`   |

Options:

1. **Separate `NixProfileAdapter`** alongside existing `NixAdapter` — cleanest separation
2. **Auto-detect in `NixAdapter`** — try `nix-env`, fall back to `nix profile`
3. **Unify with discriminator** — `NixGeneration.ProfileType` field

This shapes the `NixGeneration` domain type and the adapter interface. I cannot decide without knowing your preference for adapter granularity.

---

## Codebase Health

| Metric                 | Value                                                                             |
| ---------------------- | --------------------------------------------------------------------------------- |
| Build                  | ✅ Clean                                                                          |
| Tests                  | 290 pass, 0 fail                                                                  |
| Test files             | 63                                                                                |
| Lint warnings          | ~115 (pre-existing, not introduced this session)                                  |
| Largest source files   | compiledbinaries.go (592), systemcache.go (542), nodepackages.go (540)            |
| Packages without tests | 7 (cmd/clean-wizard, errors, pkg/errors, fileutil, gitfilterrepo, testutil, test) |
| Cache types            | 27 enum values                                                                    |
| Cleaners               | 13+ implementations                                                               |

## evo-x2 Cache Footprint (as of this report)

| Cache                  | Size         | Covered By                |
| ---------------------- | ------------ | ------------------------- |
| /nix/store             | 81 GB        | NixCleaner                |
| go-build               | 27 GB        | GoCacheCleaner            |
| goimports              | 8.2 GB       | SystemCache (NEW)         |
| pip                    | 6.5 GB       | SystemCache               |
| nix (~/.cache)         | 3.1 GB       | SystemCache (NEW)         |
| gopls                  | 2.1 GB       | SystemCache (NEW)         |
| net.imput.helium       | 2.4 GB       | NOT COVERED               |
| puppeteer              | 1.2 GB       | SystemCache               |
| JetBrains              | 722 MB       | SystemCache (NEW)         |
| pnpm                   | 717 MB       | NodePackageManagerCleaner |
| comgr                  | 661 MB       | SystemCache (NEW)         |
| ms-playwright          | 641 MB       | SystemCache (NEW)         |
| mozilla                | 541 MB       | SystemCache (NEW)         |
| golangci-lint          | —            | GolangciLintCacheCleaner  |
| bun                    | 166 MB       | SystemCache (NEW)         |
| zig                    | 69 MB        | SystemCache (NEW)         |
| uv                     | 60 MB        | SystemCache (NEW)         |
| tinygo                 | 52 MB        | SystemCache (NEW)         |
| mesa_shader_cache      | 49 MB        | SystemCache (NEW)         |
| **Total covered**      | **~131 GB**  |                           |
| **Total NEW coverage** | **~17.3 GB** |                           |
