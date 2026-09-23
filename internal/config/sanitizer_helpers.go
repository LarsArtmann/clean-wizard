package config

// sanitizePathValue applies the configured string normalizations (trim
// whitespace, clean path) to a single path entry. Returns the sanitized value.
func (cs *ConfigSanitizer) sanitizePathValue(path string) string {
	if cs.rules.TrimWhitespace {
		path = strings.TrimSpace(path)
	}

	if cs.rules.NormalizePaths {
		path = filepath.Clean(path)
	}

	return path
}

// finalizePathList applies remove-duplicates and sort rules to a sanitized
// path list before it is written back to settings.
func (cs *ConfigSanitizer) finalizePathList(list []string) []string {
	if cs.rules.RemoveDuplicates {
		list = cs.removeDuplicates(list)
	}

	if cs.rules.SortArrays {
		cs.sortStrings(list)
	}

	return list
}
