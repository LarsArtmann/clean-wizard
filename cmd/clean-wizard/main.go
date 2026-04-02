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
	rootCmd.AddCommand(commands.NewGitHistoryCommand())

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Get version info
	info := version.Get()

	// Handle command execution with styled output
	err := fang.Execute(ctx, rootCmd,
		fang.WithVersion(info.Version),
		fang.WithCommit(info.Commit),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)

	cancel()

	if err != nil {
		os.Exit(1)
	}
}
