package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHeaderFooterValidateInvalidApplyOn(t *testing.T) {
	cfg := Default()
	cfg.HeaderFooter.ApplyOn = "unknown"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "header_footer.apply_on") {
		t.Fatalf("expected apply_on validation error, got %v", err)
	}
}

func TestHeaderFooterValidateCellBounds(t *testing.T) {
	cfg := Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.Header.Grid.Cells = []HeaderFooterCell{
		{
			Row:    2,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []HeaderFooterBlock{
				{Type: "text", Value: "Header"},
			},
		},
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "header_footer.header.grid.cells[0].row") {
		t.Fatalf("expected row bounds error, got %v", err)
	}
}

func TestLoadRejectsDeprecatedHeaderFooterKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "md2pdf.yaml")
	content := []byte("header_footer:\n  footer_left: \"legacy\"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(LoadOptions{BaseDir: dir})
	if err == nil || !strings.Contains(err.Error(), "deprecated key header_footer.footer_left") {
		t.Fatalf("expected deprecated key error, got %v", err)
	}
}

func TestHeaderFooterConfiguredButDisabled(t *testing.T) {
	cfg := Default()
	cfg.HeaderFooter.Enabled = false
	cfg.HeaderFooter.Header.Grid.Cells = []HeaderFooterCell{
		{
			Row: 1,
			Col: 1,
			Blocks: []HeaderFooterBlock{
				{Type: "text", Value: "Configured"},
			},
		},
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "header_footer.enabled must be true") {
		t.Fatalf("expected explicit enabled error, got %v", err)
	}
}

func TestHeaderFooterValidateFooterReserveAbovePt(t *testing.T) {
	cfg := Default()
	cfg.HeaderFooter.FooterReserveAbovePt = -1
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "header_footer.footer_reserve_above_pt") {
		t.Fatalf("expected footer_reserve_above_pt validation error, got %v", err)
	}
}
