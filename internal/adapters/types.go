package adapters

// TODO: This file is temporary - remove once domain.NixGeneration is complete
// TODO: Migrate to domain package completely

import "time"

// NixGeneration represents a Nix generation (temporary)
// TODO: Move this to domain package
type NixGeneration struct {
	ID      int       `json:"id"`
	Path    string    `json:"path"`
	Date    time.Time `json:"date"`
	Current bool      `json:"current"`
}