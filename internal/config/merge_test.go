package config

import "testing"

func TestMergeMapSupportsNullUnset(t *testing.T) {
	base := map[string]any{
		"pdf": map[string]any{
			"engine":   "xelatex",
			"template": "default.tex",
		},
	}
	overlay := map[string]any{
		"pdf": map[string]any{
			"template": nil,
		},
	}

	merged := MergeMap(base, overlay)
	pdf := merged["pdf"].(map[string]any)
	if _, ok := pdf["template"]; ok {
		t.Fatalf("expected template to be removed when set to null")
	}
	if got := pdf["engine"]; got != "xelatex" {
		t.Fatalf("expected engine to remain untouched, got %v", got)
	}
}

func TestSetNestedValue(t *testing.T) {
	root := map[string]any{}
	SetNestedValue(root, []string{"toc", "depth"}, 4)

	toc, ok := root["toc"].(map[string]any)
	if !ok {
		t.Fatalf("toc nested object was not created")
	}
	if got := toc["depth"]; got != 4 {
		t.Fatalf("expected depth=4, got %v", got)
	}
}
