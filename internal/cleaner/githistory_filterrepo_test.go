package cleaner

import (
	"context"
	"testing"
	"time"
)

func TestDetectFilterRepoProvider(t *testing.T) {
	t.Parallel()

	// Reset detector for clean test
	ResetDetector()

	provider := DetectFilterRepoProvider()

	// Provider should be one of the valid values
	if provider < FilterRepoNone || provider > FilterRepoNix {
		t.Errorf("Invalid provider: %d", provider)
	}

	// Verify the provider matches what we can detect
	// Note: We don't assert specific values because the test environment varies
	// Instead, we verify that the detection logic is consistent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	systemAvailable := isSystemInstallAvailable(ctx)
	nixAvailable := isNixAvailable(ctx)

	// Verify detection priority: system > nix > none
	if systemAvailable && provider != FilterRepoSystem {
		t.Errorf("System install available but provider is %s, expected system", provider)
	}

	if !systemAvailable && nixAvailable && provider != FilterRepoNix {
		t.Errorf("Nix available (no system) but provider is %s, expected nix", provider)
	}

	if !systemAvailable && !nixAvailable && provider != FilterRepoNone {
		t.Errorf("No provider available but provider is %s, expected none", provider)
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
