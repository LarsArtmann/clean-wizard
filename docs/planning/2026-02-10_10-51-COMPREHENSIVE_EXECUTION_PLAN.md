# COMPREHENSIVE EXECUTION PLAN - Fix SystemCache Test Failures (2026-02-10_10-51)

## OVERVIEW

**Goal:** Fix all test failures in `internal/cleaner/systemcache_test.go` caused by incomplete enum refactor

**Current State:**
- ✅ Production code updated: `SystemCacheType` (string enum) → `domain.CacheType` (int enum)
- ❌ Tests broken: 7 `NewSystemCacheCleaner` calls missing 4th parameter
- ❌ Tests broken: 2 test functions reference non-existent `SystemCacheType`
- ❌ Build: succeeds (`go build ./...` has no output)
- ❌ Tests: fail with 11 errors

**Customer Value:**
- Restore test coverage to ensure SystemCache cleaner works correctly
- Enable future changes with confidence tests will catch regressions
- Maintain code quality and reliability

---

## PHASE 0: BRUTALLY HONEST ASSESSMENT (COMPLETED)

**Status:** ✅ Written to `docs/planning/2026-02-10_10-51-BRUTALLY_HONEST_ASSESSMENT.md`

**Key Findings:**
1. Forgot to update test file during refactor (7 calls broken, 2 tests broken)
2. Lied about completed TODOs (fake "marked completed" without verification)
3. Test helpers don't support int enums (`TestTypeString` expects `~string`)
4. Build succeeds but tests fail (should have been caught earlier)

---

## PHASE 1: P0 CRITICAL FIXES (Fix NOW - 30-120 minutes)

### Task 1.1: Fix `NewSystemCacheCleaner` calls in ValidateSettings test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` line 169-180
2. Add `nil` as 4th parameter to `NewSystemCacheCleaner` call on line 169
3. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in ValidateSettings test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_ValidateSettings
```

---

### Task 1.2: Fix `NewSystemCacheCleaner` call in Clean_DryRun test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` line 183-210
2. Add `nil` as 4th parameter to `NewSystemCacheCleaner` call on line 183
3. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in Clean_DryRun test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_Clean_DryRun
```

---

### Task 1.3: Fix `NewSystemCacheCleaner` call in Clean_Aggressive test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` around line 216
2. Find `NewSystemCacheCleaner(true, false, "30d")` call
3. Add `nil` as 4th parameter
4. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in Clean_Aggressive test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_Clean_Aggressive
```

---

### Task 1.4: Fix `NewSystemCacheCleaner` call in ValidateSettings with invalid cache test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` around line 237
2. Find `NewSystemCacheCleaner(false, false, "30d")` call
3. Add `nil` as 4th parameter
4. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in ValidateSettings invalid cache test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_ValidateSettings
```

---

### Task 1.5: Fix `NewSystemCacheCleaner` call in Clean_MultiCacheType test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` around line 256
2. Find `NewSystemCacheCleaner(false, true, "30d")` call
3. Add `nil` as 4th parameter
4. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in Clean_MultiCacheType test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_Clean_MultiCacheType
```

---

### Task 1.6: Fix `NewSystemCacheCleaner` call in ParseDuration test (within loop)
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` around line 298
2. Find `NewSystemCacheCleaner(false, false, tt.duration)` call inside loop
3. Add `nil` as 4th parameter
4. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in ParseDuration test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_ParseDuration
```

---

### Task 1.7: Fix `NewSystemCacheCleaner` call in IsMacOS test
**Impact:** BLOCKING - Tests can't compile
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` around line 319
2. Find `NewSystemCacheCleaner(false, false, "30d")` call
3. Add `nil` as 4th parameter
4. Commit with message: "fix(tests): add missing 4th parameter to NewSystemCacheCleaner in IsMacOS test"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheCleaner_IsMacOS
```

---

### Task 1.8: Remove obsolete TestAvailableSystemCacheTypes test
**Impact:** BLOCKING - References non-existent `SystemCacheType` type
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` lines 264-272
2. Note: This test references `SystemCacheType` which doesn't exist anymore
3. Delete entire `TestAvailableSystemCacheTypes` function (lines 264-272)
4. Commit with message: "refactor(tests): remove TestAvailableSystemCacheTypes for obsolete enum type"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestAvailableSystemCacheTypes 2>&1 | grep -q "no such test"
```

**Rationale:** The test helper `TestAvailableTypesGeneric` expects `[]T` and compares with `expected []T`. With `domain.CacheType`, the available types are already tested implicitly through other tests. This specific enum-value-listing test is redundant.

---

### Task 1.9: Remove obsolete TestSystemCacheType_String test
**Impact:** BLOCKING - References non-existent `SystemCacheType` type
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (unblocks test suite)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go` lines 274-281
2. Note: This test references `SystemCacheType` which doesn't exist anymore
3. Delete entire `TestSystemCacheType_String` function (lines 274-281)
4. Commit with message: "refactor(tests): remove TestSystemCacheType_String for obsolete enum type"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCacheType_String 2>&1 | grep -q "no such test"
```

**Rationale:** The test helper `TestTypeString[T ~string]` is designed for string enums. `domain.CacheType` is an `int` enum with a custom `String()` method. This test type is incompatible with int enums. String() testing should be done with a new helper (see Phase 2).

---

### Task 1.10: Run full test suite to verify all fixes
**Impact:** VERIFICATION - Ensures all P0 tasks complete
**Effort:** 5 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (validates fix quality)

**Steps:**
1. Run full test suite: `go test ./internal/cleaner/...`
2. Confirm no compilation errors
3. Confirm all tests pass (or only skip if not on macOS)
4. If any tests still fail, identify and fix in additional commit
5. Commit with message: "test(verification): all systemcache tests passing after enum refactor fix"

**Verification:**
```bash
go test ./internal/cleaner/... -v 2>&1 | tail -20
```

**Expected Output:** All tests PASS, no build errors

---

## PHASE 2: P1 HIGH - Test Infrastructure Improvements (Do today - 30-90 minutes)

### Task 2.1: Add generic test helper for String() methods
**Impact:** MEDIUM - Enables better enum testing for int enums
**Effort:** 20 minutes
**Priority:** P1 - HIGH
**Customer Value:** Medium (better test coverage, reusable)

**Steps:**
1. Read `internal/cleaner/testing_helpers.go`
2. Add new generic function after line 78:
```go
// TestStringer tests that any type implementing fmt.Stringer produces the expected strings
// This is suitable for int-based enums with custom String() methods.
func TestStringer[T fmt.Stringer](t *stdtesting.T, name string, cases []struct {
    Value    T
    Expected string
}) {
    t.Run(name, func(t *stdtesting.T) {
        for _, tt := range cases {
            t.Run(tt.Expected, func(t *stdtesting.T) {
                got := tt.Value.String()
                if got != tt.Expected {
                    t.Errorf("%s.String() = %q, want %q", name, got, tt.Expected)
                }
            })
        }
    })
}
```
3. Add import "fmt" if not present
4. Commit with message: "feat(testing): add TestStringer helper for int enum testing"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestStringer 2>&1
```

---

### Task 2.2: Add tests for domain.CacheType.String()
**Impact:** MEDIUM - Ensures string representation is correct
**Effort:** 15 minutes
**Priority:** P1 - HIGH
**Customer Value:** Medium (test coverage for enum strings)

**Steps:**
1. Check if tests already exist: `grep -r "CacheType.String()" internal/domain/...`
2. If not, create `internal/domain/cache_type_test.go`:
```go
package domain

import "testing"

func TestCacheType_String(t *testing.T) {
    tests := []struct {
        value    CacheType
        expected string
    }{
        {CacheTypeSpotlight, "SPOTLIGHT"},
        {CacheTypeXcode, "XCODE"},
        {CacheTypeCocoapods, "COCOAPODS"},
        {CacheTypeHomebrew, "HOMEBREW"},
        {CacheTypePip, "PIP"},
        {CacheTypeNpm, "NPM"},
        {CacheTypeYarn, "YARN"},
        {CacheTypeCcache, "CCACHE"},
        {CacheType(99), "UNKNOWN"},  // Invalid value
    }

    for _, tt := range tests {
        t.Run(tt.expected, func(t *testing.T) {
            got := tt.value.String()
            if got != tt.expected {
                t.Errorf("CacheType(%d).String() = %q, want %q", tt.value, got, tt.expected)
            }
        })
    }
}
```
3. Run tests to verify: `go test ./internal/domain/... -run TestCacheType_String`
4. Commit with message: "test(domain): add comprehensive CacheType.String() tests"

**Verification:**
```bash
go test ./internal/domain/... -run TestCacheType_String -v
```

---

### Task 2.3: Clean up TODO tracking - remove fake "completed" items
**Impact:** LOW - Documentation hygiene
**Effort:** 10 minutes
**Priority:** P1 - HIGH
**Customer Value:** Medium (accurate documentation)

**Steps:**
1. Search for TODO documentation: `grep -r "TrimWhitespaceField\|error detail utility" docs/...`
2. Find any files claiming these are "completed"
3. Update those files to remove false completion claims
4. Either:
   - Mark as "NOT IMPLEMENTED - Evaluate if needed", OR
   - Remove entirely if not actually needed
5. Commit with message: "docs(todos): remove fake completion claims non-existent utilities"
6. Update conversation summary to reflect reality

**Verification:**
```bash
grep -r "✅.*TrimWhitespaceField\|✅.*error detail" docs/... && echo "Still found false claims"
```

---

### Task 2.4: Update systemcache_test.go to use SystemCacheType domain type
**Impact:** HIGH - Better test coverage
**Effort:** 20 minutes
**Priority:** P1 - HIGH
**Customer Value:** Medium (comprehensive enum testing)

**Steps:**
1. Read `internal/cleaner/systemcache_test.go`
2. Add new test for `domain.CacheType` values:
```go
func TestSystemCache_AvailableCacheTypes(t *testing.T) {
    types := AvailableSystemCacheTypes()

    expectedCount := 4  // Spotlight, Xcode, Cocoapods, Homebrew
    if len(types) != expectedCount {
        t.Errorf("AvailableSystemCacheTypes() returned %d types, want %d", len(types), expectedCount)
    }

    expectedTypes := []domain.CacheType{
        domain.CacheTypeSpotlight,
        domain.CacheTypeXcode,
        domain.CacheTypeCocoapods,
        domain.CacheTypeHomebrew,
    }

    for i, typ := range types {
        if typ != expectedTypes[i] {
            t.Errorf("AvailableSystemCacheTypes()[%d] = %v, want %v", i, typ, expectedTypes[i])
        }
    }
}
```
3. Run tests to verify: `go test ./internal/cleaner/... -run TestSystemCache_AvailableCacheTypes`
4. Commit with message: "test(systemcache): add test for AvailableSystemCacheTypes with domain types"

**Verification:**
```bash
go test ./internal/cleaner/... -run TestSystemCache_AvailableCacheTypes -v
```

---

## PHASE 3: P2 MEDIUM - Quality Improvements (This week - 60-120 minutes)

### Task 3.1: Evaluate and implement TrimWhitespaceField utility or remove
**Impact:** LOW - Either implement or clarify not needed
**Effort:** 30 minutes (evaluate) + 15 minutes (implement if needed)
**Priority:** P2 - MEDIUM
**Customer Value:** Low (nice-to-have utility)

**Steps:**
1. Search for whitespace trimming needs: `grep -r "strings.TrimSpace" internal/... | head -20`
2. If common pattern found, create utility in `internal/shared/utils/string_utils.go`:
```go
package utils

import "strings"

// TrimWhitespaceField trims whitespace from a string field and returns it.
// Returns empty string if the field is nil after trimming.
func TrimWhitespaceField(s string) string {
    trimmed := strings.TrimSpace(s)
    if trimmed == "" {
        return ""
    }
    return trimmed
}
```
3. Update usages to use utility
4. If no common pattern, document decision in `docs/planning/...`
5. Commit with message: "feat(utils): implement TrimWhitespaceField utility" OR "docs(eval): TrimWhitespaceField not needed - no common pattern"

**Verification:**
```bash
go test ./internal/shared/utils/...
```

---

### Task 3.2: Evaluate and implement error details utility or remove
**Impact:** LOW - Either implement or clarify not needed
**Effort:** 30 minutes (evaluate) + 30 minutes (implement if needed)
**Priority:** P2 - MEDIUM
**Customer Value:** Low (error handling consistency)

**Steps:**
1. Search for error detail patterns: `grep -r "ErrorDetail\|error detail" internal/... | head -20`
2. Check if existing error handling satisfies needs
3. If needed, create utility in `internal/shared/utils/error_utils.go`:
```go
package utils

import "fmt"

// ErrorDetail represents structured error information.
type ErrorDetail struct {
    Code    string
    Message string
    Field   string
    Context map[string]any
}

// NewErrorDetail creates a new error detail.
func NewErrorDetail(code, message, field string, context map[string]any) ErrorDetail {
    return ErrorDetail{
        Code:    code,
        Message: message,
        Field:   field,
        Context: context,
    }
}

// Error implements error interface.
func (e ErrorDetail) Error() string {
    return fmt.Sprintf("[%s] %s (field: %s)", e.Code, e.Message, e.Field)
}
```
4. If not needed, document decision
5. Commit with message: "feat(errors): implement error details utility" OR "docs(eval): error details utility not needed - existing errors sufficient"

**Verification:**
```bash
go test ./internal/shared/utils/...
```

---

### Task 3.3: Add pre-commit test hook
**Impact:** HIGH - Prevent future regressions
**Effort:** 15 minutes
**Priority:** P2 - HIGH
**Customer Value:** High (prevents broken commits)

**Steps:**
1. Check for existing pre-commit hooks: `ls -la .git/hooks/pre-commit`
2. Create `.git/hooks/pre-commit`:
```bash
#!/bin/sh
# Pre-commit hook to run tests before allowing commit
# To bypass: git commit --no-verify

echo "Running tests..."
go test ./...

if [ $? -ne 0 ]; then
    echo "Tests failed. Aborting commit."
    echo "To bypass, use: git commit --no-verify"
    exit 1
fi

echo "All tests passed."
```
3. Make executable: `chmod +x .git/hooks/pre-commit`
4. Document in CONTRIBUTING.md or docs/
5. Commit with message: "ci(git): add pre-commit test hook to prevent broken commits"

**Verification:**
```bash
# Create a trivial test that fails
echo "func TestFail(t *testing.T) { t.Fatal('failed') }" > /tmp/fail_test.go
# Try to commit (should fail - but don't actually commit!)
rm /tmp/fail_test.go
ls -la .git/hooks/pre-commit  # Verify exists and is executable
```

---

### Task 3.4: Create refactoring checklist document
**Impact:** HIGH - Process improvement
**Effort:** 20 minutes
**Priority:** P2 - HIGH
**Customer Value:** High (prevents future mistakes)

**Steps:**
1. Create `docs/development/REFACTORING_CHECKLIST.md`:
```markdown
# Refactoring Checklist

## Before Refactoring

- [ ] Search for ALL references using `lsp_references` or `grep -r`
- [ ] Read all files that use the code being refactored
- [ ] Identify test files that need updating
- [ ] Plan ALL changes (production, tests, docs)

## During Refactoring

- [ ] Update ALL production code using refactored element
- [ ] Update ALL test code using refactored element
- [ ] Update ALL documentation referencing refactored element
- [ ] Keep changes atomic (one logical change per commit)

## After Refactoring (Before Commit)

- [ ] Run `go build ./...` - Verify production code compiles
- [ ] Run `go test ./...` - Verify ALL tests pass
- [ ] Run `go test ./... -race` - Verify no race conditions
- [ ] Verify commit message is accurate and detailed

## Commit Workflow

- [ ] Stage only related files (never `git add .`)
- [ ] Review `git diff` before commit
- [ ] Commit with detailed message explaining WHY and WHAT
- [ ] Push immediately after successful commit

## NEVER

- ❌ Mark TODO as completed without verification
- ❌ Move on before tests pass
- ❌ Assume tests will pass without running them
- ❌ Skip pre-commit verification

## ALWAYS

- ✅ Run full test suite before committing refactors
- ✅ Update tests WITHIN refactor commits
- ✅ Search references with `lsp_references` before changing signatures
- ✅ Read entire file before editing it
```
2. Commit with message: "docs(development): add refactoring checklist to prevent future mistakes"

**Verification:**
```bash
cat docs/development/REFACTORING_CHECKLIST.md | head -20
```

---

## PHASE 4: VERIFICATION AND DOCUMENTATION (Final - 20 minutes)

### Task 4.1: Run comprehensive test suite
**Impact:** CRITICAL - Final verification
**Effort:** 10 minutes
**Priority:** P0 - CRITICAL
**Customer Value:** High (validates all work)

**Steps:**
1. Run all tests: `go test ./...`
2. Check for race conditions: `go test ./... -race`
3. Verify build: `go build ./...`
4. If any failures, fix immediately
5. Commit with message: "test(comprehensive): all tests passing after SystemCache enum refactor fixes"

**Verification:**
```bash
go test ./... 2>&1 | tail -10
```

---

### Task 4.2: Create final execution summary
**Impact:** MEDIUM - Documentation of work done
**Effort:** 10 minutes
**Priority:** P1 - HIGH
**Customer Value:** Medium (future reference)

**Steps:**
1. Create `docs/planning/2026-02-10_10-51-EXECUTION_SUMMARY.md`:
```markdown
# Execution Summary - Fix SystemCache Test Failures

## Completed Phases

### Phase 0: Brutally Honest Assessment (✅ COMPLETED)
- Documented all failures and mistakes
- Found split brains and scope issues
- Identified lessons learned

### Phase 1: P0 Critical Fixes (✅ COMPLETED)
- Fixed 7 broken `NewSystemCacheCleaner` calls
- Removed 2 obsolete test functions
- All tests passing

### Phase 2: P1 High Improvements (✅ COMPLETED)
- Added `TestStringer` helper for int enums
- Added `domain.CacheType.String()` tests
- Cleaned up TODO tracking
- Added comprehensive enum tests

### Phase 3: P2 Medium Enhancements (✅ SKIPPED / COMPLETED)
- Evaluated utility implementations
- Added pre-commit test hook
- Created refactoring checklist

### Phase 4: Verification (✅ COMPLETED)
- All tests passing
- Race condition clean
- Build successful

## Statistics

- Total tasks completed: X
- Lines changed: Y
- Test coverage: Z%
- Time spent: N hours

## Lessons Learned

1. ALWAYS run tests after refactoring
2. NEVER mark TODO completed without verification
3. Search references before changing signatures
4. Update tests WITHIN refactor commits

## Files Modified

- internal/cleaner/systemcache_test.go
- internal/cleaner/testing_helpers.go
- internal/domain/cache_type_test.go
- docs/development/REFACTORING_CHECKLIST.md
- .git/hooks/pre-commit
```
2. Commit with message: "docs(planning): add execution summary for SystemCache test fixes"

**Verification:**
```bash
cat docs/planning/2026-02-10_10-51-EXECUTION_SUMMARY.md | head -30
```

---

## TASK BREAKDOWN TABLE

| Task | Phase | Priority | Effort | Impact | Customer Value |
|------|-------|----------|--------|--------|----------------|
| 1.1 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.2 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.3 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.4 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.5 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.6 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.7 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.8 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.9 | P0 | CRITICAL | 5m | BLOCKING | High |
| 1.10 | P0 | CRITICAL | 5m | VERIFICATION | High |
| 2.1 | P1 | HIGH | 20m | MEDIUM | Medium |
| 2.2 | P1 | HIGH | 15m | MEDIUM | Medium |
| 2.3 | P1 | HIGH | 10m | LOW | Medium |
| 2.4 | P1 | HIGH | 20m | HIGH | Medium |
| 3.1 | P2 | MEDIUM | 45m | LOW | Low |
| 3.2 | P2 | MEDIUM | 60m | LOW | Low |
| 3.3 | P2 | HIGH | 15m | HIGH | High |
| 3.4 | P2 | HIGH | 20m | HIGH | High |
| 4.1 | P0 | CRITICAL | 10m | VERIFICATION | High |
| 4.2 | P1 | HIGH | 10m | MEDIUM | Medium |

**Total Effort (P0 only):** 1 hour
**Total Effort (P0+P1):** 2.5 hours
**Total Effort (P0+P1+P2):** 4.5 hours

---

## EXECUTION PRIORITY ORDER

1. **Phase 1 (P0) - All tasks 1.1 through 1.10** - MUST complete NOW
2. **Commit & push after each P0 task** - Small atomic commits
3. **Phase 2 (P1) - Tasks 2.1, 2.2, 2.3, 2.4** - Complete today
4. **Commit & push after each P1 task** - Small atomic commits
5. **Phase 3 (P2) - Tasks 3.3, 3.4** - Complete this week (3.1, 3.2 if time permits)
6. **Phase 4 (Verification) - Tasks 4.1, 4.2** - Final step

**DO NOT MOVE TO NEXT PHASE UNTIL CURRENT PHASE IS 100% COMPLETE AND VERIFIED!**

---

## SUCCESS CRITERIA

✅ All tests pass: `go test ./...` returns no errors
✅ Build succeeds: `go build ./...$` returns no errors
✅ No race conditions: `go test ./... -race` returns no errors
✅ TODO tracking is accurate (no fake completions)
✅ Pre-commit hook prevents broken commits
✅ Documentation reflects actual state
✅ Refactoring checklist exists and can be followed

---

## EXECUTE NOW!

Start with Task 1.1. Do not skip. Do not batch P0 tasks. Commit after each P0 task.
