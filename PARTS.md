# Clean Wizard - Reusable Components Analysis

> Analysis of components that could be extracted as standalone libraries/SDKs.
> Created: 2026-02-28

## Executive Summary

| Component             | Recommendation          | Rationale                                         |
| --------------------- | ----------------------- | ------------------------------------------------- |
| Result[T]             | **Keep Internal**       | samber/mo covers this well; no unique value       |
| Type-Safe Enums       | **Consider Extraction** | YAML/JSON helpers + pattern are valuable together |
| Human Duration Parser | **Consider Extraction** | Fills gap in stdlib (days support)                |
| Cleaner Registry      | **Keep Internal**       | Domain-specific, generic registries exist         |
| Format Utilities      | **Keep Internal**       | Thin wrapper around go-humanize                   |
| Config Validation     | **Keep Internal**       | Domain-specific rules                             |

---

## 1. Result[T] Type

**Location:** `internal/result/type.go` (161 lines)

### Description

Railway Oriented Programming error handling via generic Result[T] type. Similar to Rust's `Result<T, E>`.

### Current Implementation

```go
type Result[T any] struct {
    value T
    err   error
}

// Constructors
func Ok[T any](value T) Result[T]
func Err[T any](err error) Result[T]

// Queries
func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool

// Extraction
func (r Result[T]) Unwrap() (T, error)
func (r Result[T]) Value() T           // panics on error
func (r Result[T]) SafeValue() (T, error)
func (r Result[T]) Error() error       // panics on success
func (r Result[T]) SafeError() (error, bool)
func (r Result[T]) UnwrapOr(default_ T) T

// Transformations
func Map[T, U any](r Result[T], fn func(T) U) Result[U]
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U]
func (r Result[T]) OrElse(fallback Result[T]) Result[T]

// Validation
func (r Result[T]) Validate(predicate func(T) bool, errorMsg string) Result[T]
func (r Result[T]) ValidateWithError(predicate func(T) bool, err error) Result[T]

// Side Effects
func (r Result[T]) Tap(fn func(T)) Result[T]
```

### Alternatives

| Library             | Stars | Features                           | Notes                                 |
| ------------------- | ----- | ---------------------------------- | ------------------------------------- |
| **samber/mo**       | ~3k   | Option, Result, Either, Future, IO | Most popular; comprehensive FP types  |
| larsartmann/uniflow | -     | Railway-oriented errors            | Policy-recommended for error handling |
| TeaEntityLab/fpGo   | ~200  | Monad, FP features                 | Less maintained                       |
| markphelps/optional | ~400  | Optional types only                | No Result type                        |

### samber/mo Result Comparison

```go
// samber/mo Result
type Result[T any] struct {
    value T
    err   error
}

// Methods: Ok, Err, IsOk, IsErr, Unwrap, UnwrapOr, UnwrapOrElse
// Map, MapError, FlatMap, Fold, Match, ToOption
```

**Our implementation adds:**

- `SafeValue()` - non-panicking extraction with error
- `SafeError()` - non-panicking error extraction
- `Validate()` / `ValidateWithError()` - predicate-based validation
- `Tap()` - side effects without breaking chain

**samber/mo adds:**

- `MapError()` - transform error type
- `Fold()` - combine both branches
- `Match()` - pattern matching
- `ToOption()` - convert to Option type
- Integration with other monads (Option, Either)

### Unique Value Proposition

**Minimal.** Our implementation is clean but samber/mo is:

- More comprehensive (Option, Either, Future, IO in one package)
- Battle-tested with larger community
- Policy-recommended (`HOW_TO_GOLANG.md` line 191-196)

### Recommendation

**KEEP INTERNAL** — Replace with `samber/mo.Result` if consistency with library policy is desired. Our `Validate` and `Tap` methods can be added as extension functions.

---

## 2. Type-Safe Enum System

**Location:** `internal/domain/type_safe_enums.go` (539 lines), `internal/domain/execution_enums.go` (377 lines), `internal/domain/operation_settings.go` (353 lines)

### Description

Int-based enums with compile-time safety, YAML/JSON serialization, validation, and helper methods.

### Current Implementation

**Pattern (per enum):**

```go
type RiskLevelType int

const (
    RiskLevelLowType RiskLevelType = iota
    RiskLevelMediumType
    RiskLevelHighType
    RiskLevelCriticalType
)

func (rl RiskLevelType) String() string
func (rl RiskLevelType) IsValid() bool
func (rl RiskLevelType) Values() []RiskLevelType
func (rl RiskLevelType) MarshalJSON() ([]byte, error)
func (rl *RiskLevelType) UnmarshalJSON(data []byte) error
func (rl RiskLevelType) MarshalYAML() (any, error)
func (rl *RiskLevelType) UnmarshalYAML(value *yaml.Node) error
```

**Generic Helpers:**

```go
func UnmarshalYAMLEnum[T ~int](value *yaml.Node, target *T, valueMap map[string]T, errorMsg string) error
func UnmarshalYAMLEnumWithDefault[T ~int](...) T
func UnmarshalJSONEnum[T any](data []byte, target *T, valueMap map[string]T, errorMsg string) error
```

**Features:**

- Case-insensitive string parsing
- Integer representation support
- Helpful error messages with valid options
- Documentation references in errors
- TypeSafeEnum interface for generics

**15+ enums defined:**

- RiskLevelType, ValidationLevelType, ChangeOperationType
- CleanStrategyType, SizeEstimateStatusType
- PlatformType, CleanerType, etc.

### Alternatives

| Tool/Library                      | Type     | Features                                     | Notes                     |
| --------------------------------- | -------- | -------------------------------------------- | ------------------------- |
| **go-enum** (abice/go-enum)       | Code Gen | YAML/JSON/sql, String(), Values(), IsValid() | Most popular (~400 stars) |
| **enumer** (dmarkham/enumer)      | Code Gen | String, JSON, YAML, SQL, Text                | Well-maintained           |
| **stringer** (golang.org/x/tools) | Code Gen | String() only                                | Official, minimal         |
| **go-enum-encoding**              | Code Gen | JSON, YAML, BSON                             | Focused on serialization  |

### go-enum Generated Code Comparison

go-enum generates similar boilerplate but typically:

- Requires build step (`go generate`)
- Generates to separate files
- Doesn't include our helpful error messages with docs references
- Case sensitivity is often strict

### Unique Value Proposition

**MODERATE.** Our implementation provides:

1. **Better error messages** - Shows valid strings AND integers with documentation links
2. **Flexible parsing** - Case-insensitive + integer fallback
3. **Zero build step** - No code generation required
4. **YAML + JSON** - Full serialization support out of box
5. **Generic helpers** - `UnmarshalYAMLEnum[T]` reduces boilerplate significantly

However, code generation tools are more popular and generate similar functionality.

### Recommendation

**CONSIDER EXTRACTION** — Extract just the generic helpers (`UnmarshalYAMLEnum`, `UnmarshalJSONEnum`, `UnmarshalYAMLEnumWithDefault`) into a small library. The enum definitions themselves can remain manual or use go-enum.

**Potential library:** `github.com/larsartmann/go-enum-helpers`

```go
// What to extract
package enumhelpers

func UnmarshalYAMLEnum[T ~int](...) error
func UnmarshalYAMLEnumWithDefault[T ~int](...) T
func UnmarshalJSONEnum[T any](...) error
```

---

## 3. Human Duration Parser

**Location:** `internal/domain/duration_parser.go` (76 lines)

### Description

Parses human-readable durations like "7d", "24h", "30m" extending Go's stdlib with day support.

### Current Implementation

```go
func ParseCustomDuration(durationStr string) (time.Duration, error)
func ValidateCustomDuration(durationStr string) error
```

**Supported formats:**

- Standard Go durations: "24h", "30m", "1h30m"
- Days: "7d", "0.5d" (converts to hours)
- Whitespace trimming

**Not supported:**

- Combined: "7d12h" (planned but not implemented)
- Weeks: "1w"
- Months/Years: variable length, complex

### Alternatives

| Library                         | Features                      | Notes                       |
| ------------------------------- | ----------------------------- | --------------------------- |
| **time.ParseDuration** (stdlib) | ns, us, ms, s, m, h           | No days, weeks              |
| **github.com/xeonx/timeago**    | "2d", "1w", natural language  | Focused on "ago" format     |
| **github.com/hako/durafmt**     | Formatting, not parsing       | Converts duration to string |
| **github.com/cenkalti/backoff** | Exponential backoff durations | Different use case          |

**Gap in ecosystem:** No widely-used library for parsing "7d" format.

### Unique Value Proposition

**HIGH.** Go's stdlib `time.ParseDuration` explicitly does NOT support days:

> A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

This is a common pain point for CLI tools that accept duration configuration.

### Recommendation

**EXTRACT** — Create `github.com/larsartmann/go-duration` with:

```go
package duration

import "time"

// Parse extends time.ParseDuration with day ("d") support.
// Formats: "7d", "24h", "30m", "1h30m", "0.5d"
func Parse(s string) (time.Duration, error)

// MustParse is like Parse but panics on error.
func MustParse(s string) time.Duration

// Validate checks if a duration string is valid.
func Validate(s string) error

// Extend to support:
// - "1w" (weeks = 7 days)
// - "7d12h30m" (combined)
```

This fills a real gap in the Go ecosystem.

---

## 4. Cleaner Registry Pattern

**Location:** `internal/cleaner/registry.go` (140 lines), `internal/cleaner/cleaner.go`

### Description

Thread-safe registry for managing cleaner plugins with polymorphic operations.

### Current Implementation

```go
type Registry struct {
    cleaners map[string]Cleaner
    mu       sync.RWMutex
}

func NewRegistry() *Registry
func (r *Registry) Register(name string, c Cleaner)
func (r *Registry) Get(name string) (Cleaner, bool)
func (r *Registry) List() []Cleaner
func (r *Registry) Names() []string
func (r *Registry) Count() int
func (r *Registry) Available(ctx context.Context) []Cleaner
func (r *Registry) CleanAll(ctx context.Context) map[string]result.Result[domain.CleanResult]
func (r *Registry) Unregister(name string)
func (r *Registry) Clear()
```

**Cleaner Interface:**

```go
type Cleaner interface {
    Name() string
    Type() CleanerType
    Clean(ctx context.Context) result.Result[domain.CleanResult]
    IsAvailable(ctx context.Context) bool
    Scan(ctx context.Context) result.Result[domain.ScanResult]
}
```

### Alternatives

| Library                 | Purpose           | Notes                   |
| ----------------------- | ----------------- | ----------------------- |
| **hashicorp/go-plugin** | RPC-based plugins | Overkill for in-process |
| **google/wire**         | Compile-time DI   | Different purpose       |
| **samber/do/v2**        | DI container      | Policy-recommended      |
| Manual sync.Map         | Concurrent maps   | More primitives         |

### Unique Value Proposition

**LOW.** This is a domain-specific registry. Generic patterns:

- `map[string]T` with `sync.RWMutex` is a well-known pattern
- DI frameworks like `samber/do` handle registration differently
- `hashicorp/go-plugin` is for out-of-process plugins

The registry is well-designed but tightly coupled to the Cleaner domain.

### Recommendation

**KEEP INTERNAL** — The registry is domain-specific. If we wanted to extract it, we'd need to:

1. Genericize the interface type
2. Remove domain-specific methods like `CleanAll`, `Available`

The pattern is simple enough that extraction adds little value.

---

## 5. Format Utilities

**Location:** `internal/format/format.go` (61 lines)

### Description

Human-readable formatting for bytes, durations, dates, and numbers.

### Current Implementation

```go
func Size(bytes int64) string      // IEC binary: "1.5 GiB"
func Duration(d time.Duration) string  // Human: "1.5 s"
func Date(t time.Time) string      // "2006-01-02" or "never"
func DateTime(t time.Time) string  // "2006-01-02 15:04:05" or "never"
func Number(n int64) string        // "1,234,567"
```

### Alternatives

| Library                           | Features                     | Notes               |
| --------------------------------- | ---------------------------- | ------------------- |
| **github.com/dustin/go-humanize** | Bytes, time, numbers, commas | What we already use |
| **github.com/labstack/gommon**    | Random utilities             | Broader scope       |

### Current Usage

We're a thin wrapper around `go-humanize`:

```go
func Size(bytes int64) string {
    return humanize.IBytes(uint64(bytes))
}
func Number(n int64) string {
    return humanize.Comma(n)
}
```

### Unique Value Proposition

**NONE.** We're just wrapping `go-humanize` with slightly different API. The "never" handling for zero dates is the only addition.

### Recommendation

**KEEP INTERNAL** — Use `go-humanize` directly in consuming code. The wrapper adds no value.

---

## 6. Configuration Validation Framework

**Location:** `internal/config/validator.go`

### Description

Multi-level configuration validation with structured results.

### Current Implementation

- Structure validation
- Field constraint validation
- Cross-field validation
- Business logic validation
- Security validation
- ValidationResult with errors/warnings/sanitized data

### Alternatives

| Library                                | Features                     | Notes              |
| -------------------------------------- | ---------------------------- | ------------------ |
| **go-playground/validator**            | Struct tags, extensive rules | Most popular       |
| **huma validation**                    | Schema-based                 | Fromhuma framework |
| **github.com/go-ozzo/ozzo-validation** | Rule-based                   | Flexible           |

### Unique Value Proposition

**LOW.** Our validation is domain-specific (cleaner configs). Generic validators cover the patterns better.

### Recommendation

**KEEP INTERNAL** — Domain-specific rules don't generalize well.

---

## Summary Table

| Component              | Lines | Extract? | Library Name      | Priority |
| ---------------------- | ----- | -------- | ----------------- | -------- |
| Result[T]              | 161   | No       | -                 | -        |
| Type-Safe Enum Helpers | ~100  | Maybe    | `go-enum-helpers` | Low      |
| Human Duration Parser  | 76    | **Yes**  | `go-duration`     | **High** |
| Cleaner Registry       | 140   | No       | -                 | -        |
| Format Utilities       | 61    | No       | -                 | -        |
| Config Validation      | -     | No       | -                 | -        |

---

## Recommended Actions

### Immediate: Extract Human Duration Parser

```bash
# Create new repository
gh repo create larsartmann/go-duration --public

# Structure
go-duration/
├── duration.go       # Parse, MustParse, Validate
├── duration_test.go  # Comprehensive tests
├── README.md         # Usage examples
└── go.mod
```

### Future Consideration: Enum Helpers

If we find ourselves defining many enums across projects, extract the generic helpers into `go-enum-helpers`.

### Replace Result[T] with samber/mo

Align with library policy by adopting `samber/mo.Result` and adding extension methods for `Validate` and `Tap`.

---

## References

- Library Policy: `/Users/larsartmann/projects/library-policy/HOW_TO_GOLANG.md`
- samber/mo: https://github.com/samber/mo
- go-enum: https://github.com/abice/go-enum
- go-humanize: https://github.com/dustin/go-humanize
