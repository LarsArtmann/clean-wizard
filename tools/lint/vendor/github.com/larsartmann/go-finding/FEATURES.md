# FEATURES.md — go-finding

> A unified data model and pipeline for Go static analysis tools.
> Seven tools detect issues. Zero tools route them to remediation. This library fixes that.

---

## Status Legend

| Status               | Meaning                                                     |
| -------------------- | ----------------------------------------------------------- |
| FULLY_FUNCTIONAL     | Fully implemented, tested, production-ready                 |
| PARTIALLY_FUNCTIONAL | Works but has known limitations or rough edges              |
| EXPERIMENTAL         | Implemented and tested, but API may change                  |
| PLANNED              | Type/constant exists, tested, but no backend implementation |
| BROKEN               | Mentioned in docs/comments but no code exists               |

---

## 1. Core Data Model

### 1.1 Finding Type

**Status:** FULLY_FUNCTIONAL

The central type representing a single issue detected by a static analysis tool.

| Field       | Type                | Purpose                                                                                                                      |
| ----------- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| ID          | `ID`                | Stable unique identifier (`tool:rule:file:line:col`; hash-based `tool:rule:<hash>` for file-level findings with `Line == 0`) |
| Rule        | `RuleName`          | Rule/check name (e.g., `STRONG_ID`)                                                                                          |
| ToolName    | `ToolName`          | Source tool name (e.g., `govet`)                                                                                             |
| Message     | `string`            | Human-readable description                                                                                                   |
| Severity    | `Severity`          | info / warning / error / critical                                                                                            |
| Position    | `Position`          | Where the issue is (file, line, column, offset)                                                                              |
| Category    | `Category`          | Domain classification (security, style, etc.)                                                                                |
| Tags        | `[]Tag`             | Multiple classification labels                                                                                               |
| FixStrategy | `FixStrategy`       | none / suggest / direct / ai                                                                                                 |
| Suggestion  | `string`            | Human-readable fix description                                                                                               |
| BeforeCode  | `string`            | Code before the fix                                                                                                          |
| AfterCode   | `string`            | Code after the fix                                                                                                           |
| Range       | `*Range`            | Span-based findings (start/end positions)                                                                                    |
| Snippet     | `string`            | Surrounding code context                                                                                                     |
| Confidence  | `Confidence`        | Named type, 0.0–1.0 scale                                                                                                    |
| GroupID     | `GroupID`           | Logical group this finding belongs to (e.g., clone group)                                                                    |
| Related     | `[]RelatedRef`      | Related findings with optional `*Range` span                                                                                 |
| Suppression | `*Suppression`      | If suppressed                                                                                                                |
| Metadata    | `map[string]string` | Tool-specific key-value pairs                                                                                                |
| Edits       | `[]TextEdit`        | Typed multi-edit fix list (authoritative when set; `BeforeCode`/`AfterCode` become the first-edit display summary) (v1.14.0) |

Key methods: `Validate()` (decomposed into 6 per-field validators for low complexity), `IsValid()`, `Clone()`, `Key()`, `Equal()`, `String()`, `Preview()`, `HasFix()`, `HasSuggestion()`, `IsSuppressed()`, `NormalizedConfidence()`

### 1.2 Builder API

**Status:** FULLY_FUNCTIONAL

Fluent builder for constructing `Finding` values with validation.

```go
f, err := NewBuilder(RuleName("nilcheck"), ToolName("govet"), "possible nil deref", SeverityError, Pos("main.go", 42, 5)).
    WithFixStrategy(FixStrategyDirect).
    WithBeforeCode("x.foo").
    WithAfterCode("x.foo()").
    Build()
```

- `NewBuilder(rule RuleName, toolName ToolName, message, severity, pos)` — required fields (branded types prevent ID/Rule/Tool mixups at compile time)
- `WithID()`, `WithCategory()`, `WithTags()` (appends), `WithFixStrategy()`, `WithSuggestion()`, `WithBeforeCode()`, `WithAfterCode()`, `WithRange()`, `WithSnippet()`, `WithGroupID()`, `WithConfidence()`, `WithRelated()` (appends), `WithSuppression()`, `WithMetadata()` — optional
- `Build()` — returns validated `Finding` or error
- `BuildOrDefault()` — returns validated `Finding` or zero-value `Finding{}` on error (v1.3.0)
- `MustBuild()` — panics on invalid state
- Default confidence: `ConfidenceFull` (1.0) since v1.3.0; override with `.WithConfidence()`

### 1.3 Template Factory

**Status:** FULLY_FUNCTIONAL

Pre-configured builder factory for stamping common fields (tool name, category, fix strategy, tags) once and building many findings:

```go
tmpl := NewTemplate("my-linter").
    WithCategory(CategoryStyle).
    WithFixStrategy(FixStrategySuggest)
f1 := tmpl.Build("R1", "msg 1", SeverityInfo, Pos("a.go", 1, 1))
```

`Template.Builder(rule, msg, sev, pos) *Builder` returns a pre-configured `*Builder` for per-finding chaining (confidence, suggestion, before/after code) before the terminal `Build()`/`MustBuild()`/`BuildOrDefault()` call. Eliminates the `makeFindingWithConfidence` / `buildFixableFinding` factory patterns consumers reinvent when they need both template-level defaults AND per-finding overrides.

---

## 2. Position & Range

**Status:** FULLY_FUNCTIONAL

### 2.1 Position

| Field  | Type       | Notes                                                  |
| ------ | ---------- | ------------------------------------------------------ |
| File   | `FilePath` | Required (branded type — string literals auto-convert) |
| Line   | `int`      | 1-based; 0 = not set                                   |
| Column | `int`      | 1-based; 0 = not set                                   |
| Offset | `int`      | 0-based byte offset; -1 = not set                      |

Methods: `IsValid()`, `HasFile()`, `Equal()`, `Compare()`, `String()`, `HasOffset()`, `IsZero()`, `HasLocation()`
Constructors: `Pos(file, line, column)`, `FilePos(file)` (v1.3.0 — for file-level findings with no line number)

> **v1.3.0 change:** `Finding.Validate()` and `Finding.IsValid()` now accept file-level positions (File set, Line=0). Previously required Line>0 via `Position.IsValid()`. Now uses `Position.HasFile()` (File != ""). Use `FilePos("config.yaml")` for config-level or project-level findings.

### 2.2 Range

Span from `Start` to `End` position.

Methods: `IsValid()`, `HasEnd()`, `LineCount()`, `Length()`, `Equal()`, `Compare()`, `Contains(Position)`, `Overlaps(Range)`, `Intersection(Range)`, `Adjacent(Range)`

Constructors: `NewRange(file, startLine, startCol, endLine, endCol)`, `NewRangePtr(...)`

---

## 3. Severity

**Status:** FULLY_FUNCTIONAL

Four levels, ordered by urgency:

| Level    | String       | SARIF Mapping                             |
| -------- | ------------ | ----------------------------------------- |
| Info     | `"info"`     | `note`                                    |
| Warning  | `"warning"`  | `warning`                                 |
| Error    | `"error"`    | `error`                                   |
| Critical | `"critical"` | `error` (lossy — preserved in Properties) |

Methods: `IsValid()`, `Compare()`, `GreaterThan()`, `LessThan()`, `GreaterThanOrEqual()`, `LessThanOrEqual()`, `Badge()`, `Emoji()`, `PriorityString()` (v1.3.0)

Parsing: `ParseSeverity(s)`, `MustParseSeverity(s)`, `SeverityFromLevel(level, fallback)` (v1.3.0 — maps common aliases, returns fallback for unknown)

11 pre-registered aliases: `warn`, `high`, `medium`, `low`, `fatal`, `critical`, `crit`, `note`, `advice`, `optional`, `suggestion` (extensible via `RegisterSeverityAlias()`)

---

## 4. Fix Strategy

**Status:** FULLY_FUNCTIONAL (except AI — see below)

| Strategy | String      | Auto-apply | Description                    |
| -------- | ----------- | ---------- | ------------------------------ |
| None     | `"none"`    | No         | No fix available               |
| Suggest  | `"suggest"` | No         | Human-readable suggestion      |
| Direct   | `"direct"`  | Yes        | Automatically applicable       |
| AI       | `"ai"`      | No         | **PLANNED** — needs AI backend |

Methods: `IsValid()`, `CanAutoApply()`, `NeedsAI()`

> **Honest assessment:** `FixStrategyAI` is a placeholder. `NeedsAI()` returns `true` for it, but no AI backend exists. The pipeline triage groups `ai` with `suggest` (no auto-apply). Safe to use as a marker for future AI integration.

---

## 5. Confidence, Category & Tags

### 5.1 Confidence

**Status:** FULLY_FUNCTIONAL

Named float64 type with range [0.0, 1.0]. Five canonical levels:

| Level  | Value | String     |
| ------ | ----- | ---------- |
| None   | 0.0   | `"none"`   |
| Low    | 0.25  | `"low"`    |
| Medium | 0.5   | `"medium"` |
| High   | 0.75  | `"high"`   |
| Full   | 1.0   | `"full"`   |

Methods: `IsValid()`, `Clamp()`, `Compare()`, `String()`

Parsing: `ParseConfidence(string) (Confidence, error)` — inverse of `String()`. Accepts named levels, decimal strings (e.g. `"0.42"`), and empty string (defaults to `ConfidenceLow`). Case-insensitive, trims whitespace. Returns `ErrInvalidConfidence` sentinel, matchable via `errors.Is`.

### 5.2 Category

**Status:** FULLY_FUNCTIONAL

16 predefined domain categories:

`security`, `style`, `performance`, `correctness`, `complexity`, `duplication`, `error-handling`, `migration`, `type-safety`, `structure`, `configuration`, `documentation`, `testing`, `unused`, `best-practice`, `naming`

Plus arbitrary custom categories accepted. Methods: `IsStandard()`, `IsValid()`, `String()`, `Compare()`, `IsSecurity()`. Parsing: `ParseCategory(s)` (returns error), `MustParseCategory(s)` (panics).

### 5.3 Tags

**Status:** FULLY_FUNCTIONAL

Multi-label classification for richer filtering:

`security`, `performance`, `style`, `correctness`, `bug`, `deprecated`, `documentation`, `complexity`, `test`, `build`

Methods: `IsStandard()`, `IsValid()`, `String()`

Plus arbitrary custom tags accepted.

---

## 6. Suppression

**Status:** FULLY_FUNCTIONAL

Mark findings as suppressed with reason and optional expiry.

| Field     | Type              | Purpose                               |
| --------- | ----------------- | ------------------------------------- |
| Kind      | `SuppressionKind` | `in-source`, `in-config`, `in-review` |
| Rule      | `RuleName`        | Which rule is suppressed              |
| Reason    | `string`          | Why                                   |
| ExpiresAt | `*time.Time`      | Optional TTL                          |

Methods: `IsExpired(now)`, `IsValid()`, `IsActive(now)`, `SuppressionKind.IsValid()`

> `IsActive(now)` is a convenience combining `IsValid() && !IsExpired(now)`.

> **Note:** The library stores and checks suppression data. It does NOT parse `//nolint` or `//lint:ignore` directives — that is the caller's responsibility.

---

## 7. Report Container

**Status:** FULLY_FUNCTIONAL

Top-level container for a tool run.

| Feature                 | Method                                                         | Thread-safe  |
| ----------------------- | -------------------------------------------------------------- | ------------ |
| Create                  | `NewReport(toolInfo)`                                          | Yes          |
| Create from findings    | `NewReportFromFindings(tool, []F)`                             | Yes (v1.3.0) |
| Add single finding      | `AddFinding(f)`                                                | Yes (mutex)  |
| Add multiple findings   | `AddFindings([])`                                              | Yes (mutex)  |
| Recompute stats         | `ComputeSummary()`                                             | Yes (mutex)  |
| Filter by severity      | `BySeverity(sev)`                                              | Read-only    |
| Filter by category      | `ByCategory(cat)`                                              | Read-only    |
| Filter by fix strategy  | `ByFixStrategy(fs)`                                            | Read-only    |
| Find by ID              | `FindByID(id)` → `*Finding` (shallow copy; `Clone()` for deep) | Read-only    |
| Find by rule            | `FindByRule(rule)`                                             | Read-only    |
| Active (non-suppressed) | `ActiveFindings()`                                             | Read-only    |
| Generic filter          | `Filter(predicates...)` → `*Report`                            | Read-only    |
| Transform               | `Map(func) → *Report`                                          | Read-only    |
| Iterate                 | `All()` → `iter.Seq[Finding]`                                  | Read-only    |
| Count                   | `Len()`                                                        | Read-only    |

Summary stats: `Total`, `BySeverity`, `ByCategory`, `ByFixStrategy`, `FilesAffected`, `FilesScanned` (always serialized — "scanned 0" is coverage evidence, not an omission), `SkippedModules` (nested go.mod modules a tool deliberately did not analyze), `Suppressed`

---

## 8. Filtering & Sorting

**Status:** FULLY_FUNCTIONAL

### 8.1 Predicates

Composable filter functions:

|`BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFixStrategy`, `ByTool`, `ByRule`, `ByFile`, `NotSuppressed`, `WithFix`, `WithSuggestion`, `Negate`

### 8.2 Core Operations

| Operation | Function |
| --------- | --------------------------------- | ---------------------------------------- | -------------------- |
| Filter | `Filter(findings, predicates...)` |
| | In-place filter | `FilterInPlace(findings, predicates...)` | GC-safe: zeroes tail |

### 8.3 Grouping

`GroupBy(findings, keyFn)`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`

### 8.4 Sorting

`SortByPosition(findings)` — file, line, column
`SortBySeverity(findings)` — most severe first

---

## 9. Report Merging & Deduplication

**Status:** FULLY_FUNCTIONAL

### 9.1 Merge

Combines multiple reports into one with optional deduplication.

```go
merged := finding.Combine(reports,
    finding.WithDeduplication(true),
    finding.WithDeduplicateBy(finding.DeduplicateByID),
)
```

### 9.2 Deduplication Strategies

| Strategy                | Match By                    |
| ----------------------- | --------------------------- |
| `DeduplicateByID`       | Exact ID match (default)    |
| `DeduplicateByPosition` | tool + file + line + column |
| `DeduplicateByRule`     | rule + file + line + column |

### 9.3 Cross-Tool Correlation

**Status:** FULLY_FUNCTIONAL

Finds related findings across tools using two strategies depending on data shape:

- **Range-based findings:** `IntervalIndex[T]` provides O(n + k) overlap queries via a sorted-slice index (not a tree) (`interval_index.go`)
- **Point-only findings:** line-proximity heuristic (same file, within 5 lines)
- **Mixed:** cross-correlates range and point findings in the same file

```go
correlations := finding.Correlate(allFindings)
```

Returns `[]Correlation` with `FindingIDs`, `Reason`, and `Score CorrelationScore`. Capped at 10,000 correlations. Complexity: O(n log n) index build + O(n + k) per query for range findings; O(n) per file for point findings.

> **Honest scope:** Spatial + proximity heuristics only — no semantic analysis. Good for surface-level grouping; not a substitute for rule-level deduplication.

---

## 10. ID Generation & Parsing

**Status:** FULLY_FUNCTIONAL

### Generation

`GenerateID(toolName, rule, pos)` produces:

- `"tool:rule:file:line:col"` when line > 0 and column > 0 (human-readable; column omitted when 0)
- `"tool:rule:<sha256hash>"` when line == 0 (hash-based)

### Parsing

`ParseID(id)` → `ParsedID{Tool, Rule, File, Line, Column}`

`IsHashID(id)` → checks if hash-based

Handles Windows paths with colons correctly.

---

## 11. JSON Serialization

**Status:** FULLY_FUNCTIONAL

> **Deterministic since v1.5.0:** All JSON/SARIF marshal calls pass `json.Deterministic(true)`, ensuring Go map keys serialize in sorted order on every run. Output is safe for snapshot/diff testing without post-hoc normalization. 8 byte-identity regression tests guard this.

| Operation        | Function                                                 | Notes                        |
| ---------------- | -------------------------------------------------------- | ---------------------------- |
| Report → JSON    | `report.PrettyJSON()`                                    | Pretty-printed               |
| Report → Writer  | `report.WriteJSON(w)`                                    | Streaming, avoids allocation |
| JSON → Report    | `ReportFromJSON(data)` → `(*Report, dropped, error)`     | Drops invalid findings       |
| Finding → JSON   | `finding.LineJSON()`                                     | Compact single-line          |
| Finding → Writer | `finding.WriteJSON(w)`                                   | Streaming                    |
| JSON → Finding   | `FromJSON(data)` → `(Finding, error)`                    | Returns value, validates     |
| JSON → []Finding | `FindingsFromJSON(data)` → `([]Finding, dropped, error)` | Drops invalid                |

---

## 12. SARIF 2.1.0 Interchange

**Status:** FULLY_FUNCTIONAL

### Export

| Method                                       | Description                                                           |
| -------------------------------------------- | --------------------------------------------------------------------- |
| `report.ToSARIF()`                           | Full report → SARIF JSON (excludes suppressed by default)             |
| `report.ToSARIFWithOpts(opts...)`            | Functional options: `WithIncludeSuppressed()`, `WithMinSeverity(sev)` |
| `report.WriteSARIF(ctx, w)`                  | Streaming SARIF output                                                |
| `report.WriteSARIFWithOpts(ctx, w, opts...)` | Streaming output with functional options                              |

### Import

`FindingsFromSARIF(ctx, data)` → `([]Finding, error)` — round-trip fidelity via property bag

### Round-Trip Fidelity

- All non-standard fields preserved in `properties` bag (`go-finding/*` prefix)
- `Finding.Snippet` round-trips via SARIF `region.snippet`
- `RelatedRef.Range` end positions round-trip via related location regions
- Suppressed findings round-trip when `WithIncludeSuppressed()` is used (emits SARIF `suppressions` arrays)
- `SeverityCritical` maps to SARIF `"error"` (no `"critical"` level in SARIF 2.1.0); original preserved in Properties

---

## 13. LSP Diagnostic Conversion

**Status:** FULLY_FUNCTIONAL

### Finding → LSP

`finding.ToLSP()` → `LSPDiagnostic`

Handles: severity mapping, 0-based conversion, Range, related information, diagnostic tags

`LSPDiagnosticTagUnnecessary = 1`, `LSPDiagnosticTagDeprecated = 2` constants with `Tags []LSPDiagnosticTag`. Tags round-trip via `Metadata["go-finding/lsp-diagnostic-tags"]`, and `ToLSP()` re-emits them from that metadata key (negative values rejected as invalid), so tags survive repeated conversions. Related information includes proper end positions from `RelatedRef.Range`.

### LSP → Finding

`FromLSP(fileURI, diag)` → `Finding`

Preserves: end position as Range, related information, related range end positions, diagnostic tags in Metadata, raw LSP severity in Metadata

**Full round-trip fidelity via `LSPDiagnosticData`:** When `ToLSP()` populates `diag.Data`, `FromLSP()` restores: ID, Severity (including SeverityCritical), FixStrategy, Confidence, Category, GroupID, Tags, BeforeCode, AfterCode, Suggestion, Snippet, Suppression, Metadata, and RelatedRef.FindingID (preserved, not regenerated).

---

## 14. go/analysis Integration

**Status:** FULLY_FUNCTIONAL

| Function                                                      | Purpose                              |
| ------------------------------------------------------------- | ------------------------------------ |
| `FromDiagnostic(diag, fset, toolName, ruleCode, ...severity)` | `go/analysis.Diagnostic` → `Finding` |
| `ToDiagnostic(finding, fset)`                                 | `Finding` → `go/analysis.Diagnostic` |
| `FromTokenPosition(pos)`                                      | `token.Position` → `Position`        |
| `NodePosition(fset, node)`                                    | `ast.Node` → `Position`              |
| `NodeRange(fset, node)`                                       | `ast.Node` → `Range`                 |
| `FormatDiagnostic(diag, fset, name)`                          | Go-vet-style formatted string        |

Auto-detects suggested fixes and sets `FixStrategyDirect` with `AfterCode`. `ToDiagnostic` converts back with position resolution, SuggestedFix generation, and Related conversion.

### 14.1 AnalyzerDetector

**Status:** FULLY_FUNCTIONAL

`analysis.AnalyzerDetector` wraps any `go/analysis.Analyzer` as a `finding.Detector`, running it against Go source and converting its diagnostics to Findings:

```go
det := analysis.NewAnalyzerDetector(analyzer, []string{"./..."},
    analysis.WithSeverity(finding.SeverityWarning),
    analysis.WithFileSet(fset),
)
findings, err := det.Detect(ctx)
```

---

## 15. Structured Errors

**Status:** FULLY_FUNCTIONAL

Category-based error types with `errors.Is` / `errors.As` support:

| Error           | Category     | Factory                          |
| --------------- | ------------ | -------------------------------- |
| `ErrValidation` | `validation` | `NewValidationError(msg, cause)` |
| `ErrIO`         | `io`         | `NewIOError(msg, cause)`         |
| `ErrParse`      | `parse`      | `NewParseError(msg, cause)`      |
| `ErrConflict`   | `conflict`   | `NewConflictError(msg, cause)`   |
| `ErrInternal`   | `internal`   | `NewInternalError(msg, cause)`   |

`FindingError` supports: `WithFinding()`, `WithPosition()`, `Unwrap()`, `Is()` for sentinel matching, `ErrorCode()` (returns `"finding.<category>"`), `ErrorFamily()` (maps category to `go-error-family` families for `errorfamily.Classify()` integration)

Helpers: `IsFindingError(err)`, `CategoryOf(err)`, `IsCategory(err, cat)`

---

## 16. Pipeline

**Status:** FULLY_FUNCTIONAL

The pipeline orchestrates a detect → triage → fix → verify loop.

### 16.1 Detector Interface

```go
type Detector interface {
    Name() string
    Detect(ctx context.Context) ([]finding.Finding, error)
}
```

Adapters: `DetectorFunc`, `NamedDetectorFunc(name, fn)`

### 16.2 Configuration

| Option                       | Type                   | Default | Description                                 |
| ---------------------------- | ---------------------- | ------- | ------------------------------------------- |
| `MaxIterations`              | `int`                  | 5       | Prevents infinite loops                     |
| `ParallelDetectors`          | `bool`                 | `true`  | Concurrent detector execution               |
| `Timeout`                    | `time.Duration`        | 10min   | Pipeline timeout                            |
| `VerifyAfterFix`             | `bool`                 | `false` | Re-run detectors post-fix                   |
| `GracefulDegradation`        | `bool`                 | `false` | Continue on detector failures               |
| `DryRun`                     | `bool`                 | `false` | Detect + triage only (no fixes)             |
| `Retry`                      | `*RetryConfig`         | `nil`   | Exponential backoff retries                 |
| `Metrics`                    | `*Metrics`             | `nil`   | Timing/count collection                     |
| `CorrelateFindings`          | `bool`                 | `false` | Cross-tool correlation                      |
| `Processors`                 | `[]FindingTransformer` | `nil`   | Composable finding transforms               |
| `FixProviders`               | `[]FixProvider`        | `nil`   | Custom fix providers (e.g., AST)            |
| `OnFinding`                  | `func(Finding)`        | `nil`   | Per-finding callback                        |
| `OnFix`                      | `func(Finding, bool)`  | `nil`   | Per-fix callback                            |
| `OnIteration`                | `func(int, []Finding)` | `nil`   | Per-iteration callback                      |
| `StageHooks`                 | `[]StageHook`          | `nil`   | Per-stage before/after hooks with abort     |
| `FixRollbackAllFiles`        | `bool`                 | `false` | All-or-nothing rollback (default: per-file) |
| `DetectorTimeouts`           | `map[string]Duration`  | `nil`   | Per-detector timeout overrides              |
| `Logger`                     | `*slog.Logger`         | `nil`   | Structured logging                          |
| `TriageFunc`                 | `TriageFunc`           | `nil`   | Custom triage categorization                |
| `ByteLevelConflictDetection` | `bool`                 | `false` | Byte-level (vs position) conflict detection |

Config validation: `config.Validate()` returns joined errors for invalid values. `pipeline.New()` rejects invalid configs.

### 16.3 Pipeline Loop

1. **Detect** — Run detectors (parallel or sequential)
2. **Transform** — Run `FindingTransformer` chain on raw findings
3. **Triage** — Categorize by `FixStrategy` (direct / suggest / none)
4. **Apply** — Apply direct fixes with conflict detection
5. **Repeat** — Until stable (zero findings) or `MaxIterations`

### 16.4 Finding Processors

**Status:** FULLY_FUNCTIONAL

Composable transforms that run on findings between detection and triage.

```go
type FindingTransformer interface {
    Name() string
    Transform(ctx context.Context, findings []finding.Finding) ([]finding.Finding, error)
}
```

Adapters:

- `TransformerFunc(fn)` — wraps a function as a `FindingTransformer` (name: `""`)
- `NamedTransformerFunc(name, fn)` — wraps with a custom name

Processors are executed in order from `Config.Processors`. Use cases: filtering, enrichment, normalization, severity adjustment, deduplication.

### 16.5 Pipeline Result

| Field             | Type                    | Description                                                                           |
| ----------------- | ----------------------- | ------------------------------------------------------------------------------------- |
| `Stable()`        | method → `bool`         | Reached zero findings (`Reason == ReasonStable`)                                      |
| `TotalIterations` | `int`                   | Iterations executed                                                                   |
| `Iterations`      | `[]Iteration`           | Per-iteration details                                                                 |
| `TotalDetected`   | `int`                   | Total findings across iterations                                                      |
| `Verification`    | `*VerifyResult`         | Post-fix verification                                                                 |
| `PartialErrors`   | `map[string]error`      | Per-detector failures                                                                 |
| `Correlations`    | `[]finding.Correlation` | Cross-tool correlations                                                               |
| `Outcomes`        | `[]FixOutcome`          | Per-finding fix outcomes, deduped to the first outcome per finding identity (v1.12.0) |
| `Metrics`         | `MetricsSnapshot`       | Timing and counts                                                                     |

### 16.6 Conflict Detection

**Status:** FULLY_FUNCTIONAL

- Detects overlapping fixes in the same file
- `FilterConflictingFixes()` — returns only non-conflicting fixes
- `AnalyzeConflicts()` — detailed `Conflict` with reasons
- When multiple fixes overlap in same group, keeps first, marks rest as conflicts

### 16.7 Fix Application

**Status:** PARTIALLY_FUNCTIONAL

Byte-level fix engine with composable provider architecture.

#### FixEdit — Byte-Level Edit Operations

```go
type FixEdit struct {
    Offset      int            // 0-based byte offset
    Length      int            // bytes to remove (0 = insert)
    Replacement []byte         // bytes to write
    Source      finding.Finding
}
```

Methods: `EndOffset()`, `IsInsert()`, `IsDelete()`, `Overlaps(FixEdit)`, `Validate()`

Edits are applied in descending offset order with a frontier boundary to prevent overlapping writes.

#### FixProvider Interface

```go
type FixProvider interface {
    Name() string
    CanHandle(f finding.Finding) bool
    Edits(content []byte, f finding.Finding) ([]FixEdit, error)
}
```

**Default provider chain (tried in order):**

| Provider            | Name            | Handles                                                                                                                                                                                               |
| ------------------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EditListProvider`  | `"edit-list"`   | `HasEditList()` findings: byte offsets directly, line/col via the shared line index; cross-file lists fail `ErrEditCrossFile`, stale offsets fail `ErrEditStale` (loud, never half-applies) (v1.14.0) |
| `OffsetProvider`    | `"byte-offset"` | `HasCodeChange()` and `Range != nil` with `Length() > 0`                                                                                                                                              |
| `LineProvider`      | `"line-column"` | `HasCodeChange()` and `Position.Line > 0`                                                                                                                                                             |
| `SubstringProvider` | `"substring"`   | Fallback: `HasCodeChange()` and `BeforeCode != ""`                                                                                                                                                    |

Domain-specific providers (Go AST, Rust syn, etc.) can be registered via:

- `NewFixEngineWithProviders(providers...)` — standalone engine
- `Config.FixProviders` — pipeline integration
- `NewFixApplierWithProviders(rootDir, providers...)` — direct applier

> **Known limitation:** `SubstringProvider` uses substring matching — ambiguous when the same text appears multiple times. Strongly prefer domain-specific providers for production use.

#### FixEngine (in-memory)

`NewFixEngine()` provides pure `Apply(content []byte, fixes []finding.Finding) ([]byte, []finding.Finding, int)` that transforms byte content without filesystem access. Delegates to providers, sorts edits descending by offset, applies with overlap protection.

`ApplyWithOutcomes(content, fixes)` returns a `FixApplyResult` with one `FixOutcome` per input finding: `applied`, `no-change`, `refused` (provider matched but produced zero edits), `conflict`, `invalid` (edit dropped as invalid/out of bounds), or `failed` (provider error). Applied counts are per FINDING, not per edit: a multi-edit fix counts once (v1.14.0). Helpers: `OutcomeFor(id)`, `OutcomeCounts()`, `HasErrors()`. `FixOutcome` and `FixApplyResult` marshal to deterministic JSON (`errors` carried as message strings); failed outcomes carry a typed `*finding.FindingError` with the finding's position, so consumers can classify via `errorfamily`.

#### FixApplier (filesystem)

`NewFixApplier(rootDir)` applies fixes to actual files with backup/rollback support.

`ApplyWithReport(ctx, fixes)` returns an `ApplyReport`: applied count, applied findings, per-finding outcomes, shift maps, and the rolled-back file list. Soft per-finding failures (provider resolve errors, refused findings) never abort the run or roll anything back.

Rollback policy (`RollbackPolicy`):

- `RollbackPolicyFailingFile` (default) — restores only the failing file; earlier files keep their fixes
- `RollbackPolicyAllFiles` — legacy all-or-nothing: restores every file modified in the run

Set via `SetRollbackPolicy`, `Config.FixRollbackAllFiles`, or config-file `fixRollbackAllFiles`.

Both support:

- File backup before modification
- Rollback on failure (scope governed by `RollbackPolicy`)
- Context cancellation support mid-application
- Deterministic application order (sorted by path, descending by offset)

### 16.8 Verification

**Status:** FULLY_FUNCTIONAL

Re-runs all detectors after fixes and categorizes findings:

| Category      | Meaning                              |
| ------------- | ------------------------------------ |
| `Fixed`       | Original findings no longer detected |
| `Remaining`   | Original findings still present      |
| `NewFindings` | Fresh findings introduced by fixes   |
| `Modified`    | Same key, different content          |

`VerifyResult` also exposes `Resolved int` (count of resolved findings) and `Modified []finding.Finding`.

`DiffFindings(original, post)` — standalone utility for comparing finding sets.

### 16.9 Metrics

**Status:** FULLY_FUNCTIONAL

Thread-safe metrics collection:

| Metric                | Method                                     |
| --------------------- | ------------------------------------------ |
| Stage durations       | `RecordStage()`, `StageTiming()`           |
| Detector timing       | `RecordDetector()`                         |
| Findings per detector | `RecordDetector()`                         |
| Fixes applied         | `RecordFixes(count)`                       |
| Fix outcome counts    | `RecordOutcome(status)`, `OutcomeCounts()` |
| Total duration        | `TotalDuration()`                          |
| Point-in-time copy    | `Snapshot()` → `MetricsSnapshot`           |

Auto-populated on `Pipeline.Run()` via `PipelineResult.Metrics`.

### 16.10 Retry

**Status:** FULLY_FUNCTIONAL

Configurable exponential backoff with jitter for flaky detectors:

```go
config := pipeline.DefaultRetryConfig()
// MaxRetries: 3, BaseDelay: 100ms, MaxDelay: 5s
```

`NewRetryDetector(inner, config)` wraps any `Detector` with retry logic.

Validation: `RetryConfig.Validate()` checks constraints.

### 16.11 Partial Success

**Status:** FULLY_FUNCTIONAL

When `GracefulDegradation` is enabled:

- `DetectPartial()` collects findings from successful detectors
- Failed detector errors stored in `PartialResult.Errors` map
- `PipelineResult.PartialErrors` exposes per-detector failures
- `FormatPartialErrors()` formats collected errors

### 16.12 File Backup & Rollback

**Status:** FULLY_FUNCTIONAL

- Creates backup copies before file modification
- `Backup(path)`, `Restore(path)`, `RollbackAll(paths)`
- Thread-safe, can be enabled/disabled
- On fix application failure: restores current file, then rolls back all previously modified files

### 16.13 Flight Recorder (Execution Trace)

**Status:** FULLY_FUNCTIONAL

Go runtime execution trace flight recorder for pipeline observability. Continuously buffers `runtime/trace` data in memory and snapshots on demand or automatically when a stage exceeds a slow-stage threshold.

```go
hook, err := pipeline.NewFlightRecorderHook(pipeline.FlightRecorderConfig{
    SlowStageThreshold: 30 * time.Second,
    OutputDir:          "/tmp/go-finding-traces",
})
cfg := pipeline.Config{StageHooks: []pipeline.StageHook{hook}}
defer hook.Close()
```

| Component                       | Description                                                                                                                               |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `FlightRecorderConfig`          | `MinAge` (buffer retention, default 30s), `MaxBytes` (default 4 MiB), `SlowStageThreshold` (auto-snapshot trigger), `OutputDir`, `Logger` |
| `NewFlightRecorderHook(config)` | Creates and starts the recorder (only one active at a time)                                                                               |
| `DefaultFlightRecorderConfig()` | Sensible defaults                                                                                                                         |
| `Snapshot(ctx, reason)`         | Manual on-demand snapshot, returns file path                                                                                              |
| `Close()`                       | Stops recorder, waits for in-flight async snapshots                                                                                       |

Implements `StageHook` — `OnStageEvent` **never returns an error** (trace collection is purely diagnostic and must not affect pipeline control flow).

CLI flags: `-trace` (enable), `-trace-dir` (output directory), `-trace-slow` (auto-snapshot threshold, e.g. `30s`). Config file `flightRecorder` section: `enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes` (full parity with `pipeline.FlightRecorderFileConfig`). See [Flight Recorder Guide](docs/guides/flight-recorder.md).

---

## 17. Built-in Detectors

### 17.1 Go Vet Detector

**Status:** PARTIALLY_FUNCTIONAL

Runs `go vet -json ./...` and converts JSON output to Findings.

- Category: `correctness`
- Severity: `warning`
- FixStrategy: `suggest`
- Handles non-zero exit codes (still parses output)

### 17.2 Staticcheck Detector

**Status:** PARTIALLY_FUNCTIONAL

Runs `staticcheck -f json ./...` and converts JSON output to Findings.

- Severity: `error` or `warning` based on staticcheck output
- FixStrategy: `suggest`
- Confidence: `0.8`
- Category mapping: `SA*`→correctness (takes precedence), S/Q→style, U→unused, P/R/F→performance, default→correctness

> **Note:** Both detectors require the respective tools to be installed and available in `$PATH`.

---

## 18. CLI Tool

**Status:** PARTIALLY_FUNCTIONAL

Binary: `go-finding`

### Flags

| Flag                      | Default | Description                                                                 |
| ------------------------- | ------- | --------------------------------------------------------------------------- |
| `-dir`                    | `.`     | Root directory to analyze                                                   |
| `-format`                 | `text`  | Output format: `text`, `markdown`, `csv`, `tsv`, `json`, `sarif`            |
| `-min-severity`           | `info`  | Minimum severity filter (deprecated alias: `-severity`)                     |
| `-max-iterations`         | `1`     | Pipeline iterations                                                         |
| `-parallel`               | `true`  | Run detectors in parallel                                                   |
| `-verify`                 | `false` | Re-run detectors after fixes                                                |
| `-timeout`                | `10m`   | Pipeline timeout                                                            |
| `-config`                 | (none)  | YAML/JSON config file                                                       |
| `-output`                 | (none)  | Write output to file (default: stdout)                                      |
| `-version`                | `false` | Print version and exit                                                      |
| `-cpuprof`                | (none)  | CPU profile output                                                          |
| `-memprof`                | (none)  | Memory profile output                                                       |
| `-filter-generated`       | `false` | Filter out findings from auto-generated files                               |
| `-filter-generated-types` | `all`   | Comma-separated generator types (sqlc, templ, mockgen, protobuf, …)         |
| `-generated-exclude`      | (none)  | Comma-separated glob patterns to exclude from generated filtering           |
| `-generated-include`      | (none)  | Comma-separated glob patterns restricting generated-filtering scope         |
| `-byte-level-conflict`    | `false` | Enable precise byte-level conflict detection for overlapping fixes          |
| `-fix-provider`           | (none)  | Comma-separated fix provider names to enable (e.g., `go-ast`)               |
| `-fix-rollback-all`       | `false` | All-or-nothing rollback on hard file failures (default: per-file)           |
| `-include-suppressed`     | `true`  | Include suppressed findings in the report                                   |
| `-trace`                  | `false` | Enable Go execution trace flight recorder for pipeline diagnostics          |
| `-trace-dir`              | (none)  | Directory for trace snapshot files (default: temp dir)                      |
| `-trace-slow`             | `0`     | Auto-snapshot trace when a stage exceeds this duration (e.g. `30s`)         |
| `-trace-max-files`        | `0`     | Keep at most N trace snapshots; prunes oldest beyond the cap (v1.9.0)       |
| `-trace-gzip`             | `false` | Write `.trace.gz` compressed snapshots; gunzip for `go tool trace` (v1.9.0) |

### Config File (YAML/JSON)

```yaml
maxIterations: 3
parallelDetectors: true
verifyAfterFix: false
timeout: "5m"
detectorTimeouts:
  staticcheck: "30s"
  govet: "10s"
detectors:
  - name: govet
  - name: staticcheck
fixProviders:
  - go-ast
filterGenerated: true
filterGenTypes: "all"
```

### Default Behavior

Without `-config`: uses govet + staticcheck with the flag values.

### Plugin Detectors

`RegisterDetector(name, builder)` — thread-safe, allows adding custom detectors at runtime.

### Output Formats

| Format     | Description                                                                                                                              |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `text`     | Human-readable: `file:line:col [SEVERITY] rule: message` + `Suggestion:` lines (`finding.FormatTextRich` with badges/💡 is library-only) |
| `markdown` | Markdown table with auto-aligned columns (via go-output)                                                                                 |
| `json`     | Full JSON report                                                                                                                         |
| `csv`      | CSV with auto-quoting, clean data export without footer (via go-output)                                                                  |
| `tsv`      | Tab-separated, clean data export without footer (via go-output)                                                                          |
| `sarif`    | SARIF 2.1.0                                                                                                                              |

Severity-badged table output exists as a library API (`finding.FormatTable`), not as a CLI format.

Metrics summary printed to stderr when available.

---

## 19. Testing

**Status:** FULLY_FUNCTIONAL

Test categories:

- Unit tests per source file
- Integration tests (`pipeline/integration_test.go`, `cmd/go-finding/integration_test.go`)
- E2E tests (`cmd/go-finding/e2e_test.go`)
- Fuzz tests (`fuzz_test.go`, `id_fuzz_test.go`, `json_fuzz_test.go`, `lsp_fuzz_test.go`, `merge_fuzz_test.go`, `sarif_fuzz_test.go`)
- Property-based tests (`property_test.go`)
- Benchmarks (`bench_test.go`)
- Regression tests colocated with the behavior they protect (per-subject naming, e.g. `finding_validate_test.go`, `interval_tree_test.go`)

---

## 20. Examples

**Status:** PARTIALLY_FUNCTIONAL

Runnable examples:

| Example                       | Description                                 |
| ----------------------------- | ------------------------------------------- |
| `examples/basic/`             | Direct Finding construction                 |
| `examples/builder/`           | Builder API usage                           |
| `pipeline/examples/outcomes/` | Per-finding outcomes + rollback policy demo |

> **Note:** Examples have no test files (compile-only check via `example_compile_test.go`).

---

## 21. Extensibility & Advanced Features (v0.7.0+)

### 21.1 DetectorRegistry

**Status:** FULLY_FUNCTIONAL

Thread-safe named detector constructor registry for plugin-style architecture:

```go
registry := finding.NewDetectorRegistry()
registry.MustRegister("my-tool", func() finding.Detector { ... })
det, err := registry.Build("my-tool")
all, err := registry.BuildAll() // sorted by name
```

Methods: `Register`, `MustRegister`, `Build`, `BuildAll`, `Names`, `Has`. Thread-safe via RWMutex.

### 21.2 IntervalIndex[T]

**Status:** FULLY_FUNCTIONAL

Generic interval index for O(n + k) overlap queries (sorted-slice implementation):

```go
idx := finding.NewIntervalIndex(intervals)
overlaps := idx.Query(start, end)
```

Used internally by Correlate for spatial finding correlation.

### 21.3 LineShiftMap

**Status:** FULLY_FUNCTIONAL

Tracks how line numbers and columns shift after byte-level edits:

```go
shiftMap := pipeline.NewLineShiftMap(originalContent, edits)
newLine := shiftMap.ShiftedLine(originalLine)
newPos := shiftMap.ShiftedPosition(pos) // shifts Line + Column
newRange := shiftMap.ShiftedRange(r)    // shifts both endpoints
```

Column shifting applies to single-line edits (no line-count change) on the same line. Multi-line edits shift subsequent lines only.

### 21.4 MergeIter

**Status:** FULLY_FUNCTIONAL

Streaming merge via `iter.Seq[Finding]`, avoiding intermediate slice allocation:

```go
for f := range finding.MergeIter(reports, finding.WithDeduplication(true)) {
    process(f)
}
```

### 21.5 ConfigFile (Pipeline)

**Status:** FULLY_FUNCTIONAL

JSON config loading for library use:

```go
cfg, err := pipeline.ConfigFromFile(data)
detectors, err := configFile.ResolveDetectors(registry)
providers, err := configFile.ResolveProviders(providerMap)
```

### 21.6 StageHook

**Status:** FULLY_FUNCTIONAL

Per-stage before/after hooks with abort capability:

```go
config.StageHooks = []pipeline.StageHook{
    pipeline.StageHookFunc(func(_ context.Context, e pipeline.StageEvent) error {
        if e.Stage == pipeline.StageDetect {
            log.Println("detection complete")
        }
        return nil
    }),
}
```

### 21.7 GoASTProvider

**Status:** FULLY_FUNCTIONAL

AST-aware fix provider for `.go` files using `go/parser`. Disambiguates BeforeCode occurrences structurally rather than via substring matching:

```go
applier, err := pipeline.NewFixApplierWithProviders(rootDir, &goast.Provider{})
```

Separate subpackage (`pipeline/goast`) keeps `go/parser` as an opt-in within the pipeline module.

### 21.8 GeneratedFileFilter

**Status:** FULLY_FUNCTIONAL

FindingTransformer that removes findings from auto-generated Go files (sqlc, protobuf, mockgen, templ, etc.) via `gogenfilter/v3`:

```go
filter, err := pipeline.NewGeneratedFileFilter(nil)
if err != nil {
    return err
}
config.Processors = []pipeline.FindingTransformer{filter}
```

Configurable per-generator type, include/exclude patterns.

### 21.9 ToolAdapter[O]

**Status:** FULLY_FUNCTIONAL

Generic adapter converting any tool's output type to Findings:

```go
adapter := finding.NewToolAdapter("my-tool", runFunc, parseFunc, convertFunc)
```

### 21.10 CategoryForLinter / LinterRegistry

**Status:** FULLY_FUNCTIONAL

84 linter→category mappings with case-insensitive lookup:

```go
cat := finding.CategoryForLinter("gosec") // CategorySecurity
finding.RegisterLinterCategory("my-linter", finding.CategoryPerformance)
```

### 21.11 Severity Aliases / ParseSeverity

**Status:** FULLY_FUNCTIONAL

11 pre-registered severity aliases (warn, high, medium, low, fatal, critical, crit, note, advice, optional, suggestion). `SeverityFromLevel(level, fallback)` convenience function maps aliases with fallback (v1.3.0). `PriorityString()` reverse mapping (v1.3.0):

```go
sev, err := finding.ParseSeverity("warn") // SeverityWarning
```

---

## 22. Convenience APIs (v1.3.0+)

### 22.1 Template

**Status:** FULLY_FUNCTIONAL

Pre-configured builder factory: stamp tool name, category, fix strategy, and tags once, then build many findings:

```go
tmpl := NewTemplate("my-linter").
    WithCategory(CategoryStyle).
    WithFixStrategy(FixStrategySuggest)
f1 := tmpl.Build("R1", "msg 1", SeverityInfo, Pos("a.go", 1, 1))
f2 := tmpl.Build("R2", "msg 2", SeverityWarning, Pos("b.go", 2, 3))
```

`Template.Builder(rule, msg, sev, pos) *Builder` returns a pre-configured `*Builder` for per-finding chaining:

```go
f := tmpl.Builder("R3", "msg 3", SeverityWarning, pos).
    WithConfidence(ConfidenceHigh).
    WithSuggestion("use humanize.Bytes").
    MustBuild()
```

Eliminates the `newMigrationFinding` / `buildFixableFinding` / `makeFindingWithConfidence` / `IssueBuilderFactory` patterns that consumers reinvent.

### 22.2 ApplySimpleFixes

**Status:** FULLY_FUNCTIONAL

BeforeCode→AfterCode string replacement for the 80% case where consumers don't need the full pipeline FixEngine:

```go
results := ApplySimpleFixes(findingsWithDirectFixes)
for file, fileResults := range results {
    for _, r := range fileResults {
        if r.Applied { /* success */ }
    }
}
```

Groups findings by file, reads each file, applies `strings.Replace` with count=1 per finding, writes back. Skips findings without BeforeCode/AfterCode. See [Fix Engine Guide](docs/guides/fix-engine.md#simple-fixes-core-package) for full usage.

### 22.3 External Tool Helpers

**Status:** FULLY_FUNCTIONAL

Standardized helpers for the "run CLI tool → parse JSON" pattern:

```go
path, err := CheckBinary("golangci-lint")  // wraps exec.LookPath with finding error
output, err := RunCmd(ctx, "golangci-lint", "run", "--out-format", "json", "./...")
```

Both return `NewIOError` on failure for `errors.Is(err, ErrIO)` matching.

---

## 23. Tool SDK (`toolsdk` sub-module, v1.10.0+)

**Status:** FULLY_FUNCTIONAL

The `github.com/larsartmann/go-finding/toolsdk` sub-module is the stable
contract external tools implement so BuildFlow can consume them as DAG
providers without importing BuildFlow internals. A tool declares one `Spec`
(typically a package-level var) and self-registers at init:

```go
var Provider = toolsdk.Register(toolsdk.Spec{
    Name:        "branching-flow",
    Description: "Semantic analysis (14 analyzers)",
    Trigger:     toolsdk.OnGoFiles(),
    Detect:      myDetector, // a finding.Detector
})
```

- `Spec` — name, description, `Trigger`, `DependsOn`, `Inputs`, and any of
  `Detect` (`finding.Detector`), `Repair` (`Repairer`), `HealthCheck`.
  `Register` validates and panics on malformed specs (programming error).
- `Trigger` constructors: `OnGoFiles`, `OnGoModule`, `OnFiles`, `AnyLanguage`.
- `Repairer`/`RepairerFunc`/`RepairResult` — repair contract; BuildFlow
  measures fix counts by re-detection, never self-reporting.
- Registry: `All()` returns specs in registration order; test-isolation
  helpers `SnapshotForTest`/`RestoreForTest`/`ResetForTest`.
- Dry-run plumbing: `WithDryRun`/`DryRunFromContext`.

The module depends only on the core module. Evidence: `toolsdk/spec.go`,
`toolsdk/registry.go`, `toolsdk/triggers.go`, `toolsdk/dryrun.go`.

---

## Summary Matrix

| Feature                                      | Status               | Notes                                                                                                      |
| -------------------------------------------- | -------------------- | ---------------------------------------------------------------------------------------------------------- |
| Finding type                                 | FULLY_FUNCTIONAL     | Core data model with branded types (ID, RuleName, ToolName, FilePath)                                      |
| Builder API                                  | FULLY_FUNCTIONAL     | Fluent construction with validation                                                                        |
| Position & Range                             | FULLY_FUNCTIONAL     | Full spatial algebra (Contains, Overlaps, Intersection, Adjacent)                                          |
| Severity (4 levels)                          | FULLY_FUNCTIONAL     | With comparison operators                                                                                  |
| FixStrategy (none/suggest/direct)            | FULLY_FUNCTIONAL     | Production auto-fix for `direct`                                                                           |
| FixStrategy (ai)                             | PLANNED              | Constant exists, no AI backend                                                                             |
| Category (16 standard + custom)              | FULLY_FUNCTIONAL     | Domain classification                                                                                      |
| Tags (multi-label)                           | FULLY_FUNCTIONAL     | Singular Tag field removed                                                                                 |
| Suppression                                  | FULLY_FUNCTIONAL     | With TTL/expiry support                                                                                    |
| Report container                             | FULLY_FUNCTIONAL     | Thread-safe, with summary statistics                                                                       |
| Filtering & sorting                          | FULLY_FUNCTIONAL     | Composable predicates + grouping                                                                           |
| Report merging                               | FULLY_FUNCTIONAL     | 3 deduplication strategies                                                                                 |
| Cross-tool correlation                       | FULLY_FUNCTIONAL     | IntervalIndex for range findings + proximity for points; capped at 10K                                     |
| ID generation & parsing                      | FULLY_FUNCTIONAL     | Hash-based fallback, Windows path handling                                                                 |
| JSON serialization                           | FULLY_FUNCTIONAL     | Streaming support, drops invalid findings                                                                  |
| SARIF 2.1.0 export/import                    | FULLY_FUNCTIONAL     | Round-trip via property bag                                                                                |
| LSP conversion                               | FULLY_FUNCTIONAL     | Position, severity, rule, message, related ranges, diagnostic tags survive                                 |
| go/analysis integration                      | FULLY_FUNCTIONAL     | Bidirectional conversion (Diagnostic ↔ Finding)                                                            |
| AnalyzerDetector                             | FULLY_FUNCTIONAL     | Wraps `go/analysis.Analyzer` as a `Detector`                                                               |
| Structured errors                            | FULLY_FUNCTIONAL     | 5 categories, errors.Is support                                                                            |
| Pipeline (detect→fix→verify)                 | FULLY_FUNCTIONAL     | Iterative loop with configurable behavior                                                                  |
| Finding transformers                         | FULLY_FUNCTIONAL     | Composable transforms between detect and triage (FindingTransformer)                                       |
| Conflict detection                           | FULLY_FUNCTIONAL     | Overlapping fix detection                                                                                  |
| FixEdit (byte-level edits)                   | FULLY_FUNCTIONAL     | Offset, Length, Replacement with Overlaps/Validate                                                         |
| FixProvider interface                        | FULLY_FUNCTIONAL     | Composable providers: Offset, Line, Substring + custom                                                     |
| Fix application                              | PARTIALLY_FUNCTIONAL | Byte-level FixEngine + filesystem FixApplier, backup/rollback                                              |
| Verification                                 | FULLY_FUNCTIONAL     | Diff-based: fixed / remaining / new                                                                        |
| Metrics                                      | FULLY_FUNCTIONAL     | Thread-safe, snapshot support                                                                              |
| Retry (exponential backoff)                  | FULLY_FUNCTIONAL     | With jitter                                                                                                |
| Partial success                              | FULLY_FUNCTIONAL     | Graceful degradation on detector failure                                                                   |
| File backup & rollback                       | FULLY_FUNCTIONAL     | Automatic on fix failure                                                                                   |
| Go vet detector                              | PARTIALLY_FUNCTIONAL | Requires `go vet` in PATH                                                                                  |
| Staticcheck detector                         | PARTIALLY_FUNCTIONAL | Requires `staticcheck` in PATH                                                                             |
| CLI tool                                     | PARTIALLY_FUNCTIONAL | 6 output formats (text, markdown, csv, tsv, json, sarif), config, profiling                                |
| Plugin detector registry                     | FULLY_FUNCTIONAL     | Thread-safe `RegisterDetector`                                                                             |
| Per-detector timeouts                        | FULLY_FUNCTIONAL     | `DetectorTimeouts` map in Config + CLI config file                                                         |
| Structured logging (slog)                    | FULLY_FUNCTIONAL     | Optional `Logger *slog.Logger` in Config                                                                   |
| Stage hooks/callbacks                        | FULLY_FUNCTIONAL     | `StageHooks` with abort capability                                                                         |
| Diff function                                | FULLY_FUNCTIONAL     | `Diff(before, after)` by ID, `DiffResult.HasChanges()`, `Stats()`                                          |
| FormatText / FormatTextRich / FormatMarkdown | FULLY_FUNCTIONAL     | FormatText: `[SEVERITY]` tag. FormatTextRich: emoji badges, category, 💡 (v1.3.0)                          |
| Config validation                            | FULLY_FUNCTIONAL     | Both pipeline and CLI configs                                                                              |
| Examples                                     | PARTIALLY_FUNCTIONAL | 3 runnable examples (basic, builder, outcomes), compile-tested                                             |
| `RelatedRef.Range`                           | FULLY_FUNCTIONAL     | Span-based related locations with SARIF/LSP round-trip                                                     |
| LSP diagnostic tags                          | FULLY_FUNCTIONAL     | `Unnecessary`/`Deprecated` preserved in metadata                                                           |
| SARIF `region.snippet`                       | FULLY_FUNCTIONAL     | Native SARIF snippet round-trip support                                                                    |
| Comprehensive `doc.go`                       | FULLY_FUNCTIONAL     | Full package documentation with examples and architecture notes                                            |
| DetectorRegistry                             | FULLY_FUNCTIONAL     | Thread-safe plugin-style detector constructor registry                                                     |
| IntervalIndex[T]                             | FULLY_FUNCTIONAL     | Generic O(n + k) overlap queries; used by Correlate                                                        |
| LineShiftMap                                 | FULLY_FUNCTIONAL     | Byte-offset-aware line+column shift tracking after edits                                                   |
| MergeIter                                    | FULLY_FUNCTIONAL     | Streaming iter.Seq merge with deduplication                                                                |
| ConfigFile (pipeline)                        | FULLY_FUNCTIONAL     | JSON config loading + ResolveDetectors/ResolveProviders                                                    |
| StageHook                                    | FULLY_FUNCTIONAL     | Per-stage before/after hooks with abort capability                                                         |
| GoASTProvider                                | FULLY_FUNCTIONAL     | AST-aware fix provider for .go files (go/parser)                                                           |
| GeneratedFileFilter                          | FULLY_FUNCTIONAL     | Removes findings from auto-generated files (sqlc, protobuf, etc.)                                          |
| ToolAdapter[O]                               | FULLY_FUNCTIONAL     | Generic tool→Finding converter adapter                                                                     |
| CategoryForLinter                            | FULLY_FUNCTIONAL     | 84 linter→category mappings, case-insensitive                                                              |
| Severity aliases                             | FULLY_FUNCTIONAL     | 11 severity aliases + SeverityFromLevel + PriorityString (v1.3.0)                                          |
| SubstringProvider column-aware               | FULLY_FUNCTIONAL     | Nearest-position heuristic with line+column disambiguation                                                 |
| Template factory                             | FULLY_FUNCTIONAL     | Pre-configured builder: stamp common fields once, build many findings                                      |
| Template.Builder (per-finding chaining)      | FULLY_FUNCTIONAL     | Returns *Builder from template for confidence/suggestion overrides (v1.6.0)                                |
| ParseConfidence                              | FULLY_FUNCTIONAL     | Inverse of Confidence.String(); named levels + decimals + empty default (v1.6.0)                           |
| Deterministic JSON/SARIF output              | FULLY_FUNCTIONAL     | json.Deterministic(true) on all 8 marshal calls; 8 byte-identity regression tests                          |
| ValidateAll (batch validation)               | FULLY_FUNCTIONAL     | Batch helper returning map[int]error for invalid findings                                                  |
| FlightRecorderHook                           | FULLY_FUNCTIONAL     | Go runtime execution trace flight recorder for pipeline observability                                      |
| FlightRecorder config-file integration       | FULLY_FUNCTIONAL     | FlightRecorderFileConfig + ResolveFlightRecorder(); 5 fields via YAML/JSON                                 |
| go-arch-lint module boundary CI              | FULLY_FUNCTIONAL     | 11 components, one-directional flow enforced; wired into ci.yml arch-check job                             |
| Docs-freshness CI                            | FULLY_FUNCTIONAL     | Staleness + code-doc sync checks; wired into ci.yml docs-freshness job                                     |
| Configuration guide                          | FULLY_FUNCTIONAL     | docs/guides/configuration.md: all CLI flags, config file, library ConfigFile API                           |
| Troubleshooting guide                        | FULLY_FUNCTIONAL     | docs/guides/troubleshooting.md: build, config, pipeline, fix, flight recorder errors                       |
| Multi-module benchmark                       | FULLY_FUNCTIONAL     | docs/reports/2026-08-08_multi-module-vs-monolith.md: zero runtime overhead                                 |
| Finding groups (GroupID)                     | FULLY_FUNCTIONAL     | Branded `GroupID`; JSON/SARIF/LSP round-trip; `Report.GroupFindings()` (v1.7.0)                            |
| Per-finding fix outcomes                     | FULLY_FUNCTIONAL     | `ApplyWithOutcomes`: applied/no-change/refused/conflict/invalid/failed (v1.7.0)                            |
| Rollback policies                            | FULLY_FUNCTIONAL     | Per-file default + all-files opt-in; config-file `fixRollbackAllFiles` (v1.7.0)                            |
| Fix outcome metrics                          | FULLY_FUNCTIONAL     | `Metrics.RecordOutcome`/`OutcomeCounts` + CLI `Fix outcomes:` summary (v1.7.0)                             |
| GroupID validation                           | FULLY_FUNCTIONAL     | `GroupID.IsValid()`: machine-safe identifiers, enforced by `Validate()` (v1.8.0)                           |
| Sorted finding groups                        | FULLY_FUNCTIONAL     | `Report.GroupFindingsSorted()` + `Group`; `Template.WithGroupID` (v1.8.0)                                  |
| Fix outcome callback                         | FULLY_FUNCTIONAL     | `Config.OnFixOutcome` status callback; `OnFix` deprecated (v1.8.0)                                         |
| Result-level fix outcomes                    | FULLY_FUNCTIONAL     | `PipelineResult.Outcomes`: deduped first-outcome-per-finding record (v1.12.0)                              |
| Dry-run fix planning                         | FULLY_FUNCTIONAL     | `FixApplier.ApplyDryRun`: full ApplyReport, zero writes (v1.8.0)                                           |
| Unsafe-path outcome surfacing                | FULLY_FUNCTIONAL     | Traversal findings become `failed` outcomes instead of silent drops (v1.8.0)                               |
| staticcheck fix extension                    | FULLY_FUNCTIONAL     | Optional `before`/`after` JSON fields make findings auto-fixable (v1.8.0)                                  |
| Release preflight gate                       | FULLY_FUNCTIONAL     | `scripts/release-preflight.sh`: structural checks as code before tagging (v1.8.0)                          |
| Flight-recorder rotation + gzip              | FULLY_FUNCTIONAL     | `MaxFiles` pruning + `Compress` `.trace.gz` snapshots, rotation serialized under `writeMu` (v1.9.0)        |
| Tool SDK (`toolsdk` sub-module)              | FULLY_FUNCTIONAL     | `Spec` + triggers + registry plugin contract for external tools (v1.10.0)                                  |
| Multi-edit fixes (typed edit lists)          | FULLY_FUNCTIONAL     | `Finding.Edits` + `TextEdit`; lossless go/analysis round-trip; `EditListProvider` first in chain (v1.14.0) |
| Module-coverage evidence                     | FULLY_FUNCTIONAL     | `Summary.FilesScanned` always serialized + `Summary.SkippedModules` (v1.14.0)                              |
| Detector coverage channel                    | FULLY_FUNCTIONAL     | `DetectorV2.DetectV2` returns `DetectResult{Findings, FilesScanned, SkippedModules}`; `RunDetect` + `Report.AddDetectResult` feed it end-to-end (v1.15.0) |

---

_Assisted-by: Crush <crush@charm.land>_
