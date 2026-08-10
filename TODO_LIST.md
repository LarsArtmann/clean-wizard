# TODO LIST

**Last Updated:** 2026-08-10
**Focus:** Actionable items for the next 2-4 weeks
**Source:** Verified against code on 2026-08-10; harvested from 2026-07-06 through 2026-08-05 status reports

---

## Critical (Do First)

| #   | Task                                                                                                                                                                                            | Impact | Effort | Source                        |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----------------------------- |
| 1   | Migrate 5 command files to classified errors: `init.go`, `githistory.go`, `config.go`, `clean_select.go`, `profile.go` (30+ bare `fmt.Errorf` calls classify as Transient regardless of nature) | HIGH   | MED    | 2026-07-06 hardening          |
| 2   | Classify `ErrGitNotAvailable` as Infrastructure (`errorfamily.NewInfrastructure`) — currently defaults to Transient                                                                             | MED    | LOW    | 2026-07-06 hardening          |
| 3   | Enrich scan JSON output with `family`/`code`/`retryable` fields — only clean JSON was enriched; scan uses disjoint schema                                                                       | MED    | LOW    | 2026-07-06 hardening          |
| 4   | Fix scan JSON swallowing marshal errors — `outputScanJSON` prints error and returns silently instead of propagating                                                                             | MED    | LOW    | 2026-07-06 hardening          |
| 5   | Wire `errorfamily.HandleError` in `main.go` or remove dead message templates (3 templates registered but never consumed)                                                                        | LOW    | LOW    | 2026-07-06 hardening          |

## High Priority

| #   | Task                                                                                                          | Impact | Effort | Source                        |
| --- | ------------------------------------------------------------------------------------------------------------- | ------ | ------ | ----------------------------- |
| 6   | Wire `OperationSettings` from YAML config → cleaner constructors (cleaners use hardcoded defaults)            | HIGH   | HIGH   | 2026-07-06 DI/workflow        |
| 7   | Add BDD tests for execution layer (Ginkgo) — workflow DAG, retry, parallel execution                          | HIGH   | MED    | 2026-07-06 BDD audit          |
| 8   | Add BDD tests for Docker, Homebrew, Go cleaners (9 of 13 cleaners have NO BDD tests)                          | HIGH   | HIGH   | 2026-07-06 BDD audit          |
| 9   | Migrate `docker_parsing.go` `sizeMultiplier` map to `humanize.ParseBytes` (H007 violation, mirrors b7692ff)   | MED    | LOW    | 2026-08-05 linter             |

## Medium Priority

| #   | Task                                                                                                | Impact | Effort | Source                  |
| --- | --------------------------------------------------------------------------------------------------- | ------ | ------ | ----------------------- |
| 10  | Implement `scan --profile` filtering or remove the flag (currently warns but shows all cleaners)    | MED    | MED    | 2026-07-06 DI/workflow  |
| 11  | Logger globals (`L`, `StdLogger`) → DI-injected logger — root cause of test race conditions         | MED    | MED    | 2026-07-06 DI/workflow  |
| 12  | Split files over 350 lines: `compiledbinaries.go` (585), `docker.go` (524), `nodepackages.go` (523) | MED    | MED    | Code quality            |
| 13  | Add CLI command tests: profile, config, scan, init (clean integration test exists)                  | MED    | HIGH   | 2026-07-06 BDD audit    |
| 14  | Extract `"go-build*"` string constant in `golang_cache_cleaner.go`                                  | LOW    | LOW    | Lint (goconst)          |
| 15  | Improve Nix size estimation (currently hardcoded 50MB/generation)                                   | MED    | MED    | FEATURES audit          |
| 16  | Add tests for `getRegistryName` reverse lookup (`scan.go:246`)                                      | MED    | LOW    | Pre-existing            |
| 17  | Move `/tmp/go-humanize-linter` into repo (`tools/lint/`) so CI can reproduce the H007 check          | MED    | LOW    | 2026-08-05 linter       |
| 18  | Wire `go-humanize-linter` into `flake.nix` `checks` (or pre-commit) — H007 violations can regress  | MED    | LOW    | 2026-08-05 linter       |

## Low Priority / Polish

| #   | Task                                                                             | Impact | Effort | Source           |
| --- | -------------------------------------------------------------------------------- | ------ | ------ | ---------------- |
| 19  | Remove `infertypeargs` warnings (15+ places with unnecessary explicit type params) | LOW    | LOW    | Lint diagnostics |
| 20  | Add Gherkin `.feature` files for top 3 cleaners                                  | MED    | MED    | BDD testing      |
| 21  | Standardize BDD test naming (`*_ginkgo_test.go` → consistent pattern)            | LOW    | LOW    | BDD testing      |
| 22  | Add `--dry-run` to scan command (parity with clean)                              | LOW    | LOW    | Pre-existing     |
| 23  | Add `--keep-generations` flag for Nix cleaner                                    | LOW    | LOW    | Pre-existing     |
| 24  | ~~Fix `TestBooleanSettingsCleaners/Cargo` flake — counts 2 items (registry + git) but test expects 1~~ DONE 2026-08-10 | LOW    | LOW    | 2026-08-05 linter |
| 25  | `ParseNumberAndUnit` in `fsutil.go:483` is a single-callsite helper after b7692ff — inline or delete | LOW    | LOW    | 2026-08-05 linter |
| 26  | Add regression test that `parseSize("garbage")` produces a wrapped error chain (errorfamily.Classify → Rejection) | LOW    | LOW    | 2026-08-05 linter |

## Architecture Planning (Long Term)

| #   | Task                                                                                  | Impact | Effort | Source                    |
| --- | ------------------------------------------------------------------------------------- | ------ | ------ | ------------------------- |
| 27  | Split `internal/domain/` into `enums/`, `operations/`, `types/` sub-packages          | HIGH   | HIGH   | Architecture deepening   |
| 28  | Split `internal/cleaner/` into per-domain sub-packages (nix/, docker/, golang/, etc.) | HIGH   | HIGH   | Architecture deepening   |
| 29  | Register individual cleaners as DI providers (enables per-cleaner config)             | HIGH   | HIGH   | 2026-07-06 DI/workflow    |
| 30  | Make adapters interface-backed with `do.As` aliasing                                  | MED    | HIGH   | 2026-07-06 DI/workflow    |

---

**Status:** 26 actionable items (5 Critical, 4 High, 9 Medium, 8 Low, 4 Long-term)

**Build:** `GOEXPERIMENT=jsonv2 go build ./...` ✅ PASS
**Tests:** `GOEXPERIMENT=jsonv2 go test ./... -short` — 23/23 packages PASS
