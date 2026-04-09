package main

import (
	"context"
	"os"
	"syscall"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/LarsArtmann/clean-wizard/internal/version"
	"github.com/charmbracelet/fang"
)

func main() {
	rootCmd := commands.NewRootCmd()

	rootCmd.AddCommand(commands.NewCleanCommand())
	rootCmd.AddCommand(commands.NewScanCommand())
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewProfileCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())
	rootCmd.AddCommand(commands.NewGitHistoryCommand())

	info := version.Get()

	err := fang.Execute(context.Background(), rootCmd,
		fang.WithVersion(info.Version),
		fang.WithCommit(info.Commit),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)
	if err != nil {
		os.Exit(1)
	}
}
