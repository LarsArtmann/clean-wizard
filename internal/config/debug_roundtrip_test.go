package config

import (
	"fmt"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestDebugKoanfPath(t *testing.T) {
	yamlContent := `
profiles:
  daily:
    operations:
      - name: docker
        settings:
          docker:
            prune_mode: "VOLUMES"
`
	k := loadKoanfFromYAML(t, yamlContent)

	fmt.Printf("raw full=%#v\n", k.Raw())

	raw := k.Get("profiles.daily.operations.0.settings")
	fmt.Printf("raw=%#v\n", raw)

	exists := k.Exists("profiles.daily.operations.0.settings")
	fmt.Printf("exists=%v\n", exists)

	existsOp := k.Exists("profiles.daily.operations.0")
	fmt.Printf("existsOp=%v\n", existsOp)

	if raw != nil {
		out, err := yamlv3.Marshal(raw)
		fmt.Printf("marshal err=%v out=%q\n", err, string(out))

		var s domain.OperationSettings
		err = yamlv3.Unmarshal(out, &s)
		fmt.Printf("unmarshal err=%v docker=%+v\n", err, s.Docker)
	}
}
