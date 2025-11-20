package benchmark

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkSystemMetrics benchmarks system metric collection performance
func BenchmarkSystemMetrics(b *testing.B) {
	for b.Loop() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		_ = m
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	for b.Loop() {
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
	for b.Loop() {
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
	for i := 0; b.Loop(); i++ {
		// Create memory pressure
		var objects [][]byte
		for range 10 {
			objects = append(objects, make([]byte, 1024*10)) // 10KB each
		}
		// Keep objects alive to prevent optimization
		runtime.KeepAlive(objects)
		// Force GC occasionally
		if i%100 == 0 {
			runtime.GC()
		}
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent goroutine operations
func BenchmarkConcurrentOperations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			for range 10 {
				wg.Go(func() {
					// Simulate work
					time.Sleep(time.Nanosecond * 10)
				})
			}
			wg.Wait()
		}
	})
}
