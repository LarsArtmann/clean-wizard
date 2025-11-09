package scan

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/LarsArtmann/clean-wizard/internal/types"
)

// Scanner interface defines scanning capabilities
type Scanner interface {
	// Scan performs system scan with type-safe result
	Scan(ctx context.Context) result.Result[types.ScanResult]
	
	// ScanType returns what type of scanner this is
	ScanType() types.CleanType
	
	// Name returns human-readable name
	Name() string
	
	// IsEnabled checks if scanner is enabled for this system
	IsEnabled() bool
}

// MockScanner implements Scanner for testing
type MockScanner struct {
	scanType types.CleanType
	name     string
	results  types.ScanResult
	err      error
}

// NewMockScanner creates a mock scanner with default data
func NewMockScanner() *MockScanner {
	return &MockScanner{
		scanType: types.CleanTypeTempFiles,
		name:     "Mock Scanner",
		results: types.ScanResult{
			CleanableItems: 3,
			TotalBytes:     1024 * 1024 * 100, // 100MB
			ScanTime:       0,
			Items: []types.ScanItem{
				{Path: "/tmp/example1", Size: 1024 * 1024 * 50, Type: types.CleanTypeTempFiles},
				{Path: "/tmp/example2", Size: 1024 * 1024 * 30, Type: types.CleanTypeTempFiles},
				{Path: "/tmp/example3", Size: 1024 * 1024 * 20, Type: types.CleanTypeTempFiles},
			},
		},
	}
}

// NewEmptyMockScanner creates a mock scanner with empty results
func NewEmptyMockScanner() *MockScanner {
	return &MockScanner{
		scanType: types.CleanTypeTempFiles,
		name:     "Empty Mock Scanner",
		results: types.ScanResult{
			CleanableItems: 0,
			TotalBytes:     0,
			ScanTime:       0,
			Items:          []types.ScanItem{},
		},
	}
}

// NewErrorMockScanner creates a mock scanner that returns an error
func NewErrorMockScanner(err error) *MockScanner {
	return &MockScanner{
		scanType: types.CleanTypeTempFiles,
		name:     "Error Mock Scanner",
		err:      err,
	}
}

// Scan implements Scanner interface
func (m *MockScanner) Scan(ctx context.Context) result.Result[types.ScanResult] {
	if m.err != nil {
		return result.Err[types.ScanResult](m.err)
	}
	return result.Ok(m.results)
}

// ScanType implements Scanner interface
func (m *MockScanner) ScanType() types.CleanType {
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
func NewScanner(scanType types.CleanType, verbose bool) (Scanner, error) {
	// TODO: Implement real scanners
	switch scanType {
	case types.CleanTypeTempFiles:
		return NewMockScanner(), nil
	case types.CleanTypePackageCache:
		return NewMockScanner(), nil
	case types.CleanTypeNixStore:
		return NewMockScanner(), nil
	case types.CleanTypeHomebrew:
		return NewMockScanner(), nil
	default:
		return NewMockScanner(), nil
	}
}

// NewNixScanner creates a Nix scanner
func NewNixScanner(verbose bool) Scanner {
	// TODO: Implement real Nix scanner
	return NewMockScanner()
}
