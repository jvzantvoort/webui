package utils

import (
	"os"
	"path/filepath"
)

// FileExists reports whether a file or directory exists at path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// AbsPath returns an absolute path, resolving relative paths against base.
func AbsPath(base, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(base, rel)
}
