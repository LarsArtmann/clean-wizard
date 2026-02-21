# Clean Wizard Architecture

**Version:** 1.0.0
**Last Updated:** 2026-02-20

---

## Overview

Clean Wizard is a Go-based system cleanup tool for macOS (with partial Linux support). It provides a modular, extensible architecture for cleaning various development artifacts, caches, and temporary files.

## Core Architecture

### Layer Structure

```
cmd/clean-wizard/           # CLI entry point
    └── commands/           # Cobra commands (clean, scan, init, profile, config)

internal/
    ├── cleaner/            # Cleaner implementations (13 cleaners)
    ├── domain/             # Core domain types and enums
    ├── result/             # Result[T] type for railway-oriented programming
    ├── shared/
    │   └── context/        # Generic Context[T] system
    └── adapters/           # External system adapters
```

---

## Key Components

### 1. Cleaner Interface

All cleaners implement the `Cleaner` interface defined in `internal/cleaner/cleaner.go`:

```go
type Cleaner interface {
    Name() string
    Clean(ctx context.Context) result.Result[domain.CleanResult]
    IsAvailable(ctx context.Context) bool
    Scan(ctx context.Context) result.Result[[]domain.ScanItem]
}
```

**Methods:**

- `Name()` — Unique identifier for registry operations
- `Clean(ctx)` — Execute cleaning operation, returns typed result
- `IsAvailable(ctx)` — Check if cleaner can run on current system
- `Scan(ctx)` — Preview cleanable items for dry-run mode

### 2. Cleaner Registry

Thread-safe registry pattern for managing cleaners (`internal/cleaner/registry.go`):

```go
registry := cleaner.NewRegistry()
registry.Register("docker", dockerCleaner)
registry.Register("golang", golangCleaner)

// Get all available cleaners
available := registry.Available(ctx)

// Clean all at once
results := registry.CleanAll(ctx)
```

**Features:**

- Thread-safe with `sync.RWMutex`
- Polymorphic operations over all cleaners
- Automatic availability filtering

### 3. Generic Context System

Type-safe context holder for validation, error handling, and sanitization (`internal/shared/context/context.go`):

```go
type Context[T any] struct {
    Context     context.Context
    ValueType   T
    Metadata    map[string]string
    Permissions []string
}
```

**Common Usage Patterns:**

```go
// Validation context
valCtx := context.NewContext[ValidationConfig](ctx, NewValidationConfig())
valCtx = valCtx.WithMetadata("trace_id", "123")

// Error context
errCtx := context.NewContext[ErrorConfig](ctx, errorConfig)

// Sanitization context
sanCtx := context.NewContext[SanitizationConfig](ctx, sanConfig)
```

**Legacy Compatibility:**

- `ToLegacyValidationContext()` / `FromLegacyValidationContext()` for backward compatibility
- Deprecated types marked for v2.0 removal

### 4. Result[T] Type

Railway-oriented programming for composable error handling (`internal/result/type.go`):

```go
// Create results
result := result.Ok(cleaningResult)
result := result.Err[domain.CleanResult](errors.New("failed"))

// Chain operations
result := result.Ok(data).
    Validate(validator, "validation failed").
    AndThen(processData).
    Tap(logSuccess)

// Safe unwrapping
value, err := result.SafeValue()
value := result.UnwrapOr(defaultValue)
```

**Key Methods:**

- `Map()` — Transform success values
- `AndThen()` / `FlatMap()` — Chain operations that return Result
- `Validate()` — Add validation predicates
- `OrElse()` — Provide fallback results

---

## Domain Enums

All enums follow the `TypeSafeEnum[T]` interface:

```go
type TypeSafeEnum[T any] interface {
    String() string
    IsValid() bool
    Values() []T
}
```

### Available Enums

| Enum                     | Values                                   | File                    |
| ------------------------ | ---------------------------------------- | ----------------------- |
| `RiskLevelType`          | LOW, MEDIUM, HIGH, CRITICAL              | `type_safe_enums.go`    |
| `ValidationLevelType`    | NONE, BASIC, COMPREHENSIVE, STRICT       | `type_safe_enums.go`    |
| `ChangeOperationType`    | ADDED, REMOVED, MODIFIED                 | `type_safe_enums.go`    |
| `CleanStrategyType`      | aggressive, conservative, dry-run        | `type_safe_enums.go`    |
| `SizeEstimateStatusType` | KNOWN, UNKNOWN                           | `type_safe_enums.go`    |
| `CacheCleanupMode`       | DISABLED, ENABLED                        | `operation_settings.go` |
| `DockerPruneMode`        | ALL, IMAGES, CONTAINERS, VOLUMES, BUILDS | `operation_settings.go` |
| `BuildToolType`          | GO, RUST, NODE, PYTHON, JAVA, SCALA      | `operation_settings.go` |
| `CacheType`              | SPOTLIGHT, XCODE, COCOAPODS, ...         | `operation_settings.go` |
| `PackageManagerType`     | NPM, PNPM, YARN, BUN                     | `operation_settings.go` |

### YAML/JSON Serialization

All enums support both string and integer representations:

```yaml
risk_level: HIGH      # String form
risk_level: 2         # Integer form (equivalent to HIGH)
```

---

## Implemented Cleaners

| Cleaner                               | Purpose                                | Platform     |
| ------------------------------------- | -------------------------------------- | ------------ |
| `BuildCacheCleaner`                   | Build tool caches (Go, Cargo, etc.)    | macOS, Linux |
| `CargoCleaner`                        | Rust/Cargo artifacts                   | macOS, Linux |
| `DockerCleaner`                       | Docker images, containers, volumes     | macOS, Linux |
| `GolangCleaner`                       | Go module cache, test cache            | macOS, Linux |
| `HomebrewCleaner`                     | Homebrew cache and old versions        | macOS        |
| `NixCleaner`                          | Nix store generations                  | macOS, Linux |
| `NodePackagesCleaner`                 | npm/yarn/pnpm caches                   | macOS, Linux |
| `ProjectExecutablesCleaner`           | Compiled binaries in projects          | macOS, Linux |
| `ProjectsManagementAutomationCleaner` | Project automation files (deprecated)  | macOS, Linux |
| `SystemCacheCleaner`                  | System caches (Spotlight, Xcode, etc.) | macOS        |
| `TempFilesCleaner`                    | Temporary files                        | macOS, Linux |

---

## Clean Operation Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   CLI       │────▶│   Registry  │────▶│   Cleaner   │
│  Commands   │     │             │     │   Clean()   │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                    │
                           ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │ IsAvailable │     │   Result[T] │
                    │   Check     │     │  Ok / Err   │
                    └─────────────┘     └─────────────┘
```

1. **CLI** parses command and flags
2. **Registry** filters to available cleaners
3. **Cleaner.Scan()** previews items (dry-run) or **Cleaner.Clean()** executes
4. **Result[T]** wraps success/error for type-safe handling
5. **Aggregation** combines results for reporting

---

## Configuration

Configuration is loaded via Viper from:

1. `./clean-wizard.yaml` (project-local)
2. `~/.config/clean-wizard/config.yaml` (user-level)
3. Environment variables (prefix: `CLEAN_WIZARD_`)

---

## Testing Strategy

- **BDD Tests**: Ginkgo/Gomega for behavioral specifications
- **Unit Tests**: Standard Go testing for isolated components
- **Test Helpers**: `test_factories.go`, `test_helpers.go` for DRY test setup

Run tests:

```bash
go test ./...
```

---

## Future Considerations

| Topic                | Status        | Notes                                  |
| -------------------- | ------------- | -------------------------------------- |
| Dependency Injection | Investigating | Consider `samber/do/v2`                |
| Plugin Architecture  | Deferred      | Interface already supports plugins     |
| Linux Support        | Partial       | SystemCache needs Linux paths          |
| Complexity Reduction | In Progress   | 21 functions >10 cyclomatic complexity |

---

## Key Files Reference

| File                                           | Purpose                                    |
| ---------------------------------------------- | ------------------------------------------ |
| `internal/cleaner/cleaner.go`                  | Cleaner interface definition               |
| `internal/cleaner/registry.go`                 | Thread-safe cleaner registry               |
| `internal/domain/type_safe_enums.go`           | Core domain enums                          |
| `internal/domain/operation_settings.go`        | Operation-specific enums                   |
| `internal/domain/types.go`                     | Domain types (CleanResult, ScanItem, etc.) |
| `internal/result/type.go`                      | Result[T] railway type                     |
| `internal/shared/context/context.go`           | Generic Context[T] system                  |
| `internal/shared/context/validation_config.go` | Validation context types                   |
