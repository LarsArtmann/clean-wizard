# Architecture Review — 2026-05-03

## Scalability/Modularity Assessment

### Current Architecture Strengths

1. **Clean Registry Pattern** — `cleaner.Registry` provides thread-safe, polymorphic access to all cleaners via a consistent `Cleaner` interface
2. **Type-Safe Enums** — 19 iota-based enums with generic `EnumString`, `EnumIsValid`, `EnumValues` helpers eliminate stringly-typed code
3. **Result Type** — `result.Result[T]` with `Map`, `AndThen`, `Fold`, `Partition` provides functional error handling
4. **Flow Control** — `FlowBuilder[T]`, `ParallelFlow[T]`, `BranchFlow[T]` enable composable pipeline construction
5. **Adapter Pattern** — External tools wrapped in `internal/adapters/` (Nix, HTTP, exec, cache, rate limiter)

### Architecture Diagram (Current)

```
cmd/clean-wizard/commands/     CLI Layer (Cobra commands)
        │
        ├── domain/             Domain types, enums, interfaces
        │   ├── types.go        Core types: CleanResult, ScanResult, SizeEstimate
        │   ├── type_safe_enums.go  4 enums (Risk, Validation, ChangeOp, Strategy)
        │   ├── operation_types.go  15 OperationType constants
        │   ├── operation_settings.go 12 settings structs + 6 enum types
        │   ├── execution_enums.go   6 execution/config enums
        │   ├── enum_macros.go       Generic enum helpers
        │   └── interfaces.go        OperationHandler, Scanner, GenerationCleaner
        │
        ├── cleaner/            13+ Cleaner implementations
        │   ├── cleaner.go      Cleaner interface, CleanerBase, AgeBasedCleaner
        │   ├── registry.go     Thread-safe Registry
        │   ├── registry_factory.go  DefaultRegistry factory
        │   └── [13 cleaner files]
        │
        ├── config/             Configuration (koanf/yaml)
        ├── result/             Result[T], FlowBuilder, ParallelFlow, BranchFlow
        ├── adapters/           External tool adapters
        ├── middleware/          Validation middleware
        ├── conversions/        Unit conversions
        ├── format/             Byte formatting
        └── pkg/errors/         Error types
```

### Modularity Score: 7/10

**Good:**
- Clean package boundaries
- Domain types isolated in `internal/domain/`
- Registry pattern for extensibility
- Adapter pattern for external deps

**Needs Improvement:**
- `internal/domain/` is a "god package" with 20+ files
- `internal/cleaner/` is flat — 50+ files in one package
- Some cleaners reach into OS directly instead of through adapters
- Error packages scattered: `internal/errors/`, `internal/pkg/errors/`, `internal/adapters/errors.go`, `pkg/errors/`

### Service Orientation Score: 6/10

**Missing:**
- No formal service layer between CLI commands and domain
- Commands directly construct and call cleaners
- No dependency injection — cleaners construct their own dependencies
- No event/command system for async operations

### Composability Score: 8/10

**Excellent:**
- `result.Result[T]` enables clean function composition
- `FlowBuilder[T]` and `ParallelFlow[T]` enable pipeline composition
- `BranchFlow[T]` enables conditional flow composition
- Generic enum macros work across all enum types

**Weakness:**
- Cleaners don't compose well — each is monolithic
- No middleware chain for cleaner operations (scan → validate → clean → report)
- AgeBasedCleaner is the only composable extension point

## Recommendations

### 1. Split `internal/domain/` into Sub-Packages
```
internal/domain/
├── enums/           All enum types and macros
├── operations/      OperationType, OperationSettings, defaults
├── types/           Core domain types (CleanResult, ScanResult)
└── interfaces.go    Core interfaces
```

### 2. Split `internal/cleaner/` by Domain
```
internal/cleaner/
├── registry.go
├── cleaner.go
├── nix/
├── docker/
├── golang/
├── homebrew/
├── githistory/
└── ...
```

### 3. Consolidate Error Packages
```
internal/errors/  →  Single package with categorized errors
Remove: pkg/errors/, internal/pkg/errors/ (split brain)
```

### 4. Add Service Layer
```
internal/service/
├── clean_service.go    Orchestrates scan → clean → report
├── scan_service.go     Orchestrates multi-cleaner scanning
└── config_service.go   Configuration management
```

### 5. Extract Test Infrastructure
```
internal/cleaner/test_*.go  →  internal/testutil/cleaner/
```
Files: `test_assertions.go`, `test_factories.go`, `test_helpers.go`, `test_interfaces.go`, `testing_helpers.go`, `ginkgo_test_helpers.go` — 6 test utility files in production package.
