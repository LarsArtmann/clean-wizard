# How to Use Clean Wizard

A comprehensive step-by-step guide to using Clean Wizard effectively for system cleanup.

## 🎯 Prerequisites

- **Go 1.25+** installed on your system
- **Nix package manager** (if using Nix cleanup features)
- **macOS or Linux** system
- **Basic terminal knowledge**

## 🚀 Quick Start Guide

### Step 1: Installation

```bash
# Option 1: Install from source
go install github.com/LarsArtmann/clean-wizard@latest

# Option 2: Build locally
git clone https://github.com/LarsArtmann/clean-wizard.git
cd clean-wizard
go build -o clean-wizard ./cmd/clean-wizard
sudo mv clean-wizard /usr/local/bin/  # Optional: add to PATH
```

### Step 2: Initial Setup

```bash
# Initialize your configuration
clean-wizard init
```

This creates an interactive setup wizard that guides you through:

- Safety preferences
- Disk usage limits
- Protected paths configuration
- Default cleaning profiles

### Step 3: Scan Your System

```bash
# Basic scan to see what can be cleaned
clean-wizard scan

# Verbose scan with detailed information
clean-wizard scan --verbose

# Scan with specific validation level
clean-wizard scan --validation-level comprehensive
```

### Step 4: Perform First Cleanup

```bash
# Always start with dry run to see what will happen
clean-wizard clean --dry-run

# If satisfied, perform actual cleanup
clean-wizard clean

# Use a specific profile
clean-wizard clean --profile daily
```

## 🔧 Common Workflows

### Daily Cleanup Routine

```bash
# Quick daily maintenance
clean-wizard clean --profile daily

# With safety checks
clean-wizard clean --profile daily --validation-level comprehensive
```

### Weekly Comprehensive Cleanup

```bash
# Scan first to see what needs cleaning
clean-wizard scan --verbose

# Comprehensive cleanup
clean-wizard clean --profile comprehensive

# Review results
clean-wizard profile show comprehensive
```

### Monthly Deep Clean

```bash
# Check disk usage first
df -h

# Aggressive cleanup with confirmation
clean-wizard clean --profile aggressive --validation-level strict

# Verify results
clean-wizard scan
```

## 📊 Understanding Scan Results

When you run `clean-wizard scan`, you'll see output like:

```
🔍 Scanning system...
✅ Scan completed!
📦 Nix Store: 2.3 GB cleanable
🍺 Homebrew: 150 MB cleanable
📁 Package Caches: 500 MB cleanable
💡 Total: ~3 GB can be recovered
```

### Interpreting the Results

- **Nix Store**: Old Nix package generations
- **Homebrew**: Homebrew cache and old versions
- **Package Caches**: npm, cargo, go module caches
- **Total**: Estimated space you can recover

## 🎨 Using Profiles

### Built-in Profiles

1. **Daily Profile** (`daily`)
   - Quick, safe cleanup
   - Keeps last 3 Nix generations
   - Homebrew autoremove
   - Perfect for routine maintenance

2. **Comprehensive Profile** (`comprehensive`)
   - Weekly cleanup
   - More aggressive Nix cleanup
   - All package managers
   - System temporary files

3. **Aggressive Profile** (`aggressive`)
   - Monthly deep clean
   - Maximum space recovery
   - All cache types
   - Requires confirmation

### Using Profiles

```bash
# List available profiles
clean-wizard profile list

# Show profile details
clean-wizard profile show daily

# Use a profile
clean-wizard clean --profile daily

# Override profile dry-run setting
clean-wizard clean --profile comprehensive --dry-run
```

## 🔒 Safety Best Practices

### Always Use Dry Run First

```bash
# Check what will be deleted
clean-wizard clean --dry-run --verbose

# Only then run actual cleanup
clean-wizard clean
```

### Use Validation Levels

```bash
# Basic validation (default)
clean-wizard clean --validation-level basic

# Comprehensive validation (recommended)
clean-wizard clean --validation-level comprehensive

# Strict validation (highest safety)
clean-wizard clean --validation-level strict
```

### Check Protected Paths

```bash
# View current protected paths
clean-wizard config show

# Ensure important directories are protected
# Common protected paths:
# - /Users
# - /System
# - /Applications
# - /Library
# - /nix/store
```

## 📁 Configuration Management

### Creating Custom Profiles

```bash
# Create a new profile interactively
clean-wizard profile create

# Example profile configuration:
# custom-profile.yaml
profiles:
  developer:
    description: "Developer workspace cleanup"
    operations:
      - name: "nix-generations"
        enabled: true
        settings:
          generations: 5
      - name: "go-cache"
        enabled: true
      - name: "node-modules"
        enabled: false  # Keep node modules
```

### Using Configuration Files

```bash
# Use specific configuration file
clean-wizard clean --config /path/to/config.yaml

# Validate configuration
clean-wizard config validate

# Edit configuration
clean-wizard config edit
```

## 🐛 Troubleshooting

### Common Issues

1. **Permission Denied**

   ```bash
   # Check if you have sufficient permissions
   clean-wizard scan --verbose

   # Some operations may require sudo
   sudo clean-wizard clean --profile aggressive
   ```

2. **Configuration Not Found**

   ```bash
   # Initialize configuration
   clean-wizard init

   # Check default location
   ls -la ~/.clean-wizard.yaml
   ```

3. **Dry Run Shows Nothing**

   ```bash
   # Try more comprehensive scan
   clean-wizard scan --validation-level comprehensive

   # Check if you're already clean
   clean-wizard clean --dry-run --verbose
   ```

### Getting Help

```bash
# General help
clean-wizard --help

# Command-specific help
clean-wizard clean --help
clean-wizard scan --help
clean-wizard profile --help

# Verbose output for debugging
clean-wizard clean --verbose --dry-run
```

## 📈 Advanced Usage

### Custom Workflows

```bash
# Create a script for regular maintenance
#!/bin/bash
# daily-cleanup.sh

echo "🧹 Starting daily cleanup..."
clean-wizard scan
clean-wizard clean --profile daily --dry-run

read -p "Continue with cleanup? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    clean-wizard clean --profile daily
    echo "✅ Daily cleanup complete!"
fi
```

### Integration with Cron

```bash
# Add to crontab for automated cleanup
# Edit crontab
crontab -e

# Add weekly comprehensive cleanup
0 2 * * 0 /usr/local/bin/clean-wizard clean --profile comprehensive --validation-level strict

# Add daily quick cleanup
0 1 * * 1-6 /usr/local/bin/clean-wizard clean --profile daily
```

### Monitoring Results

```bash
# Create a log of cleanups
clean-wizard clean --verbose | tee cleanup-$(date +%Y%m%d).log

# Track disk space over time
df -h | tee disk-usage-$(date +%Y%m%d).log
```

## 🎯 Tips for Effective Usage

1. **Start conservative** - Begin with daily profile and dry runs
2. **Monitor results** - Keep track of what gets cleaned
3. **Customize profiles** - Adapt to your specific needs
4. **Regular schedule** - Set up automated cleanups
5. **Backup first** - Ensure important data is backed up
6. **Check permissions** - Some operations may require elevated privileges

## 🔗 Next Steps

- Read [USAGE.md](USAGE.md) for detailed command reference
- Check [WHAT_THIS_PROJECT_IS_NOT.md](WHAT_THIS_PROJECT_IS_NOT.md) for limitations
- Review [README.md](README.md) for project overview
- Explore configuration examples in the repository
