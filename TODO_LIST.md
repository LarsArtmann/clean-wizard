# TODO LIST

**Last Updated:** 2026-02-13
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

---

## PENDING TODOs

### Priority 1 - Critical

| #   | Task                                                                                               | Impact | Status      |
| --- | -------------------------------------------------------------------------------------------------- | ------ | ----------- |
| 1   | Generic Context System - unify ValidationContext, ErrorDetails, SanitizationChange into Context[T] | 90%    | NOT_STARTED |
| 2   | Domain Model Enhancement - add Validate(), Sanitize(), ApplyProfile() to Config struct             | 50%    | NOT_STARTED |

### Priority 2 - Enum Refactoring

| #   | Task                                                            | Status            |
| --- | --------------------------------------------------------------- | ----------------- |
| 3   | NodePackages: refactor local string enum to domain integer enum | REQUIRES_REFACTOR |
| 4   | BuildCache: decide on tools vs languages abstraction            | REQUIRES_DECISION |

### Priority 3 - Complexity Reduction

| #   | Task                                      | Current → Target | Status      |
| --- | ----------------------------------------- | ---------------- | ----------- |
| 5   | Reduce LoadWithContext complexity         | 20 → <10         | NOT_STARTED |
| 6   | Reduce validateProfileName complexity     | 16 → <10         | NOT_STARTED |
| 7   | Reduce 19 other high-complexity functions | >10 → <10        | NOT_STARTED |

### Priority 4 - Test Helper Refactoring

| #   | Task                      | Files    | Status      |
| --- | ------------------------- | -------- | ----------- |
| 8   | Refactor BDD test helpers | 8+ files | NOT_STARTED |

### Priority 5 - Type Model Improvements

| #   | Task                                           | Status      |
| --- | ---------------------------------------------- | ----------- |
| 9   | Add IsValid(), Values(), String() to all enums | NOT_STARTED |
| 10  | Enhance Result type for validation chaining    | NOT_STARTED |

### Priority 6 - Cleaner Improvements

| #   | Task                                         | Status      |
| --- | -------------------------------------------- | ----------- |
| 11  | Fix Language Version Manager NO-OP           | NOT_STARTED |
| 12  | Fix Docker size reporting (returns 0)        | NOT_STARTED |
| 13  | Fix Cargo size reporting                     | NOT_STARTED |
| 14  | Improve dry-run estimates (hardcoded values) | NOT_STARTED |
| 15  | Add Linux support for SystemCache cleaner    | NOT_STARTED |

### Priority 7 - Documentation

| #   | Task                           | Status      |
| --- | ------------------------------ | ----------- |
| 16  | Create ARCHITECTURE.md         | NOT_STARTED |
| 17  | Document CleanerRegistry usage | NOT_STARTED |
| 18  | Create ENUM_QUICK_REFERENCE.md | NOT_STARTED |

### Priority 8 - Future Considerations

| #   | Task                                              | Status      |
| --- | ------------------------------------------------- | ----------- |
| 19  | Investigate RiskLevelType manual Viper processing | NOT_STARTED |
| 20  | Add samber/do/v2 dependency injection             | NOT_STARTED |
| 21  | Plugin architecture for cleaners                  | DEFERRED    |

---

## VERIFICATION CHECKLIST

### Immediate:

- [x] Timeout protection on exec calls
- [x] Cleaner interface compliance
- [ ] Refactor enum inconsistencies (NodePackages, BuildCache)

### Short-term:

- [ ] Implement Generic Context System
- [ ] Reduce function complexity

### Long-term:

- [ ] Complete domain model enhancements
- [ ] Add comprehensive documentation

---

## SOURCE FILES WITH PENDING WORK

| File                                          | Pending Tasks |
| --------------------------------------------- | ------------- |
| IMPLEMENTATION_STATUS.md                      | 3 tasks       |
| REFACTORING_PLAN.md                           | 12 tasks      |
| SELF_REFLECTION_AND_PLAN.md                   | 17 tasks      |
| COMPREHENSIVE_ARCHITECTURAL_TODO_LIST.md      | 20 tasks      |
| COMPREHENSIVE_GITHUB_ISSUES_EXECUTION_PLAN.md | 32 tasks      |
| COMPREHENSIVE_IMPROVEMENT_PLAN_2026-02-09.md  | 17 tasks      |
| COMPREHENSIVE_REFLECTION_2026-02-11.md        | 12 tasks      |
| FEATURES.md                                   | 5+ features   |
| ENUM_USAGE_ANALYSIS.md                        | 10 tasks      |
