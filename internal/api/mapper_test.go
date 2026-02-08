package api

import (
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
						Name:        "nix-cleanup",
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
		t.Fatalf("Expected successful mapping, got error: %v", domainConfigResult.Error())
	}

	domainConfig, err := domainConfigResult.Unwrap()
	if err != nil {
		t.Fatalf("Expected successful unwrap, got error: %v", err)
	}

	// Validate mapped values
	if domainConfig.Version != publicConfig.Version {
		t.Errorf("Expected version %s, got %s", publicConfig.Version, domainConfig.Version)
	}

	if domainConfig.SafeMode.IsEnabled() != publicConfig.SafeMode {
		t.Errorf("Expected safeMode %v, got %v", publicConfig.SafeMode, domainConfig.SafeMode)
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
	if operation.Name != "nix-cleanup" {
		t.Errorf("Expected operation name 'nix-cleanup', got %s", operation.Name)
	}

	if operation.RiskLevel != domain.RiskLow {
		t.Errorf("Expected risk level LOW, got %v", operation.RiskLevel)
	}
}

func TestMapConfigToPublic_ValidDomainConfig(t *testing.T) {
	// Create domain config
	domainConfig := &domain.Config{
		Version:  "1.0.0",
		SafeMode: domain.SafeModeEnabled,
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Enabled:     domain.ProfileStatusEnabled,
				Operations: []domain.CleanupOperation{
					{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   domain.RiskMedium,
						Enabled:     domain.ProfileStatusEnabled,
						Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
				},
			},
		},
	}

	// Map to public
	publicConfigResult := MapConfigToPublic(domainConfig)

	// Validate successful mapping
	if publicConfigResult.IsErr() {
		t.Fatalf("Expected successful mapping, got error: %v", publicConfigResult.Error())
	}

	publicConfig, err := publicConfigResult.Unwrap()
	if err != nil {
		t.Fatalf("Failed to unwrap result: %v", err)
	}

	// Validate mapped values
	if publicConfig.Version != domainConfig.Version {
		t.Errorf("Expected version %s, got %s", domainConfig.Version, publicConfig.Version)
	}

	if publicConfig.MaxDiskUsage != int32(domainConfig.MaxDiskUsage) {
		t.Errorf("Expected maxDiskUsage %d, got %d", domainConfig.MaxDiskUsage, publicConfig.MaxDiskUsage)
	}

	// Validate mapped profile
	profile, exists := publicConfig.Profiles["test"]
	if !exists {
		t.Fatalf("Expected profile 'test' to exist")
	}

	if profile.Name != "test" {
		t.Errorf("Expected profile name 'test', got %s", profile.Name)
	}
}

func TestMapCleanResultToPublic_ValidResult(t *testing.T) {
	// Create domain clean result
	now := time.Now()
	domainResult := domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{
			Known: 1024 * 1024 * 100, // 100MB - must be set when ItemsRemoved > 0
		},
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
		t.Fatalf("Expected successful mapping, got error: %v", publicResultResult.Error())
	}

	publicResult, err := publicResultResult.Unwrap()
	if err != nil {
		t.Fatalf("Failed to unwrap result: %v", err)
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
