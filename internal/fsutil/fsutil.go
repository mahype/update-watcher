package fsutil

import (
	"os"
	"path/filepath"
)

// FileExists returns true if the path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists returns true if the path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FindParentDir walks up the directory tree from startPath and returns the
// first directory for which the check function returns true. If no match is
// found, startPath is returned unchanged.
func FindParentDir(startPath string, check func(dir string) bool) string {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return startPath
	}

	dir := absPath
	for {
		if check(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return startPath
}
