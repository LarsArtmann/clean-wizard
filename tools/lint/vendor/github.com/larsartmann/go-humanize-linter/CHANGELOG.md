# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Nothing yet.

### Changed

- Nothing yet.

### Fixed

- Nothing yet.

## [0.5.0] - 2026-10-23

### Changed

- **Walker skip rules: hidden directories (leading `.`), backup trees (trailing `.bak`), and `archived`/`forks` basenames are never scanned.** Frozen and upstream-owned copies (e.g. a `BuildFlow.vendor.bak/` vendored backup or a `.bf-jscpd-worktree/` duplicate) double-reported findings that no one can fix at the live source — 33 of 34 unfixable findings in the 2026-10-08 fleet cleanup came from such trees. Regression-guarded by `TestWalkGoDir_SkipsHiddenBackupArchivedForks`.
- **BREAKING (library API): `VerifySuppressions(dir)` and `VerifySuppressionsInFiles(fset, files)` no longer take a run report.** Staleness is now judged against a fresh unsuppressed re-detection instead of the caller's report — the detection pipeline drops suppressed findings before any report is built, so report-based matching flagged every working `//nolint:gohumanize` directive as stale (the findings it suppressed were absent by construction; even the repo's own `h001_suppressed` fixture was flagged). Re-detecting also makes the verdict rule-set independent: a directive stays valid when its rule is merely disabled in the current run. Regression-guarded by `TestVerifySuppressions_UsedSuppression` (now genuinely end-to-end) and the new `TestCLI_VerifySuppressions_UsedSuppressionNotFlagged`.
- **Go 1.27 is now the minimum toolchain** (go.mod `go 1.27`). The background "bumper" sweep raised the directive on 2026-09-19 (incident #4, commit `b7f00c2`, two hours after v0.4.0, against the recorded revert-on-sight decision) and later sweeps carried go-finding v1.10.0→v1.15.0, go-linter-sdk v0.3.1→v0.4.0, golang.org/x/tools v0.50.0→v0.52.0, and go-error-family v0.10.0→v0.12.0. Resolved by adoption (TODO_LIST T33 closed): the CI workflow, the release workflow, and the GitHub Action now pin `go-version: "1.27"`. Consumers on Go 1.26 must stay on v0.4.0.
- **golangci-lint custom-plugin base bumped v2.12.2 → v2.15.0** (`.custom-gcl.yml`, CI lint job). v2.12.2's bundled staticcheck (honnef.co/go/tools v0.7.0) IR builder panics on Go 1.27 stdlib (`internal/poll`: `unexpected expr: *ast.KeyValueExpr`) — hidden for a month behind the Setup-Go failure and surfacing the moment CI actually ran under Go 1.27. v2.15.0 ships staticcheck v0.8.2 and lints Go 1.27 code cleanly. The now-working plugin path also exposed three expected self-detections in test fixtures carrying inline H001 patterns (`TestCLI_MinConfidence`, `TestCLI_BehaviorDelta`, `TestVerifySuppressions_UsedSuppression`) — suppressed with scoped `//nolint:gohumanize:H001`, matching the existing convention.

### Fixed

- **`go-humanize-linter .` scanned nothing** (0 findings, exit 0): the walker's hidden-directory skip rule (`strings.HasPrefix(name, ".")`) matched the walk root's own basename `"."`, so `filepath.SkipDir` aborted the walk before a single file was visited. Every other root spelling (`internal`, `./internal/cli`, absolute paths) worked, which masked the bug — the most common CLI invocation was a silent no-op. The walk root is now exempt from all skip rules (the caller explicitly named it). Regression-guarded by `TestWalkGoDir_DotRootScansEverything` (covers `.` and `./`).
- **H0SUP no longer judges bare `//nolint:all` directives.** A blanket `all` usually exists for another linter's benefit (e.g. a golangci-lint autofix workaround), so calling it stale because gohumanize itself has no finding in that function was a false positive. Staleness verification now applies only to directives that name gohumanize or scope a rule ID; misspelled-name detection is unchanged. Regression: `TestVerifySuppressions_BareAllNotJudged` + `TestTargetsGohumanize`.
- **H012 suggestion pointed at a nonexistent root-package API**: the finding message, suggestion text, `--explain` entry, and rule docs said `humanize.WordSeries` / `humanize.OxfordWordSeries`, but those functions live in the `english` subpackage (`english.WordSeries`, `english.OxfordWordSeries`; available since go-humanize v1.0.1) — following the old advice verbatim does not compile. All linter output and docs now reference `english.`.
- **H009 false positive on CSV/round-trip float writers** (seen in go-health-dashboard's `/health/export`): `strconv.FormatFloat(v, 'g', -1, 64)` beside `WriteByte(',')` CSV field delimiters was diagnosed as manual-commaf. Root cause: source-level `-1` parses as `ast.UnaryExpr`, and `getBasicLit` strips the minus — turning the old `"!= -1"` exclusion into dead code (the old unit test fed a `BasicLit("-1")` shape the parser can never produce). Precision is now resolved sign-aware (`signedIntLiteral`), and literal verbs other than `'f'` are excluded, mirroring the Sprintf path's rejection of `%.Ng`. Regression-guarded by realistic-shape unit tests and the `h009_negative` CSV-writer fixture.
- Documentation health pass (2026-09-26): rebuilt TODO_LIST (open work only), refreshed ROADMAP (H012 shipped, plugin-index decline, per-statement ADR 0006 status, renamed the colliding output-warnings idea), refreshed FEATURES validation/coverage/interface rows, corrected the go-directive incident trail in AGENTS.md, and annotated + archived the point-in-time reports under `docs/status/` and `docs/planning/` (fully-resolved reports moved to `archived/`).

## [0.4.0] - 2026-09-19

### Added

- **H012 `manual-word-series` rule** (`pattern_word_series.go`, `rule_word_series.go`) — detects hand-rolled word-series joining ("a, b, and c") that should use `humanize.WordSeries` / `humanize.OxfordWordSeries`. Requires a comma `strings.Join` AND a conjunction in position (Join separator like `Join(x, " and ")` or a concatenation operand like `+ " and " +`); Full confidence needs the complete clone (prefix join of all but the last element + `x[len(x)-1]` access). Plain `strings.Join(x, ", ")` and prose-only conjunctions stay clean. Corpus-swept same day: 2 true positives, 0 false positives. H011 remains reserved for a future manual-si-parse rule (zero corpus demand so far).
- **H010 fixture pack** — aliased-import, dot-import, scoped `//nolint:gohumanize:H010`, and rune-filter-loop fixtures close the coverage gap with H007's alias testing.
- **H012 regression guards** — the conjunction-position filter (prose "and" does not count) is now protected twice over: expanded `h012_negative` fixtures (Sprintf-format prose, bare-variable prose, prefix join without conjunction) and a white-box unit test `TestCollectWordSeriesEvidence_ConjunctionPosition` that proves the discrimination in both directions (prose never counts, Join separator / concat operand always count).

### Changed

- **H002 suggestion enrichment** — when the flagged function references math/big, the suggestion now points at `humanize.BigComma` for `*big.Int` values (new `h002_bigint` fixture).
- **H009 suggestion enrichment** — mentions `humanize.CommafWithDigits(f, n)` for custom precision.
- **Library examples** — `doc.go` and README now show the ternary `ExitCodeByConfidence` exit code (ADR 0003) instead of the legacy binary `ExitCodeFromReport`, and README documents that upgrading to a release with new rules (like v0.3.0's H010) requires regenerating `--behavior-delta` baselines.
- **CI hygiene** — dependabot actions-group PR rebased and merged (main had been red since 2026-09-13 on a stale gogenfilter test; root cause documented in `docs/validation/2026-09-18_h010-sweep.md` era commits and fixed before v0.3.0). The five `DEPLOY_KEY_*` secrets were deleted after green CI.
- **Security hygiene** — the four now-unused read-only CI deploy keys were removed from the dependency repos' settings (go-linter-sdk, go-finding, gogenfilter, go-error-family), completing the going-public auth cleanup; all four repos verified at zero deploy keys.

### Fixed

- Nothing yet.

## [0.3.0] - 2026-09-18

### Added

- **H010 `manual-comma-parse` rule** (`pattern_comma_parse.go`, `rule_comma_parse.go`) — detects hand-rolled comma-grouped number parsing (`"1,234,567"` → number) that should use `humanize.ParseComma` / `humanize.ParseCommaf`, added in go-humanize v1.1.0. Fires only on corroborating signals: a comma-strip form (`strings.ReplaceAll/Replace(x, ",", "")`, `Split(",")+Join("")`, or a `','`-filter loop with string rebuild) AND a strconv number parse (`Atoi`/`ParseInt`/`ParseUint`/`ParseFloat`/`ParseComplex`). Confidence tiers: **Full** (strip inlined inside the strconv call), **High** (strip + parse in the same function), **Medium** (indirect strip + parse). Negative filters keep CSV field splitting, non-comma stripping, display-only cleanup, and comma-rejecting validators clean. Registered everywhere H001-H009 are (CLI, library, plugin, `--explain`, `--rules`).

### Changed

- **gogenfilter sqlc detection realignment** — gogenfilter v3.5.0 (pinned here since v3.6.0) made the bare sqlc default filenames (`models.go`, `querier.go`, `batch.go`, `db.go`, `copyfrom.go`) config-aware: they are no longer classified as generated from the filename alone, only via the `// Code generated by sqlc` content marker, a `sqlc.yaml`-declared output directory, or the `.sql.go` suffix. `TestIsGeneratedFile_FilenameOnly` now asserts the new semantics; this eliminates the previously documented hand-written-`batch.go` false-positive class.
- **Documentation corrections** — `doc.go`'s rule list was missing H008/H009 and pointed H004 at nonexistent `humanize.Plural` (correct: `english.Plural`/`english.PluralWord` from the `english` package); README H004 row had the same error. README now also notes `BytesN`/`IBytesN` (go-humanize v1.1.0+).

### Validation

- **H010 real-world sweep (2026-09-18)** — 169 repos under `~/projects/`: 1 borderline finding (locale-aware dual-format monetary parser; documented triage guidance), 0 clear false positives, 0 Full-tier hits. Full report: `docs/validation/2026-09-18_h010-sweep.md`.

## [0.2.0] - 2026-09-02

### Validation

- **Real-world validation sweep (2026-08-10)** — All 9 rules and all new features validated against 158 Go projects. Zero corpus findings (33 projects now use go-humanize). `--verify-suppressions` found 3 stale directives. `--min-confidence` filtering verified via synthetic fixtures. gogenfilter integration assessed at ~2.8% false-positive rate (no app code missed). Full report: `docs/validation/2026-08-10_real-world-sweep.md`.

### Added

- **gogenfilter integration** (`pattern_generated.go`) — Generated-file detection now delegates to `github.com/LarsArtmann/gogenfilter/v3` for two-phase detection (sqlc, templ, protobuf, deepcopy-gen, wire, moq, mockgen, mockery, easyjson, counterfeiter, plus generic `// Code generated by` fallback). A small legacy-suffix fallback (`_gen.go`, `.gen.go`, `_templ.go`) remains for generators gogenfilter does not enumerate. `walker.go` reads each file once and passes bytes to both the generated check and the parser.
- **`--min-confidence <level>` CLI flag** — Filter findings by confidence level (`low`, `medium`, `high`, `full`, default: `low`). Uses `finding.ByConfidenceAtLeast` from go-finding to produce a filtered report.
- **`--verify-suppressions` CLI flag** — Report `//nolint:gohumanize` directives that suppress zero findings, and `//nolint` directives that use a misspelled linter name containing "humanize" (e.g. `go-humanize-linter`). Uses pseudo-rule ID `H0SUP`.
- **Confidence-aware exit codes** — Exit 0 = clean, exit 1 = at least one high/full-confidence finding (must fix), exit 2 = only medium/low-confidence findings remain (triage). Lets CI distinguish "review" from "fix now".
- **H0SUP pseudo-rule** — Diagnostic rule ID for suppression-verification findings. Not registered in `AllRules()` (meta-diagnostic, not a code pattern).
- **Suppression verification system** (`suppression.go`) — `VerifySuppressions(dir, report)` detects stale and misspelled `//nolint` directives. Exported for CLI and future plugin use.
- **Plugin confidence filtering and suppression verification** — golangci-lint v2 module plugin now supports `minConfidence` and `verifySuppressions` settings in `.golangci.yml`. Findings are filtered by confidence before reporting. Plugin diagnostics use the finding's specific position (via `findingToTokenPos`) instead of the function declaration.
- **Shared `ParseConfidenceLevel`** (`confidence.go`) — Exported core-package function parsing confidence level strings. Used by both CLI and plugin paths.
- **`VerifySuppressionsInFiles`** (`suppression.go`) — Plugin-path variant of `VerifySuppressions` that collects directives from pre-parsed `pass.Files` instead of walking a directory.
- **`hasSwitchStatement()` helper** (`pattern_bytes.go`) — Reports whether a function contains any `switch` or type-switch statement. Used by the H001 size-bucket false-positive filter.
- **SDK-provided ternary exit codes** — CLI uses `linter.ExitCodeByConfidence(report, threshold)` from `go-linter-sdk` (the local `exitCodeFromReport()` was replaced by the upstream function, resolving TODO T21).
- `TestParseConfidenceLevel`, `TestExitCodeByConfidence`, `TestFilterReportByConfidence`, `TestCLI_MinConfidence`, `TestCLI_VerifySuppressions_UnknownLinterName`, `TestCLI_VerifySuppressions_StaleSuppression` — tests for confidence thresholding and suppression verification.
- `TestSuppressesGohumanize`, `TestHasUnknownHumanizeLinterName`, `TestVerifySuppressions_UnknownLinterName`, `TestVerifySuppressions_StaleSuppression`, `TestVerifySuppressions_UsedSuppression` — unit tests for suppression verification logic.
- `TestRuleBytes_SizeBucket_NoFalsePositive` — regression test for H001 size-bucket filter.
- Testdata fixture: `testdata/h001_sizebucket/` (switch + slice-based size-bucket lookups that must not trigger H001).
- **ADR 0002** (`docs/adr/0002-suppression-verification.md`) — Documents the design decision to run suppression verification as a separate post-detection pass rather than integrating into the main detection loop.
- **ADR 0003** (`docs/adr/0003-confidence-aware-exit-codes.md`) — Documents the ternary exit-code scheme (0/1/2). Originally implemented as a CLI-local function; now uses the SDK's `linter.ExitCodeByConfidence` (see Resolution section).
- **Exported `RuleIDH001`–`RuleIDH009` constants** — single source of truth for rule IDs in `rules.go`, imported by CLI, plugin, and all rule files. Eliminates duplicated string literals.
- **Package-level `var` detection for H007** — `scanFileLevelByteMultiplierMaps` detects `var byteMultipliers = map[string]int64{"KB": 1024}` at package scope. Type-checks map value type (`int*`) to avoid false positives on lookup sets like `map[string]bool`.
- **Import-alias-aware detection** — `buildImportAliases(file)` resolves import aliases (e.g., `str "strings"`) from the AST. Threaded through all pattern helpers via variadic `aliases ...map[string]string` parameter. ADR 0001 documents the syntactic-resolution design decision.
- **Configurable plugin rules** — golangci-lint v2 module plugin supports `enable`/`disable` settings via `.golangci.yml` `linters.settings.custom.gohumanize.settings`. Uses `github.com/golangci/plugin-module-register` for v2-compatible `init()` registration.
- **`--config <file>` CLI flag** — Load rule enable/disable settings from a YAML file. Config-first/CLI-override precedence with set-union semantics.
- **GitHub Action composite** (`action.yml`) — Reusable GitHub Action with inputs: `path`, `enable`, `disable`, `format`, `version`.
- **golangci-lint v2 module plugin registration** — `plugin/plugin.go` refactored to use `register.Plugin("gohumanize", newPlugin)` with `humanizePlugin` implementing `register.LinterPlugin`. `.custom-gcl.yml` build config and `.golangci.custom.yml` runtime config included.
- **`.golangci.custom.yml`** — Example runtime config showing the `linters.settings.custom.gohumanize.type: "module"` section required for module plugin discovery.
- **CI version pinning** — `govulncheck@v1.6.0`, `golangci-lint v2.12.2` (were both `@latest`).
- **ADR 0001** (`docs/adr/0001-import-alias-detection.md`) — Documents the trade-off between syntactic import-alias resolution and full `go/types` integration.
- `TestWriteReport_TextFile`, `TestWriteReport_JSONFile`, `TestWriteReport_FileHandleClosed`, `TestCLI_OutputToFile` — tests for `--output` flag.
- `TestLoadConfig_Valid`, `TestLoadConfig_EmptyFile`, `TestLoadConfig_MissingFile`, `TestLoadConfig_InvalidYAML` — tests for `--config` flag.
- `TestRuleParseBytes_PackageVarMultiplierMap`, `TestRuleParseBytes_AliasedImport` — tests for package-level var and aliased import detection.
- 8 plugin internal tests: `TestParseRuleIDs` (5 subtests), `TestFilterRules_*`, `TestNewPluginWith*`, `TestBuildAnalyzersProducesConfiguredDetector`.
- Testdata fixtures: `testdata/h007_package_var/`, `testdata/h007_aliased_import/`.
- **H008** (manual-ordinal): Detects `switch n%10/100` with st/nd/rd/th cases. Suggests `humanize.Ordinal`.
- **H009** (manual-commaf): Detects `%.Nf` + manual separator grouping. Suggests `humanize.Commaf`.
- Scoped `//nolint:gohumanize:H001` directives — each rule can be suppressed individually.
- Go-style `//lint:ignore gohumanize` alternative syntax.
- `HumanizeDetector` facade with `Run` method.
- `--list-files <dir>` CLI flag for debugging walker scope.
- `--explain Hxxx` CLI flag — prints a one-paragraph rationale for any rule.
- `TestSinglechecker_CleanCode` + `TestSinglechecker_PositiveFinding` — closes 0% coverage gap on the singlechecker binary.
- **Dot-import support** — `buildImportAliases` and `isPackageCall` now handle dot imports (`. "strings"` → bare `HasSuffix` calls). `hasTimeThresholdComparison` also supports dot-imported time constants.
- **H009/H002 overlap disambiguation** — H002 is suppressed when H009 fires on the same function, preventing double diagnostics for float-with-comma formatting.
- **Per-statement `//nolint` suppression** — `//nolint:gohumanize` directives placed inside a function body (on any line between `{` and `}`) are now honoured as function-level suppressions. Developers and AIs can place the directive on the specific statement that triggers the finding instead of only on the function declaration. Uses a shared `commentAssociatedWithFunc` helper across the CLI, plugin, and verification paths.
- **Plugin integration test** (`TestCustomGCLIntegration`) — End-to-end test that builds the `custom-gcl` binary via `golangci-lint custom`, runs it on a temp module with a known H001 pattern, and asserts the diagnostic appears. Gated behind `testing.Short()` (requires network + git clone).
- **`--behavior-delta` and `--save-baseline` CLI flags** — Save a finding baseline with `--save-baseline baseline.json`, then compare future runs with `--behavior-delta baseline.json`. Exits 0 if no delta, 1 if findings added or removed. Comparison is by (rule, file, line) tuple — message text changes don't trigger a delta.
- **`action.yml` `save-baseline` and `behavior-delta` inputs** — The GitHub Action now exposes both baseline flags for CI delta workflows.
- **ADR 0004** (`docs/adr/0004-behavior-delta.md`) — Documents the (rule, file, line) comparison key decision, exit code semantics, and the `0o600` baseline file mode rationale.
- **ADR 0005** (`docs/adr/0005-h009-h002-disambiguation.md`) — Documents the unidirectional suppression of H002 when H009's pattern matches, and why H009 (more specific) wins over H002.
- **ADR 0006** (`docs/adr/0006-per-statement-suppression.md`) — Documents the line-range matching approach (Approach B) for in-body suppression, why per-statement token.Pos (Approach A) was rejected, and when to upgrade.
- **CLI suppression scoping fix** — `checkFuncDecls` now uses `funcSuppressions` + `isSuppressedRule` instead of the binary `hasNoLintDirective`. Scoped `//nolint:gohumanize:H001` directives now work identically in CLI and plugin paths. The dead `hasNoLintDirective` function has been removed.
- **`looksLikeRuleID` helper** — Distinguishes rule IDs (H followed by digits) from comma-separated linter names in `isSuppressedRule`, preventing `//nolint:gohumanize,other` from accidentally scoping to the linter name "other" instead of suppressing all H-rules.
- `TestCLI_ScopedSuppression_H001Only`, `TestCLI_ScopedSuppression_DoesNotSuppressOtherRule`, `TestCLI_InBodyDirectiveIsolation`, `TestCLI_InBodyDirectiveStandalone` — tests for scoped and in-body suppression in the CLI path.
- `TestVerifySuppressions_InBodyStale`, `TestVerifySuppressions_InBodyUsed` — tests for in-body directive verification (stale detection and used-suppression detection).
- `TestFuncSuppressionsAssociation` — replaces `TestHasNoLintDirective`, testing the new `funcSuppressions` + `isSuppressedRule` API with scoped-suppression cases.
- Testdata fixtures: `testdata/h001_scoped_h001/`, `testdata/h001_scoped_h002_only/`, `testdata/h001_inbody_isolation/`, `testdata/h001_suppressed_inbody_standalone/`.
- Extended `TestCustomGCLIntegration` with `t.Run` subtests for `minConfidence: "full"` and `verifySuppressions: true`.
- Release notes draft (`docs/release-notes-v0.2.0.md`).
- CONTRIBUTING.md suppression behavior section documenting three suppression paths, scoped directives, and namespace gotcha.
- **Dot-import H003 testdata** (`testdata/h003_dot_import/`) — Exercises `isTimeDurationIdentifier` and the dot-import branch of `isDotImportOf`, restoring core coverage above 88.4%.
- `TestLoadConfigRules_*`, `TestAppendSuppressionFindings`, `TestSaveBaselineAndNotify`, `TestRunBehaviorDelta_*`, `TestLoadBaseline_MalformedJSON`, `TestSaveBaseline_FilePermissions`, `TestRuleRelTime_DotImport` — unit tests for CLI helpers and dot-import H003 detection.
- **`RuleIDH0SUP` exported constant** — The pseudo-rule ID for suppression verification is now an exported constant in `rules.go`, replacing the private `suppressionVerificationRuleID`.
- **`docs/DOMAIN_LANGUAGE.md`** — Ubiquitous-language glossary covering rule IDs, corroborating signals, confidence levels, suppression directives, and detection architecture.
- **`action.yml` inputs** — Added `min-confidence` and `verify-suppressions` inputs to the GitHub Action.
- `TestRunDetector_H0SUPBypassesConfidenceFilter`, `TestRunDetector_FiltersByConfidence`, `TestRunDetector_VerifySuppressions` — analysistest-based plugin tests for confidence filtering and suppression verification.
- `TestRuleParseBytes_DotImport` — verifies H007 detects dot-imported `strings` functions.
- `TestH009H002_NoOverlap` — verifies H002 is suppressed when H009 fires.
- Testdata fixtures: `testdata/h007_dot_import/`, `testdata/h009_h002_overlap/`, `testdata/analysistest/h0supbypass/`.
- Analysistest fixtures H002-H009 wired into `TestAnalyzerAnalysistest` (was: H001 only).
- `TestRuleCountConsistency` — anti-ghost-rule guard asserting `AllRules()` and `allRuleDetectors()` stay in sync.
- `TestHasOrdinalSwitch` — 8-case table-driven test for H008's pattern detector.
- White-box tests for `pattern_commaf.go` helpers (`walkFormatFloatVerbs`, `scanDottedPercentFloat`, `scanBarePercentFloat`, `hasFormatFloatPrecision`).
- Negative testdata for H008 (`h008_negative/`) and H009 (`h009_negative/`).
- `--output <file>` CLI flag for writing reports to a file instead of stdout.
- `cachedDetectorByID` using `sync.OnceValue` to avoid rebuilding the detector lookup map on every `NewHumanizeDetector` call.
- `govulncheck` step in CI.
- Self-scan CI step (build linter, run on own source, assert 0 findings).
- `CONTRIBUTING.md` rewrite with rule-addition and CLI-flag checklists, correct dev commands.
- Markdown docs per rule: `docs/rules/H001.md` through `H009.md`.
- Release workflow at `.github/workflows/release.yml` (tagged builds).
- CI coverage reporting via Codecov.
- `--version` stderr warning when built without ldflags (dev builds).

### Changed

- **`go-linter-sdk` bumped to v0.2.0** — The published SDK now provides `RuleMeta.ToolName` / `WithToolName`, which `rule_*.go` already set. This unblocks the custom-gcl plugin build (the temp golangci-lint module resolves the PUBLISHED SDK — local `replace` directives do not propagate) and the `TestCustomGCLIntegration` end-to-end test. Closes TODO T19.
- **Local absolute-path `replace` directives removed from `go.mod`** — Both `go-linter-sdk` and `go-finding` targets pointed at `~/projects/*` siblings that only exist on the dev machine (any other environment, including CI, fails at `go build`). Local dev keeps sibling resolution via `go.work`; published v0.2.0/v1.6.0 match the sibling content.
- **H0SUP bypasses confidence filtering** — Suppression-verification findings (`H0SUP`) now skip the `minConfidence` check in the plugin path. Without this, `minConfidence: "full"` would silently hide all stale-directive diagnostics (they are `ConfidenceHigh`, below `ConfidenceFull`). `RuleIDH0SUP` exported as a public constant in `rules.go`.
- **H004 suggestion text corrected** — Now references `github.com/dustin/go-humanize/english.Plural` and `english.PluralWord` (the correct sub-package APIs), not the non-existent `humanize.Plural`.
- **H001 confidence levels refined** — Unit-slice matches without division-by-1024 are lowered from `ConfidenceFull` to `ConfidenceMedium`. Switch-based and unit-slice size-bucket lookups without div1024 are suppressed entirely (`return nil`).
- **CLI `main()` refactored** — Extracted into `run()` + `runScan()` for testability and to stay under the cyclop complexity threshold. New sentinel errors `errNoPath` and `errInvalidConfidence`.
- `isPackageCall` now accepts variadic `aliases ...map[string]string` for import-alias resolution — backward compatible (existing callers compile without changes).
- `flake.nix` lint script improved: proper exit-code propagation via `output=$(...); code=$?; ... exit $code` pattern (was `grep -v ... || true` which swallowed exit codes).
- `plugin/plugin.go` doc comment rewritten with complete 4-step integration guide including the critical `linters.settings.custom` section.
- Coverage: core 88.9%, CLI 57.7%, plugin 97.1% (all packages up from prior release: core 88.4%, CLI 38.4%, plugin 95.7%).
- Dependencies: `github.com/golangci/plugin-module-register v0.1.2` and `gopkg.in/yaml.v3 v3.0.1` promoted from indirect to direct.
- `plugin.run` → `plugin.analyzeHumanize` (better grep-ability).
- `printRules()` now writes to stdout (was: stderr) so it pipes cleanly.
- Refactored suppression parser to support both `//nolint:` and `//lint:ignore` flavours with colon-scoped rule IDs.
- `gofumpt` + `goimports -local github.com/larsartmann/` formatting applied.
- `TestRuleBytes_SuppressedByDirective` → `TestCLI_SuppressedByDirective` (clarity).
- `hasPercentNF` split into `walkFormatFloatVerbs` + `scanDottedPercentFloat` + `scanBarePercentFloat` so each helper stays under the cyclop/gocognit/nestif thresholds.
- Replaced magic numbers in `pattern_ordinal.go` and `pattern_commaf.go` with named constants (`ordinalModTen`, `minOrdinalSuffixesHit`, `percentChar`, `floatVerb`, …).
- `TestHasEqualsOneBranch` and `TestHasCommaOrSeparator` extracted their case tables into shared `boolSrcCase`-typed helpers so the test functions stay under the funlen threshold.
- `ruleExplanations` map switched to local `h001`–`h009` constants so `goconst` doesn't flag the repeated rule-ID strings.
- Removed `//nolint:erraudit` directives from `walker.go` (the tool is not a golangci-lint linter — referenced directly in prose comments instead).
- `nix run .#lint` now filters the informational "Found unknown linters in //nolint directives: gohumanize" warning that golangci-lint emits for the project's own plugin-loaded analyzer.

### Removed

- **`HumanizeDetector.RunOverPackage`** — Dead code; the golangci-lint plugin path uses `runDetector` (in `plugin/plugin.go`) which inlines the same iteration logic with additional features (generated-file skipping, confidence filtering, suppression verification, per-finding position reporting via `findingToTokenPos`). `RunOverPackage` had no live callers.

### Fixed

- **Plugin-path H0SUP diagnostics were silently dropped by golangci-lint's nolint filter** — Stale-directive H0SUP findings were anchored at the `//nolint:gohumanize` directive itself, and golangci-lint suppresses any `gohumanize`-reported issue covered by that directive (self-referential suppression). The plugin path now re-anchors H0SUP findings to the file position (1:1) with the directive coordinates embedded in the message, e.g. `... (directive at main.go:3)`. The CLI path has no nolint filter and keeps the precise directive position. Verified end-to-end via `TestCustomGCLIntegration/verify_suppressions` against a real `custom-gcl` binary. Latent since the test was written: the custom-gcl build always failed to compile before the SDK v0.2.0 bump, so the subtest never executed.
- **`TestIsGeneratedFile_FilenameOnly` expectations restored to gogenfilter v3.4.0 reality** — A prior docs-sync commit flipped `models.go`/`querier.go`/`batch.go` to `want=false` based on the false claim that gogenfilter deliberately ignores bare sqlc default names. gogenfilter v3.4.0's `sqlcFilenamePatterns` DOES include them, and the 2026-08-10 validation sweep explicitly rated this behavior as correct. Expectations restored to `true` with an accurate comment; AGENTS.md claim corrected.
- Pre-existing lint gate failures resolved (was failing on `main`): funlen (test table extracted to package-level `funcSuppressionCases`), gosec G306 (0o600 for temp files), noctx (`exec.CommandContext`), G204 (scoped `//nolint:gosec` with justification), tparallel (top-level `t.Parallel()`), unparam (dead `name` parameter dropped from `suppression_test.go` helper). `golangci-lint run ./...` now reports 0 issues.
- **H004 suggested a non-existent API** — `humanize.Plural` does not exist in the root package. The correct functions are `english.PluralWord(n, singular, plural)` and `english.Plural(n, singular, plural)` in the `english` sub-package. This caused downstream projects to apply suggested fixes that would not compile.
- **H001 false-positive on size-bucket lookup tables** — Functions with a `switch` statement or `[]string` of byte-unit labels but no division by 1024 are size-bucket lookup tables, not byte formatters. They are no longer flagged.
- **golangci-lint v2 module plugin discovery** — Root cause identified: `.golangci.yml` requires `linters.settings.custom.gohumanize.type: "module"` section for golangci-lint to discover the linter at runtime. Without it, the custom binary reports "unknown linters: gohumanize". Documented in plugin.go and `.golangci.custom.yml`.
- H007 no longer false-positives on `map[string]bool` lookup sets — `isByteUnitMultiplierMapLiteral` type-checks the map value to ensure it's numeric (`int*`).
- Rule ID string literals replaced with exported `RuleIDH001`–`RuleIDH009` constants across all 9 rule files, `rules.go`, and `main.go` — eliminates duplication and `goconst` warnings.
- **H009** (manual-commaf) is now properly registered in `AllRules()` and `allRuleDetectors()` — was a "ghost rule" (detector and test existed but the rule was invisible to the CLI / plugin / registry).
- Multi-digit float precision bug in `scanDottedPercentFloat` and `scanBarePercentFloat` — `%.10f` and `%34f` were incorrectly rejected because the scanners only matched repeated same-digit patterns.
- Stale "7 rules" comments in `rules.go` and `bench_test.go` corrected to "9 rules".
- `go.mod` `replace` directives removed (were re-added for local dev after v0.1.0, blocking `go install` and plugin loading).
- 9 lint issues across `pattern_commaf.go`, `pattern_ordinal.go`, `pattern_helpers_test.go`, and `cmd/go-humanize-linter/main.go` resolved (was: 9 issues, now: 0).
- **H002/H009 false positive on `strings.Join(args, " ")`** — Space is not a thousands separator. Removed `" "` from the separator list in `isCommaSeparatorCall` (`pattern_comma.go`). Regression tests added: `TestRuleComma_StringsJoinSpace_NoFalsePositive` and `TestRuleCommaf_StringsJoinSpace_NoFalsePositive`.
- **`findingToTokenPos` panic on out-of-range line numbers** — Added defensive `tf.LineCount() < f.Position.Line` guard before `token.File.LineStart`, which panics with "illegal line number (past end of file)" on out-of-range input. Regression test: `TestFindingToTokenPos_OutOfRangeLine`.
- **Dishonest `//nolint:gosec` comment** — `WalkGoDir` is a public API, so "trusted project dirs" was inaccurate. Comment rewritten to accurately describe the trust model: path comes from `filepath.WalkDir` on the caller-supplied directory; trust is delegated to the caller.
- **`action.yml` and `CONTRIBUTING.md` missing capital-L `GOPRIVATE` variant** — Both listed only `github.com/larsartmann/*` but `gogenfilter` is at `github.com/LarsArtmann/gogenfilter/v3`. Both case variants are now documented.

## [0.1.0] - 2026-07-30

### Added

- **H001** (manual-bytes-format): Detects hand-rolled byte-size formatting — KMGTPE index trick, []string unit slices, 2+ byte-unit strings with division by 1024. Suggests `humanize.Bytes` / `humanize.IBytes`.
- **H002** (manual-comma-format): Detects thousands-separator insertion loops (mod-3, step-by-3, digit-conversion fallback). Suggests `humanize.Comma`.
- **H003** (manual-reltime-format): Detects relative-time formatting with `time.Since`/`time.Sub` + "ago" strings + time thresholds. Suggests `humanize.RelTime` / `humanize.Time`.
- **H004** (manual-plural): Detects English pluralization via `if x == 1` switches and singular/plural parameter pairs. Suggests `humanize.Plural` / `humanize.PluralWord`. Two-stage false-positive filter: (1) branch must contain string literals, (2) function must return string type.
- **H005** (manual-si-format): Detects SI-prefix formatting (division by 1000/1e6 + K/M suffix). Suggests `humanize.SI`.
- **H006** (manual-ftoa): Detects nested `strings.TrimRight(strings.TrimRight(x, "0"), ".")` for trailing-zero stripping. Suggests `humanize.Ftoa`.
- **H007** (manual-parse-bytes): Detects hand-rolled byte-size string parsing (multiple HasSuffix/CutSuffix on byte units, or map[string]int64 multiplier lookups). Suggests `humanize.ParseBytes`.
- CLI binary (`cmd/go-humanize-linter/`) with `--enable`, `--disable`, `--format text|json|sarif`, `--quiet` flags.
- golangci-lint plugin wrapper (`plugin/`) exposing all rules as `analysis.Analyzer` named `gohumanize`.
- Standalone singlechecker (`cmd/gohumanize/`) for testing the plugin path without golangci-lint.
- `DetectFuncDecl()` shared per-function entry point used by both CLI and plugin.
- `flake.nix` with test, test-race, bench, build, vet, lint, and coverage apps.
- CI workflow with test+vet and lint jobs.
- Negative testdata for H001, H003, H004, and H005.
- 26 library tests + 3 CLI integration tests + 3 plugin tests, all passing with race detector.

### Changed

- H001 message now varies by trigger: KMGTPE index trick, unit string slice, or unit-string count.
- patterns.go split into 8 focused files: pattern_bytes.go, pattern_comma.go, pattern_time.go, pattern_plural.go, pattern_si.go, pattern_ftoa.go, pattern_parsebytes.go, pattern_helpers.go.

### Removed

- Temporary `replace` directives in go.mod (resolved after go-linter-sdk v0.1.0 tag).
