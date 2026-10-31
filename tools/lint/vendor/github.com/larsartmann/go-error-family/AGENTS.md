# go-error-family

Structured error protocol library. Library only — no `main`, no build system, no external deps. Full API reference: `SKILL.md`.

**Status:** v0.11.0 released (2026-09-29, discoverability release with diagnose v0.2.6; prior: v0.10.3 retracted the proxy-broken v0.5.x/v0.6.0 root tags + diagnose/agent v0.1.0). All 7 modules proxy-indexed; retraction propagation verified through the default module proxy (`go list -m` serves the retract rationale). pkg.go.dev verified 2026-09-29: root @v0.11.0 renders all 30 examples, diagnose @v0.2.6 renders `Example` (RuleSpec); **known gap:** the v0.11.0 tag predates the announcement-withdrawal README fix, so pkg.go.dev renders a dead Discussion #12 link until the next release (TODO_LIST #2). CI + Release + Deploy Website green, tests pass with `-race`, golangci-lint v2.13.2 = 0 issues in all 7 modules, erraudit (`1c6809a`) 0 findings (re-verified 2026-09-29 battery; the structure-linter `assets/` advisory is dispositioned N/A for a library). Release-era narratives: `docs/history/agents-archive.md`.
**Workspace modules:** root (zero-dep), `agent`, `bridge` (oops integration), `diagnose`, `diagnose/git`, `diagnose/postgres`, `examples`, `website`

## Quick Start

```bash
go test ./... -count=1 -timeout 120s -race   # per module; root ./... does NOT span module dirs
golangci-lint run ./...                        # lint (all modules)
go build ./...                                 # build check
md-go-validator .                              # all markdown Go blocks must parse (zero skip-validate anywhere; 114/114 valid 2026-09-29)
```

**md-go-validator gotcha:** the system binary (`/run/current-system/sw/bin/md-go-validator`) can lag the source in `~/projects/md-go-validator` until the next NixOS rebuild — run `go run ./cmd/md-go-validator` from that repo for current behavior (prose directive mentions no longer poison blocks; `//nolint` is no longer an implicit skip).

## Release Runbook

1. **Push tags SEPARATELY from the branch** (`git push origin master && git push origin vX.Y.Z`). A combined push can silently drop the tag webhook (v0.10.2 incident) and release.yml never fires.
2. **Verify the Release run starts within ~2 min of the tag push** (`gh run list --workflow=release.yml`). If missing, dispatch manually: `gh workflow run release.yml -f tag=vX.Y.Z` (workflow_dispatch fallback).
3. Sub-module tags go up before the root tag; verify proxy indexing with `go list -m -versions` before announcing. Submodules pin the PREVIOUS root version at tag time (chicken-and-egg); root pins ride the post-release Dependabot PRs.
4. **Submodule tags get no Release run** — `release.yml` triggers on root `v*` tags only. No GitHub Release for `diagnose/vX.Y.Z` etc.; expected, don't dispatch for them.
5. **The proxy version-list cache lags the tag** (minutes to hours); exact-version resolution works immediately. Verify with `go list -m <module>@vX.Y.Z` through the default proxy — don't poll `@v/list`.

## Publishing Rules

- Anything published under Lars's identity (Discussions, Reddit, release marketing) needs an explicit channel + consent decision, even inside a "do everything" directive. In-repo decisions with recorded defaults stay autonomous (Discussion #12 withdrawal, 2026-09-29).
- Before publishing, render title + body exactly as the target surface will show them and read it once as a stranger.
- Announcements never enter `CHANGELOG.md` or release notes — an announcement is not a change to the artifact.

## Docs Layout

Living docs own all current state. Historical snapshots: `docs/status/archived/` (inline `~~strikethrough~~` resolutions), `docs/planning/archived/`, `docs/feedback/archived/`, `docs/history/`. New status reports go in `docs/status/`; when fully dispositioned, strike inline and `git mv` to `archived/`. Never cite archived files without the `archived/` path prefix.

## Architecture Decision: Libraries Classify, Applications Enrich

**go-error-family (classification) and samber/oops (enrichment) are complementary, not competing.** The `bridge/` package is the seam where they meet.

- **LIBRARY code** (clients, SDKs, domain packages) imports `go-error-family` only and returns classified errors. A library knows its own domain contract (404 = Rejection, timeout = Transient) but must NOT presume the application's observability stack — so it never imports oops.
- **APPLICATION code** imports oops for enrichment (stack traces, trace IDs, request context) and, if it also needs behavioral decisions, wraps library errors via the bridge.

The classification protocol is the **six interfaces** (`Coded`/`Classified`/`Contextual`/`Retryable`/`ExitCoder`/`HTTPStatuser`) — the sole public contract. `Error` is a reference implementation, not the contract; domain types implement only the interfaces they need.

## Surprising Behaviors

- **`Classify(nil)` returns `Rejection`**, not a zero value. Intentional: nil error = caller's fault.
- **`Classify` defaults unknown errors to `Transient`** (retryable). Fail-open design — unknown errors get retried. Same for `ParseFamily` with unrecognized strings.
- **`errors.Is` matches on `code + family` only**, ignoring message. Two `*Error`s with different messages but same code and family will match.
- **`Wrap(nil, ...)` returns `nil`** — nil-safe, but means you can't construct an error wrapping nil.
- **`WithContext`/`WithCause`/`WithTimestamp`/`WithExitCode` are copy-on-write** — they return a NEW `*Error`, not the same pointer. Safe to chain from shared/sentinel errors. Do NOT assume identity preservation.
- **Template placeholders use `{key}`, not `{{.key}}`** — the old syntax collided with Go's `text/template`.
- **Consumer interfaces (`Coded`, `Classified`, `Contextual`, `Retryable`, `ExitCoder`, `HTTPStatuser`) embed `error`** — required for Go 1.26's `errors.AsType[T]()`. Don't remove the embedding.
- **`HandleErrorWithContext` is the canonical entry point** — `HandleError` and `HandleErrorWithConfig` delegate to it.
- **Package-level `Classify`/`RegisterClassification`/`RegisterTemplate` delegate to `DefaultRegistry`** — for test isolation or scoped handling, construct a `NewRegistry()` and pass it via `HandleConfig.Registry`.
- **`CommandRunner` defaults to `DefaultCommandRunner{}`** — rules with a nil `Runner` field use the real system commands. Tests inject mocks.
- **`Error()`/`Summary()` use `safeCauseString` for panic recovery** — a wrapped cause whose `Error()` panics is omitted from the message instead of crashing the process.
- **`ExitCode(err)` checks `ExitCoder` before family** — `*Error` returns 0 (meaning "use family default") unless `WithExitCode` was called.
- **`WrapOnce` is idempotent** — returns the existing `*Error` unchanged if the chain already contains one.
- **`Orchestration` is the 6th family** — internal coordination failures, severity 5, exit 70, HTTP 500. `Corruption` severity is 6. Order: Transient(1)<Rejection(2)<Conflict(3)<Infrastructure(4)<Orchestration(5)<Corruption(6).
- **Conditional requests (issue #5):** 304 is a success path (write it, return `nil`; NEVER classify). 412 = `NewConflict(...).WithHTTPStatus(412)`. 428 = `NewRejection(...).WithHTTPStatus(428)`. 416 = `Rejection` + `WithHTTPStatus(416)`. Guarded by `Example_conditionalRequests`.
- **Embedding `*Error` in a custom wrapper struct does NOT compile** — the embedded field `Error` shadows the promoted `Error() string` method, and a declared `Error()` method collides with the field. Use `WithHTTPStatus`/`WithExitCode` or a named field + explicit `Error`/`Unwrap` forwarding.
- **Submodule tags serve the SUBDIRECTORY go.mod via the proxy** — auditing a submodule tag means reading `<subdir>/go.mod` AT the tag, never the repo-root go.mod (misreading this briefly "proved" a poisoned `bridge/v0.2.0` that was fine).
- **go.work.sum can hold stale checksums because `GOPRIVATE` skips sumdb** — fix: delete the stale line and rebuild (workspace `use` re-resolves locally).
- **bridge/examples fail to build between pin-bump and tag-push** — their external imports (oops) force loading unpublished sibling go.mods → `unknown revision`. Expected mid-release state; verify after tags land. `go test ./...` from root does NOT span module dirs — use per-module invocations.

## API Surface (no API changes since v0.10.0)

**Family adapters** (`family.go`/`retry.go`, single-source via `familyData`): `Severity()` (total order above), `HTTPStatus()` (Rejection→400, Conflict→409, Transient/Infrastructure→503, Corruption/Orchestration→500), `RetryPolicy()` (advisory: Transient 3 attempts 100ms-5s, others single attempt — library does not run the loop). `familyInfo` includes `Audience` — adding a Family requires only one `familyData` entry.

**Error methods** (`error.go`): `WithContextMap`, `WithContextf`, `WithContextAny` (type-switched to string), `WithExitCode`, `WithHTTPStatus`, `ExitCode()`, `HTTPStatus()`, `JSON()` (canonical `{family,code,message,context,retryable,timestamp}`). All `With*` copy-on-write.

**Constructors** (`constructors.go`): `WrapOnce`/`WrapOncef` (idempotent), `Wrap{Family}f` variants (nil-safe).

**Registry** (`registry.go`): `Clone()` (deep-copy), `RegisterTemplates(map)`, `TemplateForCode(code)` (registry→builtin), `RegisterClassifier(s)` (predicate classifiers for dynamic errors, atomic copy-on-write; no `UnregisterClassifier` — Go funcs aren't comparable), `RegisterClassificationType[T]`/`...For[T]` (generic sugar; top-level functions because Go forbids type params on methods).

**Stdlib taxonomy** (`stdlib.go`): `RegisterStdlibDefaults(reg)` — context/sql/os mappings with documented rationale (DeadlineExceeded→Transient, Canceled→Rejection).

**Boundary helpers:** `Code(err)` (public code extraction), `HTTPStatus(err)`/`HTTPHandler(fn)` (checks `HTTPStatuser` override first, then family; **never leaks `err.Error()`** — message only from a registered template), `LogError`/`LogErrorContext` (Transient→Warn, others→Error; logs family/code/retryable/exit_code/context.*), `HandleConfig.Logger` hook (same record in the `HandleError` call), `errorfamilytest` subpackage (all 7 assert helpers, each with a godoc example).

## Classification Precedence

`Classify(err)` checks in order — first match wins:

1. **Multi-error** (`errors.Join`) → classify each sub-error, pick the **worst by severity** (deterministic regardless of order; fail-closed: any non-Transient sub → non-Transient result)
2. `Classified` interface → `ErrorFamily()`
3. `Retryable` interface → infer `Transient` (true) or `Rejection` (false)
4. Registered sentinels via `errors.Is` chain walk (atomic.Pointer to immutable map — lock-free)
5. Registered classifiers — in registration order, first `ok=true` wins
6. Default → `Transient`

A type implementing both `Classified` and `Retryable` uses `Classified`. Registering a sentinel for an already-`Classified` error has no effect. Classifiers only run when all earlier steps miss.

## Registry Pattern

Injectable `Registry` (`registry.go`); zero value unusable — `NewRegistry()`. `DefaultRegistry` backs all package-level convenience functions. Custom registries = test isolation / scoped handling. `Registry.sentinels` is `atomic.Pointer[sentinelMap]` to an immutable snapshot (~285 ns/0 allocs at 50 sentinels). `resolveTemplate(code, cfg, reg)` is the single template path (override → registry → builtin); templates are cohesive What/Why/Fix units.

## Agent Is Analysis-Only

`DebugAgent` has one method: `Analyze` (root cause + `FixStep` suggestions). The library does NOT execute fixes. `Involvement`/`RiskLevel` belong to the consumer.

## Diagnostic Rule Pattern

Use the matching helpers (`HasContextKey`, `ContextValue`, `ResolveContextKey`, `HasContextSubstring`, `FamilyIs`, `ErrorCodeContains`) and execution helpers (`RunCommand`, `CommandExists`). Rules run concurrently via `Runner.Run`; results sort by confidence descending. `DiagnosticResult.Fix = {Summary, Command}` — exact shell command, no prose parsing. `GitRule`/`PostgresRule` live in their own submodules; `DefaultRunner()` includes only zero-dep rules. New rules should follow the runnable `ExampleRuleSpec` in `diagnose/example_test.go`. `ContextKey` type replaces raw strings in specs; `NetworkRule.Run` returns `StatusUnknown` when no host is found.

## Partial Success

Not a library type — a consumption pattern. Recipe in SKILL.md (collect outcomes, `Classify` each failure, worst family for exit code).

## Test Coverage

FEATURES.md owns the coverage table (single source — an AGENTS copy of the table drifted from it once); all 7 modules ≥ 89% on the latest live `-race -cover` sweep. 16 fuzz targets (root 11, bridge 5). `errorfamilytest` is intentionally thin.

## Adoption Reality (audited 2026-07-23)

Root package: 50+ consumers. `New`/`Wrap` ~750 · `Classify` ~130 · `IsRetryable` ~40 · `HandleError*` ~35 · `RegisterClassification`/`Template` ~27 · `ExitCode` ~24. Under-adopted: `LogError` ~3, `HTTPHandler` ~5, `errorfamilytest` ~3, `diagnose` ~3. Bridge: ZERO external consumers — near-zero oops ecosystem adoption; consumers skip enrichment (classify→handle, not classify→enrich→handle). Reference implementation: `examples/cmd/bridge/` + `examples/checkout/` (three patterns + decision guide).

## Bridge Submodule (`bridge/`)

Separate module (depends on both libraries); root stays zero-dep. Correct, tested (94.4%), fuzzed.

| API                        | Purpose                                                                               |
| -------------------------- | ------------------------------------------------------------------------------------- |
| `bridge.Wrap(err, family)` | Attach a Family to any error, preserving OopsError context                            |
| `bridge.AutoWrap(err)`     | Infer Family from oops metadata (tags + domain), then wrap                            |
| `bridge.InferFamily(err)`  | Derive Family from oops tags (explicit) → domain (structural) → Transient (fail-open) |
| `ClassifiedError`          | Embeds `oops.OopsError`; satisfies `Classified`, `Coded`, `Retryable`, `Contextual`   |

Tag overrides (first): `retryable`, `transient`, `conflict`, `corruption`/`corrupted`, `rejection`/`rejected`, `infrastructure`/`infra`. Domain defaults (second): `validation`/`auth`→Rejection, `database`/`network`/`cache`/`queue`→Transient, `storage`/`infra`/`startup`→Infrastructure, `data`/`schema`/`migration`→Corruption. Surprising: `Wrap(nil, family)` returns a zero-OopsError ClassifiedError (`Error()` = `[family]`, `Unwrap()` = nil) — nil is still classifiable. Future `ClassifiedError` methods must handle the zero-OopsError case.

## Lint Configuration

- `//nolint:gochecknoglobals` on each legitimate package-level var (registries, immutable lookup tables, rule specs) — BuildFlow's pre-commit hook re-enables the linter if disabled in config.
- G304 excluded for `diagnose/rules_filesystem.go` via `.golangci.yml` path exclusion. Do NOT use inline `//nolint:gosec` there — inline directives break when golines wraps lines.
- `exhaustruct_v5` is enabled; most project types excluded (intentional optional fields) + test files. `Registry` excluded (`mu` zero value is correct).
- `mnd` ignores `family.go` (intentional status/exit/severity literals with comments). `varnamelen` allows `tc`, `f`, `w`, `ag`. Test files exclude `err113`/`testpackage`/`fatcontext`/`funlen`/`containedctx`/`cyclop`/`gocyclo`/`gocognit`/`maintidx`.
- depguard denies `encoding/json/v2` repo-wide (lax deny-only) — the lint canary for the no-GOEXPERIMENT policy.
- art-dupl: `.art-dupl-baseline.json` (committed, `-t 1`) holds the 3 accepted clone groups; `art-dupl check -t 1 .` reports only NEW clones. The baseline serves the plain-mode gate ONLY — `check` takes no mode flags (rejects `--suggest-generics`, exit 1; exit codes: 0 green, 1 usage error, 2 new clones; gate proven both ways 2026-09-30). Routine gate `-t 5`; deep sweeps `-t 1` are exploratory (`--suggest-generics` welcome, drop `--type-aware` — overridden), their hashes are NOT baseline-comparable. Accepted groups (do NOT fix): (1) `strTrue`/`strFalse` per-package const blocks — unexported, can't cross module boundaries; (2) the parallel `fmt.Formatter` methods (`error.go` ↔ `bridge/bridge.go`) — distinct types across modules with substantively different verbose branches; generic extraction would need a dependency the zero-dep root can't have; (3) the `host/port := r.resolveHost/Port(err)` pairs (`diagnose/rules_network.go` ↔ `diagnose/postgres/rules_postgres.go`) — 2-line idiomatic per-rule setup; cross-module extraction costs more than it saves.
- golangci-lint pinned `v2.13.2` everywhere (ci.yml + release.yml) — keep pins in sync with the local binary; do NOT re-add `//nolint:recvcheck` (v2.13.2 doesn't fire it).
- Go directives are per-module TRUE DEPENDENCY FLOORS (v0.10.2) — root/diagnose/agent/git/postgres at `go 1.26`, bridge/examples at `go 1.26.0` (dep-forced by x/text v0.42.0), go.work at `go 1.26.7` (toolchain, see its comment). Do NOT normalize to a single value.
- BuildFlow: 0 failed steps (no-cache verified); skips: `go-auto-upgrade` (jsonv1tov2 migrator), `pnpm-audit` (no lockfile at root — upstream candidate), `go-structure-linter` (embedded snapshot ignores project config; run the CLI directly), `branching-flow` (phantom analyzer ignores its own IsIgnored), `nix-hash-fix` (no vendorHash in flake). `checks.format` (treefmt: gofumpt/goimports/golines/nixfmt) green — run `nix fmt` before claiming golines-clean.

## Known Limitations

- **json/v2 guards:** the fleet `go-auto-upgrade` jsonv1tov2 migrator re-introduces `encoding/json/v2` whenever the machine-global `GOEXPERIMENT=jsonv2` is set (required by the BuildFlow repo). Guards: `.buildflow.yml` skip + depguard deny + CI `GOWORK=off` build. **CI-parity reflex: `GOEXPERIMENT= go build ./...` before claiming green — the shell env lies on this machine.**
- **`.buildflow.yml` skips are upstream candidates:** pnpm-audit needs subdirectory lockfile discovery; branching-flow needs `IsIgnored` wired into `pkg/phantom`. File upstream with repros (verify-before-filing gate). The `flat` structure preset in `.go-structure-linter.yaml` is deliberate: the root package IS the published API.
- **Website build (pnpm 11):** build-script approvals live in `website/pnpm-workspace.yaml` (`allowBuilds: esbuild: true`) — pnpm 11 ignores `pnpm.*` in package.json; a placeholder value silently disables the whole key (cmdguard incident). `minimumReleaseAgeStrict: true` set 2026-09-28 (fail-loud supply-chain cooldown; pnpm 11 already defaults a 1-day `minimumReleaseAge`).
- **Website Dependabot security updates fail on the pnpm lockfile — leave enabled.** No per-directory security-update toggle exists (verified 2026-09-28; `open-pull-requests-limit: 0` does NOT disable them). Jobs only fire when an alert exists; treat a red security job as the ALERT signal and remediate via `nix develop -c pnpm audit` in `website/` + `pnpm update --depth Infinity`.
- **`applyContext` substitutes `{key}` via `strings.ReplaceAll` without HTML escaping** — safe for CLI stderr; escape before embedding in HTML.
- **Examples are a separate module** (requires root + diagnose) — keeps the root truly zero-dep; CI builds via `working-directory: ./examples`.
- **Erraudit zero findings (2026-09-28):** `//nolint:legacyerrors // <reason>` is the suppression syntax for deliberate unpropagatable writes; it does NOT trip nolintlint and survives golines wrapping.
- **Website:** Astro 7 + Starlight + Tailwind v4; Firebase Hosting `errorfamily` in `lars-software`; domain `errorfamily.lars.software`; deploy via `website-deploy.yml` (secret `FIREBASE_SERVICE_ACCOUNT_LARS_SOFTWARE`, pinned `FirebaseExtended/action-hosting-deploy@500ac625`); manual `nix run .#deploy` from `website/`. NOT part of the Go workspace (own `flake.nix`; build = `nix run .#build`, no `packages.default`).
- **Website guard canary (`website-check.yml`):** fails on typescript major ≠ 6, frozen-lockfile drift, `astro check`, or `astro build` on every `website/**` push/PR — ends the three-time TS-7 deploy-breaker class structurally.
- **Website `typescript` must stay on 6.x:** `astro check` needs the programmatic API the TS 7+ native compiler drops (withastro/roadmap#1321). **Lockfile-sync rule: after ANY `website/package.json` change, run `nix develop -c pnpm install` and commit the regenerated lockfile in the same change.** Run website commands via `nix develop -c pnpm ...` — the system `node` is a Bun shim. An untracked `website/bun.lock` may appear; never commit it.
