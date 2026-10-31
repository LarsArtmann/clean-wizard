package config

import (
	"os"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/parsers/yaml"
)

func TestReloadedViolations_Debug(t *testing.T) {
	path := t.TempDir() + "/cw.yaml"
	t.Setenv("CONFIG_PATH", path)

	os.WriteFile(path, nil, 0o600) // create file so Save targets it

	cfg := GetDefaultConfig()
	_ = cfg

	// Build the saved file through the real Save path: temp HOME trick.
	t.Setenv("HOME", t.TempDir())

	saved := readTestSavedYAML(t, path)
	t.Logf("yaml len=%d", len(saved))

	k2 := koanf.New(".")
	if err := k2.Load(file.Provider(path), yaml.Parser()); err != nil {
		t.Fatalf("load yaml: %v", err)
	}

	parsed, err := parseConfig(k2)
	if err != nil {
		t.Fatalf("parseConfig: %v", err)
	}

	v := NewConfigValidator()
	result := v.ValidateConfig(parsed)
	t.Logf("valid=%v errors=%d", result.IsValid, len(result.Errors))
	for _, e := range result.Errors {
		t.Logf("ERROR field=%q msg=%q suggest=%q", e.Field, e.Message, e.Suggestion)
	}
}

func readTestSavedYAML(t *testing.T, path string) string {
	t.Helper()
	// Trigger Save of defaults+extra profile through the public Save API.
	cfg := GetDefaultConfig()
	if err := Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	return string(data)
}
