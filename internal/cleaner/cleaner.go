package cleaner

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Cleaner defines the interface for all cleaner implementations.
type Cleaner interface {
	// Name returns the unique identifier for this cleaner.
	// Used for result tracking and registry operations.
	Name() string

	// Clean executes the cleaning operation and returns the result.
	Clean(ctx context.Context) result.Result[domain.CleanResult]

	// IsAvailable checks if the cleaner can run on this system.
	IsAvailable(ctx context.Context) bool
}
