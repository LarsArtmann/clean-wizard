package domain

import "context"

// OperationHandler interface for all cleaning operations with type-safe settings.
type OperationHandler interface {
	Type() OperationType
	IsAvailable(ctx context.Context) bool
	GetStoreSize(ctx context.Context) int64
	ValidateSettings(settings *OperationSettings) error
}

// OperationHandler interface for generation-based cleaners (Nix).
type GenerationCleaner interface {
	OperationHandler
	ListGenerations(ctx context.Context) []NixGeneration
	CleanOldGenerations(ctx context.Context, keepCount int) CleanResult
}

// OperationHandler interface for package-based cleaners (Homebrew).
type PackageCleaner interface {
	OperationHandler
	ListPackages(ctx context.Context) []string
	CleanOldPackages(ctx context.Context, settings *OperationSettings) CleanResult
}

// Scanner interface for all scanning operations.
type Scanner interface {
	Scan(ctx context.Context, req ScanRequest) ScanResult
}
