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
		Version:     "1.0.0",
		SafeMode:    true,
		MaxDiskUsage: 80,
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
							DryRun:             false,
							Verbose:            true,
							TimeoutSeconds:     300,
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

	if domainConfig.SafeMode != publicConfig.SafeMode {
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
		SafeMode: true,
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Enabled:     true,
				Operations: []domain.CleanupOperation{
					{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   domain.RiskMedium,
						Enabled:     true,
						Settings: domain.DefaultSettings(domain.OperationTypeNixGenerations),
					},
				},
			},
		},
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
			Version:     "1.0.0",
			SafeMode:    true,
			MaxDiskUsage: 80,
			ProtectedPaths: []string{"/System"},
			Profiles: map[string]*PublicProfile{
				"test": {
					Name:        "test",
					Description: "Test profile",
					Enabled:     true,
					Operations: []PublicOperation{
						{
							Name:        "temp-cleanup",
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
		DryRun:     boolPtr(true),
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
	if domainRequest.Items[0].ScanType != domain.ScanTypeTemp {
		t.Errorf("Expected ScanTypeTemp, got %v", domainRequest.Items[0].ScanType)
	}

	if domainRequest.Items[0].Path != string(OperationTypeTempFiles) {
		t.Errorf("Expected temp-files path, got %s", domainRequest.Items[0].Path)
	}
}

func TestMapStrategy_Conversions(t *testing.T) {
	// Test all strategy conversions
	testCases := []struct {
		public  PublicStrategy
		domain  domain.CleanStrategyType
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
		public  PublicRiskLevel
		domain  domain.RiskLevelType
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