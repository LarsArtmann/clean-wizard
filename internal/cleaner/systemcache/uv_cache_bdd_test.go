package systemcache_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner/systemcache"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

// setEnvForSpec sets an environment variable for the current spec and
// restores the previous state afterwards (t.Setenv is unavailable inside
// Ginkgo specs).
func setEnvForSpec(key, value string) {
	previous, had := os.LookupEnv(key)
	Expect(os.Setenv(key, value)).To(Succeed())
	DeferCleanup(func() {
		if had {
			Expect(os.Setenv(key, previous)).To(Succeed())

			return
		}

		Expect(os.Unsetenv(key)).To(Succeed())
	})
}

// installFakeUvForSpec puts a scripted `uv` on PATH: `uv cache dir` reports
// cacheDir, `uv cache clean` empties cacheDir and writes markerPath. It is
// the spec-level twin of the helper in the internal test package.
func installFakeUvForSpec(cacheDir, markerPath string) {
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"cache\" ] && [ \"$2\" = \"dir\" ]; then\n" +
		"\tprintf '%s\\n' \"$FAKE_UV_CACHE_DIR\"\n" +
		"\texit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"cache\" ] && [ \"$2\" = \"clean\" ]; then\n" +
		"\tfind \"$FAKE_UV_CACHE_DIR\" -mindepth 1 -delete\n" +
		"\t: > \"$FAKE_UV_CLEAN_MARKER\"\n" +
		"\texit 0\n" +
		"fi\n" +
		"exit 1\n"

	binDir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(binDir, "uv"), []byte(script), 0o755)).To(Succeed())

	setEnvForSpec("PATH", binDir)
	setEnvForSpec("FAKE_UV_CACHE_DIR", cacheDir)
	setEnvForSpec("FAKE_UV_CLEAN_MARKER", markerPath)
}

func seedFileForSpec(path string, size int) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
	Expect(os.WriteFile(path, make([]byte, size), 0o644)).To(Succeed())
}

func newUvOnlyCleanerForSpec() *systemcache.SystemCacheCleaner {
	cleaner, err := systemcache.NewSystemCacheCleaner(false, false, "30d", []enums.CacheType{enums.CacheTypeUv})
	Expect(err).NotTo(HaveOccurred())

	return cleaner
}

var _ = Describe("uv cache cleaning", func() {
	var (
		ctx       context.Context
		cacheDir  string
		seededGen string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cacheDir = filepath.Join(GinkgoT().TempDir(), "uv")
		seededGen = filepath.Join(cacheDir, "archive-v0", "entry.bin")
	})

	Context("when the uv binary is installed", func() {
		BeforeEach(func() {
			setEnvForSpec("HOME", GinkgoT().TempDir())
			seedFileForSpec(seededGen, 128)
			installFakeUvForSpec(cacheDir, filepath.Join(GinkgoT().TempDir(), "cleaned.marker"))
		})

		It("scans the cache directory the binary resolves, not the static default", func() {
			cleaner := newUvOnlyCleanerForSpec()

			scanResult := cleaner.Scan(ctx)
			Expect(scanResult.IsOk()).To(BeTrue())

			items := scanResult.Value()
			Expect(items).To(HaveLen(1))
			Expect(items[0].Path).To(Equal(cacheDir))
			Expect(items[0].Size).To(Equal(int64(128)))
		})

		It("cleans through uv cache clean instead of deleting a hardcoded path", func() {
			cleaner := newUvOnlyCleanerForSpec()

			cleanResult := cleaner.Clean(ctx)
			Expect(cleanResult.IsOk()).To(BeTrue())
			Expect(cleanResult.Value().ItemsRemoved).To(Equal(uint(1)))

			Expect(seededGen).NotTo(BeAnExistingFile())
		})

		It("reports the freed bytes as a known size estimate", func() {
			cleaner := newUvOnlyCleanerForSpec()

			cleanResult := cleaner.Clean(ctx)
			Expect(cleanResult.IsOk()).To(BeTrue())

			estimate := cleanResult.Value().SizeEstimate
			Expect(estimate.Status).To(Equal(enums.SizeEstimateStatusKnown))
			Expect(estimate.Known).To(Equal(uint64(128)))
		})
	})

	Context("when the uv binary is missing", func() {
		var staticDir string

		BeforeEach(func() {
			home := GinkgoT().TempDir()
			setEnvForSpec("HOME", home)
			setEnvForSpec("PATH", "")

			staticDir = filepath.Join(home, ".cache", "uv")
			seedFileForSpec(filepath.Join(staticDir, "entry.bin"), 128)
		})

		It("scans the static default cache path", func() {
			cleaner := newUvOnlyCleanerForSpec()

			scanResult := cleaner.Scan(ctx)
			Expect(scanResult.IsOk()).To(BeTrue())

			items := scanResult.Value()
			Expect(items).To(HaveLen(1))
			Expect(items[0].Path).To(Equal(staticDir))
		})

		It("falls back to removing the static default cache path", func() {
			cleaner := newUvOnlyCleanerForSpec()

			cleanResult := cleaner.Clean(ctx)
			Expect(cleanResult.IsOk()).To(BeTrue())
			Expect(cleanResult.Value().ItemsRemoved).To(Equal(uint(1)))

			Expect(staticDir).NotTo(BeAnExistingFile())
		})
	})
})
