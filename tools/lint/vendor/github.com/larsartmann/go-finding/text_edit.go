package finding

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// TextEdit is a single replacement operation within a finding's suggested
// fix: the span [Start, End) is replaced by NewText. It mirrors go/analysis's
// analysis.TextEdit and the LSP TextEdit concept, so multi-edit fixes (e.g.
// remove a declaration AND rewrite a condition) travel losslessly on the
// Finding instead of being truncated to the first edit.
//
// Conventions:
//   - An insertion is expressed either by an unset End (Position{Offset: -1})
//     or by End == Start (zero-length span); both resolve to a byte-level
//     no-removal insertion, mirroring go/analysis's Pos == End convention.
//   - Empty NewText means a pure deletion of the span.
//   - Empty Start.File means "same file as the finding" (single-file default).
//   - Start and End follow the Position zero-value convention: Position{}
//     (Offset 0) means byte 0, not "unset"; the unset form is
//     Position{Offset: -1}.
//
// BeforeCode/AfterCode on Finding remain the single-edit display summary;
// Edits is the authoritative machine representation when set.
type TextEdit struct {
	Start   Position `json:"start"`             // Span start (inclusive).
	End     Position `json:"end"`               // Span end (exclusive); unset (Offset -1) or equal to Start = insertion.
	NewText string   `json:"newText,omitempty"` // Replacement text; empty = pure deletion.
}

// HasSpan reports whether the edit replaces a span, as opposed to being a
// pure insertion at Start (End unset, following the same convention as
// [Range.HasEnd]).
func (e TextEdit) HasSpan() bool {
	return e.End.Line > 0 || e.End.Offset >= 0
}

// IsInsertion reports whether this is a pure insertion (no span, replacement
// text present).
func (e TextEdit) IsInsertion() bool {
	return !e.HasSpan() && e.NewText != ""
}

// IsDeletion reports whether this is a pure deletion (span present, no
// replacement text).
func (e TextEdit) IsDeletion() bool {
	return e.HasSpan() && e.NewText == ""
}

// EffectiveFile returns the file the edit targets: Start.File, or the given
// fallback (the owning finding's file) when empty.
func (e TextEdit) EffectiveFile(fallback FilePath) FilePath {
	if e.Start.File != "" {
		return e.Start.File
	}

	return fallback
}

var (
	errEditNoLocation  = errors.New("text edit: Start must have a file, byte offset, or line")
	errEditInverted    = errors.New("text edit: End before Start")
	errEditCrossFile   = errors.New("text edit: End.File differs from Start.File")
	errEditNegOffset   = errors.New("text edit: negative byte offset")
	errEditOffsetCross = errors.New("text edit: Start byte offset set but End byte offset unset")
)

// Validate checks the edit for internal consistency. It does NOT check
// overlap between edits of the same list (an application-time concern,
// handled by the pipeline FixEngine) nor bounds against file content.
func (e TextEdit) Validate() error {
	if !e.Start.HasFile() && !e.Start.HasOffset() && e.Start.Line <= 0 {
		return fmt.Errorf("%w: %+v", errEditNoLocation, e.Start)
	}

	if e.Start.Offset < -1 || e.End.Offset < -1 {
		return fmt.Errorf("%w: start %+v, end %+v", errEditNegOffset, e.Start, e.End)
	}

	if !e.HasSpan() {
		return nil
	}

	if e.Start.File != "" && e.End.File != "" && e.End.File != e.Start.File {
		return fmt.Errorf("%w: %q vs %q", errEditCrossFile, e.End.File, e.Start.File)
	}

	if e.Start.HasOffset() {
		if !e.End.HasOffset() {
			return fmt.Errorf("%w: start %+v, end %+v", errEditOffsetCross, e.Start, e.End)
		}

		if e.End.Offset < e.Start.Offset {
			return fmt.Errorf("%w: end offset %d < start offset %d", errEditInverted, e.End.Offset, e.Start.Offset)
		}

		return nil
	}

	if e.Start.Line > 0 && e.End.Line > 0 {
		if e.End.Line < e.Start.Line {
			return fmt.Errorf("%w: end line %d < start line %d", errEditInverted, e.End.Line, e.Start.Line)
		}

		if e.End.Line == e.Start.Line && e.Start.Column > 0 && e.End.Column > 0 &&
			e.End.Column < e.Start.Column {
			return fmt.Errorf("%w: end column %d < start column %d", errEditInverted, e.End.Column, e.Start.Column)
		}
	}

	return nil
}

// Equal reports whether two edits are identical, including all position fields.
func (e TextEdit) Equal(other TextEdit) bool {
	return e.Start.Equal(other.Start) && e.End.Equal(other.End) && e.NewText == other.NewText
}

// Compare orders edits by Start, then End, then NewText. It provides a
// deterministic sort key for order-independent list equality.
func (e TextEdit) Compare(other TextEdit) int {
	if c := e.Start.Compare(other.Start); c != 0 {
		return c
	}

	if c := e.End.Compare(other.End); c != 0 {
		return c
	}

	return strings.Compare(e.NewText, other.NewText)
}

// editsEqual reports whether two edit lists contain the same edits regardless
// of order. Edit order is not semantically meaningful — the pipeline applies
// lists in descending offset order — so reordered lists describe the same fix.
func editsEqual(a, b []TextEdit) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	if slices.EqualFunc(a, b, TextEdit.Equal) {
		return true
	}

	sortedA := slices.Clone(a)
	sortedB := slices.Clone(b)

	slices.SortFunc(sortedA, TextEdit.Compare)
	slices.SortFunc(sortedB, TextEdit.Compare)

	return slices.EqualFunc(sortedA, sortedB, TextEdit.Equal)
}
