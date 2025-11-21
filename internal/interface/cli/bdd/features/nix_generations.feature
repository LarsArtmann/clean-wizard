Feature: Nix Generation Management
  As a system administrator
  I want to safely manage Nix store generations
  To maintain system stability and reclaim storage space

  Scenario: List available Nix generations when system is available
    Given the system has configuration
    And with a "nix_cleanup" Nix profile
    And with dry run mode
    And with verbose mode
    And the Nix system is available
    When I list available Nix generations
    Then the system should have at least 1 generations
    And no error should have occurred

  Scenario: List Nix generations when system is unavailable
    Given the system has configuration
    And with a "nix_cleanup" Nix profile
    And with dry run mode
    And the Nix system is unavailable
    When I list available Nix generations
    Then an error should have occurred with error type "SystemUnavailable"
    And no generations should be listed

  Scenario: Clean old Nix generations in dry run mode
    Given the system has configuration
    And with a "nix_cleanup" Nix profile
    And with dry run mode
    And the Nix system is available
    And I list available Nix generations
    And the system should have at least 3 generations
    When I clean old Nix generations with keep count 2
    Then the cleaning should be successful
    And the cleaning should report operations planned
    And no actual generations should be removed

  Scenario: Clean Nix generations when few generations exist
    Given the system has configuration
    And with a "nix_cleanup" Nix profile
    And with dry run mode
    And the Nix system is available
    And I list available Nix generations
    And the system should have exactly 2 generations
    When I clean old Nix generations with keep count 3
    Then the cleaning should be successful
    And no generations should be removed
    And no error should have occurred

  Scenario: Clean Nix generations with risk assessment
    Given the system has configuration
    And with a "nix_cleanup" Nix profile
    And with dry run mode
    And the Nix system is available
    And I list available Nix generations
    And the system should have at least 1 generations
    When I clean old Nix generations with keep count 1
    Then the cleaning should report risk level assessment
    And the cleaning should be successful
    And no error should have occurred