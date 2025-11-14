package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewGenerateCommand creates configuration generation command
func NewGenerateCommand() *cobra.Command {
	var outputFile string
	var template string

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration templates",
		Long: `Create configuration file templates for different use cases.

Available templates:
- simple    : Basic configuration for everyday use
- working    : Complete configuration with all features
- minimal    : Bare minimum configuration for testing
- advanced   : Production-ready configuration with all options`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				template = args[0]
			}

			return generateConfigTemplate(template, outputFile)
		},
	}

	// Command flags
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "clean-wizard.yaml", "Output configuration file")
	generateCmd.Flags().StringVarP(&template, "template", "t", "working", "Configuration template (simple, working, minimal, advanced)")

	return generateCmd
}

// generateConfigTemplate generates the specified configuration template
func generateConfigTemplate(template string, outputFile string) error {
	var content string

	switch template {
	case "simple":
		content = getSimpleTemplate()
	case "working":
		content = getWorkingTemplate()
	case "minimal":
		content = getMinimalTemplate()
	case "advanced":
		content = getAdvancedTemplate()
	default:
		return fmt.Errorf("unknown template: %s. Available: simple, working, minimal, advanced", template)
	}

	// Write to file
	err := os.WriteFile(outputFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("✅ Configuration template '%s' generated: %s\n", template, outputFile)
	fmt.Printf("💡 Edit the file to customize your cleanup strategy\n")
	fmt.Printf("🏃 Run 'clean-wizard scan --config %s' to test configuration\n", outputFile)

	return nil
}

// getSimpleTemplate returns basic configuration template
func getSimpleTemplate() string {
	return `# Clean-wizard basic configuration
# Generated with: clean-wizard generate simple

version: "1.0.0"
safe_mode: true
max_disk_usage: 50

# Protected system paths (never cleaned)
protected:
  - "/System"
  - "/Library"
  - "/Applications"
  - "/usr"
  - "/etc"
  - "/var"
  - "/bin"
  - "/sbin"

# Cleanup profiles
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup for routine maintenance"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean old Nix generations"
        risk_level: "LOW"
        enabled: true
        settings:
          keep_count: 3
`
}

// getWorkingTemplate returns complete configuration template
func getWorkingTemplate() string {
	return `# Clean-wizard working configuration
# Generated with: clean-wizard generate working

version: "1.0.0"
safe_mode: true
max_disk_usage: 50

# Protected system paths (never cleaned)
protected:
  - "/"
  - "/System"
  - "/Library"
  - "/Applications"
  - "/usr"
  - "/etc"
  - "/var"
  - "/bin"
  - "/sbin"

# Cleanup profiles
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup for routine maintenance"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean old Nix generations"
        risk_level: "LOW"
        enabled: true
        settings:
          keep_count: 3
  
  weekly:
    name: "weekly"
    description: "Weekly cleanup for deeper maintenance"
    enabled: false
    operations:
      - name: "package-caches"
        description: "Clean package manager caches"
        risk_level: "MEDIUM"
        enabled: true
        settings:
          keep_recent: true
`
}

// getMinimalTemplate returns bare minimum configuration
func getMinimalTemplate() string {
	return `# Clean-wizard minimal configuration
# Generated with: clean-wizard generate minimal

version: "1.0.0"
safe_mode: true
max_disk_usage: 50

protected: []

profiles:
  basic:
    name: "basic"
    description: "Basic cleanup"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean Nix generations"
        risk_level: "LOW"
        enabled: true
`
}

// getAdvancedTemplate returns production-ready configuration
func getAdvancedTemplate() string {
	return `# Clean-wizard advanced production configuration
# Generated with: clean-wizard generate advanced

version: "1.0.0"
safe_mode: true
max_disk_usage: 70

# Protected system paths (never cleaned)
protected:
  - "/"
  - "/System"
  - "/Library"
  - "/Applications"
  - "/usr"
  - "/usr/local"
  - "/etc"
  - "/var"
  - "/bin"
  - "/sbin"
  - "/opt"

# User-defined protected paths
user_protected:
  - "/Users/Documents"
  - "/Users/Downloads/Important"
  - "/home/user/projects"

# Cleanup profiles
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup for routine maintenance"
    enabled: true
    operations:
      - name: "nix-generations"
        description: "Clean old Nix generations"
        risk_level: "LOW"
        enabled: true
        settings:
          keep_count: 3
          keep_current: true
  
  weekly:
    name: "weekly"
    description: "Weekly cleanup for deeper maintenance"
    enabled: true
    operations:
      - name: "package-caches"
        description: "Clean package manager caches"
        risk_level: "MEDIUM"
        enabled: true
        settings:
          keep_recent: true
          max_age_days: 30
  
  monthly:
    name: "monthly"
    description: "Monthly cleanup for deep maintenance"
    enabled: false
    operations:
      - name: "temp-files"
        description: "Clean temporary files"
        risk_level: "MEDIUM"
        enabled: true
        settings:
          max_age_days: 7
          exclude_patterns: ["*.log", "*.tmp"]
      
      - name: "logs"
        description: "Clean old log files"
        risk_level: "MEDIUM"
        enabled: true
        settings:
          max_age_days: 14
          keep_critical: true

# Configuration settings
settings:
  dry_run_default: true
  confirm_before_delete: true
  backup_important_files: true
  log_cleanup_operations: true
  max_cleanup_time_minutes: 60
`
}
