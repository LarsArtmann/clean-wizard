package docker

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
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

// ParseDockerSize converts a Docker size string (e.g. "2.5GB", "100MB", "1.84kB", "0B") to bytes.
// Unit parsing is delegated to humanize.ParseBytes, which natively handles all
// SI and IEC unit suffixes (case-insensitive); a number without a unit is treated as bytes.
func ParseDockerSize(sizeStr string) (int64, error) {
	// Empty string means nothing to parse; Docker reports zero as "0B" (handled by humanize).
	if sizeStr == "" {
		return 0, nil
	}

	parsed, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size format %q: %w", sizeStr, err)
	}

	return int64(parsed), nil
}
