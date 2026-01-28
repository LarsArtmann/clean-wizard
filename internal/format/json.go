package format

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// JSONOutput represents the JSON structure for clean command output.
type JSONOutput struct {
	Success      bool            `json:"success"`
	CleanedAt    time.Time       `json:"cleaned_at"`
	DurationMs   int64           `json:"duration_ms"`
	ItemsRemoved uint            `json:"items_removed"`
	ItemsFailed  uint            `json:"items_failed"`
	FreedBytes   uint64          `json:"freed_bytes"`
	FreedHuman   string          `json:"freed_human"`
	Cleaners     []CleanerResult `json:"cleaners"`
	DryRun       bool            `json:"dry_run,omitempty"`
	Errors       []string        `json:"errors,omitempty"`
}

// CleanerResult represents individual cleaner results in JSON output.
type CleanerResult struct {
	Name         string `json:"name"`
	ItemsRemoved uint   `json:"items_removed"`
	ItemsFailed  uint   `json:"items_failed"`
	FreedBytes   uint64 `json:"freed_bytes"`
	FreedHuman   string `json:"freed_human"`
	Status       string `json:"status"` // "success", "skipped", "failed"
	Error        string `json:"error,omitempty"`
}

// CleanResultsToJSON converts clean results to JSON output format.
func CleanResultsToJSON(results map[string]domain.CleanResult, duration time.Duration, dryRun bool, skipped, failed map[string]error) ([]byte, error) {
	output := JSONOutput{
		Success:      len(failed) == 0,
		CleanedAt:    time.Now(),
		DurationMs:   duration.Milliseconds(),
		ItemsRemoved: 0,
		ItemsFailed:  0,
		FreedBytes:   0,
		DryRun:       dryRun,
		Cleaners:     make([]CleanerResult, 0, len(results)),
		Errors:       make([]string, 0),
	}

	// Process successful cleaners
	for name, result := range results {
		output.ItemsRemoved += result.ItemsRemoved
		output.ItemsFailed += result.ItemsFailed
		output.FreedBytes += result.FreedBytes

		cleanerResult := CleanerResult{
			Name:         name,
			ItemsRemoved: result.ItemsRemoved,
			ItemsFailed:  result.ItemsFailed,
			FreedBytes:   result.FreedBytes,
			FreedHuman:   Bytes(int64(result.FreedBytes)),
			Status:       "success",
		}

		output.Cleaners = append(output.Cleaners, cleanerResult)
	}

	// Process skipped cleaners
	for name, err := range skipped {
		output.Cleaners = append(output.Cleaners, CleanerResult{
			Name:   name,
			Status: "skipped",
			Error:  err.Error(),
		})
	}

	// Process failed cleaners
	for name, err := range failed {
		output.Cleaners = append(output.Cleaners, CleanerResult{
			Name:   name,
			Status: "failed",
			Error:  err.Error(),
		})
		output.Errors = append(output.Errors, fmt.Sprintf("%s: %v", name, err))
		output.Success = false
	}

	// Set human-readable freed space
	output.FreedHuman = Bytes(int64(output.FreedBytes))

	return json.MarshalIndent(output, "", "  ")
}
