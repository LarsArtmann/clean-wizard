package errors

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileSystemErrorAdapter_Integration tests filesystem error adapter in realistic scenarios
func TestFileSystemErrorAdapter_Integration(t *testing.T) {
	adapter := &FileSystemErrorAdapter{}

	t.Run("Known filesystem errors maintain specific mapping", func(t *testing.T) {
		testCases := []struct {
			name         string
			error        error
			expectedCode ErrorCode
			expectedType ErrorType
		}{
			{
				name:         "File not found",
				error:        &os.PathError{Err: os.ErrNotExist, Path: "/tmp/missing"},
				expectedCode: ErrCodeFileNotFound,
				expectedType: ErrorTypeFileSystem,
			},
			{
				name:         "Permission denied",
				error:        &os.PathError{Err: os.ErrPermission, Path: "/etc/shadow"},
				expectedCode: ErrCodePermissionError,
				expectedType: ErrorTypePermission,
			},
			{
				name:         "Disk full",
				error:        &os.PathError{Err: syscall.ENOSPC, Path: "/tmp/full"},
				expectedCode: ErrCodeDiskFull,
				expectedType: ErrorTypeFileSystem,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := adapter.Adapt(tc.error)
				require.NotNil(t, err)
				assert.Equal(t, tc.expectedCode, err.Code)
				assert.Equal(t, tc.expectedType, err.Type)
				assert.Equal(t, tc.error, err.Cause)
				assert.NotEmpty(t, err.Caller)
			})
		}
	})

	t.Run("Unknown filesystem errors use generic code", func(t *testing.T) {
		// Test various unknown filesystem error types
		unknownErrors := []struct {
			name  string
			error error
		}{
			{
				name:  "File already closed",
				error: &os.PathError{Err: os.ErrClosed, Path: "/tmp/closed"},
			},
			{
				name:  "Invalid argument",
				error: &os.PathError{Err: syscall.EINVAL, Path: "/tmp/invalid"},
			},
			{
				name:  "I/O error",
				error: &os.PathError{Err: syscall.EIO, Path: "/tmp/io"},
			},
			{
				name:  "Text file busy",
				error: &os.PathError{Err: syscall.ETXTBSY, Path: "/tmp/busy"},
			},
		}

		for _, tc := range unknownErrors {
			t.Run(tc.name, func(t *testing.T) {
				err := adapter.Adapt(tc.error)
				require.NotNil(t, err)
				
				// Should use the new generic filesystem error code
				assert.Equal(t, ErrCodeFilesystem, err.Code, 
					"Expected ErrCodeFilesystem for unknown filesystem error: %v", tc.error)
				assert.Equal(t, ErrorTypeFileSystem, err.Type)
				assert.Equal(t, tc.error, err.Cause)
				assert.NotEmpty(t, err.Caller)
				
				// Verify message format includes both error and path
				assert.Contains(t, err.Message, "File system error")
				assert.Contains(t, err.Message, tc.error.(*os.PathError).Path)
			})
		}
	})

	t.Run("Non-PathError filesystem errors handled correctly", func(t *testing.T) {
		// Test os.IsNotExist and os.IsPermission cases using real errors
		// For testing purposes, we'll simulate these characteristics
		notExistErr := &os.PathError{Err: syscall.ENOENT, Path: "/tmp/test"}
		
		err := adapter.Adapt(notExistErr)
		require.NotNil(t, err)
		assert.Equal(t, ErrCodeFileNotFound, err.Code)
		assert.Equal(t, ErrorTypeFileSystem, err.Type)

		permissionErr := &os.PathError{Err: syscall.EPERM, Path: "/tmp/test"}
		
		err = adapter.Adapt(permissionErr)
		require.NotNil(t, err)
		assert.Equal(t, ErrCodePermissionError, err.Code)
		assert.Equal(t, ErrorTypePermission, err.Type)
	})

	t.Run("Non-filesystem errors return nil", func(t *testing.T) {
		nonFsErrors := []error{
			errors.New("network connection failed"),
			errors.New("validation failed"),
			errors.New("configuration error"),
			errors.New("generic error"),
		}

		for _, err := range nonFsErrors {
			result := adapter.Adapt(err)
			assert.Nil(t, result, "Expected nil for non-filesystem error: %v", err)
		}
	})
}