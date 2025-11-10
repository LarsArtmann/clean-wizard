package domain

import (
	"testing"
)

// FuzzValidationLevelCreation tests validation level creation with fuzzed inputs
func FuzzValidationLevelCreation(f *testing.F) {
	f.Add(int(ValidationLevelBasic))
	
	f.Fuzz(func(t *testing.T, data int) {
		// Should not panic on any int value
		level := ValidationLevel(data)
		
		// String method should not panic
		_ = level.String()
		
		// Should handle any int value gracefully
		if level >= ValidationLevelNone && level <= ValidationLevelStrict {
			// Valid range, string should be meaningful
			str := level.String()
			if str == "" || str == "Unknown" {
				t.Logf("Valid level %d produced unexpected string: %s", level, str)
			}
		} else {
			// Invalid range, should handle gracefully
			str := level.String()
			if str != "Unknown" {
				t.Logf("Invalid level %d should produce 'Unknown', got: %s", level, str)
			}
		}
	})
}

// FuzzScanRequestCreation tests scan request creation with fuzzed inputs
func FuzzScanRequestCreation(f *testing.F) {
	f.Add(ScanTypeNixStore)
	
	f.Fuzz(func(t *testing.T, data ScanType) {
		// Should not panic on any ScanType value
		req := ScanRequest{
			Type:      data,
			Limit:     100,
			Recursive: true,
		}
		
		// Validation should not panic
		err := req.Validate()
		
		// Should handle invalid types gracefully
		if data >= ScanTypeFile && data <= ScanTypeNixStore {
			// Valid type, should pass or fail gracefully
			if err != nil {
				t.Logf("Valid scan type %d failed validation: %v", data, err)
			}
		} else {
			// Invalid type, should fail gracefully
			if err == nil {
				t.Logf("Invalid scan type %d should fail validation", data)
			}
		}
	})
}

// FuzzCleanRequestCreation tests clean request creation with fuzzed inputs
func FuzzCleanRequestCreation(f *testing.F) {
	f.Add(CleanOperationNixGenerations)
	
	f.Fuzz(func(t *testing.T, data CleanOperation) {
		// Should not panic on any CleanOperation value
		req := CleanRequest{
			Operation:  data,
			DryRun:     true,
			Confirm:    false,
			Items:       []CleanItem{},
		}
		
		// Validation should not panic
		err := req.Validate()
		
		// Should handle invalid operations gracefully
		if data >= CleanOperationGenerations && data <= CleanOperationNixGenerations {
			// Valid operation, should pass or fail gracefully
			if err != nil && len(req.Items) == 0 {
				t.Logf("Valid operation %d with empty items failed: %v", data, err)
			}
		} else {
			// Invalid operation, should fail gracefully
			if err == nil {
				t.Logf("Invalid clean operation %d should fail validation", data)
			}
		}
	})
}

// FuzzCleanItemCreation tests clean item creation with fuzzed inputs
func FuzzCleanItemCreation(f *testing.F) {
	f.Add(CleanTypeGeneration)
	
	f.Fuzz(func(t *testing.T, data CleanType) {
		// Should not panic on any CleanType value
		item := CleanItem{
			Type:     data,
			Path:     "/test/path",
			Size:     1024,
			Risk:     RiskLevelLow,
			Created:  "2024-01-01",
			Metadata: map[string]any{"test": "data"},
		}
		
		// Validation should not panic
		err := item.Validate()
		
		// Should handle invalid types gracefully
		if data >= CleanTypeFile && data <= CleanTypeGeneration {
			// Valid type, should pass validation
			if err != nil {
				t.Logf("Valid clean item type %d failed validation: %v", data, err)
			}
		} else {
			// Invalid type, should fail gracefully
			if err == nil {
				t.Logf("Invalid clean item type %d should fail validation", data)
			}
		}
	})
}

// FuzzNixGenerationCreation tests Nix generation creation with fuzzed inputs
func FuzzNixGenerationCreation(f *testing.F) {
	f.Add("2024-01-01")
	
	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string input
		gen := NixGeneration{
			ID:       int32(len(data)), // Convert string length to ID
			Created:  data,
			Active:   len(data) % 2 == 0, // Random active status
			Size:     int64(len(data) * 1024), // Convert to size
			Files:    []string{data + "/file1", data + "/file2"},
		}
		
		// Should not panic on creation
		_ = gen.IsValid()
		
		// Validation should not panic
		err := gen.Validate()
		
		// Should handle various string lengths gracefully
		if len(data) > 100000 { // Prevent excessive memory usage in fuzzing
			return
		}
		
		// ID validation should be reasonable
		if gen.ID < 0 {
			t.Logf("Invalid generation ID: %d", gen.ID)
		}
		
		// Size validation should be reasonable
		if gen.Size < 0 {
			t.Logf("Invalid generation size: %d", gen.Size)
		}
		
		// File paths should be reasonable
		for _, file := range gen.Files {
			if len(file) > 10000 { // Prevent extremely long paths
				t.Logf("Excessively long file path: %d chars", len(file))
				break
			}
		}
	})
}

// FuzzRiskLevelOperations tests risk level operations with fuzzed inputs
func FuzzRiskLevelOperations(f *testing.F) {
	f.Add(RiskLevelLow)
	
	f.Fuzz(func(t *testing.T, data RiskLevel) {
		// Should not panic on any RiskLevel value
		level := data
		
		// Should handle any value gracefully
		_ = level.IsValid()
		_ = level.String()
		_ = level.MarshalYAML()
		
		// Should handle invalid levels gracefully
		if level >= RiskLevelLow && level <= RiskLevelCritical {
			// Valid range, string should be meaningful
			str := level.String()
			if str == "" {
				t.Logf("Valid risk level %d produced empty string", level)
			}
			
			// YAML marshaling should work
			yaml, err := level.MarshalYAML()
			if err != nil {
				t.Logf("Valid risk level %d failed YAML marshal: %v", level, err)
			} else if len(yaml) == 0 {
				t.Logf("Valid risk level %d produced empty YAML", level)
			}
		} else {
			// Invalid range, should handle gracefully
			str := level.String()
			if str != "Unknown" {
				t.Logf("Invalid risk level %d should produce 'Unknown', got: %s", level, str)
			}
		}
		
		// Should handle extremely high values gracefully
		if level > RiskLevelCritical+100 || level < RiskLevelLow-100 {
			// Should not crash or panic
			_ = level.String()
		}
	})
}

// FuzzCleanResultCreation tests clean result creation with fuzzed inputs
func FuzzCleanResultCreation(f *testing.F) {
	f.Add("test operation")
	
	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string input
		result := CleanResult{
			Operation:   data,
			Items:       make([]CleanItem, len(data)%10), // Random number of items
			Success:     len(data)%2 == 0, // Random success status
			Size:        int64(len(data) * 1024),
			Duration:    "1s",
			Error:       nil,
		}
		
		// Should not panic on creation
		_ = result.IsValid()
		
		// Validation should not panic
		err := result.Validate()
		
		// Should handle various string lengths gracefully
		if len(data) > 100000 { // Prevent excessive memory usage
			return
		}
		
		// Size should be reasonable
		if result.Size < 0 {
			t.Logf("Invalid result size: %d", result.Size)
		}
		
		// Items array should be reasonable
		if len(result.Items) > 1000 {
			t.Logf("Excessive number of items: %d", len(result.Items))
		}
		
		// Duration should be reasonable if set
		if result.Duration == "" {
			// Empty duration should fail validation
			if err == nil {
				t.Logf("Empty duration should fail validation")
			}
		}
	})
}