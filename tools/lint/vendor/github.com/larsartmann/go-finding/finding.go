package finding

// keySeparator is used by Key() to build deterministic composite keys.
// Using NUL ensures no collision with visible characters in field values.
const keySeparator = "\x00"

// Finding represents a single issue detected by a static analysis tool.
type Finding struct {
	// Identity
	ID       ID       `json:"id"`       // Stable unique identifier (e.g., "tool:rule:file:42:5")
	Rule     RuleName `json:"rule"`     // Rule/check name (e.g., "STRONG_ID", "clone-detected")
	ToolName ToolName `json:"toolName"` // Source tool name (e.g., "branching-flow", "art-dupl")

	// Core
	Message  string   `json:"message"`  // Human-readable description
	Severity Severity `json:"severity"` // info, warning, error, critical
	Position Position `json:"position"` // Where the issue is

	// Classification
	Category Category `json:"category,omitempty"` // Domain: "security", "style", "duplication", etc.
	Tags     []Tag    `json:"tags,omitempty"`     // Multiple tags for richer classification
	// Fix
	FixStrategy FixStrategy `json:"fixStrategy"`          // none, suggest, direct, ai
	Suggestion  string      `json:"suggestion,omitempty"` // Human-readable fix description
	BeforeCode  string      `json:"beforeCode,omitempty"` // Code before the fix
	AfterCode   string      `json:"afterCode,omitempty"`  // Code after the fix
	// Edits is the full typed edit list of the fix. When set, it is the
	// authoritative machine representation; BeforeCode/AfterCode carry the
	// first edit as a display summary. Empty for single-edit findings built
	// via Builder/BeforeCode only. When both representations are set, they
	// are intentionally loose: the display pair may be truncated or
	// contextualized for humans, so Validate does not require it to agree
	// with Edits[0] (ADR #19). Consumers must apply Edits, never reconstruct
	// the fix from the display pair.
	Edits []TextEdit `json:"edits,omitempty"`

	// Context
	Range       *Range       `json:"range,omitempty"`       // For span-based findings
	Snippet     string       `json:"snippet,omitempty"`     // Surrounding code context
	Confidence  Confidence   `json:"confidence,omitempty"`  // 0.0-1.0
	GroupID     GroupID      `json:"groupId,omitempty"`     // Logical group this finding belongs to
	Related     []RelatedRef `json:"related,omitempty"`     // Related findings
	Suppression *Suppression `json:"suppression,omitempty"` // If suppressed

	// Extensibility
	// Design decision: We intentionally have only Metadata (map[string]string), NOT a
	// Properties map[string]any. Use string-valued metadata for extensibility.
	// If you need complex values, JSON-serialize them into a string value.
	// Rationale: keeps the struct simple, avoids type-assertion boilerplate,
	// and Metadata is fully typed as string→string which is lossless for
	// interchange (SARIF, JSON, CLI flags, env vars).
	//
	// Key namespacing: use "toolName.key" format to prevent collisions between
	// tools. For example, "govet.category" vs "staticcheck.category". The
	// "go-finding/" prefix is reserved for internal use (SARIF round-trip,
	// LSP diagnostic tags, etc.).
	Metadata map[string]string `json:"metadata,omitempty"` // Tool-specific key-value pairs
}

// NewFinding creates a Finding with an auto-generated ID, default fix strategy,
// and clamped confidence.
func NewFinding(
	rule RuleName, toolName ToolName, message string,
	severity Severity,
	pos Position,
	confidence Confidence,
) Finding {
	return Finding{
		ID:          GenerateID(toolName, rule, pos),
		Rule:        rule,
		ToolName:    toolName,
		Message:     message,
		Severity:    severity,
		Position:    pos,
		FixStrategy: FixStrategyNone,
		Confidence:  confidence.Clamp(),
	}
}

// RelationKind describes the type of relationship between two findings.
type RelationKind string

// Standard relation kinds for relating findings.
const (
	RelationCloneOf RelationKind = "clone-of"
	RelationCauses  RelationKind = "causes"
	RelationWraps   RelationKind = "wraps"
	RelationRelated RelationKind = "related"
)

// IsValid returns true if the relation kind is a recognized standard value.
func (r RelationKind) IsValid() bool {
	switch r {
	case RelationCloneOf, RelationCauses, RelationWraps, RelationRelated:
		return true
	}

	return false
}

// RelatedRef links to another finding.
type RelatedRef struct {
	FindingID ID           `json:"findingId"`       // ID of the related finding
	Relation  RelationKind `json:"relation"`        // e.g., RelationCloneOf, RelationCauses
	Position  Position     `json:"position"`        // Quick access to related location
	Range     *Range       `json:"range,omitempty"` // Span of the related location
}

// IsValid returns true if the reference has a non-empty FindingID and a valid Relation.
func (r RelatedRef) IsValid() bool {
	return r.FindingID != "" && r.Relation.IsValid()
}
