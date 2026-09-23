package cleaner

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// Byte conversion constants for disk size math and formatting.
const (
	BytesPerKB = 1024
	BytesPerMB = 1024 * 1024
	BytesPerGB = 1024 * 1024 * 1024
	BytesPerTB = 1024 * 1024 * 1024 * 1024
)

// ParseByteSize parses a human-readable size string (e.g. "2.5GB", "100MB",
// "1.84kB", "3.1KiB", "0B") into bytes. Unit parsing is delegated to
// humanize.ParseBytes, which natively handles all SI and IEC unit suffixes
// (case-insensitive); a number without a unit is treated as bytes.
func ParseByteSize(sizeStr string) (int64, error) {
	parsed, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid size format %q: %w", sizeStr, err)
	}

	return int64(parsed), nil
}
