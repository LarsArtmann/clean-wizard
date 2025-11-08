package main

import (
	"fmt"
	"testing"
)

func TestSimpleEmpty(t *testing.T) {
	mockScanner := CreateEmptyMockScanner()
	fmt.Printf("Mock scanner: %+v\n", mockScanner)

	output := RunCommandWithMock([]string{"scan"}, mockScanner)
	fmt.Printf("Output: %s\n", output.Stdout)
}
