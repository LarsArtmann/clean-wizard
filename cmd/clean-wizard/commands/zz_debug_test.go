package commands

import (
	"os"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/config"
)

func TestSavedFileNow_Debug(t *testing.T) {
	path := t.TempDir() + "/cw.yaml"
	t.Setenv("CONFIG_PATH", path)

	if err := runProfileCreateCommand(nil, nil, "test-profile", "desc"); err != nil {
		t.Fatalf("create: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	t.Logf("file bytes=%d", len(raw))

	if _, err := config.LoadFromPath(path); err != nil {
		t.Logf("reload err: %v", err)
	} else {
		t.Logf("reload OK")
	}
}
