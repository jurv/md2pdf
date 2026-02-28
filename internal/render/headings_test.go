package render

import (
	"strings"
	"testing"

	"github.com/julien/md2pdf/internal/config"
)

func TestExtractFirstH1StripsEntrypointTitle(t *testing.T) {
	in := []byte("```md\n# ignored\n```\n\n# Real Title\n\nIntro\n")
	title, out, found := ExtractFirstH1(in, true)
	if !found {
		t.Fatalf("expected heading to be found")
	}
	if title != "Real Title" {
		t.Fatalf("expected title %q, got %q", "Real Title", title)
	}
	text := string(out)
	if strings.Contains(text, "\n# Real Title\n") {
		t.Fatalf("expected first h1 to be stripped, got: %q", text)
	}
	if !strings.Contains(text, "Intro") {
		t.Fatalf("expected remaining body content")
	}
}

func TestApplyHeadingPolicyDefaultNumberingAndTOCClass(t *testing.T) {
	cfg := config.Default()
	in := []byte("# Doc\n\n## Section\n\n### Subsection\n\n#### Deep\n")
	out, err := ApplyHeadingPolicy(in, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(out)
	for _, needle := range []string{
		"# Doc {.unlisted}",
		"## 1 Section",
		"### 1.1 Subsection",
		"#### Deep {.unlisted}",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, text)
		}
	}
}

func TestApplyHeadingPolicyResetsCountersWhenCrossingAboveRange(t *testing.T) {
	cfg := config.Default()
	in := []byte("## First\n\n### Child\n\n# New Part\n\n## Second\n")
	out, err := ApplyHeadingPolicy(in, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, "## 1 First") {
		t.Fatalf("missing first numbering:\n%s", text)
	}
	if !strings.Contains(text, "### 1.1 Child") {
		t.Fatalf("missing child numbering:\n%s", text)
	}
	if !strings.Contains(text, "# New Part {.unlisted}") {
		t.Fatalf("expected h1 to be marked unlisted:\n%s", text)
	}
	if !strings.Contains(text, "## 1 Second") {
		t.Fatalf("expected numbering reset after h1:\n%s", text)
	}
}

func TestApplyHeadingPolicySupportsPerLevelNotations(t *testing.T) {
	cfg := config.Default()
	cfg.Heading.FromLevel = 2
	cfg.Heading.ToLevel = 4
	cfg.Heading.Separator = "-"
	cfg.Heading.Suffix = ")"
	cfg.Heading.Notation = map[string]string{
		"2": "roman_upper",
		"3": "alpha_lower",
		"4": "decimal",
	}
	cfg.TOC.FromLevel = 2
	cfg.TOC.ToLevel = 4
	in := []byte("## A\n### B\n#### C\n### D\n")
	out, err := ApplyHeadingPolicy(in, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(out)
	for _, needle := range []string{
		"## I) A",
		"### I-a) B",
		"#### I-a-1) C",
		"### I-b) D",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, text)
		}
	}
}

func TestApplyHeadingPolicyAvoidsDuplicateManualNumberPrefix(t *testing.T) {
	cfg := config.Default()
	in := []byte("## 1 Glossary\n### 1.1 Terms\n")
	out, err := ApplyHeadingPolicy(in, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(out)
	if strings.Contains(text, "## 1 1 Glossary") {
		t.Fatalf("expected level 2 duplicate prefix to be removed, got:\n%s", text)
	}
	if strings.Contains(text, "### 1.1 1.1 Terms") {
		t.Fatalf("expected level 3 duplicate prefix to be removed, got:\n%s", text)
	}
	for _, needle := range []string{
		"## 1 Glossary",
		"### 1.1 Terms",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, text)
		}
	}
}
