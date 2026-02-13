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
	pruneMode domain.DockerPruneMode
}

// NewDockerCleaner creates Docker cleaner.
func NewDockerCleaner(verbose, dryRun bool, pruneMode domain.DockerPruneMode) *DockerCleaner {
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

// Name returns the unique identifier for this cleaner.
func (dc *DockerCleaner) Name() string {
	return "docker"
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
// Deprecated: Use scan-specific methods that include size parsing.
func (dc *DockerCleaner) scanDockerResources(ids []string, resourceType DockerResourceType) []domain.ScanItem {
	items := make([]domain.ScanItem, 0, len(ids))

	for _, id := range ids {
		dc.addScanItem(&items, id, resourceType, 0)
	}

	return items
}

// addScanItem adds a single scan item to the items slice.
func (dc *DockerCleaner) addScanItem(items *[]domain.ScanItem, id string, resourceType DockerResourceType, size int64) {
	if id == "" {
		return
	}

	*items = append(*items, domain.ScanItem{
		Path:     fmt.Sprintf("docker:%s:%s", resourceType, id),
		Size:     size,
		Created:  time.Time{},
		ScanType: domain.ScanTypeTemp,
	})

	if dc.verbose {
		fmt.Printf("Found %s: %s (size: %s)\n", resourceType, id, formatBytes(size))
	}
}

// formatBytes formats bytes as a human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// scanDanglingImages scans for dangling Docker images with size information.
func (dc *DockerCleaner) scanDanglingImages(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := dc.execWithTimeout(ctx, "docker", "images", "-f", "dangling=true", "--format", "{{.ID}}\t{{.Size}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan dangling images: %w", err))
	}

	items := dc.parseDockerResourceOutput(string(output), dockerImage)
	return result.Ok(items)
}

// scanUnusedContainers scans for stopped Docker containers with size information.
func (dc *DockerCleaner) scanUnusedContainers(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := dc.execWithTimeout(ctx, "docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}}\t{{.Size}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan unused containers: %w", err))
	}

	items := dc.parseDockerResourceOutput(string(output), dockerContainer)
	return result.Ok(items)
}

// scanUnusedVolumes scans for unused Docker volumes with size information.
func (dc *DockerCleaner) scanUnusedVolumes(ctx context.Context) result.Result[[]domain.ScanItem] {
	// Use docker system df -v to get volume sizes
	cmd := dc.execWithTimeout(ctx, "docker", "system", "df", "-v", "--format", "{{json .Volumes}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to basic listing if json format not available
		return dc.scanVolumesFallback(ctx)
	}

	items := dc.parseVolumeJSONOutput(string(output))
	return result.Ok(items)
}

// scanVolumesFallback is a fallback method for scanning volumes when system df fails.
func (dc *DockerCleaner) scanVolumesFallback(ctx context.Context) result.Result[[]domain.ScanItem] {
	cmd := dc.execWithTimeout(ctx, "docker", "volume", "ls", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to scan volumes: %w", err))
	}

	volumes := strings.Split(strings.TrimSpace(string(output)), "\n")
	items := dc.scanDockerResources(volumes, dockerVolume)
	return result.Ok(items)
}

// parseDockerResourceOutput parses Docker output with format "ID\tSIZE".
func (dc *DockerCleaner) parseDockerResourceOutput(output string, resourceType DockerResourceType) []domain.ScanItem {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	items := make([]domain.ScanItem, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 1 || parts[0] == "" {
			continue
		}

		id := parts[0]
		var size int64
		if len(parts) > 1 {
			size = dc.parseDockerSizeFromOutput(parts[1], resourceType)
		}

		dc.addScanItem(&items, id, resourceType, size)
	}

	return items
}

// parseDockerSizeFromOutput parses size from Docker output based on resource type.
func (dc *DockerCleaner) parseDockerSizeFromOutput(sizeStr string, resourceType DockerResourceType) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0
	}

	// Container sizes have format "X (virtual Y)" - extract the actual size
	if resourceType == dockerContainer {
		// Format: "0B (virtual 123MB)" or "1.84kB (virtual 500MB)"
		if idx := strings.Index(sizeStr, "(virtual"); idx > 0 {
			sizeStr = strings.TrimSpace(sizeStr[:idx])
		}
	}

	size, err := ParseDockerSize(sizeStr)
	if err != nil && dc.verbose {
		fmt.Printf("Warning: failed to parse size '%s': %v\n", sizeStr, err)
	}
	return size
}

// parseVolumeJSONOutput parses JSON output from docker system df -v for volumes.
func (dc *DockerCleaner) parseVolumeJSONOutput(output string) []domain.ScanItem {
	output = strings.TrimSpace(output)
	if output == "" || output == "null" || output == "[]" {
		return []domain.ScanItem{}
	}

	// Simple parsing for volume entries - format varies by Docker version
	// This handles basic cases; complex JSON parsing would require encoding/json
	items := make([]domain.ScanItem, 0)

	// Extract volume names using simple string parsing
	// Look for "Name":"volume_name" patterns
	namePattern := `"Name":"`
	idx := 0
	for {
		nameIdx := strings.Index(output[idx:], namePattern)
		if nameIdx == -1 {
			break
		}
		nameIdx += idx + len(namePattern)
		endIdx := strings.Index(output[nameIdx:], `"`)
		if endIdx == -1 {
			break
		}

		volumeName := output[nameIdx : nameIdx+endIdx]
		if volumeName != "" {
			dc.addScanItem(&items, volumeName, dockerVolume, 0)
		}

		idx = nameIdx + endIdx + 1
	}

	return items
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
		case domain.DockerPruneAll:
			totalBytes = int64(2 * 1024 * 1024 * 1024) // Estimate 2GB for everything
		case domain.DockerPruneImages:
			totalBytes = int64(500 * 1024 * 1024) // Estimate 500MB for images
		case domain.DockerPruneContainers:
			totalBytes = int64(100 * 1024 * 1024) // Estimate 100MB for containers
		case domain.DockerPruneVolumes:
			totalBytes = int64(100 * 1024 * 1024) // Estimate 100MB for volumes
		case domain.DockerPruneBuilds:
			totalBytes = int64(500 * 1024 * 1024) // Estimate 500MB for build cache
		}

		itemsRemoved := 1

		cleanResult := conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyDryRunType), itemsRemoved, totalBytes)
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
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	}

	return result.Ok(finalResult)
}

// pruneDocker executes appropriate Docker prune command based on mode.
func (dc *DockerCleaner) pruneDocker(ctx context.Context) result.Result[domain.CleanResult] {
	var args []string

	switch dc.pruneMode {
	case domain.DockerPruneAll:
		args = []string{"system", "prune", "-af", "--volumes"}
		if dc.verbose {
			fmt.Println("  Running full prune: docker system prune -af --volumes")
		}

	case domain.DockerPruneImages:
		args = []string{"image", "prune", "-af"}
		if dc.verbose {
			fmt.Println("  Running image prune: docker image prune -af")
		}

	case domain.DockerPruneContainers:
		args = []string{"container", "prune", "-f"}
		if dc.verbose {
			fmt.Println("  Running container prune: docker container prune -f")
		}

	case domain.DockerPruneVolumes:
		args = []string{"volume", "prune", "-f"}
		if dc.verbose {
			fmt.Println("  Running volume prune: docker volume prune -f")
		}

	case domain.DockerPruneBuilds:
		args = []string{"builder", "prune", "-af"}
		if dc.verbose {
			fmt.Println("  Running builder prune: docker builder prune -af")
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

	// Parse reclaimed space from docker output
	bytesFreed, err := ParseDockerReclaimedSpace(string(output))
	if err != nil && dc.verbose {
		fmt.Printf("  Warning: failed to parse reclaimed space: %v\n", err)
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}
