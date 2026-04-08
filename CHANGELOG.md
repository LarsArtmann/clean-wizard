# CHANGELOG

**Last Updated:** 2026-04-05

---

## [Unreleased]

### Added

#### 2026-04-03

- Enum Consolidation Refactor: Consolidated all 19 iota-based enum types across 4 files onto unified `enum_macros.go` helpers (52% line reduction)
- All enums now use `EnumString`, `EnumIsValid`, `EnumValues`, `EnumMarshalJSON`, `EnumUnmarshalJSON`, `EnumMarshalYAML`, `EnumUnmarshalYAML`
- YAML marshaling now returns strings instead of ints

#### 2026-04-02

- Unit tests for cleanerMetadata (`cleaner_types_test.go` - 4 tests)
- Init() validation for operationTypeToCleanerType entries

### Changed

- Error messages simplified to consistent format
- Git History dry-run default changed from true to false

### Removed

- Dead `UnmarshalYAMLEnum`, `UnmarshalJSONEnum`, `UnmarshalYAMLEnumWithDefault` helpers
- `TypeSafeEnum` interface
- Langversion cleaner stub (CleanerTypeLangVersionMgr)
- 49 deprecation warnings across 45+ files

### Fixed

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
- Error Details Utility (`internal/pkg/errors/detail_helpers.go`)
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
