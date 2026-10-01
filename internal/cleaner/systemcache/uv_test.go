package systemcache

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

func TestParseUvCacheDirOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:    "plain path with trailing newline",
			output:  "/home/user/.cache/uv\n",
			want:    "/home/user/.cache/uv",
			wantErr: false,
		},
		{
			name:    "path with surrounding blank lines",
			output:  "\n\n/data/uv-cache\n\n",
			want:    "/data/uv-cache",
			wantErr: false,
		},
		{
			name:    "custom UV_CACHE_DIR style path",
			output:  "/mnt/cache/uv\n",
			want:    "/mnt/cache/uv",
			wantErr: false,
		},
		{
			name:    "empty output",
			output:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace-only output",
			output:  "  \n \n",
			want:    "",
			wantErr: true,
		},
		{
			name:    "relative path",
			output:  "relative/uv\n",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseUvCacheDirOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUvCacheDirOutput() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("parseUvCacheDirOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUvBinaryAvailable_RespectsPath(t *testing.T) {
	// t.Setenv forbids t.Parallel
	binDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(binDir, uvBinary),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o755,
	); err != nil {
		t.Fatalf("failed to write fake uv binary: %v", err)
	}

	t.Setenv("PATH", binDir)

	if !uvBinaryAvailable() {
		t.Error("uvBinaryAvailable() = false with fake uv on PATH, want true")
	}

	t.Setenv("PATH", "")

	if uvBinaryAvailable() {
		t.Error("uvBinaryAvailable() = true with empty PATH, want false")
	}
}

// installFakeUvBinary puts a scripted `uv` on PATH: `uv cache dir` reports
// $FAKE_UV_CACHE_DIR, `uv cache clean` truncates the seeded entry (simulating
// entry removal with pure shell builtins; the script runs with a restricted
// PATH, so no external commands are available) and writes a marker file so
// tests can prove the command ran.
func installFakeUvBinary(t *testing.T, cacheDir, seedPath, markerPath string) {
	t.Helper()

	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"cache\" ] && [ \"$2\" = \"dir\" ]; then\n" +
		"\tprintf '%s\\n' \"$FAKE_UV_CACHE_DIR\"\n" +
		"\texit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"cache\" ] && [ \"$2\" = \"clean\" ]; then\n" +
		"\t: > \"$FAKE_UV_SEED_FILE\"\n" +
		"\t: > \"$FAKE_UV_CLEAN_MARKER\"\n" +
		"\texit 0\n" +
		"fi\n" +
		"exit 1\n"

	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, uvBinary), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake uv binary: %v", err)
	}

	t.Setenv("PATH", binDir)
	t.Setenv("FAKE_UV_CACHE_DIR", cacheDir)
	t.Setenv("FAKE_UV_SEED_FILE", seedPath)
	t.Setenv("FAKE_UV_CLEAN_MARKER", markerPath)
}

func seedFile(t *testing.T, path string, size int) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent dir for %s: %v", path, err)
	}

	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), size), 0o644); err != nil {
		t.Fatalf("failed to seed file %s: %v", path, err)
	}
}

func fileEmpty(t *testing.T, path string) bool {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}

	return info.Size() == 0
}

func newUvOnlyCleaner(t *testing.T) *SystemCacheCleaner {
	t.Helper()

	cleaner, err := NewSystemCacheCleaner(false, false, "30d", []enums.CacheType{enums.CacheTypeUv})
	if err != nil {
		t.Fatalf("NewSystemCacheCleaner() error = %v", err)
	}

	return cleaner
}

func TestCleanUvCache_RunsUvCacheCleanCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cacheDir := filepath.Join(t.TempDir(), "uv")
	seeded := filepath.Join(cacheDir, "archive-v0", "entry.bin")

	const seededSize = 128
	seedFile(t, seeded, seededSize)

	marker := filepath.Join(t.TempDir(), "cleaned.marker")
	installFakeUvBinary(t, cacheDir, seeded, marker)

	cleaner := newUvOnlyCleaner(t)

	cleanResult := cleaner.Clean(context.Background())
	if cleanResult.IsErr() {
		t.Fatalf("Clean() error = %v", cleanResult.Error())
	}

	result := cleanResult.Value()
	if result.ItemsRemoved != 1 {
		t.Errorf("Clean() ItemsRemoved = %d, want 1", result.ItemsRemoved)
	}

	if result.SizeEstimate.Status != enums.SizeEstimateStatusKnown || result.SizeEstimate.Known != seededSize {
		t.Errorf(
			"Clean() SizeEstimate = {Known: %d, Status: %v}, want {Known: %d, Status: %v}",
			result.SizeEstimate.Known,
			result.SizeEstimate.Status,
			seededSize,
			enums.SizeEstimateStatusKnown,
		)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("uv cache clean was not invoked (marker missing): %v", err)
	}

	if !fileEmpty(t, seeded) {
		t.Errorf("seeded cache entry %s not emptied after clean", seeded)
	}
}

func TestCleanUvCache_FallsBackToStaticPathWithoutBinary(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("PATH", "")

	staticDir := filepath.Join(fakeHome, ".cache", "uv")
	seedFile(t, filepath.Join(staticDir, "entry.bin"), 128)

	cleaner := newUvOnlyCleaner(t)

	cleanResult := cleaner.Clean(context.Background())
	if cleanResult.IsErr() {
		t.Fatalf("Clean() error = %v", cleanResult.Error())
	}

	if result := cleanResult.Value(); result.ItemsRemoved != 1 {
		t.Errorf("Clean() ItemsRemoved = %d, want 1", result.ItemsRemoved)
	}

	if _, err := os.Stat(staticDir); !os.IsNotExist(err) {
		t.Errorf("static cache dir %s still exists after fallback clean", staticDir)
	}
}

func TestScanUvCache_ReportsResolvedDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cacheDir := filepath.Join(t.TempDir(), "uv")
	seedFile(t, filepath.Join(cacheDir, "entry.bin"), 128)

	installFakeUvBinary(t, cacheDir, filepath.Join(cacheDir, "entry.bin"), filepath.Join(t.TempDir(), "cleaned.marker"))

	cleaner := newUvOnlyCleaner(t)

	scanResult := cleaner.Scan(context.Background())
	if scanResult.IsErr() {
		t.Fatalf("Scan() error = %v", scanResult.Error())
	}

	items := scanResult.Value()
	if len(items) != 1 {
		t.Fatalf("Scan() returned %d items, want 1", len(items))
	}

	if items[0].Path != cacheDir {
		t.Errorf("Scan() item path = %q, want resolved dir %q", items[0].Path, cacheDir)
	}

	if items[0].Size != 128 {
		t.Errorf("Scan() item size = %d, want 128", items[0].Size)
	}
}

func TestScanUvCache_FallsBackToStaticPathWithoutBinary(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("PATH", "")

	staticDir := filepath.Join(fakeHome, ".cache", "uv")
	seedFile(t, filepath.Join(staticDir, "entry.bin"), 128)

	cleaner := newUvOnlyCleaner(t)

	scanResult := cleaner.Scan(context.Background())
	if scanResult.IsErr() {
		t.Fatalf("Scan() error = %v", scanResult.Error())
	}

	items := scanResult.Value()
	if len(items) != 1 {
		t.Fatalf("Scan() returned %d items, want 1", len(items))
	}

	if items[0].Path != staticDir {
		t.Errorf("Scan() item path = %q, want static dir %q", items[0].Path, staticDir)
	}
}

// TestUvCommands_RespectUVCacheDirEnv runs the real uv binary against a
// relocated cache to prove env-var awareness end to end (issue #55).
func TestUvCommands_RespectUVCacheDirEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-uv integration test in short mode")
	}

	if !uvBinaryAvailable() {
		t.Skip("uv binary not installed")
	}

	cacheDir := filepath.Join(t.TempDir(), "uv-cache")
	t.Setenv("UV_CACHE_DIR", cacheDir)

	resolved, err := uvCacheDir(context.Background())
	if err != nil {
		t.Fatalf("uvCacheDir() error = %v", err)
	}

	if resolved != cacheDir {
		t.Errorf("uvCacheDir() = %q, want %q", resolved, cacheDir)
	}

	seeded := filepath.Join(cacheDir, "seeded", "entry.bin")
	seedFile(t, seeded, 128)

	if err := runUvCacheClean(context.Background()); err != nil {
		t.Fatalf("runUvCacheClean() error = %v", err)
	}

	if _, err := os.Stat(seeded); !os.IsNotExist(err) {
		t.Errorf("seeded cache entry %s still exists after real uv cache clean", seeded)
	}
}
