package commands

import (
	"testing"
)

func TestCleanerMetadataCompleteness(t *testing.T) {
	allCleanerTypes := []CleanerType{
		CleanerTypeNix,
		CleanerTypeHomebrew,
		CleanerTypeTempFiles,
		CleanerTypeNodePackages,
		CleanerTypeGoPackages,
		CleanerTypeCargoPackages,
		CleanerTypeBuildCache,
		CleanerTypeDocker,
		CleanerTypeSystemCache,
		CleanerTypeProjectsManagementAutomation,
		CleanerTypeCompiledBinaries,
		CleanerTypeProjectExecutables,
		CleanerTypeGolangciLintCache,
	}

	for _, ct := range allCleanerTypes {
		t.Run(string(ct), func(t *testing.T) {
			entry, ok := cleanerMetadata[ct]
			if !ok {
				t.Errorf("CleanerType %q has no entry in cleanerMetadata", ct)

				return
			}

			if entry.RegistryName == "" {
				t.Errorf("CleanerType %q has empty RegistryName", ct)
			}

			if entry.DisplayName == "" {
				t.Errorf("CleanerType %q has empty DisplayName", ct)
			}

			if entry.Description == "" {
				t.Errorf("CleanerType %q has empty Description", ct)
			}

			if entry.Icon == "" {
				t.Errorf("CleanerType %q has empty Icon", ct)
			}
		})
	}
}

func TestCleanerMetadataRegistryNameUniqueness(t *testing.T) {
	seen := make(map[string]CleanerType, len(cleanerMetadata))

	for ct, entry := range cleanerMetadata {
		if existing, ok := seen[entry.RegistryName]; ok {
			t.Errorf(
				"Duplicate RegistryName %q: found in both %q and %q",
				entry.RegistryName,
				existing,
				ct,
			)
		}

		seen[entry.RegistryName] = ct
	}
}

func TestRegistryNameToCleanerTypeCompleteness(t *testing.T) {
	for ct, entry := range cleanerMetadata {
		found, ok := registryNameToCleanerType[entry.RegistryName]
		if !ok {
			t.Errorf("RegistryName %q has no reverse mapping", entry.RegistryName)

			continue
		}

		if found != ct {
			t.Errorf(
				"RegistryName %q maps to %q, expected %q",
				entry.RegistryName,
				found,
				ct,
			)
		}
	}
}

func TestRegistryNameToCleanerTypeRoundTrip(t *testing.T) {
	for _, entry := range cleanerMetadata {
		ct, ok := registryNameToCleanerType[entry.RegistryName]
		if !ok {
			t.Errorf("RegistryName %q not found in registryNameToCleanerType", entry.RegistryName)

			continue
		}

		roundTripped, ok := cleanerMetadata[ct]
		if !ok {
			t.Errorf("CleanerType %q from reverse lookup not in cleanerMetadata", ct)

			continue
		}

		if roundTripped.RegistryName != entry.RegistryName {
			t.Errorf(
				"Round-trip failed for %q: got %q",
				entry.RegistryName,
				roundTripped.RegistryName,
			)
		}
	}
}
