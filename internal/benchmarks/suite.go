package benchmarks

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BenchmarkSuite holds all benchmark configurations
type BenchmarkSuite struct {
	NixCleaner   *cleaner.NixCleaner
	Config       *domain.Config
	RateLimiter  *adapters.RateLimiter
	CacheManager *adapters.CacheManager
}

// NewBenchmarkSuite creates a new benchmark suite with test data
func NewBenchmarkSuite() *BenchmarkSuite {
	// Create test configuration
	cfg := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Applications"},
		Profiles: map[string]*domain.Profile{
			"nix-cleanup": {
				Name:        "Nix Cleanup",
				Description: "Clean Nix generations",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							NixGenerations: &domain.NixGenerationsSettings{
								Generations: 3,
								Optimize:    true,
							},
						},
					},
				},
				Enabled: true,
			},
		},
	}

	return &BenchmarkSuite{
		NixCleaner:   cleaner.NewNixCleaner(true, false), // Dry run, not enhanced
		Config:       cfg,
		RateLimiter:  adapters.NewRateLimiter(100, 50),
		CacheManager: adapters.NewCacheManager(5*time.Minute, 10*time.Minute),
	}
}

// BenchmarkMetrics collects comprehensive performance metrics
type BenchmarkMetrics struct {
	Operations     int64            `json:"operations"`
	Duration       time.Duration    `json:"duration"`
	BytesProcessed int64            `json:"bytes_processed"`
	ItemsProcessed int64            `json:"items_processed"`
	MemoryUsage    runtime.MemStats `json:"memory_usage"`
	ThroughputQPS  float64          `json:"throughput_qps"`
	LatencyP95     time.Duration    `json:"latency_p95"`
	ErrorCount     int64            `json:"error_count"`
}

// CollectMetrics gathers performance metrics for a benchmark
func CollectMetrics(ops int64, duration time.Duration, bytes int64, items int64, errors int64) BenchmarkMetrics {
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	return BenchmarkMetrics{
		Operations:     ops,
		Duration:       duration,
		BytesProcessed: bytes,
		ItemsProcessed: items,
		MemoryUsage:    memStats,
		ThroughputQPS:  float64(ops) / duration.Seconds(),
		ErrorCount:     errors,
	}
}

// PrintBenchmarkResults outputs formatted benchmark results
func PrintBenchmarkResults(b *testing.B, name string, metrics BenchmarkMetrics) {
	b.Logf("\n=== %s Benchmark Results ===", name)
	b.Logf("Operations: %d", metrics.Operations)
	b.Logf("Duration: %v", metrics.Duration)
	b.Logf("Throughput: %.2f ops/sec", metrics.ThroughputQPS)
	b.Logf("Bytes Processed: %d", metrics.BytesProcessed)
	b.Logf("Items Processed: %d", metrics.ItemsProcessed)
	b.Logf("Errors: %d", metrics.ErrorCount)
	b.Logf("Memory Usage: %.2f MB", float64(metrics.MemoryUsage.Alloc)/1024/1024)
	b.Logf("Memory Allocs: %d", metrics.MemoryUsage.Mallocs)
	b.Logf("GC Cycles: %d", metrics.MemoryUsage.NumGC)
	b.Logf("================================\n")
}

// RunBenchmarkWithCleanup runs a benchmark with automatic cleanup
func RunBenchmarkWithCleanup(b *testing.B, name string, setup func() error, benchmark func(b *testing.B), cleanup func() error) {
	b.Run(name, func(b *testing.B) {
		if setup != nil {
			if err := setup(); err != nil {
				b.Fatalf("Setup failed: %v", err)
			}
		}

		defer func() {
			if cleanup != nil {
				if err := cleanup(); err != nil {
					b.Logf("Cleanup warning: %v", err)
				}
			}
		}()

		benchmark(b)
	})
}

// Warmup performs warmup operations before benchmarking
func Warmup(warmupCount int, warmupFunc func() error) error {
	for i := range warmupCount {
		if err := warmupFunc(); err != nil {
			return fmt.Errorf("warmup failed at iteration %d: %w", i, err)
		}
	}
	return nil
}

// ===== CONVERSION BENCHMARKS =====

// BenchmarkNewCleanResult benchmarks clean result creation
func BenchmarkNewCleanResult(b *testing.B) {

	for b.Loop() {
		result := conversions.NewCleanResult(domain.StrategyDryRun, 100, 1024*1024*100)
		_ = result
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), 1024*1024*100*int64(b.N), int64(b.N)*100, 0)
	PrintBenchmarkResults(b, "NewCleanResult", metrics)
}

// BenchmarkCombineCleanResults benchmarks result combination
func BenchmarkCombineCleanResults(b *testing.B) {
	// Pre-create results for combination
	results := make([]domain.CleanResult, 10)
	for i := range results {
		results[i] = conversions.NewCleanResult(domain.StrategyDryRun, 10, 1024*1024)
	}

	for b.Loop() {
		combined := conversions.CombineCleanResults(results)
		_ = combined
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), int64(b.N)*10*1024*1024, int64(b.N)*100, 0)
	PrintBenchmarkResults(b, "CombineCleanResults", metrics)
}

// ===== CONFIG BENCHMARKS =====

// BenchmarkConfigValidation benchmarks configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	validator := config.NewConfigValidator()
	cfg := NewBenchmarkSuite().Config

	for b.Loop() {
		result := validator.ValidateConfig(cfg)
		_ = result
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), 0, int64(b.N), 0)
	PrintBenchmarkResults(b, "ConfigValidation", metrics)
}

// BenchmarkConfigSanitization benchmarks configuration sanitization
func BenchmarkConfigSanitization(b *testing.B) {
	sanitizer := config.NewConfigSanitizer()
	validator := config.NewConfigValidator()
	cfg := NewBenchmarkSuite().Config

	for b.Loop() {
		validationResult := validator.ValidateConfig(cfg)
		sanitizer.SanitizeConfig(cfg, validationResult)
		_ = validationResult
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), 0, int64(b.N), 0)
	PrintBenchmarkResults(b, "ConfigSanitization", metrics)
}

// ===== NIX CLEANER BENCHMARKS =====

// BenchmarkNixListGenerations benchmarks Nix generation listing
func BenchmarkNixListGenerations(b *testing.B) {
	suite := NewBenchmarkSuite()
	ctx := context.Background()

	// Warmup
	if err := Warmup(3, func() error {
		result := suite.NixCleaner.ListGenerations(ctx)
		if result.IsErr() {
			return result.Error()
		}
		return nil
	}); err != nil {
		b.Fatalf("Warmup failed: %v", err)
	}

	for b.Loop() {
		result := suite.NixCleaner.ListGenerations(ctx)
		if result.IsErr() {
			b.Errorf("ListGenerations returned error: %v", result.Error())
		}
		_ = result
	}

	// Estimate bytes processed (rough estimate)
	bytesProcessed := int64(b.N) * 5 * 50 * 1024 * 1024 // 5 generations * 50MB each
	itemsProcessed := int64(b.N) * 5

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), bytesProcessed, itemsProcessed, 0)
	PrintBenchmarkResults(b, "NixListGenerations", metrics)
}

// BenchmarkNixCleanOldGenerations benchmarks Nix generation cleaning
func BenchmarkNixCleanOldGenerations(b *testing.B) {
	suite := NewBenchmarkSuite()
	ctx := context.Background()

	// Warmup
	if err := Warmup(3, func() error {
		result := suite.NixCleaner.CleanOldGenerations(ctx, 3)
		if result.IsErr() {
			return result.Error()
		}
		return nil
	}); err != nil {
		b.Fatalf("Warmup failed: %v", err)
	}

	var totalBytes int64
	var totalItems int64

	for b.Loop() {
		result := suite.NixCleaner.CleanOldGenerations(ctx, 3)
		if result.IsOk() {
			cleanResult := result.Value()
			totalBytes += cleanResult.FreedBytes
			totalItems += int64(cleanResult.ItemsRemoved)
		} else if result.IsErr() {
			b.Errorf("CleanOldGenerations failed: %v", result.Error())
		}
		_ = result
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), totalBytes, totalItems, 0)
	PrintBenchmarkResults(b, "NixCleanOldGenerations", metrics)
}

// ===== ADAPTER BENCHMARKS =====

// BenchmarkRateLimiter benchmarks rate limiting operations
func BenchmarkRateLimiter(b *testing.B) {
	ctx := context.Background()

	// Create a high-limit rate limiter for benchmarking
	rl := adapters.NewRateLimiter(10000, 1000)

	for b.Loop() {
		err := rl.Wait(ctx)
		if err != nil {
			b.Errorf("RateLimiter.Wait failed: %v", err)
		}
	}

	metrics := CollectMetrics(int64(b.N), b.Elapsed(), 0, int64(b.N), 0)
	PrintBenchmarkResults(b, "RateLimiter", metrics)
}

// BenchmarkCacheOperations benchmarks cache set/get operations
func BenchmarkCacheOperations(b *testing.B) {
	cache := adapters.NewCacheManager(5*time.Minute, 10*time.Minute)

	// Pre-populate cache for get benchmarks
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key_%d", i)
		cache.Set(keys[i], fmt.Sprintf("value_%d", i), time.Minute)
	}

	// Benchmark both set and get operations
	for i := 0; b.Loop(); i++ {
		key := keys[i%len(keys)]

		// Set operation
		cache.Set(key, fmt.Sprintf("value_%d", i), time.Minute)

		// Get operation
		_, found := cache.Get(key)
		if !found && i < len(keys) {
			b.Errorf("Expected key %s to be found", key)
		}
	}

	metrics := CollectMetrics(int64(b.N*2), b.Elapsed(), 0, int64(b.N*2), 0)
	PrintBenchmarkResults(b, "CacheOperations", metrics)
}

// BenchmarkHTTPClient benchmarks HTTP client operations
func BenchmarkHTTPClient(b *testing.B) {
	client := adapters.NewHTTPClient()
	ctx := context.Background()

	// Use a public API for benchmarking
	url := "https://httpbin.org/json"

	for b.Loop() {
		resp, err := client.Get(ctx, url)
		if err != nil {
			b.Skipf("HTTP request failed (likely network issue): %v", err)
			return
		}
		_ = resp
	}

	bytesPerReq := int64(500) // Rough estimate
	metrics := CollectMetrics(int64(b.N), b.Elapsed(), bytesPerReq*int64(b.N), int64(b.N), 0)
	PrintBenchmarkResults(b, "HTTPClient", metrics)
}

// ===== COMPREHENSIVE BENCHMARKS =====

// BenchmarkEndToEndWorkflow benchmarks complete cleaning workflow
func BenchmarkEndToEndWorkflow(b *testing.B) {
	validator := config.NewConfigValidator()
	sanitizer := config.NewConfigSanitizer()
	suite := NewBenchmarkSuite()
	ctx := context.Background()

	for b.Loop() {
		// Validate configuration
		validationResult := validator.ValidateConfig(suite.Config)
		if !validationResult.IsValid {
			b.Errorf("Validation failed: %v", validationResult.Errors)
		}

		// Sanitize configuration
		sanitizer.SanitizeConfig(suite.Config, validationResult)

		// List generations
		generationsResult := suite.NixCleaner.ListGenerations(ctx)
		if generationsResult.IsErr() {
			b.Errorf("ListGenerations returned error: %v", generationsResult.Error())
		}

		// Clean old generations (dry run)
		cleanResult := suite.NixCleaner.CleanOldGenerations(ctx, 3)
		if cleanResult.IsErr() {
			b.Errorf("CleanOldGenerations returned error: %v", cleanResult.Error())
		}
	}

	metrics := CollectMetrics(int64(b.N*4), b.Elapsed(), 0, int64(b.N*4), 0)
	PrintBenchmarkResults(b, "EndToEndWorkflow", metrics)
}

// ===== MEMORY BENCHMARKS =====

// BenchmarkMemoryUsage tracks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	_ = NewBenchmarkSuite()

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Simulate memory-intensive operations
	configs := make([]*domain.Config, b.N)
	results := make([]domain.CleanResult, b.N)

	for i := 0; b.Loop(); i++ {
		configs[i] = &domain.Config{
			Version:      "1.0.0",
			SafeMode:     true,
			MaxDiskUsage: 50,
			Protected:    make([]string, 10),
			Profiles:     make(map[string]*domain.Profile),
		}

		results[i] = conversions.NewCleanResult(domain.StrategyDryRun, 100, 1024*1024)
	}

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	memoryUsed := memAfter.Alloc - memBefore.Alloc
	allocsDiff := memAfter.Mallocs - memBefore.Mallocs

	b.Logf("\n=== Memory Usage Results ===")
	b.Logf("Memory Before: %.2f MB", float64(memBefore.Alloc)/1024/1024)
	b.Logf("Memory After: %.2f MB", float64(memAfter.Alloc)/1024/1024)
	b.Logf("Memory Used: %.2f MB", float64(memoryUsed)/1024/1024)
	b.Logf("Memory Per Config: %.2f KB", float64(memoryUsed)/float64(b.N)/1024)
	b.Logf("Allocs Difference: %d", allocsDiff)
	b.Logf("Allocs Per Config: %.2f", float64(allocsDiff)/float64(b.N))
	b.Logf("========================\n")

	// Prevent compiler optimizations
	_ = configs
	_ = results
}
