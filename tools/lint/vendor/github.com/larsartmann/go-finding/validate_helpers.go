package finding

// isValidLowercaseHyphen returns true if s is a non-empty string matching
// the lowercase-hyphenated convention (e.g., "security", "go-vet").
// The first rune must be a lowercase letter; subsequent runes may be
// lowercase letters, digits, or hyphens.
func isValidLowercaseHyphen(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if i == 0 {
			if r < 'a' || r > 'z' {
				return false
			}

			continue
		}

		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}

	return true
}
