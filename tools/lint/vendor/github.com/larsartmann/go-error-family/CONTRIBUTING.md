# Contributing to go-error-family

Thank you for contributing! This guide covers everything you need to know.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Architecture](#architecture)
- [Pull Request Process](#pull-request-process)
- [Commit Messages](#commit-messages)

## Code of Conduct

Be respectful, inclusive, constructive, and collaborative.

## Development Setup

### Prerequisites

| Tool          | Version | Purpose          |
| ------------- | ------- | ---------------- |
| Go            | 1.26+   | Language runtime |
| golangci-lint | latest  | Code linting     |

### Setup

```bash
# Clone
git clone https://github.com/LarsArtmann/go-error-family.git
cd go-error-family

# Verify
export PATH="$PATH:$(go env GOPATH)/bin"
go build ./...
go test ./... -count=1 -timeout 120s -race
golangci-lint run ./...
```

With Nix:

```bash
nix develop                           # dev shell with Go + linters
nix run .#test                        # run tests
nix run .#lint                        # run linter
nix run .#coverage                    # coverage report
nix flake check                       # all checks (build + lint)
```

## Code Standards

### Principles

1. **Zero third-party dependencies** — this is a library; stdlib only (`encoding/json`)
2. **Composition over inheritance** — interfaces and struct embedding, not class hierarchies
3. **Small interfaces** — each error type implements only what it needs (`Coded`, `Classified`, `Contextual`, `Retryable`)
4. **Early returns** — guard clauses over nested conditionals
5. **Strong types** — make impossible states unrepresentable (e.g. `ContextKey` instead of raw strings)

### Naming

| Kind       | Convention             | Example                   |
| ---------- | ---------------------- | ------------------------- |
| Packages   | lowercase, single word | `diagnose`, `agent`       |
| Interfaces | behavioral             | `DiagnosticRule`, `Coded` |
| Functions  | PascalCase             | `HandleError`, `Classify` |
| Constants  | PascalCase             | `KeyHost`, `KeyPort`      |

### Error Construction

Use family-specific constructors — never call `New` or `Wrap` with a raw `Family` constant directly when a named constructor exists:

```go
// Good
err := errorfamily.NewRejection("file.not_found", "config missing")
err := errorfamily.WrapTransient(cause, "db.timeout", "query timed out")

// Avoid — use the named constructor instead
err := errorfamily.New(errorfamily.Rejection, "file.not_found", "config missing")
```

## Testing

### Commands

```bash
# All tests (race-detector enabled) — root package only
go test ./... -count=1 -timeout 120s -race

# Submodule tests (must run from within the submodule)
cd diagnose/git && go test -race ./...
cd diagnose/postgres && go test -race ./...

# Single package
go test ./diagnose/git/ -v -run TestGitRule

# Coverage
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

### Test Patterns

- Use table-driven tests for classification and formatting logic
- Use mock `CommandRunner` for diagnostic rules that shell out (`diagnose/git`, `diagnose/postgres`)
- Use `Error.WithTimestamp(ts)` for deterministic timestamp assertions
- Fuzz tests live in `fuzz_test.go` at the root

### Coverage Targets

| Package              | Target |
| -------------------- | ------ |
| root (`errorfamily`) | 95%+   |
| `agent`              | 100%   |
| `diagnose/git`       | 95%+   |
| `diagnose/postgres`  | 80%+   |
| `diagnose` (core)    | 80%+   |

## Architecture

```
errorfamily/          — root package: types, constructors, classification, CLI boundary
  error.go              Error struct (reference implementation)
  family.go             Family enum + Audience/Tone metadata
│   interfaces.go       Coded, Classified, Contextual, Retryable, ExitCoder
  constructors.go       New/Wrap + family-specific shortcuts
  classify.go           Classify(), RegisterClassification()
  handle.go             HandleError(), HandleErrorWithContext(), template system

diagnose/             — concurrent diagnostic rules (zero-dep core)
  diagnose.go           Runner, DiagnosticRule, RuleSpec, CommandRunner, ContextKey
  command.go            RunCommand, CommandExists
  rules_filesystem.go   FilesystemRule
  rules_network.go      NetworkRule

diagnose/git/         — submodule: GitRule (opt-in)
diagnose/postgres/    — submodule: PostgresRule (opt-in)

agent/                — analysis-only debug agent
  agent.go              DebugAgent interface, AgentResult, FixStep
```

### Key Design Decisions

- **No `main` package** — library only
- **Consumer interfaces embed `error`** — required for Go 1.26's `errors.AsType[T]()`
- **Agent is analysis-only** — the library never executes `FixStep.Command`
- **Diagnostic submodules are opt-in** — `DefaultRunner()` includes only zero-dep rules
- **`Classify(nil)` returns Rejection** — nil error is the caller's fault

## Adding a Diagnostic Submodule

If you want to add a new diagnostic rule that depends on external tools (e.g. `docker`, `kubectl`, `redis-cli`), create a submodule instead of adding it to the core `diagnose` package.

### Steps

1. Create a new directory under `diagnose/<name>/`
2. Add a `go.mod` with `module github.com/larsartmann/go-error-family/diagnose/<name>`
3. Add a `go.sum` via `go mod tidy`
4. Register the submodule in `go.work`
5. Implement the rule with an optional `CommandRunner` field for testability
6. Add table-driven tests with a mock `CommandRunner`
7. Update `README.md` and `AGENTS.md`

### Template

```go
package mytool

import (
    "context"

    "github.com/larsartmann/go-error-family/diagnose"
)

type MyToolRule struct {
    Runner diagnose.CommandRunner // defaults to DefaultCommandRunner{}
}

var myToolSpec = diagnose.RuleSpec{
    ContextKeys:  []diagnose.ContextKey{"mytool_host"},
    CodeContains: []string{"mytool."},
}

func (r *MyToolRule) Name() string              { return "mytool" }
func (r *MyToolRule) Applicable(err error) bool { return myToolSpec.Matches(err) }

func (r *MyToolRule) Run(ctx context.Context, err error) (*diagnose.DiagnosticResult, error) {
    runner := r.Runner
    if runner == nil {
        runner = diagnose.DefaultCommandRunner{}
    }
    // ... run checks via runner.RunCommand ...
    return &diagnose.DiagnosticResult{
        Status:     diagnose.StatusHealthy,
        Confidence: diagnose.ConfidenceHigh,
        Summary:    "MyTool is reachable",
    }, nil
}
```

## Pull Request Process

1. **Self-review** — run `go test ./... -race` and `golangci-lint run ./...` locally
2. **Small PRs** — focused changes are easier to review
3. **Explain why** — not just what changed, but the rationale
4. **Update docs** — if you add an API, update README.md and AGENTS.md
5. **Never ship a `replace` directive in a tagged `go.mod`** — Go strips `replace` on fetch, so consumers hit an unresolvable module-graph edge (the v0.6.0 incident). Use `go.work` for local multi-module development; tag real versions instead.

### Branch Naming

```
feat/description
fix/description
docs/description
refactor/description
test/description
```

## Commit Messages

```
<type>(<scope>): <subject>

<body>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `ci`

Example:

```
feat(diagnose): add CommandRunner interface for testable rules

Allows diagnostic rules to accept an injectable command runner
instead of calling system commands directly, enabling mock-based
testing without shell dependencies.
```

---

Thank you for contributing to go-error-family!
