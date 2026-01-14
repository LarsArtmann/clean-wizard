package commands

import "github.com/spf13/cobra"

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean-wizard",
		Short: "Safe system cleanup tool",
		Long:  `A professional system cleanup tool that safely removes old files, package caches, and temporary data.`,
	}
}
