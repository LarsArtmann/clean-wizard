package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// analyzeConfigChanges analyzes differences between current and proposed configuration
func (vm *ValidationMiddleware) analyzeConfigChanges(current, proposed *domain.Config) []ConfigChange {
	changes := []ConfigChange{}

	// Analyze basic fields
	if current.SafeMode != proposed.SafeMode {
		changes = append(changes, ConfigChange{
			Field:     "safe_mode",
			OldValue:  current.SafeMode,
			NewValue:  proposed.SafeMode,
			Operation: vm.getChangeOperation(current.SafeMode, proposed.SafeMode),
			Risk:      vm.assessChangeRisk("safe_mode", current.SafeMode, proposed.SafeMode),
		})
	}

	if current.MaxDiskUsage != proposed.MaxDiskUsage {
		changes = append(changes, ConfigChange{
			Field:     "max_disk_usage",
			OldValue:  current.MaxDiskUsage,
			NewValue:  proposed.MaxDiskUsage,
			Operation: vm.getChangeOperation(current.MaxDiskUsage, proposed.MaxDiskUsage),
			Risk:      vm.assessChangeRisk("max_disk_usage", current.MaxDiskUsage, proposed.MaxDiskUsage),
		})
	}

	// Analyze protected paths
	pathsChanges := vm.analyzePathChanges("protected", current.Protected, proposed.Protected)
	changes = append(changes, pathsChanges...)

	// Analyze profiles
	profilesChanges := vm.analyzeProfileChanges(current.Profiles, proposed.Profiles)
	changes = append(changes, profilesChanges...)

	return changes
}

// analyzePathChanges analyzes path array changes
func (vm *ValidationMiddleware) analyzePathChanges(field string, current, proposed []string) []ConfigChange {
	changes := []ConfigChange{}

	currentSet := vm.makeStringSet(current)
	proposedSet := vm.makeStringSet(proposed)

	// Check for added paths
	for _, path := range proposed {
		if !currentSet[path] {
			changes = append(changes, ConfigChange{
				Field:     field,
				OldValue:  nil,
				NewValue:  path,
				Operation: "added",
				Risk:      "low",
			})
		}
	}

	// Check for removed paths
	for _, path := range current {
		if !proposedSet[path] {
			changes = append(changes, ConfigChange{
				Field:     field,
				OldValue:  path,
				NewValue:  nil,
				Operation: "removed",
				Risk:      "high", // Removing protected paths is risky
			})
		}
	}

	return changes
}

// analyzeProfileChanges analyzes profile map changes
func (vm *ValidationMiddleware) analyzeProfileChanges(current, proposed map[string]*domain.Profile) []ConfigChange {
	changes := []ConfigChange{}

	// Check for added profiles
	for name, profile := range proposed {
		if current[name] == nil {
			changes = append(changes, ConfigChange{
				Field:     fmt.Sprintf("profiles.%s", name),
				OldValue:  nil,
				NewValue:  profile.Name,
				Operation: "added",
				Risk:      vm.assessProfileRisk(profile),
			})
		}
	}

	// Check for removed profiles
	for name, profile := range current {
		if proposed[name] == nil {
			changes = append(changes, ConfigChange{
				Field:     fmt.Sprintf("profiles.%s", name),
				OldValue:  profile.Name,
				NewValue:  nil,
				Operation: "removed",
				Risk:      "low", // Removing profiles is generally safe
			})
		}
	}

	// Check for modified profiles
	for name, proposedProfile := range proposed {
		if currentProfile := current[name]; currentProfile != nil {
			if currentProfile.Name != proposedProfile.Name ||
				currentProfile.Description != proposedProfile.Description ||
				len(currentProfile.Operations) != len(proposedProfile.Operations) {
				changes = append(changes, ConfigChange{
					Field:     fmt.Sprintf("profiles.%s", name),
					OldValue:  currentProfile.Name,
					NewValue:  proposedProfile.Name,
					Operation: "modified",
					Risk:      vm.assessProfileRisk(proposedProfile),
				})
			}
		}
	}

	return changes
}

// Helper methods for change analysis

func (vm *ValidationMiddleware) getChangeOperation(old, new any) string {
	if old == nil && new != nil {
		return "added"
	}
	if old != nil && new == nil {
		return "removed"
	}
	return "modified"
}

func (vm *ValidationMiddleware) assessChangeRisk(field string, old, new any) string {
	switch field {
	case "safe_mode":
		if old == true && new == false {
			return "high"
		}
		return "low"
	case "max_disk_usage":
		if old.(int) < new.(int) {
			return "medium"
		}
		return "low"
	case "protected":
		if new == nil {
			return "critical"
		}
		return "low"
	default:
		return "low"
	}
}

func (vm *ValidationMiddleware) assessProfileRisk(profile *domain.Profile) string {
	maxRisk := domain.RiskLow
	for _, op := range profile.Operations {
		if op.RiskLevel == domain.RiskCritical {
			return "critical"
		}
		if op.RiskLevel == domain.RiskHigh {
			maxRisk = domain.RiskHigh
		} else if op.RiskLevel == domain.RiskMedium && maxRisk == domain.RiskLow {
			maxRisk = domain.RiskMedium
		}
	}
	return string(maxRisk)
}

func (vm *ValidationMiddleware) makeStringSet(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}
