package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewNodePackageManagerCleaner(t *testing.T) {
	tests := []struct {
		name             string
		verbose          bool
		dryRun           bool
		packageManagers  []NodePackageManagerType
		wantErr          bool
		wantPackageCount int
	}{
		{
			name:             "valid configuration with all PMs",
			verbose:          false,
			dryRun:           false,
			packageManagers:  AvailableNodePackageManagers(),
			wantErr:          false,
			wantPackageCount: 4,
		},
		{
			name:             "valid configuration with single PM",
			verbose:          true,
			dryRun:           true,
			packageManagers:  []NodePackageManagerType{NodePackageManagerNPM},
			wantErr:          false,
			wantPackageCount: 1,
		},
		{
			name:             "valid configuration with no PMs",
			verbose:          false,
			dryRun:           false,
			packageManagers:  []NodePackageManagerType{},
			wantErr:          false,
			wantPackageCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewNodePackageManagerCleaner(tt.verbose, tt.dryRun, tt.packageManagers)

			if cleaner == nil {
				t.Fatal("NewNodePackageManagerCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}

			if len(cleaner.packageManagers) != tt.wantPackageCount {
				t.Errorf("packageManagers count = %d, want %d", len(cleaner.packageManagers), tt.wantPackageCount)
			}
		})
	}
}

func TestNodePackageManagerCleaner_Type(t *testing.T) {
	cleaner := NewNodePackageManagerCleaner(false, false, AvailableNodePackageManagers())

	if cleaner.Type() != domain.OperationTypeNodePackages {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeNodePackages)
	}
}

func TestNodePackageManagerCleaner_IsAvailable(t *testing.T) {
	tests := []struct {
		name              string
		packageManagers   []NodePackageManagerType
		shouldBeAvailable bool
	}{
		{
			name:              "all package managers",
			packageManagers:   AvailableNodePackageManagers(),
			shouldBeAvailable: true, // At least npm should be available
		},
		{
			name:              "empty package managers",
			packageManagers:   []NodePackageManagerType{},
			shouldBeAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewNodePackageManagerCleaner(false, false, tt.packageManagers)
			available := cleaner.IsAvailable(context.Background())

			if available != tt.shouldBeAvailable {
				t.Errorf("IsAvailable() = %v, want %v", available, tt.shouldBeAvailable)
			}
		})
	}
}

func TestNodePackageManagerCleaner_ValidateSettings(t *testing.T) {
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
			name:     "nil node packages settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with all PMs",
			settings: &domain.OperationSettings{
				NodePackages: &domain.NodePackagesSettings{
					PackageManagers: []string{"npm", "pnpm", "yarn", "bun"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with single PM",
			settings: &domain.OperationSettings{
				NodePackages: &domain.NodePackagesSettings{
					PackageManagers: []string{"npm"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings with no PMs",
			settings: &domain.OperationSettings{
				NodePackages: &domain.NodePackagesSettings{
					PackageManagers: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid package manager",
			settings: &domain.OperationSettings{
				NodePackages: &domain.NodePackagesSettings{
					PackageManagers: []string{"invalid-pm"},
				},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid PMs",
			settings: &domain.OperationSettings{
				NodePackages: &domain.NodePackagesSettings{
					PackageManagers: []string{"npm", "invalid-pm"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewNodePackageManagerCleaner(false, false, AvailableNodePackageManagers())

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodePackageManagerCleaner_Clean_DryRun(t *testing.T) {
	tests := []struct {
		name            string
		packageManagers []NodePackageManagerType
		wantItems       uint
		shouldTest      bool // Only test if PMs are available
	}{
		{
			name:            "dry-run with all PMs",
			packageManagers: AvailableNodePackageManagers(),
			wantItems:       4,
			shouldTest:      true, // Always test all PMs (at least one should be available)
		},
		{
			name:            "dry-run with single PM",
			packageManagers: []NodePackageManagerType{NodePackageManagerNPM},
			wantItems:       1,
			shouldTest:      false, // Skip if npm not installed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewNodePackageManagerCleaner(false, true, tt.packageManagers)

			// Skip test if no PMs are available
			if !cleaner.IsAvailable(context.Background()) {
				t.Skipf("Skipping test: no available package managers")
				return
			}

			result := cleaner.Clean(context.Background())
			if result.IsErr() {
				t.Fatalf("Clean() error = %v", result.Error())
			}

			cleanResult := result.Value()

			if cleanResult.ItemsRemoved != tt.wantItems {
				t.Errorf("Clean() removed %d items, want %d", cleanResult.ItemsRemoved, tt.wantItems)
			}

			if cleanResult.Strategy != domain.StrategyDryRun {
				t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
			}

			if cleanResult.FreedBytes == 0 {
				t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
			}
		})
	}
}

func TestNodePackageManagerCleaner_Clean_NoAvailableManagers(t *testing.T) {
	cleaner := NewNodePackageManagerCleaner(false, false, []NodePackageManagerType{})

	result := cleaner.Clean(context.Background())
	if !result.IsErr() {
		t.Error("Clean() should return error when no package managers are available")
	}
}

func TestNodePackageManagerCleaner_AvailableNodePackageManagers(t *testing.T) {
	pms := AvailableNodePackageManagers()

	if len(pms) != 4 {
		t.Errorf("AvailableNodePackageManagers() returned %d PMs, want 4", len(pms))
	}

	expectedPMs := []NodePackageManagerType{
		NodePackageManagerNPM,
		NodePackageManagerPNPM,
		NodePackageManagerYarn,
		NodePackageManagerBun,
	}

	for i, pm := range pms {
		if pm != expectedPMs[i] {
			t.Errorf("AvailableNodePackageManagers()[%d] = %v, want %v", i, pm, expectedPMs[i])
		}
	}
}

func TestNodePackageManagerType_String(t *testing.T) {
	tests := []struct {
		pm   NodePackageManagerType
		want string
	}{
		{NodePackageManagerNPM, "npm"},
		{NodePackageManagerPNPM, "pnpm"},
		{NodePackageManagerYarn, "yarn"},
		{NodePackageManagerBun, "bun"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := string(tt.pm)
			if got != tt.want {
				t.Errorf("string(%v) = %v, want %v", tt.pm, got, tt.want)
			}
		})
	}
}

func TestGetHomeDir(t *testing.T) {
	// This test verifies GetHomeDir doesn't crash
	// Actual behavior depends on environment variables

	// Set HOME explicitly
	t.Setenv("HOME", "/test/home")
	home, err := GetHomeDir()
	if err != nil {
		t.Errorf("GetHomeDir() error = %v", err)
	}
	if home != "/test/home" {
		t.Errorf("GetHomeDir() = %v, want /test/home", home)
	}

	// Test fallback on Windows (USERPROFILE)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "C:\\Users\\test")
	home, err = GetHomeDir()
	if err != nil {
		t.Errorf("GetHomeDir() error = %v", err)
	}
	if home != "C:\\Users\\test" {
		t.Errorf("GetHomeDir() = %v, want C:\\Users\\test", home)
	}

	// Test error case
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	_, err = GetHomeDir()
	if err == nil {
		t.Error("GetHomeDir() should return error when no home directory can be determined")
	}
}
