package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// TestTypeSafeSchemaRules_DeepCopy verifies GetTypeSafeSchemaRules returns immutable copy
func TestTypeSafeSchemaRules_DeepCopy(t *testing.T) {
	// Create original populated instance
	original := &TypeSafeValidationRules{
		UniquePaths:           true,
		UniqueProfiles:        false,
		RequireSafeMode:       true,
		MaxRiskLevel:          domain.RiskLevelHighType,
		BackupRequired:        domain.RiskLevelCriticalType,
		ProtectedSystemPaths:  []string{"/System", "/Applications", "/Library"},
		DefaultProtectedPaths: []string{"/usr/local", "/opt"},
		MaxDiskUsage: &NumericValidationRule{
			Required: true,
			Min:      intPtr(10),
			Max:      intPtr(95),
			Message:  "invalid disk usage",
		},
		MinProtectedPaths: &NumericValidationRule{
			Required: false,
			Min:      intPtr(1),
			Max:      intPtr(50),
			Message:  "insufficient protected paths",
		},
		MaxProfiles: &NumericValidationRule{
			Required: true,
			Min:      intPtr(1),
			Max:      intPtr(10),
			Message:  "too many profiles",
		},
		ProfileNamePattern: &StringValidationRule{
			Required: true,
			Pattern:  `^[a-zA-Z0-9_-]+$`,
			Message:  "invalid profile name",
		},
		PathPattern: &StringValidationRule{
			Required: true,
			Pattern:  `^/[/a-zA-Z0-9_.-]*$`,
			Message:  "invalid path format",
		},
	}

	// Get the copy
	copied := original.GetTypeSafeSchemaRules()

	// Verify they are different instances
	if copied == original {
		t.Fatal("FAILED: GetTypeSafeSchemaRules returned same instance, not a copy")
	}

	// Test 1: Mutate boolean fields on copy
	copied.UniquePaths = !original.UniquePaths
	copied.UniqueProfiles = !original.UniqueProfiles
	copied.RequireSafeMode = !original.RequireSafeMode
	copied.MaxRiskLevel = domain.RiskLevelLowType
	copied.BackupRequired = domain.RiskLevelMediumType

	// Verify original unchanged
	if original.UniquePaths == copied.UniquePaths {
		t.Error("FAILED: modified copied.UniquePaths affected original")
	}
	if original.UniqueProfiles == copied.UniqueProfiles {
		t.Error("FAILED: modified copied.UniqueProfiles affected original")
	}
	if original.RequireSafeMode == copied.RequireSafeMode {
		t.Error("FAILED: modified copied.RequireSafeMode affected original")
	}
	if original.MaxRiskLevel == copied.MaxRiskLevel {
		t.Error("FAILED: modified copied.MaxRiskLevel affected original")
	}
	if original.BackupRequired == copied.BackupRequired {
		t.Error("FAILED: modified copied.BackupRequired affected original")
	}

	// Test 2: Mutate slices on copy
	if len(copied.ProtectedSystemPaths) > 0 {
		copied.ProtectedSystemPaths[0] = "MODIFIED"
		copied.ProtectedSystemPaths = append(copied.ProtectedSystemPaths, "NEW_PATH")
	}
	if len(copied.DefaultProtectedPaths) > 0 {
		copied.DefaultProtectedPaths[0] = "MODIFIED_TOO"
		copied.DefaultProtectedPaths = append(copied.DefaultProtectedPaths, "ANOTHER_NEW_PATH")
	}

	// Verify original slices unchanged
	if len(original.ProtectedSystemPaths) == len(copied.ProtectedSystemPaths) {
		t.Error("FAILED: slice length modification affected original")
	}
	if original.ProtectedSystemPaths[0] == copied.ProtectedSystemPaths[0] {
		t.Error("FAILED: slice element modification affected original")
	}
	if len(original.DefaultProtectedPaths) == len(copied.DefaultProtectedPaths) {
		t.Error("FAILED: default slice length modification affected original")
	}
	if original.DefaultProtectedPaths[0] == copied.DefaultProtectedPaths[0] {
		t.Error("FAILED: default slice element modification affected original")
	}

	// Test 3: Mutate numeric rules on copy
	if copied.MaxDiskUsage != nil {
		copied.MaxDiskUsage.Required = !original.MaxDiskUsage.Required
		copied.MaxDiskUsage.Min = intPtr(999)
		copied.MaxDiskUsage.Max = intPtr(888)
		copied.MaxDiskUsage.Message = "MODIFIED MESSAGE"
	}
	if copied.MinProtectedPaths != nil {
		copied.MinProtectedPaths.Required = !original.MinProtectedPaths.Required
		copied.MinProtectedPaths.Min = intPtr(777)
		copied.MinProtectedPaths.Max = intPtr(666)
		copied.MinProtectedPaths.Message = "MODIFIED MESSAGE TOO"
	}
	if copied.MaxProfiles != nil {
		copied.MaxProfiles.Required = !original.MaxProfiles.Required
		copied.MaxProfiles.Min = intPtr(555)
		copied.MaxProfiles.Max = intPtr(444)
		copied.MaxProfiles.Message = "MODIFIED MAX PROFILES MESSAGE"
	}

	// Verify original numeric rules unchanged
	if original.MaxDiskUsage != nil && copied.MaxDiskUsage != nil {
		if original.MaxDiskUsage.Required == copied.MaxDiskUsage.Required {
			t.Error("FAILED: modified copied.MaxDiskUsage.Required affected original")
		}
		if *original.MaxDiskUsage.Min == *copied.MaxDiskUsage.Min {
			t.Error("FAILED: modified copied.MaxDiskUsage.Min affected original")
		}
		if *original.MaxDiskUsage.Max == *copied.MaxDiskUsage.Max {
			t.Error("FAILED: modified copied.MaxDiskUsage.Max affected original")
		}
		if original.MaxDiskUsage.Message == copied.MaxDiskUsage.Message {
			t.Error("FAILED: modified copied.MaxDiskUsage.Message affected original")
		}
	}

	if original.MinProtectedPaths != nil && copied.MinProtectedPaths != nil {
		if original.MinProtectedPaths.Required == copied.MinProtectedPaths.Required {
			t.Error("FAILED: modified copied.MinProtectedPaths.Required affected original")
		}
		if *original.MinProtectedPaths.Min == *copied.MinProtectedPaths.Min {
			t.Error("FAILED: modified copied.MinProtectedPaths.Min affected original")
		}
		if *original.MinProtectedPaths.Max == *copied.MinProtectedPaths.Max {
			t.Error("FAILED: modified copied.MinProtectedPaths.Max affected original")
		}
		if original.MinProtectedPaths.Message == copied.MinProtectedPaths.Message {
			t.Error("FAILED: modified copied.MinProtectedPaths.Message affected original")
		}
	}

	if original.MaxProfiles != nil && copied.MaxProfiles != nil {
		if original.MaxProfiles.Required == copied.MaxProfiles.Required {
			t.Error("FAILED: modified copied.MaxProfiles.Required affected original")
		}
		if *original.MaxProfiles.Min == *copied.MaxProfiles.Min {
			t.Error("FAILED: modified copied.MaxProfiles.Min affected original")
		}
		if *original.MaxProfiles.Max == *copied.MaxProfiles.Max {
			t.Error("FAILED: modified copied.MaxProfiles.Max affected original")
		}
		if original.MaxProfiles.Message == copied.MaxProfiles.Message {
			t.Error("FAILED: modified copied.MaxProfiles.Message affected original")
		}
	}

	// Test 4: Mutate string rules on copy
	if copied.ProfileNamePattern != nil {
		copied.ProfileNamePattern.Required = !original.ProfileNamePattern.Required
		copied.ProfileNamePattern.Pattern = "MODIFIED_PATTERN"
		copied.ProfileNamePattern.Message = "MODIFIED STRING MESSAGE"
	}
	if copied.PathPattern != nil {
		copied.PathPattern.Required = !original.PathPattern.Required
		copied.PathPattern.Pattern = "ANOTHER_MODIFIED_PATTERN"
		copied.PathPattern.Message = "ANOTHER MESSAGE MODIFIED"
	}

	// Verify original string rules unchanged
	if original.ProfileNamePattern != nil && copied.ProfileNamePattern != nil {
		if original.ProfileNamePattern.Required == copied.ProfileNamePattern.Required {
			t.Error("FAILED: modified copied.ProfileNamePattern.Required affected original")
		}
		if original.ProfileNamePattern.Pattern == copied.ProfileNamePattern.Pattern {
			t.Error("FAILED: modified copied.ProfileNamePattern.Pattern affected original")
		}
		if original.ProfileNamePattern.Message == copied.ProfileNamePattern.Message {
			t.Error("FAILED: modified copied.ProfileNamePattern.Message affected original")
		}
	}

	// Test 5: Nil field handling
	originalWithNils := &TypeSafeValidationRules{
		UniquePaths:  true,
		MaxRiskLevel: domain.RiskLevelHighType,
		// All pointer fields left as nil
	}

	copiedWithNils := originalWithNils.GetTypeSafeSchemaRules()

	// Should not panic and should handle nil correctly
	if copiedWithNils.MaxDiskUsage != nil {
		t.Error("FAILED: nil MaxDiskUsage became non-nil in copy")
	}
	if copiedWithNils.ProfileNamePattern != nil {
		t.Error("FAILED: nil ProfileNamePattern became non-nil in copy")
	}

	// Boolean fields should still work
	copiedWithNils.UniquePaths = false
	if originalWithNils.UniquePaths == copiedWithNils.UniquePaths {
		t.Error("FAILED: boolean field modification affected original with nils")
	}
}

// TestTypeSafeSchemaRules_DeepCopyEmpty verifies empty instance handling
func TestTypeSafeSchemaRules_DeepCopyEmpty(t *testing.T) {
	original := &TypeSafeValidationRules{}
	copied := original.GetTypeSafeSchemaRules()

	if copied == original {
		t.Fatal("FAILED: GetTypeSafeSchemaRules returned same instance for empty struct")
	}

	// Should be able to modify copy without affecting original
	copied.UniquePaths = true
	copied.RequireSafeMode = true
	copied.MaxRiskLevel = domain.RiskLevelCriticalType

	if original.UniquePaths == copied.UniquePaths {
		t.Error("FAILED: empty original affected by copy modification")
	}
	if original.RequireSafeMode == copied.RequireSafeMode {
		t.Error("FAILED: empty original RequireSafeMode affected by copy modification")
	}
	if original.MaxRiskLevel == copied.MaxRiskLevel {
		t.Error("FAILED: empty original MaxRiskLevel affected by copy modification")
	}
}

// Helper function for pointer creation
func intPtr(i int) *int {
	return &i
}
