package finding

// Range represents a span in source code from Start to End.
type Range struct {
	Start Position `json:"start"` // Required: start position
	End   Position `json:"end"`   // Optional: end position
}

// SameFile reports whether the range's Start and End are in the same file.
// Returns true when End.File is empty (single-file convention) or matches Start.File.
func (r Range) SameFile() bool {
	return r.End.File == "" || r.End.File == r.Start.File
}

// IsValid returns true if the range has a valid start position.
func (r Range) IsValid() bool {
	return r.Start.IsValid()
}

// IsInverted returns true if the range has both Start and End lines set
// and End is before Start.
func (r Range) IsInverted() bool {
	if r.Start.Line <= 0 || r.End.Line <= 0 {
		return false
	}

	if r.End.Line < r.Start.Line {
		return true
	}

	if r.End.Line == r.Start.Line && r.Start.Column > 0 && r.End.Column > 0 &&
		r.End.Column < r.Start.Column {
		return true
	}

	return false
}

// HasEnd returns true if the range has an end position set.
func (r Range) HasEnd() bool {
	return r.End.Line > 0 || r.End.Offset >= 0
}

// EndOrStart returns the effective end position of the range.
// A range with no end (End.Line == 0) is treated as a single point
// at Start. This matches the convention used for line-based range
// arithmetic (overlap, intersection, extension), where a single-point
// range's effective end equals its start.
func (r Range) EndOrStart() Position {
	if r.End.Line == 0 {
		return r.Start
	}

	return r.End
}

// EndOffsetOrStart returns the effective end byte offset of the range.
// A range with no offset end (End.Offset < 0) is treated as a single
// point at Start. This is the offset-based counterpart to EndOrStart.
func (r Range) EndOffsetOrStart() int {
	if r.End.Offset < 0 {
		return r.Start.Offset
	}

	return r.End.Offset
}

// LineCount returns the number of lines spanned by the range.
// Returns 1 if End is not set (single-line range). Returns 0 if Start has no line info.
// For inverted ranges (End.Line < Start.Line), returns the absolute span.
func (r Range) LineCount() int {
	if r.Start.Line == 0 {
		return 0
	}

	if r.End.Line == 0 {
		return 1
	}

	span := r.End.Line - r.Start.Line
	if span < 0 {
		span = -span
	}

	return span + 1
}

// IsSingleLine reports whether the range spans exactly one line.
func (r Range) IsSingleLine() bool {
	return r.LineCount() == 1
}

// Length returns the byte length of the range (End.Offset - Start.Offset).
// Returns 0 if either offset is not set. Returns 0 if End < Start.
func (r Range) Length() int {
	if r.Start.Offset < 0 || r.End.Offset < 0 {
		return 0
	}

	length := r.End.Offset - r.Start.Offset

	return max(length, 0)
}

// Equal reports whether two ranges are identical.
func (r Range) Equal(other Range) bool {
	return r.Start.Equal(other.Start) && r.End.Equal(other.End)
}

// Compare returns -1, 0, or 1 depending on whether r is less than, equal to,
// or greater than other. Ranges are ordered by start position, then end position.
func (r Range) Compare(other Range) int {
	if c := r.Start.Compare(other.Start); c != 0 {
		return c
	}

	return r.End.Compare(other.End)
}

// sameFileAs checks if this range is in the same file as another range.
func (r Range) sameFileAs(other Range) bool {
	return r.Start.File == other.Start.File
}

// hasLineInfo checks if both ranges have line information available.
func (r Range) hasLineInfo(other Range) bool {
	return r.Start.Line > 0 && other.Start.Line > 0
}

// containsSameFile checks if both positions are in the same file.
func (r Range) containsSameFile(p Position) bool {
	return r.Start.File == p.File && (r.End.File == "" || r.End.File == p.File)
}

// hasLineRange checks if p is within the range's line bounds.
// Called only when both r.Start.Line and p.Line are greater than zero.
func (r Range) hasLineRange(p Position) bool {
	if p.Line < r.Start.Line {
		return false
	}

	if r.End.Line > 0 && p.Line > r.End.Line {
		return false
	}

	return r.checkColumnRange(p)
}

// checkColumnRange checks if p's column is within the range's column bounds.
func (r Range) checkColumnRange(p Position) bool {
	if r.Start.Column > 0 && p.Column > 0 && p.Column < r.Start.Column {
		return false
	}

	if r.End.Column > 0 && p.Column > 0 && p.Column > r.End.Column {
		return false
	}

	return true
}

// containsByOffset checks if position is within range using byte offsets.
// Respects the Offset sentinel convention: End.Offset < 0 means "unset", so a
// range without an end offset is treated as a single point at Start.Offset
// (via [Range.EndOffsetOrStart]). Offset 0 is a valid byte offset (start of file).
func (r Range) containsByOffset(p Position) bool {
	if r.Start.Offset < 0 || p.Offset < 0 {
		return false
	}

	if p.Offset < r.Start.Offset {
		return false
	}

	if p.Offset > r.EndOffsetOrStart() {
		return false
	}

	return true
}

// Contains reports whether the position is within the range.
// Checks same file, line range, and offset when line ranges aren't available.
func (r Range) Contains(p Position) bool {
	if !r.containsSameFile(p) {
		return false
	}

	// If both range and position have line info, use line-based check exclusively
	if r.Start.Line > 0 && p.Line > 0 {
		return r.hasLineRange(p)
	}

	return r.containsByOffset(p)
}

// HasOffset reports whether the byte offset is set (not the -1 sentinel).
// Position{} (the Go zero value) has Offset=0, which HasOffset reports as true —
// byte 0 is a valid offset. Constructors (Pos, NewRange, FromLSP, SARIF import)
// set Offset to -1 when no byte offset is available, so HasOffset returns false.
func (p Position) HasOffset() bool {
	return p.Offset >= 0
}

// Pos is a convenience constructor for Position.
// It creates a Position with the given file, line, and column.
// Offset is set to -1 (unset) since byte offset is not provided.
func Pos(file FilePath, line, column int) Position {
	return Position{File: file, Line: line, Column: column, Offset: -1}
}

// NewRange creates a Range with the given file, start/end lines, and columns.
// Offsets are set to -1 (unset) since byte offsets are not provided.
func NewRange(file FilePath, startLine, startCol, endLine, endCol int) Range {
	return Range{
		Start: Position{File: file, Line: startLine, Column: startCol, Offset: -1},
		End:   Position{File: file, Line: endLine, Column: endCol, Offset: -1},
	}
}

// NewRangePtr creates a pointer to a Range with the given file, start/end lines, and columns.
func NewRangePtr(file FilePath, startLine, startCol, endLine, endCol int) *Range {
	r := NewRange(file, startLine, startCol, endLine, endCol)

	return &r
}

// RangeLinesEq checks if two ranges have equal start/end lines (ignoring columns/files).
func RangeLinesEq(a, b Range) bool {
	return a.Start.Line == b.Start.Line && a.End.Line == b.End.Line
}
