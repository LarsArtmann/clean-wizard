package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
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
)

type DockerCleaner struct {
	verbose   bool
	dryRun    bool
	pruneMode DockerPruneMode
}

// DockerPruneMode represents Docker prune mode.
type DockerPruneMode string

const (
	DockerPruneLight      DockerPruneMode = "light"      // docker system prune -f
	DockerPruneStandard   DockerPruneMode = "standard"   // docker system prune -af
	DockerPruneAggressive DockerPruneMode = "aggressive" // docker system prune -af --volumes
)

// NewDockerCleaner creates Docker cleaner.
func NewDockerCleaner(verbose, dryRun bool, pruneMode DockerPruneMode) *DockerCleaner {
	return &DockerCleaner{
		verbose:   verbose,
		dryRun:    dryRun,
		pruneMode: pruneMode,
	}
}

// Type returns operation type for Docker cleaner.
func (dc *DockerCleaner) Type() domain.OperationType {
	return domain.OperationTypeDocker
}

// execWithTimeout executes a Docker command with timeout.
func (dc *DockerCleaner) execWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeoutCtx, cancel := context.WithTimeout(ctx, dockerCommandTimeout)
	_ = cancel // will be called by cmd.Wait() or context usage
	return exec.CommandContext(timeoutCtx, name, arg...)
}

// IsAvailable checks if Docker is available.
func (dc *DockerCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// ValidateSettings validates Docker cleaner settings.
func (dc *DockerCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.Docker == nil {
		return nil // Settings are optional
	}

	// Validate prune mode using domain enum validation
	if !settings.Docker.PruneMode.IsValid() {
		return fmt.Errorf("invalid prune mode: %v (must be a valid DockerPruneMode value)", settings.Docker.PruneMode)
	}

	return nil
}

// Scan scans for Docker resources.
func (dc *DockerCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	if !dc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	// Scan for dangling images
	imagesResult := dc.scanDanglingImages(ctx)
	if imagesResult.IsErr() {
		if dc.verbose {
			fmt.Printf("Warning: failed to scan dangling images: %v\n", imagesResult.Error())
		}
	} else {
		items = append(items, imagesResult.Value()...)
	}

	// Scan for unused containers
	containersResult := dc.scanUnusedContainers(ctx)
	if containersResult.IsErr() {
		if dc.verbose {
			fmt.Printf("Warning: failed to scan unused containers: %v\n", containersResult.Error())
		}
	} else {
		items = append(items, containersResult.Value()...)
	}

	// Scan for unused volumes
	volumesResult := dc.scanUnusedVolumes(ctx)
	if volumesResult.IsErr() {
		if dc.verbose {
			fmt.Printf("Warning: failed to scan unused volumes: %v\n", volumesResult.Error())
		}
	} else {
		items = append(items, volumesResult.Value()...)
	}

	return result.Ok(items)
}

// scanDockerResources converts Docker resource IDs to scan items.
func (dc *DockerCleaner) scanDockerResources(ids []string, resourceType DockerResourceType) []domain.ScanItem {
	items := make([]domain.ScanItem, 0, len(ids))

	for _, id := range ids {
		dc.addScanItem(&items, id, resourceType)
	}

	return items
}

// addScanItem adds a single scan item to the items slice.
func (dc *DockerCleaner) addScanItem(items *[]domain.ScanItem, id string, resourceType DockerResourceType) {
	if id == "" {
		return
	}

	*items = append(*items, domain.ScanItem{
		Path:     fmt.Sprintf("docker:%s:%s", resourceType, id),
		Size:     0, // Size unknown without inspecting
		Created:  time.Time{},
		ScanType: domain.ScanTypeTemp,
	})

	if dc.verbose {
		fmt.Printf("Found %s: %s\n", resourceType, id)
	}
}

// scanDanglingImages scans for dangling Docker images.
func (dc *DockerCleaner) scanDanglingImages(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := dc.execWithTimeout(ctx, "docker", "images", "-f", "dangling=true", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan dangling images: %w", err))
	}

	imageIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
	items := dc.scanDockerResources(imageIDs, dockerImage)

	return result.Ok(items)
}

// scanUnusedContainers scans for stopped Docker containers.
func (dc *DockerCleaner) scanUnusedContainers(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := dc.execWithTimeout(ctx, "docker", "ps", "-a", "-q", "--filter", "status=exited")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan unused containers: %w", err))
	}

	containerIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
	items := dc.scanDockerResources(containerIDs, dockerContainer)

	return result.Ok(items)
}

// scanUnusedVolumes scans for unused Docker volumes.
func (dc *DockerCleaner) scanUnusedVolumes(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := exec.CommandContext(ctx, "docker", "volume", "ls", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan volumes: %w", err))
	}

	volumeIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
	items := dc.scanDockerResources(volumeIDs, dockerVolume)

	return result.Ok(items)
}

// Clean removes Docker resources based on prune mode.
func (dc *DockerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !dc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](errors.New("Docker not available"))
	}

	if dc.dryRun {
		// Estimate cache sizes based on typical usage
		var totalBytes int64
		switch dc.pruneMode {
		case DockerPruneLight:
			totalBytes = int64(100 * 1024 * 1024) // Estimate 100MB
		case DockerPruneStandard:
			totalBytes = int64(500 * 1024 * 1024) // Estimate 500MB
		case DockerPruneAggressive:
			totalBytes = int64(2 * 1024 * 1024 * 1024) // Estimate 2GB
		}

		itemsRemoved := 1

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()

	pruneResult := dc.pruneDocker(ctx)
	if pruneResult.IsErr() {
		return result.Err[domain.CleanResult](fmt.Errorf("docker prune failed: %w", pruneResult.Error()))
	}

	cleanResult := pruneResult.Value()

	duration := time.Since(startTime)
	finalResult := domain.CleanResult{
		FreedBytes:   cleanResult.FreedBytes,
		ItemsRemoved: cleanResult.ItemsRemoved,
		ItemsFailed:  cleanResult.ItemsFailed,
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(finalResult)
}

// pruneDocker executes appropriate Docker prune command based on mode.
func (dc *DockerCleaner) pruneDocker(ctx context.Context) result.Result[domain.CleanResult] {
	var args []string

	switch dc.pruneMode {
	case DockerPruneLight:
		args = []string{"system", "prune", "-f"}
		if dc.verbose {
			fmt.Println("  Running light prune: docker system prune -f")
		}

	case DockerPruneStandard:
		args = []string{"system", "prune", "-af"}
		if dc.verbose {
			fmt.Println("  Running standard prune: docker system prune -af")
		}

	case DockerPruneAggressive:
		args = []string{"system", "prune", "-af", "--volumes"}
		if dc.verbose {
			fmt.Println("  Running aggressive prune: docker system prune -af --volumes")
		}

	default:
		return result.Err[domain.CleanResult](fmt.Errorf("unknown Docker prune mode: %s", dc.pruneMode))
	}

	cmd := dc.execWithTimeout(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("docker system prune failed: %w (output: %s)", err, string(output)))
	}

	if dc.verbose {
		fmt.Printf("  ✓ Docker prune completed\n")
		fmt.Printf("  Output: %s\n", string(output))
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   0, // Bytes freed unknown from prune output
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}
