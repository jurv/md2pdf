package config

import (
	"strings"
	"testing"
)

func TestBackgroundDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Background.ImageFit != "cover" {
		t.Fatalf("expected background.image_fit default to be cover, got %q", cfg.Background.ImageFit)
	}
	if cfg.Background.ApplyOn != "all_pages" {
		t.Fatalf("expected background.apply_on default to be all_pages, got %q", cfg.Background.ApplyOn)
	}
}

func TestBackgroundImageFitValidation(t *testing.T) {
	cfg := Default()
	cfg.Background.ImageFit = "invalid"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "background.image_fit") {
		t.Fatalf("expected background.image_fit validation error, got %v", err)
	}
}

func TestBackgroundApplyOnValidation(t *testing.T) {
	cfg := Default()
	cfg.Background.ApplyOn = "invalid"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "background.apply_on") {
		t.Fatalf("expected background.apply_on validation error, got %v", err)
	}
}

func TestBackgroundConfigurationIsValid(t *testing.T) {
	cfg := Default()
	cfg.Background.Image = "assets/background.png"
	cfg.Background.ImageFit = "contain"
	cfg.Background.ApplyOn = "toc_and_body"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected background configuration to be valid, got %v", err)
	}
}
