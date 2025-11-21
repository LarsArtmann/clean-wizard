// TYPE-SAFE-EXEMPT: Using reflect.DeepEqual for deep comparison of complex structs
package config

import (
	"fmt"
	"reflect"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// analyzeConfigChanges analyzes differences between current and proposed configuration
func (vm *ValidationMiddleware) analyzeConfigChanges(current, proposed *config.Config) []ConfigChange {
	changes := []ConfigChange{}

	// Analyze basic fields
	if current.SafetyLevel != proposed.SafetyLevel {
		changes = append(changes, ConfigChange{
			Field:     "safe_mode",
			OldValue:  current.SafetyLevel,
			NewValue:  proposed.SafetyLevel,
			Operation: vm.getChangeOperation(current.SafetyLevel, proposed.SafetyLevel),
			Risk:      vm.assessChangeRisk("safe_mode", current.SafetyLevel, proposed.SafetyLevel),
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
				Operation: OperationAdded,
				Risk:      shared.RiskLow,
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
				Operation: OperationRemoved,
				Risk:      shared.RiskHigh, // Removing protected paths is risky
			})
		}
	}

	return changes
}

// analyzeProfileChanges analyzes profile map changes
func (vm *ValidationMiddleware) analyzeProfileChanges(current, proposed map[string]*shared.Profile) []ConfigChange {
	changes := []ConfigChange{}

	// Check for added profiles
	for name, profile := range proposed {
		if current[name] == nil {
			// Guard against nil profile before accessing fields
			if profile == nil {
				continue
			}
			changes = append(changes, ConfigChange{
				Field:     fmt.Sprintf("profiles.%s", name),
				OldValue:  nil,
				NewValue:  profile.Name,
				Operation: OperationAdded,
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
				Operation: OperationRemoved,
				Risk:      shared.RiskLow, // Removing profiles is generally safe
			})
		}
	}

	// Check for modified profiles
	for name, proposedProfile := range proposed {
		if currentProfile := current[name]; currentProfile != nil {
			// Guard against nil proposed profile
			if proposedProfile == nil {
				continue
			}
			// Deep comparison instead of just checking length
			if currentProfile.Name != proposedProfile.Name ||
				currentProfile.Description != proposedProfile.Description ||
				!reflect.DeepEqual(currentProfile.Operations, proposedProfile.Operations) {
				changes = append(changes, ConfigChange{
					Field:     fmt.Sprintf("profiles.%s", name),
					OldValue:  currentProfile.Name,
					NewValue:  proposedProfile.Name,
					Operation: OperationModified,
					Risk:      vm.assessProfileRisk(proposedProfile),
				})
			}
		}
	}

	return changes
}

// Helper methods for change analysis

func (vm *ValidationMiddleware) getChangeOperation(old, new any) ChangeOperation {
	if old == nil && new != nil {
		return OperationAdded
	}
	if old != nil && new == nil {
		return OperationRemoved
	}
	return OperationModified
}

func (vm *ValidationMiddleware) assessChangeRisk(field string, old, new any) shared.RiskLevel {
	switch field {
	case "safe_mode":
		if old == true && new == false {
			return shared.RiskHigh
		}
		return shared.RiskLow
	case "max_disk_usage":
		// Safe type assertions
		oldVal, oldOk := old.(int)
		newVal, newOk := new.(int)
		if !oldOk || !newOk {
			return shared.RiskHigh // Conservative risk for unexpected types
		}
		if oldVal < newVal {
			return shared.RiskMedium
		}
		return shared.RiskLow
	case "protected":
		if new == nil {
			return shared.RiskCritical
		}
		return shared.RiskLow
	default:
		return shared.RiskLow
	}
}

func (vm *ValidationMiddleware) assessProfileRisk(profile *shared.Profile) shared.RiskLevel {
	// Guard against nil profile
	if profile == nil {
		return shared.RiskHigh
	}

	maxRisk := shared.RiskLow
	for _, op := range profile.Operations {
		if op.RiskLevel == shared.RiskCritical {
			return shared.RiskCritical
		}
		if op.RiskLevel == shared.RiskHigh {
			maxRisk = shared.RiskHigh
		} else if op.RiskLevel == shared.RiskMedium && maxRisk == shared.RiskLow {
			maxRisk = shared.RiskMedium
		}
	}
	return maxRisk
}

func (vm *ValidationMiddleware) makeStringSet(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}
