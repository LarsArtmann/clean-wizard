package testdata

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// MockNixGenerations provides consistent mock data across the codebase
var MockNixGenerations = []domain.NixGeneration{
	{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: time.Now().Add(-24 * time.Hour), Current: true},
	{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: time.Now().Add(-48 * time.Hour), Current: false},
	{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: time.Now().Add(-72 * time.Hour), Current: false},
	{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: time.Now().Add(-96 * time.Hour), Current: false},
	{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: time.Now().Add(-120 * time.Hour), Current: false},
}