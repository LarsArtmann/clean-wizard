# Usage Reference

Complete reference for all Clean Wizard commands, flags, and options.

## 📋 Command Overview

```
clean-wizard
├── clean        # Perform system cleanup
├── scan         # Scan for cleanable items
├── init         # Initialize configuration
├── profile      # Manage cleaning profiles
└── config       # Manage configuration
```

## 🧹 Global Flags

These flags are available on all commands:

| Flag                 | Short | Type   | Default   | Description                                          |
| -------------------- | ----- | ------ | --------- | ---------------------------------------------------- |
| `--verbose`          | `-v`  | bool   | `false`   | Enable verbose output                                |
| `--dry-run`          |       | bool   | `false`   | Show what would be deleted without actually deleting |
| `--force`            | `-f`  | bool   | `false`   | Force cleanup without confirmation                   |
| `--profile`          | `-p`  | string | `"daily"` | Configuration profile to use                         |
| `--validation-level` |       | string | `"basic"` | Validation level: none, basic, comprehensive, strict |
| `--config`           | `-c`  | string |           | Configuration file path                              |

### Validation Levels

- **none** - Skip all validation (not recommended)
- **basic** - Validate essential configuration (default)
- **comprehensive** - Validate all settings and cross-checks
- **strict** - Highest safety, requires all safety features enabled

## 🚀 Commands

### `clean-wizard clean`

Perform system cleanup based on configuration or profile.

#### Usage

```bash
clean-wizard clean [flags]
```

#### Flags Specific to `clean`

| Flag        | Type   | Default   | Description                                 |
| ----------- | ------ | --------- | ------------------------------------------- |
| `--dry-run` | bool   | `false`   | Show what would be cleaned without doing it |
| `--config`  | string |           | Configuration file path                     |
| `--profile` | string | `"daily"` | Cleaning profile to use                     |

#### Examples

```bash
# Basic cleanup with default profile
clean-wizard clean

# Dry run to see what would be cleaned
clean-wizard clean --dry-run

# Use specific profile
clean-wizard clean --profile comprehensive

# Verbose cleanup with custom config
clean-wizard clean --verbose --config ~/custom-config.yaml

# Force cleanup without prompts
clean-wizard clean --force --profile aggressive

# High safety cleanup
clean-wizard clean --validation-level strict
```

#### Output Format

```
🧹 Starting system cleanup...
📄 Loading configuration from /Users/user/.clean-wizard.yaml...
🔍 Applying validation level: comprehensive
🏷️  Using profile: daily (Daily cleanup)
✅ Configuration applied: safe_mode=true, profiles=3
🔍 Running in DRY-RUN mode (from flag) - no files will be deleted
⚙️  Settings loaded: &{NixGenerations:0xc000123abc}
🎛️  Nix Generations Settings: &{Generations:3 Optimize:true DryRun:true}
🎯 Cleanup Results (SUCCESS):
   • Duration: 2.345s
   • Status: 5 items would be cleaned
   • Space freed: 2.3 GB

💡 This was a dry run - no files were actually deleted
   🏃 Run 'clean-wizard clean' without --dry-run to perform cleanup

✅ Cleanup completed successfully
```

---

### `clean-wizard scan`

Scan your system for cleanable items and show size estimates.

#### Usage

```bash
clean-wizard scan [flags]
```

#### Flags Specific to `scan`

| Flag        | Short | Type   | Default | Description                    |
| ----------- | ----- | ------ | ------- | ------------------------------ |
| `--verbose` | `-v`  | bool   | `false` | Show detailed scan information |
| `--profile` | `-p`  | string |         | Filter results by profile      |

#### Examples

```bash
# Basic scan
clean-wizard scan

# Verbose scan with details
clean-wizard scan --verbose

# Scan with specific validation level
clean-wizard scan --validation-level comprehensive

# Filter by profile
clean-wizard scan --profile daily
```

#### Output Format

```
🔍 Scanning system...
✅ Scan completed!
📦 Nix Store: 2.3 GB cleanable
🍺 Homebrew: 150 MB cleanable
📁 Package Caches: 500 MB cleanable
💡 Total: ~3 GB can be recovered
```

---

### `clean-wizard init`

Interactive setup wizard that creates a comprehensive cleaning configuration.

#### Usage

```bash
clean-wizard init [flags]
```

#### Flags Specific to `init`

| Flag        | Short | Type | Default | Description                      |
| ----------- | ----- | ---- | ------- | -------------------------------- |
| `--force`   | `-f`  | bool | `false` | Overwrite existing configuration |
| `--minimal` |       | bool | `false` | Create minimal configuration     |

#### Examples

```bash
# Interactive setup
clean-wizard init

# Force overwrite existing config
clean-wizard init --force

# Create minimal configuration
clean-wizard init --minimal
```

#### Output Format

```
🧹 Clean Wizard Setup
======================
Let's create the perfect cleaning configuration for your system!

? Enable safe mode? › Yes
? Enable dry run by default? › Yes
? Enable automatic backups? › Yes
? Maximum disk usage percentage? › 90

✅ Configuration created successfully!
```

---

### `clean-wizard profile`

Manage cleaning profiles.

#### Usage

```bash
clean-wizard profile [subcommand] [flags]
```

#### Subcommands

##### `profile list`

List all available profiles.

```bash
clean-wizard profile list
```

##### `profile show [profile]`

Show details of a specific profile.

```bash
clean-wizard profile show daily
clean-wizard profile show comprehensive
```

##### `profile create`

Create a new profile interactively.

```bash
clean-wizard profile create
```

##### `profile delete [profile]`

Delete a profile.

```bash
clean-wizard profile delete custom-profile
```

#### Examples

```bash
# List all profiles
clean-wizard profile list

# Show daily profile details
clean-wizard profile show daily

# Create custom profile
clean-wizard profile create

# Delete unused profile
clean-wizard profile delete old-profile
```

---

### `clean-wizard config`

Manage configuration files.

#### Usage

```bash
clean-wizard config [subcommand] [flags]
```

#### Subcommands

##### `config show`

Display current configuration.

```bash
clean-wizard config show
```

##### `config edit`

Edit configuration in default editor.

```bash
clean-wizard config edit
```

##### `config validate`

Validate configuration file.

```bash
clean-wizard config validate
clean-wizard config validate --config /path/to/config.yaml
```

##### `config reset`

Reset configuration to defaults.

```bash
clean-wizard config reset
```

#### Examples

```bash
# Show current configuration
clean-wizard config show

# Edit configuration
clean-wizard config edit

# Validate configuration
clean-wizard config validate

# Validate specific file
clean-wizard config validate --config ~/backup-config.yaml

# Reset to defaults
clean-wizard config reset
```

---

## 📊 Exit Codes

| Code | Meaning             |
| ---- | ------------------- |
| 0    | Success             |
| 1    | General error       |
| 2    | Configuration error |
| 3    | Validation error    |
| 4    | Permission error    |
| 5    | Network/IO error    |

## 🔧 Configuration File

Clean Wizard uses YAML configuration stored at `~/.clean-wizard.yaml`.

### Configuration Structure

```yaml
version: "1.0.0"
safe_mode: true
max_disk_usage: 90
protected:
  - "/"
  - "/System"
  - "/Library"
  - "/Applications"
  - "/Users"
  - "/nix/store"

profiles:
  daily:
    name: "daily"
    description: "Daily cleanup for routine maintenance"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean old Nix generations"
        risk_level: "low"
        enabled: true
        settings:
          nix_generations:
            generations: 3
            optimize: true
            dry_run: false
      - name: "homebrew"
        description: "Homebrew cleanup"
        risk_level: "low"
        enabled: true
        settings:
          homebrew:
            autoremove: true
            prune: "recent"

  comprehensive:
    name: "comprehensive"
    description: "Weekly comprehensive cleanup"
    enabled: true
    operations:
      - name: "nix-generations"
        enabled: true
        settings:
          nix_generations:
            generations: 1
            optimize: true
      - name: "homebrew"
        enabled: true
        settings:
          homebrew:
            autoremove: true
            prune: "all"
      - name: "package-caches"
        enabled: true
        settings:
          package_caches:
            pnpm: true
            cargo: true
            go: true
```

### Configuration Fields

| Field            | Type     | Required | Description                                 |
| ---------------- | -------- | -------- | ------------------------------------------- |
| `version`        | string   | Yes      | Configuration version                       |
| `safe_mode`      | bool     | No       | Enable safety features (default: true)      |
| `max_disk_usage` | int      | No       | Maximum disk usage percentage (default: 90) |
| `protected`      | []string | Yes      | Paths that should never be cleaned          |
| `profiles`       | object   | Yes      | Cleaning profiles configuration             |

### Profile Configuration

| Field         | Type     | Required | Description                 |
| ------------- | -------- | -------- | --------------------------- |
| `name`        | string   | Yes      | Profile identifier          |
| `description` | string   | Yes      | Profile description         |
| `enabled`     | bool     | Yes      | Whether profile is active   |
| `operations`  | []object | Yes      | List of cleaning operations |

### Operation Configuration

| Field         | Type   | Required | Description                             |
| ------------- | ------ | -------- | --------------------------------------- |
| `name`        | string | Yes      | Operation identifier                    |
| `description` | string | Yes      | Operation description                   |
| `risk_level`  | string | Yes      | Risk level: low, medium, high, critical |
| `enabled`     | bool   | Yes      | Whether operation is active             |
| `settings`    | object | No       | Operation-specific settings             |

## 🎨 Environment Variables

| Variable                        | Default                | Description                |
| ------------------------------- | ---------------------- | -------------------------- |
| `CLEAN_WIZARD_CONFIG`           | `~/.clean-wizard.yaml` | Path to configuration file |
| `CLEAN_WIZARD_PROFILE`          | `daily`                | Default profile to use     |
| `CLEAN_WIZARD_VALIDATION_LEVEL` | `basic`                | Default validation level   |
| `NO_COLOR`                      |                        | Disable colored output     |
| `CLEAN_WIZARD_DRY_RUN`          | `false`                | Default dry-run setting    |

## 📝 Examples and Workflows

### Development Environment Setup

```bash
# Initialize with development-friendly settings
clean-wizard init --minimal

# Create developer profile
cat << EOF > ~/.clean-wizard.yaml
version: "1.0.0"
safe_mode: true
max_disk_usage: 85
protected:
  - "/Users/$USER/Development"
  - "/Users/$USER/.config"
profiles:
  dev:
    description: "Development environment cleanup"
    enabled: true
    operations:
      - name: "nix-generations"
        enabled: true
        settings:
          nix_generations:
            generations: 5  # Keep more for development
      - name: "go-cache"
        enabled: false      # Keep for faster builds
EOF

# Validate configuration
clean-wizard config validate
```

### Automated Cleanup Script

```bash
#!/bin/bash
# automated-cleanup.sh

set -euo pipefail

# Configuration
CONFIG_FILE="${HOME}/.clean-wizard.yaml"
LOG_FILE="${HOME}/.clean-wizard.log"
DATE=$(date +%Y-%m-%d)

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if configuration exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    log "Configuration not found. Initializing..."
    clean-wizard init --minimal
fi

# Scan first
log "Scanning system for cleanup opportunities..."
clean-wizard scan --validation-level comprehensive | tee -a "$LOG_FILE"

# Perform daily cleanup
log "Starting daily cleanup..."
clean-wizard clean --profile daily --validation-level strict --verbose | tee -a "$LOG_FILE"

log "Cleanup completed for $DATE"
```

### CI/CD Integration

```yaml
# .github/workflows/cleanup.yml
name: System Cleanup

on:
  schedule:
    - cron: "0 2 * * 0" # Weekly on Sunday at 2 AM
  workflow_dispatch:

jobs:
  cleanup:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: "1.25"

      - name: Install Clean Wizard
        run: |
          go install github.com/LarsArtmann/clean-wizard@latest

      - name: Initialize Configuration
        run: |
          clean-wizard init --minimal

      - name: Scan System
        run: |
          clean-wizard scan --validation-level comprehensive

      - name: Perform Cleanup
        run: |
          clean-wizard clean --profile comprehensive --validation-level strict
```

## 🔍 Debug Mode

For troubleshooting, use verbose mode and dry-run:

```bash
# Enable maximum verbosity
clean-wizard clean --verbose --dry-run --validation-level strict

# Check configuration loading
CLEAN_WIZARD_CONFIG=./debug-config.yaml clean-wizard config show --verbose

# Test specific operation
clean-wizard clean --profile daily --dry-run --verbose 2>&1 | tee debug.log
```

## 📚 Additional Resources

- [HOW_TO_USE.md](HOW_TO_USE.md) - Step-by-step usage guide
- [WHAT_THIS_PROJECT_IS_NOT.md](WHAT_THIS_PROJECT_IS_NOT.md) - Project limitations
- [README.md](README.md) - Project overview and installation
- Repository: https://github.com/LarsArtmann/clean-wizard
- Issues: https://github.com/LarsArtmann/clean-wizard/issues
