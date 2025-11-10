package domain

import "context"

// Cleaner interface for all cleaning operations
type Cleaner interface {
	IsAvailable(ctx context.Context) bool
	GetStoreSize(ctx context.Context) int64
	ValidateSettings(settings map[string]any) error
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
	CleanOldPackages(ctx context.Context, settings map[string]any) CleanResult
}

// Scanner interface for all scanning operations
type Scanner interface {
	Scan(ctx context.Context, req ScanRequest) ScanResult
}
