package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ProfileBenchmark profiles specific operations for optimization
func ProfileBenchmark(b *testing.B, name string, operation func() error, iterations int) {
	b.Logf("\n=== Profiling %s ===", name)

	// Warmup
	for range 5 {
		if err := operation(); err != nil {
			b.Fatalf("Warmup operation failed: %v", err)
		}
	}

	// Actual benchmarking
	start := time.Now()
	errors := int64(0)

	for range iterations {
		if err := operation(); err != nil {
			errors++
		}
	}

	duration := time.Since(start)
	successRate := float64(int(iterations)-int(errors)) / float64(iterations) * 100

	b.Logf("Operation: %s", name)
	b.Logf("Iterations: %d", iterations)
	b.Logf("Duration: %v", duration)
	b.Logf("Average per operation: %v", duration/time.Duration(iterations))
	b.Logf("Operations per second: %.2f", float64(iterations)/duration.Seconds())
	b.Logf("Success rate: %.2f%%", successRate)
	b.Logf("Errors: %d", errors)
	b.Logf("===============================\n")

	if errors > int64(iterations/10) { // More than 10% errors
		b.Errorf("High error rate: %d/%d (%.2f%%)", errors, iterations, float64(errors)/float64(iterations)*100)
	}
}

// ConcurrentBenchmark tests concurrent performance
func ConcurrentBenchmark(b *testing.B, name string, operation func() error, concurrency, iterations int) {
	b.Logf("\n=== Concurrent Profiling %s ===", name)
	b.Logf("Concurrency: %d", concurrency)
	b.Logf("Total Iterations: %d", iterations)

	ctx := context.Background()
	rateLimiter := adapters.NewRateLimiter(1000, 100)

	// Create workers
	jobs := make(chan int, iterations)
	results := make(chan error, iterations)

	// Start workers
	for i := range concurrency {
		go func(workerID int) {
			for range jobs {
				// Rate limit to prevent overwhelming
				if err := rateLimiter.Wait(ctx); err != nil {
					results <- err
					continue
				}

				results <- operation()
			}
		}(i)
	}

	// Distribute jobs
	start := time.Now()
	for i := range iterations {
		jobs <- i
	}
	close(jobs)

	// Collect results
	errors := int64(0)
	for range iterations {
		if err := <-results; err != nil {
			errors++
		}
	}

	duration := time.Since(start)
	successRate := float64(int(iterations)-int(errors)) / float64(iterations) * 100

	b.Logf("Operation: %s", name)
	b.Logf("Duration: %v", duration)
	b.Logf("Throughput: %.2f ops/sec", float64(iterations)/duration.Seconds())
	b.Logf("Per-worker throughput: %.2f ops/sec", float64(iterations)/duration.Seconds()/float64(concurrency))
	b.Logf("Success rate: %.2f%%", successRate)
	b.Logf("Errors: %d", errors)
	b.Logf("=====================================\n")

	if errors > int64(iterations/10) { // More than 10% errors
		b.Errorf("High error rate in concurrent test: %d/%d (%.2f%%)", errors, iterations, float64(errors)/float64(iterations)*100)
	}
}

// ScaleBenchmark tests performance at different scales
func ScaleBenchmark(b *testing.B, name string, operation func(scale int) error, scales []int) {
	b.Logf("\n=== Scale Testing %s ===", name)

	for _, scale := range scales {
		iterations := 1000 // Fixed iterations per scale

		start := time.Now()
		errors := int64(0)

		for range iterations {
			if err := operation(scale); err != nil {
				errors++
			}
		}

		duration := time.Since(start)
		successRate := float64(int(iterations)-int(errors)) / float64(iterations) * 100

		b.Logf("Scale: %d", scale)
		b.Logf("Duration: %v", duration)
		b.Logf("Ops/sec: %.2f", float64(iterations)/duration.Seconds())
		b.Logf("Success rate: %.2f%%", successRate)
		b.Logf("Errors: %d", errors)
		b.Logf("--------")

		if int(errors) > iterations/10 {
			b.Errorf("High error rate at scale %d: %d/%d", scale, errors, iterations)
		}
	}

	b.Logf("===============================\n")
}

// StressBenchmark performs stress testing
func StressBenchmark(b *testing.B, name string, operation func() error, duration time.Duration) {
	b.Logf("\n=== Stress Testing %s (Duration: %v) ===", name, duration)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	operations := int64(0)
	errors := int64(0)
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			operations++
			if err := operation(); err != nil {
				errors++
			}
		}
	}

	actualDuration := time.Since(start)
	successRate := float64(operations-errors) / float64(operations) * 100

	b.Logf("Operation: %s", name)
	b.Logf("Target Duration: %v", duration)
	b.Logf("Actual Duration: %v", actualDuration)
	b.Logf("Total Operations: %d", operations)
	b.Logf("Errors: %d", errors)
	b.Logf("Success Rate: %.2f%%", successRate)
	b.Logf("Throughput: %.2f ops/sec", float64(operations)/actualDuration.Seconds())
	b.Logf("=======================================\n")

	if errors > operations/10 {
		b.Errorf("High error rate during stress test: %d/%d (%.2f%%)", errors, operations, float64(errors)/float64(operations)*100)
	}
}

// ===== SPECIFIC PROFILING BENCHMARKS =====

// ProfileConfigurationValidation profiles configuration validation performance
func ProfileConfigurationValidation(b *testing.B) {
	suite := NewBenchmarkSuite()
	validator := config.NewConfigValidator()

	operation := func() error {
		result := validator.ValidateConfig(suite.Config)
		if !result.IsValid {
			return fmt.Errorf("validation failed: %v", result.Errors)
		}
		return nil
	}

	ProfileBenchmark(b, "ConfigurationValidation", operation, 10000)

	// Concurrent version
	ConcurrentBenchmark(b, "ConfigurationValidation-Concurrent", operation, 10, 10000)
}

// ProfileNixOperations profiles Nix cleaner operations
func ProfileNixOperations(b *testing.B) {
	ctx := context.Background()

	operation := func() error {
		nixCleaner := cleaner.NewNixCleaner(true, false) // Dry run
		result := nixCleaner.ListGenerations(ctx)
		if result.IsErr() {
			return fmt.Errorf("list generations failed: %w", result.Error())
		}
		return nil
	}

	ProfileBenchmark(b, "NixListGenerations", operation, 1000)

	// Concurrent version
	ConcurrentBenchmark(b, "NixListGenerations-Concurrent", operation, 5, 1000)
}

// ProfileConversions profiles conversion operations
func ProfileConversions(b *testing.B) {
	operation := func() error {
		result := conversions.NewCleanResult(domain.StrategyDryRun, 100, 1024*1024)
		_ = result
		return nil
	}

	ProfileBenchmark(b, "CleanResultCreation", operation, 100000)

	// Concurrent version
	ConcurrentBenchmark(b, "CleanResultCreation-Concurrent", operation, 20, 100000)
}

// ProfileCacheOperations profiles caching performance
func ProfileCacheOperations(b *testing.B) {
	suite := NewBenchmarkSuite()

	// Pre-populate cache
	for i := range 1000 {
		suite.CacheManager.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	getOperation := func() error {
		key := fmt.Sprintf("key_%d", 1000%1000) // Cycle through keys
		_, found := suite.CacheManager.Get(key)
		if !found {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return nil
	}

	setOperation := func() error {
		key := fmt.Sprintf("set_key_%d", 1000%1000)
		value := fmt.Sprintf("set_value_%d", 1000%1000)
		suite.CacheManager.Set(key, value, time.Minute)
		return nil
	}

	ProfileBenchmark(b, "CacheGet", getOperation, 100000)
	ProfileBenchmark(b, "CacheSet", setOperation, 100000)

	// Concurrent versions
	ConcurrentBenchmark(b, "CacheGet-Concurrent", getOperation, 10, 100000)
	ConcurrentBenchmark(b, "CacheSet-Concurrent", setOperation, 10, 100000)
}

// ProfileRateLimiting profiles rate limiter performance
func ProfileRateLimiting(b *testing.B) {
	suite := NewBenchmarkSuite()
	ctx := context.Background()

	operation := func() error {
		return suite.RateLimiter.Wait(ctx)
	}

	ProfileBenchmark(b, "RateLimiter", operation, 100000)

	// Concurrent version with contention
	ConcurrentBenchmark(b, "RateLimiter-Concurrent", operation, 50, 100000)
}

// ProfileScaleConfigurationValidation profiles configuration validation at different scales
func ProfileScaleConfigurationValidation(b *testing.B) {
	validator := config.NewConfigValidator()
	scales := []int{1, 10, 50, 100, 500, 1000}

	operation := func(scale int) error {
		// Create config with 'scale' profiles
		cfg := &domain.Config{
			Version:      "1.0.0",
			SafeMode:     true,
			MaxDiskUsage: 50,
			Protected:    []string{"/System", "/Applications"},
			Profiles:     make(map[string]*domain.Profile, scale),
		}

		// Populate profiles
		for i := range scale {
			cfg.Profiles[fmt.Sprintf("profile_%d", i)] = &domain.Profile{
				Name:        fmt.Sprintf("Profile %d", i),
				Description: fmt.Sprintf("Profile %d for testing", i),
				Operations: []domain.CleanupOperation{
					{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
				},
				Enabled: true,
			}
		}

		result := validator.ValidateConfig(cfg)
		if !result.IsValid {
			return fmt.Errorf("validation failed at scale %d: %v", scale, result.Errors)
		}
		return nil
	}

	ScaleBenchmark(b, "ScaleConfigurationValidation", operation, scales)
}

// StressTestCacheCaching performs stress test on cache
func StressTestCacheCaching(b *testing.B) {
	suite := NewBenchmarkSuite()

	operation := func() error {
		// Mix of gets and sets
		key := fmt.Sprintf("stress_key_%d", 10000%1000)

		if 10000%2 == 0 {
			suite.CacheManager.Set(key, fmt.Sprintf("stress_value_%d", 10000), time.Minute)
		} else {
			suite.CacheManager.Get(key)
		}
		return nil
	}

	StressBenchmark(b, "CacheStress", operation, 10*time.Second)
}

// StressTestNixOperations performs stress test on Nix operations
func StressTestNixOperations(b *testing.B) {
	ctx := context.Background()

	operation := func() error {
		nixCleaner := cleaner.NewNixCleaner(true, false) // Dry run
		result := nixCleaner.ListGenerations(ctx)
		if result.IsErr() {
			return fmt.Errorf("stress list generations error: %w", result.Error())
		}
		return nil
	}

	StressBenchmark(b, "NixStress", operation, 5*time.Second)
}
