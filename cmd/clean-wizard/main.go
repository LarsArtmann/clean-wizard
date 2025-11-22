package main

import (
	"fmt"
	"os"
	"strings"

	appconfig "github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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
func parseValidationLevel(level string) shared.ValidationLevelType {
	switch strings.ToLower(level) {
	case "none":
		return shared.ValidationLevelNoneType
	case "basic":
		return shared.ValidationLevelBasicType
	case "comprehensive":
		return shared.ValidationLevelComprehensiveType
	case "strict":
		return shared.ValidationLevelStrictType
	default:
		return shared.ValidationLevelBasicType
	}
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
	if len(os.Args) < 2 {
		fmt.Println(colorize("Clean Wizard - Safe System Cleaning Tool", "cyan"))
		fmt.Println("\nUsage: clean-wizard [command] [options]")
		fmt.Println("\nAvailable Commands:")
		fmt.Println("  clean    - Clean system based on configuration")
		fmt.Println("  scan     - Scan for cleanable files")
		fmt.Println("  profile  - Manage configuration profiles")
		fmt.Println("  generate - Generate default configuration")
		fmt.Println("\nGlobal Options:")
		fmt.Println("  -v, --verbose         Enable verbose output")
		fmt.Println("      --dry-run        Show what would be deleted without actually deleting")
		fmt.Println("      --force          Force cleanup without confirmation")
		fmt.Println("      --profile <name> Configuration profile to use")
		fmt.Println("      --validation-level <level> Validation level: none, basic, comprehensive, strict")
		os.Exit(0)
	}

	command := os.Args[1]
	switch command {
	case "clean":
		fmt.Println(colorize("🧹 Clean command (not yet implemented)", "yellow"))
	case "scan":
		fmt.Println(colorize("🔍 Scan command (not yet implemented)", "yellow"))
	case "profile":
		fmt.Println(colorize("⚙️  Profile command (not yet implemented)", "yellow"))
	case "generate":
		fmt.Println(colorize("📄 Generating default configuration...", "cyan"))
		defaultConfig, err := config.CreateDefaultConfig()
		if err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Failed to create config: %s", err), "red"))
			os.Exit(1)
		}
		
		configPath := os.Getenv("HOME") + "/.clean-wizard.yaml"
		if err := appconfig.SaveConfigToFile(defaultConfig, configPath); err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Failed to save config: %s", err), "red"))
			os.Exit(1)
		}
		
		fmt.Println(colorize(fmt.Sprintf("✅ Default configuration saved to: %s", configPath), "green"))
	default:
		fmt.Println(colorize(fmt.Sprintf("❌ Unknown command: %s", command), "red"))
		os.Exit(1)
	}
}