# Branching-Flow Mixins Analysis Report

**Date:** 2026-04-01  
**Tool:** branching-flow `mixins`  
**Command:** `branching-flow mixins . --type-strategy semantic --show-all-locations --format markdown`

## Executive Summary

| Metric                   | Value                              |
| ------------------------ | ---------------------------------- |
| Files Analyzed           | 133                                |
| Structs Found            | 185                                |
| Composition Health Score | **99/100** (good)                  |
| Mixin Opportunities      | 20 (2 medium, 18 low)              |
| Anti-Patterns            | 1 "Base naming", 2 "Large structs" |

## Key Themes

### 1. Validation Type Duplication (Highest Impact)

ValidationError, ValidationWarning, ValidationResult duplicated across 3 packages:

- `internal/config/` — config-specific validation types
- `internal/domain/` — domain-level validation types
- `internal/shared/context/` — context-level validation types

**Affected findings:** #3, #4, #13–#18, #20 (9 of 20)

This is the biggest structural issue. These types share 2–10 identical fields but live in separate packages with different Go type identities, forcing conversions between them.

### 2. Builder Pattern Duplication

SafeProfile/SafeProfileBuilder and SafeConfig/SafeConfigBuilder share 4–5 fields.

**Affected findings:** #2, #5

### 3. Test Config Duplication

BooleanSettingsCleanerTestConfig and BooleanSettingsTestConfig share 5 fields.

**Affected finding:** #1

### 4. Result Type Fragmentation

PublicCleanResult, JSONOutput, SanitizationResult, SanitizeConfigResult scattered across packages with overlapping fields.

**Affected findings:** #6, #7

### 5. Generic vs Non-Generic Rule Duplication

TypeSafeValidationRules vs ConfigValidationRules share 7 fields. NumericValidationRule vs ValidationRule[T] share 2 fields.

**Affected findings:** #8, #9

## Prioritized Action Plan

### Phase 1: Validation Type Consolidation (3 items, ~30 min)

| Priority | Finding | Action                                                                       | Est. Time |
| -------- | ------- | ---------------------------------------------------------------------------- | --------- |
| P1       | #16     | Unify domain.ValidationContext + context.ValidationConfig (10 shared fields) | 12 min    |
| P1       | #17     | Unify domain.ValidationError + context.ValidationError (3 shared fields)     | 8 min     |
| P1       | #20     | Unify context.ValidationError + context.ValidationWarning (3 shared fields)  | 8 min     |

**Rationale:** These are the most impactful — 10 shared fields in #16 is the highest count in the entire report. Validation types are used everywhere, so consolidating them reduces cognitive load and eliminates conversion boilerplate.

### Phase 2: Cross-Package Result Types (2 items, ~20 min)

| Priority | Finding       | Action                                                                                   | Est. Time |
| -------- | ------------- | ---------------------------------------------------------------------------------------- | --------- |
| P2       | #3, #4, #10   | Consolidate config.ValidationResult, config.ConfigChangeResult, context.ValidationResult | 12 min    |
| P2       | #13, #14, #15 | Consolidate config.ValidationWarning with domain/context equivalents                     | 8 min     |

### Phase 3: Builder & Config Dedup (2 items, ~15 min)

| Priority | Finding | Action                                                      | Est. Time |
| -------- | ------- | ----------------------------------------------------------- | --------- |
| P3       | #2      | Extract SafeProfileMixin for SafeProfile/SafeProfileBuilder | 6 min     |
| P3       | #5      | Extract SafeMixin for SafeConfig/SafeConfigBuilder          | 6 min     |

### Phase 4: Remaining Low-Priority Items (~25 min)

| Priority | Finding | Action                                                                         |
| -------- | ------- | ------------------------------------------------------------------------------ |
| P4       | #1      | Extract BooleanSettingsMixin for test configs                                  |
| P4       | #6      | Align PublicCleanResult + JSONOutput                                           |
| P4       | #8      | Unify TypeSafeValidationRules + ConfigValidationRules                          |
| P4       | #9      | Consider making NumericValidationRule generic or embedding ValidationRule[int] |
| P4       | #11     | Extract ScanDisplayMixin for git history display/result                        |
| P4       | #12     | Extract shared scanner/operator mixin in compiled binaries                     |

### Anti-Patterns

1. **Base naming:** `CleanerBase` — consider renaming to reflect purpose (e.g., `CleanerCore`, `CleanerFoundation`)
2. **Large structs:** 2 found — investigate and split by responsibility

## Decision Required

The validation type consolidation (Phase 1–2) is the most impactful change but requires an architectural decision:

**Option A:** Pick one canonical package (e.g., `domain/`) and have all others import from it.  
**Option B:** Create a dedicated `internal/validation/types/` package with shared types.  
**Option C:** Extract shared fields into mixins embedded by each package-specific type.

Each approach has tradeoffs around coupling, import cycles, and semantic clarity.

## Raw Output

Full branching-flow output is available by re-running:

```bash
branching-flow mixins . --type-strategy semantic --show-all-locations --format markdown
```
