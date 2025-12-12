package main

import (
	"os"
	"testing"

	testutils "github.com/LarsArtmann/clean-wizard/test/utils"
)

// Run all comprehensive tests
func main() {
	// Create testing instance
	t := &testing.T{}
	
	// Validate test infrastructure
	testutils.AssertTestInfrastructure(t)
	
	// Run comprehensive tests
	testutils.RunComprehensiveTests(t)
	
	// Run specific test suites
	testutils.RunCommandTests(t)
	testutils.RunBDDTests(t)
	testutils.RunPerformanceTests(t)
	testutils.RunIntegrationTests(t)
	
	// Generate final test report
	testutils.GenerateTestReport(t)
	
	// Exit with appropriate code
	if t.Failed() {
		os.Exit(1)
	}
	os.Exit(0)
}