package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/samber/do/v2"
)

// AdaptersPackage registers the external-system adapters that production
// cleaners consume as DI services and aliases each concrete adapter to its
// consumer-facing interface via do.As. Consumers resolve interfaces (e.g.
// adapters.NixStore), keeping implementations swappable.
//
// CacheManager, HTTPClient, and RateLimiter deliberately have NO registration:
// no production code consumes them yet (ADR-0003). Registering services
// without consumers made the container describe an architecture the code does
// not have; they return the moment a real consumer exists.
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
}

// The nix adapter participates in container health checks (doctor command):
// the alias service forwards HealthCheck to the concrete adapter.
var _ do.HealthcheckerWithContext = (*adapters.NixAdapter)(nil)

// NixStore resolves the Nix store interface from the DI container.
func NixStore(i do.Injector) (adapters.NixStore, error) {
	return do.Invoke[adapters.NixStore](i)
}
