package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegrationFramework tests the integration test framework itself
func TestIntegrationFramework(t *testing.T) {
	t.Run("IntegrationTestContext_CreationAndCleanup", func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		// Verify context is properly initialized
		require.NotNil(t, ctx)
		require.NotEmpty(t, ctx.TempDir())
		require.NotNil(t, ctx.Context())

		// Verify temp directory exists
		ctx.AssertDirExists(ctx.TempDir())

		// Test temp file creation
		testFile := ctx.TempFile("test.txt", "test content")
		ctx.AssertFileExists(testFile)

		// Test temp subdirectory creation
		subdir := ctx.TempSubdir("subdir")
		ctx.AssertDirExists(subdir)
	})

	t.Run("IntegrationTestContext_ProjectRootDetection", func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		projectRoot := ctx.ProjectRoot()
		require.NotEmpty(t, projectRoot)
		
		// Verify go.mod exists in project root
		ctx.AssertFileExists(filepath.Join(projectRoot, "go.mod"))
	})

	t.Run("CommandExecutor_BasicExecution", func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		executor := CreateTestExecutor(ctx)
		
		// Test help command (should always succeed)
		result := executor.RunCommand(t, []string{"--help"})
		result.AssertSuccess(t)
		result.AssertContains(t, "A professional system cleanup tool")
	})

	t.Run("TestConfigurationBuilder_Creation", func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		// Test basic configuration creation
		config := NewTestConfigurationBuilder().
			WithVersion("2.0.0").
			WithSafeMode(false).
			WithMaxDiskUsage(75).
			Build()

		require.Equal(t, "2.0.0", config.Version)
		require.Equal(t, false, config.SafeMode)
		require.Equal(t, 75, config.MaxDiskUsage)
		require.NotNil(t, config.Profiles)

		// Test configuration file creation
		configFile := NewTestConfigurationBuilder().
			WithDailyProfile().
			BuildAsFile(ctx, "test-config.yaml")

		ctx.AssertFileExists(configFile)
	})

	t.Run("FileSystemTestHelper_FileOperations", func(t *testing.T) {
		ctx := NewIntegrationTestContext(t)
		defer ctx.Cleanup()

		fs := NewFileSystemTestHelper(ctx)

		// Test file creation and existence
		_ = fs.CreateFile("test.txt", "test content")
		fs.AssertFileExists("test.txt")

		// Test directory creation and existence
		_ = fs.CreateDirectory("testdir")
		fs.AssertDirectoryExists("testdir")

		// Test file existence check
		require.True(t, fs.FileExists("test.txt"))
		require.False(t, fs.FileExists("nonexistent.txt"))

		// Test directory existence check
		require.True(t, fs.DirectoryExists("testdir"))
		require.False(t, fs.DirectoryExists("nonexistent"))
	})
}

// RunAllIntegrationFrameworkTests runs all framework tests
func RunAllIntegrationFrameworkTests(t *testing.T) {
	TestIntegrationFramework(t)
}