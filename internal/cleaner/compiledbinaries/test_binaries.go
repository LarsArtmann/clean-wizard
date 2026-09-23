package compiledbinaries

// Test binary fixtures shared by the compiledbinaries Ginkgo suite.

const (
	// TestBinarySize1MB is the size of the first test binary in MB.
	TestBinarySize1MB = 20
	// TestBinarySize2MB is the size of the second test binary in MB.
	TestBinarySize2MB = 15
	// TestBinaryTotalSizeMB is the total size of test binaries in MB.
	TestBinaryTotalSizeMB = TestBinarySize1MB + TestBinarySize2MB

	// Bytes per MB for size calculations.
	bytesPerMBForTest = 1024 * 1024
)

// StandardTestBinaries returns a slice of BinaryInfo for testing.
// This eliminates duplicate binary setup code across compiled binaries tests.
//
// Returns two test binaries with sizes 20MB and 15MB respectively.
func StandardTestBinaries() []BinaryInfo {
	return []BinaryInfo{
		{
			Path:     "/path/to/binary1",
			Size:     TestBinarySize1MB * bytesPerMBForTest,
			Category: CategoryTest,
		},
		{
			Path:     "/path/to/binary2",
			Size:     TestBinarySize2MB * bytesPerMBForTest,
			Category: CategoryBin,
		},
	}
}

// StandardTestBinariesTotalSize returns the total size of StandardTestBinaries.
func StandardTestBinariesTotalSize() int64 {
	return TestBinaryTotalSizeMB * bytesPerMBForTest
}
