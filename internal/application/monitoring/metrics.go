package monitoring

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/shared/errors"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// MetricsCollector collects performance and operation metrics
type MetricsCollector struct {
	mutex      sync.RWMutex
	metrics    map[string]*Metric
	counters   map[string]*Counter
	histograms map[string]*Histogram
	gauges     map[string]*Gauge
	timers     map[string]*Timer
	logger     errors.Logger
}

// Metric represents a generic metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Description string                 `json:"description"`
	Unit        string                 `json:"unit"`
	Labels      map[string]interface{} `json:"labels"`
	LastUpdated time.Time              `json:"last_updated"`
}

// MetricType represents metric type
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
	MetricTypeTimer
)

// String returns string representation of metric type
func (mt MetricType) String() string {
	switch mt {
	case MetricTypeCounter:
		return "COUNTER"
	case MetricTypeGauge:
		return "GAUGE"
	case MetricTypeHistogram:
		return "HISTOGRAM"
	case MetricTypeTimer:
		return "TIMER"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(mt))
	}
}

// IsValid checks if metric type is valid
func (mt MetricType) IsValid() bool {
	return mt >= MetricTypeCounter && mt <= MetricTypeTimer
}

// Counter represents a counter metric
type Counter struct {
	Metric
	value uint64
	mutex sync.RWMutex
}

// NewCounter creates new counter
func NewCounter(name, description, unit string) *Counter {
	return &Counter{
		Metric: Metric{
			Name:        name,
			Type:        MetricTypeCounter,
			Description: description,
			Unit:        unit,
			Labels:      make(map[string]interface{}),
			LastUpdated: time.Now(),
		},
		value: 0,
	}
}

// Increment increments counter
func (c *Counter) Increment(delta uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += delta
	c.LastUpdated = time.Now()
}

// Get returns counter value
func (c *Counter) Get() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// Reset resets counter to zero
func (c *Counter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value = 0
	c.LastUpdated = time.Now()
}

// Gauge represents a gauge metric
type Gauge struct {
	Metric
	value float64
	mutex sync.RWMutex
}

// NewGauge creates new gauge
func NewGauge(name, description, unit string) *Gauge {
	return &Gauge{
		Metric: Metric{
			Name:        name,
			Type:        MetricTypeGauge,
			Description: description,
			Unit:        unit,
			Labels:      make(map[string]interface{}),
			LastUpdated: time.Now(),
		},
		value: 0.0,
	}
}

// Set sets gauge value
func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = value
	g.LastUpdated = time.Now()
}

// Get returns gauge value
func (g *Gauge) Get() float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.value
}

// Add adds value to gauge
func (g *Gauge) Add(delta float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value += delta
	g.LastUpdated = time.Now()
}

// Histogram represents a histogram metric
type Histogram struct {
	Metric
	buckets []float64
	counts  []uint64
	sum     float64
	count   uint64
	mutex   sync.RWMutex
}

// NewHistogram creates new histogram
func NewHistogram(name, description string, buckets []float64) *Histogram {
	return &Histogram{
		Metric: Metric{
			Name:        name,
			Type:        MetricTypeHistogram,
			Description: description,
			Unit:        "value",
			Labels:      make(map[string]interface{}),
			LastUpdated: time.Now(),
		},
		buckets: buckets,
		counts:  make([]uint64, len(buckets)+1),
	}
}

// Observe records a value in histogram
func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.sum += value
	h.count++
	h.LastUpdated = time.Now()

	// Find appropriate bucket
	bucketIndex := len(h.buckets)
	for i, bucket := range h.buckets {
		if value <= bucket {
			bucketIndex = i
			break
		}
	}

	h.counts[bucketIndex]++
}

// GetStats returns histogram statistics
func (h *Histogram) GetStats() HistogramStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := HistogramStats{
		Buckets: make([]BucketStats, len(h.buckets)),
		Count:   h.count,
		Sum:     h.sum,
	}

	if h.count > 0 {
		stats.Mean = h.sum / float64(h.count)
	}

	for i, bucket := range h.buckets {
		cumulativeCount := uint64(0)
		for j := 0; j <= i; j++ {
			cumulativeCount += h.counts[j]
		}

		stats.Buckets[i] = BucketStats{
			Bucket:            bucket,
			Count:             h.counts[i],
			CumulativeCount:   cumulativeCount,
			CumulativePercent: float64(cumulativeCount) / float64(h.count) * 100.0,
		}
	}

	return stats
}

// Reset resets histogram
func (h *Histogram) Reset() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.counts = make([]uint64, len(h.buckets)+1)
	h.sum = 0.0
	h.count = 0
	h.LastUpdated = time.Now()
}

// Timer represents a timer metric
type Timer struct {
	histogram *Histogram
}

// NewTimer creates new timer
func NewTimer(name, description string, buckets []float64) *Timer {
	return &Timer{
		histogram: NewHistogram(name, description, buckets),
	}
}

// Time measures execution time
func (t *Timer) Time(fn func() result.Result[any]) result.Result[any] {
	start := time.Now()
	result := fn()
	duration := time.Since(start).Seconds()

	t.histogram.Observe(duration)

	return result
}

// Record records a duration
func (t *Timer) Record(duration time.Duration) {
	t.histogram.Observe(duration.Seconds())
}

// GetStats returns timer statistics
func (t *Timer) GetStats() HistogramStats {
	return t.histogram.GetStats()
}

// Reset resets timer
func (t *Timer) Reset() {
	t.histogram.Reset()
}

// HistogramStats represents histogram statistics
type HistogramStats struct {
	Buckets []BucketStats `json:"buckets"`
	Count   uint64        `json:"count"`
	Sum     float64       `json:"sum"`
	Mean    float64       `json:"mean"`
}

// BucketStats represents bucket statistics
type BucketStats struct {
	Bucket            float64 `json:"bucket"`
	Count             uint64  `json:"count"`
	CumulativeCount   uint64  `json:"cumulative_count"`
	CumulativePercent float64 `json:"cumulative_percent"`
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	MemoryTotal    uint64    `json:"memory_total"`
	MemoryFree     uint64    `json:"memory_free"`
	GoroutineCount int       `json:"goroutine_count"`
	LastUpdated    time.Time `json:"last_updated"`
}

// NewMetricsCollector creates new metrics collector
func NewMetricsCollector(logger errors.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics:    make(map[string]*Metric),
		counters:   make(map[string]*Counter),
		histograms: make(map[string]*Histogram),
		gauges:     make(map[string]*Gauge),
		timers:     make(map[string]*Timer),
		logger:     logger,
	}
}

// RegisterCounter registers a counter metric
func (mc *MetricsCollector) RegisterCounter(name, description, unit string) result.Result[*Counter] {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.counters[name]; exists {
		return result.Err[*Counter](
			errors.NewErrorFactory("metrics").
				AlreadyExists(fmt.Sprintf("counter %s already exists", name)).
				ToResult())
	}

	counter := NewCounter(name, description, unit)
	mc.counters[name] = counter
	mc.metrics[name] = &counter.Metric

	mc.logger.Info("Counter registered",
		"name", name,
		"description", description,
		"unit", unit)

	return result.Ok(counter)
}

// RegisterGauge registers a gauge metric
func (mc *MetricsCollector) RegisterGauge(name, description, unit string) result.Result[*Gauge] {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.gauges[name]; exists {
		return result.Err[*Gauge](
			errors.NewErrorFactory("metrics").
				AlreadyExists(fmt.Sprintf("gauge %s already exists", name)).
				ToResult())
	}

	gauge := NewGauge(name, description, unit)
	mc.gauges[name] = gauge
	mc.metrics[name] = &gauge.Metric

	mc.logger.Info("Gauge registered",
		"name", name,
		"description", description,
		"unit", unit)

	return result.Ok(gauge)
}

// RegisterHistogram registers a histogram metric
func (mc *MetricsCollector) RegisterHistogram(name, description string, buckets []float64) result.Result[*Histogram] {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.histograms[name]; exists {
		return result.Err[*Histogram](
			errors.NewErrorFactory("metrics").
				AlreadyExists(fmt.Sprintf("histogram %s already exists", name)).
				ToResult())
	}

	histogram := NewHistogram(name, description, buckets)
	mc.histograms[name] = histogram
	mc.metrics[name] = &histogram.Metric

	mc.logger.Info("Histogram registered",
		"name", name,
		"description", description,
		"buckets", buckets)

	return result.Ok(histogram)
}

// RegisterTimer registers a timer metric
func (mc *MetricsCollector) RegisterTimer(name, description string, buckets []float64) result.Result[*Timer] {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if _, exists := mc.timers[name]; exists {
		return result.Err[*Timer](
			errors.NewErrorFactory("metrics").
				AlreadyExists(fmt.Sprintf("timer %s already exists", name)).
				ToResult())
	}

	timer := NewTimer(name, description, buckets)
	mc.timers[name] = timer
	mc.metrics[name] = &timer.histogram.Metric

	mc.logger.Info("Timer registered",
		"name", name,
		"description", description,
		"buckets", buckets)

	return result.Ok(timer)
}

// GetCounter returns a counter by name
func (mc *MetricsCollector) GetCounter(name string) result.Result[*Counter] {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if counter, exists := mc.counters[name]; exists {
		return result.Ok(counter)
	}

	return result.Err[*Counter](
		errors.NewErrorFactory("metrics").
			NotFound(fmt.Sprintf("counter %s not found", name)).
			ToResult())
}

// GetGauge returns a gauge by name
func (mc *MetricsCollector) GetGauge(name string) result.Result[*Gauge] {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if gauge, exists := mc.gauges[name]; exists {
		return result.Ok(gauge)
	}

	return result.Err[*Gauge](
		errors.NewErrorFactory("metrics").
			NotFound(fmt.Sprintf("gauge %s not found", name)).
			ToResult())
}

// GetHistogram returns a histogram by name
func (mc *MetricsCollector) GetHistogram(name string) result.Result[*Histogram] {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if histogram, exists := mc.histograms[name]; exists {
		return result.Ok(histogram)
	}

	return result.Err[*Histogram](
		errors.NewErrorFactory("metrics").
			NotFound(fmt.Sprintf("histogram %s not found", name)).
			ToResult())
}

// GetTimer returns a timer by name
func (mc *MetricsCollector) GetTimer(name string) result.Result[*Timer] {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if timer, exists := mc.timers[name]; exists {
		return result.Ok(timer)
	}

	return result.Err[*Timer](
		errors.NewErrorFactory("metrics").
			NotFound(fmt.Sprintf("timer %s not found", name)).
			ToResult())
}

// GetAllMetrics returns all registered metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	metrics := make(map[string]*Metric)
	for name, metric := range mc.metrics {
		metrics[name] = metric
	}

	return metrics
}

// GetSystemMetrics returns current system metrics
func (mc *MetricsCollector) GetSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemMetrics{
		MemoryUsage:    float64(m.Alloc) / float64(m.Sys) * 100.0,
		MemoryTotal:    m.Sys,
		MemoryFree:     m.Sys - m.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		LastUpdated:    time.Now(),
	}
}

// ExportMetrics exports metrics in Prometheus format
func (mc *MetricsCollector) ExportMetrics() string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	var output strings.Builder

	// Export counters
	for name, counter := range mc.counters {
		output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, counter.Description))
		output.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, counter.Type.String()))
		output.WriteString(fmt.Sprintf("%s %d\n\n", name, counter.Get()))
	}

	// Export gauges
	for name, gauge := range mc.gauges {
		output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, gauge.Description))
		output.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, gauge.Type.String()))
		output.WriteString(fmt.Sprintf("%s %f\n\n", name, gauge.Get()))
	}

	// Export histograms
	for name, histogram := range mc.histograms {
		stats := histogram.GetStats()

		output.WriteString(fmt.Sprintf("# HELP %s %s\n", name, histogram.Description))
		output.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, histogram.Type.String()))
		output.WriteString(fmt.Sprintf("%s_sum %f\n", name, stats.Sum))
		output.WriteString(fmt.Sprintf("%s_count %d\n", name, stats.Count))

		for _, bucket := range stats.Buckets {
			output.WriteString(fmt.Sprintf("%s_bucket{le=\"%f\"} %d\n", name, bucket.Bucket, bucket.CumulativeCount))
		}

		output.WriteString(fmt.Sprintf("%s_bucket{le=\"+Inf\"} %d\n\n", name, stats.Count))
	}

	return output.String()
}

// ResetAllMetrics resets all metrics
func (mc *MetricsCollector) ResetAllMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	for _, counter := range mc.counters {
		counter.Reset()
	}

	for _, histogram := range mc.histograms {
		histogram.Reset()
	}

	for _, timer := range mc.timers {
		timer.Reset()
	}

	mc.logger.Info("All metrics reset")
}

// InitializeDefaultMetrics initializes default metrics
func (mc *MetricsCollector) InitializeDefaultMetrics() result.Result[any] {
	// Operation metrics
	mc.RegisterCounter("operations_total", "Total number of operations", "operations")
	mc.RegisterCounter("operations_successful", "Number of successful operations", "operations")
	mc.RegisterCounter("operations_failed", "Number of failed operations", "operations")

	// Performance metrics
	mc.RegisterTimer("operation_duration", "Operation execution time", []float64{0.001, 0.01, 0.1, 1.0, 10.0})
	mc.RegisterHistogram("bytes_cleaned", "Bytes cleaned by operations", []float64{1024, 1048576, 1073741824, 10737418240})

	// System metrics
	mc.RegisterGauge("goroutines", "Number of goroutines", "goroutines")
	mc.RegisterGauge("memory_usage", "Memory usage percentage", "percent")
	mc.RegisterGauge("cpu_usage", "CPU usage percentage", "percent")

	mc.logger.Info("Default metrics initialized")

	return result.Ok[any](nil)
}

// PerformanceMonitor monitors system performance
type PerformanceMonitor struct {
	metricsCollector *MetricsCollector
	monitorInterval  time.Duration
	stopChan         chan struct{}
	logger           errors.Logger
}

// NewPerformanceMonitor creates new performance monitor
func NewPerformanceMonitor(
	collector *MetricsCollector,
	interval time.Duration,
	logger errors.Logger,
) *PerformanceMonitor {
	return &PerformanceMonitor{
		metricsCollector: collector,
		monitorInterval:  interval,
		stopChan:         make(chan struct{}),
		logger:           logger,
	}
}

// Start starts performance monitoring
func (pm *PerformanceMonitor) Start() {
	pm.logger.Info("Starting performance monitoring", "interval", pm.monitorInterval.String())

	ticker := time.NewTicker(pm.monitorInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				pm.collectMetrics()
			case <-pm.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.logger.Info("Stopping performance monitoring")
	close(pm.stopChan)
}

// collectMetrics collects system metrics
func (pm *PerformanceMonitor) collectMetrics() {
	systemMetrics := pm.metricsCollector.GetSystemMetrics()

	// Update gauges
	if goroutineGauge, err := pm.metricsCollector.GetGauge("goroutines"); err.IsOk() {
		goroutineGauge.Set(float64(systemMetrics.GoroutineCount))
	}

	if memoryGauge, err := pm.metricsCollector.GetGauge("memory_usage"); err.IsOk() {
		memoryGauge.Set(systemMetrics.MemoryUsage)
	}

	pm.logger.Debug("System metrics collected",
		"goroutines", systemMetrics.GoroutineCount,
		"memory_usage", systemMetrics.MemoryUsage)
}

// CreateDefaultBuckets creates default histogram buckets
func CreateDefaultBuckets() []float64 {
	return []float64{0.001, 0.01, 0.1, 1.0, 10.0, 30.0, 60.0, 300.0}
}

// CreateTimerBuckets creates timer-specific buckets
func CreateTimerBuckets() []float64 {
	return []float64{0.0001, 0.001, 0.01, 0.1, 1.0, 5.0, 10.0, 30.0}
}

// CreateSizeBuckets creates size-specific buckets
func CreateSizeBuckets() []float64 {
	return []float64{1024, 1048576, 1073741824, 10737418240, 107374182400}
}
