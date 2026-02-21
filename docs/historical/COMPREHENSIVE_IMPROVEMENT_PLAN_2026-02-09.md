# Clean Wizard: Comprehensive Improvement Analysis & Execution Plan

**Date**: February 9, 2026
**Status**: Analysis Complete - Ready for Execution
**Priority**: CRITICAL - Disk Space Blocking Development

---

## Executive Summary

Clean Wizard is a **world-class Go application** (90.1/100 quality score) with excellent architecture, type safety, and production-ready code. However, immediate critical issues and strategic improvements must be addressed:

**CRITICAL BLOCKER**: Disk is 100% full - must be resolved before any work
**STRATEGIC GAP**: 21 of 25 enum tasks incomplete (84% remaining)
**ARCHITECTURAL OPPORTUNITIES**: 12 high-impact improvements identified

---

## Part 1: What Did We Forget? What Could Be Done Better?

### 1.1 Immediate Oversight Analysis

**What Was Forgotten:**

1. **Disk Space Management** ❌ CRITICAL
   - No automated cleanup of build artifacts
   - No disk space monitoring before running tests
   - No cache size limits in CI/CD
   - **Impact**: Currently blocking all development

2. **Cleaner Interface Extraction** ❌ HIGH IMPACT
   - All 11+ cleaners have identical `Clean()` signature
   - No shared interface for polymorphism
   - Duplicate patterns across all cleaners
   - **Impact**: Can't iterate over cleaners, no mock interface

3. **Error Context Preservation** ❌ MEDIUM IMPACT
   - `validate.go:12` loses context in error messages
   - Variables `items` and `validItems` not included in error
   - **Impact**: Difficult debugging, lower error quality score

4. **Enum Unification Opportunity** ❌ MEDIUM IMPACT
   - Binary enums use `unmarshalBinaryEnum()`
   - Standard enums use `UnmarshalYAMLEnum()`
   - Two separate code paths for similar functionality
   - **Impact**: Duplication, maintenance overhead

5. **Integration Testing Gap** ❌ HIGH IMPACT
   - No end-to-end tests for enum-based configs
   - Can't verify full workflow works with int/string enums
   - **Impact**: Risk of breaking changes, low confidence

6. **Library Underutilization** ❌ MEDIUM IMPACT
   - No DI container (manual wiring everywhere)
   - `samber/do` mentioned in memory but not used
   - Rate limiter exists but not consistently applied
   - **Impact**: Boilerplate, inconsistency

**What Could Be Done Better:**

1. **Task Granularity** - Original 25-task list too coarse
   - Task #7: "Verify all cleaners" should be 11 separate tasks
   - Task #21: High complexity reduction should be 21 separate tasks
   - **Fix**: Break down into sub-tasks for better tracking

2. **Documentation Timing** - Docs written before implementation
   - Tasks 1-3 were all documentation-heavy
   - Should prioritize working code first
   - **Fix**: Implement features, verify, then document

3. **Batch Processing** - Tasks done sequentially instead of batching
   - Similar tasks could be done together
   - Enum tasks could be batched by type
   - **Fix**: Group related tasks for efficiency

4. **Test Automation** - Some verification still manual
   - Need automated regression testing
   - Performance benchmarks exist but no regression detection
   - **Fix**: Add automated checks to CI

5. **Migration Path** - RiskLevelType manual processing not addressed
   - Open question from status report remains unanswered
   - Should either fix or document rationale
   - **Fix**: Investigate Viper enum support, decide on approach

### 1.2 Existing Code Analysis

**What Works Well:**

✅ **Type-Safe Enums** - 15+ enum types fully implemented

- RiskLevelType, ValidationLevelType, ChangeOperationType, CleanStrategyType
- ExecutionMode, SafeMode, ProfileStatus, GenerationStatus, ScanMode
- OptimizationMode, HomebrewMode
- CacheCleanupMode, DockerPruneMode, BuildToolType
- CacheType, VersionManagerType, PackageManagerType

✅ **Enum YAML/JSON Support** - Dual format handling

- String format (human-readable): "DISABLED", "ENABLED"
- Integer format (machine-readable): 0, 1
- Case-insensitive matching
- Helpful error messages with valid options

✅ **JSON Schema** - Comprehensive validation

- Validates all 12 enum types with integer values
- Pattern validation for duration fields
- Nested validation for operation settings

✅ **Benchmark Tests** - Performance baseline established

- Enum marshal/unmarshal: sub-nanosecond
- Full config round-trip: <10ns
- Zero allocations for enum operations

✅ **Documentation** - Well-structured

- YAML_ENUM_FORMATS.md (400+ lines)
- Config schema documentation
- Usage examples and migration guides

**What Still Needs Work:**

❌ **Cleaner Interface Missing** - No polymorphism

```go
// Current: Each cleaner has its own Clean() method
func (c *NixCleaner) Clean(ctx context.Context) Result[domain.CleanResult]
func (c *HomebrewCleaner) Clean(ctx context.Context) Result[domain.CleanResult]
// ... repeated 11 times

// Needed: Shared interface
type Cleaner interface {
    Name() string
    IsAvailable(ctx context.Context) bool
    Clean(ctx context.Context) Result[domain.CleanResult]
    Estimate(ctx context.Context) Result[domain.SizeEstimate]
}
```

❌ **High Complexity Functions** - 21 functions >10 cyclomatic complexity

- `config.LoadWithContext`: 20 complexity (nested loops, switch statements)
- `config.TestIntegration_ValidationSanitizationPipeline`: 19 complexity
- `config.(*ConfigValidator).validateProfileName`: 16 complexity
- And 18 more...

❌ **Manual RiskLevelType Processing** - Inconsistent with other enums

```go
// internal/config/config.go:86-108
var riskLevelStr string
v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", name, i), &riskLevelStr)
switch strings.ToUpper(riskLevelStr) {
case "LOW": op.RiskLevel = domain.RiskLow
// ... manual processing instead of using enum unmarshaler
```

❌ **Context Loss in validate.go** - Error messages lack context

```go
// internal/cleaner/validate.go:12
func ValidateItem(items, validItems []string, item, itemName, validValuesDescription string) error {
    // ... validation code
    return fmt.Errorf("invalid %s: %s (must be %s)", itemName, item, validValuesDescription)
    // Context variables 'items' and 'validItems' lost
}
```

---

## Part 2: What Should We Still Improve?

### 2.1 Architecture Improvements (High Impact)

#### Priority 1: Extract Generic Cleaner Interface

**Impact**: HIGH | **Effort**: LOW (1 day) | **ROI**: 9/10

**Current State**:

- 11+ cleaner implementations with identical patterns
- No shared interface
- Can't iterate over cleaners programmatically
- Hard to mock for testing

**Improvement**:

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
    IsAvailable(ctx context.Context) bool
    Clean(ctx context.Context) result.Result[domain.CleanResult]
    Estimate(ctx context.Context) result.Result[domain.SizeEstimate]
}

type CleanerRegistry struct {
    cleaners map[string]Cleaner
}

func (r *CleanerRegistry) Register(name string, cleaner Cleaner) {
    r.cleaners[name] = cleaner
}

func (r *CleanerRegistry) Get(name string) (Cleaner, bool) {
    cleaner, ok := r.cleaners[name]
    return cleaner, ok
}

func (r *CleanerRegistry) List() []Cleaner {
    cleaners := make([]Cleaner, 0, len(r.cleaners))
    for _, cleaner := range r.cleaners {
        cleaners = append(cleaners, cleaner)
    }
    return cleaners
}
```

**Files to Modify**:

- `internal/cleaner/interface.go` (NEW)
- `internal/cleaner/nix.go` (implement interface)
- `internal/cleaner/homebrew.go` (implement interface)
- `internal/cleaner/docker.go` (implement interface)
- `internal/cleaner/cargo.go` (implement interface)
- `internal/cleaner/golang_cleaner.go` (implement interface)
- `internal/cleaner/nodepackages.go` (implement interface)
- `internal/cleaner/buildcache.go` (implement interface)
- `internal/cleaner/systemcache.go` (implement interface)
- `internal/cleaner/tempfiles.go` (implement interface)
- `internal/cleaner/projectsmanagementautomation.go` (implement interface)
- `cmd/clean-wizard/commands/clean.go` (use interface)

**Benefits**:
✅ Can iterate over []Cleaner
✅ Single type for all cleaners
✅ Easier testing (mock interface)
✅ Enables generic operations
✅ Reduces code duplication

---

#### Priority 2: Fix Context Propagation in validate.go

**Impact**: MEDIUM | **Effort**: LOW (0.5 days) | **ROI**: 7/10

**Current State**:

- Error messages lose context
- Debugging difficult
- Lower error handling quality score

**Improvement**:

```go
// internal/cleaner/validate.go
package cleaner

import (
    "fmt"
    "strings"
)

func ValidateItem(items, validItems []string, item, itemName, validValuesDescription string) error {
    if err := ValidateItems(items, item); err != nil {
        return fmt.Errorf("%s validation failed: %w (items: %v, valid: %v)",
            itemName, err, items, validItems)
    }
    if err := ValidateItems(validItems, item); err != nil {
        return fmt.Errorf("valid items check failed: %w (valid: %v)",
            err, validItems)
    }
    return fmt.Errorf("invalid %s: %s (must be one of %s)", itemName, item,
        formatValidItems(validItems))
}

func formatValidItems(items []string) string {
    if len(items) == 0 {
        return "none"
    }
    if len(items) <= 3 {
        return strings.Join(items, ", ")
    }
    return fmt.Sprintf("%s, ... (%d total)", strings.Join(items[:3], ", "), len(items))
}
```

**Benefits**:
✅ Context preserved in error messages
✅ Better debugging experience
✅ Improves error handling quality score from 90.1/100 to 95+

---

#### Priority 3: Unify Binary Enum Unmarshaling

**Impact**: MEDIUM | **Effort**: LOW (1 day) | **ROI**: 6/10

**Current State**:

- Binary enums use `unmarshalBinaryEnum()`
- Standard enums use `UnmarshalYAMLEnum()`
- Two separate code paths

**Improvement**:

```go
// Option 1: Make UnmarshalYAMLEnum work for binary enums
// internal/domain/type_safe_enums.go

func UnmarshalYAMLEnumBinary[T ~int](
    value *yaml.Node,
    target *T,
    disabled T,
    enabled T,
    name string,
) error {
    return UnmarshalYAMLEnum(value, target, map[string]T{
        "DISABLED": disabled,
        "ENABLED":  enabled,
        "0":        disabled,
        "1":        enabled,
        "FALSE":    disabled,
        "TRUE":     enabled,
    }, "invalid "+name)
}

// Option 2: Keep both but share error formatting
// Both functions call a shared error formatter
```

**Benefits**:
✅ Reduces code duplication
✅ Consistent error messages
✅ Easier maintenance

---

### 2.2 Testing Improvements (High Impact)

#### Priority 4: Add Integration Tests for Enum-Based Configs

**Impact**: HIGH | **Effort**: MEDIUM (2 days) | **ROI**: 8/10

**Current State**:

- Only unit tests for enum unmarshaling
- No end-to-end workflow tests
- Can't verify full config→execute→result cycle

**Improvement**:

```go
// tests/integration/enum_workflow_test.go (NEW)
package integration_test

import (
    "context"
    "testing"
    "github.com/LarsArtmann/clean-wizard/internal/config"
    "github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestEnumWorkflow_IntegerFormat(t *testing.T) {
    // Load config with integer enums
    cfg, err := config.LoadWithContext(context.Background(), "testdata/int-enum-config.yaml")
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    // Verify enum values are correct
    if cfg.Profiles["default"].Operations[0].Settings.GoPackages.CleanCache != domain.CacheCleanupEnabled {
        t.Errorf("Expected CacheCleanupEnabled, got %v",
            cfg.Profiles["default"].Operations[0].Settings.GoPackages.CleanCache)
    }

    // Execute cleanup
    // Verify results
}

func TestEnumWorkflow_StringFormat(t *testing.T) {
    // Load config with string enums
    cfg, err := config.LoadWithContext(context.Background(), "testdata/string-enum-config.yaml")
    // ... same verification
}

func TestEnumWorkflow_MixedFormat(t *testing.T) {
    // Load config with mixed enums
    cfg, err := config.LoadWithContext(context.Background(), "testdata/mixed-enum-config.yaml")
    // ... same verification
}

func TestEnumRoundTrip_YAML(t *testing.T) {
    // Load config → Marshal YAML → Unmarshal YAML
    // Verify no data loss
}

func TestEnumRoundTrip_JSON(t *testing.T) {
    // Load config → Marshal JSON → Unmarshal JSON
    // Verify no data loss
}
```

**Files to Create**:

- `tests/integration/enum_workflow_test.go` (NEW)
- `tests/integration/testdata/int-enum-config.yaml` (NEW)
- `tests/integration/testdata/string-enum-config.yaml` (NEW)
- `tests/integration/testdata/mixed-enum-config.yaml` (NEW)

**Benefits**:
✅ Verifies end-to-end workflows
✅ Confirms no breaking changes
✅ Tests actual usage patterns
✅ High confidence in enum system

---

#### Priority 5: Verify All Cleaners Handle Enum Types Correctly

**Impact**: HIGH | **Effort**: MEDIUM (2 days) | **ROI**: 8/10

**Current State**:

- Enum types exist but not verified in practice
- Cleaners might have hardcoded string comparisons
- Risk of enum values not being used correctly

**Improvement**:
For each cleaner (11 total):

1. Review all enum usage
2. Verify no hardcoded string comparisons
3. Ensure switch statements use enum constants
4. Add unit tests for enum handling
5. Add integration test with enum-based config

**Files to Review**:

- `internal/cleaner/nix.go`
- `internal/cleaner/homebrew.go`
- `internal/cleaner/docker.go`
- `internal/cleaner/cargo.go`
- `internal/cleaner/golang_cleaner.go`
- `internal/cleaner/nodepackages.go`
- `internal/cleaner/buildcache.go`
- `internal/cleaner/systemcache.go`
- `internal/cleaner/tempfiles.go`
- `internal/cleaner/projectsmanagementautomation.go`

**Benefits**:
✅ Ensures enums are actually used
✅ Catches hardcoded strings
✅ Verifies correct enum usage patterns
✅ Prevents type-safety regressions

---

### 2.3 Code Quality Improvements (Medium Impact)

#### Priority 6: Reduce Cyclomatic Complexity

**Impact**: MEDIUM | **Effort**: MEDIUM (3 days) | **ROI**: 6/10

**Current State**:

- 21 functions exceed 10 cyclomatic complexity
- Top offenders: LoadWithContext (20), validateProfileName (16)

**Improvement Strategy**:
For each high-complexity function:

1. Extract helper functions for repeated patterns
2. Use strategy pattern for complex conditional logic
3. Split large functions into focused sub-functions (<30 lines each)

**Example Refactor** (LoadWithContext):

```go
// BEFORE: LoadWithContext (20 complexity)
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
    v := viper.New()
    // ... 100+ lines with 20 cyclomatic complexity
    for name, profile := range config.Profiles {
        for i := range profile.Operations {
            // ... nested logic
            switch strings.ToUpper(riskLevelStr) {
                // ... multiple cases
            }
            // ... more nested logic
        }
    }
    // ... more complexity
}

// AFTER: Reduced complexity by extraction
func LoadWithContext(ctx context.Context) (*domain.Config, error) {
    v := viper.New()
    // ... basic setup
    config, err := loadConfigBase(v)
    if err != nil {
        return nil, err
    }
    config, err = populateProfiles(v, config)
    if err != nil {
        return nil, err
    }
    return validateAndReturn(config)
}

func loadConfigBase(v *viper.Viper) (*domain.Config, error) { /* extracted */ }
func populateProfiles(v *viper.Viper, config *domain.Config) (*domain.Config, error) { /* extracted */ }
func validateAndReturn(config *domain.Config) (*domain.Config, error) { /* extracted */ }
```

**Functions to Refactor**:

1. `config.LoadWithContext` (20) → Extract profile loading
2. `config.TestIntegration_ValidationSanitizationPipeline` (19) → Extract test scenarios
3. `config.(*ConfigValidator).validateProfileName` (16) → Extract name validation rules
4. `errors.(ErrorCode).String` (15) → Extract error code formatting
5. `config.(*EnhancedConfigLoader).SaveConfig` (15) → Extract save operations
6. `cleaner.(*HomebrewCleaner).Clean` (15) → Extract cleanup steps
7. `config.(*BDDTestRunner).runScenario` (14) → Extract scenario execution
8. `errors.(*CleanWizardError).Error` (13) → Extract error message building
9. `domain.(*OperationSettings).ValidateSettings` (13) → Extract validation rules
10. `domain.DefaultSettings` (13) → Extract default generation
11. And 10 more...

**Benefits**:
✅ More readable functions
✅ Easier to test individual parts
✅ Lower cognitive load
✅ Better separation of concerns

---

### 2.4 Documentation Improvements (Low Impact)

#### Priority 7: Create Comprehensive Architecture Documentation

**Impact**: LOW | **Effort**: MEDIUM (2 days) | **ROI**: 5/10

**Current State**:

- Scattered documentation across multiple files
- No central architecture guide
- Design decisions not documented

**Improvement**:
Create `ARCHITECTURE.md` covering:

- Overall architecture diagram
- Layer responsibilities
- Data flow between layers
- Design decisions and rationale
- Type safety philosophy
- Testing strategy
- Extension points

**Benefits**:
✅ Onboarding resource for new developers
✅ Design rationale preserved
✅ Easier maintenance
✅ Consistency in future development

---

## Part 3: Well-Established Libraries to Leverage

### 3.1 Dependency Injection: samber/do/v2

**Current State**:

- Manual dependency wiring everywhere
- No DI container
- Boilerplate code

**Improvement**:

```go
import "github.com/samber/do/v2"

type Service struct {
    cleanerRegistry do.Injector[CleanerRegistry]
    configLoader    do.Injector[ConfigLoader]
}

func NewService(injector *do.Injector) *Service {
    return &Service{
        cleanerRegistry: do.Inject[CleanerRegistry](injector),
        configLoader:    do.Inject[ConfigLoader](injector),
    }
}
```

**Benefits**:
✅ Automatic dependency resolution
✅ Easier testing (mock injection)
✅ Less boilerplate
✅ Explicit dependencies

---

### 3.2 Validation: github.com/go-playground/validator/v10

**Current State**:

- Custom validation logic
- Manual error formatting
- Scattered validation rules

**Improvement**:

```go
type OperationSettings struct {
    NixGenerations *NixGenerationsSettings `json:"nix_generations,omitempty" validate:"omitempty,dive"`
}

func (os *OperationSettings) Validate() error {
    validate := validator.New()
    return validate.Struct(os)
}
```

**Benefits**:
✅ Standard validation library
✅ Built-in error formatting
✅ Struct tags for declarative validation
✅ Less custom code

---

### 3.3 Logging: github.com/sirupsen/logrus (ALREADY USED)

**Current State**:

- logrus is already a dependency
- Used in some places
- Not consistent across codebase

**Improvement**:

- Ensure consistent logging patterns
- Add structured logging
- Add log levels appropriately

**Benefits**:
✅ Consistent logging
✅ Better observability
✅ Easier debugging
✅ Production-ready

---

### 3.4 HTTP Client: github.com/go-resty/resty/v2 (ALREADY USED)

**Current State**:

- resty is already a dependency
- Used in some places
- Could be more widely adopted

**Benefits**:
✅ Already in dependency tree
✅ Retry logic built-in
✅ Timeout handling
✅ Request/response logging

---

## Part 4: Execution Plan (Sorted by Work vs Impact)

### Phase 1: Critical Blocking Issues (Week 1)

#### Task 1.1: Clean Disk Space ✅ WORKAROUND IDENTIFIED

- **Priority**: CRITICAL
- **Impact**: UNBLOCKS ALL WORK
- **Effort**: 0.5 days
- **Status**: Disk is 100% full, Nix store has read-only Go modules
- **Workaround**: Use `nix-collect-garbage -d` to clean Nix store, or rebuild project in clean environment
- **Actions**:
  - Nix Go modules in `/Users/larsartmann/go/pkg/mod` are read-only (cannot be deleted)
  - Clean Nix store: `nix-collect-garbage -d`
  - OR: Use temporary directory for development
  - OR: Ask user to free disk space and try `go mod download` to refresh modules
  - Verify disk space available
- **NOTE**: This is blocking test execution, but code changes can still be made

#### Task 1.2: Run Full Test Suite

- **Priority**: CRITICAL
- **Impact**: ESTABLISHES BASELINE
- **Effort**: 0.5 days
- **Status**: Blocked by disk space (Nix Go modules are read-only)
- **Workaround**: Proceed with code changes that don't require test execution, verify later
- **Actions**:
  - Run `just test` to verify all tests pass (when disk space available)
  - Document any test failures
  - Fix immediate issues
  - Capture coverage report

---

### Phase 2: High Impact, Low Effort (Week 1-2) - Pareto 1% → 51%

#### Task 2.1: Extract Generic Cleaner Interface

- **Priority**: HIGH
- **Impact**: Enables polymorphism, reduces duplication
- **Effort**: 1 day
- **Files**: 12 files (1 new, 11 modified)
- **Actions**:
  1. Create `internal/cleaner/interface.go` with Cleaner interface
  2. Create `CleanerRegistry` type
  3. Update all 11 cleaners to implement interface
  4. Update `cmd/clean-wizard/commands/clean.go` to use interface
  5. Add tests for interface implementation
  6. Verify all cleaners work correctly

#### Task 2.2: Fix Context Propagation in validate.go

- **Priority**: HIGH
- **Impact**: Better error messages, easier debugging
- **Effort**: 0.5 days
- **Files**: 1 file modified
- **Actions**:
  1. Update `ValidateItem()` to preserve context in errors
  2. Add `formatValidItems()` helper
  3. Update error messages to include items/validItems
  4. Add tests for error context preservation
  5. Verify error quality score improvement

#### Task 2.3: Unify Binary Enum Unmarshaling

- **Priority**: MEDIUM
- **Impact**: Reduces duplication, consistent errors
- **Effort**: 1 day
- **Files**: 2 files modified
- **Actions**:
  1. Create `UnmarshalYAMLEnumBinary()` helper
  2. Update all binary enums to use unified function
  3. Remove `unmarshalBinaryEnum()` if possible
  4. Update tests to verify behavior unchanged
  5. Verify error messages consistent

#### Task 2.4: Add Integration Tests for Enum Workflows

- **Priority**: HIGH
- **Impact**: Verifies end-to-end functionality
- **Effort**: 2 days
- **Files**: 4 files (1 new test file, 3 test configs)
- **Actions**:
  1. Create `tests/integration/enum_workflow_test.go`
  2. Create test configs: int-enum, string-enum, mixed-enum
  3. Implement TestEnumWorkflow_IntegerFormat
  4. Implement TestEnumWorkflow_StringFormat
  5. Implement TestEnumWorkflow_MixedFormat
  6. Implement TestEnumRoundTrip_YAML
  7. Implement TestEnumRoundTrip_JSON
  8. Run integration tests with `-tags=integration`
  9. Fix any failures
  10. Add to CI/CD pipeline

---

### Phase 3: High Impact, Medium Effort (Week 2-3) - Pareto 4% → 64%

#### Task 3.1: Verify All Cleaners Handle Enums Correctly

- **Priority**: HIGH
- **Impact**: Ensures type safety in practice
- **Effort**: 2 days
- **Files**: 11 cleaner files reviewed
- **Actions**:
  For each cleaner (11 total):
  1. Review all enum usage
  2. Check for hardcoded string comparisons
  3. Verify switch statements use enum constants
  4. Add unit test for enum handling if missing
  5. Add integration test with enum-based config
  6. Document findings

#### Task 3.2: Add Enum Validation to Config Boundaries

- **Priority**: HIGH
- **Impact**: Catches invalid enums at config load time
- **Effort**: 1 day
- **Files**: 2 files modified
- **Actions**:
  1. Add enum validation to `config.LoadWithContext()`
  2. Validate all enum values after loading
  3. Add helpful error messages for invalid enums
  4. Add tests for validation failure cases
  5. Verify validation catches invalid configs

#### Task 3.3: Reduce Complexity in Top 5 Functions

- **Priority**: MEDIUM
- **Impact**: Better maintainability
- **Effort**: 1 day
- **Files**: 5 files modified
- **Actions**:
  1. Refactor `config.LoadWithContext` (20 complexity)
  2. Refactor `config.TestIntegration_ValidationSanitizationPipeline` (19 complexity)
  3. Refactor `config.(*ConfigValidator).validateProfileName` (16 complexity)
  4. Refactor `errors.(ErrorCode).String` (15 complexity)
  5. Refactor `config.(*EnhancedConfigLoader).SaveConfig` (15 complexity)
  6. Extract helper functions
  7. Verify all tests still pass
  8. Measure complexity reduction

---

### Phase 4: Medium Impact, Medium Effort (Week 3-4)

#### Task 4.1: Reduce Complexity in Remaining Functions

- **Priority**: MEDIUM
- **Impact**: Better maintainability
- **Effort**: 2 days
- **Files**: 16 files modified
- **Actions**:
  1. Refactor remaining 16 high-complexity functions
  2. Apply same extraction strategy as Task 3.3
  3. Verify all tests pass
  4. Update complexity metrics

#### Task 4.2: Add Edge Case Tests for Enum Unmarshaling

- **Priority**: MEDIUM
- **Impact**: Better error handling
- **Effort**: 1 day
- **Files**: 1 test file modified
- **Actions**:
  1. Add test for negative integers
  2. Add test for out-of-range integers
  3. Add test for mixed case strings
  4. Add test for empty values
  5. Add test for null values
  6. Verify all tests pass

#### Task 4.3: Test Enum Round-Trip Serialization

- **Priority**: MEDIUM
- **Impact**: Ensures no data corruption
- **Effort**: 1 day
- **Files**: 1 test file modified
- **Actions**:
  1. Test YAML → Go → YAML for all enum types
  2. Test JSON → Go → JSON for all enum types
  3. Verify no data loss
  4. Verify consistent formatting
  5. Add to CI/CD pipeline

---

### Phase 5: Low Impact, High Effort (Week 5-6)

#### Task 5.1: Create Comprehensive Architecture Documentation

- **Priority**: LOW
- **Impact**: Better onboarding
- **Effort**: 2 days
- **Files**: 1 new file
- **Actions**:
  1. Create `ARCHITECTURE.md`
  2. Document overall architecture
  3. Document layer responsibilities
  4. Document data flow
  5. Document design decisions
  6. Document type safety philosophy
  7. Document testing strategy
  8. Document extension points

#### Task 5.2: Investigate and Fix RiskLevelType Manual Processing

- **Priority**: MEDIUM
- **Impact**: Consistency with other enums
- **Effort**: 1 day
- **Files**: 2 files modified
- **Actions**:
  1. Investigate Viper enum support
  2. Test using RiskLevelType directly with Viper
  3. Fix manual processing if possible
  4. Document rationale if manual processing required
  5. Add tests for both approaches

#### Task 5.3: Add Dependency Injection with samber/do/v2

- **Priority**: LOW
- **Impact**: Less boilerplate, easier testing
- **Effort**: 2 days
- **Files**: Multiple files modified
- **Actions**:
  1. Add samber/do/v2 to go.mod
  2. Create DI container in cmd/clean-wizard
  3. Update main.go to use DI
  4. Update all services to use DI
  5. Add tests with mock injection
  6. Verify all tests pass

---

### Phase 6: Final Validation (Week 7)

#### Task 6.1: Verify Full Integration and Run All Tests

- **Priority**: CRITICAL
- **Impact**: Ensures everything works
- **Effort**: 1 day
- **Actions**:
  1. Run full test suite: `just test`
  2. Run integration tests: `go test ./tests/integration -tags=integration`
  3. Run BDD tests: `go test ./tests/bdd`
  4. Run benchmarks: `go test -bench=. ./tests/benchmark`
  5. Check coverage: `go test -cover ./...`
  6. Fix any failures
  7. Document final status

#### Task 6.2: Create Quick Reference Guide for Enum Types

- **Priority**: LOW
- **Impact**: Easier lookup for developers
- **Effort**: 0.5 days
- **Files**: 1 new file
- **Actions**:
  1. Create `docs/ENUM_QUICK_REFERENCE.md`
  2. List all 15+ enum types
  3. Show all values for each enum
  4. Show usage examples
  5. Add links to full documentation

---

## Part 5: What We Learned

### 5.1 What Went Well

1. **Type-Safe Enum System** - Comprehensive implementation
   - 15+ enum types fully implemented
   - Dual format support (int/string)
   - Excellent error messages
   - High performance (sub-nanosecond)

2. **Architecture Foundation** - Clean architecture in place
   - Proper layering
   - Zero circular dependencies
   - Railway programming pattern
   - Good separation of concerns

3. **Documentation** - Well-structured
   - YAML_ENUM_FORMATS.md comprehensive
   - JSON schema for validation
   - Clear usage examples

4. **Testing Infrastructure** - Good foundation
   - Benchmark tests in place
   - Test structure organized
   - BDD framework available

### 5.2 What Could Be Improved

1. **Disk Space Management** - Critical oversight
   - Should have automated cleanup
   - Should monitor disk space before running tests
   - Should have cache size limits

2. **Task Granularity** - Too coarse initially
   - Should break down large tasks into sub-tasks
   - Better progress tracking

3. **Documentation Timing** - Too early
   - Should implement features first
   - Document after verification

4. **Batch Processing** - Sequential instead of parallel
   - Could group similar tasks
   - More efficient execution

5. **Library Utilization** - Underutilized
   - samber/do available but not used
   - Rate limiter exists but not consistently applied

### 5.3 What We Would Do Differently

1. **Start with Integration Tests** - Verify end-to-end first
2. **Use Dependency Injection** - From the beginning
3. **Extract Interfaces Early** - Cleaner interface should have been first
4. **Monitor Disk Space** - Prevent blocking issues
5. **Prioritize by Impact** - Pareto principle from day one

---

## Part 6: Open Questions

### 6.1 RiskLevelType Manual Processing

**Question**: Should we fix the manual RiskLevelType processing in `internal/config/config.go:86-108` or keep it?

**Context**:

- All other enums use type-safe `UnmarshalYAML()` methods
- RiskLevelType is defined as an enum in `internal/domain/type_safe_enums.go`
- But config loader manually processes it as a string
- This creates inconsistency and potential for errors

**Options**:

1. **Fix it** - Use standard enum unmarshaler
   - Pros: Consistent, type-safe, less code
   - Cons: Need to test Viper compatibility, might break existing configs

2. **Keep it** - Document as necessary workaround
   - Pros: Works now, no risk of breaking
   - Cons: Inconsistent, more code to maintain

3. **Investigate first** - Test Viper enum support
   - Pros: Data-driven decision
   - Cons: Takes time, may be inconclusive

**Recommendation**: Investigate Viper enum support first, then decide based on findings.

---

### 6.2 Dependency Injection Adoption

**Question**: Should we adopt samber/do/v2 for dependency injection?

**Context**:

- Manual dependency wiring everywhere
- Boilerplate code
- Harder to test

**Options**:

1. **Adopt fully** - Use DI container everywhere
   - Pros: Automatic resolution, less boilerplate
   - Cons: Learning curve, adds dependency

2. **Adopt selectively** - Use DI for complex services only
   - Pros: Best of both worlds
   - Cons: Inconsistent patterns

3. **Skip for now** - Keep manual wiring
   - Pros: No new dependency, existing code works
   - Cons: More boilerplate, harder to test

**Recommendation**: Adopt selectively for complex services, evaluate after first phase.

---

### 6.3 Plugin Architecture

**Question**: Should we implement a plugin architecture for cleaners?

**Context**:

- Currently hardcoded list of cleaners
- Adding new cleaners requires modifying core code
- Plugin system mentioned in ARCHITECTURAL_ANALYSIS_2026-02-08_05-48.md

**Options**:

1. **Implement now** - Add plugin system
   - Pros: Extensible, community contributions
   - Cons: Complexity, 5+ days work

2. **Implement later** - After core features stable
   - Pros: Focus on type safety first
   - Cons: Harder to add later

3. **Defer decision** - Wait for user demand
   - Pros: No premature optimization
   - Cons: Might be blocking feature

**Recommendation**: Defer decision, focus on core type safety first.

---

## Part 7: Success Metrics

### 7.1 Technical Metrics

- [ ] All tests passing (current: blocked by disk space)
- [ ] Test coverage > 85% (current: varies by package, avg ~70%)
- [ ] Cyclomatic complexity < 10 for all functions (current: 21 functions >10)
- [ ] Error handling quality score > 95 (current: 90.1)
- [ ] Zero lint warnings in production code
- [ ] Zero security vulnerabilities
- [ ] All critical paths covered by integration tests

### 7.2 Architecture Metrics

- [ ] Cleaner interface implemented and used by all cleaners
- [ ] All enums using unified unmarshaling
- [ ] All error messages preserve context
- [ ] All high-complexity functions refactored
- [ ] Zero circular dependencies (already achieved)

### 7.3 Developer Experience Metrics

- [ ] Onboarding time < 1 hour (with architecture docs)
- [ ] New cleaner can be added in < 30 minutes
- [ ] Error messages are actionable
- [ ] Configuration format is clear and documented

---

## Conclusion

Clean Wizard is already **world-class** with excellent architecture and type safety. The improvements outlined in this plan are **strategic enhancements** that will:

1. **Unblock Development** - Clean disk space, verify baseline
2. **Strengthen Type Safety** - Integration tests, cleaner verification
3. **Improve Code Quality** - Cleaner interface, complexity reduction
4. **Enhance Developer Experience** - Better error messages, documentation

**Recommended Execution**:

- Start with Phase 1 (critical blockers)
- Prioritize Phase 2 (high impact, low effort) - Pareto 1% → 51%
- Proceed to Phase 3-4 based on results
- Evaluate Phase 5-6 based on business needs

**Timeline**: 7 weeks total, with incremental value delivery every week.

---

**Next Immediate Actions**:

1. Clean disk space ✅ IN PROGRESS
2. Run full test suite to establish baseline
3. Begin Phase 2: Extract Cleaner Interface
4. Continue with high-impact, low-effort improvements

**Document Status**: Complete ✅
**Ready for Execution**: Yes ✅
**Confidence Level**: High ✅
