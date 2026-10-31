package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

const (
	// NixDryRunStoreSizeGB is the estimated Nix store size for dry-run mode.
	NixDryRunStoreSizeGB = 50
	// NixGCPostDelay is the delay after garbage collection to allow async operations.
	NixGCPostDelay = 500 * time.Millisecond

	// Byte conversion constants.
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024

	// NixGenerationMinFields is the minimum number of fields expected in generation output.
	NixGenerationMinFields = 3
)

// boolToGenerationStatus converts boolean to GenerationStatus enum.
func boolToGenerationStatus(b bool) enums.GenerationStatus {
	return enums.GenerationStatusFromBool(b)
}

// NixAdapter wraps Nix package manager operations.
type NixAdapter struct {
	timeout time.Duration
	retries int
	dryRun  bool
}

// NewNixAdapter creates Nix adapter with configuration.
func NewNixAdapter(timeout time.Duration, retries int) *NixAdapter {
	return &NixAdapter{ //nolint:exhaustruct_v5
		timeout: timeout,
		retries: retries,
	}
}

// SetDryRun configures dry-run mode for the adapter.
func (n *NixAdapter) SetDryRun(dryRun bool) {
	n.dryRun = dryRun
}

// ListGenerations lists Nix generations.
// In dry-run mode, still lists real generations but won't actually delete them.
func (n *NixAdapter) ListGenerations(ctx context.Context) result.Result[[]types.NixGeneration] {
	if !n.IsAvailable(ctx) {
		return result.Err[[]types.NixGeneration](errors.New("nix not available"))
	}

	// Use nix-env without --profile to let it use the default user profile
	cmd := n.execWithTimeout(ctx, "nix-env", "--list-generations")

	output, err := cmd.Output()
	if err != nil {
		return result.Err[[]types.NixGeneration](fmt.Errorf("failed to list generations: %w", err))
	}

	var generations []types.NixGeneration

	lines := strings.SplitSeq(string(output), "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		gen, err := n.ParseGeneration(line)
		if err != nil {
			return result.Err[[]types.NixGeneration](
				fmt.Errorf("failed to parse generation: %w", err),
			)
		}

		generations = append(generations, gen)
	}

	return result.Ok(generations)
}

// GetStoreSize returns Nix store size with dry-run isolation.
func (n *NixAdapter) GetStoreSize(ctx context.Context) result.Result[int64] {
	// If dry-run, return estimated size without system calls
	if n.dryRun {
		return result.Ok(int64(NixDryRunStoreSizeGB * bytesPerGB))
	}

	// Real system call for production mode
	cmd := n.execWithTimeout(ctx, "du", "-sb", "/nix/store")

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

// CollectGarbage runs garbage collection on the Nix store.
// Uses the modern "nix store gc" command (Nix 2.4+).
// In dry-run mode, returns a success result without actually running garbage collection.
func (n *NixAdapter) CollectGarbage(ctx context.Context) result.Result[types.CleanResult] {
	// Dry-run GC: the freed size depends on unique closures and cannot be
	// known before the run, so the estimate is honestly Unknown; the real run
	// measures before/after disk usage.
	if n.dryRun {
		cleanResult := conversions.NewCleanResultWithSizeEstimate(
			enums.StrategyAggressiveType,
			1,
			0,
			types.SizeEstimate{Known: 0, Status: enums.SizeEstimateStatusUnknown},
		)

		return result.Ok(cleanResult)
	}

	startTime := time.Now()

	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(
			fmt.Errorf("failed to get pre-gc store size: %w", err),
		)
	}

	cmd := n.execWithTimeout(ctx, "nix", "store", "gc")

	err = cmd.Run()
	if err != nil {
		return conversions.ToCleanResultFromError(fmt.Errorf("failed to run nix store gc: %w", err))
	}

	time.Sleep(NixGCPostDelay)

	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(
			fmt.Errorf("failed to get post-gc store size: %w", err),
		)
	}

	bytesFreed := max(beforeSize-afterSize, 0)

	cleanResult := conversions.NewCleanResultWithTiming(
		enums.StrategyAggressiveType,
		1,
		bytesFreed,
		time.Since(startTime),
	)

	return result.Ok(cleanResult)
}

// getActualStoreSize helper function to get real store size.
func (n *NixAdapter) getActualStoreSize(ctx context.Context) (int64, error) {
	cmd := n.execWithTimeout(ctx, "du", "-sb", "/nix/store")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get store size: %w", err)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid du output: %s", string(output))
	}

	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse store size: %w", err)
	}

	return size, nil
}

// RemoveGeneration removes specific Nix generation.
// In dry-run mode, returns a success result without actually deleting.
func (n *NixAdapter) RemoveGeneration(
	ctx context.Context,
	genID types.NixGenerationID,
) result.Result[types.CleanResult] {
	// In dry-run mode, return success without actually removing. The size
	// estimate is honestly Unknown: generation deletion removes profile
	// symlinks only; store paths are freed by garbage collection.
	if n.dryRun {
		cleanResult := conversions.NewCleanResultWithSizeEstimate(
			enums.StrategyConservativeType,
			1,
			0,
			types.SizeEstimate{Known: 0, Status: enums.SizeEstimateStatusUnknown},
		)

		return result.Ok(cleanResult)
	}

	startTime := time.Now()

	// Get store size before removal
	beforeSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(
			fmt.Errorf("failed to get pre-remove store size for generation %d: %w", genID, err),
		)
	}

	// Remove the specific generation
	cmd := n.execWithTimeout(ctx, "nix-env", "--delete-generations", strconv.Itoa(int(genID)))

	err = cmd.Run()
	if err != nil {
		return conversions.ToCleanResultFromError(
			fmt.Errorf("failed to remove generation %d: %w", genID, err),
		)
	}

	// Get store size after removal
	afterSize, err := n.getActualStoreSize(ctx)
	if err != nil {
		return conversions.ToCleanResultFromError(
			fmt.Errorf("failed to get post-remove store size for generation %d: %w", genID, err),
		)
	}

	bytesFreed := max(beforeSize-afterSize,
		// Guard against negative values
		0)

	// Use centralized conversion with proper timing
	cleanResult := conversions.NewCleanResultWithTiming(
		enums.StrategyConservativeType,
		1,
		bytesFreed,
		time.Since(startTime),
	)

	return result.Ok(cleanResult)
}

// ParseGeneration parses generation line from nix-env output.
// Expected format when using --list-generations (without --profile):
//
//	"32   2026-01-12 08:03:14"
//	"33   2026-01-15 21:14:05   (current)"
func (n *NixAdapter) ParseGeneration(line string) (types.NixGeneration, error) {
	fields := strings.Fields(line)
	if len(fields) < NixGenerationMinFields {
		return types.NixGeneration{}, fmt.Errorf("invalid generation line: %s", line)
	}

	// Parse generation ID from first field
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return types.NixGeneration{}, fmt.Errorf("invalid generation ID: %s", fields[0])
	}

	// Parse date and time from second and third fields
	dateTimeStr := fmt.Sprintf("%s %s", fields[1], fields[2])

	date, err := time.Parse("2006-01-02 15:04:05", dateTimeStr)
	if err != nil {
		return types.NixGeneration{}, fmt.Errorf("invalid date/time: %s %s", fields[1], fields[2])
	}

	// Build the profile path using the standard Nix profile directory
	// On macOS with Nix installed via single-user or multi-user:
	// User profiles are typically in ~/.local/state/nix/profiles/ (Darwin)
	// or ~/.nix-profile/ (Linux fallback)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return types.NixGeneration{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	path := fmt.Sprintf("%s/.local/state/nix/profiles/profile-%d-link", homeDir, id)

	// Check if this is the current generation
	isCurrent := strings.Contains(line, "current")

	return types.NixGeneration{
		ID:      types.NixGenerationID(id),
		Path:    path,
		Date:    date,
		Current: boolToGenerationStatus(isCurrent),
	}, nil
}

// IsAvailable checks if Nix is available and accessible.
func (n *NixAdapter) IsAvailable(ctx context.Context) bool {
	// Check if nix command exists
	versionCmd := n.execWithTimeout(ctx, "nix", "--version")

	return versionCmd.Run() == nil
}

// ContextHealthchecker mirrors do.HealthcheckerWithContext so the adapters
// package stays free of the DI library while still proving the contract its
// services must satisfy for container health checks.
type ContextHealthchecker interface {
	HealthCheck(context.Context) error
}

var _ ContextHealthchecker = (*NixAdapter)(nil)

// HealthCheck reports whether the nix binary can be executed. It is the
// container-domain fault signal consumed by do health checks; the scan-domain
// availability signal is IsAvailable. Both probe `nix --version`, but only
// HealthCheck carries a classified not-available error for the doctor surface.
func (n *NixAdapter) HealthCheck(ctx context.Context) error {
	timeout := n.timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Own exec path: adapters.ExecWithTimeout is broken for deadline-less
	// contexts (ticketed #56) and must not gain new callers.
	cmd := exec.CommandContext(checkCtx, "nix", "--version")
	if err := cmd.Run(); err != nil {
		return cleaner.NewNotAvailableError("nix", fmt.Sprintf("nix --version failed: %v", err))
	}

	return nil
}
