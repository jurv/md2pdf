package config

import (
	"strings"
	"testing"
)

func TestPlantUMLStyleValidationInvalidAlign(t *testing.T) {
	cfg := Default()
	cfg.Style.PlantUML.Align = "middle"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.plantuml.align") {
		t.Fatalf("expected style.plantuml.align validation error, got %v", err)
	}
}

func TestPlantUMLStyleValidationNegativeSpacing(t *testing.T) {
	cfg := Default()
	cfg.Style.PlantUML.SpaceBeforePt = -1

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.plantuml.space_before_pt") {
		t.Fatalf("expected style.plantuml.space_before_pt validation error, got %v", err)
	}
}

func TestPlantUMLStyleValidationAcceptsValidValues(t *testing.T) {
	cfg := Default()
	cfg.Style.PlantUML.Align = "right"
	cfg.Style.PlantUML.SpaceBeforePt = 8
	cfg.Style.PlantUML.SpaceAfterPt = 4

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid plantuml style config, got %v", err)
	}
}
