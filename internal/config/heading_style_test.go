package config

import (
	"strings"
	"testing"
)

func TestHeadingStyleValidationInvalidColor(t *testing.T) {
	cfg := Default()
	cfg.Style.Headings.H2.Color = "not-a-color!"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.headings.h2.color") {
		t.Fatalf("expected style.headings.h2.color validation error, got %v", err)
	}
}

func TestHeadingStyleValidationInvalidSize(t *testing.T) {
	cfg := Default()
	size := 0.0
	cfg.Style.Headings.H2.SizePt = &size

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.headings.h2.size_pt") {
		t.Fatalf("expected style.headings.h2.size_pt validation error, got %v", err)
	}
}

func TestHeadingStyleValidationInvalidSpacing(t *testing.T) {
	cfg := Default()
	space := -1.0
	cfg.Style.Headings.H3.SpaceBeforePt = &space

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.headings.h3.space_before_pt") {
		t.Fatalf("expected style.headings.h3.space_before_pt validation error, got %v", err)
	}
}

func TestHeadingStyleValidationAcceptsPositiveSize(t *testing.T) {
	cfg := Default()
	size := 18.0
	spaceBefore := 10.0
	spaceAfter := 8.0
	cfg.Style.Headings.H1.SizePt = &size
	cfg.Style.Headings.H1.Color = "#1F4E79"
	cfg.Style.Headings.H1.SpaceBeforePt = &spaceBefore
	cfg.Style.Headings.H1.SpaceAfterPt = &spaceAfter

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected heading style configuration to be valid, got %v", err)
	}
}

func TestHeadingStyleDefaultKeepsWithNext(t *testing.T) {
	cfg := Default()
	if !cfg.Style.Headings.KeepWithNext {
		t.Fatalf("expected style.headings.keep_with_next to default to true")
	}
}

func TestHeadingStyleAllowsKeepWithNextDisabled(t *testing.T) {
	cfg := Default()
	cfg.Style.Headings.KeepWithNext = false
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled keep_with_next to validate, got %v", err)
	}
}
