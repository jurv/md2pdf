package frontmatter

import "testing"

func TestParseMarkdownWithFrontMatter(t *testing.T) {
	in := []byte("---\ntitle: Example\ntoc:\n  mode: on\n---\n# Heading\nBody")
	front, body, err := ParseMarkdown(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if front["title"] != "Example" {
		t.Fatalf("unexpected title: %v", front["title"])
	}
	if got := string(body); got != "# Heading\nBody" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestParseMarkdownWithoutFrontMatter(t *testing.T) {
	in := []byte("# Heading\nBody")
	front, body, err := ParseMarkdown(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if front != nil {
		t.Fatalf("expected nil front matter for plain markdown")
	}
	if got := string(body); got != string(in) {
		t.Fatalf("unexpected body: %q", got)
	}
}
