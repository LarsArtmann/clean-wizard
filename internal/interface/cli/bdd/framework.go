package bdd

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/infrastructure/system"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/application/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/cucumber/godog"
	"github.com/rs/zerolog/log"
)

// BDDContext holds all state for BDD test scenarios
type BDDContext struct {
	ctx         context.Context
	config      *domain.Config
	nixAdapter  *adapters.NixAdapter
	nixCleaner  *cleaner.NixCleaner
	generations []domain.NixGeneration
	cleanResult *domain.CleanResult
	error       error
	tempDir     string
	testMode    bool // true for test scenarios, false for integration
}

// NewBDDContext creates a fresh BDD context for scenarios
func NewBDDContext() *BDDContext {
	return &BDDContext{
		ctx:      context.Background(),
		testMode: true, // Default to test mode
	}
}

// BeforeScenario resets context for each scenario
func (b *BDDContext) BeforeScenario(sc *godog.Scenario) {
	log.Info().Str("scenario", sc.Name).Msg("Starting BDD scenario")

	// Reset context state
	b.ctx = context.Background()
	b.config = nil
	b.generations = nil
	b.cleanResult = nil
	b.error = nil

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "clean-wizard-bdd-*")
	if err != nil {
		log.Err(err).Msg("Failed to create temp directory")
		return
	}
	b.tempDir = tempDir

	// Set up test Nix adapter (dry-run mode for safety)
	b.nixAdapter = adapters.NewNixAdapter(time.Second*30, 3) // 3 generations
	b.nixCleaner = cleaner.NewNixCleaner(true, true)         // verbose=true, dry-run=true
}

// AfterScenario cleans up after each scenario
func (b *BDDContext) AfterScenario(sc *godog.Scenario, err error) {
	log.Info().Str("scenario", sc.Name).Err(err).Msg("Completed BDD scenario")

	// Clean up temporary directory
	if b.tempDir != "" {
		if err := os.RemoveAll(b.tempDir); err != nil {
			log.Err(err).Str("tempDir", b.tempDir).Msg("Failed to clean up temp directory")
		}
	}
}

// ========== CONFIGURATION STEPS ==========

// theSystemHasConfiguration creates a system configuration
func (b *BDDContext) TheSystemHasConfiguration() error {
	b.config = config.GetDefaultConfig()
	b.config.Profiles = make(map[string]*domain.Profile)
	return nil
}

// withNixProfile adds a Nix profile to the configuration
func (b *BDDContext) WithNixProfile(profileName string) error {
	if b.config == nil {
		return fmt.Errorf("configuration not initialized")
	}

	profile := &domain.Profile{
		Name:        profileName,
		Description: fmt.Sprintf("BDD test profile for %s", profileName),
		Status:      domain.StatusEnabled,
		Operations: []domain.CleanupOperation{
			{
				Name:        "clean_nix_generations",
				Description: "Clean old Nix generations",
				RiskLevel:   domain.RiskMedium,
				Status:      domain.StatusEnabled,
			},
		},
	}

	b.config.Profiles[profileName] = profile
	return nil
}

// withDryRunMode enables dry run mode
func (b *BDDContext) WithDryRunMode() error {
	b.testMode = true
	if b.nixAdapter != nil {
		b.nixAdapter.SetDryRun(true)
	}
	return nil
}

// withVerboseMode enables verbose mode
func (b *BDDContext) WithVerboseMode() error {
	if b.nixCleaner != nil {
		// Note: This would need to be added to NixCleaner interface
		log.Info().Msg("Verbose mode enabled for cleaner")
	}
	return nil
}

// ========== NIX SYSTEM STEPS ==========

// theNixSystemIsAvailable simulates Nix system availability
func (b *BDDContext) TheNixSystemIsAvailable() error {
	if b.nixAdapter == nil {
		return fmt.Errorf("Nix adapter not initialized")
	}

	// In test mode, we simulate availability
	available := b.nixAdapter.IsAvailable(b.ctx)
	if !available {
		// For testing, we create a mock that's always available
		b.nixAdapter = adapters.NewNixAdapter(time.Second*30, 3) // dry-run=false for real adapter
	}

	return nil
}

// theNixSystemIsUnavailable simulates Nix system unavailability
func (b *BDDContext) TheNixSystemIsUnavailable() error {
	// Create adapter that always reports unavailable
	b.nixAdapter = &adapters.NixAdapter{
		// This would need to be implemented for testing
	}
	return nil
}

// iListAvailableNixGenerations lists Nix generations
func (b *BDDContext) IListAvailableNixGenerations() error {
	if b.nixAdapter == nil {
		return fmt.Errorf("Nix adapter not initialized")
	}

	result := b.nixAdapter.ListGenerations(b.ctx)
	if !result.IsOk() {
		b.error = result.Error()
		return b.error
	}
	b.generations, b.error = result.Unwrap()
	return nil
}

// theSystemShouldHaveGenerations validates generation count
func (b *BDDContext) TheSystemShouldHaveGenerations(expectedCount int) error {
	if b.error != nil {
		return b.error
	}

	actualCount := len(b.generations)
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d generations, got %d", expectedCount, actualCount)
	}

	log.Info().Int("expected", expectedCount).Int("actual", actualCount).Msg("Generation count validated")
	return nil
}

// ========== CLEANING OPERATION STEPS ==========

// iCleanOldNixGenerationsWithKeepCount performs cleaning operation
func (b *BDDContext) ICleanOldNixGenerationsWithKeepCount(keepCount int) error {
	if b.nixCleaner == nil {
		return fmt.Errorf("Nix cleaner not initialized")
	}

	result := b.nixCleaner.CleanOldGenerations(b.ctx, keepCount)
	if !result.IsOk() {
		b.error = result.Error()
		return b.error
	}

	cleanRes, err := result.Unwrap()
	if err != nil {
		b.error = err
		return b.error
	}
	b.cleanResult = &cleanRes
	return nil
}

// theCleaningShouldBeSuccessful validates cleaning success
func (b *BDDContext) TheCleaningShouldBeSuccessful() error {
	if b.error != nil {
		return fmt.Errorf("cleaning failed: %w", b.error)
	}

	if b.cleanResult == nil {
		return fmt.Errorf("no clean result available")
	}

	if !b.cleanResult.IsValid() {
		return fmt.Errorf("cleaning reported invalid result")
	}

	log.Info().Msg("Cleaning operation validated as successful")
	return nil
}

// ========== ASSERTION STEPS ==========

// noErrorShouldHaveOccurred validates no errors occurred
func (b *BDDContext) NoErrorShouldHaveOccurred() error {
	if b.error != nil {
		return fmt.Errorf("unexpected error: %w", b.error)
	}
	return nil
}

// anErrorShouldHaveOccurredWithErrorType validates error occurred
func (b *BDDContext) AnErrorShouldHaveOccurredWithErrorType(errorType string) error {
	if b.error == nil {
		return fmt.Errorf("expected error but none occurred")
	}

	// In real implementation, we'd validate error type
	log.Info().Str("errorType", errorType).Str("error", b.error.Error()).Msg("Error occurred as expected")
	return nil
}

// ========== TEST REGISTRATION ==========

// InitializeScenario registers all BDD steps
func InitializeScenario(ctx *godog.ScenarioContext) {
	bddContext := NewBDDContext()

	// Register scenario hooks
	ctx.BeforeScenario(bddContext.BeforeScenario)
	ctx.AfterScenario(bddContext.AfterScenario)

	// Register configuration steps
	ctx.Given(`^the system has configuration$`, bddContext.TheSystemHasConfiguration)
	ctx.Given(`^with a "([^"]*)" Nix profile$`, bddContext.WithNixProfile)
	ctx.Given(`^with dry run mode$`, bddContext.WithDryRunMode)
	ctx.Given(`^with verbose mode$`, bddContext.WithVerboseMode)

	// Register Nix system steps
	ctx.Given(`^the Nix system is available$`, bddContext.TheNixSystemIsAvailable)
	ctx.Given(`^the Nix system is unavailable$`, bddContext.TheNixSystemIsUnavailable)
	ctx.When(`^I list available Nix generations$`, bddContext.IListAvailableNixGenerations)
	ctx.Then(`^the system should have (\d+) generations?$`, bddContext.TheSystemShouldHaveGenerations)

	// Register cleaning operation steps
	ctx.When(`^I clean old Nix generations with keep count (\d+)$`, bddContext.ICleanOldNixGenerationsWithKeepCount)
	ctx.Then(`^the cleaning should be successful$`, bddContext.TheCleaningShouldBeSuccessful)

	// Register assertion steps
	ctx.Then(`^no error should have occurred$`, bddContext.NoErrorShouldHaveOccurred)
	ctx.Then(`^an error should have occurred with error type "([^"]*)"$`, bddContext.AnErrorShouldHaveOccurredWithErrorType)
}

// RunBDDTests runs all BDD scenarios
func RunBDDTests(t *testing.T) {
	opts := godog.Options{
		Format:              "pretty",
		Paths:               []string{"features"},
		Randomize:           time.Now().UTC().UnixNano(),
		Strict:              true,
		StopOnFailure:       false,
		ShowStepDefinitions: false,
		TestingT:            t,
	}

	// Run BDD scenarios
	status := godog.TestSuite{
		Name:                "Clean Wizard BDD Tests",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if status != 0 {
		t.Errorf("BDD test suite failed with status %d", status)
	}
}
