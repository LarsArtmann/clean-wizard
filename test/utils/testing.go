package testutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig provides standardized test configuration
type TestConfig struct {
	TempDir string
	Config  *config.Config
	Context context.Context
	Cancel  context.CancelFunc
}

// SetupTest creates standardized test environment
func SetupTest(t *testing.T) *TestConfig {
	tc := &TestConfig{}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "clean-wizard-test-*")
	require.NoError(t, err)
	tc.TempDir = tempDir

	// Create test context with timeout
	tc.Context, tc.Cancel = context.WithTimeout(context.Background(), 30*time.Second)

	// Create test configuration
	tc.Config = createTestConfig(t, tempDir)

	return tc
}

// CleanupTest cleans up test environment
func (tc *TestConfig) CleanupTest(t *testing.T) {
	if tc.Cancel != nil {
		tc.Cancel()
	}
	if tc.TempDir != "" {
		os.RemoveAll(tc.TempDir)
	}
}

// createTestConfig creates test configuration
func createTestConfig(t *testing.T, tempDir string) *config.Config {
	cfg, err := config.CreateDefaultConfig()
	require.NoError(t, err)

	// Set test paths
	cfg.Protected = []string{
		tempDir + "/protected",
	}

	// Create test profiles
	cfg.Profiles["test"] = &config.Profile{
		Name:        "test",
		Description: "Test profile",
		Status:      shared.StatusActiveType,
		Operations: []config.CleanupOperation{
			{
				Name:        "test-operation",
				Description: "Test operation",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusActiveType,
			},
		},
	}

	cfg.CurrentProfile = "test"
	return cfg
}

// AssertNoError is a wrapper around assert.NoError that includes context
func AssertNoError(t *testing.T, err error, contextMsg string) {
	if !assert.NoError(t, err, contextMsg) {
		t.FailNow()
	}
}

// AssertEqual is a wrapper around assert.Equal that includes context
func AssertEqual[T comparable](t *testing.T, expected, actual T, contextMsg string) {
	if !assert.Equal(t, expected, actual, contextMsg) {
		t.FailNow()
	}
}

// AssertNotEmpty checks if a value is not empty
func AssertNotEmpty(t *testing.T, value interface{}, contextMsg string) {
	if !assert.NotEmpty(t, value, contextMsg) {
		t.FailNow()
	}
}

// CreateTempFile creates a temporary file for testing
func CreateTempFile(t *testing.T, dir, filename, content string) string {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	return path
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T, parent, dirname string) string {
	path := filepath.Join(parent, dirname)
	err := os.MkdirAll(path, 0o755)
	require.NoError(t, err)
	return path
}

// MockCleaner implements shared.Cleaner for testing
type MockCleaner struct {
	Name        string
	Available   bool
	StoreSize   int64
	CleanResult shared.CleanResult
}

// IsAvailable returns mock availability
func (m *MockCleaner) IsAvailable(ctx context.Context) bool {
	return m.Available
}

// Cleanup returns mock clean result
func (m *MockCleaner) Cleanup(ctx context.Context, settings *shared.OperationSettings) result.Result[shared.CleanResult] {
	if settings.ConfirmBeforeDelete {
		// Mock confirmation check
		return result.Ok(m.CleanResult)
	}
	return result.Ok(m.CleanResult)
}

// GetStoreSize returns mock store size
func (m *MockCleaner) GetStoreSize(ctx context.Context) int64 {
	return m.StoreSize
}

// NewMockCleaner creates a new mock cleaner
func NewMockCleaner(name string) *MockCleaner {
	return &MockCleaner{
		Name:      name,
		Available: true,
		StoreSize: 1024 * 1024, // 1MB
		CleanResult: shared.CleanResult{
			FreedBytes:   512 * 1024, // 512KB
			ItemsRemoved: 10,
			ItemsFailed:  0,
			CleanTime:    time.Second,
			CleanedAt:    time.Now(),
			Strategy:     shared.StrategyConservative,
		},
	}
}

// TestProfiler provides performance testing utilities
type TestProfiler struct {
	StartTime time.Time
	EndTime   time.Time
}

// StartProfiling starts performance profiling
func (tp *TestProfiler) StartProfiling() {
	tp.StartTime = time.Now()
}

// EndProfiling ends performance profiling
func (tp *TestProfiler) EndProfiling() time.Duration {
	tp.EndTime = time.Now()
	return tp.EndTime.Sub(tp.StartTime)
}

// AssertDuration asserts operation completes within expected duration
func AssertDuration(t *testing.T, actual, expectedMax time.Duration, contextMsg string) {
	if !assert.Less(t, actual, expectedMax, contextMsg) {
		t.FailNow()
	}
}

// EnumTestHelper provides utilities for enum testing
type EnumTestHelper[T ~int] struct {
	EnumValues []T
	StringMap  map[T]string
}

// NewEnumTestHelper creates new enum test helper
func NewEnumTestHelper[T ~int](values []T, stringMap map[T]string) *EnumTestHelper[T] {
	return &EnumTestHelper[T]{
		EnumValues: values,
		StringMap:  stringMap,
	}
}

// TestEnumString tests enum String() method
func (eth *EnumTestHelper[T]) TestEnumString(t *testing.T, enumFunc func(T) string) {
	for _, value := range eth.EnumValues {
		expected := eth.StringMap[value]
		actual := enumFunc(value)
		AssertEqual(t, expected, actual, "String() method")
	}
}

// TestEnumValidation tests enum validation methods
func (eth *EnumTestHelper[T]) TestEnumValidation(t *testing.T, isValidFunc func(T) bool) {
	for _, value := range eth.EnumValues {
		assert.True(t, isValidFunc(value), "Enum value should be valid")
	}
}
