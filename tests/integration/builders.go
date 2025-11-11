package integration

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestConfigurationBuilder provides utilities for creating test configurations
type TestConfigurationBuilder struct {
	version      string
	safeMode     bool
	maxDiskUsage int
	protected    []string
	profiles     map[string]*domain.Profile
}

// NewTestConfigurationBuilder creates a new test configuration builder
func NewTestConfigurationBuilder() *TestConfigurationBuilder {
	return &TestConfigurationBuilder{
		version:      "1.0.0",
		safeMode:     true,
		maxDiskUsage: 50,
		protected:    []string{"/System", "/Library", "/usr", "/etc", "/var"},
		profiles:     make(map[string]*domain.Profile),
	}
}

// WithVersion sets the configuration version
func (b *TestConfigurationBuilder) WithVersion(version string) *TestConfigurationBuilder {
	b.version = version
	return b
}

// WithSafeMode sets the safe mode flag
func (b *TestConfigurationBuilder) WithSafeMode(safeMode bool) *TestConfigurationBuilder {
	b.safeMode = safeMode
	return b
}

// WithMaxDiskUsage sets the maximum disk usage percentage
func (b *TestConfigurationBuilder) WithMaxDiskUsage(maxDiskUsage int) *TestConfigurationBuilder {
	b.maxDiskUsage = maxDiskUsage
	return b
}

// WithProtectedPaths sets the protected paths
func (b *TestConfigurationBuilder) WithProtectedPaths(protected ...string) *TestConfigurationBuilder {
	b.protected = protected
	return b
}

// WithProfile adds a profile to the configuration
func (b *TestConfigurationBuilder) WithProfile(name string, profile *domain.Profile) *TestConfigurationBuilder {
	if b.profiles == nil {
		b.profiles = make(map[string]*domain.Profile)
	}
	b.profiles[name] = profile
	return b
}

// WithDailyProfile adds a daily profile with basic settings
func (b *TestConfigurationBuilder) WithDailyProfile(operations ...domain.CleanupOperation) *TestConfigurationBuilder {
	profile := &domain.Profile{
		Name:        "daily",
		Description: "Daily cleanup",
		Operations:  operations,
		Enabled:     true,
	}
	return b.WithProfile("daily", profile)
}

// WithWeeklyProfile adds a weekly profile with basic settings
func (b *TestConfigurationBuilder) WithWeeklyProfile(operations ...domain.CleanupOperation) *TestConfigurationBuilder {
	profile := &domain.Profile{
		Name:        "weekly",
		Description: "Weekly cleanup",
		Operations:  operations,
		Enabled:     true,
	}
	return b.WithProfile("weekly", profile)
}

// Build creates the configuration
func (b *TestConfigurationBuilder) Build() *domain.Config {
	return &domain.Config{
		Version:      b.version,
		SafeMode:     b.safeMode,
		MaxDiskUsage: b.maxDiskUsage,
		Protected:    b.protected,
		Profiles:     b.profiles,
	}
}

// BuildAsFile builds the configuration and saves it to a file
func (b *TestConfigurationBuilder) BuildAsFile(ctx *IntegrationTestContext, filename string) string {
	config := b.Build()
	yamlContent := b.toYAML(config) // Simplified - would use proper YAML marshaling
	
	return ctx.TempFile(filename, yamlContent)
}

// toYAML converts configuration to YAML (simplified version)
func (b *TestConfigurationBuilder) toYAML(config *domain.Config) string {
	yaml := fmt.Sprintf(`version: "%s"
safe_mode: %t
max_disk_usage: %d
protected:
`, config.Version, config.SafeMode, config.MaxDiskUsage)

	for _, path := range config.Protected {
		yaml += fmt.Sprintf(`  - "%s"
`, path)
	}

	yaml += "profiles:\n"
	for name, profile := range config.Profiles {
		yaml += fmt.Sprintf(`  %s:
    name: "%s"
    description: "%s"
    enabled: %t
    operations:
`, name, profile.Name, profile.Description, profile.Enabled)

		for i, op := range profile.Operations {
			yaml += fmt.Sprintf(`      - name: "%s"
        description: "%s"
        risk_level: "%s"
        enabled: %t
`, op.Name, op.Description, op.RiskLevel, op.Enabled)

			if op.Settings != nil {
				yaml += fmt.Sprintf(`        settings:
          type: "%s"
`, op.Settings.Type)

				if op.Settings.NixStore != nil {
					yaml += fmt.Sprintf(`          nix_store:
            keep_generations: %d
            min_age: "%s"
            dry_run: %t
`, op.Settings.NixStore.KeepGenerations, op.Settings.NixStore.MinAge.String(), op.Settings.NixStore.DryRun)
				}
			}

			if i < len(profile.Operations)-1 {
				yaml += "\n"
			}
		}
	}

	return yaml
}

// TestProfileBuilder provides utilities for creating test profiles
type TestProfileBuilder struct {
	name        string
	description string
	operations  []domain.CleanupOperation
	enabled     bool
}

// NewTestProfileBuilder creates a new test profile builder
func NewTestProfileBuilder(name, description string) *TestProfileBuilder {
	return &TestProfileBuilder{
		name:        name,
		description: description,
		operations:  []domain.CleanupOperation{},
		enabled:     true,
	}
}

// WithOperation adds an operation to the profile
func (b *TestProfileBuilder) WithOperation(operation domain.CleanupOperation) *TestProfileBuilder {
	b.operations = append(b.operations, operation)
	return b
}

// WithNixOperation adds a Nix operation to the profile
func (b *TestProfileBuilder) WithNixOperation(name, description string, riskLevel domain.RiskLevel) *TestProfileBuilder {
	operation := domain.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   riskLevel,
		Enabled:     true,
		Settings: &domain.OperationSettings{
			Type: domain.OperationTypeNixStore,
			NixStore: &domain.NixStoreSettings{
				KeepGenerations: 3,
				MinAge:          0, // Would use time.Duration
				DryRun:          true,
			},
		},
	}
	b.operations = append(b.operations, operation)
	return b
}

// WithEnabled sets whether the profile is enabled
func (b *TestProfileBuilder) WithEnabled(enabled bool) *TestProfileBuilder {
	b.enabled = enabled
	return b
}

// Build creates the profile
func (b *TestProfileBuilder) Build() *domain.Profile {
	return &domain.Profile{
		Name:        b.name,
		Description: b.description,
		Operations:  b.operations,
		Enabled:     b.enabled,
	}
}