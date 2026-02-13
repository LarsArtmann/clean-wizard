# COMPREHENSIVE MULTI-STEP EXECUTION PLAN

**Date:** 2026-02-10
**Status:** IN PROGRESS

---

## PART 1: ALREADY IMPLEMENTED (No Action Needed)

| Task                         | Status    | Location                                             |
| ---------------------------- | --------- | ---------------------------------------------------- |
| String Trimming Utility      | ✅ EXISTS | `internal/shared/utils/strings/trimming.go`          |
| Generic Validation Interface | ✅ EXISTS | `internal/shared/utils/validation/validation.go`     |
| Config Loading Utility       | ✅ EXISTS | `internal/shared/utils/config/config.go`             |
| Error Details Utility        | ✅ EXISTS | `internal/pkg/errors/detail_helpers.go`              |
| Generic Context System       | ✅ DONE   | `internal/shared/context/` (19 tests passing)        |
| CleanerRegistry Integration  | ✅ DONE   | `internal/cleaner/registry.go` (231 lines, 12 tests) |
| Deprecation Fixes            | ✅ DONE   | 49 warnings eliminated across 45+ files              |
| Docker Enum Refactoring      | ✅ DONE   | Commit 5e94e2a                                       |
| Binary Enum Unification      | ✅ DONE   | 69 lines removed, unified to `UnmarshalYAMLEnum`     |
| CLI Commands                 | ✅ DONE   | clean, scan, init, profile, config                   |

---

## PART 2: CRITICAL IMMEDIATE (Diagnostics & Quick Fixes)

### Task C1: Fix 14 Diagnostics Warnings (12 min)

**Priority:** 🔴 CRITICAL | **Impact:** 90% | **Effort:** 12 min

**Description:** Fix 14 instances of "unnecessary type arguments" in `context_test.go`

**Files:**

- `internal/shared/context/context_test.go` (lines 117, 152, 157, 198, 215, 232, 233, 234, 253, 286)

**Subtasks:**

- [ ] Fix type argument redundancy (12min)

**Verification:**

```bash
go build ./...
go test ./internal/shared/context/...
```

---

## PART 3: HIGH PRIORITY (Architectural Improvements)

### Task H1: NodePackages Enum Refactor (30 min total)

**Priority:** 🔴 HIGH | **Impact:** 80% | **Effort:** 30 min

**Description:** Refactor NodePackages from local string enum to domain `PackageManagerType` integer enum

**Files:**

- `internal/cleaner/nodepackages.go`

**Subtasks:**

- [ ] Research current NodePackages enum usage (5min)
- [ ] Refactor local enum to use domain PackageManagerType (15min)
- [ ] Update tests to match new constructor signature (10min)

**Verification:**

```bash
go build ./internal/cleaner/nodepackages.go
go test ./internal/cleaner/...
```

---

### Task H2: BuildCache Architectural Decision (15 min)

**Priority:** 🔴 HIGH | **Impact:** 70% | **Effort:** 15 min

**Description:** Decide between merging local BuildToolType (string, tools) with domain BuildToolType (int, languages)

**Files:**

- `internal/cleaner/buildcache.go`
- `internal/domain/operation_settings.go`

**Subtasks:**

- [ ] Research BuildCache current implementation (5min)
- [ ] Document decision options (5min)
- [ ] Implement decision (5min)

**Decision Options:**

1. **Keep separate** - Rename local to avoid confusion
2. **Merge** - Use domain enum exclusively
3. **Create mapping** - Map tools to languages

**Verification:**

```bash
go build ./internal/cleaner/buildcache.go
```

---

### Task H3: Eliminate Backward Compatibility Aliases (60 min total)

**Priority:** 🟠 HIGH | **Impact:** 70% | **Effort:** 60 min

**Description:** Remove duplicate type systems with phased migration

**Files:**

- `internal/domain/`

**Subtasks:**

**Phase 1: Analysis & Planning (10min)**

- [ ] Audit domain types for aliases

**Phase 2: Replace usages (30min)**

- [ ] Replace RiskLevel usages with RiskLevelType
- [ ] Replace Strategy usages with StrategyType
- [ ] Update all callers

**Phase 3: Remove aliases (20min)**

- [ ] Remove deprecated type aliases
- [ ] Verify build succeeds

**Verification:**

```bash
go build ./...
go test ./...
```

---

### Task H4: Domain Model Enhancement (45 min total)

**Priority:** 🟠 HIGH | **Impact:** 50% | **Effort:** 45 min

**Description:** Transform anemic domain models into rich domain objects

**Files:**

- `internal/domain/config.go`

**Subtasks:**

- [ ] Add Validate() method to Config (10min)
- [ ] Add Sanitize() method to Config (10min)
- [ ] Add ApplyProfile() method to Config (15min)
- [ ] Add EstimateImpact() method to Config (10min)

**Verification:**

```bash
go build ./internal/domain/
go test ./internal/domain/...
```

---

### Task H5: Type Model Improvements (30 min total)

**Priority:** 🟠 HIGH | **Impact:** 60% | **Effort:** 30 min

**Description:** Add IsValid(), Values(), consistent String() to all enums

**Files:**

- `internal/domain/interfaces.go`

**Subtasks:**

- [ ] Audit enums for missing methods (10min)
- [ ] Add missing IsValid() methods (10min)
- [ ] Add missing Values() methods (10min)

**Note:** Most enums already have these methods. Need to verify which are missing.

**Verification:**

```bash
go build ./internal/domain/
```

---

### Task H6: Result Type Enhancement (20 min)

**Priority:** 🟠 HIGH | **Impact:** 50% | **Effort:** 20 min

**Description:** Enhance Result type for better validation chaining

**Files:**

- `internal/result/type.go`

**Subtasks:**

- [ ] Add validation chaining methods (10min)
- [ ] Add tests for new methods (10min)

**Verification:**

```bash
go test ./internal/result/...
```

---

### Task H7: Create Schema Min/Max Utility (12 min)

**Priority:** 🟠 HIGH | **Impact:** 50% | **Effort:** 12 min

**Description:** Create utility to eliminate schema logic duplication

**Files:**

- `internal/shared/utils/schema/minmax.go` (NEW)

**Subtasks:**

- [ ] Create min/max schema utility (12min)

**Verification:**

```bash
go build ./internal/shared/utils/schema/
```

---

### Task H8: LoadWithContext Complexity Reduction (60 min total)

**Priority:** 🟠 HIGH | **Impact:** 70% | **Effort:** 60 min

**Description:** Refactor config.LoadWithContext from complexity 20 to <10

**Files:**

- `internal/config/config.go`

**Subtasks:**

- [ ] Analyze current complexity (5min)
- [ ] Extract profile loading logic (15min)
- [ ] Extract operation processing (15min)
- [ ] Extract risk level processing (15min)
- [ ] Use early returns pattern (10min)

**Verification:**

```bash
golangci-lint run ./internal/config/config.go
go test ./internal/config/...
```

---

### Task H9: validateProfileName Complexity Reduction (30 min total)

**Priority:** 🟠 HIGH | **Impact:** 60% | **Effort:** 30 min

**Description:** Refactor validateProfileName from complexity 16 to <10

**Files:**

- `internal/config/validator.go`

**Subtasks:**

- [ ] Analyze current complexity (5min)
- [ ] Refactor to use early returns (15min)
- [ ] Extract validation helpers (10min)

**Verification:**

```bash
golangci-lint run ./internal/config/validator.go
```

---

## PART 4: MEDIUM PRIORITY (Quality Improvements)

### Task M1: Additional Complexity Reductions (60 min)

**Priority:** 🟡 MEDIUM | **Impact:** 50% | **Effort:** 60 min

**Description:** Reduce complexity of 3+ functions (TestIntegration_ValidationSanitizationPipeline, ErrorCode.String, SaveConfig)

**Files:**

- Multiple files

**Subtasks:**

- [ ] Identify high-complexity functions (10min)
- [ ] Refactor TestIntegration function (20min)
- [ ] Refactor ErrorCode.String (15min)
- [ ] Refactor SaveConfig (15min)

**Verification:**

```bash
golangci-lint run ./...
```

---

### Task M2: BDD Helper Refactoring Analysis (30 min)

**Priority:** 🟡 MEDIUM | **Impact:** 40% | **Effort:** 30 min

**Description:** Analyze and refactor test helpers for duplication

**Files:**

- `tests/bdd/`

**Subtasks:**

- [ ] Analyze BDD helpers for duplication (15min)
- [ ] Refactor duplicates (15min)

**Verification:**

```bash
go test ./tests/bdd/...
```

---

### Task M3: SystemCache Research & Refactoring (60 min)

**Priority:** 🟡 MEDIUM | **Impact:** 50% | **Effort:** 60 min

**Description:** Research domain.CacheType usage and implement decision

**Files:**

- `internal/cleaner/systemcache.go`
- `internal/domain/operation_settings.go`

**Subtasks:**

- [ ] Research domain.CacheType usage (15min)
- [ ] Document findings (15min)
- [ ] Implement decision (30min)

**Verification:**

```bash
go build ./internal/cleaner/systemcache.go
go test ./internal/cleaner/...
```

---

### Task M4: RiskLevelType Manual Processing Investigation (60 min)

**Priority:** 🟡 MEDIUM | **Impact:** 40% | **Effort:** 60 min

**Description:** Investigate Viper enum support and fix manual processing

**Files:**

- `internal/config/config.go:86-108`

**Subtasks:**

- [ ] Research Viper enum support (30min)
- [ ] Test using RiskLevelType directly (15min)
- [ ] Fix if possible (15min)

**Verification:**

```bash
go build ./internal/config/
go test ./internal/config/...
```

---

## PART 5: STRATEGIC (Long-term Improvements)

### Task S1: Improve Size Reporting (120 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 60% | **Effort:** 120 min

**Description:** Replace hardcoded estimates with actual calculations

**Files:**

- Multiple cleaner files

**Subtasks:**

- [ ] Audit size reporting (30min)
- [ ] Fix Docker size reporting (30min)
- [ ] Fix Cargo size reporting (30min)
- [ ] Update remaining cleaners (30min)

**Verification:**

```bash
go build ./internal/cleaner/...
```

---

### Task S2: Add Linux Support for SystemCache (120 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 50% | **Effort:** 120 min

**Description:** Extend SystemCache cleaner to support Linux

**Files:**

- `internal/cleaner/systemcache.go`

**Subtasks:**

- [ ] Research Linux cache paths (30min)
- [ ] Implement Linux cache detection (60min)
- [ ] Add tests for Linux (30min)

**Verification:**

```bash
go build ./internal/cleaner/systemcache.go
go test ./internal/cleaner/...
```

---

### Task S3: Registry Documentation (30 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 30% | **Effort:** 30 min

**Description:** Document how to use CleanerRegistry

**Files:**

- `docs/registry.md` (NEW)

**Subtasks:**

- [ ] Create registry documentation (30min)

**Verification:**

```bash
ls docs/registry.md
```

---

### Task S4: Architecture Documentation (120 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 40% | **Effort:** 120 min

**Description:** Create ARCHITECTURE.md

**Files:**

- `ARCHITECTURE.md` (NEW)

**Subtasks:**

- [ ] Create architecture documentation (120min)

**Verification:**

```bash
ls ARCHITECTURE.md
```

---

### Task S5: Enum Quick Reference (30 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 30% | **Effort:** 30 min

**Description:** Create enum quick reference guide

**Files:**

- `docs/ENUM_QUICK_REFERENCE.md` (NEW)

**Subtasks:**

- [ ] Create enum reference (30min)

**Verification:**

```bash
ls docs/ENUM_QUICK_REFERENCE.md
```

---

### Task S6: Dependency Injection Investigation (120 min)

**Priority:** 🟢 STRATEGIC | **Impact:** 40% | **Effort:** 120 min

**Description:** Evaluate samber/do/v2 for DI

**Files:**

- `cmd/clean-wizard/`
- `go.mod`

**Subtasks:**

- [ ] Research samber/do/v2 (30min)
- [ ] Create DI container prototype (60min)
- [ ] Evaluate and document findings (30min)

**Verification:**

```bash
go build ./cmd/clean-wizard/
```

---

## EXECUTION SUMMARY TABLE

| ID  | Task                             | Priority     | Impact | Effort | Status     |
| --- | -------------------------------- | ------------ | ------ | ------ | ---------- |
| C1  | Fix Diagnostics Warnings         | 🔴 CRITICAL  | 90%    | 12min  | ⏳ PENDING |
| H1  | NodePackages Enum Refactor       | 🔴 HIGH      | 80%    | 30min  | ⏳ PENDING |
| H2  | BuildCache Decision              | 🔴 HIGH      | 70%    | 15min  | ⏳ PENDING |
| H3  | Eliminate Aliases (3 phases)     | 🟠 HIGH      | 70%    | 60min  | ⏳ PENDING |
| H4  | Domain Model Enhancement         | 🟠 HIGH      | 50%    | 45min  | ⏳ PENDING |
| H5  | Type Model Improvements          | 🟠 HIGH      | 60%    | 30min  | ⏳ PENDING |
| H6  | Result Type Enhancement          | 🟠 HIGH      | 50%    | 20min  | ⏳ PENDING |
| H7  | Schema Min/Max Utility           | 🟠 HIGH      | 50%    | 12min  | ⏳ PENDING |
| H8  | LoadWithContext Complexity       | 🟠 HIGH      | 70%    | 60min  | ⏳ PENDING |
| H9  | validateProfileName Complexity   | 🟠 HIGH      | 60%    | 30min  | ⏳ PENDING |
| M1  | Additional Complexity Reductions | 🟡 MEDIUM    | 50%    | 60min  | ⏳ PENDING |
| M2  | BDD Helper Refactoring           | 🟡 MEDIUM    | 40%    | 30min  | ⏳ PENDING |
| M3  | SystemCache Research             | 🟡 MEDIUM    | 50%    | 60min  | ⏳ PENDING |
| M4  | RiskLevelType Investigation      | 🟡 MEDIUM    | 40%    | 60min  | ⏳ PENDING |
| S1  | Size Reporting                   | 🟢 STRATEGIC | 60%    | 120min | ⏳ PENDING |
| S2  | Linux SystemCache                | 🟢 STRATEGIC | 50%    | 120min | ⏳ PENDING |
| S3  | Registry Documentation           | 🟢 STRATEGIC | 30%    | 30min  | ⏳ PENDING |
| S4  | Architecture Documentation       | 🟢 STRATEGIC | 40%    | 120min | ⏳ PENDING |
| S5  | Enum Quick Reference             | 🟢 STRATEGIC | 30%    | 30min  | ⏳ PENDING |
| S6  | DI Investigation                 | 🟢 STRATEGIC | 40%    | 120min | ⏳ PENDING |

**Total Tasks:** 20
**Total Effort:** ~14.5 hours
**Estimated Completion:** 2-3 weeks

---

## VERIFICATION COMMANDS

After each task, run:

```bash
# Always verify
go build ./...
go test ./...

# For complexity tasks
golangci-lint run ./...

# For specific packages
go test ./internal/cleaner/...
go test ./internal/domain/...
go test ./internal/config/...
go test ./internal/shared/context/...
go test ./internal/result/...
```

---

## NEXT STEPS

1. **Start with Task C1** - Fix diagnostics warnings (12min, immediate value)
2. **Then Task H1-H9** - High priority architectural tasks
3. **Then Task M1-M4** - Medium priority quality improvements
4. **Finally Task S1-S6** - Strategic improvements

---

**Generated:** 2026-02-10
**Next Review:** After completing all HIGH priority tasks
