// Package linter provides the shared scaffolding for LarsArtmann Go linters
// (branching-flow, erraudit, go-structure-linter, and future
// additions).
//
// Each of those linters independently reinvented the same three layers:
//   - A Rule interface (Name/Description/Severity/Check)
//   - A registry of rules
//   - A converter from their native domain type to finding.Finding
//
// The last layer — the converter — is the most expensive duplication:
// branching-flow ships 1,871 LOC of converters, erraudit 1,214 LOC,
// just to bridge their Violation/ErrorViolation types to finding.Finding.
// go-structure-linter eliminated this entirely by aliasing
// `type Issue = finding.Finding`.
//
// This SDK codifies the go-structure-linter pattern: a Rule emits finding.Finding
// directly, so there is NO converter layer. A linter that adopts this SDK ships
// a rules.go file and a main.go one-liner — the registry, detection, and exit
// codes are shared.
//
// Design constraint: depends only on go-finding (the ecosystem hub) so any
// linter — CLI, library, LSP server, golangci plugin — can adopt it.
package linter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/larsartmann/go-finding"
)

// Category classifies what kind of issue a rule detects, for filtering and
// reporting. Maps to finding.Category at the finding boundary.
type Category string

// Category values classify what kind of issue a rule detects, for filtering
// and reporting. Each maps to finding.Category at the finding boundary.
//
// These are RECOMMENDED values, not a closed set. Define your own for
// domain-specific taxonomies:
//
//	const CategoryAPI Category = "api" // your domain category
const (
	CategoryDesign        Category = "design"    // design smells (coupling, cohesion)
	CategoryStructure     Category = "structure" // file/package layout
	CategoryErrorHandling Category = "error-handling"
	CategoryCorrectness   Category = "correctness"
	CategoryNaming        Category = "naming"        // naming conventions
	CategoryDocumentation Category = "documentation" // documentation conventions
	CategoryBestPractice  Category = "best-practice" // project best practices
	CategoryStyle         Category = "style"
	CategoryPerformance   Category = "performance"
	CategorySecurity      Category = "security"
	CategoryConfiguration Category = "configuration"
)

// Rule is a single lint check. A linter is a collection of rules. Each rule
// declares its identity (ID/Name/Description/Category/Severity) and a Check
// function that emits findings for a target directory.
//
// # Dual identity: ID vs Name
//
// Every rule has two identity fields with distinct stability contracts:
//
//   - ID() is the STABLE identifier. It never changes once published. Used for
//     registry deduplication, suppression matching (//tool:ignore ID), filter
//     config (--enable ID), and the finding.RuleName field. Think "G001",
//     "C001", "no-unused-vars".
//   - Name() is the DISPLAY name. Human-readable, can change across versions
//     without breaking suppression, filtering, or health-score computation.
//     Think "missing transaction commit", "unused variable".
//
// Rules emit finding.Finding DIRECTLY — there is no intermediate Violation or
// Issue type to convert. This is the key design decision that eliminates the
// thousands of LOC of converter code the existing linters carry. Use
// finding.NewBuilder(...) to construct each finding.
//
// # Building findings with Confidence and FixStrategy
//
// finding.Finding already carries Confidence (finding.ConfidenceLow/Medium/High)
// and FixStrategy (finding.FixStrategyNone/Suggest/Direct/AI) at the per-finding
// level. Set these on individual findings via finding.NewBuilder(...):
//
//	finding.NewBuilder(ruleID, toolName, msg, sev, pos).
//	    WithConfidence(finding.ConfidenceHigh).
//	    WithFixStrategy(finding.FixStrategyDirect).
//	    MustBuild()
//
// Per-finding values are strictly more expressive than rule-level defaults: a
// single rule can emit findings with different confidence levels and fix
// strategies depending on the matched pattern. RuleMeta.Sev is the default
// severity — individual findings can override it via the Builder.
type Rule interface {
	ID() string
	Name() string
	Description() string
	Category() Category
	Severity() finding.Severity
	IsEnabledByDefault() bool
	Check(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleFunc adapts a function to the Rule interface, filling in the identity
// fields from the supplied RuleMeta. This is the common case: most rules are a
// metadata header plus a check closure.
type RuleFunc struct {
	Meta RuleMeta
	Run  func(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleMeta is the declarative identity of a rule: ID, name, description,
// category, severity. Supplied as a struct literal so rules read like data.
//
// ID is REQUIRED — it is the stable identifier that never changes once
// published. It is used for registry deduplication, suppression matching,
// filter config, and the finding.RuleName field.
type RuleMeta struct {
	ID          string
	Name        string
	Description string
	Cat         Category
	Sev         finding.Severity
	// ToolName is the tool name stamped onto findings created via [RuleFunc.NewFinding].
	// When empty, NewFinding falls back to "linter". Set this so findings are attributed
	// to the actual linter, not a generic "linter" string. A [Registry] configured with
	// [WithToolName] auto-stamps this field at registration time if the rule does not set it.
	ToolName finding.ToolName
}

// ErrMissingFields is the sentinel error returned by RuleMeta.Validate and
// Register when a rule has one or more empty required identity fields. Use
// errors.Is to check:
//
//	if errors.Is(err, linter.ErrMissingFields) {
//	    // rule has missing ID, Name, Description, or Category
//	}
var ErrMissingFields = errors.New("linter: rule has missing required field(s)")

// Validate returns nil if all required fields are non-empty, or an error
// listing every missing field. ID, Name, Description, and Cat are required;
// Sev defaults to the zero value (empty string) which is a valid — if
// imprecise — severity for rules that set severity per-finding via the
// Builder.
//
// Call this during rule construction to fail fast on misconfigured rules
// before they reach the registry:
//
//	meta := linter.RuleMeta{ID: "x", Name: "X", ...}
//	if err := meta.Validate(); err != nil { log.Fatal(err) }
func (m RuleMeta) Validate() error {
	return validateIdentityFields(m.ID, m.Name, m.Description, m.Cat)
}

// validateRuleIdentity checks a Rule through its interface methods and returns
// an error listing every empty required field. Register panics on this error
// to surface misconfigured rules at startup. Works for any Rule implementation,
// not just RuleFunc.
func validateRuleIdentity(r Rule) error {
	return validateIdentityFields(r.ID(), r.Name(), r.Description(), r.Category())
}

// validateIdentityFields is the shared validation core for RuleMeta.Validate
// and validateRuleIdentity. Returns an error listing every empty field by name.
func validateIdentityFields(id, name, description string, cat Category) error {
	var missing []string

	if id == "" {
		missing = append(missing, "ID")
	}

	if name == "" {
		missing = append(missing, "Name")
	}

	if description == "" {
		missing = append(missing, "Description")
	}

	if cat == "" {
		missing = append(missing, "Category")
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", ErrMissingFields, strings.Join(missing, ", "))
	}

	return nil
}

// optInRule is a RuleFunc with default-enabled=false. Created via OptIn().
// Embeds RuleFunc so all methods (ID/Name/Description/Category/Severity/Check)
// are inherited; only IsEnabledByDefault is overridden.
type optInRule struct {
	RuleFunc
}

func (optInRule) IsEnabledByDefault() bool { return false }

// OptIn returns a Rule that is disabled by default — it only runs when a
// consumer explicitly enables it (e.g. via --enable <id>). Use for rules
// that are noisy, experimental, or domain-specific:
//
//	r.Register(linter.OptIn(linter.RuleFunc{Meta: ..., Run: ...}))
func OptIn(rf RuleFunc) Rule { return optInRule{rf} }

// ID implements Rule. Returns the stable identifier from RuleMeta.ID.
func (r RuleFunc) ID() string { return r.Meta.ID }

// Name implements Rule.
func (r RuleFunc) Name() string { return r.Meta.Name }

// Description implements Rule.
func (r RuleFunc) Description() string { return r.Meta.Description }

// Category implements Rule.
func (r RuleFunc) Category() Category { return r.Meta.Cat }

// Severity implements Rule.
func (r RuleFunc) Severity() finding.Severity { return r.Meta.Sev }

// IsEnabledByDefault returns true for RuleFunc. Rules created via RuleFunc are
// enabled by default — they run unless a consumer explicitly disables them.
// For opt-in rules, use OptIn().
//
// This is METADATA only — Run, DetectorFromRegistry, and DetectorsFromRegistry
// execute every registered rule regardless of this value. Consumers that want
// --enable/--disable semantics must filter the rule set themselves, e.g. via
// [FilterRules].
func (RuleFunc) IsEnabledByDefault() bool { return true }

// Check implements Rule. If Run returns an error, it is wrapped into a
// *RuleError carrying the rule's stable ID, so callers of Registry.Run and
// DetectorFromRegistry can identify which rule failed via errors.AsType
// (see [RuleError]).
func (r RuleFunc) Check(ctx context.Context, dir string) ([]finding.Finding, error) {
	findings, err := r.Run(ctx, dir)
	if err != nil {
		return findings, NewRuleError(r.Meta.ID, err)
	}

	return findings, nil
}

// NewFinding returns a [finding.Builder] pre-configured with the rule's identity:
// rule ID, tool name, default severity, and category — all drawn from [RuleMeta].
// Chain per-finding fields (confidence, suggestion, before/after code, etc.) on
// the returned builder before calling Build, MustBuild, or BuildOrDefault.
//
// This eliminates the boilerplate of repeating rule ID, tool name, severity, and
// category in every [finding.NewBuilder] call inside a rule's Run function.
//
// Example:
//
//	f := rule.NewFinding("manual byte formatting", pos).
//	    WithConfidence(finding.ConfidenceHigh).
//	    WithSuggestion("use humanize.Bytes").
//	    MustBuild()
func (r RuleFunc) NewFinding(message string, pos finding.Position) *finding.Builder {
	tool := r.Meta.ToolName
	if tool == "" {
		tool = "linter"
	}

	builder := finding.NewBuilder(
		finding.RuleName(r.Meta.ID),
		tool,
		message,
		r.Meta.Sev,
		pos,
	)

	if r.Meta.Cat != "" {
		builder = builder.WithCategory(finding.Category(r.Meta.Cat))
	}

	return builder
}
