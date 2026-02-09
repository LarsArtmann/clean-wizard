# Cleaner Interface Implementation Analysis

**Date**: February 9, 2026
**Task**: Verify all cleaners implement Cleaner interface correctly

---

## Cleaner Interface Definition

Location: `internal/cleaner/cleaner.go:10-17`

```go
type Cleaner interface {
    // Clean executes the cleaning operation and returns the result.
    Clean(ctx context.Context) result.Result[domain.CleanResult]

    // IsAvailable checks if the cleaner can run on this system.
    IsAvailable(ctx context.Context) bool
}
```

---

## Implementation Status

### Cleaners That Implement Cleaner Interface ✅

1. **buildcache** (`internal/cleaner/buildcache.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

2. **cargo** (`internal/cleaner/cargo.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

3. **docker** (`internal/cleaner/docker.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

4. **golang_cleaner** (`internal/cleaner/golang_cleaner.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

5. **golang_lint_adapter** (`internal/cleaner/golang_lint_adapter.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

6. **homebrew** (`internal/cleaner/homebrew.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

7. **langversionmanager** (`internal/cleaner/langversionmanager.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

8. **nodepackages** (`internal/cleaner/nodepackages.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

9. **projectsmanagementautomation** (`internal/cleaner/projectsmanagementautomation.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ✅ IsAvailable(ctx context.Context) bool

10. **systemcache** (`internal/cleaner/systemcache.go`)
    - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
    - ✅ IsAvailable(ctx context.Context) bool

11. **tempfiles** (`internal/cleaner/tempfiles.go`)
    - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
    - ✅ IsAvailable(ctx context.Context) bool

### Cleaners That DO NOT Implement Cleaner Interface ❌

1. **nix** (`internal/cleaner/nix.go`)
   - ❌ Clean(ctx context.Context) result.Result[domain.CleanResult] - MISSING
   - ✅ IsAvailable(ctx context.Context) bool

   **Issue**: Has `CleanOldGenerations(ctx context.Context, keepCount int)` instead of `Clean(ctx context.Context)`

   **Fix Options**:
   - Option 1: Add `Clean(ctx context.Context)` method that calls `CleanOldGenerations` with default keep count
   - Option 2: Refactor to use OperationSettings to pass parameters

2. **golang_cache_cleaner** (`internal/cleaner/golang_cache_cleaner.go`)
   - ✅ Clean(ctx context.Context) result.Result[domain.CleanResult]
   - ❌ IsAvailable(ctx context.Context) bool - MISSING

   **Issue**: No availability check

   **Fix Options**:
   - Option 1: Add `IsAvailable()` method that checks if `go` command exists
   - Option 2: Add `IsAvailable()` method that checks cache paths exist

---

## Non-Cleaner Files (Skipped)

Helper/utility files that are not expected to implement Cleaner interface:
- `cleaner.go` (defines interface)
- `helpers.go`
- `test_helpers.go`
- `testing_helpers.go`
- `validate.go`
- `fsutil.go`
- `golang_conversion.go`
- `golang_helpers.go`
- `golang_scanner.go`
- `golang_types.go`
- All `*_test.go` files

---

## Summary

- **Total Cleaners**: 13
- **Implementing Cleaner Interface**: 11 (84.6%)
- **NOT Implementing Cleaner Interface**: 2 (15.4%)

**Critical Findings**:
1. Cleaner interface already exists and is well-designed
2. Most cleaners already implement it correctly
3. Two cleaners need fixes:
   - nix.go: Needs Clean() method
   - golang_cache_cleaner.go: Needs IsAvailable() method

**Impact**:
- Without these fixes, nix and golang_cache_cleaner cannot be used polymorphically
- Cannot iterate over all cleaners as Cleaner interface
- Cannot create a unified cleaner registry

**Next Steps**:
1. Fix nix.go to implement Cleaner interface
2. Fix golang_cache_cleaner.go to implement Cleaner interface
3. Create CleanerRegistry type
4. Update cmd/clean-wizard/commands/clean.go to use Cleaner interface
5. Add tests for interface implementation
