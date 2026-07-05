# Status Report: Build Failures Fixed

**Date:** 2026-07-06 01:19
**Session Focus:** Fix 4 BuildFlow failures (nix-build, nix-build-verify, nix-hash-fix, test-race)

---

## Executive Summary

All 4 BuildFlow failures are now **FIXED**. The root causes were:

1. **flake.nix referenced non-existent `./pkg` directory** → nix-build, nix-build-verify, nix-hash-fix all failed
2. **Data race in logger tests** → test-race failed (parallel tests mutating global state)
3. **Untracked new packages missing from flake source** → nix couldn't find `internal/di` and `internal/execution`
4. **Stale vendorHash** → dependencies changed during DI/workflow migration but hash wasn't updated

**`nix build` now succeeds. `go test -race ./...` now passes all 20 packages.**

---

## a) FULLY DONE

| Fix                          | Root Cause                                                                                                 | Resolution                                                                                |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| nix-build (pkg path)         | `flake.nix` fileset included `./pkg` which doesn't exist                                                   | Removed `./pkg` from `lib.fileset.unions`                                                 |
| test-race (logger)           | All logger tests called `t.Parallel()` while mutating package globals `L` and `StdLogger`                  | Removed all `t.Parallel()` from `internal/logger/logger_test.go` (17 instances)           |
| nix-build (missing packages) | `internal/di/` and `internal/execution/` were untracked by git → nix flakes only include git-tracked files | `git add internal/di/ internal/execution/`                                                |
| nix-build (stale vendorHash) | New deps (samber/do v2, Azure/go-workflow) added during migration but vendorHash was stale                 | Updated to `sha256-F1o9yOTg3RiUApVGlJSRsYcpi/Y3J4GCPrCQ9vyorLs=` via `lib.fakeHash` trick |
| go.mod/go.sum sync           | `go mod tidy` hadn't been run after adding new dependencies                                                | Ran `go mod tidy` — added samber/do, go-workflow, and updated charm.land indirect deps    |

### Verification Results

```
go build ./...          → PASS (0 errors)
go test -race ./...     → PASS (20/20 packages ok)
nix build --no-link     → PASS (exit 0)
```

---

## b) PARTIALLY DONE

### DI + Workflow Migration (pre-existing, not this session's work)

The migration from the old `cleaner_implementations.go` to the new DI container (`internal/di/`) and workflow engine (`internal/execution/`) is **functional and compiles** but:

- The migration doc `docs/status/2026-07-06_00-35_DI-WORKFLOW-MIGRATION.md` exists but was authored before this session
- Files are now staged in git but **not committed**
- LSP shows stale diagnostics for `internal/execution/builder.go` (gopls cache — the actual code compiles fine)

---

## c) NOT STARTED

- **Committing the work** — All changes are staged/unstaged but nothing has been committed this session
- **Running full BuildFlow** — We fixed the 4 failures but didn't re-run `buildflow --fix --build-mode=full` to confirm 40/40 green
- **Deduplication work** — The user ran `art-dupl --semantic` showing 20 clone groups (at threshold 3) / 9 clone groups (at threshold 4). These were observed but not addressed

---

## d) TOTALLY FUCKED UP

Nothing. All 4 failures are cleanly resolved. No collateral damage.

---

## e) WHAT WE SHOULD IMPROVE

1. **Git-track new directories immediately** — The `internal/di/` and `internal/execution/` directories were created as part of the DI migration but never `git add`-ed. Nix flakes silently exclude untracked files, causing cryptic "no required module provides package" errors. **Lesson:** Always `git add` new packages before testing with nix.

2. **Logger global state is a design smell** — The logger package uses mutable package-level globals (`L`, `StdLogger`) that are written by `Init()`/`InitWithLevel()` and read by every log call. This makes the tests inherently non-parallelizable. A proper fix would use a `sync.RWMutex` or pass the logger via DI/context rather than globals.

3. **flake.nix fileset should validate paths** — `lib.fileset.unions` fails at build time (not eval time) if a path doesn't exist. Using `lib.fileset.maybeMissing` for optional directories would prevent this class of error.

4. **`go mod tidy` should be part of the migration workflow** — The DI migration added imports to `samber/do/v2` and `Azure/go-workflow` but `go mod tidy` was never run, leaving go.mod/go.sum out of sync.

5. **vendorHash updates should be automated** — The `lib.fakeHash` → real hash dance is manual and error-prone. BuildFlow's `nix-hash-fix` step should handle this but couldn't because the error was a missing path, not a hash mismatch.

---

## f) Up to 25 Things We Should Get Done Next

### High Priority (blocks green CI)

1. **Commit all staged changes** — The DI/workflow migration + fixes are staged but uncommitted
2. **Run full `buildflow --fix --build-mode=full`** to confirm 40/40 steps green
3. **Run `nix flake check`** to verify all flake checks pass
4. **Address the 9 clone groups** found by `art-dupl --semantic -t 4` (threshold 4, meaningful duplication)

### DI/Workflow Migration Polish

5. **Wire remaining 9 cleaners into BDD tests** — AGENTS.md notes 9 of 13 cleaners have NO BDD tests
6. **Add CLI command tests** — AGENTS.md notes "CLI command tests are missing entirely"
7. **Remove dormant result types** — `result.FlowBuilder`/`BranchFlow`/`ParallelFlow` are dormant (replaced by go-workflow)
8. **Consolidate 4 error packages** with overlapping responsibilities (split brain noted in AGENTS.md)

### Code Quality (from AGENTS.md known issues)

9. **Split `internal/domain/` god package** — 20+ files in one package
10. **Add sub-packages to `internal/cleaner/`** — 50+ files flat, no structure
11. **Fix ~40 `err113` lint violations** (dynamic errors via `fmt.Errorf`)
12. **Split 15 source files over 350 lines** into smaller files

### Logger Improvements

13. **Make logger thread-safe** — Replace globals with DI-injected logger or add `sync.RWMutex`
14. **Re-enable `t.Parallel()` on logger tests** once globals are eliminated

### Deduplication (from art-dupl output)

15. **Extract common cleaner test interface** — `test_interfaces.go` has 6+ clone sites
16. **Extract common cleaner factory pattern** — `test_factories.go` has 4+ clone sites
17. **Consolidate `Cleaner` interface definitions** — defined in `cleaner.go`, `test_assertions.go`, `test_interfaces.go`, `domain/interfaces.go`

### Nix/Build

18. **Use `lib.fileset.maybeMissing` for optional directories** in flake.nix
19. **Add a CI check that `go mod tidy` is clean** (diff-free after running)
20. **Add a pre-commit hook to `git add` new directories** before nix builds

### Testing

21. **Add race detection to CI** — the logger race was latent; CI should catch this
22. **Add integration tests for the DI container** — only unit tests exist
23. **Add integration tests for the workflow engine** — `execution_test.go` exists but should cover parallel execution paths

### Documentation

24. **Update AGENTS.md** with the new DI/workflow architecture details (partially done)
25. **Update FEATURES.md** to reflect the migration from sequential execution to DAG-based workflow

---

## g) Top #1 Question I Cannot Figure Out Myself

**The DI/workflow migration was in-progress before this session — is it complete and ready to commit, or is there unfinished work in `internal/di/` or `internal/execution/` that I should be aware of?**

The code compiles, all tests pass, and nix builds successfully. But I don't know if the migration is functionally complete (all cleaners wired through the new DI container and workflow engine) or if there are remaining old-path code that should be removed. Specifically:

- `cmd/clean-wizard/commands/cleaner_implementations.go` was deleted — is everything it provided now covered by the new `internal/di/` + `internal/execution/` packages?
- Are there any other commands still using the old direct-registry approach instead of going through DI?

---

## Files Changed This Session

| File                             | Change                                             | Reason                                                     |
| -------------------------------- | -------------------------------------------------- | ---------------------------------------------------------- |
| `flake.nix`                      | Removed `./pkg` from fileset; updated `vendorHash` | Fix nix-build failures                                     |
| `internal/logger/logger_test.go` | Removed all `t.Parallel()` (17 instances)          | Fix data race in tests                                     |
| `go.mod` / `go.sum`              | `go mod tidy` synced deps                          | Added samber/do, go-workflow; updated charm.land indirects |
| `internal/di/` (staged)          | `git add` — 6 files                                | Nix flakes require git-tracked files                       |
| `internal/execution/` (staged)   | `git add` — 7 files                                | Nix flakes require git-tracked files                       |
