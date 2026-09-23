package types

import "github.com/LarsArtmann/clean-wizard/internal/domain/operations"

// SettingsForProfile returns the operation settings configured for the named profile.
// The settings blocks of all operations in the profile are merged into a single
// view; when several operations define the same settings section the first
// definition wins. Operations without a settings block contribute nothing.
//
// Returns nil when the config, the profile, or its settings are absent —
// callers fall back to built-in defaults. Settings of disabled operations are
// still merged: a disabled operation does not run, but its settings describe
// how its cleaner should behave when enabled.
func (c *Config) SettingsForProfile(profileName string) *operations.OperationSettings {
	if c == nil {
		return nil
	}

	profile, ok := c.Profiles[profileName]
	if !ok {
		return nil
	}

	var merged *operations.OperationSettings

	for i := range profile.Operations {
		opSettings := profile.Operations[i].Settings
		if opSettings == nil {
			continue
		}

		if merged == nil {
			merged = &operations.OperationSettings{} //nolint:exhaustruct
		}

		mergeMissingSettings(merged, opSettings)
	}

	return merged
}

// mergeMissing copies every settings section from src into dst that dst does
// not define yet. Existing sections in dst are left untouched.
func mergeMissingSettings(dst, src *operations.OperationSettings) {
	setIfMissing(&dst.NixGenerations, src.NixGenerations)
	setIfMissing(&dst.TempFiles, src.TempFiles)
	setIfMissing(&dst.Homebrew, src.Homebrew)
	setIfMissing(&dst.NodePackages, src.NodePackages)
	setIfMissing(&dst.GoPackages, src.GoPackages)
	setIfMissing(&dst.CargoPackages, src.CargoPackages)
	setIfMissing(&dst.BuildCache, src.BuildCache)
	setIfMissing(&dst.Docker, src.Docker)
	setIfMissing(&dst.SystemCache, src.SystemCache)
	setIfMissing(&dst.SystemTemp, src.SystemTemp)
	setIfMissing(&dst.ProjectsManagementAutomation, src.ProjectsManagementAutomation)
	setIfMissing(&dst.ProjectExecutables, src.ProjectExecutables)
	setIfMissing(&dst.CompiledBinaries, src.CompiledBinaries)
	setIfMissing(&dst.GitHistory, src.GitHistory)
}

// setIfMissing assigns src to dst when dst is nil.
func setIfMissing[T any](dst **T, src *T) {
	if *dst == nil {
		*dst = src
	}
}
