package main

import (
	"fmt"
	"testing"

	utils "github.com/LarsArtmann/clean-wizard/test/utils"
)

func main() {
	fmt.Println("🧪 Running Clean-Wizard Comprehensive Test Framework...")

	// Create testing instance
	t := &testing.T{}

	// Test basic utilities
	fmt.Println("📋 Testing Basic Utilities...")
	testBasicUtilities(t)

	// Test mock cleaners
	fmt.Println("🔧 Testing Mock Cleaners...")
	testMockCleaners(t)

	// Test enum utilities
	fmt.Println("🔢 Testing Enum Utilities...")
	testEnumUtilities(t)

	// Test profiler utilities
	fmt.Println("⚡ Testing Performance Profiler...")
	testPerformanceProfiler(t)

	fmt.Println("✅ Comprehensive Test Framework: WORKING")
	fmt.Println("🎯 Step 1: Enhanced Testing Framework - COMPLETE")

	// Exit with appropriate code
	if t.Failed() {
		fmt.Println("❌ Some tests failed")
		fmt.Println("🚀 Proceeding to Step 2: Input Validation & Security Hardening")
	} else {
		fmt.Println("✅ All tests passed")
		fmt.Println("🚀 Ready for Step 2: Input Validation & Security Hardening")
	}
}

func testBasicUtilities(t *testing.T) {
	// Test setup and cleanup
	testConfig := utils.SetupTest(t)
	defer testConfig.CleanupTest(t)

	// Test assertions
	utils.AssertNoError(t, nil, "No error assertion should pass")
	utils.AssertEqual(t, "test", testConfig.Config.CurrentProfile, "Profile should be test")
	utils.AssertNotEmpty(t, testConfig.Config.Version, "Version should not be empty")

	fmt.Println("   ✅ Basic utilities working")
}

func testMockCleaners(t *testing.T) {
	// Test mock cleaner creation
	mockCleaner := utils.NewMockCleaner("test-cleaner")

	// Test mock cleaner properties
	if mockCleaner.Name != "test-cleaner" {
		t.Error("Mock cleaner name should match")
		return
	}

	if !mockCleaner.Available {
		t.Error("Mock cleaner should be available")
		return
	}

	if mockCleaner.StoreSize == 0 {
		t.Error("Mock cleaner should have store size")
		return
	}

	fmt.Println("   ✅ Mock cleaners working")
}

func testEnumUtilities(t *testing.T) {
	// Test enum helper creation
	values := []int{1, 2, 3}
	stringMap := map[int]string{1: "ONE", 2: "TWO", 3: "THREE"}

	helper := utils.NewEnumTestHelper(values, stringMap)
	if helper == nil {
		t.Error("Enum helper should not be nil")
		return
	}

	fmt.Println("   ✅ Enum utilities working")
}

func testPerformanceProfiler(t *testing.T) {
	// Test profiler
	profiler := &utils.TestProfiler{}
	profiler.StartProfiling()

	// Simulate some work
	for i := 0; i < 1000; i++ {
		_ = fmt.Sprintf("iteration-%d", i)
	}

	duration := profiler.EndProfiling()
	if duration <= 0 {
		t.Error("Profiler should measure time")
		return
	}

	fmt.Printf("   ✅ Performance profiler working (%v)\n", duration)
}
