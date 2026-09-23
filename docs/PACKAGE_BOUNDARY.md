# Package Boundary: config vs domain

This document clarifies the intended boundary between `internal/config` and the domain packages.

> **2026-09-23:** `internal/domain` was split into `internal/domain/enums`,
> `internal/domain/operations`, and `internal/domain/types` (dependency DAG:
> enums <- operations <- types). References to `domain.X` below map to these
> sub-packages: value types such as `Config`/`Profile` live in `types`, enums in
> `enums`, and `OperationSettings`/`ValidationError` in `operations`. The
> boundary rules described here are unchanged.

## Overview

| Package  | Responsibility                                   | Imports                         |
| -------- | ------------------------------------------------ | ------------------------------- |
| `domain` | Pure data structures, enums, validation logic    | Standard library only           |
| `config` | Configuration loading, persistence, sanitization | `domain`, `koanf`, `go-error-family`, `go-business-rules` |

> **2026-09-23:** Config validation is expressed as declarative
> `go-business-rules` rules (`internal/config/validation_rules.go`) with
> 4-level severity (`SeverityCritical` added to `operations.ValidationSeverity`,
> still stdlib-only in domain). See
> [planning/2026-09-23_go-business-rules-integration.md](planning/2026-09-23_go-business-rules-integration.md).

## Design Principles

### 1. Domain Package is the Source of Truth

The `domain` package contains:

- **Core types**: `Config`, `Profile`, `OperationSettings`
- **Enums**: `RiskLevelType`, `OperationType`, `SafeMode`
- **Validation**: `ValidationError`, `ValidationContext`, `ValidationSeverity`
- **System constants**: `PathSystem`, `PathApplications`, `PathLibrary`

```go
// domain/operation_validation.go - ValidationError is defined here
type ValidationError struct {
    Field      string             `json:"field"`
    Rule       string             `json:"rule,omitempty"`
    Value      any                `json:"value"`
    Message    string             `json:"message"`
    Severity   ValidationSeverity `json:"severity,omitempty"`
    Suggestion string             `json:"suggestion,omitempty"`
    Context    *ValidationContext `json:"context,omitempty"`
}
```

### 2. Config Package Provides Infrastructure

The `config` package uses **type aliases** to reference domain types:

```go
// config/validator.go - Type alias to domain
type ValidationError = domain.ValidationError
type ValidationContext = domain.ValidationContext
type ValidationSeverity = domain.ValidationSeverity
```

This pattern:

- Avoids code duplication (split brain prevention)
- Maintains clear dependency direction
- Allows config to extend domain types with infrastructure concerns

### 3. Dependency Direction

```
config ──imports──> domain
   │                    │
   │                    │
   └──> pkg/errors <────┘
```

**Important**: Domain should NOT import config. If you find yourself needing this:

- Move shared types to domain
- Create an orchestration layer in config
- Use dependency injection

## Anti-Patterns to Avoid

### ❌ Duplicate Type Definitions

```go
// BAD: Two ValidationError types
domain.ValidationError{Field, Message, Value}
config.ValidationError{Field, Rule, Value, Message, Severity, Suggestion, Context}
```

**Solution**: Define in domain, use type alias in config.

### ❌ Domain Depends on Config

```go
// BAD: Domain should not import config
package domain
import "github.com/LarsArtmann/clean-wizard/internal/config" // ❌
```

**Solution**: Move shared types to domain, or invert the dependency.

### ❌ Hardcoded Strings for System Paths

```go
// BAD: Magic strings scattered throughout code
protected := []string{"/System", "/Library", "/Applications"}
```

**Solution**: Use constants from `domain/system_paths.go`:

```go
// GOOD: Centralized constants
protected := domain.DefaultProtectedPaths()
```

## Current State

### Fixed Split Brains

| Type                  | Location                        | Status   |
| --------------------- | ------------------------------- | -------- |
| `SanitizationWarning` | `domain` only                   | ✅ Fixed |
| `ValidationError`     | `domain` + type alias in config | ✅ Fixed |
| `ValidationSeverity`  | `domain` + type alias in config | ✅ Fixed |
| `ValidationContext`   | `domain` + type alias in config | ✅ Fixed |
| System paths          | `domain/system_paths.go`        | ✅ Fixed |

### Type Aliases in Config Package

```go
// config/validator.go
type ValidationContext = domain.ValidationContext
type ValidationError = domain.ValidationError

// config/validator_rules.go
type ValidationSeverity = domain.ValidationSeverity

const (
    SeverityError   = domain.SeverityError
    SeverityWarning = domain.SeverityWarning
    SeverityInfo    = domain.SeverityInfo
)
```

## Best Practices

### When Adding New Types

1. **Start in domain** if it's a core concept
2. **Use type aliases** in config for infrastructure concerns
3. **Never** create duplicate types with same semantic meaning
4. **Centralize constants** in domain to avoid magic strings

### When to Extend vs Create

| Scenario                              | Action                                         |
| ------------------------------------- | ---------------------------------------------- |
| Adding validation severity levels     | Extend `domain.ValidationSeverity`             |
| Adding config-specific metadata       | New type in `config`                           |
| Adding new enum values                | Add to domain, update config aliases if needed |
| Adding helper methods for persistence | Add to `config`, use domain types              |

## Migration Path

If you encounter split brain types:

1. **Identify** the canonical location (usually domain)
2. **Extend** the domain type if config has extra fields
3. **Replace** config type with type alias
4. **Update** all references to use domain type
5. **Verify** with `go build ./...`
6. **Run tests** to ensure no regressions

## Verification

To verify the boundary is respected:

```bash
# Check that domain doesn't import config
grep -r "internal/config" internal/domain/  # Should be empty

# Check for duplicate type definitions
grep -rn "^type ValidationError" internal/   # Should find only one

# Verify build succeeds
go build ./...
```

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - Overall system architecture
- [docs/status/2026-03-22_03-16_COMPREHENSIVE_ARCHITECTURE_AUDIT.md](status/2026-03-22_03-16_COMPREHENSIVE_ARCHITECTURE_AUDIT.md) - Original audit
