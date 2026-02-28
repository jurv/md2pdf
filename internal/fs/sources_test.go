package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/julien/md2pdf/internal/config"
)

func TestResolveSourcesExplicitThenIncludeSorted(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "10.md"), "a")
	mustWrite(t, filepath.Join(dir, "20.md"), "b")
	mustWrite(t, filepath.Join(dir, "05.md"), "c")

	sources, err := ResolveSources(filepath.Join(dir, "entry.md"), config.SourcesConfig{
		Explicit: []string{"20.md"},
		Include:  []string{"*.md"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sources) != 3 {
		t.Fatalf("expected 3 unique sources, got %d (%v)", len(sources), sources)
	}
	if filepath.Base(sources[0]) != "20.md" {
		t.Fatalf("expected explicit source first, got %s", sources[0])
	}
	if filepath.Base(sources[1]) != "05.md" {
		t.Fatalf("expected sorted include source second, got %s", sources[1])
	}
	if filepath.Base(sources[2]) != "10.md" {
		t.Fatalf("expected sorted include source third, got %s", sources[2])
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write fixture %s: %v", path, err)
	}
}
