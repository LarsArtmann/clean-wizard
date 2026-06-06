package cleaner

import (
	"fmt"
	"strings"
)

const (
	bytesPerKB = 1024
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024
	bytesPerTB = 1024 * 1024 * 1024 * 1024
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

// sizeMultiplier converts Docker size unit strings to byte multipliers.
var sizeMultiplier = map[string]int64{ //nolint:gochecknoglobals
	"b":  1,
	"kb": bytesPerKB,
	"mb": bytesPerMB,
	"gb": bytesPerGB,
	"tb": bytesPerTB,
}

// ParseDockerSize converts Docker size string to bytes.
// Supports units: B, kB, MB, GB, TB (case-insensitive).
func ParseDockerSize(sizeStr string) (int64, error) {
	// Handle "0B" case or empty string
	if sizeStr == "0B" || sizeStr == "0" || sizeStr == "" {
		return 0, nil
	}

	// Parse number and unit using shared helper
	number, unit, err := ParseNumberAndUnit(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// Convert to bytes based on unit
	multiplier, ok := sizeMultiplier[strings.ToLower(unit)]
	if !ok {
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	return int64(number * float64(multiplier)), nil
}
