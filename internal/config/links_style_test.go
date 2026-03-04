package config

import (
	"strings"
	"testing"
)

func TestLinksStyleValidationInvalidColor(t *testing.T) {
	cfg := Default()
	cfg.Style.Links.Color = "invalid-color!"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.links.color") {
		t.Fatalf("expected style.links.color validation error, got %v", err)
	}
}

func TestLinksStyleValidationInvalidTOCColor(t *testing.T) {
	cfg := Default()
	cfg.Style.Links.TOCColor = "###"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.links.toc_color") {
		t.Fatalf("expected style.links.toc_color validation error, got %v", err)
	}
}

func TestLinksStyleValidationAcceptsValidValues(t *testing.T) {
	cfg := Default()
	cfg.Style.Links.Color = "#1F4E79"
	cfg.Style.Links.URLColor = "teal"
	cfg.Style.Links.CitationColor = "#A94442"
	cfg.Style.Links.TOCColor = "#2E86C1"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid links style, got %v", err)
	}
}
