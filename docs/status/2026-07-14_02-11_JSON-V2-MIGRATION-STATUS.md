# Status Report: encoding/json/v2 Migration

**Date:** 2026-07-14 02:11
**Session Focus:** Fixing build failures caused by go-auto-upgrade migrating to `encoding/json/v2`
**Commit:** `ef80b9f` — docs: comprehensive documentation health audit and encoding/json v2 migration

---

## What Happened This Session

The BuildFlow `go-auto-upgrade` migrator ran and performed two changes:

1. **Upgraded 7 source files** from `encoding/json` (v1) to `encoding/json/v2` + `encoding/json/jsontext`
2. **Bumped `go-error-family` v0.6.1 → v0.7.0** (which itself now imports `encoding/json/v2`)

Both changes require `GOEXPERIMENT=jsonv2` to compile. Without it, Go's build constraints exclude all files in `encoding/json/v2` and `encoding/json/jsontext`, producing `build constraints exclude all Go files` and cascading failures across `cmd/clean-wizard`, `test`, `tests/bdd`, `tests/benchmark`, `internal/testhelper`.

---

## A) FULLY DONE

| Item                                             | Detail                                                                                                                                                                                                                                                         |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Diagnosed root cause                             | `encoding/json/v2` requires `GOEXPERIMENT=jsonv2`; go-error-family v0.7.0 pulled it in transitively                                                                                                                                                            |
| Migrated 7 production source files               | `internal/format/json.go`, `cmd/clean-wizard/commands/config.go`, `cmd/clean-wizard/commands/scan.go`, `internal/cleaner/projectexecutables.go`, `internal/domain/enum_macros.go`, `internal/domain/githistory_types.go`, `internal/domain/type_safe_enums.go` |
| `MarshalIndent` → `Marshal` + `jsontext` options | `json.MarshalIndent(x, "", "  ")` → `json.Marshal(x, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))` in 3 call sites                                                                                                                                |
| `GOEXPERIMENT=jsonv2` in flake.nix build         | `env.GOEXPERIMENT = "jsonv2"` in `buildGoModule`                                                                                                                                                                                                               |
| `GOEXPERIMENT=jsonv2` in devShells               | Both `default` and `ci` devShells                                                                                                                                                                                                                              |
| AGENTS.md updated                                | Build commands, dependencies list, updated date                                                                                                                                                                                                                |
| `go build ./...`                                 | Passes                                                                                                                                                                                                                                                         |
| `go vet ./...`                                   | Passes                                                                                                                                                                                                                                                         |
| `go test ./... -short`                           | All 24 packages pass, 0 failures                                                                                                                                                                                                                               |
| `nix build .#`                                   | Passes                                                                                                                                                                                                                                                         |
| `gofumpt` formatting                             | All changed files clean                                                                                                                                                                                                                                        |
| Committed                                        | `ef80b9f` on master                                                                                                                                                                                                                                            |

---

## B) PARTIALLY DONE

| Item                    | What's Done                           | What's Missing                                                                                                                                                         |
| ----------------------- | ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| json/v2 migration       | All 7 **production** files migrated   | **3 test files still use `encoding/json` v1**: `internal/domain/type_safe_enums_status_test.go`, `internal/domain/enum_macros_test.go`, `internal/format/json_test.go` |
| flake.nix GOEXPERIMENT  | Build + both devShells                | Not verified with `nix flake check`                                                                                                                                    |
| AGENTS.md documentation | Build commands + dependencies updated | README.md build instructions NOT updated                                                                                                                               |

---

## C) NOT STARTED

- [ ] Migrate the 3 remaining test files to `encoding/json/v2` for consistency
- [ ] `nix flake check` (full CI validation)
- [ ] `golangci-lint run ./...` — 70 pre-existing issues (2 cyclop, 1 exhaustive, 10 exhaustruct, 2 gochecknoglobals, 5 goconst, 50 paralleltest) — none from this session's changes but should be addressed
- [ ] README.md build instructions — no mention of `GOEXPERIMENT=jsonv2` requirement
- [ ] Verify actual binary runtime execution (`nix run .# -- --help`, `nix run .# -- scan --dry-run`)
- [ ] Verify JSON output format equivalence (json/v2 has behavioral differences from v1: HTML escaping, nil map handling, error on unknown fields)
- [ ] Check if `go generate` / `govalid-generate` now works (it was failing in the original BuildFlow run)
- [ ] Add `.envrc` or direnv integration for GOEXPERIMENT (if direnv is used)
- [ ] Verify vendor hash is stable across rebuilds

---

## D) TOTALLY FUCKED UP

### 1. Initial approach was BACKWARDS

**What I did:** When I saw the build failure, my FIRST instinct was to **REVERT** all the migrations back to `encoding/json` v1. I ran `git restore` on all 7 files. The user had to explicitly intervene with **"UPGRADE!"** to correct my approach.

**Why this was wrong:** The `go-auto-upgrade` tool was doing exactly what it was supposed to do — migrating to the newer, better API. The `go-error-family` v0.7.0 dependency also requires json/v2, so reverting the source files wouldn't have fixed the transitive dependency issue anyway. The correct fix was to enable `GOEXPERIMENT=jsonv2`, not to fight the upgrade.

**Lesson:** When a dependency upgrades to a new API, embrace it. The migration tool knows more about the intended direction than my reflex to "restore working state."

### 2. flake.nix indentation error

**What I did:** The first edit to `flake.nix` produced misaligned indentation:

```nix
            env.CGO_ENABLED = 0;
          env.GOEXPERIMENT = "jsonv2";   # ← wrong indentation
```

I caught it during review and fixed it, but it should have been correct the first time. The `multiedit` tool matched `env.CGO_ENABLED = 0;` but I didn't match the surrounding indentation context properly.

---

## E) WHAT WE SHOULD IMPROVE

### Process Improvements

1. **Don't fight automated migration tools** — understand their intent first, then enable the required environment
2. **Check transitive dependencies** before reverting source files — if `go-error-family` needs json/v2, reverting source files alone can't work
3. **Run `nix flake check`** as part of verification, not just `nix build`
4. **Verify binary execution** — compiling and testing isn't enough; the binary should actually run
5. **Check ALL files for consistency** — test files using v1 while production uses v2 is a split brain
6. **Update ALL user-facing docs** — README.md is the sales page; if it says `go build ./...` without `GOEXPERIMENT=jsonv2`, new contributors will fail

### Technical Improvements

7. **`GOEXPERIMENT=jsonv2` is experimental** — it could change or be removed in future Go versions. Consider documenting this risk and monitoring for Go release notes
8. **json/v2 behavioral differences** — v2 changes defaults for HTML escaping (`<`, `>`, `&`), unknown field handling (errors by default), and nil semantics. Tests pass but real-world JSON output may differ
9. **70 lint issues** — 50 `paralleltest` violations are a testing discipline problem; 10 `exhaustruct` and 2 `cyclop` are code quality debt
10. **3 test files still on v1** — `type_safe_enums_status_test.go`, `enum_macros_test.go`, `json_test.go` should be migrated for consistency

---

## F) NEXT 50 THINGS TO DO

### Migration Completion (P0 — do now)

1. Migrate `internal/domain/type_safe_enums_status_test.go` to `encoding/json/v2`
2. Migrate `internal/domain/enum_macros_test.go` to `encoding/json/v2`
3. Migrate `internal/format/json_test.go` to `encoding/json/v2`
4. Run `nix flake check` to validate the full flake
5. Verify binary execution: `GOEXPERIMENT=jsonv2 go run ./cmd/clean-wizard --help`
6. Verify binary execution: `GOEXPERIMENT=jsonv2 go run ./cmd/clean-wizard scan --dry-run --json`

### Documentation (P1)

7. Update README.md build instructions with `GOEXPERIMENT=jsonv2` requirement
8. Add a "Prerequisites" or "Requirements" section to README.md noting Go 1.26+ with jsonv2 experiment
9. Verify CHANGELOG.md mentions the json/v2 migration
10. Add `GOEXPERIMENT=jsonv2` note to any CONTRIBUTING.md or DEVELOPMENT.md if they exist
11. Update any CI/CD documentation or GitHub Actions workflows

### Code Quality (P2)

12. Fix 50 `paralleltest` lint violations in test files
13. Fix 10 `exhaustruct` lint violations
14. Fix 2 `cyclop` complexity violations (including `runScanCommand` at complexity 21)
15. Fix 5 `goconst` violations
16. Fix 2 `gochecknoglobals` violations
17. Fix 1 `exhaustive` violation
18. Verify `go generate` / `govalid-generate` works with json/v2
19. Run `govulncheck` to check for vulnerabilities in the new dependency tree

### Architecture (P3)

20. Address `internal/domain` god package (23 files) — split by concern
21. Address `internal/cleaner` flat structure (50+ files) — create sub-packages
22. Fix mutable package-level logger globals (`L`, `StdLogger`) — causes test race conditions
23. Migrate remaining cleaners from hardcoded defaults to user profile config
24. Add BDD tests for the 9 of 13 cleaners that have NO BDD tests
25. Evaluate whether `GOEXPERIMENT=jsonv2` should be replaced with a stable approach once Go finalizes json/v2

### Testing (P3)

26. Add integration test verifying JSON output format matches between v1 and v2 for all commands
27. Add test for HTML character handling in JSON output (json/v2 behavior change)
28. Add test for unknown field handling in config parsing (json/v2 errors by default)
29. Add E2E test for `clean --dry-run --json` full pipeline
30. Add E2E test for `scan --json` full pipeline
31. Add test for error JSON output format consistency
32. Add benchmark comparing json/v1 vs json/v2 marshaling performance

### Dependency Management (P3)

33. Pin `GOEXPERIMENT=jsonv2` compatibility — add a CI matrix test for future Go versions
34. Monitor `go-error-family` for v0.7.x stability issues
35. Evaluate if `golang.org/x/sys` bump (v0.46→v0.47) introduced any regressions
36. Check if `charmbracelet` package bumps (ultraviolet, charmtone) are stable
37. Run `go mod tidy` to ensure dependency tree is clean
38. Update `vendorHash` verification process — document how to update it when deps change

### Nix Flake (P3)

39. Consider adding `GOEXPERIMENT=jsonv2` to `checks.test` and `checks.go-vet` explicitly
40. Add a `checks.lint` target for golangci-lint in flake.nix
41. Consider adding `statix` check to flake.nix
42. Add `nix flake update` CI check to catch stale inputs
43. Document the vendor hash update workflow in AGENTS.md

### DevX (P4)

44. Add a `justfile` or `flake.nix` script target that sets `GOEXPERIMENT=jsonv2` automatically for ad-hoc commands
45. Consider a `.envrc` file for direnv users
46. Add a pre-commit hook that verifies `GOEXPERIMENT=jsonv2` is set
47. Create a `Makefile`-like convenience: `nix run .#test` should work without manual env setup
48. Add VSCode settings.json with `go.toolsEnvVars` setting `GOEXPERIMENT=jsonv2`
49. Add gopls configuration note — gopls needs `GOEXPERIMENT=jsonv2` for correct analysis
50. Evaluate if `encoding/json/v2` should be wrapped in a shared `internal/json` package to centralize the experiment dependency

---

## G) Top 2 Questions I Cannot Answer Myself

### 1. Should we stay on `GOEXPERIMENT=jsonv2` or wait for it to become stable?

`encoding/json/v2` is still experimental (behind `GOEXPERIMENT`). It may change API, get renamed, or behave differently before stabilization. The Go team has not committed to a release timeline for making it default. Do you want to:

- **(A)** Keep `GOEXPERIMENT=jsonv2` and track it until stable (risk: breaking changes in future Go versions)
- **(B)** Pin to `go-error-family` v0.6.1 (pre-json/v2) until json/v2 is stable (safe, but loses error classification improvements)
- **(C)** Fork/vendor the relevant parts of `go-error-family` v0.7.0 to use json/v1 (full control, maintenance burden)

### 2. Is the `vendorHash` change expected and stable?

The `vendorHash` changed from `sha256-PjErY8...` to `sha256-rs1zRW...` as part of the dependency bumps (`go-error-family` v0.7.0, `golang.org/x/sys` v0.47, `charmbracelet` updates). This was done by the `nix-flake-update` auto-fixer, not by me. I verified `nix build` passes, but:

- Is this hash deterministic across machines?
- Should we run `nix build --recreate-lock-file` or `go mod vendor` to verify?
- The `go.sum` also changed — is it consistent with `go.mod`?

I don't have enough context on your vendor management workflow to answer this confidently.
