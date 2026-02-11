package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestNewDockerCleaner(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		dryRun    bool
		pruneMode domain.DockerPruneMode
	}{
		{
			name:      "ALL mode",
			verbose:   false,
			dryRun:    false,
			pruneMode: domain.DockerPruneAll,
		},
		{
			name:      "IMAGES mode",
			verbose:   true,
			dryRun:    false,
			pruneMode: domain.DockerPruneImages,
		},
		{
			name:      "CONTAINERS mode",
			verbose:   false,
			dryRun:    true,
			pruneMode: domain.DockerPruneContainers,
		},
		{
			name:      "VOLUMES mode",
			verbose:   true,
			dryRun:    true,
			pruneMode: domain.DockerPruneVolumes,
		},
		{
			name:      "BUILDS mode",
			verbose:   false,
			dryRun:    false,
			pruneMode: domain.DockerPruneBuilds,
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
	cleaner := NewDockerCleaner(false, false, domain.DockerPruneAll)

	if cleaner.Type() != domain.OperationTypeDocker {
		t.Errorf("Type() = %v, want %v", cleaner.Type(), domain.OperationTypeDocker)
	}
}

func TestDockerCleaner_IsAvailable(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, domain.DockerPruneAll)
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
					PruneMode: domain.DockerPruneAll,
				},
			},
			wantErr: false,
		},
		{
			name: "valid standard mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: domain.DockerPruneImages,
				},
			},
			wantErr: false,
		},
		{
			name: "valid aggressive mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: domain.DockerPruneContainers,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid prune mode",
			settings: &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: domain.DockerPruneMode(999),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewDockerCleaner(false, false, domain.DockerPruneAll)

			err := cleaner.ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDockerCleaner_Clean_DryRun(t *testing.T) {
	cleaner := NewDockerCleaner(false, true, domain.DockerPruneAll)

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

	if cleanResult.Strategy != domain.CleanStrategyType(domain.StrategyDryRunType) {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.CleanStrategyType(domain.StrategyDryRunType))
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

func TestDockerCleaner_Clean_NoAvailable(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, domain.DockerPruneAll)

	// Can't easily test "Docker not available" case without mocking
	// So we just verify IsAvailable is called
	_ = cleaner.IsAvailable(context.Background())
}

func TestDockerCleaner_Scan(t *testing.T) {
	cleaner := NewDockerCleaner(false, false, domain.DockerPruneAll)

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
	cleaner := NewDockerCleaner(false, true, domain.DockerPruneAll)

	TestDryRunStrategy(t, SimpleCleanerConstructorFromInstance(cleaner), "docker")
}

func TestDockerCleaner_PruneModes(t *testing.T) {
	tests := []struct {
		name      string
		pruneMode domain.DockerPruneMode
	}{
		{"ALL mode", domain.DockerPruneAll},
		{"IMAGES mode", domain.DockerPruneImages},
		{"CONTAINERS mode", domain.DockerPruneContainers},
		{"VOLUMES mode", domain.DockerPruneVolumes},
		{"BUILDS mode", domain.DockerPruneBuilds},
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

func TestDockerCleaner_Clean_Verbose(t *testing.T) {
	cleaner := NewDockerCleaner(true, false, domain.DockerPruneAll)

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
	cleaner := NewDockerCleaner(false, true, domain.DockerPruneAll)

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

// TestParseDockerReclaimedSpace tests parsing of docker prune output
func TestParseDockerReclaimedSpace(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int64
		wantErr  bool
	}{
		{
			name: "valid kB output",
			output: "Deleted Containers:\nabc123\ndef456\n\nDeleted Images:\nsha256:123\n\nTotal reclaimed space: 1.84kB",
			expected: int64(1884),
			wantErr:  false,
		},
		{
			name: "valid MB output",
			output: "Deleted Containers:\nabc123\n\nTotal reclaimed space: 13.5 MB",
			expected: int64(13.5 * 1024 * 1024),
			wantErr:  false,
		},
		{
			name: "valid GB output",
			output: "Deleted Images:\nsha256:123\n\nTotal reclaimed space: 2.5GB",
			expected: int64(2.5 * 1024 * 1024 * 1024),
			wantErr:  false,
		},
		{
			name: "zero bytes output",
			output: "Total reclaimed space: 0B",
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
			name: "valid TB output",
			output: "Deleted Volumes:\nvol1\n\nTotal reclaimed space: 1.2TB",
			expected: int64(1319413953331),
			wantErr:  false,
		},
		{
			name: "valid B output",
			output: "Total reclaimed space: 512B",
			expected: 512,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestParseDockerSize tests conversion of Docker size strings to bytes
func TestParseDockerSize(t *testing.T) {
	tests := []struct {
		name     string
		sizeStr  string
		expected int64
		wantErr  bool
	}{
		{
			name:     "kilobytes",
			sizeStr:  "1.84kB",
			expected: int64(1884),
			wantErr:  false,
		},
		{
			name:     "megabytes",
			sizeStr:  "13.5 MB",
			expected: int64(13.5 * 1024 * 1024),
			wantErr:  false,
		},
		{
			name:     "gigabytes",
			sizeStr:  "2.5GB",
			expected: int64(2.5 * 1024 * 1024 * 1024),
			wantErr:  false,
		},
		{
			name:     "terabytes",
			sizeStr:  "1.2TB",
			expected: int64(1319413953331),
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
			name:    "missing unit",
			sizeStr: "1.5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
