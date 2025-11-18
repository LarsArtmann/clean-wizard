package benchmarks

import (
	"testing"
)

// Benchmark entry point - runs all benchmarks
func RunAllBenchmarks(b *testing.B) {
	// Conversion benchmarks
	b.Run("Conversions", func(b *testing.B) {
		BenchmarkNewCleanResult(b)
		BenchmarkCombineCleanResults(b)
	})
	
	// Configuration benchmarks
	b.Run("Configuration", func(b *testing.B) {
		BenchmarkConfigValidation(b)
		BenchmarkConfigSanitization(b)
	})
	
	// Nix cleaner benchmarks
	b.Run("NixCleaner", func(b *testing.B) {
		BenchmarkNixListGenerations(b)
		BenchmarkNixCleanOldGenerations(b)
	})
	
	// Adapter benchmarks
	b.Run("Adapters", func(b *testing.B) {
		BenchmarkRateLimiter(b)
		BenchmarkCacheOperations(b)
		BenchmarkHTTPClient(b)
	})
	
	// End-to-end benchmarks
	b.Run("EndToEnd", func(b *testing.B) {
		BenchmarkEndToEndWorkflow(b)
	})
	
	// Memory benchmarks
	b.Run("Memory", func(b *testing.B) {
		BenchmarkMemoryUsage(b)
	})
}