package format

// Bytes formats bytes for display (alias for Size for consistency)
func Bytes(bytes uint64) string {
	return Size(int64(bytes))
}
