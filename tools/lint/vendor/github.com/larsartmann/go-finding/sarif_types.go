package finding

import (
	"encoding/json/v2"
	"errors"
	"fmt"
)

// SARIF types for Report generation.
// These are simplified representations of SARIF 2.1.0.

// SARIF confidence scale: Confidence is 0-1, SARIF rank is 0-100.
const sarifConfidenceScale = 100.0

const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
)

// SARIF suppression kind/status string constants (SARIF 2.1.0 §3.27.6-3.27.7).
const (
	sarifSuppressionKindSource     = "inSource"
	sarifSuppressionKindExternal   = "inExternalConfiguration"
	sarifSuppressionStatusAccepted = "accepted"
	sarifSuppressionStatusReview   = "underReview"
)

const (
	sarifPropID          = "go-finding/id"
	sarifPropSeverity    = "go-finding/severity"
	sarifPropFixStrategy = "go-finding/fixStrategy"
	sarifPropToolName    = "go-finding/toolName"
	sarifPropCategory    = "go-finding/category"

	sarifPropTags       = "go-finding/tags"
	sarifPropConfidence = "go-finding/confidence"
	sarifPropSuggestion = "go-finding/suggestion"
	sarifPropSnippet    = "go-finding/snippet"
	sarifPropGroupID    = "go-finding/groupId"
	sarifPropBeforeCode = "go-finding/beforeCode"
	sarifPropAfterCode  = "go-finding/afterCode"
	sarifPropEditPrefix = "go-finding/edit/"
	sarifPropPrefix     = "go-finding/"
	sarifPropMetaPrefix = "go-finding/meta/"

	sarifPropSuppressionKind   = "go-finding/suppression-kind"
	sarifPropSuppressionRule   = "go-finding/suppression-rule"
	sarifPropSuppressionReason = "go-finding/suppression-reason"
	sarifPropSuppressionExpiry = "go-finding/suppression-expiry"

	sarifPropStartOffset = "go-finding/start-offset"
	sarifPropEndOffset   = "go-finding/end-offset"
)

// sarifLog represents a SARIF log file containing run results.
type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun represents a single analysis run in a SARIF log.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool defines the static analysis tool that generated the results.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver represents the main driver tool with version information.
type sarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// sarifResult represents a single finding in SARIF format.
//
// Doctrine boundary: Properties uses map[string]any because the SARIF 2.1.0
// specification requires propertyBag.properties to accept arbitrary JSON values.
// This is the ONLY place in go-finding where map[string]any is used; the public
// Finding.Metadata type is map[string]string by design (see finding.go).
// The coercion between these two representations is centralized in
// sarifProperties() (export) and applySarifProperties() (import).
type sarifResult struct {
	RuleID       string             `json:"ruleId"`
	Level        string             `json:"level"`
	Message      sarifMessage       `json:"message"`
	Locations    []sarifLocation    `json:"locations"`
	Fixes        []sarifFix         `json:"fixes,omitempty"`
	Related      []sarifRelatedLoc  `json:"relatedLocations,omitempty"`
	Suppressions []sarifSuppression `json:"suppressions,omitempty"`
	Rank         float64            `json:"rank,omitempty"`
	Properties   map[string]any     `json:"properties,omitempty"`
}

// sarifSuppression represents a suppression entry on a SARIF result.
// SARIF 2.1.0 §3.27.
type sarifSuppression struct {
	Kind          string `json:"kind"`                    // "inSource" or "inExternalConfiguration"
	Status        string `json:"status,omitempty"`        // "accepted" or "underReview"
	Justification string `json:"justification,omitempty"` // Why it's suppressed
}

// sarifMessage represents a message in SARIF format.
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifLocation represents a location in SARIF format.
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation represents physical details of a location.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

// sarifArtifactLocation represents the artifact URI.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifArtifactContent represents the SARIF artifactContent object,
// used for region.snippet per SARIF 2.1.0 §3.30.13.
type sarifArtifactContent struct {
	Text string `json:"text,omitempty"`
}

// UnmarshalJSON accepts both the spec-compliant object form ({"text":"..."})
// and the common bare-string shorthand ("snippet"), ensuring backward
// compatibility with SARIF producers that emit bare strings.
func (s *sarifArtifactContent) UnmarshalJSON(data []byte) error {
	// Try object form first (anonymous struct avoids infinite recursion).
	var obj struct {
		Text string `json:"text,omitempty"`
	}

	objErr := json.Unmarshal(data, &obj)
	if objErr == nil {
		s.Text = obj.Text

		return nil
	}

	// Fall back to bare string.
	strErr := json.Unmarshal(data, &s.Text)
	if strErr != nil {
		return fmt.Errorf("unmarshal snippet: %w", errors.Join(objErr, strErr))
	}

	return nil
}

// sarifRegion represents a code region in a text document.
//
// Snippet is a *sarifArtifactContent per SARIF 2.1.0 §3.30.13. The snippet
// value is also carried in the go-finding/snippet property for lossless
// round-trip. See ADR #9.
type sarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
	// ByteOffset/ByteLength are pointers so byte offset 0 (start of file) is
	// distinguishable from "unset" under omitempty.
	ByteOffset *int                  `json:"byteOffset,omitempty"`
	ByteLength *int                  `json:"byteLength,omitempty"`
	Snippet    *sarifArtifactContent `json:"snippet,omitempty"`
}

// sarifFix represents a fix to be applied to the artifact.
type sarifFix struct {
	Description sarifMessage          `json:"description"`
	Changes     []sarifArtifactChange `json:"artifactChanges"`
}

// sarifArtifactChange represents a change to an artifact.
type sarifArtifactChange struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Replacements     []sarifReplacement    `json:"replacements"`
}

// sarifReplacement represents a replacement of text in an artifact.
type sarifReplacement struct {
	DeletedRegion sarifRegion  `json:"deletedRegion"`
	InsertedText  sarifMessage `json:"insertedText"`
}

// sarifRelatedLoc represents a related location in SARIF.
type sarifRelatedLoc struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
	Message          sarifMessage          `json:"message"`
	Properties       map[string]any        `json:"properties,omitempty"`
}

// severityToSARIFLevel converts a Severity to a SARIF level string.
//
// Known limitation: SeverityCritical maps to "error" because SARIF 2.1.0 does not
// have a "critical" level. The original severity is preserved in the result's
// Properties["go-finding/severity"] for round-trip fidelity. Use FromSARIFLevel
// only when Properties are not available; otherwise prefer reading the property.
func severityToSARIFLevel(
	s Severity,
) string {
	switch s {
	case SeverityInfo:
		return "note"
	case SeverityWarning:
		return "warning"
	case SeverityError, SeverityCritical:
		return "error"
	default:
		return "warning"
	}
}

// FromSARIFLevel converts a SARIF level back to Severity.
// Lossy: both SeverityCritical and SeverityError map to SARIF "error",
// so FromSARIFLevel("error") returns SeverityError. For full fidelity,
// read the "go-finding/severity" property from the result instead.
func FromSARIFLevel(level string) Severity {
	switch level {
	case "note":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityWarning
	}
}
