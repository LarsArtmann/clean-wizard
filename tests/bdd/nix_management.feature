Feature: Nix Store Management
  As a system administrator
  I want to manage Nix store generations safely
  So that I can free disk space without breaking my system

  Background:
    Given the system has Nix package manager installed
    And the clean-wizard tool is available

  Scenario: List available Nix generations
    When I run \"clean-wizard scan nix\"
    Then I should see a list of Nix generations
    And each generation should have an ID
    And each generation should have a creation date
    And the total store size should be displayed
    And the command should complete successfully

  Scenario: Clean old Nix generations safely
    Given the system has multiple Nix generations
    When I run \"clean-wizard clean --dry-run\"
    Then I should see what would be cleaned
    And I should see the estimated space freed
    And I should see how many generations would be removed
    And no actual cleaning should be performed
    And the command should complete successfully

  Scenario: Clean old Nix generations for real
    Given the system has multiple Nix generations
    And I want to keep the last (d+) generations
    When I run \"clean-wizard clean --keep 3\"
    Then old generations should be removed
    And disk space should be freed
    And the last (d+) generations should remain
    And the command should complete successfully

  Scenario: Handle Nix not available gracefully
    Given Nix package manager is not installed
    When I run \"clean-wizard scan nix\"
    Then I should see a helpful error message
    And the command should fail gracefully
    And I should not see a stack trace