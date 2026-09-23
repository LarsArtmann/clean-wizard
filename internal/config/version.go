package config

import (
	"fmt"
	"strconv"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
)

// CurrentFormatVersion is the configuration format version this binary reads
// and writes. It is deliberately decoupled from the binary version: format
// changes are rare and explicit, and every change must ship with a Migration
// entry so existing configurations keep loading.
//
// It is a var (not a const) so migration tests can simulate future formats;
// production code must treat it as immutable.
var CurrentFormatVersion = FormatVersion{Major: 1, Minor: 0, Patch: 0}

// FormatVersion identifies the on-disk configuration format. Configurations
// carry it in the top-level `version` field; the migration engine uses it to
// decide which Migration steps (if any) a file needs.
type FormatVersion struct {
	Major int
	Minor int
	Patch int
}

// ParseFormatVersion parses a "major.minor.patch" format version string.
func ParseFormatVersion(raw string) (FormatVersion, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return FormatVersion{}, errorfamily.NewRejection(
			"config.version",
			"invalid configuration format version "+strconv.Quote(raw)+`: want "major.minor.patch"`,
		)
	}

	nums := make([]int, len(parts))
	for i, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return FormatVersion{}, errorfamily.NewRejection(
				"config.version",
				"invalid configuration format version "+strconv.Quote(raw)+": "+part+" is not a non-negative integer",
			)
		}

		nums[i] = parsed
	}

	return FormatVersion{Major: nums[0], Minor: nums[1], Patch: nums[2]}, nil
}

// NormalizeFormatVersion maps the empty version (configurations written before
// version tracking existed) to the current version, matching the loader's
// backfill behavior. Any other value is parsed strictly.
func NormalizeFormatVersion(raw string) (FormatVersion, error) {
	if raw == "" {
		return CurrentFormatVersion, nil
	}

	return ParseFormatVersion(raw)
}

// String renders the version in its canonical "major.minor.patch" form.
func (v FormatVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns -1, 0, or 1 as v sorts before, equal to, or after other.
func (v FormatVersion) Compare(other FormatVersion) int {
	for _, pair := range [][2]int{
		{v.Major, other.Major},
		{v.Minor, other.Minor},
		{v.Patch, other.Patch},
	} {
		switch {
		case pair[0] < pair[1]:
			return -1
		case pair[0] > pair[1]:
			return 1
		}
	}

	return 0
}

// LessThan reports whether v sorts strictly before other.
func (v FormatVersion) LessThan(other FormatVersion) bool {
	return v.Compare(other) < 0
}
