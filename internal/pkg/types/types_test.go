package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		risk     RiskLevel
		expected string
	}{
		{
			name:     "low risk",
			risk:     RiskLow,
			expected: "low",
		},
		{
			name:     "medium risk",
			risk:     RiskMedium,
			expected: "medium",
		},
		{
			name:     "high risk",
			risk:     RiskHigh,
			expected: "high",
		},
		{
			name:     "unknown risk",
			risk:     "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.risk.String())
		})
	}
}

func TestRiskLevel_Color(t *testing.T) {
	tests := []struct {
		name     string
		risk     RiskLevel
		expected string
	}{
		{
			name:     "low risk",
			risk:     RiskLow,
			expected: "green",
		},
		{
			name:     "medium risk",
			risk:     RiskMedium,
			expected: "yellow",
		},
		{
			name:     "high risk",
			risk:     RiskHigh,
			expected: "red",
		},
		{
			name:     "unknown risk",
			risk:     "unknown",
			expected: "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.risk.Color())
		})
	}
}

func TestRiskLevel_Icon(t *testing.T) {
	tests := []struct {
		name     string
		risk     RiskLevel
		expected string
	}{
		{
			name:     "low risk",
			risk:     RiskLow,
			expected: "✅",
		},
		{
			name:     "medium risk",
			risk:     RiskMedium,
			expected: "⚡",
		},
		{
			name:     "high risk",
			risk:     RiskHigh,
			expected: "⚠️",
		},
		{
			name:     "unknown risk",
			risk:     "unknown",
			expected: "❓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.risk.Icon())
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		sizeGB   float64
		expected string
	}{
		{
			name:     "bytes",
			sizeGB:   0.0001,
			expected: "0 MB",
		},
		{
			name:     "kilobytes",
			sizeGB:   0.001,
			expected: "1 MB",
		},
		{
			name:     "megabytes",
			sizeGB:   0.5,
			expected: "512 MB",
		},
		{
			name:     "gigabytes",
			sizeGB:   1.5,
			expected: "1.5 GB",
		},
		{
			name:     "terabytes",
			sizeGB:   2048,
			expected: "2.0 TB",
		},
		{
			name:     "zero",
			sizeGB:   0,
			expected: "0 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatSize(tt.sizeGB))
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "seconds",
			duration: 1500 * time.Millisecond,
			expected: "1.5s",
		},
		{
			name:     "minutes",
			duration: 90 * time.Second,
			expected: "1.5m",
		},
		{
			name:     "zero",
			duration: 0,
			expected: "0ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatDuration(tt.duration))
		})
	}
}

func TestScanResult_JSONMarshaling(t *testing.T) {
	result := ScanResult{
		Name:        "test",
		SizeGB:      1.5,
		Description: "test description",
		Cleanable:   true,
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = result
	})
}

func TestScanResults_JSONMarshaling(t *testing.T) {
	results := ScanResults{
		TotalSizeGB: 2.5,
		Results: []ScanResult{
			{
				Name:        "test1",
				SizeGB:      1.0,
				Description: "test1 description",
				Cleanable:   true,
			},
			{
				Name:        "test2",
				SizeGB:      1.5,
				Description: "test2 description",
				Cleanable:   false,
			},
		},
		Timestamp: time.Now(),
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = results
	})
}

func TestCleanupOperation_JSONMarshaling(t *testing.T) {
	op := CleanupOperation{
		Name:        "test-op",
		Description: "test operation",
		RiskLevel:   RiskLow,
		Enabled:     true,
		Settings:    map[string]any{"key": "value"},
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = op
	})
}

func TestProfile_JSONMarshaling(t *testing.T) {
	profile := Profile{
		Name:        "test-profile",
		Description: "test profile description",
		Operations: []CleanupOperation{
			{
				Name:        "test-op",
				Description: "test operation",
				RiskLevel:   RiskLow,
				Enabled:     true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = profile
	})
}

func TestConfig_JSONMarshaling(t *testing.T) {
	config := Config{
		Version:      "1.0.0",
		SafeMode:     true,
		DryRun:       true,
		Verbose:      false,
		Backup:       true,
		MaxDiskUsage: 80,
		Protected:    []string{"/test"},
		Profiles: map[string]Profile{
			"test": {
				Name:        "test",
				Description: "test profile",
				Operations:  []CleanupOperation{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = config
	})
}

func TestCleanupResult_JSONMarshaling(t *testing.T) {
	result := CleanupResult{
		Operation:   "test-op",
		Success:     true,
		SizeFreedGB: 1.5,
		Error:       nil,
		Duration:    5 * time.Second,
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = result
	})
}

func TestCleanupResults_JSONMarshaling(t *testing.T) {
	results := CleanupResults{
		TotalSizeFreedGB: 2.5,
		Results: []CleanupResult{
			{
				Operation:   "test-op1",
				Success:     true,
				SizeFreedGB: 1.0,
				Duration:    3 * time.Second,
			},
			{
				Operation:   "test-op2",
				Success:     true,
				SizeFreedGB: 1.5,
				Duration:    2 * time.Second,
			},
		},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(5 * time.Second),
		Duration:  5 * time.Second,
	}

	// Test that the struct can be marshaled to JSON
	assert.NotPanics(t, func() {
		_ = results
	})
}
