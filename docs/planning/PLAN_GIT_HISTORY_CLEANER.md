# Git History Binary Cleaner - Execution Plan

> **Feature:** Safely remove binary data from git histories  
> **Target:** Compiled binaries, build artifacts, and accidentally committed blobs  
> **Priority:** High - Common pain point for Go projects  
> **Estimated Effort:** 3-4 days  
> **Last Updated:** 2026-02-11

---

## Overview

A specialized cleaner that removes accidentally committed binary files from git history. This is a **destructive but valuable** operation for repositories with:

- Compiled binaries (`main`, `app`, `*.exe`)
- Build artifacts (`*.o`, `*.a`, `dist/`, `bin/`)
- Large blobs committed by mistake
- Vendor directories that should be gitignored

### Why This Matters

| Problem                     | Impact                  |
| --------------------------- | ----------------------- |
| Cloned binaries in history  | Repo bloat, slow clones |
| `go build` output committed | Unnecessary conflicts   |
| `vendor/` committed         | 100MB+ repo size        |
| CI artifacts in git         | Security risk, bloat    |

### Safety-First Design

This cleaner is **intentionally cautious**:

1. **Dry-run by default** - Shows what would be removed
2. **Explicit confirmation** - Must type repository name to confirm
3. **Backup requirement** - Requires `--force` without backup
4. **History rewriting warnings** - Explains consequences clearly
5. **Team coordination alerts** - Warns about shared history changes

---

## Architecture

### Integration Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHistoryCleaner                         │
│                      (new cleaner)                           │
├─────────────────────────────────────────────────────────────┤
│  Scanner → Analyzer → SafetyChecks → Executor → Reporter    │
│     │          │           │           │          │         │
│     ▼          ▼           ▼           ▼          ▼         │
│  git log   size calc   backup check  filter    summary      │
│  --stat    mime type   force push    --strip   results      │
│            detection   warning        blobs                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Existing CleanerRegistry                        │
│         (integrates like Nix, Docker, Go cleaners)          │
└─────────────────────────────────────────────────────────────┘
```

### Component Breakdown

```
internal/cleaner/
├── githistory.go                    # Main cleaner implementation
├── githistory/
│   ├── scanner.go                   # Git log analysis, blob detection
│   ├── analyzer.go                  # Binary detection, size calculation
│   ├── safety.go                    # Backup checks, confirmation logic
│   ├── executor.go                  # git-filter-repo execution
│   └── reporter.go                  # Results formatting
└── githistory_test.go               # Comprehensive tests
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Day 1)

#### 1.1 Domain Types

**File:** `internal/domain/githistory_types.go`

```go
// GitHistoryCleanerConfig - Configuration for git history cleaning
type GitHistoryCleanerConfig struct {
    MaxFileSizeMB      int                    // Files larger than this are candidates
    TargetPaths        []string               // Specific paths to scan
    ExcludePaths       []string               // Paths to ignore
    BinaryExtensions   []string               // Extensions considered binary
    MinBlobSizeKB      int                    // Minimum blob size to consider
    AutoDetectBinaries bool                   // Use heuristics to detect binaries
    DryRun             bool                   // Simulate without changes
}

// BinaryFile represents a binary file found in history
type BinaryFile struct {
    Path        string    // File path in repo
    SizeBytes   int64     // Current size
    CommitHash  string    // First commit that added it
    CommitDate  time.Time // When it was added
    Author      string    // Who added it
    BlobHash    string    // Git blob hash
    IsDeleted   bool      // Whether currently in HEAD
}

// HistoryRewriteResult - Result of git history rewrite
type HistoryRewriteResult struct {
    FilesRemoved   []BinaryFile
    BytesRemoved   int64
    CommitsChanged int
    OldRepoSize    int64
    NewRepoSize    int64
    BackupCreated  bool
    BackupPath     string
}
```

#### 1.2 Safety Types

**File:** `internal/domain/githistory_safety.go`

```go
// SafetyLevel determines how cautious to be
type SafetyLevel int
const (
    SafetyLevelDryRun SafetyLevel = iota      // Only analyze, never modify
    SafetyLevelBackupRequired                 // Require backup before modify
    SafetyLevelForce                          // Skip backup with --force flag
)

// RewriteConsequences describes what will happen
type RewriteConsequences struct {
    WillRewriteHistory   bool
    WillRequireForcePush bool
    WillAffectOtherClones bool
    CommitsAffected      int
    EstimatedTime        time.Duration
    RequiresBackup       bool
}
```

#### 1.3 Enum Integration

Add to existing enums:

**File:** `internal/domain/type_safe_enums.go`

```go
// GitHistoryMode - Mode of operation
type GitHistoryMode int
const (
    GitHistoryModeAnalyze GitHistoryMode = iota  // Only report findings
    GitHistoryModeDryRun                         // Show what would be done
    GitHistoryModeExecute                        // Actually rewrite history
)

// BinaryDetectionMethod - How to detect binaries
type BinaryDetectionMethod int
const (
    DetectionMethodExtension BinaryDetectionMethod = iota
    DetectionMethodContent
    DetectionMethodBoth
)
```

### Phase 2: Scanner Implementation (Day 1-2)

#### 2.1 Git Scanner

**File:** `internal/cleaner/githistory/scanner.go`

```go
package githistory

// Scanner analyzes git history for binary files
type Scanner struct {
    repoPath string
    verbose  bool
}

// NewScanner creates a scanner for the given repository
func NewScanner(repoPath string, verbose bool) *Scanner

// ScanOptions configures what to look for
type ScanOptions struct {
    MaxFileSizeMB     int
    MinBlobSizeKB     int
    TargetPaths       []string
    ExcludePaths      []string
    DetectBinaries    bool
    LookbackCommits   int  // 0 = all history
}

// Scan returns all binary files found in history
func (s *Scanner) Scan(ctx context.Context, opts ScanOptions) ([]domain.BinaryFile, error)

// ScanResults contains the analysis
type ScanResults struct {
    Files           []domain.BinaryFile
    TotalSize       int64
    LargestFile     domain.BinaryFile
    OldestCommit    time.Time
    RecentCommits   int  // Files added in last 30 days
}
```

**Key Implementation Details:**

1. **Use `git rev-list` + `git cat-file`** for efficient scanning
2. **Parse `git log --numstat`** for size information
3. **Content-based detection**: Check for null bytes in first 8000 bytes
4. **Extension-based detection**: Common binary extensions

```go
// Default binary extensions for Go projects
var DefaultBinaryExtensions = []string{
    // Go build outputs
    "",           // No extension binaries (common in Go)
    ".exe",       // Windows executables
    ".dll",       // Windows libraries
    ".so",        // Linux shared libraries
    ".dylib",     // macOS libraries

    // Build artifacts
    ".o",         // Object files
    ".a",         // Static libraries
    ".out",       // Generic output

    // Archives
    ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z",

    // Media
    ".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov",

    // Documents
    ".pdf", ".doc", ".docx", ".xls", ".xlsx",

    // Other
    ".bin", ".dat", ".db", ".sqlite",
}
```

#### 2.2 Size Analyzer

**File:** `internal/cleaner/githistory/analyzer.go`

```go
// Analyzer calculates impact of removing files
type Analyzer struct {
    scanner *Scanner
}

// AnalyzeImpact determines the effect of removing specific files
func (a *Analyzer) AnalyzeImpact(
    ctx context.Context,
    files []domain.BinaryFile,
) (*RewriteImpact, error)

type RewriteImpact struct {
    CurrentRepoSizeMB    int64
    EstimatedNewSizeMB   int64
    SpaceReclaimedMB     int64
    CommitsToRewrite     int
    ObjectsToPrune       int
    EstimatedTime        time.Duration
}
```

### Phase 3: Safety & Confirmation (Day 2)

#### 3.1 Safety Checker

**File:** `internal/cleaner/githistory/safety.go`

```go
// SafetyChecker validates preconditions
 type SafetyChecker struct {
    repoPath string
}

// PreFlightChecks runs all safety checks
func (sc *SafetyChecker) PreFlightChecks(ctx context.Context) (*SafetyReport, error)

type SafetyReport struct {
    IsGitRepo           bool
    IsBareRepo          bool
    HasUncommittedChanges bool
    HasUnpushedCommits    bool
    HasRemote           bool
    RemoteURL           string
    CurrentBranch       string
    CanCreateBackup     bool
    BackupPath          string
    Warnings            []string
    Blockers            []string  // Must resolve before proceeding
}

// CheckRemoteStatus warns about shared history
func (sc *SafetyChecker) CheckRemoteStatus() (*RemoteStatus, error)

type RemoteStatus struct {
    HasRemote       bool
    RemoteName      string
    RemoteURL       string
    IsSharedRepo    bool  // Detect GitHub, GitLab, etc.
    BranchesOnRemote []string
    WarningMessage   string
}
```

#### 3.2 Confirmation Flow

**File:** `cmd/clean-wizard/commands/githistory.go` (subcommand)

```go
// createConfirmationPrompt builds the TUI confirmation
func createConfirmationPrompt(
    consequences *domain.RewriteConsequences,
    report *githistory.SafetyReport,
) *huh.Form {
    // Multi-step confirmation:
    // 1. Show what will happen
    // 2. Require typing repo name to confirm
    // 3. Show force-push warning if applicable
    // 4. Final yes/no
}
```

**Confirmation UI Flow:**

```
┌─────────────────────────────────────────────────────────────┐
│  ⚠️  WARNING: DESTRUCTIVE OPERATION                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  You are about to rewrite git history for:                  │
│  /Users/dev/projects/myapp                                  │
│                                                             │
│  This will:                                                 │
│    • Remove 12 binary files from history                    │
│    • Rewrite 47 commits                                     │
│    • Reduce repo size from 45MB to 2MB                      │
│    • Require force-push to remote                           │
│                                                             │
│  ⚠️  Team Impact:                                           │
│    Other team members will need to reclone or reset.        │
│    Coordinate with your team before proceeding.             │
│                                                             │
│  Type the repository name to confirm: [myapp      ]         │
│                                                             │
│  [ ] I have created a backup                                │
│  [ ] I have coordinated with my team                        │
│  [ ] I understand this cannot be undone                     │
│                                                             │
│         [Cancel]  [I Understand, Proceed]                   │
└─────────────────────────────────────────────────────────────┘
```

### Phase 4: Executor Implementation (Day 2-3)

#### 4.1 Execution Engine

**File:** `internal/cleaner/githistory/executor.go`

```go
// Executor runs the git-filter-repo command
type Executor struct {
    repoPath string
    verbose  bool
    dryRun   bool
}

// ExecutionOptions configures the rewrite
 type ExecutionOptions struct {
    FilesToRemove   []domain.BinaryFile
    CreateBackup    bool
    BackupPath      string
    StripBlobsBiggerThanMB int  // Alternative: size-based removal
}

// Execute runs the history rewrite
func (e *Executor) Execute(
    ctx context.Context,
    opts ExecutionOptions,
) (*domain.HistoryRewriteResult, error)
```

**Tool Selection Strategy:**

```go
// detectTool determines which tool to use
func detectTool() (HistoryRewriteTool, error) {
    // 1. Prefer git-filter-repo (fast, Python-based)
    //    Check: git filter-repo --version

    // 2. Fallback to git-filter-branch (slower, built-in)
    //    Check: git filter-branch --help

    // 3. Error if neither available
    //    Provide installation instructions
}
```

**git-filter-repo Command Generation:**

```go
// buildFilterRepoArgs creates arguments for git-filter-repo
func buildFilterRepoArgs(files []domain.BinaryFile) []string {
    args := []string{
        "--force",  // Required for non-fresh repos
        "--partial", // Allow running on dirty repos
    }

    // Add file-specific filters
    for _, file := range files {
        args = append(args, "--path", file.Path)
        args = append(args, "--invert-paths") // Remove, not keep
    }

    return args
}

// Alternative: Size-based filtering
func buildSizeFilterArgs(maxSizeMB int) []string {
    return []string{
        "--strip-blobs-bigger-than", fmt.Sprintf("%dM", maxSizeMB),
    }
}
```

**git-filter-branch Fallback:**

```go
// buildFilterBranchCommand creates the filter-branch command
func buildFilterBranchCommand(files []domain.BinaryFile) string {
    // Construct index-filter to remove files
    paths := make([]string, len(files))
    for i, f := range files {
        paths[i] = f.Path
    }

    return fmt.Sprintf(
        "git filter-branch --force --index-filter "+
        "'git rm --cached --ignore-unmatch %s' "+
        "--prune-empty --tag-name-filter cat -- --all",
        strings.Join(paths, " "),
    )
}
```

### Phase 5: Main Cleaner Integration (Day 3)

#### 5.1 GitHistoryCleaner

**File:** `internal/cleaner/githistory.go`

```go
package cleaner

// GitHistoryCleaner removes binaries from git history
type GitHistoryCleaner struct {
    config   domain.GitHistoryCleanerConfig
    verbose  bool
    dryRun   bool
}

// NewGitHistoryCleaner creates a new instance
func NewGitHistoryCleaner(
    config domain.GitHistoryCleanerConfig,
    verbose bool,
    dryRun bool,
) *GitHistoryCleaner

// Implement Cleaner interface
func (c *GitHistoryCleaner) Name() string {
    return "Git History"
}

func (c *GitHistoryCleaner) IsAvailable(ctx context.Context) bool {
    // Check if git is installed
    // Check if we're in a git repo
    _, err := exec.LookPath("git")
    if err != nil {
        return false
    }

    // Check if current directory is git repo
    cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
    return cmd.Run() == nil
}

func (c *GitHistoryCleaner) Clean(
    ctx context.Context,
) result.Result[domain.CleanResult] {
    // 1. Scan for binary files
    scanner := githistory.NewScanner(c.config.RepoPath, c.verbose)
    files, err := scanner.Scan(ctx, scanOptions)
    if err != nil {
        return result.Error[domain.CleanResult](err)
    }

    if len(files) == 0 {
        return result.Success(domain.CleanResult{
            Message: "No binary files found in history",
        })
    }

    // 2. Analyze impact
    analyzer := githistory.NewAnalyzer(scanner)
    impact, err := analyzer.AnalyzeImpact(ctx, files)
    if err != nil {
        return result.Error[domain.CleanResult](err)
    }

    // 3. Safety checks
    safety := githistory.NewSafetyChecker(c.config.RepoPath)
    report, err := safety.PreFlightChecks(ctx)
    if err != nil {
        return result.Error[domain.CleanResult](err)
    }

    if len(report.Blockers) > 0 {
        return result.Error[domain.CleanResult](
            fmt.Errorf("safety blockers: %v", report.Blockers),
        )
    }

    // 4. If dry-run, return analysis
    if c.dryRun {
        return c.createDryRunResult(files, impact)
    }

    // 5. Execute (requires explicit confirmation via TUI)
    executor := githistory.NewExecutor(c.config.RepoPath, c.verbose, c.dryRun)
    result, err := executor.Execute(ctx, executionOptions)
    if err != nil {
        return result.Error[domain.CleanResult](err)
    }

    return c.createSuccessResult(result)
}
```

#### 5.2 Registry Registration

**File:** `internal/cleaner/registry_factory.go`

```go
func DefaultRegistry() *Registry {
    registry := NewRegistry()

    // ... existing cleaners ...

    // Register Git History cleaner
    registry.Register("githistory", NewGitHistoryCleaner(
        domain.GitHistoryCleanerConfig{
            MaxFileSizeMB:      10,
            AutoDetectBinaries: true,
            MinBlobSizeKB:      50,
            BinaryExtensions:   domain.DefaultBinaryExtensions,
        },
        false,
        false,
    ))

    return registry
}
```

### Phase 6: CLI Integration (Day 3)

#### 6.1 Subcommand Structure

**File:** `cmd/clean-wizard/commands/githistory.go`

```go
package commands

func NewGitHistoryCommand() *cobra.Command {
    var (
        dryRun          bool
        verbose         bool
        maxSizeMB       int
        repoPath        string
        force           bool // Skip backup requirement
        detectBinaries  bool
    )

    cmd := &cobra.Command{
        Use:   "git-history [path]",
        Short: "Remove binary files from git history",
        Long: `Scans git history for binary files and removes them using git-filter-repo.

⚠️  WARNING: This rewrites git history and requires force-push.
    Always backup your repository first.
    Coordinate with your team before running on shared repos.

Examples:
  # Analyze only (dry-run)
  clean-wizard git-history --dry-run

  # Remove binaries larger than 10MB
  clean-wizard git-history --max-size 10

  # Target specific repository
  clean-wizard git-history /path/to/repo --dry-run

  # Force execution without backup check
  clean-wizard git-history --force`,

        RunE: func(cmd *cobra.Command, args []string) error {
            // Determine repo path
            path := "."
            if len(args) > 0 {
                path = args[0]
            }

            // Run analysis
            return runGitHistoryClean(path, dryRun, verbose, maxSizeMB, force, detectBinaries)
        },
    }

    cmd.Flags().BoolVar(&dryRun, "dry-run", true, "Analyze only, don't modify (default)")
    cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output")
    cmd.Flags().IntVar(&maxSizeMB, "max-size", 10, "Target files larger than this (MB)")
    cmd.Flags().BoolVar(&force, "force", false, "Skip backup requirement (dangerous)")
    cmd.Flags().BoolVar(&detectBinaries, "detect-binaries", true, "Use heuristics to detect binaries")

    return cmd
}
```

#### 6.2 Root Command Registration

**File:** `cmd/clean-wizard/main.go`

```go
func main() {
    rootCmd := commands.NewRootCommand()

    // Add all commands
    rootCmd.AddCommand(commands.NewCleanCommand())
    rootCmd.AddCommand(commands.NewScanCommand())
    rootCmd.AddCommand(commands.NewInitCommand())
    rootCmd.AddCommand(commands.NewProfileCommand())
    rootCmd.AddCommand(commands.NewConfigCommand())
    rootCmd.AddCommand(commands.NewGitHistoryCommand()) // NEW

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### Phase 7: Testing (Day 3-4)

#### 7.1 Test Structure

**File:** `internal/cleaner/githistory_test.go`

```go
func TestGitHistoryCleaner_IsAvailable(t *testing.T) {
    tests := []struct {
        name       string
        gitInstalled bool
        isGitRepo  bool
        want       bool
    }{
        {
            name:       "git installed in git repo",
            gitInstalled: true,
            isGitRepo:  true,
            want:       true,
        },
        {
            name:       "git installed not in repo",
            gitInstalled: true,
            isGitRepo:  false,
            want:       false,
        },
        {
            name:       "git not installed",
            gitInstalled: false,
            isGitRepo:  false,
            want:       false,
        },
    }
    // ... test implementation
}

func TestScanner_Scan(t *testing.T) {
    // Create temp git repo with known binary files
    // Verify scanner finds them
}

func TestAnalyzer_AnalyzeImpact(t *testing.T) {
    // Test impact calculation accuracy
}

func TestSafetyChecker_PreFlightChecks(t *testing.T) {
    // Test all safety conditions
}
```

#### 7.2 Integration Tests

**File:** `tests/integration/githistory_integration_test.go`

```go
func TestGitHistory_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup: Create temp repo with binaries
    repo := createTestRepoWithBinaries(t)
    defer cleanup(t, repo)

    t.Run("dry_run_reports_binaries", func(t *testing.T) {
        cleaner := createTestCleaner(repo, true) // dry-run
        result := cleaner.Clean(context.Background())

        require.True(t, result.IsSuccess())
        assert.Greater(t, result.Value.ItemsFound, 0)
        assert.Equal(t, 0, result.Value.ItemsRemoved) // Nothing actually removed
    })

    t.Run("actual_clean_removes_binaries", func(t *testing.T) {
        cleaner := createTestCleaner(repo, false)
        result := cleaner.Clean(context.Background())

        require.True(t, result.IsSuccess())
        assert.Greater(t, result.Value.ItemsRemoved, 0)
    })
}
```

#### 7.3 BDD Tests

**File:** `tests/bdd/git_history_cleaning.feature`

```gherkin
Feature: Git History Binary Removal

  Background:
    Given a git repository with compiled binaries
    And the binaries are committed to history

  Scenario: Dry-run identifies binaries without removing them
    When I run "git-history" with "--dry-run"
    Then I should see a list of binary files
    And the repository should be unchanged
    And I should see estimated space savings

  Scenario: Safety checks prevent execution with uncommitted changes
    Given there are uncommitted changes
    When I attempt to clean history
    Then the operation should be blocked
    And I should see a warning about uncommitted changes

  Scenario: Team coordination warning for shared repos
    Given the repository has a remote origin
    When I run the analysis
    Then I should see a warning about force-push requirements
    And I should see instructions for team coordination

  Scenario: Successful removal with confirmation
    Given I have created a backup
    And I confirm the operation
    When I execute the history rewrite
    Then the binaries should be removed from history
    And the repository size should be reduced
    And I should see instructions for force-push
```

---

## Configuration Schema

**File:** `schemas/config.schema.json` (update)

```json
{
  "gitHistory": {
    "type": "object",
    "description": "Git history cleaning configuration",
    "properties": {
      "enabled": {
        "type": "boolean",
        "default": true
      },
      "maxFileSizeMB": {
        "type": "integer",
        "description": "Files larger than this are candidates for removal",
        "default": 10,
        "minimum": 1
      },
      "minBlobSizeKB": {
        "type": "integer",
        "description": "Minimum blob size to consider",
        "default": 50,
        "minimum": 1
      },
      "autoDetectBinaries": {
        "type": "boolean",
        "description": "Use heuristics to detect binary content",
        "default": true
      },
      "targetPaths": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Specific paths to scan (empty = all)"
      },
      "excludePaths": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Paths to ignore"
      },
      "binaryExtensions": {
        "type": "array",
        "items": { "type": "string" },
        "description": "File extensions considered binary"
      }
    }
  }
}
```

---

## Error Handling

### Error Codes

**File:** `internal/pkg/errors/error_codes.go` (add)

```go
const (
    // Git History errors
    ErrGitNotInstalled      ErrorCode = "GIT_NOT_INSTALLED"
    ErrNotAGitRepo          ErrorCode = "NOT_A_GIT_REPO"
    ErrUncommittedChanges   ErrorCode = "UNCOMMITTED_CHANGES"
    ErrNoFilterTool         ErrorCode = "NO_FILTER_TOOL"
    ErrHistoryRewriteFailed ErrorCode = "HISTORY_REWRITE_FAILED"
    ErrSafetyCheckFailed    ErrorCode = "SAFETY_CHECK_FAILED"
    ErrNoBinariesFound      ErrorCode = "NO_BINARIES_FOUND"
    ErrBackupFailed         ErrorCode = "BACKUP_FAILED"
)
```

### Error Messages

```go
var GitHistoryErrorMessages = map[ErrorCode]string{
    ErrGitNotInstalled: "git is not installed or not in PATH",
    ErrNotAGitRepo:     "not a git repository (or any parent)",
    ErrUncommittedChanges:
        "You have uncommitted changes. " +
        "Commit or stash them before rewriting history.",
    ErrNoFilterTool:
        "Neither git-filter-repo nor git-filter-branch is available.\n" +
        "Install git-filter-repo: https://github.com/newren/git-filter-repo",
    ErrHistoryRewriteFailed: "Failed to rewrite history: %v",
    ErrSafetyCheckFailed: "Safety check failed: %v",
    ErrNoBinariesFound: "No binary files found in history",
    ErrBackupFailed: "Failed to create backup: %v",
}
```

---

## Success Metrics

| Metric                    | Target              | Measurement                    |
| ------------------------- | ------------------- | ------------------------------ |
| Binary Detection Accuracy | >95%                | Test repos with known binaries |
| False Positive Rate       | <5%                 | Manual review of flagged files |
| Safety Check Coverage     | 100%                | All blockers must be detected  |
| Execution Time            | <30s for 100MB repo | Benchmark tests                |
| Test Coverage             | >85%                | go test -cover                 |

---

## Risks and Mitigations

| Risk                 | Likelihood | Impact   | Mitigation                                     |
| -------------------- | ---------- | -------- | ---------------------------------------------- |
| Accidental data loss | Low        | Critical | Multi-step confirmation, backup requirement    |
| Team disruption      | Medium     | High     | Force-push warnings, coordination instructions |
| Tool not available   | Medium     | Medium   | Clear error messages, install instructions     |
| Large repo timeout   | Medium     | Medium   | Progress indicators, resumable operations      |
| False positives      | Medium     | Low      | Manual review, exclude patterns                |

---

## Documentation

### User-Facing Documentation

**File:** `docs/GIT_HISTORY_CLEANER.md`

````markdown
# Git History Binary Cleaner

## When to Use This Tool

- Repository is unexpectedly large
- `go build` outputs were committed accidentally
- CI artifacts or logs in history
- Vendor directories committed before go modules

## Safety Checklist

Before running:

- [ ] Created backup: `git clone --mirror origin/repo.git backup.git`
- [ ] Informed team members
- [ ] No uncommitted changes
- [ ] All important branches pushed
- [ ] Understand force-push is required

## Step-by-Step Guide

1. **Analyze** (always start here):
   ```bash
   clean-wizard git-history --dry-run
   ```
````

2. **Review findings** - Check the list of files to be removed

3. **Create backup**:

   ```bash
   git clone --mirror . ../myrepo-backup.git
   ```

4. **Execute** (requires confirmation):

   ```bash
   clean-wizard git-history
   ```

5. **Force push** (if remote exists):

   ```bash
   git push --force-with-lease origin main
   ```

6. **Notify team** - Others need to reset:
   ```bash
   git fetch origin
   git reset --hard origin/main  # Or reclone
   ```

````

### Implementation Documentation

**File:** `internal/cleaner/githistory/README.md`

```markdown
# Git History Cleaner Internal Documentation

## Architecture

See PLAN_GIT_HISTORY_CLEANER.md for full specification.

### Key Design Decisions

1. **Tool Preference**: git-filter-repo > git-filter-branch
   - filter-repo is 10-100x faster
   - Better safety checks built-in
   - Actively maintained

2. **Safety-First**: All operations require explicit confirmation
   - Dry-run is default
   - Multi-step confirmation UI
   - Backup check (unless --force)

3. **Detection Strategy**: Extension + Content
   - Extensions catch obvious binaries
   - Content analysis catches extension-less Go binaries
````

---

## Future Enhancements

| Feature                    | Priority | Description                               |
| -------------------------- | -------- | ----------------------------------------- |
| BFG Repo-Cleaner support   | Medium   | Alternative to filter-repo for Java repos |
| Interactive file selection | Medium   | TUI to pick specific files                |
| Incremental cleaning       | Low      | Clean only recent history                 |
| Automatic backup           | Low      | Built-in backup creation                  |
| LFS migration              | Low      | Convert binaries to Git LFS               |
| GitHub API integration     | Low      | Handle PRs on rewritten history           |

---

## Appendix: Comparison with Existing Tools

| Tool              | Pros                      | Cons                    | Our Approach          |
| ----------------- | ------------------------- | ----------------------- | --------------------- |
| git-filter-repo   | Fast, safe, maintained    | Requires Python install | Auto-detect, fallback |
| git-filter-branch | Built-in, no install      | Slow, deprecated        | Fallback only         |
| BFG Repo-Cleaner  | Very fast for large repos | Java dependency         | Future addition       |
| Manual rebase     | Full control              | Tedious, error-prone    | Not supported         |

---

_This plan follows the architectural patterns established in clean-wizard, maintaining type safety, comprehensive testing, and safety-first design principles._
