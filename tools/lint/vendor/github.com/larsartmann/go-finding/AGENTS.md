# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. Seven tools detect issues; zero route them to remediation. This library solves that with:

1. **Unified Finding type** — Common representation for all tools
2. **Pipeline** — Automated detect → triage → fix → verify loop
3. **SARIF output** — Standard interchange format
4. **LSP integration** — IDE support

## Module Structure (Multi-Module Go Workspace)

Unix-style decomposition — each module does one thing well, composes via replace directives.

| Module       | Path                                               | External Deps                | Depends on     |
| ------------ | -------------------------------------------------- | ---------------------------- | -------------- |
| **Core**     | `github.com/larsartmann/go-finding`                | go-error-family              | —              |
| **Pipeline** | `github.com/larsartmann/go-finding/pipeline`       | x/sync, gogenfilter          | Core           |
| **Analysis** | `github.com/larsartmann/go-finding/analysis`       | x/tools                      | Core           |
| **Tool SDK** | `github.com/larsartmann/go-finding/toolsdk`        | —                            | Core           |
| **CLI**      | `github.com/larsartmann/go-finding/cmd/go-finding` | yaml, go-output, gogenfilter | Core, Pipeline |

`go.work` coordinates all 5 modules for development. Library sub-modules (`pipeline`, `analysis`, `toolsdk`) carry `replace github.com/larsartmann/go-finding => ../` for `GOWORK=off` builds; **`cmd/go-finding` must have NO replace directives** (enforced by `scripts/replace-audit.sh` — `go install module@version` refuses modules with replaces). Consequence: the CLI is pinned to PUBLISHED sibling-module versions, so it cannot call a new pipeline/analysis API until that module is tagged and the CLI `go.mod` bumped — dedupe CLI-side instead of exporting new sibling APIs for it (learned 2026-09-22/23 dedup pass).

## Key Files

| Area                | Files                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Core types**      | `finding.go`, `finding_methods.go`, `finding_validate.go`, `finding_equal.go`, `position.go`, `range.go`, `text_edit.go`, `report.go`, `filter.go`, `merge.go`, `diff.go`, `format.go`, `json.go`, `id.go`, `errors.go`, `simple_fix.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| **Named types**     | `severity.go`, `confidence.go`, `category.go`, `category_linter.go`, `tag.go`, `fix_strategy.go`, `suppression.go`, `branded_types.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| **SARIF**           | `sarif_types.go`, `sarif_export.go`, `sarif_import.go` (hand-rolled, not go-sarif — see ADR #9)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **LSP**             | `lsp.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| **Extensibility**   | `detector.go`, `detect_result.go` (DetectResult/DetectorV2/RunDetect — the optional coverage-evidence channel; `Report.AddDetectResult` feeds Summary.FilesScanned/SkippedModules), `adapter.go` (ToolAdapter[O]), `registry.go` (DetectorRegistry), `interval_index.go` (IntervalIndex[T])                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| **gotoken**         | `gotoken/gotoken.go` (shared go/token utilities, public package)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **lockutil**        | `lockutil/lockutil.go` (shared sync.Locker helpers — `Locked`, `RLocked` — for generic mutex-guarded critical sections)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| **Pipeline**        | `pipeline/pipeline.go` (Run), `pipeline/pipeline_detect.go`, `pipeline/pipeline_iteration.go`, `pipeline/config.go`, `pipeline/config_file.go`, `pipeline/flight_recorder.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| **Fix engine**      | `pipeline/fix_engine.go`, `pipeline/fix_provider.go`, `pipeline/fix_applier.go`, `pipeline/fix_outcome.go`, `pipeline/fix_edit.go`, `pipeline/conflict.go`, `pipeline/goast/provider.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| **Pipeline extras** | `pipeline/stage_hook.go`, `pipeline/flight_recorder.go`, `pipeline/line_shift.go`, `pipeline/metrics.go`, `pipeline/retry.go`, `pipeline/partial.go`, `pipeline/generated_filter.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **Analysis**        | `analysis/analysis.go` (go/analysis ↔ Finding)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **Tool SDK**        | `toolsdk/spec.go` (Spec contract; `Trigger.Validate` + `ValidLanguages()`/`ErrInvalidTrigger`; `RepairResult.ModifiedFiles`/`Facts []Fact` — non-gating provenance, fix counts stay re-measured), `toolsdk/registry.go` (Register/All; `mustValidate` also panics on invalid Trigger), `toolsdk/triggers.go` (Trigger constructors), `toolsdk/dryrun.go`, `toolsdk/options.go` (Spec.Options declarations + `WithOptions`/`OptionsFromContext` per-run value channel; sentinels `ErrInvalidOption`/`ErrDuplicateOption`/`ErrUnknownOption`/`ErrOptionKindMismatch`; typed accessors `StringFromContext`/`IntFromContext`/`BoolFromContext` — miss → fallback, wrong type → `ErrOptionKindMismatch`; `WithOptions` snapshots the map and nil/empty CLEARS inherited values) — BuildFlow provider plugin contract (sub-module, see Module Dependencies) |
| **Detectors**       | `cmd/go-finding/internal/detectors/govet.go`, `staticcheck.go`, `helpers.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| **CLI**             | `cmd/go-finding/main.go`, `config.go`, `registry.go`, `fix_provider_registry.go`, `generated_filter.go`, `output_adapter.go`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |

## Testing & Build

```bash
nix run .#test                              # Run tests (ROOT module only — ./... does not cross module boundaries, even with go.work active; verified 2026-09-17. Test sub-modules per the GOWORK=off line below)
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector (workspace)
GOWORK=off go test ./...                    # Per-module isolation test (run in each module dir)
golangci-lint run ./...                     # Lint
bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 25  # Benchmark regression check
bash scripts/version-check.sh                                    # Verify version.go matches git tag
bash scripts/release-preflight.sh                                # Pre-tag structural gate (see below)
bash scripts/error-audit.sh                                      # erraudit gate: error-handling violations, all 5 modules (see below)
```

> **json/v2 is GA (verified 2026-09-23).** The project imports `encoding/json/v2`,
> which is stable and default-on since Go 1.27 — no `GOEXPERIMENT` is needed
> anywhere (dropped from flake, workflows, scripts, docs). `GOEXPERIMENT=nojsonv2`
> would now BREAK the build. Direct `go` commands just need a 1.27+ toolchain.

### Release preflight (mandatory before any tag)

`scripts/release-preflight.sh` runs the structural gates as code (version drift,
replace audit, tag collisions, clean tree, unpushed commits, worktree hygiene,
GOWORK=off builds, docs checks incl. the README version stamp). Born from the
v1.7.0 incident where sub-module go.mod files were tagged with a stale core
reference because the checklist lived in the operator's head.
`--bench`/`--stress` flags add the heavy gates. `scripts/release-preflight-selftest.sh`
injects three failure classes (tag collision, version drift, post-tag missing

tag) into a disposable worktree and asserts preflight FAILs on each.

**Preflight is lockstep/core-centric for target tags (observed toolsdk/v1.14.0,
2026-10-01)** — it derives all five target tags from core's `version.go`
(`scripts/release-preflight.sh:123`), so a sub-module-ONLY release (core still
v1.13.0, toolsdk at v1.14.0) fails 5× "tag vX already exists" as false
positives. For a sub-module-only tag, the real collision check is that the
actual target (`toolsdk/v1.14.0`) does not exist; every other gate must still
be green. Also: the script's GOWORK=off build/tidy loops (lines ~153/163) skip
`toolsdk` — run `GOWORK=off go -C toolsdk build ./... && vet` manually.

The Release workflow (`release.yml`) runs under a queueing concurrency group
(`cancel-in-progress: false`), so a manual dispatch can no longer race the
tag-push auto-trigger (v1.9.2 produced two concurrent GoReleaser runs, one 422
duplicate-asset loser).

### erraudit gate (2026-09-11)

`bash scripts/error-audit.sh` runs `erraudit ./... --type-aware` (hierarchical
error-handling analyzer) in all 5 modules; 0 violations is the green state.
Policy: sentinel errors (`errors.New`) + `*FindingError` + stdlib `fmt.Errorf`
wrapping are the documented model (docs/research/erraudit-violation-analysis.md;
samber/oops REJECTED at 3/10 fit). NEVER interpret runs with
`--enforce-samber-oops`/`--enforce-generic-return` as defect lists — those
flags contradict the model; the 79-violation reports they produce are
artifacts. Intentional patterns carry `//nolint:erraudit // <reason>`;
NEVER remove directives based on `erraudit nolint-audit .` — verified
2026-09-23: it false-reports type-aware-suppressing directives as STALE
(it analyzes with go/parser only; stripping its 24 "safe to remove" hits
turned the gate red at every one of those sites). Use the advisory
listing via `scripts/error-audit.sh --nolint-audit`; the only safe
staleness test is remove-one-directive → run the gate → observe red →
restore. `--no-suppress` surfaces those
documented suppressions BY DESIGN (audit mode) — a run with
`--no-suppress --enforce-samber-oops --enforce-generic-return` regenerates
the ~79-violation artifact and proves nothing. Blessed invocation:
`nix run .#error-audit`. CI wiring is blocked
because the erraudit repo is private (no `go install` from public runners
without credentials).

### Dead-gate lesson (2026-09-08: three in one day)

A check that "passes" by not actually running is worse than no check. Found:
bench-check.sh awk pattern never matched benchstat output; pre-commit hook
dead via stale local `core.hooksPath`; version-drift.sh aborted silently under
`set -euo pipefail`. Rules: (1) every check script prints an explicit OK/FAIL
verdict line, (2) verify a new gate's FAIL path once by intentionally breaking
something, (3) never pipe a check script's output through filters that can cut
the verdict line, (4) a doc claim of "verified by X" requires X to have
FINISHED and its output read before the claim is written (2026-09-23: an ADR
carried "allocation profile unchanged, verified by the bench gate" while the
bench was still running; the bench then FAILED with +706%).

## Module Dependencies (per go.mod)

| Module                      | Production Deps              | Test Deps         |
| --------------------------- | ---------------------------- | ----------------- |
| **Core** (`.`)              | go-error-family              | ginkgo/v2, gomega |
| **Pipeline** (`pipeline/`)  | x/sync, gogenfilter          | ginkgo/v2, gomega |
| **Analysis** (`analysis/`)  | x/tools                      | (stdlib testing)  |
| **Tool SDK** (`toolsdk/`)   | — (core only)                | (stdlib testing)  |
| **CLI** (`cmd/go-finding/`) | yaml, go-output, gogenfilter | gomega            |

### Old Dependencies (now isolated to sub-modules)

- `golang.org/x/tools` — go/analysis framework (**analysis module only**)
- `golang.org/x/sync` — errgroup (**pipeline module only**)
- `github.com/go-faster/yaml` — YAML config (**CLI module only**)
- `github.com/onsi/ginkgo/v2` + `gomega` — BDD testing (core + pipeline test deps)
- `github.com/LarsArtmann/gogenfilter/v3` — Auto-generated Go file detection (pipeline + CLI)
- `github.com/larsartmann/go-output` — CLI output formatting (CLI module only)

## Design Principles

1. **Minimal dependencies** — core module keeps a small, deliberate dependency surface
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags/Data fields
4. **One extensibility field** — `Finding.Metadata` is `map[string]string`. NO `Properties map[string]any`
5. **Compatible** — Works with existing Go analysis tools
6. **Resilient** — Retry logic, partial success, nil-safe metrics
7. **Multi-module** — Unix-style decomposition: each module has a single purpose and composes independently. Dual go.work + replace strategy.

## Important Behaviors (Gotchas)

_Updated 2026-09-08 diet pass; pre-diet text archived in `docs/planning/archived/2026-09-08_agents-gotchas-diet.md`._

### Environment & workflow

- **GOEXPERIMENT=jsonv2 DROPPED (2026-09-23)** — `encoding/json/v2` is GA since Go 1.27 (release notes: v2 is the default `encoding/json` backend; opt-out is `nojsonv2`). All env exports removed from flake.nix, workflows, scripts, and docs. Anything still telling you to export jsonv2 is stale. The per-module path still needs `GOWORK=off`.
- **CI/Release `go-version` pins must track the toolchain floor** — when the `go.mod`/`go.work` floor bumps, a stale `go-version:` pin in ci.yml/release.yml fails EVERY job with `go.work requires go >= 1.27` while local gates stay green (the devShell already runs the newer toolchain). This silently killed 6+ master CI runs and both v1.13.0 Release runs (2026-09-20 → 09-23) before anyone looked. Rule: a toolchain-floor bump is a SAME-COMMIT edit of every `go-version:` in both workflows.
- **Run `go test` invocations SEQUENTIALLY on this machine** — two concurrent `go test`/`go build` runs sharing `GOCACHE=/mnt/buildcache/go-build` produce transient `[build failed]` errors. A plain retry succeeds; don't misdiagnose as code/toolchain problems. This includes `go mod edit`: editing go.mod requires while a background stress/test run is active invalidates the module graph mid-run and produces phantom failures like `could not import ... (invalid package name: "")` in go/packages-based tests (v1.12.0 stress run 1, 2026-09-17). Bump requires only on a quiet tree.
- **Formatting has three coordinated signals** — treefmt is not on the devShell PATH; use `nix fmt` (canonical gate, also in the pre-commit hook via `--fail-on-change`), `golangci-lint fmt` (local autofix; `.golangci.yml` formatters mirror treefmt rules: gofumpt, goimports, golines@120), and dprint (markdown/JSON/YAML, pre-commit). Do NOT let the two Go formatters' rules diverge. If the pre-commit hook seems dead, check `git config core.hooksPath` — a stale `.githooks` value once silently disabled it.
- **`nix flake check --all-systems` evaluated and DECLINED (2026-09-08)** — `--all-systems` warns about incompatible systems on this machine (darwin derivations without remote builders configured) and would fail the check. Decision: plain `nix flake check` stays the local + CI gate; revisit `--all-systems` only if darwin builders or remote builder config are added. This closes the f/38 evaluation (previously performed but written nowhere).
- **CI billing superseded (2026-09-08 evening)** — the account-switch gate was bypassed by making the repo PUBLIC: Actions on public repos are free/unlimited. Post-flip `ci.yml` dispatched (run 34274674104) — verify every job; the last private-repo run had ALL 20 jobs fail in 43s from billing (no logs). Until the first green public run, local gates remain the quality bar (race x4, lint x4, structural scripts, go-arch-lint, dprint, `nix flake check`, stress `ginkgo --repeat=20 --race` as MANDATORY release gate).
- **Repo is PUBLIC (2026-09-08)** — visibility flipped via `gh repo edit`; module proxy resolution verified: plain `go get github.com/larsartmann/go-finding@v1.8.0` works for all modules with NO `GOPRIVATE` (5 modules since `toolsdk/v1.10.0`). Public-repo Actions are free, which dissolves the CI billing blocker. Secret scanning + push protection enabled same day.
- **pkg.go.dev is LIVE (2026-09-08 evening)** — https://pkg.go.dev/github.com/larsartmann/go-finding renders (v1.8.0, full docs); the temporary lychee `pkg.go.dev` exclude was REMOVED. Lychee now excludes only still-private sibling repos (remove entries as they go public) and CHANGELOG `v0.1.x` compare links (tags predate the v1.x scheme).
- **Multi-module release tagging** — sub-modules need directory-prefixed tags (`pipeline/v*`, `analysis/v*`, `toolsdk/v*`, `cmd/go-finding/v*`); core uses unprefixed `v*`; sub-modules have no version.go. See `docs/release-procedure.md`. The tagged commit's CI `module-isolation` job is EXPECTED red (v1.11.0 and v1.12.0 both shipped this way): the CLI module's go.sum cannot carry the new version's checksums until the tags are pushed; the resync commit lands on master immediately after. Consumers are unaffected — proxy + `go get` verify independently.
- **v1.14.0 train lessons (2026-10-11)** — (1) The published core `v1.14.0` tag self-reports `finding.Version` = "1.13.0": the version.go bump commit landed AFTER the tag was pushed, and retagging is blocked (`protect-release-tags`). Rule: the bump must be IN the tagged commit — verify `git show <tag>:version.go` before pushing any core tag. (2) `cmd/go-finding/v*` tags now point at the RESYNC commit (v1.14.0 = 13cc091), so `go install cmd/go-finding@vX` builds with the new sibling engine; the historical tags-first-then-resync order shipped CLIs pinned to the PREVIOUS engine (v1.13.0 CLI pinned pipeline v1.12.0). Library tags still go first; only the cmd tag moves after resync. (3) A sub-module tag whose module calls not-yet-tagged core APIs is consumer-broken while looking green locally — the `replace => ../` masks it in every GOWORK=off gate. Proof before tagging: clean-dir consumer module resolving the tag commit via the proxy (`go get <mod>@<sha>` + build). pipeline/v1.14.0 was therefore tagged at pre-DetectorV2 9c675c7, not HEAD. (4) Local `release-preflight.sh --bench` numbers are invalid under heavy machine load (uniform 2-4x time inflation with flat allocs while load avg was 186); the CI benchmark job on a clean runner is the authoritative bench evidence.
- **version-check.sh needs `--match 'v[0-9]*'`** — plain `git describe` picks sub-module tags alphabetically first. Any script resolving the core version from tags must use this flag.
- **Dependabot gomod PRs need a manual fix-up commit (2026-09-27 pattern)** — when a bump raises a module's `go` directive, dependabot does NOT touch `go.work`, and every workspace go command dies with `requires go >= 1.27.1, but go.work lists go 1.27` (takes down test/lint/stress/benchmark/coverage/govulncheck/go-work-sync/arch-check at once), plus the vendorHash goes stale. Fix pattern: worktree the PR branch → bump `go.work`'s `go` line → `go work sync` → `go mod tidy` per module (expect no diffs) → take the vendorHash the CI nix job PRINTS in its error (`To correct the hash mismatch ... use "sha256-..."`; a local `nix build` confirms) → commit `go.work`+`flake.nix` to the PR branch → admin-merge. Dependabot PRs are structurally unmergeable via the UI: branch protection's required contexts are bare job names (`test`) while the workflow reports matrix-expanded names (`test (1.27, ubuntu-latest)`) — merge with `gh pr merge --admin`, which the direct-push protection config (`enforce_admins: false`) permits. The `/` directory config covers ALL nested modules, so a root gomod PR supersedes the sibling-directory PRs (they auto-close on rebase as satisfied).
- **CI scripts guard the 5-module structure** — `replace-audit.sh`, `version-drift.sh`, `test-naming.sh`, `go-work-sync.sh`, `docs-freshness.sh` (backtick spans + links only), `docs-api-check.sh` (documented identifiers must exist in code; README `finding.Version` stamp must exist and match version.go), `json-deterministic-check.sh`, plus `go-arch-lint` (`.go-arch-lint.yml`, 11 components, one-directional flow cli->pipeline->core, analysis->core). All wired into ci.yml.
- **go.mod/go.sum bumps invalidate `flake.nix` `vendorHash`** — run `nix flake check` after any dependency change; a stale hash fails the go-modules fixed-output derivation with a `hash mismatch` (2026-09-10: indirect x/net, x/text, x/tools bumps stale-dated it; complete the bump with `go mod tidy` per module BEFORE updating the hash, or you iterate the bootstrap loop twice).
- **Branch protection is configured for DIRECT pushes (2026-09-23)** — master requires 17 status checks with `enforce_admins: false`: with admin enforcement ON, GitHub rejects every direct push that hasn't pre-run the checks (impossible on a new commit — the first protected push of the day bounced). Admins bypass the merge gates; the required checks still bind future PRs. Tag deletion/force-update on release tags is blocked by the `protect-release-tags` ruleset. Private vulnerability reporting is ENABLED (SECURITY.md intake works).
- **Keep the golangci-lint-action pin in sync with the nix binary** — CI pinned v2.13.1 while nixpkgs shipped v2.13.2; patch drift changed wsl_v5/varnamelen verdicts, so local "0 issues" and CI red disagreed (f/26, 2026-09-23). When buildflow bumps the nix golangci-lint, bump `version:` in ci.yml's lint job to match.
- **`/mnt/buildcache` can hit 100%** — the shared GOCACHE grew to 167G and made every write fail ("no space left on device" from go build). Recovery: `go clean -cache` (and `-fuzzcache` after fuzz campaigns). If builds fail with space errors, check `df /mnt/buildcache` first, not the code.
- **Push release tags ONE PER PUSH (≤3-batch rule superseded 2026-09-23)** — two failure classes: (1) a single push updating >3 tags creates NO workflow events (v1.9.0, silent); (2) even legal batches of 3 raced each other — the Release concurrency group is latest-queued-wins, so near-simultaneous pushes cancelled the older QUEUED runs (v1.12.0: core + analysis lost their runs). Rule: push master, then one tag per push; watch each tag's Release run to a terminal state (`gh run list --workflow=release.yml`) before the next push; re-run cancelled runs STRICTLY ONE AT A TIME (`gh run rerun <id>`, wait for `completed`). Queueing several reruns cancels the rest again. After the train: resync commit (CLI requires + go.sum) lands on master; the tagged commit's `module-isolation` red is EXPECTED until then. Full flow: `docs/release-procedure.md` "Tag pushing". Also: `toolsdk/v*` was missing from release.yml triggers until 4f37ec2 (toolsdk releases needed manual dispatch before that).
- **Master BRANCH pushes can drop the CI event too (2026-10-01, f394f54)** — the push updated origin/master but created NO ci.yml run (and no cancel, despite ci.yml's `cancel-in-progress: true` concurrency — proof the event never fired, same class as the v1.9.0 tag incident). After ANY master push, verify a CI run exists for the pushed SHA (`gh run list --workflow=ci.yml --json headSha`); if absent, check nothing is queued/in progress, then `gh workflow run ci.yml --ref master` and watch it green before tagging.
- **CI job runtime must fit its runner class** — before adding or retuning a CI job, state its expected runtime vs the runner (2-core public runners need ~45min for the count=10 benchmark job; a 15min cap silently cancelled it for a full session). A gate that times out is a dead gate.
- **`gh -R` for all cross-repo operations** — local directory names lie after renames (e.g. `~/projects/hierarchical-errors` is erraudit's pre-rename remote; `gh issue create` there filed into erraudit twice). Always `gh -R owner/repo ...` and `gh repo view` before the first cross-repo write.
- **Author commit messages AFTER `git status` confirms the batch** — the auto-commit daemon absorbs staged files seconds after staging; grand batch-sized messages routinely land over 1-3 file diffs. Check what is actually staged before writing the message.

### Core type rules

- **Branded types prevent mixups** — `ID`, `RuleName`, `ToolName`, `FilePath`, `GroupID` are distinct string types (`branded_types.go`). Use `finding.ID("x")` / `finding.FilePath("p")`; string literals auto-convert. JSON marshals as plain string. `GroupByFile` returns `map[FilePath][]Finding`.
- **Confidence is a named type** — `type Confidence float64` with `IsValid()`/`Clamp()`; `NewFinding`/`Builder` accept it (not raw float64), defaulting to `ConfidenceFull`; `ParseConfidence("high")` inverts `String()`, errors match `ErrInvalidConfidence` via `errors.Is`.
- **Position.Offset uses -1 sentinel** — `Position{}` zero value means byte 0; constructors set Offset=-1 for "unset". Check with `HasOffset()` (>= 0). Line=0 file-only positions are valid: `validateIdentity()` uses `HasFile()`; `Position.IsValid()` still requires Line>0 (backward compat).
- **GenerateID is length-prefixed** — `writeLenField` (uint32 big-endian) prevents hash collisions when field values contain colons.
- **Range end conventions** — `Range.EndOrStart` / `EndOffsetOrStart` give the effective end for single-point ranges (`End.Line == 0` or `End.Offset < 0` means "same as Start"); overlap/intersection use them.
- **FixStrategy normalized** — `NormalizeFixStrategy()` converts "" to "none" (Builder.Build, SARIF import, Equal short-circuit). `FixStrategyAI` is reserved (no backend). `HasFix()` for Direct requires BeforeCode/AfterCode OR an `Edits` list (`HasEditList()`); `HasCodeChange()` accepts any of the three.
- **Multi-edit fixes (issue #36, ADR #19)** — `TextEdit{Start, End Position; NewText}` (`text_edit.go`) + `Finding.Edits` (JSON `edits`) is the authoritative machine representation; BeforeCode/AfterCode stay the first-edit display summary. Insertion: `End` unset (`Offset` -1) or `End == Start`; `Position{}` is byte 0, NOT unset. Edit-list equality is order-independent (`editsEqual`, used by `Equal()`). Validate accepts edit-only Direct fixes; per-edit errors wrap as `finding.Edits[%d] is invalid`. SARIF round-trips edit lists incl. byte offsets (`sarifRegion.ByteOffset`/`ByteLength` are `*int` so offset 0 survives omitempty). `ApplySimpleFixes` REFUSES `len(Edits) > 1` (Reason points at pipeline.FixApplier) — never half-applies. Builder: `WithEdits(...)`. GenerateID excludes Edits deliberately.
- **tagsEqual fast path** — `Equal()` checks `slices.Equal` before clone+sort; same-order tags cost 0 allocations.
- **Validate() decomposed** — 6 per-field validators in `finding_validate.go`; add new rules there.
- **Severity aliases are API** — `RegisterSeverityAlias()`/`LookupSeverityAlias()` (RWMutex-guarded global map); `SeverityFromLevel(level, fallback)` maps strings incl. aliases; `PriorityString()` uses the `severityPriorities` map; `Badge()` derives from `Emoji()` (update only `Emoji()` for new mappings).
- **Deterministic JSON is mandatory** — all production marshal calls use `marshalOpts` / `prettyMarshalOpts` (`json.Deterministic(true)`); `marshalJSONString` wraps marshal-to-string. `encoding/json/v2` serializes map keys in unspecified order otherwise. Enforced by `json-deterministic-check.sh`. Wire types are separate structs where tags differ (`fixEditJSON`, `fixOutcomeJSON`, `fixApplyResultJSON`); errors serialize as message strings.
- **FindingError implements go-error-family** — `ErrorCode()` = `"finding.<category>"`, `ErrorFamily()` maps to errorfamily families; `errorfamily.Classify(err)` works on go-finding errors. See ADR #15.
- **`must[T]`** — `errors.go`; any new Must-constructor delegates to it.
- **doc.go must match current names** — after ANY rename, grep `doc.go` for the old symbol (godoc prose misleads otherwise).
- **Module-coverage evidence lives in Summary** — `Summary.FilesScanned` is always serialized ("scanned 0" is evidence, not an omission); `Summary.SkippedModules` carries nested go.mod modules a tool deliberately skipped; module-agnostic tools leave it nil and their JSON is unchanged.
- **Consumer count grows** — last audit (2026-07-22) counted 22 consumers (14 with Go code); never assert a fixed number without checking latest data.

### Core API surface

- **Report** — zero-value safe (value mutex); `findings` unexported: use `FindingsSnapshot()` (deep copy), `All()`, `FindByID()`; `NewReportFromFindings(tool, findings)` is the one-step creator.
- **Builder/Template** — `BuildOrDefault()` returns zero `Finding{}` on invalid input; `Template` stamps common fields (`NewTemplate` + `With*`), `Template.Builder()` returns a chainable `*Builder` for per-finding overrides. Convenience APIs: `ApplySimpleFixes` (core string replace), `CheckBinary`/`RunCmd` (external tool helpers -> `NewIOError`), `FormatTextRich` (emoji; `FormatText` keeps `[SEVERITY]`), `FormatTable`. Full docs: doc.go, README, `docs/guides/`.
- **Suppression** — `IsSuppressedAt` uses `Suppression.IsActive(now)` (valid Kind + Rule AND not expired); invalid suppressions are inactive.
- **GroupID groups findings** — optional `Finding.GroupID` (branded); JSON `groupId` (omitempty); in `Equal()`; builder `WithGroupID`; SARIF property `go-finding/groupId`; LSP via `LSPDiagnosticData.GroupID`; `Report.GroupFindings()` returns active grouped findings (map — order unspecified). See `docs/guides/finding-groups.md`.
- **LSP fidelity** — `LSPDiagnosticData` on `diag.Data` preserves ID/Severity/FixStrategy/Confidence/Category/Tags/code/Snippet/Suppression/Metadata/RelatedFindingIDs; round-trip lossless incl. SeverityCritical (LSP collapses to Error). `ToLSP()` re-emits `Tags` from `Metadata[LSPDiagnosticTagsKey]` (malformed entries skipped; negatives rejected).
- **SARIF options** — `ToSARIFWithOpts(WithIncludeSuppressed(), WithMinSeverity(sev))`; deprecated `ToSARIFFiltered` retained as alias. SARIF is hand-rolled (ADR #9).
- **context.Context on I/O** — `WriteSARIF`, `FindingsFromSARIF`, etc. take context first.

### Pipeline behavior

- **Pipeline.Run() is single-use** — second call returns `errAlreadyRan`.
- **StageHooks** — `Config.StageHooks` with `StageHook`/`StageHookFunc`; both StageBefore and StageAfter errors abort.
- **StageTiming closure must run exactly once** — `Metrics.RecordStage` uses `+=` (`metrics.go:50`); invoking the done-closure on both success and error paths double-records (past bug in `pipeline_iteration.go`).
- **RetryConfig validation uses named sentinels** — `pipeline/retry.go`; NEVER inline `errors.New` in validation returns (breaks `errors.Is`).
- **ResolveSafePath is the path traversal security boundary** — `pipeline/path_safety.go` resolves symlinks and verifies containment within root before ANY filesystem op on Finding paths. `ResolveRoot` + `ResolveSafePathFrom` batch-cache (resolve root once); `ResolveSafePath` is the single-call wrapper.
- **FixEngine** — byte-level, descending-offset application; all edits resolve against one original content snapshot; O(F+R) single pass. Provider chain (v1.14.0): EditList -> Offset -> Line -> Substring (fallback), custom providers prepended; `lineIndexAware` caches the line index lazily per file. `SubstringProvider` is column-aware. Applied counts are PER FINDING (not per edit): `apply()` tracks each edit's owning input finding (`editOwner`, parallel to `allEdits`) and `applyEditsWithConflicts` appends a finding once — a multi-edit fix counts once everywhere, O(1), no post-hoc dedup (an earlier full-equality O(n²) dedup regressed GoAST_1000 by +706% and was replaced).
- **Fix outcomes (v1.7.0)** — `ApplyWithOutcomes` returns `FixApplyResult` with one `FixOutcome` per input finding (input order): applied / no-change / refused / conflict / invalid / failed. `Apply`/`ApplyWithConflicts` delegate (unchanged shapes). `OutcomeFor`/`OutcomeCounts`/`HasErrors` query. Failed `Err` values are typed `*finding.FindingError` with position; `errors.Is`/`As` reach the provider cause. Guide: `docs/guides/outcomes.md`.
- **Rollback is per-file by default (ADR-016)** — `RollbackPolicyFailingFile` restores only the failing file; opt into all-or-nothing via `SetRollbackPolicy(RollbackPolicyAllFiles)` / `Config.FixRollbackAllFiles` / config `fixRollbackAllFiles` / CLI `-fix-rollback-all`. Soft per-finding failures never abort: applied edits stay, errors surface in outcomes + joined error. `ApplyWithReport` returns `ApplyReport`; `FailedOutcomes()` isolates failures; `Metrics.RecordOutcome`/`OutcomeCounts` aggregate; CLI prints `Fix outcomes:` summary.
- **`RolledBack` lists backed-up files, not modified files** — soft-failed files are still backed up; restoring them is a content no-op but they appear in `RolledBack` and the `(rolled back: ...)` error text. Tests must expect them.
- **NewFixApplier returns error** — propagates backup dir creation failures.
- **LineShiftMap** — `ShiftedPosition` shifts line+column; `ShiftedRange` shifts both endpoints.
- **FlightRecorderHook** — wraps Go 1.25 `runtime/trace.FlightRecorder`; snapshots on `SlowStageThreshold` or `Snapshot(ctx, reason)`; `OnStageEvent` never errors; degraded mode when Go's singleton recorder is taken (check `Degraded()`); `Close()` waits for in-flight snapshots. Config via `FlightRecorderFileConfig` + `ResolveFlightRecorder()` (5 string-encoded fields) with full CLI parity (`flightRecorder` config section vs `-trace` flags) — keep both structs in sync. CLI flags: `-trace`, `-trace-dir`, `-trace-slow`.
- **EditListProvider (v1.14.0)** — first in the chain, handles `HasEditList()` findings: offsets direct, else line/col via the shared line index. Cross-file lists fail `ErrEditCrossFile`; stale/out-of-bounds offsets fail `ErrEditStale` (both `FixOutcomeFailed`, cause reachable via `errors.Is`). Loud failure over partial application, at every layer.
- **Analysis BeforeCode/Edits** — `analysis.FromDiagnostic` reads source files to extract `BeforeCode` from TextEdits; since v1.14.0 it fills `Finding.Edits` with ALL TextEdits of the first SuggestedFix (edits 2..N were previously dropped, issue #36) and `ToDiagnostic` re-emits the full list. SuggestedFixes 1..N are ALTERNATIVES — first wins, documented limitation.

### Testing & tooling

- **`nix run .#lint` covers the ROOT module only** — same module-boundary blind spot as `.#test`: its `golangci-lint run ./...` never reaches toolsdk/pipeline/analysis/CLI, while CI lints each module separately (2026-10-11: master's `lint (toolsdk)` job was red on 4 findings while local reported "0 issues"). Lint sub-modules directly: run `golangci-lint run ./...` inside each module dir; the devShell binary matches the ci.yml action pin (2.14.0), so local per-module runs have CI parity.
- **math/rand split** — production `math/rand/v2`; tests `math/rand` (v1, `testing/quick` constraint).
- **testify is transitive only** — `// indirect` via ginkgo/slim-sprig; not used directly (banned but unavoidable).
- **NewParallelGomega** — per-module `testutil_test.go` helper (Helper + Parallel + NewWithT in one). `paralleltest` linter disabled because it can't trace `t.Parallel()` through the helper; re-enabling means inlining or nolint-ing 100+ tests.
- **makezero `always: false` (intentional)** — `always: true` flags idiomatic `make+copy` (23 FPs); `false` still catches real `make+append` over-allocation.
- **Stress gate is mandatory** — `ginkgo -r --race --repeat=20 --skip-package=examples` (core, pipeline), `go test -race -count=20` (analysis, CLI) before any tag (release-procedure step 4). CI stress now mirrors this exact split (ginkgo CLI pinned v2.32.0, job timeout 30m): the old `go test -count=20 ./...` job failed instantly on ginkgo suites ("Only -count=1 is allowed") and stressed nothing — invisible while CI was billing-dead.
- **`pipeline/examples/`** — one runnable program per subdir + `example_compile_test.go` (`go build -o /dev/null`), mirroring root `examples/`.

## CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, CSV, TSV, JSON, SARIF output (markdown/CSV/TSV via go-output adapter)
- YAML/JSON config (`-config`), severity filter (`-min-severity`), profiling
- `-filter-generated` — removes findings from auto-generated files (sqlc, protobuf, etc.)
- `-fix-provider go-ast` — enables AST-aware fix provider
- `-fix-rollback-all` — opt into all-or-nothing rollback (default: per-file)
- `-byte-level-conflict` — precise overlap detection
- `-trace` — enable Go execution trace flight recorder for diagnostics (`-trace-dir`, `-trace-slow`, `-trace-max-files`, `-trace-gzip` for config)
- `Fix outcomes:` stderr summary after fix stages (canonical order, nonzero counts only)
- Config-file `flightRecorder` section — alternative to `-trace` flags; supports `enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes` (full parity with pipeline `FlightRecorderFileConfig`)
- Dynamic detector registry (`RegisterDetector`)

## Architecture Decisions

ADR log: `docs/architecture-decisions.md` (#19 = typed edit lists for multi-edit fixes, issue #36; #16 = per-file rollback default, v1.7.0). Key structural calls:

- **SARIF hand-rolled, not go-sarif** — custom property bag, streaming + context. ADR #9.
- **go-error-family integration** — core dependency for unified error classification. ADR #15.
- **Pipeline split** — `pipeline.go` + `pipeline_detect.go`, both under 350 lines.
- **FixProvider chain** — EditList -> Offset -> Line -> Substring (fallback), custom prepended; `lineIndexAware` lazy line-index caching; GoASTProvider in `pipeline/goast/` (opt-in go/parser). Typed edit lists for multi-edit fixes: ADR #19.
- **IntervalIndex[T]** — generic sorted-slice O(n + k) overlap queries, used by Correlate (interval-tree NO-GO note in ROADMAP).
- **DetectorRegistry** — thread-safe plugin architecture (`Register`/`Build`/`BuildAll`); **MergeIter** — streaming `iter.Seq[Finding]` merge with dedup.
- **ConfigFile** — JSON config loading with `ResolveDetectors`/`ResolveProviders`.
- **go-output CLI adapter** — markdown/CSV/TSV live in the CLI module only; root `finding` stays without it. See `docs/PRO_CONTRA_go-output-integration.md`.
- **Naming history** — `FindingTransformer` (was FindingProcessor), `Conflict` (was ConflictInfo), `-min-severity` (was `-severity`, alias retained), `CategoryOf` (was GetCategory).
- **Pipeline convenience functions** — `pipeline.Detect(ctx, detectors...)`, `pipeline.ApplyToContent(content, fixes)`.
- **FixEngine guide** — `docs/guides/fix-engine.md`; outcomes guide `docs/guides/outcomes.md`.

## Test Organization

**One test file per production file.** All tests for a subject live in `<subject>_test.go` — no `_extra`/`_bugfix`/`coverage` suffixes (split by subject, put regression tests in the parent file, never name files after metrics). Shared helpers in `testutil_test.go`; one `example_test.go` per package (don't fragment).

## Removed APIs (v1.0.0)

All deprecated APIs from v0.6.0–v0.9.0 have been removed. No deprecated APIs remain.

See `docs/MIGRATION_v1.0.md` for migration details.

---

_Assisted-by: Crush <crush@charm.land>_
