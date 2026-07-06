package main

import (
	"context"
	"log/slog"
	"os"
	"syscall"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/LarsArtmann/clean-wizard/internal/version"
	"github.com/charmbracelet/fang"
	errorfamily "github.com/larsartmann/go-error-family"
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

	err := fang.Execute(
		context.Background(), rootCmd,
		fang.WithVersion(info.Version),
		fang.WithCommit(info.Commit),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)
	if err != nil {
		errorfamily.LogError(err, slog.Default())
		os.Exit(errorfamily.ExitCode(err))
	}
}
