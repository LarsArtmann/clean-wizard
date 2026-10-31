# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term, bounded work see [TODO_LIST.md](TODO_LIST.md).
> For shipped features and their status see [FEATURES.md](FEATURES.md).

---

## Current Phase: Consumer ecosystem growth

**Current version:** 1.13.0 ([unreleased work](CHANGELOG.md#unreleased) will accumulate toward the next release)

v1.0.0 locked the API (2026-06-24). v1.1.0 added multi-module workspace, branded type safety, SARIF suppression round-trip, LSP data fidelity. v1.2.0 extracted `lockutil`, defragmented tests. v1.2.1 shipped 15+ correctness/security fixes, `encoding/json/v2` migration. v1.3.0 added 12 consumer-driven convenience APIs based on a full audit of 22 consumer projects. v1.4.0 added `go-error-family` integration (unified error classification), community readiness infrastructure (SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates), and retired the "zero external deps" principle in favor of a small, deliberate dependency surface. v1.4.1 eliminated all code duplication (zero clones at `-t 1`), consolidated test setup across all 4 modules, and resolved 7 pipeline lint issues. v1.5.0 shipped deterministic JSON/SARIF output, `ValidateAll` batch validation, `FlightRecorderHook` pipeline observability, and the `tagsEqual` fix for order-insensitive equality. v1.6.0 shipped `ParseConfidence`, `Template.Builder`, flight-recorder config-file integration, exported path-safety APIs (`ResolveSafePath` family), and 7 CI structural-check scripts. v1.7.0 shipped `GroupID` finding groups, per-finding fix outcomes (`ApplyWithOutcomes`/`ApplyWithReport`), per-file rollback default (ADR-016), typed outcome errors, and outcome metrics. v1.8.0 shipped `GroupID` validation, `Config.OnFixOutcome`, `FixApplier.ApplyDryRun`, `Report.GroupFindingsSorted`, `Template.WithGroupID`, unsafe-path outcome surfacing, the staticcheck fix-extension, and the release-preflight gate (sub-module go.mod drift corrected in tags). v1.9.0 shipped flight-recorder rotation + gzip, the redesign of the bench/CI gates on measured evidence, and the release-preflight self-test; v1.9.1/v1.9.2 were pipeline-verification patch releases that resurrected the GitHub Release pipeline (34-asset signed releases). v1.10.0 added the `toolsdk` sub-module (the BuildFlow provider plugin contract, migrated out of BuildFlow). v1.11.0 was a release-integrity hardening release (Release concurrency guard, preflight unpushed-commits + worktree-hygiene gates, docs version-stamp guard, erraudit gate at 0 violations across all 5 modules, correlation proximity-sort underflow fix). v1.12.0 exposed `PipelineResult.Outcomes` (per-finding fix outcomes on the pipeline result, issue #32) and fixed the flaky linter-category registry tests plus dependency-graph hygiene. v1.13.0 added `toolsdk.Trigger.NotRequires` (ownership deference) and fixed the tag-validation error message.

The library is production-ready and API-stable. The focus now shifts to growing the consumer ecosystem, expanding language coverage, and completing the public launch.

---

## v1.0.0-v1.13.0 - API lock, consumer convenience, observability, community readiness, release integrity

**Status: Released.**

- API locked at v1.0.0. Breaking changes require major version bump.
- Multi-module workspace (core, pipeline, analysis, CLI) established.
- Branded types (`ID`, `RuleName`, `ToolName`, `FilePath`), hand-rolled SARIF, LSP round-trip fidelity.
- v1.3.0: Consumer-driven APIs (`BuildOrDefault`, `Template`, `NewReportFromFindings`, `FilePos`, `SeverityFromLevel`, `ApplySimpleFixes`, `CheckBinary`/`RunCmd`, `FormatTable`, `PriorityString`). File-level position validation relaxed.
- v1.4.0: `go-error-family` integration (`FindingError.ErrorCode()` / `ErrorFamily()`), community infrastructure (SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates), documentation accuracy sweep. "Zero external deps" principle retired.
- v1.4.1: Zero code duplication (extracted `must[T]`, `marshalJSONString`, `decodeConfig`, `fixEditJSON`), test setup consolidated (`NewParallelGomega`), 7 pipeline lint issues resolved.
- v1.5.0: Deterministic JSON/SARIF output (`json.Deterministic(true)` on all 8 marshal call sites, 8 byte-identity regression tests), `ValidateAll` batch helper, `FlightRecorderHook` (Go runtime execution trace flight recorder for pipeline observability), `tagsEqual` fix (order-insensitive tag equality aligning code with documented contract).
- v1.6.0: `ParseConfidence` + `Template.Builder` consumer convenience APIs, flight-recorder config-file integration (`FlightRecorderFileConfig` + CLI `flightRecorder` section), exported path-safety boundary (`ResolveSafePath`/`ResolveSafePathFrom`/`ResolveRoot`), per-module CHANGELOGs, go-arch-lint module boundary enforcement, 7 CI structural-check scripts.
- v1.7.0: `GroupID` finding groups (JSON/SARIF/LSP round-trip, `Report.GroupFindings()`), per-finding fix outcomes (`FixEngine.ApplyWithOutcomes`, `FixApplier.ApplyWithReport`, `Metrics.RecordOutcome`/`OutcomeCounts`), per-file rollback default (`RollbackPolicyFailingFile`, ADR-016; issue #28), typed outcome errors (`*finding.FindingError`), negative-LSP-tag rejection (`FuzzParseLSPDiagnosticTags`), LSP tag re-emission.
- v1.8.0: `GroupID` validation (machine-safe identifiers, D7), `Config.OnFixOutcome` (D3), `FixApplier.ApplyDryRun` plan/apply (D4), `Report.GroupFindingsSorted()` + `Group`, `Template.WithGroupID`, unsafe-path findings surface as failed outcomes, staticcheck detector `before`/`after` fix extension, `scripts/release-preflight.sh` (structural pre-tag gate; v1.7.0 sub-module go.mod drift corrected — verified non-breaking, see CHANGELOG Fixed).
- v1.9.0: FlightRecorder rotation (`MaxFiles`) + gzip (`Compress`), `-trace-max-files`/`-trace-gzip` CLI flags, bench gate redesigned on measured evidence (alloc-based), release-preflight `--bench`/`--stress`/`--post-tag` modes + self-test, coverage restored above the 98% gate.
- v1.9.1 / v1.9.2: pipeline-verification patch releases (test robustness, CI stress split, Homebrew formula fix) — the first live, fully-signed GitHub Releases since v1.4.0 (34 assets on v1.9.2/v1.10.0).
- v1.10.0: `toolsdk` sub-module (`toolsdk/v1.10.0`) — the BuildFlow provider plugin contract (`Spec`, triggers, registry, dry-run plumbing), migrated from BuildFlow so tools target a stable SDK instead of BuildFlow internals.
- v1.11.0: Release-integrity hardening — Release workflow concurrency guard, preflight unpushed-commits + worktree-hygiene gates, docs version-stamp guard, preflight self-test `--post-tag` coverage, correlation proximity-sort underflow fix, erraudit gate at 0 violations across all 5 modules (`scripts/error-audit.sh`).
- v1.12.0: `PipelineResult.Outcomes` (per-finding fix outcomes on the pipeline result, issue #32), flaky linter-category registry tests fixed at root cause, dependency-graph hygiene (stray requires dropped, toolsdk go directive restored).
- v1.13.0: `toolsdk.Trigger.NotRequires` (disqualifying patterns for ownership deference), tag-validation error message states the actual charset contract.

---

## Raw Ideas (not yet actionable)

These are directions worth exploring. They are **not** committed work - they exist to capture thinking before it is lost. When an idea becomes concrete enough to act on, it graduates to [TODO_LIST.md](TODO_LIST.md).

### AI-assisted remediation

`FixStrategyAI` is a reserved constant with no backend. A real implementation would need:

- A pluggable `AIProvider` interface (request -> suggested diff)
- Guardrails: sandboxed apply, verification re-run, human approval gate
- Cost/rate-limit awareness in the pipeline

This is the largest open product direction and the original motivation for the library's fix pipeline.

### Detector-level provenance channel (the RepairResult.Facts twin)

`RepairResult.Facts` (#42) gives repairs a structured provenance channel; detectors have none — hosts wanting detect-side context (e.g. how many (migrator, file) pairs a go-auto-upgrade run's failure cache suppressed at detect time, the glue-era `RunRecord.FailureCacheSuppressed` telemetry) must string-match Description or lose the data. Shape idea: an additive, non-gating annotations channel on the detect output, mirroring the Facts contract (describe the run, never claim its effect). Trigger: a second consumer needing detect-side context, or the go-auto-upgrade switch-over's detect-side telemetry delta being judged material (raised 2026-10-11; owner decision pending).

### Language expansion

The core `Finding` model and SARIF/LSP interchange are language-agnostic. The fix engine has a Go AST provider (`pipeline/goast/`) but the provider architecture supports more:

- Rust (`syn`-based provider)
- TypeScript/JavaScript (tree-sitter)
- Python (ast / libcst)

Each would live in its own subpackage to keep language-specific dependencies out of the core.

### Tooling integrations

- **IDE plugins** - VS Code / Neovim consuming LSP diagnostics from `ToLSP()`
- **LSP code action support** - `ToLSP()` emits diagnostics but not code actions (auto-fix proposals). The data model has `FixStrategy`/`BeforeCode`/`AfterCode` but LSP code actions would require `LSPCodeAction` wire types.
- **Watch mode** - re-run the pipeline on file change (`fsnotify`)
- **Interactive TUI** - triage and review findings before applying fixes
- **GitHub Actions action** - first-class SARIF upload with fix PR generation
- **Profile-guided optimization** - PGO investigation. The pipeline hot path (fix engine, merge) could benefit from PGO profiles.

### Performance

- **IntervalTree go/no-go (decided 2026-09-08: NO-GO)** - `IntervalIndex` answers overlap queries in O(n + k): it binary-searches the Start cutoff, then scans every interval with `Start < end`, filtering by End. A true interval tree would give O(log n + k). Why no-go at current scale: (1) consumers correlate 10²-10⁴ findings per run — measured `IntervalIndex_Query/10000_intervals_100_queries` is ~10-14µs and `Correlate_RangeBased/1000` ~2-3.6ms; worst-case O(n·m) at 10⁴×10⁴ stays in seconds and no consumer reports pain; (2) tree nodes are pointer-heavy and cache-hostile vs the current contiguous slice, and build cost roughly doubles; (3) if scale ever demands it, a cheaper middle path exists first: augment the sorted slice with max-End prefixes (or a second by-End view) to bound the scan window without leaving slice layout. Revisit trigger: a consumer correlating >50k findings per run or `Correlate` dominating a pprof profile.

### json/v2 stabilization watch — CLOSED (2026-09-23)

`encoding/json/v2` is GA since Go 1.27 (release notes: it is the default `encoding/json` backend; the opt-out is `GOEXPERIMENT=nojsonv2`). The GOEXPERIMENT requirement was dropped in one sweep — `flake.nix`, both workflows, all scripts, README, CONTRIBUTING, AGENTS.md, templates, and guides — verified by a full build/test/lint pass with the variable unset. Consumers on Go 1.26 or older must upgrade (the go.mod floor is already 1.27).

### Consumer ecosystem

- **Consumer migration to v1.3.0+ APIs** - 14 Go consumers can now simplify their codebases using `BuildOrDefault`, `Template`, `SeverityFromLevel`, `FilePos`, `NewReportFromFindings`, `ApplySimpleFixes`, `ParseConfidence`, and `Template.Builder`. Each consumer independently reinvented these patterns.
- **More `ToolAdapter[O]` recipes** - Pre-built adapters for revive, ineffassign, errcheck, etc.
- **go-linter-sdk integration** - The sibling `go-linter-sdk` repo now has `WithToolName`, `RuleFunc.NewFinding`, `FilterRules`, and `ExitCodeByConfidence` (implemented in the 2026-08-08 cross-repo refactor session). The pilot migration of `go-humanize-linter` eliminated 97 LOC; go-linter-sdk tagged v0.3.0. Next: port more linters onto the SDK and keep the consumer matrix current (see `docs/ecosystem.md`).

### Multi-edit feature follow-ups (from the v1.14.0-candidate surface, 2026-09-23)

- **GoASTProvider typed Edits output** - emit `Finding.Edits` instead of raw `FixEdit` so CLI dry-runs and SARIF export benefit.
- **LSP CodeAction TextEditEdits** - expose `Finding.Edits` as code-action edits (`LSPDiagnosticData` is single-fix today).
- **AlternativeFixes representation** - carry `SuggestedFixes 1..N` (alternatives) instead of documented first-wins lossiness.
- **Multi-file fixes in FixApplier** - per-file grouping + atomic cross-file rollback policy; currently refused loudly.
- **`Preview()` multi-hunk** - render multi-edit findings as hunks.
- **`pipeline/examples/multi-edit`** - runnable bridge to applier example.

Launch-track nice-to-haves consciously parked (2026-09-23 docs sweep): social preview image, GitHub Discussions, profile pin, FUNDING.yml, post-launch retro — revisit on an adoption signal.

### FlightRecorder future directions

The FlightRecorder feature (`pipeline/flight_recorder.go`) is shipped with config-file integration (`FlightRecorderFileConfig` + `ResolveFlightRecorder()`).

**Triage (2026-09-08, FR1):** each idea scored Impact × Effort from the perspective of pipeline consumers. Two shipped along the way (context propagation, multi-recorder degradation); two graduated to actionable ROADMAP items below; the rest stay parked with reasons.

| Idea                                          | Impact | Effort | Verdict                                                                                                           |
| --------------------------------------------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------- |
| Trace file rotation (max-files)               | High   | Low    | **SHIPPED v1.9.0** — `MaxFiles` config + `-trace-max-files`; prunes oldest beyond cap, serialized under `writeMu` |
| Compressed trace output (gzip)                | Med    | Low    | **SHIPPED v1.9.0** — `Compress` config + `-trace-gzip`; `.trace.gz` snapshots, `go tool trace` e2e verified       |
| Automatic pprof capture                       | Med    | Med    | Park — useful but duplicates what `runtime/pprof` flags already give operators                                    |
| Continuous trace sampling (1% knob)           | Med    | Low    | **NO-GO (decided 2026-09-08)** — rationale below the graduated list                                               |
| Core package trace helper (`finding/tracing`) | Low    | Med    | Rejected — speculative generalization; only one consumer pattern exists (pipeline)                                |
| OpenTelemetry bridge                          | Low    | High   | Rejected — adds a heavy dependency for a rare use case in static-analysis tooling                                 |
| Trace diff tool                               | Low    | High   | Rejected — niche debugging aid; belongs in consumer tooling, not the library                                      |
| AI-assisted trace analysis                    | Low    | High   | Rejected — LLM analysis is consumer territory; the library's job is capturing the trace                           |
| Context propagation                           | —      | —      | **Shipped v1.5.0** — `Snapshot(ctx, reason)` accepts context; cancelled contexts skip the write                   |
| Multiple recorder support                     | —      | —      | **Shipped v1.6.0** — `Degraded()` mode degrades gracefully when Go's singleton recorder is taken                  |

Graduated items: rotation and gzip both **shipped in v1.9.0** (see the triage table above); nothing currently graduated.

**Sampling NO-GO rationale (decided 2026-09-08, evening session):** Go's
`runtime/trace.FlightRecorder` has no in-process sampling API — once started it records
continuously, so a "1% knob" could only sample at the _run_ level (enable the recorder for
N% of pipeline runs). That is 3 lines at the caller's construction site
(`if rand.Float64() < 0.01 { NewFlightRecorderHook(...) }`) and does not belong in library
config. Sampling pays off for long-lived servers with ambient, always-on tracing;
go-finding pipelines are short, explicit runs where the recorder is already opt-in
(`-trace` / config `enabled: true`) — the user has already decided to pay the cost, and we
have no overhead measurements suggesting a problem. Revisit trigger: a long-lived consumer
embedding pipelines that measures FlightRecorder overhead as material. Backstop review
date: **2027-01-31** — if the trigger has not fired by then, re-examine whether the trigger
definition still matches how consumers actually embed pipelines (the trigger is
event-based; the date exists so the decision cannot silently rot).

### Hardening (owner decisions pending)

These are known design tensions deferred because they require breaking changes. Concrete designs from the [data-model review](docs/reviews/archived/2026-07-18_21-10_data-model-review.html).

- **Position zero-value** - `Position{}` has `Offset=0` (valid byte 0), not "unset" (`position.go:33-38`: 0=unset for Line/Column, -1=unset for Offset). Resolved pragmatically in v0.9.0 with `-1` sentinel, but a type-safe redesign using `Option[T]` generic helpers is still on the table for v2.0.
- **`Range.End` zero-value ambiguity** - same class of issue as Position.
- **FixStrategy as closed union** - Current `type FixStrategy string` (`fix_strategy.go:4`) with string constants loses type safety. v2.0 design: `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved`.
- **Pointer-as-state fields** - `Range *Range` (`finding.go:29`), `Suppression *Suppression` (`finding.go:33`), `ExpiresAt *time.Time` (`suppression.go:20`), and `RelatedRef.Range *Range` (`finding.go:96`) all encode 3 states (nil/zero/valid) in a single pointer.
- **Tags to TagSet** - `Tags []Tag` (`finding.go:21`) forces order-insensitive equality in `finding_equal.go`. v2.0: `TagSet map[Tag]struct{}`.
- **Finding sub-struct composition** - Current flat struct (`finding.go:8-48`). v2.0: compose from `Identity{}`, `Location{}`, `Classification{}`, `Fix{}`. Changes JSON shape - must batch.

---

## Open questions (owner decisions)

Questions that block work but cannot be answered from the code. Each cites the
session that raised it.

- **Homebrew/nix distribution** — GoReleaser renders a validated Homebrew formula but ships it with `skip_upload: true`: the `LarsArtmann/homebrew-tap` repo does not exist and `HOMEBREW_TAP_GITHUB_TOKEN` was never created (likewise `nur-packages` for nix). Either create the tap repo + secret, or drop/comment the brew/nix sections until wanted. (Raised in `docs/status/archived/2026-09-09_03-24_late-night-session-self-review.md` §g/1.)
- **GitHub Release backfill for v1.5.0–v1.8.0** — those tags have no GitHub Releases (GoReleaser was billing-dead; v1.4.0 was the last published until the v1.9.2 resurrection). Backfill with source-only/notes-only releases, or leave history as-is with v1.9.2 as the first real one? (§g/2; the D6 decision "forward-only" was taken while CI was dark and is worth revisiting now that releases work.)
- ~~**Stale `~/projects/hierarchical-errors` clone**~~ **RESOLVED 2026-09-27: the directory no longer exists** (already removed operator-side; nothing to decide). (§g/3.)
- ~~**v1.13.0 completion vs v1.14.0 fold-in**~~ **RESOLVED 2026-09-23: completed.** The module proxy had already fetched core + toolsdk v1.13.0 (public since 09-22), so folding was impossible without a retract. All 5 tags exist at 450824e with GitHub Releases (Latest = v1.13.0, full GoReleaser assets); proxy + consumer-compat verified for all 5 modules; version-drift green. The multi-edit feature ships in v1.14.0. Execution record: the plan's annotation log (2026-09-23 17:00-19:30 entry).
- **`ApplySimpleFixes` philosophy** — keep refusing multi-edit findings (current: safe, points users at the pipeline `FixApplier`) or teach it byte-level application so the no-pipeline path is also lossless? (§g/3, same report.)
- **`art-dupl-report.html` policy** — the generated art-dupl report is tracked and churns ~800 lines of diff per run. Keep committing it, gitignore it, or commit only at milestones? (Raised in `docs/status/archived/2026-09-23_03-16_art-dupl-dedup-pass-session-status.md` §g/1.)

---

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Web UI** - Not aligned with the library's core purpose (data model + pipeline, not presentation layer).
- **Hosted/SaaS offering** - Too costly relative to impact for a library project.
- **Non-Go language providers (pre-v2)** - Language expansion is roadmap, but not until the Go provider ecosystem is fully proven.
- **Generic `JSONToolDetector`** - Too opinionated; every tool's JSON shape differs. `CheckBinary`/`RunCmd` helpers suffice.
- **`Properties map[string]any`** - Explicitly banned. `Metadata map[string]string` stays for type safety and interchange simplicity.

---

_Assisted-by: Crush <crush@charm.land>_
