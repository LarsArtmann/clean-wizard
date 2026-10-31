# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-10-11

First release on the Go 1.27 toolchain. Consumers should read the toolchain
entry under Changed before bumping: the minimum Go version rises from 1.26 to
1.27. No exported Go API of this SDK changed; every consumer-observed type
still comes from `go-finding` (now v1.14.0). The 0.x MINOR (not PATCH) signals
the raised toolchain floor. All six known consumers already run Go 1.27, so
the practical breakage surface is empty (verified 2026-10-11,
`docs/ADOPTION.md`).

### Added

- **`docs/migration-guide.md`:** full consumer walkthrough (add dependency →
  port one rule → wire execution/exit codes → what to delete), with the
  measured pilot-port LOC table and common pitfalls (workspace tidy
  contamination, ID preservation). Linked from the README Migration path
  section.
- **Registry patterns section in README:** init-time, plugin/opt-in
  (`FilterRules`), and dynamic `Deregister` patterns, with the Register
  panic-contract reasoning documented in
  `docs/decisions/2026-09-09_register-panic-vs-error.md` (ADR: keep panics;
  decision recorded, no API change).
- **`bench_test.go`:** benchmarks for `Registry.Get`/`Has`/`Deregister` at
  100/500/1000 rules; baseline recorded in `reports/benchmarks-baseline.txt`.
- **CI:** `actionlint` job (workflow YAML linting, v1.7.12) and a
  `stranger-test` job (clean directory, proxy-only, `go get @latest`, build
  and run a hello-linter) — simulates exactly what pkg.go.dev users get.
- **`.github/dependabot.yml`:** weekly gomod + github-actions update checks.
- Nil-element coverage for `RuleErrors`: `errors.Join` output containing nil
  children is handled and now explicitly tested (race-clean).
- gitleaks full-history scan artifacts: 109 commits scanned, 0 findings;
  sensitive-docs sweep of `docs/` accepted with a dated verdict addendum
  (no credential values in history).
- **`examples/naked-return-guard`: pilot port of `branching-flow`'s
  `nakedreturn` rule.** Validates the converter-deletion claim for the
  analyzer family: the original ships analyzer.go (106 LOC) + a Detection
  type (17 LOC) + a `DetectionsToFindings` converter (33 LOC) on top of
  shared astutil infrastructure; the port is one self-contained `RuleFunc`
  (195 LOC including docs and main) emitting `finding.Finding` directly.
  Detection logic is a faithful port (naked `return` in functions >10 lines
  with named results; test files and vendor trees skipped); the original
  rule ID `NAKED_RETURN_GUARD`, severity, confidence (0.85), suggestion, and
  metadata keys are preserved for suppression/output compatibility. Covered
  by three integration tests (clean code, long-function detection, test-file
  skip).
- **`examples/oops-redundant-guard`: pilot port of `erraudit`'s
  `oops_redundant_guard` rule.** Validates the converter-deletion claim for
  the error-handling domain: the original ships the `oopspattern` detection
  core (206 LOC), a `go/analysis` autofixer (102 LOC), the violation-detector
  path in `internal/ast`, and a Violation-to-finding bridge; the port is one
  self-contained `RuleFunc` (about 350 LOC including docs and main) emitting
  `finding.Finding` directly. Detection logic is a faithful port (redundant
  `if err != nil { return oops.Wrap(err) }` + `return nil` guards); the rule
  ID `oops_redundant_guard`, severity (warning), confidence (0.85), message,
  and suggestion text are preserved. Covered by three integration tests
  (clean code, detection, if/else not flagged).
- **`SECURITY.md` completed** with "Supported Versions" and "Security
  Practices" sections plus a response-time commitment, clearing the
  `securitymd` BuildFlow gate.
- **Doc-drift guard** (`scripts/check-doc-versions.sh` + CI `doc-drift` job +
  flake `doc-versions` check): fails when a living doc hardcodes a dependency
  version that is neither `go.mod`'s nor a repo tag, or when the Go pins in
  `flake.nix` / `ci.yml` / `.golangci.yml` fall out of lockstep with `go.mod`.
  Living docs now point at `go.mod` instead of duplicating versions — the
  mechanism behind the 2026-10-11 six-doc drift incident.
- **Adoption proof and doc** (`docs/ADOPTION.md`): converter deletion verified
  at source level in all six consumers (4 MIGRATED, 2 PARTIAL with documented
  deliberate adapters; one deleted a ~4,737-LOC violation hierarchy). Includes
  the refresh procedure and the ecosystem version-distribution snapshot.
- **Flake `checks`** derivations for `go test`, `go vet`, `golangci-lint`, and
  the doc guard (previously only treefmt) — `nix flake check` now runs the
  same gates CI does, hermetically vendored.
- **Adoption smoke test**: the CI stranger-test job now exercises the
  consumer-facing helpers real linters use (`Registry.Run`,
  `ExitCodeByConfidence`, `DetectorFromRegistry`), not just a compile.
- **`scripts/bump-checklist.sh`**: encodes the dependency-bump checklist
  (build/vet/test/race/lint/tidy-drift/doc-guard) that kept getting skipped.
- **CI benchmark job**: records `BenchmarkRegistry_*` results as a workflow
  artifact each push (comparison vs `reports/benchmarks-baseline.txt` stays
  manual — shared runners are too noisy to gate on).

### Changed

- **Dependency:** `go-finding` bumped `v1.10.0` → `v1.14.0`, release by
  release (this SDK passes every change through unchanged; gate green per
  bump: build, test, race, vet, golangci-lint):
  - **v1.11.0 (2026-09-17):** release-workflow hardening (concurrency guard,
    preflight gates, docs version stamp) and a fix for an integer underflow in
    correlation proximity sorting — the only core-module behavior fix in the
    window.
  - **v1.12.0 (2026-09-17):** `pipeline` gains `PipelineResult.Outcomes`;
    remaining changes are test-only or module-hygiene (no core-module change
    for this SDK).
  - **v1.13.0 (2026-09-22):** `toolsdk`-only (`Trigger.NotRequires`) plus a
    clearer tag-validation error message — no core-module change for this SDK.
  - **v1.14.0 (2026-10-05):** report coverage evidence (`Summary.FilesScanned`
    always serialized, `Summary.SkippedModules`) — this SDK passes both
    through — and the json/v2 GA build change that let this repo drop
    `GOEXPERIMENT=jsonv2` everywhere. The new toolsdk options channel and
    switch topology live in sub-modules this SDK does not import.
- **Toolchain:** `go.mod` `go 1.26` → `go 1.27`; the flake pins `go_1_27` and
  every CI go step pins `GOTOOLCHAIN: go1.27.0`. `goimports` in treefmt is
  pinned to the same toolchain so the network-less `nix flake check` sandbox
  never tries to auto-download Go.
- **CI lint toolchain:** golangci-lint action bumped `v2.13.2` → `v2.14.0` —
  the first release built with Go 1.27; older binaries cannot decode Go 1.27
  stdlib export data (`export data version 5 is greater than maximum
  supported version 4`), which is what turned CI red on 2026-10-11. The stale
  `goexperiment.*` build tags (`arenas`, `goroutineleakprofile`,
  `runtimesecret`, `simd`) were removed from `.golangci.yml` at the same time,
  so lint sees exactly what `go build` sees.
- **`GOEXPERIMENT=jsonv2` dropped everywhere.** Go 1.27 promotes
  `encoding/json/v2` and `jsontext` to standard packages, so the experiment flag
  in `flake.nix` (both devShells and all eight apps), `.buildflow.yml`,
  `.golangci.yml` build tags, `ci.yml`, and the docs is now obsolete and was
  removed. Verified by a forced clean `GOEXPERIMENT= go build -a ./...`.
- **Dependency:** `go-finding` bumped from `v1.9.2` to `v1.10.0` (2026-09-11).
  v1.10.0 contains no core-module code changes — its only addition is the new
  `toolsdk` sub-module (BuildFlow provider plugin contract), which this SDK
  does not import. Full gate green post-bump (build, test, race, vet,
  golangci-lint, `nix flake check`). No API changes required in this SDK.
- **CI golangci-lint bumped `v2.12.2` → `v2.13.2`** to match the nixpkgs
  devShell version (one source of truth), and `exhaustruct` migrated to its
  replacement `exhaustruct_v5` (deprecated since v2.13.0); three nolint
  directives updated accordingly.
- `RuleFunc.Check` godoc now references `errors.AsType` (Go 1.26 generic),
  matching the errors-family convention.
- **Dependency:** `go-finding` bumped from `v1.7.0` to `v1.9.2` (transitives:
  `golang.org/x/mod` v0.41.0, `x/sync` v0.23.0, `x/sys` v0.48.0). Full gate
  green post-bump (build, test, race, vet, golangci-lint, `nix flake check`).
  No API changes required in this SDK.
- `dprint.json` re-added as a standalone formatter config (not wired into
  treefmt, preserving hermetic `nix flake check`).

### Removed

- The unused `PRIVATE_REPO_TOKEN` GitHub Actions secret (post-`v0.3.1` repo
  operation; `ci.yml` has had zero `secrets.` references since `7675912`).
  Rotate the underlying PAT if it served no other repo.

### Fixed

- **flake.nix `outputs` re-declares `self`** in the `inputs@{ ... }`
  destructure (regressed during a later flake rewrite; Nix 2.34.8 strict
  `@`-pattern checking made every `nix run .#*` / `nix flake check` fail with
  "function 'outputs' called with unexpected argument 'self'"). Restored to
  match the documented gotcha and the `go-finding` reference flake.
- **CI was red on Go 1.27**: the pinned golangci-lint binary predated the
  toolchain and typecheck failed on every file (see Changed). Fixed by the
  v2.14.0 bump; all eight CI jobs green on the release commit.

### Documentation

- **Adoption verified and documented** (`docs/ADOPTION.md`): per-repo verdicts
  on converter deletion, refresh procedure, metrics. The README status claim
  and AGENTS.md consumer table updated from "unproven" to verified.
- ROADMAP: Theme 2 rewritten on the verified-claim footing; Q2 (root package
  stays) and Q3 (library-only stands) resolved with rationale; stale raw ideas
  (stranger-test CI job, flake checks) removed — both shipped.
- `docs/DOMAIN_LANGUAGE.md` borrowed-vocabulary section is now version-neutral
  (points at `go.mod`) instead of pinning a go-finding version in prose.
- Twenty superseded status reports archived under `docs/status/archived/` with
  superseded annotations; the two 2026-10-11 reports annotated with the
  same-day execution results (GOEXPERIMENT pivot, adoption verification).

## [0.3.1] - 2026-09-09

Docs-and-policy release: zero Go API changes since `v0.3.0`. Exists so
pkg.go.dev stops rendering v0.3.0's frozen README, which still carries the
pre-public-flip "private dependency" install instructions.

### Added

- Support posture (open, best-effort, no SLAs): `SUPPORT.md`, `SECURITY.md`
  (private advisories for vulnerabilities), bug/feature issue templates,
  and a README "Support & security" section. Resolves ROADMAP Q6.
- `docs/DOMAIN_LANGUAGE.md`: borrowed-vocabulary section defining the
  `go-finding` v1.7.0 output types the SDK passes through (`Finding`,
  `Report`, `Severity`, `Confidence`, `FixStrategy`, `GroupID`, `Detector`),
  with the boundary note that fix outcomes and rollback belong to the
  `go-finding/pipeline` FixEngine — a module this SDK deliberately does not
  import.

### Changed

- Repository is now public. With `go-finding` and `go-error-family` public as
  well, plain `go get` needs no authentication; the README's private-dependency
  warning and CI's `GOPRIVATE`/token scaffolding were removed.

### Documentation

- Backfilled the missing `[0.3.0]` section (the tag existed since 2026-09-08
  with no changelog entry).
- Full docs-health pass (2026-09-09): every remaining unannotated item in the
  `docs/status/` and `docs/planning/` reports resolved inline;
  `2026-07-27_14-38` archived; `TODO_LIST.md` rebuilt (obsolete
  `PRIVATE_REPO_TOKEN`-creation item removed, publication work harvested);
  `ROADMAP.md` Theme 3 rewritten for the public repo; `FEATURES.md` CI/README
  rows corrected and re-verified.
- Removed the dead Go Report Card badge from the README — the service has
  been sunset.
- `CONTRIBUTING.md` de-drifted: version references now point at `go.mod`
  instead of duplicating them, and the duplicated Reporting Issues / Examples
  sections were merged back into one each.
- `AGENTS.md`: released-versions gotcha updated to v0.3.0, `dprint.json`
  orphant-config note added, proxy-only resolution verified.
- Docs-health follow-up (2026-09-09): lead-phrase strikethroughs in
  `2026-07-30_16-19` normalized to full-line (matching the file's canonical
  style); pre-existing `✅` table markers kept as-is (consistent within each
  file, each carries its own evidence).
- `examples/` doc comments audited against the code — verified current, no
  changes needed.

### Removed

- `dprint.json` orphan config. It was template residue (dockerfile plugin and
  helm/charts excludes for files this repo does not have) and was never wired
  into `flake.nix`, CI, or BuildFlow. Wiring it hermetically is possible only
  by vendoring its wasm plugins — disproportionate for a repo whose code
  formatting is already enforced by treefmt (gofumpt/goimports/golines/nixfmt).
  Markdown remains hand-formatted under the docs-health process.

## [0.3.0] - 2026-09-08

### Changed

- **Dependency:** `go-finding` bumped from `v1.6.0` to `v1.7.0` — consumes
  GroupID finding groups, per-finding fix outcomes, and the per-file rollback
  default. The SDK uses the core finding package only; no call sites were
  affected.
- Documentation synced: stale `v1.4.1` references updated and the v0.2.0
  release lessons (CI token secret, `GOTOOLCHAIN` override, library-only
  coverage gate) recorded in `AGENTS.md`.

> Process note: the `v0.3.0` tag was pushed before its CI run finished; that
> run failed in 7s (logs expired) and master was green again by `487d254`.
> Future tags go out only after the exact commit's CI is green.

## [0.2.0] - 2026-09-02

All changes are additive — zero removals, zero signature breaks.

### Added

#### Core types

- `ErrMissingFields` exported sentinel — `errors.Is(err, ErrMissingFields)` lets
  consumers distinguish validation failures from runtime failures.
- `RuleMeta.Validate() error` — checks ID, Name, Description, and Cat are
  non-empty. Available for use during rule construction, before the registry.
- `RuleFunc.NewFinding(message, pos) *finding.Builder` — pre-stamped builder
  from rule metadata (rule ID, tool name, severity, category).
- `RuleFunc.ToolName` field — optional tool name stamped onto every finding
  the rule emits.

#### Registry API

- `Registry.Get(id string) (Rule, bool)` — lookup by stable ID.
- `Registry.Has(id string) bool` — companion to `Get`.
- `Registry.Deregister(id string) bool` — runtime rule removal (plugin
  scenarios); returns true if found. Snapshot semantics documented: a rule in
  the `All()` snapshot still executes even if deregistered mid-run.
- `RunOption` functional-option type + `ContinueOnError()` option for
  `Registry.Run`. Default is fail-fast (unchanged); `ContinueOnError()` runs
  all rules regardless of individual failures, collects partial findings, and
  joins errors via `errors.Join`.
- `WithToolName(name) RegistryOption` — stamps the tool name onto all findings
  and the report header at registration time. `NewRegistry` now accepts
  options (`NewRegistry(opts ...RegistryOption)`); existing zero-arg calls
  keep compiling.
- `FilterRules(all, enable, disable) []RuleFunc` — standard --enable/--disable
  filtering for CLI and plugin entry points.
- `ExitCodeByConfidence(report, threshold) int` — tiered exit code: 0 clean, 1
  at-or-above threshold, 2 below threshold (triage mode).
- `RuleErrors(err) []*RuleError` helper — extracts all `*RuleError` values
  from joined errors (ContinueOnError mode), tree-walking `Unwrap() []error`
  and `Unwrap() error` chains.

#### Testing & examples

- `examples/minimal-linter/` — a minimal consumer linter proving the full
  Rule -> `finding.Finding` -> `Registry.Run` -> `ExitCodeFromReport` path.
  Integration tests (`TestExampleMinimalLinter_CleanDir`,
  `TestExampleMinimalLinter_MissingReadme`) verify exit codes via
  `exec.Command`.
- `examples/no-go-mod/` — pilot port of `go-structure-linter`'s `NoGoModRule`,
  rewritten with `go-linter-sdk`. Validates the converter-deletion claim: a
  rule emits `finding.Finding` directly via `finding.NewBuilder(...)`, with no
  intermediate Violation/Issue type. Integration tests verify exit codes.
- `TestRegistry_ConcurrentReadWrite` stress test exercising the registry's
  `RWMutex` under `-race` (writers + readers + runners).
- `BenchmarkRegistry_Register`, `BenchmarkRegistry_All`, and
  `BenchmarkRegistry_Run` — the first performance baseline for the registry
  hot paths, modernized to the `b.Loop()` pattern (Go 1.24+).
- Testable examples (`ExampleRegistry_Run`, `ExampleDetectorsFromRegistry`,
  `ExampleOptIn`, `ExampleRegistry_Run_continueOnError`) — visible on pkg.go
  docs, verified by `go test`.
- `FuzzNewRuleError` fuzz test verifying `Error()` never panics with nil cause
  or arbitrary rule IDs.
- `TestRuleError_Is_Canceled` and `TestRuleError_Is_DeadlineExceeded` — verify
  context errors chain correctly through `RuleError.Unwrap()`.
- `TestDeregister_DuringRun_SnapshotSemantics` and `TestRuleErrors_DeeplyNested`
  — snapshot-semantics and deep-nesting traversal coverage.

#### Infrastructure

- CI `go mod tidy` check: `git diff --exit-code go.mod go.sum` after tidy to
  catch module drift.
- CI coverage gate: the test step generates `coverage.out` and fails if total
  coverage drops below 90%.
- CI workflow uses the `PRIVATE_REPO_TOKEN` secret (PAT with read access to
  `go-finding`) instead of the default `GITHUB_TOKEN` (scoped to this repo
  only, cannot fetch the private `go-finding` dependency).

### Changed

- **Dependency:** `go-finding` bumped from `v1.4.1` to `v1.6.0`.
- **Toolchain:** Go 1.26.5 -> 1.26.7.
- `Registry.Register` now panics on any empty identity field (ID, Name,
  Description, Category), not just empty ID. Callers with complete `RuleMeta`
  are unaffected.
- `Registry.Run` accepts variadic `RunOption`s (source-compatible: existing
  `Run(ctx, dir)` calls keep compiling).
- `validateRuleIdentity` and `RuleMeta.Validate` share a single
  `validateIdentityFields` function, eliminating duplication between the
  interface-level and struct-level validation paths.
- `.golangci.yml` right-sized from first principles: removed cargo-culted
  `mnd` numbers/functions and `gosec` excludes (zero findings without them —
  this is a pure library), trimmed `varnamelen` ignore-names from 30+ to the
  6 conventional abbreviations that actually appear, added `examples/` path
  exclusions for `depguard`/`forbidigo`/`mnd` (example CLI code legitimately
  prints and uses exit codes), and allow-listed `linter.Rule` for `ireturn`
  (`OptIn` intentionally returns the interface).
- `DetectorFromRegistry(r *Registry, ...)` parameter renamed to `registry`
  for clearer public-API documentation.

### Documentation

- README expanded with a Status callout, Quick Start, data-flow and
  execution-paths diagrams (mermaid + ASCII fallbacks), runnable examples,
  and the completed API table (`WithToolName`, `ExitCodeByConfidence`,
  `FilterRules`, `RuleFunc.NewFinding`). CI badge added.
- Documentation one-home rule established: `FEATURES.md` tracks only
  capabilities that have code today; all not-yet-built capabilities live
  exclusively in `ROADMAP.md`. This kills the FEATURES<->ROADMAP PLANNED
  split-brain.
- `DOMAIN_LANGUAGE.md` entries added for `ErrMissingFields`, `RuleErrors`,
  and the `DetectorFromRegistry`/`DetectorsFromRegistry` integration paths.
- Godoc cross-reference from `ContinueOnError` to the `RuleErrors` helper,
  guiding callers to the right tool for enumerating joined errors.
- `FEATURES.md` refreshed (verification block, new feature rows for the new
  API); `ROADMAP.md` reconciled with shipped code and closed open questions
  Q1/Q4/Q5; `TODO_LIST.md` rebuilt from status reports; `CONTRIBUTING.md`
  duplicate section merged.
- All source line-number references in `DOMAIN_LANGUAGE.md` and `FEATURES.md`
  converted to durable symbol-name references (line numbers rot on every
  edit; symbol names don't).
- Status reports annotated with inline resolution markers (119 action items
  across 5 July reports); `cqrs-lint` feedback moved to
  `docs/feedback/processed/` with a maintainer-response pointer.

## [0.1.0] - 2026-07-30

Initial public release.

### Added

#### Core types

- Shared linter scaffolding: `Rule` interface, `Registry`, `RuleFunc` adapter,
  `RuleMeta` declarative identity, `Category` open string type (8 recommended
  values). Rules emit `finding.Finding` directly — no converter layer.
- `RuleError` type (`errors.go`) wrapping rule failures with the offending
  rule's stable ID, plus `ErrRuleFailed` sentinel and `NewRuleError`
  constructor. Callers recover the failing rule via
  `errors.AsType[*RuleError]` (Go 1.26+) or check
  `errors.Is(err, ErrRuleFailed)`.
- `DetectorFromRegistry(registry, toolName) finding.Detector` — adapts a
  registry to a single `finding.Detector` for BuildFlow DAG integration.
  Reads working dir via `finding.WorkingDirFromContext`.
- `DetectorsFromRegistry(registry) []finding.Detector` — returns one
  `finding.Detector` per registered rule for `go-finding/pipeline` integration
  with per-rule parallelism, timeouts, error isolation, and metrics. Each
  detector is named after the rule's ID.
- `ExitCodeFromReport(report) int` — binary exit code: 0 if clean, 1 if any
  findings. The ecosystem convention.
- `IsEnabledByDefault() bool` on the `Rule` interface. `RuleFunc` returns
  `true` by default; use `OptIn(rf)` to create a disabled-by-default rule
  that only runs when a consumer explicitly enables it.
- `OptIn(rf RuleFunc) Rule` constructor for opt-in rules (noisy,
  experimental, or domain-specific).

#### Infrastructure

- Nix flake (`flake.nix`) mirroring the `go-finding` toolchain: pins Go 1.26,
  sets `GOEXPERIMENT=jsonv2`, exposes `test`, `test-race`, `bench`, `build`,
  `vet`, `lint`, `coverage`, and `clean` apps, and wires treefmt
  (gofumpt, goimports, golines@120, nixfmt).
- BuildFlow configuration (`.buildflow.yml`) and a buildflow-managed
  `.gitignore` block.
- golangci-lint v2 configuration (`.golangci.yml`) with the
  `goexperiment.jsonv2` build tag, matching the ecosystem linter set.
- GitHub Actions CI workflow (`.github/workflows/ci.yml`): test
  (ubuntu-latest + macos-latest), lint, format check, govulncheck, and nix
  flake check. Configures VCS auth for private modules (`GOPRIVATE`);
  all actions pinned to commit SHAs.
- `.github/CODEOWNERS` file for automated review routing.
- `go.work` workspace for local cross-repo development (gitignored, not
  committed). References sibling `../go-finding` checkout.
- Project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `AGENTS.md`,
  `.gitattributes`, durable `reports/.gitkeep`.
- `.editorconfig` enforcing UTF-8/LF, tabs for Go/Makefile, 2-space for
  YAML/JSON/Nix/TOML.
- Living project docs: `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`,
  `CHANGELOG.md` — honest feature inventory, short-term work backlog,
  long-term vision, and change log.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous-language glossary defining the six
  core concepts (`Rule`, `RuleFunc`, `RuleMeta`, `Category`, `Registry`,
  `RuleError`), their relationships, and what is deliberately NOT in the
  domain.

### Changed

#### Breaking changes (during 0.x development, relative to pre-release internal state)

- **Breaking:** `Rule` interface now requires `ID() string`. Every rule must
  declare a stable ID that never changes once published — used for registry
  deduplication, suppression matching, filter config, and the
  `finding.RuleName` field on emitted findings. `Name()` remains as the
  mutable display name. Custom `Rule` implementations must add the method;
  `RuleFunc` users must set `RuleMeta.ID`.
- **Breaking:** `RuleMeta` now has a required `ID` field (first field).
  `Register` panics on empty IDs.
- **Breaking:** `RuleError.RuleName` renamed to `RuleError.RuleID`. Callers
  accessing the field directly must update; `errors.AsType` / `errors.Is`
  users are unaffected.
- **Breaking:** `Registry.Register` now deduplicates on `ID()` instead of
  `Name()`, and panics on empty IDs. Two rules may share a display `Name()`
  as long as their `ID()` values differ.
- **Breaking:** `NewRuleError` parameter renamed from `ruleName` to `ruleID`.

#### Non-breaking changes

- Pinned `go-finding` dependency to `v1.4.1` (published tag). The local
  `replace ../go-finding` directive has been removed — consumers and CI both
  fetch `v1.4.1` directly via VCS auth.
- `RuleFunc.Check` and `Registry.Run` / `DetectorFromRegistry` /
  `DetectorsFromRegistry` wrap rule execution errors into `*RuleError`, with
  a guard against double-wrapping custom `Rule` implementations.
- `errors.AsType` migration completed: `registry.go`, `registry_test.go`, and
  the `errors.go` doc example all use Go 1.26's generic
  `errors.AsType[*RuleError]`. The project is gopls-clean (zero
  `errorsastype` hints).
- Tests are black-box (`package linter_test`), satisfying `testpackage`
  natively with no suppression. Test factories (`makeRule`, `failingRule`)
  return the concrete `RuleFunc` rather than the `Rule` interface, so
  `ireturn` has nothing to flag either.
- `Category` documented as an open `string` type. The 8 built-in constants
  are recommendations; consumers can define custom categories
  (e.g., `Category("api")`).
- README "Building findings with Confidence and FixStrategy" section showing
  the `finding.NewBuilder(...)` fluent API and the Confidence/FixStrategy
  value tables.
- README "Two execution paths" section with a mermaid diagram comparing
  `Registry.Run` (sequential, fail-fast) vs `DetectorsFromRegistry`
  (parallel, per-rule isolation).
- README private-dependency note warning that `go-finding` requires
  `GOPRIVATE=github.com/larsartmann/*` + auth, with CI vs. local dev
  guidance.

### Fixed

- Build and lint failures caused by `encoding/json/v2` and
  `encoding/json/jsontext` (used transitively via `go-finding`) being excluded
  by build constraints without `GOEXPERIMENT=jsonv2`. The flake devShells and
  apps now export it durably.
- `erraudit` findings: bare `return nil, err` paths in `registry.go`
  and `rule.go` replaced with typed, rule-attributed errors.
- BuildFlow failure `function 'outputs' called with unexpected argument
  'self'`: added `self` to the flake `outputs` destructure pattern. Nix
  2.34.8 enforces strict argument checking on `@`-patterns; the sibling
  `go-finding` flake already declared it.
- Coverage output path mismatch: the `coverage` and `clean` flake apps now
  write/read `reports/coverage.out`, aligning with `AGENTS.md`, the
  `.gitignore` convention, and BuildFlow's `test-coverage` step.
- README "5-line linter" example did not compile: the import block was
  missing `"os"` (needed for `os.Exit`). Added the import; example verified
  to compile against the actual exported API.
