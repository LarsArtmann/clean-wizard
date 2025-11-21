package bdd_test

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/bdd"
)

// TestBDD_NixGenerationManagement runs BDD tests for Nix generation management
func TestBDD_NixGenerationManagement(t *testing.T) {
	// Run all BDD scenarios for Nix generation management
	bdd.RunBDDTests(t)
}

// BenchmarkBDD_FrameworkPerformance benchmarks BDD framework performance
func BenchmarkBDD_FrameworkPerformance(b *testing.B) {
	for b.Loop() {
		// Create and destroy BDD context repeatedly to measure performance
		ctx := bdd.NewBDDContext()

		// Simulate typical BDD workflow
		ctx.TheSystemHasConfiguration()
		ctx.WithNixProfile("test_profile")
		ctx.WithDryRunMode()

		// Cleanup simulated
		ctx.AfterScenario(nil, nil)
	}
}
