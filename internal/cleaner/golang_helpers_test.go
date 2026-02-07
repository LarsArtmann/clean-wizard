package cleaner

import (
	"os"
	"testing"
)

func TestGolangHelpers_getHomeDir(t *testing.T) {
	helper := &golangHelpers{}

	// Set HOME explicitly
	t.Setenv("HOME", "/test/home")
	home := helper.getHomeDir()
	if home != "/test/home" {
		t.Errorf("getHomeDir() = %v, want /test/home", home)
	}

	// Test fallback on Windows (USERPROFILE)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "C:\\Users\\test")
	home = helper.getHomeDir()
	if home != "C:\\Users\\test" {
		t.Errorf("getHomeDir() = %v, want C:\\Users\\test", home)
	}

	// Test fallback to empty string
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	home = helper.getHomeDir()
	if home != "" {
		t.Errorf("getHomeDir() = %v, want empty string", home)
	}
}

func TestGolangHelpers_getDirSize(t *testing.T) {
	testDir := t.TempDir()

	// Should return 0 for empty directory
	size := GetDirSize(testDir)
	if size != 0 {
		t.Errorf("GetDirSize() on empty dir = %v, want 0", size)
	}

	// Create test files
	os.WriteFile(testDir+"/file1.txt", []byte("12345"), 0o644)
	os.WriteFile(testDir+"/file2.txt", []byte("67890"), 0o644)

	// Create a subdirectory with files
	subDir := testDir + "/subdir"
	os.Mkdir(subDir, 0o755)
	os.WriteFile(subDir+"/file3.txt", []byte("abcde"), 0o644)

	size = GetDirSize(testDir)
	expectedSize := int64(5 + 5 + 5) // 5+5+5 = 15 bytes
	if size != expectedSize {
		t.Errorf("GetDirSize() = %v, want %v", size, expectedSize)
	}
}

func TestGolangHelpers_getDirModTime(t *testing.T) {
	testDir := t.TempDir()

	// Create a test file
	testFile := testDir + "/test.txt"
	os.WriteFile(testFile, []byte("test"), 0o644)

	// Should return non-zero time
	modTime := GetDirModTime(testDir)
	if modTime.IsZero() {
		t.Error("GetDirModTime() returned zero time")
	}
}
