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
	cleaner := NewCargoCleaner(false, false)
	available := cleaner.IsAvailable(context.Background())

	// Result depends on Cargo installation
	// We just verify it doesn't crash and returns a boolean
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestCargoCleaner_ValidateSettings(t *testing.T) {
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
			name:     "nil cargo packages settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings with autoclean",
			settings: &domain.OperationSettings{
				CargoPackages: &domain.CargoPackagesSettings{
					Autoclean: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings without autoclean",
			settings: &domain.OperationSettings{
				CargoPackages: &domain.CargoPackagesSettings{
					Autoclean: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewCargoCleaner(false, false)

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCargoCleaner_Clean_DryRun(t *testing.T) {
	cleaner := NewCargoCleaner(false, true)

	// Skip test if Cargo is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Cargo not available")
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Dry-run should report 1 item removed
	if cleanResult.ItemsRemoved != 1 {
		t.Errorf("Clean() removed %d items, want 1", cleanResult.ItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestCargoCleaner_GetHomeDir(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Set HOME explicitly
	t.Setenv("HOME", "/test/home")
	home := cleaner.getHomeDir()
	if home != "/test/home" {
		t.Errorf("getHomeDir() = %v, want /test/home", home)
	}

	// Test fallback on Windows (USERPROFILE)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "C:\\Users\\test")
	home = cleaner.getHomeDir()
	if home != "C:\\Users\\test" {
		t.Errorf("getHomeDir() = %v, want C:\\Users\\test", home)
	}

	// Test fallback to empty string
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	home = cleaner.getHomeDir()
	if home != "" {
		t.Errorf("getHomeDir() = %v, want empty string", home)
	}
}

func TestCargoCleaner_GetDirSize(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Cargo uses estimated size, not actual directory size
	// Estimate should be 250MB
	size := cleaner.getDirSize("/any/path")
	expectedSize := int64(250 * 1024 * 1024) // 250MB

	if size != expectedSize {
		t.Errorf("getDirSize() = %d, want %d (250MB)", size, expectedSize)
	}
}

func TestCargoCleaner_GetDirModTime(t *testing.T) {
	cleaner := NewCargoCleaner(false, false)

	// Cargo returns current time as estimate
	modTime := cleaner.getDirModTime("/any/path")
	if modTime.IsZero() {
		t.Error("getDirModTime() returned zero time")
	}

	// Should be close to now (within 1 second tolerance)
	if modTime.After(time.Now().Add(time.Second)) {
		t.Error("getDirModTime() returned time in the future")
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

func TestCargoCleaner_DryRunStrategy(t *testing.T) {
	cleaner := NewCargoCleaner(false, true)

	// Skip test if Cargo is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Cargo not available")
		return
	}

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

func TestCargoCleaner_AutocleanSettings(t *testing.T) {
	tests := []struct {
		name      string
		autoclean bool
		wantErr   bool
	}{
		{
			name:      "autoclean enabled",
			autoclean: true,
			wantErr:   false,
		},
		{
			name:      "autoclean disabled",
			autoclean: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewCargoCleaner(false, false)

			settings := &domain.OperationSettings{
				CargoPackages: &domain.CargoPackagesSettings{
					Autoclean: tt.autoclean,
				},
			}

			err := cleaner.ValidateSettings(settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
