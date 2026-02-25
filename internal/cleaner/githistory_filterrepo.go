package cleaner

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// FilterRepoProvider represents a way to run git-filter-repo.
type FilterRepoProvider int

const (
	// FilterRepoNone indicates no provider is available.
	FilterRepoNone FilterRepoProvider = iota
	// FilterRepoSystem indicates git-filter-repo is installed system-wide.
	FilterRepoSystem
	// FilterRepoNix indicates we can use nix to run git-filter-repo.
	FilterRepoNix
)

// String returns the provider name.
func (p FilterRepoProvider) String() string {
	switch p {
	case FilterRepoSystem:
		return "system"
	case FilterRepoNix:
		return "nix"
	default:
		return "none"
	}
}

// filterRepoDetector caches the detected provider.
type filterRepoDetector struct {
	once     sync.Once
	provider FilterRepoProvider
}

var detector filterRepoDetector

// DetectFilterRepoProvider detects how git-filter-repo can be run.
// Priority: system install > nix > none.
func DetectFilterRepoProvider() FilterRepoProvider {
	detector.once.Do(func() {
		detector.provider = detectProvider()
	})

	return detector.provider
}

// ResetDetector resets the cached detector (for testing).
func ResetDetector() {
	detector = filterRepoDetector{}
}

func detectProvider() FilterRepoProvider {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, check if git-filter-repo is available as a git subcommand
	if isSystemInstallAvailable(ctx) {
		return FilterRepoSystem
	}

	// Second, check if nix is available and git-filter-repo is in nixpkgs
	if isNixAvailable(ctx) {
		return FilterRepoNix
	}

	return FilterRepoNone
}

func isSystemInstallAvailable(ctx context.Context) bool {
	// git filter-repo --version (as git subcommand)
	cmd := exec.CommandContext(ctx, "git", "filter-repo", "--version")

	return cmd.Run() == nil
}

func isNixAvailable(ctx context.Context) bool {
	// Check if nix command exists
	if _, err := exec.LookPath("nix"); err != nil {
		return false
	}

	// Verify nix can access nixpkgs (without actually running git-filter-repo)
	// Use nix eval to check if the package exists (faster than nix run)
	cmd := exec.CommandContext(ctx, "nix", "eval", "--raw", "nixpkgs#git-filter-repo.name")

	return cmd.Run() == nil
}

// BuildFilterRepoCommand builds the command to run git-filter-repo with the given args.
func BuildFilterRepoCommand(ctx context.Context, args []string) *exec.Cmd {
	provider := DetectFilterRepoProvider()

	switch provider {
	case FilterRepoNix:
		// nix run nixpkgs#git-filter-repo -- <args>
		nixArgs := []string{"run", "nixpkgs#git-filter-repo", "--"}
		nixArgs = append(nixArgs, args...)

		return exec.CommandContext(ctx, "nix", nixArgs...)
	default:
		// System install or fallback: git filter-repo <args>
		gitArgs := append([]string{"filter-repo"}, args...)

		return exec.CommandContext(ctx, "git", gitArgs...)
	}
}

// GetInstallHint returns a hint for how to install git-filter-repo.
func GetInstallHint() string {
	provider := DetectFilterRepoProvider()

	switch provider {
	case FilterRepoNix:
		return "Using nix to run git-filter-repo"
	case FilterRepoSystem:
		return "git-filter-repo is installed system-wide"
	default:
		// Check what's available to give a good hint
		if _, err := exec.LookPath("nix"); err == nil {
			return "Run 'nix run nixpkgs#git-filter-repo -- --help' to verify nix can access it, or install with: brew install git-filter-repo"
		}

		if _, err := exec.LookPath("brew"); err == nil {
			return "Install with: brew install git-filter-repo"
		}

		return "Install with: pip install git-filter-repo or see https://github.com/newren/git-filter-repo"
	}
}
