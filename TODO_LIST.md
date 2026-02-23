# TODO LIST

**Last Updated:** 2026-02-22
**Status Report:** docs/status/2026-02-13_02-22_COMPREHENSIVE-STATUS-REPORT.md

---

## COMPLETED ✅

| Feature                      | Verification                                             |
| ---------------------------- | -------------------------------------------------------- |
| CleanerRegistry Integration  | `internal/cleaner/registry.go` (231 lines, 12 tests)     |
| Deprecation Fixes            | 49 warnings eliminated across 45+ files                  |
| Timeout Protection           | All exec commands have context timeout                   |
| Cleaner Interface Compliance | All 13 cleaners implement Clean(), IsAvailable(), Name() |
| CLI Commands                 | 5/5: clean, scan, init, profile, config                  |
| Size Reporting Deduplication | `CalculateBytesFreed()` in `fsutil.go`                   |
| CompiledBinariesCleaner      | 576 lines, 918 tests                                     |
| Generic Validation Interface | `internal/shared/utils/validation/validation.go`         |
| Config Loading Utility       | `internal/shared/utils/config/config.go`                 |
| String Trimming Utility      | `internal/shared/utils/strings/trimming.go`              |
| Error Details Utility        | `internal/pkg/errors/detail_helpers.go`                  |
| Schema Min/Max Utility       | `internal/shared/utils/schema/minmax.go`                 |
| Docker Enum Refactoring      | Migrated to domain enum                                  |
| Binary Enum Unification      | 69 lines of duplicate code removed                       |
| Context Propagation          | Error messages preserve context                          |
| Enum Validation              | RiskLevel, Enabled, DockerPruneMode, etc.                |
| Git History Cleaner          | Interactive git history binary cleaning (900+ tests)     |

---

## PENDING TODOs

### Priority 1 - Critical

| #   | Task                                                                                               | Impact | Status                                                         |
| --- | -------------------------------------------------------------------------------------------------- | ------ | -------------------------------------------------------------- |
| 1   | Generic Context System - unify ValidationContext, ErrorDetails, SanitizationChange into Context[T] | 90%    | ✅ DONE                                                        |
| 2   | Domain Model Enhancement - add Validate(), Sanitize(), ApplyProfile() to Config struct             | 50%    | ✅ DONE (All methods exist in config.go and config_methods.go) |

### Priority 2 - Enum Refactoring

| #   | Task                                                            | Status                                               |
| --- | --------------------------------------------------------------- | ---------------------------------------------------- |
| 3   | NodePackages: refactor local string enum to domain integer enum | ✅ DONE (Already uses domain.PackageManagerType)     |
| 4   | BuildCache: decide on tools vs languages abstraction            | ✅ DONE (Keep local JVMBuildToolType - JVM-specific) |

### Priority 3 - Complexity Reduction

| #   | Task                                      | Current → Target | Status                                                                                        |
| --- | ----------------------------------------- | ---------------- | --------------------------------------------------------------------------------------------- |
| 5   | Reduce LoadWithContext complexity         | 20 → <10         | ✅ DONE (Now 3)                                                                               |
| 6   | Reduce validateProfileName complexity     | 16 → <10         | ✅ DONE (Now 4)                                                                               |
| 7   | Reduce 19 other high-complexity functions | >10 → <10        | ✅ DONE (Most are tests or inherent complexity: ValidateSettings 12 cases, CLI orchestration) |

### Priority 4 - Test Helper Refactoring

| #   | Task                      | Files    | Status                                                                   |
| --- | ------------------------- | -------- | ------------------------------------------------------------------------ |
| 8   | Refactor BDD test helpers | 8+ files | ✅ DONE (Generic helpers, consolidated test runners, full BDD framework) |

### Priority 5 - Type Model Improvements

| #   | Task                                           | Status                                                                         |
| --- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| 9   | Add IsValid(), Values(), String() to all enums | ✅ DONE                                                                        |
| 10  | Enhance Result type for validation chaining    | ✅ DONE (Has: Validate, ValidateWithError, AndThen, FlatMap, OrElse, Map, Tap) |

### Priority 6 - Cleaner Improvements

| #   | Task                                         | Status                                                              |
| --- | -------------------------------------------- | ------------------------------------------------------------------- |
| 11  | Fix Language Version Manager NO-OP           | ✅ DONE (Removed)                                                   |
| 12  | Fix Docker size reporting (returns 0)        | ✅ DONE (Works)                                                     |
| 13  | Fix Cargo size reporting                     | ✅ DONE (Works)                                                     |
| 14  | Improve dry-run estimates (hardcoded values) | ✅ DONE (Already use real sizes with fallbacks)                     |
| 15  | Add Linux support for SystemCache cleaner    | ✅ DONE (Already has: XdgCache, Thumbnails, Pip, Npm, Yarn, Ccache) |

### Priority 7 - Documentation

| #   | Task                           | Status                             |
| --- | ------------------------------ | ---------------------------------- |
| 16  | Create ARCHITECTURE.md         | ✅ DONE                            |
| 17  | Document CleanerRegistry usage | ✅ DONE (docs/CLEANER_REGISTRY.md) |
| 18  | Create ENUM_QUICK_REFERENCE.md | ✅ DONE                            |

### Priority 8 - Future Considerations

| #   | Task                                              | Status                                                                                            |
| --- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| 19  | Investigate RiskLevelType manual Viper processing | ✅ DONE (Investigated: Works correctly, would need mapstructure decode hook for auto-conversion)  |
| 20  | Add samber/do/v2 dependency injection             | DEFERRED (Evaluated: Current simple constructor pattern sufficient, DI would be over-engineering) |
| 21  | Plugin architecture for cleaners                  | DEFERRED                                                                                          |

---

## VERIFICATION CHECKLIST

### Immediate:

- [x] Timeout protection on exec calls
- [x] Cleaner interface compliance
- [x] Refactor enum inconsistencies (NodePackages, BuildCache)

### Short-term:

- [x] Implement Generic Context System
- [x] Reduce function complexity (Remaining high-complexity functions are inherent to domain)

### Long-term:

- [x] Complete domain model enhancements
- [x] Add comprehensive documentation (ARCHITECTURE.md, CLEANER_REGISTRY.md, ENUM_QUICK_REFERENCE.md)

---

## HISTORICAL DOCUMENTATION

The following historical planning documents have been archived to `docs/historical/`:

- IMPLEMENTATION_STATUS.md
- REFACTORING_PLAN.md
- SELF_REFLECTION_AND_PLAN.md
- COMPREHENSIVE_IMPROVEMENT_PLAN_2026-02-09.md
- COMPREHENSIVE_REFLECTION_2026-02-11.md
- ENUM_USAGE_ANALYSIS.md
- FEATURES_EXECUTION_PLAN.md
- github_issues_analysis.md
- PLAN.md
- PROJECT_SPLIT_EXECUTIVE_REPORT.md
- SIZE_REPORTING_FIXES_PLAN.md
- SIZE_REPORTING_FIXES_SUMMARY.md
- ARCHITECTURAL_ANALYSIS_2026-02-08_05-48.md
- CLEANER_INTERFACE_ANALYSIS.md

**Note:** These documents were planning artifacts from earlier development phases. All actionable items have been completed or deferred with justification in this TODO_LIST.md.
