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

// Pattern matching
value := result.Match(
    func(v int) string { return fmt.Sprintf("success: %d", v) },
    func(e error) string { return fmt.Sprintf("error: %v", e) },
)

// Conditional branching
result.When(func(v int) { log.Printf("Value: %d", v) })
result.Unless(func(e error) { log.Printf("Error: %v", e) })
result.Filter(func(v int) bool { return v > 0 }, "must be positive")

// Fold/Reduce operations
results := []Result[int]{Ok(1), Ok(2), Ok(3)}
sum := Fold(results, 0, func(acc, v int) int { return acc + v })
all := FoldAll(results)

// Safe unwrapping
value, err := result.SafeValue()
value := result.UnwrapOr(defaultValue)
```

**Key Methods:**

- `Map()` — Transform success values
- `AndThen()` / `FlatMap()` — Chain operations that return Result
- `Match()` — Pattern matching on Ok/Err branches
- `When()` / `Unless()` — Conditional side effects
- `Filter()` — Predicate-based validation
- `Fold()` / `FoldAll()` — Reduce slice of Results to single Result
- `Sequence()` / `Traverse()` — Convert slice of Results
- `OrElse()` — Provide fallback results

### 5. BranchFlow Type

Complex conditional branching flows (`internal/result/branch_flow.go`):

```go
// Conditional branching with fallback
flow := NewBranchFlow[int]().
    Branch(func() bool { return isAdmin }, func() Result[int] { return Ok(42) }).
    Branch(func() bool { return isUser }, func() Result[int] { return Ok(10) }).
    Fallback(func() Result[int] { return Ok(1) })

result := flow.Execute()

// Value-based branching
cases := []Case[int, string]{
    {Predicate: func(i int) bool { return i < 0 }, Execute: func() Result[string] { return Ok("negative") }},
    {Predicate: func(i int) bool { return i == 0 }, Execute: func() Result[string] { return Ok("zero") }},
    {Predicate: func(i int) bool { return i > 0 }, Execute: func() Result[string] { return Ok("positive") }},
}
result := SwitchFlow(value, cases, func() Result[string] { return Ok("unknown") })
```

### 6. FlowBuilder & Pipeline

Pipeline composition (`internal/result/flow_builder.go`):

```go
// Sequential pipeline
pipeline := NewFlowBuilder[CleanResult]().
    Step("scan", func(ctx context.Context) Result[CleanResult] { return Scan(ctx) }).
    Step("validate", func(ctx context.Context, r CleanResult) Result[CleanResult] { return Validate(ctx, r) }).
    Step("clean", func(ctx context.Context, r CleanResult) Result[CleanResult] { return Clean(ctx, r) })

result := pipeline.Execute(ctx)

// Parallel execution
parallel := NewParallelFlow[CleanResult]().
    Add("docker", func(ctx context.Context) Result[CleanResult] { return CleanDocker(ctx) }).
    Add("go", func(ctx context.Context) Result[CleanResult] { return CleanGo(ctx) }).
    Add("cargo", func(ctx context.Context) Result[CleanResult] { return CleanCargo(ctx) })

results := parallel.Execute(ctx)
successful := parallel.Successful()
failed := parallel.Failed()
```

### 7. Context[T] Branching

Context composition and branching (`internal/shared/context/context.go`):

```go
// Conditional context modification
ctx := NewContext[ValidationConfig](ctx, config)
ctx = ctx.Branch(isStrict, func(c *Context[ValidationConfig]) *Context[ValidationConfig] {
    return c.WithMetadata("strict_mode", "true")
})

// Join multiple contexts
merged := Join(ctx1, ctx2, ctx3)

// Transform value type
strCtx := Transform[int, string](intCtx, func(i int) string { return strconv.Itoa(i) })
```

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
| `internal/result/branch_flow.go`               | BranchFlow and SwitchFlow                  |
| `internal/result/flow_builder.go`              | FlowBuilder and Pipeline                   |
| `internal/shared/context/context.go`           | Generic Context[T] with branching          |
| `internal/shared/context/validation_config.go` | Validation context types                   |
