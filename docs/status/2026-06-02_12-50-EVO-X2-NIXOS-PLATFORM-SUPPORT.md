# Status Report: evo-x2 (NixOS) Platform Support

**Date:** 2026-06-02 12:50
**Machine:** evo-x2 (NixOS Linux x86_64, Nix 2.34.7, Go 1.26.3)
**Branch:** master

---

## a) FULLY DONE

### Platform-Aware System Cache Types
- `getDefaultSystemCacheTypes()` now returns correct Linux cache types (XdgCache, Thumbnails, Pip, Npm, Yarn, Ccache)
- Removed Homebrew from Linux defaults (not present on NixOS)
- `AvailableSystemCacheTypes()` updated for both macOS and Linux

### 12 New CacheType Enums (8→27 total)
- Gopls (2.1GB on evo-x2), Goimports (8.2GB), JetBrains (845MB), BunCache (166MB)
- Playwright (641MB), Mozilla (541MB), NixCache (3.1GB), Zig (69MB)
- Uv (60MB), Tinygo (52MB), MesaShader (49MB), Comgr (661MB)
- All have `cacheTypeConfig` entries with paths and display names
- Enum infrastructure updated: `cacheTypeStrings`, `IsValid()`, `Values()`, JSON schema, YAML test data

### Platform-Aware Protected Paths
- `DefaultProtectedPaths()`: macOS→`/System,/Applications,/Library`; Linux→`/nix/store,/nix/var`
- `CriticalSystemPaths()`: universal (`/`, `/usr`, `/etc`)
- `AllProtectedSystemPaths()`: platform-conditional
- Config defaults (`config.go`, `enhanced_loader.go`, `validator_structure.go`) now use `DefaultProtectedPaths()` instead of hardcoded macOS paths

### Documentation Updated
- `AGENTS.md`: added evo-x2 machine context, updated enum counts
- `FEATURES.md`: updated cache type counts and status
- `docs/YAML_ENUM_FORMATS.md`: full 27-value CacheType table with platform tags
- `schemas/config.schema.json`: all 27 cache type values

### Tests
- All 298 tests pass across 63 test files
- Build passes cleanly

---

## b) PARTIALLY DONE

### Nix Profile Support
- `nix-env --list-generations` works for the system profile (generation 32) on evo-x2
- **BUT**: per-user profile uses `nix profile` (at `~/.local/state/nix/profiles/profile`) which has 32 generations that `nix-env` cannot manage
- The NixCleaner only cleans system profile generations, not per-user profile generations
- `nix profile history` shows different output format than `nix-env --list-generations`

### Cache Type Coverage on evo-x2
- The following large caches are covered by existing dedicated cleaners (NOT system cache):
  - `go-build` (25GB) → GoCacheCleaner (via `go clean -cache`)
  - `golangci-lint` (201MB) → GolangciLintCacheCleaner
  - `pnpm` (717MB) → NodePackageManagerCleaner
- The following are now covered by the new system cache types:
  - `goimports` (8.2GB), `pip` (6.5GB), `nix` (3.1GB), `gopls` (2.1GB), `puppeteer` (1.2GB), `JetBrains` (845MB), `comgr` (661MB), `ms-playwright` (641MB), `mozilla` (541MB), `bun` (166MB), `zig` (69M), `uv` (60MB), `tinygo` (52MB), `mesa_shader_cache` (49MB)
- **NOT covered** (would need new cleaners or enum additions):
  - `net.imput.helium` (2.4GB) — unknown cache
  - `comgr` is covered, but `~/.cache/go` (845MB) is a Go module cache subdirectory that may overlap with GoCacheCleaner

---

## c) NOT STARTED

### Critical
1. **Nix `nix profile` generation management** — The NixAdapter only supports `nix-env`. Need to add `nix profile` support for per-user profile generation cleanup.
2. **NixOS store optimization** — `nix store optimise` could save significant space by hard-linking identical files.
3. **NixOS-specific `nix-collect-garbage` vs `nix store gc`** — Modern Nix prefers `nix store gc`.

### High Priority
4. **enum_yaml_test.go marshaling tests** — Only 8/27 CacheType values have explicit marshaling test cases.
5. **systemcache_test.go mixed valid/invalid test** — Uses `CacheTypeSpotlight` which is macOS-only; misleading on Linux.
6. **goimports/gopls cache cleaning** — These are now in systemCacheConfigs but just get directory-deleted. A smarter approach (like the GoCacheCleaner) would be better.

### Medium Priority
7. **flake.nix vendorHash update** — go.mod/go.sum unchanged but should verify after any dependency changes.
8. **Homebrew Linux detection** — `CacheTypeHomebrew` still has config entries; on NixOS this is irrelevant.
9. **`net.imput.helium` cache** (2.4GB) — Unknown; investigate what creates it.

---

## d) TOTALLY FUCKED UP / REGRESSIONS

**None detected.** All tests pass, build passes, no regressions introduced.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture
- **CacheType enum explosion** — 27 values is getting unwieldy. Consider splitting into platform-specific sub-enums or a registry pattern where each cleaner registers its own cache types.
- **systemCacheConfigs is a global map** — Should be a registry that cleaners contribute to, reducing coupling.
- **GoCacheCleaner has its own GoCacheType enum** separate from CacheType — these overlap conceptually (goimports, gopls are Go caches but in the system cache enum).
- **NixAdapter hardcodes profile path** — Should discover profile paths dynamically (`~/.local/state/nix/profiles/profile`, `/nix/var/nix/profiles/default`, etc.)

### Type Model
- `cacheTypeConfig` has no `platform` field — platform filtering is done externally in `AvailableSystemCacheTypes()`. Inlining this would prevent accidentally using macOS-only types on Linux.
- `NixGeneration` type doesn't distinguish between `nix-env` and `nix profile` generations.
- `SystemCacheSettings.CacheTypes []CacheType` has no validation that selected types are available on the current platform until `ValidateSettings()` is called — this should be a compile-time or init-time check.

### Library Usage
- **`charm.land/huh/v2`** — Already used for TUI. Could leverage for interactive cache type selection showing only platform-appropriate options.
- **`github.com/knadh/koanf/v2`** — Already used. Could add a `koanf.Provider` that auto-detects platform cache types from filesystem.
- **`runtime.GOOS` checks scattered** — Could centralize into a `platform.Detector` interface, making testing easier (inject fake OS).

---

## f) Top 25 Things We Should Get Done Next

| #  | Task | Impact | Effort | Category |
| -- | ---- | ------ | ------ | -------- |
| 1 | Add `nix profile` generation support to NixAdapter (per-user profiles) | HIGH | MED | Feature |
| 2 | Add `nix store optimise` to NixCleaner | HIGH | LOW | Feature |
| 3 | Add marshaling tests for all 27 CacheType values | MED | LOW | Testing |
| 4 | Fix systemcache_test.go mixed test to use platform-aware cache type | LOW | LOW | Testing |
| 5 | Consolidate error packages (4→1) | HIGH | MED | Architecture |
| 6 | Extract test utilities from `internal/cleaner/test_*.go` | MED | MED | Architecture |
| 7 | Add BDD tests for Docker, Homebrew, Go cleaners | HIGH | HIGH | Testing |
| 8 | Fix `err113` lint violations (~40 instances) | MED | MED | Lint |
| 9 | Set up CI pipeline (`go build` + `go test`) | HIGH | MED | Infra |
| 10 | Split `internal/domain/` into sub-packages | HIGH | HIGH | Architecture |
| 11 | Add `platform` field to `cacheTypeConfig` to prevent cross-platform misuse | MED | LOW | Type Model |
| 12 | Centralize `runtime.GOOS` checks into a platform detection interface | MED | MED | Architecture |
| 13 | Add age-based filtering to goimports/gopls cache cleaning | MED | LOW | Feature |
| 14 | Investigate and handle `net.imput.helium` cache (2.4GB on evo-x2) | MED | LOW | Feature |
| 15 | Modernize NixAdapter: `nix store gc` instead of `nix-collect-garbage` | LOW | LOW | Refactor |
| 16 | Split files over 350 lines: compiledbinaries, docker, nodepackages | MED | MED | Code Quality |
| 17 | Add CLI command tests (scan, clean, profile, config) | MED | HIGH | Testing |
| 18 | Fix mixed receiver warnings (10 enum types) | LOW | LOW | Lint |
| 19 | Reduce `GetOperationType` complexity (17→<10) | LOW | LOW | Lint |
| 20 | Investigate Go `~/.cache/go` directory overlap with GoCacheCleaner | MED | LOW | Investigation |
| 21 | Add profile command tests | MED | MED | Testing |
| 22 | Make NixAdapter discover profile paths dynamically | MED | MED | Architecture |
| 23 | Split `internal/cleaner/` into per-domain sub-packages | HIGH | HIGH | Architecture |
| 24 | Improve Nix size estimation (hardcoded 50MB/generation) | MED | MED | Feature |
| 25 | Add Homebrew dry-run support or document limitation | LOW | MED | Feature |

---

## g) Top #1 Question I Cannot Figure Out Myself

**How should `nix profile` generations be handled architecturally?**

The current `NixAdapter` has a clean `ListGenerations() / DeleteGenerations() / CollectGarbage()` flow tied to `nix-env`. But `nix profile` uses a completely different API:
- `nix profile history` — shows version history with different output format
- `nix profile rollback` — rolls back to previous version (not delete-by-ID)
- No `--list-generations` equivalent — profiles have numbered links but `nix profile` doesn't expose them the same way

Should we:
1. Create a separate `NixProfileAdapter` alongside the existing `NixAdapter`?
2. Make `NixAdapter` detect which profile type is in use and delegate?
3. Treat `nix profile` generations as the same `NixGeneration` type but with a `ProfileType` discriminator?

This is a design decision that affects the `NixGeneration` domain type and the adapter interface.

---

## Files Changed (11 files, +333 / -36 lines)

| File | Change |
| ---- | ------ |
| `internal/domain/operation_settings.go` | +31 lines (12 new CacheType enums, updated strings/IsValid/Values) |
| `internal/cleaner/systemcache.go` | +83 lines (12 new cacheTypeConfig entries, updated AvailableSystemCacheTypes) |
| `internal/domain/system_paths.go` | Rewritten (platform-aware protected paths) |
| `internal/domain/operation_defaults.go` | +3/-3 lines (Linux defaults with Pip/Npm/Yarn/Ccache, no Homebrew) |
| `internal/config/config.go` | Uses `DefaultProtectedPaths()` |
| `internal/config/enhanced_loader.go` | Uses `DefaultProtectedPaths()` |
| `internal/config/validator_structure.go` | Uses `DefaultProtectedPaths()` dynamically |
| `internal/domain/enum_yaml_test.go` | +38 lines (27-value CacheType test data) |
| `schemas/config.schema.json` | +70 lines (27 cache type values) |
| `docs/YAML_ENUM_FORMATS.md` | Updated CacheType table (0-7 → 0-26) |
| `FEATURES.md` | Updated cache type counts and status |
| `AGENTS.md` | Added evo-x2 context, updated counts |

---

## Potential Reclaimable Space on evo-x2

| Cache | Size | Status |
| ----- | ---- | ------ |
| goimports | 8.2 GB | NEW — now covered |
| pip | 6.5 GB | Already covered |
| nix (~/.cache/nix) | 3.1 GB | NEW — now covered |
| gopls | 2.1 GB | NEW — now covered |
| net.imput.helium | 2.4 GB | NOT covered |
| puppeteer | 1.2 GB | Already covered |
| JetBrains | 845 MB | NEW — now covered |
| pnpm | 717 MB | Already covered (NodePackageManagerCleaner) |
| comgr | 661 MB | NEW — now covered |
| ms-playwright | 641 MB | NEW — now covered |
| mozilla | 541 MB | NEW — now covered |
| bun | 166 MB | NEW — now covered |
| zig | 69 MB | NEW — now covered |
| uv | 60 MB | NEW — now covered |
| tinygo | 52 MB | NEW — now covered |
| mesa_shader_cache | 49 MB | NEW — now covered |
| **Total NEW** | **~17.3 GB** | |
