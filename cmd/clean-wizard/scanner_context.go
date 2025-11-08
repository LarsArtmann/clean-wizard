package main

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// scannerKey is the context key for scanner injection
type scannerKey struct{}

// ContextWithScanner adds scanner to context for dependency injection
func ContextWithScanner(ctx context.Context, scanner types.Scanner) context.Context {
	return context.WithValue(ctx, scannerKey{}, scanner)
}

// ScannerFromContext extracts scanner from context
func ScannerFromContext(ctx context.Context) types.Scanner {
	if scanner, ok := ctx.Value(scannerKey{}).(types.Scanner); ok {
		return scanner
	}
	return nil
}

// createScannerForContext creates appropriate scanner based on context
func createScannerForContext(ctx context.Context, verbose bool) types.Scanner {
	// First try to get scanner from context (for testing)
	if scanner := ScannerFromContext(ctx); scanner != nil {
		return scanner
	}

	// Default to real scanner
	return scan.NewScanner(verbose)
}
