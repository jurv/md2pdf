package render

import "testing"

func TestShouldEnableTOC(t *testing.T) {
	md := []byte("# Title\n\n## Section")
	if !ShouldEnableTOC("auto", md, 2, 3) {
		t.Fatalf("expected toc auto to enable when heading is present")
	}
	if ShouldEnableTOC("auto", []byte("# Title only"), 2, 3) {
		t.Fatalf("expected toc auto to stay disabled when headings are out of range")
	}
	if !ShouldEnableTOC("on", []byte("plain"), 2, 3) {
		t.Fatalf("expected toc on to always enable")
	}
	if ShouldEnableTOC("off", md, 2, 3) {
		t.Fatalf("expected toc off to disable")
	}
}

func TestContainsPlantUML(t *testing.T) {
	if !ContainsPlantUML([]byte("```plantuml\nAlice -> Bob\n```")) {
		t.Fatalf("expected plantuml fence detection")
	}
	if ContainsPlantUML([]byte("```mermaid\nA-->B\n```")) {
		t.Fatalf("did not expect false positive")
	}
}

func TestLatexColorHex(t *testing.T) {
	model, value := latexColor("#1f4E79")
	if model != "HTML" || value != "1F4E79" {
		t.Fatalf("unexpected conversion: model=%q value=%q", model, value)
	}
}

func TestLatexColorNamed(t *testing.T) {
	model, value := latexColor("blue")
	if model != "" || value != "blue" {
		t.Fatalf("unexpected conversion: model=%q value=%q", model, value)
	}
}
