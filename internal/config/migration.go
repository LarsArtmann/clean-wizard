package config

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	yamlv3 "gopkg.in/yaml.v3"
)

// Migration transforms a configuration from one format version to the next.
// Migrations form a chain: each step's From must match the previous step's To,
// and the last step's To must be CurrentFormatVersion.
type Migration struct {
	From        FormatVersion
	To          FormatVersion
	Description string
	Apply       func(*types.Config) error
}

// Step renders the migration as "from → to".
func (m Migration) Step() string {
	return m.From.String() + " → " + m.To.String()
}

// migrations is the append-only chain of format migrations. It is a var so
// tests can register synthetic migrations; production code appends new steps
// here when a format change ships.
var migrations = []Migration{} //nolint:gochecknoglobals

// Migrations returns the registered migration chain in registration order.
func Migrations() []Migration {
	return slices.Clone(migrations)
}

// PlanMigration returns the migration steps that bring a configuration from
// the given version up to CurrentFormatVersion, or an error when no chain
// covers that path.
func PlanMigration(from FormatVersion) ([]Migration, error) {
	plan := []Migration{}

	for from.Compare(CurrentFormatVersion) < 0 {
		step, found := findMigration(from)
		if !found {
			return nil, errorfamily.NewRejection(
				"config.migrate",
				"no migration path from configuration format "+from.String()+" to "+CurrentFormatVersion.String(),
			)
		}

		plan = append(plan, step)
		from = step.To
	}

	return plan, nil
}

func findMigration(from FormatVersion) (Migration, bool) {
	for _, migration := range migrations {
		if migration.From == from && migration.To.Compare(from) > 0 {
			return migration, true
		}
	}

	return Migration{}, false //nolint:exhaustruct
}

// MigrationRecord captures what one migration step changed.
type MigrationRecord struct {
	From        FormatVersion
	To          FormatVersion
	Description string
	Changes     []autoconfigure.Change
}

// Step renders the record's migration as "from → to".
func (r MigrationRecord) Step() string {
	return r.From.String() + " → " + r.To.String()
}

// ApplyMigrations applies the plan to target in order, records the changes
// each step made, and stamps the current format version. The target is left
// in an intermediate state when a step fails.
func ApplyMigrations(target *types.Config, plan []Migration) ([]MigrationRecord, error) {
	records := make([]MigrationRecord, 0, len(plan))

	for _, migration := range plan {
		before := flattenConfig(target)

		if err := migration.Apply(target); err != nil {
			return nil, errorfamily.WrapCorruption(err, "config.migrate", "migration "+migration.Step()+" failed")
		}

		records = append(records, MigrationRecord{
			From:        migration.From,
			To:          migration.To,
			Description: migration.Description,
			Changes:     autoconfigure.DiffMaps(before, flattenConfig(target), ""),
		})
	}

	target.Version = CurrentFormatVersion.String()

	return records, nil
}

// Flatten field keys shared with the YAML writer; kept as constants so the
// migration diff keys can never drift from the on-disk field names.
const (
	versionKey   = "version"
	safeModeKey  = "safe_mode"
	profilesRoot = "profiles"
)

// flattenConfig renders the configuration as a flat string map so the SDK's
// diff engine can report what a migration changed. Keys are stable dotted
// paths; map iteration order is normalized by sorting profile names.
func flattenConfig(config *types.Config) map[string]string {
	flat := map[string]string{
		versionKey:               config.Version,
		safeModeKey:              config.SafeMode.String(),
		"max_disk_usage_percent": strconv.Itoa(config.MaxDiskUsage),
		"current_profile":        config.CurrentProfile,
		protectedField:           strings.Join(config.Protected, "|"),
		"last_clean":             config.LastClean.String(),
		"updated":                config.Updated.String(),
	}

	profileNames := make([]string, 0, len(config.Profiles))
	for name := range config.Profiles {
		profileNames = append(profileNames, name)
	}

	sort.Strings(profileNames)

	for _, name := range profileNames {
		profile := config.Profiles[name]
		prefix := profilesRoot + "." + name

		flat[prefix+".description"] = profile.Description
		flat[prefix+".enabled"] = profile.Enabled.String()

		for index, operation := range profile.Operations {
			operationPrefix := fmt.Sprintf("%s.operations.%d", prefix, index)

			flat[operationPrefix+".name"] = operation.Name
			flat[operationPrefix+".description"] = operation.Description
			flat[operationPrefix+".risk_level"] = operation.RiskLevel.String()
			flat[operationPrefix+".enabled"] = operation.Enabled.String()

			if operation.Settings != nil {
				settingsYAML, err := yamlv3.Marshal(
					operation.Settings,
				) //nolint:musttag // OperationSettings carries yaml tags
				if err == nil {
					flat[operationPrefix+".settings"] = string(settingsYAML)
				}
			}
		}
	}

	return flat
}
