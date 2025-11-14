package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// IntegrationTestContext provides common utilities for integration testing
type IntegrationTestContext struct {
	t       *testing.T
	tempDir string
	cleanup []func() error
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewIntegrationTestContext creates a new integration test context
func NewIntegrationTestContext(t *testing.T) *IntegrationTestContext {
	tempDir, err := ioutil.TempDir("", "clean-wizard-integration-*")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	return &IntegrationTestContext{
		t:       t,
		tempDir: tempDir,
		ctx:     ctx,
		cancel:  cancel,
		cleanup: []func() error{},
	}
}

// Cleanup cleans up all test resources
func (c *IntegrationTestContext) Cleanup() {
	c.cancel()

	for _, cleanup := range c.cleanup {
		if err := cleanup(); err != nil {
			c.t.Logf("Cleanup error: %v", err)
		}
	}

	if err := os.RemoveAll(c.tempDir); err != nil {
		c.t.Logf("Failed to remove temp directory: %v", err)
	}
}

// Context returns the test context
func (c *IntegrationTestContext) Context() context.Context {
	return c.ctx
}

// TempDir returns the temporary directory for this test
func (c *IntegrationTestContext) TempDir() string {
	return c.tempDir
}

// TempFile creates a temporary file with the given content
func (c *IntegrationTestContext) TempFile(name, content string) string {
	path := filepath.Join(c.tempDir, name)
	err := ioutil.WriteFile(path, []byte(content), 0644)
	require.NoError(c.t, err)
	
	// Add cleanup
	c.cleanup = append(c.cleanup, func() error {
		return os.Remove(path)
	})
	
	return path
}

// TempDir creates a temporary subdirectory
func (c *IntegrationTestContext) TempSubdir(name string) string {
	path := filepath.Join(c.tempDir, name)
	err := os.MkdirAll(path, 0755)
	require.NoError(c.t, err)
	
	// Add cleanup
	c.cleanup = append(c.cleanup, func() error {
		return os.RemoveAll(path)
	})
	
	return path
}

// ProjectRoot returns the project root directory
func (c *IntegrationTestContext) ProjectRoot() string {
	dir, err := os.Getwd()
	require.NoError(c.t, err)
	
	// Find project root by looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	c.t.Fatal("Could not find project root")
	return ""
}

// AssertFileExists asserts that a file exists
func (c *IntegrationTestContext) AssertFileExists(path string) {
	_, err := os.Stat(path)
	require.NoError(c.t, err, fmt.Sprintf("File should exist: %s", path))
}

// AssertFileNotExists asserts that a file does not exist
func (c *IntegrationTestContext) AssertFileNotExists(path string) {
	_, err := os.Stat(path)
	require.Error(c.t, err, fmt.Sprintf("File should not exist: %s", path))
	require.True(c.t, os.IsNotExist(err))
}

// AssertDirExists asserts that a directory exists
func (c *IntegrationTestContext) AssertDirExists(path string) {
	stat, err := os.Stat(path)
	require.NoError(c.t, err, fmt.Sprintf("Directory should exist: %s", path))
	require.True(c.t, stat.IsDir())
}

// AssertDirNotExists asserts that a directory does not exist
func (c *IntegrationTestContext) AssertDirNotExists(path string) {
	_, err := os.Stat(path)
	require.Error(c.t, err, fmt.Sprintf("Directory should not exist: %s", path))
	require.True(c.t, os.IsNotExist(err))
}

// IntegrationTestCase represents a test case for integration testing
type IntegrationTestCase struct {
	Name        string
	Description string
	Setup       func(*IntegrationTestContext)
	Test        func(*IntegrationTestContext)
	Teardown    func(*IntegrationTestContext)
}

// RunIntegrationTest runs a single integration test case
func RunIntegrationTest(t *testing.T, testCase IntegrationTestCase) {
	t.Run(testCase.Name, func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		if testCase.Setup != nil {
			testCase.Setup(ctx)
		}

		if testCase.Test != nil {
			testCase.Test(ctx)
		}
	})
}

// RunIntegrationTests runs multiple integration test cases
func RunIntegrationTests(t *testing.T, testCases []IntegrationTestCase) {
	for _, tc := range testCases {
		RunIntegrationTest(t, tc)
	}
}