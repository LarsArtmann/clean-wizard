package domain

import (
	"fmt"
)

// ValidationError represents a settings validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateSettings validates settings for the given operation type.
func (os *OperationSettings) ValidateSettings(opType OperationType) error {
	// Handle nil settings
	if os == nil {
		return nil // Optional settings
	}

	switch opType {
	case OperationTypeNixGenerations:
		if os.NixGenerations == nil {
			return nil // Optional settings
		}
		if os.NixGenerations.Generations < 0 || os.NixGenerations.Generations > 10 {
			return &ValidationError{
				Field:   "nix_generations.generations",
				Message: "generations must be between 0 and 10 (0 = keep only current)",
				Value:   os.NixGenerations.Generations,
			}
		}
	case OperationTypeTempFiles:
		if os.TempFiles == nil {
			return nil
		}
		if os.TempFiles.OlderThan == "" {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than is required",
				Value:   os.TempFiles.OlderThan,
			}
		}
		if _, err := ParseCustomDuration(os.TempFiles.OlderThan); err != nil {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than must be a valid duration (e.g., '7d', '24h')",
				Value:   os.TempFiles.OlderThan,
			}
		}
	case OperationTypeHomebrew:
		// Homebrew settings are always valid
	case OperationTypeSystemTemp:
		if os.SystemTemp == nil {
			return nil
		}
		if len(os.SystemTemp.Paths) == 0 {
			return &ValidationError{
				Field:   "system_temp.paths",
				Message: "paths are required",
				Value:   os.SystemTemp.Paths,
			}
		}
	case OperationTypeDocker:
		if os.Docker == nil {
			return nil
		}
		if !os.Docker.PruneMode.IsValid() {
			return &ValidationError{
				Field:   "docker.prune_mode",
				Message: "prune_mode must be a valid value (ALL, IMAGES, CONTAINERS, VOLUMES, or BUILDS), got: " + os.Docker.PruneMode.String(),
				Value:   os.Docker.PruneMode.String(),
			}
		}
	case OperationTypeGoPackages:
		if os.GoPackages == nil {
			return nil
		}
		if !os.GoPackages.CleanCache.IsValid() {
			return &ValidationError{
				Field:   "go_packages.clean_cache",
				Message: "clean_cache must be DISABLED or ENABLED, got: " + os.GoPackages.CleanCache.String(),
				Value:   os.GoPackages.CleanCache.String(),
			}
		}
		if !os.GoPackages.CleanTestCache.IsValid() {
			return &ValidationError{
				Field:   "go_packages.clean_test_cache",
				Message: "clean_test_cache must be DISABLED or ENABLED, got: " + os.GoPackages.CleanTestCache.String(),
				Value:   os.GoPackages.CleanTestCache.String(),
			}
		}
		if !os.GoPackages.CleanModCache.IsValid() {
			return &ValidationError{
				Field:   "go_packages.clean_mod_cache",
				Message: "clean_mod_cache must be DISABLED or ENABLED, got: " + os.GoPackages.CleanModCache.String(),
				Value:   os.GoPackages.CleanModCache.String(),
			}
		}
		if !os.GoPackages.CleanBuildCache.IsValid() {
			return &ValidationError{
				Field:   "go_packages.clean_build_cache",
				Message: "clean_build_cache must be DISABLED or ENABLED, got: " + os.GoPackages.CleanBuildCache.String(),
				Value:   os.GoPackages.CleanBuildCache.String(),
			}
		}
		if !os.GoPackages.CleanLintCache.IsValid() {
			return &ValidationError{
				Field:   "go_packages.clean_lint_cache",
				Message: "clean_lint_cache must be DISABLED or ENABLED, got: " + os.GoPackages.CleanLintCache.String(),
				Value:   os.GoPackages.CleanLintCache.String(),
			}
		}
	case OperationTypeSystemCache:
		if os.SystemCache == nil {
			return nil
		}
		for i, cacheType := range os.SystemCache.CacheTypes {
			if !cacheType.IsValid() {
				return &ValidationError{
					Field:   fmt.Sprintf("system_cache.cache_types[%d]", i),
					Message: "cache_type must be a valid value (SPOTLIGHT, XCODE, COCOAPODS, HOMEBREW, PIP, NPM, YARN, or CCACHE), got: " + cacheType.String(),
					Value:   cacheType.String(),
				}
			}
		}
	case OperationTypeBuildCache:
		if os.BuildCache == nil {
			return nil
		}
		for i, toolType := range os.BuildCache.ToolTypes {
			if !toolType.IsValid() {
				return &ValidationError{
					Field:   fmt.Sprintf("build_cache.tool_types[%d]", i),
					Message: "tool_type must be a valid value (GO, RUST, NODE, PYTHON, JAVA, or SCALA), got: " + toolType.String(),
					Value:   toolType.String(),
				}
			}
		}
	case OperationTypeNodePackages, OperationTypeCargoPackages, OperationTypeLangVersionManager, OperationTypeProjectsManagementAutomation:
		// No enum validation needed for these operation types
	}
	return nil
}