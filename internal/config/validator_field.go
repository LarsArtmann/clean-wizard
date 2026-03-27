package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// validateMaxDiskUsage validates max disk usage percentage.
func (cv *ConfigValidator) validateMaxDiskUsage(value any) error {
	usage, ok := value.(int)
	if !ok {
		return fmt.Errorf("max_disk_usage must be an integer, got %T", value)
	}

	minVal, maxVal := cv.getMaxDiskUsageBounds()
	if usage < minVal {
		return fmt.Errorf("max_disk_usage (%d) below minimum (%d)", usage, minVal)
	}

	if usage > maxVal {
		return fmt.Errorf("max_disk_usage (%d) above maximum (%d)", usage, maxVal)
	}

	return nil
}

// getMaxDiskUsageBounds safely returns min and max bounds for max disk usage.
func (cv *ConfigValidator) getMaxDiskUsageBounds() (minVal, maxVal int) {
	minVal, maxVal = 0, 100 // Safe defaults

	if cv.rules.MaxDiskUsage != nil {
		if cv.rules.MaxDiskUsage.Min != nil {
			minVal = *cv.rules.MaxDiskUsage.Min
		}

		if cv.rules.MaxDiskUsage.Max != nil {
			maxVal = *cv.rules.MaxDiskUsage.Max
		}
	}

	return minVal, maxVal
}

// validateProtectedPaths validates protected paths array.
func (cv *ConfigValidator) validateProtectedPaths(value any) error {
	paths, ok := value.([]string)
	if !ok {
		return fmt.Errorf("protected must be a string array, got %T", value)
	}

	// Explicitly reject empty arrays as required field
	if len(paths) == 0 {
		return errors.New("protected paths cannot be empty")
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

// findDuplicatePaths finds duplicate paths in the given slice.
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
