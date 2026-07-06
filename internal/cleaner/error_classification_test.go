package cleaner

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/stretchr/testify/assert"
)

func TestPathErrorClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want errorfamily.Family
	}{
		{
			name: "ENOENT → Rejection via stdlib os.ErrNotExist sentinel",
			err:  &os.PathError{Op: "open", Path: "/missing", Err: syscall.ENOENT},
			want: errorfamily.Rejection,
		},
		{
			name: "EACCES → Rejection via stdlib os.ErrPermission sentinel",
			err:  &os.PathError{Op: "open", Path: "/secret", Err: syscall.EACCES},
			want: errorfamily.Rejection,
		},
		{
			name: "ENOSPC → Rejection (disk full, permanent)",
			err:  &os.PathError{Op: "write", Path: "/tmp/full", Err: syscall.ENOSPC},
			want: errorfamily.Rejection,
		},
		{
			name: "EROFS → Rejection (read-only filesystem, permanent)",
			err:  &os.PathError{Op: "write", Path: "/ro/file", Err: syscall.EROFS},
			want: errorfamily.Rejection,
		},
		{
			name: "EIO → Transient (transient I/O, retryable)",
			err:  &os.PathError{Op: "read", Path: "/dev/sda", Err: syscall.EIO},
			want: errorfamily.Transient,
		},
		{
			name: "EBUSY → Transient (resource busy, retryable)",
			err:  &os.PathError{Op: "open", Path: "/dev/loop0", Err: syscall.EBUSY},
			want: errorfamily.Transient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errorfamily.Classify(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotAvailableError_ClassificationAndRetryability(t *testing.T) {
	t.Parallel()

	err := NewNotAvailableError("cargo", "")

	errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	errorfamilytest.AssertRetryable(t, err, false)
	errorfamilytest.AssertCode(t, err, "cleaner.cargo.not_available")
}

func TestNotAvailableError_FactoryDerivesPerCleanerCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cleanerName string
		wantCode    string
	}{
		{"cargo", "cargo", "cleaner.cargo.not_available"},
		{"go", "go", "cleaner.go.not_available"},
		{"docker", "docker", "cleaner.docker.not_available"},
		{"homebrew", "homebrew", "cleaner.homebrew.not_available"},
		{"systemcache", "systemcache", "cleaner.systemcache.not_available"},
		{"pma", "projects-management-automation", "cleaner.projects-management-automation.not_available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewNotAvailableError(tt.cleanerName, "")
			errorfamilytest.AssertCode(t, err, tt.wantCode)
		})
	}
}

func TestNotAvailableError_EmptyCodeFallsBackToDefault(t *testing.T) {
	t.Parallel()

	err := &NotAvailableError{CleanerName: "go"}
	errorfamilytest.AssertCode(t, err, "cleaner.not_available")
}

func TestExecErrNotFound_ClassifiedAsInfrastructure(t *testing.T) {
	t.Parallel()

	errorfamilytest.AssertFamily(t, exec.ErrNotFound, errorfamily.Infrastructure)
	errorfamilytest.AssertRetryable(t, exec.ErrNotFound, false)
}
