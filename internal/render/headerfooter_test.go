package render

import (
	"strings"
	"testing"

	"github.com/julien/md2pdf/internal/config"
)

func TestCompileHeaderFooterPartialIncludesFancyDefinitions(t *testing.T) {
	cfg := config.Default()
	cfg.Metadata.Title = "System Specification"
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Integral Service\n31 rue Ampere"},
			},
		},
	}
	cfg.HeaderFooter.Footer.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "bottom",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Confidential"},
			},
		},
		{
			Row:    1,
			Col:    2,
			AlignH: "right",
			AlignV: "bottom",
			Blocks: []config.HeaderFooterBlock{
				{Type: "page_number", Format: "{page}"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	for _, needle := range []string{
		`\fancypagestyle{mdtwohf}`,
		`\fancyhead[C]{`,
		`\fancyfoot[C]{`,
		`Confidential`,
		`\thepage`,
		`\newcommand{\mdtwohfactivate}`,
	} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, out)
		}
	}
}

func TestRenderTextWithTokens(t *testing.T) {
	out := renderTextWithTokens("Title: {title}\nPage {page}/{total_pages} - {date}", "My Document")
	if !strings.Contains(out, "My Document") {
		t.Fatalf("expected title replacement, got %q", out)
	}
	if !strings.Contains(out, `\thepage`) {
		t.Fatalf("expected page token replacement, got %q", out)
	}
	if !strings.Contains(out, `\pageref{LastPage}`) {
		t.Fatalf("expected total page token replacement, got %q", out)
	}
	if !strings.Contains(out, `\today`) {
		t.Fatalf("expected date token replacement, got %q", out)
	}
}

func TestBuildHeaderFooterMetadataActivationFlags(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.ApplyOn = "body_only"
	workDir := t.TempDir()

	meta, err := buildHeaderFooterMetadata(cfg, "/tmp", workDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasAfterTOC := false
	for _, pair := range meta {
		if pair[0] == "hf_activate_after_toc" {
			hasAfterTOC = true
		}
	}
	if !hasAfterTOC {
		t.Fatalf("expected hf_activate_after_toc metadata flag")
	}
}

func TestBuildHeaderFooterMetadataAllPagesStartsImmediately(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.ApplyOn = "all_pages"
	workDir := t.TempDir()

	meta, err := buildHeaderFooterMetadata(cfg, "/tmp", workDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasAtStart := false
	for _, pair := range meta {
		if pair[0] == "hf_activate_at_start" {
			hasAtStart = true
		}
	}
	if !hasAtStart {
		t.Fatalf("expected hf_activate_at_start metadata flag")
	}
}

func TestCompileHeaderFooterPartialUsesZeroMetricsRaisebox(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.Header.RaisePt = 10
	cfg.HeaderFooter.Footer.RaisePt = -12
	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Header"},
			},
		},
	}
	cfg.HeaderFooter.Footer.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "bottom",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Footer"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if !strings.Contains(out, `\raisebox{10pt}[0pt][0pt]{`) {
		t.Fatalf("expected header raisebox with zero metrics, got:\n%s", out)
	}
	if !strings.Contains(out, `\raisebox{-12pt}[0pt][0pt]{`) {
		t.Fatalf("expected footer raisebox with zero metrics, got:\n%s", out)
	}
}

func TestCompileHeaderFooterPartialIncludesFooterReserveAbove(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.FooterReserveAbovePt = 12
	cfg.HeaderFooter.Footer.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "bottom",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Footer"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if !strings.Contains(out, `\setlength{\footskip}{36pt}`) {
		t.Fatalf("expected footskip to include reserve amount, got:\n%s", out)
	}
	if !strings.Contains(out, `\addtolength{\textheight}{-12pt}`) {
		t.Fatalf("expected footer reserve textheight reduction, got:\n%s", out)
	}
}

func TestCompileHeaderFooterPartialInjectsDefaultHeaderLogoFromAssets(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.Assets.LogoHeader = "assets/logo-header.png"

	out, err := compileHeaderFooterPartial(cfg, "/tmp/project")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	if !strings.Contains(out, `\includegraphics[height=22pt,keepaspectratio]{\detokenize{/tmp/project/assets/logo-header.png}}`) {
		t.Fatalf("expected default header logo injection, got:\n%s", out)
	}
}

func TestCompileHeaderFooterPartialDoesNotInjectDefaultHeaderLogoWhenHeaderIsExplicit(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.Assets.LogoHeader = "assets/logo-header.png"
	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Custom header"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp/project")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	if strings.Contains(out, `/tmp/project/assets/logo-header.png`) {
		t.Fatalf("did not expect default header logo injection when header is explicit, got:\n%s", out)
	}
}

func TestCompileHeaderFooterPartialProvidesOpacityWrapperFallback(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{Type: "text", Value: "Header"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	for _, needle := range []string{
		`\@ifundefined{tikz}{\DeclareRobustCommand{\mdtwoapplyimageopacity}[2]{#2}}`,
		`\DeclareRobustCommand{\mdtwoapplyimageopacity}[2]{\tikz[baseline]{\node[anchor=base,inner sep=0pt,outer sep=0pt,opacity=#1]{#2};}}`,
	} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected opacity wrapper helper in header/footer partial, missing %q in:\n%s", needle, out)
		}
	}
}

func TestApplyTextStyleUsesTransparencyForHexColor(t *testing.T) {
	out := applyTextStyle("Confidential", config.TextStyleConfig{
		Color:   "#112233",
		Opacity: 0.35,
	})

	for _, needle := range []string{`\definecolor{mdtwoopacitybase}{HTML}{112233}`, `\color{mdtwoopacitybase!35!white}`, `Confidential`} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected styled text to contain %q, got %q", needle, out)
		}
	}
}

func TestApplyTextStyleOmitsTransparencyAtOpacityOne(t *testing.T) {
	out := applyTextStyle("Confidential", config.TextStyleConfig{
		Color:   "#112233",
		Opacity: 1,
	})

	for _, needle := range []string{`\mdtwoapplyimageopacity{`, `\definecolor{mdtwoopacitybase}{HTML}{112233}`} {
		if strings.Contains(out, needle) {
			t.Fatalf("did not expect %q at opacity 1, got %q", needle, out)
		}
	}
}

func TestCompileHeaderFooterPartialAppliesGlobalOpacityToPageNumbers(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.GlobalStyle.Opacity = 0.4
	cfg.HeaderFooter.Footer.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    2,
			AlignH: "right",
			AlignV: "bottom",
			Blocks: []config.HeaderFooterBlock{
				{Type: "page_number", Format: "{page}"},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	for _, needle := range []string{`!40!white`, `\thepage`} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, out)
		}
	}
}

func TestCompileHeaderFooterPartialAppliesGlobalOpacityToImages(t *testing.T) {
	cfg := config.Default()
	cfg.HeaderFooter.Enabled = true
	cfg.HeaderFooter.GlobalStyle.Opacity = 0.25
	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{Type: "image", Path: "assets/logo-header.png", HeightPt: 22},
			},
		},
	}

	out, err := compileHeaderFooterPartial(cfg, "/tmp/project")
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	for _, needle := range []string{
		`\mdtwoapplyimageopacity{0.25}{`,
		`\includegraphics[height=22pt,keepaspectratio]{\detokenize{/tmp/project/assets/logo-header.png}}`,
	} {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, out)
		}
	}
}

func TestDefaultTemplateLoadsTikzPackage(t *testing.T) {
	if !strings.Contains(defaultTemplate, `\usepackage{tikz}`) {
		t.Fatalf("expected embedded template to load tikz package")
	}
}
