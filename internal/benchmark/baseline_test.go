package benchmark

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkSystemMetrics benchmarks system metric collection performance
func BenchmarkSystemMetrics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		_ = m
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate and release memory to simulate workload
		data := make([][]byte, 100)
		for j := range data {
			data[j] = make([]byte, 1024)
		}
		_ = data
	}
}

// BenchmarkGoroutineCreation benchmarks goroutine creation/destruction
func BenchmarkGoroutineCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		done := make(chan bool)
		go func() {
			// Simulate minimal work
			time.Sleep(time.Nanosecond)
			done <- true
		}()
		<-done
	}
}

// BenchmarkGCPressure benchmarks memory allocation and garbage collection
func BenchmarkGCPressure(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create memory pressure
		var objects [][]byte
		for j := 0; j < 10; j++ {
			objects = append(objects, make([]byte, 1024*10)) // 10KB each
		}
		// Force GC occasionally
		if i%100 == 0 {
			runtime.GC()
		}
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent goroutine operations
func BenchmarkConcurrentOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Simulate work
				time.Sleep(time.Nanosecond * 10)
			}()
		}
		wg.Wait()
	}
}

// EstablishPerformanceBaseline establishes a comprehensive performance baseline
func EstablishPerformanceBaseline(b *testing.B) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	b.Logf("Performance baseline established:")
	b.Logf("- Memory Alloc: %d bytes", m.Alloc)
	b.Logf("- Heap System: %d bytes", m.HeapSys)
	b.Logf("- Goroutines: %d", runtime.NumGoroutine())
	b.Logf("- NumGC: %d", m.NumGC)
	b.Logf("- CPU Info: %d CPUs, GoMaxProcs: %d", 
		runtime.NumCPU(), runtime.GOMAXPROCS(0))
}