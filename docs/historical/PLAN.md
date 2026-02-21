# Comprehensive Multi-Step Execution Plan

Generated: 2026-02-10

## Reflection on Current Work

### Completed This Session

1. **S1**: Added LegacySanitizationChange conversion to complete generic context system
2. **S2**: Changed RiskLevel → RiskLevelType in 4 files (config package)
3. **S3**: Verified Config already has Validate(), Sanitize(), ApplyProfile()

### What Was Forgotten

1. No git commits after each self-contained change
2. Tests need verification before moving to next task
3. S4 Error Details Utility already partially exists

---

## Execution Plan (Sorted by Work vs Impact)

### Phase 1: Cleanup & Verification (Low Work, High Impact)

**P1.1: Commit current changes**

- Files modified: 24 files + new schema utility
- Scope: All changes since last commit

**P1.2: Full test suite verification**

- Command: `go test ./... -count=1`
- Expected: 17/17 packages passing

**P1.3: Remove deprecated RiskLevel alias completely**

- File: `internal/domain/types.go`
- Change: Remove `type RiskLevel = RiskLevelType`
- Impact: Forces canonical type usage everywhere

---

### Phase 2: Error Details Utility Extension (Medium Work, Medium Impact)

**P2.1: Analyze error detail patterns**

- Files: `error_constructors.go`, `detail_helpers.go`
- Goal: Identify remaining duplication

**P2.2: Add builder pattern to ErrorDetails**

- Add fluent builder functions to `detail_helpers.go`
- Example: `WithField()`, `WithValue()`, `WithExpected()`

**P2.3: Update error constructors to use builder**

- Refactor `error_constructors.go` to use new builder pattern

---

### Phase 3: Test Helper Refactoring (Medium Work, Medium Impact)

**P3.1: Analyze BDD test helpers**

- Files: `internal/config/bdd_*.go`, `internal/testing/`
- Goal: Find duplicate patterns

**P3.2: Extract common test utilities**

- Create `internal/testing/test_helpers.go` if needed
- Consolidate duplicate code

---

### Phase 4: SystemCache Research (Low Work, Clarification Needed)

**P4.1: Research CacheType usage**

- Files: `internal/domain/types.go`, `internal/cleaner/`
- Goal: Determine if string→int migration is needed

**P4.2: Decision document**

- Document whether to keep CacheType as int or change it
- Depends on real-world usage patterns

---

### Phase 5: Type Model Improvements (Ongoing)

**Current Strengths:**

- ✓ RiskLevelType: Integer enum with type-safe methods
- ✓ ValidationLevelType: Integer enum with type-safe methods
- ✓ CleanStrategyType: Integer enum with type-safe methods

**Potential Improvements:**

- [ ] CacheType: Verify consistency with other enums
- [ ] DockerPruneModeType: Verify consistency
- [ ] BuildToolType: Verify consistency

---

## Execution Order

1. P1.1: Commit changes (prerequisite for all work)
2. P1.2: Verify tests pass
3. P1.3: Remove RiskLevel alias (completes S2)
4. P2.1: Research error patterns
5. P2.2: Extend ErrorDetails builder
6. P2.3: Update constructors
7. P3.1: Analyze test helpers
8. P3.2: Consolidate duplicates
9. P4.1: Research CacheType
10. P4.2: Document decision

---

## Verification Commands

```bash
# Full test suite
go test ./... -count=1

# Build check
go build ./...

# Lint check (if available)
golangci-lint run ./...

# Coverage check
go test ./... -coverprofile=coverage.out
```

---

## Notes

- All tasks should be small enough to complete in one commit
- Run tests after each change
- Update TODO_LIST.md when tasks are complete
- Document decisions in docs/status/ if significant
