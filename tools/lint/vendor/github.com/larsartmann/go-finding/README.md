# go-finding

A Go library providing a unified data model and pipeline for static analysis tools.

[![CI](https://github.com/larsartmann/go-finding/actions/workflows/ci.yml/badge.svg)](https://github.com/larsartmann/go-finding/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-finding.svg)](https://pkg.go.dev/github.com/larsartmann/go-finding)
[![codecov](https://codecov.io/gh/larsartmann/go-finding/branch/master/graph/badge.svg)](https://codecov.io/gh/larsartmann/go-finding)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why

Seven tools detect issues. Zero tools **route them to remediation**.

Each tool invents its own types for findings. There is no standardized way to apply fixes. The manual loop — run tool, read output, fix, re-run — is slow and error-prone.

**go-finding** solves this with:

- **Unified Finding type** — Common model for all static analysis tools
- **Pipeline** — Automated detect → triage → fix → verify loop with retry, partial success, and observability hooks
- **SARIF 2.1.0** — Standard interchange format for CI/CD integration
- **LSP diagnostics** — IDE integration out of the box
- **Flight recorder** — Go execution trace capture for pipeline diagnostics, with slow-stage auto-snapshot and manual checkpoints. See [FlightRecorder Guide](docs/guides/flight-recorder.md).
- **Finding groups** — `GroupID` ties related findings together (e.g. clone groups), round-tripping through JSON, SARIF, and LSP
- **Per-finding fix outcomes** — Know exactly what happened to every fix: applied, refused, conflict, invalid, or failed — with scoped per-file rollback by default

## Installation

> **Prerequisite — Go 1.27+.**
> This library uses `encoding/json/v2`, which is stable and enabled by default
> since Go 1.27 — no `GOEXPERIMENT` setup required. (On Go 1.26 or older the
> import fails with `build constraints exclude all Go files`; upgrade instead.)

**Core types only**:

```bash
go get github.com/larsartmann/go-finding
```

**With pipeline** (adds `x/sync`, `gogenfilter`):

```bash
go get github.com/larsartmann/go-finding/pipeline
```

**CLI tool**:

```bash
go install github.com/larsartmann/go-finding/cmd/go-finding@latest
```

Each module is an independent Go module and is versioned with its own git tag:

| Module   | Import path                                        | Tag                 |
| -------- | -------------------------------------------------- | ------------------- |
| Core     | `github.com/larsartmann/go-finding`                | `v1.15.0`           |
| Pipeline | `github.com/larsartmann/go-finding/pipeline`       | `pipeline/v*`       |
| Analysis | `github.com/larsartmann/go-finding/analysis`       | `analysis/v*`       |
| Tool SDK | `github.com/larsartmann/go-finding/toolsdk`        | `toolsdk/v*`        |
| CLI      | `github.com/larsartmann/go-finding/cmd/go-finding` | `cmd/go-finding/v*` |

See [`docs/release-procedure.md`](docs/release-procedure.md) for details.

## Quick Start

**Recommended:** Use the `Builder` API for validated findings with fix strategies,
categories, and metadata. The Builder validates at construction time, preventing
invalid findings from entering your pipeline.

```go
package main

import (
    "fmt"
    "log"

    "github.com/larsartmann/go-finding"
)

func main() {
    // Builder API — validated, fluent, recommended for all new code.
    f, err := finding.NewBuilder(
        finding.RuleName("unused-var"), finding.ToolName("my-tool"),
        "variable x is unused",
        finding.SeverityWarning,
        finding.Pos("main.go", 42, 5),
    ).
        WithCategory(finding.CategoryCorrectness).
        WithConfidence(finding.ConfidenceHigh).
        WithFixStrategy(finding.FixStrategySuggest).
        WithSuggestion("remove the unused variable").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    report := finding.NewReport(finding.ToolInfo{Name: "my-tool", Version: "1.0.0"})
    report.AddFinding(f)
    report.ComputeSummary()

    sarif, _ := report.ToSARIF()
    fmt.Println(string(sarif))
}
```

> **Lower-level:** `finding.NewFinding(...)` skips validation and is intended for
> cases where you construct findings from trusted sources. Prefer `NewBuilder`
> unless you have a specific reason to skip validation.

### One-shot detection (no fix loop)

For tools that only need detection without the fix pipeline:

```go
detectors := []pipeline.Detector{myDetector1, myDetector2}
findings, err := pipeline.Detect(ctx, detectors...)
// → []finding.Finding, ready for Report/SARIF output
```

### Applying fixes to in-memory content

```go
result, applied := pipeline.ApplyToContent(fileContent, findings)
// result = modified []byte, applied = count of successful fixes
```

### Per-finding fix outcomes (v1.7.0)

Know exactly what happened to each finding during a fix run — no more guessing
whether a fix was applied, refused, conflicted, or failed:

```go
result := engine.ApplyWithOutcomes(content, fixes)
for _, oc := range result.Outcomes {
    fmt.Println(oc.Finding.ID, oc.Status) // applied / no-change / refused / conflict / invalid / failed
    if oc.Err != nil {
        fmt.Println("  cause:", oc.Err)
    }
}

// Aggregate: pipeline.Metrics.RecordOutcome + OutcomeCounts feed the CLI's
// "Fix outcomes:" summary. Full guide: docs/guides/outcomes.md
```

Plan-before-apply: `applier.ApplyDryRun(ctx, findings)` (v1.8.0) returns the
same report shape with zero writes. Whole-pipeline runs expose the same
per-finding record on `PipelineResult.Outcomes`, deduplicated to the first
outcome per finding.

### Multi-edit fixes (unreleased)

Fixes that touch more than one place carry a typed edit list. `Finding.Edits`
is the authoritative machine representation; `BeforeCode`/`AfterCode` stay as
the first-edit display summary:

```go
f := finding.NewBuilder("rename-var", "mytool", "x is a misleading name",
    finding.SeverityWarning, finding.Pos("main.go", 5, 5)).
    WithFixStrategy(finding.FixStrategyDirect).
    WithEdits(
        finding.TextEdit{Start: finding.Pos("main.go", 5, 5), End: finding.Pos("main.go", 5, 6), NewText: "count"},
        finding.TextEdit{Start: finding.Pos("main.go", 7, 3), End: finding.Pos("main.go", 7, 4), NewText: "count"},
    ).
    MustBuild()

// The pipeline applies the whole list in one pass; applied counts stay
// per finding (one fix, two edits). The go/analysis bridge fills Edits from
// every TextEdit of a suggested fix, so analyzer fixes are lossless (issue #36).
// core finding.ApplySimpleFixes refuses multi-edit lists — use the pipeline.
```

## Core Types

| Type                                  | Purpose                                                      |
| ------------------------------------- | ------------------------------------------------------------ |
| `Finding`                             | A single issue: ID, rule, severity, position, fix strategy   |
| `Report`                              | Thread-safe container for findings with summary statistics   |
| `Severity`                            | `info` / `warning` / `error` / `critical`                    |
| `FixStrategy`                         | `none` / `suggest` / `direct` / `ai`                         |
| `Position`                            | File, line, column location                                  |
| `Range`                               | Start and end positions with geometric operations            |
| `Category`                            | `security`, `style`, `performance`, `correctness`, etc.      |
| `Tag`                                 | Multi-label classification (`security`, `performance`, ...)  |
| `Confidence`                          | Named `float64` with `IsValid()`, `Clamp()`, `String()`      |
| `Suppression`                         | Expiring suppression with `IsActive(now)`                    |
| `ID`/`RuleName`/`ToolName`/`FilePath` | Branded string types preventing field mixups at compile time |

## API Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Detectors (govet, staticcheck, custom)                     │
│         ↓                                                   │
│  []finding.Finding                                          │
│         ↓                                                   │
│  Report (thread-safe container)                             │
│         ↓                                                   │
│  Filter / Group / Merge / Correlate / Diff                  │
│         ↓                                                   │
│  SARIF / JSON / LSP / Text / Markdown / CSV / TSV             │
└─────────────────────────────────────────────────────────────┘
```

Key packages:

- `finding` — core types, filtering, grouping, merging, SARIF, LSP, formatting
- `pipeline` — detect → triage → fix → verify loop
- `analysis` — `go/analysis.Diagnostic` ↔ `Finding` conversion
- `toolsdk` — declarative plugin contract for external tools (`Spec`, triggers, registry)
- `cmd/go-finding` — CLI tool with JSON/YAML config

## Filtering

```go
errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))

autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))

important := finding.Filter(findings,
    finding.BySeverityAtLeast(finding.SeverityWarning),
    finding.NotSuppressed,
    finding.WithFix,
)

byFile := finding.GroupByFile(findings)
bySeverity := finding.GroupBySeverity(findings)
```

## Merging

Combine reports from multiple tools with deduplication:

```go
merged := finding.Combine([]*finding.Report{govet, staticcheck, custom},
    finding.WithDeduplication(true),
)
```

Cross-tool correlation finds related findings:

```go
correlations := finding.Correlate(allFindings)
for _, c := range correlations {
    fmt.Printf("%.1f: %s\n", c.Score, c.Reason)
}
```

## Pipeline

The `pipeline` package runs a detect → process → triage → apply → verify loop:

```
detect ──→ process ──→ triage ──→ apply ──→ verify
  ↑                                      │
  └────────────── repeat ────────────────┘ (until stable or max iterations)
```

Each iteration runs registered detectors, applies `FindingTransformer` transforms,
categorizes findings by `FixStrategy`, applies direct fixes with conflict detection,
and optionally re-runs detectors to verify.

```go
detector := pipeline.NamedDetectorFunc("my-tool", func(ctx context.Context) ([]finding.Finding, error) {
    return []finding.Finding{...}, nil
})

cfg := pipeline.Config{
    MaxIterations:     5,
    ParallelDetectors: true,
    Timeout:           10 * time.Minute,
    VerifyAfterFix:    true,
    GracefulDegradation: true,
    DryRun:            false,
}

p, err := pipeline.New(cfg, ".", detector)
if err != nil {
    log.Fatal(err)
}
result, err := p.Run(context.Background())

fmt.Printf("Iterations: %d, Findings: %d, Stable: %v\n",
    result.TotalIterations, result.TotalDetected, result.Stable())
```

### Pipeline Features

| Feature                           | Description                                                       |
| --------------------------------- | ----------------------------------------------------------------- |
| **Parallel detection**            | errgroup-based concurrent detector execution                      |
| **Finding processors**            | Composable transforms between detection and triage                |
| **Custom triage**                 | `Config.TriageFunc` overrides default categorization              |
| **Byte-level conflict detection** | `Config.ByteLevelConflictDetection` filters overlapping edits     |
| **Fix provider chain**            | Offset → Line → Substring, plus custom AST-aware providers        |
| **Fix application**               | Byte-level edits with backup/rollback                             |
| **Verification**                  | Re-run detectors to confirm fixes                                 |
| **Retry**                         | Exponential backoff for flaky detectors                           |
| **Partial success**               | Continue with findings from successful detectors                  |
| **Metrics**                       | Optional timing and count collection with snapshots               |
| **Structured logging**            | `*slog.Logger` integration                                        |
| **Stage hooks**                   | `StageHooks` before/after events with abort (replaces `OnStage`)  |
| **Dry run**                       | Detect + triage without applying fixes                            |
| **Generated file filter**         | Removes findings from auto-generated Go source files              |
| **Flight recorder**               | Chrome Trace Event export for pipeline stage timing visualization |

### Custom Detector

```go
type MyDetector struct{}

func (d *MyDetector) Name() string { return "my-detector" }

func (d *MyDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
    findings := []finding.Finding{
        finding.NewFinding("RULE001", "my-detector", "issue found",
            finding.SeverityError,
            finding.Position{File: "main.go", Line: 10}, finding.ConfidenceHigh),
    }
    return findings, nil
}
```

### Fix Providers

The pipeline resolves findings to byte-level edits via a composable provider chain:

```go
// Default chain: EditListProvider → OffsetProvider → LineProvider → SubstringProvider
applier, err := pipeline.NewFixApplier(rootDir)
defer applier.Close()

// Custom providers for AST-aware transformations
applier, err = pipeline.NewFixApplierWithProviders(rootDir, myASTProvider)
```

`EditListProvider` (first in the chain) resolves typed `Finding.Edits` lists —
byte offsets directly, line/column via the line index. Lists spanning files
fail with `pipeline.ErrEditCrossFile`; stale or out-of-bounds offsets fail
with `pipeline.ErrEditStale` — never a silent partial application.

A built-in Go AST provider disambiguates BeforeCode occurrences structurally:

```go
import "github.com/larsartmann/go-finding/pipeline/goast"

applier, err := pipeline.NewFixApplierWithProviders(rootDir, &goast.Provider{})
```

The SubstringProvider fallback uses nearest-position matching (line + column distance)
to disambiguate multiple occurrences of the same text.

### Diff and Compare

```go
result := finding.Diff(before, after)
fmt.Println(result.Stats()) // "+2 -1 ~0 =3"
fmt.Println(result.HasChanges())
```

## Tool Adapters

The `ToolAdapter[O]` generic wraps any external tool into a `Detector`:

```go
detector := finding.NewToolAdapter("staticcheck",
    func(ctx context.Context) ([]byte, error) {
        return exec.CommandContext(ctx, "staticcheck", "-json", "./...").Output()
    },
    func(data []byte) ([]staticcheckIssue, error) {
        var issues []staticcheckIssue
        return issues, json.Unmarshal(data, &issues)
    },
    func(issue staticcheckIssue) finding.Finding {
        return finding.NewFinding(issue.Rule, "staticcheck", issue.Message,
            finding.SeverityError, finding.Pos(issue.File, issue.Line, 0),
            finding.ConfidenceHigh)
    },
)
```

### CategoryForLinter

70+ built-in linter→category mappings:

```go
cat := finding.CategoryForLinter("SA1000") // CategoryCorrectness
cat := finding.CategoryForLinter("G104")  // CategorySecurity
finding.RegisterLinterCategory("MY-RULE", finding.CategoryPerformance)
```

## SARIF

```go
// Export
sarifJSON, err := report.ToSARIF()

// Parse SARIF from another tool
findings, err := finding.FindingsFromSARIF(sarifJSON)
```

Round-trip fidelity is preserved:

- `SeverityCritical` maps to SARIF `"error"` (no critical level in SARIF 2.1.0); the original severity is stored in `Properties["go-finding/severity"]`
- `Finding.Snippet` round-trips via `region.snippet`
- `RelatedRef.Range` end positions round-trip via related location regions
- Non-standard metadata preserved in the property bag with `go-finding/*` prefix

## LSP Diagnostics

```go
lspDiag := f.ToLSP()

// Related information includes proper LSPRange when RelatedRef.Range is set
for _, rel := range lspDiag.RelatedInformation {
    fmt.Println(rel.Range.Start.Line, rel.Range.End.Line)
}

// From LSP diagnostic
f := finding.FromLSP("file:///path/to/file.go", lspDiag)
```

LSP diagnostic tags (`Unnecessary`, `Deprecated`) are preserved in `Finding.Metadata["go-finding/lsp-diagnostic-tags"]`. Related information end positions reconstruct `RelatedRef.Range`.

## go/analysis Integration

```go
// From go/analysis Diagnostic (in analysis subpackage)
f := analysis.FromDiagnostic(diag, pass.Fset, "my-analyzer", "RULE001")

// Note: Converting back to analysis.Diagnostic is supported via ToDiagnostic().
```

## JSON

```go
// Serialize a single finding
data, err := f.LineJSON()

// Deserialize with validation (returns value type)
f, err := finding.FromJSON(data)

// Line-delimited JSON stream
line, err := f.LineJSON()

// Pretty-printed report
data, err := report.PrettyJSON()
```

## Error Handling

Structured errors with categories:

```go
err := finding.NewValidationError("invalid severity", nil)
err := finding.NewIOError("read file", cause).WithPosition(pos)
err := finding.NewConflictError("overlapping fixes", cause)

finding.IsFindingError(err)
finding.CategoryOf(err) // "validation", "io", "conflict", etc.
```

## CLI

```bash
go install github.com/larsartmann/go-finding/cmd/go-finding@latest

go-finding -format sarif -output results.sarif
go-finding -format json -config config.yaml
go-finding -format csv -output findings.csv
go-finding -format markdown -min-severity warning
go-finding -filter-generated -fix-provider go-ast
```

Key flags: `-format` (text/markdown/csv/tsv/json/sarif), `-min-severity`, `-config`, `-filter-generated`, `-fix-provider`, `-byte-level-conflict`. Use `-help` for the full list.

## Development

```bash
nix run .#test                     # Run tests (all modules)
nix run .#test-race                # Run tests with race detector
nix run .#bench                    # Run benchmarks
nix run .#lint                     # Lint
```

> `encoding/json/v2` is stable since Go 1.27 — no `GOEXPERIMENT` needed anywhere.

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

`go-finding` is **MIT-licensed** and maintained on a **best-effort** basis by a
single author.

- **Bugs & feature requests:** open a [GitHub Issue](https://github.com/larsartmann/go-finding/issues).
- **No SLA.** Issues and PRs are reviewed as time permits.
- **Security vulnerabilities:** see [SECURITY.md](SECURITY.md) for private reporting.
- **Supported versions:** the latest minor release only. The API has been frozen
  since `v1.0.0`; breaking changes require a major version bump.
- **Community:** be kind and constructive — see [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The API has been frozen since `v1.0.0` (2026-06-24). Breaking changes require a major version bump.

The current version is available programmatically:

```go
fmt.Println(finding.Version) // "1.15.0"
```

## Documentation

| Document                                                                         | Purpose                                                  |
| -------------------------------------------------------------------------------- | -------------------------------------------------------- |
| [FEATURES.md](FEATURES.md)                                                       | Honest feature inventory with status                     |
| [ROADMAP.md](ROADMAP.md)                                                         | Long-term direction and future ideas                     |
| [TODO_LIST.md](TODO_LIST.md)                                                     | Short-term actionable tasks                              |
| [CHANGELOG.md](CHANGELOG.md)                                                     | Versioned change history                                 |
| [docs/USAGE_GUIDE.md](docs/USAGE_GUIDE.md)                                       | Comprehensive usage guide                                |
| [docs/MIGRATION_v1.0.md](docs/MIGRATION_v1.0.md)                                 | v1.0 migration instructions                              |
| [docs/guides/consumer-migration-v1.7.md](docs/guides/consumer-migration-v1.7.md) | Upgrade guide: outcomes, rollback default, groups        |
| [docs/guides/outcomes.md](docs/guides/outcomes.md)                               | Per-finding fix outcomes, rollback semantics, metrics    |
| [docs/guides/fix-engine.md](docs/guides/fix-engine.md)                           | Fix engine patterns (providers, edits, conflicts)        |
| [docs/guides/fix-providers.md](docs/guides/fix-providers.md)                     | Writing custom fix providers                             |
| [docs/guides/finding-groups.md](docs/guides/finding-groups.md)                   | Grouping related findings via SARIF/LSP                  |
| [docs/guides/flight-recorder.md](docs/guides/flight-recorder.md)                 | Pipeline trace flight recorder                           |
| [docs/guides/configuration.md](docs/guides/configuration.md)                     | CLI flags, YAML/JSON config, precedence                  |
| [docs/guides/troubleshooting.md](docs/guides/troubleshooting.md)                 | Common errors and fixes                                  |
| [docs/ecosystem.md](docs/ecosystem.md)                                           | How go-finding relates to the surrounding SDKs and tools |

## Related Projects

go-finding is the hub of an ecosystem of SDKs and tools. See [`docs/ecosystem.md`](docs/ecosystem.md) for the full architecture diagram and component comparison.

**Ecosystem SDKs** (shared plumbing for tools that emit findings):

- go-linter-sdk — Rule + Registry scaffolding for linters
- linter-autoconfigure-sdk — Config round-trip + finding emission for auto-configurers
- go-checker-helpers — Finding builders, fix pipeline, and safe I/O for BuildFlow checkers

**Tools in the ecosystem** (several consume go-finding today; see
[`docs/ecosystem.md`](docs/ecosystem.md) for per-tool status):

- [art-dupl](https://github.com/larsartmann/art-dupl) — Code duplication detection (go-finding integration designed, GroupID wiring pending)
- branching-flow — Go code quality analyzer
- erraudit (formerly hierarchical-errors) — Error handling pattern detector
- go-auto-upgrade — Dependency upgrade automation

**Standards:**

- [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) — Static Analysis Results Interchange Format
- [LSP](https://microsoft.github.io/language-server-protocol/) — Language Server Protocol

## License

[MIT](LICENSE)
