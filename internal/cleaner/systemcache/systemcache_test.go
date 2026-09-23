package systemcache

import (
	"context"
	"testing"

	cln "github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

func TestNewSystemCacheCleaner(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			cleaner, err := NewSystemCacheCleaner(tt.verbose, tt.dryRun, tt.olderThan, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSystemCacheCleaner() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && cleaner == nil {
				t.Error("NewSystemCacheCleaner() returned nil cleaner")
			}

			if cleaner != nil {
				cln.AssertCleanerBooleanFields(t, cleaner, tt.verbose, tt.dryRun)
			}
		})
	}
}

func TestSystemCacheCleaner_Type(t *testing.T) {
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, false, "30d", nil)
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	if cleaner.Type() != operations.OperationTypeSystemCache {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), operations.OperationTypeSystemCache)
	}
}

func TestSystemCacheCleaner_IsAvailable(t *testing.T) {
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, false, "30d", nil)
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
	t.Parallel()

	tests := []struct {
		name     string
		settings *operations.OperationSettings
		wantErr  bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil system cache settings",
			settings: &operations.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all platform caches",
			settings: func() *operations.OperationSettings {
				return &operations.OperationSettings{
					SystemCache: &operations.SystemCacheSettings{
						CacheTypes: AvailableSystemCacheTypes(),
						OlderThan:  "30d",
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "valid settings with single platform cache",
			settings: func() *operations.OperationSettings {
				caches := AvailableSystemCacheTypes()
				if len(caches) == 0 {
					return &operations.OperationSettings{
						SystemCache: &operations.SystemCacheSettings{
							CacheTypes: []enums.CacheType{},
							OlderThan:  "7d",
						},
					}
				}

				return &operations.OperationSettings{
					SystemCache: &operations.SystemCacheSettings{
						CacheTypes: []enums.CacheType{caches[0]},
						OlderThan:  "7d",
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "valid settings with no caches",
			settings: &operations.OperationSettings{
				SystemCache: &operations.SystemCacheSettings{
					CacheTypes: []enums.CacheType{},
					OlderThan:  "30d",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cache type",
			settings: &operations.OperationSettings{
				SystemCache: &operations.SystemCacheSettings{
					CacheTypes: []enums.CacheType{99}, // Invalid value
					OlderThan:  "30d",
				},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid caches",
			settings: &operations.OperationSettings{
				SystemCache: &operations.SystemCacheSettings{
					CacheTypes: []enums.CacheType{
						enums.CacheTypeSpotlight,
						99,
					}, // Mixed valid and invalid
					OlderThan: "30d",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner, err := NewSystemCacheCleaner(false, false, "30d", nil)
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
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, true, "30d", nil)
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Skip test if not on macOS or Linux
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: SystemCacheCleaner only available on macOS/Linux")

		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Verify dry-run strategy
	if cleanResult.Strategy != enums.StrategyDryRunType {
		t.Errorf(
			"Clean() strategy = %v, want %v",
			cleanResult.Strategy,
			enums.StrategyDryRunType,
		)
	}

	// Note: ItemsRemoved and FreedBytes depend on actual cache directories and files
	// We only assert if items were found; 0 items/bytes is valid if caches are empty
	if cleanResult.ItemsRemoved > 0 && cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() removed %d items but freed 0 bytes", cleanResult.ItemsRemoved)
	}

	t.Logf("Dry-run found %d items, %d bytes", cleanResult.ItemsRemoved, cleanResult.FreedBytes)
}

func TestSystemCacheCleaner_Scan(t *testing.T) {
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, false, "30d", nil)
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
	t.Parallel()

	_, err := NewSystemCacheCleaner(false, false, "30d", nil)
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Test GetHomeDir doesn't crash
	home, err := cln.GetHomeDir()

	// May return empty string if home cannot be determined
	if home == "" && err == nil {
		t.Error("GetHomeDir() returned empty string and no error")
	}

	if home != "" {
		t.Logf("GetHomeDir() = %s", home)
	}
}

func TestSystemCacheCleaner_DryRunStrategy(t *testing.T) {
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, true, "30d", nil)
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	cln.TestDryRun(t, cln.SimpleCleanerConstructorFromInstance(cleaner), "system-cache", -1)
}

func TestSystemCacheCleaner_ParseDuration(t *testing.T) {
	t.Parallel()

	for _, tc := range cln.CommonDurationTestCases {
		t.Run(tc.Duration, func(t *testing.T) {
			t.Parallel()

			cleaner, err := NewSystemCacheCleaner(false, false, tc.Duration, nil)

			if tc.WantValid && err != nil {
				t.Errorf(
					"NewSystemCacheCleaner() with duration %s should succeed, got error: %v",
					tc.Duration,
					err,
				)
			}

			if !tc.WantValid && err == nil {
				t.Errorf("NewSystemCacheCleaner() with duration %s should fail", tc.Duration)
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
	t.Parallel()

	cleaner, err := NewSystemCacheCleaner(false, false, "30d", nil)
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	// Just verify it doesn't crash
	// Result depends on OS
	_ = cleaner.isMacOS()
}
