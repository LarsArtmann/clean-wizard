package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Remove any existing config file
	configPath := filepath.Join(os.Getenv("HOME"), ".clean-wizard.yaml")
	os.Remove(configPath)

	// Load default config
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values
	assert.Equal(t, "dev", cfg.Version)
	assert.True(t, cfg.SafeMode)
	assert.True(t, cfg.DryRun)
	assert.False(t, cfg.Verbose)
	assert.True(t, cfg.Backup)
	assert.Equal(t, 90, cfg.MaxDiskUsage)
	assert.NotEmpty(t, cfg.Protected)
	assert.Contains(t, cfg.Protected, "/nix/store")
	assert.Contains(t, cfg.Protected, "/Users")
	assert.NotEmpty(t, cfg.Profiles)
	assert.Contains(t, cfg.Profiles, "daily")
	assert.Contains(t, cfg.Profiles, "comprehensive")
	assert.Contains(t, cfg.Profiles, "aggressive")
}

func TestLoad_ExistingConfig(t *testing.T) {
	// Create a test config
	testConfig := &types.Config{
		Version:      "test",
		SafeMode:     false,
		DryRun:       false,
		Verbose:      true,
		Backup:       false,
		MaxDiskUsage: 80,
		Protected:    []string{"/test"},
		Profiles: map[string]types.Profile{
			"test": {
				Name:        "test",
				Description: "test profile",
				Operations: []types.CleanupOperation{
					{
						Name:        "test-op",
						Description: "test operation",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Save the test config
	err := Save(testConfig)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify loaded values
	assert.Equal(t, "test", cfg.Version)
	assert.False(t, cfg.SafeMode)
	assert.False(t, cfg.DryRun)
	assert.True(t, cfg.Verbose)
	assert.False(t, cfg.Backup)
	assert.Equal(t, 80, cfg.MaxDiskUsage)
	assert.Equal(t, []string{"/test"}, cfg.Protected)
	assert.Contains(t, cfg.Profiles, "test")
	assert.Equal(t, "test profile", cfg.Profiles["test"].Description)

	// Clean up
	os.Remove(GetConfigPath())
}

func TestSave(t *testing.T) {
	// Create a test config
	testConfig := &types.Config{
		Version:      "test-save",
		SafeMode:     true,
		DryRun:       true,
		Verbose:      false,
		Backup:       true,
		MaxDiskUsage: 95,
		Protected:    []string{"/save-test"},
		Profiles: map[string]types.Profile{
			"save-test": {
				Name:        "save-test",
				Description: "save test profile",
				Operations: []types.CleanupOperation{
					{
						Name:        "save-op",
						Description: "save operation",
						RiskLevel:   types.RiskMedium,
						Enabled:     true,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Save the config
	err := Save(testConfig)
	require.NoError(t, err)

	// Verify the file exists
	configPath := GetConfigPath()
	assert.FileExists(t, configPath)

	// Load and verify
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "test-save", cfg.Version)
	assert.Equal(t, 95, cfg.MaxDiskUsage)
	assert.Equal(t, []string{"/save-test"}, cfg.Protected)
	assert.Contains(t, cfg.Profiles, "save-test")

	// Clean up
	os.Remove(configPath)
}

func TestGetConfigPath(t *testing.T) {
	expectedPath := filepath.Join(os.Getenv("HOME"), ".clean-wizard.yaml")
	actualPath := GetConfigPath()
	assert.Equal(t, expectedPath, actualPath)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &types.Config{
				Version:      "test",
				SafeMode:     true,
				DryRun:       true,
				Verbose:      false,
				Backup:       true,
				MaxDiskUsage: 50,
				Protected:    []string{"/test"},
				Profiles: map[string]types.Profile{
					"test": {
						Name:        "test",
						Description: "test profile",
						Operations: []types.CleanupOperation{
							{
								Name:        "test-op",
								Description: "test operation",
								RiskLevel:   types.RiskLow,
								Enabled:     true,
							},
						},
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid max disk usage - too low",
			config: &types.Config{
				MaxDiskUsage: -1,
				Protected:    []string{"/test"},
				Profiles:     map[string]types.Profile{},
			},
			wantErr: true,
			errMsg:  "max_disk_usage_percent must be between 0 and 100",
		},
		{
			name: "invalid max disk usage - too high",
			config: &types.Config{
				MaxDiskUsage: 101,
				Protected:    []string{"/test"},
				Profiles:     map[string]types.Profile{},
			},
			wantErr: true,
			errMsg:  "max_disk_usage_percent must be between 0 and 100",
		},
		{
			name: "empty protected paths",
			config: &types.Config{
				MaxDiskUsage: 50,
				Protected:    []string{},
				Profiles:     map[string]types.Profile{},
			},
			wantErr: true,
			errMsg:  "at least one protected path must be specified",
		},
		{
			name: "empty profiles",
			config: &types.Config{
				MaxDiskUsage: 50,
				Protected:    []string{"/test"},
				Profiles:     map[string]types.Profile{},
			},
			wantErr: true,
			errMsg:  "at least one profile must be specified",
		},
		{
			name: "profile without name",
			config: &types.Config{
				MaxDiskUsage: 50,
				Protected:    []string{"/test"},
				Profiles: map[string]types.Profile{
					"": {
						Name:        "",
						Description: "test profile",
						Operations:  []types.CleanupOperation{},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
			},
			wantErr: true,
			errMsg:  "profile  must have a name",
		},
		{
			name: "profile without description",
			config: &types.Config{
				MaxDiskUsage: 50,
				Protected:    []string{"/test"},
				Profiles: map[string]types.Profile{
					"test": {
						Name:        "test",
						Description: "",
						Operations:  []types.CleanupOperation{},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
			},
			wantErr: true,
			errMsg:  "profile test must have a description",
		},
		{
			name: "profile without operations",
			config: &types.Config{
				MaxDiskUsage: 50,
				Protected:    []string{"/test"},
				Profiles: map[string]types.Profile{
					"test": {
						Name:        "test",
						Description: "test profile",
						Operations:  []types.CleanupOperation{},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
			},
			wantErr: true,
			errMsg:  "profile test must have at least one operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsProtectedPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		protected []string
		want      bool
	}{
		{
			name:      "exact match",
			path:      "/Users/test",
			protected: []string{"/Users/test"},
			want:      true,
		},
		{
			name:      "prefix match",
			path:      "/Users/test/file.txt",
			protected: []string{"/Users"},
			want:      true,
		},
		{
			name:      "no match",
			path:      "/tmp/test",
			protected: []string{"/Users", "/System"},
			want:      false,
		},
		{
			name:      "empty path",
			path:      "",
			protected: []string{"/Users"},
			want:      false,
		},
		{
			name:      "empty protected",
			path:      "/Users/test",
			protected: []string{},
			want:      false,
		},
		{
			name:      "subdirectory match",
			path:      "/System/Library/test",
			protected: []string{"/System"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsProtectedPath(tt.path, tt.protected)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := getDefaultConfig()
	require.NotNil(t, cfg)

	// Verify default values
	assert.Equal(t, "dev", cfg.Version)
	assert.True(t, cfg.SafeMode)
	assert.True(t, cfg.DryRun)
	assert.False(t, cfg.Verbose)
	assert.True(t, cfg.Backup)
	assert.Equal(t, 90, cfg.MaxDiskUsage)
	assert.NotEmpty(t, cfg.Protected)
	assert.NotEmpty(t, cfg.Profiles)

	// Verify profiles exist
	assert.Contains(t, cfg.Profiles, "daily")
	assert.Contains(t, cfg.Profiles, "comprehensive")
	assert.Contains(t, cfg.Profiles, "aggressive")

	// Verify daily profile
	daily := cfg.Profiles["daily"]
	assert.Equal(t, "daily", daily.Name)
	assert.Equal(t, "Quick daily cleanup", daily.Description)
	assert.Len(t, daily.Operations, 2)
	assert.True(t, daily.Operations[0].Enabled)
	assert.Equal(t, types.RiskLow, daily.Operations[0].RiskLevel)

	// Verify comprehensive profile
	comprehensive := cfg.Profiles["comprehensive"]
	assert.Equal(t, "comprehensive", comprehensive.Name)
	assert.Equal(t, "Complete system cleanup", comprehensive.Description)
	assert.Len(t, comprehensive.Operations, 3)

	// Verify aggressive profile
	aggressive := cfg.Profiles["aggressive"]
	assert.Equal(t, "aggressive", aggressive.Name)
	assert.Equal(t, "Nuclear option - everything", aggressive.Description)
	assert.Len(t, aggressive.Operations, 3)
	assert.Equal(t, types.RiskHigh, aggressive.Operations[0].RiskLevel)
}
