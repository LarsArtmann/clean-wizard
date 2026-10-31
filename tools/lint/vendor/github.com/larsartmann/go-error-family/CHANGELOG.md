# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.12.0] - 2026-10-14

### Fixed
- **`(*Error).Unwrap` is nil-receiver safe** (DD22): a nil `*Error` unwraps
  to nil instead of panicking - defensive guard for callers that forward
  concrete pointers into `error` interfaces and unwrap without a nil check
  (the gotcha-#213 typed-nil class). Red-first test pins it.

## [0.11.0] - 2026-09-29

Discoverability release: **no public API changes.** Constructor-example
symmetry, first `errorfamilytest` examples, an executable `RuleSpec` example
for `diagnose`, the gRPC status-mapping guide, and the release-engineering
guards (website canary, art-dupl baseline, release dispatch fallback).

### Added

- **Root godoc examples: family-constructor symmetry complete** — `ExampleNewConflict`, `ExampleNewCorruption`, `ExampleNewInfrastructure`, `ExampleNewOrchestration` join `ExampleNewTransient`; each shows the family's canonical HTTP status and one behavioral field (retryability / severity / exit code). Root package now carries 30 runnable examples.
- **`errorfamilytest` gains its first godoc examples** — one per assertion helper (`AssertFamily`, `AssertCode`, `AssertRetryable`, `AssertContext`, `AssertContextMissing`, `AssertExitCode`, `AssertHTTPStatus`), showing the exact call shape on pkg.go.dev.
- **`diagnose` gains an executable `RuleSpec` example** — the data-driven rule pattern end-to-end: spec matching (`ContextKeys` + `ContextSubstr`), a custom rule type, `Runner` registration, structured `Fix`, and `ResolveContextKey` lookup.
- **Website: gRPC status-mapping guide** — the six families mapped to gRPC status codes with an interceptor pattern, mirroring the HTTP guide.

### Changed

- **Documentation: canonical references brought current with the six-family protocol (docs-health pass 2026-09-27)** — SKILL.md (the full API reference) had zero mentions of `Orchestration`: family table, severity order, HTTP mapping, audience mapping, and constructor lists now include the 6th family; `docs/DOMAIN_LANGUAGE.md` gained `Orchestration` in the Family/Audience definitions and the `HTTPStatuser` interface row; README's "four interfaces" claim corrected to the six-interface contract (with `HTTPStatuser`), and the architecture tree now lists `registry.go`, `stdlib.go`, `bridge/`, and the bridge/checkout examples; website `contributing.mdx` "four interfaces" and `quick-start.mdx` "Five Families" heading fixed (both stale since v0.8.0/v0.10.0).
- **README: test-helper and diagnostic-rules sections name the new example coverage** (`AssertContextMissing`/`AssertHTTPStatus`, `RuleSpec` example).

### Fixed

- **erraudit regression in `safeCauseString`** — the Sep 22 erraudit binary added a blank-identifier-on-`recover` rule after the 2026-09-15 zero-findings verification; the deliberate silent recovery now carries the documented `//nolint:legacyerrors` suppression with its rationale. Battery re-run confirms 0 findings across all 7 modules (2026-09-28).
- **`.gitattributes`: `*.png binary`** — structure-linter advisory; prevents diff/merge corruption for website OG images.

### Release engineering

- **`release.yml`: `workflow_dispatch` fallback + explicit `tag_name`** — a swallowed tag-push webhook (the v0.10.2 incident: the tag ref never produced a push event, so the release silently skipped) can now be re-run from the Actions UI; the runbook in AGENTS.md mandates separate tag/branch pushes and a ~2 min trigger check.
- **`website-check.yml`: website guard canary** — fails on typescript major ≠ 6, frozen-lockfile drift, `astro check`, or `astro build` on every `website/**` push/PR; ends the three-time deploy-breaker class structurally.
- **`.art-dupl-baseline.json` (threshold 1)** — records the three accepted clone groups; `art-dupl check -t 1 .` reports only new clones. Routine gate is the default `-t 5`; deep sweeps run `-t 1` against the baseline (resolves ROADMAP OQ1 as a recorded default).
- **Website: `minimumReleaseAgeStrict: true`** — the pnpm 11 one-day release cooldown now fails loudly instead of silently falling back to a too-fresh version; frozen installs verify the lockfile against supply-chain policies (492 entries, clean).

### Modules

- `go-error-family` -> **v0.11.0**, `diagnose` -> **v0.2.6**. `agent` (v0.2.5), `bridge` (v0.3.4), `diagnose/git` (v0.5.4), `diagnose/postgres` (v0.5.4), and `examples` (v0.3.2) have no Go changes in this release and stay put; their root-version pins ride the post-release Dependabot PRs.

## [0.10.3] - 2026-09-28

Consumer-protection release: **no public API changes.** Retracts every
broken pre-extraction tag family so consumers can no longer resolve them
to broken module graphs from the module proxy.

### Deprecated

- **Retracted the broken root tags `v0.5.0`, `v0.5.1`, and `v0.6.0`** — all three shipped
  local-directory `replace` directives (`./agent`, `./diagnose`) that break every consumer
  building from the module proxy (directory replaces cannot resolve outside this workspace);
  `v0.6.0` additionally pinned a phantom `v0.0.0-00010101…` pseudo-version. `v0.6.1` is
  clean (zero requires, pre-extraction subpackage layout) and stays available. The retract
  block in the root `go.mod` now carries the reason, so `go` tooling refuses quiet upgrades
  and reports the retraction. Superseded by v0.10.3.

### Fixed

- **`diagnose/v0.1.0` retracted (diagnose v0.2.5)** — its `go.mod` carried a local-directory
  replace (`=> ..`), equally unusable from the proxy. `agent/v0.1.0` retracted in agent
  v0.2.5 (replaces `=> ..` and `=> ../diagnose`). These are the only two submodule tags
  with replace directives; every other submodule tag audited clean (module path + zero
  replaces across all 60+ tags).
- **Release scope note:** only the three affected modules release here (`diagnose` v0.2.5,
  `agent` v0.2.5, root v0.10.3). `bridge`, `diagnose/git`, `diagnose/postgres`, and
  `examples` have no broken tags and no go.mod changes, so they stay at v0.3.4 / v0.5.4 /
  v0.5.4 / v0.3.2 — a coordinated 7-module bump would only widen the documented
  pin-bump-to-tag-push breakage window for the externally-dependent modules.
- **CHANGELOG correction (0.10.2):** the Modules section claimed submodule `go.mod` files
  reference root v0.10.2; they actually pin root **v0.10.1** (the new root version cannot
  exist when the submodule tags are cut). Root pins ride the post-release Dependabot PRs.

## [0.10.2] - 2026-09-22

Release-engineering release: **no public API changes.** Completes the
`exhaustruct` → `exhaustruct_v5` lint migration, re-aligns every module's
`go` directive with its true dependency floor, and clears the website's
TypeScript-7 recurrence plus 8 transitive vulnerabilities.

### Added

- **Six godoc example functions for the under-adopted API surface** (pkg.go.dev discoverability) — `ExampleWrap` (the most-used constructor previously had no example), `ExampleHandleErrorWithContext` (the canonical entry point), `ExampleIsRetryable` (including the unknown-error fail-open default), `ExampleFamily_RetryPolicy` (advisory single-attempt vs Transient defaults), `ExampleLogError` (family→severity mapping: Transient→Warn, others→Error, with time-stripped deterministic output), and `ExampleHTTPHandler` (end-to-end: handler returns a classified error, response is a safe JSON body with per-error `WithHTTPStatus(404)` override). Root package now carries 26 runnable examples.

### Changed

- **Lint: completed the `exhaustruct` → `exhaustruct_v5` migration** in `.golangci.yml` (deprecation-warning follow-up from v0.10.1) — the linter block and its os/exec/net exclusion list moved to the v5 key (`ignore-patterns`), and test files exclude `exhaustruct_v5` alongside the other structural linters. Verified 0 issues across all 7 modules on golangci-lint v2.13.2.
- **Go directives aligned with true dependency floors in all 7 modules** — `agent`, `diagnose/git`, and `diagnose/postgres` strip to `go 1.26` once `diagnose` stops carrying a 1.26.7 floor; `bridge` and `examples` settle at `go 1.26.0` because `golang.org/x/text v0.42.0` publishes a 1.26.0 floor that `go mod tidy` enforces (dep-forced patch). Self-requires bump to the coordinated release versions. This supersedes the earlier uniform-`1.26.7` normalization: the intended floor is now per-module, not repo-wide.
- **`.buildflow.yml`: corrected the `pnpm-audit` skip rationale and documented the real failure mode** — the step runs `pnpm audit` at the repo root where no lockfile exists (it lives in `website/`), failing closed with `ERR_PNPM_AUDIT_NO_LOCKFILE`; a cached ✔ had been hiding that failure until a `BUILDFLOW_NO_RESULT_CACHE=1` run exposed it. The old rationale ("Dependabot covers it") overstated Dependabot: alerts detect website vulnerabilities, but Dependabot's automatic security-update jobs fail on the pnpm lockfile. The working remediation check is `nix develop -c pnpm audit` in `website/` (documented in AGENTS.md).

### Fixed

- **Website: TypeScript 7 re-bump recurrence and stale lockfile (third occurrence of this class)** — `website/package.json` drifted to `typescript ^7.0.2` (against the documented 6.x policy) and `@astrojs/starlight ^0.42.1`/`astro ^7.3.2` while `pnpm-lock.yaml` stayed at 6.0.3/0.42.0, failing `pnpm install --frozen-lockfile` in `website-deploy` (manifest/lockfile specifier mismatch). Reverted `typescript` to `^6.0.0`, kept the in-range starlight/astro/html-validate bumps, regenerated the lockfile, and re-verified `astro check` (0 issues) + `astro build` (15 pages).
- **Website: 8 transitive dependency vulnerabilities resolved** (6 high, 2 moderate) — `devalue < 5.9.1` (DoS), `fast-uri`, `js-yaml`, and `svgo` advisories across multiple major lines, all reachable through `astro`/`starlight`/`html-validate` whose latest releases don't yet pull patched transitives. `pnpm update --depth Infinity` bumped every transitive within its declared range (no overrides needed); `pnpm audit` is clean. This also removes the alert source that kept triggering the failing Dependabot update jobs.

### Modules

Coordinated multi-module release. Submodule `go.mod` files pin sibling versions (diagnose **v0.2.4**, bridge **v0.3.4**); the root pin stays at **v0.10.1** until the post-release Dependabot pass (the new root version cannot be referenced before it exists).

- `github.com/larsartmann/go-error-family` → **v0.10.2** (no API changes; docs + CI + website fixes)
- `github.com/larsartmann/go-error-family/diagnose` → **v0.2.4** (pin-only; go-directive floor alignment)
- `github.com/larsartmann/go-error-family/agent` → **v0.2.4** (pin-only; go-directive floor alignment)
- `github.com/larsartmann/go-error-family/bridge` → **v0.3.4** (pin-only; dep-forced `go 1.26.0` floor)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.5.4** (pin-only; go-directive floor alignment)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.5.4** (pin-only; go-directive floor alignment)
- `github.com/larsartmann/go-error-family/examples` → **v0.3.2** (six new godoc examples; dep-forced `go 1.26.0` floor)

## [0.10.1] - 2026-09-15

Documentation, CI-reliability, and error-handling-hardening release. **No
public API changes** — every Go diff in the root module is comments or lint
directives. The conditional-request classification guidance (issue #5), the
json/v2 incident resolution, and the erraudit cleanup land here.

### Added

- **CI: examples module test and lint steps** — the `test` job now runs `go test -race -count=1 ./...` in `./examples` (19 bridge-reference tests + checkout tests), and the `lint` job runs golangci-lint there, matching every other workspace module. Replaces the old examples `go build` step (`go test` compiles everything it tests).
- **Website: Bridge Patterns guide** (`guides/bridge`) — the classify→enrich→handle walkthrough: libraries classify, applications enrich; the three patterns (pass-through, `AutoWrap`, explicit `Wrap`), the oops tag/domain inference cascade, the one-error-two-representations table, and the decision guide. Linked from the sidebar and from `related-tools`.
- **Documentation: conditional-request classification guidance** (README, website HTTP guide, SKILL.md) — how to handle RFC 9110 §13 conditional-request outcomes: **304 Not Modified** is a success path (write the response and return `nil`; never classify it — every family implies 4xx/5xx plus retry/exit semantics), **412 Precondition Failed** is `Conflict` + `WithHTTPStatus(412)` (client-asserted state no longer holds = version mismatch), **428 Precondition Required** (RFC 6585 §3) is `Rejection` + `WithHTTPStatus(428)` (omitted required precondition = incomplete input, not a state clash), and **416 Range Not Satisfiable** is `Rejection` + `WithHTTPStatus(416)`. Includes a compiled, tested example (`Example_conditionalRequests`). Resolves #5.
- **Documentation fix:** "The Five Families" headings in README and SKILL.md renamed to "The Six Families" — stale since `Orchestration` was added in 0.10.0.
- **depguard deny rule for `encoding/json/v2`** (`.golangci.yml`) — re-enabled depguard with a single deny-only rule (lax mode): importing `encoding/json/v2` anywhere in the workspace is now a lint error in BuildFlow and CI. Fast local canary for the no-GOEXPERIMENT policy; the hermetic nix derivations and the `GOWORK=off` CI build remain the authoritative compile-level backstop.
- **`.buildflow.yml`** (first BuildFlow config for this repo) — skips `go-auto-upgrade` (its `jsonv1tov2` migrator re-introduced the reverted `encoding/json/v2` imports on every fleet `buildflow --fix` pass, breaking nix/CI while the machine-global `GOEXPERIMENT=jsonv2` masked it locally), `pnpm-audit` (the only lockfile lives in `website/`, a separate Node project covered by Dependabot), `go-structure-linter` (BuildFlow's embedded snapshot ignores `.go-structure-linter.yaml`), and `branching-flow` (its phantom analyzer suggests breaking phantom-type changes to published API strings and does not honor its own ignore directives — `IsIgnored` is not wired into `pkg/phantom`).
- **`.go-structure-linter.yaml`** — `flat` exclusion preset (the root package IS the published library API; `/pkg/` restructure would break 50+ consumers) plus reviewed suppressions (`go-version`: the `go` directive is a consumer-friendly minimum floor; `testdata-directory` on `diagnose/mock.go`: exported mock-injection API, not a fixture). The standalone CLI passes with this config (exit 0).

### Fixed

- **Root module re-reverted to `encoding/json`** — a fleet-wide `buildflow --fix` orchestrator run invoked `go-auto-upgrade`, whose `jsonv1tov2` migrator rewrote `encoding/json` → `encoding/json/v2` in `error.go`/`http.go` (+ tests and `examples/cmd/bridge`). The rewrite compiled locally only because this machine exports `GOEXPERIMENT=jsonv2` globally; the hermetic nix derivations (`checks.x86_64-linux.build`, `build-standalone`) and CI failed to compile. Restored the documented v0.8.0 no-experiment policy; `json.MarshalWrite` became `json.NewEncoder(w).Encode`.
- **All 36 erraudit findings resolved across 6 modules** — real error handling added in `diagnose/git` (`git status`/`remote`/`ls-remote` failures now surface as `StatusUnknown` with the run error instead of being conflated with exit codes; `git remote` failure no longer misdiagnoses as "no remotes: healthy"), `diagnose/postgres` (`pg_isready` run errors recorded in Details; `IsPostgresRunning` returns false on run error), `diagnose/command.go` migrated `errors.As` → `errors.AsType[*exec.ExitError]`, and `examples/cmd/http` now attaches `WithContext("id", userID)` on the missing-id and db-timeout paths (the not-found path already did) and fails loudly on `ListenAndServe` error. Deliberate unpropagatable writes (fmt.State formatting, best-effort CLI output, post-status HTTP body writes, reachability-probe `Close`) carry `//nolint:legacyerrors` with per-site reasons.
- **Website: `astro check` broken by TypeScript 7** — a dependency bump moved the website to `typescript ^7.0.2`, whose native compiler no longer exposes the programmatic API `astro check` (`@astrojs/language-server`) relies on, failing the `website-deploy` workflow's check step on every run. Pinned back to `^6.0.0` (resolves 6.0.3) per the upstream guidance to stay on 6.x until Astro supports the native compiler; `astro check` is green again (0 errors/warnings/hints). The same fleet `buildflow --fix` pass that re-introduced `encoding/json/v2` also re-bumped `package.json` to `typescript ^7.0.2` (and `@astrojs/starlight` to `^0.42.1`) while the lockfile stayed at 6.0.3/0.42.0, so `pnpm install --frozen-lockfile` failed in `website-deploy` with a manifest/lockfile mismatch. Reverted both to the lockfile state and re-verified `astro check` + `astro build` (15 pages).
- **Workspace build: stale `go.work.sum` checksum for `diagnose v0.2.2`** — the recorded `go.mod` checksum no longer matched the tag's bits (tag was re-pointed after the sum was recorded; `GOPRIVATE` skips sumdb verification, so the drift surfaced as a SECURITY ERROR on every workspace build). Removed the stale line; `go build ./...` re-resolves via the workspace `use` directive.
- **Website: stray compiled Tailwind artifact** — `website/src/styles/global.out.css` (build output, referenced nowhere) was accidentally committed; removed and `*.out.css` added to `website/.gitignore`.
- **CI lint version split-brain (golangci-lint v2.12.2 → v2.13.2)** — CI's pinned v2.12.2 fired `recvcheck` on `Family` and `Audience` (mixed value/pointer receivers — the pointer receiver on `UnmarshalText` is required by the `encoding.TextUnmarshaler` contract), while local/BuildFlow v2.13.2 no longer flags the pattern at all. The interim `//nolint:recvcheck` directives could not satisfy both: they sat on the line before the reported finding (no suppression in v2.12.2) and were "unused" under v2.13.2 (a `nolintlint` error — which is what auto-removed them, re-breaking CI). Fix: version parity — all ten `version:` pins across `ci.yml` and `release.yml` bumped to v2.13.2 and the directives removed. All seven modules lint clean under v2.13.2 (the `exhaustruct` deprecation warning is cosmetic; the `exhaustruct` → `exhaustruct_v5` rename stays a standalone follow-up).

### Modules

Coordinated multi-module release. Submodule `go.mod` files now reference root **v0.10.1** and diagnose **v0.2.3**.

- `github.com/larsartmann/go-error-family` → **v0.10.1** (docs, CI lint parity, website fixes; no API changes)
- `github.com/larsartmann/go-error-family/diagnose` → **v0.2.3** (erraudit fixes: git/pg run errors surface as `StatusUnknown`, `errors.As` → `errors.AsType` migration)
- `github.com/larsartmann/go-error-family/agent` → **v0.2.3** (pin-only; doc comments on published protocol fields)
- `github.com/larsartmann/go-error-family/bridge` → **v0.3.3** (pin-only; lint-directive comments in `Format`)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.5.3** (`git status`/`remote`/`ls-remote` failures surface as `StatusUnknown` with the run error)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.5.3** (`pg_isready` run errors recorded; `IsPostgresRunning` false on run error)
- `github.com/larsartmann/go-error-family/examples` → **v0.3.1** (http example attaches request-id context, fails loudly on `ListenAndServe`)

## [0.10.0] - 2026-07-26

Adds the `Orchestration` family — the 6th behavioral family — for internal
coordination failures: the class
of bug where the program's own logic failed to complete an operation (rendering
output, building a CLI command, wiring dependencies). Distinct from
`Infrastructure` (the system cannot serve) and `Corruption` (source of truth is
damaged): nothing is wrong with I/O or data, the program simply has a bug or
misconfiguration in its own orchestration layer. Not retryable; no user-facing
fix (report the bug).

### Added

- **Bridge reference implementation** (`examples/cmd/bridge/` + `examples/checkout/`) — canonical example of the classify→enrich→handle flow using oops + bridge + error-family. Demonstrates three patterns: pass-through (library classification survives oops enrichment), AutoWrap (oops-first errors classified by bridge), and explicit Wrap (application assigns the family). The library layer (`examples/checkout/`) imports ONLY `go-error-family`, proving the "libraries classify, applications enrich" architecture. Includes 19 tests (8 library + 11 application) and a pattern documentation README.
- **`Orchestration` Family** (`family.go`) — 6th behavioral family. Metadata: severity 5, exit code 70 (`EX_SOFTWARE`), HTTP 500 (Internal Server Error), `ToneApologetic`, `AudienceOps`, not retryable. Sits between `Infrastructure` (severity 4) and `Corruption` (severity 6) in the total-severity order, so it dominates infrastructure outages but is itself dominated by data-integrity breaks in multi-error classification.
- **`NewOrchestration` / `WrapOrchestration` / `WrapOrchestrationf`** (`constructors.go`) — family-specific constructors matching the existing `New{Family}` / `Wrap{Family}` / `Wrap{Family}f` pattern.
- **Orchestration integration tests** (`orchestration_test.go`) — end-to-end coverage: constructor -> `Classify` -> exit code (70) -> HTTP status (500) -> retryability (false) -> severity ordering (between Infrastructure and Corruption), plus multi-error worst-severity-wins ordering and `Wrap*`/`Wrapf*` chain preservation.

### Changed

- **`Corruption` severity 5 -> 6** (`family.go`) — bumped to preserve the total order after inserting `Orchestration` at severity 5. The fail-closed retry guarantee is unchanged: any non-`Transient` sub-error (severity > 1) still makes a joined error non-retryable. The relative ordering of the five original families is unchanged.
- **`IsValid()` boundary** now spans `Rejection`..`Orchestration` (six constants).
- **Doc comments** (`family.go`, `README.md`, `FEATURES.md`) updated to reflect six families and the `Orchestration -> 500` HTTP mapping.
- **Removed 52 phantom `//nolint:hierarchical-errors` directives** across 13 Go files — the `hierarchical-errors` linter was never installed (not as a binary, not as a golangci-lint linter, not as a BuildFlow step). The directives only produced "unknown linters: hierarchical-errors" warnings on every lint run.
- **Test-file lint exclusions** (`.golangci.yml`) — added `cyclop`, `gocyclo`, `gocognit`, `maintidx` to the `_test.go` exclusion list. Test functions with many subtests legitimately exceed complexity thresholds.
- **BuildFlow pipeline verified** — 38/39 steps pass (1 skipped via config). The previously-failing `gitignore-upsetter:repair` and `nix-hash-fix` now succeed.

### Fixed

- **`examples/go.mod` phantom replace directive** — the examples module referenced `bridge v0.0.0-00010101000000-000000000000` with a local `replace` directive. Go strips `replace` on fetch, so consumers would hit an unresolvable module-graph edge. Replaced with the real `bridge v0.3.2` version and removed the `replace` directive. Same class of bug as the v0.6.0 hotfix.
- **Website deploy workflow** (`.github/workflows/website-deploy.yml`) — pinned `FirebaseExtended/action-hosting-deploy` to a valid commit SHA (`500ac625 # v0.11.0`); the previously pinned SHA did not exist in the upstream repo, failing every run at action resolution. The `FIREBASE_SERVICE_ACCOUNT_LARS_SOFTWARE` GitHub secret is now set, so production deploys to `errorfamily.lars.software` run automatically on `website/**` changes to `master` (verified green).
- **Release workflow supply-chain pin** (`.github/workflows/release.yml`) — pinned 3 occurrences of `version: latest` for `golangci-lint-action` to `version: v2.12.2`, matching the pin already in `ci.yml`. Previously every release picked up whatever version existed at tag-push time.

### Modules

Coordinated multi-module release. Submodule `go.mod` files now reference root **v0.10.0** and diagnose **v0.2.2**.

- `github.com/larsartmann/go-error-family` -> **v0.10.0** (Orchestration family + examples/go.mod fix)
- `github.com/larsartmann/go-error-family/diagnose` -> **v0.2.2** (nolint cleanup + root pin bump)
- `github.com/larsartmann/go-error-family/agent` -> **v0.2.2** (nolint cleanup + pin bumps)
- `github.com/larsartmann/go-error-family/bridge` -> **v0.3.2** (nolint cleanup + root pin bump)
- `github.com/larsartmann/go-error-family/diagnose/git` -> **v0.5.2** (nolint cleanup + pin bumps)
- `github.com/larsartmann/go-error-family/diagnose/postgres` -> **v0.5.2** (nolint cleanup + pin bumps)
- `github.com/larsartmann/go-error-family/examples` -> **v0.3.0** (bridge reference implementation + checkout example + phantom replace fix)

## [0.9.0] - 2026-07-24

Structured-logging hook for `HandleError*` and an HTTP error-path fix. All
new APIs use only the Go standard library — the root package remains
zero-dependency.

### Added

- **`HandleConfig.Logger *slog.Logger`** (`handle.go`) — optional structured-logging hook. When set, `HandleErrorWithContext` emits a single `slog` record (family, code, retryable, exit_code, and every `context.*` key) in the same call as the human-facing CLI output, classifying once and logging once. Nil (default) preserves the original behavior — fully backward compatible. Recommended for systemd/journald observability: every `HandleError` call emits a self-contained structured record without a separate `LogError` call.
- **`errorfamilytest.AssertHTTPStatus(tb, err, want)`** (`errorfamilytest/errorfamilytest.go`) — test assertion for HTTP status codes. Checks the `HTTPStatuser` interface first (per-error override via `WithHTTPStatus`), then falls back to the family default.
- **`logErrorInternal`** (`log.go`) — shared emission path for `LogError`/`LogErrorContext` and the new `HandleConfig.Logger` hook. Accepts a pre-computed family and exit code so `HandleError` does not re-classify when logging alongside human output.
- **Fuzz tests**: `FuzzWithHTTPStatus`, `FuzzRegisterClassificationType` (added to existing `fuzz_test.go`).
- **Benchmarks**: `BenchmarkWithHTTPStatus`, `BenchmarkHTTPStatusOverride`, plus `WrapOnce` benchmarks.
- **Examples**: `ExampleError_WithHTTPStatus`, `ExampleHTTPStatus`, `ExampleRegisterClassificationType`, `ExampleWrapOnce`, `ExampleError_WithExitCode`, `ExampleError_WithContextAny`.

### Changed

- **`writeHTTPError` respects per-error `HTTPStatuser` overrides** (`http.go`) — previously called `HTTPStatus(err)` which double-classified and could ignore a per-error status set via `WithHTTPStatus`. Now computes the status once from the already-classified family, then checks the `HTTPStatuser` interface for a non-zero override. Fixes a bug where `WithHTTPStatus` overrides were silently dropped on the HTTP error path.
- **`LogError` / `LogErrorContext` now emit `exit_code`** (`log.go`) — every structured log record includes the resolved BSD exit code alongside family/code/retryable.

### Modules

Coordinated multi-module release. Submodule `go.mod` files now reference root **v0.9.0** and diagnose **v0.2.1**. Submodule code is unchanged; bumps are pin-only patches so consumers of submodules transitively receive the `writeHTTPError` fix and the `HandleConfig.Logger` hook.

- `github.com/larsartmann/go-error-family` → **v0.9.0** (HandleConfig.Logger + writeHTTPError fix)
- `github.com/larsartmann/go-error-family/diagnose` → **v0.2.1** (root pin bump)
- `github.com/larsartmann/go-error-family/agent` → **v0.2.1** (root + diagnose pin bump)
- `github.com/larsartmann/go-error-family/bridge` → **v0.3.1** (root pin bump + oklog/ulid v2.1.2)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.5.1** (root + diagnose pin bump)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.5.1** (root + diagnose pin bump)
- `github.com/larsartmann/go-error-family/examples` → **v0.2.1** (root + diagnose pin bump)

## [0.8.0] - 2026-07-23

BuildFlow-inspired error handling additions. All new APIs use only the Go
standard library — the root package remains zero-dependency.

### Added

- **`HTTPStatuser` interface** (`interfaces.go`) — sixth consumer interface alongside `Coded`/`Classified`/`Contextual`/`Retryable`/`ExitCoder`. Errors implementing `HTTPStatus() int` can override the family-based HTTP status on a per-error basis. Mirrors the `ExitCoder` pattern.
- **`Error.WithHTTPStatus(status int) *Error`** (`error.go`) — copy-on-write mutator that sets a custom HTTP status code. Zero means "unset, use family default." Example: `NewRejection("battle.not_found", "...").WithHTTPStatus(404)` returns 404 instead of the family default 400.
- **`RegisterClassificationType[T error](family Family)`** (`classify.go`) — generic type-based classifier sugar over `RegisterClassifier`. One-liner for the common "type T → Family F" case.
- **`RegisterClassificationTypeFor[T error](r *Registry, family Family)`** (`classify.go`) — variant targeting a custom `Registry`. Two top-level functions because Go doesn't allow type parameters on methods.
- **`ExitCoder` interface** (`interfaces.go`) — fifth consumer interface alongside `Coded`/`Classified`/`Contextual`/`Retryable`. Errors implementing `ExitCode() int` can override the family-based exit code on a per-error basis. Embeds `error` for `errors.AsType[T]` compatibility.
- **`Error.WithExitCode(code int) *Error`** (`error.go`) — copy-on-write mutator that sets a custom exit code. Zero means "unset, use family default."
- **`Error.WithContextAny(key string, value any) *Error`** (`error.go`) — type-safe context attachment for non-string values. Uses a type switch for common scalars (string, int, int64, uint, uint64, float64, bool, []byte, time.Time, error) and falls back to `fmt.Sprint`.
- **`WrapOnce(err, family, code, msg) *Error`** (`constructors.go`) — idempotent wrap: returns the existing `*Error` unchanged if the error chain already contains one. Prevents double-wrapping at API boundaries.
- **`WrapOncef(err, family, code, format, args...) *Error`** (`constructors.go`) — formatted variant of `WrapOnce`.
- **`errorfamilytest.AssertExitCode(tb, err, want)`** — test assertion for exit codes (checks ExitCoder override first, then family default).
- **`errorfamilytest.AssertHTTPStatus(tb, err, want)`** — test assertion for HTTP status codes (checks HTTPStatuser override first, then family default).
- **`safeCauseString`** (`error.go`) — panic-recovery guard around `cause.Error()`. Defense-in-depth against third-party error types that panic on nil internal values. Applied to `Error()`, `Summary()`, and `formatVerbose()`.
- **`formatVerbose` now shows `exit_code` when non-zero** — improved `%+v` debug output.
- **Benchmarks**: `BenchmarkWrapOnceWrap`, `BenchmarkWrapOnceIdempotent`, `BenchmarkWithExitCode`, `BenchmarkExitCodeOverride`, `BenchmarkWithHTTPStatus`, `BenchmarkHTTPStatusOverride`.
- **Examples**: `ExampleWrapOnce`, `ExampleError_WithExitCode`, `ExampleError_WithContextAny`, `ExampleExitCode`, `ExampleError_WithHTTPStatus`, `ExampleHTTPStatus`, `ExampleRegisterClassificationType`.
- **Fuzz tests**: `FuzzWrapOnce`, `FuzzContextValueToString`, `FuzzWithExitCode`, `FuzzWithHTTPStatus`, `FuzzRegisterClassificationType`.
- **`Compose(errs...)`** — re-added as backward-compatibility wrapper around `errors.Join` (was removed in v0.5.0, restored per consumer feedback; commit `8cb240a`).
- **`time.Duration` case in `contextValueToString`** — common context value type (timeouts, retry intervals) now renders via `Duration.String()` instead of `fmt.Sprint`.
- **Tests**: `TestSafeCauseStringNonStringPanics` (int, nil, struct panic recovery), `TestWriteHTTPErrorMarshalFailure` (failing writer error path).
- **CI**: `GOWORK=off go list -m all` module-graph gate, consumer-simulation job (throwaway module import), `go vet ./...` step.

### Changed

- **Root module reverted from `encoding/json/v2` to `encoding/json`** — the `GOEXPERIMENT=jsonv2` requirement (introduced in v0.7.0) has been removed. The library is now a pure stdlib consumer with zero adoption friction. Only 2 call sites were affected (`Error.JSON()` and `writeHTTPError`); JSON output is byte-identical.
- **`HTTPStatus(err)` checks `HTTPStatuser` interface first** — a non-zero per-error HTTP status (via `WithHTTPStatus`) wins over the family default. Zero falls back to `Family.HTTPStatus()`. `HTTPHandler` now uses this, so per-error overrides take effect in HTTP responses.
- **`writeHTTPError` no longer double-classifies** — the `Classify` result is reused for both the response body and the status code resolution, eliminating a redundant `Classify` call on the HTTP error path.
- **`ExitCode(err)` checks `ExitCoder` interface first** — a non-zero custom exit code wins over the family default. Zero falls back to `Family.ExitCode()`.
- **`HandleErrorWithContext` and `HandleErrorDetailedWithConfig`** now resolve exit codes via a shared `resolveExitCode` helper that respects the `ExitCoder` interface.
- **`Error.Error()`, `Error.Summary()`, `Error.formatVerbose()`** use `safeCauseString` instead of `%v` format for cause rendering.
- **`contextValueToString` refactored** into `contextValueToString` + `scalarToString` dispatch, eliminating `//nolint:cyclop` suppression.
- **`WithExitCode` godoc** documents POSIX exit code wrapping behavior (0-255 range, negative values wrap via `os.Exit`).

### Modules

Coordinated multi-module release. Submodule `go.mod` files now reference root **v0.8.0** and diagnose **v0.2.0**.

- `github.com/larsartmann/go-error-family` → **v0.8.0** (new APIs + json/v2 revert)
- `github.com/larsartmann/go-error-family/diagnose` → **v0.2.0** (diagnostic rule enhancements)
- `github.com/larsartmann/go-error-family/agent` → **v0.2.0** (refactoring + dep bump)
- `github.com/larsartmann/go-error-family/bridge` → **v0.3.0** (v0.8.0 hardening + dep bump)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.5.0** (rule enhancements + dep bump)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.5.0** (rule enhancements + dep bump)
- `github.com/larsartmann/go-error-family/examples` → **v0.2.0** (dep bump + lint cleanup)

## [0.7.0] - 2026-07-09

### Changed (BREAKING)

- **Root module migrated to `encoding/json/v2`** — `error.go` (`JSON()` method) and `http.go` (`HTTPHandler` response writer) now use `encoding/json/v2` instead of `encoding/json`. This requires consumers to set `GOEXPERIMENT=jsonv2` when building on Go 1.26 (the nix devShell sets this automatically). The JSON output shape is unchanged.
- **Bridge dependencies bumped** — `samber/oops` v1.22.0 → v1.23.0, `golang.org/x/text` v0.39.0 → v0.40.0.

## [0.6.1] - 2026-07-05

Hotfix: the published `go.mod` files in 0.6.0 contained local `replace` directives plus phantom `require ... v0.0.0-00010101000000-000000000000` versions. Go strips `replace` when a module is fetched, so consumers hit unresolvable module-graph edges. `go.work` masked the defect locally.

### Fixed

- **Root `go.mod`** — removed the `replace` block and phantom `require` for `diagnose`. The root module is now genuinely zero-dependency: `GOWORK=off go list -m all` returns exactly one module.
- **`diagnose/go.mod`** — `require root v0.0.0-...` replaced with real `v0.6.0`; local `replace` removed.
- **`agent/go.mod`** — both phantom requires replaced with real `v0.6.0` + `diagnose v0.1.0`; local `replace` block removed.

### Changed

- **`examples/` is now a separate Go module** (`examples/go.mod`). Previously the root module required `diagnose` solely because `examples/cmd/custom_rule` imported it, dragging the entire `diagnose` subtree into every root consumer's module graph. Extracting examples restores the true zero-dependency invariant on root.

## [0.6.0] - 2026-07-05

Driven by consumer feedback from SEC and browser-history integrations. The root
package remains zero-dependency (all new features use only the Go standard
library: `net/http`, `log/slog`, `testing`).

### Added

- **`RegisterClassifier(func(error) (Family, bool))`** (`classify.go`/`registry.go`) — predicate-based classification for dynamic third-party errors (e.g. `*sqlite.Error`, `*pgconn.PgError`) that cannot be registered as sentinels because each occurrence is a fresh instance. Classifiers run after sentinel matching fails, in registration order; first match wins. Stored lock-free behind an `atomic.Pointer` (copy-on-write), mirroring the sentinel design. Includes package-level `RegisterClassifier`/`RegisterClassifiers` and `Registry.RegisterClassifier`/`RegisterClassifiers`; `Registry.Clone()` now copies classifiers too.
- **`Code(err) string`** (`classify.go`) — public one-liner for extracting a machine-readable error code from any error (walks the unwrap chain for the `Coded` interface). Replaces the 5-line `errors.As` boilerplate. `HandleError`'s internal `extractCode` now delegates to it.
- **`TemplateForCode(code) (MessageTemplate, bool)`** (`registry.go`/`handle.go`) — resolves a registered message template by code without wiring the full CLI pipeline. Lets HTTP/gRPC consumers look up user-facing messages directly. Available as both a `Registry` method and a package-level convenience function.
- **`Wrap{Family}f` formatted variants** (`constructors.go`) — `WrapRejectionf`, `WrapConflictf`, `WrapTransientf`, `WrapCorruptionf`, `WrapInfrastructuref`, completing the symmetry with `Newf`/`Wrapf`. All are nil-safe.
- **`HTTPStatus(err) int`** and **`HTTPHandler(fn) http.Handler`** (`http.go`) — classify→status-code helper and a ready-made net/http middleware. `HTTPHandler` wraps an error-returning handler and writes a **safe** JSON response (family/code/message) — it never leaks the raw `err.Error()`; the message comes from a registered `MessageTemplate` when available.
- **`LogError(err, *slog.Logger)`** and **`LogErrorContext(ctx, err, logger)`** (`log.go`) — structured `log/slog` logging with `family`, `code`, `retryable`, and each context key (prefixed `context.`). Transient errors log at Warn; all others at Error. Nil error is a no-op; nil logger falls back to `slog.Default`.
- **`errorfamilytest` subpackage** — test assertion helpers (`AssertFamily`, `AssertCode`, `AssertRetryable`, `AssertContext`, `AssertContextMissing`) mirroring `net/http/httptest`, keeping the `testing` import out of the production package.

### Changed

- **Classification pipeline** now has six steps (was five): registered classifiers run as step 5, before the Transient default. This is additive — errors that already declare a family (via `Classified`) bypass classifiers entirely, so the hot path is unaffected. Doc comments on `Classify` (both package-level and `Registry` method) updated.
- **`HTTPStatus` doc** now documents the rationale for each family→status mapping (notably why Corruption→500 rather than 422: a data-integrity break is the server's problem, not the client's).
- **`Code()` vs `ErrorCode()` clarified**: doc comments now explain that `ErrorCode()` is the canonical `Coded` interface contract while `Code()` is an ergonomic accessor on `*Error` (sibling of `Family()`/`Message()`); both are intentional, neither is deprecated.

## [0.5.0] - 2026-06-22

First release since `v0.4.0`. Consolidates the copy-on-write error refactor, module extraction, severity-ordered multi-error classification, lock-free sentinel lookup, structured diagnostic fixes, new family adapters, and the bridge `%s` format fix.

### Added

- **`Registry` type** (`registry.go`) — injectable classification sentinels + message templates. Replaces global mutable maps with a construct-and-pass type for test isolation (no `t.Cleanup(Unregister...)` needed) and scoped error handling within a single binary. Zero value is not usable — use `NewRegistry()`.
- **`NewRegistry()`** constructor and **`DefaultRegistry`** package-level var — backward-compatible defaults for all convenience functions (`Classify`, `RegisterClassification`, `RegisterTemplate`, etc.).
- **`HandleConfig.Registry`** field — pass a custom registry to `HandleError*` functions. Falls back to `DefaultRegistry` when nil.
- **`Registry.Clone()`** — deep-copy for inherit-and-extend patterns (start from `DefaultRegistry`, clone, register scope-specific overrides without touching the global).
- **`Registry.RegisterTemplates(map)`** — batch template registration, matching the existing `RegisterClassifications` batch.
- **`Family.Severity() int`** — total order for multi-error classification (Transient < Rejection < Conflict < Infrastructure < Corruption).
- **`Family.HTTPStatus() int`** — canonical family → HTTP status mapping (Rejection→400, Conflict→409, Transient→503, Corruption→500, Infrastructure→503).
- **`Family.RetryPolicy() RetryPolicy`** — advisory retry defaults per family (Transient: 3 attempts, 100ms–5s; others: single attempt). The library does not run the loop.
- **`Error.JSON() ([]byte, error)`** — canonical JSON (`{family,code,message,context,retryable,timestamp}`) for API boundaries.
- **`Error.WithContextMap(map[string]string)`** and **`Error.WithContextf(key, format, args...)`** — batch and formatted context attachment.
- **`RegisterStdlibDefaults(reg)`** (`stdlib.go`) — maps `context`/`sql`/`os` errors with documented rationale for ambiguous cases (DeadlineExceeded→Transient, Canceled→Rejection, etc.).
- **`diagnose/`** and **`agent/`** are now independent Go modules with their own `go.mod`, enabling independent versioning. Import paths are unchanged.
- **`go.work`** expanded to 6 workspace modules (root, diagnose, agent, bridge, diagnose/git, diagnose/postgres).
- **Experimental stability notices** in package docs for `agent`, `diagnose`, `diagnose/git`, `diagnose/postgres`, and `bridge`. Root package documented as the stable classification core.
- **Fuzz tests**: `FuzzParseFamily`, `FuzzParseFamilyRoundTrip`, `FuzzClassify`, `FuzzClassifyPlainError`, `FuzzErrorFormatting` (root); `FuzzFormat` (bridge).

### Changed (BREAKING)

- **Copy-on-write errors:** `WithContext`, `WithCause`, and `WithTimestamp` now return a NEW `*Error` instead of mutating the receiver. Fixes a data race when errors are shared across goroutines (e.g. package-level sentinels). Previous chaining calls that assumed identity preservation still compile but now get a distinct pointer.
- **Template placeholder syntax** changed from `{{.key}}` to `{key}`. The old syntax collided with Go's `text/template`. Migration: replace all `{{.key}}` with `{key}` in registered templates.
- **Severity-ordered multi-error classification:** `Classify` on an `errors.Join` result now returns the worst (highest-severity) sub-error instead of the first non-Transient one. Classification is deterministic regardless of join argument order; fail-closed retry semantics preserved.
- **Lock-free sentinel lookup:** `Registry.sentinels` changed from `map[error]Family` to `atomic.Pointer[sentinelMap]` (copy-on-write). At 50 registered sentinels, `Classify` dropped from ~1330 ns/3 allocs/1832 B to ~285 ns/0 allocs/0 B.
- **Structured diagnostic fixes:** `DiagnosticResult.SuggestedFix string` replaced with `Fix struct{Summary, Command string}`. Diagnostic rules now emit the remediation summary and shell command as distinct fields. The `agent` no longer parses suggestions — `extractCommand` and `looksLikeCommand` (40+ lines of heuristic prose parsing) are deleted; `FixStep.Command` comes directly from `diagnose.Fix.Command`.
- **Module extraction:** root module no longer contains `diagnose/` and `agent/` as sub-packages — they are separate modules. Consumers using `go.work` see no difference. Local `replace` directives added until published versions resolve the extraction.
- **`agent.Config.Enabled`** now returns `(nil, error)` instead of a synthetic `AgentResult`. Calling `Analyze` on a disabled agent is a programming error, not a silent no-op.

### Changed

- Package-level `RegisterClassification`/`RegisterClassifications`/`UnregisterClassification`/`RegisterTemplate`/`UnregisterTemplate` now delegate to `DefaultRegistry` (backward compatible).
- Template resolution (override → registry → built-in default) extracted into a single shared `resolveTemplate` helper used by both `renderCLI` and `resolveSuggestedFix`, eliminating split-brain divergence. Templates are cohesive units (What/Why/Fix belong together).
- README retry wording clarified: `IsRetryable` returns a binary signal; backoff, jitter, and idempotency are the consumer's responsibility.

### Removed

- **`Compose(errs...)`** — use stdlib `errors.Join` directly. `Classify` already classifies multi-errors. One less API surface to learn.

### Fixed

- **Data race** in `WithContext`/`WithCause`/`WithTimestamp` — these methods mutated the receiver's fields and returned the same pointer. All three now use copy-on-write via a shared `clone()` helper.
- **Bridge `Format(%s)`** returned an empty string when the wrapped error's message was empty — added a `[family]` fallback matching `Error()` and `%v`. Found by `FuzzFormat`.

### Modules

Coordinated multi-module release.

- `github.com/larsartmann/go-error-family` → **v0.5.0** (breaking changes + new APIs)
- `github.com/larsartmann/go-error-family/diagnose` → **v0.1.0** (first tagged release — structured `Fix`, `MockCommandRunner`)
- `github.com/larsartmann/go-error-family/agent` → **v0.1.0** (first tagged release — structured `FixStep`)
- `github.com/larsartmann/go-error-family/bridge` → **v0.2.0** (format fix + root v0.5.0 bump)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.4.0** (structured `Fix`, `MockCommandRunner.Set`)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.4.0** (structured `Fix`, `MockCommandRunner.Set`)

## [0.4.0] - 2026-06-17

### Added

- `Family` and `Audience` now implement `encoding.TextMarshaler`/`TextUnmarshaler`, enabling YAML/JSON config decoding of error families and audiences (e.g. unmarshalling a family from a config struct tag)
- `ParseAudience(string)` and `ParseStatus(string)` — case-insensitive string parsing, completing the enum-parse trio alongside `ParseFamily`
- `Audience.IsValid()` — validation mirroring `Family.IsValid()`, giving all three enums (`Family`, `Audience`, `Status`) a consistent validation API
- `Family.Audience()` — exposes the audience metadata (User vs Operator) for each family; audience is now a first-class field in the family metadata table
- `diagnose.Status.IsValid()` — validation consistency with `Family.IsValid()` and `Audience.IsValid()`
- `Compose` rationale documentation and `example_test.go` covering the new enum APIs
- Integration tests for `FilesystemRule` (temp-dir filesystem) and `NetworkRule` (localhost DNS/TCP), raising diagnose core coverage to ~77%

### Changed

- **BREAKING:** Removed `HandleConfig.Diagnose` bool field. Diagnostics now run whenever `DiagnosticFunc` is set — no separate enable flag. Consumers using `Diagnose: true` must drop that field; diagnostic behavior is unchanged when a `DiagnosticFunc` is configured.
- **BREAKING:** `agent.Config.Enabled` now returns `(nil, error)` instead of a synthetic `AgentResult`. Calling `Analyze` on a disabled agent is a programming error rather than a silent no-op result.
- `familyInfo` gained an `Audience` field; adding a new `Family` now requires only a single entry in the `familyData` table (previously audience was implicit).

### Fixed

- `lookupRegistered` now snapshots the classification registry map before iterating, so `errors.Is` chain walks run lock-free — eliminates a deadlock risk under concurrent sentinel registration.
- `NetworkRule` returns `StatusUnknown` when no host is found in the error context, preventing undefined DNS resolution behavior.

### Modules

Coordinated multi-module release. Submodule `go.mod` files retain their root dependency at **v0.3.0** — a valid lower bound since none use v0.4.0-only APIs. Consumers pulling root v0.4.0 get it automatically via MVS.

- `github.com/larsartmann/go-error-family` → **v0.4.0** (breaking changes + new APIs)
- `github.com/larsartmann/go-error-family/diagnose/git` → **v0.3.0** (new `Runner` field on `GitRule`; was skipped in v0.3.0 root release)
- `github.com/larsartmann/go-error-family/diagnose/postgres` → **v0.3.0** (new `Runner` field on `PostgresRule`; was skipped in v0.3.0 root release)
- `github.com/larsartmann/go-error-family/bridge` → **v0.1.1** (lint fix, transitive dependency update)

## [0.3.0] - 2026-06-01

### Added

- `HandleErrorWithContext(ctx, err, cfg)` — new entry point that propagates caller context to diagnostic functions (fixes context.Background() hardcode)
- `HandleErrorDetailedWithConfig(err, cfg)` — template-aware structured result for HTTP/gRPC consumers
- `CommandRunner` interface in `diagnose` package — injectable command execution for testable diagnostic rules
- `DefaultCommandRunner` struct — zero-value default wrapping `RunCommand`/`CommandExists`
- `ContextKey` typed string with exported constants (`KeyHost`, `KeyPort`, `KeyPath`, `KeyDBHost`, etc.)
- `DiagnosticResult.Context` field — surfaces the error context that triggered the rule
- `ErrorContext(err)` helper in `diagnose` package — extracts context from any error
- `Error.WithTimestamp(ts)` mutator — for testing and deterministic construction
- `Compose(errs...)` helper — combines errors via `errors.Join` for partial-success patterns
- Package-level `Example` functions: `ExampleNewTransient`, `ExampleClassify`, `ExampleHandleError`, `ExampleWrapRejection`, `ExampleParseFamily`
- Expanded git tests with mock CommandRunner: dirty tree, merge conflicts, unreachable remote, no git binary (coverage 98.5%)
- Expanded postgres tests with mock CommandRunner: pg_isready success/failure, suggestStartFix variants (coverage 81.0%)
- `UnregisterClassification` — removes a previously registered sentinel classification (for test cleanup)
- `UnregisterTemplate` — removes a previously registered message template (for test cleanup)
- `diagnose.MockCommandRunner` — shared, deterministic mock for diagnostic rule tests
- `diagnose.NewMockCommandRunner()` — constructor for `MockCommandRunner`
- `diagnose.ResolveRunner(r)` — helper that returns `r` if non-nil, otherwise `DefaultCommandRunner{}`
- `TestRunnerContextCancelledMidRun` — verifies early return when context is cancelled mid-run

### Changed

- `HandleErrorWithConfig` now delegates to `HandleErrorWithContext(context.Background(), ...)`
- `HandleErrorDetailed` now uses the full template resolution chain (registered templates, consumer overrides, family fallbacks)
- `RuleSpec.ContextKeys` field type changed from `[]string` to `[]ContextKey`
- All diagnostic rules use typed `ContextKey` constants instead of raw strings
- All diagnostic rules populate `DiagnosticResult.Context` from the error's `ErrorContext()`
- `GitRule` and `PostgresRule` accept optional `Runner diagnose.CommandRunner` field (defaults to `DefaultCommandRunner`)
- `HandleError` benchmark uses `io.Discard` to suppress stderr output (532ns/op vs 1095ns/op before)
- Improved godoc on `Family`, `Tone`, `Audience`, `HandleResult`, `MessageTemplate`, `DiagnosticFinding`
- Updated `diagnose` package doc comment with custom rule pattern and CommandRunner usage
- `Runner.Run` refactored into three focused methods (`applicableRules`, `runRules`, `sortByConfidence`) for lower cyclomatic complexity
- `Runner.Run` now respects context cancellation via buffered channels with `select` on `ctx.Done()` (previously could hang on slow rules)
- Git and postgres test packages migrated from local mock types to shared `diagnose.MockCommandRunner` (~115 lines of duplicated mock code removed)
- Git and postgres `cmdRunner()` methods consolidated via `diagnose.ResolveRunner()`

### Fixed

- `HandleErrorWithConfig` now passes caller context to `DiagnosticFunc` instead of `context.Background()`
- `HandleError` benchmark no longer writes ~1M lines to stderr during benchmark runs
- `NetworkRule.resolveHost` uses `net/url.Parse` and `net.Dialer` for TCP dial instead of raw string splitting
- `FilesystemRule` uses `filepath.Ext` instead of `strings.Contains(".")` for file vs directory detection
- `Compose` doc comment corrected; `extractCommand` updated to match diagnostic fix formats

## [0.2.0] - 2026-05-26

### Changed

- **BREAKING: Modularized diagnostic rules** — `GitRule` moved to `diagnose/git` submodule, `PostgresRule` moved to `diagnose/postgres` submodule. `DefaultRunner()` now includes only zero-dependency rules (`FilesystemRule`, `NetworkRule`). Consumers must opt into git/postgres diagnostics via explicit submodule import:
  ```go
  import "github.com/larsartmann/go-error-family/diagnose/git"
  runner := diagnose.NewRunner(&git.GitRule{})
  ```
- Exported all diagnostic rule helpers: `RuleSpec`, `HasContextKey`, `ContextValue`, `ResolveContextKey`, `HasContextSubstring`, `FamilyIs`, `ErrorCodeContains`, `RunCommand`, `CommandExists`
- G304 gosec exclusion for `diagnose/rules_filesystem.go` moved from inline `//nolint` to `.golangci.yml` path-based rule (eliminates golines formatting conflict)
- Extracted string constants in postgres submodule for goconst compliance

### Added

- Fuzz tests: `FuzzParseFamily`, `FuzzParseFamilyRoundTrip`, `FuzzClassify`, `FuzzClassifyPlainError`, `FuzzErrorFormatting`
- Benchmark suite: 16 benchmarks covering `Classify`, `HandleError`, `Runner.Run`, `ParseFamily`, and more
- Runnable examples in `examples/`:
  - `cmd/cli` — CLI boundary handler with contextual messages
  - `cmd/http` — HTTP middleware with family-to-status-code mapping
  - `cmd/custom_rule` — How to implement `DiagnosticRule` from scratch
- Integration tests for git submodule using temp git repos (clean, dirty, context key resolution)
- Expanded postgres test suite with 13 `Applicable` cases, table-driven `resolveHost`/`resolvePort` tests

### Fixed

- `lookupRegistered` deadlock risk eliminated — map snapshot copied before `errors.Is` iteration (lock-free)

## [0.1.1] - 2026-05-16

### Changed

- **License changed from Proprietary to MIT** — the project is now open source
- Rewrote README to accurately reflect the actual API (previous version documented fabricated agent APIs)

### Fixed

- README AI Agent section documented non-existent `Involvement` levels, `ConfirmFunc`, and `FixStep.Risk` — replaced with actual `agent.Config{Enabled, Timeout}` and `FixStep{Description, Command, Rationale}` API
- README contained a dead link to a non-existent design doc in `docs/` — removed
- `Newf` code example was missing the `errorfamily.` package prefix on `Rejection` — fixed

### Added

- Badges (Go Reference, Go Report Card) and Installation section to README
- CLI Boundary section documenting `HandleErrorWithConfig`, `HandleErrorDetailed`, `HandleConfig`, `MessageTemplate`
- Template Resolution subsection (What/Why/Fix/WayOut lookup precedence)
- Classification precedence explanation
- `ParseFamily`, `Audience()`, `Tone()`, `RegisterClassifications` (batch) documentation
- Architecture tree now accurately reflects exports (e.g. `context.go` is internal)
- Repo made public with description and topics

## [0.1.0] - 2026-05-10

### Added

- `Family` enum: Rejection, Conflict, Transient, Corruption, Infrastructure
- Small interfaces: `Coded`, `Classified`, `Contextual`, `Retryable` (each embeds `error`)
- `Error` struct: reference implementation with `Is`, `Unwrap`, `Format`, `WithContext`, `Summary`
- Family-specific constructors: `NewRejection`, `WrapTransient`, etc.
- `Classify(err)` — universal classification for any error
- `ExitCode(err)` — BSD sysexits.h exit codes from Family
- `IsRetryable(err)` — retry decision from Family
- `HandleError(err)` — CLI boundary handler with structured messages
- `HandleErrorWithConfig` — configurable handler with template overrides and diagnostics
- `HandleErrorDetailed` — structured result for HTTP/gRPC handlers
- `MessageTemplate` — Wix-style What/Why/Fix/WayOut templates with `{{.key}}` substitution
- `RegisterTemplate` — global template registry
- `RegisterClassification` / `RegisterClassifications` — map third-party errors to families
- Diagnostic rules: `PostgresRule`, `FilesystemRule`, `NetworkRule`, `GitRule`
- `diagnose.Runner` — concurrent rule execution with confidence-sorted results
- `agent.DebugAgent` interface — root cause analysis and `FixStep` suggestions
- `ParseFamily` — case-insensitive string-to-Family (defaults to Transient for unknowns)
- `Audience` and `Tone` types for presentation-layer decisions
