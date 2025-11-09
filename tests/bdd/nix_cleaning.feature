Feature: Nix Store Cleaning
  As a Sr. Software Architect using Nix
  I want to safely clean old Nix generations
  So that I can recover disk space without breaking my development environment

  Scenario: List available Nix generations
    Given I have Nix installed
    When I run "clean-wizard scan"
    Then I should see generation numbers and dates
    And I should see current generation marked

  Scenario: Clean old Nix generations safely
    Given I have multiple Nix generations
    And I have at least 2 generations
    When I run "clean-wizard clean --dry-run"
    Then I should see which generations would be deleted
    And current generation should not be deleted
    And I should get confirmation before real deletion

  Scenario: Prevent deletion of current generation
    Given I have Nix generations
    And generation 300 is current
    When I clean old generations
    Then generation 300 should still be present
    And my development environment should still work

  Scenario: Clean with dry-run mode
    Given I want to test Nix cleaning
    When I run "clean-wizard clean --dry-run"
    Then I should see what would be deleted
    And no generations should actually be deleted
    And I should get confirmation message

  Scenario: Clean with backup protection
    Given I have important Nix generations
    When I run "clean-wizard clean --dry-run"
    Then important generations should be protected
    And I should see space estimation
    And I should get confirmation prompt

  Scenario: Verify type-safe operations
    Given I am cleaning Nix store
    When I run cleaning operations
    Then all operations should be type-safe
    And no invalid states should be possible
    And error handling should be consistent
