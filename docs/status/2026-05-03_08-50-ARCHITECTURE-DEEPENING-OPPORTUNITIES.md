# Improve Codebase Architecture — Deepening Opportunities — 2026-05-03

## Glossary (Applied)

- **Module** — anything with an interface and implementation (function, class, package)
- **Depth** — leverage at the interface: lots of behavior behind a small interface
- **Seam** — where an interface lives; a place behavior can be altered without editing in place
- **Adapter** — a concrete thing satisfying an interface at a seam

## Deepening Opportunities

### 1. Error Package Consolidation (Split Brain)

**Files:** `internal/errors/`, `internal/pkg/errors/`, `internal/adapters/errors.go`, `pkg/errors/`
**Problem:** Four separate error packages with overlapping responsibilities. `internal/errors/` has 40 lines, `internal/pkg/errors/` has 500+ lines with 9 files, `internal/adapters/errors.go` duplicates patterns. Callers must choose between packages arbitrarily.
**Solution:** Consolidate into `internal/errors/` with sub-categories. Remove `pkg/errors/` (empty wrapper). Merge `adapters/errors.go` functions into adapter-specific error types.
**Benefits:** Locality — one place for all error definitions. Callers don't need to guess which package. Leverage — error construction patterns centralized.

### 2. Domain Package Split (God Package)

**Files:** `internal/domain/` — 20+ files, 4000+ lines
**Problem:** All domain types, enums, settings, interfaces, validation, defaults in one flat package. High cognitive load to find anything. New developers can't predict where a type lives.
**Solution:** Split into `internal/domain/enums/`, `internal/domain/operations/`, `internal/domain/types/`, `internal/domain/interfaces.go`. The enum macros stay in enums, the settings stay in operations, core types stay in types.
**Benefits:** Locality — related types grouped together. Depth — each sub-package has a clear, small interface. Deletion test passes — removing any sub-package concentrates complexity, proving each earns its keep.

### 3. Cleaner Package Decomposition (Flat 50-File Package)

**Files:** `internal/cleaner/` — 50+ files in one package
**Problem:** 13 cleaner implementations, 6 test helper files, registry, factories, metrics, validation — all flat. The package has no internal structure. Test helpers (`test_*.go`, `ginkgo_test_helpers.go`, `testing_helpers.go`) live alongside production code.
**Solution:** Extract test utilities to `internal/testutil/cleaner/`. Group complex cleaners (GitHistory has 7 files, Docker has 2) into sub-packages if they exceed ~350 lines. Keep the `Cleaner` interface and `Registry` in the root.
**Benefits:** Test code doesn't pollute production package. Locality — each cleaner's related files are together. Deletion test — removing a cleaner sub-package doesn't cascade.

### 4. Service Layer Introduction (Missing Seam)

**Files:** `cmd/clean-wizard/commands/*.go`
**Problem:** CLI commands directly construct registries, call cleaners, and handle results. There's no seam between "CLI orchestration" and "business logic." Testing commands requires full system integration.
**Solution:** Introduce `internal/service/` with `CleanService`, `ScanService`, `ConfigService`. Commands call services; services call domain/cleaners. This creates a seam where mock services can be injected for CLI testing.
**Benefits:** Leverage — one service interface replaces duplicated command orchestration. Locality — business logic concentrated in services. Two adapters (real + mock) at the service seam = real seam.

### 5. DefaultSettings Platform Awareness (Broken Test)

**Files:** `internal/domain/operation_defaults.go`, `internal/cleaner/systemcache.go`
**Problem:** `DefaultSettings(OperationTypeSystemCache)` returns `CacheTypeSpotlight` on line 73-78 regardless of platform. On Linux, `ValidateSettings` then rejects these as "not supported on current platform." The defaults are wrong for non-macOS.
**Solution:** Make `DefaultSettings` platform-aware for `OperationTypeSystemCache` — call `AvailableSystemCacheTypes()` (already exists in cleaner package) or pass platform context. Alternatively, defer cache type validation to runtime.
**Benefits:** Locality — defaults are correct per platform. Depth — callers don't need to know about platform differences.

### 6. Remove Dead Code (Unused Functions/Constants)

**Files:** `internal/config/config.go:52`, `internal/config/sanitizer_nix.go:9`, `internal/adapters/nix.go:28`
**Problem:** `readConfigFile`, `nixGenerationsMin`, `bytesPerKB` are defined but never used. The compiler doesn't catch them (they're package-level).
**Solution:** Remove unused functions and constants. Run `staticcheck` or `golangci-lint` regularly.
**Benefits:** Deletion test passes — removing them concentrates nothing. Less cognitive load.

### 7. SizeEstimate Exhaustive Construction

**Files:** `internal/cleaner/golang_cache_cleaner.go:200`
**Problem:** `domain.SizeEstimate` construction missing `Status` field (`exhaustruct` warning). The field defaults to zero value (`SizeEstimateStatusKnown`), which may be incorrect when actual size is unknown.
**Solution:** Always explicitly set `Status` field when constructing `SizeEstimate`. Use `SizeEstimateStatusUnknown` when size cannot be determined.
**Benefits:** Impossible states unrepresentable — the type already encodes this, but constructors bypass it.
