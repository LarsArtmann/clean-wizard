package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	verbose         bool
	dryRun          bool
	force           bool
	profileName     string
	validationLevel string
)

// colorize adds color to output based on type
func colorize(text, color string) string {
	colors := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"purple": "\033[35m",
		"cyan":   "\033[36m",
		"reset":  "\033[0m",
	}

	if !strings.Contains(os.Getenv("NO_COLOR"), "1") {
		return colors[color] + text + colors["reset"]
	}
	return text
}

// parseValidationLevel converts string to ValidationLevel
func parseValidationLevel(level string) config.ValidationLevel {
	return commands.ParseValidationLevel(level)
}

func init() {
	// Configure zerolog with colorful output
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    false,
			TimeFormat: "2006-01-02 15:04:05",
		},
	).With().Timestamp().Caller().Logger()

	// Set global log level based on verbosity
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	rootCmd := commands.NewRootCmd()

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Force cleanup without confirmation")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "daily", "Configuration profile to use")
	rootCmd.PersistentFlags().StringVar(&validationLevel, "validation-level", "basic", "Validation level: none, basic, comprehensive, strict")

	// Add subcommands
	rootCmd.AddCommand(
		commands.NewProfileCommand(),
		commands.NewScanCommand(verbose, parseValidationLevel(validationLevel)),
		commands.NewCleanCommand(),
		commands.NewGenerateCommand(),
	)

	// Handle command execution with proper error handling
	if err := rootCmd.Execute(); err != nil {
		// Log fatal errors with context
		if verbose {
			log.Fatal().Err(err).Msg(colorize("Command execution failed", "red"))
		} else {
			fmt.Println(colorize(fmt.Sprintf("❌ Error: %s", err), "red"))
			os.Exit(1)
		}
	}
}
