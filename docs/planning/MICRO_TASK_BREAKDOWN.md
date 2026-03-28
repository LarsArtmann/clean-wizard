# Micro-Task Breakdown - Clean Wizard Consolidation

**Generated:** 2026-03-28 10:46 CET  
**Total Tasks:** 38 micro-tasks  
**Target Duration:** 12 minutes per task

---

## Phase 1: Quick Wins (18 micro-tasks, ~3 hours)

### Task Group 1.1: Delete Ghost Directories (4 micro-tasks)

| # | Task | Time | Action | Verification |
|---|------|------|--------|--------------|
| 1.1.1 | Verify internal/application is empty | 2 min | `find internal/application -name "*.go"` | Count = 0 |
| 1.1.2 | Verify internal/infrastructure is empty | 2 min | `find internal/infrastructure -name "*.go"` | Count = 0 |
| 1.1.3 | Delete internal/application directory | 1 min | `rm -rf internal/application/` | Directory gone |
| 1.1.4 | Delete internal/infrastructure directory | 1 min | `rm -rf internal/infrastructure/` | Directory gone |

**Dependencies:** None  
**Total Time:** 6 minutes

### Task Group 1.2: Remove Old GolangciLintCleaner (4 micro-tasks)

| # | Task | Time | Action | Verification |
|---|------|------|--------|--------------|
| 1.2.1 | Identify all usages of GolangciLintCleaner | 5 min | `grep -r "GolangciLintCleaner\|golang_lint_adapter"` | List files |
| 1.2.2 | Update GoCacheCleaner to remove old usage | 10 min | Edit golang_cache_cleaner.go | Build passes |
| 1.2.3 | Update CLI to remove old references | 10 min | Edit cleaner_implementations.go, cleaner_config.go | Build passes |
| 1.2.4 | Delete golang_lint_adapter.go | 5 min | `rm internal/cleaner/golang_lint_adapter.go` | File gone |

**Dependencies:** 1.2.1 before 1.2.2 before 1.2.3 before 1.2.4  
**Total Time:** 30 minutes

### Task Group 1.3: Fix Unused Parameters (7 micro-tasks)

| # | Task | Time | File | Line | Fix |
|---|------|------|------|------|-----|
| 1.3.1 | Fix buildcache.go unused ctx | 5 min | buildcache.go | 201 | Prefix with `_` or use |
| 1.3.2 | Fix githistory.go unused totalBytes | 5 min | githistory.go | 286 | Prefix with `_` or use |
| 1.3.3 | Fix golang_cache_cleaner.go unused ctx | 5 min | golang_cache_cleaner.go | 294 | Prefix with `_` or use |
| 1.3.4 | Fix nodepackages.go unused ctx | 5 min | nodepackages.go | 229 | Prefix with `_` or use |
| 1.3.5 | Fix systemcache.go unused ctx x2 | 10 min | systemcache.go | 368,385 | Prefix with `_` or use |
| 1.3.6 | Fix enhanced_loader.go unused ctx | 5 min | enhanced_loader.go | 14 | Prefix with `_` or use |
| 1.3.7 | Fix enhanced_loader_private unused params | 10 min | enhanced_loader_private.go | 57,91 | Prefix with `_` or use |

**Dependencies:** None (can be done in parallel)  
**Total Time:** 45 minutes

### Task Group 1.4: Remove Deprecated Aliases (3 micro-tasks)

| # | Task | Time | Action | Verification |
|---|------|------|--------|--------------|
| 1.4.1 | List all deprecated aliases | 5 min | `grep -E "RiskLow\|RiskMedium" domain/types.go` | List aliases |
| 1.4.2 | Update domain/types.go to remove aliases | 10 min | Edit types.go | Build passes |
| 1.4.3 | Update all references to use type-safe enums | 15 min | `grep -r "domain.RiskLow\|domain.RiskMedium"` | Update refs |

**Dependencies:** 1.4.1 before 1.4.2 before 1.4.3  
**Total Time:** 30 minutes

---

## Phase 2: Medium Effort (20 micro-tasks, ~4 hours)

### Task Group 2.1: Add samber/lo (6 micro-tasks)

| # | Task | Time | Action | Verification |
|---|------|------|--------|--------------|
| 2.1.1 | Research samber/lo API | 10 min | Read docs | Understand Map, Filter, Reduce |
| 2.1.2 | Add samber/lo to go.mod | 2 min | `go get github.com/samber/lo` | In go.mod |
| 2.1.3 | Replace manual Filter with lo.Filter | 10 min | Edit registry.go | Build passes |
| 2.1.4 | Replace manual Map with lo.Map | 10 min | Edit registry.go, cleaner files | Build passes |
| 2.1.5 | Replace manual Reduce with lo.Reduce | 10 min | Edit result handling | Build passes |
| 2.1.6 | Review and test all changes | 10 min | `go test ./...` | Tests pass |

**Dependencies:** None  
**Total Time:** 52 minutes

### Task Group 2.2: Add samber/mo (3 micro-tasks)

| # | Task | Time | Action | Verification |
|---|------|------|--------|--------------|
| 2.2.1 | Research samber/mo API | 10 min | Read docs | Understand Option, Either |
| 2.2.2 | Evaluate mo vs result.Result overlap | 10 min | Compare types | Decision doc |
| 2.2.3 | Add samber/mo to go.mod or skip | 2 min | `go get github.com/samber/mo` | In go.mod or skip |

**Dependencies:** None  
**Total Time:** 22 minutes (or skip if not needed)

### Task Group 2.3: Reduce Function Complexity (8 micro-tasks)

| # | Task | Time | Function | Strategy |
|---|------|------|----------|----------|
| 2.3.1 | Profile runCleanCommand complexity | 5 min | clean.go | Measure current |
| 2.3.2 | Extract profile handling to function | 15 min | clean.go | Extract handleProfile() |
| 2.3.3 | Extract mode handling to function | 15 min | clean.go | Extract handleMode() |
| 2.3.4 | Extract interactive selection to function | 15 min | clean.go | Extract handleSelection() |
| 2.3.5 | Profile ValidateSettings complexity | 5 min | operation_validation.go | Measure current |
| 2.3.6 | Group validation by operation type | 15 min | operation_validation.go | Map-based validation |
| 2.3.7 | Profile validateEnumDefaults | 5 min | operation_defaults.go | Measure current |
| 2.3.8 | Convert to map-based validation | 15 min | operation_defaults.go | Map-based validation |

**Dependencies:** 2.3.1→2.3.4, 2.3.5→2.3.6, 2.3.7→2.3.8  
**Total Time:** 90 minutes

---

## Phase 3: Long-term (Optional, 14+ hours)

### Task Group 3.1: Integration Tests

| # | Task | Time | Coverage Target |
|---|------|------|----------------|
| 3.1.1 | Test full cleaning workflow | 60 min | End-to-end |
| 3.1.2 | Test configuration loading | 60 min | Config parsing |
| 3.1.3 | Test error handling paths | 60 min | Error scenarios |

### Task Group 3.2: Fuzz Testing

| # | Task | Time | Target |
|---|------|------|--------|
| 3.2.1 | Add fuzz test for parseSize | 40 min | Size parsing |
| 3.2.2 | Add fuzz test for duration parsing | 40 min | Duration parsing |
| 3.2.3 | Add fuzz test for YAML loading | 40 min | Config parsing |

### Task Group 3.3: CLI Flag Integration

| # | Task | Time | Target |
|---|------|------|--------|
| 3.3.1 | Complete --dry-run propagation | 40 min | All cleaners |
| 3.3.2 | Complete --verbose propagation | 40 min | All cleaners |
| 3.3.3 | Add config path override | 40 min | Config loading |

---

## Summary Table

| Phase | Task Groups | Micro-Tasks | Total Time |
|-------|-------------|-------------|------------|
| Phase 1 | 4 | 18 | ~3 hours |
| Phase 2 | 3 | 17 | ~4 hours |
| Phase 3 | 3 | 9 | ~14 hours |
| **Total** | **10** | **44** | **~21 hours** |

---

## Execution Order

```
1. Phase 1.1 (Ghost Dirs) - 6 min
   ↓
2. Phase 1.2 (Old Cleaner) - 30 min
   ↓
3. Phase 1.3 (Unused Params) - 45 min
   ↓
4. Phase 1.4 (Deprecated Aliases) - 30 min
   ↓
5. Phase 2.1 (samber/lo) - 52 min
   ↓
6. Phase 2.2 (samber/mo) - 22 min (or skip)
   ↓
7. Phase 2.3 (Complexity) - 90 min
   ↓
8. Phase 3 (Optional) - 14+ hours
```

---

## Git Commit Strategy

Each micro-task should result in a commit:

```
git add <files>
git commit -m "<task>: <description>

<what changed>
<why it changed>

Time: <X> min
Refs: #ghost-cleanup
"
```

Example:
```
git commit -m "chore: delete ghost internal/application directory

Remove empty application/ directory that was never used.
Contains no Go files, only empty subdirectories.

Time: 1 min
Refs: #ghost-cleanup
"
```

---

_Document generated by AI Assistant - Micro-Task Breakdown_
