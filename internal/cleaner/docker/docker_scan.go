package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// This file carries the Docker cleaner's scan (discovery) side: resource
// enumeration, output parsing, and size estimation. The cleaner lifecycle
// (Scan/Clean entry points, prune execution) lives in docker.go.

// scanDockerResources converts Docker resource IDs to scan items.
//
// Deprecated: Use scan-specific methods that include size parsing.
func (dc *DockerCleaner) scanDockerResources(
	ids []string,
	resourceType DockerResourceType,
) []types.ScanItem {
	items := make([]types.ScanItem, 0, len(ids))

	for _, id := range ids {
		dc.addScanItem(&items, types.DockerResourceID(id), resourceType, 0)
	}

	return items
}

// addScanItem adds a single scan item to the items slice.
func (dc *DockerCleaner) addScanItem(
	items *[]types.ScanItem,
	id types.DockerResourceID,
	resourceType DockerResourceType,
	size int64,
) {
	if id == "" {
		return
	}

	*items = append(*items, types.ScanItem{
		Path:     fmt.Sprintf("docker:%s:%s", resourceType, id),
		Size:     size,
		Created:  time.Time{},
		ScanType: types.ScanTypeTemp,
	})

	if dc.GetVerbose() {
		fmt.Printf("Found %s: %s (size: %s)\n", resourceType, id, format.Bytes(size))
	}
}

// scanDanglingImages scans for dangling Docker images with size information.
func (dc *DockerCleaner) scanDanglingImages(ctx context.Context) result.Result[[]types.ScanItem] {
	cmd := adapters.ExecWithTimeout(ctx, dockerCommandTimeout,
		"docker", "images", "-f", "dangling=true", "--format", "{{.ID}}\t{{.Size}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]types.ScanItem](fmt.Errorf("failed to scan dangling images: %w", err))
	}

	items := dc.parseDockerResourceOutput(string(output), dockerImage)

	return result.Ok(items)
}

// scanAndAccumulate handles the common pattern of scanning resources and accumulating their sizes.
// Returns true if processing should continue (pruneMode is DockerPruneAll), false to return early.
func (dc *DockerCleaner) scanAndAccumulate(
	ctx context.Context,
	checkMode enums.DockerPruneMode,
	scanFunc func(context.Context) result.Result[[]types.ScanItem],
	totalBytes *int64,
) bool {
	if dc.pruneMode == enums.DockerPruneAll || dc.pruneMode == checkMode {
		if result := scanFunc(ctx); result.IsOk() {
			for _, item := range result.Value() {
				*totalBytes += item.Size
			}
		}

		if dc.pruneMode != enums.DockerPruneAll {
			return false
		}
	}

	return true
}

// scanUnusedContainers scans for stopped Docker containers with size information.
func (dc *DockerCleaner) scanUnusedContainers(
	ctx context.Context,
) result.Result[[]types.ScanItem] {
	cmd := adapters.ExecWithTimeout(ctx, dockerCommandTimeout,
		"docker", "ps", "-a", "--filter", "status=exited", "--format", "{{.ID}}\t{{.Size}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]types.ScanItem](
			fmt.Errorf("failed to scan unused containers: %w", err),
		)
	}

	items := dc.parseDockerResourceOutput(string(output), dockerContainer)

	return result.Ok(items)
}

// scanUnusedVolumes scans for unused Docker volumes with size information.
func (dc *DockerCleaner) scanUnusedVolumes(ctx context.Context) result.Result[[]types.ScanItem] {
	// Use docker system df -v to get volume sizes
	cmd := adapters.ExecWithTimeout(ctx, dockerCommandTimeout,
		"docker", "system", "df", "-v", "--format", "{{json .Volumes}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to basic listing if json format not available
		return dc.scanVolumesFallback(ctx)
	}

	items := dc.parseVolumeJSONOutput(string(output))

	return result.Ok(items)
}

// scanVolumesFallback is a fallback method for scanning volumes when system df fails.
func (dc *DockerCleaner) scanVolumesFallback(ctx context.Context) result.Result[[]types.ScanItem] {
	cmd := adapters.ExecWithTimeout(ctx, dockerCommandTimeout, "docker", "volume", "ls", "-q")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[[]types.ScanItem](fmt.Errorf("failed to scan volumes: %w", err))
	}

	volumes := strings.Split(strings.TrimSpace(string(output)), "\n")
	items := dc.scanDockerResources(volumes, dockerVolume)

	return result.Ok(items)
}

// parseDockerResourceOutput parses Docker output with format "ID\tSIZE".
func (dc *DockerCleaner) parseDockerResourceOutput(
	output string,
	resourceType DockerResourceType,
) []types.ScanItem {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	items := make([]types.ScanItem, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", DockerParseTabSplitCount)
		if len(parts) < 1 || parts[0] == "" {
			continue
		}

		id := parts[0]

		var size int64
		if len(parts) > 1 {
			size = dc.parseDockerSizeFromOutput(parts[1], resourceType)
		}

		dc.addScanItem(&items, types.DockerResourceID(id), resourceType, size)
	}

	return items
}

// parseDockerSizeFromOutput parses size from Docker output based on resource type.
func (dc *DockerCleaner) parseDockerSizeFromOutput(
	sizeStr string,
	resourceType DockerResourceType,
) int64 {
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
	if err != nil && dc.GetVerbose() {
		fmt.Printf("Warning: failed to parse size '%s': %v\n", sizeStr, err)
	}

	return size
}

// estimateSizeFromScan scans Docker resources and returns the total estimated size.
func (dc *DockerCleaner) estimateSizeFromScan(ctx context.Context) int64 {
	var totalBytes int64

	switch dc.pruneMode {
	case enums.DockerPruneAll, enums.DockerPruneImages:
		if !dc.scanAndAccumulate(
			ctx,
			enums.DockerPruneImages,
			dc.scanDanglingImages,
			&totalBytes,
		) {
			return totalBytes
		}

		fallthrough
	case enums.DockerPruneContainers:
		if !dc.scanAndAccumulate(
			ctx,
			enums.DockerPruneContainers,
			dc.scanUnusedContainers,
			&totalBytes,
		) {
			return totalBytes
		}

		fallthrough
	case enums.DockerPruneVolumes:
		if !dc.scanAndAccumulate(
			ctx,
			enums.DockerPruneVolumes,
			dc.scanUnusedVolumes,
			&totalBytes,
		) {
			return totalBytes
		}

		fallthrough
	case enums.DockerPruneBuilds:
		// Build cache size estimation - try to get from docker system df
		cmd := adapters.ExecWithTimeout(
			ctx,
			dockerCommandTimeout,
			"docker",
			"system",
			"df",
			"--format",
			"{{.BuildCache}}",
		)
		if output, err := cmd.CombinedOutput(); err == nil {
			if size, parseErr := ParseDockerSize(
				strings.TrimSpace(string(output)),
			); parseErr == nil &&
				size > 0 {
				totalBytes += size
			}
		}
	}

	return totalBytes
}

// parseVolumeJSONOutput parses JSON output from docker system df -v for volumes.
func (dc *DockerCleaner) parseVolumeJSONOutput(output string) []types.ScanItem {
	output = strings.TrimSpace(output)
	if output == "" || output == "null" || output == "[]" {
		return []types.ScanItem{}
	}

	// Simple parsing for volume entries - format varies by Docker version
	// This handles basic cases; complex JSON parsing would require encoding/json
	items := make([]types.ScanItem, 0)

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
			dc.addScanItem(&items, types.DockerResourceID(volumeName), dockerVolume, 0)
		}

		idx = nameIdx + endIdx + 1
	}

	return items
}
