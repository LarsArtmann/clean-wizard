# Clean Wizard

**Reclaim disk space across macOS and Linux. One command, 13 specialized cleaners, zero guesswork.**

[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-300%2B-brightgreen.svg)](https://github.com/LarsArtmann/clean-wizard/actions)
[![Website](https://img.shields.io/badge/website-cleanwizard.lars.software-8b5cf6)](https://cleanwizard.lars.software)

---

## Why Clean Wizard?

Your dev machine accumulates gigabytes of stale data: old Nix generations, Go build caches, Docker images, Homebrew downloads, npm/pnpm/yarn/bun caches, Gradle/Maven artifacts, Xcode DerivedData, temp files, and more.

Clean Wizard finds and removes all of it — safely, with dry-run previews, confirmation dialogs, and protected generations.

**Reclaim 5-20 GB in seconds.**

## Quick Start

### Install

```bash
go install github.com/LarsArtmann/clean-wizard@latest
```

Or build from source:

```bash
git clone https://github.com/LarsArtmann/clean-wizard.git
cd clean-wizard
go build -o clean-wizard ./cmd/clean-wizard/
```

### Use

```bash
# Interactive TUI — pick what to clean
clean-wizard clean

# Quick daily cleanup (safe, no system-level operations)
clean-wizard clean --mode quick

# Preview without changing anything
clean-wizard clean --dry-run

# JSON output for scripts and CI
clean-wizard clean --json
```

## 13 Specialized Cleaners

| Cleaner              | What It Cleans                                       | Platforms |
| -------------------- | ---------------------------------------------------- | --------- |
| **Nix**              | Old generations, garbage collection                  | Linux     |
| **Homebrew**         | Cache downloads, dead symlinks, autoremove           | macOS     |
| **Docker**           | Stopped containers, dangling images, volumes, builds | Both      |
| **Cargo**            | Rust registry and git cache                          | Both      |
| **Go**               | Build cache, test cache, module cache, lint cache    | Both      |
| **Node**             | npm, pnpm, yarn, bun caches                          | Both      |
| **BuildCache**       | Gradle, Maven, SBT artifacts                         | Both      |
| **SystemCache**      | Spotlight, Xcode DerivedData, CocoaPods, pip, ccache | Both      |
| **TempFiles**        | Age-based temporary file removal                     | Both      |
| **ProjectExec**      | Old compiled scripts in `~/projects`                 | Both      |
| **CompiledBinaries** | Large stale binaries                                 | Both      |
| **GitHistory**       | Large blobs bloating git repos                       | Both      |
| **Golangci-lint**    | golangci-lint cache directory                        | Both      |

Each cleaner auto-detects whether its target tool is installed and available. Unavailable cleaners are silently skipped.

## Preset Modes

| Mode           | Flag                | What Runs                                                  |
| -------------- | ------------------- | ---------------------------------------------------------- |
| **Quick**      | `--mode quick`      | Homebrew, Go, Node, TempFiles, BuildCache                  |
| **Standard**   | `--mode standard`   | All available cleaners                                     |
| **Aggressive** | `--mode aggressive` | Everything including system caches and full Docker volumes |

## Safety First

- **Dry-run mode** — `--dry-run` previews every action without touching the filesystem
- **Confirmation dialogs** — explicit yes/no before any deletion
- **Protected generations** — current Nix generation is never deleted
- **Availability detection** — only shows cleaners for installed tools
- **Error classification** — transient failures auto-retry; unavailable tools are skipped, not crashed

## Architecture

Built with a type-safe, extensible architecture:

- **Type-safe enums** — 27 cache types as compile-time constants, not strings
- **Registry pattern** — thread-safe cleaner registry with runtime registration
- **Dependency injection** — `samber/do v2` container with typed accessors
- **Workflow engine** — `Azure/go-workflow` DAG with parallel execution and retry support
- **Functional error handling** — `Result[T]` type, no unchecked panics
- **Error classification** — `go-error-family` drives retry decisions, skip/failed classification, and BSD sysexits exit codes
- **300+ tests** — unit, integration, BDD (Ginkgo), benchmarks, fuzz

```
clean-wizard/
├── cmd/clean-wizard/          # CLI entry point (Cobra)
├── internal/
│   ├── cleaner/               # 13 cleaner implementations + registry
│   ├── di/                    # Dependency injection (samber/do v2)
│   ├── execution/             # Workflow orchestration (Azure/go-workflow)
│   ├── domain/                # Type-safe enums, settings, interfaces
│   ├── config/                # YAML configuration loading (Koanf)
│   ├── adapters/              # External tool adapters (Nix, Exec, HTTP)
│   ├── result/                # Result[T] functional error handling
│   └── format/                # Byte formatting, JSON output
├── tests/bdd/                 # Ginkgo BDD tests
└── docs/                      # Architecture documentation
```

## CLI Reference

### `clean-wizard clean`

```bash
clean-wizard clean [flags]
```

| Flag                  | Description                                             | Default                              |
| --------------------- | ------------------------------------------------------- | ------------------------------------ |
| `--mode`, `-m`        | Preset: `quick`, `standard`, `aggressive`               | `standard`                           |
| `--config`, `-c`      | Path to config file                                     | `~/.config/clean-wizard/config.yaml` |
| `--profile`, `-p`     | Configuration profile                                   | `""`                                 |
| `--dry-run`           | Preview without making changes                          | `false`                              |
| `--json`              | Machine-readable JSON output                            | `false`                              |
| `--verbose`           | Detailed logging                                        | `false`                              |
| `--yes`, `-y`         | Skip confirmation prompts                               | `false`                              |
| `--retries`           | Retry attempts per cleaner (0=disabled)                 | `3`                                  |
| `--retry-profile`     | Preset: `default`, `aggressive`, `conservative`, `none` | `""`                                 |
| `--concurrency`, `-C` | Max concurrent cleaners (0=unlimited)                   | `0`                                  |

### `clean-wizard scan`

Scans and reports reclaimable space without cleaning:

```bash
clean-wizard scan                  # Scan all available cleaners
clean-wizard scan --json           # JSON output
clean-wizard scan --verbose        # Detailed breakdown
```

## Configuration

```yaml
# ~/.config/clean-wizard/config.yaml

presets:
  quick:
    cleaners: [homebrew, go, node, tempfiles, buildcache]
  standard:
    cleaners: [homebrew, go, node, cargo, tempfiles, buildcache, systemcache, docker, nix]
  aggressive:
    cleaners: [all]
    include_dangerous: true

nix:
  keep_generations: 5

docker:
  timeout: 2m
  include_volumes: true

tempfiles:
  older_than: 7d
  exclude_paths:
    - /tmp/important-*
```

## Testing

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# BDD scenarios
go test ./tests/bdd/...

# Coverage report
go test -cover ./...
```

## Development

```bash
# Build
go build -o clean-wizard ./cmd/clean-wizard/

# Lint
golangci-lint run ./...

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o clean-wizard-darwin ./cmd/clean-wizard/
```

Requires `GOEXPERIMENT=jsonv2` (set automatically in the Nix devShell via `nix develop`).

## Comparison with SystemNix

Clean Wizard is the successor to [SystemNix](https://github.com/LarsArtmann/SystemNix) (a POSIX shell script):

|                 | SystemNix  | Clean Wizard                 |
| --------------- | ---------- | ---------------------------- |
| Language        | Shell      | Type-safe Go                 |
| Dry-run         | No         | Yes                          |
| Interactive TUI | No         | Yes                          |
| JSON output     | No         | Yes                          |
| Test coverage   | Manual     | 300+ tests                   |
| Configuration   | Hardcoded  | YAML profiles                |
| Error handling  | Exit codes | Classified errors with retry |

## Documentation

Full documentation lives at **[cleanwizard.lars.software](https://cleanwizard.lars.software)**.

## Contributing

1. Fork and clone the repository
2. Create a feature branch (`git switch -c feature/new-cleaner`)
3. Run tests (`go test -race ./...`)
4. Ensure linting passes (`golangci-lint run ./...`)
5. Open a pull request

See the [contributing guide](https://cleanwizard.lars.software/contributing/) for details.

## Links

- **Website & Docs:** [cleanwizard.lars.software](https://cleanwizard.lars.software)
- **GitHub:** [github.com/LarsArtmann/clean-wizard](https://github.com/LarsArtmann/clean-wizard)
- **Issue Tracker:** [github.com/LarsArtmann/clean-wizard/issues](https://github.com/LarsArtmann/clean-wizard/issues)
- **Predecessor:** [SystemNix](https://github.com/LarsArtmann/SystemNix)
- **Origin:** [Setup-Mac #111](https://github.com/LarsArtmann/Setup-Mac/issues/111)

## License

MIT License. See [LICENSE](LICENSE).
