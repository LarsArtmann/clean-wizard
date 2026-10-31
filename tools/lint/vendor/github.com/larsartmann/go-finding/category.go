package finding

import (
	"cmp"
	"errors"
	"fmt"
)

// Category classifies the domain of a finding.
type Category string

// Standard category constants for findings.
const (
	CategorySecurity      Category = "security"
	CategoryStyle         Category = "style"
	CategoryPerformance   Category = "performance"
	CategoryCorrectness   Category = "correctness"
	CategoryComplexity    Category = "complexity"
	CategoryDuplication   Category = "duplication"
	CategoryErrorHandling Category = "error-handling"
	CategoryMigration     Category = "migration"
	CategoryTypeSafety    Category = "type-safety"
	CategoryStructure     Category = "structure"
	CategoryConfiguration Category = "configuration"
	CategoryDocumentation Category = "documentation"
	CategoryTesting       Category = "testing"
	CategoryUnused        Category = "unused"
	CategoryBestPractice  Category = "best-practice"
	CategoryNaming        Category = "naming"
)

// IsStandard returns true if the category is one of the predefined standard constants.
// Use IsStandard for allow-list filtering (e.g., "only show findings in known categories").
// Use IsValid for input validation (accepts any well-formed custom category).
func (c Category) IsStandard() bool {
	switch c {
	case CategorySecurity, CategoryStyle, CategoryPerformance, CategoryCorrectness,
		CategoryComplexity, CategoryDuplication, CategoryErrorHandling, CategoryMigration,
		CategoryTypeSafety, CategoryStructure, CategoryConfiguration, CategoryDocumentation,
		CategoryTesting, CategoryUnused, CategoryBestPractice, CategoryNaming:
		return true
	}

	return false
}

// IsValid returns true if the category is a non-empty string matching the
// lowercase-hyphenated convention (e.g., "security", "go-vet").
// This rejects typos like "Security" or "SOME_CATEGORY".
// Use IsValid for input validation (accepts custom categories).
// Use IsStandard to check for predefined constants only (allow-list filtering).
func (c Category) IsValid() bool {
	return isValidLowercaseHyphen(string(c))
}

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}

// IsSecurity reports whether the category is security-related.
func (c Category) IsSecurity() bool {
	return c == CategorySecurity
}

// Compare returns -1, 0, or +1 depending on whether c is less than, equal to,
// or greater than other. Categories have no inherent priority ordering, so the
// comparison is lexicographic by string value, providing a stable total ordering
// suitable for deterministic sorting.
func (c Category) Compare(other Category) int {
	return cmp.Compare(string(c), string(other))
}

// errInvalidCategory is returned when parsing an invalid category string.
var errInvalidCategory = errors.New("invalid category")

// ParseCategory parses a string into a Category.
// It accepts the standard category names (e.g., "security", "style", "correctness").
// Returns an error if the string is not a valid category.
func ParseCategory(s string) (Category, error) {
	cat := Category(s)
	if cat.IsValid() {
		return cat, nil
	}

	return "", fmt.Errorf("%w: %q must match ^[a-z][a-z0-9-]*$", errInvalidCategory, s)
}

// MustParseCategory parses a string into a Category, panicking on invalid input.
func MustParseCategory(s string) Category {
	return must(ParseCategory(s))
}
