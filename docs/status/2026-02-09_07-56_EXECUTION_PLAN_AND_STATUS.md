# Clean Wizard: Execution Plan & Comprehensive Status Report

**Date**: February 9, 2026, 07:56 UTC  
**Status**: Docker Refactoring COMPLETE - Ready for Next Phase  
**Quality Score**: 90.1/100  
**Branch**: master  
**Last Commit**: cd63d36 (docs(status): add comprehensive status report)

---

## 1. WHAT DID I FORGET? WHAT COULD BE DONE BETTER?

### What I Forgot:

1. **Deprecation Warnings** - 49 warnings for Strategy constants still present
   - Should have fixed these as part of Docker refactoring
   - Technical debt accumulating

2. **Complexity Reduction** - Task 3.3 at 5%, never actually started refactoring
   - Identified functions but no actual work done
   - Should prioritize this before adding new features

3. **Investigation Before Question** - Asked about BuildCache direction without investigating existing usage
   - Should check how domain.CacheType is used elsewhere first
   - Need data before asking architectural questions

4. **Library Research** - Didn't check if samber/do/v2 (DI) is already in go.mod
   - Memory says it should be used, but didn't verify
   - Should check existing dependencies before planning

5. **Test Coverage Analysis** - Didn't check current coverage metrics
   - Claims ~70% average but didn't verify
   - Should baseline before improvements

### What Could Be Done Better:

1. **More Aggressive Cleanup** - Should fix warnings as we go, not accumulate them
2. **Smaller Commits** - Docker refactoring was good, but could be even more granular
3. **Pre-Investigation** - Research existing patterns before asking questions
4. **Metrics-Driven** - Get baseline metrics before claiming improvements

### What Could Still Be Improved:

1. **Fix all 49 deprecation warnings** (1 day work, high impact on code quality)
2. **Extract Cleaner interface** (1 day work, enables polymorphism)
3. **Refactor SystemCache cleaner** (1 day work, fixes enum inconsistency)
4. **Reduce complexity in top 5 functions** (2 days work, better maintainability)
5. **Add integration tests for remaining cleaners** (2 days work, confidence)

---

## 2. COMPREHENSIVE MULTI-STEP EXECUTION PLAN

### Phase 1: Immediate Cleanup (Week 1, Days 1-2)

#### Step 1.1: Fix Deprecation Warnings

**Work**: 0.5 days | **Impact**: HIGH | **Files**: ~15 files

**Details**:

- Replace `domain.StrategyDryRun` with `domain.CleanStrategyDryRun`
- Replace `domain.StrategyConservative` with `domain.CleanStrategyConservative`
- Replace deprecated RiskLevel constants with RiskLevelType
- Run tests after each file to verify

**Acceptance Criteria**:

- [ ] Zero deprecation warnings
- [ ] All tests pass
- [ ] No behavioral changes

#### Step 1.2: Extract Cleaner Interface

**Work**: 1 day | **Impact**: HIGH | **Files**: 12 files

**Details**:

```go
// internal/cleaner/interface.go (NEW)
package cleaner

import (
    "context"
    "github.com/LarsArtmann/clean-wizard/internal/domain"
    "github.com/LarsArtmann/clean-wizard/internal/result"
)

type Cleaner interface {
    Name() string
    Description() string
    Type() domain.OperationType
    IsAvailable(ctx context.Context) bool
    Clean(ctx context.Context) result.Result[domain.CleanResult]
    Scan(ctx context.Context) result.Result[[]domain.ScanItem]
    ValidateSettings(settings *domain.OperationSettings) error
}
```

**Implementation Order**:

1. Create `interface.go` with interface definition
2. Update NixCleaner to implement interface (reference implementation)
3. Update HomebrewCleaner
4. Update DockerCleaner (already close)
5. Update remaining 8 cleaners
6. Update `cmd/clean-wizard/commands/clean.go` to use interface
7. Add tests verifying all cleaners implement interface

**Acceptance Criteria**:

- [ ] All 11 cleaners implement Cleaner interface
- [ ] Can iterate over []Cleaner
- [ ] cmd/clean-wizard uses interface
- [ ] Tests verify interface compliance

#### Step 1.3: Create Cleaner Registry

**Work**: 0.5 days | **Impact**: MEDIUM | **Files**: 2 files

**Details**:

```go
// internal/cleaner/registry.go (NEW)
package cleaner

import "sync"

type Registry struct {
    cleaners map[string]Cleaner
    mu       sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        cleaners: make(map[string]Cleaner),
    }
}

func (r *Registry) Register(name string, c Cleaner) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.cleaners[name] = c
}

func (r *Registry) Get(name string) (Cleaner, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    c, ok := r.cleaners[name]
    return c, ok
}

func (r *Registry) List() []Cleaner {
    r.mu.RLock()
    defer r.mu.RUnlock()
    list := make([]Cleaner, 0, len(r.cleaners))
    for _, c := range r.cleaners {
        list = append(list, c)
    }
    return list
}

func (r *Registry) Available(ctx context.Context) []Cleaner {
    all := r.List()
    available := make([]Cleaner, 0, len(all))
    for _, c := range all {
        if c.IsAvailable(ctx) {
            available = append(available, c)
        }
    }
    return available
}
```

**Acceptance Criteria**:

- [ ] Registry implemented with thread-safety
- [ ] All cleaners registered on startup
- [ ] Can filter by availability
- [ ] Tests cover concurrent access

---

### Phase 2: Enum Consistency (Week 1, Days 3-5)

#### Step 2.1: Research Domain.CacheType Usage

**Work**: 0.5 days | **Impact**: HIGH | **Files**: Investigation only

**Investigation Questions**:

- Where is `domain.CacheType` defined?
- Where is it used outside of SystemCache cleaner?
- What was the original design intent?
- Is there documentation about the abstraction?

**Deliverable**: Analysis document with findings

#### Step 2.2: Refactor SystemCache Cleaner

**Work**: 1 day | **Impact**: HIGH | **Files**: 3 files

**Options Based on Research**:

**Option A**: If CacheType is meant for languages:

- Map languages to system cache types
- Java/Scala → Xcode cache
- Node → CocoaPods cache
- etc.

**Option B**: If CacheType should be cache types:

- Update domain enum to match cleaner
- Keep lowercase strings
- Add missing types

**Option C**: Keep both with explicit mapping:

- Create conversion functions
- Document rationale

**Acceptance Criteria**:

- [ ] SystemCache cleaner uses consistent enum
- [ ] No case sensitivity issues
- [ ] All cache types supported
- [ ] Tests updated and passing

#### Step 2.3: Refactor NodePackages Cleaner

**Work**: 0.5 days | **Impact**: MEDIUM | **Files**: 2 files

**Details**:

- Local enum: string type ("npm", "pnpm", "yarn", "bun")
- Domain enum: integer type
- Solution: Use domain enum directly

**Acceptance Criteria**:

- [ ] NodePackages uses domain.PackageManagerType
- [ ] No type conversions needed
- [ ] Tests updated

#### Step 2.4: Investigate BuildCache Direction

**Work**: 0.5 days | **Impact**: HIGH | **Files**: Investigation

**Top #1 Question to Answer**:
Should BuildCache cleaner use:

1. Domain enum (languages) - map languages to build tools?
2. Local enum (build tools) - update domain to match?
3. Both with mapping layer?

**Investigation**:

- Check original design docs
- Check how domain.BuildToolType is used
- Check user-facing config format
- Determine intended abstraction

**Deliverable**: Decision document with rationale

---

### Phase 3: Code Quality (Week 2)

#### Step 3.1: Reduce Complexity in LoadWithContext

**Work**: 1 day | **Impact**: MEDIUM | **Files**: 2 files

**Current**: 20 cyclomatic complexity
**Target**: <10

**Strategy**:

1. Extract profile loading to separate function
2. Extract operation processing to separate function
3. Extract risk level processing to separate function
4. Use early returns

**Refactor Plan**:

```go
// BEFORE
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
    // ... 100+ lines with nested loops and switches
}

// AFTER
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
    v := viper.New()
    if err := setupViper(v); err != nil {
        return nil, err
    }

    config, err := loadConfigBase(v)
    if err != nil {
        return nil, err
    }

    if err := populateProfiles(v, config); err != nil {
        return nil, err
    }

    return validateAndReturn(config)
}
```

**Acceptance Criteria**:

- [ ] Complexity < 10
- [ ] All tests pass
- [ ] No behavioral changes
- [ ] Better testability

#### Step 3.2: Reduce Complexity in validateProfileName

**Work**: 0.5 days | **Impact**: MEDIUM | **Files**: 1 file

**Current**: 16 cyclomatic complexity
**Strategy**: Extract validation rules to separate functions

#### Step 3.3: Reduce Complexity in Remaining Top 3

**Work**: 1 day | **Impact**: MEDIUM | **Files**: 3 files

- TestIntegration_ValidationSanitizationPipeline (19)
- ErrorCode.String (15)
- EnhancedConfigLoader.SaveConfig (15)

---

### Phase 4: Testing & Documentation (Week 3)

#### Step 4.1: Add Integration Tests for Remaining Cleaners

**Work**: 2 days | **Impact**: HIGH | **Files**: 5 files

**Cleaners needing tests**:

- SystemCache (after refactoring)
- NodePackages (after refactoring)
- BuildCache (after decision)
- LangVersionManager
- TempFiles

**Test Pattern**:

```go
func TestCleaner_EnumWorkflow(t *testing.T) {
    // Load config with enum values
    // Create cleaner
    // Verify enum usage
    // Test clean operation
}
```

#### Step 4.2: Add Edge Case Tests

**Work**: 1 day | **Impact**: MEDIUM | **Files**: 2 files

**Test Cases**:

- Negative integers for enums
- Out-of-range integers
- Mixed case strings
- Empty values
- Null values
- Unicode in strings

#### Step 4.3: Create Architecture Documentation

**Work**: 1 day | **Impact**: LOW | **Files**: 1 file

**Content**:

- Overall architecture diagram
- Layer responsibilities
- Data flow
- Design decisions
- Extension points

---

## 3. SORTED BY WORK REQUIRED VS IMPACT

### Pareto 1% → 51% (High Impact, Low Effort)

| #   | Task                       | Work | Impact | Priority |
| --- | -------------------------- | ---- | ------ | -------- |
| 1   | Fix deprecation warnings   | 0.5d | HIGH   | CRITICAL |
| 2   | Extract Cleaner interface  | 1d   | HIGH   | CRITICAL |
| 3   | Create Cleaner registry    | 0.5d | MEDIUM | HIGH     |
| 4   | Research CacheType usage   | 0.5d | HIGH   | HIGH     |
| 5   | Refactor NodePackages enum | 0.5d | MEDIUM | MEDIUM   |

### Pareto 4% → 64% (High Impact, Medium Effort)

| #   | Task                                  | Work | Impact | Priority |
| --- | ------------------------------------- | ---- | ------ | -------- |
| 6   | Refactor SystemCache cleaner          | 1d   | HIGH   | HIGH     |
| 7   | Reduce LoadWithContext complexity     | 1d   | MEDIUM | MEDIUM   |
| 8   | Add integration tests (5 cleaners)    | 2d   | HIGH   | HIGH     |
| 9   | Investigate BuildCache direction      | 0.5d | HIGH   | HIGH     |
| 10  | Reduce validateProfileName complexity | 0.5d | MEDIUM | MEDIUM   |

### Medium Impact, Medium Effort

| #   | Task                                  | Work | Impact | Priority |
| --- | ------------------------------------- | ---- | ------ | -------- |
| 11  | Reduce remaining complexity (3 funcs) | 1d   | MEDIUM | MEDIUM   |
| 12  | Add edge case tests                   | 1d   | MEDIUM | MEDIUM   |
| 13  | Refactor BuildCache (after decision)  | 1d   | MEDIUM | MEDIUM   |
| 14  | Refactor LangVersionManager           | 0.5d | LOW    | LOW      |
| 15  | Create architecture docs              | 1d   | LOW    | LOW      |

### Lower Priority

| #   | Task                                        | Work | Impact | Priority |
| --- | ------------------------------------------- | ---- | ------ | -------- |
| 16  | Add benchmark regression detection          | 0.5d | LOW    | LOW      |
| 17  | Investigate RiskLevelType manual processing | 0.5d | MEDIUM | LOW      |
| 18  | Add dependency injection (samber/do)        | 2d   | LOW    | LOW      |
| 19  | Create enum quick reference                 | 0.5d | LOW    | LOW      |
| 20  | Add fuzz testing                            | 1d   | LOW    | LOW      |

---

## 4. EXISTING CODE ANALYSIS

### What We Can Reuse:

1. **UnmarshalYAMLEnum** - Already generic, works for all enum types
2. **Testing Framework** - `TestTypeString`, `TestDryRunStrategy` helpers
3. **Result Type** - `result.Result[T]` for error handling
4. **ValidateSettings Pattern** - Consistent across cleaners

### What Libraries Are Available:

**From go.mod check needed:**

- `github.com/samber/do/v2` - Dependency injection (mentioned in memory)
- `github.com/spf13/viper` - Config loading (already used)
- `gopkg.in/yaml.v3` - YAML parsing (already used)
- `github.com/stretchr/testify` - Testing (already used)

### Type Model Improvements:

**Current Issues**:

1. Deprecated constants mixed with new type-safe ones
2. Some enums use integers, some use strings (inconsistency)
3. No unified way to check if enum is valid

**Proposed Improvements**:

```go
// Add to all enum types
func (e EnumType) IsValid() bool {
    return e >= EnumMin && e <= EnumMax
}

// Add String() method consistently
func (e EnumType) String() string {
    // implementation
}

// Add validation at boundaries
func ValidateEnum(value, min, max int, name string) error {
    if value < min || value > max {
        return fmt.Errorf("invalid %s: %d", name, value)
    }
    return nil
}
```

---

## 5. TOP #1 QUESTION

### BuildCache Cleaner: What is the Correct Architectural Direction?

**The Dilemma**:

**Local Enum** (current cleaner implementation):

```go
type BuildToolType string
const (
    BuildToolGradle BuildToolType = "gradle"
    BuildToolMaven  BuildToolType = "maven"
    BuildToolSBT    BuildToolType = "sbt"
)
```

**Domain Enum** (current domain definition):

```go
type BuildToolType int
const (
    BuildToolGO     BuildToolType = iota // "GO"
    BuildToolRUST                        // "RUST"
    BuildToolNODE                        // "NODE"
    BuildToolPYTHON                      // "PYTHON"
    BuildToolJAVA                        // "JAVA"
    BuildToolSCALA                       // "SCALA"
)
```

**The Problem**:

- Domain enum represents **languages** (JAVA, SCALA)
- Cleaner enum represents **build tools** (gradle, maven, sbt)
- These are different abstractions entirely
- Java can use gradle OR maven
- Scala uses sbt

**Options**:

**Option A: Domain Should Represent Languages**

- Refactor cleaner to map languages to their default build tools
- Java → gradle (default), but could also use maven
- Loss of precision but consistent with domain

**Option B: Domain Should Represent Build Tools**

- Update domain enum to: GRADLE, MAVEN, SBT
- More precise but changes domain model
- May affect other parts of system

**Option C: Both Concepts Needed**

- Languages in domain
- Build tools in cleaner
- Explicit mapping between them
- More complex but accurate

**What I Need From You**:

1. **What is the domain concept?** Is `BuildToolType` meant to be:
   - Languages that have build tools?
   - Specific build tool implementations?
   - Something else?

2. **User perspective**: When users configure BuildCache, do they think:
   - "Clean Java caches" (language-centric)
   - "Clean Gradle caches" (tool-centric)

3. **Scope**: Is BuildCache cleaner meant to support:
   - Only Java ecosystem (gradle, maven, sbt)?
   - All languages (Go, Rust, Node, etc.)?

4. **Original intent**: Was there a design decision documented about this?

**My Recommendation**:

Investigate first (Step 2.4), but I'm leaning toward **Option B** (domain should represent build tools) because:

- More precise
- Aligns with cleaner's actual functionality
- Users think in tools ("I use Gradle")
- Languages don't directly map to cache locations

But I need your input on the original design intent before proceeding.

---

## a) FULLY DONE

1. **Docker Cleaner Refactoring** - Migrated from local enum to domain enum ✅
2. **Fixed 26 Compilation Errors** - Across 4 files ✅
3. **Updated All Tests** - Docker tests use domain enum ✅
4. **All Tests Passing** - Full suite passes ✅
5. **Status Report Created** - docs/status/2026-02-09_07-48_DOCKER_REFACTOR_COMPLETE.md ✅
6. **Committed and Pushed** - cd63d36 on master ✅

---

## b) PARTIALLY DONE

1. **Complexity Reduction Task** - 5% complete
   - Identified 21 high-complexity functions
   - No actual refactoring done yet
   - Task deprioritized for Docker fix

---

## c) NOT STARTED

### HIGH PRIORITY:

1. Fix 49 deprecation warnings
2. Extract Cleaner interface
3. Refactor SystemCache cleaner (enum inconsistency)
4. Investigate BuildCache architectural direction

### MEDIUM PRIORITY:

5. Refactor NodePackages cleaner
6. Reduce complexity in top 5 functions
7. Add integration tests for remaining cleaners

### LOW PRIORITY:

8. Refactor BuildCache cleaner (after decision)
9. Refactor LangVersionManager cleaner
10. Add dependency injection

---

## d) TOTALLY FUCKED UP

**NONE** ✅

All tests passing, no critical issues.

---

## e) WHAT WE SHOULD IMPROVE

1. **Fix Deprecation Warnings** - 49 warnings accumulating
2. **Extract Cleaner Interface** - Enable polymorphism
3. **Enum Consistency** - 4 cleaners still have mismatches
4. **Complexity Reduction** - 21 functions need refactoring
5. **Test Coverage** - Add integration tests for all cleaners
6. **Documentation** - Architecture docs missing
7. **Type Safety** - Add IsValid() to all enum types

---

## f) TOP 25 THINGS TO GET DONE NEXT

### CRITICAL (Do First):

1. Fix 49 deprecation warnings (0.5d, HIGH impact)
2. Extract Cleaner interface (1d, HIGH impact)
3. Research domain.CacheType usage (0.5d, HIGH impact)

### HIGH PRIORITY (Do Next):

4. Refactor SystemCache cleaner (1d, HIGH impact)
5. Add integration tests for 5 cleaners (2d, HIGH impact)
6. Investigate BuildCache direction (0.5d, HIGH impact)
7. Create Cleaner registry (0.5d, MEDIUM impact)
8. Reduce LoadWithContext complexity (1d, MEDIUM impact)

### MEDIUM PRIORITY (Do After):

9. Refactor NodePackages cleaner (0.5d, MEDIUM impact)
10. Reduce validateProfileName complexity (0.5d, MEDIUM impact)
11. Reduce remaining 3 high-complexity functions (1d, MEDIUM impact)
12. Add edge case tests (1d, MEDIUM impact)
13. Refactor BuildCache (after decision) (1d, MEDIUM impact)

### LOWER PRIORITY:

14. Refactor LangVersionManager (0.5d, LOW impact)
15. Create architecture documentation (1d, LOW impact)
16. Add benchmark regression detection (0.5d, LOW impact)
17. Investigate RiskLevelType manual processing (0.5d, MEDIUM impact)
18. Add dependency injection (2d, LOW impact)
19. Create enum quick reference (0.5d, LOW impact)
20. Add fuzz testing (1d, LOW impact)
21. Reduce complexity in remaining 16 functions (3d, LOW impact)
22. Add plugin architecture (5d, LOW impact)
23. Add structured logging (1d, LOW impact)
24. Create migration guides (1d, LOW impact)
25. Add performance monitoring (1d, LOW impact)

---

## g) TOP #1 QUESTION

**BuildCache Cleaner: What is the Correct Architectural Direction?**

**Context**:

- Local enum: Build tools ("gradle", "maven", "sbt")
- Domain enum: Languages ("JAVA", "SCALA", "GO", "RUST", etc.)
- These represent different abstractions

**Question**: Should we:

1. **Refactor cleaner** to use languages (map JAVA → gradle/maven)?
2. **Update domain** to use build tools (GRADLE, MAVEN, SBT)?
3. **Keep both** with explicit mapping layer?

**What I need from you**:

- Original design intent for domain.BuildToolType
- User perspective: do they think in languages or tools?
- Scope: should BuildCache support all languages or just Java ecosystem?

**My lean**: Option 2 (domain should represent build tools) for precision, but need your confirmation.

---

**Report Generated**: 2026-02-09 07:56 UTC  
**Next Action**: Waiting for your instructions on which task to tackle first
