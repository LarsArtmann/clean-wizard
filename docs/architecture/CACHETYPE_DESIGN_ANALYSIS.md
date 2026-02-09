# CacheType Design Analysis

**Date**: 2026-02-09
**Status**: Analyzed - Design Intent Clarified

## Summary

The relationship between `domain.CacheType` and `SystemCacheType` has been analyzed. The design is **intentional** - `domain.CacheType` is a comprehensive enum for all cache types, while `SystemCacheType` is a local subset specific to macOS system cache handling.

## Findings

### domain.CacheType (Comprehensive)

**Location**: `internal/domain/operation_settings.go:707`

**Type**: `int`-based enum (type-safe)

**Values** (8 total):
```go
const (
    CacheTypeSpotlight CacheType = iota  // macOS Spotlight
    CacheTypeXcode                        // Xcode DerivedData
    CacheTypeCocoapods                    // CocoaPods cache
    CacheTypeHomebrew                     // Homebrew cache
    CacheTypePip                          // Python pip cache
    CacheTypeNpm                          // Node.js npm cache
    CacheTypeYarn                         // Yarn cache
    CacheTypeCcache                       // ccache
)
```

**Purpose**: Global enum representing ALL cache types that can be configured in the system. Used in:
- `OperationSettings.SystemCache.CacheTypes` - configuration
- YAML/JSON serialization
- Cross-cutting cache type operations

### SystemCacheType (Local Subset)

**Location**: `internal/cleaner/systemcache.go:25`

**Type**: `string`-based enum

**Values** (4 total):
```go
const (
    SystemCacheSpotlight SystemCacheType = "spotlight"
    SystemCacheXcode     SystemCacheType = "xcode"
    SystemCacheCocoaPods SystemCacheType = "cocoapods"
    SystemCacheHomebrew  SystemCacheType = "homebrew"
)
```

**Purpose**: Local enum for `SystemCacheCleaner` which only handles **macOS system caches**. Does NOT handle language-specific caches (pip, npm, yarn, ccache).

## Design Intent

### Why Two Enums?

1. **Separation of Concerns**:
   - `domain.CacheType` - Configuration/serialization layer
   - `SystemCacheType` - Implementation layer for macOS system cache cleaner

2. **Different Responsibilities**:
   - Language-specific caches (pip, npm, yarn, ccache) are handled by other cleaners:
     - `NodePackageManagerCleaner` - handles npm/yarn
     - `LanguageVersionManagerCleaner` - handles pyenv (pip)
     - `BuildCacheCleaner` - handles ccache
   - `SystemCacheCleaner` only handles macOS-specific system caches

3. **Type Safety**:
   - `domain.CacheType` uses int-based enum for performance and type safety
   - `SystemCacheType` uses string-based enum for readability in logs/debugging

## Relationship Map

```
domain.CacheType (comprehensive)
├── SystemCacheType subset (SystemCacheCleaner)
│   ├── CacheTypeSpotlight  → SystemCacheSpotlight
│   ├── CacheTypeXcode      → SystemCacheXcode
│   ├── CacheTypeCocoapods  → SystemCacheCocoaPods
│   └── CacheTypeHomebrew   → SystemCacheHomebrew
├── CacheTypePip            → Handled by LanguageVersionManagerCleaner
├── CacheTypeNpm            → Handled by NodePackageManagerCleaner
├── CacheTypeYarn           → Handled by NodePackageManagerCleaner
└── CacheTypeCcache         → Handled by BuildCacheCleaner
```

## Validation Flow

```
User Config (YAML)
    ↓
domain.CacheType (unmarshal)
    ↓
SystemCacheSettings.Validate()
    ↓
SystemCacheType conversion & validation
    ↓
SystemCacheCleaner execution
```

## Conversion Functions

The `CacheTypeToLowerSlice()` function in `systemcache.go` handles the conversion:

```go
// domain.CacheType → lowercase strings → SystemCacheType validation
CacheTypeSpotlight → "spotlight" → SystemCacheSpotlight
```

## Recommendation

**No refactoring required**. The current design is architecturally sound:

1. ✅ `domain.CacheType` correctly represents all possible cache types
2. ✅ `SystemCacheType` correctly represents only the cache types it handles
3. ✅ Separation prevents SystemCacheCleaner from accidentally handling caches it shouldn't
4. ✅ Each cleaner handles its own domain-specific cache types

## Files Affected

- `internal/domain/operation_settings.go` - domain.CacheType definition
- `internal/cleaner/systemcache.go` - SystemCacheType definition and conversion
- `internal/cleaner/systemcache_test.go` - Test validation

## Related Documentation

- See `ENUM_USAGE_ANALYSIS.md` for broader enum usage patterns
- See `docs/status/2026-02-09_07-48_DOCKER_REFACTOR_COMPLETE.md` for enum consistency notes
