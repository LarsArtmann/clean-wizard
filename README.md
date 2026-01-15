> 🚨 **PROJECT ORIGIN**

> This project originated from: [Setup-Mac](https://github.com/LarsArtmann/Setup-Mac)
>
> **GitHub Issue:** [Setup-Mac #111](https://github.com/LarsArtmann/Setup-Mac/issues/111)

# Clean Wizard

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**A simple TUI tool to clean old Nix generations.**

Just run `clean-wizard clean` to scan and interactively select which generations to delete.

## ✨ Features

- 🎯 **Interactive TUI** - Select which generations to delete with a beautiful interface
- 🧠 **Smart scanning** - Automatically detects and lists all Nix generations
- 📊 **Size estimates** - Shows estimated space each generation occupies
- 🎨 **Beautiful UI** - Powered by Charm Bracelet's Huh library
- ⚡ **Fast & efficient** - Clean exactly what you want, nothing more

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
clean-wizard clean
```

That's it. The tool will:
1. Scan for all Nix generations
2. Show you which ones can be cleaned (old generations)
3. Let you interactively select which ones to delete
4. Confirm before making changes
5. Clean them and show results

## 📸 Demo

```bash
$ clean-wizard clean
🔍 Scanning for Nix generations...
✓ Current generation: 300 (from 1 day ago)
✓ Found 4 old generations

# TUI interface appears - select which generations to clean

# After selection:

🗑️  Cleaning 2 generation(s)...

Will delete:
  • Generation 299 (from 2 days ago) ~ 50 MB
  • Generation 298 (from 3 days ago) ~ 50 MB

Total space to free: 100 MB

# Confirm dialog appears

🧹 Cleaning...
  ✓ Removed generation 299
  ✓ Removed generation 298
  🔄 Running garbage collection...

✅ Cleanup completed in 2.3s
   • Removed 2 generation(s)
   • Freed approximately 100 MB
```

## 🛠️ Commands

### `clean-wizard clean`

Interactively scan and clean old Nix generations.

**No configuration needed. No profiles. No setup.** Just run it.

## 🔒 Safety Features

### Protected Generations

- The current (active) generation is never shown for deletion
- Confirmation dialog before any deletions
- Clear display of what will be deleted

### What It Cleans

- **Nix generations** - Old, historical Nix generations (current one is always protected)

## 🏗️ Architecture

Clean Wizard is built with:
- **Cobra** - CLI framework
- **Huh** (by Charm Bracelet) - Beautiful TUI forms and selections
- **BubbleTea** (by Charm Bracelet) - Terminal user interface framework

## 🧪 Testing

```bash
# Run tests
go test ./...

# Build and test CLI
go build -o clean-wizard ./cmd/clean-wizard/
./clean-wizard --help
./clean-wizard clean --help
```

## 🔗 Links

- [GitHub Repository](https://github.com/LarsArtmann/clean-wizard)
- [Report Issues](https://github.com/LarsArtmann/clean-wizard/issues)
- [Huh Library](https://github.com/charmbracelet/huh)
- [Charm Bracelet](https://charm.sh)

---

**Made with ❤️ to simplify Nix cleanup**
