package config

import (
	"testing"

	"github.com/knadh/koanf/parsers/yaml"

	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

func TestConfigYAMLMap_Debug(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Profiles["test-profile"] = &types.Profile{
		Name:        "test-profile",
		Description: "desc",
	}

	m := configYAMLMap(cfg)
	t.Logf("map keys: %v", func() []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}())

	profiles, ok := m["profiles"].(map[string]any)
	t.Logf("profiles ok=%v count=%d", ok, len(profiles))

	data, err := yaml.Parser().Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	t.Logf("YAML:\n%s", string(data))
}
