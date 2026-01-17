package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// NodePackageManagerType represents different Node.js package managers.
type NodePackageManagerType string

const (
	NodePackageManagerNPM    NodePackageManagerType = "npm"
	NodePackageManagerPNPM   NodePackageManagerType = "pnpm"
	NodePackageManagerYarn   NodePackageManagerType = "yarn"
	NodePackageManagerBun    NodePackageManagerType = "bun"
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
	verbose          bool
	dryRun           bool
	packageManagers   []NodePackageManagerType
}

// NewNodePackageManagerCleaner creates Node.js package manager cleaner.
func NewNodePackageManagerCleaner(verbose, dryRun bool, packageManagers []NodePackageManagerType) *NodePackageManagerCleaner {
	return &NodePackageManagerCleaner{
		verbose:          verbose,
		dryRun:           dryRun,
		packageManagers:   packageManagers,
	}
}

// Type returns operation type for Node package manager cleaner.
func (npmc *NodePackageManagerCleaner) Type() domain.OperationType {
	return domain.OperationTypeNodePackages
}

// IsAvailable checks if any Node.js package manager is available.
func (npmc *NodePackageManagerCleaner) IsAvailable(ctx context.Context) bool {
	for _, pm := range npmc.packageManagers {
		if npmc.isPackageManagerAvailable(pm) {
			return true
		}
	}
	return false
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

	// Validate package managers are valid
	validPackageManagers := map[NodePackageManagerType]bool{
		NodePackageManagerNPM:   true,
		NodePackageManagerPNPM:  true,
		NodePackageManagerYarn:  true,
		NodePackageManagerBun:   true,
	}

	for _, pm := range settings.NodePackages.PackageManagers {
		pmStr := NodePackageManagerType(pm)
		if !validPackageManagers[pmStr] {
			return fmt.Errorf("invalid package manager: %s (must be npm, pnpm, yarn, or bun)", pm)
		}
	}

	return nil
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
		// Yarn cache location is typically ~/.yarn/cache
		homeDir, err := getHomeDir()
		if err != nil {
			return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
		}

		cachePath := fmt.Sprintf("%s/.yarn/cache", homeDir)
		items = append(items, domain.ScanItem{
			Path:     cachePath,
			Size:     0, // Size unknown without checking
			Created:  time.Time{},
			ScanType: domain.ScanTypeTemp,
		})

		if npmc.verbose {
			fmt.Printf("Found yarn cache: %s\n", cachePath)
		}

	case NodePackageManagerBun:
		// Bun cache location is typically ~/.bun/install/cache
		homeDir, err := getHomeDir()
		if err != nil {
			return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
		}

		cachePath := fmt.Sprintf("%s/.bun/install/cache", homeDir)
		items = append(items, domain.ScanItem{
			Path:     cachePath,
			Size:     0, // Size unknown without checking
			Created:  time.Time{},
			ScanType: domain.ScanTypeTemp,
		})

		if npmc.verbose {
			fmt.Printf("Found bun cache: %s\n", cachePath)
		}
	}

	return result.Ok(items)
}

// Clean removes Node.js package manager caches.
func (npmc *NodePackageManagerCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !npmc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("no Node.js package managers available"))
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

// cleanPackageManager cleans cache for a specific package manager.
func (npmc *NodePackageManagerCleaner) cleanPackageManager(ctx context.Context, pm NodePackageManagerType) result.Result[domain.CleanResult] {
	switch pm {
	case NodePackageManagerNPM:
		cmd := exec.CommandContext(ctx, "npm", "cache", "clean", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("npm cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ npm cache cleaned")
		}

	case NodePackageManagerPNPM:
		cmd := exec.CommandContext(ctx, "pnpm", "store", "prune")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("pnpm store prune failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ pnpm store pruned")
		}

	case NodePackageManagerYarn:
		cmd := exec.CommandContext(ctx, "yarn", "cache", "clean")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("yarn cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ yarn cache cleaned")
		}

	case NodePackageManagerBun:
		cmd := exec.CommandContext(ctx, "bun", "pm", "cache", "rm")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("bun cache clean failed: %w (output: %s)", err, string(output)))
		}

		if npmc.verbose {
			fmt.Println("  ✓ bun cache cleaned")
		}
	}

	// Return success without specific byte count (unknown for most PMs)
	cleanResult := domain.CleanResult{
		FreedBytes:   0,
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanResult)
}

// getHomeDir returns the user's home directory.
func getHomeDir() (string, error) {
	// Try getting from HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// Fallback to user profile directory
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile, nil
	}

	return "", fmt.Errorf("unable to determine home directory")
}
