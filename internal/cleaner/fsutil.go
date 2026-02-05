package cleaner

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// GetHomeDir returns user's home directory.
func GetHomeDir() (string, error) {
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.HomeDir, nil
	}

	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile, nil
	}

	return "", fmt.Errorf("unable to determine home directory")
}

// GetDirSize returns total size of directory recursively.
func GetDirSize(path string) int64 {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}

	return size
}

// GetDirModTime returns most recent modification time in directory.
func GetDirModTime(path string) time.Time {
	var modTime time.Time

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.ModTime().After(modTime) {
			modTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return time.Time{}
	}

	return modTime
}