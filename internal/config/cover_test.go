package config

import (
	"strings"
	"testing"
)

func TestCoverImageFitValidation(t *testing.T) {
	cfg := Default()
	cfg.Cover.ImageFit = "invalid"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "cover.image_fit") {
		t.Fatalf("expected cover.image_fit validation error, got %v", err)
	}
}

func TestCoverImageFitAllowsEmptyValue(t *testing.T) {
	cfg := Default()
	cfg.Cover.ImageFit = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected empty cover.image_fit to be accepted, got %v", err)
	}
}

func TestCoverExternalTemplateAndImageConflict(t *testing.T) {
	cfg := Default()
	cfg.Cover.Mode = "external_template"
	cfg.Cover.ExternalTemplate = "cover.tex"
	cfg.Cover.Image = "assets/cover.png"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "cover.image cannot be set when cover.mode=external_template") {
		t.Fatalf("expected external_template/image conflict error, got %v", err)
	}
}

func TestCoverImageWithModeNoneIsValid(t *testing.T) {
	cfg := Default()
	cfg.Cover.Mode = "none"
	cfg.Cover.Image = "assets/cover.png"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected cover.image with mode=none to be valid, got %v", err)
	}
}
