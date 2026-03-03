package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// FilterRepoProvider represents the provider type for git-filter-repo.
type FilterRepoProvider int

const (
	FilterRepoSystem FilterRepoProvider = iota
	FilterRepoNix
)

// DetectFilterRepoProvider detects which provider is available.
func DetectFilterRepoProvider() FilterRepoProvider {
	// Check if nix is available
	if _, err := exec.LookPath("nix"); err == nil {
		return FilterRepoNix
	}
	return FilterRepoSystem
}

// LogFilterRepoCommand logs the git-filter-repo command if verbose.
func LogFilterRepoCommand(verbose bool, args []string) {
	if !verbose {
		return
	}

	provider := DetectFilterRepoProvider()
	if provider == FilterRepoNix {
		fmt.Printf("Running: nix run nixpkgs#git-filter-repo -- %s\n", strings.Join(args, " "))
	} else {
		fmt.Printf("Running: git filter-repo %s\n", strings.Join(args, " "))
	}
}
