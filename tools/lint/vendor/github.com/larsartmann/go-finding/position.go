package finding

import (
	"cmp"
	"fmt"
)

// OffsetUnknown is the sentinel value for "no byte offset set."
// Use this instead of raw -1 literals to centralize the convention.
const OffsetUnknown = -1

// Position represents a location in source code.
//
// Sentinel values:
//   - Line, Column: 0 means "not set" (1-based, so 0 is never valid).
//   - Offset: -1 means "not set" (0-based, so 0 means "start of file" which IS valid).
//
// Constructors (Pos, NewRange, NewFinding, FromLSP, SARIF import) set Offset to -1
// when no byte offset is available. The zero value Position{} has Offset=0, which
// means "byte offset 0" (start of file) — NOT "unset". This is a deliberate design
// choice: Offset=0 is a valid byte position, and the zero value should not lie about
// having data.
//
// File-level positions: A Position with a file path but Line=0 is valid for findings
// that apply to an entire file (config issues, project checks). Use [FilePos] to create
// such positions. [Position.IsValid] still requires Line>0 for backward compatibility;
// use [Position.HasFile] to check only whether a file is set.
//
// Use HasLocation() to check for a meaningful position (file + line).
// Use HasOffset() to check whether a byte offset is set (Offset >= 0).
// Use IsZero() to check for the completely-uninitialized state (all fields at their
// "unset" sentinels: empty File, Line=0, Column=0, Offset=-1).
type Position struct {
	File   FilePath `json:"file"`             // Required: file path
	Line   int      `json:"line,omitempty"`   // 1-based line number; 0 = not set
	Column int      `json:"column,omitempty"` // 1-based column number; 0 = not set
	Offset int      `json:"offset,omitempty"` // 0-based byte offset; -1 = not set
}

// IsValid returns true if the position has a file set and a non-zero line number.
// Line 0 means "not set" per the sentinel convention, so IsValid returns false
// for positions that lack a line number.
//
// For checking only whether a file path is present, use [Position.HasFile].
func (p Position) IsValid() bool {
	return p.File != "" && p.Line > 0
}

// HasFile reports whether the position has a file path set, regardless of
// line/column completeness.
func (p Position) HasFile() bool {
	return p.File != ""
}

// FilePos creates a Position for a file-level finding (no line/column).
// Use this for findings that apply to an entire file, such as configuration
// issues, project-level checks, or file-wide linting rules.
// The Offset is set to OffsetUnknown (-1) since no byte position applies.
func FilePos(file FilePath) Position {
	return Position{File: file, Offset: OffsetUnknown}
}

// IsZero reports whether the position is completely uninitialized (all fields
// at their "unset" sentinel values: empty File, Line=0, Column=0, Offset=-1).
//
// Note: Position{} (the Go zero value) has Offset=0 (byte 0), NOT Offset=-1 (unset),
// so Position{}.IsZero() returns false. This is correct: Position{} has a valid byte
// offset of 0, even though it lacks a file path. Use HasLocation() or IsValid() to
// check for a meaningful position.
func (p Position) IsZero() bool {
	return p.File == "" && p.Line == 0 && p.Column == 0 && p.Offset < 0
}

// HasLocation reports whether the position has a file and line number.
func (p Position) HasLocation() bool {
	return p.File != "" && p.Line > 0
}

// Equal reports whether two positions are identical.
func (p Position) Equal(other Position) bool {
	return p.File == other.File &&
		p.Line == other.Line &&
		p.Column == other.Column &&
		p.Offset == other.Offset
}

// Compare returns -1, 0, or 1 depending on whether p is less than, equal to,
// or greater than other. Positions are ordered by file, then line, then column,
// then offset. This is consistent with Equal: Compare returns 0 iff Equal returns true.
func (p Position) Compare(other Position) int {
	if c := cmp.Compare(p.File, other.File); c != 0 {
		return c
	}

	if c := cmp.Compare(p.Line, other.Line); c != 0 {
		return c
	}

	if c := cmp.Compare(p.Column, other.Column); c != 0 {
		return c
	}

	return cmp.Compare(p.Offset, other.Offset)
}

// String returns a human-readable representation.
func (p Position) String() string {
	if p.Line == 0 {
		return string(p.File)
	}

	if p.Column == 0 {
		return fmt.Sprintf("%s:%d", p.File, p.Line)
	}

	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}
