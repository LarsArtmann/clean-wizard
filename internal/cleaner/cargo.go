package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// CargoCleaner handles Rust/Cargo cleanup.

const cargoCommandTimeout = 5 * time.Minute

// execWithTimeout executes a Cargo command with timeout.
func (cc *CargoCleaner) execWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeoutCtx, cancel := context.WithTimeout(ctx, cargoCommandTimeout)
	_ = cancel // will be called by cmd.Wait() or context usage
	return exec.CommandContext(timeoutCtx, name, arg...)
}

type CargoCleaner struct {
	verbose bool
	dryRun  bool
}

// NewCargoCleaner creates Cargo cleaner.
func NewCargoCleaner(verbose, dryRun bool) *CargoCleaner {
	return &CargoCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// Type returns operation type for Cargo cleaner.
func (cc *CargoCleaner) Type() domain.OperationType {
	return domain.OperationTypeCargoPackages
}

// IsAvailable checks if Cargo is available.
func (cc *CargoCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("cargo")
	return err == nil
}

// ValidateSettings validates Cargo cleaner settings.
func (cc *CargoCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.CargoPackages == nil {
		return nil // Settings are optional
	}

	// All Cargo settings are valid by default
	return nil
}

// Scan scans for Cargo caches.
func (cc *CargoCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	// Get CARGO_HOME environment variable
	cargoHome := os.Getenv("CARGO_HOME")
	if cargoHome == "" {
		// Fallback to default ~/.cargo
		if homeDir, err := GetHomeDir(); err == nil && homeDir != "" {
			cargoHome = fmt.Sprintf("%s/.cargo", homeDir)
		}
	}

	if cargoHome != "" {
		// Add registry cache location
		registryCache := fmt.Sprintf("%s/registry", cargoHome)
		items = append(items, domain.ScanItem{
			Path:     registryCache,
			Size:     GetDirSize(registryCache),
			Created:  GetDirModTime(registryCache),
			ScanType: domain.ScanTypeTemp,
		})

		if cc.verbose {
			fmt.Printf("Found Cargo registry cache: %s\n", registryCache)
		}

		// Add source cache location
		sourceCache := fmt.Sprintf("%s/git", cargoHome)
		items = append(items, domain.ScanItem{
			Path:     sourceCache,
			Size:     GetDirSize(sourceCache),
			Created:  GetDirModTime(sourceCache),
			ScanType: domain.ScanTypeTemp,
		})

		if cc.verbose {
			fmt.Printf("Found Cargo source cache: %s\n", sourceCache)
		}
	}

	return result.Ok(items)
}

// Clean removes Cargo caches.
func (cc *CargoCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !cc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("Cargo not available"))
	}

	if cc.dryRun {
		// Estimate cache sizes based on typical usage
		totalBytes := int64(500 * 1024 * 1024) // Estimate 500MB for Cargo
		itemsRemoved := 1

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	bytesFreed := int64(0)

	// Check if cargo-cache tool is available (optional extension)
	if cc.hasCargoCacheTool() {
		cacheToolResult := cc.cleanWithCargoCacheTool(ctx)
		if cacheToolResult.IsErr() {
			if cc.verbose {
				fmt.Printf("Warning: cargo-cache tool failed, falling back to manual clean: %v\n", cacheToolResult.Error())
			}
			// Fall through to manual clean
		} else {
			cacheCleanResult := cacheToolResult.Value()
			itemsRemoved++
			bytesFreed += int64(cacheCleanResult.FreedBytes)

			duration := time.Since(startTime)
			finalResult := domain.CleanResult{
				FreedBytes:   uint64(bytesFreed),
				ItemsRemoved: uint(itemsRemoved),
				ItemsFailed:  0,
				CleanTime:    duration,
				CleanedAt:    time.Now(),
				Strategy:     domain.StrategyConservative,
			}
			return result.Ok(finalResult)
		}
	}

	// Manual cleanup using cargo clean command
	cleanResult := cc.cleanWithCargoClean(ctx)
	if cleanResult.IsErr() {
		return result.Err[domain.CleanResult](fmt.Errorf("cargo clean failed: %w", cleanResult.Error()))
	}

	itemsRemoved++
	bytesFreed += int64(cleanResult.Value().FreedBytes)

	duration := time.Since(startTime)
	finalResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  0,
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(finalResult)
}

// cleanWithCargoCacheTool cleans using cargo-cache extension.
func (cc *CargoCleaner) cleanWithCargoCacheTool(ctx context.Context) result.Result[domain.CleanResult] {
	return cc.executeCargoCleanCommand(
		ctx,
		"cargo-cache", []string{"--autoclean"},
		"cargo-cache --autoclean failed: %w (output: %s)",
		"  ✓ Cargo cache cleaned with cargo-cache tool",
	)
}

// cleanWithCargoClean cleans using standard cargo clean command.
func (cc *CargoCleaner) cleanWithCargoClean(ctx context.Context) result.Result[domain.CleanResult] {
	return cc.executeCargoCleanCommand(
		ctx,
		"cargo", []string{"clean"},
		"cargo clean failed: %w (output: %s)",
		"  ✓ Cargo cache cleaned",
	)
}

// executeCargoCleanCommand is a helper that executes a cargo command and returns a clean result.
func (cc *CargoCleaner) executeCargoCleanCommand(
	ctx context.Context,
	cmdName string,
	args []string,
	errorFormat string,
	successMessage string,
) result.Result[domain.CleanResult] {
	cmd := cc.execWithTimeout(ctx, cmdName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf(errorFormat, err, string(output)))
	}

	if cc.verbose {
		fmt.Println(successMessage)
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   0,
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	})
}

// hasCargoCacheTool checks if cargo-cache tool is available.
func (cc *CargoCleaner) hasCargoCacheTool() bool {
	_, err := exec.LookPath("cargo-cache")
	return err == nil
}
