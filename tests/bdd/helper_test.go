package bdd

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// mockGenerations creates mock Nix generations for testing in CI environments.
func mockGenerations(count int) []domain.NixGeneration {
	gens := make([]domain.NixGeneration, count)
	for i := range count {
		status := domain.GenerationStatusHistorical
		if i == 0 {
			status = domain.GenerationStatusCurrent
		}
		gens[i] = domain.NixGeneration{
			ID:      300 - i,
			Path:    fmt.Sprintf("/nix/var/nix/profiles/default-%d-link", 300-i),
			Date:    time.Now().Add(-time.Duration(i*24) * time.Hour),
			Current: status,
		}
	}
	return gens
}

// getGenerationsOrMock attempts to list real Nix generations, falling back to mocks on error.
func getGenerationsOrMock(ctx context.Context, nixCleaner *cleaner.NixCleaner, mockCount int) result.Result[[]domain.NixGeneration] {
	generations := nixCleaner.ListGenerations(ctx)
	if generations.IsErr() {
		return result.Ok(mockGenerations(mockCount))
	}
	return generations
}

// newTestContext creates a NixTestContext with sensible defaults for testing.
func newTestContext() *NixTestContext {
	return &NixTestContext{
		ctx:    context.Background(),
		output: nil,
		dryRun: true,
	}
}
