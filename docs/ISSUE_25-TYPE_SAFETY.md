# Issue #25: ELIMINATE MAP[STRING]ANY CANCER

## 🚨 PRIORITY: CRITICAL

## 📋 ISSUE SUMMARY
**Title**: ELIMINATE MAP[STRING]ANY CANCER  
**Type**: TECHNICAL DEBT • ARCHITECTURAL IMPROVEMENT  
**Impact**: PRODUCTION TYPE SAFETY • RUNTIME ERROR PREVENTION  
**Estimate**: 2 hours

## 🎯 PROBLEM STATEMENT
Critical map[string]any instances remain throughout infrastructure layer, creating runtime type safety vulnerabilities and making it impossible to enforce compile-time invariants.

**Current Status**: 27% COMPLETE (48 → 35 instances eliminated)

## 📊 IMPACT ANALYSIS

### **🚨 HIGH RISK AREAS**:
- **cmd/ layer**: CLI command arguments use map[string]any (type safety vulnerability)
- **internal/middleware/**: Request/response contexts use map[string]any (runtime errors)
- **internal/config/**: Validation results still contain map[string]any (type leakage)
- **internal/cleaner/**: Operation parameters use map[string]any (execution safety)
- **internal/domain/interfaces/**: Generic interfaces return map[string]any (contract violation)
- **internal/pkg/errors/**: Error contexts use map[string]any (debugging issues)

### **🔥 BUSINESS IMPACT**:
- **Runtime Errors**: map[string]any allows invalid states to reach production
- **Debugging Difficulty**: No compile-time guarantees for data structures
- **IDE Support**: Poor autocomplete and type checking
- **Refactoring Risk**: Changes can silently break type contracts

## 🎯 ACCEPTANCE CRITERIA
- [ ] Zero map[string]any instances in cmd/ layer
- [ ] Zero map[string]any instances in internal/ layer  
- [ ] All configuration types strongly typed
- [ ] All adapter outputs properly typed
- [ ] Zero compilation errors
- [ ] All existing tests passing

## 🏗️ IMPLEMENTATION PLAN

### **Phase 1: cmd/ Layer Type Safety** (45 min)
**Files Affected**:
- `cmd/clean-wizard/commands/clean.go`
- `cmd/clean-wizard/commands/scan.go`

**Tasks**:
1. Replace command argument maps with typed structs
2. Add typed command options for all CLI commands
3. Update command parsing to use typed inputs
4. Add command validation with typed rules

**Expected Reduction**: 8 instances

### **Phase 2: Middleware Layer Type Safety** (45 min)
**Files Affected**:
- `internal/middleware/validation.go`
- `internal/middleware/validation_test.go`

**Tasks**:
1. Replace middleware map[string]any with typed contexts
2. Add typed validation result structures
3. Update error handling to use typed errors
4. Add middleware type safety tests

**Expected Reduction**: 6 instances

### **Phase 3: Config Layer Type Safety** (30 min)
**Files Affected**:
- `internal/config/validation_types.go`
- `internal/config/enhanced_loader.go`

**Tasks**:
1. Replace validation result map[string]any with typed structs
2. Update configuration loading to use typed results
3. Add configuration type safety tests
4. Remove type leakage in validation

**Expected Reduction**: 5 instances

### **Phase 4: Adapter Layer Type Safety** (30 min)
**Files Affected**:
- `internal/cleaner/nix.go`
- `internal/domain/interfaces.go`

**Tasks**:
1. Replace cleaner output maps with typed results
2. Add typed Nix command interfaces
3. Update domain interfaces to use typed contracts
4. Add adapter type safety tests

**Expected Reduction**: 6 instances

## 🧪 TESTING STRATEGY
1. **Unit Tests**: Ensure each typed module compiles and functions correctly
2. **Integration Tests**: Verify end-to-end workflows still function
3. **Type Safety Tests**: Use Go's type checker to verify no map[string]any
4. **Regression Tests**: Ensure existing functionality is preserved

## 📈 SUCCESS METRICS
- **Type Safety**: 95% → 100% (final 5% improvement)
- **map[string]any Instances**: 35 → 0 (100% elimination)
- **Compilation Safety**: Zero type-related errors
- **Runtime Safety**: Zero type-related runtime errors
- **Developer Experience**: Full IDE support and autocomplete

## 🔗 DEPENDENCIES
- ✅ Typed Settings Implementation (COMPLETED)
- ✅ Domain Layer Type Safety (COMPLETED)
- ✅ Validation Framework (COMPLETED)
- 🔄 Integration Test Suite (Issue #26)

## 📋 NOTES
This is the **final frontier** in achieving complete type safety. Each remaining map[string]any instance represents a potential runtime error and loss of compile-time guarantees.

**Risk Level**: LOW (changes are isolated and well-understood)
**Rollback Plan**: Complete safety with typed settings already available

## 🎯 DEFINITION OF DONE
- [ ] grep -r "map\[string\]any" . --include="*.go" returns 0 instances
- [ ] go test ./... passes completely
- [ ] go build ./... completes without errors
- [ ] All CLI commands function with typed inputs
- [ ] Integration tests validate end-to-end functionality

---

**Issue Created**: 2025-11-10  
**Milestone**: v0.2.0 Type Safety Excellence  
**Assignee**: TBD  
**Labels**: technical-debt, type-safety, high-priority