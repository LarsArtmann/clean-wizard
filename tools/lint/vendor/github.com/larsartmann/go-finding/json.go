package finding

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"slices"
)

// marshalOpts ensures deterministic map key ordering in all JSON output.
// All production json.Marshal / json.MarshalWrite calls must include this option
// to guarantee reproducible output (required for snapshot/diff stability).
var marshalOpts = json.Deterministic(true)

// prettyMarshalOpts ensures deterministic, indented JSON output.
// Used by PrettyJSON, PrettyJSONFiltered, WriteJSON, ToSARIFWithOpts, WriteSARIFWithOpts.
var prettyMarshalOpts = []json.Options{
	json.Deterministic(true),
	jsontext.WithIndentPrefix(""),
	jsontext.WithIndent("  "),
}

// Sentinel errors for JSON validation.
var (
	ErrInvalidFinding = errors.New("invalid finding: missing required fields")
	ErrInvalidReport  = errors.New("invalid report: missing tool name")
)

// reportJSON is the JSON representation of Report. It exists because
// Report.findings is unexported for thread safety; encoding/json cannot
// access unexported fields.
type reportJSON struct {
	Tool     ToolInfo  `json:"tool"`
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"`
}

// MarshalJSON implements json.Marshaler.
func (r *Report) MarshalJSON() ([]byte, error) {
	data, err := withReadLockErr(r, func() ([]byte, error) {
		return json.Marshal(reportJSON{
			Tool:     r.Tool,
			Findings: r.findings,
			Summary:  r.Summary,
		}, marshalOpts)
	})
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler.
// Safe for concurrent use — acquires a write lock.
func (r *Report) UnmarshalJSON(data []byte) error {
	var dto reportJSON

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return fmt.Errorf("unmarshal report: %w", err)
	}

	withLock(r, func() struct{} {
		r.Tool = dto.Tool
		r.findings = dto.Findings
		r.Summary = dto.Summary

		return struct{}{}
	})

	return nil
}

// IsInvalid returns true if the finding is invalid (has missing required fields).
// Intended for use with [slices.DeleteFunc].
func IsInvalid(f Finding) bool {
	return !f.IsValid()
}

// FilterInvalid is a deprecated alias for [IsInvalid]. The name was misleading:
// it returns true when the finding IS invalid (should be filtered out), not
// when it should be kept.
//
// Deprecated: Use [IsInvalid] instead.
func FilterInvalid(f Finding) bool {
	return IsInvalid(f)
}

// marshalJSONString converts a marshal result (bytes, err) into a string,
// wrapping any marshaling error with consistent context.
func marshalJSONString(bytes []byte, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}

	return string(bytes), nil
}

// JSON returns a compact JSON representation of the report.
// Shorthand for MarshalJSON that returns a string.
func (r *Report) JSON() (string, error) {
	return marshalJSONString(r.MarshalJSON())
}

// PrettyJSON returns a formatted JSON representation of the report.
// Includes all findings, including suppressed ones.
func (r *Report) PrettyJSON() (string, error) {
	return marshalJSONString(json.Marshal(r, prettyMarshalOpts...))
}

// PrettyJSONFiltered returns a formatted JSON representation with only
// active (non-suppressed) findings. Unlike PrettyJSON, this excludes
// suppressed findings from the output.
func (r *Report) PrettyJSONFiltered() (string, error) {
	filtered := withReadLock(r, func() *Report {
		result := &Report{
			Tool:     r.Tool,
			findings: make([]Finding, 0, len(r.findings)),
			Summary:  Summary{},
		}
		for _, f := range r.findings {
			if !f.IsSuppressed() {
				result.findings = append(result.findings, f)
			}
		}

		return result
	})

	filtered.ComputeSummary()

	return marshalJSONString(json.Marshal(filtered, prettyMarshalOpts...))
}

// FromJSON parses a Finding from JSON and validates required fields.
func FromJSON(data []byte) (Finding, error) {
	var f Finding

	err := json.Unmarshal(data, &f)
	if err != nil {
		return Finding{}, fmt.Errorf("unmarshal finding: %w", err)
	}

	err = f.Validate()
	if err != nil {
		return Finding{}, fmt.Errorf("%w: %w", ErrInvalidFinding, err)
	}

	return f, nil
}

// ReportFromJSON parses a Report from JSON and validates required fields.
// Invalid findings are silently dropped. Use the returned count to detect data loss.
func ReportFromJSON(data []byte) (*Report, int, error) {
	var r Report

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal report: %w", err)
	}

	if r.Tool.Name == "" {
		return nil, 0, ErrInvalidReport
	}

	before := len(r.findings)
	r.findings = slices.DeleteFunc(r.findings, FilterInvalid)

	return &r, before - len(r.findings), nil
}

// FindingsFromJSON parses a slice of Findings from JSON and validates each one.
// Invalid findings are silently dropped. Use the returned count to detect data loss.
func FindingsFromJSON(data []byte) ([]Finding, int, error) {
	var findings []Finding

	err := json.Unmarshal(data, &findings)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal findings: %w", err)
	}

	before := len(findings)
	findings = slices.DeleteFunc(findings, FilterInvalid)

	return findings, before - len(findings), nil
}

// LineJSON returns compact JSON (single line).
func (f Finding) LineJSON() (string, error) {
	return marshalJSONString(json.Marshal(f, marshalOpts))
}

// WriteJSON writes compact JSON directly to w.
// Avoids the intermediate string allocation of LineJSON.
func (f Finding) WriteJSON(w io.Writer) error {
	err := json.MarshalWrite(w, f, marshalOpts)
	if err != nil {
		return fmt.Errorf("encoding finding JSON: %w", err)
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("writing trailing newline: %w", err)
	}

	return nil
}

// WriteJSON writes pretty-printed JSON directly to w.
// Avoids the intermediate string allocation of PrettyJSON.
// Safe for concurrent use.
func (r *Report) WriteJSON(w io.Writer) error {
	err := json.MarshalWrite(w, r, prettyMarshalOpts...)
	if err != nil {
		return fmt.Errorf("encoding report JSON: %w", err)
	}

	return nil
}
