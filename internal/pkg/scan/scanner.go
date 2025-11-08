package scan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/errors"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// Scanner implements the types.Scanner interface
type Scanner struct {
	verbose bool
}

// NewScanner creates a new scanner instance
func NewScanner(verbose bool) *Scanner {
	return &Scanner{verbose: verbose}
}

// Scan performs system scanning
func (s *Scanner) Scan(ctx context.Context) (*types.ScanResults, error) {
	results := &types.ScanResults{
		Timestamp: time.Now(),
		Results:   []types.ScanResult{},
	}

	// Scan Nix store
	if nixResult, err := s.scanNixStore(ctx); err == nil {
		results.Results = append(results.Results, *nixResult)
		results.TotalSizeGB += nixResult.SizeGB
	} else if s.verbose {
		fmt.Printf("Warning: Failed to scan Nix store: %v\n", err)
	}

	// Scan Homebrew
	if brewResult, err := s.scanHomebrew(ctx); err == nil {
		results.Results = append(results.Results, *brewResult)
		results.TotalSizeGB += brewResult.SizeGB
	} else if s.verbose {
		fmt.Printf("Warning: Failed to scan Homebrew: %v\n", err)
	}

	// Scan package caches
	if cacheResult, err := s.scanPackageCaches(ctx); err == nil {
		results.Results = append(results.Results, *cacheResult)
		results.TotalSizeGB += cacheResult.SizeGB
	} else if s.verbose {
		fmt.Printf("Warning: Failed to scan package caches: %v\n", err)
	}

	// Scan system temp files
	if tempResult, err := s.scanTempFiles(ctx); err == nil {
		results.Results = append(results.Results, *tempResult)
		results.TotalSizeGB += tempResult.SizeGB
	} else if s.verbose {
		fmt.Printf("Warning: Failed to scan temp files: %v\n", err)
	}

	// Scan Docker
	if dockerResult, err := s.scanDocker(ctx); err == nil {
		results.Results = append(results.Results, *dockerResult)
		results.TotalSizeGB += dockerResult.SizeGB
	} else if s.verbose {
		fmt.Printf("Warning: Failed to scan Docker: %v\n", err)
	}

	return results, nil
}

// Name returns the scanner name
func (s *Scanner) Name() string {
	return "system-scanner"
}

// scanNixStore scans the Nix store for cleanable items
func (s *Scanner) scanNixStore(ctx context.Context) (*types.ScanResult, error) {
	// Check if nix is available
	if _, err := exec.LookPath("nix"); err != nil {
		return nil, errors.NotFoundError("nix")
	}

	// Get nix store size
	cmd := exec.CommandContext(ctx, "nix", "store", "du", "/nix/store")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.ScanFailedError("nix store", err)
	}

	// Parse output to get size in bytes
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return nil, errors.ScanFailedError("nix store", fmt.Errorf("no output"))
	}

	// The first line contains the total size
	firstLine := strings.Fields(lines[0])
	if len(firstLine) < 1 {
		return nil, errors.ScanFailedError("nix store", fmt.Errorf("invalid output format"))
	}

	// Convert bytes to GB
	bytesStr := firstLine[0]
	var bytes int64
	_, err = fmt.Sscanf(bytesStr, "%d", &bytes)
	if err != nil {
		return nil, errors.ScanFailedError("nix store", err)
	}

	sizeGB := float64(bytes) / (1024 * 1024 * 1024)

	// Estimate cleanable size (roughly 20-30% of total store)
	cleanableGB := sizeGB * 0.25

	return &types.ScanResult{
		Name:        "nix-store",
		SizeGB:      cleanableGB,
		Description: "Nix store optimization and old generations",
		Cleanable:   true,
	}, nil
}

// scanHomebrew scans Homebrew for cleanable items
func (s *Scanner) scanHomebrew(ctx context.Context) (*types.ScanResult, error) {
	// Check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return nil, errors.NotFoundError("brew")
	}

	// Get Homebrew cache size
	cachePath := filepath.Join(os.Getenv("HOME"), "Library/Caches/Homebrew")
	sizeGB, err := s.getDirSizeGB(cachePath)
	if err != nil {
		return nil, errors.ScanFailedError("homebrew cache", err)
	}

	// Also check for old downloads
	downloadsPath := filepath.Join(os.Getenv("HOME"), "Library/Caches/Homebrew/downloads")
	downloadsSizeGB, err := s.getDirSizeGB(downloadsPath)
	if err == nil {
		sizeGB += downloadsSizeGB
	}

	return &types.ScanResult{
		Name:        "homebrew",
		SizeGB:      sizeGB,
		Description: "Homebrew caches and downloads",
		Cleanable:   true,
	}, nil
}

// scanPackageCaches scans package manager caches
func (s *Scanner) scanPackageCaches(ctx context.Context) (*types.ScanResult, error) {
	var totalSizeGB float64
	var foundCaches bool

	// Check npm cache
	npmCachePath := filepath.Join(os.Getenv("HOME"), ".npm")
	if npmSizeGB, err := s.getDirSizeGB(npmCachePath); err == nil && npmSizeGB > 0 {
		totalSizeGB += npmSizeGB
		foundCaches = true
	}

	// Check pnpm cache
	pnpmCachePath := filepath.Join(os.Getenv("HOME"), ".pnpm-store")
	if pnpmSizeGB, err := s.getDirSizeGB(pnpmCachePath); err == nil && pnpmSizeGB > 0 {
		totalSizeGB += pnpmSizeGB
		foundCaches = true
	}

	// Check Yarn cache
	yarnCachePath := filepath.Join(os.Getenv("HOME"), ".cache/yarn")
	if yarnSizeGB, err := s.getDirSizeGB(yarnCachePath); err == nil && yarnSizeGB > 0 {
		totalSizeGB += yarnSizeGB
		foundCaches = true
	}

	// Check Go cache
	goCachePath := filepath.Join(os.Getenv("HOME"), ".cache/go")
	if goSizeGB, err := s.getDirSizeGB(goCachePath); err == nil && goSizeGB > 0 {
		totalSizeGB += goSizeGB
		foundCaches = true
	}

	// Check Cargo cache
	cargoPath := filepath.Join(os.Getenv("HOME"), ".cargo")
	if cargoSizeGB, err := s.getDirSizeGB(cargoPath); err == nil && cargoSizeGB > 0 {
		totalSizeGB += cargoSizeGB
		foundCaches = true
	}

	if !foundCaches {
		return nil, errors.NotFoundError("package caches")
	}

	return &types.ScanResult{
		Name:        "package-caches",
		SizeGB:      totalSizeGB,
		Description: "Package manager caches (npm, pnpm, yarn, go, cargo)",
		Cleanable:   true,
	}, nil
}

// scanTempFiles scans system temporary files
func (s *Scanner) scanTempFiles(ctx context.Context) (*types.ScanResult, error) {
	var totalSizeGB float64
	var foundTempFiles bool

	// Check system temp directories
	tempDirs := []string{
		"/tmp",
		"/var/tmp",
		filepath.Join(os.Getenv("HOME"), ".cache"),
		filepath.Join(os.Getenv("HOME"), "Library/Caches"),
	}

	for _, dir := range tempDirs {
		if sizeGB, err := s.getDirSizeGB(dir); err == nil && sizeGB > 0 {
			totalSizeGB += sizeGB
			foundTempFiles = true
		}
	}

	// Check for specific temp file patterns
	tempPatterns := []string{
		"/tmp/nix-build-*",
		"/tmp/nix-shell-*",
		"/tmp/go-build*",
	}

	for _, pattern := range tempPatterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, match := range matches {
				if sizeGB, err := s.getDirSizeGB(match); err == nil && sizeGB > 0 {
					totalSizeGB += sizeGB
					foundTempFiles = true
				}
			}
		}
	}

	if !foundTempFiles {
		return nil, errors.NotFoundError("temp files")
	}

	return &types.ScanResult{
		Name:        "temp-files",
		SizeGB:      totalSizeGB,
		Description: "System temporary files and caches",
		Cleanable:   true,
	}, nil
}

// scanDocker scans Docker system for cleanable items
func (s *Scanner) scanDocker(ctx context.Context) (*types.ScanResult, error) {
	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, errors.NotFoundError("docker")
	}

	// Get Docker system info
	cmd := exec.CommandContext(ctx, "docker", "system", "df")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.ScanFailedError("docker system", err)
	}

	// Parse output to get size
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, errors.ScanFailedError("docker system", fmt.Errorf("invalid output format"))
	}

	// Skip header line and find the "Images" or "Containers" line
	var sizeGB float64
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			// Try to parse size field (usually the 3rd column)
			sizeStr := fields[2]
			if parsedSize, err := s.parseSizeString(sizeStr); err == nil {
				sizeGB += parsedSize
			}
		}
	}

	if sizeGB == 0 {
		return nil, errors.NotFoundError("docker data")
	}

	return &types.ScanResult{
		Name:        "docker",
		SizeGB:      sizeGB,
		Description: "Docker images, containers, and build cache",
		Cleanable:   true,
	}, nil
}

// getDirSizeGB gets the size of a directory in GB
func (s *Scanner) getDirSizeGB(path string) (float64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return float64(size) / (1024 * 1024 * 1024), nil
}

// parseSizeString parses a size string like "1.2GB" or "500MB" to GB
func (s *Scanner) parseSizeString(sizeStr string) (float64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	if strings.HasSuffix(sizeStr, "GB") {
		var size float64
		_, err := fmt.Sscanf(sizeStr, "%fGB", &size)
		return size, err
	}

	if strings.HasSuffix(sizeStr, "MB") {
		var size float64
		_, err := fmt.Sscanf(sizeStr, "%fMB", &size)
		return size / 1024, err
	}

	if strings.HasSuffix(sizeStr, "KB") {
		var size float64
		_, err := fmt.Sscanf(sizeStr, "%fKB", &size)
		return size / (1024 * 1024), err
	}

	// Assume bytes if no unit
	var size float64
	_, err := fmt.Sscanf(sizeStr, "%f", &size)
	return size / (1024 * 1024 * 1024), err
}
