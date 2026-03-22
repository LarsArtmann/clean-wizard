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

	// Type returns the operation type for this cleaner.
	// Used for categorization and filtering.
	Type() domain.OperationType

	// Clean executes the cleaning operation and returns the result.
	Clean(ctx context.Context) result.Result[domain.CleanResult]

	// IsAvailable checks if the cleaner can run on this system.
	IsAvailable(ctx context.Context) bool

	// Scan scans for cleanable items and returns them as scan items.
	// This is used for dry-run estimation and preview functionality.
	Scan(ctx context.Context) result.Result[[]domain.ScanItem]
}

// NixStoreSizer defines the interface for cleaners that can report store size.
type NixStoreSizer interface {
	// GetStoreSize returns the size of the Nix store in bytes.
	// Returns 0 if Nix is not available.
	GetStoreSize(ctx context.Context) int64
}
