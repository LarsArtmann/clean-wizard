package format

import (
	"testing"
	"time"
)

var zeroTime = time.Time{}
var testDate = time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)

var commonTimeTests = []struct {
	name     string
	input    time.Time
	expected string
}{
	{"zero time", zeroTime, "never"},
	{"unix epoch", time.Unix(0, 0), ""},
}

func runFormattingTests[T any](t *testing.T, tests []struct {
	name     string
	input    T
	expected string
}, formatFn func(T) string,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFn(tt.input)
			if result != tt.expected {
				t.Errorf("result = %v, want %v", result, tt.expected)
			}
		})
	}
}

type dateTimeTestStruct struct {
	name     string //nolint:unused // needed for type compatibility with anonymous struct
	input    time.Time
	expected string
}

func runDateTimeTests(t *testing.T, tests []struct {
	name     string
	input    time.Time
	expected string
}, formatFn func(time.Time) string, customCheck func(t *testing.T, result string, tt dateTimeTestStruct),
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFn(tt.input)
			customCheck(t, result, tt)
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"bytes", int64(512), "512 B"},
		{"kilobytes", int64(1536), "1.5 KiB"},
		{"megabytes", int64(1572864), "1.5 MiB"},
		{"gigabytes", int64(1610612736), "1.5 GiB"},
		{"terabytes", int64(1649267441664), "1.5 TiB"},
		{"petabytes", int64(1688849860263936), "1.5 PiB"},
		{"exabytes", int64(1729382256910270464), "1.5 EiB"},
		{"zero", int64(0), "0 B"},
		{"negative", int64(-1024), "0 B"},
	}

	runFormattingTests(t, tests, Size)
}

func TestDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"nanoseconds", 500 * time.Nanosecond, "500 ns"},
		{"microseconds", 1500 * time.Nanosecond, "1500 ns"},
		{"milliseconds", 1500 * time.Microsecond, "1.5 ms"},
		{"seconds", 1500 * time.Millisecond, "1.5 s"},
		{"minutes", 1500 * time.Second, "25.0 m"},
		{"hours", 1500 * time.Minute, "25.0 h"},
		{"zero", time.Duration(0), "0 ns"},
	}

	runFormattingTests(t, tests, Duration)
}

func TestDate(t *testing.T) {
	dateTests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"valid date", testDate, "2023-12-25"},
	}
	dateTests = append(dateTests, commonTimeTests...)
	for i := range dateTests {
		if dateTests[i].name == "unix epoch" {
			dateTests[i].expected = "1970-01-01"
		}
	}

	runFormattingTests(t, dateTests, Date)
}

func TestDateTime(t *testing.T) {
	dateTimeTests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"valid datetime", testDate, "2023-12-25 15:30:45"},
	}
	dateTimeTests = append(dateTimeTests, commonTimeTests...)
	for i := range dateTimeTests {
		if dateTimeTests[i].name == "unix epoch" {
			dateTimeTests[i].expected = "1970-01-01 00:00:00"
		}
	}

	runDateTimeTests(
		t,
		dateTimeTests,
		DateTime,
		func(t *testing.T, result string, tt dateTimeTestStruct) {
			t.Helper()

			if tt.expected == "never" && result != "never" {
				t.Errorf("DateTime(%v) = %v, want %v", tt.input, result, tt.expected)
			} else if tt.expected != "never" {
				if len(result) < 19 {
					t.Errorf("DateTime(%v) = %v, expected at least 19 characters", tt.input, result)
				}
			}
		},
	)
}

func TestNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"small number", int64(42), "42"},
		{"thousands", int64(1234), "1,234"},
		{"millions", int64(1234567), "1,234,567"},
		{"billions", int64(1234567890), "1,234,567,890"},
		{"zero", int64(0), "0"},
		{"negative", int64(-1234), "-1,234"},
		{"single digit", int64(5), "5"},
		{"hundreds", int64(999), "999"},
	}

	runFormattingTests(t, tests, Number)
}
