# Clean Wizard: Path to SystemNix Parity

**Analysis Date:** 2026-02-09  
**Goal:** Fully replace SystemNix `clean-quick`, `clean`, `clean-aggressive` commands  
**Current State:** ⚠️ Partial feature coverage with critical gaps

---

## Executive Summary

Clean Wizard currently has the **architectural foundation** to replace SystemNix but is missing **critical features** that would make it your "ultimate MacBook storage cleanup tool."

**Gap Summary:**
| SystemNix Command | Clean Wizard Equivalent | Feature Parity | Critical Gaps |
|-------------------|------------------------|----------------|---------------|
| `clean-quick` | `clean --mode quick` | ~85% | Docker light, Nix temp files |
| `clean` | `clean --mode standard` | ~75% | Nix optimization, Nix profiles, iOS |
| `clean-aggressive` | `clean --mode aggressive` | ~60% | Language managers (NO-OP), full Xcode |

---

## 1. Feature Gap Analysis

### 1.1 `clean-quick` Comparison

| Feature | SystemNix | Clean Wizard | Status |
|---------|-----------|--------------|--------|
| **Homebrew** | ✅ `brew autoremove && brew cleanup` | ✅ `brew cleanup` | ✅ Matching |
| **npm** | ✅ `npm cache clean --force` | ✅ `npm cache clean --force` | ✅ Matching |
| **pnpm** | ✅ `pnpm store prune` | ✅ `pnpm store prune` | ✅ Matching |
| **Go** | ✅ `go clean -cache` | ✅ `go clean -cache -testcache -modcache` | ✅ Better |
| **Temp Files** | ✅ `/tmp/nix-build-*` | ✅ Configurable age-based | ⚠️ Partial |
| **Build Cache** | ❌ Not included | ✅ Gradle, Maven, SBT | ✅ Clean Wizard |
| **Docker Light** | ✅ `docker system prune -f` | ❌ Not implemented | 🔴 **MISSING** |
| **Nix Temp Files** | ✅ `/tmp/nix-build-*`, `/tmp/nix-shell-*` | ❌ Not implemented | 🔴 **MISSING** |

### `clean-quick` Gap: **2 missing features**

---

### 1.2 `clean` (Standard) Comparison

| Feature | SystemNix | Clean Wizard | Status |
|---------|-----------|--------------|--------|
| **Everything from quick** | ✅ | ⚠️ Partial | See above |
| **Nix GC** | ✅ `--delete-older-than 1d` | ✅ Count-based (N generations) | ⚠️ Different strategy |
| **Nix Store Optimization** | ✅ `nix-store --optimize` | ❌ Not implemented | 🔴 **MISSING** |
| **Nix Profile Wipe** | ✅ `nix profile wipe-history` | ❌ Not implemented | 🔴 **MISSING** |
| **Docker Full Prune** | ✅ `docker system prune -af` | ✅ `docker system prune -af --volumes` | ✅ Better |
| **Xcode Simulators** | ✅ `xcrun simctl delete unavailable` | ❌ Not implemented | 🔴 **MISSING** |
| **Size Before/After** | ✅ `du -sh` output | ❌ Not implemented | 🔴 **MISSING** |
| **Disk Space Display** | ✅ `df -h` | ❌ Not implemented | 🔴 **MISSING** |

### `clean` Gap: **5 missing features**

---

### 1.3 `clean-aggressive` Comparison

| Feature | SystemNix | Clean Wizard | Status |
|---------|-----------|--------------|--------|
| **Everything from clean** | ✅ | ⚠️ Partial | See above |
| **Nix ALL Generations** | ✅ No threshold | ⚠️ Count-based only | ⚠️ Partial |
| **Nix ALL Profiles** | ✅ Full wipe | ❌ Not implemented | 🔴 **MISSING** |
| **NVM Cleanup** | ✅ Deletes `~/.nvm/versions/node/*` | ⚠️ Scans only (NO-OP) | 🔴 **BROKEN** |
| **Pyenv Cleanup** | ✅ Deletes `~/.pyenv/versions/*` | ⚠️ Scans only (NO-OP) | 🔴 **BROKEN** |
| **Rbenv Cleanup** | ✅ Deletes `~/.rbenv/versions/*` | ⚠️ Scans only (NO-OP) | 🔴 **BROKEN** |
| **Full Cache Wipe** | ✅ `rm -rf ~/.cache` | ⚠️ Via BuildCache only | ⚠️ Partial |
| **Xcode DerivedData** | ✅ Full wipe | ✅ Via SystemCache | ✅ Matching |
| **Docker with Volumes** | ✅ `--volumes` | ✅ `--volumes` | ✅ Matching |
| **iOS ALL Simulators** | ✅ `xcrun simctl delete all` | ❌ Not implemented | 🔴 **MISSING** |
| **Confirmation Prompt** | ✅ Explicit interactive pause | ⚠️ Generic Yes/No | ⚠️ Different |

### `clean-aggressive` Gap: **8 missing/broken features**

---

## 2. Priority Work Items

### 🔴 P1: CRITICAL - Language Version Manager NO-OP

**Issue:** `langversionmanager.go` scans for versions but NEVER deletes them.

**SystemNix behavior:**
```bash
# This actually frees space
rm -rf ~/.nvm/versions/node/*
rm -rf ~/.pyenv/versions/*
rm -rf ~/.rbenv/versions/*
```

**Clean Wizard current behavior:**
```go
// This is a NO-OP by default to avoid destructive behavior
// Returns (FreedBytes: 0, Warning: "...")
```

**Required Work:**
- [ ] Implement actual deletion logic for NVM
- [ ] Implement actual deletion logic for Pyenv
- [ ] Implement actual deletion logic for Rbenv
- [ ] Add aggressive mode flag to enable destructive cleanup
- [ ] Add safety confirmation for language manager cleanup
- [ ] Update presets to include language managers in aggressive mode

**Impact:** Blocks `clean-aggressive` parity

---

### 🔴 P2: CRITICAL - Docker Light Prune

**Issue:** `clean-quick` needs `docker system prune -f` but Clean Wizard only has full prune.

**SystemNix behavior:**
```bash
# Light prune - removes stopped containers, networks, images older than undefined
docker system prune -f
```

**Clean Wizard current state:**
- Only has `DockerPruneAll` mode
- No light prune option

**Required Work:**
- [ ] Add `DockerPruneLight` enum value
- [ ] Implement light prune command: `docker system prune -f`
- [ ] Update quick mode preset to include Docker with light prune
- [ ] Add Docker availability check before prune

**Impact:** Blocks `clean-quick` parity

---

### 🟠 P3: HIGH - Nix Temp Files

**Issue:** `/tmp/nix-build-*` and `/tmp/nix-shell-*` not cleaned.

**SystemNix behavior:**
```bash
rm -rf /tmp/nix-build-* /tmp/nix-shell-*
```

**Clean Wizard current state:**
- TempFiles cleaner exists but doesn't include Nix-specific patterns
- Standard temp paths: `/tmp` with exclusions

**Required Work:**
- [ ] Add Nix-specific temp path pattern to TempFiles cleaner
- [ ] Auto-detect if running on system with Nix installed
- [ ] Add `--include-nix-temp` flag to enable Nix temp cleanup
- [ ] Include in quick mode by default (SystemNix does this)

**Impact:** Blocks `clean-quick` parity

---

### 🟠 P4: HIGH - Nix Store Optimization

**Issue:** `nix-store --optimize` not implemented.

**SystemNix behavior:**
```bash
nix-store --optimize
# or with sudo
sudo -S nix-store --optimize
```

**Clean Wizard current state:**
- NixCleaner only does garbage collection
- No optimization implementation

**Required Work:**
- [ ] Add `--optimize` flag to NixCleaner
- [ ] Implement `nix-store --optimize` command
- [ ] Handle sudo prompts gracefully
- [ ] Add size comparison before/after optimization

**Impact:** Blocks `clean` parity

---

### 🟠 P5: HIGH - Nix Profile Management

**Issue:** `nix profile wipe-history` not implemented.

**SystemNix behavior:**
```bash
nix profile wipe-history --profile /Users/$(whoami)/.local/state/nix/profiles/profile
```

**Clean Wizard current state:**
- No profile management implemented
- Only generational cleanup

**Required Work:**
- [ ] Add NixProfileCleaner or extend NixCleaner
- [ ] Implement `nix profile list` to find profiles
- [ ] Implement `nix profile wipe-history` for each profile
- [ ] Add `--include-profiles` flag

**Impact:** Blocks `clean` parity

---

### 🟠 P6: HIGH - Time-Based vs Count-Based Nix GC

**Issue:** SystemNix uses `--delete-older-than 1d`, Clean Wizard uses count-based keep.

**SystemNix behavior:**
```bash
nix-collect-garbage -d --delete-older-than 1d
```

**Clean Wizard behavior:**
```go
// Keeps N generations, deletes older
NewNixCleaner(verbose, dryRun, keepCount)
```

**Required Work:**
- [ ] Add `--older-than` duration flag to NixCleaner
- [ ] Implement time-based deletion logic
- [ ] Allow both count-based AND time-based (use whichever is more conservative)
- [ ] Update documentation to explain the difference

**Impact:** Blocks `clean` parity (different behavior, not necessarily worse)

---

### 🟡 P7: MEDIUM - Xcode iOS Simulators

**Issue:** `xcrun simctl delete unavailable` not implemented.

**SystemNix behavior:**
```bash
xcrun simctl delete unavailable 2>/dev/null || echo "⚠️  Xcode not found"
```

**Clean Wizard current state:**
- SystemCacheCleaner has Xcode DerivedData
- No iOS simulator cleanup

**Required Work:**
- [ ] Add iOS simulator cleanup to SystemCacheCleaner
- [ ] Implement `xcrun simctl delete unavailable`
- [ ] Implement `xcrun simctl delete all` for aggressive mode
- [ ] Check for Xcode command line tools availability

**Impact:** Blocks `clean` and `clean-aggressive` parity

---

### 🟡 P8: MEDIUM - Size Reporting Before/After

**Issue:** SystemNix shows `du -sh /nix/store` before/after. Clean Wizard has estimates.

**SystemNix behavior:**
```bash
@echo "📊 Current store size:"
@du -sh /nix/store
# ... cleanup ...
@echo "📊 New store size:"
@du -sh /nix/store
```

**Clean Wizard current state:**
- Hardcoded estimates (50MB per generation)
- No real filesystem queries

**Required Work:**
- [ ] Implement real filesystem size queries for Nix store
- [ ] Implement real filesystem size queries for Homebrew
- [ ] Implement real filesystem size queries for Docker
- [ ] Add `--show-sizes` flag to display before/after
- [ ] Update TUI to show actual freed bytes

**Impact:** Quality of life improvement

---

### 🟡 P9: MEDIUM - Disk Space Display

**Issue:** `df -h /` not implemented.

**SystemNix behavior:**
```bash
@df -h / | tail -1 | awk '{print "  Available: " $4 " of " $2 " (" $5 " used)"}'
```

**Clean Wizard current state:**
- No disk space display

**Required Work:**
- [ ] Add disk space summary to end of cleanup
- [ ] Show: Total, Used, Available, Percentage
- [ ] Highlight freed space
- [ ] Add `--show-disk-space` flag

**Impact:** Quality of life improvement

---

### 🟡 P10: MEDIUM - iOS Simulator Full Delete

**Issue:** Aggressive mode needs `xcrun simctl delete all`, not just unavailable.

**SystemNix behavior:**
```bash
@echo "📱 Removing all iOS simulators..."
xcrun simctl delete all 2>/dev/null || true
```

**Clean Wizard current state:**
- No iOS simulator cleanup at all

**Required Work:**
- [ ] Implement `xcrun simctl delete all` for aggressive mode
- [ ] Add confirmation warning (deletes ALL simulators)
- [ ] Require explicit `--aggressive` flag

**Impact:** Blocks `clean-aggressive` parity

---

## 3. Implementation Roadmap

### Phase 1: Quick Mode Parity (Week 1)

| Task | Effort | Priority |
|------|--------|----------|
| Implement Docker light prune (P2) | 2 hours | 🔴 CRITICAL |
| Implement Nix temp files cleanup (P3) | 1 hour | 🔴 CRITICAL |
| Update README to match current state | 2 hours | 🟡 MEDIUM |
| Update quick mode preset to include Docker + Nix temp | 30 min | 🟡 MEDIUM |

**Deliverable:** `clean-wizard clean --mode quick` matches SystemNix `clean-quick`

---

### Phase 2: Standard Mode Parity (Week 2)

| Task | Effort | Priority |
|------|--------|----------|
| Implement Nix store optimization (P4) | 3 hours | 🔴 CRITICAL |
| Implement Nix profile management (P5) | 3 hours | 🔴 CRITICAL |
| Implement time-based GC (P6) | 2 hours | 🟠 HIGH |
| Implement iOS simulator cleanup (P7) | 2 hours | 🟡 MEDIUM |
| Implement size reporting (P8) | 2 hours | 🟡 MEDIUM |
| Implement disk space display (P9) | 1 hour | 🟡 MEDIUM |
| Update standard mode preset | 30 min | 🟡 MEDIUM |

**Deliverable:** `clean-wizard clean --mode standard` matches SystemNix `clean`

---

### Phase 3: Aggressive Mode Parity (Week 3)

| Task | Effort | Priority |
|------|--------|----------|
| Fix Language Version Manager NO-OP (P1) | 5 hours | 🔴 CRITICAL |
| Implement iOS simulator full delete (P10) | 1 hour | 🟡 MEDIUM |
| Update aggressive mode preset to include language managers | 30 min | 🟡 MEDIUM |
| Add aggressive confirmation dialog | 1 hour | 🟡 MEDIUM |

**Deliverable:** `clean-wizard clean --mode aggressive` matches SystemNix `clean-aggressive`

---

### Phase 4: Polish (Week 4)

| Task | Effort | Priority |
|------|--------|----------|
| Update README to reflect full capabilities | 2 hours | 🟡 MEDIUM |
| Add comprehensive documentation | 4 hours | 🟡 MEDIUM |
| Write BDD tests for all three modes | 4 hours | 🟡 MEDIUM |
| Create man page / shell completion | 2 hours | 🟢 LOW |

**Deliverable:** Production-ready documentation and tests

---

## 4. Estimated Timeline

| Phase | Duration | Total Effort |
|-------|----------|---------------|
| Phase 1: Quick Mode | 1 week | ~6 hours |
| Phase 2: Standard Mode | 1 week | ~14 hours |
| Phase 3: Aggressive Mode | 1 week | ~8 hours |
| Phase 4: Polish | 1 week | ~12 hours |
| **Total** | **4 weeks** | **~40 hours** |

---

## 5. Updated Preset Definitions

### 5.1 Quick Mode (Target)

```go
case "quick":
    return []CleanerType{
        CleanerTypeHomebrew,           // brew autoremove + cleanup
        CleanerTypeNodePackages,        // npm, pnpm, yarn, bun caches
        CleanerTypeGoPackages,          // go clean -cache -testcache -modcache
        CleanerTypeTempFiles,           // Including /tmp/nix-build-* /tmp/nix-shell-*
        CleanerTypeBuildCache,          // Gradle, Maven, SBT caches
        CleanerTypeDocker,              // docker system prune -f (light)
    }
```

**vs Current:** Adds Docker (light) + Nix temp files

---

### 5.2 Standard Mode (Target)

```go
case "standard":
    // All available cleaners from quick
    // Plus:
    return []CleanerType{
        CleanerTypeNix,                 // gc --delete-older-than 1d + --optimize
        CleanerTypeNixProfiles,         // nix profile wipe-history (NEW)
        CleanerTypeDocker,              // docker system prune -af --volumes
        CleanerTypeSystemCache,         // Spotlight, Xcode DerivedData, iOS (NEW)
        // ... all other cleaners ...
    }
```

**vs Current:** Adds NixProfiles + SystemCache includes iOS

---

### 5.3 Aggressive Mode (Target)

```go
case "aggressive":
    // All cleaners from standard
    // Plus:
    return []CleanerType{
        CleanerTypeLanguageManagers,    // NVM, Pyenv, Rbenv (ACTUALLY DELETES)
        CleanerTypeIOSSimulators,      // xcrun simctl delete all (NEW)
        // ... all others ...
    }
```

**vs Current:** Language managers will NOW DELETE (currently NO-OP) + iOS full delete

---

## 6. Documentation Updates Required

### README.md (Current vs Target)

**Current:** Only mentions Nix generations
**Target:** Comprehensive multi-cleaner tool

**Sections to add:**
- [ ] Multi-cleaner overview (11 cleaners)
- [ ] Preset modes (quick/standard/aggressive)
- [ ] Full feature matrix
- [ ] Installation for macOS
- [ ] Examples for each preset mode
- [ ] Safety features documentation

---

## 7. Testing Requirements

### BDD Scenarios Needed

```gherkin
Feature: Clean Quick Mode
  As a user who runs daily cleanup
  I want to clean common caches without system changes
  So that I can free space quickly

  Scenario: Running quick mode cleans Homebrew
    Given Homebrew is installed
    When I run "clean-wizard clean --mode quick"
    Then Homebrew cache should be cleaned
    And Nix store should NOT be changed
    And Docker should use light prune

  Scenario: Running quick mode includes Nix temp files
    Given Nix is installed
    And /tmp/nix-build-* exists
    When I run "clean-wizard clean --mode quick"
    Then Nix temp files should be deleted
```

```gherkin
Feature: Aggressive Mode Language Manager Cleanup
  As a user with many Node versions
  I want to remove old Node versions
  So that I can free gigabytes of space

  Scenario: NVM versions are actually deleted
    Given NVM is installed
    And ~/.nvm/versions/node/* exists
    When I run "clean-wizard clean --mode aggressive"
    Then old Node versions should be deleted
    And FreedBytes should be > 0
```

---

## 8. Success Criteria

### Quick Mode Parity ✅
- [ ] Homebrew cleanup matches SystemNix
- [ ] Node packages cleanup matches SystemNix
- [ ] Go cache cleanup matches SystemNix
- [ ] Temp files includes `/tmp/nix-build-*`
- [ ] Docker uses light prune
- [ ] No Nix store changes in quick mode

### Standard Mode Parity ✅
- [ ] Nix garbage collection (time-based or count-based)
- [ ] Nix store optimization works
- [ ] Nix profile cleanup works
- [ ] Docker full prune works
- [ ] Xcode DerivedData cleaned
- [ ] iOS unavailable simulators cleaned
- [ ] Size before/after displayed

### Aggressive Mode Parity ✅
- [ ] Language version managers ACTUALLY DELETE
- [ ] All Nix generations deleted (no keep count)
- [ ] All Nix profiles wiped
- [ ] iOS ALL simulators deleted
- [ ] Explicit confirmation for destructive actions
- [ ] Disk space freed is accurate (not 0)

---

## 9. Code References

### Key Files to Modify

| File | Purpose |
|------|---------|
| `internal/cleaner/langversionmanager.go` | Fix NO-OP, implement actual deletion |
| `internal/cleaner/docker.go` | Add light prune mode |
| `internal/cleaner/tempfiles.go` | Add Nix temp patterns |
| `internal/cleaner/nix.go` | Add optimization, profile cleanup, time-based GC |
| `internal/cleaner/systemcache.go` | Add iOS simulator cleanup |
| `cmd/clean-wizard/commands/clean.go` | Update preset definitions |
| `README.md` | Complete rewrite |

### New Files to Create

| File | Purpose |
|------|---------|
| `docs/features/quick-mode.md` | Quick mode documentation |
| `docs/features/standard-mode.md` | Standard mode documentation |
| `docs/features/aggressive-mode.md` | Aggressive mode documentation |
| `docs/testing/bdd-scenarios.md` | BDD test scenarios |

---

## 10. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Language manager deletion too destructive | Medium | High | Add explicit `--force-aggressive` flag |
| Nix optimization takes too long | Low | Medium | Add timeout, show progress |
| iOS simulator deletion causes issues | Medium | Low | Require confirmation |
| User confusion between modes | Medium | Low | Clear TUI messaging |

---

## Conclusion

Clean Wizard has the **foundation** to become your ultimate MacBook cleanup tool, but needs:

1. **Critical fix:** Language Version Manager must actually delete
2. **Missing features:** Docker light prune, Nix temp files, Nix optimization, Nix profiles, iOS simulators
3. **Documentation:** Complete rewrite of README

**Timeline:** 4 weeks, ~40 hours of implementation work

**Recommendation:** Implement Phase 1 (Quick Mode parity) first, validate, then proceed with remaining phases.

---

*Document created: 2026-02-09*  
*For questions or updates, open an issue.*