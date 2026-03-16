package config

import "testing"

func TestFigureStyleDefaultsEnableCaptions(t *testing.T) {
	cfg := Default()
	if !cfg.Style.Figures.CaptionEnabled {
		t.Fatalf("expected figure captions to be enabled by default")
	}
}

func TestFigureStyleValidationAcceptsDisabledCaptions(t *testing.T) {
	cfg := Default()
	cfg.Style.Figures.CaptionEnabled = false

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected figure style configuration to be valid, got %v", err)
	}
}
