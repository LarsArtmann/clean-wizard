package main

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
)

func main() {
	rootCmd := commands.NewRootCmd()

	// Add all CLI commands
	rootCmd.AddCommand(commands.NewCleanCommand())
	rootCmd.AddCommand(commands.NewScanCommand())
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewProfileCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())

	// Handle command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		os.Exit(1)
	}
}
