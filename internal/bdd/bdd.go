package bdd

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/interface/cli/bdd"
)

// RunBDDTests runs all BDD scenarios (re-exported from cli/bdd package)
func RunBDDTests(t *testing.T) {
	bdd.RunBDDTests(t)
}

// NewBDDContext creates a fresh BDD context for scenarios (re-exported from cli/bdd package)
func NewBDDContext() *bdd.BDDContext {
	return bdd.NewBDDContext()
}