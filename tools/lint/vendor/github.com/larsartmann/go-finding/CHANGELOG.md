# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Sub-module changelogs:** [pipeline/CHANGELOG.md](pipeline/CHANGELOG.md),
> [analysis/CHANGELOG.md](analysis/CHANGELOG.md),
> [cmd/go-finding/CHANGELOG.md](cmd/go-finding/CHANGELOG.md),
> [toolsdk/CHANGELOG.md](toolsdk/CHANGELOG.md).

## [Unreleased]

### Added

- Nothing yet.

### Fixed

- Nothing yet.

## [1.15.0] - 2026-10-12

Coverage-evidence release: detectors that can measure their own scan coverage
now flow that evidence end-to-end. Sub-module release rides along:
[pipeline/v1.15.0](pipeline/CHANGELOG.md) (every detect path dispatches
through the new channel).

### Added

- **`DetectResult` + `DetectorV2` coverage channel (T07b)** — detectors that
  can measure their own coverage implement `DetectorV2.DetectV2(ctx)` and
  return `DetectResult{Findings, FilesScanned, SkippedModules}`; hosts call
  `RunDetect(ctx, d)` (uses v2 when implemented, falls back to plain `Detect`
  with coverage zeroed — documented, not an error) and `Report.AddDetectResult`
  to feed findings plus coverage into a report. `FilesScanned` accumulates
  across detectors, `SkippedModules` unions in first-seen order, and both
  survive `ComputeSummary` recomputation. Answers "did the tool actually see
  every file?" from the interchange (Summary.FilesScanned/SkippedModules
  shipped in v1.14.0 with no way for a Detector to populate them).
  Source-compatible: `Detector` is unchanged, `DetectorV2` embeds it.

### Fixed

- **CI and Release workflow Go pins matched to the toolchain floor** — both
  workflows pinned `go-version: "1.26"` while `go.mod`/`go.work` require
  `go 1.27`, so every master CI run since 2026-09-20 failed at setup and both
  v1.13.0 Release runs died in the test job (`go.work requires go >= 1.27`),
  leaving v1.13.0 with no GitHub Release. Pins are now `1.27`; the red runs
  were invisible to local gates, which already ran Go 1.27.

## [1.14.0] - 2026-10-05

Root release: the toolsdk options channel (PR #41), the toolsdk switch-topology
API, report coverage evidence, and the json/v2 GA build change. Sub-module
releases ride along: [toolsdk/v1.14.0](toolsdk/CHANGELOG.md) (options channel)
and [toolsdk/v1.15.0](toolsdk/CHANGELOG.md) (`Spec.SwitchCases`). Train
completed 2026-10-11: [pipeline/v1.14.0](pipeline/CHANGELOG.md) (typed edit
lists end-to-end), [analysis/v1.14.0](analysis/CHANGELOG.md) (lossless
multi-edit round-trips), and
[cmd/go-finding/v1.14.0](cmd/go-finding/CHANGELOG.md) (CLI pinned to the
v1.14.0 siblings).

### Added

- **toolsdk options channel (PR #41)** — tools can declare typed `Option`s
  (name, kind, default) on their `Spec`; consumers pass per-run values that are
  validated against the declarations (`ErrUnknownOption` on typos,
  `ErrInvalidOption`/`ErrDuplicateOption` on malformed specs).
- **toolsdk `Spec.SwitchCases` + `ExcludedRuleIDs` (toolsdk/v1.15.0)** — SDK
  tools can declare conditional repair paths keyed on diagnose finding rule
  IDs, mirroring BuildFlow's `domain/tool.SwitchCase` field-for-field; mutual
  exclusion of mutating repair cases is encoded via `ExcludedRuleIDs`
  (go-workflow runs every predicate-matching case in parallel).
- **`Summary.FilesScanned` always serialized + `Summary.SkippedModules`** —
  a report can now say "scanned 0 files" as coverage evidence instead of
  omitting the field (clean full scans emit `"filesScanned": 0`), and monorepo
  coverage boundaries have a home: `SkippedModules` carries the nested go.mod
  modules a tool deliberately did not analyze; module-agnostic tools leave it
  nil and their JSON is unchanged.
- **`finding.TextEdit` and `Finding.Edits` (issue #36)** — typed multi-edit fix
  representation: `Start`/`End` positions plus `NewText`, with `HasSpan`,
  `IsInsertion`, `IsDeletion`, `EffectiveFile`, and `Validate` helpers,
  order-independent list equality, and builder support via `WithEdits`. When
  set, `Edits` is the authoritative machine representation of the fix;
  `BeforeCode`/`AfterCode` remain the first-edit display summary. Direct fixes
  now validate with either representation ("requires BeforeCode, AfterCode, or
  Edits"). SARIF export/import round-trips edit lists, including byte offsets
  via `region.byteOffset`/`byteLength`.
- **`pipeline.EditListProvider` (issue #36)** — resolves a finding's typed edit
  list (`Finding.Edits`) to byte-level edits: byte offsets directly, line/column
  via the shared line index. Runs first in the default provider chain (before
  Offset, Line, and Substring). Loud failures instead of partial application:
  lists spanning files fail with `ErrEditCrossFile`; stale or out-of-bounds
  offsets fail with `ErrEditStale` (both surface as `FixOutcomeFailed` with the
  provider cause in the error chain). See
  [pipeline/CHANGELOG.md](pipeline/CHANGELOG.md).

### Changed

- **`GOEXPERIMENT=jsonv2` no longer required** — `encoding/json/v2` is stable
  and enabled by default since Go 1.27, and the toolchain floor was already
  `go 1.27`. The env var was removed from the flake, CI/Release workflows,
  scripts, and all docs. Consumers must build with Go 1.27+ (previously:
  Go 1.26 + the experiment flag).
- **`ApplySimpleFixes` refuses multi-edit findings** — findings whose `Edits`
  list more than one edit are skipped with an actionable reason pointing at the
  pipeline `FixApplier`, instead of silently fixing only the first edit.
- **`FixEngine` applied counts are per finding, not per edit (pipeline)** — a
  multi-edit fix counts once in `Apply`/`ApplyWithConflicts` results and
  `ApplyReport.Applied`, matching "number of successful fixes" semantics.
- **Multi-edit suggested fixes are lossless (analysis)** —
  `FromDiagnosticWithSource` carries all `TextEdits` of a diagnostic's first
  `SuggestedFix` into `Finding.Edits` (previously edits 2..N were silently
  dropped, issue #36); `ToDiagnostic` reconstructs the full edit list.
  Alternative fixes (`SuggestedFixes` 1..N) remain first-wins and are now
  documented as such.

## [1.13.0] - 2026-09-22

### Added

- **`toolsdk.Trigger.NotRequires`** — disqualifying file patterns that express
  ownership deference: a tool whose NotRequires pattern matches does not run
  (e.g. a standalone formatter deferring to a repo's own treefmt config).
  Field-for-field parity with BuildFlow's `domain/tool.Trigger`; see
  [toolsdk/CHANGELOG.md](toolsdk/CHANGELOG.md).

### Fixed

- **Tag validation error message** — the finding-validation error for invalid
  `Tags` entries claimed "tags must be non-empty" even for non-empty tags that
  violated the lowercase-hyphenated convention; the message now states the
  actual contract (first char a-z, then a-z0-9-).

## [1.12.0] - 2026-09-17

### Added

- **`PipelineResult.Outcomes` (pipeline)** — per-finding fix outcomes are now exposed on the
  pipeline result itself, populated alongside the `Config.OnFixOutcome` callback whenever fix
  application ran. Deduplicated to the first outcome per finding identity, so re-detected
  findings whose fix already applied no longer force every in-process consumer to hand-roll a
  collector plus dedup (issue #32). Empty in `DryRun` mode or when nothing was fixable.

### Fixed

- **Flaky linter-category registry tests** — `TestRegisterLinterCategory` overrode the shared
  global registry entry `govet` while parallel tests asserted the built-in mapping, so the
  stress gate failed nondeterministically (observed as `govet maps to correctness` receiving
  `security` on CI repeat=20). The global function is now pinned with a name unique to the
  test and no pre-state assertion (the registry also persists across `-count` iterations);
  override semantics are pinned on an isolated registry. Test-only; no library behavior change.
- **Dependency graph hygiene** — dropped stray `go-finding/pipeline` requires from the
  analysis, pipeline, and toolsdk go.mod files (declared dependencies that no code imports),
  restored the toolsdk `go` directive to match go.work (`1.26.7`), and re-attached the toolsdk
  package doc comment to its `package` statement (revive `package-comments`).
- **`cmd/go-finding` go.sum resync** — checksums now match the published v1.11.0 tags and the
  indirect `go-branded-id` bump (v0.5.1 → v0.6.0).

## [1.11.0] - 2026-09-17

### Added

- **Release workflow concurrency guard** — the Release workflow now runs under a
  queueing concurrency group (`cancel-in-progress: false`), so a manual dispatch
  can no longer race the tag-push auto-trigger into two concurrent GoReleaser
  runs (the v1.9.2 422 duplicate-asset failure is now impossible by construction).
- **Preflight unpushed-commits gate** — `release-preflight.sh` fails when commits
  exist in `@{u}..HEAD` or the branch has no upstream, closing the gap that let
  CI be dispatched twice against unpushed commits on 2026-09-08.
- **Preflight worktree-hygiene gate** — fails on any git worktree registered
  outside the repo (the /tmp strays that went unnoticed for two days) and on
  prunable worktree registrations.
- **Docs version-stamp guard** — `docs-api-check.sh` now requires the README
  `finding.Version` example to exist and to match `version.go`, so the
  current-version stamp can no longer silently lag a release (v1.9.1/v1.9.2
  slipped this way).
- **Preflight self-test covers `--post-tag` failure path** — the self-test
  injects a missing tag in post-tag mode and asserts preflight FAILs; previously
  only the happy path was proven.

### Fixed

- **Integer underflow in correlation proximity sorting** — `correlateByProximity`
  computed line deltas with subtraction (`a.Line - b.Line`), which wraps for
  negative deltas under wide line ranges and silently mis-sorts correlated
  findings. Replaced with `cmp.Compare` (8a9b7c8); ordering is pinned by test.
- **Error-handling hardening — erraudit gate now 0 violations across all 5 modules** —
  fixed 3 critical context-loss errors (`FileBackup` backup-dir failures now carry
  the directory via `ioErrorAt`; `NewPipeline` config-validation and fix-applier
  failures now include `rootDir`), migrated the govet/staticcheck detectors from
  `errors.As` to `errors.AsType`, and documented 27 intentional patterns with
  reasoned `//nolint:erraudit` directives (`nolint-audit`: 0 stale).
- **New gate: `scripts/error-audit.sh`** — runs the erraudit hierarchical
  error-handling analyzer in every module with an explicit OK/FAIL verdict;
  the FAIL path is proven by injected-violation self-test. Policy and the
  samber/oops rejection are documented in
  `docs/research/erraudit-violation-analysis.md` §9. CI wiring is blocked
  until the erraudit repository is public.
- **Version drift fixed** — `cmd/go-finding` now requires `pipeline v1.10.0`
  (the v1.10.0 release bumped sub-module requires on master but missed this
  one); alignment tags `pipeline/v1.10.0`, `analysis/v1.10.0`, and
  `cmd/go-finding/v1.10.0` created at master (push pending, ≤3-per-push rule).
- **flake.nix vendorHash refreshed** after the indirect dependency bumps
  (`golang.org/x/net`, `golang.org/x/text`, `golang.org/x/tools`); the module
  sums left half-tidied by that bump were completed with `go mod tidy`, and
  `nix flake check` passes again.

## [1.10.0] - 2026-09-10

No core-module code changes.

### Added

- **`toolsdk` sub-module** (`github.com/larsartmann/go-finding/toolsdk`): the BuildFlow
  provider plugin contract, migrated from `github.com/larsartmann/buildflow/tool-sdk`.
  External tools declare a `Spec` (name, trigger, inputs, `Detect`/`Repair` via core
  `finding` interfaces) and self-register at init; BuildFlow discovers them via `All()`
  and converts each into a DAG provider. Depends only on the core module. First release:
  `toolsdk/v1.10.0`, aligned with this root release.

## [1.9.2] - 2026-09-08

Release-pipeline fix only: no code changes in any module.

### Fixed

- **Homebrew formula template** — the GoReleaser brew `test` stanza used `#{{bin}}`, which Go's template engine evaluates as a call to an undefined function `bin` (first live GoReleaser run failed on exactly this after successfully building, checksumming, and cosign-verifying all platform archives). The correct stanza is `system "#{bin}/go-finding --version"` — `#{bin}` is Ruby/Homebrew interpolation evaluated at install time.

## [1.9.1] - 2026-09-08

Test-only patch release: no production code changes in any module. Tags exist so the Release workflow finally runs green on a public repo (the v1.9.0 run failed in its test job on an environment-flaky test, before reaching GoReleaser).

### Fixed

- **`TestFinding_WriteJSON_NewlineError` made environment-robust** — The test pinned the Nth write call, but the encoder's internal flush granularity is not contractual and varies across environments (the same commit passed CI's test job and failed the Release run's). It now fails exactly the semantic target — the single-byte `\n` trailing-newline write — via a writer that rejects only that payload. Also: `docs-api-check.sh` now exits loudly (code 2) when ripgrep is missing instead of reporting every identifier as absent, and its CI job installs ripgrep; the CI stress job is split per suite type (ginkgo suites reject `go test -count>1`).

## [1.9.0] - 2026-09-08

No core-module API changes. Pipeline: flight-recorder `MaxFiles` rotation + `Compress` (gzip) snapshots, with rotation now serialized against concurrent writes — see `pipeline/CHANGELOG.md`. CLI: `-trace-max-files` / `-trace-gzip` flags; staticcheck parser regression-guarded against real tool output.

### Changed

- **Benchmark regression gate redesigned (evidence-based)** — Sustained full-suite benchmark runs thermally depress later benchmarks on the dev machine (identical code measured -9%..+150% depending on suite position; `IntervalIndex_Query/1000` = 9.3µ fresh vs 20.1µ late-suite, n=10, p=0.000 both). `bench-check.sh` now fails on allocation regressions beyond +10% (deterministic signal, ±0% on unchanged code) and on time regressions beyond 3.5x (algorithmic blowouts). Fine-grained cross-session time comparison stays a manual, filtered-benchmark exercise. Baseline regenerated at v1.9.0.
- **`release-preflight.sh` hardened** — new `--post-tag` mode (all 4 tags must exist + `version-check.sh` runs), `--bench` mode now captures benchmarks itself (previously compared against a stale temp file — a dead-gate bug found by executing the mode), and `scripts/release-preflight-selftest.sh` injects a synthetic tag collision + version drift into a disposable worktree and asserts the gate FAILS on each (wired as a CI job).

## [1.8.0] - 2026-09-08

### Added

- **`Report.GroupFindingsSorted()` + `Group`** — Deterministic variant of `GroupFindings()`: groups sorted by `GroupID`, members in report order, `nil` when no grouped findings. For tests, serialization, and stable output where map iteration order must not leak.
- **`Template.WithGroupID`** — Stamps one `GroupID` onto every finding built from the template (whole template = one clone group); per-finding groups remain available via `Builder.WithGroupID`.

### Changed

- **`GroupID` validation (D7)** — A set `GroupID` must now be a machine-safe identifier: no whitespace or control characters, at most 128 bytes (empty = not grouped, always valid). The charset stays permissive (underscores, dots, uppercase allowed) so existing consumer formats keep validating; the guard protects SARIF property keys, LSP data, and map keys from corruption. `GroupID.IsValid()` is the standalone check.
- **Unsafe-path fix findings now surface as failed outcomes (pipeline)** — Findings whose `Position.File` resolves outside the applier root (e.g. `../../etc/passwd`) were previously dropped silently during grouping. They are now reported as `failed` outcomes with a validation-category error and included in the joined error return; safe findings in the same run still apply. See `pipeline/CHANGELOG.md`.

### Fixed

- **Sub-module go.mod core references realigned (v1.8.0 tags)** — The v1.7.0 sub-module tags carried stale `go-finding v1.6.0` requires (the sync fix landed on master after tagging). Verified non-breaking: `pipeline@v1.7.0` and `analysis@v1.7.0` compile standalone against core v1.6.0, so proxy consumers were unaffected in practice — but the metadata was semantically wrong. The v1.8.0 sub-module tags reference core v1.8.0. `scripts/release-preflight.sh` now gates version drift (and every structural check) before any tag exists.

## [1.7.0] - 2026-09-08

### Added

- **`GroupID` on `Finding`** — New optional `GroupID GroupID` field groups findings that belong to the same logical set (e.g., a clone group of N duplicated code blocks from art-dupl). Branded string type like `ID`/`RuleName`/`FilePath`; JSON key `groupId`, omitted when empty. Round-trips losslessly through JSON, SARIF (property `go-finding/groupId`), and LSP (`LSPDiagnosticData.GroupID`). Builder support via `WithGroupID`. `Report.GroupFindings()` returns active findings grouped by `GroupID`, excluding suppressed and ungrouped findings.
- **`ToLSP` emits LSP diagnostic tags from metadata** — Completes the LSP tag round-trip: a finding that carries `Metadata[LSPDiagnosticTagsKey]` (as written by `FromLSP`) now produces an `LSPDiagnostic` with `Tags` populated, so LSP clients (faded/struck-through rendering for `Unnecessary`) see the tags on every conversion.

### Changed

- **Fix rollback scope is now per-file by default** — Previously, any file failure rolled back every file modified earlier in the run, discarding clean fixes because of one bad file (see issue #28). Now only the failing file is restored and earlier files keep their applied fixes. Opt back into all-or-nothing via `FixApplier.SetRollbackPolicy(RollbackPolicyAllFiles)`, `pipeline.Config.FixRollbackAllFiles`, or the config-file field `fixRollbackAllFiles`. Soft per-finding failures (provider resolve errors, refused findings) no longer abort a run: they are reported while all successfully applied fixes stay on disk.

### Fixed

- **`ToLSP` no longer emits negative diagnostic tags** — Malformed `Metadata[LSPDiagnosticTagsKey]` values (e.g. `"-1,abc,3"`) previously produced `LSPDiagnostic.Tags` containing negative numbers, which are invalid per the LSP specification (tags are 1-based: 1 = Unnecessary, 2 = Deprecated). Negative values are now rejected during parsing. Found by the new `FuzzParseLSPDiagnosticTags` fuzz target and pinned by a golden-wire test.

### Dependencies

- Refreshed indirect `golang.org/x/` dependencies that drifted in via an automated housekeeping commit (`c04af34`): `x/mod` 0.38.0 → 0.40.0, `x/net` 0.57.0 → 0.58.0, `x/text` 0.40.0 → 0.41.0, `x/tools` 0.48.0 → 0.49.0 (core + pipeline + CLI modules, all `// indirect`; no direct dependency changes). Reviewed and confirmed intentional: kept instead of reverted per no-churn policy. Full test suite green.

## [1.6.0] - 2026-08-08

### Added

- **`pipeline.FlightRecorderFileConfig` + `ConfigFile.ResolveFlightRecorder()`** — Flight recorder configuration via JSON config files. The `FlightRecorderFileConfig` struct exposes 5 fields (`Enabled`, `OutputDir`, `SlowStageThreshold`, `MinAge`, `MaxBytes`) as string-encoded durations matching the ConfigFile convention. `ResolveFlightRecorder()` constructs a `*FlightRecorderHook` from the config section, returning `(nil, nil)` when disabled. Enables pipeline consumers to configure trace recording without CLI flags.
- **CLI `flightRecorder` config section** — YAML/JSON config files now support a `flightRecorder` section as an alternative to `-trace`, `-trace-dir`, and `-trace-slow` flags. Fields: `enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes` (full parity with `pipeline.FlightRecorderFileConfig`).
- **4 CI structural-check scripts** — `scripts/replace-audit.sh` (verifies replace directives), `scripts/version-drift.sh` (cross-checks go.mod version references), `scripts/test-naming.sh` (enforces test naming conventions), `scripts/go-work-sync.sh` (verifies go.work sync idempotency). Wired into `.github/workflows/ci.yml` as `structural-checks` and `go-work-sync` jobs.
- **`scripts/docs-freshness.sh`** — Checks documentation freshness against code changes. Flags docs not modified in N days (default 180) and docs whose referenced source files have been modified since the doc was last updated. Wired into ci.yml as `docs-freshness` job.
- **Per-module CHANGELOGs** — `pipeline/CHANGELOG.md`, `analysis/CHANGELOG.md`, `cmd/go-finding/CHANGELOG.md` with cross-references from root CHANGELOG.md.
- **go-arch-lint module boundary enforcement** — `.go-arch-lint.yml` defines 11 component boundaries across all 4 modules. Enforces one-directional dependency flow (cmd → pipeline → core, analysis → core). Wired into ci.yml as `arch-check` job.
- **E2E test for 5-field flightRecorder config** — `TestRun_E2E_TraceViaConfigFile_AllFields` verifies `minAge` and `maxBytes` reach the pipeline when set via YAML config.
- **resolveSafePath edge case tests** — Circular symlinks (self-referential and mutual), dangling symlinks, root-is-symlink, and root-is-symlink path traversal in `pipeline/path_safety_test.go`.
- **FlightRecorder edge case tests** — Snapshot write error (permission denied), slow last-stage snapshot during pipeline run, concurrent stage events safety with race detector in `pipeline/flight_recorder_test.go`.
- **TOCTOU path safety tests** — Symlink swap attack tests for `resolveSafePath` in `pipeline/path_safety_test.go`.
- **LSP serialization benchmarks** — `BenchmarkToLSP`, `BenchmarkFromLSP`, `BenchmarkLSPRoundTrip` in `bench_test.go`.
- **FlightRecorder user guide** — `docs/guides/flight-recorder.md` covering CLI flags, config-file integration, programmatic API, and snapshot workflow.
- **`ParseConfidence(string) (Confidence, error)`** — Inverse of `Confidence.String()`. Parses named levels ("none", "low", "medium", "high", "full"), decimal strings (e.g. "0.42"), and empty string (defaults to `ConfidenceLow`). Case-insensitive, trims whitespace. Returns `ErrInvalidConfidence` sentinel error, matchable via `errors.Is`. Eliminates the per-consumer switch statement every CLI linter reinvents for its `--min-confidence` flag.
- **`Template.Builder(rule, message, severity, pos) *Builder`** — Returns a pre-configured `*Builder` from the template, allowing per-finding chaining (confidence, suggestion, before/after code, metadata) before the terminal `Build()`/`MustBuild()`/`BuildOrDefault()` call. `Template.Build()` delegates to `Builder().BuildOrDefault()` for backward compatibility. Eliminates the `makeFindingWithConfidence` / `buildFixableFinding` factory patterns consumers reinvent when they need both template-level defaults AND per-finding overrides.
- **Configuration guide** — `docs/guides/configuration.md` covers all CLI flags, YAML/JSON config file format, CLI vs config precedence, library ConfigFile API, and runtime defaults.
- **Troubleshooting guide** — `docs/guides/troubleshooting.md` covers build/setup errors, config file errors, pipeline runtime errors, fix application errors, flight recorder issues, and finding validation errors.
- **Multi-module benchmark report** — `docs/reports/2026-08-08_multi-module-vs-monolith.md` quantifies zero runtime overhead, negligible build overhead, and significant dependency isolation benefit.
- **sanitizeFilename fuzz test** — `FuzzSanitizeFilename` in `pipeline/flight_recorder_fuzz_test.go` verifies non-empty output, no consecutive hyphens, no leading/trailing hyphens, and only safe filename characters across random inputs.
- **SARIF schema compliance edge cases** — 5 new tests in `sarif_properties_test.go`: multiple findings, file-level positions, minimal findings, empty reports, and suppressed findings.
- **`pipeline.ResolveSafePath`, `ResolveSafePathFrom`, `ResolveRoot`** — Exported the path-traversal security boundary as public API so external callers can validate finding file paths without duplicating logic. `ResolveSafePath(rootDir, relPath)` is the convenience wrapper; `ResolveRoot(rootDir)` (resolve symlinks once) + `ResolveSafePathFrom(resolvedRoot, relPath)` (per-path resolution) enable batch optimization.
- **`pipeline.FlightRecorderHook.Degraded()`** — When a second flight recorder is created while one is already active (Go singleton limit), `NewFlightRecorderHook` now returns a degraded hook instead of an error. `Degraded()` returns true; all snapshot operations silently no-op. Callers can detect this state without error handling.
- **`scripts/json-deterministic-check.sh`** — CI script enforcing `json.Deterministic(true)` on all production marshal calls via paren-depth-aware source scanning. Excludes `doc.go`. Wired into ci.yml.

### Changed

- **doc.go FlightRecorder section** — Updated field list to include `minAge` and `maxBytes`; added soft reference to the project's FlightRecorder guide.
- **version-drift.sh** — Excludes `// indirect` lines from grep to prevent false positives when a module appears as both direct and indirect dependency.
- **docs/guides/flight-recorder.md** — Updated YAML and JSON config examples to include all 5 fields (`minAge`, `maxBytes`).
- **README.md** — Added FlightRecorder to feature list with link to guide; enhanced Pipeline description with observability mention.
- **docs/DOMAIN_LANGUAGE.md** — Added Observability section (StageHook, FlightRecorderHook, Metrics, MetricsSnapshot) and Pipeline Configuration section (Iteration, CompletionReason, DryRun, GracefulDegradation, ByteLevelConflictDetection, VerifyAfterFix).
- **`pipeline.FlightRecorderHook.Snapshot` signature** — Now accepts `context.Context` as first parameter: `Snapshot(ctx, reason)`. Cancelled contexts skip the trace write and return `context.Cause(ctx)` wrapped in a descriptive error. **Breaking change** for callers using the v1.5.0 `Snapshot(reason)` signature — update to `Snapshot(ctx, reason)`.
- **`pipeline.NewFlightRecorderHook` singleton behavior** — On "flight recorder already enabled" runtime error, now returns `(degradedHook, nil)` instead of `(nil, err)`. Callers that previously checked `err != nil` for this case now get a safe no-op degraded hook.
- **Centralized JSON marshal options** — `marshalOpts` and `prettyMarshalOpts` package-level variables in `json.go` replace duplicated `json.Deterministic(true)` literals across all marshal methods. SARIF export (`ToSARIFWithOpts`, `WriteSARIFWithOpts`) now reuses `prettyMarshalOpts`.
- **`scripts/docs-freshness.sh`** — Tightened Go-file reference matching to only backtick code spans (`` `filename.go` ``) and markdown link targets (`[text](filename.go)`), excluding prose mentions that caused false positives.

### Fixed

- **`pipeline.FixEdit.MarshalJSON` non-deterministic output** — Missing `json.Deterministic(true)` caused non-reproducible fix serialization (map keys in random order). Now uses `marshalOpts`.

## [1.5.0] - 2026-08-06

Deterministic output, batch validation, pipeline observability, and tag equality fix.

### Added

- **`ValidateAll(findings []Finding) map[int]error`** — Batch validation helper. Returns a map of index → error for all invalid findings in a slice, or nil if all are valid. Lazy-allocates only when at least one finding is invalid. Use for validating detector output or config-loaded findings without manual loop boilerplate.
- **`pipeline.FlightRecorderHook`** — Go runtime execution trace flight recorder for pipeline observability. Wraps Go 1.25+'s `runtime/trace.FlightRecorder` to continuously buffer trace data in memory, snapshotting on demand or automatically when a pipeline stage exceeds a configurable slow-stage threshold. Invaluable for diagnosing slow detectors, fix contention, or unexpected pipeline stalls.
  - `NewFlightRecorderHook(config FlightRecorderConfig) (*FlightRecorderHook, error)` — creates and starts the recorder
  - `FlightRecorderConfig` — `MinAge` (buffer retention, default 30s), `MaxBytes` (default 4 MiB), `SlowStageThreshold` (auto-snapshot trigger, default off), `OutputDir` (default temp), `Logger`
  - `DefaultFlightRecorderConfig()` — sensible defaults
  - Implements `StageHook` — `OnStageEvent` never returns an error (trace collection is purely diagnostic)
  - `Snapshot(reason) (path, error)` — manual on-demand snapshot
  - `Close()` — stops recorder, waits for in-flight async snapshots
  - CLI flags: `-trace` (enable), `-trace-dir` (output directory), `-trace-slow` (auto-snapshot threshold, e.g. `30s`)
- **Determinism regression tests** — `TestDeterminism_*_RawBytesIdentical` test suite (8 tests) verifying all JSON/SARIF marshal paths produce byte-identical output across 100 invocations. Tests compare raw `[]byte`, not parsed structures. The SARIF test is proven to fail without `json.Deterministic(true)`.

### Changed

- Go module dependencies and Nix flake toolchain versions synchronized across all sub-modules.

### Fixed

- **Non-deterministic JSON/SARIF output** — All `json.Marshal`/`json.MarshalWrite` calls in production code now pass `json.Deterministic(true)`, ensuring Go map keys (SARIF `properties` bag, `Summary.bySeverity`, `Summary.byCategory`, `Finding.Metadata`, etc.) serialize in sorted order on every run. Previously, `encoding/json/v2` serialized `map[string]any` keys (SARIF properties) in unspecified order, making every consumer that snapshots or diffs SARIF output flaky. Affects `Report.MarshalJSON`, `Report.PrettyJSON`, `Report.PrettyJSONFiltered`, `Report.WriteJSON`, `Finding.LineJSON`, `Finding.WriteJSON`, `Report.ToSARIFWithOpts`, and `Report.WriteSARIFWithOpts`.
- **`Finding.Equal()` tag-order sensitivity** — Was using `slices.Equal` for tag comparison, making findings with the same tags in different order compare as unequal. The documentation and ROADMAP always claimed order-insensitive equality. Fixed with a `tagsEqual` helper that clones, sorts, then compares. This is a semantic behavior change that aligns the implementation with the documented contract.
- **`FlightRecorderHook` concurrent `WriteTo` race** — `runtime/trace.FlightRecorder.WriteTo` is not concurrency-safe. Two slow stages firing snapshots concurrently would produce an "already in progress" error. Fixed by adding a `writeMu sync.Mutex` that serializes WriteTo calls.
- **`sanitizeFilename("")` produced malformed filenames** — Empty reason strings produced filenames like `go-finding-trace-000-.trace`. Now returns `"snapshot"` as the default reason label when sanitization yields an empty string.

## [1.4.1] - 2026-07-28

Internal refactoring and test-infrastructure sweep. No public API changes.

### Changed

- **Code duplication eliminated (zero clones at `-t 5`).** Extracted internal helpers that the same pattern was reinvented across files: `must[T any]` generic (errors.go, used by `MustParseCategory`/`MustParseSeverity`/`Builder.MustBuild`), `marshalJSONString` (json.go, used by 4 marshal-to-string methods), `decodeConfig` (pipeline/config_file.go), package-level `fixEditJSON` wire type (pipeline/fix_edit.go), and `must(err)` in examples.
- **`Severity.Badge()` now derives from `Emoji()`.** Single switch instead of duplicate 4-case logic. `Severity.PriorityString()` uses a `severityPriorities` map lookup instead of a switch.
- **Test setup consolidated.** `NewParallelGomega(t *testing.T)` helper added to `testutil_test.go` in all 4 modules, replacing 112 occurrences of the `t.Parallel()` + `NewWithT(t)` two-line boilerplate.
- **`paralleltest` linter disabled (again).** Commit `59121e8` had inadvertently re-added `paralleltest` to the `enable` list in `.golangci.yml` (with a comment stating it _should_ be disabled), causing `golangci-lint` to report 93 false positives. Moved to the `disable` list where the documentation intended. The linter cannot trace `t.Parallel()` through the `NewParallelGomega` helper wrapper.
- Go module dependencies and Nix flake toolchain versions synchronized across all sub-modules.

### Fixed

- **7 pipeline lint issues resolved.** `Pipeline.Run()` decomposed via `runVerification()` extraction (funlen). `Pipeline.runIteration()` decomposed via `applyStage()` extraction (funlen). `applyTriage` complexity reduced via `shiftFindingsPositions()`/`shiftFindingSlice()` extraction (gocognit). `goast/provider.go` exhaustruct fixed with explicit zero values. `convenience.go` `errgroup.Wait()` error wrapped (wrapcheck). `fix_applier_test.go` gosec G703 suppressed with justification. `saboteurProvider` unused receivers removed (revive).
- **Release gate unblocked.** Lint (`nix run .#lint`) was failing the pre-release checklist due to the `paralleltest` config drift above; now passes cleanly across all 4 modules.

## [1.4.0] - 2026-07-26

Error classification integration, community readiness infrastructure, and documentation accuracy sweep. The core module gains its first production dependency (`go-error-family`) to enable unified error classification across consumer codebases.

### Added

- **`FindingError.ErrorCode()` and `FindingError.ErrorFamily()`** — Integration with `go-error-family` library. `ErrorCode()` returns `"finding.<category>"`; `ErrorFamily()` maps `ErrorCategory` to error families (Rejection, Conflict, Transient, Infrastructure) for use with `errorfamily.Classify()`.
- `SECURITY.md` — Vulnerability reporting policy (GitHub private advisories, 72h SLA).
- `CODE_OF_CONDUCT.md` — Contributor Covenant v2.1.
- `.github/ISSUE_TEMPLATE/bug_report.md` — Structured bug report template with environment section.
- `.github/ISSUE_TEMPLATE/feature_request.md` — Feature request template with API sketch.
- `.github/ISSUE_TEMPLATE/config.yml` — Issue chooser with security advisory contact link.
- `.github/PULL_REQUEST_TEMPLATE.md` — PR checklist with GOEXPERIMENT test/lint commands.
- **ADR #15** — `go-error-family` as core dependency decision documented.

### Changed

- README: Added prominent `GOEXPERIMENT=jsonv2` prerequisite block at top of Installation.
- README: Removed `GOPRIVATE` warning (no longer needed once public).
- README: Added `## Support` section (MIT, best-effort, latest-minor-only policy).
- README: Fixed stale version references (`v1.2.1` → `v1.3.0`, `"1.2.0"` → `"1.3.0"`).
- README: Removed redundant `Requires Go 1.26 or later.` line (now covered by Prerequisites block).
- SECURITY.md: Replaced stale hardcoded version table with prose-only "latest minor release" policy.
- GitHub repo: Set description and 11 topics for discoverability.
- `PUBLIC_OR_PRIVATE.md`: Corrected false "public" banner (repo was private; GitHub API confirmed).
- Dependabot: Expanded to track all 4 sub-modules (core, pipeline, analysis, CLI).
- Removed inline `exhaustruct` `nolint` comments superseded by `.golangci.yml` config exclusions.
- Extracted `writeSuggestionLine` helper to reduce `FormatText`/`FormatTextRich` duplication.
- **Retired "zero external deps" principle** — Core module now depends on `go-error-family`. All docs, README, AGENTS.md, and ADR #9 updated to reflect a small, deliberate dependency surface instead of zero-dep claims.

### Fixed

- Replaced deprecated `ToSARIFFiltered` reference in godoc with `ToSARIFWithOpts`.
- `doc.go`: Fixed stale category count ("14 predefined" → "16 predefined").
- `doc.go`: Added `ErrorCode()` and `ErrorFamily()` documentation to Error Handling section.
- `FEATURES.md`: Fixed stale linter count ("89 mappings" → "84 mappings" in two places).
- `go.mod`: Moved `go-error-family` from indirect to direct require block.

## [1.3.0] - 2026-07-22

Consumer-driven API improvements. 11 additive APIs eliminating the boilerplate every consumer independently reinvents, without breaking changes to existing types or signatures. Driven by a full audit of all 22 consumer projects.

### Added

- **`NewReportFromFindings(tool, findings)`** — One-step report creation (NewReport + AddFindings + ComputeSummary). Eliminates the 4-line boilerplate in 5+ consumers.
- **`Builder.BuildOrDefault() Finding`** — Returns zero-value Finding on validation error instead of panicking. Eliminates the SafeBuildFinding / buildFinding error-swallowing pattern.
- **`Template` type** — Pre-configured builder factory: stamp tool name, category, fix strategy, and tags once, then build many findings with `Build(rule, msg, sev, pos)`. Eliminates `newMigrationFinding` / `buildFixableFinding` patterns.
- **`FilePos(file FilePath) Position`** — Constructor for file-level positions (Line=0, Offset=-1). For findings that apply to an entire file (config issues, project checks).
- **`SeverityFromLevel(level, fallback) Severity`** — Maps severity strings (canonical + aliases) to Severity, returns fallback for unknown. Eliminates consumer-side `mapSeverity()` switches.
- **`Severity.PriorityString() string`** — Reverse mapping: Critical→"critical", Error→"high", Warning→"medium", Info→"low".
- **`FormatTable(w, findings)`** — Severity-badged table output (SEVERITY, LOCATION, RULE, MESSAGE columns).
- **`FormatTextRich(w, findings)`** — Rich text output with emoji severity badges, category display, and 💡 suggestion prefix. `FormatText` retains the original `[SEVERITY]` format for backward compatibility.
- **`ApplySimpleFixes(findings) map[FilePath][]SimpleFixResult`** — BeforeCode→AfterCode string replacement in core package. 80% case for consumers that don't need the full pipeline FixEngine.
- **`CheckBinary(name) (string, error)`** — Wraps `exec.LookPath` with finding error. Standardizes the "run CLI tool → parse JSON" pattern.
- **`RunCmd(ctx, name, args) ([]byte, error)`** — Wraps `exec.CommandContext` with finding error. For external tool integration.

### Changed

- **Builder default Confidence is now `ConfidenceFull`** — `NewBuilder` now passes `ConfidenceFull` (1.0) instead of `0` (ConfidenceNone). More semantically correct for deterministic static analysis. Override with `.WithConfidence()`.
- **`validateIdentity()` uses `HasFile()` instead of `IsValid()`** — File-level positions (File set, Line=0) now pass validation. `Position.IsValid()` still requires Line>0 for backward compatibility. Use `HasFile()` for file-only checks.
- **Severity aliases expanded** — Added "optional"→Info and "crit"→Critical to the pre-registered alias map.
- **DefaultLinterRegistry expanded** — Added 20+ new golangci-lint linters (gofumpt, nolintlint, depguard, nakedret, bidichk, tagliatelle, mirror, etc.).

## [1.2.1] - 2026-07-19

Post-v1.2.0 correctness, security, and dependency modernization release. 40 commits covering 15+ bug fixes (range geometry, thread safety, SARIF/LSP round-trip fidelity, path traversal), `encoding/json/v2` migration, go-output v0.30.4, dependabot PR merges, and comprehensive skills audit sweep (8 review reports + 2 architecture diagrams).

### Fixed

- **`Range.containsByOffset` respected Offset=0 sentinel** — Previously used `> 0` instead of `>= 0` for the end offset upper bound, causing single-point ranges at offset 0 and ranges with `End.Offset = -1` (unset sentinel) to incorrectly contain all higher offsets. Now uses `EndOffsetOrStart()` for consistency with overlap/intersection logic.
- **`Report.UnmarshalJSON` data race** — Wrote to struct fields without acquiring the write lock. Now uses `withLock` for thread safety, consistent with all other Report methods.
- **`Finding.IsSuppressedAt` inconsistent with `Suppression.IsActive`** — `IsSuppressedAt` didn't validate the suppression (missing `IsValid` check), so findings with invalid suppressions (e.g., missing `Rule`) were incorrectly treated as suppressed. Now delegates to `Suppression.IsActive`.
- **SARIF `Suppression.Rule` lost in round-trip** — Export didn't include `Suppression.Rule`; import fabricated it from `Finding.Rule`. Now exports/imports via `go-finding/suppression-rule` property.
- **SARIF `Suppression.Reason` fabricated from Kind** — Import used the Kind string as Reason when Reason was empty. Now exports/imports the actual Reason via `go-finding/suppression-reason` property.
- **SARIF `Confidence` stored unclamped** — Export property used raw `f.Confidence` instead of `f.NormalizedConfidence()`, allowing invalid values (>1.0) to round-trip. Now clamps to [0.0, 1.0].
- **LSP `SeverityCritical` lost in round-trip** — LSP collapses `critical` and `error` into the same severity. Now stores the exact severity in `LSPDiagnosticData.Severity` and restores it in `FromLSP`.
- **`correlateByProximity` false correlations at Line=0** — Findings with `Position.Line == 0` (unset) correlated with confidence 1.0 due to zero line difference. Now skips findings without line info.
- **Pipeline metrics dropped on error/cancel paths** — `metricsResult` was only set on the success path. Now set before every early return so metrics are available for debugging failures.
- **Pipeline `TotalIterations` zero on error/cancel** — Was only set after the loop exited normally. Now set before every early return.
- **Pipeline `TotalDetected` counted duplicates** — Used raw `p.findings` (accumulated across iterations) instead of the deduplicated set. Now uses `collectAllFindings` for accurate unique count.
- **Pipeline StageAfter hook errors silently discarded** — All `StageAfter` and `StageBefore Verify` hook errors were swallowed with `_ =`. Now propagated, aborting the pipeline per documented contract.
- **Pipeline `StageAfter Triage` conflicts always 0** — Passed `iter.Conflicts` before conflict detection ran. Conflicts are now correctly reported in `StageAfter Apply` (where they're computed), and the `Conflicts` field doc updated.
- **Pipeline suggest findings not shifted after direct fixes** — Only `iter.findings` was shifted; `iter.suggest` retained stale line numbers. Now both are shifted.
- **Pipeline provider errors during byte-level conflict detection swallowed** — Errors were logged but never surfaced to callers. Now recorded in `result.PartialErrors`.
- **Pipeline silent file read failure in conflict detection** — When a file couldn't be read, all fixes bypassed conflict detection with no log entry. Now logs a warning.
- **SARIF `Position.Offset` lost in round-trip** — Byte offsets were not preserved through SARIF export/import. Now carried via `go-finding/start-offset` and `go-finding/end-offset` properties.
- **LSP round-trip lost Snippet, Suppression, Metadata, RelatedRef.FindingID** — `LSPDiagnosticData` now carries these fields; `FromLSP` restores them. `RelatedRef.FindingID` is preserved instead of regenerated.
- **`resolveLineCol` discarded underlying error** — Replaced detailed error (line/col/contentLen context) with bare sentinel. Now wraps with `fmt.Errorf("%w: %w", ...)`.
- **Provider errors lack finding context** — FixEngine `resolveErrors` now include the finding ID in the error message.
- **Pipeline `collectAllFindings` called twice** — When `VerifyAfterFix` was enabled, findings were collected once for verification and again for `TotalDetected`. Now cached.
- **Path traversal vulnerability in `filterByFileEdits`** — Findings with `Position.File = "../../etc/passwd"` caused reads outside rootDir. Now uses shared `resolveSafePath` containment check.
- **TOCTOU race in `groupFindingsBySafePath`** — Used unresolved path as map key; symlink swap between validation and I/O could redirect writes outside rootDir. Now uses resolved (symlink-evaluated) path as map key.
- **FixApplier rollback errors silently swallowed** — `Restore` and `RollbackAll` errors were discarded with `_ =`. Now propagated to the caller so file corruption is visible.

### Changed

- **`applyTriage` signature updated** — Now accepts `*PipelineResult` to record provider errors in `PartialErrors`.
- **Error messages normalized** — Retry config errors now use lowercase field names consistent with config errors. Missing colons added to error wrapping in CLI config and pipeline config file.
- **`dedupKey` duplication eliminated** — Extracted `positionDedupKey` helper in `merge.go`.
- **`Suppression.IsExpired` boundary documented** — At the exact `ExpiresAt` instant, the suppression is still active (expired only strictly after). Documented and tested.
- **SARIF `sarifRegion.Snippet` spec compliance fixed** — Changed `Snippet string` to `Snippet *sarifArtifactContent` per SARIF 2.1.0 §3.30.13 (object form `{"text": "..."}`). Custom `UnmarshalJSON` accepts both the spec-compliant object and bare string (backward compat).
- **Migrated to `encoding/json/v2`** — All 9 JSON-handling files across all modules now use `encoding/json/v2`. Requires `GOEXPERIMENT=jsonv2` env var for all Go tool invocations (set automatically in all nix apps and devShells).
- **go-output v0.30.1 API migration** — `cmd/go-finding/output_adapter.go` updated: `TableData`→`Table`, `NewTableData`→`NewTable`, `RenderTableData`→`RenderTable`.
- **GOEXPERIMENT=jsonv2 propagation in flake.nix** — Added `export GOEXPERIMENT=jsonv2` to all 9 `mkApp` scripts and `devShells.ci`.

## [1.2.0] - 2026-07-06

Internal refactor release: new shared `lockutil` package, test suite defragmentation, and version constant sync.

### Added

- **`lockutil` package** — Generic helpers `lockutil.Locked(sync.Locker, fn) T` and `lockutil.RLocked(*sync.RWMutex, fn) T` that eliminate `m.mu.Lock()/defer m.mu.Unlock()` boilerplate around short critical sections. Returns generic `T` so callers can return values directly; use `struct{}` for side-effect-only sections. Stdlib-only, follows `gotoken` precedent as a public shared utility subpackage.

### Changed

- **Mutex boilerplate migrated to `lockutil`** — Report, Metrics, DetectorRegistry, LinterRegistry, FileBackup, GoASTProvider, CLI detector/fix-provider registries, and SARIF import now use `Locked`/`RLocked` instead of manual `Lock()/defer Unlock()` patterns. Behavior-preserving refactor; no public API changes beyond the new `lockutil` import.
- **`PrettyJSONFiltered` TOCTOU fixed (again)** — Consolidated the previous manual fix into the `withReadLock` helper, making the single-snapshot guarantee structural rather than convention-based.
- **SARIF import guard inlined** — `checkSARIFReadContext` unexported helper replaced with direct `ctx.Err()` guard at each entry point. No API change.
- **`lineProviderOffset` extracted** — Shared helper deduplicating offset resolution between insertion and replacement edit builders in the fix provider chain.

### Fixed

- **`version.go` synced to 1.2.0** — Was stale at `1.0.0` despite the v1.1.0 release. The `Version` constant now reflects the actual release.

### Test Suite

- **Test defragmentation** — Eliminated `_extra_test.go`, `_bugfix_test.go`, and `coverage_test.go` naming. All tests consolidated into canonical `<subject>_test.go` files (e.g., `assert_extra_test.go` → `testutil_test.go`, `coverage_test.go` → `validate_test.go`). Shared helpers centralized in `testutil_test.go`. No production code affected.

## [1.1.0] - 2026-07-06

Multi-module workspace split, type safety improvements, and SARIF/LSP fidelity.

### Breaking Changes

- **`Position.File` type changed from `string` to `FilePath`** — Compile-time safety preventing file path mixups with IDs, rule names, and tool names. String literals auto-convert (`Position{File: "main.go"}` still compiles). String variables need explicit wrapping: `Position{File: FilePath(someVar)}`. Constructors `Pos`, `NewRange`, `NewRangePtr` now accept `FilePath`. JSON serialization unchanged.
- **`GroupByFile` return type changed** — Now returns `map[FilePath][]Finding` instead of `map[string][]Finding`.
- **`ByFile` parameter type changed** — Now accepts `FilePath` instead of `string`.
- **`FindingError.File` type changed** — Now `FilePath` instead of `string`.
- **`FixGroup.File` type changed** — Now `FilePath` instead of `string`.
- **`Suppression.Rule` type changed** — Now `RuleName` instead of `string`.
- **CLI `-severity` renamed to `-min-severity`** — Deprecated alias `-severity` retained for backward compatibility.
- **`ToSARIFFiltered` / `WriteSARIFFiltered` removed** — Use `ToSARIFWithOpts(WithMinSeverity(sev))` / `WriteSARIFWithOpts(w, WithMinSeverity(sev))` instead.
- **`FromLSP` parameter changed** — First parameter is now `FilePath` instead of `string`.

### Added

- **Multi-module Go workspace** — Project split into 4 independently versioned modules coordinated via `go.work` + `replace` directives:
  - Core (`.`): stdlib-only Finding domain model (zero external production deps)
  - Pipeline (`pipeline/`): detect→triage→fix→verify loop
  - Analysis (`analysis/`): go/analysis adapter
  - CLI (`cmd/go-finding/`): govet + staticcheck detectors
- **`pipeline.Detect()`** — One-shot concurrent detection without the full pipeline. The simplest entry point for running detectors.
- **`pipeline.ApplyToContent()`** — Apply fixes to in-memory `[]byte` content without filesystem operations.
- **`SARIFOption` pattern** — Functional options for SARIF export: `WithIncludeSuppressed()`, `WithMinSeverity(sev)`. New APIs: `ToSARIFWithOpts()`, `WriteSARIFWithOpts()`.
- **SARIF suppression round-trip** — Suppressed findings now emit SARIF `suppressions` arrays when `WithIncludeSuppressed()` is used, instead of being silently dropped. Full round-trip preserves Kind, Reason, Rule, and ExpiresAt.
- **`LSPDiagnosticData`** — `LSPDiagnostic.Data` field carries go-finding-specific data (ID, FixStrategy, Confidence, Category, Tags, code data) for lossless LSP round-trip.
- **`RelationKind.IsValid()`** — Validates relation kind against standard values.
- **`RelatedRef.IsValid()`** — Now checks both FindingID and Relation validity.
- **`Position.HasFile()`** — Check whether a file path is set, independent of line completeness.
- **`Position.OffsetUnknown = -1`** — Named constant for the "no offset" sentinel.
- **`Finding.IsInvalid()`** — Inverse of IsValid, deprecated `FilterInvalid` as alias.
- **Exported registry errors** — `ErrDetectorRegistered`, `ErrUnknownDetector` now exported for `errors.Is`.
- **CLI `--include-suppressed`** — Controls whether suppressed findings appear in SARIF output.
- **`docs/guides/fix-engine.md`** — Full guide covering standalone FixEngine, FixApplier, custom providers, and conflict detection.

### Fixed

- **`IsAutoFixable()` now normalizes FixStrategy first** — Was checking raw strategy, missing findings with empty strategy.
- **`Position.IsValid()` requires `Line > 0`** — Was returning true for positions with only a file path.
- **`columnCoincident` / `offsetCoincident`** — Renamed from `columnAdjacent`/`offsetAdjacent`. Fixed `> 0` to `>= 0` for inclusive boundary.
- **Staticcheck SA codes map to `CategoryCorrectness`** — Were incorrectly mapped to `CategoryStyle`.
- **`file_backup.go` race condition** — Changed `enabled bool` to `enabled atomic.Bool`.
- **`PrettyJSONFiltered` TOCTOU** — Now snapshots findings once instead of twice.
- **`FixApplier.ApplyWithDetails`** — Now returns only applied fixes instead of input fixes.
- **`groupFindingsBySafePath` O(N) syscalls** — Cached resolved root (O(N) → O(1)).
- **`escapeMarkdownCell` panic** — Fixed `maxLen > 0` to `maxLen > 3`.
- **`DetectorFunc.Name()` / `TransformerFunc.Name()`** — Changed from `"anonymous"` to `""`.
- **`retry.Config` validation** — `BaseDelay > 0` required when `MaxRetries > 0`.
- **`writeOutput` Close() error** — Now checked and propagated.
- **CLI SIGINT/SIGTERM handling** — Added `signal.NotifyContext` for graceful shutdown.
- **`gotoken.NodeByteRange`** — Added `start <= end` validation.
- **IntervalIndex complexity doc** — Corrected O(log n+k) to O(n+k).

### Deprecated

- **`ToSARIFFiltered(minSeverity)`** — Removed. Use `ToSARIFWithOpts(WithMinSeverity(sev))`.
- **`WriteSARIFFiltered(ctx, w, minSeverity)`** — Removed. Use `WriteSARIFWithOpts(ctx, w, WithMinSeverity(sev))`.
- **`FilterInvalid`** — Deprecated as alias for `Filter(IsInvalid)`.

### Removed

- `Report.Findings` field (use `FindingsSnapshot()`)
- `Report.Merge()` (use `MergeInto()`)
- `Config.OnStage` (use `StageHooks`)
- `Metrics.RecordFix()` (use `RecordFixes(1)`)
- `CountBySeverity()` free function (use `Report.CountBySeverity()`)
- `SeverityAliases()` (use `RegisterSeverityAlias()`/`LookupSeverityAlias()`)
- `GetCategory()` (use `CategoryOf()`)
- `ConflictInfo` (use `Conflict`)
- `LSPRelatedInfo` (use `LSPRelated`)
- `FindingProcessor` (use `FindingTransformer`)

## [1.0.0] - 2026-06-24

First stable release. The API is now frozen — future breaking changes require a major version bump.

### Breaking Changes — Deprecated APIs Removed

All APIs deprecated since v0.6.0–v0.9.0 have been removed. See `docs/MIGRATION_v1.0.md` for the full migration guide.

- **`Report.Findings` field unexported** → `findings`. Use `FindingsSnapshot()` for a deep copy, `All()` for iteration, `FindByID()` for single lookups, or `Len()` for count.
- **`Report.Merge()` removed** → Use `Report.MergeInto()` which returns a new Report without mutating the receiver.
- **`Config.OnStage` removed** → Use `Config.StageHooks` with `StageHook`/`StageHookFunc` for before/after events with abort capability.
- **`Metrics.RecordFix()` removed** → Use `Metrics.RecordFixes(1)` which batches multiple recordings in a single mutex acquisition.
- **`CountBySeverity()` free function removed** → Use `Report.CountBySeverity(sev)` method.
- **`SeverityAliases()` removed** → Use `RegisterSeverityAlias()` / `LookupSeverityAlias()`.
- **`GetCategory()` removed** → Use `CategoryOf()` (Go convention: no `Get` prefix).
- **`HasFix` free function removed** → Use `WithFix` (consistent with `BySeverity`, `ByCategory` naming).
- **`HasSuggestion` free function removed** → Use `WithSuggestion`.

### Added

- **Branded primitive types** — `ID`, `RuleName`, `ToolName`, `FilePath` in `branded_types.go`. Compile-time type safety preventing ID/Rule/Tool/File mixups. JSON marshals identically to string.
- **`Validate()` decomposition** — Monolithic `Validate()` split into 6 per-field validators (`validateIdentity`, `validateClassification`, `validateFix`, `validateReferences`, `validateSpatial`, `validateSuppression`).
- **Thread-safe `SeverityAliases`** — Global alias map guarded by `sync.RWMutex`. Use `RegisterSeverityAlias()` / `LookupSeverityAlias()`.
- **`FindingTransformer`** — Renamed from `FindingProcessor`/`Process()`. Pipeline uses `Config.Processors []FindingTransformer`.
- **`CategoryOf`** — Renamed from `GetCategory` (Go convention: no `Get` prefix).
- **`Conflict`** — Renamed from `ConflictInfo`. `AnalyzeConflicts() []Conflict`.
- **`LSPRelated`** — Renamed from `LSPRelatedInfo`.
- **CSV and TSV output formats** — CLI `-format csv` and `-format tsv` produce clean, machine-readable data exports via go-output's delimited writers.
- **Markdown output via go-output** — CLI `-format markdown` uses go-output's `MarkdownTable` renderer with auto-aligned column widths.
- **Format validation** — Unknown `-format` values now return a clear error listing all supported formats.

### Changed

- **`Report` custom JSON marshaling** — `MarshalJSON`/`UnmarshalJSON` implemented via internal `reportJSON` DTO to support the unexported `findings` field.
- **CLI adapter pattern** — `cmd/go-finding/output_adapter.go` adapts `[]Finding` → go-output's `TableData`. Root `finding` package stays dependency-free.
- **go-output dependency** — Added `github.com/larsartmann/go-output` v0.17.2 as CLI-only dep. Root `finding` package unaffected.

## [0.9.1] - 2026-06-18

### Fixed

- **Fuzz target naming collisions** — Three fuzz function names were regex substrings of other names, causing Go's `-fuzz` flag (which uses regex matching) to match multiple targets and refuse to run. `FuzzFilter` → `FuzzFilterPredicates`, `FuzzGroupBy` → `FuzzGroupByKey`, `FuzzApplyEditsToContent_Overlapping` → `FuzzApplyOverlappingEdits`. Seed corpus directories renamed to match. This broke `buildflow check test-fuzz` (20/23 targets failed).
- **FuzzMergeByPosition oracle** — The fuzz test asserted that equal dedup keys always merge to one finding. This was wrong: `DeduplicateByPosition` intentionally skips findings with empty `Position.File` (a position without a file is not a meaningful dedup key, see `merge.go:146`). Corrected the oracle to assert the real two-case semantics: merge only when keys match AND both files are non-empty.

### Changed

- **`sortEditsDescending` extracted** — Moved from inline `slices.SortFunc` call in `fix_engine.go` to a named helper in `fix_edit.go`. Pure internal refactor; no behavior change. The helper is used by both `FixEngine.ApplyWithConflicts` and the fuzz test harness.

### Added

- **`pipeline/export_test.go`** — Standard Go `export_test.go` bridge pattern. Centralizes the `rangeFix`, `offsetFix`, `offsetFixWithID` test helpers and exposes them to the external `pipeline_test` package via `ExportRangeFix` / `ExportOffsetFix` / `ExportOffsetFixWithID` wrappers. Eliminates helper duplication across `bdd_test.go` and `fix_engine_test.go`.

### Removed

- **9 unused transitive test dependencies from `go.sum`** — `testify`, `go-spew`, `gkampitakis/*`, `joshdk/go-junit`, `kr/*`, `maruel/natural`, `mfridman/tparse`, `pmezard/go-difflib`, `tidwall/gjson`. Canonicalized via BuildFlow's `go-mod-tidy` step (verified deterministic across two independent runs).

## [0.9.0] - 2026-06-18

### Fixed — Split-Brain Resolution (11 issues)

- **Position zero-value semantics (#1)** — `Position.Offset` now uses `-1` as the explicit "unset" sentinel. `Position{}` (zero value) has `Offset=0` meaning "byte 0" (valid). `IsZero()` and `HasOffset()` no longer return contradictory results. All constructors updated.
- **FixStrategy `""` vs `"none"` (#2)** — `NormalizeFixStrategy()` helper converts empty string to `FixStrategyNone`. Called by `Builder.Build()`, `Equal()`, and SARIF import. `Finding.Normalized()` method added for explicit normalization.
- **HasFix/Validate alignment (#3)** — `HasFix()` now requires `BeforeCode` or `AfterCode` for `FixStrategyDirect`, aligning with `Validate()`. Documented fixability lattice: `IsAutoFixable() ⟹ HasFix()`.
- **Category/Tags consistency (#4)** — `Validate()` now rejects findings where `Category` and `Tags` contain conflicting standard categories.
- **Summary.DurationMs removed (#5)** — Vestigial caller-set field removed. `Metrics.TotalDuration()` is the single source of truth for timing.
- **Finding identity documented (#6)** — `GenerateID()` output is canonical identity (ADR #12). `Key()` documented as fallback.
- **SARIF fix-region consistency (#7)** — `findingFixRegion()` now respects `Range.End` coordinates, matching the location region for multi-line findings.
- **SARIF property bag doctrine (#8)** — Documented boundary between `map[string]string` (public API) and `map[string]any` (SARIF wire format requirement).
- **Suppression validation (#9)** — `Suppression.IsValid()` now calls `Kind.IsValid()`, rejecting unknown kinds like `"bogus"`.
- **Category/Tag twin methods (#10)** — Documented `IsValid` (input validation) vs `IsStandard` (allow-list filtering) contract.
- **HasFix name shadowing (#11)** — Added `WithFix`/`WithSuggestion` as canonical `FilterFunc` names. Deprecated `HasFix`/`HasSuggestion` free functions.

### Added

- **Finding.Normalized()** — Returns a copy with `FixStrategy` canonicalized (empty → "none").
- **NormalizeFixStrategy()** — Helper that converts `""` to `FixStrategyNone`.
- **WithFix / WithSuggestion FilterFuncs** — Canonical replacements for deprecated `HasFix`/`HasSuggestion` free functions.
- **ADR #12** — Canonical finding identity definition.
- **ADR #13** — Category/Tags deprecation plan for v1.0.0.
- **splitbrain_test.go** — 6 integration tests verifying all critical split-brain fixes.

### Removed

- **Summary.DurationMs** — Field was caller-set, never computed by `ComputeSummary()`. Use `Metrics.TotalDuration()` instead.

### Changed

- **Position.IsZero()** — Now checks `Offset < 0` instead of `Offset == 0`. `Position{}.IsZero()` returns `false` (Offset=0 is valid byte offset).
- **Position.Offset** — Constructors (`Pos`, `NewRange`, `FromLSP`, SARIF import) now set `Offset: -1` when no byte offset is available.
- **HasFix()** — Returns `false` for `FixStrategyDirect` without code data (was `true`). Aligns with `Validate()`.

## [0.8.0] - 2026-06-17

### Added

- **Category.Compare()** — Lexicographic comparison method matching the Severity/Confidence pattern, for deterministic sorting.
- **SubstringProvider column-aware disambiguation** — When multiple occurrences of BeforeCode exist and Position.Column is set, picks the occurrence closest to the target byte offset (line + column). Previously used line-distance only.
- **LineShiftMap.ShiftedPosition / ShiftedRange** — Extends line shifting to full positions (line + column) and ranges (both endpoints). Single-line edits shift columns; multi-line edits shift subsequent lines.
- **ConfigFile integration tests** — End-to-end proof that ConfigFile → ResolveDetectors → Pipeline.Run works.
- **DetectorRegistry integration tests** — End-to-end proof that BuildAll → Pipeline.Run works.
- **Type alias backward compat test** — Verifies pipeline.Detector == finding.Detector compile-time identity.
- **FuzzCategoryForLinter** — Fuzz target for case-insensitive linter lookup with seed corpus.
- **Godoc examples** — IntervalIndex, MergeIter, DetectorRegistry, ConfigFile, LineShiftMap.
- **v1.0 Migration Guide** — `docs/MIGRATION_v1.0.md` with per-API migration instructions.
- **Benchmark regression CI gate** — `scripts/bench-check.sh` + CI job fails on >25% regression vs baseline.
- **Dependabot** — `.github/dependabot.yml` for gomod + github-actions.
- **CLI resolveFixProviders error surfacing** — Unknown fix provider names now return an error instead of being silently skipped.
- **CLI `-filter-generated-types` new detectors** — Added 7 new generator types from gogenfilter v3.2.0: `counterfeiter`, `easyjson`, `ent`, `go-swagger`, `gqlgen`, `mockery`, `msgp`. Registry now covers all 19 upstream `FilterOption` values.

### Changed

- **resolveFixProviders returns error** — Signature changed to `([]pipeline.FixProvider, error)`; unknown provider names surface `ErrUnknownFixProvider` with available names listed.
- **slices.Sorted modernization** — Replaced `slices.Collect(maps.Keys(m))` + `slices.Sort()` with `slices.Sorted(maps.Keys(m))` in 6 locations (correlate, fix_applier, partial, CLI registries).
- **LineShiftEntry.ByteDelta** — Added byte-delta tracking to LineShiftEntry for column shifting on single-line edits.
- **Pipeline shifts Position + Range** — `pipeline_detect.go` now calls `ShiftedPosition` and `ShiftedRange` instead of only `ShiftedLine`.
- **AGENTS.md trimmed** — Reduced from 399 → 115 lines by removing session changelogs (kept in git history + CHANGELOG.md).
- **RELEASE_CRITERIA.md rewritten** — Concrete pass/fail thresholds, owner-decision blockers documented.
- **OnStage unified into fireStageHook** — Eliminated the split brain where every stage boundary called both `notifyStage()` (legacy `OnStage` callback) and `fireStageHook()` (`StageHooks`). Now `fireStageHook` fires the legacy `OnStage` callback internally on `StageAfter` events, giving a single notification call per stage boundary. `Config.OnStage` is marked deprecated in favor of `Config.StageHooks` which provides before/after events with context and abort capability.
- **gogenfilter v3.1.0 → v3.2.0** — Picks up 7 new generator detectors (counterfeiter, easyjson, ent, go-swagger, gqlgen, mockery, msgp), absolute-path content detection fix, new `FilterWithContent`/`FilterDetailedWithContent` APIs (pre-read content to avoid redundant I/O), and `FilterResult.Is()` helper. `flake.nix` vendorHash updated.

### Removed

- **FixStrategyResolver + DefaultResolver** — Removed unused interface and its only implementation. `DefaultResolver.CanAutoApply` merely delegated to the existing `FixStrategy.CanAutoApply()` method, and the pipeline triage path never referenced the resolver (it calls `Finding.IsAutoFixable()` directly). Zero consumers, zero tests. Pre-v1.0 cleanup.
- **MiddlewareFunc / ComposeMiddleware / RunFunc** — Removed `pipeline/middleware.go` entirely. The middleware pattern was never wired into `Config` or `Pipeline.Run`, had zero consumers, and no test file existed. `StageHook` (per-stage before/after with abort capability) already covers the same cross-cutting concerns at finer granularity. Pre-v1.0 cleanup.

## [0.7.0] - 2026-06-09

### Added

- **Report.Findings deprecation (ADR 10)** — `Findings` field marked `// Deprecated:`. Internal code uses `readFindings()` helper for thread-safe access. CLI output uses `FindingsSnapshot()`. Field will be unexported in v1.0.0.
- **Pipeline stage hooks** — `StageHook` interface with `Before`/`After` methods, `StageHookFunc` adapter, `StageEvent` struct. Wired into detect, process, triage, apply, and verify stages. Pre-hook errors abort pipeline. Configured via `Config.StageHooks`.
- **LineShiftMap** — Byte-offset-aware line shift tracking after fix edits. `NewLineShiftMap(content, edits)` builds a shift map from `[]FixEdit`, `ShiftedLine(original)` maps original line numbers to post-edit line numbers. Uses `buildLineOffsetIndex` for O(1) line-to-byte lookup.
- **IntervalIndex[T]** — Generic sorted-scan interval overlap queries with O(log n + k) complexity. `NewIntervalIndex[T](intervals)` + `Query(start, end)` returns all overlapping `[Start, End)` intervals. Uses binary search cutoff + linear scan.
- **MergeIter()** — Streaming `iter.Seq[Finding]` merge that reads each report's findings under `RLock` via `readFindings()`, clones each finding, and supports early termination via iterator break.
- **ConfigFile** — `pipeline/config_file.go` with `ConfigFile` struct, `ConfigFromFile(path)`, `ConfigFromReader(r)`. Parses YAML/JSON config for timeout, max iterations, severity filtering.
- **DetectorRegistry** — Thread-safe named detector constructor registry with `Register`, `MustRegister`, `Build`, `BuildAll`, `Names`, `Has`. Sentinel errors with `%w` wrapping for typed error matching.
- **MiddlewareFunc / ComposeMiddleware** — Composable pipeline middleware pattern. `MiddlewareFunc func(next RunFunc) RunFunc` wraps pipeline execution for cross-cutting concerns. `ComposeMiddleware` chains multiple middleware in registration order.
- **FixStrategyResolver** — `FixStrategyResolver` interface with `CanAutoApply(strategy) bool`. `DefaultResolver` delegates to `FixStrategy.CanAutoApply()`.
- **Benchmark CI job** — Separate `benchmark` job in `.github/workflows/ci.yml` running `go test -bench`.
- **LinterRegistry** — `LinterRegistry` type for linter→category mapping inspection.
- **AnalyzerDetector** — `analysis.AnalyzerDetector` adapter wrapping `go/analysis.Analyzer` as a `Detector`.
- **Byte-level conflict tests** — Integration tests for `ByteLevelConflictDetection` config option.

### Changed

- **Detector moved to root package** — `Detector` interface, `DetectorFunc`, `NamedDetectorFunc` moved from `pipeline/` to root package. Pipeline re-exports as type aliases for zero-breaking-change backward compatibility.
- **SeverityAliases map** — `SeverityAliases` exported map with 9 common aliases (warn, high, medium, low, fatal, critical, note, advice, suggestion). `ParseSeverity` checks aliases after canonical names.
- **CategoryForLinter registry** — 70+ linter→category mappings with case-insensitive lookup. Thread-safe via `sync.RWMutex`.
- **ToolAdapter[O] generic** — `NewToolAdapter(name, run, parse, convert)` wires exec→parse→convert pipeline.
- **Pipeline single-use enforced** — `Pipeline.Run()` returns `errAlreadyRan` on second invocation.
- **Merge deprecated** — `Report.Merge()` deprecated in favor of `MergeInto()`.

### Fixed

- **linterCategories data race** — Added `sync.RWMutex` guarding the global `linterCategories` map. Race detector now passes 30/30 consecutive runs.
- **Test deduplication** — Unified 5 separate test groups into table-driven tests (RegisterLinterCategory, DeduplicateStrategies, Metrics_RecordFixes, Finding_Validate).
- **Dead code removed** — `Iteration.Failed` field removed (declared but never written).
- **5 lint issues fixed** — perfsprint (errors.New), revive (MiddlewareFunc rename), modernize (slices.Backward), wsl_v5, gci formatting.

## [0.6.1] - 2026-06-09

### Fixed

- **FixApplier symlink-based path traversal** — Added `filepath.EvalSymlinks` to path validation, preventing malicious symlink chains from escaping the root directory (e.g., `../../../etc/passwd` via symlink).

### Changed

- **`finding.go` split into 4 focused files** — `finding.go` (95 lines, core struct + builder), `finding_methods.go` (162 lines, methods), `finding_validate.go` (118 lines, validation), `finding_equal.go` (149 lines, equality). No API changes.
- **`position.go` split into 2 focused files** — `position.go` (78 lines, Position type), `range.go` (368 lines, Range type). No API changes.
- **SARIF test constants** — Replaced raw property strings in tests with named constants for consistency.
- **`merge_test.go` split** — Separated correlation tests into `merge_correlate_test.go` (123 lines).
- **flake.nix `proxyVendor`** — Added `proxyVendor = true` for deterministic Go module downloads in Nix sandbox.

### Added

- **ADR 10** — `Report.Findings` encapsulation strategy documented in `docs/architecture-decisions.md`.

## [0.6.0] - 2026-06-08

(See [0.5.0] entry — v0.6.0 tag was created from same commit series as v0.5.0 release artifacts.)

## [0.5.0] - 2026-06-08

### Added

- **`RelatedRef.Range *Range`** — Span-based related locations with deep-copy in `Clone()`, validation in `Validate()` (rejects inverted ranges), and value equality via `equalRelated()`. Enables full geometric relationships between related findings.
- **`LSPDiagnosticTag` type and constants** — `Unnecessary = 1` and `Deprecated = 2` per LSP spec; `Tags []LSPDiagnosticTag` field on `LSPDiagnostic`.
- **LSP diagnostic tag round-trip** — `FromLSP()` preserves `DiagnosticTag` values in `Metadata["go-finding/lsp-diagnostic-tags"]` as comma-separated integers and reconstructs them on `ToLSP()`.
- **LSP related information range support** — `ToLSP()` emits proper `LSPRange` end positions from `rel.Range`; `FromLSP()` reconstructs `RelatedRef.Range` from LSP related info end positions.
- **SARIF `region.snippet` support** — `SarifRegion.Snippet` field enables native SARIF snippet round-trip. `findingRegion()` populates it on export; `applySarifPosition()` reads it back on import. The property bag `go-finding/snippet` remains as fallback.
- **SARIF related location range round-trip** — `sarifRelatedLocs()` writes `RelatedRef.Range` end coordinates into related location regions; `findingFromSarResult()` reconstructs `RelatedRef.Range` from `physicalLocation.region` end coordinates.
- **JSON Schema `relatedRef.range`** — `docs/schemas/finding.schema.json` now includes `range` property on the `relatedRef` definition referencing `#/$defs/range`.
- **Comprehensive `doc.go`** — Expanded from ~40% to full package documentation covering Diff, SARIF, LSP, Formatting, JSON, ID Generation, Pipeline, Fix Providers, Filtering, Merging, Correlation, Suppression, Error Handling, and Known Limitations.
- **`TriageFunc` config option** — Customizable triage function in `pipeline.Config`. `DefaultTriageFunc` preserves existing behavior; `TriageResult` moved to `config.go` for discovery.
- **`ByteLevelConflictDetection` config option** — Opt-in byte-level edit conflict detection that groups fixes by file, reads content, runs `FilterConflictingEdits` per file, and gracefully degrades on read errors.
- **`DeduplicateBy.String()`** — Human-readable names for deduplication strategies (`"id"`, `"position"`, `"rule"`). Consistent with every other named type in the codebase.
- **Correlate complexity documentation** — `Correlate` godoc now includes a `# Complexity` section explaining O(n·k) normal case, O(k²) worst case for dense clustering, and the `maxCorrelations` (10,000) cap behavior.
- **Position semantic documentation** — `Position` doc comment now explicitly documents the `Offset=0` semantic trap where `IsZero()` and `HasOffset()` both return true for the zero value.

### Fixed

- **flake.nix infinite recursion** — `goPkg = goPkg` self-reference caused `nix build` to infinite-loop. Fixed to `goPkg = pkgs.go_1_26`.
- **flake.nix duplicate `checks.build`** — `checks.build` was defined in two separate `perSystem` blocks causing immediate Nix evaluation error. Consolidated into single `checks = { format; build; test; }` block.
- **flake.nix stale `vendorHash`** — Updated to match current `go.sum`.
- **`Equal()` for `RelatedRef`** — Previously used `slices.Equal()` which compared `*Range` pointers by identity. Now uses `equalRelated()` for proper deep value comparison.
- **`Clone()` for `RelatedRef`** — Previously did shallow `copy()` of the `Related` slice, sharing `*Range` pointers between original and clone. Now deep-copies each `RelatedRef` and its `Range` pointer.

### Changed

- **flake.nix** — Added `maintainers = [ lib.maintainers.larsartmann ]` to meta block.
- **`.golangci.yml`** — Normalized indentation from 4-space to 2-space YAML.

### Testing

- Coverage: root 97.2%, analysis 98.5%, pipeline 93.6%, internal/detectors 95.9%, cmd/go-finding 90.7%. Total 93.4%.
- 9+ new tests: `TestToLSP_RelatedWithRange`, `TestFromLSP_RelatedWithRange`, `TestFromLSP_PreservesDiagnosticTags`, `TestSARIF_RoundTrip_RelatedRefRange`, `TestSARIF_RegionSnippet`, `TestFinding_Equal_RelatedRefRange`, `TestFinding_Validate_InvertedRelatedRange`, `TestDeduplicateBy_String`, plus `TestClone` updated and schema round-trip extended.
- `go test -race -count=1 ./...` passes; `golangci-lint run ./...` reports 0 issues.

## [0.4.3] - 2026-06-01

### Changed — Breaking

- **`context.Context` added to SARIF I/O functions** — All SARIF export and import functions that perform I/O now accept `context.Context` as their first parameter, enabling cancellation and timeout control in production pipelines. Affected signatures:
  - `Report.WriteSARIF(ctx context.Context, w io.Writer) error`
  - `Report.WriteSARIFFiltered(ctx context.Context, w io.Writer, minSeverity Severity) error`
  - `FindingsFromSARIF(ctx context.Context, data []byte) ([]Finding, error)`
  - `FindingsFromReader(ctx context.Context, r io.Reader) ([]Finding, error)`
  - `Report.WriteTo(w io.Writer)` unchanged (already satisfies `io.WriterTo`).

## [0.4.2] - 2026-06-01

### Added

- **`Summary.FilesScanned`** — New field on `Summary` counting total files scanned (including clean files with no findings). Complements existing `FilesAffected` which only counts files with findings. Tools can now display "X files scanned" in success paths where `FilesAffected` is 0.
- **`Category.IsValid()` rune-level validation** — Enforces `^[a-z][a-z0-9-]*$` format at the rune level, rejecting invalid categories like `"Security"`, `"UPPERCASE"`, or `"has space"`. Uses De Morgan's law for clarity (per staticcheck QF1001).

### Fixed

- **`ErrorCategory.IsValid()` format validation** — Now enforces the same lowercase-hyphenated convention as `Category.IsValid()`. Rejects typos like `"Validation"`, `"HAS_SPACE"`, `"has space"`.
- **SARIF import FixStrategy assignment** — Results with actual code replacements (`InsertedText`) were incorrectly assigned `FixStrategySuggest`, preventing pipeline auto-application. Now: replacements → `FixStrategyDirect`, description-only → `FixStrategySuggest`.
- **`FileBackup` permission preservation** — `Restore()` now uses the original file's `Mode()` instead of hardcoded `0o600`. Previously, restoring an executable (`0755`) would strip its execute bits.
- **CLI config duration parse errors** — `toPipelineConfig()` now returns `(Config, error)` and reports malformed duration strings (e.g., `"abc"`) instead of silently falling back to defaults.
- **`Config.DetectorTimeouts` negative validation** — `Config.Validate()` now rejects negative per-detector timeouts, which would previously panic at runtime via `context.WithTimeout`.
- **CLI redundant timeout wrapping** — Removed double `context.WithTimeout` application. Pipeline already wraps `ctx` internally; the CLI's second wrap with the same value was redundant and misleading.

### Changed

- **Upgraded `gogenfilter/v3`** — Multiple dependency bumps to pre-release versions with critical fixes for generated-file detection edge cases.
- **Nix flake overhaul** — Full `nix-review` improvements including `buildGoModule` with fileset source filtering, E2E sandbox fix, and added `lake.nix` configuration.
- **Code quality improvements** — `byFindingID` uses `cmp.Compare` instead of manual three-way comparison; `FixApplier` control flow made more explicit; `gci` formatting fixed in `pipeline/config.go`.
- **Deprecated `justfile` removed** — Build automation fully migrated to Nix flakes. `AGENTS.md` and `CONTRIBUTING.md` updated to reference `nix` commands.

## [0.4.1] - 2026-05-27

### Added

- **`GeneratedFileFilter` pipeline processor** — `pipeline.GeneratedFileFilter` wraps `gogenfilter/v3` to automatically remove findings from auto-generated Go source files (sqlc, protobuf, mockgen, templ, wire, stringer, deepcopy-gen, oapi-codegen, moq, go-enum, and generic `// Code generated by` comments). Configurable per-generator type with include/exclude glob patterns. Gracefully keeps findings when source files can't be read.
- **`gogenfilter/v3` dependency** — `github.com/LarsArtmann/gogenfilter/v3 v3.0.2` for two-phase generated file detection (filename-first zero-I/O, then content-based).
- **CLI `-filter-generated` flag** — Enables generated file filtering in the pipeline.
- **CLI `-filter-generated-types` flag** — Comma-separated generator types to filter (default: `all`). Supported: `all`, `sqlc`, `templ`, `go-enum`, `protobuf`, `oapi-codegen`, `deepcopy-gen`, `wire`, `moq`, `mockgen`, `stringer`, `generic`.
- **CLI `-generated-exclude` flag** — Comma-separated glob patterns for files to always exclude.
- **CLI `-generated-include` flag** — Comma-separated glob patterns restricting filter scope.
- **Config file `filterGenerated` field** — YAML/JSON config file support for generated file filtering.
- **Config file `filterGenTypes` field** — Generator types in config file (default: `"all"`).
- **Config file `generatedExclude` field** — Exclude patterns in config file.
- **Config file `generatedInclude` field** — Include patterns in config file.

### Changed

- **CLI `run()` refactored** — Extracted `cliFlags` struct, `parseFlags()`, and `writeResults()` functions. Resolves pre-existing `funlen` lint violation (148 → <120 lines).
- **`.golangci.yml` depguard** — Added `github.com/LarsArtmann/gogenfilter` to allow-lists.

## [0.4.0] - 2026-05-27

### Fixed

- **golangci-lint issues resolved (7 → 0)** — Fixed contextcheck, exhaustruct, revive, and other lint warnings across pipeline, analysis, and CLI packages.
- **`FormatPartialErrors` error wrapping** — Changed from `%v` with `fmt.Sprintf` to `%s` with `strings.Join` for cleaner error messages.
- **`.golangci.yml` config** — Replaced non-existent `gomodguard_v2` with `gomodguard`.

### Changed

- **Named constants for magic strings** — Extracted SARIF property keys (`go-finding/edit/*`), LSP severity key, conflict reason strings, merge tool names, analysis relation constant, and parseSeverity to named constants. All previously hardcoded strings are now exported package-level constants.
- **`ErrPositionUnresolvable` sentinel error** — Replaced 4 `//nolint:nilerr` directives in `fix_provider.go` with an explicit sentinel error. Callers already skip on error, so behavior is unchanged but now type-safe and lint-clean.
- **`parseSeverity` refactored** — Switch/case replaced with map-based lookup using `Severity.String()`.

### Added

- **Per-detector timeouts** — `DetectorTimeouts` config for individual detector timeout control.
- **`FormatText` / `FormatMarkdown`** — Human-readable and markdown table output formatters.
- **`Finding.ToDiagnostic()`** — Reverse conversion from Finding to `analysis.Diagnostic`.
- **`slog` logging integration** — Pipeline now accepts `*slog.Logger` for structured logging.
- **`OnStage` callback** — Pipeline lifecycle callback for monitoring stage transitions.
- **`FixApplier` lifecycle** — `Start()` and `Stop()` methods for provider lifecycle management.
- **Godoc examples** — Added `ExampleDiff`, `ExampleFormatText`, `ExampleFormatMarkdown`.

## [0.2.1] - 2026-05-01

### Breaking

- **`NewFinding` signature changed** — Now takes 6 parameters (added `confidence float64` as last parameter). Confidence is clamped to `[0.0, 1.0]`.
  - Before: `NewFinding(rule, toolName, message string, severity Severity, pos Position) Finding`
  - After: `NewFinding(rule, toolName, message string, severity Severity, pos Position, confidence float64) Finding`
- **`Builder.Build()` returns `(Finding, error)`** — Previously panicked on invalid state. Now returns `ErrInvalidBuilder` sentinel error. Callers handling the old panic must now check error.
  - Added `Builder.MustBuild() Finding` for panic-on-error use cases (tests, examples).
- **`WithTags` parameter type changed** — From `...string` to `...Tag`. Use the new `Tag` type constants (`TagSecurity`, `TagPerformance`, etc.) or `Tag("custom")`.
- **`Finding.Tag string` deprecated** — Use `Finding.Tags []Tag` instead. `Tag` field still exists for backward compatibility but will be removed in v1.0.
- **`FixApplier` applies fixes in deterministic order** — Files are now sorted alphabetically before applying fixes. Previously, map iteration order was non-deterministic.

### Added

- **`Tag` type with standard constants** — `type Tag string` with `TagSecurity`, `TagPerformance`, `TagStyle`, `TagCorrectness`, `TagBug`, `TagDeprecated`, `TagDocumentation`, `TagComplexity`, `TagTest`, `TagBuild`.
- **`Finding.Tags []Tag` field** — Multi-tag classification support (JSON: `"tags"`).
- **`Finding.Validate() error`** — Comprehensive structural validation (ID, Rule, ToolName, Message, Severity, Position, FixStrategy, Confidence range).
- **`Finding.Preview() string`** — Returns unified-diff-style preview of `BeforeCode`/`AfterCode` fix changes.
- **`Finding.HasCategory() bool`** — Convenience check for non-empty category.
- **`Report.Filter(predicates ...FilterFunc) *Report`** — Returns new filtered report.
- **`Report.Map(fn func(Finding) Finding) *Report`** — Returns new report with transformed findings.
- **`Report.WriteJSON(w io.Writer) error`** — Streaming JSON output without buffer allocation.
- **`Finding.WriteJSON(w io.Writer) error`** — Single-finding streaming JSON output.
- **`Builder.MustBuild() Finding`** — Panics on invalid builder; convenience for tests and examples.
- **CLI `-output` flag** — Write output to file instead of stdout.
- **CLI `-version` flag** — Print version and exit.
- **`Pipeline.CorrelateFindings` config** — Optional post-detection correlation stage.
- **`PipelineResult.Correlations`** — Correlation results when correlation is enabled.
- **`finding.Version` constant** — Programmatic version checking via `finding.Version` (`"0.2.1"`).
- **`examples/` directory** — Standalone examples (`basic/`, `builder/`, `pipeline/`) with compile checks.
- **`docs/integration-guide.md`** — Real-world tool integration guide.
- **`docs/release-procedure.md`** — Release process documentation.
- **`docs/architecture-decisions.md`** — 5 documented architectural decisions.
- **`FEATURES.md`** — Comprehensive, honest feature inventory.

### Changed

- **`Merge()` performance** — 2.4x faster with pre-allocated slice and batch appending (408K → 173K ns/op for 1000 findings).
- **`Report.AddFinding` / `Report.AddFindings`** — Now thread-safe via `*sync.Mutex`.
- **`FixStrategyAI` in `HasFix()`** — Now treated as Suggest-equivalent (requires `AfterCode`).
- **`Correlation` JSON tags** — Changed to camelCase (`findingIds`).
- **`DeduplicateByPosition` key** — Now includes `ToolName` for cross-tool deduplication.
- **`Severity.Compare`** — Total ordering for invalid severities via string comparison tiebreaker.
- **`Finding.Equal`** — Uses `floatEq` with 1e-9 epsilon for float comparison.
- **`Range.LineCount()`** — Returns absolute span for inverted ranges.
- **SARIF metadata round-trip** — Preserves non-string values via `fmt.Sprintf("%v", v)`.
- **SARIF `FindingsFromSARIF`** — Decomposed from cognitive complexity 90 to thin loop with 4 extracted helpers.
- **`math/rand` → `math/rand/v2`** — In retry jitter.
- **`detectResult` ghost type eliminated** — Pipeline uses `PartialResult` directly.
- **`findingKey` extracted** — Now `Finding.Key()` method instead of duplicate helpers.

### Fixed

- **`OnFix` callback accuracy** — Fires only for actually-applied fixes, not skipped ones.
- **`Pipeline.OnFinding` data race** — Added `callbackMu sync.Mutex` for concurrent callback safety.
- **`Report` copylocks on JSON marshal** — Uses `*sync.Mutex` instead of value `sync.Mutex`.
- **`Correlate()` O(n²) hang** — Hard-limited at `maxCorrelations = 10000`.
- **`FilterInvalid` global mutable state** — Changed from `var` to function.
- **`Position.HasEnd()`** — Checks `End.Line > 0 || End.Offset >= 0` (offset 0 is valid).
- **`Adjacent()` fallback** — No longer falls back to offset-based when line info present.
- **`govet.go parsePosn`** — Uses `strconv.Atoi` with error checking.
- **`FixApplier.replaceNearestToLine`** — Finds closest occurrence instead of first match.
- **Backup path collisions** — Include nanosecond timestamp suffix.
- **`Report.FindByRule`** — Uses `ActiveFindings()` for consistency.
- **Global `log.SetFlags(0)` init** — Removed unexpected global logger side effect.
- **SARIF suggestion-only fixes** — Export as fix descriptions without replacements.
- **SARIF `BeforeCode` and `RelatedRef.FindingID` round-trip** — Now preserved in properties.
- **`hasLineRange` dead code** — Removed unreachable branch.
- **`FixApplier` concurrent backup paths** — Uses unique temp dir with `os.MkdirTemp`.
- **Flaky `TestProperty_IDRoundTrip`** — Seeded with deterministic `rand.NewSource(42)`.

### Testing

- Coverage: **root 99.6%**, **pipeline 98.0%**, **detectors 96.1%**, **CLI 95.4%**
- Fuzz tests for SARIF import (1.6M execs, zero panics) and merge
- Property-based tests for ID generation round-trip
- Comprehensive edge-case tests for FixEngine, FileBackup, FixApplier, RetryConfig, Verifier
- CI stress test (`-count=20`) with govulncheck

## [0.1.3] - 2026-04-19

### Breaking

- **`pipeline.New()` now returns `(*Pipeline, error)`** — Previously returned `*Pipeline` with no error. Now validates config via `Config.Validate()` and rejects invalid configs (negative iterations/timeout, invalid retry settings).

### Added

- **`PipelineResult.PartialErrors`** — `map[string]error` surfacing per-detector errors when `GracefulDegradation` is enabled. Callers can now inspect which detectors failed during partial success.
- **`PipelineResult.Metrics`** — `MetricsSnapshot` populated automatically after `Run()` completes, including `TotalDuration`, `StageDurations`, `DetectorTimes`, and `FindingsFound`.
- **CLI metrics output** — Metrics summary printed to stderr after each run (detect/fix counts, duration).
- **Tests for config validation, partial errors, metrics snapshot** — `TestNew_RejectsInvalidConfig`, `TestNew_ValidConfig_NoError`, `TestPipelineRun_PartialErrorsSurfaced`, `TestPipelineRun_MetricsInResult`.

### Changed

- **Test helpers moved to `_test.go`** — `testutil.go` merged into `testutil_test.go`; `suppression_test_util.go` renamed to `suppression_test_util_test.go`. No test-only code ships in production builds.
- **`FixStrategyAI` documented** as phantom/placeholder — not wired to any implementation.
- **`map[string]bool` → `map[string]struct{}`** in `sarif_test.go` for idiomatic Go sets.
- **Removed dead `detectorSpec.Args` field** from CLI.
- **Removed unused `//nolint:goconst` directive** in `finding_extra_test.go`.
- **Extracted "changed" sentinel to const** in clone test for goconst compliance.

### Fixed

- **Metrics snapshot defer ordering bug** — `TotalDuration` was always zero because snapshot was taken before deferred `SetEnd()`. Moved snapshot into deferred cleanup so it captures correct end time.

### Testing

- Coverage: 93.1% root, 87.1% pipeline, 71.6% detectors, 57.3% CLI
- All tests pass with `-race`, `go vet` clean

## [0.1.2] - 2026-04-19

### Changed

- **SARIF property key constants** — Extracted 10 hardcoded property strings into named constants (`sarifPropID`, etc.)
- **SARIF nil safety** — Fixed nil `Region` dereference panics in `applySarifPosition` and `findingFromSarResult` for malformed SARIF input
- **SortByPosition refactor** — Replaced hand-rolled three-way comparison with `Position.Compare`
- **map[string]bool → map[string]struct{}** — Idiomatic Go set in `pipeline.collectAllFindings`
- **Retry MaxDelay validation** — `MaxDelay == 0` with `BaseDelay > 0` now returns an error instead of busy-looping
- **Merge nil-safety** — Passing nil `*Report` in the reports slice no longer panics

### Fixed

- Hand-rolled `HasPrefix` check replaced with `strings.HasPrefix` in SARIF metadata parsing
- SARIF filtered results capacity hint for reduced allocations
- Missing test coverage for `NewFinding`, `SuppressionKind.IsValid`, `Severity.GTE/LTE`, `Finding.String()`, `Report.AddFindings`, `Merge` with nil reports
- Missing `staticcheckCategory` F-prefix test case

## [0.1.1] - 2026-04-19

### Changed

- **SARIF decomposition** — `FindingsFromSARIF` refactored from cognitive complexity 90 to thin loop with 4 extracted helpers
- **Sentinel error migration** — All validation errors use `errors.New` sentinels + `fmt.Errorf("%w")` wrapping (err113 compliance)
- **golangci-lint zero issues** — Full lint compliance across 80+ enabled linters
- **CI upgraded** — Multi-OS matrix (ubuntu + macos), dedicated coverage job with 75% enforcement, tag-triggered builds
- **Code formatting** — Applied golines across entire codebase

- **CLI integration tests** — Coverage from 24% to 59% (loadConfig, validate, profiling, output)
- **SARIF parse benchmark** — `BenchmarkFromSARIF` for the `FindingsFromSARIF` hot path
- **CONTRIBUTING.md** — Fixed Go version (1.26), added golangci-lint commands, expanded pre-submit checklist

### Fixed

- Missing `FixStrategyNone` case in `HasFix()` switch (exhaustive linter, potential silent bug)
- Indentation bug in `findingToSARIF` (line 221)
- Error wrapping in CLI `setupProfiling` (wrapcheck compliance)
- Unused test helper extraction and table-driven test modernization

## [0.1.0] - 2026-04-11

### Added

- **Core types** — `Finding`, `Severity`, `FixStrategy`, `Position`, `Range`, `Category`, `Suppression`, `Report`
  - `NewFinding()` constructor with auto-generated ID
  - `IsValid()`, `HasFix()`, `HasSuggestion()`, `IsSuppressed()`, `Clone()`
  - `Position`/`Range` with `Overlaps()`, `Intersection()`, `Adjacent()` geometric operations
- **Filtering** — `Filter`, `BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFile`, `ByRule`, `ByTool`, `ByFixStrategy`
- **Grouping** — `GroupBy`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`
- **Merging** — `Merge` with deduplication by ID, position, or rule; `Correlate` for cross-tool correlation
- **SARIF 2.1.0** — Full round-trip: `FindingsToSARIF` + `FindingsFromSARIF` with `ToSARIFFiltered`
- **LSP Diagnostic conversion** — `FromLSP`, `ToLSP`, `FromLSPRelated`
- **go/analysis integration** — `FromDiagnostic`, `AnalysisDiagnostic`
- **JSON serialization** — `FromJSON`, `ReportFromJSON`, `PrettyJSON`, `LineJSON` with dropped-finding counts
- **Structured errors** — `FindingError` with categories (Validation, Conflict, Apply, Verify, Pipeline)
- **ID generation** — Stable `tool:rule:file:line:col` format with FNV-1a 128-bit hash
- **Pipeline package** (`pipeline/`) — detect → triage → fix → verify loop
  - Parallel and sequential detector execution via `errgroup`
  - Fix conflict detection and resolution
  - AST-aware fix application with text fallback
  - Post-fix verification by re-running detectors
  - Metrics collection with timing, counts, and snapshots
  - Exponential backoff retry wrapper for flaky detectors
  - Partial success — continue with findings from successful detectors
  - `Config.Validate()` + `RetryConfig.Validate()` with sentinel errors
- **CLI tool** (`cmd/go-finding`) — JSON/YAML/SARIF/text output, pprof profiling, version via ldflags
- **Built-in detectors** (`internal/detectors/`) — govet + staticcheck JSON wrappers
- **CI/CD** — GitHub Actions with multi-OS matrix, coverage enforcement

### Testing

- 81.1% total coverage (root: 91.7%, pipeline: 84.0%, detectors: 71.6%, CLI: 24.1%)
- Fuzz tests for ID generation/parsing, merge, filter, dedup, correlate
- Property-based tests using `testing/quick` for filter, group, merge, ID round-trip
- Integration tests for backup/restore, graceful degradation, retry, verify

### Documentation

- `README.md` — project overview and quick start
- `CONTRIBUTING.md` — contribution guidelines and development setup
- `AGENTS.md` — AI assistant context for development
- `docs/USAGE_GUIDE.md` — comprehensive usage guide
- `cmd/go-finding/config.example.yaml` — sample configuration

### Dependencies

- `gopkg.in/yaml.v3` — YAML config support in CLI
- `golang.org/x/sync` — errgroup for parallel detection
- `golang.org/x/tools` — go/analysis framework integration

---

[Unreleased]: https://github.com/larsartmann/go-finding/compare/v1.14.0...HEAD
[1.14.0]: https://github.com/larsartmann/go-finding/compare/v1.13.0...v1.14.0
[1.13.0]: https://github.com/larsartmann/go-finding/compare/v1.12.0...v1.13.0
[1.12.0]: https://github.com/larsartmann/go-finding/compare/v1.11.0...v1.12.0
[1.11.0]: https://github.com/larsartmann/go-finding/compare/v1.10.0...v1.11.0
[1.10.0]: https://github.com/larsartmann/go-finding/compare/v1.9.2...v1.10.0
[1.9.2]: https://github.com/larsartmann/go-finding/compare/v1.9.1...v1.9.2
[1.9.1]: https://github.com/larsartmann/go-finding/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/larsartmann/go-finding/compare/v1.8.0...v1.9.0
[1.8.0]: https://github.com/larsartmann/go-finding/compare/v1.7.0...v1.8.0
[1.7.0]: https://github.com/larsartmann/go-finding/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/larsartmann/go-finding/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/larsartmann/go-finding/compare/v1.4.1...v1.5.0
[1.4.1]: https://github.com/larsartmann/go-finding/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/larsartmann/go-finding/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/larsartmann/go-finding/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/larsartmann/go-finding/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/larsartmann/go-finding/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/larsartmann/go-finding/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/larsartmann/go-finding/compare/v0.9.1...v1.0.0
[0.9.1]: https://github.com/larsartmann/go-finding/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/larsartmann/go-finding/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/larsartmann/go-finding/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/larsartmann/go-finding/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/larsartmann/go-finding/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/larsartmann/go-finding/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/larsartmann/go-finding/compare/v0.4.3...v0.5.0
[0.4.3]: https://github.com/larsartmann/go-finding/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/larsartmann/go-finding/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/larsartmann/go-finding/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/larsartmann/go-finding/compare/v0.2.1...v0.4.0
[0.2.1]: https://github.com/larsartmann/go-finding/compare/v0.1.3...v0.2.1
[0.1.0]: https://github.com/larsartmann/go-finding/releases/tag/v0.1.0
