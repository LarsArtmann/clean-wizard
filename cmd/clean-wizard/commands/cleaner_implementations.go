package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// runCleaner runs a specific cleaner and returns the result.
func runCleaner(
	ctx context.Context,
	cleanerType CleanerType,
	dryRun, verbose bool,
) (domain.CleanResult, error) {
	name := getCleanerName(cleanerType)
	_ = name // Avoid unused variable warning

	var (
		result domain.CleanResult
		err    error
	)

	switch cleanerType {
	case CleanerTypeNix:
		result, err = runNixCleaner(ctx, dryRun, verbose)
	case CleanerTypeHomebrew:
		result, err = runHomebrewCleaner(ctx, dryRun, verbose)
	case CleanerTypeTempFiles:
		result, err = runTempFilesCleaner(ctx, dryRun, verbose)
	case CleanerTypeNodePackages:
		result, err = runNodePackageManagerCleaner(ctx, dryRun, verbose)
	case CleanerTypeGoPackages:
		result, err = runGoCleaner(ctx, dryRun, verbose)
	case CleanerTypeCargoPackages:
		result, err = runCargoCleaner(ctx, dryRun, verbose)
	case CleanerTypeBuildCache:
		result, err = runBuildCacheCleaner(ctx, dryRun, verbose)
	case CleanerTypeDocker:
		result, err = runDockerCleaner(ctx, dryRun, verbose)
	case CleanerTypeSystemCache:
		result, err = runSystemCacheCleaner(ctx, dryRun, verbose)
	case CleanerTypeProjectsManagementAutomation:
		result, err = runProjectsManagementAutomationCleaner(ctx, dryRun, verbose)
	case CleanerTypeCompiledBinaries:
		result, err = runCompiledBinariesCleaner(ctx, dryRun, verbose)
	case CleanerTypeLangVersionMgr:
		return domain.CleanResult{}, errors.New("langversion cleaner is not implemented yet")
	default:
		return domain.CleanResult{}, errors.New("unknown cleaner type: " + string(cleanerType))
	}

	if err != nil {
		return domain.CleanResult{}, err
	}

	return result, nil
}

// runNixCleaner executes the Nix cleaner.
func runNixCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	nixAdapter := cleaner.NewNixCleaner(verbose, dryRun)

	if !nixAdapter.IsAvailable(ctx) {
		return domain.CleanResult{}, errors.New("nix not available on this system")
	}

	keepCount := 5
	result := nixAdapter.CleanOldGenerations(ctx, keepCount)

	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	return result.Value(), nil
}

// runHomebrewCleaner executes the Homebrew cleaner.
func runHomebrewCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	return runCleanerWithConfig(
		ctx,
		verbose,
		dryRun,
		homebrewCleanerFactory,
		domain.HomebrewModeAll,
	)
}

// runTempFilesCleaner executes the TempFiles cleaner.
func runTempFilesCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	defaultTempPaths := []string{filepath.Join("/", "tmp")}
	defaultExcludes := []string{}

	return runGenericCleanerWithError(
		ctx,
		verbose,
		dryRun,
		func(v, d bool) (cleaner.Cleaner, error) {
			return cleaner.NewTempFilesCleaner(v, d, "7d", defaultExcludes, defaultTempPaths)
		},
	)
}

// runGenericCleaner executes a cleaner using a factory function.
func runGenericCleaner(
	ctx context.Context, verbose, dryRun bool,
	factory func(bool, bool) cleaner.Cleaner,
) (domain.CleanResult, error) {
	cleanerInstance := factory(verbose, dryRun)

	result := cleanerInstance.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	return result.Value(), nil
}

// runGenericCleanerWithError executes a cleaner using a factory function that may return an error.
// Returns the error to the caller for proper handling.
func runGenericCleanerWithError(
	ctx context.Context, verbose, dryRun bool,
	factory func(bool, bool) (cleaner.Cleaner, error),
) (domain.CleanResult, error) {
	cleanerInstance, err := factory(verbose, dryRun)
	if err != nil {
		return domain.CleanResult{}, fmt.Errorf("failed to create cleaner: %w", err)
	}

	result := cleanerInstance.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	return result.Value(), nil
}

// runCleanerWithConfig executes a cleaner that requires a single configuration parameter.
// T is the cleaner configuration type (e.g., domain.HomebrewMode, domain.DockerPruneMode).
func runCleanerWithConfig[T any](
	ctx context.Context,
	verbose, dryRun bool,
	factory func(bool, bool, T) cleaner.Cleaner,
	config T,
) (domain.CleanResult, error) {
	return runGenericCleaner(ctx, verbose, dryRun, func(v, d bool) cleaner.Cleaner {
		return factory(v, d, config)
	})
}

// homebrewCleanerFactory creates a Homebrew cleaner with the specified mode.
func homebrewCleanerFactory(v, d bool, mode domain.HomebrewMode) cleaner.Cleaner {
	return cleaner.NewHomebrewCleaner(v, d, mode)
}

// dockerCleanerFactory creates a Docker cleaner with the specified prune mode.
func dockerCleanerFactory(v, d bool, pruneMode domain.DockerPruneMode) cleaner.Cleaner {
	return cleaner.NewDockerCleaner(v, d, pruneMode)
}

// runManagerCleaner executes a cleaner with manager-based configuration.
// T is the manager type (e.g., cleaner.NodePackageManagerType, cleaner.LangVersionManagerType).
func runManagerCleaner[T any](
	ctx context.Context,
	verbose, dryRun bool,
	availableManagers []T,
	factory func(bool, bool, []T) cleaner.Cleaner,
) (domain.CleanResult, error) {
	return runGenericCleaner(ctx, verbose, dryRun, func(v, d bool) cleaner.Cleaner {
		return factory(v, d, availableManagers)
	})
}

// runCleanerWithNodeManagers executes the Node package manager cleaner.
func runCleanerWithNodeManagers(
	ctx context.Context, verbose, dryRun bool,
	managers []domain.PackageManagerType,
) (domain.CleanResult, error) {
	return runManagerCleaner(ctx, verbose, dryRun, managers, nodeManagerFactory)
}

// nodeManagerFactory is a factory function for Node package managers.
// This adapter bridges the type gap between *NodePackageManagerCleaner and cleaner.Cleaner.
func nodeManagerFactory(v, d bool, managers []domain.PackageManagerType) cleaner.Cleaner {
	return cleaner.NewNodePackageManagerCleaner(v, d, managers)
}

// runNodePackageManagerCleaner executes the Node package manager cleaner.
func runNodePackageManagerCleaner(
	ctx context.Context,
	dryRun, verbose bool,
) (domain.CleanResult, error) {
	return runCleanerWithNodeManagers(ctx, verbose, dryRun, cleaner.AvailableNodePackageManagers())
}

// runGoCleaner executes the Go cleaner.
// Skips cleaning if other Go processes are running to avoid cache corruption.
func runGoCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	if hasOtherGoProcesses() {
		return domain.CleanResult{}, errors.New(
			"other Go processes detected (e.g., running tests, builds, or IDE operations). " +
				"Skipping Go cache cleaning to avoid cache corruption. " +
				"Please wait for other Go operations to complete and try again.",
		)
	}

	return runGenericCleanerWithError(
		ctx,
		verbose,
		dryRun,
		func(v, d bool) (cleaner.Cleaner, error) {
			return cleaner.NewGoCleaner(
				v,
				d,
				cleaner.GoCacheGOCACHE|cleaner.GoCacheTestCache|cleaner.GoCacheModCache|cleaner.GoCacheBuildCache,
			)
		},
	)
}

// hasOtherGoProcesses checks if there are other Go processes running
// that might be using the Go cache.
func hasOtherGoProcesses() bool {
	// Check for common Go-related processes
	goProcesses := []string{"go", "gopls", "golangci-lint", "dlv"}

	return slices.ContainsFunc(goProcesses, isProcessRunning)
}

// isProcessRunning checks if a process with the given name is currently running.
// This is a platform-independent check that looks for other instances.
func isProcessRunning(name string) bool {
	// Use pgrep on Unix systems to check for processes
	// Exclude the current clean-wizard process itself
	cmd := exec.Command("pgrep", "-x", name)

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Parse output to get PIDs
	pids := strings.Fields(string(output))

	// Get current process PID
	currentPID := os.Getpid()

	// Check if any found PID is not the current process
	for _, pidStr := range pids {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		if pid != currentPID {
			return true
		}
	}

	return false
}

// runCargoCleaner executes the Cargo cleaner.
func runCargoCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	return runGenericCleaner(ctx, verbose, dryRun, func(v, d bool) cleaner.Cleaner {
		return cleaner.NewCargoCleaner(v, d)
	})
}

// runBuildCacheCleaner executes the Build Cache cleaner.
func runBuildCacheCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	return runGenericCleanerWithError(
		ctx,
		verbose,
		dryRun,
		func(v, d bool) (cleaner.Cleaner, error) {
			return cleaner.NewBuildCacheCleaner(v, d, "30d", []string{}, []string{})
		},
	)
}

// runDockerCleaner executes the Docker cleaner.
func runDockerCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	return runCleanerWithConfig(
		ctx,
		verbose,
		dryRun,
		dockerCleanerFactory,
		domain.DockerPruneAll,
	)
}

// runSystemCacheCleaner executes the System Cache cleaner.
func runSystemCacheCleaner(ctx context.Context, dryRun, verbose bool) (domain.CleanResult, error) {
	return runGenericCleanerWithError(
		ctx,
		verbose,
		dryRun,
		func(v, d bool) (cleaner.Cleaner, error) {
			return cleaner.NewSystemCacheCleaner(v, d, "30d", nil)
		},
	)
}

// runProjectsManagementAutomationCleaner executes Projects Management Automation cleaner.
func runProjectsManagementAutomationCleaner(
	ctx context.Context,
	dryRun, verbose bool,
) (domain.CleanResult, error) {
	return runGenericCleaner(ctx, verbose, dryRun, func(v, d bool) cleaner.Cleaner {
		return cleaner.NewProjectsManagementAutomationCleaner(v, d)
	})
}

// runCompiledBinariesCleaner executes the Compiled Binaries cleaner.
func runCompiledBinariesCleaner(
	ctx context.Context,
	dryRun, verbose bool,
) (domain.CleanResult, error) {
	return runGenericCleaner(ctx, verbose, dryRun, func(v, d bool) cleaner.Cleaner {
		return cleaner.NewCompiledBinariesCleaner(
			v,
			d,
			cleaner.DefaultMinSizeMB,
			cleaner.DefaultOlderThan,
			nil,
			[]string{},
		)
	})
}
