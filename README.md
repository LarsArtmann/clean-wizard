> 🚨 **PROJECT ORIGIN**
>
> This project originated from: [Setup-Mac](https://github.com/LarsArtmann/Setup-Mac)
>
> **GitHub Issue:** [Setup-Mac #111](https://github.com/LarsArtmann/Setup-Mac/issues/111)

# Clean Wizard

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-200%2B-green.svg)](https://github.com/LarsArtmann/clean-wizard/actions)

**A comprehensive system cleanup tool for macOS with type-safe architecture and interactive TUI.**

Clean Wizard cleans old caches, temporary files, and unused data across 11 different cleanup targets — from Nix generations to Homebrew, Docker containers to Go build caches, npm packages to language version managers.

## ✨ Features

### 11 Specialized Cleaners

| Cleaner | Target | Status |
|---------|--------|--------|
| **Nix** | Nix store and generations | ✅ Production Ready |
| **Homebrew** | Homebrew cache and autoremove | ✅ Production Ready |
| **Docker** | Containers, images, volumes, builds | ✅ Production Ready |
| **Cargo** | Rust Cargo cache and registries | ✅ Production Ready |
| **Go** | Go module, test, and build cache | ✅ Production Ready |
| **Node** | npm, pnpm, yarn, bun package caches | ✅ Production Ready |
| **BuildCache** | Gradle, Maven, SBT build caches | ✅ Production Ready |
| **SystemCache** | macOS Spotlight, Xcode, CocoaPods | ✅ Production Ready |
| **TempFiles** | Age-based temporary files | ✅ Production Ready |
| **LangVersion** | NVM, Pyenv, Rbenv version managers | 🚧 NO-OP (Planned) |
| **Projects** | Project automation (development) | 🚧 In Progress |

### Preset Modes

| Mode | Purpose | What's Cleaned |
|------|---------|----------------|
| **Quick** | Daily cleanup | Homebrew, Go, Node, TempFiles, BuildCache |
| **Standard** | Full cleanup | All available cleaners including Nix and Docker |
| **Aggressive** | Nuclear cleanup | All cleaners including language versions |

### Interactive TUI

- **Beautiful Forms** - Powered by Charm Bracelet's Huh library
- **Multi-Select** - Choose which cleaners to run
- **Progress Tracking** - See each cleaner execute in real-time
- **Size Estimates** - See freed space before confirming

### Safety Features

- **Dry-Run Mode** - Preview what would be cleaned without making changes
- **Confirmation Dialog** - Explicit Yes/No before any deletion
- **Protected Generations** - Current Nix generation is never deleted
- **Availability Detection** - Only shows cleaners for available tools

### Output Options

| Flag | Output |
|------|--------|
| *(default)* | Interactive TUI with forms |
| `--json` | Machine-readable JSON output |
| `--dry-run` | Preview mode (no changes) |
| `--verbose` | Detailed logging |

## 🎬 Quick Start

### Installation

```bash
# Install from source
go install github.com/LarsArtmann/clean-wizard@latest

# Or build locally
git clone https://github.com/LarsArtmann/clean-wizard.git
cd clean-wizard
go build -o clean-wizard ./cmd/clean-wizard/
```

### Basic Usage

```bash
# Standard mode - interactively select what to clean
clean-wizard clean

# Quick mode - daily cleanup without system changes
clean-wizard clean --mode quick

# Aggressive mode - all cleaners including dangerous ones
clean-wizard clean --mode aggressive

# Dry-run - preview without making changes
clean-wizard clean --dry-run

# JSON output for automation
clean-wizard clean --json

# With verbose logging
clean-wizard clean --verbose
```

The tool will:

1. Scan for all available cleaners (based on installed tools)
2. Show you which ones can be run
3. Let you select which ones to execute (or use preset modes)
4. Confirm before making any changes
5. Clean the selected targets and show results

## 📸 Demo

### Interactive TUI

```bash
$ clean-wizard clean
🔍 Scanning for available cleaners...
✓ Found 9 available cleaners

# TUI interface appears - multi-select form
# Select which cleaners to run

✅ Running cleaners...
  ✓ Homebrew: Freed 245 MB
  ✓ Docker: Freed 1.2 GB
  ✓ Go: Freed 512 MB
  ✓ Node: Freed 128 MB

✅ Cleanup completed in 2.3s
   • Freed 2.1 GB total
```

### Quick Mode

```bash
$ clean-wizard clean --mode quick --dry-run
🔍 Quick mode - excludes Nix, Docker, System

Would clean:
  • Homebrew cache
  • Go module cache
  • npm/pnpm/yarn/bun caches
  • Temporary files (>7 days)
  • Build caches (Gradle, Maven, SBT)

Estimated space: 500 MB
```

### JSON Output

```bash
$ clean-wizard clean --json
{
  "success": true,
  "freed_bytes": 2147483648,
  "duration_ms": 2300,
  "cleaners": {
    "homebrew": {"freed_bytes": 245000000, "success": true},
    "docker": {"freed_bytes": 1200000000, "success": true},
    "go": {"freed_bytes": 512000000, "success": true},
    "node": {"freed_bytes": 128000000, "success": true}
  }
}
```

## 🛠️ Commands

### `clean-wizard clean` (Main Command)

**Synopsis:**
```bash
clean-wizard clean [flags]
```

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--mode, -m` | Preset mode: quick, standard, aggressive | standard |
| `--config, -c` | Path to config file | ~/.config/clean-wizard/config.yaml |
| `--dry-run` | Preview without making changes | false |
| `--json` | Output as JSON | false |
| `--verbose` | Enable verbose logging | false |
| `--dry-run` | Simulate deletion | false |

**Examples:**

```bash
# Normal mode - select cleaners interactively
clean-wizard clean

# Quick daily cleanup
clean-wizard clean --mode quick

# Aggressive cleanup (requires confirmation)
clean-wizard clean --mode aggressive

# Preview what would be cleaned
clean-wizard clean --dry-run

# Automation-friendly output
clean-wizard clean --json

# With verbose logging
clean-wizard clean --verbose

# Dry-run in quick mode
clean-wizard clean --mode quick --dry-run

# Combined flags
clean-wizard clean --mode standard --dry-run --json
```

### Preset Modes

#### Quick Mode (`--mode quick`)

Daily cleanup that excludes system-critical paths:

```bash
# What's included
clean-wizard clean --mode quick

# Cleaners active in quick mode:
# - Homebrew (brew cleanup + autoremove)
# - Go (go clean -cache -testcache -modcache)
# - Node (npm, pnpm, yarn, bun caches)
# - TempFiles (age-based, excludes system)
# - BuildCache (Gradle, Maven, SBT)

# What's excluded in quick mode:
# - Nix store (never touch current gen)
# - Docker (requires explicit selection)
# - SystemCache (Xcode, Spotlight, etc.)
# - Language Versions (requires aggressive)
```

#### Standard Mode (`--mode standard`)

Full cleanup with all available cleaners:

```bash
# What's included
clean-wizard clean

# Cleaners active in standard mode:
# - All from quick mode
# - Nix (gc --delete-older-than 1d or keep N gens)
# - Docker (system prune -af --volumes)
# - Cargo (cargo cache --autoclean)
# - SystemCache (Spotlight, Xcode, CocoaPods)
```

#### Aggressive Mode (`--mode aggressive`)

Nuclear cleanup including dangerous operations:

```bash
# What's included
clean-wizard clean --mode aggressive

# Additional cleaners in aggressive mode:
# - All from standard mode
# - Language Version Managers (NVM, Pyenv, Rbenv)
# - All Nix generations (no keep count)
# - iOS Simulators (xcrun simctl delete all)
# - Full Docker with volumes
```

**⚠️ Warning:** Aggressive mode includes destructive operations. Always use `--dry-run` first to preview.

## 🔒 Safety Features

### Protected Operations

**Current Generation Protection:**
- Nix current generation is never shown for deletion
- System always keeps at least the active profile

**Confirmation Requirements:**
- Explicit Yes/No dialog before any deletion
- Lists exactly what will be removed
- Shows estimated space to be freed
- Can be bypassed with `--yes` flag (not recommended)

**Dry-Run Mode:**
```bash
# Safe preview - shows what would happen
clean-wizard clean --dry-run

# In dry-run mode:
# - No files are actually deleted
# - Shows estimate of freed space
# - Lists what would be cleaned
# - No confirmation required
```

### What Each Cleaner Cleans

| Cleaner | Target | Risk Level |
|---------|--------|------------|
| Nix | `/nix/store/*-generation` | Low |
| Homebrew | `~/Library/Caches/Homebrew` | Low |
| Docker | Containers, images, volumes | Medium |
| Cargo | `~/.cargo/registry`, `~/.cargo/git` | Low |
| Go | `go build cache`, `go mod cache` | Low |
| Node | `~/.npm`, `~/.pnpm`, `~/.cache/yarn`, `~/.bun` | Low |
| BuildCache | `~/.gradle`, `~/.m2`, `~/.sbt` | Low |
| SystemCache | `~/Library/Caches/*`, Xcode DerivedData | Medium |
| TempFiles | `/tmp/*` (age-based) | Low |
| LangVersion | `~/.nvm/versions`, `~/.pyenv/versions` | **High** |
| Projects | Project-specific cleanup | Medium |

---

## 📋 Configuration

### YAML Configuration File

Clean Wizard supports a YAML configuration file for customization:

```yaml
# ~/.config/clean-wizard/config.yaml

# Preset definitions
presets:
  quick:
    cleaners:
      - homebrew
      - go
      - node
      - tempfiles
      - buildcache
  
  standard:
    cleaners:
      - homebrew
      - go
      - node
      - cargo
      - tempfiles
      - buildcache
      - systemcache
      - docker
      - nix
  
  aggressive:
    cleaners:
      - all
    include_dangerous: true
    confirm: true

# Cleaner-specific settings
nix:
  keep_generations: 5
  optimization: true

docker:
  timeout: 2m
  include_volumes: true

tempfiles:
  older_than: 7d
  exclude_paths:
    - /tmp/important-*

# Output settings
output:
  json: false
  verbose: false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLEAN_WIZARD_CONFIG` | Path to config file |
| `CLEAN_WIZARD_DRY_RUN` | Default dry-run mode |
| `CLEAN_WIZARD_VERBOSE` | Default verbose mode |

---

## 🏗️ Architecture

Clean Wizard is built with a focus on type safety and extensibility:

```
clean-wizard/
├── cmd/
│   └── clean-wizard/          # CLI entry point
├── internal/
│   ├── cleaner/               # Cleaner implementations
│   │   ├── registry.go         # Thread-safe cleaner registry
│   │   ├── registry_factory.go # Default cleaner setup
│   │   ├── nix.go             # Nix store cleanup
│   │   ├── homebrew.go        # Homebrew cleanup
│   │   ├── docker.go          # Docker cleanup
│   │   ├── golang_*.go        # Go cache cleanup
│   │   ├── node_*.go          # Node package managers
│   │   └── ...
│   ├── domain/                # Type-safe enums and constants
│   ├── config/                # YAML configuration
│   └── adapters/             # External tool adapters
├── api/                        # REST API layer
├── schemas/                    # JSON schemas
└── tests/
    ├── unit/                  # Unit tests (200+)
    └── integration/           # BDD tests with Godog
```

### Key Design Principles

1. **Type-Safe Enums** - Compile-time safety for all constants
2. **Registry Pattern** - Centralized cleaner management with thread-safety
3. **Result Types** - Explicit success/failure with error handling
4. **Interface-Based** - All cleaners implement common `Cleaner` interface
5. **Testable** - 200+ tests with BDD scenarios

### Domain Enums

```go
// Example: Type-safe strategy constants
type CleanStrategyType int
const (
    StrategyDryRunType CleanStrategyType = iota
    StrategyConservativeType
    StrategyAggressiveType
)

// Example: Risk level with explicit values
type RiskLevelType int
const (
    RiskLowType RiskLevelType = iota
    RiskMediumType
    RiskHighType
    RiskCriticalType
)
```

---

## 🧪 Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestDockerCleaner ./internal/cleaner/

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# BDD tests
go test -tags=bdd ./tests/integration/...
```

### Test Coverage

| Test Type | Count | Status |
|-----------|-------|--------|
| Unit Tests | 200+ | ✅ All Passing |
| Integration Tests | 10+ | ✅ Passing |
| BDD Tests (Godog) | 5 | ✅ Passing |
| Benchmark Tests | 15 | ✅ Passing |

### Test Categories

- **Unit Tests** - Individual cleaner functionality
- **Integration Tests** - Clean command and registry
- **BDD Tests** - User scenarios with Godog
- **Benchmark Tests** - Performance measurement

---

## 🔧 Development

### Building

```bash
# Build for current platform
go build -o clean-wizard ./cmd/clean-wizard/

# Build with optimization
go build -ldflags="-s -w" -o clean-wizard ./cmd/clean-wizard/

# Cross-compile for macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o clean-wizard-darwin-arm64 ./cmd/clean-wizard/

# Run locally
./clean-wizard clean
```

### Linting

```bash
# Run golangci-lint
golangci-lint run ./...

# Run with specific linters
golangci-lint run --enable=staticcheck,gosimple ./...
```

### Adding a New Cleaner

```go
// 1. Implement the Cleaner interface
type MyCleaner struct {
    verbose bool
    dryRun  bool
}

func (mc *MyCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    // Implementation
    return result.Result[domain.CleanResult]{
        FreedBytes: 1024,
        Success:    true,
    }
}

func (mc *MyCleaner) IsAvailable(ctx context.Context) bool {
    // Check if tool is installed
    return true
}

// 2. Register in registry_factory.go
func DefaultRegistry() *Registry {
    registry := NewRegistry()
    registry.Register("mycleaner", NewMyCleaner(false, false))
    // ...
    return registry
}

// 3. Add domain enum if needed
// 4. Add tests
// 5. Update configuration schema
```

---

## 📊 Comparison with SystemNix

Clean Wizard and [SystemNix](https://github.com/LarsArtmann/SystemNix) serve similar purposes but with different approaches:

| Feature | SystemNix (justfile) | Clean Wizard (Go) |
|---------|---------------------|-------------------|
| Language | POSIX shell | Type-safe Go |
| Type Safety | ❌ Shell strings | ✅ Compile-time enums |
| Dry-Run Mode | ❌ Not available | ✅ Preview mode |
| Interactive TUI | ❌ CLI only | ✅ Beautiful forms |
| JSON Output | ❌ Not available | ✅ Machine-readable |
| Registry Pattern | ❌ Manual | ✅ Thread-safe registry |
| Test Coverage | ❌ Manual | ✅ 200+ tests |
| Configuration | ❌ Hardcoded | ✅ YAML profiles |

### Feature Parity

| Mode | SystemNix | Clean Wizard | Parity |
|------|-----------|--------------|--------|
| Quick Mode | ✅ | ~85% | 🟡 Partial |
| Standard Mode | ✅ | ~75% | 🟡 Partial |
| Aggressive Mode | ✅ | ~60% | 🔴 Poor |

### Missing in Clean Wizard

- Language Version Manager (NO-OP, needs implementation)
- Docker light prune (quick mode)
- Nix store optimization
- Nix profile management
- iOS simulator cleanup
- Size before/after display

---

## 🤝 Contributing

### Ways to Contribute

1. **Fix Language Version Manager** - Currently NO-OP, needs actual deletion logic
2. **Implement Docker light prune** - For quick mode parity
3. **Add Nix store optimization** - `nix-store --optimize`
4. **Implement iOS simulator cleanup** - `xcrun simctl delete`
5. **Write BDD tests** - Expand test coverage
6. **Improve documentation** - README, examples, guides
7. **Report issues** - Bug reports and feature requests

### Development Workflow

```bash
# Fork the repository
# Clone your fork
git clone https://github.com/YOUR-USERNAME/clean-wizard.git

# Create feature branch
git checkout -b feature/new-cleaner

# Make changes
# ...

# Run tests
go test ./...

# Commit with descriptive message
git commit -m "feat(cleaner): add new cleaner implementation"

# Push to your fork
git push origin feature/new-cleaner

# Create Pull Request
```

---

## 📝 Roadmap

### Immediate (This Week)

- [ ] Update README documentation
- [ ] Fix Go build cache location gap
- [ ] Implement Docker light prune
- [ ] Add Nix temp files cleanup

### Short-Term (This Month)

- [ ] Fix Language Version Manager NO-OP
- [ ] Implement Nix store optimization
- [ ] Implement Nix profile management
- [ ] Add iOS simulator cleanup
- [ ] Complete quick mode parity

### Long-Term (This Quarter)

- [ ] Complete standard mode parity
- [ ] Complete aggressive mode parity
- [ ] Write comprehensive BDD tests
- [ ] Create complete documentation
- [ ] Add performance benchmarks

---

## 🔗 Links

- **GitHub Repository:** https://github.com/LarsArtmann/clean-wizard
- **Issue Tracker:** https://github.com/LarsArtmann/clean-wizard/issues
- **Wiki:** https://github.com/LarsArtmann/clean-wizard/wiki
- **SystemNix (Comparison Target):** https://github.com/LarsArtmann/SystemNix

### Dependencies

- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **TUI Library:** [Huh](https://github.com/charmbracelet/huh)
- **TUI Framework:** [BubbleTea](https://github.com/charmbracelet/bubbletea)
- **Testing:** [Testify](https://github.com/stretchr/testify), [Godog](https://github.com/cucumber/godog)
- **Configuration:** [Viper](https://github.com/spf13/viper)

---

## 📜 License

MIT License - See [LICENSE](LICENSE) for details.

---

**Built with ❤️ for a cleaner macOS experience**

*The ultimate tool for keeping your MacBook clean and fast*