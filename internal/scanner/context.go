package scanner

import (
	"context"
	"errors"

	"github.com/LarsArtmann/clean-wizard/internal/scan"
)

var (
	ErrScannerNotFound = errors.New("scanner not found in context")
)

type contextKey string

const ScannerContextKey contextKey = "scanner"

// ContextWithScanner adds scanner to context
func ContextWithScanner(ctx context.Context, scanner scan.Scanner) context.Context {
	return context.WithValue(ctx, ScannerContextKey, scanner)
}

// ScannerFromContext retrieves scanner from context
func ScannerFromContext(ctx context.Context) scan.Scanner {
	if scanner, ok := ctx.Value(ScannerContextKey).(scan.Scanner); ok {
		return scanner
	}
	return nil
}

// CreateScannerForContext creates scanner from context with fallback
func CreateScannerForContext(ctx context.Context, verbose bool) (scan.Scanner, error) {
	if scanner := ScannerFromContext(ctx); scanner != nil {
		return scanner, nil
	}

	// Fallback to real scanner
	// TODO: Replace with proper factory pattern
	return scan.NewNixScanner(verbose), nil
}
