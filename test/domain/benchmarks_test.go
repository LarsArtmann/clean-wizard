package domain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// BenchmarkEnumHelper_String benchmarks String() method across all enum types
func BenchmarkEnumHelper_String(b *testing.B) {
	b.Run("RiskLevel", func(b *testing.B) {
		levels := []RiskLevelType{RiskLow, RiskMedium, RiskHigh, RiskCritical}
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = levels[i%len(levels)].String()
		}
	})

	b.Run("CleanStrategy", func(b *testing.B) {
		strategies := []CleanStrategyType{StrategyAggressive, StrategyConservative, StrategyDryRun}
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = strategies[i%len(strategies)].String()
		}
	})

	b.Run("ScanType", func(b *testing.B) {
		types := []ScanTypeType{ScanTypeNixStoreType, ScanTypeHomebrewType, ScanTypeSystemType, ScanTypeTempType}
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = types[i%len(types)].String()
		}
	})
}

// BenchmarkEnumHelper_IsValid benchmarks IsValid() method across all enum types
func BenchmarkEnumHelper_IsValid(b *testing.B) {
	b.Run("RiskLevel", func(b *testing.B) {
		levels := []RiskLevelType{RiskLow, RiskMedium, RiskHigh, RiskCritical}
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = levels[i%len(levels)].IsValid()
		}
	})

	b.Run("CleanStrategy", func(b *testing.B) {
		strategies := []CleanStrategyType{StrategyAggressive, StrategyConservative, StrategyDryRun}
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = strategies[i%len(strategies)].IsValid()
		}
	})
}

// BenchmarkResult_Creation benchmarks result type creation
func BenchmarkResult_Creation(b *testing.B) {
	b.Run("CleanResult", func(b *testing.B) {
		ts := time.Now() // Pre-compute timestamp outside the hot loop
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			_ = CleanResult{
				FreedBytes:   uint64(i),
				ItemsRemoved: uint(i),
				ItemsFailed:  0,
				CleanTime:    time.Duration(i),
				CleanedAt:    ts, // Use cached timestamp
				Strategy:     StrategyAggressive,
			}
		}
	})
}

// BenchmarkMemory_Usage benchmarks memory usage patterns
func BenchmarkMemory_Usage(b *testing.B) {
	b.Run("EnumValues", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = RiskLow.Values()
			_ = StrategyAggressive.Values()
			_ = ScanTypeNixStoreType.Values()
		}
	})
}

// PrintBenchmarkSummary prints benchmark summary for reporting
func PrintBenchmarkSummary(b *testing.B, testName string) {
	// This function can be called from test runners to print summary
	// Implementation depends on specific reporting needs
	fmt.Printf("Benchmark summary for %s completed\n", testName)
}
