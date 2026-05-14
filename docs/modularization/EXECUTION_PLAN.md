# Execution Plan — Clean Wizard Modularization

**Date:** 2026-05-14 | **Status:** Draft | **Depends on:** [PROPOSAL.md](./PROPOSAL.md)

---

## Overview

20 tasks organized in 6 phases. Each task leaves the project in a buildable, testable state and is independently revertable via a single commit.

---

## Phase A: Pre-Modularization Cleanup (1% → 51% impact)

These tasks fix structural problems that would complicate the module split. They are the highest-impact items because they eliminate coupling _before_ boundaries are drawn.

---

### Task A1: Delete dead code (`internal/api`)

**Impact:** 1% → 51% | **Effort:** 5 min | **Depends on:** None

`internal/api/` (3 files, ~650 lines) is never imported by any file. Delete it.

**Steps:**

1. `trash internal/api/`
2. `go build ./...`
3. `go test ./...`

**Verification:** Build + tests pass. No import errors.

**Rollback:** `git checkout -- internal/api/`

---

### Task A2: Consolidate error packages (3 → 1)

**Impact:** 1% → 51% | **Effort:** 30 min | **Depends on:** None

Three overlapping error packages:

- `internal/errors` — 7 simple constructors
- `internal/pkg/errors` — structured `CleanWizardError` with builder
- `pkg/errors` — minimal `BaseError` + `New()`

**Steps:**

1. Choose `internal/pkg/errors` as the canonical package (most complete)
2. Move useful constructors from `internal/errors` into `internal/pkg/errors`
3. Move `BaseError` from `pkg/errors` into `internal/pkg/errors` (or delete if unused)
4. Update all import paths across the codebase
5. Delete `internal/errors/` and `pkg/errors/`
6. `go build ./...` && `go test ./...`

**Verification:** All references updated. Build passes. No duplicate error constructors.

**Rollback:** `git revert HEAD`

---

### Task A3: Extract test helpers from `internal/cleaner/`

**Impact:** 1% → 51% | **Effort:** 30 min | **Depends on:** None

Move non-`_test.go` test helper files out of the production package:

| File                     | Lines | Destination                     |
| ------------------------ | ----- | ------------------------------- |
| `test_interfaces.go`     | ~120  | `internal/cleaner/testhelpers/` |
| `test_factories.go`      | 94    | `internal/cleaner/testhelpers/` |
| `test_helpers.go`        | 213   | `internal/cleaner/testhelpers/` |
| `test_assertions.go`     | 323   | `internal/cleaner/testhelpers/` |
| `testing_helpers.go`     | 88    | `internal/cleaner/testhelpers/` |
| `ginkgo_test_helpers.go` | 304   | `internal/cleaner/testhelpers/` |

**Steps:**

1. Create `internal/cleaner/testhelpers/` package
2. Move files, update package declaration from `cleaner` to `testhelpers`
3. Export any symbols that test files reference (they're in the same package currently)
4. Update all `_test.go` files to import the new package
5. `go build ./...` && `go test ./...`

**Result:** `internal/cleaner` no longer imports `testify`, `ginkgo`, or `gomega` in production code.

**Verification:** `grep -r 'testify\|ginkgo\|gomega' internal/cleaner/*.go` returns nothing (only `_test.go` files).

**Rollback:** `git revert HEAD`

---

### Task A4: Remove production testify imports from `internal/domain`

**Impact:** 1% → 51% | **Effort:** 15 min | **Depends on:** A3

`internal/domain` imports `stretchr/testify` in non-test files (test helpers).

**Steps:**

1. Identify non-`_test.go` files in `internal/domain/` that import `testify`
2. Either move the helpers to a `testhelpers` sub-package, or convert to `testing.T` + manual assertions
3. `go build ./...` && `go test ./...`

**Verification:** `grep -r 'testify' internal/domain/*.go` (excluding `_test.go`) returns nothing.

**Rollback:** `git revert HEAD`

---

## Phase B: Extract Core Module (4% → 64% impact)

This is the foundational extraction. Everything depends on core.

---

### Task B1: Create core module skeleton

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** A2, A4

**Steps:**

1. Create `core/` directory
2. Create `core/go.mod`:
   ```
   module github.com/LarsArtmann/clean-wizard/core
   go 1.26.2
   require gopkg.in/yaml.v3 v3.0.1
   ```
3. Create `go.work` at repo root:
   ```
   go 1.26.2
   use (
       .
       ./core
   )
   ```

**Verification:** `go work sync` succeeds. Root build still works.

**Rollback:** Delete `core/go.mod` and `go.work`.

---

### Task B2: Move `internal/domain` → `core/domain`

**Impact:** 4% → 64% | **Effort:** 20 min | **Depends on:** B1

**Steps:**

1. `mkdir -p core/domain`
2. Move all files from `internal/domain/` to `core/domain/`
3. Update package declarations (already `domain`, no change needed)
4. Find-and-replace import paths across entire codebase:
   - `github.com/LarsArtmann/clean-wizard/internal/domain` → `github.com/LarsArtmann/clean-wizard/core/domain`
5. `go mod tidy` in root and core
6. `go build ./...` && `go test ./...`

**Verification:** All 29 packages compile. Tests pass. Domain types resolve from new path.

**Rollback:** `git revert HEAD`

---

### Task B3: Move `internal/result` → `core/result`

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** B1

**Steps:**

1. `mkdir -p core/result`
2. Move files from `internal/result/` to `core/result/`
3. Find-and-replace import paths
4. `go mod tidy` in both modules
5. `go build ./...` && `go test ./...`

**Verification:** Build passes. `result.Result[T]` resolves from new path.

**Rollback:** `git revert HEAD`

---

### Task B4: Move error package → `core/errors`

**Impact:** 4% → 64% | **Effort:** 15 min | **Depends on:** B1, A2

**Steps:**

1. Move `internal/pkg/errors/` to `core/errors/`
2. Update package declaration from `errors` to `errors` (same name, no change)
3. Find-and-replace import paths:
   - `github.com/LarsArtmann/clean-wizard/internal/pkg/errors` → `github.com/LarsArtmann/clean-wizard/core/errors`
4. `go mod tidy`
5. `go build ./...` && `go test ./...`

**Verification:** Build passes. Error types resolve from new path.

**Rollback:** `git revert HEAD`

---

### Task B5: Move `internal/format` → `core/format`

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** B1

**Steps:**

1. Move `internal/format/` to `core/format/`
2. Update import paths
3. Add `github.com/dustin/go-humanize` to `core/go.mod`
4. `go mod tidy` in both modules
5. `go build ./...` && `go test ./...`

**Verification:** Build passes. Formatting functions resolve from new path.

**Rollback:** `git revert HEAD`

---

### Task B6: Move `internal/shared/utils/validation` → `core/validation`

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** B3

**Steps:**

1. Move `internal/shared/utils/validation/` to `core/validation/`
2. Update import paths
3. `go mod tidy`
4. `go build ./...` && `go test ./...`

**Verification:** Build passes.

**Rollback:** `git revert HEAD`

---

### Task B7: Move `internal/testing` → `core/testing`

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** B1

**Steps:**

1. Move `internal/testing/` to `core/testing/`
2. Update import paths
3. `go mod tidy`
4. `go build ./...` && `go test ./...`

**Verification:** Build passes.

**Rollback:** `git revert HEAD`

---

### Task B8: Create `go.work` and verify core module

**Impact:** 4% → 64% | **Effort:** 10 min | **Depends on:** B1-B7

**Steps:**

1. Verify `go.work` includes both `.` and `./core`
2. `go work sync`
3. `go build ./...` at root — all packages compile
4. `go test ./...` at root — all tests pass
5. `go build ./...` in `core/` — core builds independently
6. `go test ./...` in `core/` — core tests pass independently
7. Verify `core/go.mod` has no internal dependencies (only external deps)

**Verification:**

- `grep 'github.com/LarsArtmann' core/go.mod` returns nothing (no replace/require of self)
- Root builds with workspace
- Core builds in isolation

**Rollback:** Remove `go.work`, revert imports.

---

## Phase C: Extract Adapters Module (4% → 64% impact)

---

### Task C1: Create adapters module and move files

**Impact:** 4% → 64% | **Effort:** 20 min | **Depends on:** B8

**Steps:**

1. Create `adapters/` directory and `adapters/go.mod`:
   ```
   module github.com/LarsArtmann/clean-wizard/adapters
   go 1.26.2
   require (
       github.com/LarsArtmann/clean-wizard/core v0.0.0
       github.com/caarlos0/env/v6 v6.10.1
       github.com/go-resty/resty/v2 v2.17.2
       github.com/maypok86/otter/v2 v2.3.0
       golang.org/x/time v0.15.0
   )
   ```
2. Move `internal/adapters/` files to `adapters/`
3. Update import paths:
   - `github.com/LarsArtmann/clean-wizard/internal/adapters` → `github.com/LarsArtmann/clean-wizard/adapters`
   - `github.com/LarsArtmann/clean-wizard/internal/domain` → `github.com/LarsArtmann/clean-wizard/core/domain`
   - `github.com/LarsArtmann/clean-wizard/internal/result` → `github.com/LarsArtmann/clean-wizard/core/result`
   - `github.com/LarsArtmann/clean-wizard/internal/conversions` → Update to new conversions location
4. Update `go.work` to include `./adapters`
5. `go work sync`
6. `go mod tidy` in all three modules
7. `go build ./...` && `go test ./...`

**Verification:** Adapters build independently. Root builds with workspace.

**Rollback:** `git revert HEAD`

---

## Phase D: Extract Config Module (4% → 64% impact)

---

### Task D1: Remove logger dependency from config

**Impact:** 4% → 64% | **Effort:** 15 min | **Depends on:** None (can run in parallel)

**Steps:**

1. In `internal/config/config.go`, replace `logger.Info(...)` / `logger.Error(...)` with `slog.Info(...)` / `slog.Error(...)`
2. Remove `internal/logger` import from config files
3. Add `log/slog` import
4. `go build ./...` && `go test ./...`

**Verification:** Config no longer imports logger. Tests pass.

**Rollback:** `git revert HEAD`

---

### Task D2: Create config module and move files

**Impact:** 4% → 64% | **Effort:** 25 min | **Depends on:** B8, D1

**Steps:**

1. Create `config/` directory and `config/go.mod`:
   ```
   module github.com/LarsArtmann/clean-wizard/config
   go 1.26.2
   require (
       github.com/LarsArtmann/clean-wizard/core v0.0.0
       github.com/knadh/koanf/v2 v2.3.4
       github.com/knadh/koanf/parsers/yaml v1.1.0
       github.com/knadh/koanf/providers/file v1.2.1
   )
   ```
2. Move all `internal/config/` files to `config/`
3. Update import paths in config files:
   - `internal/domain` → `core/domain`
   - `internal/pkg/errors` → `core/errors`
   - `internal/shared/utils/strings` → update path
4. Update import paths in all other packages that import config
5. Update `go.work` to include `./config`
6. `go work sync`
7. `go mod tidy` in all modules
8. `go build ./...` && `go test ./...`

**Verification:** Config builds independently. Root builds with workspace.

**Rollback:** `git revert HEAD`

---

## Phase E: Extract Cleaners Module (4% → 64% impact)

---

### Task E1: Move conversions to cleaners

**Impact:** 4% → 64% | **Effort:** 15 min | **Depends on:** B8

`internal/conversions` is primarily used by cleaners (16 files). Move it into the cleaners module.

**Steps:**

1. Note: conversions also imported by `internal/adapters/nix.go` and `internal/middleware/validation.go`
2. Decision: Move conversions to `cleaners/conversions/` sub-package
3. For adapters and middleware, they use `ToCleanResult()` and similar — these should be inlined or the callers updated
4. Alternative: keep conversions in `core` and accept the `result` dependency (cleaner boundary)
5. **Revised approach:** Keep `conversions` in core. It's a pure function package with no external deps beyond `domain` and `result`. Both are in core already.
6. Update `internal/conversions` → `core/conversions`
7. `go build ./...` && `go test ./...`

**Verification:** All packages that import conversions resolve from new path.

**Rollback:** `git revert HEAD`

---

### Task E2: Create cleaners module and move files

**Impact:** 4% → 64% | **Effort:** 45 min | **Depends on:** C1, E1

This is the largest single task — moving ~40 files.

**Steps:**

1. Create `cleaners/` directory and `cleaners/go.mod`:
   ```
   module github.com/LarsArtmann/clean-wizard/cleaners
   go 1.26.2
   require (
       github.com/LarsArtmann/clean-wizard/core v0.0.0
       github.com/LarsArtmann/clean-wizard/adapters v0.0.0
       github.com/cockroachdb/errors v1.12.0
       golang.org/x/sys v0.42.0
   )
   ```
2. Move all cleaner implementation files from `internal/cleaner/` to `cleaners/`
3. Move `internal/cleaner/testhelpers/` to `cleaners/testhelpers/`
4. Update all import paths:
   - `internal/cleaner` → `cleaners` (or `github.com/LarsArtmann/clean-wizard/cleaners`)
   - `internal/domain` → `core/domain`
   - `internal/result` → `core/result`
   - `internal/errors` / `internal/pkg/errors` → `core/errors`
   - `internal/adapters` → `adapters`
   - `internal/conversions` → `core/conversions`
   - `internal/format` → `core/format`
   - `internal/shared/utils/*` → update paths
   - `internal/testing` → `core/testing`
5. Update `go.work` to include `./cleaners`
6. `go work sync`
7. `go mod tidy` in all modules
8. `go build ./...` && `go test ./...`

**Verification:**

- Cleaners build independently: `cd cleaners && go build ./...`
- Cleaner tests pass independently: `cd cleaners && go test ./...`
- Root builds with workspace
- Root tests pass

**Rollback:** `git revert HEAD`

---

## Phase F: Finalize (20% → 80% impact)

---

### Task F1: Update root module imports

**Impact:** 20% → 80% | **Effort:** 20 min | **Depends on:** E2

**Steps:**

1. Update root `go.mod` — remove direct dependencies now in sub-modules
2. Add sub-module dependencies to root `go.mod`:
   ```
   require (
       github.com/LarsArtmann/clean-wizard/core v0.0.0
       github.com/LarsArtmann/clean-wizard/cleaners v0.0.0
       github.com/LarsArtmann/clean-wizard/adapters v0.0.0
       github.com/LarsArtmann/clean-wizard/config v0.0.0
   )
   ```
3. Update all remaining import paths in:
   - `cmd/clean-wizard/main.go`
   - `cmd/clean-wizard/commands/`
   - `internal/logger/`
   - `internal/testhelper/`
   - `internal/version/`
   - `internal/shared/utils/` (remaining: config, gitfilterrepo, strings, schema, fileutil)
   - `tests/bdd/`
   - `tests/benchmark/`
   - `test/`
4. `go mod tidy` in root
5. `go build ./...` && `go test ./...`

**Verification:** Root module has only TUI + Cobra direct deps. All other deps come from sub-modules.

**Rollback:** `git revert HEAD`

---

### Task F2: Clean up empty directories

**Impact:** 20% → 80% | **Effort:** 5 min | **Depends on:** F1

**Steps:**

1. Remove empty directories left after migration:
   - `internal/cleaner/` (if empty)
   - `internal/domain/` (if empty)
   - `internal/adapters/` (if empty)
   - `internal/config/` (if empty)
   - `internal/conversions/` (if empty)
   - `internal/format/` (if empty)
   - `internal/testing/` (if empty)
   - `internal/pkg/` (if empty)
   - `internal/shared/` (if empty after moves)
2. `go build ./...` && `go test ./...`

**Verification:** No empty directories. Build passes.

**Rollback:** `git revert HEAD`

---

### Task F3: Verify per-module independence

**Impact:** 20% → 80% | **Effort:** 10 min | **Depends on:** F2

**Steps:**

1. `cd core && go build ./... && go test ./... && go vet ./... && go mod tidy`
2. `cd adapters && go build ./... && go test ./... && go vet ./... && go mod tidy`
3. `cd config && go build ./... && go test ./... && go vet ./... && go mod tidy`
4. `cd cleaners && go build ./... && go test ./... && go vet ./... && go mod tidy`
5. `cd / && go build ./... && go test ./... && go vet ./... && go mod tidy`
6. Verify `go mod tidy` changes nothing in any module
7. Verify `go mod verify` passes in all modules
8. Verify `go work sync` produces no changes

**Verification:** Every module builds, tests, and passes vet independently. No module has dirty go.mod.

**Rollback:** N/A (read-only verification)

---

### Task F4: Update flake.nix

**Impact:** 20% → 80% | **Effort:** 20 min | **Depends on:** F3

**Steps:**

1. Update `flake.nix` build to use `go.work`
2. Add per-module test targets:
   - `nix run .#test-core`
   - `nix run .#test-adapters`
   - `nix run .#test-config`
   - `nix run .#test-cleaners`
3. Add aggregated root test target
4. Verify `nix build` passes
5. Verify `nix flake check` passes

**Verification:** Nix builds all modules. Per-module tests work.

**Rollback:** `git revert HEAD`

---

### Task F5: Update documentation

**Impact:** 20% → 80% | **Effort:** 15 min | **Depends on:** F3

**Steps:**

1. Update `AGENTS.md` — new module structure, build/test commands per module
2. Update `README.md` — if it references internal structure
3. Update `TODO_LIST.md` — mark modularization items as done
4. Update `FEATURES.md` — if it references module structure

**Verification:** Documentation accurately reflects the new structure.

**Rollback:** `git revert HEAD`

---

## Task Dependency Graph

```
A1 ──┐
A2 ──┤
A3 ──┼── B1 ── B2 ──┐
A4 ──┘    │   B3 ──┤
          │   B4 ──┤
          │   B5 ──┤
          │   B6 ──┤
          │   B7 ──┤
          │        ↓
          │       B8 ── C1 ──┐
          │                 │
          │       E1 ── E2 ←┤
          │                 │
          └── D1 ── D2 ←───┘
                            │
                            ↓
                           F1 ── F2 ── F3 ── F4
                                             └── F5
```

**Critical path:** A3 → B1 → B8 → E2 → F1 → F3

---

## Pareto Impact Summary

| Phase          | Tasks        | Impact       | Effort       |
| -------------- | ------------ | ------------ | ------------ |
| A: Pre-cleanup | A1-A4        | 1% → 51%     | ~80 min      |
| B: Core module | B1-B8        | 4% → 64%     | ~95 min      |
| C: Adapters    | C1           | 4% → 64%     | ~20 min      |
| D: Config      | D1-D2        | 4% → 64%     | ~40 min      |
| E: Cleaners    | E1-E2        | 4% → 64%     | ~60 min      |
| F: Finalize    | F1-F5        | 20% → 80%    | ~70 min      |
| **Total**      | **20 tasks** | **0% → 80%** | **~365 min** |

---

## Risk Mitigation Per Task

| Risk                             | Tasks Affected | Mitigation                                                      |
| -------------------------------- | -------------- | --------------------------------------------------------------- |
| Import path typos                | All move tasks | Use `sed` + `gofmt` for find-replace; verify with `go build`    |
| Missed imports                   | E2, F1         | `grep -r 'internal/cleaner\|internal/domain' .` after each task |
| Test failures from moved helpers | A3, E2         | Move helpers first; run full test suite                         |
| go.work sync issues              | B8, C1, D2, E2 | `go work sync` after every module creation                      |
| Circular dependency              | C1, D2, E2     | `go vet ./...` catches cycles; DAG pre-verified                 |
