package adapters

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// NixStore abstracts Nix store operations for the nix cleaner. Consumers depend
// on this interface, not on *NixAdapter, so implementations (real, mock, or
// remote) are interchangeable. The DI container aliases *NixAdapter to it.
type NixStore interface {
	// SetDryRun toggles dry-run behavior for destructive operations.
	SetDryRun(dryRun bool)
	// IsAvailable reports whether the nix binary can be executed.
	IsAvailable(ctx context.Context) bool
	// GetStoreSize reports the total size of the Nix store in bytes.
	GetStoreSize(ctx context.Context) result.Result[int64]
	// ListGenerations lists all Nix store generations.
	ListGenerations(ctx context.Context) result.Result[[]types.NixGeneration]
	// RemoveGeneration deletes a single generation.
	RemoveGeneration(ctx context.Context, genID types.NixGenerationID) result.Result[types.CleanResult]
	// CollectGarbage runs nix store garbage collection.
	CollectGarbage(ctx context.Context) result.Result[types.CleanResult]
}

// Compile-time proof that the concrete adapter satisfies the interface.
var _ NixStore = (*NixAdapter)(nil)

// HTTPRequester abstracts the HTTP verb methods of HTTPClient.
type HTTPRequester interface {
	Get(ctx context.Context, url string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, body any) (*HTTPResponse, error)
	Put(ctx context.Context, url string, body any) (*HTTPResponse, error)
	Delete(ctx context.Context, url string) (*HTTPResponse, error)
}

var _ HTTPRequester = (*HTTPClient)(nil)

// Limiter abstracts rate limiting decisions.
type Limiter interface {
	Wait(ctx context.Context) error
	Allow() bool
}

var _ Limiter = (*RateLimiter)(nil)

// KeyValueCache abstracts the in-process cache. Values expire lazily on read.
type KeyValueCache interface {
	Set(key string, value any, expiration time.Duration)
	Get(key string) (any, bool)
	GetWithExpiration(key string) (any, time.Time, bool)
	Delete(key string)
	Clear()
	ItemCount() int
	FlushExpired()
}

var _ KeyValueCache = (*CacheManager)(nil)
