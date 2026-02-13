package cleaner

import (
	"fmt"
	"strings"
)

// ParseDockerReclaimedSpace extracts "Total reclaimed space: X" from docker prune output.
// Returns 0 if no reclaimed space is found (which is valid).
func ParseDockerReclaimedSpace(output string) (int64, error) {
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if strings.Contains(line, "Total reclaimed space:") {
			// Extract: "Total reclaimed space: 2.5GB"
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}
			sizeStr := strings.TrimSpace(parts[1])
			// Parse size string (e.g., "2.5GB", "100MB", "1.84kB", "0B")
			return ParseDockerSize(sizeStr)
		}
	}
	return 0, nil // No space found (0 is valid)
}

// ParseDockerSize converts Docker size string to bytes.
// Supports units: B, kB, MB, GB, TB (case-insensitive).
func ParseDockerSize(sizeStr string) (int64, error) {
	// Remove any whitespace
	sizeStr = strings.TrimSpace(sizeStr)

	// Handle "0B" case or empty string
	if sizeStr == "0B" || sizeStr == "0" || sizeStr == "" {
		return 0, nil
	}

	// Parse number and unit
	var number float64
	var unit string
	n, err := fmt.Sscanf(sizeStr, "%f%s", &number, &unit)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// Convert to bytes based on unit
	switch strings.ToLower(unit) {
	case "b":
		return int64(number), nil
	case "kb":
		return int64(number * 1024), nil
	case "mb":
		return int64(number * 1024 * 1024), nil
	case "gb":
		return int64(number * 1024 * 1024 * 1024), nil
	case "tb":
		return int64(number * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}
