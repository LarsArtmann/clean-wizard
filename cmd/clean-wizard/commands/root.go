package commands

import (
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command.
//
// Usage errors (bad flags, unknown commands, wrong argument counts) are
// classified as Rejection so the CLI boundary exits 1 with a machine-readable
// code instead of the blanket Transient default (exit 75, empty code).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "clean-wizard",
		Short: "Safe system cleanup tool",
		Long:  `A professional system cleanup tool that safely removes old files, package caches, and temporary data.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return errorfamily.NewRejection("cli.unknown_command", fmt.Sprintf("unknown command %q", args[0]))
		},
	}

	// FlagErrorFunc inherits to all subcommands: flag-parse usage errors reach
	// the CLI boundary as classified Rejection errors (exit 1, code cli.flag)
	// instead of the blanket Transient default (exit 75, empty code).
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return errorfamily.WrapRejection(err, "cli.flag", "invalid flag usage")
	})

	return root
}

// exactArgsClassified wraps cobra.ExactArgs so that argument-count usage
// errors reach the CLI boundary as classified Rejection errors.
func exactArgsClassified(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(n)(cmd, args); err != nil {
			return errorfamily.WrapRejection(err, "cli.args", "invalid argument count")
		}

		return nil
	}
}
