package scan

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// TODO: Migrate to domain.Scanner interface
// TODO: Remove this entire package once migration complete
// DEPRECATED: Use internal/domain package instead

// Scanner interface defines scanning capabilities
type Scanner interface {
	// Scan performs system scan with type-safe result
	Scan(ctx context.Context) result.Result[domain.ScanResult]

	// ScanType returns what type of scanner this is
	ScanType() domain.ScanType

	// Name returns human-readable name
	Name() string

	// IsEnabled checks if scanner is enabled for this system
	IsEnabled() bool
}

// MockScanner implements Scanner for testing
type MockScanner struct {
	scanType domain.ScanType
	name     string
	results  domain.ScanResult
	err      error
}

// NewMockScanner creates a mock scanner with default data
func NewMockScanner() *MockScanner {
	return &MockScanner{
		scanType: domain.ScanTypeTemp,
		name:     "Mock Scanner",
		results: domain.ScanResult{
			TotalBytes:   1024 * 1024 * 100, // 100MB
			TotalItems:   3,
			ScannedPaths: []string{"/tmp"},
			ScanTime:     0,
			ScannedAt:    time.Now(),
		},
	}
}

// NewEmptyMockScanner creates a mock scanner with empty results
func NewEmptyMockScanner() *MockScanner {
	return &MockScanner{
		scanType: domain.ScanTypeTemp,
		name:     "Empty Mock Scanner",
		results: domain.ScanResult{
			TotalBytes:   0,
			TotalItems:   0,
			ScannedPaths: []string{},
			ScanTime:     0,
			ScannedAt:    time.Now(),
		},
	}
}

// NewErrorMockScanner creates a mock scanner that returns an error
func NewErrorMockScanner(err error) *MockScanner {
	return &MockScanner{
		scanType: domain.ScanTypeTemp,
		name:     "Error Mock Scanner",
		err:      err,
	}
}

// Scan implements Scanner interface
func (m *MockScanner) Scan(ctx context.Context) result.Result[domain.ScanResult] {
	if m.err != nil {
		return result.Err[domain.ScanResult](m.err)
	}
	return result.Ok(m.results)
}

// ScanType implements Scanner interface
func (m *MockScanner) ScanType() domain.ScanType {
	return m.scanType
}

// Name implements Scanner interface
func (m *MockScanner) Name() string {
	return m.name
}

// IsEnabled implements Scanner interface
func (m *MockScanner) IsEnabled() bool {
	return true
}

// NewScanner creates a real scanner based on clean type
// TODO: Remove once domain.Scanner is adopted
func NewScanner(scanType domain.ScanType, verbose bool) (Scanner, error) {
	// TODO: Implement real scanners
	switch scanType {
	case domain.ScanTypeTemp:
		return NewMockScanner(), nil
	case domain.ScanTypeSystem:
		return NewMockScanner(), nil
	case domain.ScanTypeNixStore:
		return NewMockScanner(), nil
	case domain.ScanTypeHomebrew:
		return NewMockScanner(), nil
	default:
		return NewMockScanner(), nil
	}
}

// NewNixScanner creates a Nix scanner
// TODO: Remove once domain.Scanner is adopted
func NewNixScanner(verbose bool) Scanner {
	// TODO: Implement real Nix scanner
	return NewMockScanner()
}
