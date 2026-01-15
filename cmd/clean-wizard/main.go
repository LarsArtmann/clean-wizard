package main

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
)

func main() {
	rootCmd := commands.NewRootCmd()

	// Add only the clean command
	rootCmd.AddCommand(commands.NewCleanCommand())

	// Handle command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		os.Exit(1)
	}
}
