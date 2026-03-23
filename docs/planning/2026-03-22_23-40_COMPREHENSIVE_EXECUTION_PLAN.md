# Clean Wizard - Comprehensive Multi-Step Execution Plan

**Date:** 2026-03-22  
**Status:** Ready for Execution  
**Estimated Total Effort:** 40-60 hours  
**Priority:** Sorted by Impact/Effort Ratio

---

## 🔍 REFLECTION: WHAT WAS MISSED & WHAT COULD BE BETTER

### 1. Library Usage Analysis

**Current State:**
- Custom `Result[T]` type (161 lines) - Reinventing the wheel
- Custom `Context[T]` type - Could use existing solutions
- Manual enum unmarshaling - Repetitive boilerplate

**What We Should Use:**

| Library | Purpose | Current Replacement |
|---------|---------|---------------------|
| `github.com/samber/mo` | Option, Result, Either types | Custom Result[T] (161 lines) |
| `github.com/invopop/validation` | Validation framework | Custom validation (400+ lines) |
| `github.com/go-playground/validator/v10` | Struct validation | Manual validation |
| `go.uber.org/zap` | Structured logging | fmt.Println scattered |
| `github.com/prometheus/client_golang` | Metrics | None - blind to usage |
| `github.com/hashicorp/go-multierror` | Error aggregation | Manual error handling |

### 2. Architecture Improvements

**DRY Violations Found:**
- Each enum type has ~50 lines of boilerplate (YAML, JSON, String)
- Cleaner constructors follow identical pattern (13× repetition)
- Test helpers have massive duplication

**Missing Abstractions:**
- No `AgeBasedCleaner` interface for time-based cleanup
- No `SizeEstimator` strategy pattern
- No `ProgressReporter` abstraction

### 3. Code Quality Issues

**Inconsistencies:**
- Mixed error wrapping styles (some wrap, some don't)
- Inconsistent context usage (some params unused)
- No structured logging (only fmt.Println)
- No tracing for long operations

**Performance Gaps:**
- Cleaners run sequentially (could parallelize)
- No caching of scan results
- Repeated os.Stat calls

### 4. Missing Features

**Observability:**
- No metrics on space freed
- No tracking of cleaner success rates
- No performance timing
- No user feedback mechanism

**User Experience:**
- No shell completions
- No man pages
- No `--verbose` levels (just on/off)
- No config profiles beyond risk levels

---

## 📋 COMPREHENSIVE MULTI-STEP EXECUTION PLAN

### PHASE 1: FOUNDATION IMPROVEMENTS (Week 1)

#### Step 1.1: Fix Critical Go Build Cache Gap ⭐ CRITICAL
**Impact:** HIGH | **Effort:** 1h | **Value/Effort:** ⭐⭐⭐⭐⭐

**Problem:** Misses `/private/var/folders/*/T/go-build*` on macOS  
**Impact:** Hundreds of MB to several GB uncleaned  
**Existing Code:** Already documented in `docs/issues/go-build-cache-gap.md`

**Implementation:**
```go
// Use os.TempDir() instead of hardcoded "/tmp"
func (gcc *GoCacheCleaner) cleanGoBuildCache(ctx context.Context) result.Result[domain.CleanResult] {
    locations := []string{
        os.TempDir(),           // /private/var/folders/... on macOS
        "/tmp",                  // Fallback
        filepath.Join(gcc.helper.getHomeDir(), "Library", "Caches"), // User cache
    }
    // Check all locations
}
```

**Verification:**
- Test on macOS: `go test ./internal/cleaner/... -run TestGoCache`
- Verify finds caches in `/private/var/folders`

---

#### Step 1.2: Create Generic Enum Generator
**Impact:** HIGH | **Effort:** 4h | **Value/Effort:** ⭐⭐⭐⭐

**Problem:** Each enum has ~50 lines of boilerplate  
**Existing Pattern:** `internal/domain/type_safe_enums.go` - 539 lines of repetition

**Implementation:**
```go
//go:generate go run ./cmd/enumgen -type=RiskLevel -values=Conservative,Moderate,Aggressive
// Generate: String(), MarshalJSON(), UnmarshalJSON(), MarshalYAML(), UnmarshalYAML(), IsValid(), Values()
```

**Reuse Existing:**
- Pattern from `UnmarshalYAMLEnum[T]` in `type_safe_enums.go`
- Just automate it with code generation

**Alternative:** Use `github.com/dmarkham/enumer` or `golang.org/x/tools/cmd/stringer` with enhancements

**Verification:**
- Generate all 8+ enum types
- Remove 400+ lines of boilerplate
- All tests pass

---

#### Step 1.3: Add Structured Logging with Zap
**Impact:** HIGH | **Effort:** 3h | **Value/Effort:** ⭐⭐⭐⭐

**Problem:** Scattered `fmt.Println`, no log levels, no structured output  
**Library:** `go.uber.org/zap` (industry standard)

**Implementation:**
```go
// internal/logger/logger.go
package logger

import "go.uber.org/zap"

var L *zap.Logger

func Init(verbose bool) {
    config := zap.NewProductionConfig()
    if verbose {
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    }
    L, _ = config.Build()
}

// Usage in cleaners:
logger.L.Info("cleaning cache",
    zap.String("cleaner", "docker"),
    zap.Int64("bytes_freed", bytes),
    zap.Duration("duration", elapsed),
)
```

**Migration Strategy:**
1. Add zap dependency
2. Create logger package
3. Replace fmt.Println in one cleaner (Docker as pilot)
4. Migrate remaining cleaners incrementally

**Verification:**
- All tests pass
- JSON output in production
- Human-readable in development

---

### PHASE 2: TYPE MODEL IMPROVEMENTS (Week 2)

#### Step 2.1: Create Age-Based Cleaner Interface
**Impact:** MEDIUM | **Effort:** 4h | **Value/Effort:** ⭐⭐⭐

**Problem:** No standardized way to handle age-based cleanup  
**Existing Code:** SystemCache has custom duration parsing

**New Interface:**
```go
// internal/cleaner/interfaces.go
type AgeBasedCleaner interface {
    Cleaner
    SetMaxAge(age time.Duration)
    GetMaxAge() time.Duration
}

// Generic implementation
func CleanWithAgeFilter(ctx context.Context, paths []string, maxAge time.Duration) result.Result[domain.CleanResult]
```

**Reuse Existing:**
- Duration parsing from `internal/domain/duration_parser.go`
- File walk logic from `internal/cleaner/fsutil.go`

**Verification:**
- Implements for SystemCache, TempFiles
- Tests for age filtering
- Integration with --older-than flag

---

#### Step 2.2: Add Cleaner Composition Framework
**Impact:** HIGH | **Effort:** 6h | **Value/Effort:** ⭐⭐⭐⭐

**Problem:** Cleaner logic is monolithic  
**Pattern:** Strategy + Decorator + Pipeline

**New Types:**
```go
// Strategy pattern for size estimation
type SizeEstimator interface {
    Estimate(path string) (int64, error)
}

// Decorator for retry logic
type RetryCleaner struct {
    base     Cleaner
    maxAttempts int
}

// Pipeline for chained operations
type CleanerPipeline []Cleaner
```

**Reuse Existing:**
- `Result[T]` chaining with AndThen/FlatMap
- Context propagation pattern

**Verification:**
- Refactor one cleaner (Cargo) as pilot
- All existing tests pass
- New composition tests added

---

#### Step 2.3: Implement Parallel Cleaner Execution
**Impact:** HIGH | **Effort:** 4h | **Value/Effort:** ⭐⭐⭐⭐

**Problem:** Cleaners run sequentially  
**Existing Code:** `internal/cleaner/registry.go` - sequential iteration

**Implementation:**
```go
func (r *Registry) CleanParallel(ctx context.Context) []CleanResult {
    cleaners := r.Available(ctx)
    results := make([]CleanResult, len(cleaners))
    
    var wg sync.WaitGroup
    for i, c := range cleaners {
        wg.Add(1)
        go func(idx int, cleaner Cleaner) {
            defer wg.Done()
            results[idx] = cleaner.Clean(ctx)
        }(i, c)
    }
    wg.Wait()
    return results
}
```

**Considerations:**
- Add `IsThreadSafe()` to Cleaner interface
- Respect dependency order (e.g., Docker before BuildKit)
- Limit concurrency (configurable)

**Verification:**
- Benchmark shows speedup
- No race conditions
- Progress reporting still works

---

### PHASE 3: OBSERVABILITY & METRICS (Week 3)

#### Step 3.1: Add Prometheus Metrics
**Impact:** MEDIUM | **Effort:** 4h | **Value/Effort:** ⭐⭐⭐

**Library:** `github.com/prometheus/client_golang`

**Metrics to Track:**
```go
var (
    bytesFreedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "clean_wizard_bytes_freed_total",
            Help: "Total bytes freed by cleaner",
        },
        []string{"cleaner"},
    )
    
    cleanupDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "clean_wizard_cleanup_duration_seconds",
            Help: "Time spent cleaning",
        },
        []string{"cleaner"},
    )
    
    cleanupsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "clean_wizard_cleanups_total",
            Help: "Total number of cleanups",
        },
        []string{"cleaner", "status"},
    )
)
```

**Verification:**
- Metrics endpoint at `:9090/metrics`
- Grafana dashboard JSON
- Documentation

---

#### Step 3.2: Add OpenTelemetry Tracing
**Impact:** MEDIUM | **Effort:** 6h | **Value/Effort:** ⭐⭐⭐

**Library:** `go.opentelemetry.io/otel`

**Implementation:**
```go
func (dc *DockerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    ctx, span := tracer.Start(ctx, "DockerCleaner.Clean")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("cleaner.name", dc.Name()),
        attribute.Bool("dry_run", dc.dryRun),
    )
    
    // ... cleanup logic
}
```

**Verification:**
- Traces visible in Jaeger/Zipkin
- Performance bottleneck identification
- Error tracing

---

### PHASE 4: USER EXPERIENCE (Week 4)

#### Step 4.1: Add Shell Completions
**Impact:** LOW | **Effort:** 2h | **Value/Effort:** ⭐⭐

**Library:** Cobra (already has support)

**Implementation:**
```go
// In cmd/clean-wizard/commands/root.go
rootCmd.AddCommand(&cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate completion script",
})
```

**Verification:**
- `clean-wizard completion bash > /etc/bash_completion.d/clean-wizard`
- Tab completion works

---

#### Step 4.2: Add Config Profile Templates
**Impact:** MEDIUM | **Effort:** 3h | **Value/Effort:** ⭐⭐⭐

**Problem:** Only 3 hardcoded profiles (conservative, balanced, aggressive)  
**Feature:** User-defined profiles + community templates

**Implementation:**
```go
type ProfileTemplate struct {
    Name        string
    Description string
    Category    string // "dev", "ci", "minimal", etc.
    Config      domain.Config
}

var BuiltInTemplates = []ProfileTemplate{
    DeveloperProfile,
    CIProfile,
    MinimalProfile,
    // ...
}
```

**Verification:**
- `clean-wizard profile list --templates`
- `clean-wizard profile create --from-template=developer`

---

### PHASE 5: ADVANCED FEATURES (Week 5-6)

#### Step 5.1: Implement Plugin Architecture
**Impact:** HIGH | **Effort:** 12h | **Value/Effort:** ⭐⭐⭐

**Library:** `github.com/hashicorp/go-plugin` or native Go plugins

**Design:**
```go
// Plugin interface
 type CleanerPlugin interface {
     Name() string
     IsAvailable(ctx context.Context) bool
     Clean(ctx context.Context) result.Result[domain.CleanResult]
 }
 
 // Plugin loader
 type PluginLoader struct {
     directory string
 }
```

**Verification:**
- Load external cleaner plugin
- Hot-reload capability
- Plugin marketplace structure

---

#### Step 5.2: Add Windows Support
**Impact:** HIGH | **Effort:** 16h | **Value/Effort:** ⭐⭐⭐

**New Package:** `internal/cleaner/windows/`

**Cleaners to Implement:**
- Windows Temp: `%TEMP%`, `%LOCALAPPDATA%\Temp`
- Windows Update Cache
- Browser caches (Chrome, Edge, Firefox on Windows)
- Recycle Bin
- Thumbnail cache
- DNS cache

**Verification:**
- CI/CD with Windows runner
- Test on Windows 10/11
- Documentation

---

## 🎯 IMPACT/EFFORT SORTED SUMMARY

| Rank | Step | Impact | Effort | Value/Effort | Phase |
|------|------|--------|--------|--------------|-------|
| 1 | Fix Go Build Cache Gap | ⭐⭐⭐⭐⭐ | 1h | ⭐⭐⭐⭐⭐ | 1.1 |
| 2 | Parallel Cleaner Execution | ⭐⭐⭐⭐⭐ | 4h | ⭐⭐⭐⭐⭐ | 2.3 |
| 3 | Structured Logging (Zap) | ⭐⭐⭐⭐⭐ | 3h | ⭐⭐⭐⭐ | 1.3 |
| 4 | Enum Generator | ⭐⭐⭐⭐ | 4h | ⭐⭐⭐⭐ | 1.2 |
| 5 | Cleaner Composition | ⭐⭐⭐⭐ | 6h | ⭐⭐⭐⭐ | 2.2 |
| 6 | Age-Based Cleaner Interface | ⭐⭐⭐ | 4h | ⭐⭐⭐ | 2.1 |
| 7 | Prometheus Metrics | ⭐⭐⭐ | 4h | ⭐⭐⭐ | 3.1 |
| 8 | OpenTelemetry Tracing | ⭐⭐⭐ | 6h | ⭐⭐⭐ | 3.2 |
| 9 | Config Profile Templates | ⭐⭐⭐ | 3h | ⭐⭐⭐ | 4.2 |
| 10 | Windows Support | ⭐⭐⭐⭐⭐ | 16h | ⭐⭐⭐ | 5.2 |
| 11 | Plugin Architecture | ⭐⭐⭐⭐ | 12h | ⭐⭐⭐ | 5.1 |
| 12 | Shell Completions | ⭐⭐ | 2h | ⭐⭐ | 4.1 |

---

## 🔧 TYPE MODEL IMPROVEMENTS ARCHITECTURE

### Current State
```go
// Monolithic cleaner
type DockerCleaner struct {
    verbose bool
    dryRun  bool
    // ... 20+ fields
}

func (dc *DockerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
    // 200+ lines of mixed logic
}
```

### Improved Architecture
```go
// Composable behaviors
type BaseCleaner struct {
    name    string
    verbose bool
    dryRun  bool
}

type SizeEstimable struct {
    estimator SizeEstimator
}

type AgeFilterable struct {
    maxAge time.Duration
}

// Docker cleaner composed of behaviors
type DockerCleaner struct {
    *BaseCleaner
    *SizeEstimable
    ContainerCleaner
    ImageCleaner
    VolumeCleaner
}
```

### Benefits
- **Single Responsibility:** Each behavior is testable independently
- **DRY:** Reuse behaviors across cleaners
- **Extensible:** Add new behaviors without changing existing code
- **Testable:** Mock individual behaviors

---

## 📚 LIBRARY RECOMMENDATIONS

### Must Add (High Value)
1. **`go.uber.org/zap`** - Structured logging (production ready)
2. **`github.com/prometheus/client_golang`** - Metrics
3. **`github.com/samber/mo`** - Option/Result types (can replace custom)

### Should Add (Medium Value)
4. **`github.com/go-playground/validator/v10`** - Struct validation
5. **`github.com/hashicorp/go-multierror`** - Error aggregation
6. **`go.opentelemetry.io/otel`** - Distributed tracing

### Consider (Low Priority)
7. **`github.com/dmarkham/enumer`** - Enum generation
8. **`github.com/hashicorp/go-plugin`** - Plugin system

---

## ✅ EXECUTION CHECKLIST

For each step:
- [ ] Read existing code thoroughly
- [ ] Research library alternatives
- [ ] Create branch: `feature/step-X.Y-description`
- [ ] Implement with tests
- [ ] Run `go build ./...`
- [ ] Run `go test ./... -short`
- [ ] Update documentation
- [ ] Commit with detailed message
- [ ] Push to remote
- [ ] Update this plan with completion status

---

## 🚨 CRITICAL QUESTIONS

### Q1: Should we replace custom Result[T] with samber/mo?
**Trade-off:**
- Keep: Full control, no external dependency
- Replace: Less code to maintain, community-tested

**Recommendation:** Keep for now (already working), consider for v2.0

### Q2: How do we handle breaking changes?
**Options:**
- Semantic versioning (v1.x → v2.0)
- Feature flags
- Migration scripts

**Recommendation:** Use feature flags, gradual migration

### Q3: What's the priority: Windows support or Plugin architecture?
**Analysis:**
- Windows: Larger user base, immediate value
- Plugin: Long-term extensibility, smaller impact

**Recommendation:** Windows first (Step 5.2), then plugins (Step 5.1)

---

## 📊 SUCCESS METRICS

After completing this plan:
- [ ] Zero critical bugs
- [ ] 50%+ reduction in boilerplate code
- [ ] Sub-second cleanup for common operations
- [ ] Complete observability (logs, metrics, traces)
- [ ] Windows parity with macOS/Linux
- [ ] Plugin marketplace ready

---

**Next Action:** Execute Step 1.1 (Fix Go Build Cache Gap)

**Status:** ✅ Ready to begin execution

---

*Plan created by: Parakletos*  
*Last updated: 2026-03-22 23:40*
