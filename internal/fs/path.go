package fs

import (
	"path/filepath"
)

func ResolveOptionalPath(baseDir, candidate string) string {
	if candidate == "" {
		return ""
	}
	if filepath.IsAbs(candidate) {
		return filepath.Clean(candidate)
	}
	return filepath.Clean(filepath.Join(baseDir, candidate))
}
