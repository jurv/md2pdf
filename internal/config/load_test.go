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

func TestLoadSupportsLegacyHeadingStyleAliases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "md2pdf.yaml")
	content := []byte(`
style:
  headings:
    color: "#1F4E79"
    h2_size_pt: 18
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(LoadOptions{BaseDir: dir})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.Style.Headings.H1.Color != "#1F4E79" || cfg.Style.Headings.H2.Color != "#1F4E79" {
		t.Fatalf("expected legacy headings.color alias to be applied to h1/h2, got h1=%q h2=%q", cfg.Style.Headings.H1.Color, cfg.Style.Headings.H2.Color)
	}
	if cfg.Style.Headings.H2.SizePt == nil || *cfg.Style.Headings.H2.SizePt != 18 {
		t.Fatalf("expected legacy h2_size_pt alias to populate h2.size_pt, got %#v", cfg.Style.Headings.H2.SizePt)
	}
}
