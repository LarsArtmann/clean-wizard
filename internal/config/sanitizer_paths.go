package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeProtectedPaths sanitizes protected paths array
func (cs *ConfigSanitizer) sanitizeProtectedPaths(cfg *domain.Config, result *SanitizationResult) {
	sanitizedPaths := make([]string, 0, len(cfg.Protected))

	for i, path := range cfg.Protected {
		original := path

		// Trim whitespace
		if cs.rules.TrimWhitespace {
			path = strings.TrimSpace(path)
		}

		// Expand home directory
		if cs.rules.ExpandHomeDir {
			if strings.HasPrefix(path, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					path = filepath.Join(home, path[2:])
				}
			}
		}

		// Normalize path
		if cs.rules.NormalizePaths {
			path = filepath.Clean(path)
		}

		// Ensure absolute path
		if !filepath.IsAbs(path) {
			path = "/" + path
		}

		// Validate existence if enabled
		if cs.rules.ValidateExists {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				result.Warnings = append(result.Warnings, SanitizationWarning{
					Field:     fmt.Sprintf("protected[%d]", i),
					Original:  original,
					Sanitized: path,
					Reason:    "path does not exist but will be protected",
				})
			}
		}

		if original != path {
			result.addChange(fmt.Sprintf("protected[%d]", i), original, path, "path normalized")
		}

		sanitizedPaths = append(sanitizedPaths, path)
	}

	// Remove duplicates
	if cs.rules.RemoveDuplicates {
		sanitizedPaths = cs.removeDuplicates(sanitizedPaths)
	}

	// Sort paths
	if cs.rules.SortArrays {
		cs.sortStrings(sanitizedPaths)
	}

	cfg.Protected = sanitizedPaths
}

// Helper methods for path sanitization

func (cs *ConfigSanitizer) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (cs *ConfigSanitizer) sortStrings(slice []string) {
	// Simple bubble sort - sufficient for small arrays
	for i := 0; i < len(slice)-1; i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] > slice[j] {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}
