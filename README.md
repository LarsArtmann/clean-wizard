> 🚨 **PROJECT ORIGIN**
>
> This project originated from: [Setup-Mac](https://github.com/LarsArtmann/Setup-Mac)
>
> **GitHub Issue:** [Setup-Mac #111](https://github.com/LarsArtmann/Setup-Mac/issues/111)

# Clean Wizard

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Inspired by](https://img.shields.io/badge/inspired%20by-GoReleaser%20Wizard-purple.svg)](https://github.com/LarsArtmann/GoReleaser-Wizard)

**The interactive system cleaning wizard that replaces all your cleanup commands with a unified, intelligent solution.**

Stop running multiple `just clean*` commands. Stop guessing what to clean. Get a comprehensive cleaning solution with one command.

## ✨ Features

- 🎯 **Interactive wizard** - Guides you through every cleaning option
- 🧠 **Smart scanning** - Automatically detects cleanable items and sizes
- 🚀 **Multiple profiles** - Daily, comprehensive, and aggressive cleaning modes
- 🔒 **Safety built-in** - Protected paths, dry runs, and backup creation
- 📊 **Real-time feedback** - Progress visualization and size estimates
- 🎨 **Beautiful TUI** - Powered by Charm Bracelet for a modern terminal experience
- ⚡ **Fast & efficient** - Optimized cleaning operations for macOS/Linux

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
# Initialize configuration
clean-wizard init

# Scan for cleanable items
clean-wizard scan

# Perform cleanup
clean-wizard clean
```

## 📸 Demo

```bash
$ clean-wizard init
🧹 Clean Wizard Setup
======================
Let's create the perfect cleaning configuration for your system!

? Enable safe mode? › Yes
? Enable dry run by default? › Yes
? Enable automatic backups? › Yes
? Maximum disk usage percentage? › 90

✅ Configuration created successfully at ~/.clean-wizard.yaml
```

```bash
$ clean-wizard scan
🔍 Analyzing system state...
✅ Configuration applied: safe_mode=ENABLED, profiles=2
🏷️  Using profile: daily (Quick daily cleanup for routine maintenance)

📊 Scan Results:
   • Total generations: 5
   • Current generation: 1
   • Cleanable generations: 4
   • Store size: 250.0 MB

💡 You can clean up 4 old generations to free space

✅ Scan completed!
```

## 🛠️ Commands

### `clean-wizard init`

Interactive setup wizard that creates a comprehensive cleaning configuration.

**Options:**

- `--force, -f` - Overwrite existing configuration
- `--minimal` - Create minimal configuration

### `clean-wizard scan`

Scans your system for cleanable items and shows size estimates.

**Options:**

- `--verbose, -v` - Show detailed scan information
- `--profile, -p` - Filter results by profile

### `clean-wizard clean`

Performs system cleanup based on configuration or profile.

**Options:**

- `--profile, -p` - Cleaning profile to use (daily, comprehensive, aggressive)
- `--dry-run` - Show what would be cleaned without doing it
- `--force, -f` - Skip confirmation prompts
- `--verbose, -v` - Show detailed output

### `clean-wizard config`

Manage configuration files.

**Subcommands:**

- `show` - Display current configuration
- `edit` - Edit configuration in default editor
- `validate` - Validate configuration file
- `reset` - Reset to defaults

### `clean-wizard profile`

Manage cleaning profiles.

**Subcommands:**

- `list` - List available profiles
- `show [profile]` - Show profile details
- `create` - Create new profile
- `delete [profile]` - Delete profile

## 🎯 Cleaning Profiles

### Daily Profile

Quick daily cleanup for routine maintenance.

**Operations:**

- Clean old Nix generations
- Homebrew autoremove and recent prune
- Go package caches

### Comprehensive Profile

Complete system cleanup for weekly maintenance.

**Operations:**

- Nix store optimization
- Full Homebrew cleanup
- All package caches (npm, pnpm, cargo, etc.)
- System temporary files
- Docker system prune

### Aggressive Profile

Nuclear option - everything that can be cleaned.

**Operations:**

- Remove all Nix generations
- Clean all language versions (Node, Python, Ruby)
- Remove all cache directories
- System logs and temp files
- Docker volumes and images
- iOS simulator data

## 🔒 Safety Features

### Protected Paths

The following paths are never cleaned by default:

- `/nix/store` - Nix store
- `/Users` - User directories
- `/System` - System files
- `/Applications` - Applications
- `/Library` - Library files

### Safe Mode

- Prevents dangerous operations
- Shows warnings before risky actions
- Can be disabled with `--force` flag

### Dry Run Mode

- Shows what would be cleaned
- No actual changes made
- Perfect for testing configurations

### Backup Creation

- Creates backups before aggressive operations
- Configurable backup location
- Can be disabled if not needed

## 🏗️ Architecture

Clean Wizard is built with inspiration from GoReleaser-Wizard and uses:

- **Cobra** - CLI framework
- **Viper** - Configuration management
- **Charm Bracelet** - Beautiful TUI components
- **Huh** - Interactive forms and selections
- **Bubbletea** - Terminal user interface

## 📦 What It Cleans

### Package Managers

- **Nix** - Store optimization, generation cleanup
- **Homebrew** - Autoremove, cache cleanup
- **npm/pnpm** - Package cache cleanup
- **Go** - Module cache cleanup
- **Cargo** - Rust package cache cleanup

### System Files

- **Temporary files** - System temp directories
- **Spotlight metadata** - Search index cleanup
- **System logs** - Log file cleanup
- **Docker** - System prune, volume cleanup

### Development Tools

- **iOS Simulators** - Derived data, simulator cleanup
- **Language versions** - Node, Python, Ruby versions
- **Build caches** - Gradle, Maven, etc.

## 🧪 Testing

```bash
# Run tests
go test ./...

# Build and test CLI
go build -o clean-wizard ./cmd/clean-wizard/
./clean-wizard --help
./clean-wizard init
./clean-wizard scan
./clean-wizard clean --dry-run
```

## 📚 Examples

### Basic Daily Cleanup

```bash
clean-wizard clean --profile daily
```

### Comprehensive Cleanup with Dry Run

```bash
clean-wizard clean --profile comprehensive --dry-run
```

### Aggressive Cleanup (Force)

```bash
clean-wizard clean --profile aggressive --force
```

### Create Custom Profile

```bash
clean-wizard profile create
```

## 🔧 Configuration

The configuration file is located at `~/.clean-wizard.yaml`:

```yaml
defaults:
  safe_mode: true
  dry_run: true
  confirm_before_cleanup: true
  backup_enabled: true
  max_disk_usage_percent: 90

safety:
  protected_paths:
    - /nix/store
    - /Users
    - /System
    - /Applications
    - /Library

profiles:
  daily:
    description: "Quick daily cleanup"
    operations:
      - nix: { generations: 1, optimize: false }
      - homebrew: { autoremove: true, prune: "recent" }
      - caches: { go: true, npm: false }
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [GoReleaser-Wizard](https://github.com/LarsArtmann/GoReleaser-Wizard) - Inspiration for the wizard interface
- [Charm Bracelet](https://charm.sh) - Beautiful terminal UI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

## 🔗 Links

- [GitHub Repository](https://github.com/LarsArtmann/clean-wizard)
- [Report Issues](https://github.com/LarsArtmann/clean-wizard/issues)
- [GoReleaser-Wizard](https://github.com/LarsArtmann/GoReleaser-Wizard)

---

**Made with ❤️ to simplify system cleanup**
