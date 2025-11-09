package format

// Bytes formats bytes for display (alias for Size for consistency)
func Bytes(bytes int64) string {
	return Size(bytes)
}
