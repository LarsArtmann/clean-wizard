# Roadmap

Long-term direction and raw ideas not yet refined into actionable tasks.
When an idea becomes bounded and actionable, it moves to `TODO_LIST.md`.

**Last updated:** 2026-09-30 (bridge README showcase idea harvested from 2026-09-30_06-08_art-dupl-deep-sweep-triage.md)

---

## Direction

go-error-family is a stable classification core with a growing ecosystem
of opt-in modules (`diagnose`, `agent`, `bridge`). The taxonomy is proven across
multiple consumers (DiscordSync, browser-history, SwettySwipperWeb). v0.11.0 is
released (2026-09-29, discoverability release) on top of v0.10.3's retraction of
every proxy-broken pre-extraction tag; all 7 modules are proxy-indexed and the
pkg.go.dev example surface covers all six families, the boundary APIs, the test
helpers, and the `RuleSpec` diagnostic pattern (30 root examples). The focus
now is: converting that discoverability into adoption for the under-used
boundary handlers (`HTTPHandler`, `LogError`, `diagnose`) — the bridge
announcement (TODO_LIST #1) is the open lever — and keeping the supply-chain
gates honest (the BuildFlow `pnpm-audit` step is
skipped because it cannot see the `website/` subdirectory lockfile;
Dependabot's npm updater cannot drive the pnpm website lockfile — the working
check is a manual `pnpm audit` inside `website/`).

## Themes

### 1. Consumer Discoverability

The #1 cross-cutting theme from all three consumer feedback sessions: surprising
behaviors are documented in SKILL.md but NOT in godoc where consumers actually
look. The library is correct; the documentation surface needs to meet consumers
where they are.

**Raw ideas:**

- ~~Example functions (`ExampleClassify`, `ExampleWrap`, etc.) visible on pkg.go.dev~~ — **SHIPPED (2026-09-18):** 26 runnable examples covering all six families, the classification precedence (`Classify`, `IsRetryable`, `ParseFamily`), registry patterns, CLI handling (`HandleError`, `HandleErrorWithContext`, `HandleErrorDetailed`), and the HTTP boundary (`HTTPHandler`, `HTTPStatus`, `Error_JSON`, conditional requests). ~~Remaining gap: `errorfamilytest` and `diagnose` subpackages have no examples yet.~~ — **SHIPPED (v0.11.0):** 7 assertion-helper examples plus an executable `ExampleRuleSpec`. Remaining gap: `agent` and `bridge` have no examples.
- A "common patterns" section that grows from real consumer usage
- Consider whether `Code()` vs `ErrorCode()` dual accessors should converge in a future major version

### 2. HTTP Story Parity

The CLI story (`HandleError`) is universally praised (10/10). The HTTP story is
weaker (6/10). `HTTPHandler` and `HTTPStatus` exist but consumers still build
custom layers.

**Raw ideas:**

- ~~Per-error HTTP status override (`WithHTTPStatus`)~~ — **SHIPPED (v0.8.0)**, mirrors `ExitCoder`/`WithExitCode` pattern. Update: still under-adopted (~5 consumers for `HTTPHandler`); needs discoverability work.
- Error code in JSON responses is solved for `HTTPHandler` but not for consumers using their own HTTP layer
- Consider an `httperror` subpackage with richer response shaping
- OpenAPI/schema generation for the error JSON shape

### 3. Release Pipeline Hardening

The v0.6.0 phantom-`replace` incident exposed that `go.work` masks
consumer-facing bugs. CI needs to verify the module graph from the consumer's
perspective. The `GOWORK=off go list -m all` gate, consumer-simulation job,
and `go vet` all shipped in v0.8.0; the remaining gaps are tooling-level.

**Raw ideas:**

- Release automation script for coordinated multi-module tag cutting — reinforced 2026-09-27: the `Release` workflow silently did not fire on the v0.10.2 tag push (root-caused 2026-09-28: dropped tag-ref push webhook on a combined branch+tag push; `workflow_dispatch` fallback + runbook in place — `docs/status/archived/2026-09-28_02-10_release-yml-no-fire-root-cause-v0.10.2.md`); an automation script with explicit trigger verification would end this class. Status: the runbook + `workflow_dispatch` fallback shipped 2026-09-28 (live-tested on v0.10.3 and v0.11.0) — the script itself is the last piece of this class
- ~~Deprecation notes for broken tags (v0.6.0 family)~~ — **SHIPPED (v0.10.3, 2026-09-28)**: `retract [v0.5.0, v0.6.0]` in root go.mod; `diagnose/v0.1.0` and `agent/v0.1.0` retracted by the v0.2.5 releases. Full audit found exactly 5 broken tags of 60+; v0.6.1 verified clean and left available. Retraction verified via `go list -m -retracted` against the origin.
- Pin-bump hygiene: submodules should bump root pins in lockstep on releases

### 4. Ecosystem Growth

The library has 50+ consumers importing the root classification core, but
higher-level APIs are under-adopted: `LogError` (~3 consumers), `HTTPHandler`
(~5), `errorfamilytest` (~3), `diagnose` (~3). The bridge module has **zero**
external consumers despite being correct, tested (94.4%), and fuzzed.

The bridge gap is demand and demonstration, not quality. The root cause is
that `samber/oops` adoption is near-zero across the ecosystem, and consumers
skip the enrichment layer entirely (classify→handle, not classify→enrich→
handle). The reference implementation shipped (2026-07-26:
`examples/cmd/bridge/` + `examples/checkout/`) demonstrates the full
classify→enrich→handle flow with three patterns (pass-through, AutoWrap,
explicit Wrap) and documents when to use each. The remaining gap is **demand
and discoverability** — getting the pattern in front of consumers who already
use oops.

**Raw ideas:**

- ~~**Reference implementation for oops + bridge + error-family stack**~~ — **SHIPPED (2026-07-26):** `examples/cmd/bridge/` + `examples/checkout/`. Three patterns, 19 tests, pattern documentation in `cmd/bridge/README.md`. ~~Next: website guide page~~ — **guide page SHIPPED (v0.10.1):** `website/src/content/docs/guides/bridge.mdx`. Remaining: public announcement (TODO_LIST #1).
- A bridge reference-implementation showcase section in the main README (short pointer to `examples/cmd/bridge/` + the website guide; distinct from the pending public announcement, TODO_LIST #1) — source: 2026-09-30 report §f21
- More diagnostic submodules (`redis`, `docker`, `kubectl`)
- Bridge packages for other error enrichment libraries beyond oops (only after oops bridge has proven consumers)
- Integration guides for common frameworks (Echo, Gin, Chi) — ~~including a gRPC status-mapping guide (family → `codes.Internal`/`Unavailable`/`InvalidArgument`), a recurring consumer ask~~ — **gRPC guide SHIPPED (v0.11.0):** `guides/grpc` on the website (six-family table per Google's HTTP↔gRPC mapping + interceptor)
- ~~End-to-end `HTTPHandler` middleware example (net/http or Chi) and example symmetry for the `New{Family}` constructor set~~ — **RESOLVED (v0.11.0):** constructor symmetry complete (`ExampleNewConflict/Corruption/Infrastructure/Orchestration`); the middleware shape was already demonstrated end-to-end by `ExampleHTTPHandler` (v0.10.2) — a second one would duplicate it
- Compile-verify documentation code samples for new guides (the gRPC interceptor snippet is docs-only by design — extract to `examples/` or a build-tagged test before it rots)
- `examples/cmd/grpc` mirroring the gRPC guide; E2E httptest coverage for `examples/cmd/http`
- Benchmark suite comparing classification overhead across versions
- Type `DiagnosticResult.Details` values (kills the stringly `"true"`/`"false"` constants at the root of the accepted `strTrue`/`strFalse` clone) — breaking protocol change, next major only

## Open Questions

Unresolved decisions that gate work elsewhere. They are not tasks — see TODO_LIST for the actionable fallout.

1. ~~**art-dupl threshold policy**~~ — **RESOLVED (2026-09-28, recorded default):** routine gate `-t 5`; occasional deep sweeps `-t 1` against the committed `.art-dupl-baseline.json` (3 accepted clone groups). Policy lives in AGENTS.md "Lint Configuration"; wired via `art-dupl check -t 1 .` reporting only new clones. **Refined 2026-09-30 (gate-only decision):** the baseline serves the plain-mode gate ONLY — deep sweeps (`--suggest-generics` welcome) are exploratory and their hashes are NOT baseline-comparable; full policy in AGENTS.md Lint Configuration. CI enforces the baseline gate (`-t 1`) — decided 2026-09-30.
2. **Fleet churn: fix at the root or per-repo canaries forever?** The json/v2 re-imports, TS-7 bumps, and go-directive drift originate in the machine-global `GOEXPERIMENT=jsonv2` export plus the orchestrator's migrator — outside this repo. Is a fleet-level fix planned, or is the per-repo canary defense the accepted permanent architecture? (Decides whether the website canaries in TODO_LIST #4 are scaffolding or permanent.)
3. **Bridge & enrichment APIs: invest or freeze?** Bridge has zero external consumers (audited 2026-07-23; root cause: near-zero oops adoption ecosystem-wide); `LogError` ~3, `HTTPHandler` ~5, `errorfamilytest` ~3, `diagnose` ~3 external call sites. Grow adoption (announcements, examples, guides) or freeze the surface and invest in reliability/coverage? (Gates roughly a third of the discoverability backlog.)
4. **Post-tag edits to released artifacts:** may a released CHANGELOG section or GitHub Release body ever be edited after the tag? The v0.11.0 withdrawal needed one (dead-link removal — done and justified), but the deeper rule is that nothing that can expire enters a released section. Decide: dead-link-removal exemption vs strict never-after-tag (source: 2026-09-29 report §g3).
5. **GitHub Discussions:** disabled again after the withdrawn announcement. Record it as permanently off, or as "off until there's community traffic" with a revisit trigger? (Source: 2026-09-29 report §g2.)
