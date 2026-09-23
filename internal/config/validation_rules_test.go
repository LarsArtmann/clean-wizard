package config

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	businessrules "github.com/LarsArtmann/go-business-rules/v2"
)

func TestSeverityCriticalConstant(t *testing.T) {
	t.Parallel()

	if SeverityCritical != "critical" {
		t.Errorf("expected SeverityCritical to be %q, got %q", "critical", string(SeverityCritical))
	}
}

func TestMapViolations_SeverityBucketing(t *testing.T) {
	t.Parallel()

	set := newConfigRuleSet()
	set.add(configRule("critical_field", "structure", "required",
		businessrules.SeverityCritical, "critical rule",
		func() error { return errStatic("critical failure") },
	), "v", nil)
	set.add(configRule("error_field", "field", "range",
		businessrules.SeverityError, "error rule",
		func() error { return errStatic("error failure") },
	), "v", nil)
	set.add(configRule("warning_field", "cross_field", "max_count",
		businessrules.SeverityWarning, "warning rule",
		func() error { return errStatic("warning failure") },
	), "v", nil)
	set.add(configRule("info_field", "security", "security",
		businessrules.SeverityInfo, "info rule",
		func() error { return errStatic("info failure") },
	), "v", nil)

	outcome := businessrules.NewValidator().AddRules(set.rules...).Build()
	result := mapViolations(set, outcome)

	if result.IsValid {
		t.Fatal("expected invalid result when error-severity rules fail")
	}

	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors (critical+error), got %d: %v", len(result.Errors), result.Errors)
	}

	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings (warning+info), got %d: %v", len(result.Warnings), result.Warnings)
	}

	severityByField := map[string]operations.ValidationSeverity{}
	for _, validationError := range result.Errors {
		severityByField[validationError.Field] = validationError.Severity
	}

	if severityByField["critical_field"] != operations.SeverityCritical {
		t.Errorf("expected critical_field to carry critical severity, got %q",
			severityByField["critical_field"])
	}

	if severityByField["error_field"] != operations.SeverityError {
		t.Errorf("expected error_field to carry error severity, got %q",
			severityByField["error_field"])
	}
}

func TestValidateConfig_CriticalSecurityEscalation(t *testing.T) {
	t.Parallel()

	cfg := CreateTestConfig(WithProtectedPaths([]string{"/System", "../escape"}))
	result := NewConfigValidator().ValidateConfig(cfg)

	if result.IsValid {
		t.Fatal("expected invalid config for parent directory reference")
	}

	found := false

	for _, validationError := range result.Errors {
		if validationError.Field == "protected" &&
			validationError.Severity == SeverityCritical &&
			validationError.Rule == "security" {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("expected critical security error for '..' path, got: %+v", result.Errors)
	}
}

func TestValidateConfig_BridgePreservesStructuredMetadata(t *testing.T) {
	t.Parallel()

	cfg := CreateTestConfig(WithMaxDiskUsage(150))
	result := NewConfigValidator().ValidateConfig(cfg)

	if result.IsValid {
		t.Fatal("expected invalid config for max_disk_usage=150")
	}

	var found *ValidationError

	for i := range result.Errors {
		if result.Errors[i].Field == "max_disk_usage" {
			found = &result.Errors[i]

			break
		}
	}

	if found == nil {
		t.Fatalf("expected max_disk_usage error, got: %+v", result.Errors)
	}

	if found.Value != 150 {
		t.Errorf("expected value 150, got %v", found.Value)
	}

	if found.Rule != "range" {
		t.Errorf("expected rule kind %q, got %q", "range", found.Rule)
	}

	if found.Suggestion != "Max disk usage must be between 10% and 95%" {
		t.Errorf("expected configured suggestion, got %q", found.Suggestion)
	}

	if found.Context == nil {
		t.Fatal("expected validation context with min/max bounds")
	}

	if found.Context.MinValue != 10 || found.Context.MaxValue != 95 {
		t.Errorf("expected bounds [10, 95], got %v..%v",
			found.Context.MinValue, found.Context.MaxValue)
	}
}

func TestValidateConfig_WarningsAreBridged(t *testing.T) {
	t.Parallel()

	cfg := CreateTestConfig(WithProtectedPaths([]string{"/System", "/System", "/Library"}))
	result := NewConfigValidator().ValidateConfig(cfg)

	if !result.IsValid {
		t.Fatalf("duplicate paths must warn, not fail: %+v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Fatal("expected duplicate-path warning")
	}

	found := false

	for _, warning := range result.Warnings {
		if warning.Field == "protected" &&
			warning.Suggestion == "Remove duplicate paths from protected list" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected duplicate-path warning with suggestion, got: %+v", result.Warnings)
	}
}

func TestValidateConfig_DeterministicViolationOrder(t *testing.T) {
	t.Parallel()

	cfg := CreateTestConfig(WithVersion(""), WithMaxDiskUsage(150), WithEmptyProfiles())
	validator := NewConfigValidator()

	fingerprint := func(result *ValidationResult) string {
		out := ""

		var outSb181 strings.Builder
		for _, validationError := range result.Errors {
			outSb181.WriteString(validationError.Field + "|" + validationError.Message + "\n")
		}
		out += outSb181.String()

		return out
	}

	first := fingerprint(validator.ValidateConfig(cfg))
	second := fingerprint(validator.ValidateConfig(cfg))

	if first == "" {
		t.Fatal("expected violations for broken config")
	}

	if first != second {
		t.Errorf("violation order not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestValidateConfig_StructuredMetadataForContexts(t *testing.T) {
	t.Parallel()

	cfg := CreateTestConfig()
	cfg.SafeMode = enums.SafeModeDisabled
	cfg.Profiles["daily"].Operations[0].RiskLevel = enums.RiskLevelCriticalType

	validator := NewConfigValidator()
	result := validator.ValidateConfig(cfg)

	for _, warning := range result.Warnings {
		if warning.Field == "safe_mode" && warning.Context != nil {
			if warning.Context.Metadata["safe_mode"] == "" {
				t.Errorf("expected safe_mode metadata, got %v", warning.Context.Metadata)
			}

			return
		}
	}

	t.Fatalf("expected safe_mode cross-field warning with context, got: %+v", result.Warnings)
}

// errStatic is a fixed failure used by synthetic bridge-test rules.
type errStatic string

func (e errStatic) Error() string { return string(e) }
