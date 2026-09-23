package commands

import (
	"bytes"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/stretchr/testify/require"
)

func executeRoot(t *testing.T, args ...string) error {
	t.Helper()

	var out bytes.Buffer

	root := NewRootCmd()
	root.AddCommand(
		NewCleanCommand(),
		NewScanCommand(),
		NewInitCommand(),
		NewProfileCommand(),
		NewConfigCommand(),
		NewGitHistoryCommand(),
	)
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)

	return root.Execute()
}

func TestUsageErrorsAreClassified(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantCode string
	}{
		{
			name:     "unknown shorthand flag is Rejection",
			args:     []string{"clean", "-j"},
			wantCode: "cli.flag",
		},
		{
			name:     "unknown flag is Rejection",
			args:     []string{"clean", "--bogus-flag"},
			wantCode: "cli.flag",
		},
		{
			name:     "unknown command is Rejection",
			args:     []string{"scna"},
			wantCode: "cli.unknown_command",
		},
		{
			name:     "wrong argument count is Rejection",
			args:     []string{"profile", "show"},
			wantCode: "cli.args",
		},
		{
			name:     "too many arguments is Rejection",
			args:     []string{"profile", "delete", "a", "b"},
			wantCode: "cli.args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := executeRoot(t, tt.args...)
			require.Error(t, err)

			errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
			errorfamilytest.AssertCode(t, err, tt.wantCode)
			errorfamilytest.AssertExitCode(t, err, 1)
		})
	}
}

func TestRootNoArgsPrintsHelpWithoutError(t *testing.T) {
	t.Parallel()

	err := executeRoot(t)
	require.NoError(t, err)
}
