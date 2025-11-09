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

// GetStoreSize returns Nix store size as bytes
func (n *NixAdapter) GetStoreSize(ctx context.Context) result.Result[int64] {
	cmd := exec.CommandContext(ctx, "du", "-sb", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		return result.Err[int64](fmt.Errorf("failed to get store size: %w", err))
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return result.Err[int64](fmt.Errorf("invalid du output: %s", string(output)))
	}

	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return result.Err[int64](fmt.Errorf("failed to parse size: %w", err))
	}

	return result.Ok(size)
}

// CollectGarbage removes old Nix generations with real byte calculation
func (n *NixAdapter) CollectGarbage(ctx context.Context) result.Result[domain.CleanResult] {
	// Get store size before garbage collection
	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get pre-gc store size: %w", err))
	}

	// Run actual nix-collect-garbage command
	cmd := exec.CommandContext(ctx, "nix-collect-garbage", "-d")
	err = cmd.Run()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to collect garbage: %w", err))
	}

	// Get store size after garbage collection
	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get post-gc store size: %w", err))
	}

	bytesFreed := beforeSize - afterSize
	if bytesFreed < 0 {
		bytesFreed = 0 // Shouldn't happen but guard against it
	}

	return result.Ok(domain.CleanResult{
		ItemsRemoved: 1,
		FreedBytes:   bytesFreed,
		ItemsFailed:  0,
		CleanTime:    time.Since(time.Now()),
		CleanedAt:    time.Now(),
		Strategy:     "NIX_GC",
	})
}

// getActualStoreSize helper function to get real store size
func (n *NixAdapter) getActualStoreSize(ctx context.Context) (int64, error) {
	cmd := exec.CommandContext(ctx, "du", "-sb", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid du output: %s", string(output))
	}

	return strconv.ParseInt(fields[0], 10, 64)
}

// RemoveGeneration removes specific Nix generation with real byte calculation
func (n *NixAdapter) RemoveGeneration(ctx context.Context, genID int) result.Result[domain.CleanResult] {
	// Get store size before removal
	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get pre-remove store size: %w", err))
	}

	// Remove the specific generation
	cmd := exec.CommandContext(ctx, "nix-env", "--delete-generations", fmt.Sprintf("%d", genID))
	err = cmd.Run()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to remove generation %d: %w", genID, err))
	}

	// Get store size after removal
	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get post-remove store size: %w", err))
	}

	bytesFreed := beforeSize - afterSize
	if bytesFreed < 0 {
		bytesFreed = 0 // Guard against negative values
	}

	return result.Ok(domain.CleanResult{
		ItemsRemoved: 1,
		FreedBytes:   bytesFreed,
		ItemsFailed:  0,
		CleanTime:    time.Since(time.Now()),
		CleanedAt:    time.Now(),
		Strategy:     "REMOVE_GENERATION",
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

