package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	domainConfig "github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	cleaner "github.com/LarsArtmann/clean-wizard/internal/infrastructure/cleaners"
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

// handleCleanCommand implements the clean command with full functionality
func handleCleanCommand() error {
	fmt.Println(colorize("🧹 Starting clean operation...", "cyan"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Parse command line arguments
	profileName := getProfileName()
	if profileName != "" {
		if _, exists := cfg.Profiles[profileName]; exists {
			cfg.CurrentProfile = profileName
			fmt.Println(colorize(fmt.Sprintf("✅ Using profile: %s", profileName), "green"))
		} else {
			return fmt.Errorf("profile not found: %s", profileName)
		}
	}

	// Determine if dry run
	dryRunMode := getDryRunMode()
	if dryRunMode {
		fmt.Println(colorize("🔍 DRY RUN MODE - No files will be deleted", "yellow"))
	}

	// Determine if force
	forceMode := getForceMode()
	if forceMode {
		fmt.Println(colorize("⚡ FORCE MODE - Skipping confirmations", "red"))
	}

	// Create operation settings
	settings := &shared.OperationSettings{
		ExecutionMode:       shared.ExecutionModeSequentialType,
		Verbose:             verbose,
		TimeoutSeconds:      300, // 5 minutes
		ConfirmBeforeDelete: !forceMode,
	}

	// Initialize cleaners
	cleanersList := []shared.Cleaner{
		cleaner.NewNixCleaner(),
		cleaner.NewHomebrewCleaner(verbose, dryRunMode),
		cleaner.NewNpmCleaner(verbose, dryRunMode),
		cleaner.NewPnpmCleaner(verbose, dryRunMode),
		cleaner.NewTempFileCleaner(verbose, dryRunMode),
	}

	// Execute cleanup
	ctx := context.Background()
	totalCleaned := int64(0)
	totalItems := int32(0)

	for _, cleaner := range cleanersList {
		if !cleaner.IsAvailable(ctx) {
			if verbose {
				fmt.Println(colorize(fmt.Sprintf("⚠️  %s not available on this system", getCleanerName(cleaner)), "yellow"))
			}
			continue
		}

		fmt.Println(colorize(fmt.Sprintf("🔧 Cleaning with %s...", getCleanerName(cleaner)), "cyan"))

		result := cleaner.Cleanup(ctx, settings)
		if result.IsOk() {
			cleanResult := result.Value()
			totalCleaned += int64(cleanResult.FreedBytes)
			totalItems += int32(cleanResult.ItemsRemoved)

			fmt.Println(colorize(fmt.Sprintf(
				"✅ %s: %d items cleaned, %s freed",
				getCleanerName(cleaner),
				cleanResult.ItemsRemoved,
				formatBytes(int64(cleanResult.FreedBytes)),
			), "green"))
		} else {
			fmt.Println(colorize(fmt.Sprintf(
				"❌ %s failed: %s",
				getCleanerName(cleaner),
				result.Error(),
			), "red"))
		}
	}

	// Show summary
	fmt.Println(colorize("\n📊 CLEANUP SUMMARY", "cyan"))
	fmt.Println(colorize(fmt.Sprintf("Total items cleaned: %d", totalItems), "green"))
	fmt.Println(colorize(fmt.Sprintf("Total space freed: %s", formatBytes(totalCleaned)), "green"))

	if dryRunMode {
		fmt.Println(colorize("\n🔍 This was a DRY RUN - No files were actually deleted", "yellow"))
	}

	return nil
}

// getProfileName returns the profile name from command line args
func getProfileName() string {
	for i, arg := range os.Args {
		if arg == "--profile" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

// getDryRunMode returns true if dry-run flag is set
func getDryRunMode() bool {
	for _, arg := range os.Args {
		if arg == "--dry-run" {
			return true
		}
	}
	return dryRun
}

// getForceMode returns true if force flag is set
func getForceMode() bool {
	for _, arg := range os.Args {
		if arg == "--force" {
			return true
		}
	}
	return force
}

// getCleanerName returns a human-readable name for the cleaner
func getCleanerName(c interface{}) string {
	switch c.(type) {
	case *cleaner.NixCleaner:
		return "Nix Cleaner"
	case *cleaner.HomebrewCleaner:
		return "Homebrew Cleaner"
	case *cleaner.NpmCleaner:
		return "npm Cleaner"
	case *cleaner.PnpmCleaner:
		return "pnpm Cleaner"
	case *cleaner.TempFileCleaner:
		return "Temp File Cleaner"
	default:
		return "Unknown Cleaner"
	}
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleScanCommand implements scan command with full functionality
func handleScanCommand() error {
	fmt.Println(colorize("🔍 Starting scan operation...", "cyan"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Parse command line arguments
	profileName := getProfileName()
	if profileName != "" {
		if _, exists := cfg.Profiles[profileName]; exists {
			fmt.Println(colorize(fmt.Sprintf("✅ Using profile: %s", profileName), "green"))
		} else {
			return fmt.Errorf("profile not found: %s", profileName)
		}
	}

	// Initialize scanners (reuse cleaners for now)
	scanners := []shared.Cleaner{
		cleaner.NewNixCleaner(),
		cleaner.NewHomebrewCleaner(false, true), // verbose=false, dry-run=true for scanning
		cleaner.NewNpmCleaner(false, true),
		cleaner.NewPnpmCleaner(false, true),
		cleaner.NewTempFileCleaner(false, true),
	}

	// Execute scan
	ctx := context.Background()
	totalScanned := int64(0)
	totalEstimate := int64(0)

	for _, scanner := range scanners {
		if !scanner.IsAvailable(ctx) {
			if verbose {
				fmt.Println(colorize(fmt.Sprintf("⚠️  %s not available on this system", getCleanerName(scanner)), "yellow"))
			}
			continue
		}

		fmt.Println(colorize(fmt.Sprintf("🔍 Scanning with %s...", getCleanerName(scanner)), "cyan"))

		// Get store size for estimation
		storeSize := scanner.GetStoreSize(ctx)
		if storeSize > 0 {
			totalEstimate += storeSize / 2 // Assume 50% can be cleaned
			totalScanned++

			fmt.Println(colorize(fmt.Sprintf(
				"✅ %s: %s total, estimated %s cleanable",
				getCleanerName(scanner),
				formatBytes(storeSize),
				formatBytes(storeSize/2),
			), "green"))
		} else {
			fmt.Println(colorize(fmt.Sprintf(
				"⚠️  %s: No scannable items found",
				getCleanerName(scanner),
			), "yellow"))
		}
	}

	// Show summary
	fmt.Println(colorize("\n📊 SCAN SUMMARY", "cyan"))
	fmt.Println(colorize(fmt.Sprintf("Total areas scanned: %d", totalScanned), "green"))
	fmt.Println(colorize(fmt.Sprintf("Estimated space recoverable: %s", formatBytes(totalEstimate)), "green"))
	fmt.Println(colorize("\n💡 Run 'clean-wizard clean --dry-run' to see what would be deleted", "cyan"))

	return nil
}

// handleProfileCommand implements profile management command
func handleProfileCommand() error {
	if len(os.Args) < 3 {
		fmt.Println(colorize("⚙️  Profile Management", "cyan"))
		fmt.Println("\nUsage: clean-wizard profile [subcommand] [options]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  list                    - List all profiles")
		fmt.Println("  show <name>            - Show profile details")
		fmt.Println("  create <name>           - Create new profile")
		fmt.Println("  delete <name>           - Delete profile")
		fmt.Println("  set-active <name>       - Set active profile")
		return nil
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "list":
		return handleProfileList()
	case "show":
		if len(os.Args) < 4 {
			return fmt.Errorf("profile name required: clean-wizard profile show <name>")
		}
		return handleProfileShow(os.Args[3])
	case "create":
		if len(os.Args) < 4 {
			return fmt.Errorf("profile name required: clean-wizard profile create <name>")
		}
		return handleProfileCreate(os.Args[3])
	case "delete":
		if len(os.Args) < 4 {
			return fmt.Errorf("profile name required: clean-wizard profile delete <name>")
		}
		return handleProfileDelete(os.Args[3])
	case "set-active":
		if len(os.Args) < 4 {
			return fmt.Errorf("profile name required: clean-wizard profile set-active <name>")
		}
		return handleProfileSetActive(os.Args[3])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// handleProfileList lists all available profiles
func handleProfileList() error {
	fmt.Println(colorize("📋 Available Profiles", "cyan"))

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println(colorize("No profiles found. Run 'clean-wizard profile create <name>' to create one.", "yellow"))
		return nil
	}

	for name, profile := range cfg.Profiles {
		status := "Inactive"
		if cfg.CurrentProfile == name {
			status = colorize("✅ Active", "green")
		}

		fmt.Printf("📝 %s - %s (%s)\n",
			colorize(name, "blue"),
			profile.Description,
			status)
		fmt.Printf("   Operations: %d\n", len(profile.Operations))
		fmt.Printf("   Status: %s\n", profile.Status.String())
		fmt.Println()
	}

	return nil
}

// handleProfileShow shows profile details
func handleProfileShow(name string) error {
	fmt.Printf(colorize("📝 Profile Details: %s\n", "cyan"), name)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	profile, exists := cfg.Profiles[name]
	if !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	fmt.Printf("Name: %s\n", profile.Name)
	fmt.Printf("Description: %s\n", profile.Description)
	fmt.Printf("Status: %s\n", profile.Status.String())
	fmt.Printf("Operations: %d\n", len(profile.Operations))
	fmt.Println("\nOperations:")
	for i, op := range profile.Operations {
		fmt.Printf("%d. %s - %s\n", i+1, op.Name, op.Description)
		fmt.Printf("   Risk Level: %s\n", op.RiskLevel.String())
		fmt.Printf("   Status: %s\n", op.Status.String())
	}

	return nil
}

// handleProfileCreate creates a new profile
func handleProfileCreate(name string) error {
	fmt.Printf(colorize("📝 Creating Profile: %s\n", "cyan"), name)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile already exists: %s", name)
	}

	// Create basic profile
	profile := &domainConfig.Profile{
		Name:        name,
		Description: fmt.Sprintf("%s profile - auto-generated", name),
		Status:      shared.StatusActiveType,
		Operations: []domainConfig.CleanupOperation{
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusActiveType,
			},
		},
	}

	cfg.Profiles[name] = profile

	// Save configuration
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println(colorize(fmt.Sprintf("✅ Profile '%s' created successfully", name), "green"))
	fmt.Println(colorize("Run 'clean-wizard profile show <name>' to view details", "cyan"))

	return nil
}

// handleProfileDelete deletes a profile
func handleProfileDelete(name string) error {
	fmt.Printf(colorize("🗑️  Deleting Profile: %s\n", "cyan"), name)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	// Don't allow deleting current active profile
	if cfg.CurrentProfile == name {
		return fmt.Errorf("cannot delete active profile. Switch to another profile first.")
	}

	delete(cfg.Profiles, name)

	// Save configuration
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println(colorize(fmt.Sprintf("✅ Profile '%s' deleted successfully", name), "green"))

	return nil
}

// handleProfileSetActive sets active profile
func handleProfileSetActive(name string) error {
	fmt.Printf(colorize("⚙️  Setting Active Profile: %s\n", "cyan"), name)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	cfg.CurrentProfile = name
	cfg.Updated = time.Now()

	// Save configuration
	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println(colorize(fmt.Sprintf("✅ Profile '%s' set as active", name), "green"))

	return nil
}

// saveConfig saves configuration to default location
func saveConfig(cfg *domainConfig.Config) error {
	configPath := os.Getenv("HOME") + "/.clean-wizard.yaml"
	return config.SaveConfigToFile(cfg, configPath)
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
		if err := handleCleanCommand(); err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Clean command failed: %s", err), "red"))
			os.Exit(1)
		}
	case "scan":
		if err := handleScanCommand(); err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Scan command failed: %s", err), "red"))
			os.Exit(1)
		}
	case "profile":
		if err := handleProfileCommand(); err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Profile command failed: %s", err), "red"))
			os.Exit(1)
		}
	case "generate":
		fmt.Println(colorize("📄 Generating default configuration...", "cyan"))
		defaultConfig, err := domainConfig.CreateDefaultConfig()
		if err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Failed to create config: %s", err), "red"))
			os.Exit(1)
		}

		configPath := os.Getenv("HOME") + "/.clean-wizard.yaml"
		if err := config.SaveConfigToFile(defaultConfig, configPath); err != nil {
			fmt.Println(colorize(fmt.Sprintf("❌ Failed to save config: %s", err), "red"))
			os.Exit(1)
		}

		fmt.Println(colorize(fmt.Sprintf("✅ Default configuration saved to: %s", configPath), "green"))
	default:
		fmt.Println(colorize(fmt.Sprintf("❌ Unknown command: %s", command), "red"))
		os.Exit(1)
	}
}
