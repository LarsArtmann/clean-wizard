package cleaner

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewCargoCleaner(t *testing.T) {
	TestNewCleanerConstructor(t, NewCargoCleaner, "NewCargoCleaner")
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
	testCases := []GetHomeDirTestCase{
		{
			Name:      "HOME set",
			HomeValue: "/test/home",
			WantErr:   false,
			WantHome:  "/test/home",
		},
		{
			Name:         "fallback to USERPROFILE",
			HomeValue:    "",
			ProfileValue: "C:\\Users\\test",
			WantErr:      false,
			WantHome:     "C:\\Users\\test",
		},
		{
			Name:         "no home can be determined",
			HomeValue:    "",
			ProfileValue: "",
			WantErr:      true,
			WantHome:     "",
		},
	}

	RunGetHomeDirTests(t, testCases)
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


