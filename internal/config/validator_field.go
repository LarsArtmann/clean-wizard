package config

import (
	"fmt"
	"path/filepath"
)

// validateMaxDiskUsage validates max disk usage percentage
func (cv *ConfigValidator) validateMaxDiskUsage(value any) error {
	usage, ok := value.(int)
	if !ok {
		return fmt.Errorf("max_disk_usage must be an integer, got %T", value)
	}

	min, max := cv.getMaxDiskUsageBounds()
	if usage < min {
		return fmt.Errorf("max_disk_usage (%d) below minimum (%d)", usage, min)
	}
	if usage > max {
		return fmt.Errorf("max_disk_usage (%d) above maximum (%d)", usage, max)
	}

	return nil
}

// getMaxDiskUsageBounds safely returns min and max bounds for max disk usage
func (cv *ConfigValidator) getMaxDiskUsageBounds() (min, max int) {
	min, max = 0, 100 // Safe defaults

	if cv.rules.MaxDiskUsage != nil {
		if cv.rules.MaxDiskUsage.Min != nil {
			min = *cv.rules.MaxDiskUsage.Min
		}
		if cv.rules.MaxDiskUsage.Max != nil {
			max = *cv.rules.MaxDiskUsage.Max
		}
	}
	return min, max
}

// validateProtectedPaths validates protected paths array
func (cv *ConfigValidator) validateProtectedPaths(value any) error {
	paths, ok := value.([]string)
	if !ok {
		return fmt.Errorf("protected must be a string array, got %T", value)
	}

	// Explicitly reject empty arrays as required field
	if len(paths) == 0 {
		return fmt.Errorf("protected paths cannot be empty")
	}

	for i, path := range paths {
		if path == "" {
			return fmt.Errorf("protected[%d] cannot be empty", i)
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("protected[%d] must be absolute path: %s", i, path)
		}
	}

	return nil
}

// findDuplicatePaths finds duplicate paths in the given slice
func (cv *ConfigValidator) findDuplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	duplicates := []string{}

	for _, path := range paths {
		if seen[path] {
			duplicates = append(duplicates, path)
		} else {
			seen[path] = true
		}
	}

	return duplicates
}
