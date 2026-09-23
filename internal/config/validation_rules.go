package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	businessrules "github.com/LarsArtmann/go-business-rules/v2"
)

// Validation level tags classify every rule by the pipeline stage it belongs
// to. They surface on RuleEvaluated events and feed ValidationError.Rule.
const (
	levelStructure   = "structure"
	levelField       = "field"
	levelCrossField  = "cross_field"
	levelBusiness    = "business_logic"
	levelSecurityTag = "security"
)

// configRuleSet pairs the declarative rule list with per-rule structured
// metadata. businessrules violations only carry the check error as text, so
// the maps restore the typed data (field value, validation context) when the
// outcome is bridged back into ValidationResult.
type configRuleSet struct {
	rules    []businessrules.Rule
	values   map[string]any
	contexts map[string]*ValidationContext
}

func newConfigRuleSet() *configRuleSet {
	return &configRuleSet{
		rules:    []businessrules.Rule{},
		values:   map[string]any{},
		contexts: map[string]*ValidationContext{},
	}
}

// add registers a rule; value and context are optional structured metadata
// attached under the rule name.
func (set *configRuleSet) add(
	rule businessrules.Rule, value any, context *ValidationContext,
) {
	set.rules = append(set.rules, rule)
	set.values[rule.Name()] = value

	if context != nil {
		set.contexts[rule.Name()] = context
	}
}

// configRule builds a named rule tagged with its validation level and
// semantic rule kind (required, range, format, ...).
func configRule( //nolint:ireturn // rules are consumed through the Rule interface
	field, level, ruleKind string,
	severity businessrules.Severity,
	message string,
	check func() error,
) businessrules.Rule {
	return businessrules.NewRule(field, check, severity, message).
		WithTags(level, ruleKind)
}

// withSuggestion attaches the fix hint as rule description metadata; the
// bridge surfaces it as ValidationError.Suggestion.
func withSuggestion( //nolint:ireturn // rules are consumed through the Rule interface
	rule businessrules.Rule, suggestion string,
) businessrules.Rule {
	if impl, ok := rule.(businessrules.RuleImpl); ok {
		return impl.WithDescription(suggestion)
	}

	return rule
}

// mapViolations bridges a businessrules outcome into the project's
// ValidationResult: Error/Critical violations become blocking errors,
// Warning/Info violations become non-blocking warnings.
func mapViolations(
	set *configRuleSet, outcome businessrules.ValidationResultError,
) *ValidationResult {
	result := &ValidationResult{ //nolint:exhaustruct
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Sanitized: nil,
	}

	outcome.ForEach(func(violation businessrules.ViolationError) {
		severity := validationSeverityFromBusinessRules(violation.Rule.Severity())

		message := violation.Context
		if message == "" {
			message = violation.Rule.Message()
		}

		if severity == operations.SeverityError || severity == operations.SeverityCritical {
			result.Errors = append(result.Errors, ValidationError{
				Field:      violation.Rule.Name(),
				Rule:       ruleKindOf(violation.Rule),
				Value:      set.values[violation.Rule.Name()],
				Message:    message,
				Severity:   severity,
				Suggestion: suggestionOf(violation.Rule),
				Context:    set.contexts[violation.Rule.Name()],
			})

			return
		}

		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      violation.Rule.Name(),
			Message:    message,
			Suggestion: suggestionOf(violation.Rule),
			Context:    set.contexts[violation.Rule.Name()],
		})
	})

	result.IsValid = len(result.Errors) == 0

	return result
}

// validationSeverityFromBusinessRules maps the library's 4-level severity
// onto the domain severity enum.
func validationSeverityFromBusinessRules(
	severity businessrules.Severity,
) operations.ValidationSeverity {
	switch severity {
	case businessrules.SeverityCritical:
		return operations.SeverityCritical
	case businessrules.SeverityWarning:
		return operations.SeverityWarning
	case businessrules.SeverityInfo:
		return operations.SeverityInfo
	case businessrules.SeverityError:
		return operations.SeverityError
	default:
		return operations.SeverityError
	}
}

func ruleKindOf(rule businessrules.Rule) string {
	if tagged, ok := rule.(interface{ Tags() []string }); ok {
		if tags := tagged.Tags(); len(tags) > 1 {
			return tags[1]
		}
	}

	return ""
}

func suggestionOf(rule businessrules.Rule) string {
	if describable, ok := rule.(interface{ Description() string }); ok {
		return describable.Description()
	}

	return ""
}

// buildConfigRules compiles every validation level (structure, field,
// cross-field, business logic, security) into one declarative rule set.
// Rules run in registration order for deterministic violation output.
func (cv *ConfigValidator) buildConfigRules(cfg *types.Config) *configRuleSet {
	set := newConfigRuleSet()
	cv.addStructureRules(set, cfg)
	cv.addFieldRules(set, cfg)
	cv.addCrossFieldRules(set, cfg)
	cv.addBusinessLogicRules(set, cfg)
	cv.addSecurityRules(set, cfg)

	return set
}

func (cv *ConfigValidator) addStructureRules(set *configRuleSet, cfg *types.Config) {
	set.add(withSuggestion(
		configRule("version", levelStructure, "required", businessrules.SeverityError,
			"Configuration version is required",
			requiredCheck(cfg.Version != "", errors.New("Configuration version is required"))),
		"Add version field with semantic version (e.g., '1.0.0')",
	), cfg.Version, nil)

	set.add(withSuggestion(
		configRule("profiles", levelStructure, "required", businessrules.SeverityError,
			"At least one profile is required",
			requiredCheck(len(cfg.Profiles) > 0, errors.New("At least one profile is required"))),
		"Add a profile with at least one operation",
	), cfg.Profiles, nil)

	set.add(withSuggestion(
		configRule("protected", levelStructure, "required", businessrules.SeverityError,
			"Protected paths cannot be empty",
			requiredCheck(len(cfg.Protected) > 0, errors.New("Protected paths cannot be empty"))),
		"Add system paths like "+strings.Join(types.DefaultProtectedPaths(), ", "),
	), cfg.Protected, nil)

	if minPaths := cv.protectedPathsMinimum(); minPaths > 0 {
		set.add(withSuggestion(
			configRule("protected", levelStructure, "min_items", businessrules.SeverityError,
				"At least one protected path is required",
				func() error {
					if len(cfg.Protected) == 0 || len(cfg.Protected) >= minPaths {
						return nil
					}

					return fmt.Errorf(
						"protected paths (%d) below required minimum (%d)",
						len(cfg.Protected), minPaths,
					)
				}),
			fmt.Sprintf("Protect at least %d paths", minPaths),
		), cfg.Protected, nil)
	}
}

// protectedPathsMinimum returns the configured minimum protected-path count,
// or 0 when unset (the required-empty rule already covers that case).
func (cv *ConfigValidator) protectedPathsMinimum() int {
	if cv.rules.MinProtectedPaths == nil || cv.rules.MinProtectedPaths.Min == nil {
		return 0
	}

	return *cv.rules.MinProtectedPaths.Min
}

func requiredCheck(passes bool, failure error) func() error {
	return func() error {
		if passes {
			return nil
		}

		return failure
	}
}

func (cv *ConfigValidator) addFieldRules(set *configRuleSet, cfg *types.Config) {
	minUsage, maxUsage := cv.getMaxDiskUsageBounds()
	set.add(withSuggestion(
		configRule("max_disk_usage", levelField, "range", businessrules.SeverityError,
			"Max disk usage out of range",
			func() error {
				return cv.validateMaxDiskUsage(cfg.MaxDiskUsage)
			}),
		diskUsageSuggestion(cv.rules.MaxDiskUsage, minUsage, maxUsage),
	), cfg.MaxDiskUsage, &ValidationContext{ //nolint:exhaustruct
		MinValue: minUsage,
		MaxValue: maxUsage,
	})

	set.add(withSuggestion(
		configRule("protected", levelField, "format", businessrules.SeverityError,
			"Protected paths must be absolute",
			func() error {
				return cv.validateProtectedPaths(cfg.Protected)
			}),
		"Ensure all paths are valid absolute paths",
	), cfg.Protected, nil)

	if cv.rules.UniquePaths {
		set.add(withSuggestion(
			configRule("protected", levelField, "unique", businessrules.SeverityWarning,
				"Protected paths should be unique",
				func() error {
					if duplicates := cv.findDuplicatePaths(cfg.Protected); len(duplicates) > 0 {
						return fmt.Errorf("Duplicate protected paths found: %v", duplicates)
					}

					return nil
				}),
			"Remove duplicate paths from protected list",
		), cfg.Protected, nil)
	}

	if cv.rules.MaxProfiles != nil && cv.rules.MaxProfiles.Max != nil {
		maxProfiles := *cv.rules.MaxProfiles.Max

		set.add(withSuggestion(
			configRule("profiles", levelField, "max_count", businessrules.SeverityWarning,
				"Profile count exceeds recommended limit",
				func() error {
					if len(cfg.Profiles) > maxProfiles {
						return fmt.Errorf(
							"Profile count (%d) exceeds recommended limit (%d)",
							len(cfg.Profiles), maxProfiles,
						)
					}

					return nil
				}),
			"Consider consolidating profiles to improve maintainability",
		), cfg.Profiles, nil)
	}
}

func diskUsageSuggestion(rule *ValidationRule[int], minUsage, maxUsage int) string {
	if rule != nil && rule.Message != "" {
		return rule.Message
	}

	return fmt.Sprintf("Set max_disk_usage between %d and %d", minUsage, maxUsage)
}

func (cv *ConfigValidator) addCrossFieldRules(set *configRuleSet, cfg *types.Config) {
	set.add(withSuggestion(
		configRule("safe_mode", levelCrossField, "risk_consistency", businessrules.SeverityWarning,
			"Safe mode and risk level are inconsistent",
			func() error {
				if cfg.SafeMode.IsEnabled() {
					return nil
				}

				maxRisk := cv.findMaxRiskLevel(cfg)
				if maxRisk != enums.RiskLevelCriticalType {
					return nil
				}

				return errors.New("Critical risk operations enabled while safe_mode is false")
			}),
		"Enable safe_mode or review critical risk operations",
	), nil, &ValidationContext{ //nolint:exhaustruct
		Metadata: map[string]string{
			"max_risk_level": cv.findMaxRiskLevel(cfg).String(),
			"safe_mode":      fmt.Sprintf("%v", cfg.SafeMode), //nolint:goconst
		},
	})

	for _, name := range sortedProfileNames(cfg) {
		profile := cfg.Profiles[name]
		if profile == nil {
			continue
		}

		if cv.rules.MaxOperations == nil || cv.rules.MaxOperations.Max == nil {
			continue
		}

		maxOperations := *cv.rules.MaxOperations.Max
		field := fmt.Sprintf("profiles.%s.operations", name)
		set.add(withSuggestion(
			configRule(field, levelCrossField, "max_operations", businessrules.SeverityWarning,
				"Operation count exceeds recommended limit",
				func() error {
					if len(profile.Operations) > maxOperations {
						return fmt.Errorf(
							"Profile '%s' has %d operations, exceeding recommended limit (%d)",
							name, len(profile.Operations), maxOperations,
						)
					}

					return nil
				}),
			"Consider splitting operations into multiple profiles",
		), len(profile.Operations), &ValidationContext{ //nolint:exhaustruct
			Metadata: map[string]string{
				"operation_count": strconv.Itoa(len(profile.Operations)),
				"max_operations":  strconv.Itoa(maxOperations),
			},
		})
	}
}

func (cv *ConfigValidator) addBusinessLogicRules(set *configRuleSet, cfg *types.Config) {
	for _, name := range sortedProfileNames(cfg) {
		profile := cfg.Profiles[name]
		if profile == nil {
			continue
		}

		operationsField := fmt.Sprintf("profiles.%s.operations", name)
		set.add(withSuggestion(
			configRule(operationsField, levelBusiness, "business_logic", businessrules.SeverityError,
				"Profile must have at least one operation",
				func() error {
					if len(profile.Operations) == 0 {
						return fmt.Errorf("Profile '%s' must have at least one operation", name)
					}

					return nil
				}),
			"Add at least one valid operation to profile",
		), len(profile.Operations), nil)

		for _, operation := range profile.Operations {
			cv.addOperationRules(set, cfg, name, operation)
		}
	}
}

func (cv *ConfigValidator) addOperationRules(
	set *configRuleSet, cfg *types.Config, profileName string, operation types.CleanupOperation,
) {
	unsafeCriticalCheck := func() error {
		if !cfg.SafeMode.IsEnabled() && operation.RiskLevel == enums.RiskLevelCriticalType {
			return fmt.Errorf(
				"Critical risk operation '%s' not allowed in unsafe mode", operation.Name,
			)
		}

		return nil
	}

	riskField := fmt.Sprintf("profiles.%s.operations.%s.risk_level", profileName, operation.Name)
	set.add(withSuggestion(
		configRule(riskField, levelBusiness, "business_logic", businessrules.SeverityError,
			"Critical risk operation in unsafe mode", unsafeCriticalCheck),
		"Enable safe mode or remove critical risk operation",
	), operation.RiskLevel, nil)

	if operation.Settings != nil {
		settingsField := fmt.Sprintf(
			"profiles.%s.operations.%s.settings", profileName, operation.Name,
		)
		settings := operation.Settings
		opType := operations.GetOperationType(operation.Name)

		set.add(withSuggestion(
			configRule(settingsField, levelBusiness, "validation", businessrules.SeverityError,
				"Operation settings are invalid",
				func() error {
					if err := settings.ValidateSettings(opType); err != nil {
						return fmt.Errorf("invalid settings for operation '%s': %w", operation.Name, err)
					}

					return nil
				}),
			"Fix operation settings according to validation rules",
		), settings, nil)
	}

	protectedConflictField := fmt.Sprintf(
		"profiles.%s.operations.%s", profileName, operation.Name,
	)
	set.add(withSuggestion(
		configRule(protectedConflictField, levelBusiness, "protected_conflict",
			businessrules.SeverityWarning,
			"Operation may affect protected paths",
			func() error {
				return cv.validateProtectedPathsConflict(cfg.Protected, operation)
			}),
		"Review operation scope and protected paths configuration",
	), nil, nil)
}

func (cv *ConfigValidator) addSecurityRules(set *configRuleSet, cfg *types.Config) {
	for _, path := range cfg.Protected {
		if path == "/" {
			set.add(withSuggestion(
				configRule("protected", levelSecurityTag, "security", businessrules.SeverityWarning,
					"Protecting root directory may prevent system operations",
					func() error {
						return errors.New("Protecting root directory '/' may prevent system operations")
					}),
				"Consider protecting specific system directories instead",
			), path, &ValidationContext{ //nolint:exhaustruct
				Metadata: map[string]string{"protected_path": path},
			})
		}

		if strings.Contains(path, "..") {
			set.add(withSuggestion(
				configRule("protected", levelSecurityTag, "security",
					businessrules.SeverityCritical,
					"Protected path contains parent directory reference",
					func() error {
						return errors.New("Protected path contains parent directory reference '..'")
					}),
				"Use absolute paths without parent directory references",
			), path, nil)
		}
	}

	for _, name := range sortedProfileNames(cfg) {
		profile := cfg.Profiles[name]
		if profile == nil {
			continue
		}

		for _, operation := range profile.Operations {
			requiresSafeMode := func() error {
				if operation.RiskLevel == enums.RiskLevelCriticalType &&
					!cfg.SafeMode.IsEnabled() {
					return fmt.Errorf(
						"Critical risk operation '%s' requires safe mode enabled", operation.Name,
					)
				}

				return nil
			}

			riskField := fmt.Sprintf(
				"profiles.%s.operations.%s.risk_level", name, operation.Name,
			)
			set.add(withSuggestion(
				configRule(riskField, levelSecurityTag, "security",
					businessrules.SeverityError,
					"Critical risk operation requires safe mode", requiresSafeMode),
				"Enable safe mode or remove critical risk operations",
			), operation.RiskLevel, nil)
		}
	}
}

func sortedProfileNames(cfg *types.Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}
