package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// BDDFeature represents a BDD feature for behavior-driven development
type BDDFeature struct {
	Name        string
	Description  string
	Background  string
	Scenarios   []BDDScenario
}

// BDDScenario represents a single BDD test scenario
type BDDScenario struct {
	Name        string
	Description string
	Given       []BDDGiven
	When        []BDDWhen
	Then        []BDDThen
}

// BDDGiven represents the initial state in BDD
type BDDGiven struct {
	Description string
	Setup       func() (*domain.Config, error)
}

// BDDWhen represents the action in BDD
type BDDWhen struct {
	Description string
	Action      func(*domain.Config) (*ValidationResult, error)
}

// BDDThen represents the expected outcome in BDD
type BDDThen struct {
	Description string
	Validate    func(*ValidationResult) error
}

// BDDTestRunner provides comprehensive BDD test execution
type BDDTestRunner struct {
	t         *testing.T
	feature   BDDFeature
	validator *ConfigValidator
}

// NewBDDTestRunner creates a new BDD test runner
func NewBDDTestRunner(t *testing.T, feature BDDFeature) *BDDTestRunner {
	return &BDDTestRunner{
		t:         t,
		feature:   feature,
		validator: NewConfigValidator(),
	}
}

// RunFeature executes all scenarios in a BDD feature
func (b *BDDTestRunner) RunFeature() {
	b.t.Logf("Feature: %s", b.feature.Name)
	b.t.Logf("Description: %s", b.feature.Description)
	if b.feature.Background != "" {
		b.t.Logf("Background: %s", b.feature.Background)
	}

	for _, scenario := range b.feature.Scenarios {
		b.runScenario(scenario)
	}
}

// runScenario executes a single BDD scenario
func (b *BDDTestRunner) runScenario(scenario BDDScenario) {
	b.t.Logf("Scenario: %s", scenario.Name)
	b.t.Logf("Description: %s", scenario.Description)

	// Setup Given conditions
	var cfg *domain.Config
	var setupErr error

	for i, given := range scenario.Given {
		b.t.Logf("  Given %d: %s", i+1, given.Description)
		if given.Setup == nil {
			b.t.Errorf("Given step %d has no setup function", i+1)
			return
		}

		cfg, setupErr = given.Setup()
		if setupErr != nil {
			b.t.Errorf("Given setup failed: %v", setupErr)
			return
		}
	}

	// Execute When actions
	var result *ValidationResult
	var actionErr error

	for i, when := range scenario.When {
		b.t.Logf("  When %d: %s", i+1, when.Description)
		if when.Action == nil {
			b.t.Errorf("When step %d has no action function", i+1)
			return
		}

		result, actionErr = when.Action(cfg)
		if actionErr != nil {
			b.t.Errorf("When action failed: %v", actionErr)
			return
		}
	}

	// Validate Then expectations
	for i, then := range scenario.Then {
		b.t.Logf("  Then %d: %s", i+1, then.Description)
		if then.Validate == nil {
			b.t.Errorf("Then step %d has no validation function", i+1)
			return
		}

		if err := then.Validate(result); err != nil {
			b.t.Errorf("Then validation failed: %v", err)
			return
		}
	}

	b.t.Logf("✓ Scenario passed: %s", scenario.Name)
}