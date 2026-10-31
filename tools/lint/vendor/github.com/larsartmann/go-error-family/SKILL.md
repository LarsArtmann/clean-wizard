# go-error-family — AI Agent Context

**Module:** `github.com/larsartmann/go-error-family`
**Go:** 1.26+ (uses `errors.AsType`) | **Third-party deps:** zero | **Kind:** library (no `main`, no build system)

---

## What This Library Is

A structured error protocol for Go. Every error gets a behavioral **Family** (retry/no-retry), a machine-readable **code**, human-readable context, and optional diagnostics/agent analysis. Designed for the CLI/HTTP boundary — the place where errors leave your program and meet a human or a downstream system.

## Architecture at a Glance

```
errorfamily/          ← root module: types, constructors, classification, CLI boundary
  error.go              Error struct (reference implementation)
  family.go             Family enum + Audience/Tone metadata
  interfaces.go         Coded, Classified, Contextual, Retryable, ExitCoder, HTTPStatuser
  constructors.go       New/Wrap + family shortcuts (incl. Wrap{Family}f formatted variants)
  classify.go           Classify(), Code(), IsRetryable, ExitCode, HTTPStatus, Classifier, RegisterClassification(s), RegisterClassifier(s), RegisterClassificationType[T]
  registry.go           Registry type (injectable sentinels + classifiers + templates), DefaultRegistry, NewRegistry(), TemplateForCode()
  handle.go             HandleError(), HandleErrorWithContext(), HandleErrorDetailed(), template system
  http.go               HTTPStatus(), HTTPHandler() — classify→status-code net/http middleware
  log.go                LogError(), LogErrorContext() — structured slog logging
  errorfamilytest/      test assertion helpers (AssertFamily, AssertCode, ...) — mirrors httptest

diagnose/             ← OWN MODULE (v0.x experimental): concurrent diagnostic rules
  go.mod                independent module (depends on root)
  diagnose.go           Runner, DiagnosticRule, RuleSpec, CommandRunner, ContextKey, ErrorContext
  command.go           RunCommand, CommandExists, DefaultCommandRunner
  rules_filesystem.go   FilesystemRule
  rules_network.go      NetworkRule

diagnose/git/         ← submodule: GitRule (opt-in)
  rules_git.go          GitRule

diagnose/postgres/    ← submodule: PostgresRule (opt-in)
  rules_postgres.go     PostgresRule, IsPostgresRunning

agent/                ← OWN MODULE (v0.x experimental): analysis-only debug agent
  go.mod                independent module (depends on root + diagnose)
  agent.go              DebugAgent interface, deterministic analyzer

bridge/               ← submodule: samber/oops integration (opt-in, depends on both libraries)
  bridge.go             ClassifiedError (satisfies Classified, Coded, Retryable, Contextual), Wrap
  classify.go           InferFamily, AutoWrap (tag/domain → Family mapping)

examples/             ← OWN MODULE: runnable examples (depends on root + diagnose + bridge)
  checkout/             LIBRARY layer: imports only errorfamily (proves "libraries classify")
  cmd/bridge/           REFERENCE IMPL: oops + bridge + error-family (3 patterns, HTTP server)
  cmd/cli/              CLI boundary handler
  cmd/http/             HTTP middleware
  cmd/custom_rule/      custom DiagnosticRule
```

---

## The Six Families

| Family           | Retry?  | Exit | Whose fault | Audience | Tone          | When                                                        |
| ---------------- | ------- | ---- | ----------- | -------- | ------------- | ----------------------------------------------------------- |
| `Rejection`      | No      | 1    | User        | User     | Instructional | Bad input, unauthorized, not found                          |
| `Conflict`       | No      | 1    | User        | User     | Explanatory   | Version mismatch, duplicate, state clash                    |
| `Transient`      | **Yes** | 75   | System      | All      | Reassuring    | Temporary infra failure (the only retryable one)            |
| `Corruption`     | No      | 65   | System      | Ops      | Urgent        | Source of truth damaged, unparseable data                   |
| `Infrastructure` | No      | 69   | System      | Ops      | Apologetic    | System cannot serve, nil deps, startup fail                 |
| `Orchestration`  | No      | 70   | Program     | Ops      | Apologetic    | Internal coordination failure — the program's own logic bug |

Only `Transient` is retryable. Everything else is not. This is the core design decision.

### Family Methods

```go
// Classification
func (f Family) IsRetryable() bool      // true only for Transient
func (f Family) IsValid() bool          // true if within defined range
func (f Family) String() string         // "rejection", "transient", etc.

// Ordering & multi-error (powers errors.Join worst-severity selection)
func (f Family) Severity() int          // total order: Transient(1) < Rejection(2) < Conflict(3) < Infrastructure(4) < Orchestration(5) < Corruption(6)

// Process & network boundaries
func (f Family) ExitCode() int          // BSD sysexits.h code (see table above)
func (f Family) HTTPStatus() int        // canonical family→HTTP (Rejection→400, Conflict→409, Transient→503, Infrastructure→503, Orchestration→500, Corruption→500)
func (f Family) RetryPolicy() RetryPolicy  // advisory: Transient→{3 attempts, 100ms–5s backoff}; others→{1 attempt}

// Presentation metadata
func (f Family) Audience() Audience     // who to notify: User, Ops, or All
func (f Family) Tone() Tone             // presentation tone hint
func (f Family) DefaultMessage() string // generic human-readable message
func (f Family) DefaultWhy() string     // generic "why" explanation
func (f Family) DefaultFix() string     // generic fix suggestion

// Config / serialization (implements encoding.TextMarshaler/TextUnmarshaler)
func (f Family) MarshalText() ([]byte, error)   // YAML/JSON config: "transient"
func (f *Family) UnmarshalText(text []byte) error // case-insensitive parse, defaults to Transient
```

### Audience & Tone Types

```go
type Audience int // AudienceUser, AudienceOps, AudienceAll
type Tone string  // "instructional", "explanatory", "reassuring", "urgent", "apologetic"
```

Audience mapping: Rejection/Conflict → User, Corruption/Infrastructure/Orchestration → Ops, Transient → All.

---

## Consumer Interfaces

All embed `error` (required for Go 1.26 `errors.AsType[T]()` — do not remove):

```go
type Coded interface {       // machine-readable identity (e.g. "db.timeout")
    error
    ErrorCode() string
}

type Classified interface {   // behavioral classification
    error
    ErrorFamily() Family
}

type Contextual interface {   // factual key-value details
    error
    ErrorContext() map[string]string
}

type Retryable interface {    // retry hint (consulted only when Classified is absent: true→Transient, false→Rejection)
    error
    IsRetryable() bool
}

type ExitCoder interface {    // per-error exit code override (0 = use family default)
    error
    ExitCode() int
}

type HTTPStatuser interface { // per-error HTTP status override (0 = use family default)
    error
    HTTPStatus() int
}
```

`*Error` implements all six. Third-party error types can implement whichever subset makes sense.

---

## Quick API Reference

### Error Struct Methods

```go
// Accessors (beyond the interface methods)
func (e *Error) ErrorCode() string                  // from Coded
func (e *Error) ErrorFamily() Family                // from Classified
func (e *Error) ErrorContext() map[string]string    // from Contextual (returns a copy)
func (e *Error) IsRetryable() bool                  // from Retryable

// Direct accessors (no interface assertion needed)
func (e *Error) Code() string                       // same as ErrorCode()
func (e *Error) Family() Family                     // same as ErrorFamily()
func (e *Error) Message() string                    // human-readable technical message
func (e *Error) Cause() error                       // underlying error in the chain
func (e *Error) Timestamp() time.Time               // when the error was created

// Mutators (chainable — all copy-on-write, return a NEW *Error)
func (e *Error) WithContext(key, value string) *Error
func (e *Error) WithContextMap(ctx map[string]string) *Error    // bulk set from a map
func (e *Error) WithContextf(key, format string, args ...any) *Error  // printf-style context value
func (e *Error) WithContextAny(key string, value any) *Error   // type-safe: string, int, int64, uint, uint64, float64, bool, []byte, time.Time, error, nil
func (e *Error) WithCause(cause error) *Error
func (e *Error) WithTimestamp(ts time.Time) *Error   // deterministic timestamp for tests
func (e *Error) WithExitCode(code int) *Error        // override family exit code (0 = use default)
func (e *Error) WithHTTPStatus(status int) *Error    // override family HTTP status (0 = use default)

// Serialization
func (e *Error) JSON() ([]byte, error)              // canonical JSON for API boundaries: {family,code,message,context,retryable,timestamp}

// Helpers
func (e *Error) HasContext(key string) bool
func (e *Error) ContextValue(key string) string
func (e *Error) Summary() string                    // "code: message" (no family prefix)
```

Formatting (implements `fmt.Formatter`):

```go
func formatting(err *Error) {
    fmt.Sprintf("%v", err)    // [family:code] message[: cause]
    fmt.Sprintf("%+v", err)   // verbose: context, timestamp, cause chain
    fmt.Sprintf("%s", err)    // message only
}
```

### Creating Errors

```go
// Direct constructors
err := errorfamily.NewTransient("db.timeout", "query took too long")
err := errorfamily.WrapRejection(originalErr, "validation", "invalid email format")

// With context (chainable)
err := errorfamily.NewTransient("db.connection", "could not connect").
    WithContext("host", "localhost").
    WithContext("port", "5432")

// Generic constructors
errorfamily.New(family, code, message) *Error
errorfamily.Newf(family, code, format, args...) *Error
errorfamily.Wrap(err, family, code, message) *Error    // nil-safe: returns nil if err is nil
errorfamily.Wrapf(err, family, code, format, args...) *Error
errorfamily.WrapOnce(err, family, code, message) *Error  // idempotent: returns existing *Error if already classified
errorfamily.WrapOncef(err, family, code, format, args...) *Error

// Family shortcuts (New + Wrap + formatted Wrap for each)
NewRejection / NewConflict / NewTransient / NewCorruption / NewInfrastructure / NewOrchestration
WrapRejection / WrapConflict / WrapTransient / WrapCorruption / WrapInfrastructure / WrapOrchestration
WrapRejectionf / WrapConflictf / WrapTransientf / WrapCorruptionf / WrapInfrastructuref / WrapOrchestrationf  // printf-style
```

**When to use `New*` vs `Wrap*`:**

- `New*` — the error originates here (no underlying cause). Use for validation failures, domain rule violations, sentinel errors.
- `Wrap*` — the error has a cause you're classifying for the caller. `Wrap(nil, ...)` returns `nil` (nil-safe). Use when translating a third-party error into a behavioral family.

### Domain Error Helpers (the "errkit" pattern)

Every non-trivial consumer builds a thin domain layer over the constructors. This gives reusable, typed error factories with consistent codes:

```go
// internal/errors/users.go
package errors

import "github.com/larsartmann/go-error-family"

func UserNotFound(userID string) error {
    return errorfamily.NewRejection("user.not_found", "user not found").
        WithContext("user_id", userID)
}

func DBTimeout(cause error) error {
    if cause == nil { return nil }  // nil-safe (Wrap* already is, but explicit for domain helpers)
    return errorfamily.WrapTransient(cause, "db.timeout", "database query timed out")
}

func OrderConflict(orderID string, cause error) error {
    return errorfamily.WrapConflict(cause, "order.duplicate", "order already exists").
        WithContext("order_id", orderID)
}
```

Callers import your domain errors, not raw constructors. Codes and families live in one place.

### Classification

```go
family := errorfamily.Classify(err)     // always returns a Family (never panics)
code := errorfamily.Code(err)           // extract code from any error in the chain ("" if none)
retryable := errorfamily.IsRetryable(err)
exitCode := errorfamily.ExitCode(err)
httpStatus := errorfamily.HTTPStatus(err)  // classify → status code

family := errorfamily.ParseFamily("transient")  // parse from string (case-insensitive, defaults to Transient if unrecognized)

// Register third-party sentinels (call from init()) — for stable error VALUES
errorfamily.RegisterClassification(sql.ErrConnDone, errorfamily.Transient)
errorfamily.RegisterClassifications(map[error]errorfamily.Family{...})

// Register a classifier (call from init()) — for DYNAMIC errors (new instance each time)
errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
    var sq *sqlite.Error
    if errors.As(err, &sq) {
        switch sq.Code() {
        case 5, 6: return errorfamily.Transient, true // BUSY, LOCKED
        case 19:   return errorfamily.Conflict, true  // CONSTRAINT
        }
    }
    return errorfamily.Transient, false
})

// Register a type-based classifier (sugar over RegisterClassifier for "type T → Family F")
errorfamily.RegisterClassificationType[*sqlite.Error](errorfamily.Transient)
// Or on a custom registry:
errorfamily.RegisterClassificationTypeFor[*sqlite.Error](reg, errorfamily.Transient)

// Register stdlib taxonomy (context/sql/os errors with documented rationale)
errorfamily.RegisterStdlibDefaults(errorfamily.DefaultRegistry)
// Maps: context.DeadlineExceeded→Transient, context.Canceled→Rejection,
// sql.ErrNoRows→Rejection, sql.ErrConnDone→Transient, os.ErrNotExist→Rejection, etc.

// Combine errors for partial-success patterns — use stdlib errors.Join
combined := errors.Join(err1, err2)  // Classify picks the worst Family automatically
```

**Classification precedence** (first match wins):

1. **Multi-error** (`errors.Join`) → classify each sub-error, worst severity wins
2. `Classified` interface → `ErrorFamily()`
3. `Retryable` interface → infer `Transient` (true) or `Rejection` (false)
4. Registered sentinels via `errors.Is` chain walk (lock-free, atomic.Pointer)
5. Registered classifiers (`RegisterClassifier`) — predicate funcs for dynamic errors
6. Default → `Transient` (fail-open)

### Injectable Registry (test isolation, scoped handling)

Package-level functions (`Classify`, `RegisterClassification`, `RegisterTemplate`) delegate to `DefaultRegistry`. For test isolation or scoped error handling within a binary, construct a custom registry:

```go
reg := errorfamily.NewRegistry()
reg.RegisterClassification(sql.ErrConnDone, errorfamily.Transient)
reg.RegisterTemplate("custom.code", errorfamily.MessageTemplate{What: "Custom"})
reg.RegisterTemplates(map[string]errorfamily.MessageTemplate{  // batch variant
    "db.timeout":  {What: "DB timed out on {host}"},
    "auth.failed": {What: "Invalid credentials"},
})
errorfamily.RegisterStdlibDefaults(reg)  // context/sql/os taxonomy onto this registry

// Clone — deep-copy with inherit-and-extend semantics
child := reg.Clone()
child.RegisterClassification(myErr, errorfamily.Corruption)

// Pass via HandleConfig.Registry
code := errorfamily.HandleErrorWithConfig(err, errorfamily.HandleConfig{
    Registry: reg,
})

// Or classify directly
family := reg.Classify(err)
```

No `t.Cleanup(Unregister...)` needed — the registry is local, no global state mutated.

### CLI Boundary (main.go pattern)

```go
func main() {
    if err := run(); err != nil {
        os.Exit(errorfamily.HandleError(err))  // classify → format → stderr → exit code
    }
}
```

**Canonical entry point when you have a `context.Context`:**

```go
exitCode := errorfamily.HandleErrorWithContext(ctx, err, errorfamily.HandleConfig{})
```

`HandleError` and `HandleErrorWithConfig` both delegate to `HandleErrorWithContext`.

### Structured Result (HTTP/gRPC)

```go
result := errorfamily.HandleErrorDetailed(err)
// result.ExitCode, result.Message, result.SuggestedFix

// Template-aware structured result
result := errorfamily.HandleErrorDetailedWithConfig(err, cfg)
```

### HTTP Middleware (net/http)

```go
// Thin helper: classify → status code
w.WriteHeader(errorfamily.HTTPStatus(err))  // Rejection→400, Conflict→409, Transient→503, ...

// Ready-made middleware: wrap an error-returning handler, write safe JSON
mux.Handle("/api/orders", errorfamily.HTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
    if err := createOrder(r); err != nil {
        return errorfamily.WrapConflict(err, "order.duplicate", "already exists")
    }
    return nil
}))
```

`HTTPHandler` writes `{"family","code","message"}` where `message` comes only from a
registered `MessageTemplate` — it NEVER includes the raw `err.Error()` (no internal leak).

### Conditional Requests (304, 412, 428, 416)

Every family maps to a 4xx/5xx, so RFC 9110 §13 conditional-request outcomes need explicit handling:

| Outcome                                 | Recipe                                                                                                                                                                                                                                                             |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 304 Not Modified                        | NOT an error. Write the 304 (plus `ETag`) yourself, return `nil` (`HTTPHandler` treats nil as fully handled). Never classify it.                                                                                                                                   |
| 412 Precondition Failed                 | `NewConflict(...).WithHTTPStatus(http.StatusPreconditionFailed)` — client-asserted state (`If-Match`) no longer holds = version mismatch = Conflict semantics.                                                                                                     |
| 428 Precondition Required (RFC 6585 §3) | `NewRejection(...).WithHTTPStatus(http.StatusPreconditionRequired)` — request omitted a required precondition = incomplete input = Rejection semantics (NOT Conflict: nothing contradicts current state; the fix is "fix the request", not "refresh and reapply"). |
| 416 Range Not Satisfiable               | `NewRejection(...).WithHTTPStatus(http.StatusRequestedRangeNotSatisfiable)` — unsatisfiable `Range` = bad input value.                                                                                                                                             |

```go
// RFC 9110 §13.2.2: If-Match evaluates before If-None-Match
if im := r.Header.Get("If-Match"); im != "" && !etagMatches(im, currentETag) {
    return errorfamily.NewConflict("etag.precondition_failed", "If-Match precondition failed").
        WithHTTPStatus(http.StatusPreconditionFailed) // 412, not family default 409
}
if etagMatches(r.Header.Get("If-None-Match"), currentETag) {
    w.Header().Set("ETag", currentETag)
    w.WriteHeader(http.StatusNotModified) // success outcome — no error, no body
    return nil
}
```

`WithHTTPStatus` is the intended behavioral/wire-level seam (same pattern as 404-over-Rejection).
For error types NOT built with this package, implement the `HTTPStatuser` interface on your own type.

**TRAP:** do NOT build a wrapper that embeds `*errorfamily.Error` to add `HTTPStatus()` — it does not compile.
The embedded field named `Error` shadows the promoted `Error() string` method (wrapper fails `error` satisfaction),
and declaring your own `Error() string` method collides with the embedded field name ("field and method with the
same name Error"). Use `.WithHTTPStatus(...)` on the `*Error`, or a named field + explicit `Error`/`Unwrap` forwarding.

### Look Up a Template Without the CLI Pipeline

```go
tmpl, ok := errorfamily.TemplateForCode("db.timeout")  // registry → built-in defaults
// tmpl.What / Why / Fix / WayOut
```

### Structured Logging (log/slog)

```go
errorfamily.LogError(err, slog.Default())
// Transient → WARN; all others → ERROR
// attrs: family, code, retryable, context.<key>...
errorfamily.LogErrorContext(ctx, err, logger)  // propagates context
```

### Configurable Handler

```go
exitCode := errorfamily.HandleErrorWithConfig(err, errorfamily.HandleConfig{
    Output: os.Stderr,
    DiagnosticFunc: func(ctx context.Context, err error) []errorfamily.DiagnosticFinding {
        return nil // adapt diagnose.Runner results here
    },
    TemplateOverride: map[string]errorfamily.MessageTemplate{
        "db.timeout": {What: "DB timed out on {host}", Fix: "Check {host}"},
    },
    OnDiagnosed: func(err error, findings []errorfamily.DiagnosticFinding) {
        for _, f := range findings {
            slog.Warn("diagnostic", "rule", f.RuleName, "summary", f.Summary)
        }
    },
})
```

### Diagnostics

```go
// One-shot with zero-dep built-in rules (Filesystem, Network)
results := diagnose.RunAuto(ctx, err)

// Custom runner with opt-in submodules
import (
    "github.com/larsartmann/go-error-family/diagnose/git"
    "github.com/larsartmann/go-error-family/diagnose/postgres"
)
runner := diagnose.NewRunner(&git.GitRule{}, &postgres.PostgresRule{}, &myCustomRule{})
results := runner.Run(ctx, err)
// results sorted by confidence desc; nil if no rules applicable
```

**DiagnosticResult fields:** `RuleName`, `Status` (Healthy/Degraded/Failed/Unknown), `Summary`, `Details` (map[string]string), `Fix` (struct with `Summary` and `Command`), `Confidence` (0.0–1.0), `Duration`, `Context` (the error context that triggered the rule, `map[string]string`).

**Standalone helpers (postgres submodule):**

```go
func IsPostgresRunning(ctx context.Context, host, port string) bool  // pg_isready or TCP check
```

### Partial Success (Recipe)

This library does not provide batch/multi-error types — partial success is a **consumption pattern**, not a classification concern. Use the existing primitives to compose what your domain needs:

```go
// 1. Process items, collect per-item results.
type outcome struct {
    value Item
    err   error
}
var results []outcome
for _, item := range items {
    v, err := process(item)
    results = append(results, outcome{value: v, err: err})
}

// 2. Separate successes from failures.
var successes []Item
var failures []outcome
for _, r := range results {
    if r.err == nil {
        successes = append(successes, r.value)
    } else {
        failures = append(failures, r)
    }
}

// 3. Use Classify to decide what to do with each failure.
for _, f := range failures {
    switch errorfamily.Classify(f.err) {
    case errorfamily.Transient:
        // retry (backoff, jitter, idempotency are the consumer's responsibility)
    case errorfamily.Rejection:
        // skip, log, or surface to user
    case errorfamily.Corruption:
        // escalate to ops
    }
}

// 4. If you need a single exit code, pick the one with the highest ExitCode().
// Exit codes map: Transient(75) > Infrastructure(69) > Corruption(65) > Conflict(1) = Rejection(1).
// Note: exit codes ≠ severity. Transient is retryable (not "worst"); Corruption is severe but low exit code.
worst := errorfamily.Transient
for _, f := range failures {
    if errorfamily.Classify(f.err).ExitCode() > worst.ExitCode() {
        worst = errorfamily.Classify(f.err)
    }
}
```

Why not a built-in `ErrorBatch` type? Because batch semantics vary by domain — some consumers want fail-fast, some want collect-all, some want per-item retry with circuit breakers. The library provides the **classification vocabulary**; you compose the **collection strategy**.

### Agent (Analysis-Only)

```go
ag := agent.New(agent.Config{Enabled: true})
result, err := ag.Analyze(ctx, err, diagnosis)
// result.RootCause, result.Confidence, result.Explanation, result.FixSteps
// FixSteps have Description, Command, Rationale — consumer decides whether to execute
```

---

## Surprising Behaviors (Gotchas)

| Behavior                                                    | Why                                                                                                                                                                 |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Classify(nil)` returns `Rejection`                         | nil error = caller's fault                                                                                                                                          |
| `Classify` defaults unknown → `Transient`                   | Fail-open: unknown errors get retried                                                                                                                               |
| `ParseFamily("unknown")` → `Transient`                      | Same fail-open design                                                                                                                                               |
| `errors.Is` matches on **code + family** only               | Two `*Error`s with different messages but same code+family match                                                                                                    |
| `Wrap(nil, ...)` returns `nil`                              | Nil-safe, but can't construct error wrapping nil                                                                                                                    |
| `WithContext`/`WithCause`/`WithTimestamp` are copy-on-write | They return a NEW `*Error`, not the same pointer — safe to chain from shared sentinels                                                                              |
| `Error.ErrorContext()` returns a **copy**                   | Mutations won't affect the original                                                                                                                                 |
| Template `{key}` uses `strings.ReplaceAll`                  | Not html/template — just simple substitution; NOT HTML-escaped (unsafe for HTML rendering)                                                                          |
| `DiagnosticFunc` is a function type, not interface          | Avoids circular import between root and diagnose packages                                                                                                           |
| `diagnose/` and `agent/` are separate modules               | Opt-in: skip them unless you need infrastructure debugging or AI analysis                                                                                           |
| `HTTPStatus(nil)` returns `400`                             | Nil reaching the HTTP layer classifies as Rejection — check nil first                                                                                               |
| 304 Not Modified must never be classified                   | Every family implies 4xx/5xx + retry/exit semantics a 304 doesn't have — write it, return nil                                                                       |
| Embedding `*Error` in a custom wrapper struct won't compile | Embedded field `Error` shadows the promoted `Error()` method; own `Error()` method collides with the field name. Use `WithHTTPStatus` or a named field + forwarding |

---

## Template Resolution Order

1. `HandleConfig.TemplateOverride[code]` — per-call override
2. `Registry.lookupTemplate(code)` — registry templates via `RegisterTemplate()` (uses `DefaultRegistry` unless `HandleConfig.Registry` is set)
3. `defaultMessages[code]` — built-in exact-match (see handle.go)
4. `family.DefaultMessage()` — generic fallback

`SuggestedFix` resolution follows the same chain via `resolveSuggestedFix()`.

All lookups are exact code match (case-insensitive). No substring matching.

Built-in codes: `file.not_found`, `permission.denied`, `db.timeout`, `db.connection`, `db.error`, `config.invalid`, `config.not_found`, `conflict`, `validation`, `timeout`, `connection.refused`, `git.error`.

---

## Diagnostic Rule Pattern

### Adding a New Rule

1. Implement `diagnose.DiagnosticRule` (3 methods: `Name`, `Applicable`, `Run`)
2. Use `diagnose.RuleSpec` for matching — define it as a package-level `var`:

```go
var mySpec = diagnose.RuleSpec{
    ContextKeys:   []diagnose.ContextKey{diagnose.KeyHost}, // typed string constants
    CodeContains:  []string{"my."},                          // matches if error code contains substring
    ContextSubstr: []string{"my_thing"},                     // matches if any context value contains substring
    Extra:         func(err error) bool { ... },             // custom logic
}

func (r *MyRule) Applicable(err error) bool { return mySpec.Matches(err) }
```

3. Use matching helpers from `diagnose` package: `HasContextKey`, `ContextValue`, `ResolveContextKey`, `HasContextSubstring`, `FamilyIs`, `ErrorCodeContains`
4. Use execution helpers: `diagnose.RunCommand`, `diagnose.CommandExists`
5. Extract context from any error: `diagnose.ErrorContext(err)` → `map[string]string`
6. Rules run concurrently via `Runner.Run`; results sorted by confidence descending

### Testability: CommandRunner

Rules that shell out should accept a `CommandRunner` for mock injection:

```go
type MyRule struct {
    Runner diagnose.CommandRunner // nil → DefaultCommandRunner{}
}
```

### Built-in Rules

| Rule             | Module                                                     | Matches On                                                                                                                                                                | Checks                        |
| ---------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| `FilesystemRule` | `github.com/larsartmann/go-error-family/diagnose`          | Keys: `path`, `file`, `dir`, `directory`, `config_path`, `output_path`. Codes: `file`, `dir`, `path`, `config`, `permission`                                              | Existence, permissions, write |
| `NetworkRule`    | `github.com/larsartmann/go-error-family/diagnose`          | Keys: `host`, `port`, `url`, `endpoint`, `address`, `remote`. Codes: `network`, `connect`, `dial`, `timeout`. Substr: `connection refused`, `no such host`, `i/o timeout` | DNS, TCP, port reachability   |
| `GitRule`        | `github.com/larsartmann/go-error-family/diagnose/git`      | Keys: `git`, `repository`, `repo`, `branch`, `git_dir`. Codes: `git`. Substr: `git`                                                                                       | Repo, tree, merge, remote     |
| `PostgresRule`   | `github.com/larsartmann/go-error-family/diagnose/postgres` | Keys: `db_host`, `db_port`, `db_name`, `database_url`, `postgres_host`. Codes: `db.`, `database`. Substr: `postgres`, `postgresql`, `database`, `sql` + Transient family  | pg_isready, TCP, start cmd    |

---

## Bridge Submodule (samber/oops integration)

`bridge/` is a separate Go module that connects go-error-family with samber/oops. It has its own `go.mod` with both libraries as dependencies. The root package remains zero-dependency.

### ClassifiedError

```go
// Wraps any error with a behavioral Family + oops context
type ClassifiedError struct {
    oops.OopsError           // preserves all oops methods (Stacktrace, Sources, etc.)
    // satisfies: Classified, Coded, Retryable, Contextual, fmt.Formatter
}

// Manual family assignment
classified := bridge.Wrap(err, errorfamily.Transient)

// Automatic inference from oops metadata
classified := bridge.AutoWrap(err)          // tags -> domain -> Transient (fail-open)
family := bridge.InferFamily(err)           // just the Family, no wrapping
```

### InferFamily cascade

1. **Tags** (developer-intentional) — `retryable`, `transient`, `conflict`, `corruption`/`corrupted`, `rejection`/`rejected`, `infrastructure`/`infra`
2. **Domain** (structural) — `validation`/`auth` -> Rejection, `database`/`network`/`cache`/`queue` -> Transient, `storage`/`infra`/`startup` -> Infrastructure, `data`/`schema`/`migration` -> Corruption
3. **Default** — `Transient` (fail-open, consistent with root Classify)

### What ClassifiedError bridges

| oops method  | error-family interface                          | Notes                                      |
| ------------ | ----------------------------------------------- | ------------------------------------------ |
| `.Code()`    | `ErrorCode() string` (Coded)                    | Converts `any` to string via fmt.Sprint    |
| `.Context()` | `ErrorContext() map[string]string` (Contextual) | Non-strings converted via fmt.Sprint       |
| `.Domain()`  | Included in `ErrorContext()["domain"]`          |                                            |
| `.Tags()`    | Included in `ErrorContext()["tags"]`            |                                            |
| —            | `ErrorFamily() Family` (Classified)             | From the attached Family                   |
| —            | `IsRetryable() bool` (Retryable)                | Derived from Family                        |
| `.Is()`      | `Is(target error) bool`                         | Delegates to OopsError.Is + original error |
| `Format()`   | `fmt.Formatter`                                 | `%+v` shows oops stacktrace when present   |

### Original error preservation

`Wrap(err, family)` always preserves the original error in the `Unwrap()` chain, even when `err` is not an OopsError. `errors.Is(classified, originalErr)` always works.

### Import

```go
import "github.com/larsartmann/go-error-family/bridge"
```

### Reference Implementation

`examples/cmd/bridge/` is the canonical reference implementation demonstrating the full
classify→enrich→handle flow. It shows three patterns:

1. **Pass-through** — library errors (classified with error-family) flow through oops
   enrichment unchanged. `Classify()` finds the family through the chain. No bridge needed.
2. **AutoWrap** — application-created errors built with oops, classified by `bridge.AutoWrap`.
3. **Explicit Wrap** — `bridge.Wrap` when the application knows the exact family.

The library layer (`examples/checkout/`) imports ONLY `go-error-family`, proving the
"libraries classify, applications enrich" architecture. See `examples/cmd/bridge/README.md`
for the full pattern documentation and decision guide.

---

## Testing

```bash
go test ./...                                    # all tests
go test -cover ./...                             # with coverage
go test -coverprofile=cover.out ./... && go tool cover -func=cover.out  # detailed coverage
go test -run TestName ./...                      # specific test
go test -bench=. -run=^$ ./...                    # benchmarks only
```

**Test assertion helpers** — import the `errorfamilytest` subpackage (mirrors `net/http/httptest`, keeps `testing` out of production code):

```go
import "github.com/larsartmann/go-error-family/errorfamilytest"

errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
errorfamilytest.AssertCode(t, err, "user.not_found")
errorfamilytest.AssertRetryable(t, err, false)
errorfamilytest.AssertContext(t, err, "user_id", "42")
errorfamilytest.AssertContextMissing(t, sentinelErr, "field")
errorfamilytest.AssertExitCode(t, err, 1)              // checks ExitCoder override first
errorfamilytest.AssertHTTPStatus(t, err, 404)          // checks HTTPStatuser override first
```

Test files by area (run `find . -name '*_test.go'` for the canonical list):

- **Root:** `error_test.go`, `family_test.go`, `classify_test.go`, `registry_test.go`, `handle_test.go`, `handle_context_test.go`, `template_test.go`, `http_test.go`, `log_test.go`, `retry_test.go`, `stdlib_test.go`, `example_test.go`, `benchmark_test.go`, `fuzz_test.go`
- **errorfamilytest:** `errorfamilytest_test.go` — tests all assert helpers (happy + failure paths)
- **agent:** `agent_test.go`
- **bridge:** `wrap_test.go`, `autowrap_test.go`, `infer_test.go`, `fuzz_test.go`
- **diagnose:** `helpers_test.go`, `rules_test.go`, `rules_integration_test.go`, `rules_network_test.go`, `runner_test.go`, `mock_test.go`, `benchmark_test.go`
- **diagnose/git:** `scenario_test.go`, `mock_test.go`, `integration_test.go`
- **diagnose/postgres:** `rules_postgres_test.go`

**Coverage:** root 97.0% | errorfamilytest 96.3% | agent 100% | bridge 94.1% | diagnose 83.9% | git 98.5% | postgres 80.3%
(rules that shell out to system commands are tested via `CommandRunner` mocks in git/postgres; diagnose core coverage reflects shell-out rules tested via integration)

### Test Style

Standard `testing.T` table-driven tests. No external test frameworks. Same-package tests (no `_test` suffix on package name — tests access internals).

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name string
        // ...
        want string
    }{
        {name: "basic", want: "expected"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := thing(); got != tt.want {
                t.Errorf("thing() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

---

## Code Conventions

- **Zero third-party dependencies** — stdlib only (`encoding/json`)
- **Interfaces embed `error`** — for `errors.AsType[T]()` compatibility
- **Data-driven patterns** — `familyData` array, `defaultMessages` map, `ruleSpec` structs
- **Thread-safe registries** — `Registry.sentinels` and `Registry.classifiers` are `atomic.Pointer` to immutable snapshots: reads (the `Classify` hot path) load the pointer once and iterate lock-free/allocation-free; rare writers serialize under `r.mu` (write lock) and publish a new snapshot via copy-on-write. Templates use `r.mu` directly (RWLock).
- **Nil-safe** — `Wrap(nil, ...)` returns nil; `Classify(nil)` returns `Rejection`
- **`maps.Clone`** for defensive copies in `ErrorContext()`
- **Constructors set `timestamp: time.Now().UTC()`**
- **Context values are always `string`** — not `any`
- **Error codes use dot-notation** — e.g. `db.timeout`, `file.not_found`, `config.invalid`
- **No `main` package, no build system** — this is a library consumers import

---

## Key Files for Common Tasks

| Task                           | File(s)                                                                                  |
| ------------------------------ | ---------------------------------------------------------------------------------------- |
| Add a new Family               | `family.go` — one entry in `familyData` slice (const + metadata + methods auto-derived)  |
| Add a new constructor shortcut | `constructors.go`                                                                        |
| Change classification logic    | `classify.go` (pipeline) + `registry.go` (Registry.Classify)                             |
| Add/modify message templates   | `handle.go` (`defaultMessages`) or `RegisterTemplate()` / `Registry.RegisterTemplates()` |
| Register stdlib error taxonomy | `stdlib.go` — `RegisterStdlibDefaults(reg)`                                              |
| Add a diagnostic rule          | New file in `diagnose/`, implement `DiagnosticRule`, add to `DefaultRunner()`            |
| Change CLI boundary behavior   | `handle.go`                                                                              |
| HTTP error responses           | `http.go`                                                                                |
| Structured error logging       | `log.go`                                                                                 |
| Test assertion helpers         | `errorfamilytest/errorfamilytest.go`                                                     |
| Modify agent analysis          | `agent/agent.go`                                                                         |
| Understand the Error struct    | `error.go`                                                                               |
| Understand consumer interfaces | `interfaces.go`                                                                          |

---

## MessageTemplate (Wix-style)

```go
type MessageTemplate struct {
    What   string  // "Could not connect to {host}"
    Why    string  // "The database is not reachable."
    Fix    string  // "Check that {host} is running."
    WayOut string  // "Run with --verbose for more details."
}
```

`{key}` placeholders are replaced from error context values. Empty fields fall back to family defaults.

---

## Dependency Graph

```
agent → errorfamily (root)
agent → diagnose

bridge → errorfamily (root)
bridge → samber/oops

diagnose → errorfamily (root)

errorfamily → (stdlib only)
errorfamilytest → errorfamily (root)
```

The root package has no dependency on `diagnose`, `agent`, or `bridge`. `DiagnosticFunc` in `handle.go` is a function type to avoid circular imports — the consumer wires `diagnose.Runner` to it. The `bridge/` module is the only one with an external dependency (`samber/oops`); it exists for consumers who already use oops for enrichment and want go-error-family's behavioral classification on top.
