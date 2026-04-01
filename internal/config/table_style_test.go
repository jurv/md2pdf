package config

import "testing"

func TestTableStyleDefaults(t *testing.T) {
	cfg := Default()

	if cfg.Style.Tables.RowSpacingFactor != 1.2 {
		t.Fatalf("expected default table row spacing factor 1.2, got %v", cfg.Style.Tables.RowSpacingFactor)
	}
	if !cfg.Style.Tables.ZebraEnabled {
		t.Fatalf("expected zebra striping to be enabled by default")
	}
	if cfg.Style.Tables.ZebraColor != "#F5F5F5" {
		t.Fatalf("expected default zebra color #F5F5F5, got %q", cfg.Style.Tables.ZebraColor)
	}
}

func TestTableStyleValidationRejectsNonPositiveRowSpacingFactor(t *testing.T) {
	cfg := Default()
	cfg.Style.Tables.RowSpacingFactor = 0

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid table row spacing factor to fail validation")
	}
}

func TestTableStyleValidationRejectsInvalidZebraColor(t *testing.T) {
	cfg := Default()
	cfg.Style.Tables.ZebraColor = "rgba(0,0,0,0.1)"

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid zebra color to fail validation")
	}
}

func TestTableStyleValidationAllowsDisabledZebra(t *testing.T) {
	cfg := Default()
	cfg.Style.Tables.ZebraEnabled = false

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled zebra style to be valid, got %v", err)
	}
}
