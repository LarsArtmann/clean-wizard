# Status Report: go-business-rules Integration + Brutal Self-Review

**Date:** 2026-09-23 14:49
**Session scope:** Analyze `/home/lars/projects/go-business-rules` → adopt what helps `clean-wizard`
**Verification state at writing:** `go build ./...` green · 35 test packages pass (`-short`) · golangci-lint 0 findings on changed files · `nix fmt` clean

---

## Session Summary (what this report covers)

Adopted `github.com/LarsArtmann/go-business-rules/v2@v2.2.0` (Lars's own severity-aware
validation library) as the engine behind clean-wizard config validation. Rewrote all five
validation levels as declarative rules, added a 4th severity level (`critical`), fixed a real
UX bug (warnings silently dropped at config load), preserved JSON output shape, added 7 tests,
and documented everything.

---

## a) FULLY DONE

| # | Work | Evidence |
|---|------|----------|
| A1 | Library research: read go-business-rules API surface locally (rule.go, errors.go, validation_result.go, validator.go, README, FEATURES) — no guessing, verified exact signatures incl. the `NewViolationFromError` string-only context limitation that shaped the bridge design | session log |
| A2 | Dependency added: `go-business-rules/v2 v2.2.0` (+ transitive `go-finding v1.13.0`) via nix devShell (go 1.27 — host go 1.26 + GOTOOLCHAIN=local blocks it, documented gotcha); `go mod tidy` promotes it to a direct require | `go.mod:10` |
| A3 | `SeverityCritical` ("critical") added to `operations.ValidationSeverity`; alias + const in config | `operation_validation.go:17`, `validator_rules.go` |
| A4 | All five validation levels (structure / field / cross-field / business logic / security) expressed as tagged `businessrules.Rule`s with per-level tags `[level, kind]`; `kind` preserves the old rule strings (`required`/`range`/`format`/`security`/…) | `internal/config/validation_rules.go` |
| A5 | Bridge `mapViolations`: Critical+Error → `Errors` bucket, Warning+Info → `Warnings` bucket; `IsValid` semantics preserved; suggestion rides `WithDescription`, kind surfaces in `ValidationError.Rule` — JSON shape unchanged | `validation_rules.go:82` |
| A6 | UX bug fix: validation warnings are now logged at the load boundary (`logger.Warn`, field/message/suggestion); previously computed and silently discarded. Error suggestions logged too | `config.go:147-152` |
| A7 | Deleted `validator_structure.go`, `validator_crossfield.go` (`git rm`, history preserved); trimmed `addCriticalRiskError`/`validateBusinessLogic`/`validateSecurityConstraints` while keeping conflict-check predicates as rule check functions | git history |
| A8 | `MinProtectedPaths` — declared-but-unenforced constraint now enforced (guarded to not double-fire on empty paths) | `validation_rules.go:214` |
| A9 | Deterministic violation order: sorted profile iteration (previously map-iteration random) + sequential rule execution | `sortedProfileNames` |
| A10 | Intentional escalation: `..` parent-directory reference in a protected path is now `severity: "critical"` (was plain error) — the flagship demonstration of severity-aware validation | tested |
| A11 | 7 new tests: severity bucketing matrix, critical escalation, structured metadata preservation (Value/Context/bounds), warning bridging, deterministic order, safe-mode context metadata, const guard | `validation_rules_test.go` |
| A12 | Lint cleanliness: 0 golangci-lint findings on changed files (fixed 5× exhaustruct via project-convention `nolint`, 1× goconst, 2× ireturn, 1× ST1005 by lowercasing + switching `%v`→`%w` for settings errors) | buildflow |
| A13 | Docs: assessment + adoption doc, AGENTS.md dependency entry + "Config Validation Engine" section, TODO_LIST #30/#31, PACKAGE_BOUNDARY.md import row updated (was stale: viper/pkg/errors) | 4 files |
| A14 | Non-goals documented honestly: Stream/otel/composites/domain-layer rewrite deliberately NOT adopted with reasons | planning doc |

## b) PARTIALLY DONE

| # | Item | What's missing |
|---|------|----------------|
| B1 | **Warning surfacing is implemented but untested** — no test drives `validateLoadedConfig` through a warning-producing config and asserts the `logger.Warn` output; logger is a package global (known issue #11), which makes it awkward. The *code* is done; the *verification* is not | logging test |
| B2 | **License verification of the new dependency** — buildflow's go-licenses step failed in the first (out-of-shell) run; I never re-confirmed it passed inside the devShell run (tail cut the summary). `go-finding` license = unverified claim | re-run `buildflow -s license-check` |
| B3 | **Performance comparison** — benchmarks pass (25µs/op ValidateConfig) but I have no before/after baseline; the rewrite allocates a rule set + maps per call. "Behavior compatible" was proven; "performance compatible" was not measured | old-vs-new bench |
| B4 | **Consolidation opportunity only partially taken** — `ValidationError` vs `ValidationWarning` remain two near-identical shapes (pre-existing split brain); I mapped violations onto them instead of unifying. Info-level violations land in Warnings with **no severity marker** (`ValidationWarning` has no severity field) — info vs warning is indistinguishable downstream | severity field or unification |
| B5 | **Docs freshness** — CHANGELOG.md and FEATURES.md were NOT updated for the behavior changes (critical severity now emitted, warnings now logged); only AGENTS.md/planning/TODO/PACKAGE_BOUNDARY got updates | CHANGELOG entry |
| B6 | **gopls stale-diagnostics hygiene** — the editor kept reporting "go-business-rules not in go.mod" (stale: LSP runs host go 1.26); I diagnosed it but never ran `lsp_restart` to confirm and clear the noise | LSP restart |

## c) NOT STARTED

| # | Item | Why it's listed |
|---|------|-----------------|
| C1 | `ValidationMiddleware.validateChangeBusinessRules` if-chains → tagged rules (TODO #30) | registered, not started |
| C2 | `RuleEvaluated` events → per-rule timing in `-v` verbose output (TODO #31) | registered, not started |
| C3 | `operations`-level `ValidateSettings` rewrite on the rules engine — blocked by the stdlib-only domain boundary; would require an architecture decision first | TODO #30's bigger sibling |
| C4 | Full (non-`-short`) test suite run this session | AGENTS.md documents `-short` as the norm; not run, not claimed |
| C5 | `schemas/config.schema.json` review — whether it enumerates validation output severities and needs `"critical"` added | unknown, unchecked |
| C6 | `PathPattern`, `MaxRiskLevel`, `BackupRequired` fields in `ConfigValidationRules` are declared-but-dead (pre-existing ghost config) — left in place | pre-existing |

## d) TOTALLY FUCKED UP!

| # | Item | Detail |
|---|------|--------|
| D1 | **REAL BUG I INTRODUCED: rule-name collisions corrupt bridged Value/Context metadata.** `configRuleSet.values` / `.contexts` are maps keyed by rule *name*, but six rules share the name `"protected"` (required, min_items, format, unique, root-path warning, parent-reference critical). `add()` overwrites per name, and security rules are added LAST — so `values["protected"]` ends up holding a single path string (not the array) and `contexts["protected"]` holds the last security rule's metadata. Any non-security "protected" violation then bridges with the WRONG Value and WRONG Context. Found during this self-review, NOT by my tests — `TestValidateConfig_BridgePreservesStructuredMetadata` uses max_disk_usage (unique name) and no test probes Value/Context on colliding names. **This is exactly the "tested what I changed, not what I could break" failure.** | `validation_rules.go:52-56` |
| D2 | **First buildflow format run executed outside the nix devShell** and failed 4 Go steps on the documented GOTOOLCHAIN gotcha — wasted a full run (~4 min) because I didn't check the env precondition before delegating; AGENTS.md documents this exact trap | session log |
| D3 | **3 failed edits during the session** (stale context after buildflow's auto-formatter rewrote my file mid-edit; one structurally impossible interface assertion in `withSuggestion` that I caught only by reasoning about return-type matching) — each recovered, but the interface-assertion bug would have silently dropped ALL suggestions had I not caught it pre-build | session log |
| D4 | **TODO_LIST numbering inserted out of order** (#30/#31 placed before #29) — cosmetic, but sloppy | TODO_LIST.md |

## e) WHAT WE SHOULD IMPROVE!

1. **Test what the bridge drops, not just what it keeps.** The D1 bug class = property tests: "for every violation, the bridged Value/Context matches the rule that produced it." A table-driven collision test would have caught it.
2. **Bridge keys must be collision-proof.** Composite key (name+level+kind) or unique rule names — decided with Lars (see question 3).
3. **Read env preconditions before delegating to orchestrators.** `buildflow` requires the devShell here; one `git grep GOTOOLCHAIN AGENTS.md` would have saved a run.
4. **Benchmark before/after for engine swaps.** "Refactors must not regress performance" is a claim, and claims need baselines.
5. **Restart LSP after go.mod toolchain changes** instead of leaving stale error diagnostics polluting every tool result.
6. **Update CHANGELOG.md in the same change** when user-observable behavior shifts (severity labels, warning logging) — not "later".
7. **Logging paths deserve tests even when the logger is a global** — inject a capture logger or assert via slog handler; "implemented" ≠ "verified".
8. **The `requiredCheck(capturedBool)` pattern is inconsistent** with the closure-reads-cfg pattern used elsewhere in the same file — pick one lifecycle model.
9. **`ValidateField` still severity-blind** (returns bare `error`) — either deprecate it or make it return the bridge result; it's the last severity-losing API in the package.
10. **GOEXPERIMENT/toolchain gotchas keep burning minutes across sessions** — a `direnv`/`nix develop` auto-entering setup (or buildflow doctor gate) would kill this class of failure permanently.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

*Sorted by impact×effort within tiers. Items 1-8 are direct session follow-ups; 9+ are project backlog noticed this session.*

**Tier 1 — Fix & harden this session's work (do first)**

| # | Task |
|---|------|
| 1 | **Fix D1 collision bug**: collision-proof bridge keys (or unique rule names) + regression test proving protected-path violations carry their own Value/Context |
| 2 | Add a table-driven "every rule's metadata survives the bridge" test over ALL rule names (collision matrix) |
| 3 | Test the warning-surfacing path in `validateLoadedConfig` (capture logger or slog handler) |
| 4 | Re-run `buildflow -s license-check` in the devShell to verify go-finding/go-business-rules licenses |
| 5 | Add CHANGELOG.md entry: critical severity introduced, warnings now logged, `..` escalation |
| 6 | Review `schemas/config.schema.json` for a severity enum needing `"critical"` |
| 7 | Benchmark old-vs-new ValidateConfig (checkout prior commit, same bench) and record a baseline in TODO or a bench note |
| 8 | Restart gopls / verify stale "not in go.mod" diagnostics clear; if not, note as toolchain issue |

**Tier 2 — Complete the adoption**

| # | Task |
|---|------|
| 9 | Migrate `validateChangeBusinessRules` middleware if-chains to tagged rules (TODO #30) |
| 10 | Wire `RuleEvaluated` events into `-v` verbose output for per-rule timing (TODO #31) |
| 11 | Add severity to `ValidationWarning` (or unify with ValidationError) so Info ≠ Warning downstream |
| 12 | Decide + execute: `ValidateField` returns bridge result (severity-aware) or gets deprecated |
| 13 | Unify the rule-construction lifecycle: all checks read cfg live in closures (kill `requiredCheck` bool capture) |
| 14 | Architecture decision: move `operations.ValidateSettings` validation onto the rules engine — decide whether domain stays stdlib-only (validation logic migrates to a validation layer above domain) |
| 15 | Enforce `MaxRiskLevel`/`BackupRequired` constraints from `ConfigValidationRules` (currently dead) |
| 16 | Implement or delete `PathPattern` in `ConfigValidationRules` (declared, never read — ghost config) |
| 17 | Consider `businessrules.When` composites where conditionals read naturally (safe-mode rules) |
| 18 | Emit validation results as structured JSON in `--json` mode with the new severity field |

**Tier 3 — Project backlog (noticed, pre-existing)**

| # | Task |
|---|------|
| 19 | Logger globals (`L`, `StdLogger`) → DI-injected (TODO #11; blocks B1-style tests everywhere) |
| 20 | `scan --profile` filtering or remove the flag (TODO #10) |
| 21 | Split >350-line files: compiledbinaries (585), docker (524), nodepackages (523) (TODO #12) |
| 22 | CLI command tests: profile, config, scan, init (TODO #13) |
| 23 | BDD tests for remaining cleaners (7 of 13 lack them; TODO #8 follow-up) |
| 24 | Fix `nix flake check` treefmt-check go1.27 toolchain download (TODO #29) |
| 25 | Wire go-humanize-linter into flake checks (TODO #18) |
| 26 | Move `/tmp/go-humanize-linter` into repo `tools/lint/` (TODO #17) |
| 27 | Pre-existing release.yml finding: pin sbom-action by SHA (the one buildflow gate error) |
| 28 | Improve Nix size estimation, hardcoded 50MB/generation (TODO #15) |
| 29 | `getRegistryName` reverse-lookup tests (TODO #16) |
| 30 | Settings only flow with `--profile`; preset/interactive runs use factory defaults (Known Issues) |
| 31 | `NixGenerationsSettings.DryRun`/`Optimize` + `BuildCacheSettings.ToolTypes` have no constructor consumption (Known Issues) |
| 32 | Gherkin `.feature` files for top cleaners (TODO #20) |
| 33 | Standardize BDD test file naming (TODO #21) |
| 34 | `--dry-run` for scan command (TODO #22) |
| 35 | `--keep-generations` flag for Nix cleaner (TODO #23) |
| 36 | `parseSize("garbage")` wrapped-error regression test (TODO #26) |
| 37 | Remove `infertypeargs` warnings, 15+ explicit type params (TODO #19) |
| 38 | Inline-or-delete single-callsite `ParseNumberAndUnit` (TODO #25) |
| 39 | mdx/website docs: document the new warning logging + critical severity in configuration guide |
| 40 | Consider surfacing validation warnings in `config` CLI command output (not just load-time logs) |

**Tier 4 — Larger bets (ROADMAP fuel)**

| # | Task |
|---|------|
| 41 | Severity → errorfamily.Family mapping policy (should Critical map differently than Error at exit-code layer?) |
| 42 | otel listener adoption if/when clean-wizard grows telemetry |
| 43 | Config validation as user-facing `doctor` subcommand reusing the rule set |
| 44 | Rule set extensibility: user-supplied custom rules in YAML config |
| 45 | Property test: violations(Build) ≡ violations(bridge output) invariants (mirror go-business-rules' gopter test) |
| 46 | Deduplicate `ValidationResult` consumers: sanitizer/middleware/loader all append warnings ad hoc |
| 47 | Snapshot/golden tests for ValidationError JSON output across the rule matrix |
| 48 | Evaluate `govalid` structural layer (go-business-rules README pairing) for `OperationSettings` structs |
| 49 | CI: add a `buildflow --build-mode dev` job inside the devShell so gate failures surface pre-push |
| 50 | docs-health HARVEST: route items 1-18 into TODO_LIST.md with priorities; 41-50 into ROADMAP.md |

## g) QUESTIONS I CAN NOT FIGURE OUT MYSELF

1. **Domain boundary vs. rules engine:** `operations.ValidateSettings` (domain, stdlib-only by PACKAGE_BOUNDARY.md) still validates by hand. Do you want validation logic to *migrate out of domain* into a config-adjacent validation layer (breaking the "domain = stdlib only" rule), or should the domain keep hand-rolled validation permanently?
2. **Strict-mode semantics:** should `ValidationLevelStrictType` treat Warning-severity violations as blocking (errors)? Right now strict only adds extra checks; "strict" meaning warnings-block is undefined and only you can decide the product intent.
3. **Bridge collision fix — JSON compatibility:** fixing D1 either (a) keeps `Field` strings identical and keys metadata maps internally by a composite key, or (b) makes rule names unique like `protected.format`, which would change `ValidationError.Field` values in JSON output (a breaking change for any consumer matching on field names). Which do you want — internal-only fix (a) or cleaner rule names (b)?

---

**Override note:** status-report skill specifies styled HTML output; user explicitly requested `.md` at `docs/status/` — user instruction wins, no HTML dashboard this time.

**Per skill contract:** the auto-commit daemon picks this file up; no manual commit (Crush harness forbids commits without explicit user request). Section (f) items 1-18 are HARVEST candidates for TODO_LIST.md on the next docs-health run.

**WAITING FOR INSTRUCTIONS.**
