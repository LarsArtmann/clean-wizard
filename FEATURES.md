# Clean Wizard Features

> **Last Updated:** 2026-02-09  
> **Version:** Based on codebase analysis  
> **Status:** BRUTALLY HONEST ASSESSMENT

---

## Overview

Clean Wizard is a system cleanup tool designed to safely remove old files, package caches, and temporary data. It supports multiple package managers, caches, and system components across macOS and Linux.

---

## Feature Status Legend

| Status                      | Meaning                                                     |
| --------------------------- | ----------------------------------------------------------- |
| ✅ **FULLY_FUNCTIONAL**     | Feature is complete, tested, and works as intended          |
| ⚠️ **PARTIALLY_FUNCTIONAL** | Feature works but has limitations or known issues           |
| 🔧 **NEEDS_IMPROVEMENT**    | Feature exists but needs refinement or has technical debt   |
| 🚧 **BROKEN**               | Feature does not work correctly or is incomplete            |
| 📝 **PLANNED**              | Feature is planned but not yet implemented                  |
| 🧪 **MOCKED**               | Feature returns mock/simulated data instead of real results |

---

## Core Cleaners (11 Total)

### 1. Nix Cleaner ❄️

| Aspect                     | Status                  | Details                                                          |
| -------------------------- | ----------------------- | ---------------------------------------------------------------- |
| **Overall**                | ✅ FULLY_FUNCTIONAL     | Core feature, well-tested                                        |
| **Availability Detection** | ✅ Working              | Checks for `nix` command                                         |
| **Generation Listing**     | ⚠️ PARTIALLY_FUNCTIONAL | Returns mock data when Nix unavailable; real data when available |
| **Generation Cleanup**     | ✅ Working              | Removes old generations, keeps current + N others                |
| **Garbage Collection**     | ✅ Working              | Runs `nix-collect-garbage` after cleanup                         |
| **Dry Run Mode**           | ✅ Working              | Estimates 50MB per generation                                    |
| **Size Estimation**        | 🧪 MOCKED               | Uses hardcoded 50MB estimate per generation                      |
| **Configurability**        | ✅ Working              | Configurable keep count (default: 5)                             |

**Notes:**

- Original purpose of the tool
- Most mature cleaner
- Current generation always protected
- Mock data returned in CI/testing environments

---

### 2. Homebrew Cleaner 🍺

| Aspect                       | Status              | Details                                |
| ---------------------------- | ------------------- | -------------------------------------- |
| **Overall**                  | ✅ FULLY_FUNCTIONAL | Well-implemented                       |
| **Availability Detection**   | ✅ Working          | Checks for `brew` command              |
| **Scanning**                 | ✅ Working          | Lists outdated packages                |
| **Cleanup (`brew cleanup`)** | ✅ Working          | Removes cached downloads               |
| **Prune (`brew prune`)**     | ✅ Working          | Removes dead symlinks                  |
| **Dry Run Mode**             | 🚧 BROKEN           | Not supported - prints warning only    |
| **Mode Selection**           | ✅ Working          | Supports `all` and `unused_only` modes |

**Notes:**

- Dry-run explicitly not supported (Homebrew limitation)
- Suggests manual `brew cleanup -n` for preview

---

### 3. Docker Cleaner 🐳

| Aspect                     | Status              | Details                                            |
| -------------------------- | ------------------- | -------------------------------------------------- |
| **Overall**                | ✅ FULLY_FUNCTIONAL | Recently refactored                                |
| **Availability Detection** | ✅ Working          | Checks for `docker` command                        |
| **Scanning**               | ✅ Working          | Scans dangling images, stopped containers, volumes |
| **Prune Modes**            | ✅ Working          | ALL, IMAGES, CONTAINERS, VOLUMES, BUILDS           |
| **System Prune**           | ✅ Working          | `docker system prune -af --volumes`                |
| **Dry Run Mode**           | 🧪 MOCKED           | Returns hardcoded estimates (100MB-2GB)            |
| **Timeout Handling**       | ✅ Working          | 2-minute timeout on operations                     |
| **Size Reporting**         | 🚧 BROKEN           | Real freed bytes not parsed from output            |

**Notes:**

- Size freed always reported as 0 (output parsing not implemented)
- Docker must be running, not just installed

---

### 4. Go Cleaner 🐹

| Aspect                     | Status              | Details                                                   |
| -------------------------- | ------------------- | --------------------------------------------------------- |
| **Overall**                | ✅ FULLY_FUNCTIONAL | Comprehensive implementation                              |
| **Availability Detection** | ✅ Working          | Checks for `go` command                                   |
| **Cache Types**            | ✅ Working          | GOCACHE, GOTESTCACHE, GOMODCACHE, Build Cache, Lint Cache |
| **Cache Cleaning**         | ✅ Working          | Uses `go clean -cache`, `go clean -testcache`, etc.       |
| **Lint Cache**             | ✅ Working          | Cleans `golangci-lint` cache                              |
| **Scanning**               | ✅ Working          | Scans all Go cache locations                              |
| **Dry Run Mode**           | 🧪 MOCKED           | Estimates 200MB                                           |
| **Type Safety**            | ✅ Working          | Bit-flag based cache type selection                       |

**Notes:**

- Most sophisticated cleaner with type-safe configuration
- Supports selective cache cleaning via bit flags
- Golangci-lint cache cleaning as bonus feature

---

### 5. Cargo Cleaner (Rust) 🦀

| Aspect                     | Status                  | Details                                         |
| -------------------------- | ----------------------- | ----------------------------------------------- |
| **Overall**                | ⚠️ PARTIALLY_FUNCTIONAL | Basic implementation                            |
| **Availability Detection** | ✅ Working              | Checks for `cargo` command                      |
| **Scanning**               | ✅ Working              | Scans `~/.cargo/registry` and `~/.cargo/git`    |
| **Standard Clean**         | ✅ Working              | Runs `cargo clean`                              |
| **cargo-cache Tool**       | ✅ Working              | Uses `cargo-cache --autoclean` if available     |
| **Dry Run Mode**           | 🧪 MOCKED               | Estimates 500MB                                 |
| **Size Reporting**         | 🚧 BROKEN               | Real freed bytes not tracked                    |
| **Fallback Logic**         | ✅ Working              | Falls back to manual clean if cargo-cache fails |

**Notes:**

- Only cleans local project, not global cache
- cargo-cache tool is optional enhancement

---

### 6. Node Package Manager Cleaner 📦

| Aspect                     | Status              | Details                     |
| -------------------------- | ------------------- | --------------------------- |
| **Overall**                | ✅ FULLY_FUNCTIONAL | Multi-PM support            |
| **Package Managers**       | ✅ Working          | npm, pnpm, yarn, bun        |
| **Availability Detection** | ✅ Working          | Checks each PM individually |
| **npm Cache Clean**        | ✅ Working          | `npm cache clean --force`   |
| **pnpm Store Prune**       | ✅ Working          | `pnpm store prune`          |
| **Yarn Cache Clean**       | ✅ Working          | `yarn cache clean`          |
| **Bun Cache Clean**        | ✅ Working          | `bun pm cache rm`           |
| **Scanning**               | ✅ Working          | Discovers cache locations   |
| **Dry Run Mode**           | 🧪 MOCKED           | Estimates 100MB per PM      |

**Notes:**

- Gracefully handles unavailable package managers
- Cache paths discovered dynamically where possible

---

### 7. Build Cache Cleaner 🔨

| Aspect                  | Status                  | Details                                                  |
| ----------------------- | ----------------------- | -------------------------------------------------------- |
| **Overall**             | ⚠️ PARTIALLY_FUNCTIONAL | Limited tool coverage                                    |
| **Availability**        | ✅ Working              | Always available (file-based)                            |
| **Gradle Support**      | ✅ Working              | Cleans `~/.gradle/caches`                                |
| **Maven Support**       | ✅ Working              | Removes `~/.m2/repository/**/*.part` files               |
| **SBT Support**         | ✅ Working              | Cleans `~/.ivy2/cache`                                   |
| **Age-Based Filtering** | ✅ Working              | Configurable `older_than` duration                       |
| **Dry Run Mode**        | ✅ Working              | Correctly previews actions                               |
| **Other Build Tools**   | 📝 PLANNED              | Go, Rust, Node, Python exist in enum but NOT implemented |

**Notes:**

- Domain enum has 6 build tools (Go, Rust, Node, Python, Java, Scala)
- Only Java (Gradle, Maven) and Scala (SBT) actually implemented
- Other tools listed but silently ignored

---

### 8. System Cache Cleaner ⚙️

| Aspect                   | Status                  | Details                                                     |
| ------------------------ | ----------------------- | ----------------------------------------------------------- |
| **Overall**              | ⚠️ PARTIALLY_FUNCTIONAL | macOS only                                                  |
| **Platform Support**     | 🚧 BROKEN               | macOS only - Linux not supported                            |
| **Availability Check**   | 🔧 NEEDS_IMPROVEMENT    | Checks `GOOS`/`OSTYPE` env vars only                        |
| **Spotlight Cache**      | ✅ Working              | `~/Library/Metadata/CoreSpotlight/SpotlightKnowledgeEvents` |
| **Xcode DerivedData**    | ✅ Working              | `~/Library/Developer/Xcode/DerivedData`                     |
| **CocoaPods Cache**      | ✅ Working              | `~/Library/Caches/CocoaPods`                                |
| **Homebrew Cache**       | ✅ Working              | `~/Library/Caches/Homebrew`                                 |
| **Age-Based Filtering**  | ✅ Working              | Configurable `older_than` duration                          |
| **Dry Run Mode**         | ✅ Working              | Correctly previews actions                                  |
| **Extended Cache Types** | 📝 PLANNED              | Pip, npm, yarn, ccache exist in enum but NOT implemented    |

**Notes:**

- Domain enum has 8 cache types
- Only 4 actually implemented in cleaner
- Platform detection is fragile (env vars vs runtime check)

---

### 9. Temporary Files Cleaner 🗂️

| Aspect                  | Status              | Details                                |
| ----------------------- | ------------------- | -------------------------------------- |
| **Overall**             | ✅ FULLY_FUNCTIONAL | Robust implementation                  |
| **Availability**        | ✅ Working          | Always available                       |
| **Age-Based Filtering** | ✅ Working          | Configurable `older_than` duration     |
| **Path Configuration**  | ✅ Working          | Custom base paths supported            |
| **Exclusion Patterns**  | ✅ Working          | Prefix-based exclusions                |
| **Recursive Scanning**  | ✅ Working          | Full directory tree walk               |
| **File-Only Cleanup**   | ✅ Working          | Directories preserved, files removed   |
| **Dry Run Mode**        | ✅ Working          | Correctly previews with accurate sizes |
| **Size Calculation**    | ✅ Working          | Real file sizes in dry-run             |

**Notes:**

- Only removes files, never directories (safety feature)
- Uses `filepath.Walk` for traversal
- Respects exclusion patterns

---

### 10. Language Version Manager Cleaner 🗑️

| Aspect              | Status     | Details                                             |
| ------------------- | ---------- | --------------------------------------------------- |
| **Overall**         | 🚧 BROKEN  | NO-OP implementation                                |
| **Availability**    | ✅ Working | Always available (file-based)                       |
| **Scanning**        | ✅ Working | Scans NVM, Pyenv, Rbenv directories                 |
| **NVM Support**     | 🚧 BROKEN  | Scans only, does NOT clean                          |
| **Pyenv Support**   | 🚧 BROKEN  | Scans only, does NOT clean                          |
| **Rbenv Support**   | 🚧 BROKEN  | Scans only, does NOT clean                          |
| **Clean Operation** | 🚧 BROKEN  | NO-OP - prints warning only                         |
| **Domain Enum**     | 📝 PLANNED | GVM, SDKMAN, Jenv exist in enum but NOT implemented |

**Notes:**

- **CRITICAL:** Cleaner explicitly does NOTHING on clean
- Comment in code: "This is a NO-OP by default to avoid destructive behavior"
- Destructive nature acknowledged but not addressed
- Essentially a placeholder cleaner

---

### 11. Projects Management Automation Cleaner ⚙️

| Aspect                     | Status     | Details                                             |
| -------------------------- | ---------- | --------------------------------------------------- |
| **Overall**                | 🚧 BROKEN  | Requires external tool                              |
| **Availability Detection** | ✅ Working | Checks for `projects-management-automation` command |
| **Scanning**               | 🧪 MOCKED  | Returns hardcoded path estimate                     |
| **Cache Clearing**         | 🚧 BROKEN  | Only works if external tool installed               |
| **Dry Run Mode**           | 🧪 MOCKED  | Estimates 100MB                                     |
| **Size Estimation**        | 🧪 MOCKED  | Hardcoded 100MB estimate                            |

**Notes:**

- Requires separate `projects-management-automation` CLI tool
- Unlikely to be available on most systems
- Effectively non-functional for typical users

---

## CLI Features

### Command Structure

| Command                | Status              | Description                    |
| ---------------------- | ------------------- | ------------------------------ |
| `clean-wizard clean`   | ✅ FULLY_FUNCTIONAL | Main cleanup command with TUI  |
| `clean-wizard scan`    | 📝 PLANNED          | Documented but NOT implemented |
| `clean-wizard init`    | 📝 PLANNED          | Documented but NOT implemented |
| `clean-wizard profile` | 📝 PLANNED          | Documented but NOT implemented |
| `clean-wizard config`  | 📝 PLANNED          | Documented but NOT implemented |

**Notes:**

- Only `clean` command is actually implemented
- Other commands documented in USAGE.md but return "unknown command"
- Significant documentation/implementation gap

### Clean Command Features

| Feature                    | Status              | Details                       |
| -------------------------- | ------------------- | ----------------------------- |
| **Interactive TUI**        | ✅ FULLY_FUNCTIONAL | Beautiful Charm Huh forms     |
| **Multi-Select**           | ✅ FULLY_FUNCTIONAL | Select multiple cleaners      |
| **Availability Detection** | ✅ FULLY_FUNCTIONAL | Shows only available cleaners |
| **Dry Run Mode**           | ✅ FULLY_FUNCTIONAL | `--dry-run` flag works        |
| **Verbose Mode**           | ✅ FULLY_FUNCTIONAL | `--verbose` flag works        |
| **JSON Output**            | ✅ FULLY_FUNCTIONAL | `--json` flag works           |
| **Preset Modes**           | ✅ FULLY_FUNCTIONAL | quick, standard, aggressive   |
| **Confirmation Prompt**    | ✅ FULLY_FUNCTIONAL | Yes/No before execution       |
| **Result Aggregation**     | ✅ FULLY_FUNCTIONAL | Totals across all cleaners    |
| **Progress Display**       | ✅ FULLY_FUNCTIONAL | Per-cleaner progress          |
| **Encouraging Messages**   | ✅ FULLY_FUNCTIONAL | Celebrates >1GB freed         |

### Preset Modes

| Mode           | Cleaners Included                         | Status     |
| -------------- | ----------------------------------------- | ---------- |
| **quick**      | Homebrew, Node, Go, TempFiles, BuildCache | ✅ Working |
| **standard**   | All available cleaners                    | ✅ Working |
| **aggressive** | All available cleaners                    | ✅ Working |

**Note:** Standard and aggressive are currently identical (both use all cleaners).

---

## Configuration System

| Feature                 | Status               | Details                             |
| ----------------------- | -------------------- | ----------------------------------- |
| **YAML Configuration**  | ✅ FULLY_FUNCTIONAL  | Full schema support                 |
| **Profile System**      | ✅ FULLY_FUNCTIONAL  | Multiple profiles supported         |
| **Operation Settings**  | ✅ FULLY_FUNCTIONAL  | Type-safe per-cleaner settings      |
| **Enum Type Safety**    | ✅ FULLY_FUNCTIONAL  | Compile-time enum safety            |
| **Validation**          | ✅ FULLY_FUNCTIONAL  | Comprehensive validation rules      |
| **Default Settings**    | ✅ FULLY_FUNCTIONAL  | Sensible defaults for all cleaners  |
| **Config File Loading** | 🔧 NEEDS_IMPROVEMENT | CLI flags exist but not fully wired |
| **Hot Reload**          | 📝 PLANNED           | Not implemented                     |

### Configuration Enums (Type-Safe)

| Enum                   | Values                                                        | Status                                 |
| ---------------------- | ------------------------------------------------------------- | -------------------------------------- |
| **CacheCleanupMode**   | DISABLED, ENABLED                                             | ✅ Working                             |
| **DockerPruneMode**    | ALL, IMAGES, CONTAINERS, VOLUMES, BUILDS                      | ✅ Working                             |
| **BuildToolType**      | GO, RUST, NODE, PYTHON, JAVA, SCALA                           | ⚠️ Partial (only JAVA/SCALA used)      |
| **CacheType**          | SPOTLIGHT, XCODE, COCOAPODS, HOMEBREW, PIP, NPM, YARN, CCACHE | ⚠️ Partial (only first 4 used)         |
| **VersionManagerType** | NVM, PYENV, GVM, RBENV, SDKMAN, JENV                          | ⚠️ Partial (only NVM/PYENV/RBENV used) |
| **PackageManagerType** | NPM, PNPM, YARN, BUN                                          | ✅ Working                             |
| **RiskLevel**          | LOW, MEDIUM, HIGH, CRITICAL                                   | ✅ Working                             |
| **ValidationLevel**    | NONE, BASIC, COMPREHENSIVE, STRICT                            | ✅ Working                             |
| **CleanStrategy**      | AGGRESSIVE, CONSERVATIVE, DRY_RUN                             | ✅ Working                             |
| **HomebrewMode**       | UNUSED_ONLY, ALL                                              | ✅ Working                             |
| **OptimizationMode**   | DISABLED, ENABLED                                             | ✅ Working                             |
| **ExecutionMode**      | NORMAL, DRY_RUN                                               | ✅ Working                             |

---

## Testing & Quality

| Aspect                   | Status       | Details                        |
| ------------------------ | ------------ | ------------------------------ |
| **Unit Tests**           | ✅ EXTENSIVE | 200+ tests across packages     |
| **BDD Tests**            | ✅ WORKING   | Godog-based BDD scenarios      |
| **Integration Tests**    | ✅ WORKING   | Real cleaner integration tests |
| **Fuzz Tests**           | ✅ WORKING   | Multiple fuzzing targets       |
| **Benchmark Tests**      | ✅ WORKING   | Performance benchmarks         |
| **Test Coverage**        | ⚠️ MODERATE  | Good but not comprehensive     |
| **Mock Implementations** | ✅ WORKING   | Mock data for CI environments  |

---

## Architecture Highlights

| Pattern                  | Status               | Details                                    |
| ------------------------ | -------------------- | ------------------------------------------ |
| **Registry Pattern**     | ✅ FULLY_FUNCTIONAL  | Clean registry for all cleaners            |
| **Factory Functions**    | ✅ FULLY_FUNCTIONAL  | DefaultRegistry, DefaultRegistryWithConfig |
| **Result Type**          | ✅ FULLY_FUNCTIONAL  | Generic result.Result[T] type              |
| **Adapter Pattern**      | ✅ FULLY_FUNCTIONAL  | External tool adapters (Nix, etc.)         |
| **Middleware**           | ✅ FULLY_FUNCTIONAL  | Validation middleware                      |
| **Type-Safe Enums**      | ✅ FULLY_FUNCTIONAL  | Compile-time enum safety                   |
| **Dependency Injection** | 🔧 NEEDS_IMPROVEMENT | Some hardcoded dependencies                |

---

## Known Issues & Limitations

### Critical Issues

1. **Language Version Manager Cleaner is NO-OP** 🚧
   - Scans but never cleans
   - Documented as "destructive" but no actual implementation

2. **Projects Management Automation requires external tool** 🚧
   - Depends on tool most users won't have
   - Effectively non-functional

3. **Most CLI commands not implemented** 🚧
   - Only `clean` works
   - `scan`, `init`, `profile`, `config` documented but missing

### Minor Issues

4. **Dry-run uses hardcoded estimates** 🧪
   - Most cleaners estimate rather than calculate
   - Inconsistent estimation sizes

5. **Size reporting often returns 0** 🧪
   - Docker cleaner doesn't parse output
   - Several cleaners don't track actual bytes freed

6. **Platform detection is fragile** 🔧
   - SystemCache uses env vars instead of runtime detection
   - May fail in containers or unusual environments

7. **Enum/Implementation mismatch** 🔧
   - Several enums have values not used in implementations
   - Dead code in domain layer

---

## Feature Matrix Summary

| Cleaner          | Available | Scan | Clean | Dry-Run | Size Accurate | Status              |
| ---------------- | --------- | ---- | ----- | ------- | ------------- | ------------------- |
| Nix              | ✅        | ✅   | ✅    | 🧪      | 🧪            | ✅ Production Ready |
| Homebrew         | ✅        | ✅   | ✅    | 🚧      | 🧪            | ✅ Production Ready |
| Docker           | ✅        | ✅   | ✅    | 🧪      | 🚧            | ✅ Production Ready |
| Go               | ✅        | ✅   | ✅    | 🧪      | ⚠️            | ✅ Production Ready |
| Cargo            | ✅        | ✅   | ✅    | 🧪      | 🚧            | ⚠️ Basic            |
| Node Packages    | ✅        | ✅   | ✅    | 🧪      | 🧪            | ✅ Production Ready |
| Build Cache      | ✅        | ✅   | ✅    | ✅      | ✅            | ⚠️ Limited Tools    |
| System Cache     | ⚠️        | ✅   | ✅    | ✅      | ✅            | ⚠️ macOS Only       |
| Temp Files       | ✅        | ✅   | ✅    | ✅      | ✅            | ✅ Production Ready |
| Lang Version Mgr | ✅        | ✅   | 🚧    | 🚧      | N/A           | 🚧 Non-Functional   |
| Projects Mgmt    | 🚧        | 🧪   | 🚧    | 🧪      | 🧪            | 🚧 Non-Functional   |

---

## Recommendations

### For Users

1. **Use with confidence:** Nix, Homebrew, Docker, Go, Node, Temp Files cleaners
2. **Use with caution:** Build Cache (limited), System Cache (macOS only)
3. **Don't rely on:** Language Version Manager, Projects Management Automation

### For Contributors

1. **Priority 1:** Implement actual cleaning for Language Version Manager
2. **Priority 2:** Add remaining CLI commands (scan, init, profile, config)
3. **Priority 3:** Improve size reporting across all cleaners
4. **Priority 4:** Implement remaining enum values (BuildToolType, CacheType, VersionManagerType)
5. **Priority 5:** Add Linux support for System Cache cleaner

---

## Conclusion

Clean Wizard has a **solid foundation** with excellent architecture and type safety. The core cleaners (Nix, Homebrew, Docker, Go, Node, Temp Files) are production-ready and work well.

However, there are **significant gaps**:

- ~20% of cleaners are non-functional or placeholders
- ~80% of documented CLI commands don't exist
- Several enums have unused values (dead code)

**Overall Project Status:** ⚠️ **PARTIALLY_FUNCTIONAL** - Core features work well, but peripheral features need significant work.

---

_This assessment was generated by thorough code analysis. For questions or corrections, please open an issue._
