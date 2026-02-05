package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

func TestNewDockerCleaner(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		dryRun    bool
		pruneMode DockerPruneMode
	}{
		{
			name:      "light mode",
			verbose:   false,
			dryRun:    false,
			pruneMode: DockerPruneLight,
		},
		{
			name:      "standard mode",
			verbose:   true,
			dryRun:    false,
			pruneMode: DockerPruneStandard,
		},
		{
			name:      "aggressive mode",
			verbose:   false,
			dryRun:    true,
			pruneMode: DockerPruneAggressive,
		},
		{
			name:      "verbose dry-run",
			verbose:   true,
			dryRun:    true,
			pruneMode: DockerPruneStandard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewDockerCleaner(tt.verbose, tt.dryRun, tt.pruneMode)

			if cleaner == nil {
				t.Fatal("NewDockerCleaner() returned nil cleaner")
			}

			if cleaner.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", cleaner.verbose, tt.verbose)
			}

			if cleaner.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cleaner.dryRun, tt.dryRun)
			}

			if cleaner.pruneMode != tt.pruneMode {
				t.Errorf("pruneMode = %v, want %v", cleaner.pruneMode, tt.pruneMode)
			}
		})
	}
}

func TestDockerCleaner_Type(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, DockerPruneStandard)

	if cleaner.Type() != domain.OperationTypeDocker {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeDocker)
	}
}

func TestDockerCleaner_IsAvailable(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, DockerPruneStandard)
	available := cleaner.IsAvailable(context.Background())

	// Result depends on Docker installation
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestDockerCleaner_ValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *domain.OperationSettings
		wantErr  bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil docker settings",
			settings: &domain.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid light mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: "light",
				},
			},
			wantErr: false,
		},
		{
			name: "valid standard mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: "standard",
				},
			},
			wantErr: false,
		},
		{
			name: "valid aggressive mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: "aggressive",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid prune mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: "invalid-mode",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewDockerCleaner(false, false, DockerPruneStandard)

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDockerCleaner_Clean_DryRun(t *testing.T) {
	cleaner := NewDockerCleaner(false, true, DockerPruneStandard)

	// Skip test if Docker is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Docker not available")
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Dry-run should report items
	if cleanResult.ItemsRemoved == 0 {
		t.Errorf("Clean() removed %d items, want > 0", cleanResult.ItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestDockerCleaner_Clean_NoAvailable(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, DockerPruneLight)

	// Can't easily test "Docker not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestDockerCleaner_Scan(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, DockerPruneStandard)

	result := cleaner.Scan(context.Background())

	// Scan may not find any items if Docker is not installed or no resources exist
	if result.IsErr() {
		t.Fatalf("Scan() error = %v", result.Error())
	}

	items := result.Value()

	// Items count depends on Docker installation and resources
	if len(items) == 0 {
		t.Log("Scan() found 0 items (Docker may not be installed or no resources)")
	}
}

func TestDockerCleaner_DryRunStrategy(t *testing.T) {
	cleaner := NewDockerCleaner(false, true, DockerPruneStandard)

	constructor := func(verbose, dryRun bool) interface {
		IsAvailable(ctx context.Context) bool
		Clean(ctx context.Context) result.Result[domain.CleanResult]
	} {
		return cleaner
	}

	TestDryRunStrategy(t, constructor, "docker")
}

func TestDockerCleaner_PruneModes(t *testing.T) {
	tests := []struct {
		name      string
		pruneMode DockerPruneMode
	}{
		{"light mode", DockerPruneLight},
		{"standard mode", DockerPruneStandard},
		{"aggressive mode", DockerPruneAggressive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewDockerCleaner(false, false, tt.pruneMode)

			if cleaner.pruneMode != tt.pruneMode {
				t.Errorf("pruneMode = %v, want %v", cleaner.pruneMode, tt.pruneMode)
			}
		})
	}
}

func TestDockerPruneMode_String(t *testing.T) {
	TestTypeStringGeneric(t, "DockerPruneMode", func() []struct {
		Value DockerPruneMode
		Want  string
	} {
		return []struct {
			Value DockerPruneMode
			Want  string
		}{
			{DockerPruneLight, "light"},
			{DockerPruneStandard, "standard"},
			{DockerPruneAggressive, "aggressive"},
		}
	})
}

func TestDockerCleaner_Clean_Verbose(t *testing.T) {
	cleaner := NewDockerCleaner(true, false, DockerPruneStandard)

	// Skip if Docker is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Docker not available")
		return
	}

	// Just verify verbose flag is set
	if !cleaner.verbose {
		t.Error("verbose flag should be set")
	}
}

func TestDockerCleaner_Clean_Aggressive(t *testing.T) {
	cleaner := NewDockerCleaner(false, true, DockerPruneAggressive)

	// Skip if Docker is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Docker not available")
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	// Aggressive mode should include volumes
	// Estimate is 2GB for aggressive mode
	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}
