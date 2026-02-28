package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/julien/md2pdf/internal/config"
)

func ResolveSources(inputPath string, src config.SourcesConfig) ([]string, error) {
	baseDir := filepath.Dir(inputPath)
	if len(src.Explicit) == 0 && len(src.Include) == 0 {
		return []string{inputPath}, nil
	}

	ordered := make([]string, 0)
	seen := make(map[string]struct{})

	addPath := func(path string) {
		canonical := canonicalPath(path)
		if _, exists := seen[canonical]; exists {
			return
		}
		seen[canonical] = struct{}{}
		ordered = append(ordered, path)
	}

	for _, item := range src.Explicit {
		resolved := item
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(baseDir, item)
		}
		resolved = filepath.Clean(resolved)
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("explicit source not found: %s", resolved)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("explicit source is a directory, expected file: %s", resolved)
		}
		addPath(resolved)
	}

	included := make([]string, 0)
	for _, pattern := range src.Include {
		resolved := pattern
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(baseDir, pattern)
		}
		matches, err := filepath.Glob(filepath.Clean(resolved))
		if err != nil {
			return nil, fmt.Errorf("invalid include glob %q: %w", pattern, err)
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			included = append(included, filepath.Clean(match))
		}
	}
	sort.Strings(included)
	for _, item := range included {
		addPath(item)
	}

	if len(ordered) == 0 {
		return nil, fmt.Errorf("multi-source configuration resolved to an empty source list")
	}
	return ordered, nil
}

func canonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}
