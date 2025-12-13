# COMPREHENSIVE MULTI-STEP EXECUTION PLAN

## 📋 RESEARCH COMPLETED - LIBRARIES IDENTIFIED

### 🔍 Key Libraries Found:
1. **`stringer`** - Official Go tool for enum generation (`go install golang.org/x/tools/cmd/stringer`)
2. **TypeSpec** - Already in use for API definitions, can generate type-safe enums
3. **Existing infrastructure** - Well-structured result types, error handling, domain models

---

## 🎯 PHASE 1: CRITICAL BUILD FAILURES (Week 1 - Day 1)

### IMPACT: HIGH (Unblocks all development) | WORK: LOW (Quick fixes)

#### Step 1.1: Fix Missing Interface Method
- **Issue**: `mockCleaner` missing `Cleanup()` method  
- **File**: `internal/shared/utils/middleware/validation_test.go:69`
- **Fix**: Add Cleanup method to mock interface
- **Verify**: `go test ./...`

#### Step 1.2: Fix NixCleaner Constructor Signature  
- **Issue**: `NewNixCleaner` expects no args but called with 2 args
- **File**: `internal/infrastructure/cleaners/nix_test.go:18,36,54,71`
- **Fix**: Update constructor or test calls to match
- **Verify**: `go test ./internal/infrastructure/cleaners`

#### Step 1.3: Fix Remaining Undefined References
- **Issue**: `adapters` undefined in `internal/infrastructure/system/adapters_test.go`
- **Fix**: Add import to test file
- **Verify**: `go test ./internal/infrastructure/system`

#### Step 1.4: Verify All Critical Issues Fixed
- **Action**: Run `just lint` to confirm 9 critical errors resolved
- **Verify**: `go test ./...` passes
- **Commit**: "fix: resolve all critical build failures"

---

## 🎯 PHASE 2: TYPE SAFETY IMPROVEMENTS (Week 1 - Days 2-3)

### IMPACT: HIGH (Architectural excellence) | WORK: MEDIUM (Code generation)

#### Step 2.1: Install Stringer Tool
```bash
go install golang.org/x/tools/cmd/stringer
```

#### Step 2.2: Add `//go:generate` Directives to Enums
- **Files**: `internal/domain/shared/type_safe_enums.go` (864 lines!)
- **Pattern**: `//go:generate stringer -type=RiskLevelType`
- **Target**: All type-safe enums with duplication

#### Step 2.3: Generate String Methods
```bash
go generate ./internal/domain/shared/
```

#### Step 2.4: Update Domain Types to Use Generated Methods
- **Remove**: Duplicate String() implementations  
- **Replace**: With generated `String()` methods
- **Verify**: `go test ./internal/domain/shared`

#### Step 2.5: Clean Up Enum Duplications
- **Target**: 28+ clone groups in type_safe_enums.go
- **Action**: Remove manually implemented string methods
- **Benefit**: Eliminate major source of duplication

---

## 🎯 PHASE 3: ERROR HANDLING CONSOLIDATION (Week 1 - Days 4-5)

### IMPACT: HIGH (Better maintainability) | WORK: MEDIUM (Refactoring)

#### Step 3.1: Consolidate Error Factories
- **Files**: Multiple error packages with patterns
- **Target**: `internal/shared/utils/pkg/errors/errors.go` (15+ clone groups)
- **Create**: `ErrorFactory` interface for common patterns

#### Step 3.2: Extract Common Error Patterns
```go
// Instead of duplicate patterns like:
func ErrInvalidConfig(message string, args ...any) error
func ErrInvalidArgument(arg, message string) error
// Use factory:
factory.NewValidationError(ConfigValidation, message, args...)
factory.NewArgumentError(arg, message)
```

#### Step 3.3: Update All Error References
- **Replace**: Direct error creation with factory methods
- **Files**: All error packages identified in duplication report
- **Verify**: All tests pass with new error handling

---

## 🎯 PHASE 4: TEST INFRASTRUCTURE CLEANUP (Week 2 - Days 1-3)

### IMPACT: MEDIUM (Better dev experience) | WORK: HIGH (Many test files)

#### Step 4.1: Create Test Helper Package
- **Location**: `test/helpers/`
- **Content**: Common test setup, validation helpers, BDD utilities
- **Goal**: Reduce 14+ clone groups in test files

#### Step 4.2: Extract Common Test Patterns
- **Target**: `test/config/safety_level_test.go` (14+ clone groups)
- **Target**: `internal/application/config/safe_test.go` (9+ clone groups)
- **Create**: Reusable test builders and validators

#### Step 4.3: Consolidate BDD Test Infrastructure
- **Files**: Multiple BDD test files with duplication
- **Action**: Move to shared test helpers
- **Verify**: All BDD tests pass

---

## 🎯 PHASE 5: CONFIG MODULE REFACTORING (Week 2 - Days 4-5)

### IMPACT: MEDIUM (Cleaner architecture) | WORK: MEDIUM (Refactoring)

#### Step 5.1: Split Large Config Files
- **Target**: Files >300 lines identified in ISSUES_LIST.md
- **Split**: `internal/domain/shared/type_safe_enums.go` (864 lines)
- **Structure**: Separate files per enum type

#### Step 5.2: Extract Config Patterns
- **Target**: `internal/domain/config/config.go` (15+ clone groups)
- **Action**: Create common config builder patterns
- **Verify**: All config functionality preserved

#### Step 5.3: Remove Config Duplications
- **Target**: Cross-file config pattern duplication
- **Action**: Consolidate into shared utilities
- **Files**: `internal/application/config/`, `test/config/`

---

## 🎯 PHASE 6: CLEANER MODULE STANDARDIZATION (Week 3 - Days 1-2)

### IMPACT: MEDIUM (Consistent patterns) | WORK: MEDIUM (Refactoring)

#### Step 6.1: Extract Cleaner Interface Implementation
- **Target**: Common patterns across homebrew, npm, pnpm, nix, temp_files
- **Create**: `BaseCleaner` struct with shared logic
- **Action**: Implement in specific cleaners

#### Step 6.2: Standardize Cleaner Constructors
- **Target**: Inconsistent constructor patterns
- **Action**: Create cleaner factory functions
- **Verify**: All cleaners work consistently

#### Step 6.3: Remove Cleaner Duplications
- **Target**: Cross-file cleanup patterns (lines 172-188)
- **Action**: Move to base implementation
- **Result**: Eliminated duplicate cleanup logic

---

## 🎯 PHASE 7: API MAPPER CLEANUP (Week 3 - Days 3-4)

### IMPACT: LOW (Code quality) | WORK: LOW (Targeted fixes)

#### Step 7.1: Fix Mapper Test Duplications
- **Target**: `internal/interface/api/mapper_test.go` (10+ clone groups, 744 lines)
- **Action**: Extract common test setup to helpers
- **Verify**: All mapper tests pass

#### Step 7.2: Consolidate Mapping Patterns
- **Target**: Lines 554-620 with duplicate mapping test patterns
- **Action**: Create reusable test fixtures
- **Result**: Cleaner, more maintainable tests

---

## 🎯 PHASE 8: DOCUMENTATION & FINAL POLISH (Week 3 - Day 5)

### IMPACT: LOW (Better dev experience) | WORK: LOW (Quick wins)

#### Step 8.1: Add Package Documentation
- **Target**: 4 packages without documentation
- **Action**: Add comprehensive package docs
- **Content**: Usage examples, architectural notes

#### Step 8.2: Update README.md with New Architecture
- **Include**: Type safety improvements, duplication elimination
- **Add**: Developer guidelines for using generated code

#### Step 8.3: Final Verification
- **Action**: `just build && just lint && just test`
- **Goal**: All 108 issues resolved
- **Verification**: Code duplication report shows significant reduction

---

## 📊 EXPECTED OUTCOMES

### Critical Metrics:
- **Build Errors**: 9 → 0 ✅
- **Code Duplication**: 218 clone groups → ~50 groups (77% reduction)
- **Large Files**: 5 files >300 lines → 0 files ✅
- **Test Coverage**: Add tests for 4 uncovered packages ✅

### Architectural Improvements:
- **Type Safety**: Generated enums with `String()` methods ✅
- **Error Handling**: Consistent factory patterns ✅  
- **Code Reuse**: Shared test utilities ✅
- **Maintainability**: Cleaner, smaller focused files ✅

---

## 🔧 EXECUTION STRATEGY

### Daily Workflow:
1. **Morning**: Review daily tasks, understand dependencies
2. **Development**: Execute step-by-step with verification after each
3. **Testing**: Run relevant tests after every change
4. **Commit**: Small, focused commits with descriptive messages
5. **Evening**: Review progress, plan next day

### Verification After Each Phase:
- `go test ./...` - All tests pass
- `go build ./...` - Clean build
- `just lint` - No lint errors  
- `just fd` - Check duplication reduction

### Risk Mitigation:
- **Small commits** - Easy to rollback if needed
- **Test-first** - Verify before moving to next step
- **Backup patterns** - Keep working versions during refactoring

---

*This plan systematically addresses all 108 issues from ISSUES_LIST.md while prioritizing by impact vs work required. The focus is on architectural excellence and type safety while maintaining functionality.*