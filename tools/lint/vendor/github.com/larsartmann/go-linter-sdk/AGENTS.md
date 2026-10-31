# AGENTS.md — go-linter-sdk

Shared scaffolding for LarsArtmann Go linters (`Rule` interface, registry, detector adapter, exit codes). Eliminates the per-linter converter layer by codifying the `go-structure-linter` pattern: rules emit `finding.Finding` directly.

## Go 1.27 and `encoding/json/v2`

`go-finding` uses `encoding/json/v2` and `encoding/json/jsontext`. As of Go 1.27
the `encoding/json/v2` experiment is **on by default**, so plain `go build` /
`go test` work with no environment override. Never set `GOEXPERIMENT=none` — it
disables the toolchain's default experiment set and the build fails with
`build constraints exclude all Go files`. (Historic: on Go ≤ 1.26 every command
needed `GOEXPERIMENT=jsonv2`; that scaffolding was removed 2026-10-11.)

| How                   | Command                                               |
| --------------------- | ----------------------------------------------------- |
| Nix app (recommended) | `nix run .#test`, `nix run .#build`, `nix run .#lint` |
| Dev shell             | `nix develop`, then `go test ./...`                   |
| Inline                | `go test ./...`                                       |

## Build Commands

```bash
nix run .#test          # go test ./... -count=1
nix run .#test-race     # go test ./... -race -count=1
nix run .#build         # go build ./...
nix run .#lint          # golangci-lint run ./...
nix run .#vet           # go vet ./...
nix run .#coverage      # go test ./... -coverprofile=reports/coverage.out
nix flake check         # validate flake
```

No Makefile. No justfile. Everything via `flake.nix`.

## Architecture

Core types in package `linter`:

- **`Rule` interface** — `ID()` / `Name()` / `Description()` / `Category()` / `Severity()` / `IsEnabledByDefault()` / `Check(ctx, dir) ([]Finding, error)`. Rules emit `finding.Finding` directly (no intermediate Violation/Issue type). Dual identity: `ID()` is the stable key (registry dedup, suppression, filtering, finding emission); `Name()` is the mutable display name.
- **`RuleFunc` struct** — adapter combining `RuleMeta` header + `Run` closure to satisfy `Rule`. Most rules use this. Enabled by default.
- **`OptIn(rf)`** — wraps a `RuleFunc` as disabled-by-default. Use for noisy, experimental, or domain-specific rules.
- **`Registry`** — thread-safe (`sync.RWMutex`) collection of rules. `Register` panics on duplicate ID or empty identity fields (ID, Name, Description, Category). `Get(id)` / `Has(id)` for lookup; `Deregister(id)` for runtime removal.
- **`Registry.Run`** — standalone execution. Fail-fast by default; pass `ContinueOnError()` to run all rules regardless of failures, collecting partial findings and joining errors.
- **`RuleMeta.Validate()`** — checks ID, Name, Description, Cat are non-empty. Called by `Register` through the `Rule` interface.
- **`DetectorFromRegistry`** — adapts a registry to a single `finding.Detector` for BuildFlow integration. Reads working dir from context.
- **`DetectorsFromRegistry`** — returns `[]finding.Detector` (one per rule) for `go-finding/pipeline` integration with per-rule parallelism, timeouts, and error isolation. Each detector is named after the rule's ID.
- **`ExitCodeFromReport`** — binary: 0 if clean, 1 if any findings.
- **`Category`** is an open `string` type. The 8 built-in constants are recommendations; consumers can define custom categories (`Category("api")`).
- **`finding.Confidence` and `finding.FixStrategy`** are per-finding, set via `finding.NewBuilder(...).WithConfidence(...).WithFixStrategy(...)`. More expressive than rule-level defaults.

### Error Wrapping Pattern

`Registry.Run`, `RuleFunc.Check`, and `DetectorFromRegistry`/`DetectorsFromRegistry` wrap rule execution errors into `*RuleError` (carrying the rule ID). Use `errors.AsType[*RuleError](err)` to extract which rule failed, or `errors.Is(err, ErrRuleFailed)` for a boolean check. `Registry.Run` avoids double-wrapping: if the error is already a `*RuleError` (from `RuleFunc.Check`), it passes through.

## Dependencies

- **`go-finding`** (public repo; the pinned version lives in `go.mod` — the
  single source of truth; don't duplicate it here, and the doc-drift guard
  `scripts/check-doc-versions.sh` enforces that). Bumped repeatedly through
  2026-10, each bump gate green: build/test/race/lint/vet/flake-check. Resolved
  as a real published tag (no local `replace` directive); consumers and CI both
  fetch it from the module proxy. For local cross-repo dev against an
  uncommitted sibling checkout, add a temporary `go.work` (workspace) or a
  `replace ../go-finding` line — neither is committed.
- **go-finding is a multi-module repo; its sub-modules are irrelevant to this
  SDK's go.mod.** The repo ships root `github.com/larsartmann/go-finding` plus
  tagged sub-modules `toolsdk`, `pipeline`, `analysis`, `cmd/go-finding`. The
  sub-modules are versioned **independently, not in lockstep**: their tags can
  lead or lag the root release (check the module proxy for current tags). This SDK
  imports only the root module, so `go get github.com/larsartmann/go-finding@latest`
  is always the complete bump.
- Go 1.27 (`encoding/json/v2` is on by default; no `GOEXPERIMENT` needed).

## Gotchas & conventions

- **Flake `outputs` and the `self` argument.** The `outputs` lambda uses the
  flake-parts `inputs@{ ... }` destructure. Nix 2.34.8+ enforces strict
  argument checking on `@`-patterns, so a bare multi-argument `outputs` that
  does not tolerate `self` breaks with
  `function 'outputs' called with unexpected argument 'self'`. Historically
  this required `self,` first in the destructure (applied `c13366c`, `59e3cd4`,
  clobbered `802ff0a`, restored 2026-09-17). **Verified 2026-10-11: the current
  flake evaluates cleanly (`nix flake check --no-build` → all checks passed)
  under flake-parts with `inputs@` and no explicit `self`**, so the requirement
  is not load-bearing today. If a `nix` eval fails with that error, inspect the
  destructure before suspecting inputs or the lockfile.
- **`reports/` must exist for BuildFlow's `test-coverage` step.** It is kept
  durable via `reports/.gitkeep` plus a `.gitignore` exception
  (`!reports/`, `reports/*`, `!reports/.gitkeep`) — don't remove the marker.
  The `coverage` flake app writes `reports/coverage.out`.
- **`.golangci.yml` no longer needs the `goexperiment.jsonv2` build tag.**
  Removed 2026-10-11: Go 1.27 enables the `encoding/json/v2` experiment by
  default, so golangci-lint compiles `go-finding`'s transitive imports without
  it. Setting `GOEXPERIMENT=none` reintroduces the failure.
- **The Go version is pinned in three places; keep them in lockstep.** `go.mod`
  (`go 1.27`), `flake.nix` (`goPkg = pkgs.go_1_27`), and `.golangci.yml`
  (`run.go: 1.27.0`). treefmt's `goimports` formatter shells out to `go`, so the
  flake overrides `pkgs.gotools` to run against `goPkg`; without that the
  sandbox Go (nixpkgs default, which can lag `go.mod`'s floor) is older and
  the toolchain auto-switch tries to download a newer Go, which fails in the
  network-less `nix flake check` sandbox (root cause of the 2026-10-11 treefmt
  failure).
- **Doc drift is guarded, not just frowned upon.**
  `scripts/check-doc-versions.sh` (CI job `doc-drift`) fails when a living doc
  hardcodes a dependency version that is neither `go.mod`'s nor a repo tag, or
  when the Go pins above fall out of lockstep with `go.mod`. Living docs point
  at `go.mod`; don't re-hardcode versions there. Genuinely historical
  references take a `docguard:allow` line marker (used sparingly — each one is
  a promise to revisit on the next bump). Removing `**/*.md` / `docs/**` from
  CI `paths-ignore` was deliberate: doc-only commits are exactly the commits
  that need the guard.
- **Tests are black-box (`package linter_test`), not white-box.** This was a
  deliberate decision to satisfy `testpackage` natively (no suppression) and
  test the real consumer contract. Test factories (`makeRule`, `failingRule`)
  return the concrete `RuleFunc`, not the `Rule` interface, so `ireturn` has
  nothing to flag either. There are no unexported symbols that tests need to
  reach — if one is added, use an `export_test.go` shim rather than reverting
  to white-box.
- **Error matching: `errors.AsType` for types, `errors.Is` for sentinels.**
  `registry.go` uses `errors.AsType[*RuleError]` (Go 1.26 generic) docguard:allow (when the API was introduced, not a pin); `errors.Is`
  is correct for the `ErrRuleFailed` sentinel. Do not "migrate" `errors.Is`
  calls — see the `hierarchical-errors` skill for the decision tree.
- **Everything is public; CI needs no auth.** The repo went public 2026-09-08,
  and both deps (`go-finding`, `go-error-family`) are public too — CI resolves
  them from the module proxy with no `GOPRIVATE` and no token secret (the old
  `PRIVATE_REPO_TOKEN` scaffolding was removed). Assume anything committed is
  world-visible, including `docs/status` and `docs/planning`.
  A one-off tidy on a fresh machine may still need `GIT_CONFIG_COUNT=1
GIT_CONFIG_KEY_0="url.git@github.com:.insteadOf"
GIT_CONFIG_VALUE_0="https://github.com/"` if global git config is read-only
  and a command must fetch over VCS instead of the proxy. The global
  `GOPRIVATE=github.com/larsartmann/*` + SSH `insteadOf` setup deliberately
  STAYS (decided 2026-09-09): the still-private sibling repos
  (`branching-flow`, `erraudit`) need VCS auth, Go has no per-module
  `GOPRIVATE`, and this repo resolves fine regardless (proxy works when
  `GOPRIVATE` is empty; SSH works when it is not). (Verified
  2026-09-09: proxy-only `go mod download` + `go mod verify` succeed with no
  auth at all — the workaround is only needed when deliberately bypassing
  the proxy.)
- **setup-go injects `GOTOOLCHAIN=local` and it beats workflow-level env.**
  setup-go's version manifest lags `go.dev` by hours, so `go-version: "1.27"`
  can resolve an older patch than `go.mod` requires (`go: go.mod requires go
  > = 1.27`). Only a STEP-level`env:`takes precedence over the`GITHUB_ENV`injection — every go step in`ci.yml`pins`GOTOOLCHAIN: go1.27.0` so Go
  > downloads the exact patch.
- **The coverage gate measures library packages only.** `examples/` are
  exercised via `exec.Command` from integration tests, so their statements
  never attribute (`./...` totals 44.3%; the library alone is 98.7%). CI runs
  `go list ./... | grep -v /examples/ | xargs go test ...` — don't "fix" this
  back to `./...`.
- **Released versions: v0.3.0 (2026-09-08), v0.3.1 (2026-09-09).** Tag only
  when CI is green on the exact commit — v0.3.0 was pushed before its CI run
  finished (failed in 7s); v0.3.1 waited for green on `88bf9fa` first.
  - **A commit whose CI never went green can carry a stale `go.sum`:** the
    v1.7.0 bump commit's 7s-red CI meant its `go.sum` was never validated by docguard:allow (historical incident; the version is history, not a pin)
    the tidy check; the next push failed on "Check go.mod is tidy".
    After ANY dependency change, run `go mod tidy` locally and
    commit the result before pushing.
  - pkg.go.dev freezes the tagged README; v0.3.1 exists mostly to refresh it
    (docs-only releases are fine for that — PATCH version).
- **`dprint.json` re-exists (re-added 2026-09-09) but is NOT wired into treefmt.**
  It is a standalone formatter config; wiring it into treefmt would fetch wasm
  plugins from the network at runtime and break hermetic `nix flake check` (only
  vendored plugins would work). `nix fmt` covers Go + Nix; markdown is
  hand-formatted under the docs-health process — that is a decision, not a
  gap.
- **Public repo, open support posture (decided 2026-09-09).** Anything
  committed is world-visible (including `docs/status` and `docs/planning`).
  Issues/PRs from strangers are accepted, best-effort, no SLAs — `SUPPORT.md`
  and `SECURITY.md` (private advisories) are the policy docs. Branch
  protection is deliberately OFF: the auto-commit daemon pushes directly to
  master, and required-status checks would wedge it.
- **Doc one-home rule: FEATURES = what exists, ROADMAP = what's next.**
  `FEATURES.md` tracks only capabilities that have code today
  (`FULLY_FUNCTIONAL` / `PARTIALLY_FUNCTIONAL` / `BROKEN`). Not-yet-built
  capabilities live exclusively in `ROADMAP.md` until they graduate to
  `TODO_LIST.md`. This kills the FEATURES↔ROADMAP PLANNED split-brain that
  recurred across sessions 2-5.
- **Layering vs sibling libs (decided 2026-09-17).** The SDK deliberately does
  NOT depend on `cmdguard` (CLI framework) or `go-output` (renderer): it is a
  library that stops at findings/report/exit codes, and its single-dep policy
  (`go-finding` only) would be destroyed by cmdguard's transitive cobra/fang/
  lipgloss/samber-do/koanf tree. The intended stack for consumer linters:
  rules (this SDK) → report (go-finding) → CLI shell (cmdguard) → rendering
  (go-output). The README's "Companion libraries" section recommends cmdguard
  and go-output to end users; keep that section accurate if the siblings'
  import paths change.
- **User decisions recorded 2026-09-09 (P0 of the 47-task plan):**
  1. **Version label: v0.3.1 stood; v0.4.0 taken 2026-10-11.** The original
     rule ("no next-minor until real Go API movement") targeted docs-only
     adjacent tags; the 1.26→1.27 toolchain floor bump is a consumer-visible
     requirement change that a PATCH would hide, so the 0.x MINOR is the
     honest slot. All six consumers already ran Go 1.27 (verified), so the
     practical breakage was empty. Future floor/API bumps: MINOR. Docs-only
     pkg.go.dev refreshes: PATCH.
  2. **PAT rotation: user-owned.** The `PRIVATE_REPO_TOKEN` Actions secret is
     deleted; whether the underlying token still lives in the GitHub account
     and is used elsewhere is only knowable by the user — rotate at will.
  3. **Daemon push policy: (b) — the assistant pushes at session end.** The
     daemon commits but never pushes; pushing manually at end-of-session is
     now a mandatory step. (Option (c), daemon-pushes-on-green-CI, stays off
     the table while branch protection might ever be enabled.)
- **Auto-git daemon manages commits.** The daemon commits automatically with
  generated messages (sometimes low-quality). Do not rewrite history to fix
  them — `CHANGELOG.md` is the narrative of record. Focus on keeping the
  CHANGELOG accurate rather than the `git log` pretty.
- **This checkout is shared; parallel sessions are real.** (2026-10-11: a
  multiedit was rejected with "file has been modified since it was last read"
  because another Crush session rewrote AGENTS.md mid-edit.) Before the first
  edit of a session, run `git log --oneline -5` and treat rapid daemon commits
  as a live-session signal; before editing any file you read more than a few
  minutes ago, re-read it. Never force a clobber over an unexpected change —
  investigate it on its merits. A local `go.work` appearing or disappearing
  between commands is another session's doing, not corruption.
- **Post-release habit (do this every release).** After tagging: verify
  pkg.go.dev renders `@latest` (frozen README!) and check the "Imported by"
  count to see adoption. pkg.go.dev indexes on proxy activity — a `go get` of
  the new tag triggers it.
- **NEVER run `go mod tidy` / `go work sync` with a `go.work` workspace
  active** (resolved 2026-09-09 — this is the R3 tidy-mystery root cause).
  Workspace mode pulls sibling repos' test dependencies (ginkgo/gomega from
  go-finding) into this repo's `go.mod`, so CI's tidy check fails on an
  otherwise-identical go.sum — explaining the "green on `487d254`, red later"
  contradiction. Always tidy with `GOWORK=off go mod tidy` (or outside the
  workspace dir). Recorded in docs/migration-guide.md pitfalls too.
- **`go.work`'s `use ../go-finding` is a local-only convenience and is
  gitignored (never committed).** BuildFlow's `go-work-sync` reports it as a
  "stale module reference" because `DetectGoWorkDrift` only scans directories
  under the repo root, so an out-of-repo `use` can never be "found on disk"; it
  is an expected false positive for sibling-checkout workspaces, not drift to
  repair. Deleting `go.work` clears it (local builds then use the published
  `go-finding` tag from the proxy); re-add it only while testing uncommitted
  sibling changes and expect the warning again.
- **go-structure-linter's `internal/` suggestion is a deliberate non-fix.**
  This repo is a public library: its API is the root package `linter`, consumed
  by import path, and there is no non-public implementation worth hiding. The
  finding is a warning (it does not gate); do not "fix" it by inventing an
  `internal/` package.

## Consumers

Six direct consumers at `v0.3.1` (verified 2026-10-11 via `go.mod` grep plus a
project-dependency-graph scan of `~/projects`; roughly 60 further indirect
importers):

| Consumer              | Notes                                       |
| --------------------- | ------------------------------------------- |
| `go-structure-linter` | The original pilot target.                  |
| `samber-linter`       | Linter built on the SDK.                    |
| `go-humanize-linter`  | Linter built on the SDK.                    |
| `cmdguard/lint`       | The CLI-framework sibling's lint submodule. |
| `project-meta/linter` | Meta tooling consumer.                      |
| `BuildFlow/execution` | Direct dependency in the orchestrator.      |

The SDK's pilot rules were ported from `go-structure-linter`, `branching-flow`,
and `erraudit`. Converter deletion is **verified at the source level
(2026-10-11)**: no consumer retains a custom Issue/Violation converter layer
(4 MIGRATED, 2 PARTIAL with documented deliberate adapters). Per-repo evidence,
verdicts, and the refresh procedure: `docs/ADOPTION.md`.
