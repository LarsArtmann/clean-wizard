# go-composable-business-types/id Integration Analysis

**Date:** 2026-03-17
**Library:** `github.com/larsartmann/go-composable-business-types/id`
**Status:** Analysis Complete - Ready for Implementation

---

## Executive Summary

The `go-composable-business-types/id` library provides **branded, strongly-typed identifiers** using Go generics. It prevents mixing different entity IDs at compile time, eliminating an entire class of runtime bugs. For clean-wizard, this library offers significant type-safety improvements for identifiers currently implemented as raw `string` or `int` types.

### Key Benefits for clean-wizard

| Benefit                | Impact                                                         |
| ---------------------- | -------------------------------------------------------------- |
| Compile-time ID safety | Prevent accidental mixing of ProfileID, OperationID, CleanerID |
| Self-documenting APIs  | Function signatures clearly show expected ID types             |
| Serialization support  | JSON/YAML/SQL compatible with existing config format           |
| Zero-value handling    | Null serialization for unset IDs                               |
| Performance            | ~1-2 ns/op for NewID, ~1 ns/op for Get                         |

---

## Library Capabilities

### Core Features

```go
// Type-safe ID creation
userID := id.NewID[UserBrand]("user-123")
orderID := id.NewID[OrderBrand, int64](456)

// Compile-time safety - these would fail:
// GetOrder(userID)  // ERROR: type mismatch
// GetUser(orderID)  // ERROR: type mismatch
```

### Supported Operations

| Operation            | Description              |
| -------------------- | ------------------------ |
| `NewID[B, V](v V)`   | Create new branded ID    |
| `Get() V`            | Extract underlying value |
| `IsZero() bool`      | Check if zero value      |
| `Equal(other) bool`  | Type-safe equality       |
| `Compare(other) int` | Ordering comparison      |
| `Or(default) ID`     | Default value fallback   |

### Serialization Support

- **JSON**: String/number based on underlying type, zero → `null`
- **SQL**: `Scan()` and `Value()` for database operations
- **Binary**: Efficient binary encoding
- **Text**: XML/TOML compatible
- **Gob**: Go-specific encoding

---

## Current ID Usage Analysis

### 1. Profile Identifiers (`internal/domain/config.go`)

**Current State:**

```go
type Config struct {
    Profiles       map[string]*Profile `json:"profiles" yaml:"profiles"`
    CurrentProfile string              `json:"current_profile,omitempty" yaml:"current_profile,omitempty"`
}

type Profile struct {
    Name        string             `json:"name" yaml:"name"`
    // ...
}
```

**Problem:** Profile names are raw strings. Functions accept `name string` parameters, allowing any string to be passed.

**Recommended Type:**

```go
type ProfileBrand struct{}
type ProfileID = id.ID[ProfileBrand, string]
```

---

### 2. Operation Identifiers (`internal/domain/operation_types.go`)

**Current State:**

```go
type OperationType string

const (
    OperationTypeNixGenerations OperationType = "nix-generations"
    OperationTypeTempFiles    OperationType = "temp-files"
    // ... 15 total
)

func GetOperationType(name string) OperationType
```

**Problem:** Operations are strings. The `GetOperationType()` function accepts any string and falls back to raw conversion.

**Recommended Type:**

```go
type OperationBrand struct{}
type OperationID = id.ID[OperationBrand, string]
```

---

### 3. Cleaner Identifiers (`internal/cleaner/registry.go`)

**Current State:**

```go
const (
    CleanerNix              = "nix"
    CleanerHomebrew         = "homebrew"
    CleanerDocker           = "docker"
    // ... 12 total
)

func (r *Registry) Register(name string, c Cleaner)
func (r *Registry) Get(name string) (Cleaner, bool)
```

**Problem:** Cleaner names are string constants. Registry methods accept any string.

**Recommended Type:**

```go
type CleanerBrand struct{}
type CleanerID = id.ID[CleanerBrand, string]
```

---

### 4. Generation Identifiers (`internal/domain/types.go`)

**Current State:**

```go
type NixGeneration struct {
    ID   int    `json:"id"`
    Path string `json:"path"`
    // ...
}
```

**Problem:** Generation ID is raw `int`. Could be confused with other integer IDs.

**Recommended Type:**

```go
type GenerationBrand struct{}
type GenerationID = id.ID[GenerationBrand, int]
```

---

### 5. Git Hash Identifiers (`internal/domain/githistory_types.go`)

**Current State:**

```go
type GitHistoryFile struct {
    BlobHash   string `json:"blob_hash"`
    CommitHash string `json:"commit_hash"`
    // ...
}
```

**Problem:** Git hashes are raw strings. Easy to accidentally swap blob/commit hashes.

**Recommended Types:**

```go
type BlobHashBrand struct{}
type CommitHashBrand struct{}
type BlobHash = id.ID[BlobHashBrand, string]
type CommitHash = id.ID[CommitHashBrand, string]
```

---

## Integration Strategy

### Phase 1: Domain Types (Recommended First)

Start with low-risk, high-impact changes in the domain layer:

1. **ProfileID** in `internal/domain/config.go`
2. **OperationID** in `internal/domain/operation_types.go`
3. **GenerationID** in `internal/domain/types.go`

### Phase 2: Infrastructure Layer

Update registry and adapters:

4. **CleanerID** in `internal/cleaner/registry.go`
5. Update registry methods to accept `CleanerID` instead of `string`

### Phase 3: Git History Types

6. **BlobHash** and **CommitHash** in `internal/domain/githistory_types.go`

---

## Implementation Examples

### Example 1: Profile ID

```go
// internal/domain/profile_id.go
package domain

import "github.com/larsartmann/go-composable-business-types/id"

type ProfileBrand struct{}

// ProfileID is a strongly-typed identifier for profiles.
type ProfileID = id.ID[ProfileBrand, string]

// NewProfileID creates a new ProfileID from a string.
func NewProfileID(name string) ProfileID {
    return id.NewID[ProfileBrand](name)
}
```

**Updated Config:**

```go
type Config struct {
    Profiles       map[ProfileID]*Profile `json:"profiles" yaml:"profiles"`
    CurrentProfile ProfileID              `json:"current_profile,omitempty"`
    // ...
}
```

### Example 2: Generation ID

```go
// internal/domain/generation_id.go
package domain

type GenerationBrand struct{}
type GenerationID = id.ID[GenerationBrand, int]

func NewGenerationID(id int) GenerationID {
    return id.NewID[GenerationBrand](id)
}
```

**Updated NixGeneration:**

```go
type NixGeneration struct {
    ID   GenerationID `json:"id"`
    Path string       `json:"path"`
    // ...
}
```

### Example 3: Cleaner Registry

```go
// internal/cleaner/cleaner_id.go
package cleaner

import "github.com/larsartmann/go-composable-business-types/id"

type CleanerBrand struct{}

// CleanerID is a strongly-typed identifier for cleaners.
type CleanerID = id.ID[CleanerBrand, string]

// Predefined cleaner IDs
var (
    CleanerNixID      = id.NewID[CleanerBrand]("nix")
    CleanerHomebrewID = id.NewID[CleanerBrand]("homebrew")
    CleanerDockerID   = id.NewID[CleanerBrand]("docker")
    // ...
)
```

**Updated Registry Methods:**

```go
func (r *Registry) Register(id CleanerID, c Cleaner)
func (r *Registry) Get(id CleanerID) (Cleaner, bool)
```

---

## Migration Path

### Backward Compatibility

The library's serialization maintains backward compatibility:

- String-based IDs serialize as JSON strings
- Int-based IDs serialize as JSON numbers
- Zero values serialize as `null`

**Example:**

```json
// Before (raw string)
{"current_profile": "development"}

// After (ProfileID) - SAME JSON!
{"current_profile": "development"}
```

### Gradual Migration Strategy

1. **Add ID types** alongside existing code
2. **Update functions** to accept both old and new types
3. **Migrate call sites** incrementally
4. **Remove old types** once migration complete

---

## Cost-Benefit Analysis

### Costs

| Cost              | Description                | Mitigation                                       |
| ----------------- | -------------------------- | ------------------------------------------------ |
| Import dependency | New external dependency    | Library is lightweight, no transitive deps       |
| Type verbosity    | More types to learn        | Use type aliases (`type ProfileID = id.ID[...]`) |
| Migration effort  | Update function signatures | Gradual migration, backward compatible           |

### Benefits

| Benefit               | Impact | Priority |
| --------------------- | ------ | -------- |
| Compile-time safety   | High   | Critical |
| Self-documenting APIs | Medium | High     |
| Refactoring safety    | High   | Critical |
| IDE support           | Medium | Medium   |

---

## Recommended Action

### Immediate Steps

1. **Add dependency** to `go.mod`:

   ```bash
   go get github.com/larsartmann/go-composable-business-types/id
   ```

2. **Create domain ID types**:
   - `internal/domain/profile_id.go`
   - `internal/domain/operation_id.go`
   - `internal/domain/generation_id.go`

3. **Update Config struct** to use `ProfileID`

4. **Update Registry** to use `CleanerID`

### Success Criteria

- [ ] All profile operations use `ProfileID`
- [ ] All operation lookups use `OperationID`
- [ ] Registry uses `CleanerID` for registration/lookup
- [ ] No regression in JSON/YAML serialization
- [ ] All tests pass

---

## Conclusion

The `go-composable-business-types/id` library provides **significant type-safety benefits** with **minimal overhead** for clean-wizard. The branded ID pattern eliminates an entire class of runtime errors where identifiers are accidentally swapped or misused.

**Recommendation:** Proceed with Phase 1 implementation (domain types) as a proof of concept. The backward-compatible serialization ensures no breaking changes to existing configuration files.

---

## References

- Library: `/Users/larsartmann/projects/go-composable-business-types/id/`
- Domain types: `internal/domain/`
- Registry: `internal/cleaner/registry.go`
- Config: `internal/domain/config.go`
