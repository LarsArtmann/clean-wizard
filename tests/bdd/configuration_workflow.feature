Feature: Configuration-driven cleanup workflow
  As a system administrator
  I want to use configuration files to control cleanup operations
  So that I can customize and automate cleanup strategies

  Background:
    Given system has clean-wizard tool available
    And a working directory for configuration files

  Scenario: Scan with valid configuration file
    Given I have a valid configuration file at "working-config.yaml"
    And the configuration includes:
      | field        | value                    |
      | version      | "1.0.0"                 |
      | safe_mode    | true                     |
      | max_disk_usage | 50                    |
    When I run "clean-wizard scan --config working-config.yaml"
    Then I should see "Loading configuration from working-config.yaml"
    And I should see "Configuration applied: safe_mode=true"
    And I should see scan results with generations
    And command should complete successfully

  Scenario: Clean with valid configuration file (dry-run)
    Given I have a valid configuration file at "working-config.yaml"
    And the configuration includes a daily profile
    When I run "clean-wizard clean --config working-config.yaml --dry-run"
    Then I should see "Loading configuration from working-config.yaml"
    And I should see "Configuration applied: safe_mode=true"
    And I should see "Using daily profile configuration"
    And I should see "Running in DRY-RUN mode"
    And I should see cleanup results with items cleaned
    And command should complete successfully

  Scenario: Scan with invalid configuration file
    Given I have an invalid configuration file at "invalid-config.yaml"
    When I run "clean-wizard scan --config invalid-config.yaml"
    Then I should see "failed to load configuration"
    And command should fail with an error

  Scenario: Clean with missing configuration file
    When I run "clean-wizard clean --config missing-config.yaml"
    Then I should see "failed to load configuration"
    And command should fail with an error

  Scenario: Clean with basic validation level
    Given I have a configuration file with minimal protected paths
    When I run "clean-wizard clean --config basic-config.yaml --validation-level basic"
    Then I should see "Applying validation level: Basic"
    And command should complete successfully

  Scenario: Clean with strict validation level on unsafe configuration
    Given I have a configuration file with safe_mode set to false
    When I run "clean-wizard clean --config unsafe-config.yaml --validation-level strict"
    Then I should see "strict validation failed: safe_mode must be enabled"
    And command should fail with validation error

  Scenario: Use validation level none to bypass validation
    Given I have a configuration file with missing protected paths
    When I run "clean-wizard scan --config incomplete-config.yaml --validation-level none"
    Then I should not see any validation errors
    And command should complete successfully

  Scenario: Profile-based configuration works
    Given I have a configuration file with multiple profiles
    And the profiles include "daily" and "weekly"
    When I run "clean-wizard scan --config multi-profile-config.yaml"
    Then I should see "Using daily profile configuration"
    And scan results should reflect daily profile settings