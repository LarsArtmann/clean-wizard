# Contributing to go-finding

Thank you for your interest in contributing to `go-finding`. This document outlines the development setup, coding standards, and PR process.

## Development Setup

### Prerequisites

- Go 1.26 or later
- `golangci-lint` (for linting)

### Nix Setup (optional)

If you use [Nix](https://nixos.org/), you can get a reproducible development environment:

```bash
# Enter dev shell (if flake.nix is configured)
nix develop

# Or with direnv (auto-loads on cd)
echo "use flake" > .envrc
direnv allow
```

### Getting Started

```bash
git clone https://github.com/larsartmann/go-finding.git
cd go-finding
go mod download
```

### Build & Test

```bash
# Run all tests (must use -race)
go test -race -count=1 ./...

# Run tests with coverage
go test -race -count=1 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Build all packages
go build ./...

# Run linter
golangci-lint run ./...

# Run vet
go vet ./...
```

### Using Nix (recommended)

If you have [Nix](https://nixos.org/) with flakes enabled:

```bash
nix run .#test                              # Run all tests with -race
nix run .#bench                             # Run benchmarks
nix run .#lint                              # golangci-lint
nix run .#check                             # All checks (fmt + lint + test)
nix build                                   # Build the CLI binary
```

### Using Go directly

```bash
go test -race -count=1 ./...
go test -bench=. -benchmem ./...
golangci-lint run ./...
go test -coverprofile=coverage.out ./...
go vet ./...
```

## Project Structure

Multi-module Go workspace — each module has a single purpose and composes independently.

```
go-finding/
│
├── Core module (github.com/larsartmann/go-finding)
│   ├── finding.go              # Core Finding type
│   ├── finding_builder.go      # Fluent builder API + Template factory
│   ├── finding_methods.go      # Finding methods (HasFix, IsSuppressedAt, etc.)
│   ├── finding_validate.go     # Validate() and 6 per-field validators
│   ├── finding_equal.go        # Equality semantics (tag-order-insensitive)
│   ├── validate_helpers.go     # Shared validation utilities
│   ├── severity.go             # Severity enum, Badge, SeverityFromLevel, PriorityString
│   ├── confidence.go           # Confidence named type (0.0–1.0)
│   ├── category.go             # Category constants
│   ├── category_linter.go      # LinterRegistry: linter→category mapping (84+ linters)
│   ├── fix_strategy.go         # FixStrategy enum
│   ├── tag.go                  # Tag type with standard constants
│   ├── suppression.go          # Suppression handling with TTL
│   ├── branded_types.go        # ID, RuleName, ToolName, FilePath branded string types
│   ├── position.go             # Position type, FilePos constructor
│   ├── range.go                # Range type
│   ├── range_overlap.go        # Range overlap/intersection/extension logic
│   ├── report.go               # Report container (thread-safe), NewReportFromFindings
│   ├── report_query.go         # Report query methods (FindByID, FindingsSnapshot)
│   ├── filter.go               # Filtering and grouping
│   ├── merge.go                # Merge, dedup, MergeIter streaming
│   ├── correlate.go            # Correlation scoring between findings
│   ├── diff.go                 # Diff (before/after finding sets)
│   ├── interval_index.go       # Generic IntervalIndex[T] for overlap queries
│   ├── format.go               # FormatText/FormatTextRich/FormatMarkdown/FormatTable
│   ├── simple_fix.go           # ApplySimpleFixes (BeforeCode→AfterCode replacement)
│   ├── detector.go             # Detector interface, CheckBinary, RunCmd helpers
│   ├── adapter.go              # ToolAdapter[O] generic adapter for external tools
│   ├── registry.go             # DetectorRegistry (thread-safe plugin architecture)
│   ├── sarif_export.go         # SARIF 2.1.0 export (hand-rolled, context-aware)
│   ├── sarif_import.go         # SARIF 2.1.0 import
│   ├── sarif_types.go          # SARIF struct types and constants
│   ├── id.go                   # ID generation (length-prefixed hash) and parsing
│   ├── errors.go               # Structured errors (FindingError, ErrorCategory, must[T])
│   ├── context.go              # Context helpers
│   ├── json.go                 # JSON marshaling (encoding/json/v2)
│   ├── lsp.go                  # LSP diagnostic conversion (lossless round-trip)
│   ├── doc.go                  # Package documentation
│   ├── version.go              # Version constants
│   ├── gotoken/                # Shared go/token utilities (public package, stdlib only)
│   └── lockutil/               # Generic sync.Locker helpers (Locked/RLocked)
│
├── analysis/                   # go/analysis.Diagnostic integration (x/tools dep)
│   ├── analysis.go             # FromDiagnostic converts go/analysis → Finding
│   └── adapter.go              # Adapter for single-pass analyzers
│
├── pipeline/                   # Pipeline package (x/sync, gogenfilter deps)
│   ├── pipeline.go             # Pipeline orchestrator (Run entry point)
│   ├── pipeline_detect.go      # Detection stage (parallel via errgroup)
│   ├── pipeline_iteration.go   # Detect→triage→fix→verify iteration loop
│   ├── config.go               # Config struct with validation
│   ├── config_file.go          # JSON/YAML config file loading
│   ├── convenience.go          # Detect() and ApplyToContent() one-shot helpers
│   ├── result.go               # PipelineResult type
│   ├── stage.go                # Stage type and Stage constants
│   ├── stage_hook.go           # StageHook interface + StageHookFunc
│   ├── adapters.go             # Detector, FindingTransformer interfaces
│   ├── conflict.go             # Fix conflict detection (AnalyzeConflicts)
│   ├── fix_engine.go           # Byte-level edit engine (descending-offset)
│   ├── fix_outcome.go          # Per-finding fix outcomes (FixOutcome, FixApplyResult)
│   ├── fix_edit.go             # FixEdit type (JSON wire format: fixEditJSON)
│   ├── fix_provider.go         # Composable fix providers (chain of responsibility)
│   ├── fix_provider_helpers.go # OffsetProvider, LineProvider, SubstringProvider
│   ├── fix_applier.go          # Filesystem fix application with backup/rollback
│   ├── file_backup.go          # File backup/restore for fix rollback
│   ├── flight_recorder.go      # Chrome Trace Event export (runtime/trace)
│   ├── generated_filter.go     # Auto-generated Go file detection (gogenfilter)
│   ├── line_shift.go           # LineShiftMap for post-edit position tracking
│   ├── metrics.go              # Timing/count metrics with snapshots
│   ├── retry.go                # Retry with exponential backoff
│   ├── partial.go              # Partial success handling
│   ├── path_safety.go          # Path validation (resolveSafePath)
│   ├── verify.go               # Verification stage (re-run detectors)
│   ├── goast/                  # AST-aware fix provider (opt-in go/parser dep)
│   │   └── provider.go
│   └── internal/benchutil/     # Benchmark utilities
│       └── benchutil.go
│
├── cmd/go-finding/             # CLI tool (own module: yaml, go-output deps)
│   ├── main.go                 # Entry point, CLI flag parsing
│   ├── config.go               # Config loading from YAML/JSON
│   ├── registry.go             # Detector registration
│   ├── fix_provider_registry.go # Fix provider registration
│   ├── generated_filter.go     # CLI generated file filter
│   ├── output_adapter.go       # []Finding → output.Table (markdown/CSV/TSV)
│   └── internal/detectors/     # Built-in detectors
│       ├── govet.go
│       ├── staticcheck.go
│       └── helpers.go
│
├── examples/                   # Standalone examples
│   ├── basic/main.go
│   └── builder/main.go
│
├── docs/                       # Documentation, guides, reviews
├── scripts/                    # Build/test helper scripts
├── go.work                     # Go workspace (coordinates all 4 modules)
├── flake.nix                   # Nix flake (devShell, build, test, lint)
└── .golangci.yml               # Linter configuration
```

## Coding Standards

### Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Prefer composition over inheritance
- Use functional options for configuration
- Early returns over nested conditionals
- Descriptive names over comments
- No one-letter variable names except in tight loops

### Types

- All exported enums are string types (`Severity`, `Category`, `FixStrategy`, `ErrorCategory`)
- Use `IsValid()` methods for validation
- Make impossible states unrepresentable
- Immutable data: `Finding` structs are data, not state machines

### Error Handling

Use structured `FindingError` with categories:

```go
return finding.NewValidationError("invalid severity", nil)
return finding.NewIOError("read file", err).WithPosition(pos)
```

Never return bare `fmt.Errorf` from library code. Wrap external errors:

```go
return fmt.Errorf("detector %s: %w", d.Name(), err)
```

### Testing Requirements

- All new code must have tests
- Test behavior, not implementation
- Integration tests over unit tests where practical
- Real implementations over mocks
- Use `Example*()` functions for GoDoc examples
- Use `testing.F` for fuzz tests when testing parsing/serialization

### Naming Conventions

- Filter predicates: `BySeverity`, `ByCategory`, `ByFile`
- Configuration options: `WithDeduplication`, `WithDeduplicateBy`
- Constructor functions: `NewReport`, `NewMetrics`, `NewFixApplier`
- Boolean getters: `IsValid`, `HasFix`, `HasEnd`, `IsSuppressed`

## PR Process

### Before Submitting

1. Run the full test suite: `go test -race -count=1 ./...`
2. Run vet: `go vet ./...`
3. Run linter: `golangci-lint run ./...`
4. Verify all examples build: `go build ./...`
5. Ensure GoDoc examples pass: `go test -run Example ./...`
6. Check test coverage has not decreased
7. Verify benchmarks still work: `go test -bench=. -benchmem ./...`

### PR Guidelines

- One logical change per PR
- Write clear commit messages explaining why, not what
- Include tests for all new functionality
- Update documentation if adding public API surface
- Keep PRs focused and reviewable

### Commit Messages

Use conventional commit format:

```
feat: add support for custom deduplication strategies
fix: handle Windows paths in ID generation correctly
docs: add usage guide for pipeline configuration
test: add fuzz tests for ID parsing
refactor: extract severity comparison to standalone method
```

## Architecture Decisions

### Zero Dependencies for Core Types

The root `finding` package uses only standard library types. This makes it safe to import in any Go project without dependency conflicts.

### Lossless Conversions

All conversions (Finding → SARIF, Finding → LSP Diagnostic) preserve the original data. Round-trip conversions should be lossless.

### String-Based Enums

`Severity`, `Category`, `FixStrategy` are string types, not int-based enums. This ensures JSON serialization is human-readable and avoids int↔string conversion errors.

### Pipeline as Orchestrator

The pipeline uses a functional stage pattern. Each stage is a pure function that transforms input to output. The pipeline manages iteration, context, and error propagation.
