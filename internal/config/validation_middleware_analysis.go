package config

import (
	"reflect"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// analyzeConfigChanges analyzes differences between current and proposed configuration.
func (vm *ValidationMiddleware) analyzeConfigChanges(current, proposed *domain.Config) []ConfigChange {
	changes := []ConfigChange{}

	// Analyze basic fields
	if current.SafeMode != proposed.SafeMode {
		changes = append(changes, *vm.createFieldChange("safe_mode", current.SafeMode, proposed.SafeMode))
	}

	if current.MaxDiskUsage != proposed.MaxDiskUsage {
		changes = append(changes, *vm.createFieldChange("max_disk_usage", current.MaxDiskUsage, proposed.MaxDiskUsage))
	}

	// Analyze protected paths
	pathsChanges := vm.analyzePathChanges("protected", current.Protected, proposed.Protected)
	changes = append(changes, pathsChanges...)

	// Analyze profiles
	profilesChanges := vm.analyzeProfileChanges(current.Profiles, proposed.Profiles)
	changes = append(changes, profilesChanges...)

	return changes
}

// analyzePathChanges analyzes path array changes.
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
				Risk:      domain.RiskLow,
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
				Risk:      domain.RiskHigh, // Removing protected paths is risky
			})
		}
	}

	return changes
}

// analyzeProfileChanges analyzes profile map changes.
func (vm *ValidationMiddleware) analyzeProfileChanges(current, proposed map[string]*domain.Profile) []ConfigChange {
	changes := []ConfigChange{}

	// Check for added profiles
	for name, profile := range proposed {
		if current[name] == nil {
			// Guard against nil profile before accessing fields
			if profile == nil {
				continue
			}
			changes = append(changes, ConfigChange{
				Field:     "profiles." + name,
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
				Field:     "profiles." + name,
				OldValue:  profile.Name,
				NewValue:  nil,
				Operation: OperationRemoved,
				Risk:      domain.RiskLow, // Removing profiles is generally safe
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
					Field:     "profiles." + name,
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

func (vm *ValidationMiddleware) assessChangeRisk(field string, old, new any) domain.RiskLevel {
	switch field {
	case "safe_mode":
		if old == true && new == false {
			return domain.RiskHigh
		}
		return domain.RiskLow
	case "max_disk_usage":
		// Safe type assertions
		oldVal, oldOk := old.(int)
		newVal, newOk := new.(int)
		if !oldOk || !newOk {
			return domain.RiskHigh // Conservative risk for unexpected types
		}
		if oldVal < newVal {
			return domain.RiskMedium
		}
		return domain.RiskLow
	case "protected":
		if new == nil {
			return domain.RiskCritical
		}
		return domain.RiskLow
	default:
		return domain.RiskLow
	}
}

func (vm *ValidationMiddleware) assessProfileRisk(profile *domain.Profile) domain.RiskLevel {
	// Guard against nil profile
	if profile == nil {
		return domain.RiskHigh
	}

	maxRisk := domain.RiskLow
	for _, op := range profile.Operations {
		if op.RiskLevel == domain.RiskCritical {
			return domain.RiskCritical
		}
		if op.RiskLevel == domain.RiskHigh {
			maxRisk = domain.RiskHigh
		} else if op.RiskLevel == domain.RiskMedium && maxRisk == domain.RiskLow {
			maxRisk = domain.RiskMedium
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

// createFieldChange creates a ConfigChange for a field comparison.
func (vm *ValidationMiddleware) createFieldChange(field string, oldValue, newValue any) *ConfigChange {
	return &ConfigChange{
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operation: vm.getChangeOperation(oldValue, newValue),
		Risk:      vm.assessChangeRisk(field, oldValue, newValue),
	}
}
