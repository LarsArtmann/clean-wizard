package finding

// Branded primitive types for compile-time safety.
// These prevent accidental mixing of string fields that represent
// distinct domain concepts (e.g., putting a Rule where an ID belongs).
// JSON serialization is identical to raw string — branded types marshal
// as strings with no overhead.

// ID is the stable unique identifier for a finding (tool:rule:file:line:col).
type ID string

// RuleName is the rule or check name (e.g., "nilcheck", "STRONG_ID").
type RuleName string

// ToolName is the source tool name (e.g., "govet", "staticcheck").
type ToolName string

// FilePath is a path to a source file.
type FilePath string

// GroupID groups findings that belong to the same logical set
// (e.g., a clone group of N duplicated code blocks).
//
// A set GroupID must be a machine-safe identifier: no whitespace, no
// control characters, at most 128 bytes. The empty GroupID means
// "not grouped" and is always valid. See [GroupID.IsValid].
type GroupID string

// maxGroupIDLen is the upper bound for a GroupID value. Group IDs travel
// into SARIF property keys, LSP data, and map keys; unbounded values are
// a corruption risk for every wire format at once.
const maxGroupIDLen = 128

// IsValid reports whether the GroupID is usable as a machine identifier.
// The empty value ("not grouped") is valid. Non-empty values must not
// contain whitespace or control characters and must fit in 128 bytes.
// The charset is deliberately permissive (underscores, dots, uppercase
// are fine) — this guards the wire formats, not naming style.
func (g GroupID) IsValid() bool {
	if g == "" {
		return true
	}

	if len(g) > maxGroupIDLen {
		return false
	}

	for _, r := range g {
		if r <= 0x20 || r == 0x7f {
			return false
		}
	}

	return true
}
