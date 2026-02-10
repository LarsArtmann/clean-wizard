package config

import (
	"reflect"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// analyzeConfigChanges analyzes differences between current and proposed configuration.
func (vm *ValidationMiddleware) analyzeConfigChanges(current, proposed *domain.Config) []ConfigChange {
	changes := []ConfigChange{}

	// Analyze basic fields
	changes = append(changes, vm.analyzeSimpleFieldChanges(current, proposed)...)

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
				Operation: domain.ChangeOperationType(domain.ChangeOperationAddedType),
				Risk:      domain.RiskLevelType(domain.RiskLevelLowType),
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
				Operation: domain.ChangeOperationType(domain.ChangeOperationRemovedType),
				Risk:      domain.RiskLevelType(domain.RiskLevelHighType), // Removing protected paths is risky
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
				Operation: domain.ChangeOperationType(domain.ChangeOperationAddedType),
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
				Operation: domain.ChangeOperationType(domain.ChangeOperationRemovedType),
				Risk:      domain.RiskLevelType(domain.RiskLevelLowType), // Removing profiles is generally safe
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
					Operation: domain.ChangeOperationType(domain.ChangeOperationModifiedType),
					Risk:      vm.assessProfileRisk(proposedProfile),
				})
			}
		}
	}

	return changes
}

// Helper methods for change analysis

func (vm *ValidationMiddleware) getChangeOperation(old, new any) domain.ChangeOperationType {
	if old == nil && new != nil {
		return domain.ChangeOperationType(domain.ChangeOperationAddedType)
	}
	if old != nil && new == nil {
		return domain.ChangeOperationType(domain.ChangeOperationRemovedType)
	}
	return domain.ChangeOperationType(domain.ChangeOperationModifiedType)
}

func (vm *ValidationMiddleware) assessChangeRisk(field string, old, new any) domain.RiskLevelType {
	switch field {
	case "safe_mode":
		if old == true && new == false {
			return domain.RiskLevelType(domain.RiskLevelHighType)
		}
		return domain.RiskLevelType(domain.RiskLevelLowType)
	case "max_disk_usage":
		// Safe type assertions
		oldVal, oldOk := old.(int)
		newVal, newOk := new.(int)
		if !oldOk || !newOk {
			return domain.RiskLevelType(domain.RiskLevelHighType) // Conservative risk for unexpected types
		}
		if oldVal < newVal {
			return domain.RiskLevelType(domain.RiskLevelMediumType)
		}
		return domain.RiskLevelType(domain.RiskLevelLowType)
	case "protected":
		if new == nil {
			return domain.RiskLevelType(domain.RiskLevelCriticalType)
		}
		return domain.RiskLevelType(domain.RiskLevelLowType)
	default:
		return domain.RiskLevelType(domain.RiskLevelLowType)
	}
}

func (vm *ValidationMiddleware) assessProfileRisk(profile *domain.Profile) domain.RiskLevelType {
	// Guard against nil profile
	if profile == nil {
		return domain.RiskLevelType(domain.RiskLevelHighType)
	}

	return maxRiskLevelFromOperations(profile.Operations, domain.RiskLevelType(domain.RiskLevelLowType))
}

func (vm *ValidationMiddleware) makeStringSet(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}

// analyzeSimpleFieldChanges analyzes changes for simple comparable fields.
func (vm *ValidationMiddleware) analyzeSimpleFieldChanges(current, proposed *domain.Config) []ConfigChange {
	changes := []ConfigChange{}

	fieldComparisons := []struct {
		name        string
		currentVal  any
		proposedVal any
	}{
		{"safe_mode", current.SafeMode, proposed.SafeMode},
		{"max_disk_usage", current.MaxDiskUsage, proposed.MaxDiskUsage},
	}

	for _, field := range fieldComparisons {
		if field.currentVal != field.proposedVal {
			changes = append(changes, *vm.createFieldChange(field.name, field.currentVal, field.proposedVal))
		}
	}

	return changes
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
