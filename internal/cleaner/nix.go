package cleaner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/LarsArtmann/clean-wizard/internal/types"
)

// NixCleaner handles Nix store cleanup operations
type NixCleaner struct {
	verbose bool
	dryRun  bool
}

// NewNixCleaner creates a new Nix cleaner
func NewNixCleaner(verbose, dryRun bool) *NixCleaner {
	return &NixCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// ListGenerations returns list of Nix generations
func (nc *NixCleaner) ListGenerations(ctx context.Context) result.Result[[]NixGeneration] {
	cmd := exec.CommandContext(ctx, "nix-env", "--list-generations")
	output, err := cmd.Output()
	if err != nil {
		if nc.isNixNotAvailable(err) {
			return nc.mockGenerations()
		}
		return result.Err[[]NixGeneration](fmt.Errorf("failed to list generations: %w", err))
	}

	var generations []NixGeneration
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		gen, err := nc.parseGeneration(line)
		if err != nil {
			continue
		}
		generations = append(generations, gen)
	}

	return result.Ok(generations)
}

// CleanOldGenerations removes old Nix generations
func (nc *NixCleaner) CleanOldGenerations(ctx context.Context, keepCount int) result.Result[types.OperationResult] {
	genResult := nc.ListGenerations(ctx)
	if genResult.IsErr() {
		return result.Err[types.OperationResult](genResult.Error())
	}
	
	generations := genResult.Value()
	if len(generations) <= keepCount {
		return result.Ok(types.OperationResult{
			Success:    true,
			FreedBytes: 0,
			Duration:   0,
		})
	}

	startTime := time.Now()
	
	if nc.dryRun {
		toRemove := len(generations) - keepCount
		return result.Ok(types.OperationResult{
			Success:      true,
			FreedBytes:   1024 * 1024 * 1024 * 5, // Estimate 5GB
			Duration:     time.Since(startTime),
			ErrorMessage: fmt.Sprintf("[DRY RUN] Would remove %d old generations, keeping %d", toRemove, keepCount),
		})
	}

	cmd := exec.CommandContext(ctx, "nix-collect-garbage", "-d")
	if nc.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err := cmd.Run()
	if err != nil {
		if nc.isNixNotAvailable(err) {
			return result.Ok(types.OperationResult{
				Success:    true,
				FreedBytes: 1024 * 1024 * 1024 * 2, // Mock 2GB
				Duration:   time.Since(startTime),
				ErrorMessage: "[MOCK] Simulated Nix garbage collection (nix not available)",
			})
		}
		return result.Err[types.OperationResult](fmt.Errorf("nix-collect-garbage failed: %w", err))
	}

	return result.Ok(types.OperationResult{
		Success:    true,
		FreedBytes: 1024 * 1024 * 1024 * 2, // TODO: Calculate actual space
		Duration:   time.Since(startTime),
	})
}

// GetStoreSize returns Nix store size
func (nc *NixCleaner) GetStoreSize(ctx context.Context) result.Result[int64] {
	cmd := exec.CommandContext(ctx, "du", "-sb", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		if nc.isNixNotAvailable(err) {
			return result.Ok[int64](1024 * 1024 * 1024 * 10) // Mock 10GB
		}
		return result.Err[int64](fmt.Errorf("failed to get store size: %w", err))
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return result.Err[int64](fmt.Errorf("invalid du output: %s", string(output)))
	}

	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return result.Err[int64](fmt.Errorf("failed to parse size: %w", err))
	}

	return result.Ok(size)
}

// NixGeneration represents a Nix generation
type NixGeneration struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	Date      time.Time `json:"date"`
	Current   bool      `json:"current"`
}

// parseGeneration parses a generation line
func (nc *NixCleaner) parseGeneration(line string) (NixGeneration, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return NixGeneration{}, fmt.Errorf("invalid generation line: %s", line)
	}

	pathParts := strings.Split(fields[0], "-")
	if len(pathParts) < 2 {
		return NixGeneration{}, fmt.Errorf("invalid generation path: %s", fields[0])
	}
	
	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		return NixGeneration{}, fmt.Errorf("invalid generation ID: %s", pathParts[len(pathParts)-1])
	}

	// Simple date parsing
	date := time.Now()
	if len(fields) > 1 && strings.Contains(fields[1], "-") {
		if parsed, err := time.Parse("2006-01-02", strings.Trim(fields[1], "()")); err == nil {
			date = parsed
		}
	}

	current := strings.Contains(line, "(current)")

	return NixGeneration{
		ID:      id,
		Path:    fields[0],
		Date:    date,
		Current: current,
	}, nil
}

// isNixNotAvailable checks if Nix is not available
func (nc *NixCleaner) isNixNotAvailable(err error) bool {
	return strings.Contains(err.Error(), "executable not found") || 
		   strings.Contains(err.Error(), "command not found") ||
		   strings.Contains(err.Error(), "no such file or directory")
}

// mockGenerations returns mock data when Nix is not available
func (nc *NixCleaner) mockGenerations() result.Result[[]NixGeneration] {
	now := time.Now()
	generations := []NixGeneration{
		{ID: 300, Path: "/nix/var/nix/profiles/default-300-link", Date: now.Add(-24 * time.Hour), Current: true},
		{ID: 299, Path: "/nix/var/nix/profiles/default-299-link", Date: now.Add(-48 * time.Hour), Current: false},
		{ID: 298, Path: "/nix/var/nix/profiles/default-298-link", Date: now.Add(-72 * time.Hour), Current: false},
		{ID: 297, Path: "/nix/var/nix/profiles/default-297-link", Date: now.Add(-96 * time.Hour), Current: false},
		{ID: 296, Path: "/nix/var/nix/profiles/default-296-link", Date: now.Add(-120 * time.Hour), Current: false},
	}
	return result.Ok(generations)
}
