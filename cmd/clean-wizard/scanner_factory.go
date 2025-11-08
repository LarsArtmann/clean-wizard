package main

import (
	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// scannerFactory is responsible for creating scanner instances
type scannerFactory struct {
	useMock bool
	mockScanner *scan.MockScanner
}

// newScannerFactory creates a new scanner factory
func newScannerFactory() *scannerFactory {
	return &scannerFactory{
		useMock: false,
		mockScanner: nil,
	}
}

// createScanner creates a scanner instance based on configuration
func (f *scannerFactory) createScanner(verbose bool) types.Scanner {
	if f.useMock && f.mockScanner != nil {
		return f.mockScanner
	}
	return scan.NewScanner(verbose)
}

// setMockScanner sets a mock scanner for testing
func (f *scannerFactory) setMockScanner(mockScanner *scan.MockScanner) {
	f.mockScanner = mockScanner
	f.useMock = true
}

// resetToRealScanner resets to use real scanner
func (f *scannerFactory) resetToRealScanner() {
	f.useMock = false
	f.mockScanner = nil
}

// Global scanner factory instance
var globalScannerFactory = newScannerFactory()

// createScanner creates a scanner instance (exported for testing)
func createScanner(verbose bool) types.Scanner {
	return globalScannerFactory.createScanner(verbose)
}

// setMockScannerForTesting sets a mock scanner for testing (exported for testing)
func setMockScannerForTesting(mockScanner *scan.MockScanner) {
	globalScannerFactory.setMockScanner(mockScanner)
}

// resetToRealScannerForTesting resets to real scanner for testing (exported for testing)
func resetToRealScannerForTesting() {
	globalScannerFactory.resetToRealScanner()
}