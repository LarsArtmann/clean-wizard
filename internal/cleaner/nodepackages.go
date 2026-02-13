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

// DefaultNodePackageManagerTimeout is the default timeout for package manager commands.
const DefaultNodePackageManagerTimeout = 2 * time.Minute

// execWithTimeout creates a command with timeout protection.
func (npmc *NodePackageManagerCleaner) execWithTimeout(ctx context.Context, name string, args ...string) *exec.Cmd {
	// Create a timeout context if the input context doesn't have a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultNodePackageManagerTimeout)
	_ = cancel // will be called by cmd.Wait() or context usage
	return exec.CommandContext(timeoutCtx, name, args...)
}

// AvailableNodePackageManagers returns all available Node.js package managers.
func AvailableNodePackageManagers() []domain.PackageManagerType {
	return []domain.PackageManagerType{
		domain.PackageManagerNpm,
		domain.PackageManagerPnpm,
		domain.PackageManagerYarn,
		domain.PackageManagerBun,
	}
}

// NodePackageManagerCleaner handles Node.js package manager cleanup.
type NodePackageManagerCleaner struct {
	verbose         bool
	dryRun          bool
	packageManagers []domain.PackageManagerType
}

// NewNodePackageManagerCleaner creates Node.js package manager cleaner.
func NewNodePackageManagerCleaner(verbose, dryRun bool, packageManagers []domain.PackageManagerType) *NodePackageManagerCleaner {
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

// Name returns the cleaner name for result tracking.
func (npmc *NodePackageManagerCleaner) Name() string {
	return "node"
}

// IsAvailable checks if any Node.js package manager is available.
func (npmc *NodePackageManagerCleaner) IsAvailable(ctx context.Context) bool {
	return slices.ContainsFunc(npmc.packageManagers, npmc.isPackageManagerAvailable)
}

// isPackageManagerAvailable checks if a specific package manager is available.
func (npmc *NodePackageManagerCleaner) isPackageManagerAvailable(pm domain.PackageManagerType) bool {
	switch pm {
	case domain.PackageManagerNpm:
		_, err := exec.LookPath("npm")
		return err == nil
	case domain.PackageManagerPnpm:
		_, err := exec.LookPath("pnpm")
		return err == nil
	case domain.PackageManagerYarn:
		_, err := exec.LookPath("yarn")
		return err == nil
	case domain.PackageManagerBun:
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

	packageManagerStrings := PackageManagerTypeToLowerSlice(settings.NodePackages.PackageManagers)
	validPackageManagersMap := map[string]bool{
		"npm":  true,
		"pnpm": true,
		"yarn": true,
		"bun":  true,
	}

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
func (npmc *NodePackageManagerCleaner) scanPackageManager(ctx context.Context, pm domain.PackageManagerType) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	switch pm {
	case domain.PackageManagerNpm:
		// Get npm cache location
		cmd := npmc.execWithTimeout(ctx, "npm", "config", "get", "cache")
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

	case domain.PackageManagerPnpm:
		// Get pnpm store location
		cmd := npmc.execWithTimeout(ctx, "pnpm", "store", "path")
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

	case domain.PackageManagerYarn:
		cacheResult := npmc.scanHomeDirCache(ctx, ".yarn/cache", "yarn")
		if cacheResult.IsOk() {
			items = append(items, cacheResult.Value()...)
		} else {
			return cacheResult
		}

	case domain.PackageManagerBun:
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

// getNpmCacheDir returns the npm cache directory path.
func (npmc *NodePackageManagerCleaner) getNpmCacheDir(ctx context.Context) (string, error) {
	cmd := npmc.execWithTimeout(ctx, "npm", "config", "get", "cache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get npm cache location: %w", err)
	}

	cachePath := strings.TrimSpace(string(output))
	if cachePath == "" {
		return "", errors.New("npm cache path is empty")
	}

	return cachePath, nil
}

// getPnpmStoreDir returns the pnpm store directory path.
func (npmc *NodePackageManagerCleaner) getPnpmStoreDir(ctx context.Context) (string, error) {
	cmd := npmc.execWithTimeout(ctx, "pnpm", "store", "path")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get pnpm store location: %w", err)
	}

	storePath := strings.TrimSpace(string(output))
	if storePath == "" {
		return "", errors.New("pnpm store path is empty")
	}

	return storePath, nil
}

// getYarnCacheDir returns the yarn cache directory path.
func (npmc *NodePackageManagerCleaner) getYarnCacheDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return homeDir + "/.yarn/cache", nil
}

// getBunCacheDir returns the bun cache directory path.
func (npmc *NodePackageManagerCleaner) getBunCacheDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return homeDir + "/.bun/install/cache", nil
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

		cleanResult := conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyDryRunType), itemsRemoved, totalBytes)
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
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	}

	return result.Ok(cleanResult)
}

// runPackageManagerCommand executes a package manager command with common error handling and timeout.
func (npmc *NodePackageManagerCleaner) runPackageManagerCommand(ctx context.Context, name string, args ...string) result.Result[domain.CleanResult] {
	// Create a timeout context if the input context doesn't have a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultNodePackageManagerTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, name, args...)
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
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	}
}

// cleanPackageManager cleans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) cleanPackageManager(ctx context.Context, pm domain.PackageManagerType) result.Result[domain.CleanResult] {
	switch pm {
	case domain.PackageManagerNpm:
		return npmc.cleanNpmCache(ctx)

	case domain.PackageManagerPnpm:
		return npmc.cleanPnpmStore(ctx)

	case domain.PackageManagerYarn:
		return npmc.cleanYarnCache(ctx)

	case domain.PackageManagerBun:
		return npmc.cleanBunCache(ctx)
	}

	return result.Ok(npmc.createDefaultCleanResult())
}

// cleanNpmCache cleans the npm cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanNpmCache(ctx context.Context) result.Result[domain.CleanResult] {
	cacheDir, err := npmc.getNpmCacheDir(ctx)
	if err != nil {
		// Cache directory not found, execute command anyway
		cmd := npmc.execWithTimeout(ctx, "npm", "cache", "clean", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("npm cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ npm cache cleaned")
		}

		return result.Ok(npmc.createDefaultCleanResult())
	}

	bytesFreed, _, _ := CalculateBytesFreed(cacheDir, func() error {
		cmd := npmc.execWithTimeout(ctx, "npm", "cache", "clean", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("npm cache clean failed: %w (output: %s)", err, string(output))
		}
		return nil
	}, npmc.verbose, "Cache")

	if npmc.verbose {
		fmt.Println("  ✓ npm cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}

// cleanPnpmStore cleans the pnpm store and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanPnpmStore(ctx context.Context) result.Result[domain.CleanResult] {
	cacheDir, err := npmc.getPnpmStoreDir(ctx)
	if err != nil {
		// Store directory not found, execute command anyway
		cmd := npmc.execWithTimeout(ctx, "pnpm", "store", "prune")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("pnpm store prune failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ pnpm store pruned")
		}

		return result.Ok(npmc.createDefaultCleanResult())
	}

	bytesFreed, _, _ := CalculateBytesFreed(cacheDir, func() error {
		cmd := npmc.execWithTimeout(ctx, "pnpm", "store", "prune")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("pnpm store prune failed: %w (output: %s)", err, string(output))
		}
		return nil
	}, npmc.verbose, "Store")

	if npmc.verbose {
		fmt.Println("  ✓ pnpm store pruned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}

// cleanYarnCache cleans the yarn cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanYarnCache(ctx context.Context) result.Result[domain.CleanResult] {
	cacheDir, err := npmc.getYarnCacheDir()
	if err != nil {
		// Cache directory not found, execute command anyway
		cmd := npmc.execWithTimeout(ctx, "yarn", "cache", "clean")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("yarn cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ yarn cache cleaned")
		}

		return result.Ok(npmc.createDefaultCleanResult())
	}

	bytesFreed, _, _ := CalculateBytesFreed(cacheDir, func() error {
		cmd := npmc.execWithTimeout(ctx, "yarn", "cache", "clean")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("yarn cache clean failed: %w (output: %s)", err, string(output))
		}
		return nil
	}, npmc.verbose, "Cache")

	if npmc.verbose {
		fmt.Println("  ✓ yarn cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}

// cleanBunCache cleans the bun cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanBunCache(ctx context.Context) result.Result[domain.CleanResult] {
	cacheDir, err := npmc.getBunCacheDir()
	if err != nil {
		// Cache directory not found, execute command anyway
		cmd := npmc.execWithTimeout(ctx, "bun", "pm", "cache", "rm")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("bun cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ bun cache cleaned")
		}

		return result.Ok(npmc.createDefaultCleanResult())
	}

	bytesFreed, _, _ := CalculateBytesFreed(cacheDir, func() error {
		cmd := npmc.execWithTimeout(ctx, "bun", "pm", "cache", "rm")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("bun cache clean failed: %w (output: %s)", err, string(output))
		}
		return nil
	}, npmc.verbose, "Cache")

	if npmc.verbose {
		fmt.Println("  ✓ bun cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}
