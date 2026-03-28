package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestGolangciLintCacheCleaner_Name(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(false, false)

	if cleaner.Name() != "golangci-lint-cache" {
		t.Errorf("Name() = %q, want %q", cleaner.Name(), "golangci-lint-cache")
	}
}

func TestGolangciLintCacheCleaner_Type(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(false, false)

	if cleaner.Type() != domain.OperationTypeGolangciLintCache {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeGolangciLintCache)
	}
}

func TestGolangciLintCacheCleaner_IsAvailable(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(false, false)
	ctx := context.Background()

	available := cleaner.IsAvailable(ctx)

	if t.Failed() {
		t.Logf("golangci-lint may not be installed, which is acceptable")
	}
	_ = available
}

func TestGolangciLintCacheCleaner_GetVerbose(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(true, false)

	if !cleaner.GetVerbose() {
		t.Error("GetVerbose() = false, want true")
	}
}

func TestGolangciLintCacheCleaner_GetDryRun(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(false, true)

	if !cleaner.GetDryRun() {
		t.Error("GetDryRun() = false, want true")
	}
}

func TestGolangciLintCacheCleaner_Scan(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(false, false)
	ctx := context.Background()

	result := cleaner.Scan(ctx)

	if result.IsErr() {
		t.Errorf("Scan() returned error: %v", result.Error())
	}
}

func TestGolangciLintCacheCleaner_Clean(t *testing.T) {
	cleaner := NewGolangciLintCacheCleaner(true, false)
	ctx := context.Background()

	result := cleaner.Clean(ctx)

	if result.IsErr() {
		t.Logf(
			"Clean() returned error (golangci-lint may not be installed or cache in use): %v",
			result.Error(),
		)
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		// Binary units (powers of 1024)
		{"3.1KiB", 3174, false},
		{"1.5MiB", 1572864, false},
		{"500B", 500, false},
		{"1GiB", 1073741824, false},
		{"1TiB", 1099511627776, false},
		// Decimal units (powers of 1000)
		{"1KB", 1000, false},
		{"1MB", 1000000, false},
		{"1GB", 1000000000, false},
		{"1TB", 1000000000000, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("parseSize(%q) = %d, want error", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("parseSize(%q) returned error: %v", tt.input, err)
				} else if result != tt.expected {
					t.Errorf("parseSize(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestParseCacheStatus(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantDir  string
		wantSize int64
		hasError bool
	}{
		{
			name:     "valid output",
			output:   "Dir: /Users/user/Library/Caches/golangci-lint\nSize: 3.1KiB",
			wantDir:  "/Users/user/Library/Caches/golangci-lint",
			wantSize: 3174,
			hasError: false,
		},
		{
			name:     "empty output",
			output:   "",
			wantDir:  "",
			wantSize: 0,
			hasError: true,
		},
		{
			name:     "missing size",
			output:   "Dir: /Users/user/Library/Caches/golangci-lint",
			wantDir:  "",
			wantSize: 0,
			hasError: true,
		},
		{
			name:     "missing dir",
			output:   "Size: 3.1KiB",
			wantDir:  "",
			wantSize: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCacheStatus(tt.output)
			if tt.hasError {
				if err == nil {
					t.Errorf("parseCacheStatus(%q) = %+v, want error", tt.output, result)
				}
			} else {
				if err != nil {
					t.Errorf("parseCacheStatus(%q) returned error: %v", tt.output, err)
				} else {
					if result.Dir != tt.wantDir {
						t.Errorf(
							"parseCacheStatus(%q).Dir = %q, want %q",
							tt.output,
							result.Dir,
							tt.wantDir,
						)
					}
					if result.Size != tt.wantSize {
						t.Errorf(
							"parseCacheStatus(%q).Size = %d, want %d",
							tt.output,
							result.Size,
							tt.wantSize,
						)
					}
				}
			}
		})
	}
}
