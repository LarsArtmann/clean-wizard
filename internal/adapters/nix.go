package adapters

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// boolToGenerationStatus converts boolean to GenerationStatus enum.
func boolToGenerationStatus(b bool) domain.GenerationStatus {
	if b {
		return domain.GenerationStatusCurrent
	}
	return domain.GenerationStatusHistorical
}

// NixAdapter wraps Nix package manager operations.
type NixAdapter struct {
	timeout time.Duration
	retries int
	dryRun  bool
}

// NewNixAdapter creates Nix adapter with configuration.
func NewNixAdapter(timeout time.Duration, retries int) *NixAdapter {
	return &NixAdapter{
		timeout: timeout,
		retries: retries,
	}
}

// SetDryRun configures dry-run mode for the adapter.
func (n *NixAdapter) SetDryRun(dryRun bool) {
	n.dryRun = dryRun
}

// ListGenerations lists Nix generations with dry-run isolation.
func (n *NixAdapter) ListGenerations(ctx context.Context) result.Result[[]domain.NixGeneration] {
	if !n.IsAvailable(ctx) {
		return result.Err[[]domain.NixGeneration](errors.New("nix not available"))
	}

	// If dry-run, return mock data without system calls
	if n.dryRun {
		return result.Ok([]domain.NixGeneration{
			{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: time.Now().Add(-24 * time.Hour), Current: domain.GenerationStatusCurrent},
			{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: time.Now().Add(-48 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: time.Now().Add(-72 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: time.Now().Add(-96 * time.Hour), Current: domain.GenerationStatusHistorical},
			{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: time.Now().Add(-120 * time.Hour), Current: domain.GenerationStatusHistorical},
		})
	}

	// Real system call for production mode
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

// GetStoreSize returns Nix store size with dry-run isolation.
func (n *NixAdapter) GetStoreSize(ctx context.Context) result.Result[int64] {
	// If dry-run, return estimated size without system calls
	if n.dryRun {
		return result.Ok(int64(50 * 1024 * 1024 * 1024)) // 50GB estimate
	}

	// Real system call for production mode
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

// CollectGarbage removes old Nix generations using centralized conversion.
func (n *NixAdapter) CollectGarbage(ctx context.Context) result.Result[domain.CleanResult] {
	// Get store size before garbage collection
	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to get pre-gc store size: %w", err))
	}

	// Run actual nix-collect-garbage command
	cmd := exec.CommandContext(ctx, "nix-collect-garbage", "-d")
	err = cmd.Run()
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to collect garbage: %w", err))
	}

	// Get store size after garbage collection
	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to get post-gc store size: %w", err))
	}

	bytesFreed := max(beforeSize-afterSize,
		// Shouldn't happen but guard against it
		0)

	// Use centralized conversion with proper timing
	cleanResult := conversions.NewCleanResultWithTiming(domain.StrategyAggressive, 1, bytesFreed, time.Since(time.Now()))
	return result.Ok(cleanResult)
}

// getActualStoreSize helper function to get real store size.
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

// RemoveGeneration removes specific Nix generation using centralized conversion.
func (n *NixAdapter) RemoveGeneration(ctx context.Context, genID int) result.Result[domain.CleanResult] {
	// Get store size before removal
	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to get pre-remove store size for generation %d: %w", genID, err))
	}

	// Remove the specific generation
	cmd := exec.CommandContext(ctx, "nix-env", "--delete-generations", strconv.Itoa(genID))
	err = cmd.Run()
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to remove generation %d: %w", genID, err))
	}

	// Get store size after removal
	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to get post-remove store size for generation %d: %w", genID, err))
	}

	bytesFreed := max(beforeSize-afterSize,
		// Guard against negative values
		0)

	// Use centralized conversion with proper timing
	cleanResult := conversions.NewCleanResultWithTiming(domain.StrategyConservative, 1, bytesFreed, time.Since(time.Now()))
	return result.Ok(cleanResult)
}

// ParseGeneration parses generation line from nix-env output.
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
		Current: boolToGenerationStatus(strings.Contains(line, "current")),
	}, nil
}

// IsAvailable checks if Nix is available and accessible.
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
