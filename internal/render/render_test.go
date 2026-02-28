package render

import "testing"

func TestShouldEnableTOC(t *testing.T) {
	md := []byte("# Title\n\n## Section")
	if !ShouldEnableTOC("auto", md) {
		t.Fatalf("expected toc auto to enable when heading is present")
	}
	if !ShouldEnableTOC("on", []byte("plain")) {
		t.Fatalf("expected toc on to always enable")
	}
	if ShouldEnableTOC("off", md) {
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
