package render

import (
	"path/filepath"
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

func TestMetadataArgsIncludesPlantUMLStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Style.PlantUML.Align = "right"
	cfg.Style.PlantUML.SpaceBeforePt = 9
	cfg.Style.PlantUML.SpaceAfterPt = 3

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"plantuml_align=right",
		"plantuml_space_before_pt=9",
		"plantuml_space_after_pt=3",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
}

func TestMetadataArgsCoverImageImplicitBuiltinMode(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.Mode = "none"
	cfg.Cover.Image = "assets/cover.png"
	cfg.Cover.ImageFit = "cover"

	args, err := metadataArgs(cfg, "/tmp/project", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"cover_mode_first_page_background=true",
		"cover_image=" + filepath.Clean("/tmp/project/assets/cover.png"),
		"cover_image_fit_cover=true",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
	if strings.Contains(joined, "cover_mode_builtin=true") {
		t.Fatalf("did not expect cover_mode_builtin for implicit cover.image mode, got %q", joined)
	}
}

func TestMetadataArgsCoverImageFitContain(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.Mode = "builtin"
	cfg.Cover.Image = "assets/cover.png"
	cfg.Cover.ImageFit = "contain"

	args, err := metadataArgs(cfg, "/tmp/project", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "cover_image_fit_contain=true") {
		t.Fatalf("expected contain fit metadata, got %q", joined)
	}
	if strings.Contains(joined, "cover_image_fit_cover=true") {
		t.Fatalf("did not expect cover fit metadata when contain is selected, got %q", joined)
	}
}

func TestDefaultTemplateDefinesCoverImageHelpers(t *testing.T) {
	for _, needle := range []string{
		`\newcommand{\mdtwoaddcoverimagecover}[1]{`,
		`\newcommand{\mdtwoaddcoverimagecontain}[1]{`,
		`\newcommand{\mdtwoaddcoverimagestretch}[1]{`,
		`\AddToShipoutPictureBG*`,
		`$if(cover_mode_first_page_background)$`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestDefaultTemplateSkipsInlineTitleWhenBuiltinCoverIsEnabled(t *testing.T) {
	for _, needle := range []string{
		`$if(title_render_inline)$`,
		`$if(cover_mode_builtin)$`,
		`\maketitle`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestMetadataArgsIncludesHeadingStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Style.Headings.H1.Color = "#1F4E79"
	h2 := 18.0
	spaceBefore := 12.0
	spaceAfter := 8.0
	cfg.Style.Headings.H2.SizePt = &h2
	cfg.Style.Headings.H2.SpaceBeforePt = &spaceBefore
	cfg.Style.Headings.H2.SpaceAfterPt = &spaceAfter

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"heading_style_enabled=true",
		"heading_h1_color_model=HTML",
		"heading_h1_color_value=1F4E79",
		"heading_h2_size_pt=18",
		"heading_h2_space_before_pt=12",
		"heading_h2_space_after_pt=8",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
}

func TestDefaultTemplateDefinesHeadingStyleHooks(t *testing.T) {
	for _, needle := range []string{
		`\usepackage{sectsty}`,
		`\usepackage{titlesec}`,
		`\newcommand{\mdtwohoneStyle}{`,
		`$if(heading_style_enabled)$`,
		`$if(heading_h2_space_before_pt)$`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestDefaultTemplateDefinesPlantUMLStyleHooks(t *testing.T) {
	for _, needle := range []string{
		`\newcommand{\mdtwoplantumlalign}`,
		`\newcommand{\mdtwoplantumlspacebefore}`,
		`\newcommand{\mdtwoplantumlspaceafter}`,
		`\str_if_in:nnTF {#2} {plantuml-images/}`,
		`\mdtwo_includegraphics_with_plantuml_style:nn`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestMetadataArgsIncludesLinkStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Style.Links.Color = "#1F4E79"
	cfg.Style.Links.URLColor = "teal"
	cfg.Style.Links.CitationColor = "#A94442"
	cfg.Style.Links.TOCColor = "#2E86C1"

	args, err := metadataArgs(cfg, "/tmp", true, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"hyperref_link_color_model=HTML",
		"hyperref_link_color_value=1F4E79",
		"hyperref_url_color_value=teal",
		"hyperref_cite_color_model=HTML",
		"hyperref_cite_color_value=A94442",
		"hyperref_toc_link_color_model=HTML",
		"hyperref_toc_link_color_value=2E86C1",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
}

func TestDefaultTemplateDefinesLinkStyleHooks(t *testing.T) {
	for _, needle := range []string{
		`mdtwolinkcolor`,
		`mdtwotoclinkcolor`,
		`$if(hyperref_link_color_value)$`,
		`\hypersetup{linkcolor=mdtwotoclinkcolor}`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestMetadataArgsCoverImageSimpleModeKeepsAllPagesHeaderFooterStart(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.Mode = "none"
	cfg.Cover.Image = "assets/cover.png"
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.ApplyOn = "all_pages"

	args, err := metadataArgs(cfg, "/tmp/project", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "hf_activate_at_start=true") {
		t.Fatalf("expected hf_activate_at_start for first-page-background mode, got %q", joined)
	}
	if strings.Contains(joined, "hf_activate_after_cover=true") {
		t.Fatalf("did not expect hf_activate_after_cover for first-page-background mode, got %q", joined)
	}
}
