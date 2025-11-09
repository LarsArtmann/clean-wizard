package adapters

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// NixAdapter wraps Nix package manager operations
type NixAdapter struct {
	timeout time.Duration
	retries int
}

// NewNixAdapter creates Nix adapter with configuration
func NewNixAdapter(timeout time.Duration, retries int) *NixAdapter {
	return &NixAdapter{
		timeout: timeout,
		retries: retries,
	}
}

// ListGenerations lists Nix generations with domain types
func (n *NixAdapter) ListGenerations(ctx context.Context) result.Result[[]domain.NixGeneration] {
	if !n.IsAvailable(ctx) {
		return result.Err[[]domain.NixGeneration](fmt.Errorf("nix not available"))
	}

	cmd := exec.CommandContext(ctx, "nix-env", "--list-generations", "--profile", "/nix/var/nix/profiles/default")
	output, err := cmd.Output()
	if err != nil {
		return result.Err[[]domain.NixGeneration](fmt.Errorf("failed to list generations: %w", err))
	}

	var generations []domain.NixGeneration
	lines := strings.SplitSeq(string(output), "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		gen, err := n.ParseGeneration(line)
		if err != nil {
			return result.Err[[]domain.NixGeneration](fmt.Errorf("failed to parse generation: %w", err))
		}

		generations = append(generations, gen)
	}

	return result.Ok(generations)
}

// GetStoreSize returns Nix store size with domain types
func (n *NixAdapter) GetStoreSize(ctx context.Context) result.Result[domain.CleanResult] {
	cmd := exec.CommandContext(ctx, "du", "-sb", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get store size: %w", err))
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return result.Err[domain.CleanResult](fmt.Errorf("invalid du output: %s", string(output)))
	}

	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to parse size: %w", err))
	}

	return result.Ok(domain.CleanResult{
		ItemsRemoved: 0,
		FreedBytes:   size,
		ItemsFailed:  0,
		Strategy:     "STORE_SIZE",
		CleanTime:    time.Since(time.Now()),
	})
}

// CollectGarbage removes old Nix generations with domain types
func (n *NixAdapter) CollectGarbage(ctx context.Context) result.Result[domain.CleanResult] {
	cmd := exec.CommandContext(ctx, "nix-collect-garbage", "-d")
	err := cmd.Run()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to collect garbage: %w", err))
	}

	// Calculate bytes freed (mock implementation)
	return result.Ok(domain.CleanResult{
		ItemsRemoved: 1,
		FreedBytes:   0, // Calculate from before/after
		ItemsFailed:  0,
		Strategy:     "NIX_GC",
		CleanTime:    time.Since(time.Now()),
	})
}

// RemoveGeneration removes specific Nix generation
func (n *NixAdapter) RemoveGeneration(ctx context.Context, genID int) result.Result[domain.CleanResult] {
	cmd := exec.CommandContext(ctx, "nix-env", "--delete-generations", fmt.Sprintf("%d", genID))
	err := cmd.Run()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to remove generation %d: %w", genID, err))
	}

	return result.Ok(domain.CleanResult{
		ItemsRemoved: 1,
		FreedBytes:   0, // Calculate from before/after
		ItemsFailed:  0,
		Strategy:     "REMOVE_GENERATION",
		CleanTime:    time.Since(time.Now()),
	})
}

// ParseGeneration parses generation line from nix-env output
func (n *NixAdapter) ParseGeneration(line string) (domain.NixGeneration, error) {
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return domain.NixGeneration{}, fmt.Errorf("invalid generation line: %s", line)
	}

	pathParts := strings.Split(fields[0], "-")
	if len(pathParts) < 2 {
		return domain.NixGeneration{}, fmt.Errorf("invalid generation path: %s", fields[0])
	}

	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		return domain.NixGeneration{}, fmt.Errorf("invalid generation ID: %s", pathParts[len(pathParts)-1])
	}

	// Simple date parsing
	date := time.Now()
	if len(fields) > 1 && strings.Contains(fields[1], "-") {
		parts := strings.Split(fields[1], "-")
		if len(parts) == 3 {
			if parsed, err := time.Parse("2006-01-02", strings.Join(parts, "-")); err == nil {
				date = parsed
			}
		}
	}

	return domain.NixGeneration{
		ID:      id,
		Path:    fields[0],
		Date:    date,
		Current: strings.Contains(line, "current"),
	}, nil
}

// IsAvailable checks if Nix is available and accessible
func (n *NixAdapter) IsAvailable(ctx context.Context) bool {
	// First check if nix command exists
	versionCmd := exec.CommandContext(ctx, "nix", "--version")
	if err := versionCmd.Run(); err != nil {
		return false
	}

	// Then check if we can access profiles (the actual operation we need)
	listCmd := exec.CommandContext(ctx, "nix-env", "--list-generations", "--profile", "/nix/var/nix/profiles/default")
	err := listCmd.Run()
	return err == nil
}

