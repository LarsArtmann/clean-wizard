package finding

import "maps"

// Builder provides a fluent API for constructing Finding values.
// Use NewBuilder with the required fields, then chain With* methods
// for optional fields, and call Build to obtain the result.
//
// Example:
//
//	f := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, Pos("main.go", 42, 5)).
//		WithFixStrategy(FixStrategyDirect).
//		WithBeforeCode("x.foo").
//		WithAfterCode("x.foo()").
//		Build()
type Builder struct {
	f Finding
}

// NewBuilder creates a builder seeded with the required fields.
// The ID is auto-generated from the provided arguments.
// Confidence defaults to ConfidenceFull (1.0), appropriate for deterministic
// static analysis. Override with WithConfidence if needed.
func NewBuilder(rule RuleName, toolName ToolName, message string, severity Severity, pos Position) *Builder {
	return &Builder{f: NewFinding(rule, toolName, message, severity, pos, ConfidenceFull)}
}

// WithID overrides the auto-generated ID.
func (b *Builder) WithID(id ID) *Builder {
	b.f.ID = id

	return b
}

// WithCategory sets the category.
func (b *Builder) WithCategory(cat Category) *Builder {
	b.f.Category = cat

	return b
}

// WithTags sets multiple tags.
func (b *Builder) WithTags(tags ...Tag) *Builder {
	b.f.Tags = append(b.f.Tags, tags...)

	return b
}

// WithFixStrategy sets the fix strategy.
func (b *Builder) WithFixStrategy(fs FixStrategy) *Builder {
	b.f.FixStrategy = fs

	return b
}

// WithSuggestion sets the human-readable fix suggestion.
func (b *Builder) WithSuggestion(s string) *Builder {
	b.f.Suggestion = s

	return b
}

// WithBeforeCode sets the code before the fix.
func (b *Builder) WithBeforeCode(code string) *Builder {
	b.f.BeforeCode = code

	return b
}

// WithAfterCode sets the code after the fix.
func (b *Builder) WithAfterCode(code string) *Builder {
	b.f.AfterCode = code

	return b
}

// WithEdits appends typed text edits to the fix. When set, the edit list is
// the authoritative fix representation; BeforeCode/AfterCode remain the
// display summary of the first edit.
func (b *Builder) WithEdits(edits ...TextEdit) *Builder {
	b.f.Edits = append(b.f.Edits, edits...)

	return b
}

// WithRange sets the source range.
func (b *Builder) WithRange(r Range) *Builder {
	b.f.Range = &r

	return b
}

// WithSnippet sets the surrounding code context.
func (b *Builder) WithSnippet(s string) *Builder {
	b.f.Snippet = s

	return b
}

// WithGroupID sets the logical group this finding belongs to
// (e.g., a clone group identifier).
func (b *Builder) WithGroupID(g GroupID) *Builder {
	b.f.GroupID = g

	return b
}

// WithConfidence sets the confidence level (clamped to [0.0, 1.0]).
func (b *Builder) WithConfidence(c Confidence) *Builder {
	b.f.Confidence = c.Clamp()

	return b
}

// WithRelated appends related references.
func (b *Builder) WithRelated(refs ...RelatedRef) *Builder {
	b.f.Related = append(b.f.Related, refs...)

	return b
}

// WithSuppression sets the suppression info.
func (b *Builder) WithSuppression(s Suppression) *Builder {
	b.f.Suppression = &s

	return b
}

// WithMetadata copies the given metadata into the finding.
func (b *Builder) WithMetadata(m map[string]string) *Builder {
	if b.f.Metadata == nil {
		b.f.Metadata = make(map[string]string, len(m))
	}

	maps.Copy(b.f.Metadata, m)

	return b
}

// Build returns the constructed Finding.
// Returns a detailed validation error if required fields are missing or invalid.
// FixStrategy is normalized: empty string becomes FixStrategyNone.
func (b *Builder) Build() (Finding, error) {
	b.f = b.f.Normalized()

	err := b.f.Validate()
	if err != nil {
		return Finding{}, err
	}

	return b.f.Clone(), nil
}

// MustBuild returns the constructed Finding or panics if required fields are missing.
// Use this only when the builder is fully configured and invalid state is a programmer error.
func (b *Builder) MustBuild() Finding {
	return must(b.Build())
}

// BuildOrDefault returns the constructed Finding, or a zero-value Finding{} if
// validation fails. This eliminates the error-swallowing boilerplate
// (SafeBuildFinding / buildFinding) that consumers universally reinvent.
// Use Build when you need to handle validation errors explicitly.
func (b *Builder) BuildOrDefault() Finding {
	f, err := b.Build()
	if err != nil {
		return Finding{}
	}

	return f
}

// Template is a pre-configured builder factory: stamp common fields
// (tool name, category, fix strategy, tags) once, then build many findings
// with varying rule/message/severity/position. This eliminates the
// newMigrationFinding / buildFixableFinding / IssueBuilderFactory patterns
// that consumers reinvent for batch finding creation.
type Template struct {
	Tool        ToolName
	Category    Category
	FixStrategy FixStrategy
	Tags        []Tag
	GroupID     GroupID
}

// NewTemplate creates a Template with the given tool name.
// Chain WithCategory, WithFixStrategy, WithTags to configure common fields,
// then call Build for each finding.
func NewTemplate(toolName ToolName) *Template {
	return &Template{Tool: toolName}
}

// WithCategory sets the category on the template.
func (t *Template) WithCategory(cat Category) *Template {
	t.Category = cat

	return t
}

// WithFixStrategy sets the fix strategy on the template.
func (t *Template) WithFixStrategy(fs FixStrategy) *Template {
	t.FixStrategy = fs

	return t
}

// WithTags sets tags on the template. These are stamped onto every finding
// built from this template.
func (t *Template) WithTags(tags ...Tag) *Template {
	t.Tags = append(t.Tags, tags...)

	return t
}

// WithGroupID sets the logical group on the template. Every finding built
// from it joins the same group (e.g. N occurrences of one cloned block).
// Callers that need per-finding groups should use [Builder.WithGroupID]
// on the per-finding builder instead of a template-level stamp.
func (t *Template) WithGroupID(g GroupID) *Template {
	t.GroupID = g

	return t
}

// Builder creates a pre-configured [*Builder] from the template, stamping the
// pre-configured tool name, category, fix strategy, and tags. Unlike [Build],
// which returns a final [Finding], Builder returns the intermediate [*Builder]
// so the caller can chain additional per-finding fields (confidence, suggestion,
// before/after code, metadata, etc.) before calling [Builder.Build],
// [Builder.MustBuild], or [Builder.BuildOrDefault].
//
// This eliminates the per-consumer factory wrapper (e.g. makeFindingWithConfidence)
// that every linter reinvents when it needs both template-level defaults AND
// per-finding confidence/suggestion:
//
//	tmpl := finding.NewTemplate("my-linter").
//	    WithCategory(finding.CategoryStyle).
//	    WithFixStrategy(finding.FixStrategySuggest)
//
//	f := tmpl.Builder(rule, msg, finding.SeverityWarning, pos).
//	    WithConfidence(finding.ConfidenceHigh).
//	    WithSuggestion("use foo.Bar() instead").
//	    MustBuild()
func (t *Template) Builder(rule RuleName, message string, severity Severity, pos Position) *Builder {
	b := NewBuilder(rule, t.Tool, message, severity, pos)

	if t.Category != "" {
		b = b.WithCategory(t.Category)
	}

	if t.FixStrategy != "" {
		b = b.WithFixStrategy(t.FixStrategy)
	}

	if len(t.Tags) > 0 {
		b = b.WithTags(t.Tags...)
	}

	if t.GroupID != "" {
		b = b.WithGroupID(t.GroupID)
	}

	return b
}

// Build creates a Finding from the template, stamping the pre-configured
// tool name, category, fix strategy, and tags. Returns a zero-value Finding
// if validation fails (delegates to Builder.BuildOrDefault).
//
// For per-finding confidence, suggestion, or other overrides, use [Template.Builder]
// instead — it returns a [*Builder] for further chaining before terminal Build.
func (t *Template) Build(rule RuleName, message string, severity Severity, pos Position) Finding {
	return t.Builder(rule, message, severity, pos).BuildOrDefault()
}
