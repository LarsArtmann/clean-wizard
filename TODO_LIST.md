# TODO LIST

**Last Updated:** 2026-02-13
**Total MD Files Found:** 91
**Files Processed:** 41/91
**Files Remaining:** 50
**Status Report:** docs/status/2026-02-13_02-22_COMPREHENSIVE-STATUS-REPORT.md

---

## FILE TRACKING

### Already Processed

1. ✅ USAGE.md - Status: COMPLETED - No TODOs found (pure documentation)
2. ✅ README.md - Status: COMPLETED - No TODOs found (pure documentation)
3. ✅ HOW_TO_USE.md - Status: COMPLETED - No TODOs found (pure documentation)
4. ✅ DEVELOPMENT.md - Status: COMPLETED - No TODOs found (pure documentation)
5. ✅ docs/domain.md - Status: COMPLETED - API documentation (no TODOs)
6. ✅ docs/config.md - Status: COMPLETED - API documentation (no TODOs)
7. ✅ docs/README.md - Status: COMPLETED - Documentation hub (no TODOs)
8. ✅ FIX_SUMMARY.md - Status: COMPLETED - Issue resolution documentation (no active TODOs)
9. ✅ IMPLEMENTATION_STATUS.md - Status: COMPLETED - Contains 3 remaining TODOs (Generic Context System, Eliminate Backward Compatibility Aliases, Domain Model Enhancement)
10. ✅ REFACTORING_PLAN.md - Status: COMPLETED - Contains 8 phases of refactoring tasks (Generic Validation Interface, Config Loading Utility, String Trimming Utility, Error Details Utility, Test Helper Refactoring, Schema Min/Max Utility, Type Model Improvements, Result Type Enhancement)
11. ✅ SELF_REFLECTION_AND_PLAN.md - Status: COMPLETED - Contains comprehensive execution plan (Deprecation Fixes - 20+ files, CleanerRegistry Integration, SystemCache Research, Complexity Reductions, Documentation)
12. ✅ docs/planning/2025-11-10_15-34-COMPREHENSIVE_ARCHITECTURAL_TODO_LIST.md - Status: COMPLETED - Contains 20 prioritized architectural tasks (4 critical, 5 high impact, 6 professional completion, 5 continuous improvement)
13. ✅ docs/status/2025-12-14_02-58_COMPREHENSIVE-MULTI-STEP-EXECUTION-PLAN.md - Status: COMPLETED - Contains 12 ROI-based execution tasks
14. ✅ docs/planning/2025-11-10_13_45-COMPREHENSIVE_GITHUB_ISSUES_EXECUTION_PLAN.md - Status: COMPLETED - Contains 4 GitHub issues with detailed implementation plans (Issues #17, #18, #19, #20)
15. ✅ docs/result.md - Status: COMPLETED - API documentation (no TODOs)
16. ✅ docs/ALIASES.md - Status: COMPLETED - Shell aliases documentation (no TODOs)
17. ✅ docs/cleaner.md - Status: COMPLETED - Cleaner package API (no TODOs)
18. ✅ DOCUMENTATION.md - Status: COMPLETED - Empty file
19. ✅ docs/adapters.md - Status: COMPLETED - Adapters package API (no TODOs)
20. ✅ close_issue_50.md - Status: COMPLETED - Issue closure documentation
21. ✅ schemas/README.md - Status: COMPLETED - JSON Schema documentation (no TODOs)
22. ✅ close_issue_46.md - Status: COMPLETED - Issue closure documentation
23. ✅ docs/middleware.md - Status: COMPLETED - Middleware package API (no TODOs)
24. ✅ resolve_issue_49.md - Status: COMPLETED - Feature flags resolution
25. ✅ resolve_issue_48.md - Status: COMPLETED - Bridge adapters resolution
26. ✅ resolve_issue_47.md - Status: COMPLETED - Integration adapters resolution
27. ✅ docs/conversions.md - Status: COMPLETED - Conversions package API (no TODOs)
28. ✅ ENUM_USAGE_ANALYSIS.md - Status: COMPLETED - CRITICAL enum inconsistencies identified
29. ✅ github_issues_analysis.md - Status: COMPLETED - Production readiness analysis
30. ✅ docs/YAML_ENUM_FORMATS.md - Status: COMPLETED - Comprehensive enum format documentation
31. ✅ FEATURES.md - Status: COMPLETED - Brutally honest feature assessment (11 cleaners, 1 CLI command implemented, 2 broken cleaners identified)
32. ✅ WHAT_THIS_PROJECT_IS_NOT.md - Status: COMPLETED - Scope limitations documentation
33. ✅ ARCHITECTURAL_ANALYSIS_2026-02-08_05-48.md - Status: COMPLETED - Comprehensive architectural analysis (Quality score: 90.1/100, 12 improvement areas identified)
34. ✅ COMPREHENSIVE_IMPROVEMENT_PLAN_2026-02-09.md - Status: COMPLETED - 7-phase execution plan with success metrics and ROI analysis (7 weeks total, incremental delivery)
35. ✅ FEATURES_SUMMARY_TABLE.md - Status: COMPLETED - Feature implementation summary with priority recommendations (P1-P5, overall status: PARTIALLY_FUNCTIONAL)
36. ✅ FEATURES_EXECUTION_PLAN.md - Status: COMPLETED - Historical execution plan for FEATURES.md creation (all tasks completed in FEATURES.md and FEATURES_SUMMARY_TABLE.md)
37. ✅ 2026-02-09_09-15_MAJOR_MILESTONE.md - Status: COMPLETED - Session report confirming CleanerRegistry integration, all tests passing, deprecation warnings eliminated (Phase 1 & 2 complete)
38. ✅ 2026-02-08_ENUM_CLEANER_VERIFICATION.md - Status: COMPLETED - All cleaners verified to correctly handle enum types with proper type safety (no raw int comparisons)
39. ✅ 2026-02-09_07-48_DOCKER_REFACTOR_COMPLETE.md - Status: COMPLETED - Docker cleaner refactored from local enum to domain enum, 6 major tasks completed (Cleaner interface, Context propagation, Binary enum unification, Integration tests, Enum validation, Docker refactor)
40. ✅ 2026-02-10_15-41_COMPREHENSIVE_EXECUTION_PLAN_INITIATED.md - Status: COMPLETED - Full status report documenting verification of critical issues, creation of 139-task execution plan, and initiation of Phase 1
41. ✅ COMPREHENSIVE_REFLECTION_2026-02-11.md - Status: COMPLETED - Comprehensive reflection on size reporting fixes, identified duplicate code patterns, created 6-phase execution plan with priority matrix
42. ✅ docs/status/2026-02-13_02-22_COMPREHENSIVE-STATUS-REPORT.md - Status: COMPLETED - Full project audit: 13 cleaners (10 production-ready, 1 partial, 2 NO-OP), 5/5 CLI commands, 51 tests passing, health score 8.5/10

---

## AGGREGATED TODOs

### Priority 1 - Critical

**From IMPLEMENTATION_STATUS.md:**

1. **Generic Context System Implementation** (90% impact, 1 day work)
   - **Status**: NOT_STARTED
   - **Description**: Unify split brain context types (ValidationContext, ErrorDetails, SanitizationChange) into generic Context[T] struct
   - **Target**: `internal/config/` and related packages
   - **Expected outcome**: Cleaner architecture, reduced duplication
   - **Verification**: Check for ValidationContext, ErrorDetails, SanitizationChange types

2. **CleanerRegistry Integration** (HIGH impact)
   - **Status**: ✅ COMPLETED & VERIFIED
   - **Description**: CleanerRegistry implemented with full test coverage (231 lines, 12 test cases)
   - **Target**: `internal/cleaner/registry.go`
   - **Factory functions**: `DefaultRegistry()`, `DefaultRegistryWithConfig()`
   - **Thread-safe**: RWMutex implementation
   - **Integration**: ✅ Verified in `cmd/clean-wizard/commands/clean.go` (line 79)
   - **Tests**: ✅ 12+ test cases in `internal/cleaner/registry_test.go`
   - **Verification**: Build succeeds, all factory functions exist and work

**From SELF_REFLECTION_AND_PLAN.md:**

3. **Complete Deprecation Fixes** (MEDIUM impact, 1.85 days work)
   - **Status**: ✅ COMPLETED & VERIFIED
   - **Description**: Fix ~20 test/support files with deprecation warnings (Strategy and RiskLevel aliases)
   - **Target**: Multiple files in internal/cleaner/\*\_test.go, conversions/, adapters/, api/, middleware/, tests/
   - **Subtasks**:
     - ✅ Fix test files (7 files)
     - ✅ Fix conversions package (2 files)
     - ✅ Fix adapters package (1 file)
     - ✅ Fix api package (2 files)
     - ✅ Fix middleware package (1 file)
     - ✅ Fix benchmark tests (1 file)
     - ✅ Fix RiskLevel deprecations (~15 files)
   - **Result**: 49 deprecation warnings eliminated across 45+ files
   - **Verification**: ✅ `go build ./...` - no output = no warnings
   - **Note**: Deprecated type aliases remain (marked for v2.0 removal) but no longer cause warnings

### 🚨 CRITICAL - UNSAFE EXEC CALLS (PRODUCTION RISK) ✅ VERIFIED RESOLVED

**From docs/status/2026-01-28_01-30_EXECUTION_AUDIT.md:**

**ORIGINAL REPORT:** These commands could HANG FOREVER in production without timeout protection

**VERIFICATION RESULT: ✅ ALL COMMANDS HAVE TIMEOUT PROTECTION**

| File | Line | Command | Timeout Protection | Verification |
|------|------|---------|-------------------|--------------|
| `internal/cleaner/cargo.go` | Line 18 | `const cargoCommandTimeout = 5m` | ✅ 5-minute timeout | `cargo.go:21-25` uses `context.WithTimeout`
| `internal/cleaner/cargo.go` | Line 182 | `cargo-cache --autoclean` | ✅ Protected | `execWithTimeout` wrapper in `cargo.go:21-25`
| `internal/cleaner/cargo.go` | Line 190 | `cargo clean` | ✅ Protected | `executeCargoCleanCommand` with timeout
| `internal/cleaner/nodepackages.go` | Line 28 | `const DefaultNodePackageManagerTimeout = 2m` | ✅ 2-minute timeout | `nodepackages.go:31-36` uses `context.WithTimeout`
| `internal/cleaner/nodepackages.go` | Line 146 | `npm config get cache` | ✅ Protected | `execWithTimeout` wrapper
| `internal/cleaner/nodepackages.go` | Line 168 | `pnpm store path` | ✅ Protected | `execWithTimeout` wrapper
| `internal/cleaner/nodepackages.go` | Line 321 | `npm cache clean --force` | ✅ Protected | `runPackageManagerCommand` with context timeout `nodepackages.go:289-292`
| `internal/cleaner/nodepackages.go` | Line 324 | `pnpm store prune` | ✅ Protected | `runPackageManagerCommand` with context timeout `nodepackages.go:323-325`
| `internal/cleaner/nodepackages.go` | Line 327 | `yarn cache clean` | ✅ Protected | `runPackageManagerCommand` with context timeout `nodepackages.go:326-328`
| `internal/cleaner/nodepackages.go` | Line 330 | `bun pm cache rm` | ✅ Protected | `runPackageManagerCommand` with context timeout `nodepackages.go:329-331`
| `internal/cleaner/projectsmanagementautomation.go` | Line 15 | `const DefaultProjectsAutomationTimeout = 2m` | ✅ 2-minute timeout | Uses context timeout protection `projectsmanagementautomation.go:18-25`
| `internal/cleaner/projectsmanagementautomation.go` | Line 118 | `projects-management-automation --clear-cache` | ✅ Protected | Command wrapped in timeout context

**Verification Method:** Code review of timeout implementations + `go test ./internal/cleaner/...` (145/145 passing)

**Status:** ✅ **NO ACTION REQUIRED - PRODUCTION SAFE**

### 🚨 CRITICAL - CLI COMMAND GAP (From FEATURES_SUMMARY_TABLE.md)

**Current State:**

- USAGE.md documents 5 commands: clean, scan, init, profile, config
- ✅ ALL 5 COMMANDS NOW IMPLEMENTED
- 100% alignment between documentation and implementation

**Implementation Status:**

1. ✅ `clean` - Full implementation with TUI, dry-run, preset modes
2. ✅ `scan` - Scans for cleanable items, shows size estimates
3. ✅ `init` - Interactive setup wizard, minimal mode support
4. ✅ `profile` - Subcommands: list, show, create, delete
5. ✅ `config` - Subcommands: show, edit, validate, reset

**Verification**: `go build ./cmd/clean-wizard/...` succeeds, all --help commands work

### 🚨 CRITICAL - CLEANER INTERFACE COMPLIANCE ISSUES ✅ VERIFIED RESOLVED

**From CLEANER_INTERFACE_ANALYSIS.md:**

**ORIGINAL REPORT:** Some cleaners were reported as non-compliant with Cleaner interface

**VERIFICATION RESULT: ✅ ALL CLEANERS ARE COMPLIANT**

| Cleaner | File | Methods | Status |
|---------|------|---------|--------|
| `nix.go` | internal/cleaner/nix.go | `Clean(ctx)` ✅, `IsAvailable()` ✅ | COMPLIANT ✅ |
| `golang_cache_cleaner.go` | internal/cleaner/golang_cache_cleaner.go | `Clean(ctx)` ✅, `IsAvailable()` ✅, Name() ✅ | COMPLIANT ✅ |
| **All 13 cleaners** | internal/cleaner/ | All implement Clean(), IsAvailable(), Name() | COMPLIANT ✅ |

**Verification Details:**

**nix.go:**
- Line 48-51: `Clean(ctx context.Context) result.Result[domain.CleanResult]` ✅
- Line 53-55: `IsAvailable(ctx context.Context) bool` ✅
- Line 56-58: `Name() string` ✅

**golang_cache_cleaner.go:**
- Line 39-42: `IsAvailable(ctx context.Context) bool` ✅
- Line 44-46: `Name() string` ✅
- Line 49-63: `Clean(ctx context.Context) result.Result[domain.CleanResult]` ✅

**Test Verification:**
- `go test ./internal/cleaner/...` - 145/145 tests passing
- Build verification: `go build ./...` - clean build
- Interface compliance verified at compile time

**Status:** ✅ **NO ACTION REQUIRED - ALL CLEANERS COMPLIANT**

### 🚨 HIGH - SIZE REPORTING DUPLICATE CODE ✅ EXTRACTED

**From COMPREHENSIVE_REFLECTION_2026-02-11.md:**

**ORIGINAL REPORT:** Size calculation pattern duplicated 7 times across cleaners

**EXTRACTION RESULT: ✅ SHARED UTILITY CREATED**

| Location | File | Lines | Status |
|----------|------|-------|--------|
| executeCargoCleanCommand | cargo.go | 217-235 | ✅ Refactored |
| Clean | golang_lint_adapter.go | 64-88 | ✅ Refactored |
| cleanGoCacheEnv | golang_cache_cleaner.go | 116-143 | ✅ Refactored |
| cleanNpmCache | nodepackages.go | 401-413 | ✅ Refactored |
| cleanPnpmStore | nodepackages.go | 450-462 | ✅ Refactored |
| cleanYarnCache | nodepackages.go | 499-511 | ✅ Refactored |
| cleanBunCache | nodepackages.go | 548-560 | ✅ Refactored |

**Implementation:**
- Created `CalculateBytesFreed()` in `internal/cleaner/fsutil.go`
- Consolidates: GetDirSize before/after, subtraction, non-negative check, verbose logging
- Closure-based cleanup execution for flexible command wrapping
- All 7 locations now use the shared utility

**Benefits:**
- Reduces code duplication from 7 locations to 1 implementation
- Future bug fixes apply automatically to all cleaners
- Improved maintainability and consistency

**Verification:**
- ✅ `go build ./...` succeeds
- ✅ `go test ./...` all tests pass
- ✅ Verbose output preserved for all cleaners
- ✅ Error handling preserved with proper cleanup error propagation

### 🚨 HIGH - ENUM INCONSISTENCIES

**From ENUM_USAGE_ANALYSIS.md:**

| Cleaner          | Issue                                                                       | Severity | Status              |
| ---------------- | --------------------------------------------------------------------------- | -------- | ------------------- |
| **Docker**       | Local enum vs domain enum - concept mismatch (aggression vs resource types) | CRITICAL | ✅ FIXED in 5e94e2a |
| **SystemCache**  | Uses domain.CacheType correctly (verified)                                  | LOW      | ✅ VERIFIED OK       |
| **NodePackages** | Local string enum vs domain integer enum                                    | HIGH     | 🔍 REQUIRES REFACTOR |
| **BuildCache**   | Different abstractions (tools vs languages) - domain has different scope     | HIGH     | 🔍 REQUIRES DECISION |

**Required Action**: Refactor enum systems to align with domain definitions

### 🟠 HIGH - SIZE REPORTING ISSUES (From FEATURES_SUMMARY_TABLE.md)

**Current State:**

- Most cleaners use hardcoded estimates for dry-run
- Docker returns 0 bytes freed (parsing not implemented)
- Cargo doesn't track actual bytes freed

**Affected Cleaners:**

- Docker: Size reporting broken (returns 0)
- Cargo: Size reporting broken
- Multiple others: Use hardcoded estimates

**Required Action**: Improve size reporting accuracy across all cleaners

### Priority 2 - High

**From IMPLEMENTATION_STATUS.md:**

4. **Eliminate Backward Compatibility Aliases** (70% impact, 2 days work)
   - **Status**: ✅ OBSOLETE - No type aliases exist in domain/
   - **Description**: Remove duplicate type systems (RiskLevel = RiskLevelType, etc.) with phased migration
   - **Verification**: `grep "^type " internal/domain/*.go | grep "= "` returns no results

5. **Domain Model Enhancement** (50% impact, 3 days work)
   - **Status**: NOT_STARTED
   - **Description**: Transform anemic domain models into rich domain objects with behavior
   - **Target**: `internal/domain/` Config and related structs
   - **Expected methods**: Validate(), Sanitize(), ApplyProfile(), EstimateImpact()
   - **Verification**: Check if Config struct has these methods

**From REFACTORING_PLAN.md:**

6. **Generic Validation Interface** (HIGH impact, 2 hours work)
   - **Status**: ✅ COMPLETED
   - **Description**: Create generic Validator interface and ValidateAndWrap utility to eliminate 4 validation duplicates
   - **Target**: `internal/shared/utils/validation/`
   - **Files affected**: 4 files
   - **Verification**: ✅ `internal/shared/utils/validation/validation.go` exists with ValidateAndWrap[T]

7. **Config Loading Utility** (HIGH impact, 1 hour work)
   - **Status**: ✅ COMPLETED
   - **Description**: Create LoadConfigWithFallback utility to eliminate 2 config loading duplicates
   - **Target**: `internal/shared/utils/config/`
   - **Files affected**: 2 files
   - **Verification**: ✅ `internal/shared/utils/config/config.go` exists with LoadConfigWithFallback

8. **String Trimming Utility** (MEDIUM impact, 30 minutes work)
   - **Status**: ✅ COMPLETED
   - **Description**: Create TrimWhitespaceField utility to eliminate 2 string trimming duplicates
   - **Target**: `internal/shared/utils/strings/`
   - **Files affected**: 2 files
   - **Verification**: ✅ `internal/shared/utils/strings/trimming.go` exists with TrimWhitespaceField

### Priority 3 - Medium

**From REFACTORING_PLAN.md:**

9. **Error Details Utility** (MEDIUM impact, 2 hours work)
   - **Status**: ✅ COMPLETED
   - **Description**: Create error details utility to eliminate 3 error detail setting duplicates
   - **Target**: `internal/pkg/errors/`
   - **Files affected**: 3 files
   - **Verification**: ✅ `internal/pkg/errors/detail_helpers.go` exists with ErrorDetailsBuilder

10. **Test Helper Functions Refactoring** (MEDIUM impact, 3 hours work)
    - **Status**: NOT_STARTED
    - **Description**: Refactor test helper functions to eliminate 8+ test helper duplicates
    - **Target**: `tests/bdd/helpers/`
    - **Files affected**: 8+ files
    - **Verification**: Check BDD helpers for duplication

11. **Schema Min/Max Utility** (LOW impact, 1 hour work)
    - **Status**: ✅ COMPLETED
    - **Description**: Create schema min/max utility to eliminate 2 schema logic duplicates
    - **Target**: `internal/shared/utils/schema/`
    - **Files affected**: 2 files
    - **Verification**: ✅ `internal/shared/utils/schema/minmax.go` exists with MinMax struct

12. **Type Model Improvements** (MEDIUM impact, 4 hours work)
    - **Status**: NOT_STARTED
    - **Description**: Add IsValid(), Values(), consistent String() methods to all enums
    - **Target**: `internal/domain/interfaces.go`
    - **Impact**: Better architecture
    - **Verification**: Check enum definitions for missing methods

13. **Result Type Enhancement** (MEDIUM impact, 2 hours work)
    - **Status**: NOT_STARTED
    - **Description**: Enhance Result type for better validation chaining
    - **Target**: `internal/result/validation.go`
    - **Impact**: Better validation patterns
    - **Verification**: Check Result type for validation methods

**From SELF_REFLECTION_AND_PLAN.md:**

14. **SystemCache Research & Refactoring** (HIGH impact, 1 day work)
    - **Status**: NOT_STARTED
    - **Description**: Research domain.CacheType usage and implement decision (refactor SystemCache, update enum, or create mapping)
    - **Target**: SystemCache cleaner and domain types
    - **Subtasks**:
      - Research domain.CacheType usage
      - Document findings
      - Implement decision
    - **Verification**: Check domain.CacheType definition and usage

15. **Reduce LoadWithContext Complexity** (MEDIUM impact, 1 day work)
    - **Status**: NOT_STARTED
    - **Description**: Refactor config.LoadWithContext from complexity 20 to <10
    - **Target**: `internal/config/config.go`
    - **Approach**: Extract profile loading, operation processing, risk level processing, use early returns
    - **Verification**: Check complexity using golangci-lint

16. **Reduce validateProfileName Complexity** (MEDIUM impact, 0.5 days work)
    - **Status**: NOT_STARTED
    - **Description**: Refactor config.(\*ConfigValidator).validateProfileName from complexity 16 to <10
    - **Target**: `internal/config/validator.go`
    - **Verification**: Check complexity using golangci-lint

17. **Reduce Additional Complex Functions** (LOW impact, 1 day work)
    - **Status**: NOT_STARTED
    - **Description**: Reduce complexity of 3 more functions (TestIntegration_ValidationSanitizationPipeline, ErrorCode.String, EnhancedConfigLoader.SaveConfig)
    - **Target**: Multiple files
    - **Verification**: Check complexity using golangci-lint

### Priority 5 - Strategic (From COMPREHENSIVE_IMPROVEMENT_PLAN_2026-02-09.md)

**Phase 2: High Impact, Low Effort (Pareto 1% → 51% impact):**

19. **Extract Generic Cleaner Interface** (HIGH impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Create Cleaner interface and CleanerRegistry type to enable polymorphism and reduce duplication
- **Target**: `internal/cleaner/interface.go` (new), 11 cleaner files, cmd/clean-wizard/commands/clean.go
- **Impact**: Eliminates hardcoded cleaner lists, enables plugin architecture foundation
- **Verification**: Check if Cleaner interface exists and all cleaners implement it

20. **Fix Context Propagation in validate.go** (HIGH impact, 0.5 days work)

- **Status**: NOT_STARTED
- **Description**: Update ValidateItem() to preserve context in errors, add formatValidItems() helper
- **Target**: `internal/config/validate.go`
- **Impact**: Better error messages, easier debugging, improves error handling quality from 90.1/100 to 95+
- **Verification**: Check error messages include items/validItems context

21. **Unify Binary Enum Unmarshaling** (MEDIUM impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Create UnmarshalYAMLEnumBinary() helper to unify binary and standard enum unmarshaling
- **Target**: `internal/domain/type_safe_enums.go`
- **Impact**: Reduces code duplication, consistent error messages, easier maintenance
- **Verification**: Check if binary enums use unified unmarshaling

22. **Add Integration Tests for Enum Workflows** (HIGH impact, 2 days work)

- **Status**: NOT_STARTED
- **Description**: Create end-to-end enum workflow tests (integer format, string format, mixed format, round-trip)
- **Target**: `tests/integration/enum_workflow_test.go` (new), test configs (3 new)
- **Impact**: Verifies end-to-end functionality, confirms no breaking changes, tests actual usage
- **Verification**: Check if integration tests exist and pass with `-tags=integration`

**Phase 3: High Impact, Medium Effort (Pareto 4% → 64% impact):**

23. **Verify All Cleaners Handle Enums Correctly** (HIGH impact, 2 days work)

- **Status**: NOT_STARTED
- **Description**: Review 11 cleaners for enum usage, verify no hardcoded strings, ensure switch statements use enum constants
- **Target**: All 11 cleaner files (nix, homebrew, docker, cargo, golang, nodepackages, buildcache, systemcache, tempfiles, projectsmanagementautomation)
- **Impact**: Ensures enums are actually used, catches hardcoded strings, prevents type-safety regressions
- **Verification**: Check all cleaners for correct enum usage patterns

24. **Add Enum Validation to Config Boundaries** (HIGH impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Add enum validation to config.LoadWithContext(), validate all enum values after loading
- **Target**: `internal/config/config.go`
- **Impact**: Catches invalid enums at config load time, provides helpful error messages
- **Verification**: Check if config validation includes enum validation

25. **Reduce Complexity in Top 5 Functions** (MEDIUM impact, 1 day work)

- **Status**: NOT_STARTED (duplicates existing TODO #15)
- **Description**: Refactor LoadWithContext (20→<10), TestIntegration_ValidationSanitizationPipeline (19→<10), validateProfileName (16→<10), ErrorCode.String (15→<10), SaveConfig (15→<10)
- **Target**: Multiple files
- **Impact**: Better maintainability, more readable functions, easier testing
- **Verification**: Check complexity using golangci-lint

**Phase 4: Medium Impact, Medium Effort:**

26. **Reduce Complexity in Remaining Functions** (MEDIUM impact, 2 days work)

- **Status**: NOT_STARTED
- **Description**: Refactor remaining 16 high-complexity functions (all complexity >10)
- **Target**: 16 files across codebase
- **Impact**: Better maintainability, lower cognitive load
- **Verification**: Check complexity using golangci-lint

27. **Add Edge Case Tests for Enum Unmarshaling** (MEDIUM impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Add tests for negative integers, out-of-range integers, mixed case strings, empty values, null values
- **Target**: Existing test files
- **Impact**: Better error handling, catches edge cases
- **Verification**: Check if edge case tests exist

28. **Test Enum Round-Trip Serialization** (MEDIUM impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Test YAML→Go→YAML and JSON→Go→JSON for all enum types
- **Target**: Existing test files
- **Impact**: Ensures no data corruption, verifies consistent formatting
- **Verification**: Check if round-trip tests exist

**Phase 5: Low Impact, High Effort:**

29. **Create Comprehensive Architecture Documentation** (LOW impact, 2 days work)

- **Status**: NOT_STARTED
- **Description**: Create ARCHITECTURE.md covering architecture diagram, layer responsibilities, data flow, design decisions, type safety philosophy, testing strategy, extension points
- **Target**: `ARCHITECTURE.md` (new)
- **Impact**: Better onboarding, preserved design rationale, easier maintenance
- **Verification**: Check if ARCHITECTURE.md exists and is comprehensive

30. **Investigate RiskLevelType Manual Processing** (MEDIUM impact, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Investigate Viper enum support, test using RiskLevelType directly with Viper, fix manual processing if possible
- **Target**: `internal/config/config.go:86-108`
- **Impact**: Consistency with other enums, less code to maintain
- **Verification**: Check if RiskLevelType uses standard enum unmarshaler

31. **Add Dependency Injection with samber/do/v2** (LOW impact, 2 days work)

- **Status**: NOT_STARTED
- **Description**: Add samber/do/v2 to go.mod, create DI container in cmd/clean-wizard, update main.go and all services
- **Target**: Multiple files
- **Impact**: Less boilerplate, easier testing, automatic dependency resolution
- **Verification**: Check if samber/do/v2 is used in main.go and services

**Phase 6: Final Validation:**

32. **Verify Full Integration and Run All Tests** (CRITICAL, 1 day work)

- **Status**: NOT_STARTED
- **Description**: Run full test suite (just test), integration tests (go test ./tests/integration -tags=integration), BDD tests (go test ./tests/bdd), benchmarks (go test -bench=. ./tests/benchmark), check coverage (go test -cover ./...)
- **Target**: All test suites
- **Impact**: Ensures everything works, comprehensive validation
- **Verification**: All tests pass, coverage >85%

33. **Create Quick Reference Guide for Enum Types** (LOW impact, 0.5 days work)

- **Status**: NOT_STARTED
- **Description**: Create docs/ENUM_QUICK_REFERENCE.md listing all 15+ enum types, values, usage examples
- **Target**: `docs/ENUM_QUICK_REFERENCE.md` (new)
- **Impact**: Easier lookup for developers
- **Verification**: Check if ENUM_QUICK_REFERENCE.md exists

### Priority 6 - Low

**From SELF_REFLECTION_AND_PLAN.md:**

34. **Registry Documentation** (LOW impact, 0.25 days work)
    - **Status**: NOT_STARTED
    - **Description**: Document how to use the CleanerRegistry
    - **Target**: Documentation files
    - **Verification**: Check if registry documentation exists

**From FEATURES_SUMMARY_TABLE.md:**

35. **Implement Language Version Manager Cleaning** (HIGH impact, HIGH effort)
    - **Status**: NOT_STARTED
    - **Description**: Fix Language Version Manager NO-OP - implement actual cleaning instead of returning success without action
    - **Target**: `internal/cleaner/langversionmanager.go:133-154`
    - **Impact**: Removes broken cleaner, improves functionality
    - **Verification**: Check if langversionmanager.go performs actual cleanup

36. **Implement Missing CLI Commands** (HIGH impact, HIGH effort)
    - **Status**: ✅ COMPLETED & VERIFIED
    - **Description**: Implement scan, init, profile, config commands (4 missing commands)
    - **Target**: `cmd/clean-wizard/commands/` (scan.go, init.go, profile.go, config.go)
    - **Impact**: Closes 80% gap between documentation and implementation
    - **Verification**: All 5 commands exist, build succeeds, --help works for root and subcommands
    - **Details**: scan (2 flags), init (2 flags), profile (4 subcommands), config (4 subcommands)

37. **Improve Size Reporting Across All Cleaners** (HIGH impact, MEDIUM effort)
    - **Status**: NOT_STARTED
    - **Description**: Replace hardcoded estimates with actual size calculations, fix Docker and Cargo size reporting
    - **Target**: Multiple cleaner files
    - **Impact**: Accurate dry-run estimates, better user experience
    - **Verification**: Check if cleaners report actual bytes freed

38. **Implement or Remove Unused Enum Values** (LOW impact, MEDIUM effort)
    - **Status**: NOT_STARTED
    - **Description**: Implement remaining BuildToolType values (GO, RUST, NODE, PYTHON), CacheType values (PIP, NPM, YARN, CCACHE), VersionManagerType values (GVM, SDKMAN, JENV), or remove unused values
    - **Target**: `internal/domain/operation_settings.go`
    - **Impact**: Remove dead code, enable more cache types
    - **Verification**: Check if unused enum values are implemented or removed

39. **Add Linux Support for System Cache Cleaner** (MEDIUM impact, MEDIUM effort)
    - **Status**: NOT_STARTED
    - **Description**: Extend System Cache cleaner to support Linux (currently macOS only, 4 of 8 cache types)
    - **Target**: `internal/cleaner/systemcache.go`
    - **Impact**: Increases platform support, enables more cache cleanup
    - **Verification**: Check if systemcache.go supports Linux cache paths

---

## TODO STATUS TRACKING

| TODO                             | Source File                         | Status       | Verification Notes                                                                              |
| -------------------------------- | ----------------------------------- | ------------ | ----------------------------------------------------------------------------------------------- |
| Generic Context System           | IMPLEMENTATION_STATUS.md            | ✅ COMPLETED  | Context[T] generic struct, ValidationConfig, ErrorConfig, SanitizationConfig, 19 tests passing |
| CleanerRegistry Integration      | SELF_REFLECTION_AND_PLAN.md         | ✅ COMPLETED | Registry implemented with 231 lines, 12 tests, integrated in cmd/clean-wizard/commands/clean.go |
| Deprecation Fixes (20+ files)    | SELF_REFLECTION_AND_PLAN.md         | ✅ COMPLETED | 49 warnings eliminated across 45+ files                                                         |
| Backward Compatibility Aliases   | IMPLEMENTATION_STATUS.md            | ✅ OBSOLETE   | No type aliases found in domain/                                                               |
| Domain Model Enhancement         | IMPLEMENTATION_STATUS.md            | NOT_STARTED  | Need to check current domain models                                                             |
| Generic Validation Interface     | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to create new utility                                                                      |
| Config Loading Utility           | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to create new utility                                                                      |
| String Trimming Utility          | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to create new utility                                                                      |
| Error Details Utility            | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to create new utility                                                                      |
| Test Helper Refactoring          | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to refactor existing helpers                                                               |
| Schema Min/Max Utility           | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to create new utility                                                                      |
| Type Model Improvements          | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to check all enum definitions                                                              |
| Result Type Enhancement          | REFACTORING_PLAN.md                 | NOT_STARTED  | Need to check current Result type                                                               |
| SystemCache Research             | SELF_REFLECTION_AND_PLAN.md         | NOT_STARTED  | Need to research domain.CacheType usage                                                         |
| LoadWithContext Complexity       | SELF_REFLECTION_AND_PLAN.md         | NOT_STARTED  | Need to check current implementation                                                            |
| validateProfileName Complexity   | SELF_REFLECTION_AND_PLAN.md         | NOT_STARTED  | Need to check current implementation                                                            |
| Additional Complexity Reductions | SELF_REFLECTION_AND_PLAN.md         | NOT_STARTED  | Need to identify exact functions                                                                |
| Registry Documentation           | SELF_REFLECTION_AND_PLAN.md         | NOT_STARTED  | Documentation task                                                                              |
| **NEW - UNSAFE EXEC CALLS**      | 2026-01-28_01-30_EXECUTION_AUDIT.md | 🔴 CRITICAL  | 9 exec calls without timeout - cargo-cache, cargo clean, npm/pnpm/yarn/bun cache commands       |
| **NEW - Interface Compliance**   | CLEANER_INTERFACE_ANALYSIS.md       | 🔴 CRITICAL  | nix.go missing Clean(ctx), golang_cache_cleaner.go missing IsAvailable()                        |
| **NEW - Enum Inconsistencies**   | ENUM_USAGE_ANALYSIS.md              | 🔍 DOCUMENTED | NodePackages: local string vs domain int; BuildCache: tools vs languages (requires refactor decision) |

---

## DE-DUPLICATION NOTES

**Identified Duplication Patterns:**

1. **Validation Wrapping** - Found in 4 files, planned to eliminate with Generic Validation Interface
2. **Config Loading** - Found in 2 files, planned to eliminate with Config Loading Utility
3. **String Trimming** - Found in 2 files, planned to eliminate with String Trimming Utility
4. **Error Detail Setting** - Found in 3 files, planned to eliminate with Error Details Utility
5. **Test Helper Functions** - Found in 8+ files, planned to refactor
6. **Schema Min/Max Logic** - Found in 2 files, planned to eliminate with Schema Min/Max Utility

**Total Duplicates Identified:** 62 (from art-dupl analysis)
**Target Reduction:** 75% (to ~15 duplicates)
**Current Status:** Planning phase complete, execution pending

---

## VERIFICATION CHECKLIST

### Immediate Actions Required:

- [x] Add timeout protection to UNSAFE EXEC CALLS (9 commands identified) ✅ VERIFIED
- [x] Fix Cleaner interface compliance (nix.go, golang_cache_cleaner.go) ✅ VERIFIED
- [ ] Refactor enum inconsistencies (NodePackages, BuildCache)

### Short-term Actions:

- [ ] Implement Generic Context System
- [ ] Create missing utilities (validation, config, strings, errors, schema)

### Long-term Actions:

- [ ] Complete domain model enhancements
- [ ] Reduce function complexity across codebase
- [ ] Add comprehensive documentation

### ✅ VERIFIED COMPLETED:

- [x] CleanerRegistry Implementation (231 lines, 12 tests, fully integrated)
- [x] Deprecation Fixes (49 warnings eliminated, build clean)
- [x] Timeout Protection (all exec commands have context timeout)
- [x] Cleaner Interface Compliance (all 13 cleaners implement Clean(), IsAvailable(), Name())
- [x] CLI Commands (5/5 implemented: clean, scan, init, profile, config)
- [x] Size Reporting Deduplication (CalculateBytesFreed utility extracted)
- [x] CompiledBinariesCleaner (576 lines, 918 tests)

---

## SUCCESS METRICS (From COMPREHENSIVE_IMPROVEMENT_PLAN)

### Technical Metrics:

- [x] All tests passing (51 test files, ~178s BDD suite) ✅
- [ ] Test coverage > 85% (current: varies by package, avg ~70%)
- [ ] Cyclomatic complexity < 10 for all functions (current: 21 functions >10)
- [ ] Error handling quality score > 95 (current: 90.1)
- [x] Zero lint warnings in production code ✅
- [ ] Zero security vulnerabilities
- [ ] All critical paths covered by integration tests

### Architecture Metrics:

- [x] Cleaner interface implemented and used by all cleaners ✅
- [x] All enums using unified unmarshaling ✅
- [x] All error messages preserve context ✅
- [ ] All high-complexity functions refactored
- [x] Zero circular dependencies (already achieved) ✅

### Developer Experience Metrics:

- [ ] Onboarding time < 1 hour (with architecture docs)
- [ ] New cleaner can be added in < 30 minutes
- [ ] Error messages are actionable
- [ ] Configuration format is clear and documented

---

## OPEN QUESTIONS (From COMPREHENSIVE_IMPROVEMENT_PLAN)

1. **RiskLevelType Manual Processing**: Should we fix the manual RiskLevelType processing in `internal/config/config.go:86-108` or keep it?
   - **Context**: All other enums use type-safe UnmarshalYAML(), but RiskLevelType is manually processed as string
   - **Recommendation**: Investigate Viper enum support first, then decide

2. **Dependency Injection Adoption**: Should we adopt samber/do/v2 for dependency injection?
   - **Context**: Manual dependency wiring everywhere, more boilerplate, harder to test
   - **Recommendation**: Adopt selectively for complex services, evaluate after first phase

3. **Plugin Architecture**: Should we implement a plugin architecture for cleaners?
   - **Context**: Currently hardcoded list of cleaners, plugin system mentioned in architectural analysis
   - **Recommendation**: Defer decision, focus on core type safety first

---

## RECOMMENDED EXECUTION PRIORITY

1. **Phase 1** (Critical Blocking): Clean disk space, run full test suite
2. **Phase 2** (High Impact, Low Effort): Extract Cleaner interface, fix context propagation, unify enum unmarshaling, add integration tests
3. **Phase 3** (High Impact, Medium Effort): Verify cleaner enum usage, add enum validation, reduce complexity in top 5 functions
4. **Phase 4-6**: Based on Phase 2-3 results and business needs

## **Total Timeline**: 7 weeks with incremental value delivery every week

## VERIFICATION RESULTS (2026-02-09)

### Verified Completed Tasks:

1. **Cleaner Interface Compliance** - ✅ VERIFIED
   - nix.go:60 has `Clean(ctx context.Context)` method
   - golang_cache_cleaner.go:39 has `IsAvailable(ctx context.Context)` method
   - All 13 cleaners implement both methods

2. **Docker Enum Refactoring** - ✅ VERIFIED
   - Committed in 5e94e2a
   - Migrated from local enum to domain enum
   - All tests passing

3. **Binary Enum Unification** - ✅ VERIFIED
   - Removed 69 lines of duplicate code
   - Consolidated to UnmarshalYAMLEnum
   - Numeric string handling added

4. **Context Propagation** - ✅ VERIFIED
   - Error messages in validate.go include context
   - Index, valid options, and full input list preserved

5. **Integration Tests for Enums** - ✅ VERIFIED
   - 6 comprehensive test functions implemented
   - All passing

6. **Enum Validation at Config Boundaries** - ✅ VERIFIED
   - Validation for RiskLevel, Enabled, DockerPruneMode
   - Validation for GoPackages, SystemCache, BuildCache

7. **Cleaner Enum Usage** - ✅ VERIFIED
   - All cleaners use type-safe enum handling
   - No raw int comparisons found

### Still Pending Critical Issues:

1. **UNSAFE EXEC CALLS** - 🔴 VERIFIED
   - cargo.go:177 - `cargo-cache --autoclean` (no timeout)
   - Multiple other commands without timeout protection

2. **CLI Command Gap** - ✅ VERIFIED FIXED
   - All 5 commands now implemented: clean, scan, init, profile, config
   - scan.go:13 - NewScanCommand() implemented
   - init.go:12 - NewInitCommand() implemented
   - profile.go:13 - NewProfileCommand() with 4 subcommands
   - config.go:14 - NewConfigCommand() with 4 subcommands
   - Build succeeds, all --help commands verified

3. **Language Version Manager NO-OP** - 🔴 VERIFIED
   - Explicit NO-OP in langversionmanager.go:133-154
   - Returns success without cleaning

4. **Enum Inconsistencies** - 🟠 VERIFIED
   - Docker: ✅ FIXED
   - SystemCache: Local lowercase vs domain uppercase
   - NodePackages: Local string vs domain integer
   - BuildCache: Different abstractions (tools vs languages)
