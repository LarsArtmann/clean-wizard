package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/clean-wizard/cmd/clean-wizard/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version     = "dev"
	verbose     bool
	dryRun      bool
	force       bool
	profileName string
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

	// Add subcommands
	rootCmd.AddCommand(
		commands.NewScanCommand(false),
		commands.NewCleanCommand(),
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