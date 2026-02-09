package benchmark

import (
	"errors"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// BenchmarkResult_Ok benchmarks creating successful results.
func BenchmarkResult_Ok(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = result.Ok[int](42)
	}
}

// BenchmarkResult_Err benchmarks creating error results.
func BenchmarkResult_Err(b *testing.B) {
	b.ReportAllocs()
	err := errors.New("test error")
	for b.Loop() {
		_ = result.Err[int](err)
	}
}

// BenchmarkResult_IsOk benchmarks checking if result is OK.
func BenchmarkResult_IsOk(b *testing.B) {
	r := result.Ok[int](42)
	b.ReportAllocs()

	for b.Loop() {
		_ = r.IsOk()
	}
}

// BenchmarkResult_Value benchmarks getting value from result.
func BenchmarkResult_Value(b *testing.B) {
	r := result.Ok[int](42)
	b.ReportAllocs()

	for b.Loop() {
		_ = r.Value()
	}
}

// BenchmarkCleanResult_Validate benchmarks CleanResult validation.
func BenchmarkCleanResult_Validate(b *testing.B) {
	cr := domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{Known: 1024, Status: domain.SizeEstimateStatusKnown},
		ItemsRemoved: 5,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    testTime,
		Strategy:     domain.CleanStrategyType(domain.StrategyAggressiveType),
	}
	b.ReportAllocs()

	for b.Loop() {
		_ = cr.Validate()
	}
}

var testTime = time.Now()
