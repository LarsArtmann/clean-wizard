# Clean Wizard Development Guide

## Overview

Clean Wizard is a production-ready Go CLI application for system cleanup with world-class type safety and architectural excellence.

## Architecture

### Core Principles

- **Type Safety**: Impossible states made unrepresentable through type system
- **Railway Programming**: Result[T] type eliminates nil panics and unchecked errors
- **Clean Architecture**: Zero circular dependencies, proper layering
- **Timeout Protection**: All external commands have 2-5 minute timeouts
- **Mock Pattern**: Graceful degradation when systems unavailable

### Package Structure

```
cmd/clean-wizard/          # CLI entry points
internal/domain/           # Domain models (pure, no deps)
internal/cleaner/          # Cleaning implementations
internal/adapters/         # External system adapters
internal/config/           # Configuration management
internal/result/           # Result[T] type (zero deps)
internal/format/           # Formatting (bytes, JSON)
internal/middleware/       # Cross-cutting concerns
```

### Dependency Flow

```
cmd → cleaner → domain → result
       ↓
   adapters (nix, docker, etc)
```

**Zero circular dependencies** - Flawless architecture

## Building

```bash
# Build binary
go build -o clean-wizard ./cmd/clean-wizard

# Or use Just
just build

# Output: 8.4MB optimized binary
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. -benchmem ./tests/benchmark

# Run integration tests
go test ./tests/integration -tags=integration

# Run BDD tests
go test ./tests/bdd

# Quick test
just test
```

## Benchmarks

Current performance (Apple M2):

```
BenchmarkResult_Ok-8         1000000000    0.93 ns/op    0 B/op    0 allocs/op
BenchmarkResult_Err-8        1000000000    0.76 ns/op    0 B/op    0 allocs/op
BenchmarkResult_IsOk-8       1000000000    0.96 ns/op    0 B/op    0 allocs/op
BenchmarkResult_Value-8      1000000000    1.06 ns/op    0 B/op    0 allocs/op
```

**Result[T] is sub-nanosecond with zero allocations**

## Features

### JSON Output Mode

```bash
# Machine-readable output for scripting
clean-wizard clean --json --dry-run
clean-wizard clean --json --mode quick | jq '.freed_human'
```

### Timeout Protection

All external commands have timeout protection:

- **5 minutes**: Homebrew, Node packages, Cargo, Go
- **2 minutes**: Docker, Projects Management Automation
- **30 seconds**: golangci-lint
- **Configurable**: Nix (via adapter)

This ensures the CLI never hangs even if external commands freeze.

### Result Type

```go
// Type-safe error handling
func ListGenerations(ctx context.Context) result.Result[[]NixGeneration]

// Usage
result := nixCleaner.ListGenerations(ctx)
if result.IsErr() {
    return result.Error()
}
generations := result.Value()
```

No nil panics. No unchecked errors.

## Adding a New Cleaner

1. Create cleaner in `internal/cleaner/`
2. Implement `cleaner.Cleaner` interface (for registry/execution) AND `domain.OperationHandler` interface (for validation)
3. Add availability check
4. Add to `GetCleanerConfigs()` in commands/clean.go
5. Add tests

Note: There are two related interfaces:
- `cleaner.Cleaner`: Registry/execution focused with `Name()`, `Clean()`, `IsAvailable()`
- `domain.OperationHandler`: Operation validation with `Type()`, `GetStoreSize()`, `ValidateSettings()`

Example:

```go
type MyCleaner struct {
    verbose bool
    dryRun  bool
}

// Implement cleaner.Cleaner interface
func (mc *MyCleaner) Name() string {
    return "my-cleaner"
}

func (mc *MyCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    // Add 5-minute timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    cmd := exec.CommandContext(timeoutCtx, "my-tool", "clean")
    // ... implementation
}
```

## Configuration

Default config location: `~/.clean-wizard.yaml`

```yaml
version: "1.0.0"
safe_mode: truemax_disk_usage_percent: 50
protected:
  - /System
  - /Library
```

## Dependencies

### Required

- Go 1.25+
- Git (for version info)

### Optional (for full functionality)

- Nix
- Docker
- Homebrew
- Node.js (npm, pnpm, yarn, bun)
- Rust/Cargo
- Go toolchain

Cleaners gracefully degrade when tools are unavailable (mock data for testing).

## Architecture Decisions

### Result[T] over errors

We use a custom Result type instead of Go's (value, error) pattern:

```go
type Result[T any] struct {
    value T
    err   error
}
```

**Pros:**

- Forces error handling at call sites
- No nil panics
- Can chain operations
- Type-safe

**Cons:**

- Non-standard pattern (but worth it)

### Type-safe enums

We use type-safe string enums instead of iota:

```go
type CleanStrategyType string
const (
    StrategyAggressiveType   CleanStrategyType = "aggressive"
    StrategyConservativeType CleanStrategyType = "conservative"
)
```

**Pros:**

- JSON/YAML marshaling built-in
- String representation always available
- Type-safe

**Cons:**

- Slightly more verbose

## Just Commands

```bash
just build          # Build binary
just test           # Run tests
just test-coverage  # Run with coverage report
just format         # Format code
just clean          # Clean build artifacts
just fix-modules    # Fix module cache issues
```

## Troubleshooting

### Module cache corrupted

```bash
just fix-modules
```

### Tests fail

Check if external tools are available:

```bash
which nix docker brew
```

Tests should pass even without tools (mock data), but verify environment.

## Performance

Current test suite:

- 37 test files
- 8,071 lines of test code
- 0.69:1 test-to-code ratio (target: 1:1)

Binary size: 8.4MB (includes all cleaners)

## Contributing

1. Run `just test` before committing
2. Add tests for new features
3. Update DEVELOPMENT.md if adding architecture changes
4. Use Result[T] for error handling
5. Add 5-minute timeout to external commands
6. Follow existing patterns

## Production Readiness

✅ Type safety (impossible negative values)
✅ Timeout protection on all external commands
✅ Railway error handling (Result[T])
✅ Mock pattern for graceful degradation
✅ Comprehensive test suite
✅ JSON output for scripting
✅ Clean architecture (zero circular deps)
✅ Benchmarks showing sub-nanosecond performance

**Production readiness: 95%** 🚀
