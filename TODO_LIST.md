# TODO LIST

**Last Updated:** 2026-02-09  
**Total MD Files Found:** 91  
**Files Processed:** 14/91  
**Files Remaining:** 77

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

### Remaining Files (in processing order)
12. docs/result.md

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
   - **Status**: NOT_STARTED
   - **Description**: Integrate CleanerRegistry into cmd/clean-wizard/commands/clean.go and add tests
   - **Target**: `internal/cleaner/registry.go` and `cmd/clean-wizard/commands/clean.go`
   - **Expected outcome**: Dynamic cleaner discovery and management
   - **Verification**: Check if registry.go exists and if it's used in clean.go

**From SELF_REFLECTION_AND_PLAN.md:**

3. **Complete Deprecation Fixes** (MEDIUM impact, 1.85 days work)
   - **Status**: NOT_STARTED
   - **Description**: Fix ~20 test/support files with deprecation warnings (Strategy and RiskLevel aliases)
   - **Target**: Multiple files in internal/cleaner/*_test.go, conversions/, adapters/, api/, middleware/, tests/
   - **Subtasks**:
     - Fix test files (7 files)
     - Fix conversions package (2 files)
     - Fix adapters package (1 file)
     - Fix api package (2 files)
     - Fix middleware package (1 file)
     - Fix benchmark tests (1 file)
     - Fix RiskLevel deprecations (~15 files)
   - **Verification**: Run go build and check for deprecation warnings

### Priority 2 - High

**From IMPLEMENTATION_STATUS.md:**

4. **Eliminate Backward Compatibility Aliases** (70% impact, 2 days work)
   - **Status**: NOT_STARTED
   - **Description**: Remove duplicate type systems (RiskLevel = RiskLevelType, etc.) with phased migration
   - **Target**: `internal/domain/` type definitions
   - **Subtasks**:
     - Phase 1: Mark aliases as deprecated
     - Phase 2: Replace all usages
     - Phase 3: Remove aliases entirely
   - **Verification**: Check domain types for type aliases

5. **Domain Model Enhancement** (50% impact, 3 days work)
   - **Status**: NOT_STARTED
   - **Description**: Transform anemic domain models into rich domain objects with behavior
   - **Target**: `internal/domain/` Config and related structs
   - **Expected methods**: Validate(), Sanitize(), ApplyProfile(), EstimateImpact()
   - **Verification**: Check if Config struct has these methods

**From REFACTORING_PLAN.md:**

6. **Generic Validation Interface** (HIGH impact, 2 hours work)
   - **Status**: NOT_STARTED
   - **Description**: Create generic Validator interface and ValidateAndWrap utility to eliminate 4 validation duplicates
   - **Target**: `internal/shared/utils/validation/`
   - **Files affected**: 4 files
   - **Verification**: Check if validation utility exists

7. **Config Loading Utility** (HIGH impact, 1 hour work)
   - **Status**: NOT_STARTED
   - **Description**: Create LoadConfigWithFallback utility to eliminate 2 config loading duplicates
   - **Target**: `internal/shared/utils/config/`
   - **Files affected**: 2 files
   - **Verification**: Check if config utility exists

8. **String Trimming Utility** (MEDIUM impact, 30 minutes work)
   - **Status**: NOT_STARTED
   - **Description**: Create TrimWhitespaceField utility to eliminate 2 string trimming duplicates
   - **Target**: `internal/shared/utils/strings/`
   - **Files affected**: 2 files
   - **Verification**: Check if string utility exists

### Priority 3 - Medium

**From REFACTORING_PLAN.md:**

9. **Error Details Utility** (MEDIUM impact, 2 hours work)
   - **Status**: NOT_STARTED
   - **Description**: Create error details utility to eliminate 3 error detail setting duplicates
   - **Target**: `internal/pkg/errors/`
   - **Files affected**: 3 files
   - **Verification**: Check if error utility exists

10. **Test Helper Functions Refactoring** (MEDIUM impact, 3 hours work)
    - **Status**: NOT_STARTED
    - **Description**: Refactor test helper functions to eliminate 8+ test helper duplicates
    - **Target**: `tests/bdd/helpers/`
    - **Files affected**: 8+ files
    - **Verification**: Check BDD helpers for duplication

11. **Schema Min/Max Utility** (LOW impact, 1 hour work)
    - **Status**: NOT_STARTED
    - **Description**: Create schema min/max utility to eliminate 2 schema logic duplicates
    - **Target**: `internal/shared/utils/schema/`
    - **Files affected**: 2 files
    - **Verification**: Check if schema utility exists

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
    - **Description**: Refactor config.(*ConfigValidator).validateProfileName from complexity 16 to <10
    - **Target**: `internal/config/validator.go`
    - **Verification**: Check complexity using golangci-lint

17. **Reduce Additional Complex Functions** (LOW impact, 1 day work)
    - **Status**: NOT_STARTED
    - **Description**: Reduce complexity of 3 more functions (TestIntegration_ValidationSanitizationPipeline, ErrorCode.String, EnhancedConfigLoader.SaveConfig)
    - **Target**: Multiple files
    - **Verification**: Check complexity using golangci-lint

### Priority 4 - Low

**From SELF_REFLECTION_AND_PLAN.md:**

18. **Registry Documentation** (LOW impact, 0.25 days work)
    - **Status**: NOT_STARTED
    - **Description**: Document how to use the CleanerRegistry
    - **Target**: Documentation files
    - **Verification**: Check if registry documentation exists

---

## TODO STATUS TRACKING

| TODO | Source File | Status | Verification Notes |
|------|-------------|--------|-------------------|
| Generic Context System | IMPLEMENTATION_STATUS.md | NOT_STARTED | Need to verify current context types in code |
| CleanerRegistry Integration | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Registry exists but not integrated in clean.go |
| Deprecation Fixes (20+ files) | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Need to identify exact files and fix them |
| Backward Compatibility Aliases | IMPLEMENTATION_STATUS.md | NOT_STARTED | Need to check domain types |
| Domain Model Enhancement | IMPLEMENTATION_STATUS.md | NOT_STARTED | Need to check current domain models |
| Generic Validation Interface | REFACTORING_PLAN.md | NOT_STARTED | Need to create new utility |
| Config Loading Utility | REFACTORING_PLAN.md | NOT_STARTED | Need to create new utility |
| String Trimming Utility | REFACTORING_PLAN.md | NOT_STARTED | Need to create new utility |
| Error Details Utility | REFACTORING_PLAN.md | NOT_STARTED | Need to create new utility |
| Test Helper Refactoring | REFACTORING_PLAN.md | NOT_STARTED | Need to refactor existing helpers |
| Schema Min/Max Utility | REFACTORING_PLAN.md | NOT_STARTED | Need to create new utility |
| Type Model Improvements | REFACTORING_PLAN.md | NOT_STARTED | Need to check all enum definitions |
| Result Type Enhancement | REFACTORING_PLAN.md | NOT_STARTED | Need to check current Result type |
| SystemCache Research | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Need to research domain.CacheType usage |
| LoadWithContext Complexity | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Need to check current implementation |
| validateProfileName Complexity | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Need to check current implementation |
| Additional Complexity Reductions | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Need to identify exact functions |
| Registry Documentation | SELF_REFLECTION_AND_PLAN.md | NOT_STARTED | Documentation task |

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
- [ ] Verify CleanerRegistry exists in internal/cleaner/
- [ ] Check if Registry is used in cmd/clean-wizard/commands/clean.go
- [ ] Run go build to identify deprecation warnings
- [ ] Check domain types for aliases (RiskLevel = RiskLevelType)
- [ ] Verify LoadWithContext complexity using golangci-lint
- [ ] Research domain.CacheType usage

### Short-term Actions:
- [ ] Implement Generic Context System
- [ ] Complete deprecation fixes for all files
- [ ] Integrate CleanerRegistry into codebase
- [ ] Create missing utilities (validation, config, strings, errors, schema)

### Long-term Actions:
- [ ] Complete domain model enhancements
- [ ] Reduce function complexity across codebase
- [ ] Add comprehensive documentation