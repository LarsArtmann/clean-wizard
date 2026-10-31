package finding

// Tag is a sub-classification label for a finding.
type Tag string

// Standard tag constants for common classification labels.
const (
	TagSecurity      Tag = "security"
	TagPerformance   Tag = "performance"
	TagStyle         Tag = "style"
	TagCorrectness   Tag = "correctness"
	TagBug           Tag = "bug"
	TagDeprecated    Tag = "deprecated"
	TagDocumentation Tag = "documentation"
	TagComplexity    Tag = "complexity"
	TagTest          Tag = "test"
	TagBuild         Tag = "build"
)

// IsStandard returns true if the tag is one of the predefined standard constants.
// Use IsStandard for allow-list filtering (e.g., "only show findings with known tags").
// Use IsValid for input validation (accepts any well-formed custom tag).
func (t Tag) IsStandard() bool {
	switch t {
	case TagSecurity, TagPerformance, TagStyle, TagCorrectness,
		TagBug, TagDeprecated, TagDocumentation, TagComplexity,
		TagTest, TagBuild:
		return true
	}

	return false
}

// IsValid returns true if the tag is a non-empty lowercase-hyphenated string.
// This rejects typos like "Security" or "SOME_TAG".
// Custom tags are valid as long as they match the format.
// Use IsValid for input validation (accepts custom values).
// Use IsStandard to check for predefined constants only (allow-list filtering).
func (t Tag) IsValid() bool {
	return isValidLowercaseHyphen(string(t))
}

// String returns the string representation of the tag.
func (t Tag) String() string {
	return string(t)
}
