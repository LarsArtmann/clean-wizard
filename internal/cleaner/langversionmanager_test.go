package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewLanguageVersionManagerCleaner(t *testing.T) {
	tests := []struct {
		name         string
		verbose      bool
		dryRun       bool
		managerTypes []LangVersionManagerType
	}{
		{
			name:         "all managers",
			verbose:      false,
			dryRun:       false,
			managerTypes: AvailableLangVersionManagers(),
		},
		{
			name:         "single manager",
			verbose:      true,
			dryRun:       false,
			managerTypes: []LangVersionManagerType{LangVersionManagerNVM},
		},
		{
			name:         "dry-run mode",
			verbose:      false,
			dryRun:       true,
			managerTypes: []LangVersionManagerType{LangVersionManagerPYENV},
		},
		{
			name:         "empty managers (should use all)",
			verbose:      false,
			dryRun:       false,
			managerTypes: []LangVersionManagerType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewLanguageVersionManagerCleaner(tt.verbose, tt.dryRun, tt.managerTypes)

			if cleaner == nil {
				t.Fatal("NewLanguageVersionManagerCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}

			// If empty managers provided, should default to all available
			if len(tt.managerTypes) == 0 {
				if len(cleaner.managerTypes) != 3 {
					t.Errorf("managerTypes count = %d, want 3 (all available)", len(cleaner.managerTypes))
				}
			} else {
				if len(cleaner.managerTypes) != len(tt.managerTypes) {
					t.Errorf("managerTypes count = %d, want %d", len(cleaner.managerTypes), len(tt.managerTypes))
				}
			}
		})
	}
}

func TestLanguageVersionManagerCleaner_Type(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers())

	if cleaner.Type() != domain.OperationTypeLangVersionManager {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeLangVersionManager)
	}
}

func TestLanguageVersionManagerCleaner_IsAvailable(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers())

	available := cleaner.IsAvailable(context.Background())

	// Language version manager cleaner should always be available
	if !available {
		t.Error("IsAvailable() should always return true for LanguageVersionManagerCleaner")
	}
}

func TestLanguageVersionManagerCleaner_ValidateSettings(t *testing.T) {
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
			name:     "nil lang version manager settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all managers",
			settings: &domain.OperationSettings{
				LangVersionManager: &domain.LangVersionManagerSettings{
					ManagerTypes: []string{"nvm", "pyenv", "rbenv"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with single manager",
			settings: &domain.OperationSettings{
				LangVersionManager: &domain.LangVersionManagerSettings{
					ManagerTypes: []string{"nvm"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with no managers",
			settings: &domain.OperationSettings{
				LangVersionManager: &domain.LangVersionManagerSettings{
					ManagerTypes: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid manager type",
			settings: &domain.OperationSettings{
				LangVersionManager: &domain.LangVersionManagerSettings{
					ManagerTypes: []string{"invalid-manager"},
				},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid managers",
			settings: &domain.OperationSettings{
				LangVersionManager: &domain.LangVersionManagerSettings{
					ManagerTypes: []string{"nvm", "invalid-manager"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers())

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLanguageVersionManagerCleaner_Clean_DryRun(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, true, AvailableLangVersionManagers())

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Dry-run should report items for all manager types (3 managers)
	if cleanResult.ItemsRemoved != 3 {
		t.Errorf("Clean() removed %d items, want 3", cleanResult.ItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestLanguageVersionManagerCleaner_Scan(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers())

	result := cleaner.Scan(context.Background())

	// Scan may not find any items if version managers aren't installed
	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Items count depends on whether version managers are installed
	if len(items) == 0 {
		t.Log("Scan() found 0 items (version managers may not be installed)")
	}
}

func TestLanguageVersionManagerCleaner_GetHomeDir(t *testing.T) {
	_ = NewLanguageVersionManagerCleaner(false, false, AvailableLangVersionManagers())

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

func TestLanguageVersionManagerCleaner_DryRunStrategy(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, true, AvailableLangVersionManagers())

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Verify dry-run strategy is set
	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	// Verify no failures occurred
	if cleanResult.ItemsFailed != 0 {
		t.Errorf("Clean() failed %d items, want 0", cleanResult.ItemsFailed)
	}
}

func TestAvailableLangVersionManagers(t *testing.T) {
	expectedManagers := []LangVersionManagerType{
		LangVersionManagerNVM,
		LangVersionManagerPYENV,
		LangVersionManagerRBENV,
	}
	availableItemsTestHelper(t, expectedManagers, AvailableLangVersionManagers, "AvailableLangVersionManagers")
}

func TestLangVersionManagerType_String(t *testing.T) {
	tests := []struct {
		Item LangVersionManagerType
		Want string
	}{
		{LangVersionManagerNVM, "nvm"},
		{LangVersionManagerPYENV, "pyenv"},
		{LangVersionManagerRBENV, "rbenv"},
	}
	stringTypesTestHelper(t, tests, func(t LangVersionManagerType) string { return string(t) }, "string")
}

func TestLanguageVersionManagerCleaner_Verbose(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(true, false, AvailableLangVersionManagers())

	// Verify verbose flag is set
	if !cleaner.verbose {
		t.Error("verbose flag should be set")
	}
}

func TestLanguageVersionManagerCleaner_SingleManager(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(false, false, []LangVersionManagerType{LangVersionManagerNVM})

	if len(cleaner.managerTypes) != 1 {
		t.Errorf("managerTypes count = %d, want 1", len(cleaner.managerTypes))
	}

	if cleaner.managerTypes[0] != LangVersionManagerNVM {
		t.Errorf("managerTypes[0] = %v, want %v", cleaner.managerTypes[0], LangVersionManagerNVM)
	}
}

func TestLanguageVersionManagerCleaner_Clean_Verbose(t *testing.T) {
	cleaner := NewLanguageVersionManagerCleaner(true, false, AvailableLangVersionManagers())

	// Just verify verbose flag is set for cleaning
	if !cleaner.verbose {
		t.Error("verbose flag should be set")
	}

	// Run clean - it shouldn't crash
	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Logf("Clean() error (may be expected): %v", result.Error())
	}
}
