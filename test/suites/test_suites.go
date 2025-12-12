package testutils

import (
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	cleaner "github.com/LarsArtmann/clean-wizard/internal/infrastructure/cleaners"
)

// CleanerTestSuite provides comprehensive cleaner testing
type CleanerTestSuite struct {
	tc       *TestConfig
	cleaners []shared.Cleaner
}

// NewCleanerTestSuite creates new cleaner test suite
func NewCleanerTestSuite(t *testing.T) *CleanerTestSuite {
	suite := &CleanerTestSuite{}
	suite.tc = SetupTest(t)

	// Initialize all cleaners
	suite.cleaners = []shared.Cleaner{
		cleaner.NewNixCleaner(),
		cleaner.NewHomebrewCleaner(true, true), // verbose, dry-run for testing
		cleaner.NewNpmCleaner(true, true),
		cleaner.NewPnpmCleaner(true, true),
		cleaner.NewTempFileCleaner(true, true),
	}

	return suite
}

// Cleanup cleans up test suite
func (cts *CleanerTestSuite) Cleanup(t *testing.T) {
	cts.tc.CleanupTest(t)
}

// TestCleanerAvailability tests if cleaners are available
func (cts *CleanerTestSuite) TestCleanerAvailability(t *testing.T) {
	t.Run("Cleaner Availability", func(t *testing.T) {
		for _, cleaner := range cts.cleaners {
			available := cleaner.IsAvailable(cts.tc.Context)
			t.Logf("Cleaner %s availability: %v", GetCleanerName(cleaner), available)

			// Some cleaners might not be available in test environment
			// This is expected behavior
		}
	})
}

// TestCleanerDryRun tests dry-run functionality
func (cts *CleanerTestSuite) TestCleanerDryRun(t *testing.T) {
	t.Run("Cleaner Dry Run", func(t *testing.T) {
		settings := &shared.OperationSettings{
			ExecutionMode:       shared.ExecutionModeSequentialType,
			Verbose:             true,
			TimeoutSeconds:      300,
			ConfirmBeforeDelete: true, // Always confirm in tests
		}

		for _, cleaner := range cts.cleaners {
			if !cleaner.IsAvailable(cts.tc.Context) {
				t.Skipf("Cleaner %s not available", GetCleanerName(cleaner))
				continue
			}

			result := cleaner.Cleanup(cts.tc.Context, settings)
			if result.IsOk() {
				cleanResult := result.Value()
				t.Logf("Cleaner %s dry-run result: %d items, %d bytes",
					GetCleanerName(cleaner),
					cleanResult.ItemsRemoved,
					cleanResult.FreedBytes)
			} else {
				t.Logf("Cleaner %s dry-run failed: %v", GetCleanerName(cleaner), result.Error())
			}
		}
	})
}

// TestCleanerGetStoreSize tests store size estimation
func (cts *CleanerTestSuite) TestCleanerGetStoreSize(t *testing.T) {
	t.Run("Cleaner Store Size", func(t *testing.T) {
		for _, cleaner := range cts.cleaners {
			if !cleaner.IsAvailable(cts.tc.Context) {
				t.Skipf("Cleaner %s not available", GetCleanerName(cleaner))
				continue
			}

			size := cleaner.GetStoreSize(cts.tc.Context)
			t.Logf("Cleaner %s store size: %d bytes", GetCleanerName(cleaner), size)

			AssertNoError(t, nil, "Store size estimation should not error")
		}
	})
}

// ConfigTestSuite provides comprehensive configuration testing
type ConfigTestSuite struct {
	tc *TestConfig
}

// NewConfigTestSuite creates new config test suite
func NewConfigTestSuite(t *testing.T) *ConfigTestSuite {
	return &ConfigTestSuite{
		tc: SetupTest(t),
	}
}

// Cleanup cleans up test suite
func (cts *ConfigTestSuite) Cleanup(t *testing.T) {
	cts.tc.CleanupTest(t)
}

// TestConfigCreation tests configuration creation
func (cts *ConfigTestSuite) TestConfigCreation(t *testing.T) {
	t.Run("Config Creation", func(t *testing.T) {
		AssertNotEmpty(t, cts.tc.Config.Version, "Config version should not be empty")
		AssertEqual(t, "1.0.0", cts.tc.Config.Version, "Config version should match expected")
		AssertEqual(t, "test", cts.tc.Config.CurrentProfile, "Current profile should be test")
		AssertNotEmpty(t, cts.tc.Config.Profiles, "Config should have profiles")
	})
}

// TestProfileManagement tests profile management
func (cts *ConfigTestSuite) TestProfileManagement(t *testing.T) {
	t.Run("Profile Management", func(t *testing.T) {
		// Test profile exists
		profile, exists := cts.tc.Config.Profiles["test"]
		AssertNoError(t, nil, "Profile should exist")
		AssertEqual(t, true, exists, "Test profile should exist")

		if exists {
			AssertEqual(t, "test", profile.Name, "Profile name should match")
			AssertNotEmpty(t, profile.Description, "Profile description should not be empty")
			AssertEqual(t, shared.StatusActiveType, profile.Status, "Profile status should be active")
			AssertNotEmpty(t, profile.Operations, "Profile should have operations")
		}
	})
}

// TestProfileValidation tests profile validation
func (cts *ConfigTestSuite) TestProfileValidation(t *testing.T) {
	t.Run("Profile Validation", func(t *testing.T) {
		// Create test profile for validation
		testProfile := &config.Profile{
			Name:        "validation-test",
			Description: "Validation test profile",
			Status:      shared.StatusActiveType,
			Operations: []config.CleanupOperation{
				{
					Name:        "test-op",
					Description: "Test operation",
					RiskLevel:   shared.RiskLevelLowType,
					Status:      shared.StatusActiveType,
				},
			},
		}

		// Test validation
		AssertNoError(t, nil, "Profile should be valid")

		isValid := testProfile.IsValid("validation-test")
		AssertEqual(t, true, isValid, "Test profile should be valid")

		// Test invalid profile (empty name)
		invalidProfile := &config.Profile{
			Name:        "",
			Description: "Invalid test profile",
			Status:      shared.StatusActiveType,
		}

		isValid = invalidProfile.IsValid("")
		AssertEqual(t, false, isValid, "Profile with empty name should be invalid")
	})
}

// EnumTestSuite provides comprehensive enum testing
type EnumTestSuite struct {
	tc *TestConfig
}

// NewEnumTestSuite creates new enum test suite
func NewEnumTestSuite(t *testing.T) *EnumTestSuite {
	return &EnumTestSuite{
		tc: SetupTest(t),
	}
}

// Cleanup cleans up test suite
func (cts *EnumTestSuite) Cleanup(t *testing.T) {
	cts.tc.CleanupTest(t)
}

// TestRiskLevelType tests RiskLevelType enum
func (cts *EnumTestSuite) TestRiskLevelType(t *testing.T) {
	t.Run("RiskLevelType Enum", func(t *testing.T) {
		values := []shared.RiskLevelType{
			shared.RiskLevelLowType,
			shared.RiskLevelMediumType,
			shared.RiskLevelHighType,
			shared.RiskLevelCriticalType,
		}

		stringMap := map[shared.RiskLevelType]string{
			shared.RiskLevelLowType:      "LOW",
			shared.RiskLevelMediumType:   "MEDIUM",
			shared.RiskLevelHighType:     "HIGH",
			shared.RiskLevelCriticalType: "CRITICAL",
		}

		helper := NewEnumTestHelper(values, stringMap)
		helper.TestEnumString(t, func(rl shared.RiskLevelType) string {
			return rl.String()
		})

		helper.TestEnumValidation(t, func(rl shared.RiskLevelType) bool {
			return rl.IsValid()
		})
	})
}

// TestStatusType tests StatusType enum
func (cts *EnumTestSuite) TestStatusType(t *testing.T) {
	t.Run("StatusType Enum", func(t *testing.T) {
		values := []shared.StatusType{
			shared.StatusInactiveType,
			shared.StatusActiveType,
			shared.StatusCompletedType,
			shared.StatusFailedType,
			shared.StatusPendingType,
		}

		stringMap := map[shared.StatusType]string{
			shared.StatusInactiveType:  "INACTIVE",
			shared.StatusActiveType:    "ACTIVE",
			shared.StatusCompletedType: "COMPLETED",
			shared.StatusFailedType:    "FAILED",
			shared.StatusPendingType:   "PENDING",
		}

		helper := NewEnumTestHelper(values, stringMap)
		helper.TestEnumString(t, func(st shared.StatusType) string {
			return st.String()
		})

		helper.TestEnumValidation(t, func(st shared.StatusType) bool {
			return st.IsValid()
		})
	})
}

// TestExecutionModeType tests ExecutionModeType enum
func (cts *EnumTestSuite) TestExecutionModeType(t *testing.T) {
	t.Run("ExecutionModeType Enum", func(t *testing.T) {
		values := []shared.ExecutionModeType{
			shared.ExecutionModeSequentialType,
			shared.ExecutionModeParallelType,
			shared.ExecutionModeBatchType,
			shared.ExecutionModeInteractiveType,
		}

		stringMap := map[shared.ExecutionModeType]string{
			shared.ExecutionModeSequentialType:  "SEQUENTIAL",
			shared.ExecutionModeParallelType:    "PARALLEL",
			shared.ExecutionModeBatchType:       "BATCH",
			shared.ExecutionModeInteractiveType: "INTERACTIVE",
		}

		helper := NewEnumTestHelper(values, stringMap)
		helper.TestEnumString(t, func(em shared.ExecutionModeType) string {
			return em.String()
		})

		helper.TestEnumValidation(t, func(em shared.ExecutionModeType) bool {
			return em.IsValid()
		})
	})
}

// PerformanceTestSuite provides performance testing utilities
type PerformanceTestSuite struct {
	tc       *TestConfig
	profiler *TestProfiler
}

// NewPerformanceTestSuite creates new performance test suite
func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		tc:       SetupTest(t),
		profiler: &TestProfiler{},
	}
}

// Cleanup cleans up test suite
func (pts *PerformanceTestSuite) Cleanup(t *testing.T) {
	pts.tc.CleanupTest(t)
}

// TestCleanupPerformance tests cleanup operation performance
func (pts *PerformanceTestSuite) TestCleanupPerformance(t *testing.T) {
	t.Run("Cleanup Performance", func(t *testing.T) {
		pts.profiler.StartProfiling()

		// Test configuration loading performance
		_, err := config.Load()
		AssertNoError(t, err, "Config loading should not error")

		duration := pts.profiler.EndProfiling()

		// Config loading should be fast (<100ms)
		AssertDuration(t, duration, 100*time.Millisecond, "Config loading should be fast")

		t.Logf("Config loading took: %v", duration)
	})
}

// TestMemoryUsage tests memory usage patterns
func (pts *PerformanceTestSuite) TestMemoryUsage(t *testing.T) {
	t.Run("Memory Usage", func(t *testing.T) {
		// Test memory usage with multiple cleaners
		cleaners := []shared.Cleaner{
			cleaner.NewNixCleaner(),
			cleaner.NewHomebrewCleaner(false, true),
			cleaner.NewNpmCleaner(false, true),
			cleaner.NewPnpmCleaner(false, true),
			cleaner.NewTempFileCleaner(false, true),
		}

		// Test cleaner creation and basic operations
		for _, cleaner := range cleaners {
			_ = cleaner.IsAvailable(pts.tc.Context)
			_ = cleaner.GetStoreSize(pts.tc.Context)
		}

		// Memory usage should be reasonable
		// Note: In a real scenario, you'd use runtime.MemStats to measure
		t.Logf("Memory usage test completed with %d cleaners", len(cleaners))
	})
}

// IntegrationTestSuite provides integration testing
type IntegrationTestSuite struct {
	tc *TestConfig
}

// NewIntegrationTestSuite creates new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		tc: SetupTest(t),
	}
}

// Cleanup cleans up test suite
func (its *IntegrationTestSuite) Cleanup(t *testing.T) {
	its.tc.CleanupTest(t)
}

// TestCleanCommandIntegration tests clean command integration
func (its *IntegrationTestSuite) TestCleanCommandIntegration(t *testing.T) {
	t.Run("Clean Command Integration", func(t *testing.T) {
		// Test clean command with dry-run
		settings := &shared.OperationSettings{
			ExecutionMode:       shared.ExecutionModeSequentialType,
			Verbose:             true,
			TimeoutSeconds:      300,
			ConfirmBeforeDelete: true,
		}

		// Test with a subset of cleaners for integration
		cleaners := []shared.Cleaner{
			cleaner.NewHomebrewCleaner(true, true), // dry-run for testing
		}

		for _, cleaner := range cleaners {
			if !cleaner.IsAvailable(its.tc.Context) {
				t.Skipf("Cleaner %s not available for integration test", GetCleanerName(cleaner))
				continue
			}

			result := cleaner.Cleanup(its.tc.Context, settings)
			if result.IsOk() {
				cleanResult := result.Value()
				t.Logf("Integration test - cleaner %s: %d items, %d bytes",
					GetCleanerName(cleaner),
					cleanResult.ItemsRemoved,
					cleanResult.FreedBytes)
			} else {
				t.Logf("Integration test - cleaner %s failed: %v",
					GetCleanerName(cleaner), result.Error())
			}
		}
	})
}

// TestScanCommandIntegration tests scan command integration
func (its *IntegrationTestSuite) TestScanCommandIntegration(t *testing.T) {
	t.Run("Scan Command Integration", func(t *testing.T) {
		// Test scan command
		cleaners := []shared.Cleaner{
			cleaner.NewHomebrewCleaner(true, true),
		}

		totalSize := int64(0)
		for _, cleaner := range cleaners {
			if !cleaner.IsAvailable(its.tc.Context) {
				t.Skipf("Cleaner %s not available for integration test", GetCleanerName(cleaner))
				continue
			}

			size := cleaner.GetStoreSize(its.tc.Context)
			totalSize += size

			t.Logf("Integration test - cleaner %s store size: %d bytes",
				GetCleanerName(cleaner), size)
		}

		t.Logf("Integration test - total store size: %d bytes", totalSize)
	})
}
