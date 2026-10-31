package di

import (
	"context"
	"fmt"
	"sort"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner/factory"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
)

// HealthEntry is one service's health-check outcome. Error is nil when the
// service was constructed and passed its health check.
type HealthEntry struct {
	Service string
	Error   error
}

// Healthy reports whether the service passed construction and health check.
func (e HealthEntry) Healthy() bool { return e.Error == nil }

// HealthReport materializes every registered service and runs the container
// health check, returning one entry per service sorted by name.
//
// Materialization is mandatory: samber/do treats never-invoked lazy services
// as healthy (serviceLazy.healthcheck returns nil when !built), so a doctor
// run that skipped construction would report a container it never built.
// Construction failures are merged into the entry of the failing service.
func HealthReport(ctx context.Context, injector do.Injector) ([]HealthEntry, error) {
	buildFailures := materializeAllServices(injector)

	healthResults := injector.HealthCheckWithContext(ctx)

	entries := make([]HealthEntry, 0, len(healthResults))
	for serviceName, healthErr := range healthResults {
		merged := healthErr
		if merged == nil {
			merged = buildFailures[serviceName]
		}

		entries = append(entries, HealthEntry{Service: serviceName, Error: merged})
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Service < entries[j].Service })

	return entries, nil
}

// materializeAllServices invokes every service in the container so lazy
// singletons are constructed. Returned failures are keyed by the same service
// names HealthCheckWithContext reports, merging construction faults into the
// health report. Config/run-settings registration failures are a broken
// composition root and surface as Rejections from the accessors.
func materializeAllServices(injector do.Injector) map[string]error {
	failures := map[string]error{}

	record := func(serviceName string, err error) {
		if err != nil {
			failures[serviceName] = err
		}
	}

	_, configErr := Config(injector)
	record(do.NameOf[*types.Config](), configErr)

	_, settingsErr := Settings(injector)
	record(do.NameOf[RunSettings](), settingsErr)

	_, storeErr := NixStore(injector)
	record(do.NameOf[adapters.NixStore](), storeErr)

	registry, registryErr := CleanerRegistry(injector)
	record(do.NameOf[*cleaner.Registry](), registryErr)

	cleanerNames := cleanerRegistrationNames(registry, registryErr)
	for _, registryName := range cleanerNames {
		_, cleanerErr := Cleaner(injector, registryName)
		record(cleanerServiceName(registryName), cleanerErr)
	}

	return failures
}

// cleanerRegistrationNames lists the cleaners to materialize: the registry's
// canonical order when it built, otherwise the factory table (whose names the
// per-cleaner providers consume anyway).
func cleanerRegistrationNames(registry *cleaner.Registry, registryErr error) []string {
	if registryErr == nil && registry != nil {
		return registry.Names()
	}

	registrations := factory.CleanerRegistrations()
	names := make([]string, 0, len(registrations))

	for _, registration := range registrations {
		names = append(names, registration.Name)
	}

	return names
}

// HealthSummary is the aggregate a doctor surface renders.
type HealthSummary struct {
	Total     int
	Unhealthy int
}

// Summarize reduces a report to counts for exit-code decisions.
func Summarize(entries []HealthEntry) HealthSummary {
	summary := HealthSummary{Total: len(entries)}

	for _, entry := range entries {
		if !entry.Healthy() {
			summary.Unhealthy++
		}
	}

	return summary
}

// UnhealthyError builds the classified Infrastructure error a doctor command
// returns when any service failed (errorfamily.ExitCode → 69).
func UnhealthyError(summary HealthSummary) error {
	return fmt.Errorf(
		"%d of %d service(s) unhealthy: %w",
		summary.Unhealthy, summary.Total,
		errorfamily.NewInfrastructure(
			"doctor.unhealthy",
			"one or more DI services failed construction or health check",
		),
	)
}
