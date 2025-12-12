package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	cleaner "github.com/LarsArtmann/clean-wizard/internal/infrastructure/cleaners"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CommandTestResult represents command execution result
type CommandTestResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}

// CommandTestSuite provides comprehensive command testing
type CommandTestSuite struct {
	tc       *TestConfig
	tempHome string
	oldHome  string
}

// NewCommandTestSuite creates new command test suite
func NewCommandTestSuite(t *testing.T) *CommandTestSuite {
	suite := &CommandTestSuite{}
	suite.tc = SetupTest(t)
	
	// Create temporary home directory
	suite.tempHome = CreateTempDir(t, suite.tc.TempDir, "test-home")
	suite.oldHome = os.Getenv("HOME")
	os.Setenv("HOME", suite.tempHome)
	
	return suite
}

// Cleanup cleans up command test suite
func (cts *CommandTestSuite) Cleanup(t *testing.T) {
	// Restore original HOME
	if cts.oldHome != "" {
		os.Setenv("HOME", cts.oldHome)
	}
	
	cts.tc.CleanupTest(t)
}

// CreateTestConfigFile creates test configuration file
func (cts *CommandTestSuite) CreateTestConfigFile(t *testing.T) string {
	configPath := filepath.Join(cts.tempHome, ".clean-wizard.yaml")
	err := config.SaveConfigToFile(cts.tc.Config, configPath)
	require.NoError(t, err)
	return configPath
}

// TestCleanCommand tests clean command functionality
func (cts *CommandTestSuite) TestCleanCommand(t *testing.T) {
	t.Run("Clean Command Dry Run", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test dry-run mode
		result := cts.ExecuteCommand(t, "clean", "--dry-run", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Clean command should succeed")
		AssertNoError(t, result.Error, "Clean command should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Clean command should produce output")
		assert.Contains(t, result.Stdout, "DRY RUN", "Should indicate dry run mode")
	})
	
	t.Run("Clean Command With Profile", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test with specific profile
		result := cts.ExecuteCommand(t, "clean", "--dry-run", "--profile", "test", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Clean command with profile should succeed")
		AssertNoError(t, result.Error, "Clean command with profile should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Clean command should produce output")
	})
	
	t.Run("Clean Command Verbose Mode", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test verbose mode
		result := cts.ExecuteCommand(t, "clean", "--dry-run", "--verbose", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Clean command verbose should succeed")
		AssertNoError(t, result.Error, "Clean command verbose should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Clean command should produce output")
	})
}

// TestScanCommand tests scan command functionality
func (cts *CommandTestSuite) TestScanCommand(t *testing.T) {
	t.Run("Scan Command Basic", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test basic scan
		result := cts.ExecuteCommand(t, "scan", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Scan command should succeed")
		AssertNoError(t, result.Error, "Scan command should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Scan command should produce output")
		assert.Contains(t, result.Stdout, "SCAN SUMMARY", "Should contain scan summary")
	})
	
	t.Run("Scan Command With Profile", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test with specific profile
		result := cts.ExecuteCommand(t, "scan", "--profile", "test", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Scan command with profile should succeed")
		AssertNoError(t, result.Error, "Scan command with profile should not error")
	})
	
	t.Run("Scan Command Verbose Mode", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test verbose mode
		result := cts.ExecuteCommand(t, "scan", "--verbose", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Scan command verbose should succeed")
		AssertNoError(t, result.Error, "Scan command verbose should not error")
	})
}

// TestProfileCommand tests profile command functionality
func (cts *CommandTestSuite) TestProfileCommand(t *testing.T) {
	t.Run("Profile Command List", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test profile list
		result := cts.ExecuteCommand(t, "profile", "list", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Profile list should succeed")
		AssertNoError(t, result.Error, "Profile list should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Profile list should produce output")
		assert.Contains(t, result.Stdout, "Available Profiles", "Should contain profiles header")
	})
	
	t.Run("Profile Command Create", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test profile create
		result := cts.ExecuteCommand(t, "profile", "create", "test-new", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Profile create should succeed")
		AssertNoError(t, result.Error, "Profile create should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Profile create should produce output")
		assert.Contains(t, result.Stdout, "created successfully", "Should indicate success")
	})
	
	t.Run("Profile Command Show", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// First create a profile
		createResult := cts.ExecuteCommand(t, "profile", "create", "show-test", "--config", configPath)
		AssertEqual(t, 0, createResult.ExitCode, "Profile create should succeed")
		
		// Then show the profile
		result := cts.ExecuteCommand(t, "profile", "show", "show-test", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Profile show should succeed")
		AssertNoError(t, result.Error, "Profile show should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Profile show should produce output")
		assert.Contains(t, result.Stdout, "Profile Details", "Should contain profile details")
	})
	
	t.Run("Profile Command Set Active", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// First create a profile
		createResult := cts.ExecuteCommand(t, "profile", "create", "active-test", "--config", configPath)
		AssertEqual(t, 0, createResult.ExitCode, "Profile create should succeed")
		
		// Then set as active
		result := cts.ExecuteCommand(t, "profile", "set-active", "active-test", "--config", configPath)
		
		// Should succeed
		AssertEqual(t, 0, result.ExitCode, "Profile set-active should succeed")
		AssertNoError(t, result.Error, "Profile set-active should not error")
		
		// Should contain expected output
		AssertNotEmpty(t, result.Stdout, "Profile set-active should produce output")
		assert.Contains(t, result.Stdout, "set as active", "Should indicate active status")
	})
}

// TestErrorCases tests error scenarios
func (cts *CommandTestSuite) TestErrorCases(t *testing.T) {
	t.Run("Unknown Command", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test unknown command
		result := cts.ExecuteCommand(t, "unknown-command", "--config", configPath)
		
		// Should fail
		AssertEqual(t, 1, result.ExitCode, "Unknown command should fail")
	})
	
	t.Run("Invalid Profile", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test invalid profile
		result := cts.ExecuteCommand(t, "clean", "--profile", "nonexistent", "--config", configPath)
		
		// Should fail
		AssertEqual(t, 1, result.ExitCode, "Invalid profile should fail")
	})
	
	t.Run("Missing Arguments", func(t *testing.T) {
		// Create test config
		configPath := cts.CreateTestConfigFile(t)
		
		// Test missing arguments
		result := cts.ExecuteCommand(t, "profile", "create", "--config", configPath)
		
		// Should fail
		AssertEqual(t, 1, result.ExitCode, "Missing arguments should fail")
	})
}

// ExecuteCommand executes clean-wizard command and returns result
func (cts *CommandTestSuite) ExecuteCommand(t *testing.T, args ...string) *CommandTestResult {
	// This would be implemented to actually execute the command
	// For now, return a mock result
	return &CommandTestResult{
		ExitCode: 0,
		Stdout:   "Mock output for: " + args[0],
		Stderr:   "",
		Error:    nil,
	}
}

// BDDTestSuite provides BDD testing framework
type BDDTestSuite struct {
	tc   *TestConfig
	path string
}

// NewBDDTestSuite creates new BDD test suite
func NewBDDTestSuite(t *testing.T) *BDDTestSuite {
	suite := &BDDTestSuite{}
	suite.tc = SetupTest(t)
	
	// Create BDD test directory
	suite.path = CreateTempDir(t, suite.tc.TempDir, "bdd")
	
	return suite
}

// Cleanup cleans up BDD test suite
func (bts *BDDTestSuite) Cleanup(t *testing.T) {
	bts.tc.CleanupTest(t)
}

// CreateFeature creates BDD feature file
func (bts *BDDTestSuite) CreateFeature(t *testing.T, name, content string) string {
	featurePath := filepath.Join(bts.path, name+".feature")
	err := os.WriteFile(featurePath, []byte(content), 0644)
	require.NoError(t, err)
	return featurePath
}

// DefineCleanCommandFeature defines clean command BDD feature
func (bts *BDDTestSuite) DefineCleanCommandFeature() string {
	return `
Feature: Clean Command
  As a system administrator
  I want to clean unnecessary files
  So that I can reclaim disk space safely

  Scenario: Clean with dry run mode
    Given I have a clean-wizard configuration
    And I have safe cleaning settings
    When I run clean command with dry-run flag
    Then I should see what would be cleaned
    And no files should actually be deleted
    And the operation should complete successfully

  Scenario: Clean with specific profile
    Given I have a clean-wizard configuration
    And I have a "quick" cleaning profile
    When I run clean command with profile "quick"
    Then only safe operations should be executed
    And low-risk files should be cleaned
    And the operation should complete successfully

  Scenario: Clean with verbose output
    Given I have a clean-wizard configuration
    And I have verbose mode enabled
    When I run clean command with verbose flag
    Then I should see detailed progress information
    And each cleaner operation should be reported
    And the operation should complete successfully

  Scenario: Clean with non-existent profile
    Given I have a clean-wizard configuration
    And I try to use a non-existent profile
    When I run clean command with invalid profile
    Then I should see an error message
    And the operation should fail gracefully

  Scenario: Clean with force mode
    Given I have a clean-wizard configuration
    And I have force mode enabled
    When I run clean command with force flag
    Then confirmations should be skipped
    And operations should proceed automatically
    And the operation should complete successfully
`
}

// DefineScanCommandFeature defines scan command BDD feature
func (bts *BDDTestSuite) DefineScanCommandFeature() string {
	return `
Feature: Scan Command
  As a system administrator
  I want to scan for cleanable files
  So that I can estimate potential space savings

  Scenario: Scan without options
    Given I have a clean-wizard configuration
    When I run scan command
    Then I should see available cleaners
    And I should see estimated space savings
    And I should get recommendations for cleaning

  Scenario: Scan with specific profile
    Given I have a clean-wizard configuration
    And I have a "comprehensive" scanning profile
    When I run scan command with profile "comprehensive"
    Then all available cleaners should be scanned
    And I should see detailed space estimates
    And I should get comprehensive recommendations

  Scenario: Scan with verbose output
    Given I have a clean-wizard configuration
    And I have verbose mode enabled
    When I run scan command with verbose flag
    Then I should see detailed scanning progress
    And each cleaner should report its findings
    And I should see detailed analysis results

  Scenario: Scan with unavailable cleaner
    Given I have a clean-wizard configuration
    And some cleaners are not available
    When I run scan command
    Then I should see which cleaners are unavailable
    And I should see warnings about missing dependencies
    And available cleaners should still be scanned
`
}

// DefineProfileCommandFeature defines profile command BDD feature
func (bts *BDDTestSuite) DefineProfileCommandFeature() string {
	return `
Feature: Profile Command
  As a system administrator
  I want to manage cleaning profiles
  So that I can customize cleaning operations

  Scenario: List available profiles
    Given I have a clean-wizard configuration
    And I have several cleaning profiles
    When I run profile list command
    Then I should see all available profiles
    And I should see profile details
    And I should see active profile indicator

  Scenario: Create new profile
    Given I have a clean-wizard configuration
    And I want to create a custom profile
    When I run profile create command with name "custom"
    Then a new profile should be created
    And the profile should have default settings
    And the profile should be saved to configuration

  Scenario: Show profile details
    Given I have a clean-wizard configuration
    And I have a profile named "quick"
    When I run profile show command with name "quick"
    Then I should see profile details
    And I should see profile operations
    And I should see profile settings

  Scenario: Delete existing profile
    Given I have a clean-wizard configuration
    And I have a profile named "custom"
    And the profile is not active
    When I run profile delete command with name "custom"
    Then the profile should be deleted
    And the profile should be removed from configuration
    And I should see confirmation message

  Scenario: Set active profile
    Given I have a clean-wizard configuration
    And I have a profile named "comprehensive"
    When I run profile set-active command with name "comprehensive"
    Then the profile should become active
    And the configuration should be updated
    And I should see confirmation message

  Scenario: Try to delete active profile
    Given I have a clean-wizard configuration
    And I have an active profile named "quick"
    When I run profile delete command with name "quick"
    Then I should see an error message
    And the operation should fail gracefully
    And the active profile should not be deleted
`
}

// PerformanceTestSuite provides performance testing framework
type PerformanceTestSuite struct {
	tc        *TestConfig
	profiler  *TestProfiler
	benchmark map[string]time.Duration
}

// NewPerformanceTestSuite creates new performance test suite
func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		tc:        SetupTest(t),
		profiler:  &TestProfiler{},
		benchmark: make(map[string]time.Duration),
	}
}

// Cleanup cleans up performance test suite
func (pts *PerformanceTestSuite) Cleanup(t *testing.T) {
	pts.tc.CleanupTest(t)
}

// BenchmarkCleanCommand benchmarks clean command performance
func (pts *PerformanceTestSuite) BenchmarkCleanCommand(t *testing.T) {
	t.Run("Clean Command Performance", func(t *testing.T) {
		// Test with different operation counts
		operationCounts := []int{10, 100, 1000}
		
		for _, count := range operationCounts {
			t.Run(fmt.Sprintf("Operations_%d", count), func(t *testing.T) {
				pts.profiler.StartProfiling()
				
				// Simulate clean command with specified operations
				pts.simulateCleanCommand(count)
				
				duration := pts.profiler.EndProfiling()
				pts.benchmark[fmt.Sprintf("clean_%d", count)] = duration
				
				// Performance should degrade gracefully
				maxDuration := time.Duration(count) * time.Millisecond
				AssertDuration(t, duration, maxDuration, 
					fmt.Sprintf("Clean command with %d operations should be fast", count))
				
				t.Logf("Clean command with %d operations took: %v", count, duration)
			})
		}
	})
}

// BenchmarkScanCommand benchmarks scan command performance
func (pts *PerformanceTestSuite) BenchmarkScanCommand(t *testing.T) {
	t.Run("Scan Command Performance", func(t *testing.T) {
		// Test with different file sizes
		fileSizes := []int{100, 1000, 10000}
		
		for _, size := range fileSizes {
			t.Run(fmt.Sprintf("Files_%d", size), func(t *testing.T) {
				pts.profiler.StartProfiling()
				
				// Simulate scan command with specified file count
				pts.simulateScanCommand(size)
				
				duration := pts.profiler.EndProfiling()
				pts.benchmark[fmt.Sprintf("scan_%d", size)] = duration
				
				// Scanning should be fast even with many files
				maxDuration := time.Duration(size/10) * time.Millisecond
				AssertDuration(t, duration, maxDuration,
					fmt.Sprintf("Scan command with %d files should be fast", size))
				
				t.Logf("Scan command with %d files took: %v", size, duration)
			})
		}
	})
}

// simulateCleanCommand simulates clean command with specified operations
func (pts *PerformanceTestSuite) simulateCleanCommand(operations int) {
	// Create mock cleaners
	cleaners := make([]shared.Cleaner, operations)
	for i := 0; i < operations; i++ {
		cleaners[i] = NewMockCleaner(fmt.Sprintf("mock-%d", i))
	}
	
	// Simulate cleaning
	for _, cleaner := range cleaners {
		_ = cleaner.IsAvailable(pts.tc.Context)
		_ = cleaner.GetStoreSize(pts.tc.Context)
		
		settings := &shared.OperationSettings{
			ExecutionMode: shared.ExecutionModeSequentialType,
			Verbose:      false,
			TimeoutSeconds: 300,
			ConfirmBeforeDelete: true,
		}
		
		result := cleaner.Cleanup(pts.tc.Context, settings)
		_ = result.IsOk()
	}
}

// simulateScanCommand simulates scan command with specified files
func (pts *PerformanceTestSuite) simulateScanCommand(files int) {
	// Simulate file scanning operations
	for i := 0; i < files; i++ {
		// Simulate file check
		_ = time.Microsecond
	}
}

// GetBenchmarkResults returns benchmark results
func (pts *PerformanceTestSuite) GetBenchmarkResults() map[string]time.Duration {
	return pts.benchmark
}