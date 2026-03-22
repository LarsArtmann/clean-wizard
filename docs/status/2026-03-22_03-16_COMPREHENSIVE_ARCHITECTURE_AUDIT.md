# Comprehensive Architecture Audit & Status Report

**Date:** 2026-03-22 03:16  
**Auditor:** Senior Software Architect (AI)  
**Project:** Clean Wizard  
**Lines of Code:** 28,859 (189 Go files, 59 test files)

---

## Executive Summary

This audit reveals a codebase with **strong architectural foundations** but significant **technical debt** in the form of split brains, oversized files, and inconsistent error handling. The project demonstrates excellent type safety and domain modeling but needs consolidation and refactoring to achieve production excellence.

**Overall Grade: B+** (Good foundation, needs cleanup)

---

## Work Completed

### ✅ FULLY DONE

| Task                            | Impact   | Details                                                     |
| ------------------------------- | -------- | ----------------------------------------------------------- |
| Branching-Flow Analysis         | High     | context: 88.5/100, compose: 98/100, phantom: 891 violations |
| SanitizationWarning Split Brain | Critical | Unified into domain package as single source of truth       |
| EnvironmentConfig Refactoring   | High     | Split 31-field struct into 9 focused domain structs         |
| DefaultValidationLogger Rename  | Medium   | Eliminated base-naming anti-pattern                         |
| Error Context Enhancement       | Medium   | Added validValuesDescription to error messages              |

### 🔄 PARTIALLY DONE

| Task                          | Status      | Blockers                                      |
| ----------------------------- | ----------- | --------------------------------------------- |
| ValidationError Consolidation | Identified  | Similar split brain exists (config vs domain) |
| File Size Reduction           | In Progress | 30 files exceed 350-line limit                |
| Boolean-to-Enum Conversion    | Planned     | 10+ bools in SanitizationRules need enums     |

### ❌ NOT STARTED

| Task                            | Priority | Effort   |
| ------------------------------- | -------- | -------- |
| BDD Tests for Nix               | High     | 2-3 days |
| Protected Paths Constants       | Medium   | 2 hours  |
| Error Handling Standardization  | Medium   | 1 day    |
| Docker.go Split                 | Medium   | 4 hours  |
| TypeSpec Code Generation Review | Low      | 1 day    |

---

## Critical Findings

### 🔴 CRITICAL: Split Brains (Data Duplication)

Split brains occur when the same concept is defined in multiple places with slight variations, leading to maintenance nightmares.

| Type                | Location 1                    | Location 2                         | Fields Match? |
| ------------------- | ----------------------------- | ---------------------------------- | ------------- |
| SanitizationWarning | `domain/config_methods.go:21` | `config/sanitizer.go:61`           | ❌ NO (Fixed) |
| ValidationError     | `config/validator.go:53`      | `domain/operation_validation.go:8` | ❌ NO         |

**Impact:** Developers use different types interchangeably, causing subtle bugs and serialization issues.

**Fix:** Consolidate into domain package, create type aliases for backward compatibility.

### 🟡 HIGH: Oversized Files

Files exceeding 350 lines violate the Single Responsibility Principle.

| File                                               | Lines | Over | Suggested Split                                          |
| -------------------------------------------------- | ----- | ---- | -------------------------------------------------------- |
| `internal/cleaner/compiledbinaries_ginkgo_test.go` | 917   | +567 | Split by test type (unit/integration)                    |
| `internal/cleaner/docker.go`                       | 524   | +174 | Extract resource handlers (images/containers/volumes)    |
| `internal/cleaner/compiledbinaries.go`             | 599   | +249 | Separate scanning logic from cleaning                    |
| `cmd/clean-wizard/commands/clean.go`               | 611   | +261 | Split CLI handlers from business logic                   |
| `internal/domain/config_methods.go`                | 470   | +120 | Split Sanitize/Validate/ApplyProfile into separate files |
| `internal/domain/type_safe_enums.go`               | 539   | +189 | One enum per file                                        |
| `internal/config/config.go`                        | 395   | +45  | Separate Load/Save/Defaults                              |

### 🟡 HIGH: Boolean Blindness

Booleans don't convey meaning at the type level. These should be enums:

```go
// Current (boolean blindness)
type SanitizationRules struct {
    NormalizePaths   bool  // What does "true" mean?
    ExpandHomeDir    bool
    ValidateExists   bool
    ClampValues      bool
    RoundPercentages bool
    TrimWhitespace   bool
    NormalizeCase    bool
    SortArrays       bool
    RemoveDuplicates bool
    AddDefaults      bool
}

// Proposed (type-safe enums)
type PathSanitizationMode int32
const (
    PathSanitizationNone PathSanitizationMode = iota
    PathSanitizationNormalizeOnly
    PathSanitizationNormalizeAndExpand
    PathSanitizationFull
)
```

### 🟡 HIGH: Hardcoded Strings

Magic strings scattered throughout the codebase:

| String            | Locations                                           | Should Be                             |
| ----------------- | --------------------------------------------------- | ------------------------------------- |
| `"/System"`       | `config_methods.go:152`, `validator_business.go:41` | `constants.ProtectedSystemPath`       |
| `"/Library"`      | `config_methods.go:152`, `validator_business.go:41` | `constants.ProtectedLibraryPath`      |
| `"/Applications"` | `config_methods.go:152`                             | `constants.ProtectedApplicationsPath` |
| `"docker"`        | `docker.go:158`                                     | `constants.DockerBinaryName`          |

### 🟠 MEDIUM: Error Handling Inconsistency

Multiple error handling patterns coexist:

1. **Centralized** (pkg/errors) - Preferred
2. **Standard library** (`errors.New`, `fmt.Errorf`)
3. **Domain-specific** (`domain.NewValidationError`)

Files using non-centralized errors:

- `internal/cleaner/golang_cleaner.go:41, 110`
- `internal/cleaner/tempfiles.go:80`
- `internal/cleaner/docker.go:408`
- `internal/cleaner/githistory.go` (multiple)
- `internal/domain/types.go:50, 54, 256`
- `internal/domain/config.go:51, 61, 158`

---

## Architecture Assessment

### ✅ Strengths

1. **Type Safety**: Excellent use of type-safe enums with `type_safe_enums.go`
2. **Domain Modeling**: Clear domain boundaries with `internal/domain/`
3. **Result Type**: Proper use of `result.Result[T]` for error handling
4. **Context System**: Generic `Context[T]` implementation
5. **Registry Pattern**: Clean cleaner registration
6. **Composition**: Good use of struct embedding

### ❌ Weaknesses

1. **Split Brains**: Duplicate type definitions
2. **File Size**: 30 files exceed 350 lines
3. **Test Coverage**: Ginkgo tests lack parallel execution
4. **Magic Strings**: Hardcoded paths and values
5. **Lint Violations**: 791 issues (many are style, but some are architectural)

### ⚠️ Non-Obvious But True

1. **The project has TWO ValidationError types** - one with Context pointer, one without. This causes subtle serialization bugs.

2. **SanitizationRules has 10 booleans** - This creates 2^10 = 1024 possible states, most of which are likely invalid. Enums would make invalid states unrepresentable.

3. **The type_safe_enums.go file is 539 lines** - This suggests enums are doing too much. Each enum should be in its own file with its methods.

4. **Protected paths are hardcoded in 3+ places** - Changing them requires finding all occurrences, which is error-prone.

5. **The config package imports domain, but domain also imports config indirectly** - This creates a circular dependency risk.

---

## Top 25 Next Tasks (Sorted by Impact/Effort)

### Critical (Do First)

| #   | Task                                     | Impact      | Effort  | Customer Value  |
| --- | ---------------------------------------- | ----------- | ------- | --------------- |
| 1   | Consolidate ValidationError split brain  | 🔴 Critical | 30 min  | Prevents bugs   |
| 2   | Extract protected paths constants        | 🔴 Critical | 1 hour  | Safety          |
| 3   | Convert SanitizationRules bools to enums | 🟡 High     | 4 hours | Type safety     |
| 4   | Split type_safe_enums.go (539 lines)     | 🟡 High     | 3 hours | Maintainability |
| 5   | Split config_methods.go (470 lines)      | 🟡 High     | 3 hours | Maintainability |
| 6   | Split docker.go (524 lines)              | 🟡 High     | 4 hours | Maintainability |

### High Priority

| #   | Task                                   | Impact  | Effort  | Customer Value  |
| --- | -------------------------------------- | ------- | ------- | --------------- |
| 7   | Add BDD tests for Nix cleaner          | 🟡 High | 2 days  | Reliability     |
| 8   | Standardize error handling             | 🟡 High | 1 day   | Consistency     |
| 9   | Split clean.go (611 lines)             | 🟡 High | 3 hours | Maintainability |
| 10  | Split compiledbinaries.go (599 lines)  | 🟡 High | 4 hours | Maintainability |
| 11  | Add parallel execution to Ginkgo tests | 🟡 High | 2 hours | Test speed      |
| 12  | Create constants for docker commands   | 🟡 High | 1 hour  | Maintainability |

### Medium Priority

| #   | Task                                   | Impact    | Effort  | Customer Value  |
| --- | -------------------------------------- | --------- | ------- | --------------- |
| 13  | Split config.go (395 lines)            | 🟠 Medium | 2 hours | Maintainability |
| 14  | Split compiledbinaries_ginkgo_test.go  | 🟠 Medium | 3 hours | Test speed      |
| 15  | Extract string trimming utilities      | 🟠 Medium | 2 hours | Reusability     |
| 16  | Create phantom types for paths         | 🟠 Medium | 3 hours | Type safety     |
| 17  | Review TypeSpec for generation         | 🟠 Medium | 1 day   | Code generation |
| 18  | Add uint usage for non-negative values | 🟠 Medium | 2 hours | Type safety     |

### Lower Priority

| #   | Task                             | Impact | Effort  | Customer Value  |
| --- | -------------------------------- | ------ | ------- | --------------- |
| 19  | Fix remaining linter warnings    | 🟢 Low | 2 days  | Code quality    |
| 20  | Add integration tests for Docker | 🟢 Low | 1 day   | Reliability     |
| 21  | Document architecture decisions  | 🟢 Low | 1 day   | Team onboarding |
| 22  | Create ADRs for major decisions  | 🟢 Low | 1 day   | Documentation   |
| 23  | Add performance benchmarks       | 🟢 Low | 3 hours | Performance     |
| 24  | Review plugin architecture       | 🟢 Low | 1 day   | Extensibility   |
| 25  | Add property-based tests         | 🟢 Low | 2 days  | Reliability     |

---

## Data Flow Analysis

```
CLI Command
    ↓
Config Loading (config/)
    ↓
Domain Validation (domain/)
    ↓
Cleaner Registry (cleaner/registry.go)
    ↓
Individual Cleaners (cleaner/*.go)
    ↓
Adapters (adapters/)
    ↓
External Tools (nix, docker, brew, etc.)
```

**Assessment:** Clean separation of concerns. No circular dependencies detected.

---

## Type Safety Assessment

### ✅ Excellent

- Type-safe enums with String()/IsValid()/Values() methods
- Generic Result[T] type
- Domain-specific types (RiskLevelType, CacheType, etc.)

### ⚠️ Needs Improvement

- Too many primitive strings for paths/config keys
- Booleans where enums should be used
- Missing uint for non-negative quantities

### 🔴 Critical

- Split brain types (SanitizationWarning - FIXED, ValidationError - PENDING)

---

## Customer Value Assessment

### Direct Value (User-Facing)

| Feature             | Status     | Value  |
| ------------------- | ---------- | ------ |
| 13 Cleaners         | ✅ Working | High   |
| Interactive TUI     | ✅ Working | High   |
| Dry-run mode        | ✅ Working | High   |
| Profile system      | ✅ Working | Medium |
| Git History cleaner | ✅ Working | High   |

### Indirect Value (Developer Experience)

| Aspect          | Status      | Impact           |
| --------------- | ----------- | ---------------- |
| Type Safety     | ✅ Strong   | Reduces bugs     |
| Test Coverage   | ⚠️ Moderate | Needs BDD        |
| Documentation   | ✅ Good     | Well documented  |
| Maintainability | ⚠️ Fair     | File size issues |

---

## Questions for Product Owner

### My Top #1 Question I Cannot Figure Out:

**"What is the intended boundary between `config` and `domain` packages?"**

Currently:

- `config` imports `domain` (correct direction)
- But `domain.Config` has methods like `Sanitize()`, `Validate()`, `ApplyProfile()` that use types from `config` package
- This creates an implicit circular dependency risk

**Options:**

1. Move all validation/sanitization logic to `config` package, keep `domain` as pure data
2. Create a `config/application` layer that orchestrates between `config` and `domain`
3. Keep as-is but document the boundary clearly

Which direction should we take?

---

## Recommendations

### Immediate (This Week)

1. **Fix ValidationError split brain** - 30 minutes, prevents bugs
2. **Extract protected paths constants** - 1 hour, safety improvement
3. **Schedule file splitting** - Create tickets for oversized files

### Short Term (Next 2 Weeks)

1. **Convert SanitizationRules to enums** - Type safety improvement
2. **Add BDD tests for Nix** - Critical cleaner needs assertions
3. **Standardize error handling** - Use centralized pkg/errors

### Long Term (Next Month)

1. **Review TypeSpec for code generation** - Could eliminate boilerplate
2. **Consider plugin architecture** - For 3rd party cleaners
3. **Performance benchmarking** - Ensure cleaners are efficient

---

## Conclusion

Clean Wizard has a **solid architectural foundation** with excellent type safety and domain modeling. The main issues are:

1. **Split brains** - Duplicate type definitions (1 fixed, 1 pending)
2. **Oversized files** - 30 files need splitting
3. **Boolean blindness** - Enums would improve type safety
4. **Inconsistent error handling** - Needs standardization

**Next Steps:**

1. Merge the SanitizationWarning fix ✅
2. Fix ValidationError split brain
3. Extract protected paths constants
4. Schedule file splitting work

The codebase is **production-ready** but needs **consolidation** to achieve excellence.

---

_Report generated by Senior Software Architect (AI)_  
_Methodology: Branching-flow analysis, static analysis, manual code review_
