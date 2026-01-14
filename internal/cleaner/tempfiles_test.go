package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewTempFilesCleaner(t *testing.T) {
	tests := []struct {
		name      string
		olderThan string
		excludes  []string
		basePaths []string
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			olderThan: "24h",
			excludes:  []string{"/tmp/keep"},
			basePaths: []string{"/tmp"},
			wantErr:   false,
		},
		{
			name:      "invalid duration",
			olderThan: "invalid",
			excludes:  []string{},
			basePaths: []string{"/tmp"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewTempFilesCleaner(false, false, tt.olderThan, tt.excludes, tt.basePaths)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTempFilesCleaner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cleaner == nil {
				t.Error("NewTempFilesCleaner() returned nil cleaner")
			}
		})
	}
}

func TestTempFilesCleaner_IsAvailable(t *testing.T) {
	cleaner, err := NewTempFilesCleaner(false, false, "24h", []string{}, []string{"/tmp"})
	if err != nil {
		t.Fatalf("NewTempFilesCleaner() error = %v", err)
	}

	if !cleaner.IsAvailable(context.Background()) {
		t.Error("TempFilesCleaner should always be available")
	}
}

func TestTempFilesCleaner_Scan(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Create old file
	oldFile := filepath.Join(tmpDir, "old_file.txt")
	err := os.WriteFile(oldFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set modification time to 2 days ago
	oldTime := time.Now().Add(-48 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Create recent file
	recentFile := filepath.Join(tmpDir, "recent_file.txt")
	err = os.WriteFile(recentFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create cleaner with 24h cutoff
	cleaner, err := NewTempFilesCleaner(false, false, "24h", []string{}, []string{tmpDir})
	if err != nil {
		t.Fatalf("NewTempFilesCleaner() error = %v", err)
	}

	// Scan for files
	result := cleaner.Scan(context.Background())
	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Should find only the old file
	if len(items) != 1 {
		t.Errorf("Scan() found %d items, want 1", len(items))
	}

	if len(items) > 0 && items[0].Path != oldFile {
		t.Errorf("Scan() found %s, want %s", items[0].Path, oldFile)
	}
}

func TestTempFilesCleaner_Clean_DryRun(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Create old file
	oldFile := filepath.Join(tmpDir, "old_file.txt")
	err := os.WriteFile(oldFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set modification time to 2 days ago
	oldTime := time.Now().Add(-48 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Create cleaner with dry-run enabled
	cleaner, err := NewTempFilesCleaner(false, true, "24h", []string{}, []string{tmpDir})
	if err != nil {
		t.Fatalf("NewTempFilesCleaner() error = %v", err)
	}

	// Clean with dry-run
	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Check that dry-run reports items to clean
	if cleanResult.ItemsRemoved != 1 {
		t.Errorf("Clean() removed %d items, want 1", cleanResult.ItemsRemoved)
	}

	// File should still exist in dry-run mode
	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		t.Error("File was removed in dry-run mode")
	}
}

func TestTempFilesCleaner_Clean_Real(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Create old file
	oldFile := filepath.Join(tmpDir, "old_file.txt")
	err := os.WriteFile(oldFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set modification time to 2 days ago
	oldTime := time.Now().Add(-48 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Create cleaner with dry-run disabled
	cleaner, err := NewTempFilesCleaner(false, false, "24h", []string{}, []string{tmpDir})
	if err != nil {
		t.Fatalf("NewTempFilesCleaner() error = %v", err)
	}

	// Clean for real
	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Check that real cleaning removed items
	if cleanResult.ItemsRemoved != 1 {
		t.Errorf("Clean() removed %d items, want 1", cleanResult.ItemsRemoved)
	}

	// File should be removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("File was not removed in real cleaning mode")
	}
}

func TestTempFilesCleaner_isExcluded(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		excludes []string
		want     bool
	}{
		{
			name:     "exact match",
			path:     "/tmp/keep/file.txt",
			excludes: []string{"/tmp/keep"},
			want:     true,
		},
		{
			name:     "prefix match",
			path:     "/tmp/keep/subdir/file.txt",
			excludes: []string{"/tmp/keep"},
			want:     true,
		},
		{
			name:     "no match",
			path:     "/tmp/file.txt",
			excludes: []string{"/tmp/keep"},
			want:     false,
		},
		{
			name:     "empty excludes",
			path:     "/tmp/file.txt",
			excludes: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewTempFilesCleaner(false, false, "24h", tt.excludes, []string{"/tmp"})
			if err != nil {
				t.Fatalf("NewTempFilesCleaner() error = %v", err)
			}

			got := cleaner.isExcluded(tt.path)
			if got != tt.want {
				t.Errorf("isExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTempFilesCleaner_ValidateSettings(t *testing.T) {
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
			name:     "nil temp files settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid settings",
			settings: &domain.OperationSettings{
				TempFiles: &domain.TempFilesSettings{
					OlderThan: "7d",
					Excludes:  []string{"/tmp/keep"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing older_than",
			settings: &domain.OperationSettings{
				TempFiles: &domain.TempFilesSettings{
					OlderThan: "",
					Excludes:  []string{"/tmp/keep"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid older_than",
			settings: &domain.OperationSettings{
				TempFiles: &domain.TempFilesSettings{
					OlderThan: "invalid",
					Excludes:  []string{"/tmp/keep"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner, err := NewTempFilesCleaner(false, false, "24h", []string{}, []string{"/tmp"})
			if err != nil {
				t.Fatalf("NewTempFilesCleaner() error = %v", err)
			}

			err = cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
