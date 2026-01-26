package commands

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
)

// isGoCleanerAvailable checks if Go cleaner is available with default cache settings.
func isGoCleanerAvailable(ctx context.Context) bool {
	goCleaner, err := cleaner.NewGoCleaner(false, false, cleaner.GoCacheGOCACHE|cleaner.GoCacheTestCache|cleaner.GoCacheBuildCache)
	if err != nil {
		return false
	}
	return goCleaner.IsAvailable(ctx)
}