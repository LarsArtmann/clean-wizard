# TODO LIST

**Last Updated:** 2026-07-13
**Focus:** Actionable items for the next 2-4 weeks
**Source:** go-error-family hardening session (2026-07-06), architecture review, code quality scan, BDD testing audit

---

## Critical (Do First)

| #   | Task                                                                                                                                                                                            | Impact | Effort | Source                |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | --------------------- |
| 1   | Migrate 5 command files to classified errors: `init.go`, `githistory.go`, `config.go`, `clean_select.go`, `profile.go` (30+ bare `fmt.Errorf` calls classify as Transient regardless of nature) | HIGH   | MED    | go-error-family audit |
| 2   | Classify `ErrGitNotAvailable` as Infrastructure (`errorfamily.NewInfrastructure`) — currently defaults to Transient                                                                             | MED    | LOW    | go-error-family audit |
| 3   | Enrich scan JSON output with `family`/`code`/`retryable` fields — only clean JSON was enriched; scan uses disjoint schema                                                                       | MED    | LOW    | go-error-family audit |
| 4   | Fix scan JSON swallowing marshal errors — `outputScanJSON` prints error and returns silently instead of propagating                                                                             | MED    | LOW    | go-error-family audit |
| 5   | Wire `errorfamily.HandleError` in `main.go` or remove dead message templates (3 templates registered but never consumed)                                                                        | LOW    | LOW    | go-error-family audit |

## High Priority

| #   | Task                                                                                               | Impact | Effort | Source              |
| --- | -------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------- |
| 6   | Wire `OperationSettings` from YAML config → cleaner constructors (cleaners use hardcoded defaults) | HIGH   | HIGH   | DI/workflow session |
| 7   | Add BDD tests for execution layer (Ginkgo) — workflow DAG, retry, parallel execution               | HIGH   | MED    | BDD testing audit   |
| 8   | Add BDD tests for Docker, Homebrew, Go cleaners (9 of 13 cleaners have NO BDD tests)               | HIGH   | HIGH   | BDD testing audit   |
| 9   | Fix mixed receiver warnings (10 enum types use both pointer and value receivers)                   | LOW    | LOW    | Lint (recvcheck)    |

## Medium Priority

| #   | Task                                                                                                | Impact | Effort | Source              |
| --- | --------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------- |
| 10  | Implement `scan --profile` filtering or remove the flag (currently warns but shows all cleaners)    | MED    | MED    | DI/workflow session |
| 11  | Logger globals (`L`, `StdLogger`) → DI-injected logger — root cause of test race conditions         | MED    | MED    | DI/workflow session |
| 12  | Split files over 350 lines: `compiledbinaries.go` (585), `docker.go` (524), `nodepackages.go` (523) | MED    | MED    | Code quality        |
| 13  | Add CLI command tests: profile, config, scan, init (clean integration test exists)                  | MED    | HIGH   | BDD testing audit   |
| 14  | Reduce `GetOperationType` complexity (17 → <10) using map lookup                                    | LOW    | LOW    | Lint (cyclop)       |
| 15  | Extract `"go-build*"` string constant in `golang_cache_cleaner.go`                                  | LOW    | LOW    | Lint (goconst)      |
| 16  | Improve Nix size estimation (currently hardcoded 50MB/generation)                                   | MED    | MED    | Features audit      |
| 17  | Add tests for `getRegistryName` reverse lookup (`scan.go:246`)                                      | MED    | LOW    | TODO_LIST           |

## Low Priority / Polish

| #   | Task                                                                             | Impact | Effort | Source      |
| --- | -------------------------------------------------------------------------------- | ------ | ------ | ----------- |
| 18  | Remove `infertypeargs` warnings (8 places with unnecessary explicit type params) | LOW    | LOW    | Lint        |
| 19  | Add Gherkin `.feature` files for top 3 cleaners                                  | MED    | MED    | BDD testing |
| 20  | Fix `nix_test.go` BDD tests (remove `go:build skip_bdd` tag)                     | LOW    | LOW    | BDD testing |
| 21  | Standardize BDD test naming (`*_ginkgo_test.go` → consistent pattern)            | LOW    | LOW    | BDD testing |
| 22  | Fix pre-commit hook timeout (golangci-lint times out)                            | MED    | LOW    | TODO_LIST   |
| 23  | Add `--dry-run` to scan command (parity with clean)                              | LOW    | LOW    | TODO_LIST   |
| 24  | Add `--keep-generations` flag for Nix cleaner                                    | LOW    | LOW    | TODO_LIST   |

## Architecture Planning (Long Term)

| #   | Task                                                                                  | Impact | Effort | Source                 |
| --- | ------------------------------------------------------------------------------------- | ------ | ------ | ---------------------- |
| 25  | Split `internal/domain/` into `enums/`, `operations/`, `types/` sub-packages          | HIGH   | HIGH   | Architecture deepening |
| 26  | Split `internal/cleaner/` into per-domain sub-packages (nix/, docker/, golang/, etc.) | HIGH   | HIGH   | Architecture deepening |
| 27  | Register individual cleaners as DI providers (enables per-cleaner config)             | HIGH   | HIGH   | DI/workflow session    |
| 28  | Make adapters interface-backed with `do.As` aliasing                                  | MED    | HIGH   | DI/workflow session    |

---

**Status:** 28 actionable items (5 Critical, 3 High, 8 Medium, 7 Low, 5 Long-term)
