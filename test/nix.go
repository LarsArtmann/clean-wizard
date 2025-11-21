package mocks

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// MockNixGenerations returns standardized mock Nix generation data for testing
func MockNixGenerations() []shared.NixGeneration {
	now := time.Now()
	return []shared.NixGeneration{
		{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: now.Add(-24 * time.Hour), Status: shared.SelectedStatusSelected},
		{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: now.Add(-48 * time.Hour), Status: shared.SelectedStatusNotSelected},
		{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: now.Add(-72 * time.Hour), Status: shared.SelectedStatusNotSelected},
		{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: now.Add(-96 * time.Hour), Status: shared.SelectedStatusNotSelected},
		{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: now.Add(-120 * time.Hour), Status: shared.SelectedStatusNotSelected},
	}
}

// MockNixGenerationsResult returns mock Nix generations as a Result type
func MockNixGenerationsResult() result.Result[[]shared.NixGeneration] {
	return result.Ok(MockNixGenerations())
}

// MockNixGenerationsResultWithMessage returns mock Nix generations with a custom message
func MockNixGenerationsResultWithMessage(message string) result.Result[[]shared.NixGeneration] {
	return result.MockSuccess(MockNixGenerations(), message)
}
