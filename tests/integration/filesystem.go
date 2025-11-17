package integration

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"
)

// FileSystemTestHelper provides utilities for filesystem operations in integration tests
type FileSystemTestHelper struct {
	ctx *IntegrationTestContext
}

// NewFileSystemTestHelper creates a new filesystem test helper
func NewFileSystemTestHelper(ctx *IntegrationTestContext) *FileSystemTestHelper {
	return &FileSystemTestHelper{ctx: ctx}
}

// CreateFile creates a file with given content
func (h *FileSystemTestHelper) CreateFile(path, content string) string {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	err := os.WriteFile(fullPath, []byte(content), 0o644)
	require.NoError(h.ctx.t, err, "Failed to create file: %s", fullPath)
	return fullPath
}

// CreateDirectory creates a directory
func (h *FileSystemTestHelper) CreateDirectory(path string) string {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	err := os.MkdirAll(fullPath, 0o755)
	require.NoError(h.ctx.t, err, "Failed to create directory: %s", fullPath)
	return fullPath
}

// FileExists checks if a file exists
func (h *FileSystemTestHelper) FileExists(path string) bool {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// DirectoryExists checks if a directory exists
func (h *FileSystemTestHelper) DirectoryExists(path string) bool {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	stat, err := os.Stat(fullPath)
	return err == nil && stat.IsDir()
}

// AssertFileExists asserts that a file exists
func (h *FileSystemTestHelper) AssertFileExists(path string) {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	h.ctx.AssertFileExists(fullPath)
}

// AssertFileNotExists asserts that a file does not exist
func (h *FileSystemTestHelper) AssertFileNotExists(path string) {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	h.ctx.AssertFileNotExists(fullPath)
}

// AssertDirectoryExists asserts that a directory exists
func (h *FileSystemTestHelper) AssertDirectoryExists(path string) {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	h.ctx.AssertDirExists(fullPath)
}

// AssertDirectoryNotExists asserts that a directory does not exist
func (h *FileSystemTestHelper) AssertDirectoryNotExists(path string) {
	fullPath := filepath.Join(h.ctx.TempDir(), path)
	h.ctx.AssertDirNotExists(fullPath)
}

// Cleanup creates a temporary directory for testing files
func (h *FileSystemTestHelper) Cleanup(name string) string {
	tempDir := filepath.Join(h.ctx.TempDir(), name)
	err := os.MkdirAll(tempDir, 0o755)
	require.NoError(h.ctx.t, err)
	return tempDir
}
