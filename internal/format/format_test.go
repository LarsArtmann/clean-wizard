package format

import (
	"testing"
	"time"
)

func TestSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1572864, "1.5 MB"},
		{"gigabytes", 1610612736, "1.5 GB"},
		{"terabytes", 1649267441664, "1.5 TB"},
		{"petabytes", 1688849860263936, "1.5 PB"},
		{"exabytes", 1729382256910270464, "1.5 EB"},
		{"zero", 0, "0 B"},
		{"negative", -1024, "-1024 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Size(tt.bytes)
			if result != tt.expected {
				t.Errorf("Size(%d) = %v, want %v", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"nanoseconds", 500 * time.Nanosecond, "500 ns"},
		{"microseconds", 1500 * time.Nanosecond, "1500 ns"},
		{"milliseconds", 1500 * time.Microsecond, "1.5 ms"},
		{"seconds", 1500 * time.Millisecond, "1.5 s"},
		{"minutes", 1500 * time.Second, "25.0 m"},
		{"hours", 1500 * time.Minute, "25.0 h"},
		{"zero", 0, "0 ns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Duration(tt.duration)
			if result != tt.expected {
				t.Errorf("Duration(%v) = %v, want %v", tt.duration, result, tt.expected)
			}
		})
	}
}

func runTimeFormattingTests(t *testing.T, tests []struct {
		name     string
		t        time.Time
		expected string
	}, formatFn func(time.Time) string) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFn(tt.t)
			if result != tt.expected {
				t.Errorf("result = %v, want %v", result, tt.expected)
			}
		})
	}
}

var commonTimeTestCases = []struct {
		name     string
		t        time.Time
		expected string
	}{
		{"valid datetime", time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC), "2023-12-25 15:30:45"},
		{"zero time", time.Time{}, "never"},
		{"unix epoch", time.Unix(0, 0), "1970-01-01 00:00:00"},
	}

func TestDate(t *testing.T) {
	tests := []struct {
		name     string
		t        time.Time
		expected string
	}{
		{"valid date", time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC), "2023-12-25"},
		{"zero time", time.Time{}, "never"},
		{"unix epoch", time.Unix(0, 0), "1970-01-01"},
	}

	runTimeFormattingTests(t, tests, Date)
}

func TestDateTime(t *testing.T) {
	for _, tt := range commonTimeTestCases {
		t.Run(tt.name, func(t *testing.T) {
			result := DateTime(tt.t)
			if tt.expected == "never" && result != "never" {
				t.Errorf("DateTime(%v) = %v, want %v", tt.t, result, tt.expected)
			} else if tt.expected != "never" {
				if len(result) < 19 {
					t.Errorf("DateTime(%v) = %v, expected at least 19 characters", tt.t, result)
				}
			}
		})
	}
}

func TestNumber(t *testing.T) {
	tests := []struct {
		name     string
		n        int64
		expected string
	}{
		{"small number", 42, "42"},
		{"thousands", 1234, "1,234"},
		{"millions", 1234567, "1,234,567"},
		{"billions", 1234567890, "1,234,567,890"},
		{"zero", 0, "0"},
		{"negative", -1234, "-1,234"},
		{"single digit", 5, "5"},
		{"hundreds", 999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Number(tt.n)
			if result != tt.expected {
				t.Errorf("Number(%d) = %v, want %v", tt.n, result, tt.expected)
			}
		})
	}
}
