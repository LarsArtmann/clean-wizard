# 📊 DETAILED TASK EXECUTION PLAN - Clean Wizard

**Date:** 2025-11-09  
**Session:** COMPREHENSIVE TASK BREAKDOWN  
**Scope:** ALL GitHub Issues + Internal TODOs

---

## 🎯 PHASE 1: CRITICAL INFRASTRUCTURE (30min tasks max - 30 total)

### PRIORITY LEVEL: CRITICAL (1% → 51% Impact)

| Task ID | Description | GitHub Issue | Time (min) | Impact | Dependencies |
|---------|-------------|--------------|------------|--------|--------------|
| G1-1 | **Remove Ghost System: Delete `internal/types/` package** | internal | 30 | Critical | None |
| G1-2 | **Delete `main.go.broken` dead code** | internal | 15 | Critical | None |
| G1-3 | **Migrate consumers from legacy types to domain types** | internal | 30 | Critical | G1-1 |
| G1-4 | **Consolidate duplicate FormatSize functions** | internal | 20 | Critical | G1-1 |
| G1-5 | **Create centralized type conversion package** | #11 | 30 | Critical | None |
| G1-6 | **Implement primitive→domain conversion functions** | #11 | 30 | Critical | G1-5 |
| G1-7 | **Refactor all primitive→domain conversions** | #11 | 30 | Critical | G1-6 |
| G1-8 | **Enable godog debug mode for BDD debugging** | #12 | 15 | Critical | None |
| G1-9 | **Systematic regex pattern analysis for 8 undefined steps** | #12 | 30 | Critical | G1-8 |
| G1-10 | **Fix remaining BDD step definition patterns** | #12 | 25 | Critical | G1-9 |
| G1-11 | **Create centralized error package structure** | #4 | 25 | Critical | None |
| G1-12 | **Define custom error types with codes** | #4 | 30 | Critical | G1-11 |
| G1-13 | **Implement consistent error handling pattern** | #4 | 30 | Critical | G1-12 |
| G1-14 | **Update all error handling sites to use centralized pattern** | #4 | 30 | Critical | G1-13 |

### PRIORITY LEVEL: HIGH (4% → 64% Impact)

| Task ID | Description | GitHub Issue | Time (min) | Impact | Dependencies |
|---------|-------------|--------------|------------|--------|--------------|
| H1-1 | **Define configuration validation schema** | #5 | 30 | High | None |
| H1-2 | **Implement ConfigValidator interface** | #5 | 30 | High | H1-1 |
| H1-3 | **Add field-level validation helpers** | #5 | 25 | High | H1-2 |
| H1-4 | **Update configuration loading to include validation** | #5 | 25 | High | H1-3 |
| H1-5 | **Add configuration validation tests** | #5 | 20 | High | H1-4 |
| H1-6 | **Complete Homebrew adapter implementation** | #2 | 30 | High | None |
| H1-7 | **Add Homebrew cleaning operations** | #2 | 30 | High | H1-6 |
| H1-8 | **Implement package cache cleaners (npm, cargo, go)** | #2 | 30 | High | H1-7 |
| H1-9 | **Add BDD scenarios for other operations** | #3 | 25 | High | G1-10 |
| H1-10 | **Implement performance test framework** | #3 | 30 | High | H1-9 |

### PRIORITY LEVEL: MEDIUM (20% → 80% Impact)

| Task ID | Description | GitHub Issue | Time (min) | Impact | Dependencies |
|---------|-------------|--------------|------------|--------|--------------|
| M1-1 | **Research and select progress bar library** | #7 | 20 | Medium | None |
| M1-2 | **Implement progress bar interface** | #7 | 25 | Medium | M1-1 |
| M1-3 | **Add progress to scan operations** | #7 | 25 | Medium | M1-2 |
| M1-4 | **Add progress to clean operations** | #7 | 25 | Medium | M1-3 |
| M1-5 | **Implement colored output system** | #7 | 20 | Medium | M1-4 |
| M1-6 | **Add table formatting for results** | #7 | 25 | Medium | M1-5 |

---

## 🎯 PHASE 2: DETAILED MICRO-TASKS (15min tasks max - 150 total)

### CRITICAL INFRASTRUCTURE - Micro Tasks

| Task ID | Parent | Description | Time (min) | Sequence |
|---------|--------|-------------|------------|----------|
| G2-1 | G1-1 | Remove `internal/types/deprecated.go` file | 15 | 1 |
| G2-2 | G1-1 | Remove `internal/types/legacy.go` file | 15 | 2 |
| G2-3 | G1-1 | Remove empty `internal/types/` directory | 10 | 3 |
| G2-4 | G1-1 | Search for imports of removed types package | 15 | 4 |
| G2-5 | G1-1 | Update import statements to use domain package | 15 | 5 |
| G2-6 | G1-1 | Fix type references in command files | 15 | 6 |
| G2-7 | G1-1 | Fix type references in test files | 15 | 7 |
| G2-8 | G1-1 | Verify compilation after type migration | 10 | 8 |
| G2-9 | G1-2 | Delete `cmd/clean-wizard/main.go.broken` | 5 | 1 |
| G2-10 | G1-2 | Verify no references to broken file exist | 10 | 2 |
| G2-11 | G1-3 | Update main.go to use domain.ScanResult | 15 | 1 |
| G2-12 | G1-3 | Update command files to use domain types | 15 | 2 |
| G2-13 | G1-3 | Update test files to use domain types | 15 | 3 |
| G2-14 | G1-4 | Identify duplicate FormatSize usage | 10 | 1 |
| G2-15 | G1-4 | Remove FormatSize from legacy.go | 10 | 2 |
| G2-16 | G1-4 | Ensure all FormatSize calls use format package | 15 | 3 |
| G2-17 | G1-5 | Create `internal/conversions/` package directory | 10 | 1 |
| G2-18 | G1-5 | Create `conversions.go` file with interface | 15 | 2 |
| G2-19 | G1-5 | Define conversion function signatures | 15 | 3 |
| G2-20 | G1-6 | Implement ToCleanResult conversion function | 15 | 1 |
| G2-21 | G1-6 | Implement ToScanResult conversion function | 15 | 2 |
| G2-22 | G1-6 | Add unit tests for conversion functions | 15 | 3 |
| G2-23 | G1-7 | Refactor NixCleaner to use centralized conversions | 15 | 1 |
| G2-24 | G1-7 | Refactor Scanner to use centralized conversions | 15 | 2 |
| G2-25 | G1-7 | Refactor Command files to use centralized conversions | 15 | 3 |
| G2-26 | G1-8 | Add godog debug flag to BDD test command | 10 | 1 |
| G2-27 | G1-8 | Run BDD tests with debug output | 10 | 2 |
| G2-28 | G1-8 | Analyze debug output for pattern matching issues | 15 | 3 |
| G2-29 | G1-9 | Document each undefined step with exact pattern | 15 | 1 |
| G2-30 | G1-9 | Compare feature file patterns with step definitions | 15 | 2 |
| G2-31 | G1-9 | Identify regex escaping or character encoding issues | 15 | 3 |
| G2-32 | G1-10 | Fix regex patterns for quoted command steps | 10 | 1 |
| G2-33 | G1-10 | Fix regex patterns for parameterized And steps | 10 | 2 |
| G2-34 | G1-10 | Fix capture group patterns for step parameters | 10 | 3 |
| G2-35 | G1-10 | Test all 4 Nix scenarios end-to-end | 15 | 4 |
| G2-36 | G1-11 | Create `internal/pkg/errors/` package structure | 10 | 1 |
| G2-37 | G1-11 | Create `errors.go` with core error types | 15 | 2 |
| G2-38 | G1-11 | Create `codes.go` with error codes and categories | 15 | 3 |
| G2-39 | G1-11 | Create `handling.go` with error handling utilities | 15 | 4 |
| G2-40 | G1-11 | Create `logging.go` with error logging utilities | 15 | 5 |
| G2-41 | G1-12 | Define CleanWizardError struct with context | 15 | 1 |
| G2-42 | G1-12 | Define ErrorCode constants for categorization | 10 | 2 |
| G2-43 | G1-12 | Implement error context capture functionality | 15 | 3 |
| G2-44 | G1-13 | Implement HandleError function with structured logging | 15 | 1 |
| G2-45 | G1-13 | Implement error recovery mechanisms | 15 | 2 |
| G2-46 | G1-13 | Add user-friendly error message formatting | 15 | 3 |
| G2-47 | G1-14 | Update main.go error handling to use centralized pattern | 15 | 1 |
| G2-48 | G1-14 | Update command files error handling | 15 | 2 |
| G2-49 | G1-14 | Update adapter error handling | 15 | 3 |
| G2-50 | G1-14 | Update test error assertions | 10 | 4 |

### HIGH PRIORITY - Micro Tasks

| Task ID | Parent | Description | Time (min) | Sequence |
|---------|--------|-------------|------------|----------|
| H2-1 | H1-1 | Define ConfigValidationRules struct with tags | 15 | 1 |
| H2-2 | H1-1 | Add numeric constraints for validation | 15 | 2 |
| H2-3 | H1-1 | Add string constraints for validation | 15 | 3 |
| H2-4 | H1-1 | Add array/map constraints for validation | 15 | 4 |
| H2-5 | H1-1 | Add safety constraints for validation | 15 | 5 |
| H2-6 | H1-2 | Create ConfigValidator interface definition | 10 | 1 |
| H2-7 | H1-2 | Implement ValidateConfig method | 15 | 2 |
| H2-8 | H1-2 | Implement ValidateField method | 15 | 3 |
| H2-9 | H1-2 | Implement SanitizeConfig method | 15 | 4 |
| H2-10 | H1-3 | Create numeric validation helpers | 15 | 1 |
| H2-11 | H1-3 | Create string validation helpers | 15 | 2 |
| H2-12 | H1-3 | Create path validation helpers | 15 | 3 |
| H2-13 | H1-4 | Update config.Load to include validation | 15 | 1 |
| H2-14 | H1-4 | Add validation error formatting | 10 | 2 |
| H2-15 | H1-4 | Add validation warnings display | 10 | 3 |
| H2-16 | H1-5 | Create test cases for valid configurations | 10 | 1 |
| H2-17 | H1-5 | Create test cases for invalid configurations | 15 | 2 |
| H2-18 | H1-5 | Test validation error messages | 10 | 3 |
| H2-19 | H1-6 | Create HomebrewAdapter struct | 15 | 1 |
| H2-20 | H1-6 | Implement Homebrew IsAvailable method | 15 | 2 |
| H2-21 | H1-6 | Implement Homebrew ListGenerations method | 15 | 3 |
| H2-22 | H1-6 | Implement Homebrew CollectGarbage method | 15 | 4 |
| H2-23 | H1-7 | Create HomebrewCleaner implementation | 15 | 1 |
| H2-24 | H1-7 | Implement Homebrew cache cleaning | 15 | 2 |
| H2-25 | H1-7 | Implement Homebrew old formula cleanup | 15 | 3 |
| H2-26 | H1-8 | Create NpmCacheCleaner implementation | 15 | 1 |
| H2-27 | H1-8 | Create CargoCacheCleaner implementation | 15 | 2 |
| H2-28 | H1-8 | Create GoCacheCleaner implementation | 15 | 3 |
| H2-29 | H1-9 | Create Homebrew BDD feature file | 15 | 1 |
| H2-30 | H1-9 | Implement Homebrew step definitions | 15 | 2 |
| H2-31 | H1-9 | Create cache cleaning BDD scenarios | 15 | 3 |
| H2-32 | H1-10 | Create performance test utilities | 15 | 1 |
| H2-33 | H1-10 | Implement large file handling tests | 15 | 2 |
| H2-34 | H1-10 | Implement concurrent operation tests | 15 | 3 |
| H2-35 | H1-10 | Implement memory usage tests | 15 | 4 |

### MEDIUM PRIORITY - Micro Tasks

| Task ID | Parent | Description | Time (min) | Sequence |
|---------|--------|-------------|------------|----------|
| M2-1 | M1-1 | Research progressbar library options | 15 | 1 |
| M2-2 | M1-1 | Test schollz/progressbar/v3 compatibility | 15 | 2 |
| M2-3 | M1-1 | Evaluate bubbletea for TUI framework | 15 | 3 |
| M2-4 | M1-2 | Define ProgressBar interface | 15 | 1 |
| M2-5 | M1-2 | Create progress bar implementation | 15 | 2 |
| M2-6 | M1-2 | Add graceful fallback for non-interactive terminals | 10 | 3 |
| M2-7 | M1-3 | Add progress to Nix scanning operations | 15 | 1 |
| M2-8 | M1-3 | Add progress to generation listing | 15 | 2 |
| M2-9 | M1-3 | Add file count and size progress | 15 | 3 |
| M2-10 | M1-4 | Add progress to cleaning operations | 15 | 1 |
| M2-11 | M1-4 | Add time remaining estimates | 15 | 2 |
| M2-12 | M1-4 | Implement graceful cancellation with progress cleanup | 15 | 3 |
| M2-13 | M1-5 | Implement colored output based on operation status | 15 | 1 |
| M2-14 | M1-5 | Add status icons (✅ ⚠️ ❓ ⚡) | 10 | 2 |
| M2-15 | M1-5 | Implement verbose mode with detailed operation logging | 15 | 3 |
| M2-16 | M1-6 | Research table formatting libraries | 15 | 1 |
| M2-17 | M1-6 | Implement table formatting for results display | 15 | 2 |
| M2-18 | M1-6 | Add structured logging with different verbosity levels | 15 | 3 |

---

## 📈 EXECUTION TIMELINE

### SESSION 1: CRITICAL INFRASTRUCTURE (3-4 hours)
- **Focus**: Ghost system elimination + type conversion standardization
- **Tasks**: G1-1 through G1-7 + G2-1 through G2-25
- **Outcome**: Clean architecture with zero split-brain patterns

### SESSION 2: BDD COMPLETION + ERROR STANDARDIZATION (2-3 hours)  
- **Focus**: Complete BDD foundation + consistent error handling
- **Tasks**: G1-8 through G1-14 + G2-26 through G2-50
- **Outcome**: Robust testing + maintainable error patterns

### SESSION 3: PROFESSIONAL POLISH (2-3 hours)
- **Focus**: Configuration validation + real cleaning operations
- **Tasks**: H1-1 through H1-10 + H2-1 through H2-35
- **Outcome**: Production-ready safety + comprehensive functionality

### SESSION 4: USER EXPERIENCE EXCELLENCE (2 hours)
- **Focus**: CLI/UX improvements + visual polish
- **Tasks**: M1-1 through M1-6 + M2-1 through M2-18  
- **Outcome**: Professional user experience + tool adoption

---

## 🎯 SUCCESS METRICS

### AFTER SESSION 1:
- ✅ Zero ghost systems in repository
- ✅ Centralized type conversions implemented
- ✅ All primitive→domain boilerplate eliminated
- ✅ Compilation success with 100% type safety

### AFTER SESSION 2:
- ✅ BDD scenarios execute successfully (0 undefined steps)
- ✅ Error handling follows standardized pattern
- ✅ All error types categorized and context-rich
- ✅ Debugging and maintenance significantly improved

### AFTER SESSION 3:
- ✅ Configuration validation prevents runtime errors
- ✅ Real cleaning operations for Nix + Homebrew
- ✅ Comprehensive test coverage (>90%)
- ✅ Performance benchmarks established

### AFTER SESSION 4:
- ✅ Rich CLI experience with progress indicators
- ✅ Professional visual design and user feedback
- ✅ Documentation complete and current
- ✅ Tool ready for production deployment

---

## 🚀 IMMEDIATE NEXT STEPS

1. **Execute Session 1 Tasks** - Ghost system elimination
2. **Verify Compilation** - Ensure no regressions
3. **Run Test Suite** - Validate architectural changes
4. **Commit Changes** - Detailed commit messages for each phase
5. **Begin Session 2** - BDD completion + error standardization

**Total Estimated Time**: 10-12 hours across 4 focused sessions
**Critical Path**: Session 1 (Ghost Systems) → Session 2 (BDD/Errors) → Session 3 (Validation/Cleaning) → Session 4 (UX Polish)

---

*Generated with Crush*
*Comprehensive Task Planning*
*2025-11-09*