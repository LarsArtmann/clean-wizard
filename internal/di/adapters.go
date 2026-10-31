package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/samber/do/v2"
)

// AdaptersPackage registers the external-system adapters that production
// cleaners consume as DI services. The consumer-facing interface alias
// (adapters.NixStore) is declared by RegisterAllServices; consumers resolve
// interfaces, keeping implementations swappable.
//
// CacheManager, HTTPClient, and RateLimiter deliberately have NO registration:
// no production code consumes them yet (ADR-0003). Registering services
// without consumers made the container describe an architecture the code does
// not have; they return the moment a real consumer exists.
var AdaptersPackage = do.Package( //nolint:gochecknoglobals
	registerAdapters,
)

// registerAdapters provides the concrete adapters; the consumer-facing
// interface alias is declared by RegisterAllServices so an invalid alias
// surfaces as a classified error instead of a panic (do.MustAs panics,
// do.As returns).
func registerAdapters(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*adapters.NixAdapter, error) {
		return adapters.NewNixAdapter(0, 0), nil
	})
}

// The nix adapter participates in container health checks (doctor command):
// the alias service forwards HealthCheck to the concrete adapter.
var _ do.HealthcheckerWithContext = (*adapters.NixAdapter)(nil)

// NixStore resolves the Nix store interface from the DI container.
func NixStore(i do.Injector) (adapters.NixStore, error) { //nolint:ireturn // returning the adapter interface is the DI contract: implementations stay swappable
	return do.Invoke[adapters.NixStore](i)
}
