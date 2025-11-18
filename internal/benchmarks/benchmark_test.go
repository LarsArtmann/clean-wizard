package benchmarks

import (
	"testing"
)

// Main benchmark runner
func BenchmarkMain(b *testing.B) {
	b.Logf("Starting Clean Wizard Benchmark Suite")
	b.Logf("Go Version: %s", b.Name())
	b.Logf("CPU Count: %d", b.N)
	
	RunAllBenchmarks(b)
}

// Individual benchmark entry points
func BenchmarkConversions(b *testing.B) {
	b.Run("NewCleanResult", BenchmarkNewCleanResult)
	b.Run("CombineCleanResults", BenchmarkCombineCleanResults)
}

func BenchmarkConfiguration(b *testing.B) {
	b.Run("Validation", BenchmarkConfigValidation)
	b.Run("Sanitization", BenchmarkConfigSanitization)
}

func BenchmarkNixCleaner(b *testing.B) {
	b.Run("ListGenerations", BenchmarkNixListGenerations)
	b.Run("CleanOldGenerations", BenchmarkNixCleanOldGenerations)
}

func BenchmarkAdapters(b *testing.B) {
	b.Run("RateLimiter", BenchmarkRateLimiter)
	b.Run("CacheOperations", BenchmarkCacheOperations)
	b.Run("HTTPClient", BenchmarkHTTPClient)
}

func BenchmarkEndToEnd(b *testing.B) {
	b.Run("Workflow", BenchmarkEndToEndWorkflow)
}

func BenchmarkMemory(b *testing.B) {
	b.Run("MemoryUsage", BenchmarkMemoryUsage)
}

// Profiling benchmarks
func BenchmarkProfiling(b *testing.B) {
	b.Run("ConfigurationValidation", ProfileConfigurationValidation)
	b.Run("NixOperations", ProfileNixOperations)
	b.Run("Conversions", ProfileConversions)
	b.Run("CacheOperations", ProfileCacheOperations)
	b.Run("RateLimiting", ProfileRateLimiting)
	b.Run("ScaleConfigurationValidation", ProfileScaleConfigurationValidation)
}

func BenchmarkStress(b *testing.B) {
	b.Run("CacheStress", StressTestCacheCaching)
	b.Run("NixStress", StressTestNixOperations)
}