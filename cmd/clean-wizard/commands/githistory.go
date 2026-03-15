package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewGitHistoryCommand creates the git-history subcommand.
func NewGitHistoryCommand() *cobra.Command {
	var (
		dryRun          bool
		verbose         bool
		minSizeMB       int
		maxFiles        int
		scanPath        string
		force           bool
		createBackup    bool
		scanAllProjects bool
	)

	cmd := &cobra.Command{
		Use:   "git-history [path]",
		Short: "Interactive wizard to remove binary files from git history",
		Long: `Scans git history for binary files and provides an interactive wizard
to select which files to remove using git-filter-repo.

⚠️  WARNING: This rewrites git history and requires force-push.
    Always backup your repository first.
    Coordinate with your team before running on shared repos.

Examples:
  # Scan current directory with interactive selection
  clean-wizard git-history

  # Scan specific repository
  clean-wizard git-history /path/to/repo --dry-run

  # Scan all projects under ~/projects
  clean-wizard git-history --scan-all-projects

  # Quick mode: remove files > 10MB without interactive selection
  clean-wizard git-history --min-size 10 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine repo path
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			if scanPath != "" {
				path = scanPath
			}

			return runGitHistoryWizard(
				path,
				dryRun,
				verbose,
				minSizeMB,
				maxFiles,
				force,
				createBackup,
				scanAllProjects,
			)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Analyze only, don't modify history")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output")
	cmd.Flags().IntVar(&minSizeMB, "min-size", 1, "Minimum file size in MB to consider")
	cmd.Flags().IntVar(&maxFiles, "max-files", 100, "Maximum number of files to display")
	cmd.Flags().
		StringVar(&scanPath, "path", "", "Path to git repository (default: current directory)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts (dangerous)")
	cmd.Flags().BoolVar(&createBackup, "backup", true, "Create backup before rewriting")
	cmd.Flags().
		BoolVar(&scanAllProjects, "scan-all-projects", false, "Scan all projects under ~/projects")

	return cmd
}

// runGitHistoryWizard runs the interactive git history cleaning wizard.
func runGitHistoryWizard(
	basePath string,
	dryRun, verbose bool,
	minSizeMB, maxFiles int,
	force, createBackup, scanAllProjects bool,
) error {
	ctx := context.Background()

	fmt.Println(TitleStyle.Render("🔮 Git History Binary Cleaner"))
	fmt.Println()

	// Step 1: Select repositories
	var repos []string

	if scanAllProjects {
		homeDir, _ := os.UserHomeDir()
		projectsPath := filepath.Join(homeDir, "projects")
		fmt.Printf("📁 Scanning for git repositories in %s...\n", projectsPath)

		var err error

		repos, err = cleaner.FindGitRepositories(projectsPath, 3)
		if err != nil {
			return fmt.Errorf("failed to find repositories: %w", err)
		}

		if len(repos) == 0 {
			return fmt.Errorf("no git repositories found in %s", projectsPath)
		}

		fmt.Printf("✅ Found %d repositories\n\n", len(repos))

		// Let user select which repos to clean
		if !force {
			var selectedRepos []string

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select repositories to clean").
						Description("Choose which repositories to scan for binary files").
						Options(func() []huh.Option[string] {
							opts := make([]huh.Option[string], len(repos))
							for i, repo := range repos {
								opts[i] = huh.NewOption(filepath.Base(repo)+" ("+repo+")", repo)
							}

							return opts
						}()...).
						Value(&selectedRepos),
				),
			)

			err := form.Run()
			if err != nil {
				return fmt.Errorf("selection error: %w", err)
			}

			repos = selectedRepos
		}
	} else {
		// Single repository mode
		absPath, _ := filepath.Abs(basePath)
		repos = []string{absPath}
	}

	// Process each repository
	for _, repoPath := range repos {
		err := processRepository(
			ctx,
			repoPath,
			dryRun,
			verbose,
			minSizeMB,
			maxFiles,
			force,
			createBackup,
		)
		if err != nil {
			fmt.Printf("❌ Error processing %s: %v\n\n", repoPath, err)

			continue
		}
	}

	return nil
}

// processRepository processes a single repository.
func processRepository(
	ctx context.Context,
	repoPath string,
	dryRun, verbose bool,
	minSizeMB, maxFiles int,
	force, createBackup bool,
) error {
	fmt.Println(InfoStyle.Render("\n📂 Repository: " + repoPath))

	// Create cleaner
	c := cleaner.NewGitHistoryCleaner(
		cleaner.WithGitHistoryRepoPath(repoPath),
		cleaner.WithGitHistoryMinSizeMB(minSizeMB),
		cleaner.WithGitHistoryMaxFiles(maxFiles),
		cleaner.WithGitHistoryVerbose(verbose),
		cleaner.WithGitHistoryDryRun(dryRun),
		cleaner.WithGitHistoryCreateBackup(createBackup),
	)

	// Check availability
	if !c.IsAvailable(ctx) {
		return errors.New("not a git repository or git not available")
	}

	// Run safety checks
	fmt.Print("🔒 Running safety checks... ")

	safetyReport := c.GetSafetyReport(ctx)

	if len(safetyReport.Blockers) > 0 {
		fmt.Println(WarningStyle.Render("BLOCKED"))

		for _, blocker := range safetyReport.Blockers {
			fmt.Printf("   ❌ %s\n", blocker)
		}

		return errors.New("safety checks failed")
	}

	fmt.Println(SuccessStyle.Render("PASSED"))

	// Show warnings
	if len(safetyReport.Warnings) > 0 {
		for _, warning := range safetyReport.Warnings {
			fmt.Printf("   ⚠️  %s\n", warning)
		}
	}

	// Scan for binary files
	fmt.Print("🔍 Scanning git history for binary files... ")

	scanResult, err := c.GetScanResult(ctx)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(scanResult.Files) == 0 {
		fmt.Println(SuccessStyle.Render("No large binaries found!"))

		return nil
	}

	fmt.Println(
		SuccessStyle.Render(
			fmt.Sprintf(
				"Found %d file(s) (%s)",
				len(scanResult.Files),
				format.Bytes(scanResult.TotalBytes),
			),
		),
	)

	// Step 2: Show found files and let user select
	selectedFiles, err := selectFilesToClean(scanResult.Files, force)
	if err != nil {
		return fmt.Errorf("selection error: %w", err)
	}

	if len(selectedFiles) == 0 {
		fmt.Println("❌ No files selected. Nothing to do.")

		return nil
	}

	// Calculate selected size
	var selectedSize int64
	for _, f := range selectedFiles {
		selectedSize += f.SizeBytes
	}

	// Step 3: Show impact and get final confirmation
	impact, err := c.EstimateImpact(ctx)
	if err != nil {
		fmt.Printf("Warning: could not estimate impact: %v\n", err)
	}

	// Show summary
	fmt.Println()
	fmt.Println(TitleStyle.Render("📊 Summary"))
	fmt.Printf("   Repository:      %s\n", repoPath)
	fmt.Printf("   Files to remove: %d\n", len(selectedFiles))
	fmt.Printf("   Total size:      %s\n", format.Bytes(selectedSize))

	if impact != nil {
		fmt.Printf("   Current size:    %.1f MB\n", impact.CurrentRepoSizeMB)
		fmt.Printf("   Estimated new:   %.1f MB\n", impact.EstimatedNewSizeMB)
		fmt.Printf("   Space saved:     %.1f MB\n", impact.SpaceReclaimedMB)
	}

	fmt.Println()

	if dryRun {
		fmt.Println(InfoStyle.Render("🔍 DRY RUN MODE - No changes will be made"))
		fmt.Println()
	}

	// Final confirmation
	if !force && !dryRun {
		if !confirmAction(repoPath, len(selectedFiles), selectedSize, safetyReport) {
			fmt.Println("❌ Cancelled. No changes made.")

			return nil
		}
	}

	// Step 4: Execute
	c.SetSelectedFiles(selectedFiles)

	fmt.Println("\n🧹 Starting cleanup...")

	result := c.Clean(ctx)

	if result.IsErr() {
		return fmt.Errorf("cleanup failed: %w", result.Error())
	}

	// Show results
	cleanResult := result.Value()

	fmt.Println()
	fmt.Println(SuccessStyle.Render("✅ Cleanup completed!"))
	fmt.Printf("   Files processed: %d\n", cleanResult.ItemsRemoved)
	fmt.Printf("   Space freed:     %s\n", format.Bytes(int64(cleanResult.FreedBytes)))

	if cleanResult.SizeEstimate.Known > 0 {
		fmt.Printf("   Repository size: %s\n", format.Bytes(int64(cleanResult.SizeEstimate.Known)))
	}

	if dryRun {
		fmt.Println()
		fmt.Println(InfoStyle.Render("💡 Run without --dry-run to actually remove files"))
	} else {
		fmt.Println()

		if safetyReport.HasRemote {
			fmt.Println(WarningStyle.Render("⚠️  Next steps:"))
			fmt.Println("   1. Verify the repository is in good state")
			fmt.Printf(
				"   2. Force push: git push --force-with-lease %s %s\n",
				safetyReport.RemoteName,
				safetyReport.CurrentBranch,
			)
			fmt.Println("   3. Notify team members to reclone or reset")
		}
	}

	return nil
}

// selectFilesToClean shows an interactive multi-select for files.
func selectFilesToClean(
	files []domain.GitHistoryFile,
	force bool,
) ([]domain.GitHistoryFile, error) {
	if force {
		// In force mode, select all files
		return files, nil
	}

	// Sort files by size (already done by scanner, but ensure)
	slices.SortFunc(files, func(a, b domain.GitHistoryFile) int {
		if a.SizeBytes > b.SizeBytes {
			return -1
		} else if a.SizeBytes < b.SizeBytes {
			return 1
		}

		return 0
	})

	// Build options with size info
	type fileSelection struct {
		Index int
		File  domain.GitHistoryFile
	}

	var selectedIndices []int

	options := make([]huh.Option[int], len(files))
	for i, f := range files {
		status := ""
		if f.IsDeleted {
			status = " [deleted in HEAD]"
		}

		label := fmt.Sprintf("%s (%.1f MB)%s",
			f.Path, f.SizeMB(), status)
		options[i] = huh.NewOption(label, i)
	}

	// Show selection form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select binary files to remove from history").
				Description("Large files are shown first. Space to select, Enter to confirm").
				Options(options...).
				Value(&selectedIndices),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Collect selected files
	selected := make([]domain.GitHistoryFile, len(selectedIndices))
	for i, idx := range selectedIndices {
		selected[i] = files[idx]
	}

	return selected, nil
}

// confirmAction shows a confirmation dialog.
func confirmAction(
	repoPath string,
	fileCount int,
	totalSize int64,
	report *domain.GitHistorySafetyReport,
) bool {
	// Build warning message
	var warnMsg strings.Builder
	warnMsg.WriteString("⚠️  WARNING: DESTRUCTIVE OPERATION\n\n")
	fmt.Fprintf(&warnMsg, "You are about to rewrite git history for:\n  %s\n\n", repoPath)
	warnMsg.WriteString("This will:\n")
	fmt.Fprintf(&warnMsg, "  • Remove %d binary files from history\n", fileCount)
	fmt.Fprintf(&warnMsg, "  • Free approximately %s\n", format.Bytes(totalSize))
	warnMsg.WriteString("  • Require force-push to remote\n\n")

	if report.HasRemote {
		warnMsg.WriteString(WarningStyle.Render("Team Impact:\n"))
		warnMsg.WriteString("  Other team members will need to reclone or reset.\n")
		warnMsg.WriteString("  Coordinate with your team before proceeding.\n\n")
	}

	fmt.Println(warnMsg.String())

	// Confirmation checkboxes
	var (
		confirmed      bool
		understandRisk bool
		haveBackup     bool
		coordinated    bool
	)

	// Build form with all required confirmations
	var confirms []huh.Field

	confirmOpts := []huh.Option[bool]{
		huh.NewOption("Yes, proceed", true),
		huh.NewOption("No, cancel", false),
	}
	confirms = append(confirms,
		huh.NewSelect[bool]().
			Title("Confirm destructive operation?").
			Options(confirmOpts...).
			Value(&confirmed),
		huh.NewConfirm().
			Title("I understand this rewrites history and cannot be undone").
			Value(&understandRisk),
		huh.NewConfirm().
			Title("I have created a backup or am willing to risk data loss").
			Value(&haveBackup),
	)

	if report.HasRemote {
		confirms = append(confirms,
			huh.NewConfirm().
				Title("I have coordinated with my team").
				Value(&coordinated),
		)
	}

	form := huh.NewForm(huh.NewGroup(confirms...))

	err := form.Run()
	if err != nil {
		return false
	}

	return confirmed && understandRisk && haveBackup && (coordinated || !report.HasRemote)
}

// ScanRepoForDisplay is a helper to get scan results formatted for display.
func ScanRepoForDisplay(ctx context.Context, repoPath string, minSizeMB int) (*ScanDisplay, error) {
	c := cleaner.NewGitHistoryCleaner(
		cleaner.WithGitHistoryRepoPath(repoPath),
		cleaner.WithGitHistoryMinSizeMB(minSizeMB),
		cleaner.WithGitHistoryVerbose(false),
	)

	if !c.IsAvailable(ctx) {
		return nil, errors.New("not a git repository")
	}

	scanResult, err := c.GetScanResult(ctx)
	if err != nil {
		return nil, err
	}

	repoSize := c.GetStoreSize(ctx)

	return &ScanDisplay{
		RepoPath:   repoPath,
		Files:      scanResult.Files,
		TotalBytes: scanResult.TotalBytes,
		TotalFiles: scanResult.TotalFiles,
		RepoSize:   repoSize,
		Duration:   scanResult.Duration,
	}, nil
}

// ScanDisplay holds formatted scan results.
type ScanDisplay struct {
	RepoPath   string
	Files      []domain.GitHistoryFile
	TotalBytes int64
	TotalFiles int
	RepoSize   int64
	Duration   time.Duration
}
