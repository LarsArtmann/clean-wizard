package di

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/samber/do/v2"
)

// AdaptersPackage registers external-system adapters as DI services and aliases
// each concrete adapter to its consumer-facing interface via do.As. Consumers
// resolve interfaces (e.g. adapters.NixStore), keeping implementations swappable.
var AdaptersPackage = do.Package( //nolint:gochecknoglobals
	registerAdapters,
)

// registerAdapters provides the concrete adapters and declares their interface
// aliases. Aliasing happens eagerly at registration time (do.As validates the
// type relationship); the adapters themselves stay lazy singletons.
func registerAdapters(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*adapters.NixAdapter, error) {
		return adapters.NewNixAdapter(0, 0), nil
	})
	do.MustAs[*adapters.NixAdapter, adapters.NixStore](injector)

	do.Provide(injector, func(i do.Injector) (*adapters.CacheManager, error) {
		return adapters.NewCacheManager(adapters.DefaultCacheExpiry, time.Hour), nil
	})
	do.MustAs[*adapters.CacheManager, adapters.KeyValueCache](injector)

	do.Provide(injector, func(i do.Injector) (*adapters.HTTPClient, error) {
		return adapters.NewHTTPClient(), nil
	})
	do.MustAs[*adapters.HTTPClient, adapters.HTTPRequester](injector)

	do.Provide(injector, func(i do.Injector) (*adapters.RateLimiter, error) {
		return adapters.NewRateLimiter(0, 0), nil
	})
	do.MustAs[*adapters.RateLimiter, adapters.Limiter](injector)
}

// NixStore resolves the Nix store interface from the DI container.
func NixStore(i do.Injector) (adapters.NixStore, error) {
	return do.Invoke[adapters.NixStore](i)
}

// KeyValueCache resolves the cache interface from the DI container.
func KeyValueCache(i do.Injector) (adapters.KeyValueCache, error) {
	return do.Invoke[adapters.KeyValueCache](i)
}
