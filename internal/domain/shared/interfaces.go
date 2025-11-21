package shared

import (
	"context"
	
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// Cleaner interface for all cleaning operations with type-safe settings
type Cleaner interface {
	IsAvailable(ctx context.Context) bool
	GetStoreSize(ctx context.Context) int64
	ValidateSettings(settings *OperationSettings) error
	Cleanup(ctx context.Context, settings *OperationSettings) result.Result[CleanResult]
}

// Cleaner interface for generation-based cleaners (Nix)
type GenerationCleaner interface {
	Cleaner
	ListGenerations(ctx context.Context) []NixGeneration
	CleanOldGenerations(ctx context.Context, keepCount int) CleanResult
}

// Cleaner interface for package-based cleaners (Homebrew)
type PackageCleaner interface {
	Cleaner
	ListPackages(ctx context.Context) []string
	CleanOldPackages(ctx context.Context, settings *OperationSettings) CleanResult
}

// Scanner interface for all scanning operations
type Scanner interface {
	Scan(ctx context.Context, req ScanRequest) ScanResult
}
