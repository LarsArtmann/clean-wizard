# What This Project Is Not

Setting clear expectations about Clean Wizard's scope and limitations.

## 🚫 Important Disclaimers

Clean Wizard is a **specialized system cleanup tool** with a focused purpose. Understanding what it is **not** designed to do is as important as understanding what it can do.

---

## 🔒 Clean Wizard is NOT a Backup Solution

### What It Doesn't Do
- ❌ **Does not create system backups**
- ❌ **Does not backup personal files**
- ❌ **Does not provide version control for your data**
- ❌ **Does not protect against data loss from hardware failure**

### What You Should Use Instead
- ✅ **Time Machine** (macOS) for full system backups
- ✅ **rsync** or **Borg Backup** for custom backup solutions
- ✅ **Cloud storage** (Dropbox, Google Drive, iCloud) for file sync
- ✅ **Git** for code and configuration files

### Best Practice
```bash
# Always have backups before aggressive cleanup
# Clean Wizard complements, but never replaces, proper backup strategy

# Check your backup status first
tmutil listlocalsnapshots /
# Then use Clean Wizard
clean-wizard clean --profile aggressive --dry-run
```

---

## ⚡ Clean Wizard is NOT a System Optimizer

### What It Doesn't Do
- ❌ **Does not optimize system performance**
- ❌ **Does not speed up boot times**
- ❌ **Does not defragment drives**
- ❌ **Does not optimize memory usage**
- ❌ **Does not clean registry (macOS doesn't have one)**
- ❌ **Does not improve CPU performance**

### What It Actually Does
- ✅ **Frees disk space by removing unnecessary files**
- ✅ **Cleans package caches and temporary data**
- ✅ **Removes old package manager generations**
- ✅ **May indirectly improve some operations by reducing clutter**

### For System Optimization, Consider
- ✅ **Hardware upgrades** (SSD, more RAM)
- ✅ **System updates** (macOS updates)
- ✅ **Activity Monitor** for identifying resource hogs
- ✅ **Proper system maintenance** beyond just cleanup

---

## 🖥️ Clean Wizard is NOT a Comprehensive System Maintenance Tool

### What It Doesn't Do
- ❌ **Does not update system software**
- ❌ **Does not manage permissions**
- ❌ **Does not repair file systems**
- ❌ **Does not manage user accounts**
- ❌ **Does not handle system security**
- ❌ **Does not manage startup programs**

### Missing Maintenance Tasks
```bash
# Clean Wizard handles disk space cleanup, but you still need:

# System updates
softwareupdate --install --all

# File system checks
diskutil verifyVolume /

# Permission repairs (when needed)
diskutil repairPermissions /

# Security updates
sudo softwareupdate --install --all --restart
```

---

## 🏢 Clean Wizard is NOT for Production Server Cleanup

### Why It's Not Suitable for Production
- ❌ **Not designed for automated server environments**
- ❌ **No remote management capabilities**
- ❌ **Interactive-first design approach**
- ❌ **Not built for high-availability systems**
- ❌ **Limited logging and monitoring features**

### Production Alternatives
- ✅ **Ansible** or **Chef** for configuration management
- ✅ **Cron jobs** with custom scripts
- ✅ **Docker** for containerized cleanup
- ✅ **Kubernetes** jobs for orchestrated cleanup
- ✅ **Logrotate** for log management

### Production Cleanup Example
```bash
# Use these instead for production environments:
#!/bin/bash
# production-cleanup.sh - NOT Clean Wizard

# Automated Nix cleanup
nix-collect-garbage -d

# Systemd journal cleanup
journalctl --vacuum-time=7d

# Docker cleanup
docker system prune -f

# Package manager cleanup
apt-get autoremove -y  # Debian/Ubuntu
dnf autoremove -y      # RHEL/Fedora
```

---

## 🛡️ Clean Wizard is NOT a Security or Malware Removal Tool

### What It Doesn't Do
- ❌ **Does not scan for malware or viruses**
- ❌ **Does not remove spyware**
- ❌ **Does not protect against security threats**
- ❌ **Does not analyze suspicious files**
- ❌ **Does not provide real-time protection**

### Security Tools You Should Use
- ✅ **Malwarebytes** for malware scanning
- ✅ **ClamAV** for antivirus protection
- ✅ **Little Snitch** for network monitoring
- ✅ ** macOS built-in security** (XProtect, Gatekeeper)

### Security Best Practices
```bash
# Clean Wizard + Security Tools = Complete Protection
# 1. Security scanning (monthly)
malwarebytes-scan

# 2. System updates (as available)
softwareupdate --install --all

# 3. Clean Wizard cleanup (weekly)
clean-wizard clean --profile comprehensive

# 4. Review system logs
log show --predicate 'eventMessage contains "error"' --last 1d
```

---

## 📊 Clean Wizard is NOT a Disk Space Analyzer

### What It Doesn't Do
- ❌ **Does not provide detailed disk usage analysis**
- ❌ **Does not show file type breakdowns**
- ❌ **Does not visualize disk space usage**
- ❌ **Does not track disk space over time**
- ❌ **Does not identify large files for manual review**

### What It Does Instead
- ✅ **Estimates cleanable space for specific caches**
- ✅ **Shows potential space recovery from cleanup**
- ✅ **Focuses only on what it can safely remove**

### For Disk Analysis, Use These Tools
- ✅ **DaisyDisk** - Visual disk space analyzer
- ✅ **GrandPerspective** - File size visualization
- ✅ **ncdu** - Terminal-based disk usage analyzer
- ✅ **Disk Inventory X** - Free disk space analyzer

### Disk Analysis Workflow
```bash
# Analyze disk space first
ncdu ~/

# Then use Clean Wizard to cleanup
clean-wizard scan --verbose

# Remove specific large files manually
# Use Clean Wizard for cache cleanup
clean-wizard clean --dry-run
```

---

## 🔧 Clean Wizard is NOT a One-Click Solution

### What It Doesn't Do
- ❌ **Does not magically solve all disk space issues**
- ❌ **Does not work without user understanding**
- ❌ **Does not adapt to every unique system**
- ❌ **Does not eliminate the need for system management**

### What You Still Need to Do
- ✅ **Understand what's being cleaned**
- ✅ **Review configuration settings**
- ✅ **Backup important data**
- ✅ **Monitor results**
- ✅ **Customize profiles for your needs**

### Mindset for Using Clean Wizard
```bash
# Wrong approach: "Click and forget"
clean-wizard clean --aggressive --force

# Right approach: "Understand and verify"
clean-wizard scan --verbose
clean-wizard clean --profile daily --dry-run
# Review output, then:
clean-wizard clean --profile daily
```

---

## 📁 Clean Wizard is NOT for Custom File Management

### What It Doesn't Do
- ❌ **Does not clean your Documents folder**
- ❌ **Does not organize your files**
- ❌ **Does not find duplicate files**
- ❌ **Does not manage photo libraries**
- ❌ **Does not clean download folders automatically**

### File Management Tools
- ✅ **Finder** (macOS) for manual file management
- ✅ **Hazel** for automated file organization
- ✅ **Gemini 2** for duplicate file finding
- ✅ **Photo library tools** for media management

### Manual File Management Workflow
```bash
# Clean Wizard handles system caches only
clean-wizard clean

# You handle personal files manually
# or with specialized tools:
# - ~/Downloads cleanup
# - Duplicate file removal
# - Document organization
# - Photo library maintenance
```

---

## 🐧 Clean Wizard is NOT Cross-Platform (Yet)

### Current Limitations
- ❌ **Does not work on Windows**
- ❌ **Limited Windows support (no Windows package managers)**
- ❌ **Does not handle Linux package managers other than Nix**

### Platform Coverage
- ✅ **macOS** - Full support with Homebrew
- ✅ **Linux** - Basic support with Nix
- ❌ **Windows** - Not supported
- ❌ **BSD** - Not supported
- ❌ **Other Unix variants** - Limited support

### For Windows Users
```powershell
# Windows equivalent cleanup tools:
# - Windows Disk Cleanup (built-in)
# - PowerShell cleanup scripts
# - Third-party tools like CCleaner

# Example PowerShell cleanup:
Remove-Item $env:TEMP\* -Recurse -Force
Get-ChildItem -Path $env:LOCALAPPDATA\Temp | Remove-Item -Recurse -Force
```

---

## 🚨 Clean Wizard is NOT Infinitely Safe

### Safety Limitations
- ❌ **Cannot guarantee 100% safety**
- ❌ **May remove files you actually need**
- ❌ **Does not understand every possible use case**
- ❌ **Configuration mistakes can cause issues**

### Safety Measures You Should Take
- ✅ **Always use dry-run first**
- ✅ **Review protected paths configuration**
- ✅ **Backup before aggressive operations**
- ✅ **Start with conservative settings**
- ✅ **Monitor results and adjust**

### Safety Checklist
```bash
# Before aggressive cleanup:
# 1. Backup your data
tmutil startbackup

# 2. Check protected paths
clean-wizard config show

# 3. Test with dry run
clean-wizard clean --profile aggressive --dry-run --verbose

# 4. Review each operation carefully
# 5. Only then run actual cleanup
clean-wizard clean --profile aggressive
```

---

## 🎯 The Bottom Line

Clean Wizard is a **focused, specialized tool** designed for one primary purpose: **safe cleanup of system caches and package manager artifacts**.

### ✅ What Clean Wizard IS
- A safe way to clean Nix generations
- A tool for package manager cache cleanup
- A framework for configurable cleaning profiles
- A safety-first approach to system cleanup
- A CLI tool for terminal users

### ❌ What Clean Wizard is NOT
- A backup solution
- A system optimizer
- A comprehensive maintenance tool
- A production server solution
- A security tool
- A disk space analyzer
- A one-click fix
- A cross-platform solution
- Infinitely safe

### 🔄 How to Use Clean Wizard Effectively

1. **Understand its scope** - Use it for what it's designed for
2. **Combine with other tools** - It's part of a complete toolkit
3. **Start conservative** - Begin with safe operations
4. **Always verify** - Use dry-run and review results
5. **Customize wisely** - Adapt profiles to your needs
6. **Backup first** - Never skip the safety steps

---

## 📚 Related Documentation

- [HOW_TO_USE.md](HOW_TO_USE.md) - Proper usage guide
- [USAGE.md](USAGE.md) - Complete command reference
- [README.md](README.md) - Project overview and features

## 🔗 External Resources

### For System Maintenance
- [macOS User Guide](https://support.apple.com/guide/mac-help)
- [Nix Package Manager](https://nixos.org/manual/nix/)
- [Homebrew Documentation](https://docs.brew.sh/)

### For System Optimization
- [Apple's Performance Guide](https://support.apple.com/en-us/HT201757)
- [Understanding macOS Performance](https://developer.apple.com/documentation/)

### For Backup Solutions
- [Time Machine User Guide](https://support.apple.com/guide/mac-help/use-time-machine-to-back-up-or-restore-your-mac-mh104931/mac)
- [macOS Backup Solutions](https://macpaw.com/blog/mac-backup-software)

---

**Remember:** Clean Wizard is a valuable tool, but it's just **one piece** of your system maintenance toolkit. Use it wisely, in combination with proper backups, security tools, and system monitoring for a complete solution.