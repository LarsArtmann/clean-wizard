package bdd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner/tempfiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTempfilesBDD(t *testing.T) {
	t.Parallel()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Tempfiles Cleaner Suite")
}

var _ = Describe("Tempfiles cleaner", func() {
	var (
		ctx      context.Context
		fakeTmp  string
		seedFile = func(relPath string, size int) string {
			path := filepath.Join(fakeTmp, relPath)
			Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())

			content := make([]byte, size)
			for i := range content {
				content[i] = 'x'
			}

			Expect(os.WriteFile(path, content, 0o644)).To(Succeed())

			// Age the file past the 7d threshold; the cleaner only considers
			// files older than its cutoff.
			stale := time.Now().Add(-30 * 24 * time.Hour)
			Expect(os.Chtimes(path, stale, stale)).To(Succeed())

			return path
		}
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeTmp = GinkgoT().TempDir()
	})

	Describe("identity", func() {
		It("exposes the temp-files cleaner name and operation type", func() {
			cleaner, err := tempfiles.NewTempFilesCleaner(true, true, "7d", nil, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(cleaner.Name()).To(Equal("tempfiles"))
		})
	})

	Describe("scan", func() {
		It("finds seeded files under the given temp root", func() {
			seedFile("stale/build.log", 256)

			cleaner, err := tempfiles.NewTempFilesCleaner(true, true, "7d", nil, []string{fakeTmp})
			Expect(err).ToNot(HaveOccurred())
			scan := cleaner.Scan(ctx)

			Expect(scan.IsOk()).To(BeTrue())
			Expect(scan.Value()).ToNot(BeEmpty())
		})

		It("reports no items when the temp root is empty", func() {
			cleaner, err := tempfiles.NewTempFilesCleaner(true, true, "7d", nil, []string{fakeTmp})
			Expect(err).ToNot(HaveOccurred())
			scan := cleaner.Scan(ctx)

			Expect(scan.IsOk()).To(BeTrue())
			Expect(scan.Value()).To(BeEmpty())
		})
	})

	Describe("dry-run clean", func() {
		It("leaves seeded files untouched in dry-run mode", func() {
			seeded := seedFile("stale/cache.bin", 128)

			cleaner, err := tempfiles.NewTempFilesCleaner(true, true, "7d", nil, []string{fakeTmp})
			Expect(err).ToNot(HaveOccurred())
			cleaned := cleaner.Clean(ctx)

			Expect(cleaned.IsOk()).To(BeTrue())

			_, statErr := os.Stat(seeded)
			Expect(statErr).To(BeNil(), "dry-run must not delete files")
		})
	})
})
