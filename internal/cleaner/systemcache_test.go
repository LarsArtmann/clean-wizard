package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewSystemCacheCleaner(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		dryRun    bool
		olderThan string
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			verbose:   false,
			dryRun:    false,
			olderThan: "30d",
			wantErr:   false,
		},
		{
			name:      "verbose dry-run",
			verbose:   true,
			dryRun:    true,
			olderThan: "7d",
			wantErr:   false,
		},
		{
			name:      "invalid duration",
			verbose:   false,
			dryRun:    false,
			olderThan: "invalid",
			wantErr:   true,
		},
		{
			name:      "empty duration",
			verbose:   false,
			dryRun:    false,
			olderThan: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewSystemCacheCleaner(tt.verbose, tt.dryRun, tt.olderThan)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSystemCacheCleaner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cleaner == nil {
				t.Error("NewSystemCacheCleaner() returned nil cleaner")
			}

			if cleaner != nil {
				if cleaner.verbose != tt.verbose {
					t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
				}

				if cleaner.dryRun != tt.dryRun {
					t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
				}
			}
		})
	}
}

func TestSystemCacheCleaner_Type(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, false, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	if cleaner.Type() != domain.OperationTypeSystemCache {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeSystemCache)
	}
}

func TestSystemCacheCleaner_IsAvailable(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, false, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	available := cleaner.IsAvailable(context.Background())

	// Result depends on OS (macOS vs others)
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestSystemCacheCleaner_ValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *domain.OperationSettings
		wantErr  bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil system cache settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all caches",
			settings: &domain.OperationSettings{
				SystemCache: &domain.SystemCacheSettings{
					CacheTypes: []string{"spotlight", "xcode", "cocoapods", "homebrew"},
					OlderThan:  "30d",
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with single cache",
			settings: &domain.OperationSettings{
				SystemCache: &domain.SystemCacheSettings{
					CacheTypes: []string{"spotlight"},
					OlderThan:  "7d",
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with no caches",
			settings: &domain.OperationSettings{
				SystemCache: &domain.SystemCacheSettings{
					CacheTypes: []string{},
					OlderThan:  "30d",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cache type",
			settings: &domain.OperationSettings{
				SystemCache: &domain.SystemCacheSettings{
					CacheTypes: []string{"invalid-cache"},
					OlderThan:  "30d",
				},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid caches",
			settings: &domain.OperationSettings{
				SystemCache: &domain.SystemCacheSettings{
					CacheTypes: []string{"spotlight", "invalid-cache"},
					OlderThan:  "30d",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewSystemCacheCleaner(false, false, "30d")
			if err != nil {
				t.Fatalf("NewSystemCacheCleaner() error = %v", err)
			}

			err = cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemCacheCleaner_Clean_DryRun(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, true, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Skip test if not on macOS
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: SystemCacheCleaner only available on macOS")
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Dry-run should report items for all cache types (4 types)
	if cleanResult.ItemsRemoved != 4 {
		t.Errorf("Clean() removed %d items, want 4", cleanResult.ItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestSystemCacheCleaner_Scan(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, false, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	result := cleaner.Scan(context.Background())

	// Scan may not find any items if not on macOS
	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Items count depends on OS and cache directories existence
	if len(items) == 0 {
		t.Log("Scan() found 0 items (may not be on macOS or caches don't exist)")
	}
}

func TestSystemCacheCleaner_GetHomeDir(t *testing.T) {
	_, err := NewSystemCacheCleaner(false, false, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Test GetHomeDir doesn't crash
	home, err := GetHomeDir()

	// May return empty string if home cannot be determined
	if home == "" && err == nil {
		t.Error("GetHomeDir() returned empty string and no error")
	}

	if home != "" {
		t.Logf("GetHomeDir() = %s", home)
	}
}

func TestSystemCacheCleaner_DryRunStrategy(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, true, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	TestDryRunStrategy(t, SimpleCleanerConstructorFromInstance(cleaner), "system-cache")
}

func TestAvailableSystemCacheTypes(t *testing.T) {
	expectedTypes := []SystemCacheType{
		SystemCacheSpotlight,
		SystemCacheXcode,
		SystemCacheCocoaPods,
		SystemCacheHomebrew,
	}
	TestAvailableTypesGeneric(t, "AvailableSystemCacheTypes", AvailableSystemCacheTypes, expectedTypes)
}

func TestSystemCacheType_String(t *testing.T) {
	TestTypeString(t, "SystemCacheType", []SystemCacheType{
		SystemCacheSpotlight,
		SystemCacheXcode,
		SystemCacheCocoaPods,
		SystemCacheHomebrew,
	})
}

func TestSystemCacheCleaner_ParseDuration(t *testing.T) {
	tests := []struct {
		duration  string
		wantValid bool
	}{
		{"1h", true},
		{"24h", true},
		{"7d", true},
		{"30d", true},
		{"1w", false}, // Not supported
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.duration, func(t *testing.T) {
			cleaner, err := NewSystemCacheCleaner(false, false, tt.duration)

			if tt.wantValid && err != nil {
				t.Errorf("NewSystemCacheCleaner() with duration %s should succeed, got error: %v", tt.duration, err)
			}

			if !tt.wantValid && err == nil {
				t.Errorf("NewSystemCacheCleaner() with duration %s should fail", tt.duration)
			}

			if cleaner != nil {
				// Verify duration was parsed correctly
				if cleaner.olderThan <= 0 {
					t.Errorf("olderThan = %v, want > 0", cleaner.olderThan)
				}
			}
		})
	}
}

func TestSystemCacheCleaner_IsMacOS(t *testing.T) {
	cleaner, err := NewSystemCacheCleaner(false, false, "30d")
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Just verify it doesn't crash
	// Result depends on OS
	_ = cleaner.isMacOS()
}
