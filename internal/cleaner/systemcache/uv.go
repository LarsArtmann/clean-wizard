package systemcache

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// uvBinary is the uv executable looked up on PATH.
const uvBinary = "uv"

// uvCommandTimeout bounds `uv cache dir` and `uv cache clean` runs. uv blocks
// cache-mutating commands on its cache lock while other uv processes run, so
// the timeout doubles as a bounded wait instead of an unbounded stall; a
// timeout surfaces as a Transient failure and is retried by the workflow.
const uvCommandTimeout = 60 * time.Second

// scanUvCache scans uv's cache, preferring the location resolved by the uv
// binary (UV_CACHE_DIR / XDG_CACHE_HOME aware) over the static default path.
func (scc *SystemCacheCleaner) scanUvCache(
	ctx context.Context,
	homeDir string,
) result.Result[[]types.ScanItem] {
	uvConfig := systemCacheConfigs[enums.CacheTypeUv]

	cacheDir, err := uvCacheDir(ctx)
	if err != nil {
		if scc.GetVerbose() {
			fmt.Printf(
				"Warning: failed to resolve uv cache dir (%v), falling back to static path: %s\n",
				err,
				staticUvCachePath(homeDir),
			)
		}

		return scc.scanCachePathWithConfig(ctx, homeDir, uvConfig)
	}

	scanResult := cleaner.ScanPath(
		"",
		uvConfig.scanType,
		uvConfig.displayName,
		scc.GetVerbose(),
		"",
		cacheDir,
	)

	return result.Ok(scanResult.Items)
}

// cleanUvCache cleans uv's cache. When the uv binary is installed it delegates
// to `uv cache clean`, which respects UV_CACHE_DIR/XDG_CACHE_HOME and uv's
// cache lock (uv documents direct cache modification as unsafe). Without the
// binary the cache is unmanaged, so it falls back to removing the static
// default path like the other system caches.
func (scc *SystemCacheCleaner) cleanUvCache(
	ctx context.Context,
	homeDir string,
) result.Result[types.CleanResult] {
	if !uvBinaryAvailable() {
		return scc.removeCachePath(
			staticUvCachePath(homeDir),
			systemCacheConfigs[enums.CacheTypeUv].displayName+" cleaned",
		)
	}

	cacheDir, dirErr := uvCacheDir(ctx)
	beforeSize := int64(0)
	if dirErr == nil {
		beforeSize = cleaner.GetDirSize(cacheDir)
	}

	if err := runUvCacheClean(ctx); err != nil {
		return result.Err[types.CleanResult](err)
	}

	bytesFreed := int64(0)
	sizeEstimate := types.SizeEstimate{Status: enums.SizeEstimateStatusUnknown} //nolint:exhaustruct
	if dirErr == nil {
		afterSize := cleaner.GetDirSize(cacheDir)
		bytesFreed = max(beforeSize-afterSize, 0)
		sizeEstimate = types.SizeEstimate{
			Known:  uint64(bytesFreed),
			Status: enums.SizeEstimateStatusKnown,
		}
	}

	if scc.GetVerbose() {
		fmt.Printf(
			"  ✓ %s cleaned via uv cache clean (%s freed)\n",
			systemCacheConfigs[enums.CacheTypeUv].displayName,
			format.Bytes(bytesFreed),
		)
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		1, bytesFreed,
		sizeEstimate,
	))
}

// uvBinaryAvailable reports whether the uv executable is on PATH.
func uvBinaryAvailable() bool {
	_, err := exec.LookPath(uvBinary)

	return err == nil
}

// uvCacheDir resolves uv's cache directory via `uv cache dir`.
func uvCacheDir(ctx context.Context) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, uvCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, uvBinary, "cache", "dir")

	output, err := cmd.Output()
	if err != nil {
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("uv cache dir timed out after %v", uvCommandTimeout)
		}

		return "", fmt.Errorf("uv cache dir failed: %w (output: %s)", err, string(output))
	}

	return parseUvCacheDirOutput(string(output))
}

// parseUvCacheDirOutput extracts the cache directory from `uv cache dir`
// stdout: the first non-empty line, which must be an absolute path.
func parseUvCacheDirOutput(output string) (string, error) {
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		dir := strings.TrimSpace(line)
		if dir == "" {
			continue
		}

		if !filepath.IsAbs(dir) {
			return "", fmt.Errorf("uv cache dir is not an absolute path: %q", dir)
		}

		return dir, nil
	}

	return "", fmt.Errorf("uv cache dir produced no path (output: %q)", output)
}

// runUvCacheClean executes `uv cache clean`, which clears all cache entries
// under uv's cache lock.
func runUvCacheClean(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, uvCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, uvBinary, "cache", "clean")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf(
				"uv cache clean timed out after %v (another uv process may hold the cache lock)",
				uvCommandTimeout,
			)
		}

		return fmt.Errorf("uv cache clean failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// staticUvCachePath is the static default cache path ($HOME/.cache/uv) used
// when the uv binary is unavailable.
func staticUvCachePath(homeDir string) string {
	return filepath.Join(
		append([]string{homeDir}, systemCacheConfigs[enums.CacheTypeUv].pathComponents...)...,
	)
}
