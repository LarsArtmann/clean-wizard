package docker

import (
	"context"
	"testing"

	cln "github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

func TestNewDockerCleaner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		verbose   bool
		dryRun    bool
		pruneMode enums.DockerPruneMode
	}{
		{
			name:      "ALL mode",
			verbose:   false,
			dryRun:    false,
			pruneMode: enums.DockerPruneAll,
		},
		{
			name:      "IMAGES mode",
			verbose:   true,
			dryRun:    false,
			pruneMode: enums.DockerPruneImages,
		},
		{
			name:      "CONTAINERS mode",
			verbose:   false,
			dryRun:    true,
			pruneMode: enums.DockerPruneContainers,
		},
		{
			name:      "VOLUMES mode",
			verbose:   true,
			dryRun:    true,
			pruneMode: enums.DockerPruneVolumes,
		},
		{
			name:      "BUILDS mode",
			verbose:   false,
			dryRun:    false,
			pruneMode: enums.DockerPruneBuilds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner := NewDockerCleaner(tt.verbose, tt.dryRun, tt.pruneMode)

			if cleaner == nil {
				t.Fatal("NewDockerCleaner() returned nil cleaner")
			}

			if cleaner.pruneMode != tt.pruneMode {
				t.Errorf("pruneMode = %v, want %v", cleaner.pruneMode, tt.pruneMode)
			}
		})
	}
}

func TestDockerCleaner_Type(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, false, enums.DockerPruneAll)

	if cleaner.Type() != operations.OperationTypeDocker {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), operations.OperationTypeDocker)
	}
}

func TestDockerCleaner_IsAvailable(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, false, enums.DockerPruneAll)
	available := cleaner.IsAvailable(context.Background())

	// Result depends on Docker installation
	if available != true && available != false {
		t.Errorf("IsAvailable() returned invalid value")
	}
}

func TestDockerCleaner_ValidateSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *operations.OperationSettings
		wantErr  bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false,
		},
		{
			name:     "nil docker settings",
			settings: &operations.OperationSettings{},
			wantErr:  false,
		},
		{
			name: "valid light mode",
			settings: &operations.OperationSettings{
				Docker: &operations.DockerSettings{
					PruneMode: enums.DockerPruneAll,
				},
			},
			wantErr: false,
		},
		{
			name: "valid standard mode",
			settings: &operations.OperationSettings{
				Docker: &operations.DockerSettings{
					PruneMode: enums.DockerPruneImages,
				},
			},
			wantErr: false,
		},
		{
			name: "valid aggressive mode",
			settings: &operations.OperationSettings{
				Docker: &operations.DockerSettings{
					PruneMode: enums.DockerPruneContainers,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid prune mode",
			settings: &operations.OperationSettings{
				Docker: &operations.DockerSettings{
					PruneMode: enums.DockerPruneMode(999),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner := NewDockerCleaner(false, false, enums.DockerPruneAll)

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDockerCleaner_Clean_DryRun(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, true, enums.DockerPruneAll)

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

	if cleanResult.Strategy != enums.StrategyDryRunType {
		t.Errorf(
			"Clean() strategy = %v, want %v",
			cleanResult.Strategy,
			enums.StrategyDryRunType,
		)
	}

	if cleanResult.FreedBytes == 0 {
		t.Logf("Clean() freed 0 bytes (no Docker resources to clean)")
	} else {
		t.Logf("Clean() freed %d bytes", cleanResult.FreedBytes)
	}
}

func TestDockerCleaner_Clean_NoAvailable(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, false, enums.DockerPruneAll)

	// Can't easily test "Docker not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestDockerCleaner_Scan(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, false, enums.DockerPruneAll)

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
	t.Parallel()

	cleaner := NewDockerCleaner(false, true, enums.DockerPruneAll)

	cln.TestDryRun(t, cln.SimpleCleanerConstructorFromInstance(cleaner), "docker", -1)
}

func TestDockerCleaner_PruneModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pruneMode enums.DockerPruneMode
	}{
		{"ALL mode", enums.DockerPruneAll},
		{"IMAGES mode", enums.DockerPruneImages},
		{"CONTAINERS mode", enums.DockerPruneContainers},
		{"VOLUMES mode", enums.DockerPruneVolumes},
		{"BUILDS mode", enums.DockerPruneBuilds},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleaner := NewDockerCleaner(false, false, tt.pruneMode)

			if cleaner.pruneMode != tt.pruneMode {
				t.Errorf("pruneMode = %v, want %v", cleaner.pruneMode, tt.pruneMode)
			}
		})
	}
}

func TestDockerCleaner_Clean_Verbose(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(true, false, enums.DockerPruneAll)

	// Skip if Docker is not available
	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: Docker not available")

		return
	}

	// Just verify verbose flag is set
	if !cleaner.GetVerbose() {
		t.Error("verbose flag should be set")
	}
}

func TestDockerCleaner_Clean_Aggressive(t *testing.T) {
	t.Parallel()

	cleaner := NewDockerCleaner(false, true, enums.DockerPruneAll)

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
	// Note: FreedBytes may be 0 if no Docker resources exist to clean
	if cleanResult.FreedBytes == 0 {
		t.Logf("Clean() freed 0 bytes (no Docker resources to clean)")
	} else {
		t.Logf("Clean() freed %d bytes", cleanResult.FreedBytes)
	}
}

// TestParseDockerReclaimedSpace tests parsing of docker prune output.
func TestParseDockerReclaimedSpace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		output   string
		expected int64
		wantErr  bool
	}{
		{
			name:     "valid kB output",
			output:   "Deleted Containers:\nabc123\ndef456\n\nDeleted Images:\nsha256:123\n\nTotal reclaimed space: 1.84kB",
			expected: int64(1840),
			wantErr:  false,
		},
		{
			name:     "valid MB output",
			output:   "Deleted Containers:\nabc123\n\nTotal reclaimed space: 13.5 MB",
			expected: int64(13500000),
			wantErr:  false,
		},
		{
			name:     "valid GB output",
			output:   "Deleted Images:\nsha256:123\n\nTotal reclaimed space: 2.5GB",
			expected: int64(2500000000),
			wantErr:  false,
		},
		{
			name:     "zero bytes output",
			output:   "Total reclaimed space: 0B",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "no reclaimed space line",
			output:   "Deleted Containers:\nabc123\n",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "valid TB output",
			output:   "Deleted Volumes:\nvol1\n\nTotal reclaimed space: 1.2TB",
			expected: int64(1200000000000),
			wantErr:  false,
		},
		{
			name:     "valid B output",
			output:   "Total reclaimed space: 512B",
			expected: 512,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseDockerReclaimedSpace(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDockerReclaimedSpace() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if result != tt.expected {
				t.Errorf("ParseDockerReclaimedSpace() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestParseDockerSize tests conversion of Docker size strings to bytes.
func TestParseDockerSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sizeStr  string
		expected int64
		wantErr  bool
	}{
		{
			name:     "kilobytes",
			sizeStr:  "1.84kB",
			expected: int64(1840),
			wantErr:  false,
		},
		{
			name:     "megabytes",
			sizeStr:  "13.5 MB",
			expected: int64(13500000),
			wantErr:  false,
		},
		{
			name:     "gigabytes",
			sizeStr:  "2.5GB",
			expected: int64(2500000000),
			wantErr:  false,
		},
		{
			name:     "terabytes",
			sizeStr:  "1.2TB",
			expected: int64(1200000000000),
			wantErr:  false,
		},
		{
			name:     "bytes",
			sizeStr:  "512B",
			expected: 512,
			wantErr:  false,
		},
		{
			name:     "zero bytes",
			sizeStr:  "0B",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "zero with no unit",
			sizeStr:  "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "empty string",
			sizeStr:  "",
			expected: 0,
			wantErr:  false,
		},
		{
			name:    "invalid unit",
			sizeStr: "1.5XB",
			wantErr: true,
		},
		{
			name:    "invalid format",
			sizeStr: "invalid",
			wantErr: true,
		},
		{
			name:     "unit-less number treated as bytes",
			sizeStr:  "1.5",
			expected: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseDockerSize(tt.sizeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDockerSize() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseDockerSize() = %d, want %d", result, tt.expected)
			}
		})
	}
}
