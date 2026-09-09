# Migration to Nix Flakes Proposal

**Project:** clean-wizard
**Date:** 2026-04-09
**Status:** Draft

---

## Executive Summary

clean-wizard currently manages its development environment through an ad-hoc combination of `go.mod` tool directives, `just` recipes that auto-install tools via `go install`, Homebrew (`brew install trivy`), npm (`bun install -g jscpd`), and pip (`pip install jsonschema pyyaml`). This creates several problems:

1. **No reproducibility** — developers get different tool versions depending on when they first ran `just bootstrap`
2. **No declarative lock** — there is no single file that pins every tool to an exact version
3. **Platform drift** — macOS vs Linux inconsistencies in tool availability and paths
4. **CI/CD fragility** — GitHub Actions installs tools imperatively, introducing flakiness
5. **Polluted global environment** — `go install -tool` scatters binaries into `GOPATH/bin`

**Nix Flakes** solve all of these by providing a single `flake.nix` + `flake.lock` that declares and pins every dependency — Go compiler, linting tools, formatters, test frameworks, and the project binary itself — with hermetic, reproducible builds.

This proposal outlines a phased migration that introduces Nix Flakes alongside existing tooling, with zero disruption to the current workflow.

---

## Table of Contents

- [Current State Analysis](#current-state-analysis)
- [Proposed Architecture](#proposed-architecture)
- [Phase 1: Development Shell](#phase-1-development-shell)
- [Phase 2: Reproducible Package Build](#phase-2-reproducible-package-build)
- [Phase-3-ci-cd-integration](#phase-3-ci-cd-integration)
- [Phase 4: Advanced Features](#phase-4-advanced-features)
- [File Structure](#file-structure)
- [Migration Path](#migration-path)
- [Risk Analysis](#risk-analysis)
- [Open Questions](#open-questions)

---

## Current State Analysis

### Tooling Inventory

The following external tools are required across the project, extracted from `Justfile`, `dev/arch-lint.just`, `scripts/run_fuzz_tests.sh`, and `.github/workflows/type-safety.yml`:

#### Build & Runtime Dependencies

| Tool      | Source             | Purpose                     |
| --------- | ------------------ | --------------------------- |
| Go 1.26.1 | `go.mod` directive | Compiler and toolchain      |
| `just`    | Homebrew / system  | Task runner (replaces Make) |

#### Go-Based Development Tools (installed via `go install`/`go get -tool`)

| Tool            | Package                                    | Install Method               | Used In                          |
| --------------- | ------------------------------------------ | ---------------------------- | -------------------------------- |
| `golangci-lint` | `github.com/golangci/golangci-lint`        | `go install` / GitHub Action | `dev/arch-lint.just`             |
| `gofumpt`       | `mvdan.cc/gofumpt`                         | `go install -tool`           | `dev/arch-lint.just`             |
| `goimports`     | `golang.org/x/tools/cmd/goimports`         | `go install -tool`           | `Justfile`, `dev/arch-lint.just` |
| `go-arch-lint`  | `github.com/fe3dback/go-arch-lint` v1.14.0 | `go get -tool`               | `dev/arch-lint.just`             |
| `govulncheck`   | `golang.org/x/vuln/cmd/govulncheck`        | `go get -tool`               | `dev/arch-lint.just`             |
| `nilaway`       | `go.uber.org/nilaway/cmd/nilaway`          | `go get -tool`               | `dev/arch-lint.just`             |
| `capslock`      | `github.com/google/capslock`               | `go get -tool`               | `dev/arch-lint.just`             |
| `go-licenses`   | `github.com/google/go-licenses`            | `go get -tool`               | `dev/arch-lint.just`             |
| `goleak`        | `go.uber.org/goleak`                       | `go get -tool`               | `dev/arch-lint.just`             |
| `dupl`          | `github.com/mibk/dupl`                     | `go install -tool`           | `dev/arch-lint.just`             |
| `templ`         | `github.com/a-h/templ/cmd/templ`           | `go install -tool`           | `dev/arch-lint.just`             |
| `air`           | `github.com/cosmtrek/air`                  | `go install -tool`           | `dev/arch-lint.just`             |
| `benchcmp`      | `golang.org/x/tools/cmd/benchcmp`          | `go install -tool`           | `dev/arch-lint.just`             |
| `gocycmd`       | `github.com/fzipp/gocycmd`                 | `go install`                 | CI workflow                      |

#### Non-Go Development Tools

| Tool         | Install Method         | Purpose                            |
| ------------ | ---------------------- | ---------------------------------- |
| `trivy`      | `brew install trivy`   | Docker security scanning           |
| `jscpd`      | `bun install -g jscpd` | Multi-language duplicate detection |
| `jq`         | System package         | JSON processing in scripts         |
| `bc`         | System package         | Arithmetic in coverage scripts     |
| `node`/`npm` | System / nvm           | TypeSpec API compilation           |

#### System Tools Invoked at Runtime (by cleaners)

| Tool                      | Used By Cleaner       | Notes                      |
| ------------------------- | --------------------- | -------------------------- |
| `nix`                     | `NixCleaner`          | Nix store operations       |
| `docker`                  | `DockerCleaner`       | Container/image cleanup    |
| `brew`                    | `HomebrewCleaner`     | Homebrew cache cleanup     |
| `cargo`                   | `CargoCleaner`        | Rust cache cleanup         |
| `npm`/`yarn`/`pnpm`/`bun` | `NodePackagesCleaner` | Node cache cleanup         |
| `git`                     | `GitHistoryCleaner`   | Git history operations     |
| `git-filter-repo`         | `GitHistoryCleaner`   | History rewriting          |
| `trash`                   | `TempFilesCleaner`    | Safe file deletion (macOS) |

### Current Pain Points

1. **No version pinning for dev tools** — `go install -tool ...@latest` means every developer potentially gets a different version of `gofumpt`, `golangci-lint`, etc.

2. **Bootstrap complexity** — The `dev/arch-lint.just` contains 4 bootstrap variants (`bootstrap`, `bootstrap-diagnose`, `bootstrap-fix`, `bootstrap-quick`) that download external `bootstrap.sh` scripts at runtime. This is fragile and hard to audit.

3. **Scattered installation methods** — Tools come from `go install`, `go get -tool`, `brew`, `bun`, and `pip`. No single command can reproduce the full environment.

4. **CI instability** — The GitHub Actions workflow installs `golangci-lint` via an action but `gocycmd` via `go install`. Go version mismatch (`go.mod` says 1.26.1 but CI uses `setup-go@v4` with `"1.21"`).

5. **Fuzzy tool availability checks** — Throughout `dev/arch-lint.just`, every recipe starts with `if command -v <tool> >/dev/null 2>&1; then ... else echo "not found. Installing..."`. This is slow, error-prone, and produces inconsistent behavior.

6. **Global pollution** — `go install -tool` modifies `go.mod` and places binaries in `GOPATH/bin`, affecting all Go projects on the machine.

---

## Proposed Architecture

### Core Principle: Flake as Single Source of Truth

```
flake.nix     → Declarative specification of all dependencies and outputs
flake.lock    → Reproducible lockfile (like go.sum but for the entire environment)
```

Everything else (`Justfile`, `scripts/`, `go.mod`) remains unchanged. Nix Flakes complement the existing workflow — they do not replace it.

### Technology Choices

| Decision             | Choice                                          | Rationale                                                          |
| -------------------- | ----------------------------------------------- | ------------------------------------------------------------------ |
| Flake inputs         | `nixpkgs` only (no flake-utils, no flake-parts) | Minimal dependencies, easier to understand, Go projects are simple |
| Go version selection | `pkgs.go_1_26` from nixpkgs                     | Matches `go.mod`'s `go 1.26.1`                                     |
| Build function       | `buildGoModule`                                 | Standard for Go projects with `go.sum`                             |
| Dev shell            | `pkgs.mkShell` with `inputsFrom`                | Shares build dependencies, ensures consistency                     |
| Direnv integration   | Optional `.envrc`                               | Auto-loads shell on `cd` — but not required                        |

---

## Phase 1: Development Shell

**Goal:** Every developer gets the exact same tools with one command: `nix develop`.

### Deliverable: `flake.nix` with `devShells.default`

```nix
{
  description = "clean-wizard: System cleanup CLI with type-safe architecture";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forEachSupportedSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import nixpkgs { inherit system; };
            inherit system;
          }
        );
    in
    {
      devShells = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              # Go toolchain (matches go.mod: 1.26.1)
              go_1_26

              # Go development tools (all version-pinned via nixpkgs)
              gotools           # goimports, godoc, etc.
              golangci-lint     # comprehensive Go linter
              gofumpt           # strict Go formatter
              govulncheck       # vulnerability scanner
              go-licenses       # license compliance

              # Task runner
              just

              # General tools
              jq                # JSON processing
              git               # version control
              docker            # container management
              trivy             # container security scanning

              # Code quality
              gnumake           # for any legacy makefiles
            ];

            shellHook = ''
              echo "🧙 clean-wizard development environment"
              echo "  Go:    $(go version)"
              echo "  Just:  $(just --version)"
              echo ""
              echo "Available commands:"
              echo "  just build          — Build the binary"
              echo "  just test           — Run tests"
              echo "  just lint           — Run all linters"
              echo "  just dev            — Development with auto-reload"
            '';
          };
        }
      );
    };
}
```

### What This Provides

Running `nix develop` gives a shell with **every** development tool pre-installed and version-pinned:

| Tool            | Current Install              | With Nix                      |
| --------------- | ---------------------------- | ----------------------------- |
| Go 1.26         | `go.mod` directive           | `pkgs.go_1_26` — exact match  |
| `golangci-lint` | `go install` / GitHub Action | `pkgs.golangci-lint` — locked |
| `gofumpt`       | `go install -tool`           | `pkgs.gofumpt` — locked       |
| `goimports`     | `go install -tool`           | `pkgs.gotools` — locked       |
| `govulncheck`   | `go get -tool`               | `pkgs.govulncheck` — locked   |
| `go-licenses`   | `go get -tool`               | `pkgs.go-licenses` — locked   |
| `just`          | Homebrew / system            | `pkgs.just` — locked          |
| `trivy`         | `brew install`               | `pkgs.trivy` — locked         |
| `jq`            | System package               | `pkgs.jq` — locked            |
| `docker`        | System install               | `pkgs.docker` — locked        |

### Tools Not Yet in Nix (Phase 2 candidates)

These tools are not available in nixpkgs or are more specialized:

| Tool                   | Status                             | Phase 2 Strategy                      |
| ---------------------- | ---------------------------------- | ------------------------------------- |
| `go-arch-lint` v1.14.0 | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `capslock`             | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `nilaway`              | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `goleak` (CLI)         | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `dupl`                 | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `templ`                | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `air`                  | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `gocycmd`              | Not in nixpkgs                     | Build from source via `buildGoModule` |
| `jscpd`                | npm package                        | `pkgs.nodePackages.jscpd` or skip     |
| `benchcmp`             | Deprecated in favor of `benchstat` | Use `pkgs.benchstat`                  |

### Step-by-Step Actions

1. Create `flake.nix` with `devShells` only (no package build yet)
2. Run `nix flake lock` to generate `flake.lock`
3. Verify with `nix develop --command just test`
4. Add `.direnv/` and `result` to `.gitignore`
5. Document usage in `DEVELOPMENT.md`

### Verification

```bash
# Enter development shell
nix develop

# Verify tools are available
go version           # Should show go1.26.x
golangci-lint version
gofumpt --version
just --version
govulncheck --version

# Run existing test suite
just test

# Run linting
just lint
```

---

## Phase 2: Reproducible Package Build

**Goal:** `nix build` produces a reproducible `clean-wizard` binary.

### Deliverable: `packages.default` using `buildGoModule`

```nix
packages = forEachSupportedSystem (
  { pkgs, system }:
  {
    default = pkgs.buildGoModule {
      pname = "clean-wizard";
      version = "0.1.0"; # or derive from git tags

      src = self;

      vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Updated via `nix build`

      subPackages = [ "cmd/clean-wizard" ];

      # Build-time ldflags for version injection
      ldflags = [
        "-s"
        "-w"
        "-X github.com/LarsArtmann/clean-wizard/internal/version.Version=${version}"
      ];

      # Tests require external tools (docker, nix, etc.) — skip in Nix build
      doCheck = false;

      meta = with pkgs.lib; {
        description = "System cleanup CLI with type-safe architecture";
        homepage = "https://github.com/LarsArtmann/clean-wizard";
        license = licenses.asl20;
        mainProgram = "clean-wizard";
        platforms = platforms.unix;
      };
    };
  }
);
```

### How `vendorHash` Works

1. First build will fail with a hash mismatch error — this is expected
2. The error message provides the correct hash
3. Copy the correct hash into `flake.nix`
4. Subsequent builds are fully reproducible

```bash
# First attempt — will show the correct hash
nix build

# Update vendorHash in flake.nix, then:
nix build

# Run the built binary
./result/bin/clean-wizard --help
```

### Dynamic Version from Git

For automatic version detection from git tags:

```nix
version =
  let
    ref = self.shortRev or "dirty";
    tag = self.ref or "0.0.0";
  in
  # If we're on a tagged commit, use the tag; otherwise use short rev
  if builtins.match "v[0-9]+.*" tag != null then
    builtins.substring 1 (builtins.stringLength tag - 1) tag
  else
    "0.0.0-${ref}";
```

### Step-by-Step Actions

1. Add `packages` output to existing `flake.nix`
2. Run `nix build`, capture correct `vendorHash`
3. Verify binary works: `./result/bin/clean-wizard --help`
4. Cross-compile test: `nix build .#packages.aarch64-darwin.default`

---

## Phase 3: CI/CD Integration

**Goal:** Replace imperative tool installation in GitHub Actions with Nix.

### Deliverable: Updated `.github/workflows/type-safety.yml`

```yaml
name: Type Safety Validation

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  type-safety-check:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install Nix
        uses: DeterminateSystems/nix-installer-action@v16

      - name: Cache Nix Store
        uses: DeterminateSystems/magic-nix-cache-action@v9

      - name: Type Safety Linting
        run: nix develop --command bash -c "
          echo '🔍 TYPE SAFETY VALIDATION'
          echo 'Checking for type safety violations...'

          # 1. Check for map[string]any violations
          if grep -r 'map\[string\]any' internal/ --include='*.go' --exclude-dir=vendor; then
            echo '❌ VIOLATION: map[string]any found'
            exit 1
          fi
          echo '✅ No map[string]any violations'

          # 2. Check for interface{} abuse
          if grep -r 'interface{}' internal/ --include='*.go' --exclude-dir=vendor; then
            echo '❌ VIOLATION: interface{} found'
            exit 1
          fi
          echo '✅ No interface{} violations'

          # 3. Check for unsafe package usage
          if grep -r 'unsafe\.' internal/ --include='*.go' --exclude-dir=vendor; then
            echo '❌ VIOLATION: unsafe package usage'
            exit 1
          fi
          echo '✅ No unsafe package violations'

          # 4. Check for reflect package usage
          if grep -r 'reflect\.' internal/ --include='*.go' --exclude-dir=vendor; then
            echo '❌ VIOLATION: reflect package usage'
            exit 1
          fi
          echo '✅ No reflect package violations'

          # 5. Run golangci-lint with strict rules
          golangci-lint run --timeout=5m \
            --disable-all \
            --enable=govet,goimports,gofmt,errcheck,staticcheck,unused \
            --config=.golangci.yml || true
        "

      - name: Run Tests
        run: nix develop --command just test

      - name: Build Binary
        run: nix build

      - name: Verify Binary
        run: ./result/bin/clean-wizard --help
```

### Benefits Over Current CI

| Aspect            | Current                                                   | With Nix                                   |
| ----------------- | --------------------------------------------------------- | ------------------------------------------ |
| Go installation   | `setup-go@v4` (version "1.21" — **mismatch with go.mod**) | `nix develop` (go_1_26 — matches go.mod)   |
| golangci-lint     | `golangci-lint-action@v3` (separate install)              | Already in dev shell                       |
| gocycmd           | `go install` (imperative, unversioned)                    | Nix package (pinned)                       |
| Total setup steps | 4+ steps with `go install` scattered throughout           | 1 step: `nix develop`                      |
| Cache             | Go module cache only                                      | Full Nix store cache via `magic-nix-cache` |
| Reproducibility   | Low — depends on when CI runs                             | High — locked via `flake.lock`             |

### Critical Fix: Go Version Mismatch

The current CI uses `setup-go@v4` with `"1.21"` but `go.mod` requires `1.26.1`. Nix migration fixes this automatically since the dev shell uses `go_1_26`.

---

## Phase 4: Advanced Features

**Goal:** Polish and extend the Nix integration with developer-experience enhancements.

### 4.1 Direnv Integration (Optional)

Create `.envrc`:

```bash
# .envrc
use flake
```

This auto-loads the Nix development shell when entering the project directory. No need to type `nix develop` every time.

```bash
# One-time setup
echo ".direnv/" >> .gitignore
direnv allow
```

### 4.2 Custom Go Tools as Nix Packages

For tools not in nixpkgs (`go-arch-lint`, `capslock`, `nilaway`, `dupl`, `templ`, `air`, `gocycmd`), define them inline:

```nix
let
  # Custom Go tools built from source
  goArchLint = pkgs.buildGoModule {
    pname = "go-arch-lint";
    version = "1.14.0";
    src = pkgs.fetchFromGitHub {
      owner = "fe3dback";
      repo = "go-arch-lint";
      rev = "v1.14.0";
      hash = "sha256-..."; # Filled in during first build
    };
    vendorHash = "sha256-..."; # Filled in during first build
    subPackages = [ "cmd/go-arch-lint" ];
  };

  capslock = pkgs.buildGoModule {
    pname = "capslock";
    version = "0.1.0";
    src = pkgs.fetchFromGitHub {
      owner = "google";
      repo = "capslock";
      rev = "main";
      hash = "sha256-...";
    };
    vendorHash = "sha256-...";
  };

  nilaway = pkgs.buildGoModule {
    pname = "nilaway";
    version = "0.1.0";
    src = pkgs.fetchFromGitHub {
      owner = "uber-go";
      repo = "nilaway";
      rev = "main";
      hash = "sha256-...";
    };
    vendorHash = "sha256-...";
    subPackages = [ "cmd/nilaway" ];
  };
in
# ... use in devShells
```

### 4.3 Nix Apps for Common Workflows

Define Nix apps as shortcuts:

```nix
apps = forEachSupportedSystem (
  { pkgs, system }:
  {
    default = {
      type = "app";
      program = "${self.packages.${system}.default}/bin/clean-wizard";
    };

    lint = {
      type = "app";
      program = "${pkgs.golangci-lint}/bin/golangci-lint";
    };
  }
);
```

Usage:

```bash
nix run              # Run clean-wizard
nix run .#lint -- run --config .golangci.yml
```

### 4.4 Checks (Nix-native CI)

```nix
checks = forEachSupportedSystem (
  { pkgs, system }:
  {
    build = self.packages.${system}.default;

    fmt = pkgs.runCommandLocal "check-fmt" { } ''
      ${pkgs.gofumpt}/bin/gofumpt -l ${self} > $out 2>&1 || true
      if [ -s $out ]; then
        echo "Files need formatting:"
        cat $out
        exit 1
      fi
      echo "All files formatted correctly" > $out
    '';

    vet = pkgs.runCommandLocal "check-vet" { } ''
      cd ${self}
      ${pkgs.go_1_26}/bin/go vet ./... 2>&1 > $out
    '';
  }
);
```

Run all checks:

```bash
nix flake check
```

---

## File Structure

After migration, the project root will contain:

```
clean-wizard/
├── flake.nix              # NEW — Declarative project configuration
├── flake.lock             # NEW — Reproducible dependency lock
├── .envrc                 # NEW — Optional direnv config
├── .gitignore             # UPDATED — Add .direnv/, result, result-*
├── Justfile               # UNCHANGED — Task runner recipes
├── dev/
│   └── arch-lint.just     # UNCHANGED (initially) — Can simplify later
├── go.mod                 # UNCHANGED — Go module definition
├── go.sum                 # UNCHANGED — Go dependency lock
├── scripts/
│   └── run_fuzz_tests.sh  # UNCHANGED — Fuzz testing script
└── ...                    # All other files unchanged
```

### `.gitignore` Additions

```gitignore
# Nix
result
result-*
.direnv/
```

---

## Migration Path

### Execution Order

| Step | Action                                                   | Estimated Time | Risk   |
| ---- | -------------------------------------------------------- | -------------- | ------ |
| 1    | Create `flake.nix` with `devShells.default` only         | 30 min         | Low    |
| 2    | Generate `flake.lock` and verify tools                   | 15 min         | Low    |
| 3    | Update `.gitignore` for Nix artifacts                    | 5 min          | None   |
| 4    | Document in `DEVELOPMENT.md`                             | 15 min         | None   |
| 5    | Add `packages.default` with `buildGoModule`              | 30 min         | Low    |
| 6    | Update CI workflow to use Nix                            | 45 min         | Medium |
| 7    | Add custom Go tools (go-arch-lint, capslock, etc.)       | 2-3 hr         | Medium |
| 8    | Add direnv integration (`.envrc`)                        | 10 min         | Low    |
| 9    | Add `nix flake check` validations                        | 30 min         | Low    |
| 10   | Simplify `dev/arch-lint.just` (remove install-on-demand) | 1 hr           | Low    |

### Backward Compatibility

**Critical:** Nix is additive. The existing workflow continues to work:

- `just build` — still works (uses system Go)
- `just test` — still works
- `just lint` — still works (if tools installed)
- `go test ./...` — still works
- `scripts/run_fuzz_tests.sh` — still works

Nix users additionally get:

- `nix develop` — hermetic shell with all tools
- `nix build` — reproducible binary build
- `nix flake check` — CI-like validation locally

### What We Do NOT Change

- `go.mod` / `go.sum` — remain the source of truth for Go dependencies
- `Justfile` — remains the task runner interface
- `.golangci.yml` — remains the linter configuration
- Any Go source code — zero code changes required
- Project structure — no file reorganization

---

## Risk Analysis

### Low Risk

| Risk                                    | Mitigation                                           |
| --------------------------------------- | ---------------------------------------------------- |
| Nix not installed on developer machines | Nix is optional — existing workflow works without it |
| `flake.lock` drift                      | `nix flake update` is explicit and deliberate        |
| Build hash mismatches                   | Standard `buildGoModule` workflow — well-documented  |

### Medium Risk

| Risk                                  | Mitigation                                              |
| ------------------------------------- | ------------------------------------------------------- |
| nixpkgs doesn't have Go 1.26 yet      | Use `go_1_25` with overlay, or wait for nixpkgs update  |
| Custom tool builds are fragile        | Pin to specific git refs, cache aggressively            |
| CI migration breaks existing pipeline | Keep old workflow, add new Nix-based workflow alongside |

### Mitigated Risks

| Risk                           | Status                                                                            |
| ------------------------------ | --------------------------------------------------------------------------------- |
| Nix replaces existing tooling  | **No** — purely additive                                                          |
| Developers forced to learn Nix | **No** — `just` commands remain primary interface                                 |
| Build times increase           | **No** — Nix caches are aggressive; `nix develop` is near-instant after first run |

---

## Open Questions

1. **Go 1.26 availability in nixpkgs** — As of April 2026, `go_1_26` may or may not be in nixpkgs-unstable. If not available, we can use an overlay to build from source or fall back to `go_1_25`. The `go.mod` says `go 1.26.1` so we should verify this first.

2. **Custom tool versions** — `go-arch-lint` is pinned to v1.14.0 in `dev/arch-lint.just`. Should we maintain this exact version in Nix, or accept whatever version nixpkgs provides?

3. **macOS-specific cleaners** — `HomebrewCleaner` and `SystemCacheCleaner` are macOS-only. The Nix dev shell on Linux simply won't have these tools available, which is the correct behavior.

4. **`jscpd` integration** — This is a Node.js tool that's optional (multi-language duplicate detection). We can include it via `pkgs.nodePackages.jscpd` or leave it as an optional manual install.

5. **Binary distribution** — Once `nix build` works, should we also produce static binaries via `pkgs.buildGoModule` with `CGO_ENABLED=0`? This would enable cross-platform distribution without Nix.

6. **Flakehub / Cachix** — Should we set up a binary cache (Cachix or Flakehub) for CI artifacts? This would eliminate build times for downstream consumers.

---

## Summary

| Phase       | What                               | Why                                                     | Effort  |
| ----------- | ---------------------------------- | ------------------------------------------------------- | ------- |
| **Phase 1** | Dev shell (`devShells.default`)    | Reproducible dev environment, replaces `just bootstrap` | ~1 hr   |
| **Phase 2** | Package build (`packages.default`) | `nix build` produces the binary                         | ~30 min |
| **Phase 3** | CI/CD integration                  | Reproducible CI, fixes Go version mismatch              | ~45 min |
| **Phase 4** | Advanced (direnv, apps, checks)    | Developer experience, `nix flake check`                 | ~3 hr   |

**Total estimated effort:** 5-6 hours spread across multiple sessions.

**Recommended approach:** Start with Phase 1, verify it works, then proceed incrementally. Each phase is independently valuable and can be merged separately.
