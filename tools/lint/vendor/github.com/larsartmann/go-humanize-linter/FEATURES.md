# Features

> Honest feature inventory by status. Verified against code on 2026-09-26
> (docs-health AUDIT pass; covers the v0.4.0 rule set).
>
> Known incident (resolved 2026-10-23 in v0.5.0): bumper sweeps left go.mod
> at `go 1.27` while CI pinned `go-version: "1.26"`, keeping main CI red
> 2026-09-19 → 2026-10-23 (T33); resolved by adopting Go 1.27 everywhere.
> Every FULLY_FUNCTIONAL claim below was green as of v0.4.0 (tag `6a3267f`).

## Status legend

| Status               | Meaning                                                      |
| -------------------- | ------------------------------------------------------------ |
| FULLY_FUNCTIONAL     | Code present AND working (tests pass or exercised).          |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps, edge-case bugs, or missing pieces. |
| PLANNED              | Designed or documented but **no code exists yet**.           |
| BROKEN               | Code exists but does not work / is disabled / fails.         |

## Rules

All 11 rules are registered in `AllRules()` (`rules.go`) and `allRuleDetectors()`
(`rules.go`), and share a single per-function entry point `DetectFuncDecl()`.
(H011 is reserved for a future manual-si-parse rule.)

| Rule | Name                  | Status           | Detects                                                                                                                                                             | Suggests                                          |
| ---- | --------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| H001 | manual-bytes-format   | FULLY_FUNCTIONAL | KMGTPE index trick, unit slices, unit strings + div1024. Size-bucket lookup tables (switch/unit-slice without div1024) are excluded.                                | `humanize.Bytes` / `humanize.IBytes`              |
| H002 | manual-comma-format   | FULLY_FUNCTIONAL | mod-3, step-by-3, digit-conversion fallback                                                                                                                         | `humanize.Comma`                                  |
| H003 | manual-reltime-format | FULLY_FUNCTIONAL | time-diff + "ago" + thresholds                                                                                                                                      | `humanize.RelTime` / `humanize.Time`              |
| H004 | manual-plural         | FULLY_FUNCTIONAL | `if x == 1` with string branch + string-return-type filters                                                                                                         | `english.Plural` / `english.PluralWord`           |
| H005 | manual-si-format      | FULLY_FUNCTIONAL | division by 1000 + K/M suffix, excludes byte units                                                                                                                  | `humanize.SI`                                     |
| H006 | manual-ftoa           | FULLY_FUNCTIONAL | nested `strings.TrimRight(strings.TrimRight(x,"0"),".")`                                                                                                            | `humanize.Ftoa`                                   |
| H007 | manual-parse-bytes    | FULLY_FUNCTIONAL | 2+ HasSuffix/CutSuffix on byte units, map multiplier (func + package scope), aliased imports                                                                        | `humanize.ParseBytes`                             |
| H008 | manual-ordinal        | FULLY_FUNCTIONAL | `switch n%10`/`n%100` with st/nd/rd/th cases                                                                                                                        | `humanize.Ordinal`                                |
| H009 | manual-commaf         | FULLY_FUNCTIONAL | `%.Nf` Sprintf + manual comma/separator grouping loop                                                                                                               | `humanize.Commaf`                                 |
| H010 | manual-comma-parse    | FULLY_FUNCTIONAL | Comma strip (ReplaceAll/Replace, Split+Join, rune-filter loop) + strconv parse; CSV splits/validators excluded. Swept 2026-09-18 (169 repos): 1 borderline, 0 FPs.  | `humanize.ParseComma` / `humanize.ParseCommaf`    |
| H012 | manual-word-series    | FULLY_FUNCTIONAL | Comma join + conjunction literal (+ prefix join/last-element for Full). Plain `strings.Join(x, ", ")` without conjunction excluded. Motivated by a real corpus hit. | `english.WordSeries` / `english.OxfordWordSeries` |

## Interfaces

| Feature              | Status           | Notes                                                                                                                                                                                                                                                                          |
| -------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CLI binary           | FULLY_FUNCTIONAL | `--enable`, `--disable`, `--config`, `--format`, `--output`, `--quiet`, `--rules`, `--version`, `--list-files`, `--explain`, `--min-confidence`, `--verify-suppressions` (`cmd/go-humanize-linter/main.go`)                                                                    |
| Go library           | FULLY_FUNCTIONAL | `DefaultRegistry()`, `AllRules()`, `DetectFuncDecl()`, `HumanizeDetector` facade with `Run`, exported `RuleIDH001`–`H012` constants (`rules.go`)                                                                                                                               |
| golangci-lint plugin | FULLY_FUNCTIONAL | `plugin/plugin.go` using `plugin-module-register` v2 module plugin pattern. Configurable enable/disable, `minConfidence`, and `verifySuppressions` via `.golangci.yml` `linters.settings.custom.gohumanize.settings`. Diagnostics at finding position via `findingToTokenPos`. |
| GitHub Action        | FULLY_FUNCTIONAL | `action.yml` composite Action with inputs: path, enable, disable, format, version                                                                                                                                                                                              |
| Nix flake            | FULLY_FUNCTIONAL | `test`, `test-race`, `bench`, `build`, `vet`, `lint`, `coverage` apps                                                                                                                                                                                                          |
| CI workflow          | FULLY_FUNCTIONAL | test (-race) + vet + govulncheck + self-scan + coverage (Codecov) job and golangci-lint job (`.github/workflows/ci.yml`). Versions pinned: `govulncheck@v1.6.0`, `golangci-lint v2.12.2`. Runs red on main since 2026-09-19 — see the go-directive incident note above.        |
| Release workflow     | FULLY_FUNCTIONAL | Tag-triggered (`v*` → `.github/workflows/release.yml`); v0.1.0–v0.4.0 released through it. Publishes GitHub Release + linux-amd64 binaries (multi-arch is TODO T39). Curated release notes are mandatory — `--generate-notes` lists PRs only.                                  |

## Detection capabilities

| Feature                          | Status           | Notes                                                                                                                                                                                                                                              |
| -------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Function-scope AST scanning      | FULLY_FUNCTIONAL | All `FuncDecl`s in non-test, non-generated `.go files                                                                                                                                                                                              |
| Package-level var scanning       | FULLY_FUNCTIONAL | H007 detects `var x = map[string]int64{"KB": 1024}` at package scope                                                                                                                                                                               |
| Multi-signal detection           | FULLY_FUNCTIONAL | Every rule requires 2+ corroborating signals                                                                                                                                                                                                       |
| Word-boundary regex              | FULLY_FUNCTIONAL | Prevents false positives like "MEDIUMBLOB" matching "MB"                                                                                                                                                                                           |
| Underscore digit normalization   | FULLY_FUNCTIONAL | `normLit()` strips `1_000_000` → `1000000`                                                                                                                                                                                                         |
| Named constant detection         | FULLY_FUNCTIONAL | `hasConst1024` finds `const unit = 1024` patterns                                                                                                                                                                                                  |
| Directory skip rules             | FULLY_FUNCTIONAL | `skipDirName` (`walker.go`): hidden dirs (leading `.`), backup trees (trailing `.bak`), `archived`/`forks` basenames — frozen/upstream copies never double-report. Regression: `TestWalkGoDir_SkipsHiddenBackupArchivedForks`                      |
| Generated file skipping          | FULLY_FUNCTIONAL | gogenfilter two-phase detection (sqlc, templ, protobuf, deepcopy-gen, wire, moq, mockgen, mockery, easyjson, counterfeiter, generic `// Code generated by`) + legacy-suffix fallback (`_gen.go`, `.gen.go`, `_templ.go`) in `pattern_generated.go` |
| `//nolint:gohumanize` directives | FULLY_FUNCTIONAL | Bare, `:all`, scoped `:H001`, comma-lists, and `//lint:ignore` syntax (`pattern_helpers.go`)                                                                                                                                                       |
| Import-alias resolution          | FULLY_FUNCTIONAL | `buildImportAliases(file)` resolves `str "strings"` → `{"str": "strings"}`; dot imports (`. "strings"`) and aliased time constants supported (fixtures: `testdata/h007_dot_import/`, `h010_dot_import/`, `h003_dot_import/`)                       |
| Map-type checking (H007)         | FULLY_FUNCTIONAL | `isByteUnitMultiplierMapLiteral` checks `map[string]int*` value type to avoid flagging `map[string]bool` lookup sets                                                                                                                               |
| go/types integration             | PLANNED          | Pure syntactic analysis, no type info. ADR 0001 documents the trade-off.                                                                                                                                                                           |

## Configuration

| Feature                          | Status           | Notes                                                                                                                      |
| -------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------- |
| CLI `--enable`/`--disable` flags | FULLY_FUNCTIONAL | Per-rule filtering on the command line                                                                                     |
| CLI `--config` YAML file         | FULLY_FUNCTIONAL | Load enable/disable rules from `.gohumanize.yaml`. Config-first, CLI-override with set-union semantics.                    |
| CLI `--min-confidence` flag      | FULLY_FUNCTIONAL | Filter findings by confidence level: `low`, `medium`, `high`, `full` (default: `low`). Uses `finding.ByConfidenceAtLeast`. |
| CLI `--verify-suppressions` flag | FULLY_FUNCTIONAL | Detects stale `//nolint:gohumanize` directives (suppress zero findings) and misspelled linter names. Reports as `H0SUP`.   |
| CLI confidence-aware exit codes  | FULLY_FUNCTIONAL | Exit 0 = clean, exit 1 = high/full-confidence finding (must fix), exit 2 = only medium/low (triage).                       |
| Plugin enable/disable settings   | FULLY_FUNCTIONAL | `linters.settings.custom.gohumanize.settings.enable`/`disable` in `.golangci.yml`                                          |
| Plugin `minConfidence` setting   | FULLY_FUNCTIONAL | `linters.settings.custom.gohumanize.settings.minConfidence` in `.golangci.yml` — same levels as CLI `--min-confidence`     |
| Plugin `verifySuppressions`      | FULLY_FUNCTIONAL | `linters.settings.custom.gohumanize.settings.verifySuppressions` in `.golangci.yml` — same as CLI `--verify-suppressions`  |
| Scoped `//nolint` directives     | FULLY_FUNCTIONAL | Per-function suppression via `//nolint:gohumanize:H001`                                                                    |

## Validation

| Sweep                                                                                      | Date       | Corpus           | Result                                                                                                                                 |
| ------------------------------------------------------------------------------------------ | ---------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| H001–H007                                                                                  | 2026-07-30 | 190+ repos       | 97 findings, ~0% FP after H004 fix (`docs/validation/2026-07-30_real-world-sweep.md`)                                                  |
| H001–H009                                                                                  | 2026-07-31 | 327 projects     | 242 findings, ~0% FP (`docs/validation/2026-07-31_real-world-sweep.md`)                                                                |
| All rules + all v0.2.0 features (`--verify-suppressions`, `--min-confidence`, gogenfilter) | 2026-08-10 | 158 projects     | 0 findings; 3 stale directives found; gogenfilter FP rate ~2.8%, no app code missed (`docs/validation/2026-08-10_real-world-sweep.md`) |
| H010                                                                                       | 2026-09-18 | 169 repos        | 1 borderline (locale-aware parser), 0 clear FPs (`docs/validation/2026-09-18_h010-sweep.md`)                                           |
| H012                                                                                       | 2026-09-18 | top-level corpus | 2 true positives, 0 FPs (one prose-only FP fixed by the conjunction-position filter; see `docs/rules/H012.md`)                         |

- H004 tuned from ~60% FP to ~0% FP across two iterations (branch-string + return-type filters).
- H010 shipped under unswept-rule policy D3 (multi-signal + Full tier + 4-class negative corpus + same-day sweep).
- Corpus note: consumer repos (KeyCountdown, Kernovia, CreditReformBilanzampel) are intentionally left unfixed as the live demonstration corpus (decision 2026-09-19).

## Test coverage

Computed via `go test ./... -cover` on 2026-08-05 (pre-H010/H012; recompute
after T33 unblocks the toolchain):

| Package                          | Coverage                                                                                                       |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `go-humanize-linter` (core)      | 88.9%                                                                                                          |
| `cmd/go-humanize-linter` (CLI)   | 57.7%                                                                                                          |
| `cmd/gohumanize` (singlechecker) | 0.0% (1-liner `singlechecker.Main` wrapper — not coverable in-process; smoke-tested via `TestSinglechecker_*`) |
| `plugin`                         | 97.1%                                                                                                          |
