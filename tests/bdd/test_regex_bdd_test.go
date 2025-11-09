package bdd

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestRegexBDD(t *testing.T) {
	status := godog.TestSuite{
		Name: "Regex Test",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeTestScenario(ctx.ScenarioContext())
		},
		Options: &godog.Options{
			Format: "cucumber",
			Paths:  []string{"test_regex.feature"},
		},
	}.Run()

	if status != 0 {
		t.Errorf("Regex BDD tests failed with status: %d", status)
	}
}

func InitializeTestScenario(ctx *godog.ScenarioContext) {
	ctx.Given(`^I want to keep the last (\d+) generations$`, func(count int) error {
		return nil
	})
}