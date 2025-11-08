package scan

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// MockScanner is a mock implementation of the Scanner interface for testing
type MockScanner struct {
	Results     *types.ScanResults
	Error       error
	ScanDelay   time.Duration
	Called      bool
	CancelCheck bool // Whether to check for context cancellation
}

// NewMockScanner creates a new mock scanner
func NewMockScanner() *MockScanner {
	return &MockScanner{
		Results: &types.ScanResults{
			Timestamp: time.Now(),
			Results: []types.ScanResult{
				{
					Name:        "test-nix-store",
					SizeGB:      2.5,
					Description: "Test Nix store data",
					Cleanable:   true,
				},
				{
					Name:        "test-homebrew",
					SizeGB:      0.5,
					Description: "Test Homebrew caches",
					Cleanable:   true,
				},
				{
					Name:        "test-package-caches",
					SizeGB:      1.0,
					Description: "Test package manager caches",
					Cleanable:   true,
				},
			},
			TotalSizeGB: 4.0,
		},
		Error:       nil,
		ScanDelay:   0,
		Called:      false,
		CancelCheck: false,
	}
}

// Scan performs the mock scan
func (m *MockScanner) Scan(ctx context.Context) (*types.ScanResults, error) {
	m.Called = true

	// Simulate scan delay if specified
	if m.ScanDelay > 0 {
		select {
		case <-time.After(m.ScanDelay):
			// Continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check for context cancellation if enabled
	if m.CancelCheck {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue
		}
	}

	return m.Results, m.Error
}

// Name returns the scanner name
func (m *MockScanner) Name() string {
	return "mock-scanner"
}

// WithResults sets the results to return
func (m *MockScanner) WithResults(results *types.ScanResults) *MockScanner {
	m.Results = results
	return m
}

// WithError sets the error to return
func (m *MockScanner) WithError(err error) *MockScanner {
	m.Error = err
	return m
}

// WithScanDelay sets the scan delay
func (m *MockScanner) WithScanDelay(delay time.Duration) *MockScanner {
	m.ScanDelay = delay
	return m
}

// WithCancelCheck enables/disables context cancellation check
func (m *MockScanner) WithCancelCheck(check bool) *MockScanner {
	m.CancelCheck = check
	return m
}

// WasCalled returns whether the scanner was called
func (m *MockScanner) WasCalled() bool {
	return m.Called
}

// Reset resets the mock scanner state
func (m *MockScanner) Reset() {
	m.Called = false
}

// NewEmptyMockScanner creates a mock scanner with empty results
func NewEmptyMockScanner() *MockScanner {
	return &MockScanner{
		Results: &types.ScanResults{
			Timestamp:   time.Now(),
			Results:     []types.ScanResult{},
			TotalSizeGB: 0,
		},
		Error:       nil,
		ScanDelay:   0,
		Called:      false,
		CancelCheck: false,
	}
}

// NewErrorMockScanner creates a mock scanner that returns an error
func NewErrorMockScanner(err error) *MockScanner {
	return &MockScanner{
		Results:     nil,
		Error:       err,
		ScanDelay:   0,
		Called:      false,
		CancelCheck: false,
	}
}