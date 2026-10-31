package di

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// Benchmarks for the startup decision recorded in ADR-0004's appendix: the
// container registers lazy per-cleaner services; the registry materializes on
// first resolution. These numbers quantify registration cost vs full
// materialization cost so a future eager registry has a baseline to beat.

func BenchmarkRegisterAllServices(b *testing.B) {
	cfg := &types.Config{}
	settings := RunSettings{Verbose: false, DryRun: true}

	for b.Loop() {
		container, cleanup := New()
		if err := RegisterAllServices(container.Injector(), cfg, settings); err != nil {
			b.Fatal(err)
		}

		cleanup()
	}
}

func BenchmarkRegistryMaterialization(b *testing.B) {
	cfg := &types.Config{}
	settings := RunSettings{Verbose: false, DryRun: true}

	for b.Loop() {
		container, cleanup := New()
		if err := RegisterAllServices(container.Injector(), cfg, settings); err != nil {
			b.Fatal(err)
		}

		registry, err := CleanerRegistry(container.Injector())
		if err != nil {
			b.Fatal(err)
		}

		if len(registry.Names()) != 13 {
			b.Fatalf("registry size %d != %d", len(registry.Names()), 13)
		}

		cleanup()
	}
}
