package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// NodePackageManagerType represents different Node.js package managers.
type NodePackageManagerType string

const (
	NodePackageManagerNPM  NodePackageManagerType = "npm"
	NodePackageManagerPNPM NodePackageManagerType = "pnpm"
	NodePackageManagerYarn NodePackageManagerType = "yarn"
	NodePackageManagerBun  NodePackageManagerType = "bun"
)

// AvailableNodePackageManagers returns all available Node.js package managers.
func AvailableNodePackageManagers() []NodePackageManagerType {
	return []NodePackageManagerType{
		NodePackageManagerNPM,
		NodePackageManagerPNPM,
		NodePackageManagerYarn,
		NodePackageManagerBun,
	}
}

// NodePackageManagerCleaner handles Node.js package manager cleanup.
type NodePackageManagerCleaner struct {
	verbose         bool
	dryRun          bool
	packageManagers []NodePackageManagerType
}

// NewNodePackageManagerCleaner creates Node.js package manager cleaner.
func NewNodePackageManagerCleaner(verbose, dryRun bool, packageManagers []NodePackageManagerType) *NodePackageManagerCleaner {
	return &NodePackageManagerCleaner{
		verbose:         verbose,
		dryRun:          dryRun,
		packageManagers: packageManagers,
	}
}

// Type returns operation type for Node package manager cleaner.
func (npmc *NodePackageManagerCleaner) Type() domain.OperationType {
	return domain.OperationTypeNodePackages
}

// IsAvailable checks if any Node.js package manager is available.
func (npmc *NodePackageManagerCleaner) IsAvailable(ctx context.Context) bool {
	return slices.ContainsFunc(npmc.packageManagers, npmc.isPackageManagerAvailable)
}

// isPackageManagerAvailable checks if a specific package manager is available.
func (npmc *NodePackageManagerCleaner) isPackageManagerAvailable(pm NodePackageManagerType) bool {
	switch pm {
	case NodePackageManagerNPM:
		_, err := exec.LookPath("npm")
		return err == nil
	case NodePackageManagerPNPM:
		_, err := exec.LookPath("pnpm")
		return err == nil
	case NodePackageManagerYarn:
		_, err := exec.LookPath("yarn")
		return err == nil
	case NodePackageManagerBun:
		_, err := exec.LookPath("bun")
		return err == nil
	default:
		return false
	}
}

// ValidateSettings validates Node package manager cleaner settings.
func (npmc *NodePackageManagerCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.NodePackages == nil {
		return nil // Settings are optional
	}

	validPackageManagersMap := toStringMap(AvailableNodePackageManagers())
	packageManagerStrings := PackageManagerTypeToLowerSlice(settings.NodePackages.PackageManagers)

	return validateSettings(
		packageManagerStrings,
		validPackageManagersMap,
		"package manager",
		"npm, pnpm, yarn, or bun",
	)
}

// Scan scans for Node.js package manager caches.
func (npmc *NodePackageManagerCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	for _, pm := range npmc.packageManagers {
		if !npmc.isPackageManagerAvailable(pm) {
			continue
		}

		result := npmc.scanPackageManager(ctx, pm)
		if result.IsErr() {
			if npmc.verbose {
				fmt.Printf("Warning: failed to scan %s: %v\n", pm, result.Error())
			}
			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}

// scanPackageManager scans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) scanPackageManager(ctx context.Context, pm NodePackageManagerType) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch pm {
	case NodePackageManagerNPM:
		// Get npm cache location
		cmd := exec.CommandContext(ctx, "npm", "config", "get", "cache")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get npm cache location: %w", err))
		}

		cachePath := strings.TrimSpace(string(output))
		if cachePath != "" {
			items = append(items, domain.ScanItem{
				Path:     cachePath,
				Size:     0, // Size unknown without checking
				Created:  time.Time{},
				ScanType: domain.ScanTypeTemp,
			})

			if npmc.verbose {
				fmt.Printf("Found npm cache: %s\n", cachePath)
			}
		}

	case NodePackageManagerPNPM:
		// Get pnpm store location
		cmd := exec.CommandContext(ctx, "pnpm", "store", "path")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get pnpm store location: %w", err))
		}

		storePath := strings.TrimSpace(string(output))
		if storePath != "" {
			items = append(items, domain.ScanItem{
				Path:     storePath,
				Size:     0, // Size unknown without checking
				Created:  time.Time{},
				ScanType: domain.ScanTypeTemp,
			})

			if npmc.verbose {
				fmt.Printf("Found pnpm store: %s\n", storePath)
			}
		}

	case NodePackageManagerYarn:
		cacheResult := npmc.scanHomeDirCache(ctx, ".yarn/cache", "yarn")
		if cacheResult.IsOk() {
			items = append(items, cacheResult.Value()...)
		} else {
			return cacheResult
		}

	case NodePackageManagerBun:
		cacheResult := npmc.scanHomeDirCache(ctx, ".bun/install/cache", "bun")
		if cacheResult.IsOk() {
			items = append(items, cacheResult.Value()...)
		} else {
			return cacheResult
		}
	}

	return result.Ok(items)
}

// scanHomeDirCache scans a cache directory located under the home directory.
func (npmc *NodePackageManagerCleaner) scanHomeDirCache(ctx context.Context, cacheSuffix, pmName string) result.Result[[]domain.ScanItem] {
	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
	}

	cachePath := fmt.Sprintf("%s/%s", homeDir, cacheSuffix)
	items := []domain.ScanItem{
		{
			Path:     cachePath,
			Size:     0, // Size unknown without checking
			Created:  time.Time{},
			ScanType: domain.ScanTypeTemp,
		},
	}

	if npmc.verbose {
		fmt.Printf("Found %s cache: %s\n", pmName, cachePath)
	}

	return result.Ok(items)
}

// Clean removes Node.js package manager caches.
func (npmc *NodePackageManagerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !npmc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](errors.New("no Node.js package managers available"))
	}

	if npmc.dryRun {
		// Dry-run not fully supported for all package managers
		// Estimate cache sizes based on typical usage
		totalBytes := int64(100 * 1024 * 1024) // Estimate 100MB per available PM
		itemsRemoved := len(npmc.packageManagers)

		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, itemsRemoved, totalBytes)
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	for _, pm := range npmc.packageManagers {
		if !npmc.isPackageManagerAvailable(pm) {
			continue
		}

		result := npmc.cleanPackageManager(ctx, pm)
		if result.IsErr() {
			itemsFailed++
			if npmc.verbose {
				fmt.Printf("Warning: failed to clean %s: %v\n", pm, result.Error())
			}
			continue
		}

		cleanResult := result.Value()
		itemsRemoved++
		bytesFreed += int64(cleanResult.FreedBytes)
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanResult)
}

// runPackageManagerCommand executes a package manager command with common error handling.
func (npmc *NodePackageManagerCleaner) runPackageManagerCommand(ctx context.Context, name string, args ...string) result.Result[domain.CleanResult] {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("%s command failed: %w (output: %s)", name, err, string(output)))
	}

	if npmc.verbose {
		fmt.Printf("  ✓ %s command completed\n", name)
	}

	return result.Ok(npmc.createDefaultCleanResult())
}

// createDefaultCleanResult returns a default CleanResult for package manager operations.
func (npmc *NodePackageManagerCleaner) createDefaultCleanResult() domain.CleanResult {
	return domain.CleanResult{
		FreedBytes:   0,
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}
}

// cleanPackageManager cleans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) cleanPackageManager(ctx context.Context, pm NodePackageManagerType) result.Result[domain.CleanResult] {
	switch pm {
	case NodePackageManagerNPM:
		return npmc.runPackageManagerCommand(ctx, "npm", "cache", "clean", "--force")

	case NodePackageManagerPNPM:
		return npmc.runPackageManagerCommand(ctx, "pnpm", "store", "prune")

	case NodePackageManagerYarn:
		return npmc.runPackageManagerCommand(ctx, "yarn", "cache", "clean")

	case NodePackageManagerBun:
		return npmc.runPackageManagerCommand(ctx, "bun", "pm", "cache", "rm")
	}

	return result.Ok(npmc.createDefaultCleanResult())
}
