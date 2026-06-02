package config

import (
	"strings"
	"testing"
)

func TestDefaultEmojiStyleUsesAutoMode(t *testing.T) {
	cfg := Default()

	if cfg.Style.Emoji.Mode != "auto" {
		t.Fatalf("expected default emoji mode %q, got %q", "auto", cfg.Style.Emoji.Mode)
	}
	if cfg.Style.Emoji.Font == "" {
		t.Fatalf("expected default emoji font to be configured")
	}
	if cfg.Style.Emoji.ImageHeightEm <= 0 {
		t.Fatalf("expected default emoji image height to be > 0")
	}
}

func TestEmojiStyleValidationRejectsInvalidMode(t *testing.T) {
	cfg := Default()
	cfg.Style.Emoji.Mode = "font-fallback"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.emoji.mode") {
		t.Fatalf("expected style.emoji.mode validation error, got %v", err)
	}
}

func TestEmojiStyleValidationRejectsInvalidHeight(t *testing.T) {
	cfg := Default()
	cfg.Style.Emoji.ImageHeightEm = 0

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.emoji.image_height_em") {
		t.Fatalf("expected style.emoji.image_height_em validation error, got %v", err)
	}
}
