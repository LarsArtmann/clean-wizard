package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/LarsArtmann/clean-wizard/internal/version"
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

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get version info
	info := version.Get()

	// Handle command execution with styled output
	if err := fang.Execute(ctx, rootCmd,
		fang.WithVersion(info.Version),
		fang.WithCommit(info.Commit),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	); err != nil {
		os.Exit(1)
	}
}
