package domain

import (
	"encoding/json"
	"testing"
)

func TestRiskLevelType_Helper(t *testing.T) {
	// Test String()
	if RiskLevelLowType.String() != "LOW" {
		t.Errorf("RiskLevelLowType.String() = %q, want %q", RiskLevelLowType.String(), "LOW")
	}
	
	// Test IsValid()
	if !RiskLevelMediumType.IsValid() {
		t.Error("RiskLevelMediumType.IsValid() = false, want true")
	}
	
	risk := RiskLevelType(999)
	if risk.IsValid() {
		t.Error("Invalid RiskLevelType.IsValid() = true, want false")
	}
	
	// Test Values()
	values := RiskLevelLowType.Values()
	if len(values) != 4 {
		t.Errorf("RiskLevelType.Values() = %d, want 4", len(values))
	}
	
	// Test JSON marshal/unmarshal
	original := RiskLevelHighType
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	var unmarshaled RiskLevelType
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	
	if unmarshaled != original {
		t.Errorf("Unmarshal = %v, want %v", unmarshaled, original)
	}
}

func TestValidationLevelType_Helper(t *testing.T) {
	// Test String()
	if ValidationLevelStrictType.String() != "STRICT" {
		t.Errorf("ValidationLevelStrictType.String() = %q, want %q", ValidationLevelStrictType.String(), "STRICT")
	}
	
	// Test IsValid()
	if !ValidationLevelNoneType.IsValid() {
		t.Error("ValidationLevelNoneType.IsValid() = false, want true")
	}
	
	level := ValidationLevelType(999)
	if level.IsValid() {
		t.Error("Invalid ValidationLevelType.IsValid() = true, want false")
	}
	
	// Test Values()
	values := ValidationLevelNoneType.Values()
	if len(values) != 4 {
		t.Errorf("ValidationLevelType.Values() = %d, want 4", len(values))
	}
}

func TestChangeOperationType_Helper(t *testing.T) {
	// Test String()
	if ChangeOperationAddedType.String() != "ADDED" {
		t.Errorf("ChangeOperationAddedType.String() = %q, want %q", ChangeOperationAddedType.String(), "ADDED")
	}
	
	// Test IsValid()
	if !ChangeOperationRemovedType.IsValid() {
		t.Error("ChangeOperationRemovedType.IsValid() = false, want true")
	}
	
	op := ChangeOperationType(999)
	if op.IsValid() {
		t.Error("Invalid ChangeOperationType.IsValid() = true, want false")
	}
	
	// Test Values()
	values := ChangeOperationAddedType.Values()
	if len(values) != 3 {
		t.Errorf("ChangeOperationType.Values() = %d, want 3", len(values))
	}
}

func TestCleanStrategyType_Helper(t *testing.T) {
	// Test String()
	if StrategyAggressiveType.String() != "aggressive" {
		t.Errorf("StrategyAggressiveType.String() = %q, want %q", StrategyAggressiveType.String(), "aggressive")
	}
	
	// Test IsValid()
	if !StrategyDryRunType.IsValid() {
		t.Error("StrategyDryRunType.IsValid() = false, want true")
	}
	
	strategy := CleanStrategyType(999)
	if strategy.IsValid() {
		t.Error("Invalid CleanStrategyType.IsValid() = true, want false")
	}
	
	// Test Values()
	values := StrategyConservativeType.Values()
	if len(values) != 3 {
		t.Errorf("CleanStrategyType.Values() = %d, want 3", len(values))
	}
}

func TestScanTypeType_Helper(t *testing.T) {
	// Test String()
	if ScanTypeNixStoreType.String() != "nix_store" {
		t.Errorf("ScanTypeNixStoreType.String() = %q, want %q", ScanTypeNixStoreType.String(), "nix_store")
	}
	
	// Test IsValid()
	if !ScanTypeHomebrewType.IsValid() {
		t.Error("ScanTypeHomebrewType.IsValid() = false, want true")
	}
	
	scan := ScanTypeType(999)
	if scan.IsValid() {
		t.Error("Invalid ScanTypeType.IsValid() = true, want false")
	}
	
	// Test Values()
	values := ScanTypeSystemType.Values()
	if len(values) != 4 {
		t.Errorf("ScanTypeType.Values() = %d, want 4", len(values))
	}
}

func TestJSON_Unmarshal_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected func() interface{}
	}{
		{
			name:     "CleanStrategy lowercase",
			data:     `"conservative"`,
			expected: func() interface{} { return StrategyConservativeType },
		},
		{
			name:     "ScanType mixed case",
			data:     `"homebrew"`,
			expected: func() interface{} { return ScanTypeHomebrewType },
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			switch tt.expected().(type) {
			case CleanStrategyType:
				var cs CleanStrategyType
				if err := json.Unmarshal([]byte(tt.data), &cs); err != nil {
					t.Fatalf("Unmarshal error: %v", err)
				}
				result = cs
			case ScanTypeType:
				var st ScanTypeType
				if err := json.Unmarshal([]byte(tt.data), &st); err != nil {
					t.Fatalf("Unmarshal error: %v", err)
				}
				result = st
			case RiskLevelType:
				var rl RiskLevelType
				if err := json.Unmarshal([]byte(tt.data), &rl); err != nil {
					t.Fatalf("Unmarshal error: %v", err)
				}
				result = rl
			}
			
			if result != tt.expected() {
				t.Errorf("Result = %v, want %v", result, tt.expected())
			}
		})
	}
}