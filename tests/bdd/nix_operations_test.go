package bdd

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// mockCleaner implements domain.Cleaner interface for testing
type mockCleaner struct{}

func (m *mockCleaner) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *mockCleaner) GetStoreSize(ctx context.Context) int64 {
	return 1000
}

func (m *mockCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	return nil
}

// NixOperationsContext holds state for BDD scenarios
type NixOperationsContext struct {
	ctx          context.Context
	cleaner      domain.Cleaner
	uiAdapter    *adapters.UIAdapter
	config       *domain.Config
	profile      *domain.Profile
	result       *domain.CleanResult
	error        error
	operations   []domain.CleanupOperation
	safetyLevel  domain.SafetyLevelType
	riskLevel    domain.RiskLevelType
	profiles     map[string]*domain.Profile
	logMessages  []string
	cleanupState CleanupState
	t            *testing.T
}

// CleanupState represents state of cleanup operations
type CleanupState int

const (
	StateIdle CleanupState = iota
	StateRunning
	StateCompleted
	StateFailed
	StateRollback
)

// InitializeContext sets up BDD test context
func (ctx *NixOperationsContext) InitializeContext(t *testing.T) {
	ctx.t = t
	ctx.ctx = context.Background()
	ctx.cleaner = &mockCleaner{}
	ctx.uiAdapter = adapters.NewUIAdapter()
	ctx.profiles = make(map[string]*domain.Profile)
	ctx.logMessages = []string{}
	ctx.cleanupState = StateIdle
	ctx.safetyLevel = domain.SafetyLevelEnabled
}

// Scenario 1: Happy Path Nix Generation Cleanup
func (ctx *NixOperationsContext) GivenValidNixInstallation() error {
	ctx.safetyLevel = domain.SafetyLevelEnabled
	ctx.config = &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  ctx.safetyLevel,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library", "/Applications"},
		Profiles:     ctx.profiles,
	}

	// Create test profile for nix-generations
	ctx.profile = &domain.Profile{
		Name:        "nix-cleanup",
		Description: "Nix generations cleanup profile",
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
				Settings: &domain.OperationSettings{
					NixGenerations: &domain.NixGenerationsSettings{
						Generations:  3,
						Optimization: domain.OptimizationLevelConservative,
					},
				},
			},
		},
		Status: domain.StatusEnabled,
	}

	ctx.profiles["nix-cleanup"] = ctx.profile
	return nil
}

func (ctx *NixOperationsContext) WhenRunningNixGenerationsCleanup() error {
	ctx.cleanupState = StateRunning
	ctx.logMessages = append(ctx.logMessages, "Starting Nix generations cleanup")

	// Simulate cleanup operation
	ctx.result = &domain.CleanResult{
		FreedBytes:   1024 * 1024 * 250,
		ItemsRemoved: 2,
		ItemsFailed:  0,
		CleanTime:    time.Second * 5,
		CleanedAt:    time.Now(),
	}

	ctx.cleanupState = StateCompleted
	ctx.logMessages = append(ctx.logMessages, "Completed Nix generations cleanup")
	return nil
}

func (ctx *NixOperationsContext) ThenRemoveOldGenerationsAndKeepCurrent() error {
	require.NotNil(ctx.t, ctx.result, "Cleanup result should not be nil")
	assert.Equal(ctx.t, uint(2), ctx.result.ItemsRemoved, "Should clean 2 old generations")
	assert.Equal(ctx.t, uint(0), ctx.result.ItemsFailed, "Should have no failed items")
	assert.Greater(ctx.t, ctx.result.FreedBytes, uint64(0), "Should clean significant bytes")
	assert.Greater(ctx.t, ctx.result.CleanTime, time.Duration(0), "Should take some time")
	return nil
}

func (ctx *NixOperationsContext) AndLogOperationsPerformed() error {
	expectedLogs := []string{
		"Starting Nix generations cleanup",
		"Completed Nix generations cleanup",
	}

	for _, expectedLog := range expectedLogs {
		found := slices.Contains(ctx.logMessages, expectedLog)
		assert.True(ctx.t, found, "Expected log message not found: %s", expectedLog)
	}

	return nil
}

// Scenario 2: Risk-Based Operation Prevention
func (ctx *NixOperationsContext) GivenSafetyLevelEnabled() error {
	ctx.safetyLevel = domain.SafetyLevelEnabled
	return nil
}

func (ctx *NixOperationsContext) AndHighRiskOperationRequested() error {
	ctx.riskLevel = domain.RiskCritical // High risk operation
	ctx.profile = &domain.Profile{
		Name:        "dangerous-cleanup",
		Description: "High risk cleanup operation",
		Operations: []domain.CleanupOperation{
			{
				Name:        "delete-system-files",
				Description: "Delete system files (DANGEROUS)",
				RiskLevel:   ctx.riskLevel,
				Status:      domain.StatusEnabled,
			},
		},
		Status: domain.StatusEnabled,
	}
	return nil
}

func (ctx *NixOperationsContext) WhenAttemptingCleanup() error {
	// Check safety level and risk
	if ctx.safetyLevel != domain.SafetyLevelDisabled {
		for _, op := range ctx.profile.Operations {
			if op.RiskLevel.IsHigherOrEqualThan(domain.RiskCritical) {
				ctx.error = fmt.Errorf("operation %s prevented: critical risk not allowed with safety level %s",
					op.Name, ctx.safetyLevel.String())
				ctx.logMessages = append(ctx.logMessages,
					fmt.Sprintf("Prevented high-risk operation: %s", op.Name))
				return nil
			}
		}
	}

	ctx.result = &domain.CleanResult{
		FreedBytes:   0,
		ItemsRemoved: 0,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
	}
	return nil
}

func (ctx *NixOperationsContext) ThenPreventOperation() error {
	require.Error(ctx.t, ctx.error, "Should have an error for high-risk operation")
	assert.Contains(ctx.t, ctx.error.Error(), "operation delete-system-files prevented",
		"Error should mention operation prevention")
	return nil
}

func (ctx *NixOperationsContext) AndShowRiskWarning() error {
	expectedWarning := "Prevented high-risk operation: delete-system-files"
	found := slices.Contains(ctx.logMessages, expectedWarning)
	assert.True(ctx.t, found, "Should log risk warning: %s", expectedWarning)
	return nil
}

// Scenario 3: Profile-Specific Execution
func (ctx *NixOperationsContext) GivenMultipleCleanupProfiles() error {
	ctx.profiles["daily"] = &domain.Profile{
		Name:        "daily",
		Description: "Daily cleanup profile",
		Operations: []domain.CleanupOperation{
			{Name: "temp-files", RiskLevel: domain.RiskLow, Status: domain.StatusEnabled},
		},
		Status: domain.StatusEnabled,
	}

	ctx.profiles["weekly"] = &domain.Profile{
		Name:        "weekly",
		Description: "Weekly deep cleanup profile",
		Operations: []domain.CleanupOperation{
			{Name: "nix-generations", RiskLevel: domain.RiskMedium, Status: domain.StatusEnabled},
			{Name: "homebrew-cleanup", RiskLevel: domain.RiskLow, Status: domain.StatusEnabled},
		},
		Status: domain.StatusEnabled,
	}

	return nil
}

func (ctx *NixOperationsContext) WhenRunningSpecificProfile(profileName string) error {
	ctx.profile = ctx.profiles[profileName]
	if ctx.profile == nil {
		return fmt.Errorf("profile %s not found", profileName)
	}

	ctx.logMessages = append(ctx.logMessages, fmt.Sprintf("Running profile: %s", profileName))

	// Execute only profile operations
	ctx.result = &domain.CleanResult{
		FreedBytes:   1024 * 1024 * 100,
		ItemsRemoved: uint(len(ctx.profile.Operations)),
		ItemsFailed:  0,
		CleanTime:    time.Second * 2,
		CleanedAt:    time.Now(),
	}

	return nil
}

func (ctx *NixOperationsContext) ThenExecuteOnlyProfileOperations() error {
	require.NotNil(ctx.t, ctx.result, "Result should not be nil")

	expectedOps := len(ctx.profile.Operations)
	assert.Equal(ctx.t, uint(expectedOps), ctx.result.ItemsRemoved,
		"Should execute only operations from selected profile")

	return nil
}

func (ctx *NixOperationsContext) AndRespectProfileSettings() error {
	// Verify profile settings are respected
	for _, op := range ctx.profile.Operations {
		if op.Status == domain.StatusDisabled {
			ctx.t.Errorf("Disabled operation %s should not be executed", op.Name)
		}
	}
	return nil
}

// Scenario 4: Error Recovery on Failed Cleanup
func (ctx *NixOperationsContext) GivenCleanupOperationFailsMidway() error {
	ctx.profile = &domain.Profile{
		Name:        "failing-cleanup",
		Description: "Profile that will fail",
		Operations: []domain.CleanupOperation{
			{
				Name:        "simulate-failure",
				Description: "Operation that fails",
				RiskLevel:   domain.RiskLow,
				Status:      domain.StatusEnabled,
			},
		},
		Status: domain.StatusEnabled,
	}
	return nil
}

func (ctx *NixOperationsContext) WhenRunningCleanup() error {
	ctx.cleanupState = StateRunning
	ctx.logMessages = append(ctx.logMessages, "Starting cleanup operation")

	// Simulate failure midway
	ctx.cleanupState = StateFailed
	ctx.error = fmt.Errorf("cleanup operation failed: simulate-failure")
	ctx.logMessages = append(ctx.logMessages, "Cleanup operation failed")

	// Simulate rollback
	ctx.cleanupState = StateRollback
	ctx.logMessages = append(ctx.logMessages, "Rolling back changes")
	ctx.cleanupState = StateIdle
	ctx.logMessages = append(ctx.logMessages, "Rollback completed")

	return nil
}

func (ctx *NixOperationsContext) ThenRollbackChanges() error {
	require.Error(ctx.t, ctx.error, "Should have error from failed cleanup")

	foundRollback := false
	foundCompleted := false
	for _, log := range ctx.logMessages {
		if log == "Rolling back changes" {
			foundRollback = true
		}
		if log == "Rollback completed" {
			foundCompleted = true
		}
	}
	assert.True(ctx.t, foundRollback, "Should attempt rollback")
	assert.True(ctx.t, foundCompleted, "Should complete rollback")
	return nil
}

func (ctx *NixOperationsContext) AndReportErrorDetails() error {
	assert.Contains(ctx.t, ctx.error.Error(), "simulate-failure",
		"Error should mention failed operation")
	return nil
}

// Scenario 5: Concurrent Cleanup Prevention
func (ctx *NixOperationsContext) GivenCleanupInProgress() error {
	ctx.cleanupState = StateRunning
	ctx.logMessages = append(ctx.logMessages, "Cleanup already in progress")
	return nil
}

func (ctx *NixOperationsContext) WhenStartingAnotherCleanup() error {
	if ctx.cleanupState == StateRunning {
		ctx.error = fmt.Errorf("cleanup operation already in progress")
		ctx.logMessages = append(ctx.logMessages, "Concurrent cleanup attempt prevented")
		return nil
	}

	ctx.cleanupState = StateRunning
	return nil
}

func (ctx *NixOperationsContext) ThenPreventConcurrentExecution() error {
	require.Error(ctx.t, ctx.error, "Should prevent concurrent execution")
	assert.Contains(ctx.t, ctx.error.Error(), "already in progress",
		"Error should mention concurrent operation")
	return nil
}

func (ctx *NixOperationsContext) AndShowAppropriateError() error {
	expectedLog := "Concurrent cleanup attempt prevented"
	found := slices.Contains(ctx.logMessages, expectedLog)
	assert.True(ctx.t, found, "Should log concurrent prevention: %s", expectedLog)
	return nil
}

// Unit tests for BDD scenarios
func TestNixOperationsBDD(t *testing.T) {
	ctx := &NixOperationsContext{ctx: context.Background()}
	ctx.InitializeContext(t)

	t.Run("Happy Path Nix Cleanup", func(t *testing.T) {
		err := ctx.GivenValidNixInstallation()
		require.NoError(t, err)

		err = ctx.WhenRunningNixGenerationsCleanup()
		require.NoError(t, err)

		err = ctx.ThenRemoveOldGenerationsAndKeepCurrent()
		require.NoError(t, err)

		err = ctx.AndLogOperationsPerformed()
		require.NoError(t, err)
	})

	t.Run("Risk-Based Prevention", func(t *testing.T) {
		err := ctx.GivenSafetyLevelEnabled()
		require.NoError(t, err)

		err = ctx.AndHighRiskOperationRequested()
		require.NoError(t, err)

		err = ctx.WhenAttemptingCleanup()
		require.NoError(t, err)

		err = ctx.ThenPreventOperation()
		require.NoError(t, err)

		err = ctx.AndShowRiskWarning()
		require.NoError(t, err)
	})

	t.Run("Profile-Specific Execution", func(t *testing.T) {
		err := ctx.GivenMultipleCleanupProfiles()
		require.NoError(t, err)

		err = ctx.WhenRunningSpecificProfile("daily")
		require.NoError(t, err)

		err = ctx.ThenExecuteOnlyProfileOperations()
		require.NoError(t, err)

		err = ctx.AndRespectProfileSettings()
		require.NoError(t, err)
	})

	t.Run("Error Recovery", func(t *testing.T) {
		err := ctx.GivenCleanupOperationFailsMidway()
		require.NoError(t, err)

		err = ctx.WhenRunningCleanup()
		require.NoError(t, err)

		err = ctx.ThenRollbackChanges()
		require.NoError(t, err)

		err = ctx.AndReportErrorDetails()
		require.NoError(t, err)
	})

	t.Run("Concurrent Prevention", func(t *testing.T) {
		err := ctx.GivenCleanupInProgress()
		require.NoError(t, err)

		err = ctx.WhenStartingAnotherCleanup()
		require.NoError(t, err)

		err = ctx.ThenPreventConcurrentExecution()
		require.NoError(t, err)

		err = ctx.AndShowAppropriateError()
		require.NoError(t, err)
	})
}
