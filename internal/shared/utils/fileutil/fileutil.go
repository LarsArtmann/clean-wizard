package utils

import (
	"os"
)

// GetFileSize returns file size in bytes, or 0 if file doesn't exist or is inaccessible.
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
