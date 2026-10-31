package docker

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// dockerCommandTimeout is the timeout for Docker operations.
const dockerCommandTimeout = 2 * time.Minute

// DockerResourceType represents Docker resource types for scanning.
type DockerResourceType string

const (
	dockerImage     DockerResourceType = "image"
	dockerContainer DockerResourceType = "container"
	dockerVolume    DockerResourceType = "volume"
	// DockerParseTabSplitCount is the expected split count for parsing Docker output with tab delimiter.
	DockerParseTabSplitCount = 2
)

type DockerCleaner struct {
	cleaner.CleanerBase

	pruneMode enums.DockerPruneMode
}

// NewDockerCleaner creates Docker cleaner.
func NewDockerCleaner(verbose, dryRun bool, pruneMode enums.DockerPruneMode) *DockerCleaner {
	return &DockerCleaner{
		CleanerBase: cleaner.NewCleanerBase(verbose, dryRun),
		pruneMode:   pruneMode,
	}
}

// Type returns operation type for Docker cleaner.
func (dc *DockerCleaner) Type() operations.OperationType {
	return operations.OperationTypeDocker
}

// Name returns the unique identifier for this cleaner.
func (dc *DockerCleaner) Name() string {
	return cleaner.CleanerDocker
}

// IsAvailable checks if Docker is available.
func (dc *DockerCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("docker")

	return err == nil
}

// ValidateSettings validates Docker cleaner settings.
func (dc *DockerCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.DockerSettings { return s.Docker },
		func(d *operations.DockerSettings) error {
			if !d.PruneMode.IsValid() {
				return fmt.Errorf("invalid DockerPruneMode: %d", d.PruneMode)
			}

			return nil
		},
	)
}

// Scan scans for Docker resources.
func (dc *DockerCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

	// Consistent with Clean: unavailability is an Infrastructure error, not a
	// silent empty scan — the execution layer classifies it as skipped and
	// the doctor/scan surfaces must not dress a missing tool as "nothing to
	// clean".
	if !dc.IsAvailable(ctx) {
		return result.Err[[]types.ScanItem](cleaner.NewNotAvailableError("docker", ""))
	}

	// Scan for dangling images
	imagesResult := dc.scanDanglingImages(ctx)
	if imagesResult.IsErr() {
		if dc.GetVerbose() {
			fmt.Printf("Warning: failed to scan dangling images: %v\n", imagesResult.Error())
		}
	} else {
		items = append(items, imagesResult.Value()...)
	}

	// Scan for unused containers
	containersResult := dc.scanUnusedContainers(ctx)
	if containersResult.IsErr() {
		if dc.GetVerbose() {
			fmt.Printf("Warning: failed to scan unused containers: %v\n", containersResult.Error())
		}
	} else {
		items = append(items, containersResult.Value()...)
	}

	// Scan for unused volumes
	volumesResult := dc.scanUnusedVolumes(ctx)
	if volumesResult.IsErr() {
		if dc.GetVerbose() {
			fmt.Printf("Warning: failed to scan unused volumes: %v\n", volumesResult.Error())
		}
	} else {
		items = append(items, volumesResult.Value()...)
	}

	return result.Ok(items)
}

// Clean removes Docker resources based on prune mode.
func (dc *DockerCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	if !dc.IsAvailable(ctx) {
		return result.Err[types.CleanResult](cleaner.NewNotAvailableError("docker", ""))
	}

	if dc.GetDryRun() {
		// Scan for actual sizes instead of using hardcoded estimates
		totalBytes := dc.estimateSizeFromScan(ctx)

		return result.Ok(dc.newDryRunResult(totalBytes, 1))
	}

	// Real cleaning implementation
	startTime := time.Now()

	pruneResult := dc.pruneDocker(ctx)
	if pruneResult.IsErr() {
		return result.Err[types.CleanResult](
			fmt.Errorf("docker prune failed: %w", pruneResult.Error()),
		)
	}

	cleanResult := pruneResult.Value()

	duration := time.Since(startTime)
	finalResult := conversions.NewCleanResultWithTiming(
		enums.StrategyConservativeType,
		int(cleanResult.ItemsRemoved),
		int64(cleanResult.FreedBytes),
		duration,
	)
	finalResult.ItemsFailed = cleanResult.ItemsFailed

	return result.Ok(finalResult)
}

// verbosePrune logs the prune command about to run (when verbose) and returns its args.
func (dc *DockerCleaner) verbosePrune(label, commandLine string, args ...string) []string {
	if dc.GetVerbose() {
		fmt.Printf("  Running %s: %s\n", label, commandLine)
	}

	return args
}

// buildPruneArgs returns the docker command arguments for the prune mode.
func (dc *DockerCleaner) buildPruneArgs() []string {
	switch dc.pruneMode {
	case enums.DockerPruneAll:
		return dc.verbosePrune(
			"full prune", "docker system prune -af --volumes",
			"system", "prune", "-af", "--volumes",
		)

	case enums.DockerPruneImages:
		return dc.verbosePrune("image prune", "docker image prune -af", "image", "prune", "-af")

	case enums.DockerPruneContainers:
		return dc.verbosePrune("container prune", "docker container prune -f", "container", "prune", "-f")

	case enums.DockerPruneVolumes:
		return dc.verbosePrune("volume prune", "docker volume prune -f", "volume", "prune", "-f")

	case enums.DockerPruneBuilds:
		return dc.verbosePrune("builder prune", "docker builder prune -af", "builder", "prune", "-af")

	default:
		return nil
	}
}

// pruneDocker executes appropriate Docker prune command based on mode.
func (dc *DockerCleaner) pruneDocker(ctx context.Context) result.Result[types.CleanResult] {
	args := dc.buildPruneArgs()

	if args == nil {
		return result.Err[types.CleanResult](
			fmt.Errorf("unknown Docker prune mode: %s", dc.pruneMode),
		)
	}

	cmd := adapters.ExecWithTimeout(ctx, dockerCommandTimeout, "docker", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[types.CleanResult](
			fmt.Errorf("docker system prune failed: %w (output: %s)", err, string(output)),
		)
	}

	if dc.GetVerbose() {
		fmt.Printf("  ✓ Docker prune completed\n")
		fmt.Printf("  Output: %s\n", string(output))
	}

	// Parse reclaimed space from docker output
	bytesFreed, err := ParseDockerReclaimedSpace(string(output))
	if err != nil && dc.GetVerbose() {
		fmt.Printf("  Warning: failed to parse reclaimed space: %v\n", err)
	}

	return result.Ok(
		conversions.NewCleanResult(
			enums.StrategyConservativeType,
			1,
			bytesFreed,
		),
	)
}

// newDryRunResult creates a CleanResult for dry-run mode with the estimated size.
func (dc *DockerCleaner) newDryRunResult(totalBytes int64, itemsCount int) types.CleanResult {
	return conversions.NewCleanResult(
		enums.StrategyDryRunType,
		itemsCount,
		totalBytes,
	)
}

// GetPruneMode returns the configured Docker prune mode.
func (dc *DockerCleaner) GetPruneMode() enums.DockerPruneMode { return dc.pruneMode }
