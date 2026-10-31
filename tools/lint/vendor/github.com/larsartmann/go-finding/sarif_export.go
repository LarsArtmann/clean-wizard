package finding

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
)

// SARIFOption configures SARIF export behavior.
type SARIFOption func(*sarifExportConfig)

type sarifExportConfig struct {
	includeSuppressed bool
	minSeverity       Severity
}

// WithIncludeSuppressed causes SARIF export to include suppressed findings
// with their suppression metadata emitted as SARIF suppression entries,
// rather than dropping them entirely. This enables full round-trip fidelity
// for suppression data through SARIF export→import.
func WithIncludeSuppressed() SARIFOption {
	return func(c *sarifExportConfig) { c.includeSuppressed = true }
}

// WithMinSeverity filters findings below the given severity level.
func WithMinSeverity(sev Severity) SARIFOption {
	return func(c *sarifExportConfig) { c.minSeverity = sev }
}

func defaultSarifConfig() sarifExportConfig {
	return sarifExportConfig{
		includeSuppressed: false,
		minSeverity:       SeverityInfo,
	}
}

func sarifResultsFromFindings(findings []Finding, cfg sarifExportConfig) []sarifResult {
	results := make([]sarifResult, 0, len(findings))

	for _, f := range findings {
		if f.Severity.LessThan(cfg.minSeverity) {
			continue
		}

		if f.IsSuppressed() && !cfg.includeSuppressed {
			continue
		}

		results = append(results, findingToSARIF(f))
	}

	return results
}

func sarifDriverFromReport(r *Report) sarifDriver {
	return sarifDriver{Name: r.Tool.Name, Version: r.Tool.Version}
}

// ToSARIF converts a Report to SARIF 2.1.0 format.
// Suppressed findings are excluded by default.
//
// For full round-trip fidelity including suppression data, use ToSARIFWithOpts(WithIncludeSuppressed()).
func (r *Report) ToSARIF() ([]byte, error) {
	return r.ToSARIFWithOpts()
}

// ToSARIFWithOpts converts a Report to SARIF 2.1.0 format with the given options.
// See [WithIncludeSuppressed] and [WithMinSeverity].
func (r *Report) ToSARIFWithOpts(opts ...SARIFOption) ([]byte, error) {
	cfg := defaultSarifConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	log := r.buildsarifLog(sarifResultsFromFindings(r.readFindings(), cfg))

	data, err := json.Marshal(log, prettyMarshalOpts...)
	if err != nil {
		return nil, fmt.Errorf("marshaling SARIF: %w", err)
	}

	return data, nil
}

// WriteSARIF writes the report in SARIF 2.1.0 format directly to w.
// Streams via json.Encoder, avoiding the intermediate []byte buffer of ToSARIF.
// The context is checked for cancellation before encoding begins.
// Suppressed findings are excluded by default.
//
// For full round-trip fidelity including suppression data, use WriteSARIFWithOpts(w, WithIncludeSuppressed()).
func (r *Report) WriteSARIF(ctx context.Context, w io.Writer) error {
	return r.WriteSARIFWithOpts(ctx, w)
}

// WriteSARIFWithOpts writes the report in SARIF 2.1.0 format with the given options.
// See [WithIncludeSuppressed] and [WithMinSeverity].
func (r *Report) WriteSARIFWithOpts(ctx context.Context, w io.Writer, opts ...SARIFOption) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("writing SARIF: %w", err)
	}

	cfg := defaultSarifConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	err = json.MarshalWrite(
		w,
		r.buildsarifLog(sarifResultsFromFindings(r.readFindings(), cfg)),
		prettyMarshalOpts...,
	)
	if err != nil {
		return fmt.Errorf("encoding SARIF: %w", err)
	}

	return nil
}

// WriteTo writes the report in SARIF 2.1.0 format to w and returns the bytes written.
// Implements io.WriterTo, enabling use with io.Copy for streaming SARIF output.
//
// For context-aware cancellation, prefer WriteSARIF directly.
func (r *Report) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	err := r.WriteSARIF(context.Background(), cw)
	if err != nil {
		return cw.n, fmt.Errorf("writing SARIF: %w", err)
	}

	return cw.n, nil
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)

	return n, err //nolint:wrapcheck // passthrough writer — wrapping would be misleading
}

func (r *Report) buildsarifLog(results []sarifResult) sarifLog {
	return sarifLog{
		Version: sarifVersion,
		Schema:  sarifSchema,
		Runs: []sarifRun{
			{
				Tool:    sarifTool{Driver: sarifDriverFromReport(r)},
				Results: results,
			},
		},
	}
}

func findingToSARIF(f Finding) sarifResult {
	result := sarifResult{
		RuleID:       string(f.Rule),
		Level:        severityToSARIFLevel(f.Severity),
		Message:      sarifMessage{Text: f.Message},
		Locations:    sarifLocations(f),
		Fixes:        sarifFixes(f),
		Related:      sarifRelatedLocs(f),
		Properties:   sarifProperties(f),
		Suppressions: sarifSuppressions(f),
	}

	if f.Confidence > 0 {
		result.Rank = float64(f.NormalizedConfidence()) * sarifConfidenceScale
	}

	return result
}

// sarifSuppressions converts a Finding's Suppression to SARIF suppression entries.
// Returns nil if the finding is not suppressed.
func sarifSuppressions(f Finding) []sarifSuppression {
	if f.Suppression == nil {
		return nil
	}

	return []sarifSuppression{{
		Kind:          suppressionKindToSARIF(f.Suppression.Kind),
		Status:        suppressionStatusToSARIF(f.Suppression.Kind),
		Justification: f.Suppression.Reason,
	}}
}

// suppressionKindToSARIF maps go-finding SuppressionKind to SARIF suppression kind.
func suppressionKindToSARIF(kind SuppressionKind) string {
	switch kind {
	case SuppressionInSource:
		return sarifSuppressionKindSource
	case SuppressionInConfig, SuppressionInReview:
		return sarifSuppressionKindExternal
	default:
		return sarifSuppressionKindSource
	}
}

// suppressionStatusToSARIF maps go-finding SuppressionKind to SARIF suppression status.
func suppressionStatusToSARIF(kind SuppressionKind) string {
	switch kind {
	case SuppressionInReview:
		return sarifSuppressionStatusReview
	case SuppressionInSource, SuppressionInConfig:
		return sarifSuppressionStatusAccepted
	default:
		return sarifSuppressionStatusAccepted
	}
}

func sarifLocations(f Finding) []sarifLocation {
	return []sarifLocation{{
		PhysicalLocation: sarifPhysicalLocation{
			ArtifactLocation: sarifArtifactLocation{URI: string(f.Position.File)},
			Region:           findingRegion(f),
		},
	}}
}

func findingRegion(f Finding) *sarifRegion {
	region := &sarifRegion{
		StartLine:   f.Position.Line,
		StartColumn: f.Position.Column,
	}

	if f.Range != nil && f.Range.HasEnd() {
		region.EndLine = f.Range.End.Line
		region.EndColumn = f.Range.End.Column
	}

	if f.Snippet != "" {
		region.Snippet = &sarifArtifactContent{Text: f.Snippet}
	}

	return region
}

func findingFixRegion(f Finding) sarifRegion {
	region := sarifRegion{
		StartLine:   f.Position.Line,
		StartColumn: f.Position.Column,
	}

	if f.Range != nil && f.Range.HasEnd() {
		region.EndLine = f.Range.End.Line
		region.EndColumn = f.Range.End.Column
	} else {
		region.EndLine = f.Position.Line
		region.EndColumn = f.Position.Column
	}

	return region
}

func sarifFixes(f Finding) []sarifFix {
	if len(f.Edits) > 0 {
		return []sarifFix{{
			Description: sarifMessage{Text: f.Suggestion},
			Changes:     sarifArtifactChangesFromEdits(f),
		}}
	}

	if f.HasFix() {
		region := findingFixRegion(f)

		return []sarifFix{{
			Description: sarifMessage{Text: f.Suggestion},
			Changes: []sarifArtifactChange{{
				ArtifactLocation: sarifArtifactLocation{URI: string(f.Position.File)},
				Replacements: []sarifReplacement{{
					DeletedRegion: region,
					InsertedText:  sarifMessage{Text: f.AfterCode},
				}},
			}},
		}}
	}

	if f.HasSuggestion() {
		return []sarifFix{{
			Description: sarifMessage{Text: f.Suggestion},
		}}
	}

	return nil
}

// sarifArtifactChangesFromEdits groups the finding's typed edit list into one
// artifact change per target file. Edits with an empty Start.File belong to
// the finding's own file.
func sarifArtifactChangesFromEdits(f Finding) []sarifArtifactChange {
	byFile := make(map[FilePath][]sarifReplacement)
	order := make([]FilePath, 0, len(f.Edits))

	for _, edit := range f.Edits {
		file := edit.EffectiveFile(f.Position.File)
		if _, seen := byFile[file]; !seen {
			order = append(order, file)
		}

		byFile[file] = append(byFile[file], sarifReplacement{
			DeletedRegion: sarifEditRegion(edit),
			InsertedText:  sarifMessage{Text: edit.NewText},
		})
	}

	changes := make([]sarifArtifactChange, 0, len(order))
	for _, file := range order {
		changes = append(changes, sarifArtifactChange{
			ArtifactLocation: sarifArtifactLocation{URI: string(file)},
			Replacements:     byFile[file],
		})
	}

	return changes
}

// sarifEditRegion renders a TextEdit span as a SARIF region, carrying line and
// byte coordinates so the round-trip preserves both.
func sarifEditRegion(edit TextEdit) sarifRegion {
	region := sarifRegion{
		StartLine:   edit.Start.Line,
		StartColumn: edit.Start.Column,
	}

	if edit.Start.HasOffset() {
		region.ByteOffset = sarifIntPtr(edit.Start.Offset)
	}

	if !edit.HasSpan() {
		return region
	}

	region.EndLine = edit.End.Line
	region.EndColumn = edit.End.Column

	if edit.Start.HasOffset() && edit.End.HasOffset() {
		region.ByteLength = sarifIntPtr(max(edit.End.Offset-edit.Start.Offset, 0))
	}

	return region
}

// sarifIntPtr returns a pointer to v, or nil for negative v (the offset
// "unset" sentinel).
func sarifIntPtr(v int) *int {
	if v < 0 {
		return nil
	}

	return &v
}

func sarifRelatedLocs(f Finding) []sarifRelatedLoc {
	if len(f.Related) == 0 {
		return nil
	}

	related := make([]sarifRelatedLoc, 0, len(f.Related))

	for _, rel := range f.Related {
		region := &sarifRegion{
			StartLine:   rel.Position.Line,
			StartColumn: rel.Position.Column,
		}
		if rel.Range != nil && rel.Range.HasEnd() {
			region.EndLine = rel.Range.End.Line
			region.EndColumn = rel.Range.End.Column
		}

		sarifRel := sarifRelatedLoc{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: string(rel.Position.File)},
				Region:           region,
			},
			Message: sarifMessage{Text: string(rel.Relation)},
		}

		if rel.FindingID != "" {
			sarifRel.Properties = map[string]any{sarifPropID: string(rel.FindingID)}
		}

		related = append(related, sarifRel)
	}

	return related
}

func sarifProperties(f Finding) map[string]any {
	props := map[string]any{
		sarifPropID:          string(f.ID),
		sarifPropSeverity:    string(f.Severity),
		sarifPropFixStrategy: string(f.FixStrategy),
		sarifPropToolName:    string(f.ToolName),
	}

	if f.Category != "" {
		props[sarifPropCategory] = string(f.Category)
	}

	if len(f.Tags) > 0 {
		props[sarifPropTags] = f.Tags
	}

	if f.Confidence > 0 {
		props[sarifPropConfidence] = float64(f.NormalizedConfidence())
	}

	if f.Suggestion != "" {
		props[sarifPropSuggestion] = f.Suggestion
	}

	if f.Snippet != "" {
		props[sarifPropSnippet] = f.Snippet
	}

	if f.GroupID != "" {
		props[sarifPropGroupID] = string(f.GroupID)
	}

	if f.BeforeCode != "" {
		props[sarifPropBeforeCode] = f.BeforeCode
	}

	if f.AfterCode != "" {
		props[sarifPropAfterCode] = f.AfterCode
	}

	if f.Position.HasOffset() {
		props[sarifPropStartOffset] = f.Position.Offset
	}

	if f.Range != nil && f.Range.End.HasOffset() {
		props[sarifPropEndOffset] = f.Range.End.Offset
	}

	if f.Suppression != nil {
		props[sarifPropSuppressionKind] = string(f.Suppression.Kind)

		if f.Suppression.Rule != "" {
			props[sarifPropSuppressionRule] = string(f.Suppression.Rule)
		}

		if f.Suppression.Reason != "" {
			props[sarifPropSuppressionReason] = f.Suppression.Reason
		}

		if f.Suppression.ExpiresAt != nil {
			props[sarifPropSuppressionExpiry] = f.Suppression.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}
	}

	for k, v := range f.Metadata {
		props[sarifPropMetaPrefix+k] = v
	}

	return props
}
