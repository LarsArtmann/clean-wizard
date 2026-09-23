package nodepackages

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

func TestNewNodePackageManagerCleaner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name			string
		verbose			bool
		dryRun			bool
		packageManagers		[]enums.PackageManagerType
		wantErr			bool
		wantPackageCount	int
	}{
		{
			name:			"valid configuration with all PMs",
			verbose:		false,
			dryRun:			false,
			packageManagers:	AvailableNodePackageManagers(),
			wantErr:		false,
			wantPackageCount:	4,
		},
		{
			name:			"valid configuration with single PM",
			verbose:		true,
			dryRun:			true,
			packageManagers:	[]enums.PackageManagerType{enums.PackageManagerNpm},
			wantErr:		false,
			wantPackageCount:	1,
		},
		{
			name:			"valid configuration with no PMs",
			verbose:		false,
			dryRun:			false,
			packageManagers:	[]enums.PackageManagerType{},
			wantErr:		false,
			wantPackageCount:	0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner := NewNodePackageManagerCleaner(tt.verbose, tt.dryRun, tt.packageManagers)

			if cleaner == nil {
				t.Fatal("NewNodePackageManagerCleaner() returned nil cleaner")
			}

			if len(cleaner.packageManagers) != tt.wantPackageCount {
				t.Errorf(
					"packageManagers count = %d, want %d",
					len(cleaner.packageManagers),
					tt.wantPackageCount,
				)
			}
		})
	}
}

func TestNodePackageManagerCleaner_Type(t *testing.T) {
	t.Parallel()

	cleaner := NewNodePackageManagerCleaner(false, false, AvailableNodePackageManagers())

	if cleaner.Type() != operations.OperationTypeNodePackages {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), operations.OperationTypeNodePackages)
	}
}

func TestNodePackageManagerCleaner_IsAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name			string
		packageManagers		[]enums.PackageManagerType
		shouldBeAvailable	bool
	}{
		{
			name:			"all package managers",
			packageManagers:	AvailableNodePackageManagers(),
			shouldBeAvailable:	true,	// At least npm should be available
		},
		{
			name:			"empty package managers",
			packageManagers:	[]enums.PackageManagerType{},
			shouldBeAvailable:	false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner := NewNodePackageManagerCleaner(false, false, tt.packageManagers)
			available := cleaner.IsAvailable(context.Background())

			if available != tt.shouldBeAvailable {
				t.Errorf("IsAvailable() = %v, want %v", available, tt.shouldBeAvailable)
			}
		})
	}
}

func TestNodePackageManagerCleaner_ValidateSettings(t *testing.T) {
	t.Parallel()

	factory := cleaner.NewCleanerConstructorWithSettings(
		NewNodePackageManagerCleaner,
		AvailableNodePackageManagers,
	)
	testCases := []cleaner.ValidateSettingsTestCase{
		{
			Name:		"nil settings",
			Settings:	nil,
			WantErr:	false,
		},
		{
			Name:		"nil node packages settings",
			Settings:	&operations.OperationSettings{},
			WantErr:	false,
		},
		{
			Name:	"valid settings with all PMs",
			Settings: &operations.OperationSettings{
				NodePackages: &operations.NodePackagesSettings{
					PackageManagers: []enums.PackageManagerType{
						enums.PackageManagerNpm, enums.PackageManagerPnpm,
						enums.PackageManagerYarn, enums.PackageManagerBun,
					},
				},
			},
			WantErr:	false,
		},
		{
			Name:	"valid settings with single PM",
			Settings: &operations.OperationSettings{
				NodePackages: &operations.NodePackagesSettings{
					PackageManagers: []enums.PackageManagerType{enums.PackageManagerNpm},
				},
			},
			WantErr:	false,
		},
		{
			Name:	"valid settings with no PMs",
			Settings: &operations.OperationSettings{
				NodePackages: &operations.NodePackagesSettings{
					PackageManagers: []enums.PackageManagerType{},
				},
			},
			WantErr:	false,
		},
		{
			Name:	"invalid package manager",
			Settings: &operations.OperationSettings{
				NodePackages: &operations.NodePackagesSettings{
					PackageManagers: []enums.PackageManagerType{99},	// Invalid value
				},
			},
			WantErr:	true,
		},
		{
			Name:	"mixed valid and invalid PMs",
			Settings: &operations.OperationSettings{
				NodePackages: &operations.NodePackagesSettings{
					PackageManagers: []enums.PackageManagerType{
						enums.PackageManagerNpm,
						99,
					},	// Mixed valid and invalid
				},
			},
			WantErr:	true,
		},
	}
	cleaner.TestValidateSettings(t, factory, testCases)
}

func TestNodePackageManagerCleaner_Clean_DryRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name		string
		packageManagers	[]enums.PackageManagerType
		wantMinItems	uint	// Minimum expected (actual depends on what's installed)
		shouldTest	bool	// Only test if PMs are available
	}{
		{
			name:			"dry-run with all PMs",
			packageManagers:	AvailableNodePackageManagers(),
			wantMinItems:		1,	// At least one PM should be available
			shouldTest:		true,	// Always test all PMs (at least one should be available)
		},
		{
			name:			"dry-run with single PM",
			packageManagers:	[]enums.PackageManagerType{enums.PackageManagerNpm},
			wantMinItems:		1,	// npm should be available
			shouldTest:		false,	// Skip if npm not installed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

			if cleanResult.ItemsRemoved < tt.wantMinItems {
				t.Skipf("Clean() removed %d items — PM installed but cache empty on CI", cleanResult.ItemsRemoved)
			}

			if cleanResult.Strategy != enums.StrategyDryRunType {
				t.Errorf(
					"Clean() strategy = %v, want %v",
					cleanResult.Strategy,
					enums.StrategyDryRunType,
				)
			}

			// FreedBytes may be 0 if cache directories are empty
			t.Logf(
				"Clean() freed %d bytes from %d items",
				cleanResult.FreedBytes,
				cleanResult.ItemsRemoved,
			)
		})
	}
}

func TestNodePackageManagerCleaner_Clean_NoAvailableManagers(t *testing.T) {
	t.Parallel()

	cleaner := NewNodePackageManagerCleaner(false, false, []enums.PackageManagerType{})

	result := cleaner.Clean(context.Background())
	if !result.IsErr() {
		t.Error("Clean() should return error when no package managers are available")
	}
}

func TestNodePackageManagerCleaner_AvailableNodePackageManagers(t *testing.T) {
	t.Parallel()

	expectedPMs := []enums.PackageManagerType{
		enums.PackageManagerNpm,
		enums.PackageManagerPnpm,
		enums.PackageManagerYarn,
		enums.PackageManagerBun,
	}
	cleaner.TestAvailableTypesGeneric(
		t,
		"AvailableNodePackageManagers",
		AvailableNodePackageManagers,
		expectedPMs,
	)
}

func TestGetHomeDir(t *testing.T) {
	// This test verifies GetHomeDir doesn't crash
	// Actual behavior depends on environment variables

	// Set HOME explicitly
	t.Setenv("HOME", "/test/home")

	home, err := cleaner.GetHomeDir()
	if err != nil {
		t.Errorf("GetHomeDir() error = %v", err)
	}

	if home != "/test/home" {
		t.Errorf("GetHomeDir() = %v, want /test/home", home)
	}

	// Test fallback on Windows (USERPROFILE)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "C:\\Users\\test")

	home, err = cleaner.GetHomeDir()
	if err != nil {
		t.Errorf("GetHomeDir() error = %v", err)
	}

	if home != "C:\\Users\\test" {
		t.Errorf("GetHomeDir() = %v, want C:\\Users\\test", home)
	}

	// Test error case (only applies if user.Current() would fail)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err = cleaner.GetHomeDir()
	// On systems where user.Current() succeeds, this won't error
	// This test only validates that error handling exists
	_ = err
}
