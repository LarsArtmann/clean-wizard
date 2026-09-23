package domain

import (
	"testing"
)

func TestConfig_SettingsForProfile(t *testing.T) {
	t.Parallel()

	tempSettings := &TempFilesSettings{OlderThan: "14d"}
	dockerSettings := &DockerSettings{PruneMode: DockerPruneVolumes}

	tests := []struct {
		name        string
		config      *Config
		profileName string
		want        *OperationSettings
	}{
		{
			name:        "nil config returns nil",
			config:      nil,
			profileName: "daily",
			want:        nil,
		},
		{
			name: "unknown profile returns nil",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files", Settings: &OperationSettings{TempFiles: tempSettings}},
						},
					},
				},
			},
			profileName: "weekly",
			want:        nil,
		},
		{
			name: "profile without settings returns nil",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files"}, //nolint:exhaustruct
						},
					},
				},
			},
			profileName: "daily",
			want:        nil,
		},
		{
			name: "single operation settings are returned",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files", Settings: &OperationSettings{TempFiles: tempSettings}},
						},
					},
				},
			},
			profileName: "daily",
			want:        &OperationSettings{TempFiles: tempSettings},
		},
		{
			name: "settings from multiple operations are merged",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files", Settings: &OperationSettings{TempFiles: tempSettings}},
							{Name: "docker", Settings: &OperationSettings{Docker: dockerSettings}},
						},
					},
				},
			},
			profileName: "daily",
			want:        &OperationSettings{TempFiles: tempSettings, Docker: dockerSettings},
		},
		{
			name: "first settings section wins on conflict",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files", Settings: &OperationSettings{TempFiles: tempSettings}},
							{
								Name:     "temp-files",
								Settings: &OperationSettings{TempFiles: &TempFilesSettings{OlderThan: "1h"}},
							},
						},
					},
				},
			},
			profileName: "daily",
			want:        &OperationSettings{TempFiles: tempSettings},
		},
		{
			name: "operations without settings do not erase merged sections",
			config: &Config{ //nolint:exhaustruct
				Profiles: map[string]*Profile{
					"daily": { //nolint:exhaustruct
						Operations: []CleanupOperation{ //nolint:exhaustruct
							{Name: "temp-files"}, //nolint:exhaustruct
							{Name: "docker", Settings: &OperationSettings{Docker: dockerSettings}},
						},
					},
				},
			},
			profileName: "daily",
			want:        &OperationSettings{Docker: dockerSettings},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.config.SettingsForProfile(tt.profileName)
			if got == nil && tt.want == nil {
				return
			}

			if got == nil || tt.want == nil {
				t.Fatalf("SettingsForProfile() = %v, want %v", got, tt.want)
			}

			if got.TempFiles != tt.want.TempFiles {
				t.Errorf("TempFiles = %v, want %v", got.TempFiles, tt.want.TempFiles)
			}

			if got.Docker != tt.want.Docker {
				t.Errorf("Docker = %v, want %v", got.Docker, tt.want.Docker)
			}
		})
	}
}
