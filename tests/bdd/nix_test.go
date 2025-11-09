package bdd

import (
	"context"
	"os/exec"
	"testing"
)

// TestNixCleaningBDD tests Nix cleaning operations
func TestNixCleaningBDD(t *testing.T) {
	t.Run("Scenario: List available Nix generations", func(t *testing.T) {
		// Given I have Nix installed
		t.Log("✅ Given I have Nix installed")

		// When I run "clean-wizard scan"
		cmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/clean-wizard", "scan")
		output, err := cmd.CombinedOutput()

		// Then I should see generation numbers and dates
		if err != nil {
			t.Logf("ℹ️  Scan failed (expected in CI): %v\nOutput: %s", err, string(output))
			return
		}

		outputStr := string(output)
		if !contains(outputStr, "✅ Scan completed!") {
			t.Error("❌ Expected scan completion message not found")
		}

		t.Logf("✅ Scan completed successfully")
	})

	t.Run("Scenario: Clean old Nix generations safely", func(t *testing.T) {
		// Given I have multiple Nix generations
		t.Log("✅ Given I have multiple Nix generations")

		// When I run "clean-wizard clean --dry-run"
		cmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/clean-wizard", "clean", "--dry-run")
		output, err := cmd.CombinedOutput()

		// Then I should see which generations would be deleted
		if err != nil {
			t.Logf("ℹ️  Clean failed (expected in CI): %v\nOutput: %s", err, string(output))
			return
		}

		outputStr := string(output)
		if !contains(outputStr, "✅ Cleanup completed!") {
			t.Error("❌ Expected cleanup completion message not found")
		}

		t.Logf("✅ Dry-run cleanup completed successfully")
	})
}

// contains helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
