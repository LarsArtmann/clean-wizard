# go-error-family

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-error-family.svg)](https://pkg.go.dev/github.com/larsartmann/go-error-family)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-error-family)](https://goreportcard.com/report/github.com/larsartmann/go-error-family)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Structured error protocol for Go — behavioral classification, exit codes, and diagnostic rules.

**Share the protocol, not the implementation.**

**[Documentation](https://errorfamily.lars.software)** &middot; **[pkg.go.dev](https://pkg.go.dev/github.com/larsartmann/go-error-family)** &middot; **[Changelog](CHANGELOG.md)**

---

## The Problem

Every Go program has `if err != nil`. But what do you _do_ with that error?

| Question                                | Before                                        | After                                       |
| --------------------------------------- | --------------------------------------------- | ------------------------------------------- |
| Should the caller retry?                | `if strings.Contains(err.Error(), "timeout")` | `errorfamily.IsRetryable(err)`              |
| What exit code?                         | `os.Exit(1)` for everything                   | `errorfamily.ExitCode(err)` — context-aware |
| What message for the user?              | `fmt.Println(err)` — often internal jargon    | Structured What/Why/Fix/WayOut              |
| Is it the user's fault or the system's? | Guess from the string                         | `errorfamily.Classify(err).Audience()`      |

go-error-family answers all of these with a single concept: **Family**.

## Installation

```bash
go get github.com/larsartmann/go-error-family
```

Requires Go 1.26+ (uses `errors.AsType`). Zero third-party dependencies.

This is a Go workspace. The root module (classification core, zero third-party deps) is stable. Experimental submodules — `agent`, `bridge` (samber/oops integration), `diagnose`, `diagnose/git`, `diagnose/postgres` — have their own `go.mod` and require separate imports.

## Complementary, not competing

go-error-family **classifies**; [samber/oops](https://github.com/samber/oops) **enriches** (stack traces, trace IDs). Use both:

- **Libraries** import go-error-family only — they know their domain contract (a 404 is a Rejection, a timeout is Transient) but must not presume the app's observability stack, so they never import oops.
- **Applications** import oops for enrichment and, if they also need behavioral decisions, wrap library errors via the `bridge/` package.

Running the two together: three patterns (pass through, `AutoWrap`, explicit `Wrap`) in the [bridge guide](https://errorfamily.lars.software/guides/bridge/).

The six interfaces (`Coded`, `Classified`, `Contextual`, `Retryable`, `ExitCoder`, `HTTPStatuser`) are the sole public contract; the `Error` struct is a reference implementation, not a requirement.

## Quick Start

```go
package main

import (
    "errors"
    "os"

    "github.com/larsartmann/go-error-family"
)

func main() {
    err := errors.New("connection refused")

    family := errorfamily.Classify(err)
    // → Transient (default: unknown errors are retryable)

    if errorfamily.IsRetryable(err) {
        // yes — retrying is appropriate (backoff/jitter/idempotency are yours to implement)
    }

    os.Exit(errorfamily.ExitCode(err))
    // → 75 (EX_TEMPFAIL: temporary failure)
}
```

For your own errors, attach a Family at creation time:

```go
err := errorfamily.NewRejection("file.not_found", "config file missing").
    WithContext("path", "/etc/app/config.yaml")

os.Exit(errorfamily.HandleError(err))
// stderr: "A required resource was not found."
// stderr: "Check that the path and resource name are correct."
// exit: 1
```

See [examples/](examples/) for runnable CLI, HTTP, custom diagnostic rule, and oops-bridge demos.

## What It Gives You

- **`Family`** — behavioral classification (Rejection, Conflict, Transient, Corruption, Infrastructure, Orchestration) that maps to retry decisions, exit codes, HTTP status codes, and user-facing tone
- **Small interfaces** — `Coded`, `Classified`, `Contextual`, `Retryable`, `ExitCoder`, `HTTPStatuser` — each error type implements what it needs; the `Error` struct is just a reference implementation
- **`Classify(err)`** — universal classification for any error (multi-error → interface → sentinels → classifiers → default)
- **Multi-error support** — `errors.Join` + `Classify` picks the **worst** Family by severity, deterministically regardless of argument order
- **`ExitCode(err)`** — BSD sysexits.h exit codes derived from Family (overridable per-error via `ExitCoder` interface)
- **`Code(err)`** — one-liner error-code extraction (walks the unwrap chain for the `Coded` interface)
- **`IsRetryable(err)`** — binary retry decision derived from Family
- **`Family.HTTPStatus()`** — canonical family→HTTP status mapping (Rejection→400, Conflict→409, Transient→503, …)
- **`HTTPStatus(err)` / `HTTPHandler(fn)`** — classify→status-code helper and a ready-made net/http middleware writing safe JSON responses
- **`Family.RetryPolicy()`** — advisory retry defaults (attempts + backoff); the library does not run the loop
- **`Error.JSON()`** — canonical JSON view for API boundaries
- **`RegisterClassifier(func(error) (Family, bool))`** — predicate-based classification for dynamic third-party errors (e.g. `*sqlite.Error`) that can't be registered as sentinels
- **`Registry`** — injectable registry with `Clone()` for inherit-and-extend; test isolation and scoped error handling (no `t.Cleanup` needed)
- **`RegisterStdlibDefaults(reg)`** — pre-registered classifications for common stdlib errors (context/sql/os) with documented rationale
- **`TemplateForCode(code)`** — look up a registered message template without the full CLI pipeline (for HTTP/gRPC boundaries)
- **`LogError(err, logger)`** — structured `log/slog` logging with family/code/retryable/context fields
- **`errorfamilytest`** — test assertion helpers (`AssertFamily`, `AssertCode`, `AssertRetryable`, `AssertContext`, `AssertContextMissing`, `AssertExitCode`, `AssertHTTPStatus`), each with a godoc example
- **`HandleError(err)`** — CLI boundary handler with structured messages (What / Why / Fix / WayOut)
- **Diagnostic rules** — deterministic checks (PostgreSQL, filesystem, network, git) that auto-discover why an error occurred and emit structured `Fix{Summary, Command}`; the `RuleSpec` data-driven pattern has a runnable godoc example
- **`WrapOnce(err, family, code, msg)`** — idempotent wrap that prevents double-wrapping at API boundaries
- **`WithExitCode(code)` / `WithContextAny(key, value)`** — per-error exit code override and type-safe context attachment

## The Six Families

| Family             | Retryable | Exit Code | Whose Fault             | Tone          |
| ------------------ | --------- | --------- | ----------------------- | ------------- |
| **Rejection**      | no        | 1         | User                    | Instructional |
| **Conflict**       | no        | 1         | User (needs to resolve) | Explanatory   |
| **Transient**      | **yes**   | 75        | System                  | Reassuring    |
| **Corruption**     | no        | 65        | Data damage             | Urgent        |
| **Infrastructure** | no        | 69        | System                  | Apologetic    |
| **Orchestration**  | no        | 70        | Program (internal bug)  | Apologetic    |

Each family exposes `Audience()` (User / Ops / All) and `Tone()` for presentation-layer decisions.

## Constructors

```go
// Simple
err := errorfamily.NewRejection("config.invalid", "bad config")

// With cause
err := errorfamily.WrapTransient(dbErr, "db.timeout", "query timed out")

// With context (chainable)
err := errorfamily.NewRejection("file.not_found", "config missing").
    WithContext("path", "/etc/app/config.yaml").
    WithContext("format", "yaml")

// Formatted
err := errorfamily.Newf(errorfamily.Rejection, "file.not_found", "missing: %s", path)

// Multi-error (partial success) — use stdlib errors.Join, Classify picks the worst Family
err := errors.Join(err1, err2, err3)
```

Family-specific constructors: `NewRejection`, `NewConflict`, `NewTransient`, `NewCorruption`, `NewInfrastructure`, `NewOrchestration`. Wrap variants: `WrapRejection`, `WrapConflict`, `WrapTransient`, `WrapCorruption`, `WrapInfrastructure`, `WrapOrchestration`. Formatted wrap variants: `WrapRejectionf`, `WrapConflictf`, `WrapTransientf`, `WrapCorruptionf`, `WrapInfrastructuref`, `WrapOrchestrationf`.

## Classification

```go
// Classify any error → Family
family := errorfamily.Classify(err)

// Retry decision — binary signal only; backoff/idempotency are your responsibility
if errorfamily.IsRetryable(err) {
    retry(err)
}

// Exit code for CLI
os.Exit(errorfamily.ExitCode(err))

// Parse from string (e.g. config, HTTP headers)
family := errorfamily.ParseFamily("transient") // defaults to Transient for unknowns
```

Classification precedence — first match wins:

1. **Multi-error** (`errors.Join`) → classify each sub-error, worst severity wins
2. `Classified` interface → `ErrorFamily()`
3. `Retryable` interface → infer Transient (true) or Rejection (false)
4. Registered sentinels via `errors.Is` chain walk
5. Registered classifiers (`RegisterClassifier`) — predicate-based, for dynamic errors
6. Default → Transient (fail-open for retry)

## Registering Third-Party Errors

### Decision tree: which classification mechanism?

```
Do you OWN the error type?
├── YES → Implement the Classified interface (ErrorFamily()) on your type.
├         (Or use NewRejection/NewConflict/.../WrapRejection/... constructors.)
└── NO  → Is it a SENTINEL (a single, stable value compared by errors.Is)?
          ├── YES → RegisterClassification(sentinel, family)
          └── NO  → Is it a DYNAMIC type (new instance per error, e.g. *sqlite.Error)?
                    ├── YES → RegisterClassifier(func(error) (Family, bool))
                    └── NO  → Let Classify default to Transient (fail-open for retry)
```

### Sentinels (stable values)

For errors you don't own that are stable sentinel values (stdlib, libraries):

```go
func init() {
    errorfamily.RegisterClassification(sql.ErrConnDone, errorfamily.Transient)
    errorfamily.RegisterClassification(os.ErrPermission, errorfamily.Rejection)

    // Batch registration
    errorfamily.RegisterClassifications(map[error]errorfamily.Family{
        sql.ErrConnDone: errorfamily.Transient,
        sql.ErrTxDone:   errorfamily.Transient,
    })
}
```

### Classifiers (dynamic errors)

Some third-party errors are **dynamic** — each occurrence is a fresh instance
(e.g. `*sqlite.Error`, `*pgconn.PgError`), so they can't be matched by
`errors.Is` identity. Register a predicate-based classifier instead:

```go
func init() {
    errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
        var sqliteErr *sqlite.Error
        if errors.As(err, &sqliteErr) {
            switch sqliteErr.Code() {
            case 5, 6: return errorfamily.Transient, true // BUSY, LOCKED
            case 19:   return errorfamily.Conflict, true  // CONSTRAINT
            }
        }
        return errorfamily.Transient, false
    })
}
```

Classifiers run only after sentinels miss, in registration order; the first
returning `ok=true` wins. For test isolation, construct a `NewRegistry()` and
call its `RegisterClassifier` method instead of polluting `DefaultRegistry`.

## Implementing Your Own Error Type

You don't have to use the built-in `Error` struct. Implement the interfaces you need:

```go
package mypkg

import (
    "fmt"

    "github.com/larsartmann/go-error-family"
)

type FindingError struct {
    Category string
    Message  string
    File     string
    Line     int
    Cause    error
}

func (e *FindingError) Error() string          { return e.Message }
func (e *FindingError) Unwrap() error          { return e.Cause }
func (e *FindingError) ErrorCode() string      { return e.Category }
func (e *FindingError) ErrorFamily() errorfamily.Family {
    switch e.Category {
    case "validation":
        return errorfamily.Rejection
    case "io":
        return errorfamily.Transient
    default:
        return errorfamily.Infrastructure
    }
}
func (e *FindingError) ErrorContext() map[string]string {
    return map[string]string{"file": e.File, "line": fmt.Sprintf("%d", e.Line)}
}
```

Now `Classify()`, `IsRetryable()`, `ExitCode()`, and `HandleError()` all work with your type.

## CLI Boundary: HandleError

`HandleError` is the top-of-main handler that classifies the error, picks an exit code, formats a user-facing message, and writes to stderr:

```go
func main() {
    if err := run(); err != nil {
        os.Exit(errorfamily.HandleError(err))
    }
}
```

For more control:

```go
// Structured result (no stderr write) — for HTTP handlers, gRPC interceptors
result := errorfamily.HandleErrorDetailed(err)
fmt.Printf("exit=%d msg=%q fix=%q\n", result.ExitCode, result.Message, result.SuggestedFix)

// Template-aware structured result
result := errorfamily.HandleErrorDetailedWithConfig(err, cfg)

// Context-propagating handler — preferred when you have a context.Context
exitCode := errorfamily.HandleErrorWithContext(ctx, err, errorfamily.HandleConfig{
    Output:      myWriter,
    DiagnosticFunc: myDiagnoseFunc,
    OnDiagnosed: func(err error, findings []errorfamily.DiagnosticFinding) {
        logFindings(findings)
    },
    TemplateOverride: map[string]errorfamily.MessageTemplate{
        "file.not_found": {
            What:   "Could not find {path}",
            Why:    "The file doesn't exist at the expected location.",
            Fix:    "Check that {path} exists and is readable.",
            WayOut: "Run with --verbose for more details.",
        },
    },
})
```

### Template Resolution

Messages are resolved in this order (first match wins):

1. `HandleConfig.TemplateOverride[code]` — per-call consumer override
2. `RegisterTemplate(code, tmpl)` — global registry
3. Built-in `defaultMessages[code]` — exact-match defaults
4. `Family.DefaultMessage()` — generic family-based fallback

All lookups are exact code matches (case-insensitive). No substring matching.

Register templates globally for reusable error codes:

```go
errorfamily.RegisterTemplate("db.timeout", errorfamily.MessageTemplate{
    What: "Database connection timed out.",
    Why:  "The database server did not respond within the deadline.",
    Fix:  "Check database health and network connectivity.",
})
```

## HTTP Boundary

`HTTPStatus(err)` maps any error to its status code. `HTTPHandler` wraps an
error-returning handler and writes a **safe** JSON response (no internal leakage):

```go
func createOrder(w http.ResponseWriter, r *http.Request) error {
    order, err := parseOrder(r)
    if err != nil {
        return errorfamily.WrapRejectionf(err, "order.parse", "invalid field %s", field)
    }
    if err := repo.Save(order); err != nil {
        return errorfamily.WrapConflict(err, "order.duplicate", "order already exists")
    }
    json.NewEncoder(w).Encode(order)
    return nil
}

func registerRoutes(mux *http.ServeMux) {
    mux.Handle("/api/orders", errorfamily.HTTPHandler(createOrder))
}
// Rejection → 400, Conflict → 409, Transient → 503, ...
```

The response body contains only `family`, `code`, and a user-facing `message`
(from a registered template) — never the raw `err.Error()`:

```json
{
  "family": "conflict",
  "code": "order.duplicate",
  "message": "A conflict was detected."
}
```

For a custom response shape, write your own response and use `HTTPStatus(err)`
directly.

### Conditional Requests (304, 412, 428, 416)

Every family maps to an error status (4xx/5xx), so conditional-request outcomes
(RFC 9110 §13) need explicit handling:

| Outcome                                           | Classification                                                                                                                                                                                          |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **304 Not Modified**                              | Not an error. Write the 304 (plus `ETag`/`Cache-Control`) yourself and return `nil` — `HTTPHandler` treats `nil` as "fully handled". Classifying it would attach wrong retry/exit semantics.            |
| **412 Precondition Failed** (RFC 9110 §15.5.13)   | `Conflict` + `WithHTTPStatus(412)`. The client asserted state (`If-Match: "abc"`) that no longer holds — a version mismatch. Not retryable; remediation is "refresh and reapply", exactly `Conflict`'s. |
| **428 Precondition Required** (RFC 6585 §3)       | `Rejection` + `WithHTTPStatus(428)`. The request omitted a required precondition — incomplete input, not a state clash. Remediation is "fix the request", exactly `Rejection`'s.                        |
| **416 Range Not Satisfiable** (RFC 9110 §15.5.17) | `Rejection` + `WithHTTPStatus(416)`. The `Range` value doesn't overlap the resource — a bad input value.                                                                                                |

```go
func getWidget(w http.ResponseWriter, r *http.Request) error {
    // RFC 9110 §13.2.2: If-Match evaluates before If-None-Match
    if im := r.Header.Get("If-Match"); im != "" && !etagMatches(im, currentETag) {
        return errorfamily.NewConflict("etag.precondition_failed", "If-Match precondition failed").
            WithHTTPStatus(http.StatusPreconditionFailed) // 412, not the family default 409
    }
    if etagMatches(r.Header.Get("If-None-Match"), currentETag) {
        w.Header().Set("ETag", currentETag)
        w.WriteHeader(http.StatusNotModified) // success outcome — no error, no body
        return nil
    }
    // ...serve the current representation
    return nil
}

func registerWidgetRoute(mux *http.ServeMux) {
    mux.Handle("/widgets/{id}", errorfamily.HTTPHandler(getWidget))
}
// HTTPStatus(err) == 412 · Classify(err) == Conflict · IsRetryable(err) == false · ExitCode(err) == 1
```

Families stay behavioral (retry, exit code, tone); statuses stay wire-level.
`WithHTTPStatus` is the intended seam between them — the same pattern as a 404
Not Found `Rejection`. If you don't construct errors with this package,
implement the `HTTPStatuser` interface on your own type instead. To control the
client-facing `message` in `HTTPHandler` responses, register a template for the
code — `HTTPHandler` never leaks `err.Error()`.

For gRPC services, the [gRPC status-mapping guide](https://errorfamily.lars.software/guides/grpc/)
maps the same six families to gRPC status codes, with a ready interceptor pattern.

## Structured Logging

`LogError` logs classified fields (`family`, `code`, `retryable`, and each
context key prefixed with `context.`) at Warn for Transient errors and Error
for everything else:

```go
if err := run(); err != nil {
    errorfamily.LogError(err, slog.Default())
    // → level=WARN msg="db timeout" family=transient code=db.timeout retryable=true
}
```

`LogErrorContext(ctx, err, logger)` propagates a context for trace correlation.

## Test Helpers

The `errorfamilytest` subpackage mirrors `net/http/httptest` — it keeps
`testing` out of the production package. Every helper has a runnable-shaped
godoc example on pkg.go.dev (`ExampleAssertFamily` and friends):

```go
import "github.com/larsartmann/go-error-family/errorfamilytest"

func TestHandler(t *testing.T) {
    err := handler(req)
    errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
    errorfamilytest.AssertCode(t, err, "user.not_found")
    errorfamilytest.AssertRetryable(t, err, false)
    errorfamilytest.AssertContext(t, err, "user_id", "42")
}
```

## Diagnostic Rules

Auto-discover why an error occurred:

```go
runner := diagnose.DefaultRunner()
results := runner.Run(ctx, err)

for _, r := range results {
    if r.Status == diagnose.StatusFailed {
        fmt.Println(r.Summary)
        fmt.Println("  Fix:", r.Fix.Summary)
        if r.Fix.Command != "" {
            fmt.Println("  Run:", r.Fix.Command)
        }
    }
}
```

Built-in rules (zero-dependency, included in `DefaultRunner`):

- **FilesystemRule** — checks path existence, permissions, writability
- **NetworkRule** — checks DNS resolution, TCP connectivity

Opt-in submodules (import explicitly):

```go
import (
    "github.com/larsartmann/go-error-family/diagnose/git"
    "github.com/larsartmann/go-error-family/diagnose/postgres"
)

runner := diagnose.NewRunner(&git.GitRule{}, &postgres.PostgresRule{}, &diagnose.FilesystemRule{})
```

- **GitRule** (`diagnose/git`) — checks repo state, merge conflicts, remote reachability
- **PostgresRule** (`diagnose/postgres`) — checks `pg_isready`, TCP connectivity

Results include `Confidence` (0.0–1.0) and are sorted by confidence descending.

### Writing Custom Rules

Use `RuleSpec` for data-driven matching — the simplest and most common pattern:

```go
type RateLimitRule struct{}

const keyRetryAfter diagnose.ContextKey = "retry_after"

var rateLimitSpec = diagnose.RuleSpec{
    ContextKeys:  []diagnose.ContextKey{keyRetryAfter},
    CodeContains: []string{"rate.limit", "too_many_requests"},
}

func (r *RateLimitRule) Name() string              { return "rate_limit" }
func (r *RateLimitRule) Applicable(err error) bool { return rateLimitSpec.Matches(err) }

func (r *RateLimitRule) Run(ctx context.Context, err error) (*diagnose.DiagnosticResult, error) {
    retryAfter := diagnose.ResolveContextKey(err, []string{"retry_after"}, "unknown")
    // ... check system state ...
    return &diagnose.DiagnosticResult{
        Status:     diagnose.StatusDegraded,
        Confidence: diagnose.ConfidenceHigh,
        Summary:    "Rate limited — retry after " + retryAfter,
        Fix: diagnose.Fix{
            Summary: "Wait for the duration specified in the Retry-After header",
        },
    }, nil
}
```

For testable rules that shell out, accept a `CommandRunner`:

```go
type MyRule struct {
    Runner diagnose.CommandRunner // inject mock in tests; defaults to DefaultCommandRunner{}
}
```

See the [custom rule example](examples/README.md) for a complete working example.

## AI Debug Agent

Root cause analysis and fix suggestions from diagnostic context:

```go
ag := agent.New(agent.Config{Enabled: true})
result, _ := ag.Analyze(ctx, err, diagnosis)

fmt.Println("Root cause:", result.RootCause)
fmt.Println("Confidence:", result.Confidence)
for _, step := range result.FixSteps {
    fmt.Printf("  - %s\n    Command: %s\n", step.Description, step.Command)
}
```

The agent produces analysis but does **not** execute fixes — the consumer decides what to do with `FixStep.Command`.

`AgentResult` fields: `RootCause`, `Confidence`, `Explanation`, `FixSteps`.

## Performance

Zero-allocation hot paths. Benchmarks on AMD Ryzen 9 7950X:

| Operation                    | Time    | Allocs |
| ---------------------------- | ------- | ------ |
| `Classify` (built-in Error)  | ~9 ns   | 0      |
| `Classify` (plain error)     | ~30 ns  | 0      |
| `IsRetryable`                | ~9 ns   | 0      |
| `ExitCode`                   | ~9 ns   | 0      |
| `WithContext`                | ~8 ns   | 0      |
| `ParseFamily`                | ~12 ns  | 0      |
| `HandleError` (with context) | ~450 ns | 5      |
| `Runner.Run` (1 rule)        | ~420 ns | 5      |

`HandleError` includes template resolution and stderr write. Use `HandleErrorDetailed` when you only need the structured result.

## Architecture

```
go-error-family/
├── family.go               — Family enum + data-driven familyData
├── interfaces.go           — Coded, Classified, Contextual, Retryable, ExitCoder, HTTPStatuser (each embeds error)
├── error.go                — Reference Error struct (Is, Unwrap, Format, WithContext, accessors)
├── classify.go             — Classify, Code, IsRetryable, ExitCode, Classifier, RegisterClassifier(s)
├── constructors.go         — New, Wrap, Newf, Wrapf + family-specific shortcuts (incl. Wrap{Family}f)
├── registry.go             — injectable Registry (sentinels, classifiers, templates), Clone, DefaultRegistry
├── stdlib.go               — RegisterStdlibDefaults (context/sql/os error taxonomy)
├── handle.go               — HandleError, HandleErrorWithContext, template system, TemplateForCode
├── http.go                 — HTTPStatus, HTTPHandler (classify→status-code net/http middleware)
├── log.go                  — LogError, LogErrorContext (structured slog logging)
├── errorfamilytest/        — test assertion helpers (AssertFamily, AssertCode, ...)
├── diagnose/
│   ├── diagnose.go         — Runner, DiagnosticRule, RuleSpec, CommandRunner, ContextKey
│   ├── command.go          — RunCommand, CommandExists (exported for rule authors)
│   ├── rules_filesystem.go — FilesystemRule
│   ├── rules_network.go    — NetworkRule
│   ├── git/                — submodule: GitRule
│   └── postgres/           — submodule: PostgresRule
├── agent/
│   └── agent.go            — DebugAgent interface, Config, AgentResult, FixStep
├── bridge/                 — submodule: samber/oops integration (Wrap, AutoWrap, InferFamily)
└── examples/
    ├── cmd/cli             — CLI boundary handler example
    ├── cmd/http            — HTTP middleware with status code mapping
    ├── cmd/custom_rule     — Writing your own DiagnosticRule
    ├── cmd/bridge          — oops + bridge reference implementation (3 patterns)
    └── checkout/           — library layer that imports only errorfamily
```

## When to Use / When Not To

**Use this when** you need behavior derived from errors: retry decisions, exit codes, user-facing messages, or diagnostic investigation — especially at program boundaries (CLI top-level, HTTP handlers, gRPC interceptors).

**Don't use this when** a simple `fmt.Errorf("...: %w", err)` is enough. If you only need to wrap and propagate, the standard library is fine. This library shines when errors need to _drive behavior_.

## Philosophy

1. **Protocol, not framework** — share the vocabulary, not the implementation
2. **Each error type implements what it needs** — no god struct
3. **Presentation is separate from the error** — the CLI layer formats for humans
4. **Family = exit code = retry decision = tone** — one concept, many audiences
5. **Generic errors are structurally impossible** — Code + Family + Context guarantee specificity

## License

[MIT](LICENSE) — Copyright (c) 2026 Lars Artmann
