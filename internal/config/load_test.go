package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSupportsLegacyTOCDepthAlias(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "md2pdf.yaml")
	content := []byte("toc:\n  depth: 2\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(LoadOptions{BaseDir: dir})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.TOC.Depth != 2 || cfg.TOC.ToLevel != 2 {
		t.Fatalf("expected toc depth/to_level=2, got depth=%d to_level=%d", cfg.TOC.Depth, cfg.TOC.ToLevel)
	}
}

func TestLoadBackfillsTOCDepthFromToLevel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "md2pdf.yaml")
	content := []byte("toc:\n  to_level: 4\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(LoadOptions{BaseDir: dir})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.TOC.Depth != 4 || cfg.TOC.ToLevel != 4 {
		t.Fatalf("expected toc depth/to_level=4, got depth=%d to_level=%d", cfg.TOC.Depth, cfg.TOC.ToLevel)
	}
}
