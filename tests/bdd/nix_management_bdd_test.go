package bdd

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestNixManagementBDD(t *testing.T) {
	status := godog.TestSuite{
		Name: "Nix Management",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeScenario(ctx.ScenarioContext())
		},
		Options: &godog.Options{
			Format: "cucumber",
			Paths:  []string{"nix_management.feature"},
		},
	}.Run()

	if status != 0 {
		t.Errorf("BDD tests failed with status: %d", status)
	}
}
