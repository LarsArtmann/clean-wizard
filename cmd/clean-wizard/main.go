package main

import (
	"context"
	"os"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/charmbracelet/fang"
)

func main() {
	rootCmd := commands.NewRootCmd()

	// Add all CLI commands
	rootCmd.AddCommand(commands.NewCleanCommand())
	rootCmd.AddCommand(commands.NewScanCommand())
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewProfileCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())

	// Handle command execution with styled output
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
