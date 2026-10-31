package finding

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// ID format constants.
const (
	idPartCount      = 3  // Minimum number of parts for hash-based IDs
	hashLength       = 16 // Length of hex-encoded hash
	idPartMin        = 4  // Minimum parts to attempt parsing column and line
	positionPartsOne = 1  // Number of positional parts when only line is present
	positionPartsTwo = 2  // Number of positional parts when line and column are present
)

// GenerateID creates a stable, unique identifier for a finding.
// Format: "tool:rule:file:line:col" (human-readable)
// If line is 0, uses hash-based ID for stability.
// Deliberately NOT derived from the finding's fix data (BeforeCode/AfterCode
// or the Edits list): a fix's internal shape may be refined without changing
// the finding's identity.
func GenerateID(toolName ToolName, rule RuleName, pos Position) ID {
	if pos.Line == 0 {
		// Hash-based for position-less findings.
		// Uses length-prefixed fields to prevent ambiguity when field
		// values contain the separator character (e.g., toolName="a:b", rule="c").
		h := sha256.New()
		writeLenField(h, string(toolName))
		writeLenField(h, string(rule))
		writeLenField(h, string(pos.File))

		sum := h.Sum(make([]byte, 0, sha256.Size))
		hash := hex.EncodeToString(sum[:hashLength/2])

		var b strings.Builder
		b.Grow(len(toolName) + 1 + len(rule) + 1 + len(hash))
		b.WriteString(string(toolName))
		b.WriteByte(':')
		b.WriteString(string(rule))
		b.WriteByte(':')
		b.WriteString(hash)

		return ID(b.String())
	}

	// Normalize file path to use forward slashes
	file := filepath.ToSlash(string(pos.File))

	var b strings.Builder

	lineStr := strconv.Itoa(pos.Line)
	b.Grow(len(toolName) + 1 + len(rule) + 1 + len(file) + 1 + len(lineStr))
	b.WriteString(string(toolName))
	b.WriteByte(':')
	b.WriteString(string(rule))
	b.WriteByte(':')
	b.WriteString(file)
	b.WriteByte(':')
	b.WriteString(lineStr)

	if pos.Column == 0 {
		return ID(b.String())
	}

	b.WriteByte(':')
	b.WriteString(strconv.Itoa(pos.Column))

	return ID(b.String())
}

// extractFile extracts the file path from ID parts, excluding trailing position components.
// The parts slice is expected to be [tool, rule, file parts..., line?, column?].
// trailingCount is the number of trailing position parts (1 for line only, 2 for line:col).
func extractFile(parts []string, trailingCount int) string {
	if len(parts) < idPartCount+1 { // Need at least tool:rule:file (3 parts)
		return ""
	}

	return strings.Join(parts[2:len(parts)-trailingCount], ":")
}

// ParsedID holds the components of a parsed finding ID.
type ParsedID struct {
	Tool   ToolName
	Rule   RuleName
	File   FilePath
	Line   int
	Column int
}

// OK returns true if the ID was successfully parsed.
func (p ParsedID) OK() bool {
	return p.Tool != ""
}

// ParseID parses a finding ID and extracts its components.
func ParseID(id ID) ParsedID {
	parts := strings.Split(string(id), ":")

	if len(parts) < idPartCount {
		return ParsedID{}
	}

	tool := ToolName(parts[0])
	rule := RuleName(parts[1])

	// Handle hash-based IDs
	if len(parts) == idPartCount && len(parts[2]) == hashLength { // hex encoded hash
		return ParsedID{Tool: tool, Rule: rule}
	}

	// Try to parse position from remaining parts
	// Format: tool:rule:file:line or tool:rule:file:line:col
	// File may contain colons (e.g., Windows paths), so we need to be careful
	// We assume the last 1-2 parts are line:column

	if len(parts) >= idPartMin {
		// Try parsing last part as column
		colTest := 0

		err := parseInt(parts[len(parts)-1], &colTest)
		if err == nil {
			// Try parsing second-to-last as line
			lineTest := 0

			err2 := parseInt(parts[len(parts)-2], &lineTest)
			if err2 == nil {
				file := extractFile(parts, positionPartsTwo)

				return ParsedID{Tool: tool, Rule: rule, File: FilePath(file), Line: lineTest, Column: colTest}
			}
		}

		// No column, try line only
		var line int

		err = parseInt(parts[len(parts)-1], &line)
		if err == nil {
			file := extractFile(parts, positionPartsOne)

			return ParsedID{Tool: tool, Rule: rule, File: FilePath(file), Line: line}
		}
	}

	// Just file, no position
	file := strings.Join(parts[2:], ":")

	return ParsedID{Tool: tool, Rule: rule, File: FilePath(file)}
}

// parseInt is a helper to parse a string to int, returning nil on success.
func parseInt(s string, result *int) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("failed to parse %q as int: %w", s, err)
	}

	*result = n

	return nil
}

// IsHashID returns true if the ID appears to be hash-based.
func IsHashID(id string) bool {
	parts := strings.Split(id, ":")
	if len(parts) != idPartCount {
		return false
	}

	return isHexString(parts[2]) && len(parts[2]) == hashLength
}

// isHexString reports whether s consists entirely of hex digits.
func isHexString(s string) bool {
	for _, r := range s {
		if !isHexDigit(r) {
			return false
		}
	}

	return len(s) > 0
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// writeLenField writes a length-prefixed field to the hash.
// Format: len(field) as uint32 big-endian, then the field bytes.
// This prevents ambiguity when field values contain the same characters as separators.
func writeLenField(h io.Writer, field string) {
	var buf [4]byte

	// G115: field length fits in uint32 on all reasonable inputs.
	// Strings exceeding 4GB are impossible in practice.
	binary.BigEndian.PutUint32(buf[:], uint32(len(field))) //nolint:gosec

	_, _ = h.Write(buf[:])        //nolint:erraudit // hash.Write is documented to never return an error
	_, _ = h.Write([]byte(field)) //nolint:erraudit // hash.Write is documented to never return an error
}
