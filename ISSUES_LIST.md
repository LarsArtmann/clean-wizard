# CRITICAL CODE ISSUES (Sorted by Severity)
## Max 250 small actionable tasks

### 🔴 CRITICAL - Build/Test Failures (Fix Immediately)

1. [ ] Fix missing package: internal/adapters - referenced in internal/application/config/safe_test.go:8
2. [ ] Fix missing package: internal/bdd - referenced in internal/interface/cli/bdd/bdd_test.go:6
3. [ ] Fix missing package: internal/application/config/factories - referenced in test/config/bdd_nix_validation_test.go:8
4. [ ] Fix undefined: domain - internal/interface/api/mapper_test.go:16:19
5. [ ] Fix undefined: adapters - internal/infrastructure/system/adapters_test.go:13:8
6. [ ] Fix undefined: RiskLevelType - test/domain/benchmarks_test.go:12:15
7. [ ] Fix NewNixCleaner signature - internal/infrastructure/cleaners/nix_test.go:18:30 (too many arguments)
8. [ ] Fix mockCleaner interface - missing Cleanup method in internal/shared/utils/middleware/validation_test.go:69
9. [ ] Fix undefined: MapConfigToDomain - internal/interface/api/mapper_test.go:43:24

### 🟠 HIGH - Code Duplication (218 clone groups found)

#### Type Safe Enums (Major duplication source)
10. [ ] Refactor internal/domain/shared/type_safe_enums.go - 28+ clone groups
11. [ ] Extract common enum patterns from lines 127-129, 132-134, 176-178, etc.
12. [ ] Consolidate enum validation logic patterns
13. [ ] Remove duplicate enum string representation methods

#### Error Handling (High duplication)
14. [ ] Refactor internal/shared/utils/pkg/errors/errors.go - 15+ clone groups
15. [ ] Consolidate error creation patterns (lines 136-162)
16. [ ] Extract common error formatting functions
17. [ ] Remove duplicate error wrapping logic (lines 185-206, 210-217)
18. [ ] Fix internal/shared/errors/errors.go - duplicate patterns with lines 8-35
19. [ ] Consolidate internal/infrastructure/system/errors.go - 12+ clone groups

#### Test Code Duplication (Very high)
20. [ ] Refactor test/config/safety_level_test.go - 14+ clone groups (lines 22-158)
21. [ ] Extract common test patterns from test/config/safe_test.go
22. [ ] Consolidate internal/application/config/safe_test.go patterns - 9+ clone groups
23. [ ] Fix test/validation/duration_parser_test.go - 10+ clone groups
24. [ ] Refactor test/config/bdd_nix_validation_test.go - 8+ clone groups
25. [ ] Consolidate test/domain/cleanresult_test.go patterns
26. [ ] Fix test/validation/validation_validator_test.go duplications

#### Config Module Duplication
27. [ ] Refactor internal/domain/config/config.go - 15+ clone groups
28. [ ] Consolidate config getter/setter patterns (lines 96-143)
29. [ ] Fix internal/application/config/config.go duplicate patterns
30. [ ] Extract common config validation from type_safe_validation_rules_test.go

#### Cleaner Modules (Cross-file duplication)
31. [ ] Consolidate internal/infrastructure/cleaners/homebrew.go patterns
32. [ ] Refactor common cleaner logic across homebrew, npm, pnpm, nix, temp_files
33. [ ] Extract duplicate cleanup patterns (lines 172-188 appear in multiple files)
34. [ ] Fix internal/infrastructure/cleaners/nix.go duplications with system/nix.go

#### API Mapper Duplications
35. [ ] Refactor internal/interface/api/mapper_test.go - 10+ clone groups
36. [ ] Consolidate mapping test patterns (lines 554-620)
37. [ ] Extract common test setup patterns

### 🟡 MEDIUM - Code Quality Issues

#### Interface/Type Issues
38. [ ] Fix internal/domain/shared/interfaces.go duplicate lines 21:21, 33:33
39. [ ] Resolve internal/interface/api/types.go duplicate patterns
40. [ ] Fix internal/domain/shared/types.go multiple duplications (6+ groups)
41. [ ] Consolidate type casting patterns

#### Import/Module Structure
42. [ ] Fix missing internal/adapters package - create or remove references
43. [ ] Fix missing internal/bdd package - create or remove references
44. [ ] Fix internal/application/config/factories missing implementation
45. [ ] Review go.mod for missing package dependencies

#### Test Infrastructure
46. [ ] Fix undefined domain types in tests
47. [ ] Implement missing mockCleaner.Cleanup() method
48. [ ] Fix NewNixCleaner constructor signature mismatch
49. [ ] Review test helper functions for duplication

### 🟢 LOW - Documentation & Style

#### Documentation Issues
50. [ ] Add missing package documentation for internal modules
51. [ ] Document public API interfaces
52. [ ] Add examples for complex error handling patterns
53. [ ] Document configuration options and validation rules

#### Code Style Consistency
54. [ ] Standardize error message formatting
55. [ ] Consistent naming conventions across cleaners
56. [ ] Standardize test function naming
57. [ ] Consistent import organization

### 🔧 REFACTORING OPPORTUNITIES

#### Large Files (>300 lines)
58. [ ] Split internal/domain/shared/type_safe_enums.go (864 lines)
59. [ ] Split internal/shared/utils/pkg/errors/errors.go
60. [ ] Split internal/interface/api/mapper_test.go (>744 lines)
61. [ ] Split internal/application/config/safe_test.go
62. [ ] Split test/config/safety_level_test.go (>344 lines)

#### Common Patterns to Extract
63. [ ] Create common error factory functions
64. [ ] Extract test helper utilities
65. [ ] Create enum generation helper
66. [ ] Extract common validation patterns
67. [ ] Create base cleaner interface implementation
68. [ ] Extract HTTP client common patterns

### 📊 SPECIFIC FILE FIXES

#### Fix NewNixCleaner Signature Issues
69. [ ] internal/infrastructure/cleaners/nix_test.go:18:30
70. [ ] internal/infrastructure/cleaners/nix_test.go:36:30
71. [ ] internal/infrastructure/cleaners/nix_test.go:54:27
72. [ ] internal/infrastructure/cleaners/nix_test.go:71:27

#### Fix Undefined References
73. [ ] internal/interface/api/mapper_test.go:16:19 (domain)
74. [ ] internal/interface/api/mapper_test.go:23:18 (domain)
75. [ ] internal/interface/api/mapper_test.go:29:20 (domain)
76. [ ] internal/interface/api/mapper_test.go:31:29 (domain)
77. [ ] internal/infrastructure/system/adapters_test.go:13:8 (adapters)
78. [ ] test/domain/benchmarks_test.go:12:15 (RiskLevelType)

#### Mock Interface Issues
79. [ ] internal/shared/utils/middleware/validation_test.go:69:52
80. [ ] internal/shared/utils/middleware/validation_test.go:79:52

### 🏗️ ARCHITECTURAL IMPROVEMENTS

#### Package Structure
81. [ ] Review internal/adapters package necessity
82. [ ] Review internal/bdd package structure
83. [ ] Consolidate error handling packages
84. [ ] Review config package responsibilities

#### Interface Design
85. [ ] Review shared.Cleaner interface completeness
86. [ ] Standardize cleaner constructor patterns
87. [ ] Review domain model interfaces
88. [ ] Standardize validation interfaces

### 🔍 TESTING IMPROVEMENTS

#### Test Coverage Gaps
89. [ ] Add tests for internal/domain/config (no test files)
90. [ ] Add tests for internal/domain/shared (no test files)
91. [ ] Add tests for internal/domain/validation (no test files)
92. [ ] Add tests for cmd/clean-wizard (no test files)

#### Test Quality
93. [ ] Reduce test code duplication by 50%
94. [ ] Create shared test fixtures
95. [ ] Standardize test data setup
96. [ ] Add integration test coverage

### 📈 PERFORMANCE OPTIMIZATIONS

#### Code Efficiency
97. [ ] Review error handling performance impact
98. [ ] Optimize enum generation patterns
99. [ ] Review config loading performance
100. [ ] Optimize validation rules processing

### 🧹 CLEANUP TASKS

#### Dead Code Removal
101. [ ] Remove unused imports
102. [ ] Remove duplicate constants
103. [ ] Remove unused test helper functions
104. [ ] Clean up commented code blocks

#### Configuration Cleanup
105. [ ] Review config file duplications
106. [ ] Consolidate similar config patterns
107. [ ] Remove unused configuration options
108. [ ] Standardize config validation

---

## EXECUTION PRIORITY

### Week 1 (Critical Path)
1-9: Fix all build/test failures to get green build
10-25: Address major duplication sources
26-38: Fix high-impact architectural issues

### Week 2 (Quality Focus)
39-67: Refactor large files and extract patterns
68-88: Fix specific file issues and improve interfaces
89-100: Add missing test coverage

### Week 3 (Polish)
101-108: Code cleanup and performance
Documentation and final integration testing

---

## SUMMARY
- **Total Issues Identified**: 108 actionable tasks (under 250 limit)
- **Critical Build Failures**: 9 issues
- **Code Duplication Groups**: 218 clone groups to address
- **Missing Tests**: 4 packages without test coverage
- **Files >300 lines**: 5 files need splitting

*Generated using dupl and jscpd code analysis tools*
*Analysis date: 2025-12-07*