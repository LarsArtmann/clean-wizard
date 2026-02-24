package cleaner

import (
	"context"
	"os/exec"
	"testing"
)

func TestDetectFilterRepoProvider(t *testing.T) {
	// Reset detector for clean test
	ResetDetector()

	provider := DetectFilterRepoProvider()

	// Provider should be one of the valid values
	if provider < FilterRepoNone || provider > FilterRepoNix {
		t.Errorf("Invalid provider: %d", provider)
	}

	// If nix is available, provider should not be FilterRepoNone
	if _, err := exec.LookPath("nix"); err == nil {
		if provider == FilterRepoNone {
			t.Error("Expected non-none provider when nix is available")
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
		{FilterRepoProvider(99), "none"},
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
