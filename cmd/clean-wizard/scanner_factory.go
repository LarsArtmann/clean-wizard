package main

import (
	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// createScanner creates a scanner instance (legacy, will be deprecated)
func createScanner(verbose bool) types.Scanner {
	return scan.NewScanner(verbose)
}
