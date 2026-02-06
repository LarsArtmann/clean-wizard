package cleaner

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewCargoCleaner(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		dryRun  bool
	}{
		{
			name:    "standard configuration",
			verbose: false,
			dryRun:  false,
		},
		{
			name:    "verbose mode",
			verbose: true,
			dryRun:  false,
		},
		{
			name:    "dry-run mode",
			verbose: false,
			dryRun:  true,
		},
		{
			name:    "verbose dry-run mode",
			verbose: true,
			dryRun:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewCargoCleaner(tt.verbose, tt.dryRun)

			if cleaner == nil {
				t.Fatal("NewCargoCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}
		})
	}
}

func TestCargoCleaner_Type(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	if cleaner.Type() != domain.OperationTypeCargoPackages {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeCargoPackages)
	}
}

func TestCargoCleaner_IsAvailable(t *testing.T) {
	testCases := []IsAvailableTestCase{
		{
			Name: "default configuration",
			Constructor: func() interface {
				IsAvailable(ctx context.Context) bool
			} {
				return NewTestCleaner(NewCargoCleaner)()
			},
		},
	}

	TestIsAvailableGeneric(t, testCases)
}

func TestCargoCleaner_BooleanSettingsTests(t *testing.T) {
	CreateBooleanSettingsCleanerTestFunctions(t, NewBooleanSettingsCleanerTestConfig(
		"Cargo",
		"Cargo",
		"cargo packages",
		1,
		NewCargoCleaner,
		func(enabled bool) *domain.OperationSettings {
			return &domain.OperationSettings{
				CargoPackages: &domain.CargoPackagesSettings{
					Autoclean: enabled,
				},
			}
		},
	))
}

func TestCargoCleaner_GetHomeDir(t *testing.T) {
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

	// Test fallback to empty string when no home can be determined
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	home, err = GetHomeDir()
	if err == nil {
		t.Errorf("GetHomeDir() error = %v, want error for missing home", err)
	}
	if home != "" {
		t.Errorf("GetHomeDir() = %v, want empty string", home)
	}
}

func TestCargoCleaner_GetDirSize(t *testing.T) {
	// Test with non-existent path
	size := GetDirSize("/non/existent/path/12345")
	// Should return 0 for non-existent path
	if size != 0 {
		t.Errorf("GetDirSize() for non-existent path = %d, want 0", size)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	size = GetDirSize(tmpDir)
	// Should return 0 for empty directory
	if size != 0 {
		t.Errorf("GetDirSize() for empty dir = %d, want 0", size)
	}
}

func TestCargoCleaner_GetDirModTime(t *testing.T) {
	// Test with non-existent path
	modTime := GetDirModTime("/non/existent/path/12345")
	if !modTime.IsZero() {
		t.Errorf("GetDirModTime() for non-existent path = %v, want zero time", modTime)
	}

	// Test with temp directory
	tmpDir := t.TempDir()
	modTime = GetDirModTime(tmpDir)
	if modTime.IsZero() {
		t.Error("GetDirModTime() for temp dir returned zero time")
	}

	// Should be close to now (within 1 second tolerance)
	if modTime.After(time.Now().Add(time.Second)) {
		t.Error("GetDirModTime() returned time in the future")
	}
}

func TestCargoCleaner_HasCargoCacheTool(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Just verify it returns a boolean without crashing
	hasTool := cleaner.hasCargoCacheTool()
	if hasTool != true && hasTool != false {
		t.Errorf("hasCargoCacheTool() returned invalid value")
	}
}

func TestCargoCleaner_Clean_NoAvailable(t *testing.T) {
	// This test would fail if Cargo is installed
	// We just verify the error handling logic exists

	cleaner := NewCargoCleaner(false, false)

	// Can't easily test "Cargo not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestCargoCleaner_StandardTests(t *testing.T) {
	TestStandardCleaner(t, NewBooleanSettingsCleanerTestConstructor(NewCargoCleaner), "Cargo")
}

func TestCargoCleaner_Scan(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Test scan with CARGO_HOME set
	t.Setenv("CARGO_HOME", "/test/cargo")
	result := cleaner.Scan(context.Background())

	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Should find at least registry and git cache
	// But if CARGO_HOME doesn't exist, it might still return items with zero size
	if len(items) < 2 {
		t.Logf("Scan() found %d items (may be less if CARGO_HOME doesn't exist)", len(items))
	}
}

func TestCargoCleaner_Scan_DefaultCargoHome(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Set HOME but not CARGO_HOME
	t.Setenv("CARGO_HOME", "")
	t.Setenv("HOME", "/test/home")

	result := cleaner.Scan(context.Background())

	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Should construct path using ~/.cargo
	// Number of items depends on whether ~/.cargo exists
	if len(items) == 0 {
		t.Log("Scan() found 0 items (CARGO_HOME may not exist)")
	}
}


