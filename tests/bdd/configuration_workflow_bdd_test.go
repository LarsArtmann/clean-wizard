package bdd

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// Configuration workflow BDD test contexts
type ConfigurationWorkflowContext struct {
	tempDir       string
	configFiles   map[string]string
	commandOutput string
	commandError  error
	exitCode      int
}

func TestConfigurationWorkflowBDD(t *testing.T) {
	suite := godog.TestSuite{
		Name: "Configuration Workflow",
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			InitializeConfigurationWorkflowContext(ctx.ScenarioContext())
		},
		Options: &godog.Options{
			Format: "cucumber",
			Paths:  []string{"configuration_workflow.feature"},
			Strict: true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("BDD tests failed")
	}
}

func InitializeConfigurationWorkflowContext(sc *godog.ScenarioContext) {
	context := &ConfigurationWorkflowContext{}

	// Background steps
	sc.Given(`^system has clean-wizard tool available$`, context.cleanWizardAvailable)
	sc.Given(`^a working directory for configuration files$`, context.setupWorkingDirectory)

	// Configuration file steps
	sc.Given(`^I have a valid configuration file at "([^"]+)"$`, context.haveValidConfigFile)
	sc.Given(`^I have an invalid configuration file at "([^"]+)"$`, context.haveInvalidConfigFile)
	sc.Given(`^I have a configuration file with safe_mode set to false$`, context.haveUnsafeConfigFile)
	sc.Given(`^I have a configuration file with minimal protected paths$`, context.haveBasicConfigFile)
	sc.Given(`^I have a configuration file with missing protected paths$`, context.haveIncompleteConfigFile)
	sc.Given(`^I have a configuration file with multiple profiles$`, context.haveMultiProfileConfigFile)

	// Configuration content steps
	sc.Given(`^configuration includes:$`, context.configIncludes)
	sc.Given(`^configuration includes a daily profile$`, func() error {
		return nil // Default configs include daily profile
	})
	sc.Given(`^the profiles include "([^"]+)" and "([^"]+)"$`, func(profile1, profile2 string) error {
		return nil // Multi-profile configs include both profiles
	})
	sc.Then(`^scan results should reflect daily profile settings$`, context.shouldSeeScanResultsReflectingDailyProfile)

	// Command execution steps
	sc.When(`^I run "([^"]+)"$`, context.runCommand)
	sc.When(`^I run "([^"]+)" with exit code (\d+)$`, context.runCommandWithExitCode)

	// Verification steps
	sc.Then(`^I should see "([^"]+)"$`, context.shouldSeeOutput)
	sc.Then(`^I should see scan results with generations$`, context.shouldSeeScanResults)
	sc.Then(`^I should see cleanup results with items cleaned$`, context.shouldSeeCleanResults)
	sc.Then(`^I should see "Using daily profile configuration"$`, context.shouldSeeDailyProfile)
	sc.Then(`^I should see "Running in DRY-RUN mode"$`, context.shouldSeeDryRunMode)
	sc.Then(`^I should see "Applying validation level: ([^"]+)"$`, context.shouldSeeValidationLevel)
	sc.Then(`^I should see "([^"]+)" validation error$`, context.shouldSeeValidationError)
	sc.Then(`^I should not see any validation errors$`, context.shouldNotSeeValidationErrors)
	sc.Then(`^command should complete successfully$`, context.shouldCompleteSuccessfully)
	sc.Then(`^command should fail with an error$`, context.shouldFailWithError)
	sc.Then(`^command should fail with validation error$`, context.shouldFailWithError)
}

func (c *ConfigurationWorkflowContext) cleanWizardAvailable() error {
	// Verify clean-wizard is available in PATH or local
	cmd := exec.Command("go", "run", "./cmd/clean-wizard", "--help")
	return cmd.Run()
}

func (c *ConfigurationWorkflowContext) setupWorkingDirectory() error {
	tempDir, err := ioutil.TempDir("", "clean-wizard-bdd-*")
	if err != nil {
		return err
	}
	c.tempDir = tempDir
	c.configFiles = make(map[string]string)
	return nil
}

func (c *ConfigurationWorkflowContext) haveValidConfigFile(filename string) error {
	configContent := `version: "1.0.0"
safe_mode: true
max_disk_usage: 50
protected:
  - "/System"
  - "/Library"
  - "/Applications"
  - "/usr"
  - "/etc"
  - "/var"
  - "/bin"
  - "/sbin"
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean Nix generations"
        risk_level: "LOW"
        enabled: true`

	configPath := filepath.Join(c.tempDir, filename)
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) haveInvalidConfigFile(filename string) error {
	configContent := `version: "1.0.0"
safe_mode: true
invalid_field: true
profiles: []`

	configPath := filepath.Join(c.tempDir, filename)
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) haveUnsafeConfigFile() error {
	configContent := `version: "1.0.0"
safe_mode: false
max_disk_usage: 50
protected:
  - "/System"
profiles:
  daily:
    name: "daily"
    enabled: true
    operations:
      - name: "nix-generations"
        risk_level: "LOW"
        enabled: true`

	configPath := filepath.Join(c.tempDir, "unsafe-config.yaml")
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) haveBasicConfigFile() error {
	configContent := `version: "1.0.0"
safe_mode: true
protected:
  - "/System"
profiles:
  daily:
    name: "daily"
    enabled: true`

	configPath := filepath.Join(c.tempDir, "basic-config.yaml")
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) haveIncompleteConfigFile() error {
	configContent := `version: "1.0.0"
safe_mode: true
protected: []`

	configPath := filepath.Join(c.tempDir, "incomplete-config.yaml")
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) haveMultiProfileConfigFile() error {
	configContent := `version: "1.0.0"
safe_mode: true
protected:
  - "/System"
  - "/Library"
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup"
    enabled: true
    operations:
      - name: "nix-generations"
        risk_level: "LOW"
        enabled: true
  weekly:
    name: "weekly"
    description: "Weekly cleanup"
    enabled: true
    operations:
      - name: "package-caches"
        risk_level: "MEDIUM"
        enabled: true`

	configPath := filepath.Join(c.tempDir, "multi-profile-config.yaml")
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) configIncludes(table *godog.Table) error {
	configContent := "version: \"1.0.0\"\nsafe_mode: true\n"

	for _, row := range table.Rows {
		field := row.Cells[0].Value
		value := row.Cells[1].Value

		switch field {
		case "max_disk_usage":
			configContent += fmt.Sprintf("max_disk_usage: %s\n", value)
		default:
			configContent += fmt.Sprintf("%s: %s\n", field, value)
		}
	}

	configPath := filepath.Join(c.tempDir, "table-config.yaml")
	return ioutil.WriteFile(configPath, []byte(configContent), 0644)
}

func (c *ConfigurationWorkflowContext) runCommand(commandStr string) error {
	// Replace config file paths with temp directory paths
	commandStr = strings.ReplaceAll(commandStr, "working-config.yaml", filepath.Join(c.tempDir, "working-config.yaml"))
	commandStr = strings.ReplaceAll(commandStr, "invalid-config.yaml", filepath.Join(c.tempDir, "invalid-config.yaml"))
	commandStr = strings.ReplaceAll(commandStr, "unsafe-config.yaml", filepath.Join(c.tempDir, "unsafe-config.yaml"))
	commandStr = strings.ReplaceAll(commandStr, "basic-config.yaml", filepath.Join(c.tempDir, "basic-config.yaml"))
	commandStr = strings.ReplaceAll(commandStr, "incomplete-config.yaml", filepath.Join(c.tempDir, "incomplete-config.yaml"))
	commandStr = strings.ReplaceAll(commandStr, "multi-profile-config.yaml", filepath.Join(c.tempDir, "multi-profile-config.yaml"))

	// Change to project directory for go run
	projectDir, _ := filepath.Abs("../../../")

	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", projectDir, commandStr))
	output, err := cmd.CombinedOutput()

	c.commandOutput = string(output)
	c.commandError = err

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitError.ExitCode()
		} else {
			c.exitCode = 1
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *ConfigurationWorkflowContext) runCommandWithExitCode(commandStr string, expectedExitCode int) error {
	err := c.runCommand(commandStr)
	if c.exitCode != expectedExitCode {
		return fmt.Errorf("expected exit code %d, got %d", expectedExitCode, c.exitCode)
	}
	return err
}

func (c *ConfigurationWorkflowContext) shouldSeeOutput(expectedText string) error {
	if !strings.Contains(c.commandOutput, expectedText) {
		return fmt.Errorf("expected output to contain '%s', but got:\n%s", expectedText, c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeScanResults() error {
	if !strings.Contains(c.commandOutput, "Total generations:") {
		return fmt.Errorf("expected scan results with generations, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeCleanResults() error {
	if !strings.Contains(c.commandOutput, "items") && !strings.Contains(c.commandOutput, "would be cleaned") {
		return fmt.Errorf("expected clean results with items, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeDailyProfile() error {
	if !strings.Contains(c.commandOutput, "Using daily profile configuration") {
		return fmt.Errorf("expected to see daily profile usage, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeDryRunMode() error {
	if !strings.Contains(c.commandOutput, "DRY-RUN mode") {
		return fmt.Errorf("expected to see dry-run mode, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeValidationLevel(expectedLevel string) error {
	expectedText := fmt.Sprintf("Applying validation level: %s", expectedLevel)
	if !strings.Contains(c.commandOutput, expectedText) {
		return fmt.Errorf("expected to see validation level %s, but got:\n%s", expectedLevel, c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeValidationError(errorType string) error {
	if !strings.Contains(c.commandOutput, "validation failed") {
		return fmt.Errorf("expected to see validation error, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldNotSeeValidationErrors() error {
	if strings.Contains(c.commandOutput, "validation failed") || strings.Contains(c.commandOutput, "validation error") {
		return fmt.Errorf("expected no validation errors, but got:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldCompleteSuccessfully() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected command to complete successfully (exit code 0), but got exit code %d with output:\n%s", c.exitCode, c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldFailWithError() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected command to fail with error, but it completed successfully with output:\n%s", c.commandOutput)
	}
	return nil
}

func (c *ConfigurationWorkflowContext) shouldSeeScanResultsReflectingDailyProfile() error {
	if !strings.Contains(c.commandOutput, "Using daily profile configuration") {
		return fmt.Errorf("expected scan results to reflect daily profile, but got:\n%s", c.commandOutput)
	}
	return nil
}
