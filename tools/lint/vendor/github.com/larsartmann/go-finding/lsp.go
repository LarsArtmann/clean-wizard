package finding

import (
	"maps"
	"strconv"
	"strings"
)

// LSPSeverity represents an LSP diagnostic severity level per the LSP specification.
type LSPSeverity int

// LSP severity level constants per the LSP specification.
const (
	LSPSeverityError   LSPSeverity = 1 // Error
	LSPSeverityWarning LSPSeverity = 2 // Warning
	LSPSeverityInfo    LSPSeverity = 3 // Information
	LSPSeverityHint    LSPSeverity = 4 // Hint
)

// LSP types for conversion.
// These are simplified representations of LSP Diagnostic types.

// LSPDiagnosticTag represents a diagnostic tag per the LSP specification (3.15+).
type LSPDiagnosticTag int

// LSP diagnostic tag constants per the LSP specification.
const (
	LSPDiagnosticTagUnnecessary LSPDiagnosticTag = 1 // Unnecessary code (e.g., unused, duplicate)
	LSPDiagnosticTagDeprecated  LSPDiagnosticTag = 2 // Deprecated code
)

// LSPDiagnostic represents an LSP (Language Server Protocol) diagnostic.
// Used for converting Finding objects to LSP diagnostic format.
//
// The Data field carries go-finding-specific metadata (finding ID, fix strategy,
// confidence, category, tags) for round-trip fidelity through LSP conversion.
// LSP clients that don't understand Data will simply ignore it.
type LSPDiagnostic struct {
	Range    LSPRange           `json:"range"`
	Severity LSPSeverity        `json:"severity,omitempty"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
	Tags     []LSPDiagnosticTag `json:"tags,omitempty"`
	Related  []LSPRelated       `json:"relatedInformation,omitempty"`
	Data     *LSPDiagnosticData `json:"data,omitempty"`
}

// LSPDiagnosticData carries go-finding-specific fields in the LSP diagnostic's
// data property. This enables round-trip fidelity for fields that the standard
// LSP diagnostic type cannot represent.
type LSPDiagnosticData struct {
	ID                ID                `json:"id,omitempty"`
	Severity          Severity          `json:"severity,omitempty"`
	FixStrategy       FixStrategy       `json:"fixStrategy,omitempty"`
	Confidence        Confidence        `json:"confidence,omitempty"`
	Category          Category          `json:"category,omitempty"`
	GroupID           GroupID           `json:"groupId,omitempty"`
	Tags              []Tag             `json:"tags,omitempty"`
	BeforeCode        string            `json:"beforeCode,omitempty"`
	AfterCode         string            `json:"afterCode,omitempty"`
	Suggestion        string            `json:"suggestion,omitempty"`
	Snippet           string            `json:"snippet,omitempty"`
	Suppression       *Suppression      `json:"suppression,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	RelatedFindingIDs []string          `json:"relatedFindingIds,omitempty"`
}

// LSPRange represents a 0-based character range in a text document.
type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

// LSPPosition represents a 0-based position in a text document.
type LSPPosition struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

// LSPRelated provides related information for a diagnostic.
type LSPRelated struct {
	Location LSPLocation `json:"location"`
	Message  string      `json:"message"`
}

// LSPLocation represents the location of a diagnostic.
type LSPLocation struct {
	URI   string   `json:"uri"`
	Range LSPRange `json:"range"`
}

// ToLSP converts a Finding to LSP Diagnostic format.
// The Data field carries go-finding-specific fields (ID, FixStrategy, Confidence,
// Category, Tags, code data) for round-trip fidelity via [FromLSP].
func (f Finding) ToLSP() LSPDiagnostic {
	diag := LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{
				Line:      toZeroBased(f.Position.Line),
				Character: toZeroBased(f.Position.Column),
			},
		},
		Severity: severityToLSP(f.Severity),
		Code:     string(f.Rule),
		Source:   string(f.ToolName),
		Message:  f.Message,
		Data: &LSPDiagnosticData{
			ID:                f.ID,
			Severity:          f.Severity,
			FixStrategy:       f.FixStrategy,
			Confidence:        f.Confidence,
			Category:          f.Category,
			GroupID:           f.GroupID,
			Tags:              f.Tags,
			BeforeCode:        f.BeforeCode,
			AfterCode:         f.AfterCode,
			Suggestion:        f.Suggestion,
			Snippet:           f.Snippet,
			Suppression:       f.Suppression,
			Metadata:          f.Metadata,
			RelatedFindingIDs: collectRelatedFindingIDs(f.Related),
		},
	}

	// Set end position if available
	if f.Range != nil && f.Range.HasEnd() {
		diag.Range.End = LSPPosition{
			Line:      toZeroBased(f.Range.End.Line),
			Character: toZeroBased(f.Range.End.Column),
		}
	} else {
		// Single position diagnostic
		diag.Range.End = diag.Range.Start
	}

	// Restore LSP diagnostic tags stored in metadata by FromLSP.
	if v, ok := f.Metadata[LSPDiagnosticTagsKey]; ok {
		diag.Tags = parseLSPDiagnosticTags(v)
	}

	// Add related information
	for _, rel := range f.Related {
		lspPos := LSPPosition{
			Line:      toZeroBased(rel.Position.Line),
			Character: toZeroBased(rel.Position.Column),
		}

		lspRange := LSPRange{Start: lspPos, End: lspPos}
		if rel.Range != nil && rel.Range.HasEnd() {
			lspRange.End = LSPPosition{
				Line:      toZeroBased(rel.Range.End.Line),
				Character: toZeroBased(rel.Range.End.Column),
			}
		}

		diag.Related = append(diag.Related, LSPRelated{
			Location: LSPLocation{
				URI:   string(rel.Position.File),
				Range: lspRange,
			},
			Message: string(rel.Relation),
		})
	}

	return diag
}

// toZeroBased converts a 1-based position to a 0-based LSP position, clamping to 0.
func toZeroBased(n int) int {
	if n <= 0 {
		return 0
	}

	return n - 1
}

// LSPSeverityKey is the Metadata key for preserving raw LSP severity codes.
const LSPSeverityKey = "go-finding/lsp-severity"

// LSPDiagnosticTagsKey is the Metadata key for preserving LSP diagnostic tags.
const LSPDiagnosticTagsKey = "go-finding/lsp-diagnostic-tags"

// FromLSP creates a Finding from an LSP Diagnostic at the given file URI.
// Preserves end position in Range and related information when present.
// When the diagnostic's Data field is populated (by ToLSP), restores the
// original ID, FixStrategy, Confidence, Category, Tags, and code data.
// The raw LSP severity integer is stored in Metadata under LSPSeverityKey.
func FromLSP(fileURI FilePath, diag LSPDiagnostic) Finding {
	startLine := diag.Range.Start.Line + 1
	startChar := diag.Range.Start.Character + 1

	f := Finding{
		ID: GenerateID(
			ToolName(diag.Source),
			RuleName(diag.Code),
			Position{File: fileURI, Line: startLine, Column: startChar, Offset: -1},
		),
		Rule:        RuleName(diag.Code),
		ToolName:    ToolName(diag.Source),
		Message:     diag.Message,
		Severity:    severityFromLSP(diag.Severity),
		FixStrategy: FixStrategyNone,
		Position: Position{
			File:   fileURI,
			Line:   startLine,
			Column: startChar,
			Offset: -1,
		},
	}

	// Restore go-finding-specific fields from Data (populated by ToLSP).
	if diag.Data != nil {
		if diag.Data.ID != "" {
			f.ID = diag.Data.ID
		}

		// Restore exact severity from Data when available, since LSP collapses
		// SeverityCritical into LSPSeverityError (SeverityCritical is not representable).
		if diag.Data.Severity.IsValid() {
			f.Severity = diag.Data.Severity
		}

		f.FixStrategy = diag.Data.FixStrategy
		f.Confidence = diag.Data.Confidence
		f.Category = diag.Data.Category
		f.GroupID = diag.Data.GroupID
		f.Tags = diag.Data.Tags
		f.BeforeCode = diag.Data.BeforeCode
		f.AfterCode = diag.Data.AfterCode
		f.Suggestion = diag.Data.Suggestion
		f.Snippet = diag.Data.Snippet
		f.Suppression = diag.Data.Suppression

		// Merge original metadata with LSP-derived metadata.
		if len(diag.Data.Metadata) > 0 {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string, len(diag.Data.Metadata))
			}

			maps.Copy(f.Metadata, diag.Data.Metadata)
		}
	}

	// Preserve end position as Range when it differs from start.
	endLine := diag.Range.End.Line + 1

	endChar := diag.Range.End.Character + 1
	if endLine != startLine || endChar != startChar {
		f.Range = &Range{
			Start: f.Position,
			End:   Position{File: fileURI, Line: endLine, Column: endChar, Offset: -1},
		}
	}

	// Convert related information.
	var relatedIDs []string
	if diag.Data != nil {
		relatedIDs = diag.Data.RelatedFindingIDs
	}

	for i, rel := range diag.Related {
		relPos := Position{
			File:   FilePath(rel.Location.URI),
			Line:   rel.Location.Range.Start.Line + 1,
			Column: rel.Location.Range.Start.Character + 1,
			Offset: -1,
		}
		ref := RelatedRef{
			FindingID: GenerateID(ToolName(diag.Source), RuleName(diag.Code), relPos),
			Relation:  RelationKind(rel.Message),
			Position:  relPos,
		}

		// Restore original FindingID if available from Data.
		if i < len(relatedIDs) && relatedIDs[i] != "" {
			ref.FindingID = ID(relatedIDs[i])
		}

		endLine := rel.Location.Range.End.Line + 1
		endChar := rel.Location.Range.End.Character + 1

		if endLine != relPos.Line || endChar != relPos.Column {
			ref.Range = &Range{
				Start: ref.Position,
				End: Position{
					File:   FilePath(rel.Location.URI),
					Line:   endLine,
					Column: endChar,
					Offset: -1,
				},
			}
		}

		f.Related = append(f.Related, ref)
	}

	preserveLSPFidelity(&f, diag)

	return f
}

// preserveLSPFidelity stores raw LSP severity and diagnostic tags in the
// finding's metadata so conversions stay lossless.
func preserveLSPFidelity(f *Finding, diag LSPDiagnostic) {
	if diag.Severity <= 0 && len(diag.Tags) == 0 {
		return
	}

	if f.Metadata == nil {
		f.Metadata = make(map[string]string)
	}

	if diag.Severity > 0 {
		f.Metadata[LSPSeverityKey] = strconv.Itoa(int(diag.Severity))
	}

	if len(diag.Tags) > 0 {
		tagStrs := make([]string, len(diag.Tags))
		for i, tag := range diag.Tags {
			tagStrs[i] = strconv.Itoa(int(tag))
		}

		f.Metadata[LSPDiagnosticTagsKey] = strings.Join(tagStrs, ",")
	}
}

func severityToLSP(s Severity) LSPSeverity {
	switch s {
	case SeverityCritical, SeverityError:
		return LSPSeverityError
	case SeverityWarning:
		return LSPSeverityWarning
	case SeverityInfo:
		return LSPSeverityInfo
	default:
		return LSPSeverityWarning
	}
}

func severityFromLSP(sev LSPSeverity) Severity {
	switch sev {
	case LSPSeverityError:
		return SeverityError
	case LSPSeverityWarning:
		return SeverityWarning
	case LSPSeverityInfo, LSPSeverityHint:
		return SeverityInfo
	default:
		return SeverityWarning
	}
}

// parseLSPDiagnosticTags parses a comma-separated tag list as written by
// FromLSP into LSPDiagnosticTag values. Malformed entries and non-positive
// numbers (LSP tags start at 1) are skipped.
func parseLSPDiagnosticTags(s string) []LSPDiagnosticTag {
	parts := strings.Split(s, ",")

	tags := make([]LSPDiagnosticTag, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && n > 0 {
			tags = append(tags, LSPDiagnosticTag(n))
		}
	}

	if len(tags) == 0 {
		return nil
	}

	return tags
}

// collectRelatedFindingIDs extracts FindingIDs from related refs in order.
// Used to preserve original FindingIDs through LSP round-trip, since the LSP
// relatedInformation type has no field for arbitrary IDs.
func collectRelatedFindingIDs(refs []RelatedRef) []string {
	if len(refs) == 0 {
		return nil
	}

	ids := make([]string, len(refs))
	for i, rel := range refs {
		ids[i] = string(rel.FindingID)
	}

	return ids
}
