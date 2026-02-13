# Clean Wizard Architecture

This document describes the architecture of Clean Wizard - a comprehensive system cleanup tool for macOS and Linux with type-safe architecture.

## Table of Contents

1. [Overview](#overview)
2. [Design Principles](#design-principles)
3. [Project Structure](#project-structure)
4. [Core Components](#core-components)
5. [Data Flow](#data-flow)
6. [Type System](#type-system)
7. [Error Handling](#error-handling)
8. [Configuration](#configuration)
9. [Platform Support](#platform-support)
10. [Testing Strategy](#testing-strategy)
11. [Extension Points](#extension-points)

---

## Overview

Clean Wizard is built around a **Registry Pattern** with **Type-Safe Enums** and **Result Types**. The architecture emphasizes compile-time safety, explicit error handling, and clean separation of concerns.

### Key Architectural Decisions

| Decision               | Rationale                                     |
| ---------------------- | --------------------------------------------- |
| Type-Safe Enums        | Compile-time safety, no string comparisons    |
| Result[T] Type         | Explicit error handling without exceptions    |
| Registry Pattern       | Thread-safe cleaner management, extensibility |
| Interface-Based Design | All cleaners implement common interface       |
| Platform Abstraction   | Clean separation of macOS/Linux specific code |

---

## Design Principles

### 1. Make Impossible States Unrepresentable

The type system prevents invalid configurations at compile time:

```go
// Invalid states cannot be constructed
type RiskLevelType int
const (
    RiskLevelLowType RiskLevelType = iota
    RiskLevelMediumType
    RiskLevelHighType
    RiskLevelCriticalType
)

// IsValid() ensures runtime safety for external input
func (r RiskLevelType) IsValid() bool {
    return r >= RiskLevelLowType && r <= RiskLevelCriticalType
}
```

### 2. Railway-Oriented Error Handling

Operations return `Result[T]` instead of throwing errors:

```go
func (c *DockerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    if err := c.checkDockerAvailable(); err != nil {
        return result.Err[domain.CleanResult](err)
    }
    // ... cleaning logic
    return result.Ok(cleanResult)
}
```

### 3. Single Responsibility

Each cleaner handles exactly one domain:

- `NixCleaner` - Nix store and generations
- `HomebrewCleaner` - Homebrew cache and autoremove
- `DockerCleaner` - Docker containers, images, volumes
- `CargoCleaner` - Rust Cargo cache
- `GoCleaner` - Go module, test, and build cache
- `NodeCleaner` - npm, pnpm, yarn, bun caches
- `BuildCacheCleaner` - Gradle, Maven, SBT caches
- `SystemCacheCleaner` - OS-specific caches (macOS/Linux)
- `TempFilesCleaner` - Age-based temporary files

### 4. Composition Over Inheritance

Cleaners implement the `Cleaner` interface and compose behavior through helper functions rather than inheritance hierarchies:

```go
type Cleaner interface {
    Name() string
    Clean(ctx context.Context) result.Result[domain.CleanResult]
    IsAvailable(ctx context.Context) bool
}
```

---

## Project Structure

```
clean-wizard/
├── cmd/
│   └── clean-wizard/           # CLI entry point
├── internal/
│   ├── cleaner/                # Cleaner implementations
│   │   ├── cleaner.go           # Cleaner interface
│   │   ├── registry.go          # Thread-safe registry
│   │   ├── registry_factory.go  # Default cleaner setup
│   │   ├── helpers.go           # Shared utilities
│   │   ├── nix.go               # Nix cleanup
│   │   ├── homebrew.go          # Homebrew cleanup
│   │   ├── docker.go            # Docker cleanup
│   │   ├── cargo.go             # Cargo cleanup
│   │   ├── golang_*.go          # Go cache cleanup
│   │   ├── node_*.go            # Node package managers
│   │   ├── buildcache.go        # Build tool caches
│   │   ├── systemcache.go       # OS-specific caches
│   │   └── tempfiles.go         # Temporary files
│   ├── domain/                 # Core domain types
│   │   ├── type_safe_enums.go   # Type-safe enumerations
│   │   ├── types.go             # Domain types
│   │   ├── interfaces.go        # Domain interfaces
│   │   └── operation_*.go       # Operation settings
│   ├── result/                 # Result[T] type
│   │   └── type.go              # Railway-oriented error handling
│   ├── config/                 # Configuration
│   │   ├── config.go            # Configuration structures
│   │   ├── enhanced_loader*.go  # Configuration loading
│   │   ├── sanitizer*.go        # Input sanitization
│   │   └── validator*.go        # Validation rules
│   ├── adapters/               # External tool adapters
│   │   ├── nix.go               # Nix command adapter
│   │   ├── exec.go              # Command execution
│   │   └── environment.go       # Environment detection
│   ├── conversions/            # Type conversions
│   ├── format/                 # Output formatting
│   ├── errors/                 # Error types
│   └── middleware/             # Middleware patterns
├── tests/
│   ├── integration/            # Integration tests
│   └── bdd/                    # BDD tests (Godog)
└── schemas/
    └── config.schema.json      # JSON Schema for config
```

---

## Core Components

### Cleaner Interface

The central abstraction all cleaners implement:

```go
type Cleaner interface {
    // Name returns the unique identifier for this cleaner
    Name() string

    // Clean executes the cleaning operation
    Clean(ctx context.Context) result.Result[domain.CleanResult]

    // IsAvailable checks if the cleaner can run on this system
    IsAvailable(ctx context.Context) bool
}
```

### Registry

Thread-safe cleaner management:

```go
type Registry struct {
    cleaners map[string]Cleaner
    mu       sync.RWMutex
}

// Key methods:
func (r *Registry) Register(name string, c Cleaner)
func (r *Registry) Get(name string) (Cleaner, bool)
func (r *Registry) List() []Cleaner
func (r *Registry) Available(ctx context.Context) []Cleaner
func (r *Registry) CleanAll(ctx context.Context) map[string]result.Result[domain.CleanResult]
```

### Result[T] Type

Railway-oriented error handling:

```go
type Result[T any] struct {
    value T
    err   error
}

// Constructors
func Ok[T any](value T) Result[T]
func Err[T any](err error) Result[T]

// Accessors
func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool
func (r Result[T]) Value() T           // Panics on error
func (r Result[T]) Error() error       // Panics on success
func (r Result[T]) Unwrap() (T, error) // Safe access

// Combinators
func Map[T, U any](r Result[T], fn func(T) U) Result[U]
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U]
func (r Result[T]) OrElse(fallback Result[T]) Result[T]
func (r Result[T]) Validate(predicate func(T) bool, errorMsg string) Result[T]
```

---

## Data Flow

### Clean Operation Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   CLI/CMD    │────▶│   Registry   │────▶│   Cleaner    │
└──────────────┘     └──────────────┘     └──────────────┘
                            │                     │
                            ▼                     ▼
                     ┌──────────────┐     ┌──────────────┐
                     │ IsAvailable? │     │    Clean()   │
                     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │ Result[T]    │
                                          │ - CleanResult│
                                          │ - error      │
                                          └──────────────┘
```

### Configuration Loading Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  YAML File   │────▶│   Loader     │────▶│  Sanitizer   │
└──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │  Validator   │
                                          └──────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │    Config    │
                                          └──────────────┘
```

---

## Type System

### Type-Safe Enums

All enumerations use int-based types with validation:

```go
// Definition
type CleanStrategyType int

const (
    StrategyDryRunType CleanStrategyType = iota
    StrategyConservativeType
    StrategyAggressiveType
)

// Required methods
func (c CleanStrategyType) String() string
func (c CleanStrategyType) IsValid() bool
func (c CleanStrategyType) Values() []CleanStrategyType
func (c CleanStrategyType) MarshalYAML() (any, error)
func (c *CleanStrategyType) UnmarshalYAML(value *yaml.Node) error
```

### Available Enums

| Enum                  | Purpose                  | Values                                                                                  |
| --------------------- | ------------------------ | --------------------------------------------------------------------------------------- |
| `OperationType`       | Cleaner types            | Nix, Homebrew, Docker, Cargo, Go, Node, BuildCache, SystemCache, TempFiles, LangVersion |
| `CleanStrategyType`   | Cleaning behavior        | DryRun, Conservative, Aggressive                                                        |
| `RiskLevelType`       | Risk assessment          | Low, Medium, High, Critical                                                             |
| `CacheType`           | System cache types       | Spotlight, Xcode, Cocoapods, Homebrew, Pip, Npm, Yarn, Ccache, XdgCache, Thumbnails     |
| `BuildToolType`       | Build tools              | Go, Rust, Node, Python, Java, Scala                                                     |
| `DockerPruneMode`     | Docker operations        | All, Images, Containers, Volumes, Builds                                                |
| `PackageManagerType`  | Node.js package managers | Npm, Pnpm, Yarn, Bun                                                                    |
| `ValidationLevelType` | Validation strictness    | None, Basic, Comprehensive, Strict                                                      |

### SizeEstimate Pattern

Honest size reporting with known/unknown status:

```go
type SizeEstimate struct {
    Known  uint64                 // Known size in bytes
    Status SizeEstimateStatusType // Known or Unknown
}

// Usage in CleanResult
type CleanResult struct {
    SizeEstimate SizeEstimate      // Primary
    FreedBytes   uint64            // Deprecated: Use SizeEstimate
    // ...
}
```

---

## Error Handling

### Error Categories

| Category              | Handling                                                 |
| --------------------- | -------------------------------------------------------- |
| Tool not installed    | `IsAvailable()` returns false, cleaner skipped           |
| Permission denied     | Error returned in Result[T], logged, operation continues |
| Invalid configuration | Validation fails at load time                            |
| Runtime errors        | Error returned in Result[T], aggregated in results       |

### Error Flow

```go
// Cleaner returns error
func (c *DockerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    if !c.IsAvailable(ctx) {
        return result.Err[domain.CleanResult](
            errors.New("docker not available"),
        )
    }
    // ...
}

// Caller handles error
res := cleaner.Clean(ctx)
if res.IsErr() {
    log.Printf("Error: %v", res.Error())
    continue // or handle appropriately
}
```

---

## Configuration

### YAML Configuration Structure

```yaml
# ~/.config/clean-wizard/config.yaml

presets:
  quick:
    cleaners: [homebrew, go, node, tempfiles, buildcache]
  standard:
    cleaners: [homebrew, go, node, cargo, tempfiles, buildcache, systemcache, docker, nix]
  aggressive:
    cleaners: [all]
    include_dangerous: true

nix:
  keep_generations: 5
  optimization: true

docker:
  timeout: 2m
  include_volumes: true

tempfiles:
  older_than: 7d
  exclude_paths:
    - /tmp/important-*
```

### Configuration Loading

1. **Load**: Parse YAML into structures
2. **Sanitize**: Clean and normalize input
3. **Validate**: Apply business rules
4. **Apply Defaults**: Fill missing values
5. **Return**: Validated configuration

---

## Platform Support

### Platform Detection

Using `runtime.GOOS` for compile-time platform awareness:

```go
func (scc *SystemCacheCleaner) isMacOS() bool {
    return runtime.GOOS == "darwin"
}

func (scc *SystemCacheCleaner) isLinux() bool {
    return runtime.GOOS == "linux"
}
```

### Platform-Specific Cache Types

| Platform | Cache Types                                            |
| -------- | ------------------------------------------------------ |
| macOS    | Spotlight, Xcode, CocoaPods, Homebrew                  |
| Linux    | XdgCache, Thumbnails, Homebrew, Pip, Npm, Yarn, Ccache |

### Platform Abstraction Pattern

```go
func AvailableSystemCacheTypes() []domain.CacheType {
    switch runtime.GOOS {
    case "darwin":
        return []domain.CacheType{CacheTypeSpotlight, CacheTypeXcode, ...}
    case "linux":
        return []domain.CacheType{CacheTypeXdgCache, CacheTypeThumbnails, ...}
    default:
        return []domain.CacheType{}
    }
}
```

---

## Testing Strategy

### Test Categories

| Category          | Location              | Purpose                      |
| ----------------- | --------------------- | ---------------------------- |
| Unit Tests        | `internal/*_test.go`  | Individual component testing |
| Integration Tests | `tests/integration/`  | Cross-component testing      |
| BDD Tests         | `tests/bdd/`          | User scenario testing        |
| Benchmark Tests   | `*_benchmark_test.go` | Performance measurement      |
| Fuzz Tests        | `*_fuzz_test.go`      | Input robustness             |

### Test Commands

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# BDD tests
go test -tags=bdd ./tests/integration/...

# Benchmarks
go test -bench=. ./...

# Verbose
go test -v ./...
```

---

## Extension Points

### Adding a New Cleaner

1. **Implement the Cleaner interface**:

```go
type MyCleaner struct {
    verbose bool
    dryRun  bool
}

func (mc *MyCleaner) Name() string {
    return "mycleaner"
}

func (mc *MyCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    // Implementation
    return result.Ok(domain.CleanResult{...})
}

func (mc *MyCleaner) IsAvailable(ctx context.Context) bool {
    // Check if tool is installed
    return true
}
```

2. **Register in registry_factory.go**:

```go
func DefaultRegistry() *Registry {
    registry := NewRegistry()
    registry.Register("mycleaner", NewMyCleaner(false, false))
    return registry
}
```

3. **Add domain types** (if needed):
   - Add to `internal/domain/operation_settings.go`
   - Implement `String()`, `IsValid()`, `Values()`, YAML marshaling

4. **Add configuration** (if needed):
   - Add to config structures
   - Add validation rules
   - Update schema

5. **Add tests**:
   - Unit tests
   - Integration tests
   - BDD scenarios

### Adding a New Enum

1. Define type in appropriate file
2. Add constants with `iota`
3. Implement required methods:
   - `String() string`
   - `IsValid() bool`
   - `Values() []Type`
   - `MarshalYAML() (any, error)`
   - `UnmarshalYAML(*yaml.Node) error`

---

## Architecture Decisions Record

### ADR-001: Type-Safe Enums

**Status**: Accepted

**Context**: Need to represent fixed sets of values (strategies, risk levels, etc.)

**Decision**: Use int-based types with validation methods instead of strings

**Consequences**:

- Compile-time type safety
- No string comparison bugs
- IDE autocomplete support
- Requires explicit validation for external input

### ADR-002: Result[T] Error Handling

**Status**: Accepted

**Context**: Need explicit error handling without exceptions

**Decision**: Use generic Result[T] type with Ok/Err constructors

**Consequences**:

- Errors are always visible in function signatures
- Railway-oriented programming patterns
- No hidden control flow
- Requires explicit error checking

### ADR-003: Registry Pattern

**Status**: Accepted

**Context**: Need to manage multiple cleaners with thread-safety

**Decision**: Central registry with RWMutex protection

**Consequences**:

- Thread-safe cleaner management
- Easy to add/remove cleaners at runtime
- Polymorphic operations over all cleaners
- Single source of truth for cleaner availability

---

## References

- [Domain Types](./domain.md)
- [Configuration](./config.md)
- [Cleaner Interface](./cleaner.md)
- [Result Type](./result.md)
- [Adapters](./adapters.md)
