package nodepackages

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DefaultNodePackageManagerTimeout is the default timeout for package manager commands.
const DefaultNodePackageManagerTimeout = 2 * time.Minute

// AvailableNodePackageManagers returns all available Node.js package managers.
func AvailableNodePackageManagers() []enums.PackageManagerType {
	return []enums.PackageManagerType{
		enums.PackageManagerNpm,
		enums.PackageManagerPnpm,
		enums.PackageManagerYarn,
		enums.PackageManagerBun,
	}
}

// NodePackageManagerCleaner handles Node.js package manager cleanup.
type NodePackageManagerCleaner struct {
	cleaner.CleanerBase

	packageManagers []enums.PackageManagerType
}

// NewNodePackageManagerCleaner creates Node.js package manager cleaner.
func NewNodePackageManagerCleaner(
	verbose, dryRun bool, packageManagers []enums.PackageManagerType,
) *NodePackageManagerCleaner {
	return &NodePackageManagerCleaner{
		CleanerBase:     cleaner.NewCleanerBase(verbose, dryRun),
		packageManagers: packageManagers,
	}
}

// Type returns operation type for Node package manager cleaner.
func (npmc *NodePackageManagerCleaner) Type() operations.OperationType {
	return operations.OperationTypeNodePackages
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
func (npmc *NodePackageManagerCleaner) isPackageManagerAvailable(
	pm enums.PackageManagerType,
) bool {
	switch pm {
	case enums.PackageManagerNpm:
		_, err := exec.LookPath("npm")

		return err == nil
	case enums.PackageManagerPnpm:
		_, err := exec.LookPath("pnpm")

		return err == nil
	case enums.PackageManagerYarn:
		_, err := exec.LookPath("yarn")

		return err == nil
	case enums.PackageManagerBun:
		_, err := exec.LookPath("bun")

		return err == nil
	default:
		return false
	}
}

// ValidateSettings validates Node package manager cleaner settings.
func (npmc *NodePackageManagerCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.NodePackagesSettings { return s.NodePackages },
		func(np *operations.NodePackagesSettings) error {
			packageManagerStrings := cleaner.PackageManagerTypeToLowerSlice(np.PackageManagers)
			validPackageManagersMap := map[string]bool{
				"npm":  true,
				"pnpm": true,
				"yarn": true,
				"bun":  true,
			}

			return cleaner.ValidateStringItems(
				packageManagerStrings,
				validPackageManagersMap,
				"package manager",
				"npm, pnpm, yarn, or bun",
			)
		},
	)
}

// Scan scans for Node.js package manager caches.
func (npmc *NodePackageManagerCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

	for _, pm := range npmc.packageManagers {
		if !npmc.isPackageManagerAvailable(pm) {
			continue
		}

		result := npmc.scanPackageManager(ctx, pm)
		if result.IsErr() {
			if npmc.GetVerbose() {
				fmt.Printf("Warning: failed to scan %s: %v\n", pm, result.Error())
			}

			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}

// scanPackageManager scans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) scanPackageManager(
	ctx context.Context,
	pm enums.PackageManagerType,
) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

	switch pm {
	case enums.PackageManagerNpm:
		// Get npm cache location
		cmd := adapters.ExecWithTimeout(
			ctx,
			DefaultNodePackageManagerTimeout,
			"npm",
			"config",
			"get",
			"cache",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[[]types.ScanItem](
				fmt.Errorf("failed to get npm cache location for pm=%v: %w", pm, err),
			)
		}

		cachePath := strings.TrimSpace(string(output))
		if cachePath != "" {
			items = append(items, types.ScanItem{
				Path:     cachePath,
				Size:     0, // Size unknown without checking
				Created:  time.Time{},
				ScanType: types.ScanTypeTemp,
			})

			if npmc.GetVerbose() {
				fmt.Printf("Found npm cache: %s\n", cachePath)
			}
		}

	case enums.PackageManagerPnpm:
		// Get pnpm store location
		cmd := adapters.ExecWithTimeout(
			ctx,
			DefaultNodePackageManagerTimeout,
			"pnpm",
			"store",
			"path",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[[]types.ScanItem](
				fmt.Errorf("failed to get pnpm store location for pm=%v: %w", pm, err),
			)
		}

		storePath := strings.TrimSpace(string(output))
		if storePath != "" {
			items = append(items, types.ScanItem{
				Path:     storePath,
				Size:     0, // Size unknown without checking
				Created:  time.Time{},
				ScanType: types.ScanTypeTemp,
			})

			if npmc.GetVerbose() {
				fmt.Printf("Found pnpm store: %s\n", storePath)
			}
		}

	case enums.PackageManagerYarn:
		cacheResult := npmc.scanHomeDirCache(ctx, ".yarn/cache", "yarn")
		if cacheResult.IsOk() {
			items = append(items, cacheResult.Value()...)
		} else {
			return cacheResult
		}

	case enums.PackageManagerBun:
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
func (npmc *NodePackageManagerCleaner) scanHomeDirCache(
	_ context.Context,
	cacheSuffix, pmName string,
) result.Result[[]types.ScanItem] {
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return result.Err[[]types.ScanItem](
			fmt.Errorf(
				"failed to get home directory for cacheSuffix=%v, pmName=%v: %w",
				cacheSuffix,
				pmName,
				err,
			),
		)
	}

	cachePath := fmt.Sprintf("%s/%s", homeDir, cacheSuffix)
	items := []types.ScanItem{
		{
			Path:     cachePath,
			Size:     0, // Size unknown without checking
			Created:  time.Time{},
			ScanType: types.ScanTypeTemp,
		},
	}

	if npmc.GetVerbose() {
		fmt.Printf("Found %s cache: %s\n", pmName, cachePath)
	}

	return result.Ok(items)
}

// getNpmCacheDir returns the npm cache directory path.
func (npmc *NodePackageManagerCleaner) getNpmCacheDir(ctx context.Context) (string, error) {
	cmd := adapters.ExecWithTimeout(
		ctx,
		DefaultNodePackageManagerTimeout,
		"npm",
		"config",
		"get",
		"cache",
	)

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
	cmd := adapters.ExecWithTimeout(ctx, DefaultNodePackageManagerTimeout, "pnpm", "store", "path")

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
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return homeDir + "/.yarn/cache", nil
}

// getBunCacheDir returns the bun cache directory path.
func (npmc *NodePackageManagerCleaner) getBunCacheDir() (string, error) {
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return homeDir + "/.bun/install/cache", nil
}

// Clean removes Node.js package manager caches.
func (npmc *NodePackageManagerCleaner) Clean(
	ctx context.Context,
) result.Result[types.CleanResult] {
	if !npmc.IsAvailable(ctx) {
		return result.Err[types.CleanResult](errors.New("no Node.js package managers available"))
	}

	if npmc.GetDryRun() {
		// Scan actual cache directories to get real sizes
		scanResult := npmc.Scan(ctx)

		var (
			totalBytes   int64
			itemsRemoved int
		)

		if scanResult.IsOk() {
			items := scanResult.Value()

			itemsRemoved = len(items)
			for _, item := range items {
				// Get actual size of each cache directory
				totalBytes += cleaner.GetDirSize(item.Path)
			}
		} else {
			// Fallback to counting available package managers if scan fails
			itemsRemoved = len(npmc.packageManagers)
		}

		cleanResult := conversions.NewCleanResult(
			enums.StrategyDryRunType,
			itemsRemoved,
			totalBytes,
		)
		cleanResult.SizeEstimate = types.SizeEstimate{Known: uint64(totalBytes)} //nolint:exhaustruct

		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	counters := cleaner.NewCleanCounters()

	for _, pm := range npmc.packageManagers {
		if !npmc.isPackageManagerAvailable(pm) {
			continue
		}

		result := npmc.cleanPackageManager(ctx, pm)
		if result.IsErr() {
			counters.RecordFailure(npmc.GetVerbose(), pm, result.Error())

			continue
		}

		cleanResult := result.Value()
		counters.RecordSuccess(int64(cleanResult.FreedBytes))
	}

	return result.Ok(conversions.NewCleanResultWithFailures(
		enums.StrategyConservativeType,
		counters.ItemsRemoved, counters.ItemsFailed, counters.BytesFreed, counters.Duration(),
	))
}

// createDefaultCleanResult returns a default CleanResult for package manager operations.
func (npmc *NodePackageManagerCleaner) createDefaultCleanResult() types.CleanResult {
	return conversions.NewCleanResult(
		enums.StrategyConservativeType,
		1, 0,
	)
}

// execPackageManagerCommand executes a package manager command with the given arguments
// and returns an error if the command fails.
func (npmc *NodePackageManagerCleaner) execPackageManagerCommand(
	ctx context.Context,
	commandArgs []string,
	commandName string,
) error {
	cmd := adapters.ExecWithTimeout(
		ctx,
		DefaultNodePackageManagerTimeout,
		commandArgs[0],
		commandArgs[1:]...,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w (output: %s)", commandName, err, string(output))
	}

	return nil
}

// cleanCacheWithFallback handles cache cleaning with directory-based bytes calculation
// and fallback to direct command execution when directory is not found.
func (npmc *NodePackageManagerCleaner) cleanCacheWithFallback(
	ctx context.Context,
	cacheDir string,
	cacheDirErr error,
	commandArgs []string,
	commandName string,
	cacheLabel string,
) result.Result[types.CleanResult] {
	if cacheDirErr != nil {
		// Cache directory not found, execute command anyway
		err := npmc.execPackageManagerCommand(ctx, commandArgs, commandName)
		if err != nil {
			return result.Err[types.CleanResult](
				fmt.Errorf(
					"cacheDirErr=%w, cacheDir=%v, commandName=%v, cacheLabel=%v: %w",
					cacheDirErr,
					cacheDir,
					commandName,
					cacheLabel,
					err,
				),
			)
		}

		if npmc.GetVerbose() {
			fmt.Printf("  ✓ %s\n", commandName)
		}

		return result.Ok(npmc.createDefaultCleanResult())
	}

	bytesFreed, _, _ := cleaner.CalculateBytesFreed(cacheDir, func() error {
		return npmc.execPackageManagerCommand(ctx, commandArgs, commandName)
	}, npmc.GetVerbose(), cacheLabel)

	if npmc.GetVerbose() {
		fmt.Printf("  ✓ %s\n", commandName)
	}

	return result.Ok(conversions.NewCleanResult(
		enums.StrategyConservativeType,
		1, bytesFreed,
	))
}

// cleanPackageManager cleans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) cleanPackageManager(
	ctx context.Context, pm enums.PackageManagerType,
) result.Result[types.CleanResult] {
	switch pm {
	case enums.PackageManagerNpm:
		return npmc.cleanNpmCache(ctx)

	case enums.PackageManagerPnpm:
		return npmc.cleanPnpmStore(ctx)

	case enums.PackageManagerYarn:
		return npmc.cleanYarnCache(ctx)

	case enums.PackageManagerBun:
		return npmc.cleanBunCache(ctx)
	}

	return result.Ok(npmc.createDefaultCleanResult())
}

// cleanNpmCache cleans the npm cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanNpmCache(
	ctx context.Context,
) result.Result[types.CleanResult] {
	cacheDir, err := npmc.getNpmCacheDir(ctx)

	return npmc.cleanCacheWithFallback(ctx, cacheDir, err,
		[]string{"npm", "cache", "clean", "--force"},
		"npm cache clean",
		"Cache")
}

// cleanPnpmStore cleans the pnpm store and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanPnpmStore(
	ctx context.Context,
) result.Result[types.CleanResult] {
	cacheDir, err := npmc.getPnpmStoreDir(ctx)

	return npmc.cleanCacheWithFallback(ctx, cacheDir, err,
		[]string{"pnpm", "store", "prune"}, //nolint:goconst
		"pnpm store prune",
		"Store")
}

// cleanYarnCache cleans the yarn cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanYarnCache(
	ctx context.Context,
) result.Result[types.CleanResult] {
	cacheDir, err := npmc.getYarnCacheDir()

	return npmc.cleanCacheWithFallback(ctx, cacheDir, err,
		[]string{"yarn", "cache", "clean"},
		"yarn cache clean",
		"Cache")
}

// cleanBunCache cleans the bun cache and returns bytes freed.
func (npmc *NodePackageManagerCleaner) cleanBunCache(
	ctx context.Context,
) result.Result[types.CleanResult] {
	cacheDir, err := npmc.getBunCacheDir()

	return npmc.cleanCacheWithFallback(ctx, cacheDir, err,
		[]string{"bun", "pm", "cache", "rm"},
		"bun cache clean",
		"Cache")
}
