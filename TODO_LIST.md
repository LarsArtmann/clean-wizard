# TODO LIST

**Last Updated:** 2026-05-03
**Focus:** Actionable items for the next 2-4 weeks
**Source:** Comprehensive analysis from BDD testing, features audit, code quality scan, full code review, architecture review, and architecture deepening analysis

---

## Critical (Do First)

| # | Task | Impact | Effort | Source |
|---|------|--------|--------|--------|
| 1 | Fix `DefaultSettings(OperationTypeSystemCache)` to be platform-aware (returns macOS-only cache types on Linux) | HIGH | LOW | Code Quality Scan |
| 2 | Remove unused `readConfigFile` function in `config.go:52` | LOW | LOW | Lint |
| 3 | Remove unused `nixGenerationsMin` const in `sanitizer_nix.go:9` | LOW | LOW | Lint |
| 4 | Remove unused `bytesPerKB` const in `adapters/nix.go:28` | LOW | LOW | Lint |
| 5 | Fix `SizeEstimate` construction missing `Status` field in `golang_cache_cleaner.go:200` | MED | LOW | Lint (exhaustruct) |

## High Priority

| # | Task | Impact | Effort | Source |
|---|------|--------|--------|--------|
| 6 | Consolidate error packages: merge `internal/errors/`, `internal/pkg/errors/`, `adapters/errors.go`, `pkg/errors/` into one | HIGH | MED | Architecture Review |
| 7 | Extract test utilities from `internal/cleaner/test_*.go` to `internal/testutil/cleaner/` | MED | MED | Architecture Review |
| 8 | Add BDD tests for Docker, Homebrew, Go cleaners (using Ginkgo pattern) | HIGH | HIGH | BDD Testing |
| 9 | Fix `err113` lint violations (~40 `fmt.Errorf` calls should use sentinel errors) | MED | MED | Code Quality |
| 10 | Set up CI pipeline (at minimum: `go build` + `go test`) | HIGH | MED | TODO_LIST |

## Medium Priority

| # | Task | Impact | Effort | Source |
|---|------|--------|--------|--------|
| 11 | Split `internal/domain/` into `enums/`, `operations/`, `types/` sub-packages | HIGH | HIGH | Architecture Deepening |
| 12 | Add service layer `internal/service/` between CLI commands and domain | MED | HIGH | Architecture Deepening |
| 13 | Fix mixed receiver warnings (10 enum types use both pointer and value receivers) | LOW | LOW | Lint (recvcheck) |
| 14 | Reduce `GetOperationType` complexity (17 → <10) using map lookup | LOW | LOW | Lint (cyclop) |
| 15 | Split files over 350 lines: `compiledbinaries.go` (585), `docker.go` (524), `nodepackages.go` (523) | MED | MED | Code Quality |
| 16 | Add CLI command tests (scan, clean, profile, config) | MED | HIGH | BDD Testing |
| 17 | Extract `"go-build*"` string constant in `golang_cache_cleaner.go` | LOW | LOW | Lint (goconst) |
| 18 | Add profile command tests | MED | MED | TODO_LIST |
| 19 | Add scan command tests | MED | MED | TODO_LIST |
| 20 | Add clean command tests | MED | HIGH | TODO_LIST |

## Low Priority / Polish

| # | Task | Impact | Effort | Source |
|---|------|--------|--------|--------|
| 21 | Remove `infertypeargs` warnings (8 places with unnecessary explicit type params) | LOW | LOW | Lint |
| 22 | Add Gherkin `.feature` files for top 3 cleaners | MED | MED | BDD Testing |
| 23 | Fix `nix_test.go` BDD tests (remove `go:build skip_bdd` tag) | LOW | LOW | BDD Testing |
| 24 | Standardize BDD test naming (`*_ginkgo_test.go` → consistent pattern) | LOW | LOW | BDD Testing |
| 25 | Fix pre-commit hook timeout (golangci-lint times out) | MED | LOW | TODO_LIST |
| 26 | Add tests for `getRegistryName` reverse lookup | MED | LOW | TODO_LIST |

## Architecture Planning (Long Term)

| # | Task | Impact | Effort | Source |
|---|------|--------|--------|--------|
| 27 | Split `internal/cleaner/` into per-domain sub-packages (nix/, docker/, golang/, etc.) | HIGH | HIGH | Architecture Deepening |
| 28 | Implement Language Version Manager cleaner (currently placeholder/NO-OP) | MED | MED | Features Audit |
| 29 | Improve Nix size estimation (currently hardcoded 50MB/generation) | MED | MED | Features Audit |
| 30 | Add Homebrew dry-run support (or document limitation clearly) | LOW | MED | Features Audit |

---

## Files Analyzed

### Source Files Read
- `internal/domain/` — all files (interfaces, types, enums, settings, defaults, macros)
- `internal/cleaner/` — cleaner.go, registry.go, registry_factory.go, systemcache.go, golang_cache_cleaner.go
- `internal/result/` — type.go, flow_builder.go, branch_flow.go
- `internal/config/config.go`
- `internal/errors/errors.go`
- `internal/adapters/errors.go`
- `cmd/clean-wizard/commands/` — all files
- `tests/bdd/` — all files

### Documentation Files Read
- `FEATURES.md`, `TODO_LIST.md`, `BDD_TESTS_REVIEW.md`, `AGENTS.md`
- `go.mod`, `Justfile`

---

## Verification

- [x] Build passes: `go build ./...`
- [x] Fixed 2 test failures (SystemCache, GoCacheDeduplication)
- [x] D2 diagrams rendered to SVG
- [ ] All tests pass: `go test ./... -short` (2 failures fixed, needs full re-run)
- [ ] Lint passes: `golangci-lint run`

---

**Status:** 30 actionable items (5 Critical, 5 High, 10 Medium, 6 Low, 4 Long-term)
