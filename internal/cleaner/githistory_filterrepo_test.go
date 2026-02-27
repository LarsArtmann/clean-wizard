package cleaner

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestDetectFilterRepoProvider(t *testing.T) {
	// Reset detector for clean test
	ResetDetector()

	provider := DetectFilterRepoProvider()

	// Provider should be one of the valid values
	if provider < FilterRepoNone || provider > FilterRepoNix {
		t.Errorf("Invalid provider: %d", provider)
	}

	// If git-filter-repo is available as git subcommand, provider should be FilterRepoSystem
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if isSystemInstallAvailable(ctx) {
		if provider != FilterRepoSystem {
			t.Errorf(
				"Expected FilterRepoSystem when git filter-repo is available, got %d",
				provider,
			)
		}
	}

	// If nix can access nixpkgs#git-filter-repo (and system install not available), provider should be FilterRepoNix
	// Note: We don't just check if nix binary exists, we check if nix can actually access git-filter-repo
	if _, err := exec.LookPath("nix"); err == nil {
		// Only assert non-none if we can actually verify nixpkgs access
		cmd := exec.CommandContext(ctx, "nix", "eval", "--raw", "nixpkgs#git-filter-repo.name")
		if cmd.Run() == nil && !isSystemInstallAvailable(ctx) {
			if provider != FilterRepoNix {
				t.Errorf(
					"Expected FilterRepoNix when nix can access git-filter-repo, got %d",
					provider,
				)
			}
		}
	}
}

func TestFilterRepoProvider_String(t *testing.T) {
	tests := []struct {
		provider FilterRepoProvider
		want     string
	}{
		{FilterRepoNone, "none"},
		{FilterRepoSystem, "system"},
		{FilterRepoNix, "nix"},
		{FilterRepoProvider(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.provider.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildFilterRepoCommand(t *testing.T) {
	ResetDetector()

	ctx := context.Background()

	// Test that command is built without error
	cmd := BuildFilterRepoCommand(ctx, []string{"--version"})
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// The command path should be either "nix" or "git" depending on provider
	provider := DetectFilterRepoProvider()
	switch provider {
	case FilterRepoNix:
		if cmd.Path != "nix" && cmd.Args[0] != "nix" {
			t.Errorf("Expected nix command for Nix provider, got: %s", cmd.Path)
		}
	case FilterRepoSystem:
		if cmd.Path != "git" && cmd.Args[0] != "git" {
			t.Errorf("Expected git command for System provider, got: %s", cmd.Path)
		}
	}
}

func TestGetInstallHint(t *testing.T) {
	ResetDetector()

	hint := GetInstallHint()
	if hint == "" {
		t.Error("Expected non-empty install hint")
	}
}
