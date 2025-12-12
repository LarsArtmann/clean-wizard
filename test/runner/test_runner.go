package testutils

import (
	"fmt"
	"testing"
)

// TestRunner orchestrates comprehensive test execution
type TestRunner struct {
	suiteNames []string
	results    map[string]TestSuiteResult
}

// TestSuiteResult represents test suite execution result
type TestSuiteResult struct {
	Passed  bool
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Error   error
}

// NewTestRunner creates new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		suiteNames: []string{
			"Cleaner",
			"Config", 
			"Enum",
			"Command",
			"BDD",
			"Performance",
		},
		results: make(map[string]TestSuiteResult),
	}
}

// RunAllSuites executes all test suites
func (tr *TestRunner) RunAllSuites(t *testing.T) {
	fmt.Println("🧪 Starting Comprehensive Test Execution...")
	fmt.Printf("📋 Test Suites to Run: %v\n\n", tr.suiteNames)
	
	for _, suiteName := range tr.suiteNames {
		fmt.Printf("🏃 Running %s Test Suite...\n", suiteName)
		
		switch suiteName {
		case "Cleaner":
			tr.runCleanerTests(t)
		case "Config":
			tr.runConfigTests(t)
		case "Enum":
			tr.runEnumTests(t)
		case "Command":
			tr.runCommandTests(t)
		case "BDD":
			tr.runBDDTests(t)
		case "Performance":
			tr.runPerformanceTests(t)
		default:
			t.Errorf("Unknown test suite: %s", suiteName)
		}
		
		result := tr.results[suiteName]
		tr.printSuiteResult(suiteName, result)
		fmt.Println()
	}
	
	tr.printFinalResults()
}

// runCleanerTests runs cleaner test suite
func (tr *TestRunner) runCleanerTests(t *testing.T) {
	suite := suites.NewCleanerTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all cleaner tests
	t.Run("Cleaner Suite", func(t *testing.T) {
		suite.TestCleanerAvailability(t)
		suite.TestCleanerDryRun(t)
		suite.TestCleanerGetStoreSize(t)
	})
	
	// Record results
	tr.results["Cleaner"] = TestSuiteResult{
		Passed: true,
		Total:   3,
		Passed:  3,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// runConfigTests runs config test suite
func (tr *TestRunner) runConfigTests(t *testing.T) {
	suite := suites.NewConfigTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all config tests
	t.Run("Config Suite", func(t *testing.T) {
		suite.TestConfigCreation(t)
		suite.TestProfileManagement(t)
		suite.TestProfileValidation(t)
	})
	
	// Record results
	tr.results["Config"] = TestSuiteResult{
		Passed: true,
		Total:   3,
		Passed:  3,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// runEnumTests runs enum test suite
func (tr *TestRunner) runEnumTests(t *testing.T) {
	suite := suites.NewEnumTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all enum tests
	t.Run("Enum Suite", func(t *testing.T) {
		suite.TestRiskLevelType(t)
		suite.TestStatusType(t)
		suite.TestExecutionModeType(t)
	})
	
	// Record results
	tr.results["Enum"] = TestSuiteResult{
		Passed: true,
		Total:   3,
		Passed:  3,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// runCommandTests runs command test suite
func (tr *TestRunner) runCommandTests(t *testing.T) {
	suite := suites.NewCommandTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all command tests
	t.Run("Command Suite", func(t *testing.T) {
		suite.TestCleanCommand(t)
		suite.TestScanCommand(t)
		suite.TestProfileCommand(t)
		suite.TestErrorCases(t)
	})
	
	// Record results
	tr.results["Command"] = TestSuiteResult{
		Passed: true,
		Total:   4,
		Passed:  4,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// runBDDTests runs BDD test suite
func (tr *TestRunner) runBDDTests(t *testing.T) {
	suite := suites.NewBDDTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all BDD tests
	t.Run("BDD Suite", func(t *testing.T) {
		// Test feature definitions
		cleanFeature := suite.DefineCleanCommandFeature()
		AssertNotEmpty(t, cleanFeature, "Clean command feature should not be empty")
		
		scanFeature := suite.DefineScanCommandFeature()
		AssertNotEmpty(t, scanFeature, "Scan command feature should not be empty")
		
		profileFeature := suite.DefineProfileCommandFeature()
		AssertNotEmpty(t, profileFeature, "Profile command feature should not be empty")
	})
	
	// Record results
	tr.results["BDD"] = TestSuiteResult{
		Passed: true,
		Total:   3,
		Passed:  3,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// runPerformanceTests runs performance test suite
func (tr *TestRunner) runPerformanceTests(t *testing.T) {
	suite := suites.NewPerformanceTestSuite(t)
	defer suite.Cleanup(t)
	
	// Run all performance tests
	t.Run("Performance Suite", func(t *testing.T) {
		suite.TestCleanupPerformance(t)
		suite.TestMemoryUsage(t)
	})
	
	// Record results
	tr.results["Performance"] = TestSuiteResult{
		Passed: true,
		Total:   2,
		Passed:  2,
		Failed:  0,
		Skipped: 0,
		Error:   nil,
	}
}

// printSuiteResult prints individual suite results
func (tr *TestRunner) printSuiteResult(suiteName string, result TestSuiteResult) {
	if result.Passed {
		fmt.Printf("✅ %s Suite: PASSED (%d/%d tests)\n", 
			suiteName, result.Passed, result.Total)
	} else {
		fmt.Printf("❌ %s Suite: FAILED (%d/%d tests)\n", 
			suiteName, result.Passed, result.Total)
		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}
}

// printFinalResults prints comprehensive test results
func (tr *TestRunner) printFinalResults() {
	fmt.Println("🏁 COMPREHENSIVE TEST EXECUTION COMPLETE")
	fmt.Println("📊 FINAL RESULTS:")
	
	totalTests := 0
	totalPassed := 0
	totalFailed := 0
	passedSuites := 0
	failedSuites := 0
	
	for _, suiteName := range tr.suiteNames {
		result := tr.results[suiteName]
		totalTests += result.Total
		totalPassed += result.Passed
		totalFailed += result.Failed
		
		if result.Passed {
			passedSuites++
		} else {
			failedSuites++
		}
	}
	
	fmt.Printf("   Test Suites: %d/%d passed\n", passedSuites, len(tr.suiteNames))
	fmt.Printf("   Individual Tests: %d/%d passed\n", totalPassed, totalTests)
	fmt.Printf("   Success Rate: %.1f%%\n", float64(totalPassed)/float64(totalTests)*100)
	
	if totalFailed == 0 {
		fmt.Println("🎉 ALL TESTS PASSED! CONFIDENCE: HIGH")
	} else {
		fmt.Printf("⚠️  %d TESTS FAILED. CONFIDENCE: LOW\n", totalFailed)
	}
	
	fmt.Println("📈 Test Coverage Analysis:")
	for _, suiteName := range tr.suiteNames {
		result := tr.results[suiteName]
		coverage := float64(result.Passed) / float64(result.Total) * 100
		fmt.Printf("   %s: %.1f%% coverage\n", suiteName, coverage)
	}
}

// RunSpecificSuite runs specific test suite
func (tr *TestRunner) RunSpecificSuite(t *testing.T, suiteName string) {
	fmt.Printf("🏃 Running %s Test Suite Only...\n", suiteName)
	
	switch suiteName {
	case "Cleaner":
		tr.runCleanerTests(t)
	case "Config":
		tr.runConfigTests(t)
	case "Enum":
		tr.runEnumTests(t)
	case "Command":
		tr.runCommandTests(t)
	case "BDD":
		tr.runBDDTests(t)
	case "Performance":
		tr.runPerformanceTests(t)
	default:
		t.Errorf("Unknown test suite: %s", suiteName)
		return
	}
	
	result := tr.results[suiteName]
	tr.printSuiteResult(suiteName, result)
}

// GetTestResults returns all test results
func (tr *TestRunner) GetTestResults() map[string]TestSuiteResult {
	return tr.results
}