# Clean Wizard vs SystemNix: Feature Comparison Report

**Generated:** 2026-02-09 11:24  
**Project:** Clean Wizard (clean-wizard)  
**Comparison Target:** SystemNix (justfile)  
**Purpose:** Comprehensive feature gap analysis

---

## Executive Summary

This report provides a detailed comparison between **Clean Wizard** (a Go-based system cleanup tool) and **SystemNix** (a NixOS/macOS configuration manager with comprehensive cleanup commands). Both tools aim to clean system caches and temporary files, but differ significantly in architecture, scope, and execution strategy.

**Key Findings:**
- SystemNix provides more aggressive, complete cleanup with deeper Nix integration
- Clean Wizard offers superior architecture (type-safe Go) but has critical NO-OP implementations
- ~20% of Clean Wizard cleaners are non-functional placeholders
- SystemNix covers more edge cases (iOS simulators, Lima VMs, NuGet)
- Both agree on "Quick = Skip Nix + Docker + System"

---

## 1. Preset Mode Comparison

### Quick Mode

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Target** | Daily cache cleanup | Fast cleanup without system changes |
| **Homebrew** | ✅ `brew autoremove && brew cleanup` | ✅ Homebrew |
| **npm** | ✅ `npm cache clean --force` | ✅ Node Packages |
| **pnpm** | ✅ `pnpm store prune` | ✅ Node Packages |
| **Go** | ✅ `go clean -cache` | ✅ Go Packages |
| **Temp Files** | ✅ `/tmp/nix-*` | ✅ TempFiles |
| **Build Cache** | ❌ Not included | ✅ BuildCache |
| **Nix** | ❌ Explicitly excluded | ❌ Explicitly excluded |
| **Docker** | ✅ Light prune `docker system prune -f` | ❌ Excluded |
| **Xcode** | ❌ Not included | ❌ Excluded |
| **Safety Message** | ✅ "No Nix store changes" | ✅ Comment: "no Nix/Docker/System changes" |

### Standard/Full Mode

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Target** | Comprehensive system cleanup | All available cleaners |
| **Nix Generations** | ✅ `nix-collect-garbage -d --delete-older-than 1d` | ✅ Configurable keep count |
| **Nix Store Optimization** | ✅ `nix-store --optimize` | ❌ Not implemented |
| **Nix Profile Cleanup** | ✅ `nix profile wipe-history` | ❌ Not implemented |
| **Docker** | ✅ Full prune `docker system prune -af` | ✅ Full prune |
| **System Cache** | ✅ Spotlight, Xcode, Homebrew | ✅ Spotlight, Xcode, CocoaPods, Homebrew |
| **Cargo** | ✅ `cargo cache --autoclean` | ✅ `cargo clean` |
| **Go** | ✅ `go clean -cache -testcache -modcache` | ✅ + lint cache |
| **Temp Files** | ✅ `/tmp/nix-build-*` | ✅ Age-based configurable |
| **Language Managers** | ❌ Not included | ⚠️ Scans only (NO-OP) |

### Aggressive/Nuclear Mode

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Confirmation** | ⚠️ Interactive pause | ⚠️ Default prompt (not explicit) |
| **Nix** | ✅ ALL generations (no time threshold) | ✅ All available |
| **Nix Profiles** | ✅ All wiped | ❌ Not implemented |
| **Language Versions** | ✅ NVM, Pyenv, Rbenv all deleted | ⚠️ NO-OP (prints warning) |
| **Build Caches** | ✅ Full `~/.cache` wipe | ✅ Via BuildCache cleaner |
| **Xcode** | ✅ Full DerivedData wipe | ✅ Via SystemCache cleaner |
| **Docker** | ✅ With volumes `--volumes` | ✅ With volumes |
| **iOS Simulators** | ✅ `xcrun simctl delete all` | ❌ Not implemented |

---

## 2. Package Manager Coverage

### Node.js Package Managers

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **npm** | ✅ `npm cache clean --force` | ✅ `npm cache clean --force` |
| **pnpm** | ✅ `pnpm store prune` | ✅ `pnpm store prune` |
| **yarn** | ✅ `yarn cache clean` | ✅ `yarn cache clean` |
| **bun** | ✅ `rm -rf ~/.bun/install/cache` | ✅ `bun pm cache rm` |
| **Detection** | Assumes installed | ✅ Dynamic availability check |
| **Error Handling** | ⚠️ Silent continue | ✅ Graceful degradation |

### Go and Rust

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **Go Cache** | ✅ `go clean -cache -testcache -modcache` | ✅ `go clean -cache -testcache -modcache` |
| **Go Build Cache** | ✅ `find /var/folders... -name "go-build*"` | ✅ Via GOMODCACHE |
| **Go Lint Cache** | ❌ Not included | ✅ `golangci-lint` cache |
| **Cargo** | ✅ `cargo cache --autoclean` | ✅ `cargo clean` + cargo-cache |
| **Cargo Registry** | ✅ `~/.cargo/registry` | ✅ `~/.cargo/registry` |
| **Cargo Git** | ✅ `~/.cargo/git` | ✅ `~/.cargo/git` |

### Build Tools

| Tool | SystemNix | Clean Wizard |
|------|-----------|--------------|
| **Gradle** | ✅ `rm -rf ~/.gradle/caches/*` | ✅ Gradle support |
| **Maven** | ❌ Not included | ✅ Maven support (`.part` files) |
| **SBT** | ❌ Not included | ✅ SBT support (Ivy cache) |
| **Puppeteer** | ✅ `rm -rf ~/.cache/puppeteer` | ❌ Not implemented |
| **NuGet** | ✅ `rm -rf ~/.nuget/packages` | ❌ Not implemented |
| **Lima VM** | ✅ `rm -rf ~/Library/Caches/lima` | ❌ Not implemented |

---

## 3. System Cache Cleanup

### macOS System Caches

| Cache | SystemNix | Clean Wizard | Status |
|-------|-----------|--------------|--------|
| **Spotlight** | ✅ `rm -r ~/Library/Metadata/CoreSpotlight/...` | ✅ `rm -r ~/Library/Metadata/CoreSpotlight/...` | ✅ Matching |
| **Xcode DerivedData** | ✅ `rm -rf ~/Library/Developer/Xcode/DerivedData` | ✅ `rm -rf ~/Library/Developer/Xcode/DerivedData` | ✅ Matching |
| **CocoaPods** | ❌ Not included | ✅ `rm -rf ~/Library/Caches/CocoaPods` | ✅ Clean Wizard only |
| **Homebrew Cache** | ✅ `rm -rf ~/Library/Caches/Homebrew` | ✅ `rm -rf ~/Library/Caches/Homebrew` | ✅ Matching |
| **iOS Simulators** | ✅ `xcrun simctl delete unavailable` | ❌ Not implemented | ⚠️ SystemNix only |
| **Lima Cache** | ✅ `rm -rf ~/Library/Caches/lima` | ❌ Not implemented | ⚠️ SystemNix only |

### Platform Detection

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **macOS Support** | ✅ Always | ✅ Works |
| **Linux Support** | ⚠️ Assumes paths exist | ⚠️ Broken (env vars only) |
| **Detection Method** | ✅ `uname` check | ❌ `GOOS`/`OSTYPE` env vars |
| **Safety** | ❌ Silent failure if paths missing | ⚠️ Fragile detection |

---

## 4. Docker Cleanup Details

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **Command** | `docker system prune -af` | `docker system prune -af --volumes` |
| **Light Mode** | ✅ `docker system prune -f` | ❌ Not implemented |
| **Aggressive** | ✅ `--volumes` | ✅ `--volumes` |
| **Timeout** | ❌ Not specified | ✅ 2-minute timeout |
| **Size Parsing** | ✅ Parses freed bytes | ❌ Broken (returns 0) |
| **Running Check** | ✅ Implicit | ✅ Explicit |
| **Dangling Images** | ✅ Included | ✅ Included |
| **Stopped Containers** | ✅ Included | ✅ Included |
| **Unused Volumes** | ✅ Included | ✅ Included |
| **Build Cache** | ✅ Included | ✅ Included |

---

## 5. Language Version Manager Handling

| Manager | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **NVM (Node)** | ✅ Deletes `~/.nvm/versions/node/*` | ⚠️ Scans only, NO-OP |
| **Pyenv (Python)** | ✅ Deletes `~/.pyenv/versions/*` | ⚠️ Scans only, NO-OP |
| **Rbenv (Ruby)** | ✅ Deletes `~/.rbenv/versions/*` | ⚠️ Scans only, NO-OP |
| **GVM (Go)** | ❌ Not included | ⚠️ In enum, not implemented |
| **SDKMAN (Java)** | ❌ Not included | ⚠️ In enum, not implemented |
| **Jenv (Java)** | ❌ Not included | ⚠️ In enum, not implemented |

### Critical Issue: Clean Wizard Language Version Manager

**Location:** `internal/cleaner/langversionmanager.go:133-154`

```go
// This is a NO-OP by default to avoid destructive behavior
// Comment in code explicitly acknowledges the issue
```

**Impact:** Cleaner scans for old versions but NEVER deletes them.
- **Scan Operation:** ✅ Returns list of found versions
- **Clean Operation:** ❌ Returns `(FreedBytes: 0, Warning: "This is a NO-OP...")`

---

## 6. Nix Store Cleanup Comparison

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **Garbage Collection** | `nix-collect-garbage -d` | `nix-collect-garbage -d` |
| **Time-Based Delete** | ✅ `--delete-older-than 1d` | ❌ Not implemented |
| **Count-Based Keep** | ❌ Not available | ✅ Configurable (default: 5) |
| **Store Optimization** | ✅ `nix-store --optimize` | ❌ Not implemented |
| **Profile Wipe** | ✅ `nix profile wipe-history` | ❌ Not implemented |
| **Dry-Run Support** | ❌ Not available | ✅ Estimates 50MB/generation |
| **Safety** | Current generation protected | Current generation protected |
| **Mock Data** | ❌ Real commands only | ✅ Mock in CI/testing |
| **Size Calculation** | ✅ `du -sh /nix/store` | ⚠️ Hardcoded estimates |

---

## 7. Temporary Files Cleanup

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Nix Build Temp** | ✅ `/tmp/nix-build-*` | ❌ Not implemented |
| **Nix Shell Temp** | ✅ `/tmp/nix-shell-*` | ❌ Not implemented |
| **General Temp** | ❌ Not included | ✅ Age-based, configurable |
| **Exclusion Patterns** | ❌ Not configurable | ✅ Prefix-based exclusions |
| **Directory Preservation** | ⚠️ Risk of deletion | ✅ Files only, dirs preserved |
| **Safety Mechanism** | ⚠️ Minimal | ✅ Multiple safety checks |
| **Custom Paths** | ❌ Hardcoded | ✅ Custom base paths supported |

---

## 8. Safety and UX Features

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **Dry-Run Mode** | ❌ Not available | ✅ `--dry-run` flag |
| **Verbose Mode** | ❌ Not available | ✅ `--verbose` flag |
| **JSON Output** | ❌ Not available | ✅ `--json` flag |
| **Interactive TUI** | ❌ Command line only | ✅ Charm Huh forms |
| **Multi-Select** | ❌ All-or-nothing | ✅ Select multiple cleaners |
| **Confirmation Prompt** | ⚠️ Only aggressive | ✅ Yes/No before execution |
| **Size Reporting** | ✅ Before/after with `du` | ⚠️ Hardcoded estimates |
| **Progress Display** | ❌ Linear output | ✅ Per-cleaner progress |
| **Result Aggregation** | ❌ Manual calculation | ✅ Totals across cleaners |
| **Error Handling** | ❌ Silent continue | ✅ Graceful degradation |
| **Availability Detection** | ❌ Assumes installed | ✅ Shows only available |

---

## 9. Architectural Comparison

| Aspect | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Language** | Justfile (shell script) | Go (compiled binary) |
| **Extensibility** | Add commands to justfile | Registry pattern (unwired plugins) |
| **Type Safety** | ❌ Shell strings | ✅ Type-safe enums (compile-time) |
| **Registry Pattern** | ❌ Manual | ✅ Cleaner registry |
| **Configuration** | ❌ Hardcoded | ✅ YAML config (partially wired) |
| **Testing** | ❌ Manual | ✅ 200+ unit tests, BDD |
| **Cross-Platform** | ✅ POSIX shell | ⚠️ macOS only (Linux broken) |
| **Dependency Management** | Nix flake | Go modules |
| **Binary Size** | N/A (justfile) | ~10MB+ |

### Code Quality Metrics

| Metric | SystemNix | Clean Wizard |
|--------|-----------|--------------|
| **Test Count** | 0 | 200+ |
| **BDD Tests** | ❌ | ✅ Godog-based |
| **Coverage** | N/A | Moderate |
| **Linting** | shellcheck | golangci-lint |
| **Type Coverage** | ❌ | 100% (no `any` types) |

---

## 10. Feature Gap Matrix

### Clean Wizard has that SystemNix lacks

| Feature | Clean Wizard | SystemNix |
|---------|--------------|-----------|
| **Type-Safe Enums** | ✅ Compile-time safety | ❌ |
| **Registry Pattern** | ✅ Centralized cleaner management | ❌ |
| **Dry-Run Mode** | ✅ Preview before cleaning | ❌ |
| **Interactive TUI** | ✅ Beautiful forms | ❌ |
| **JSON Output** | ✅ Machine-readable results | ❌ |
| **Go Lint Cache** | ✅ `golangci-lint` cache cleaning | ❌ |
| **Configuration System** | ✅ YAML profiles (wired) | ❌ |
| **Maven Cleanup** | ✅ `.part` file removal | ❌ |

### SystemNix has that Clean Wizard lacks

| Feature | SystemNix | Clean Wizard |
|---------|-----------|--------------|
| **Nix Store Optimization** | ✅ `nix-store --optimize` | ❌ |
| **Nix Profile Cleanup** | ✅ `nix profile wipe-history` | ❌ |
| **iOS Simulator Cleanup** | ✅ `xcrun simctl delete` | ❌ |
| **Lima VM Cache** | ✅ `~/Library/Caches/lima` | ❌ |
| **Puppeteer Cache** | ✅ `~/.cache/puppeteer` | ❌ |
| **NuGet Packages** | ✅ `~/.nuget/packages` | ❌ |
| **Aggressive Confirmation** | ✅ Explicit pause | ⚠️ |
| **Size Before/After** | ✅ `du -sh` | ⚠️ Estimates |

### Both have (matching implementations)

| Feature | Status |
|---------|--------|
| **Nix GC** | ✅ Matching commands |
| **Homebrew cleanup** | ✅ Matching commands |
| **Docker prune** | ✅ Matching commands |
| **npm/pnpm/yarn/bun** | ✅ Matching commands |
| **Go cache clean** | ✅ SystemNix has more paths |
| **Cargo clean** | ✅ Matching commands |
| **Spotlight cleanup** | ✅ Matching paths |
| **Xcode DerivedData** | ✅ Matching paths |
| **Homebrew cache** | ✅ Matching paths |

---

## 11. Critical Issues Summary

### Clean Wizard Critical Issues

| Issue | Severity | Location |
|-------|----------|----------|
| **Language Version Manager is NO-OP** | 🔴 CRITICAL | `langversionmanager.go:133-154` |
| **Docker size reporting broken** | 🟠 HIGH | `docker.go` (returns 0) |
| **Dry-run uses hardcoded estimates** | 🟡 MEDIUM | Most cleaners |
| **Projects Mgmt requires external tool** | 🟡 MEDIUM | External dependency |
| **CLI commands missing** | 🟡 MEDIUM | Only `clean` implemented |
| **Linux platform detection broken** | 🟡 MEDIUM | Env vars vs runtime check |
| **Unused enum values** | 🟢 LOW | BuildToolType, CacheType, etc. |

### SystemNix Critical Issues

| Issue | Severity | Location |
|-------|----------|----------|
| **No dry-run mode** | 🟡 MEDIUM | Justfile limitation |
| **No size reporting** | 🟡 MEDIUM | Linear output only |
| **No type safety** | 🟢 LOW | Shell scripting |
| **Hard to extend** | 🟢 LOW | Justfile + scripts |

---

## 12. Recommendations

### For Clean Wizard Users

| Priority | Recommendation | Rationale |
|----------|----------------|-----------|
| **P1** | Use for Nix, Homebrew, Docker, Go, Node | Production-ready cleaners |
| **P2** | Avoid Language Version Manager | NO-OP implementation |
| **P3** | Don't rely on size estimates | Hardcoded, inaccurate |
| **P4** | Avoid `clean --mode quick` | Same exclusions as SystemNix |

### For Clean Wizard Contributors

| Priority | Task | Effort | Impact |
|----------|------|--------|--------|
| **P1** | Implement Language Version Manager cleaning | HIGH | CRITICAL |
| **P2** | Add Nix store optimization | MEDIUM | HIGH |
| **P3** | Fix Docker size reporting | LOW | HIGH |
| **P4** | Add iOS simulator cleanup | MEDIUM | MEDIUM |
| **P5** | Wire remaining CLI commands | HIGH | HIGH |

### For SystemNix Users

| Priority | Recommendation | Rationale |
|----------|----------------|-----------|
| **P1** | Use for aggressive Nix cleanup | Best implementation |
| **P2** | Use for language version manager cleanup | Only tool that deletes |
| **P3** | Use for iOS/Lima cleanup | Unique features |
| **P4** | Consider adding dry-run | Safety improvement |

---

## 13. Conclusion

### Architectural Winner: Clean Wizard

Clean Wizard provides superior architecture with:
- Type-safe Go implementation
- Registry pattern for cleaner management
- Interactive TUI with progress tracking
- Dry-run mode for safety
- JSON output for automation
- 200+ comprehensive tests

### Functional Winner: SystemNix

SystemNix provides more complete cleanup with:
- Actual deletion of language versions
- Nix store optimization
- Nix profile cleanup
- iOS simulator management
- Lima VM cache cleanup
- Puppeteer and NuGet support
- Better disk space reporting

### When to Use Which

| Use Case | Recommended Tool |
|----------|-----------------|
| **Nix-centric cleanup** | SystemNix (optimization + profiles) |
| **Package manager caches** | Tie (both cover similar ground) |
| **Language version managers** | SystemNix (actually deletes) |
| **Type-safe architecture** | Clean Wizard (Go type system) |
| **Aggressive cleanup** | SystemNix (destructive by design) |
| **Safety-first cleanup** | Clean Wizard (dry-run, conservative) |
| **Interactive UX** | Clean Wizard (TUI) |
| **Scriptable automation** | SystemNix (justfile) |
| **Testing** | Clean Wizard (200+ tests) |
| **Cross-platform** | Neither (both macOS-focused) |

### Final Assessment

Both tools have distinct strengths and target different use cases:
- **SystemNix** is a "kitchen sink" approach with aggressive, complete cleanup
- **Clean Wizard** is architecturally superior but functionally incomplete

**Recommendation:** Run SystemNix for comprehensive cleanup, use Clean Wizard's code architecture as a reference for future development.

---

*Report generated on 2026-02-09 11:24*  
*Clean Wizard version: Based on codebase analysis*  
*SystemNix version: Justfile (latest)*

---

## Appendix A: File References

### Clean Wizard Files Analyzed

| File | Purpose |
|------|---------|
| `FEATURES.md` | Comprehensive feature documentation |
| `internal/cleaner/registry_factory.go` | Default registry with all cleaners |
| `internal/cleaner/langversionmanager.go:133-154` | NO-OP implementation |
| `cmd/clean-wizard/commands/clean.go:597-623` | Preset mode logic |
| `internal/config/config.go` | YAML configuration system |

### SystemNix Files Analyzed

| File | Purpose |
|------|---------|
| `justfile` | Complete command definitions |
| `clean` recipe (lines 70-122) | Comprehensive cleanup |
| `clean-quick` recipe (lines 125-137) | Daily quick cleanup |
| `clean-aggressive` recipe (lines 140-187) | Nuclear cleanup |

---

## Appendix B: Test Commands

### Clean Wizard Quick Reference

```bash
# Quick mode (no Nix/Docker/System)
clean-wizard clean --mode quick

# Standard mode (all available)
clean-wizard clean

# Aggressive mode (all cleaners)
clean-wizard clean --mode aggressive

# With dry-run
clean-wizard clean --dry-run

# With JSON output
clean-wizard clean --json

# Verbose logging
clean-wizard clean --verbose
```

### SystemNix Quick Reference

```bash
# Quick daily cleanup (no Nix store changes)
just clean-quick

# Comprehensive cleanup
just clean

# Aggressive cleanup (confirmation required)
just clean-aggressive

# Deep clean (includes build caches)
just deep-clean
```

---

## Appendix C: Enum Implementation Status

### Clean Wizard Domain Enums

| Enum | Values Defined | Values Used | Gap |
|------|---------------|-------------|-----|
| **BuildToolType** | 6 (GO, RUST, NODE, PYTHON, JAVA, SCALA) | 2 (JAVA, SCALA) | 4 unused |
| **CacheType** | 8 (SPOTLIGHT, XCODE, COCOAPODS, HOMEBREW, PIP, NPM, YARN, CCACHE) | 4 (first 4) | 4 unused |
| **VersionManagerType** | 6 (NVM, PYENV, GVM, RBENV, SDKMAN, JENV) | 3 (NVM, PYENV, RBENV) | 3 unused, 1 NO-OP |
| **DockerPruneMode** | 5 (ALL, IMAGES, CONTAINERS, VOLUMES, BUILDS) | 1 (ALL) | 4 unused |

### SystemNix Enums

| Enum Type | Implementation |
|-----------|----------------|
| **Cleanup Types** | Implicit in shell conditions |
| **Risk Levels** | None (all commands run) |
| **Modes** | Quick/Standard/Aggressive (explicit) |

---

*End of Report*