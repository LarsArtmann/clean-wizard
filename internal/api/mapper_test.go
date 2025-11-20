package api

import (
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// === MAPPING LAYER TESTS ===
// Validates conversion between public API types and internal domain models

func TestMapConfigToDomain_ValidConfig(t *testing.T) {
	// Create public config
	publicConfig := &PublicConfig{
		Version:        "1.0.0",
		SafeMode:       true,
		MaxDiskUsage:   80,
		ProtectedPaths: []string{"/System", "/Library"},
		Profiles: map[string]*PublicProfile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Enabled:     true,
				Operations: []PublicOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   PublicRiskLow,
						Enabled:     true,
						Settings: OperationSettings{
							DryRun:              false,
							Verbose:             true,
							TimeoutSeconds:      300,
							ConfirmBeforeDelete: false,
						},
					},
				},
			},
		},
	}

	// Map to domain
	domainConfigResult := MapConfigToDomain(publicConfig)

	// Validate successful mapping
	if domainConfigResult.IsErr() {
		err, _ := domainConfigResult.SafeError()
		t.Fatalf("Expected successful mapping, got error: %v", err)
	}

	domainConfig, err := domainConfigResult.Unwrap()
	if err != nil {
		t.Fatalf("Expected successful unwrap, got error: %v", err)
	}

	// Validate mapped values
	if domainConfig.Version != publicConfig.Version {
		t.Errorf("Expected version %s, got %s", publicConfig.Version, domainConfig.Version)
	}

	// Compare domain SafetyLevel enum with converted boolean value
	expectedSafetyLevel := domain.SafetyLevelEnabled
	if publicConfig.SafeMode {
		expectedSafetyLevel = domain.SafetyLevelEnabled
	} else {
		expectedSafetyLevel = domain.SafetyLevelDisabled
	}

	if domainConfig.SafetyLevel != expectedSafetyLevel {
		t.Errorf("Expected safetyLevel %v, got %v", expectedSafetyLevel.String(), domainConfig.SafetyLevel.String())
	}

	if domainConfig.MaxDiskUsage != int(publicConfig.MaxDiskUsage) {
		t.Errorf("Expected maxDiskUsage %d, got %d", publicConfig.MaxDiskUsage, domainConfig.MaxDiskUsage)
	}

	// Validate mapped profile
	profile, exists := domainConfig.Profiles["daily"]
	if !exists {
		t.Fatalf("Expected profile 'daily' to exist")
	}

	if profile.Name != "daily" {
		t.Errorf("Expected profile name 'daily', got %s", profile.Name)
	}

	// Validate mapped operation
	if len(profile.Operations) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(profile.Operations))
	}

	operation := profile.Operations[0]
	if operation.Name != "nix-generations" {
		t.Errorf("Expected operation name 'nix-generations', got %s", operation.Name)
	}

	// Assert risk-related fields are correctly mapped
	if operation.RiskLevel != domain.RiskLow {
		t.Errorf("Expected risk level %v, got %v", domain.RiskLow, operation.RiskLevel)
	}

	// Assert settings are not nil and round-trip correctly
	if operation.Settings == nil {
		t.Fatal("Expected operation settings to be non-nil")
	}

	// Verify key fields from domain.DefaultSettings round-trip for nix-generations
	defaultSettings := domain.DefaultSettings(domain.OperationTypeNixGenerations)
	if defaultSettings == nil || defaultSettings.NixGenerations == nil {
		t.Fatal("Expected default nix-generations settings to be available")
	}

	if operation.Settings.NixGenerations == nil {
		t.Error("Expected nix-generations settings to be present")
	} else {
		// Verify default settings are applied correctly
		if operation.Settings.NixGenerations.Generations != defaultSettings.NixGenerations.Generations {
			t.Errorf("Expected generations %d, got %d", defaultSettings.NixGenerations.Generations, operation.Settings.NixGenerations.Generations)
		}
		if operation.Settings.NixGenerations.Optimization != defaultSettings.NixGenerations.Optimization {
			t.Errorf("Expected optimization %v, got %v", defaultSettings.NixGenerations.Optimization, operation.Settings.NixGenerations.Optimization)
		}
	}
}

func TestMapConfigToPublic_ValidDomainConfig(t *testing.T) {
	// Create domain config
	domainConfig := &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  domain.SafetyLevelEnabled,
		MaxDiskUsage: 0,                 // Use default value since not testing this field
		Protected:    []string{"/test"}, // Required for validation
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Status:      domain.StatusEnabled,
				Operations: []domain.CleanupOperation{
					{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   domain.RiskMedium,
						Status:      domain.StatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
				},
			},
		},
		Updated: time.Now(), // Required for validation
	}

	// Map to public
	publicConfigResult := MapConfigToPublic(domainConfig)

	// Validate successful mapping
	if publicConfigResult.IsErr() {
		err, _ := publicConfigResult.SafeError()
		t.Fatalf("Expected successful mapping, got error: %v", err)
	}

	publicConfig, err := publicConfigResult.Unwrap()
	if err != nil {
		t.Fatalf("Expected successful unwrap, got error: %v", err)
	}

	// Validate mapped values
	if publicConfig.Version != domainConfig.Version {
		t.Errorf("Expected version %s, got %s", domainConfig.Version, publicConfig.Version)
	}

	if publicConfig.MaxDiskUsage != int32(domainConfig.MaxDiskUsage) {
		t.Errorf("Expected maxDiskUsage %d, got %d", domainConfig.MaxDiskUsage, publicConfig.MaxDiskUsage)
	}

	// Assert risk-related fields are correctly mapped in the domain -> public direction
	profile, exists := publicConfig.Profiles["test"]
	if !exists {
		t.Fatalf("Expected profile 'test' to exist")
	}

	if len(profile.Operations) != 1 {
		t.Fatalf("Expected 1 operation in profile 'test', got %d", len(profile.Operations))
	}

	operation := profile.Operations[0]
	if operation.RiskLevel != PublicRiskMedium {
		t.Errorf("Expected operation risk level %v, got %v", PublicRiskMedium, operation.RiskLevel)
	}
}

func TestMapCleanResultToPublic_ValidResult(t *testing.T) {
	// Create domain clean result
	now := time.Now()
	domainResult := domain.CleanResult{
		FreedBytes:   1024 * 1024 * 100, // 100MB
		ItemsRemoved: 50,
		ItemsFailed:  2,
		CleanTime:    5 * time.Second,
		CleanedAt:    now,
		Strategy:     domain.StrategyAggressive,
	}

	// Map to public
	publicResultResult := MapCleanResultToPublic(domainResult)

	// Validate successful mapping
	if publicResultResult.IsErr() {
		err, _ := publicResultResult.SafeError()
		t.Fatalf("Expected successful mapping, got error: %v", err)
	}

	publicResult, err := publicResultResult.Unwrap()
	if err != nil {
		t.Fatalf("Expected successful unwrap, got error: %v", err)
	}

	// Validate mapped values
	if !publicResult.Success {
		t.Errorf("Expected success to be true, got %v", publicResult.Success)
	}

	if publicResult.FreedBytes != domainResult.FreedBytes {
		t.Errorf("Expected freedBytes %d, got %d", domainResult.FreedBytes, publicResult.FreedBytes)
	}

	if publicResult.ItemsRemoved != uint32(domainResult.ItemsRemoved) {
		t.Errorf("Expected itemsRemoved %d, got %d", domainResult.ItemsRemoved, publicResult.ItemsRemoved)
	}

	if publicResult.Strategy != PublicStrategyAggressive {
		t.Errorf("Expected strategy %s, got %s", PublicStrategyAggressive, publicResult.Strategy)
	}

	if publicResult.CleanedAt != now.Format(time.RFC3339) {
		t.Errorf("Expected cleanedAt %s, got %s", now.Format(time.RFC3339), publicResult.CleanedAt)
	}
}

func TestMapCleanRequestToDomain_ValidRequest(t *testing.T) {
	// Create public clean request
	publicRequest := &CleanRequest{
		Config: PublicConfig{
			Version:        "1.0.0",
			SafeMode:       true,
			MaxDiskUsage:   80,
			ProtectedPaths: []string{"/System"},
			Profiles: map[string]*PublicProfile{
				"test": {
					Name:        "test",
					Description: "Test profile",
					Enabled:     true,
					Operations: []PublicOperation{
						{
							Name:        "temp-files",
							Description: "Clean temp files",
							RiskLevel:   PublicRiskLow,
							Enabled:     true,
							Settings:    OperationSettings{},
						},
					},
				},
			},
		},
		Strategy:   PublicStrategyConservative,
		Operations: []OperationType{OperationTypeTempFiles, OperationTypeCacheFiles},
	}

	// Map to domain
	domainRequestResult := MapCleanRequestToDomain(publicRequest)

	// Validate successful mapping
	if domainRequestResult.IsErr() {
		err, _ := domainRequestResult.SafeError()
		t.Fatalf("Expected successful mapping, got error: %v", err)
	}

	domainRequest, err := domainRequestResult.Unwrap()
	if err != nil {
		t.Fatalf("Expected successful unwrap, got error: %v", err)
	}

	// Validate mapped strategy
	if domainRequest.Strategy != domain.StrategyConservative {
		t.Errorf("Expected strategy Conservative, got %v", domainRequest.Strategy)
	}

	// Validate operations mapped to scan items
	if len(domainRequest.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(domainRequest.Items))
	}

	// Check first operation
	if domainRequest.Items[0].ScanType != domain.ScanTypeTempType {
		t.Errorf("Expected ScanTypeTemp, got %v", domainRequest.Items[0].ScanType)
	}

	if domainRequest.Items[0].Path != "temp-files" {
		t.Errorf("Expected temp-files path, got %s", domainRequest.Items[0].Path)
	}

	// Check second operation
	if domainRequest.Items[1].ScanType != domain.ScanTypeTempType {
		t.Errorf("Expected ScanTypeTemp for cache-files, got %v", domainRequest.Items[1].ScanType)
	}

	if domainRequest.Items[1].Path != "cache-files" {
		t.Errorf("Expected cache-files path, got %s", domainRequest.Items[1].Path)
	}
}

func TestMapStrategy_Conversions(t *testing.T) {
	// Test all strategy conversions
	testCases := []struct {
		public PublicStrategy
		domain domain.CleanStrategyType
	}{
		{PublicStrategyAggressive, domain.StrategyAggressive},
		{PublicStrategyConservative, domain.StrategyConservative},
		{PublicStrategyDryRun, domain.StrategyDryRun},
	}

	for _, tc := range testCases {
		t.Run(string(tc.public), func(t *testing.T) {
			// Test public to domain
			domainStrategy, err := MapStrategyToDomain(tc.public)
			if err != nil {
				t.Errorf("Expected successful mapping, got error: %v", err)
			}

			if domainStrategy != tc.domain {
				t.Errorf("Expected domain strategy %v, got %v", tc.domain, domainStrategy)
			}

			// Test domain to public
			publicStrategy := MapStrategyToPublic(tc.domain)
			if publicStrategy != tc.public {
				t.Errorf("Expected public strategy %s, got %s", tc.public, publicStrategy)
			}
		})
	}
}

func TestMapRiskLevel_Conversions(t *testing.T) {
	// Test all risk level conversions
	testCases := []struct {
		public PublicRiskLevel
		domain domain.RiskLevelType
	}{
		{PublicRiskLow, domain.RiskLow},
		{PublicRiskMedium, domain.RiskMedium},
		{PublicRiskHigh, domain.RiskHigh},
		{PublicRiskCritical, domain.RiskCritical},
	}

	for _, tc := range testCases {
		t.Run(string(tc.public), func(t *testing.T) {
			// Test public to domain
			domainLevel, err := MapRiskLevelToDomain(tc.public)
			if err != nil {
				t.Errorf("Expected successful mapping, got error: %v", err)
			}

			if domainLevel != tc.domain {
				t.Errorf("Expected domain level %v, got %v", tc.domain, domainLevel)
			}

			// Test domain to public
			publicLevel := MapRiskLevelToPublic(tc.domain)
			if publicLevel != tc.public {
				t.Errorf("Expected public level %s, got %s", tc.public, publicLevel)
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// === NEGATIVE CASE TESTS FOR MAPCONFIGTODOMAIN ===

func TestMapConfigToDomain_NilConfig(t *testing.T) {
	// Test with nil config
	domainConfigResult := MapConfigToDomain(nil)

	// Validate error result
	if !domainConfigResult.IsErr() {
		t.Fatalf("Expected error for nil config, got success")
	}

	err, hasError := domainConfigResult.SafeError()
	if !hasError {
		t.Fatalf("Expected SafeError to return error")
	}

	expectedErrorMsg := "public config cannot be nil"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	t.Logf("✅ Nil config test passed: %v", err.Error())
}

func TestMapConfigToDomain_InvalidRiskLevel(t *testing.T) {
	// Create public config with invalid risk level
	publicConfig := &PublicConfig{
		Version:        "1.0.0",
		SafeMode:       true,
		MaxDiskUsage:   80,
		ProtectedPaths: []string{"/System"},
		Profiles: map[string]*PublicProfile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Enabled:     true,
				Operations: []PublicOperation{
					{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   PublicRiskLevel("INVALID_RISK"), // Invalid risk level
						Enabled:     true,
						Settings:    OperationSettings{},
					},
				},
			},
		},
	}

	// Map to domain
	domainConfigResult := MapConfigToDomain(publicConfig)

	// Validate error result
	if !domainConfigResult.IsErr() {
		t.Fatalf("Expected error for invalid risk level, got success")
	}

	err, hasError := domainConfigResult.SafeError()
	if !hasError {
		t.Fatalf("Expected SafeError to return error")
	}

	// Should contain error about invalid risk level
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Fatalf("Expected non-empty error message")
	}

	// Check if error has mapping failure prefix (mapping phase failure)
	if !strings.HasPrefix(errorMsg, "failed to map profile") {
		t.Errorf("Expected error to start with 'failed to map profile', got: %s", errorMsg)
	}

	t.Logf("✅ Invalid risk level test passed: %v", errorMsg)
}

func TestMapConfigToDomain_ProfileValidationFailure(t *testing.T) {
	// Create public config with invalid profile that will cause validation failure
	publicConfig := &PublicConfig{
		Version:        "1.0.0",
		SafeMode:       true,
		MaxDiskUsage:   80,
		ProtectedPaths: []string{"/System"},
		Profiles: map[string]*PublicProfile{
			"test": {
				Name:        "", // Empty name should cause validation failure
				Description: "Test profile",
				Enabled:     true,
				Operations:  []PublicOperation{}, // Empty operations might be valid
			},
		},
	}

	// Map to domain
	domainConfigResult := MapConfigToDomain(publicConfig)

	// Validate error result
	if !domainConfigResult.IsErr() {
		t.Fatalf("Expected error for profile validation failure, got success")
	}

	err, hasError := domainConfigResult.SafeError()
	if !hasError {
		t.Fatalf("Expected SafeError to return error")
	}

	errorMsg := err.Error()
	if errorMsg == "" {
		t.Fatalf("Expected non-empty error message")
	}

	// Check if error has domain config validation prefix
	if !strings.HasPrefix(errorMsg, "domain config validation failed: ") {
		t.Errorf("Expected error to start with 'domain config validation failed: ', got: %s", errorMsg)
	}

	t.Logf("✅ Profile validation failure test passed: %v", errorMsg)
}

func TestMapConfigToDomain_DomainValidationFailure(t *testing.T) {
	// Create public config that will pass mapping but fail domain validation
	// (e.g., with invalid configuration values)
	publicConfig := &PublicConfig{
		Version:        "1.0.0",
		SafeMode:       true,
		MaxDiskUsage:   -10,        // Invalid disk usage (negative)
		ProtectedPaths: []string{}, // Empty protected paths might be invalid
		Profiles: map[string]*PublicProfile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Enabled:     true,
				Operations: []PublicOperation{
					{
						Name:        "nix-generations", // Valid operation type
						Description: "Test operation",
						RiskLevel:   PublicRiskLow,
						Enabled:     true,
						Settings:    OperationSettings{},
					},
				},
			},
		},
	}

	// Map to domain
	domainConfigResult := MapConfigToDomain(publicConfig)

	// Validate error result
	if !domainConfigResult.IsErr() {
		t.Fatalf("Expected error for domain validation failure, got success")
	}

	err, hasError := domainConfigResult.SafeError()
	if !hasError {
		t.Fatalf("Expected SafeError to return error")
	}

	errorMsg := err.Error()
	if errorMsg == "" {
		t.Fatalf("Expected non-empty error message")
	}

	// Should mention validation failure
	if !strings.Contains(errorMsg, "validation") {
		t.Errorf("Expected error to mention 'validation', got: %s", errorMsg)
	}

	t.Logf("✅ Domain validation failure test passed: %v", errorMsg)
}

// === OPERATION TYPE MAPPING TESTS ===

func TestMapOperationToDomain_OperationTypeMapping(t *testing.T) {
	testCases := []struct {
		name          string
		publicOp      PublicOperation
		expectedType  domain.OperationType
		shouldError   bool
		errorContains string
	}{
		{
			name: "Valid Nix Generations",
			publicOp: PublicOperation{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   PublicRiskLow,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: domain.OperationTypeNixGenerations,
			shouldError:  false,
		},
		{
			name: "Valid Temp Files",
			publicOp: PublicOperation{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   PublicRiskMedium,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: domain.OperationTypeTempFiles,
			shouldError:  false,
		},
		{
			name: "Valid Homebrew",
			publicOp: PublicOperation{
				Name:        "homebrew-cleanup",
				Description: "Clean Homebrew",
				RiskLevel:   PublicRiskHigh,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: domain.OperationTypeHomebrew,
			shouldError:  false,
		},
		{
			name: "Valid System Temp",
			publicOp: PublicOperation{
				Name:        "system-temp",
				Description: "Clean system temp",
				RiskLevel:   PublicRiskCritical,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: "system-temp",
			shouldError:  false,
		},
		{
			name: "Unknown Operation Type",
			publicOp: PublicOperation{
				Name:        "unknown-operation",
				Description: "Unknown operation",
				RiskLevel:   PublicRiskLow,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: domain.OperationType("unknown-operation"),
			shouldError:  false, // Should not error with custom operation support
		},
		{
			name: "Custom Operation with hyphens",
			publicOp: PublicOperation{
				Name:        "my-custom-op",
				Description: "My custom operation",
				RiskLevel:   PublicRiskMedium,
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			expectedType: domain.OperationType("my-custom-op"),
			shouldError:  false, // Should not error with custom operation support
		},
		{
			name: "Invalid Risk Level",
			publicOp: PublicOperation{
				Name:        "nix-generations",
				Description: "Clean Nix generations",
				RiskLevel:   PublicRiskLevel("INVALID"),
				Enabled:     true,
				Settings:    OperationSettings{},
			},
			shouldError:   true,
			errorContains: "invalid risk level",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := MapOperationToDomain(&tc.publicOp)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
					return
				}
				if result == nil {
					t.Errorf("Expected non-nil result, got nil")
					return
				}

				// Verify the operation name is preserved
				if result.Name != tc.publicOp.Name {
					t.Errorf("Expected name '%s', got '%s'", tc.publicOp.Name, result.Name)
				}

				// Verify mapped operation type matches expected type
				actualOpType := domain.GetOperationType(result.Name)
				if actualOpType != tc.expectedType {
					t.Errorf("Expected operation type '%s', got '%s'", tc.expectedType, actualOpType)
				}

				// Verify settings are not nil (should have default settings)
				if result.Settings == nil {
					t.Errorf("Expected non-nil settings for operation type '%s'", tc.expectedType)
				} else {
					// Verify concrete type/kind of result.Settings corresponds to expected default settings
					expectedDefaultSettings := domain.DefaultSettings(actualOpType)
					if expectedDefaultSettings == nil {
						t.Errorf("Expected default settings for operation type '%s', got nil", actualOpType)
					} else {
						// Verify specific settings field matches expected default
						switch actualOpType {
						case domain.OperationTypeNixGenerations:
							if result.Settings.NixGenerations == nil {
								t.Errorf("Expected NixGenerations settings to be present for nix-generations")
							} else {
								expectedNix := expectedDefaultSettings.NixGenerations
								if result.Settings.NixGenerations.Generations != expectedNix.Generations {
									t.Errorf("Expected NixGenerations.Generations %d, got %d", expectedNix.Generations, result.Settings.NixGenerations.Generations)
								}
								if result.Settings.NixGenerations.Optimization != expectedNix.Optimization {
									t.Errorf("Expected NixGenerations.Optimization %v, got %v", expectedNix.Optimization, result.Settings.NixGenerations.Optimization)
								}
							}
						case domain.OperationTypeTempFiles:
							if result.Settings.TempFiles == nil {
								t.Errorf("Expected TempFiles settings to be present for temp-files")
							} else {
								expectedTemp := expectedDefaultSettings.TempFiles
								if result.Settings.TempFiles.OlderThan != expectedTemp.OlderThan {
									t.Errorf("Expected TempFiles.OlderThan %s, got %s", expectedTemp.OlderThan, result.Settings.TempFiles.OlderThan)
								}
							}
						case domain.OperationTypeHomebrew:
							if result.Settings.Homebrew == nil {
								t.Errorf("Expected Homebrew settings to be present for homebrew-cleanup")
							}
						case domain.OperationTypeSystemTemp:
							if result.Settings.SystemTemp == nil {
								t.Errorf("Expected SystemTemp settings to be present for system-temp")
							}
						}
					}
				}

				// Verify risk level mapping
				expectedRiskLevel, mapErr := MapRiskLevelToDomain(tc.publicOp.RiskLevel)
				if mapErr != nil {
					t.Errorf("Failed to map expected risk level: %v", mapErr)
				} else if result.RiskLevel != expectedRiskLevel {
					t.Errorf("Expected risk level '%v', got '%v'", expectedRiskLevel, result.RiskLevel)
				}
			}
		})
	}
}

func TestMapOperationToDomain_NilInput(t *testing.T) {
	result, err := MapOperationToDomain(nil)

	if err == nil {
		t.Errorf("Expected error for nil operation, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result for nil input, got %v", result)
	}

	expectedError := "public operation cannot be nil"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedError, err.Error())
	}
}
