package compiledbinaries

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	fileutil "github.com/LarsArtmann/clean-wizard/internal/shared/utils/fileutil"
)

// This file carries the discovery infrastructure of the compiledbinaries
// cleaner: duration parsing, the default directory scanner, and the trash
// operator. The cleaner type itself lives in compiledbinaries.go.

// parseAgeDuration parses a duration string like "7d", "30d", "1h", etc.
func parseAgeDuration(s string) (time.Duration, error) {
	if len(s) < MinAgeStringLength {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	var multiplier time.Duration

	unit := s[len(s)-1:]

	switch unit {
	case "d":
		multiplier = DurationUnitDays
	case "h":
		multiplier = time.Hour
	case "w":
		multiplier = DurationUnitWeeks
	case "m":
		multiplier = DurationUnitMonths
	case "y":
		multiplier = DurationUnitYears
	default:
		return 0, fmt.Errorf("unknown duration unit: %s for multiplier=%v", unit, multiplier)
	}

	var value int
	if _, err := fmt.Sscanf(s[:len(s)-1], "%d", &value); err != nil {
		return 0, fmt.Errorf(
			"invalid duration value: %s for value=%v, multiplier=%v",
			s,
			value,
			multiplier,
		)
	}

	return time.Duration(value) * multiplier, nil
}

// Default implementations

type defaultBinaryScanner struct {
	excludePatterns   []string
	includeCategories []BinaryCategory
	verbose           bool
}

func (s *defaultBinaryScanner) ScanDirectory(
	ctx context.Context,
	dir string,
	categories []BinaryCategory,
	minSize int64,
) ([]BinaryInfo, error) {
	var binaries []BinaryInfo

	categorySet := make(map[BinaryCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // Skip files we can't access
		}

		// Skip directories (but continue walking)
		if info.IsDir() {
			// Skip excluded directories
			if s.shouldSkipDirectory(path) {
				return filepath.SkipDir
			}

			return nil
		}

		// Check file size
		if info.Size() < minSize {
			return nil
		}

		// Check if it's an executable
		if info.Mode()&0o111 == 0 {
			return nil
		}

		// Determine category and check if included
		category := s.categorizeBinary(path, dir)
		if !categorySet[category] {
			return nil
		}

		// Check exclude patterns
		if s.isExcludedByPattern(path) {
			return nil
		}

		// Check default exclude binaries
		filename := filepath.Base(path)
		for _, exclude := range DefaultExcludeBinaries {
			if strings.Contains(strings.ToLower(filename), strings.ToLower(exclude)) {
				return nil
			}
		}

		binaries = append(binaries, BinaryInfo{
			Path:     path,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Category: category,
		})

		return nil
	})
	if err != nil {
		return binaries, fmt.Errorf("dir=%v, minSize=%v: %w", dir, minSize, err)
	}

	return binaries, nil
}

func (s *defaultBinaryScanner) shouldSkipDirectory(path string) bool {
	base := filepath.Base(path)

	// Skip excluded directories
	if slices.Contains(DefaultExcludeDirectories, base) {
		if s.verbose {
			fmt.Printf("Skipping excluded directory: %s\n", path)
		}

		return true
	}

	return false
}

func (s *defaultBinaryScanner) categorizeBinary(path, baseDir string) BinaryCategory {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return CategoryRoot
	}

	// Check for test binaries (*.test)
	if strings.HasSuffix(path, ".test") {
		return CategoryTest
	}

	// Split path to check directory structure
	parts := strings.Split(filepath.Dir(relPath), string(filepath.Separator))

	for _, part := range parts {
		switch part {
		case "tmp":
			return CategoryTmp
		case "bin":
			return CategoryBin
		case "dist":
			return CategoryDist
		}
	}

	// If in root of project (no subdirectory or only one level)
	if len(parts) <= 1 || (len(parts) == 1 && parts[0] == ".") {
		return CategoryRoot
	}

	return CategoryRoot
}

func (s *defaultBinaryScanner) isExcludedByPattern(path string) bool {
	for _, pattern := range s.excludePatterns {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue
		}

		if matched {
			return true
		}
		// Also try matching just the filename
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}

	return false
}

type defaultBinaryTrashOperator struct {
	verbose bool
}

func (t *defaultBinaryTrashOperator) TrashBinary(ctx context.Context, path string) error {
	return cleaner.TrashPath(ctx, path)
}

func (t *defaultBinaryTrashOperator) GetFileSize(path string) int64 {
	return fileutil.GetFileSize(path)
}

func (t *defaultBinaryTrashOperator) GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}
