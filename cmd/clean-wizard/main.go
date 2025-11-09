package main

import (
	"os"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	verbose bool
	dryRun  bool
	force   bool
	profileName string
)

func main() {
	rootCmd := NewRootCmd()

	// Set up logging
	logrus.SetOutput(os.Stderr)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("Command failed")
		os.Exit(1)
	}
}

// NewRootCmd creates and returns root command
func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "clean-wizard",
		Short:   "Interactive system cleaning wizard",
		Long:    "A comprehensive CLI/TUI tool for system cleanup",
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without doing it")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "comprehensive", "Cleaning profile to use")

	// Add commands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewScanCommand(verbose))
	rootCmd.AddCommand(commands.NewCleanCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())

	return rootCmd
}
