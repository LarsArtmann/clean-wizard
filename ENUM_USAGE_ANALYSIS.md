# Enum Usage Analysis in Cleaners

**Date:** 2025-02-09
**Status:** INCONSISTENT - Major architectural issue identified

## Summary

Analysis of enum usage across all cleaner implementations revealed significant design inconsistencies. Cleaners use local enums that don't align with domain enums, creating a mismatch between configuration/validation and actual cleaner implementation.

## Findings

### Overall Status

| Cleaner                         | Uses Enums   | Issues Found                                                                                                                                                                            |
| ------------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| buildcache.go                   | YES          | Local `BuildToolType` ("gradle", "maven", "sbt") differs from domain `BuildToolType` ("GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA")                                                  |
| cargo.go                        | NO           | None - no enums used                                                                                                                                                                    |
| docker.go                       | **CRITICAL** | Local `DockerPruneMode` ("light", "standard", "aggressive") conflicts with domain `DockerPruneMode` ("ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS")                                |
| golang_cleaner.go               | YES          | Uses local `GoCacheType` - appropriate for Go-specific cache management                                                                                                                 |
| golang_cache_cleaner.go         | YES          | Uses local `GoCacheType` - appropriate for Go-specific cache management                                                                                                                 |
| homebrew.go                     | YES          | Correctly uses `domain.HomebrewModeUnusedOnly` and `domain.HomebrewModeAll`                                                                                                             |
| langversionmanager.go           | NO           | Local `LangVersionManagerType` ("nvm", "pyenv", "rbenv") is subset of domain `VersionManagerType` ("NVM", "PYENV", "GVM", "RBENV", "SDKMAN", "JENV")                                    |
| nix.go                          | YES          | Correctly uses `domain.GenerationStatusCurrent` and `domain.GenerationStatusHistorical`                                                                                                 |
| nodepackages.go                 | **ISSUE**    | Local `NodePackageManagerType` ("npm", "pnpm", "yarn", "bun") conflicts with domain `PackageManagerType` (integer enum with same string values)                                         |
| projectsmanagementautomation.go | NO           | None - no enums used                                                                                                                                                                    |
| systemcache.go                  | **ISSUE**    | Local `SystemCacheType` ("spotlight", "xcode", "cocoapods", "homebrew") is subset of domain `CacheType` ("SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW", "PIP", "NPM", "YARN", "CCACHE") |
| tempfiles.go                    | NO           | None - no enums used                                                                                                                                                                    |

## Critical Issues

### 1. Docker Cleaner - Type Mismatch and Concept Mismatch

**Location:** `internal/cleaner/docker.go`

**Problem:**

- **Domain enum:** Represents WHAT to prune (resource types)
  - `DockerPruneAll` (0) → "ALL" - prune all resources
  - `DockerPruneImages` (1) → "IMAGES" - prune only images
  - `DockerPruneContainers` (2) → "CONTAINERS" - prune only containers
  - `DockerPruneVolumes` (3) → "VOLUMES" - prune only volumes
  - `DockerPruneBuilds` (4) → "BUILDS" - prune only build cache

- **Local enum:** Represents HOW aggressively to prune (intensity level)
  - `DockerPruneLight` ("light") → `docker system prune -f`
  - `DockerPruneStandard` ("standard") → `docker system prune -af`
  - `DockerPruneAggressive` ("aggressive") → `docker system prune -af --volumes`

**Impact:**

- `ValidateSettings` validates domain enum but cleaner uses local enum
- Conversion in integration tests: `cleaner.DockerPruneMode(pruneMode.String())` - doesn't work correctly
- Cleaner ignores the domain enum value completely

**Docker Commands Available:**

- `docker image prune` - prune images
- `docker container prune` - prune containers
- `docker volume prune` - prune volumes
- `docker builder prune` - prune build cache
- `docker system prune` - prune all (containers, networks, images, build cache)
- `docker system prune -a` - prune all including unused images
- `docker system prune -a --volumes` - prune all including volumes

**Recommendation:**
Refactor Docker cleaner to use domain enum and map each enum value to appropriate Docker commands:

- `DockerPruneImages` → `docker image prune -f`
- `DockerPruneContainers` → `docker container prune -f`
- `DockerPruneVolumes` → `docker volume prune -f`
- `DockerPruneBuilds` → `docker builder prune -f`
- `DockerPruneAll` → `docker system prune -af --volumes`

### 2. SystemCache Cleaner - Case Sensitivity and Scope Mismatch

**Location:** `internal/cleaner/systemcache.go`

**Problem:**

- **Local enum:** Lowercase strings ("spotlight", "xcode", "cocoapods", "homebrew")
- **Domain enum:** Uppercase integer representations ("SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW")
- **Scope:** Domain enum includes additional types ("PIP", "NPM", "YARN", "CCACHE")

**Impact:**

- Type mismatch between local and domain enums
- Validation uses domain enum but cleaner uses local enum
- Potential for invalid configurations

**Recommendation:**
Either:

1. Refactor to use domain enum directly (preferred for consistency)
2. Create explicit conversion functions between local and domain enums

### 3. NodePackages Cleaner - Type Mismatch

**Location:** `internal/cleaner/nodepackages.go`

**Problem:**

- **Local enum:** String type ("npm", "pnpm", "yarn", "bun")
- **Domain enum:** Integer type with same string values

**Impact:**

- Type mismatch prevents direct usage
- Requires string-to-int conversion or explicit handling

**Recommendation:**
Refactor to use domain enum directly or create conversion functions.

### 4. BuildCache Cleaner - Complete Type Mismatch

**Location:** `internal/cleaner/buildcache.go`

**Problem:**

- **Local enum:** Build tools ("gradle", "maven", "sbt")
- **Domain enum:** Languages ("GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA")

**Impact:**

- Complete mismatch in what the enums represent
- Domain enum represents languages, local enum represents build tools
- Different abstractions entirely

**Recommendation:**
Clarify the intended abstraction:

- If domain should represent languages, refactor buildcache to align
- If domain should represent build tools, update domain enum values

## Correct Implementations

### Good Examples

1. **Nix Cleaner** (`internal/cleaner/nix.go`)
   - Correctly uses `domain.GenerationStatusCurrent` and `domain.GenerationStatusHistorical`
   - No local enum duplication
   - Clean implementation

2. **Homebrew Cleaner** (`internal/cleaner/homebrew.go`)
   - Correctly uses `domain.HomebrewModeUnusedOnly` and `domain.HomebrewModeAll`
   - No local enum duplication
   - Clean implementation

3. **Go Cache Cleaners** (`internal/cleaner/golang_cache_cleaner.go`, `internal/cleaner/golang_cleaner.go`)
   - Use local `GoCacheType` which is appropriate for Go-specific cache management
   - Not conflicting with any domain enum
   - Reasonable use case for local enum

## Root Cause

The fundamental issue is **layer separation confusion**:

- **Domain layer** should define all shared concepts and enums
- **Cleaner layer** should use domain enums, not define its own
- **Local enums** are only acceptable for cleaner-specific concepts that don't exist in domain

## Refactoring Priority

### High Priority

1. **Docker cleaner** - Complete concept mismatch, affects functionality
2. **SystemCache cleaner** - Case sensitivity and scope issues

### Medium Priority

3. **NodePackages cleaner** - Type mismatch
4. **BuildCache cleaner** - Complete abstraction mismatch

### Low Priority

5. **LangVersionManager cleaner** - Subset issue (less critical)

## Next Steps

1. **Document current behavior** - ✅ Complete (this document)
2. **Design unified enum approach** - Decide on domain enum dominance
3. **Create migration plan** - Plan incremental refactoring
4. **Implement conversion functions** - If keeping both enums temporarily
5. **Refactor high-priority cleaners** - Start with Docker and SystemCache
6. **Update integration tests** - Remove type conversions
7. **Deprecate local enums** - Once domain usage is established

## Testing Implications

- Current integration tests pass because they perform string conversions
- After refactoring, integration tests will need updates
- All enum tests will need to verify correct domain enum usage
- Cleaner validation tests will need to verify domain enum validation

---

**Document created:** 2025-02-09
**Author:** Crush AI Assistant
**Status:** Needs architectural decision and refactoring plan
