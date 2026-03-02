package render

import (
	"strings"
	"testing"

	"github.com/julien/md2pdf/internal/config"
)

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

func TestPandocInputFormatSupportsInlineListBlockquotes(t *testing.T) {
	if pandocInputFormat != "markdown-blank_before_blockquote" {
		t.Fatalf("unexpected input format %q", pandocInputFormat)
	}
}

func TestDefaultTemplateDefinesStyledQuoteEnvironment(t *testing.T) {
	for _, needle := range []string{
		`\renewenvironment{quote}{`,
		`\newcommand{\mdtwoquotebarcolor}`,
		`\newcommand{\mdtwoquotetextcolor}`,
		`\newcommand{\mdtwoquotebg}[1]`,
		`\newcommand{\mdtwoquotebarwidth}`,
		`\newcommand{\mdtwoquotegap}`,
		`\newcommand{\mdtwoquotepadding}`,
		`\MakeFramed{\advance\hsize-\width \FrameRestore}`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestMetadataArgsIncludesBlockQuoteStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Style.BlockQuote.BarColor = "#AABBCC"
	cfg.Style.BlockQuote.TextColor = "gray"
	cfg.Style.BlockQuote.BackgroundColor = "#F0F2F4"
	cfg.Style.BlockQuote.BarWidthPt = 1.2
	cfg.Style.BlockQuote.GapPt = 6
	cfg.Style.BlockQuote.PaddingPt = 3

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"blockquote_bar_color_model=HTML",
		"blockquote_bar_color_value=AABBCC",
		"blockquote_text_color_value=gray",
		"blockquote_background_color_model=HTML",
		"blockquote_background_color_value=F0F2F4",
		"blockquote_bar_width_pt=1.2",
		"blockquote_gap_pt=6",
		"blockquote_padding_pt=3",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
}
