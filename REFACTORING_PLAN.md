# Comprehensive Code Duplication Analysis & Refactoring Plan

## 🎯 Executive Summary

**Goal**: Eliminate 62 code duplicates found by `art-dupl` while improving architecture and type safety.

**Impact**: Reduce code duplication by ~70%, improve maintainability, and establish patterns for future development.

## 📊 Analysis Results

### Top Duplication Patterns by Impact:

| Pattern                   | Files Affected | Impact | Work Required | Priority |
| ------------------------- | -------------- | ------ | ------------- | -------- |
| **Validation Wrapping**   | 4 files        | HIGH   | LOW           | 🥇 #1    |
| **Config Loading**        | 2 files        | HIGH   | LOW           | 🥈 #2    |
| **String Trimming**       | 2 files        | MEDIUM | LOW           | 🥉 #3    |
| **Error Detail Setting**  | 3 files        | MEDIUM | MEDIUM        | #4       |
| **Test Helper Functions** | 8+ files       | MEDIUM | MEDIUM        | #5       |
| **Schema Min/Max Logic**  | 2 files        | LOW    | LOW           | #6       |
| **BDD Test Structure**    | 10+ files      | LOW    | HIGH          | #7       |

## 🔍 Self-Reflection & Learnings

### What I Could Have Done Better:

1. **Better Initial Analysis** - Should have examined existing Result type and error infrastructure first
2. **Type System Discovery** - Should have identified the Validate() interface pattern earlier
3. **Library Awareness** - Should have leveraged existing libraries (logrus, testify) more effectively
4. **Generic Patterns** - Should have considered Go generics for type-safe utilities earlier

### Existing Infrastructure We Should Leverage:

- ✅ **Result[T]** type - Well-designed, supports chaining
- ✅ **Validate() interface** - Multiple types already implement
- ✅ **pkg/errors** - Structured error handling
- ✅ **logrus** - Established logging patterns
- ✅ **testify** - Rich testing utilities

## 🚀 Comprehensive Refactoring Plan

### Phase 1: High-Impact, Low-Effort (Quick Wins)

#### 1.1 Create Generic Validation Interface ⚡

**Files**: `internal/shared/utils/validation/`
**Impact**: Eliminates 4 validation duplicates
**Work**: 2 hours

```go
// Validator interface for types with Validate() method
type Validator interface {
    Validate() error
}

// Generic validation wrapper using our Result type
func ValidateAndWrap[T Validator](item T, itemType string) result.Result[T] {
    if err := item.Validate(); err != nil {
        return result.Err[T](fmt.Errorf("invalid %s: %w", itemType, err))
    }
    return result.Ok(item)
}
```

#### 1.2 Create Config Loading Utility ⚡

**Files**: `internal/shared/utils/config/`
**Impact**: Eliminates 2 config loading duplicates  
**Work**: 1 hour

```go
func LoadConfigWithFallback(ctx context.Context, logger *logrus.Logger) (*domain.Config, error) {
    loadedCfg, err := config.LoadWithContext(ctx)
    if err != nil {
        logger.Warnf("Could not load default configuration: %v", err)
        return nil, err // Return error for caller to handle
    }
    logger.Info("Using configuration from ~/.clean-wizard.yaml")
    return loadedCfg, nil
}
```

#### 1.3 Create String Trimming Utility ⚡

**Files**: `internal/shared/utils/strings/`
**Impact**: Eliminates 2 string trimming duplicates
**Work**: 30 minutes

```go
type TrimmableField struct {
    Name  string
    Value *string
    Path  string
}

func TrimWhitespaceField(field TrimmableField, changes *SanitizationResult) bool {
    original := *field.Value
    *field.Value = strings.TrimSpace(*field.Value)
    if original != *field.Value {
        changes.addChange(field.Path, original, *field.Value, "trimmed whitespace")
        return true
    }
    return false
}
```

### Phase 2: Medium-Impact Improvements

#### 2.1 Create Error Details Utility

**Files**: `internal/pkg/errors/`
**Impact**: Eliminates 3 error detail setting duplicates
**Work**: 2 hours

#### 2.2 Refactor Test Helper Functions

**Files**: `tests/bdd/helpers/`
**Impact**: Eliminates 8+ test helper duplicates
**Work**: 3 hours

#### 2.3 Create Schema Min/Max Utility

**Files**: `internal/shared/utils/schema/`
**Impact**: Eliminates 2 schema logic duplicates
**Work**: 1 hour

### Phase 3: Architecture Improvements

#### 3.1 Improve Type Models

**Files**: `internal/domain/interfaces.go`
**Impact**: Better compile-time guarantees
**Work**: 4 hours

```go
// Enforce Validate() interface at compile time
type (
    ValidConfig interface {
        *domain.Config
        Validate() error
    }

    ValidCleanRequest interface {
        domain.CleanRequest
        Validate() error
    }
    // ... other types
)
```

#### 3.2 Enhance Result Type for Validation Chaining

**Files**: `internal/result/validation.go`
**Impact**: Better validation composition
**Work**: 2 hours

## 🎯 Prioritized Execution Order

### Week 1: Quick Wins (High Impact, Low Effort)

1. ✅ Generic validation interface (2h) - **IMPACT: 4 duplicates eliminated**
2. ✅ Config loading utility (1h) - **IMPACT: 2 duplicates eliminated**
3. ✅ String trimming utility (0.5h) - **IMPACT: 2 duplicates eliminated**

### Week 2: Medium Impact Improvements

4. Error details utility (2h) - **IMPACT: 3 duplicates eliminated**
5. Test helper refactoring (3h) - **IMPACT: 8+ duplicates eliminated**
6. Schema min/max utility (1h) - **IMPACT: 2 duplicates eliminated**

### Week 3: Architecture Enhancement

7. Type model improvements (4h) - **IMPACT: Better architecture**
8. Result type enhancement (2h) - **IMPACT: Better validation patterns**

## 📈 Expected Metrics

### Before Refactoring:

- **Total Duplicates**: 62
- **Lines of Duplicate Code**: ~500+
- **Maintenance Overhead**: High
- **Type Safety**: Medium

### After Refactoring:

- **Total Duplicates**: ~15 (75% reduction)
- **Lines of Duplicate Code**: ~100 (80% reduction)
- **Maintenance Overhead**: Low
- **Type Safety**: High

## 🔧 Established Libraries to Leverage

| Library     | Current Usage | Proposed Usage                               |
| ----------- | ------------- | -------------------------------------------- |
| **logrus**  | ✅ Used       | ✅ Expand for config loading                 |
| **testify** | ✅ Used       | ✅ More test utilities                       |
| **viper**   | ✅ Used       | ✅ Config patterns                           |
| **cobra**   | ✅ Used       | ✅ Command patterns                          |
| **go.mod**  | ✅ Used       | ✅ Consider adding generic utility libraries |

## 🚦 Success Criteria

1. **Duplicate Reduction**: >70% reduction in duplicate lines
2. **Type Safety**: All validation uses generic interfaces
3. **Test Coverage**: No reduction in test coverage
4. **Build Time**: No significant increase
5. **Runtime Performance**: No regression

## 📋 Implementation Checklist

### For Each Refactor:

- [ ] Create utility function
- [ ] Replace duplicate code
- [ ] Run `go test ./...`
- [ ] Commit changes
- [ ] Verify with `art-dupl`
- [ ] Update documentation

### Final Verification:

- [ ] All tests pass
- [ ] Build succeeds
- [ ] `art-dupl` shows significant improvement
- [ ] No functional regressions
- [ ] Documentation updated

---

## 🎯 Next Steps

**IMMEDIATE ACTION**: Start with Phase 1.1 (Generic Validation Interface) as it provides the highest impact with lowest effort and establishes patterns for subsequent refactoring.

**ESTIMATED TOTAL TIME**: ~15.5 hours over 3 weeks
**ESTIMATED IMPACT**: 75% reduction in code duplication
