package nodepackages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// This file carries the Node package manager cleaner's discovery side:
// per-manager scans and cache/store directory resolution. The cleaner
// lifecycle (Scan/Clean entry points, cache purge execution) lives in
// nodepackages.go.

// scanPackageManager scans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) scanPackageManager(
	ctx context.Context,
	pm enums.PackageManagerType,
) result.Result[[]types.ScanItem] {
	switch pm {
	case enums.PackageManagerNpm:
		cachePath, err := npmc.getNpmCacheDir(ctx)
		if err != nil {
			return result.Err[[]types.ScanItem](
				fmt.Errorf("failed to scan npm for pm=%v: %w", pm, err),
			)
		}

		if npmc.GetVerbose() {
			fmt.Printf("Found npm cache: %s\n", cachePath)
		}

		return result.Ok([]types.ScanItem{newTempScanItem(cachePath)})

	case enums.PackageManagerPnpm:
		storePath, err := npmc.getPnpmStoreDir(ctx)
		if err != nil {
			return result.Err[[]types.ScanItem](
				fmt.Errorf("failed to scan pnpm for pm=%v: %w", pm, err),
			)
		}

		if npmc.GetVerbose() {
			fmt.Printf("Found pnpm store: %s\n", storePath)
		}

		return result.Ok([]types.ScanItem{newTempScanItem(storePath)})

	case enums.PackageManagerYarn:
		return npmc.scanHomeDirCache(ctx, ".yarn/cache", "yarn")

	case enums.PackageManagerBun:
		return npmc.scanHomeDirCache(ctx, ".bun/install/cache", "bun")
	}

	return result.Ok([]types.ScanItem{})
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

	if npmc.GetVerbose() {
		fmt.Printf("Found %s cache: %s\n", pmName, cachePath)
	}

	return result.Ok([]types.ScanItem{newTempScanItem(cachePath)})
}

// newTempScanItem builds a ScanItem for a temp cache path of unknown size.
func newTempScanItem(path string) types.ScanItem {
	return types.ScanItem{
		Path:     path,
		Size:     0, // Size unknown without checking
		Created:  time.Time{},
		ScanType: types.ScanTypeTemp,
	}
}

// getNpmCacheDir returns the npm cache directory path.
func (npmc *NodePackageManagerCleaner) getNpmCacheDir(ctx context.Context) (string, error) {
	return npmc.getCommandCacheDir(ctx, "npm cache location", "npm", "config", "get", "cache")
}

// getPnpmStoreDir returns the pnpm store directory path.
func (npmc *NodePackageManagerCleaner) getPnpmStoreDir(ctx context.Context) (string, error) {
	return npmc.getCommandCacheDir(ctx, "pnpm store location", "pnpm", "store", "path")
}

// getCommandCacheDir runs a package-manager command that prints its cache/store
// location, returning the trimmed non-empty path.
func (npmc *NodePackageManagerCleaner) getCommandCacheDir(
	ctx context.Context,
	what string,
	commandArgs ...string,
) (string, error) {
	cmd := adapters.ExecWithTimeout(ctx, DefaultNodePackageManagerTimeout, commandArgs[0], commandArgs[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get %s: %w", what, err)
	}

	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", fmt.Errorf("%s is empty", what)
	}

	return path, nil
}

// getHomeDirCacheDir returns a cache directory under the home directory.
func (npmc *NodePackageManagerCleaner) getHomeDirCacheDir(cacheSuffix string) (string, error) {
	homeDir, err := cleaner.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return homeDir + "/" + cacheSuffix, nil
}

// getYarnCacheDir returns the yarn cache directory path.
func (npmc *NodePackageManagerCleaner) getYarnCacheDir() (string, error) {
	return npmc.getHomeDirCacheDir(".yarn/cache")
}

// getBunCacheDir returns the bun cache directory path.
func (npmc *NodePackageManagerCleaner) getBunCacheDir() (string, error) {
	return npmc.getHomeDirCacheDir(".bun/install/cache")
}
