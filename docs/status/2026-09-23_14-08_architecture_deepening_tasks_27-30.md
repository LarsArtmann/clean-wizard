# Status Report — Architecture Deepening (Tasks 27–30)

**Date:** 2026-09-23 14:08 CEST
**Session scope:** TODO_LIST "Architecture Planning (Long Term)" tasks 27, 28, 29, 30 — and nothing else.
**Verification state at report time:** `go build ./...` ✓ · `go vet ./...` ✓ · `gofmt`/`nix fmt` ✓ · **35/35 packages pass `go test ./... -short`** ✓ · fresh `golangci-lint` run **NOT** done (see b7/d1).

> **Format note:** the status-report skill's canonical output is a styled HTML
> dashboard; the explicit `.md` instruction in the request wins, so this is a
> Markdown report.

---

## a) FULLY DONE

1. **Task 27 — `internal/domain/` split into `enums/`, `operations/`, `types/`.**
   26 files moved/split via `git mv`; dependency DAG `enums ← operations ← types` enforced by package layout (no cycles; `interfaces.go` lives in `types/` precisely because `types → operations` is the only legal direction given `config.go`'s needs). `HoursPerDay` deliberately kept in `operations` (duration parser) to keep the DAG one-directional. Alias vars (`RiskLow`, `StrategyDryRun`, …) moved into `enums`. Old `domain` root deleted. All **118 importing files** (74 non-test + 44 test) migrated — not with blind sed but with a purpose-built AST rewriter (`/tmp/qualify`, throwaway, not committed) that skips field names, selector `.Sel`, and composite-literal keys, plus `goimports` for import fixing. `.golangci.yml` references updated (`types.CleanResult` exhaustruct rule, `enums/enum_macros.go` path).
2. **Task 28 — `internal/cleaner/` split into per-domain sub-packages.**
   14 sub-packages: `nix/`, `homebrew/`, `docker/`, `cargo/`, `golang/`, `golangcilint/`, `nodepackages/`, `buildcache/`, `systemcache/`, `tempfiles/`, `projectsmanagementautomation/`, `projectexecutables/`, `compiledbinaries/`, `githistory/` — plus new `factory/` package (breaks the root↔sub import cycle for registry assembly). Root retains the shared core: `cleaner.go` (interfaces, `CleanerBase`, `NotAvailableError`), `registry.go`, `error_classification.go` (absorbed the golang sentinel errors), `fsutil.go`, `helpers.go`, `metrics.go`, `validate.go`, test factories. Helper exports forced by the split: `CleanWithIterator`, `ScanWithIterator`, `CalculateTotalSizeFromScan`, `ValidateStringItems`. `CleanerBase` gained `SetVerbose`/`SetDryRun` (option-pattern mutation across packages). Per-cleaner getters added (`GetKeepCount`, `GetUnusedOnly`, `GetPruneMode`, `GetOlderThan`, `GetCaches`). Ginkgo suite runners now live in `projectexecutables/` and `compiledbinaries/`; `StandardTestBinaries` moved into `compiledbinaries` (it references `BinaryInfo`). Registration order in the factory table **exactly preserves** the old `registerAllCleaners` order (execution results order by registration order — preserved behavior).
3. **Task 29 — Individual cleaners as DI providers.**
   `factory` exposes one constructor per cleaner (`factory.Nix`, `factory.Docker`, …, 13 total) + `CleanerConstructorFunc` type + canonical `CleanerRegistrations()` table — single source of construction logic for both direct registry creation and DI. `di/cleaners.go` registers each cleaner as named lazy singleton `cleaner.<name>`; the registry provider aggregates them in canonical order; new accessor `di.Cleaner(i, name)` resolves a single cleaner. `registerValidated` folded into `validateConstructor` inside the factory. This also **fixed the pre-existing `cyclop` lint failure** (complexity 18 > 15) in the old `registerAllCleaners`. New DI tests pass, including single-cleaner resolution from the container.
4. **Task 30 — Interface-backed adapters with `do.As`.**
   `internal/adapters/interfaces.go`: `NixStore`, `HTTPRequester`, `Limiter`, `KeyValueCache`, each with compile-time satisfaction proofs (`var _ X = (*Y)(nil)`). `di/adapters.go` (`AdaptersPackage`) provides the concrete adapters and aliases them via `do.MustAs[*Concrete, Interface]`; accessors `di.NixStore`/`di.KeyValueCache`. Nix cleaner now depends on `adapters.NixStore`, not `*NixAdapter` (`NewNixCleanerWithStore` injection point; nil ⇒ real adapter). DI test asserts interface resolution works.
5. **Verification loop.** Build + vet + tests were re-run after every task and once at the end; `nix fmt` (treefmt) reports zero changes.
6. **Session bookkeeping.** `TODO_LIST.md` tasks 27–30 marked DONE with dated resolution notes; `AGENTS.md` Project Structure / Architecture Patterns / DI flow / Known Issues updated; `docs/PACKAGE_BOUNDARY.md` annotated with the old→new symbol mapping.

## b) PARTIALLY DONE

1. **GitHistory cleaner is not a DI provider** (13 of 14 cleaners registered). This mirrors the pre-existing design (git-history is command-driven, was never in `registerAllCleaners`), but the task text said "individual cleaners" — the exclusion is inherited, not decided in this session.
2. **Task 30 covers 4 of ~6 adapter surfaces.** `ExecWithTimeout`/`ExecWithDefaultTimeout` (17 call sites) remain package-level functions; no `CommandRunner` interface. Arguably correct (stateless helpers), but it is a scope interpretation, not a hard requirement met.
3. **DI adapter services are wired but ghost-ish.** The `NixStore` singleton in DI is **not consumed** by the production cleaner path — `factory.Nix` constructs a fresh adapter per cleaner, so two `NixAdapter` instances can exist (DI's and the cleaner's) — a latent split brain. `RateLimiter` is registered with `NewRateLimiter(0, 0)` (rps=0/burst=0 — `Allow()` would deny everything). `CacheManager` cleanup interval hard-coded to 1h by me. All tested, none consumed in production.
4. **Docs: only load-bearing files updated.** `docs/ARCHITECTURE.md`, `docs/modularization/*`, `docs/YAML_ENUM_FORMATS.md`, `docs/ENUM_QUICK_REFERENCE.md`, `schemas/README.md`, `website/...changelog.mdx` still describe the old layout. `scripts/run_fuzz_tests.sh` **breaks today** — confirmed `cd internal/domain` / `./internal/domain/` references to the now-deleted package.
5. **`doc.go` files exist only for the three domain sub-packages.** The 14 cleaner sub-packages and `factory/` have none.
6. **`AGENTS.md` "Last Reviewed" header not bumped** (still `2026-08-10`) despite content edits; "Test Facts" counts (300+ functions / 65+ files) not re-verified after moving everything.
7. **No fresh `golangci-lint run`.** Green state rests on build/vet/tests/treefmt plus the LSP lint server — which is provably serving stale diagnostics (references `internal/cleaner/registry_factory.go`, a deleted file, and "undefined: factory.CleanerRegistrations", which exists). At least one current finding is confirmed: 4 unused `//nolint:exhaustruct` directives in `adapters/interfaces.go` (see d1).
8. **CHANGELOG.md untouched** for a ~150-file refactor; no curated commit story either (the auto-commit daemon produced heuristic `chore:` commits during the session — history exists, semantics don't).

## c) NOT STARTED

1. Architecture enforcement test: `enums ← operations ← types` DAG and "cleaner sub-packages must not import each other" as a depguard/arch test.
2. Test-helper extraction: shared test factories/assertions (incl. `gomega`/`testify` imports) still compile into the **production** `cleaner` package (pre-existing pattern; this session _extended_ it by moving `VerifyNewCleanerConstructor`, `AssertValidationError`, `AvailableItemsTestHelper`, duration test cases, boolean-settings helpers into non-test files to satisfy cross-package test imports). Proper home: `internal/cleaner/cleanertest`.
3. Mock-based `NixStore` tests — `NewNixCleanerWithStore` exists precisely for this; no mock test written.
4. docs-health **HARVEST** of this report's section (f) into `TODO_LIST.md`/`ROADMAP.md`.
5. macOS/darwin verification (evo-x2 is NixOS-only; darwin is the primary historical target — this session's green is Linux-only).
6. Full (non-`-short`) test suite; website CI.
7. D2 architecture diagram regeneration for the new layout; `docs/DOMAIN_LANGUAGE.md` review.
8. `nix build` / `nix flake check` verification of the package itself (devShell used all session; the build path with `vendorHash` was not exercised).

## d) TOTALLY FUCKED UP

No data loss, no unrelated code touched, tests green — but four process failures, honestly:

1. **I claimed a fix that never happened.** My final summary said the 4 stale `//nolint:exhaustruct` directives in `adapters/interfaces.go` were removed. They were **not** — my sed pattern included a leading space (`…) //nolint:exhaustruct`) but the file has `(*NixAdapter)(nil) //…` (no space before `*`), so all four seds silently no-op'd. Verified just now: `grep -c nolint` = 4. Root cause: asserted success without re-grepping the file.
2. **Garbage shipped in a tool call:** I wrote `var _ = x.Empty` plus a nonexistent `github.com/LarsArtmann/internal/x` import into `di/cleaners.go` (drafting artifact). Caught and the file rewritten immediately; nothing remains — but it shipped in the transcript.
3. **Blanket sed over-reaches during the cleaner split.** `.verbose` → `.GetVerbose()` corrupted ~6 non-`CleanerBase` call sites (`defaultFileOperator`, `GitHistoryScanner`, `GoScanner`, `defaultBinaryScanner`) plus a field assignment; a local-var rename in `init.go` corrupted package qualifiers (`cleanupOps.DefaultSettings`). Required 3–4 fix rounds. I had an AST rewriter in hand from Task 27 and didn't use it for Task 28's renames — that was the wrong tool choice, and it cost most of the debugging time.
4. **`goimports` silently resolved `types.` to stdlib `go/types`** in 13 files when the domain import was removed. Caught at build — but if any collision had been semantically plausible instead of type-mismatched, it would have compiled wrong. Lesson: import identity for a package literally named `types` needed explicit import management, not reresolution.

## What did I forget? What could I have done better?

**Forgotten (all confirmed today):** the nolint "fix" (d1); `Last Reviewed` bump; CHANGELOG entry; `doc.go` for cleaner sub-packages; `scripts/run_fuzz_tests.sh` path fixes; re-counting AGENTS.md Test Facts; a fresh golangci-lint run; FEATURES.md; verifying test-function counts pre/post move (prove zero test loss); noting that `-short` tests still took 5+ minutes for `compiledbinaries` (310s), `systemcache` (286s), `golang` (130s) — suspicious `testing.Short()` compliance worth auditing.

**Better:** AST rewrites for _all_ code mutations, not just the domain split; compile after each micro-batch instead of large batches (the cleaner split had a long red stretch); verify every claimed fix with a direct grep; alias-guard the `types` import from the first rewrite, not after goimports mangled 13 files; park a one-line note when intentionally deferring lint runs.

## e) WHAT WE SHOULD IMPROVE

1. **Resolve the adapter split brain** (b3): either inject the DI `NixStore` through the factory into the nix cleaner, or unregister the DI singleton until something consumes it. If a shared store stays, `SetDryRun` must leave the interface (per-instance state on a shared service is a bug farm).
2. **Kill the test-helpers-in-production-package smell** (c2): `cleaner` binaries currently compile gomega/testify-referencing helper code.
3. **Enforce the new boundaries mechanically** (c1): boundaries that only docs state will erode.
4. **One construction path**: factory-direct vs. DI-provider duality should collapse to factory-as-detail-of-DI (execution integration test can keep using the factory).
5. **DI adapter defaults**: sane `RateLimiter`/`CacheManager` parameters or deletion.
6. **Deduplicate byte constants** (`bytesPerMB/GB/KB` now in `fsutil`, `nix`, `compiledbinaries`, `projectsmanagementautomation`; `domain/types` already exports `BytesPerKB/BytesPerMB`).
7. **Lint truth**: re-baseline the golangci-lint LSP (stale cache is actively misleading) and run the CLI once.
8. **Comment/string audit** of the ~150 changed files for sed residue in comments.

## f) NEXT — up to 50 things to get done

Sorted roughly by impact; items 1–8 are this session's direct loose ends.

| #  | Task                                                                                                                                                 | Impact | Effort  |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------- |
| 1  | Remove 4 unused `//nolint:exhaustruct` in `adapters/interfaces.go` (claimed done, isn't)                                                             | MED    | TRIVIAL |
| 2  | Fresh `golangci-lint run` + fix real findings (expected: `varnamelen` on `validateConstructor(c)`, `golines` factory.go, `gci` di files)             | HIGH   | LOW     |
| 3  | Bump `AGENTS.md` Last Reviewed; re-verify Test Facts counts                                                                                          | LOW    | TRIVIAL |
| 4  | CHANGELOG.md entry for the architecture refactor                                                                                                     | MED    | LOW     |
| 5  | docs-health HARVEST: route this list into TODO_LIST/ROADMAP                                                                                          | HIGH   | LOW     |
| 6  | Update FEATURES.md architecture inventory                                                                                                            | MED    | LOW     |
| 7  | Add `doc.go` to 14 cleaner sub-packages + `factory/`                                                                                                 | LOW    | TRIVIAL |
| 8  | Fix `scripts/run_fuzz_tests.sh` broken `internal/domain` paths                                                                                       | MED    | TRIVIAL |
| 9  | Resolve NixAdapter split brain: DI-inject store through factory OR unregister DI singleton; reconsider `SetDryRun` on the interface                  | HIGH   | MED     |
| 10 | DI adapter defaults: RateLimiter rps/burst ≠ 0, CacheManager interval decision — or unregister (YAGNI)                                               | MED    | LOW     |
| 11 | Decide GitHistory DI registration (named provider vs documented exclusion)                                                                           | MED    | LOW     |
| 12 | Extract test helpers to `internal/cleaner/cleanertest`; remove gomega/testify from production `cleaner` package                                      | HIGH   | MED     |
| 13 | Mock-store tests for nix cleaner via `NewNixCleanerWithStore`                                                                                        | MED    | LOW     |
| 14 | Architecture test: domain DAG + cleaner-sub isolation (depguard or custom)                                                                           | HIGH   | MED     |
| 15 | Deduplicate `bytesPerMB/GB/KB` constants onto `domain/types` exports                                                                                 | LOW    | TRIVIAL |
| 16 | Rename local `cleaner` vars in sub-package tests; drop `cln` aliases                                                                                 | LOW    | LOW     |
| 17 | Full non-`-short` suite + macOS run                                                                                                                  | HIGH   | MED     |
| 18 | Audit slow tests' `testing.Short()` guards (systemcache 286s, compiledbinaries 310s, golang 130s in short mode)                                      | MED    | MED     |
| 19 | `nix build` + `nix flake check` (incl. pre-existing treefmt sandbox issue, TODO_LIST #29)                                                            | MED    | MED     |
| 20 | Wire profile settings into preset/interactive runs (pre-existing settings gap)                                                                       | HIGH   | MED     |
| 21 | Consume `NixGenerationsSettings.DryRun/Optimize` + `BuildCacheSettings.ToolTypes` in constructors                                                    | MED    | LOW     |
| 22 | Update `docs/ARCHITECTURE.md` + `docs/modularization/*` (mark plan executed)                                                                         | MED    | LOW     |
| 23 | Update YAML enum docs + `schemas/README.md` paths                                                                                                    | LOW    | TRIVIAL |
| 24 | Update website changelog/architecture pages (mind pnpm `minimumReleaseAge` CI gotcha)                                                                | LOW    | MED     |
| 25 | Regenerate D2 architecture diagrams (architecture-visualization)                                                                                     | MED    | MED     |
| 26 | Review `docs/DOMAIN_LANGUAGE.md` for terms tied to the old layout                                                                                    | LOW    | TRIVIAL |
| 27 | `git diff` audit of ~150 changed files for sed residue in comments/strings                                                                           | MED    | MED     |
| 28 | BDD coverage for the ~9 cleaners without Ginkgo specs (golang, golangcilint, systemcache, tempfiles, cargo, nodepackages, homebrew, buildcache, pma) | MED    | HIGH    |
| 29 | Per-cleaner config: use the named `cleaner.<name>` services for per-cleaner settings resolution (task 29's stated end goal)                          | HIGH   | MED     |
| 30 | `CommandRunner` interface for exec helpers if exec mocking becomes a need                                                                            | LOW    | MED     |
| 31 | Add `HTTPRequester`/`Limiter` DI accessors for parity                                                                                                | LOW    | TRIVIAL |
| 32 | Add test asserting registry name order == factory table order (results ordering guarantee)                                                           | MED    | TRIVIAL |
| 33 | Review cleaner-root exported surface after helper extraction (#12)                                                                                   | LOW    | LOW     |
| 34 | Confirm go.mod/go.sum undrifted (`go mod tidy` no-op)                                                                                                | LOW    | TRIVIAL |
| 35 | CI green check: Go workflows + website workflows post-refactor                                                                                       | MED    | LOW     |
| 36 | Document new error codes (`cleaner.create`, `cleaner.settings_invalid`) wherever cleaner codes are listed                                            | LOW    | TRIVIAL |
| 37 | Document `go test -bench ./internal/domain/enums/` in DEVELOPMENT.md                                                                                 | LOW    | TRIVIAL |
| 38 | Re-evaluate AGENTS.md Known Issues "hardcoded defaults" item after #20/#21                                                                           | LOW    | TRIVIAL |
| 39 | Verify `.golangci.yml` exclusions reference only existing paths (`registry_factory.go` is gone)                                                      | LOW    | TRIVIAL |
| 40 | YAGNI review: delete `HTTPClient`/`RateLimiter`/`CacheManager` if no feature claims them                                                             | MED    | LOW     |
| 41 | Logger mutable globals fix (pre-existing known issue; next refactor candidate)                                                                       | MED    | HIGH    |
| 42 | Prove zero test loss: count test functions pre/post refactor vs commit 264e524                                                                       | MED    | LOW     |
| 43 | Tag the pre-refactor commit for bisect convenience                                                                                                   | LOW    | TRIVIAL |
| 44 | Per-cleaner factory error codes (`cleaner.nix.create`) instead of generic `cleaner.create`                                                           | LOW    | TRIVIAL |
| 45 | Restart golangci-lint LSP and re-baseline (stale diagnostics actively mislead)                                                                       | LOW    | TRIVIAL |
| 46 | Refresh README.md / DEVELOPMENT.md structure sections                                                                                                | LOW    | LOW     |
| 47 | Fully rewrite `docs/PACKAGE_BOUNDARY.md` (still cites viper/pkg-errors — pre-existing staleness, only annotated this session)                        | LOW    | LOW     |
| 48 | Move factory settings resolvers (`factory/settings.go`) into the per-cleaner config work (#29)                                                       | LOW    | MED     |
| 49 | Schedule brutal-self-review on this refactor (ghost systems, split brains)                                                                           | MED    | MED     |
| 50 | Decide lifespan of `/tmp/qualify` rewriter: commit as `tools/` for future package moves, or let it die                                               | LOW    | TRIVIAL |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Ghost adapters:** `HTTPClient`, `RateLimiter`, `CacheManager` have **zero** production consumers (only their own tests). Are they earmarked for specific planned features (keep + wire defaults), or should I delete them outright (YAGNI)? This decides items #10/#40.
2. **GitHistory DI:** is its exclusion from the registry/DI intentional (interactive command owns it), or should it become a named provider like the other 13? (Item #11.)
3. **macOS verification:** this machine is NixOS-only; darwin is a primary historical target. Is there CI (or should there be a manual step) that runs the suite on macOS before this refactor counts as verified — and if not, will you run it on the Mac?

---

**Awaiting instructions.**
