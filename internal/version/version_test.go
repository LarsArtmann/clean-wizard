package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	// Version should never be empty
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should contain date-based format or git tag
	if !strings.Contains(info.Version, ".") && info.GitTag == "" {
		t.Errorf("Version should be date-based or git tag, got: %s", info.Version)
	}

	// Commit should be set
	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}

	// Date should be set
	if info.Date == "" {
		t.Error("Date should not be empty")
	}
}

func TestGenerateVersion(t *testing.T) {
	v := generateVersion()

	// Should be in format YYYY.MM.DD
	if len(v) != 10 {
		t.Errorf("Version should be 10 chars (YYYY.MM.DD), got: %s (len=%d)", v, len(v))
	}

	if !strings.Contains(v, ".") {
		t.Errorf("Version should contain dots, got: %s", v)
	}
}

func TestGetGitCommit(t *testing.T) {
	commit := getGitCommit()

	// Should return "unknown" or a valid commit hash
	if commit == "" {
		t.Error("Commit should not be empty")
	}

	if commit != "unknown" && len(commit) < 4 {
		t.Errorf("Commit hash too short: %s", commit)
	}
}

func TestIsGitDirty(t *testing.T) {
	// This test just verifies the function doesn't panic
	_ = isGitDirty()
}

func TestInfoString(t *testing.T) {
	info := Info{
		Version: "2026.02.16",
		Commit:  "abc1234",
		Date:    "2026-02-16",
		BuiltBy: "test",
	}

	s := info.String()

	if !strings.Contains(s, "2026.02.16") {
		t.Errorf("String should contain version, got: %s", s)
	}

	if !strings.Contains(s, "abc1234") {
		t.Errorf("String should contain commit, got: %s", s)
	}
}

func TestInfoShort(t *testing.T) {
	info := Info{
		Version: "2026.02.16-dirty",
	}

	if info.Short() != "2026.02.16-dirty" {
		t.Errorf("Short should return version, got: %s", info.Short())
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version() should not return empty string")
	}
}

func TestCommit(t *testing.T) {
	c := Commit()
	if c == "" {
		t.Error("Commit() should not return empty string")
	}
}

func TestGetWithDirtyRepo(t *testing.T) {
	// If repo is dirty, version should end with -dirty
	info := Get()

	// We have uncommitted changes, so it should be dirty
	// (unless committed before this test runs)
	if info.IsDirty {
		if !strings.Contains(info.Version, "-dirty") {
			t.Errorf("Dirty repo should have -dirty suffix, got: %s", info.Version)
		}
	}
}

func TestInfoStringWithoutCommit(t *testing.T) {
	info := Info{
		Version: "2026.02.16",
		Date:    "2026-02-16",
	}

	s := info.String()

	if !strings.Contains(s, "2026.02.16") {
		t.Errorf("String should contain version, got: %s", s)
	}
}
