package unit

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/pkg/scanner"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/stretchr/testify/assert"
)

// TestScannerContextInjection tests context-based dependency injection
func TestScannerContextInjection(t *testing.T) {
	t.Run("context injection works directly", func(t *testing.T) {
		// Create mock scanner
		mockScanner := scan.NewEmptyMockScanner()
		
		// Add to context
		ctx := scanner.ContextWithScanner(context.Background(), mockScanner)
		
		// Retrieve from context
		retrievedScanner := scanner.ScannerFromContext(ctx)
		
		// Should get same scanner
		assert.NotNil(t, retrievedScanner)
		assert.Equal(t, mockScanner, retrievedScanner)
	})
	
	t.Run("createScannerForContext uses context", func(t *testing.T) {
		// Create mock scanner
		mockScanner := scan.NewEmptyMockScanner()
		
		// Add to context
		ctx := scanner.ContextWithScanner(context.Background(), mockScanner)
		
		// Create scanner from context
		s := scanner.CreateScannerForContext(ctx, false)
		
		// Should get mock scanner from context
		assert.NotNil(t, s)
		assert.Equal(t, mockScanner, s)
	})
	
	t.Run("createScannerForContext falls back to real scanner", func(t *testing.T) {
		// Empty context
		ctx := context.Background()
		
		// Create scanner from context
		s := scanner.CreateScannerForContext(ctx, false)
		
		// Should get real scanner (not mock)
		assert.NotNil(t, s)
	})
}
