package di

import (
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/samber/do/v2"
)

// OverrideRegistry replaces the cleaner registry in the DI container with a test double.
// This enables tests to inject mock or pre-configured registries without modifying
// the production registration logic.
func OverrideRegistry(injector do.Injector, replacement *cleaner.Registry) {
	do.OverrideValue(injector, replacement)
}

// OverrideSettings replaces the run settings in the DI container with test values.
func OverrideSettings(injector do.Injector, settings RunSettings) {
	do.OverrideValue(injector, settings)
}
