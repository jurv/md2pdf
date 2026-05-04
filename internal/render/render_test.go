package render

import (
	"os"
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
	if pandocInputFormat != "markdown-blank_before_blockquote-blank_before_header" {
		t.Fatalf("unexpected input format %q", pandocInputFormat)
	}
}

func TestDefaultTemplateDefinesStyledQuoteEnvironment(t *testing.T) {
	for _, needle := range []string{
		`\usepackage{amssymb}`,
		`\providecommand{\hl}[1]{#1}`,
		`\@ifundefined{Shaded}{%`,
		`\input{$unicode_symbols_partial$}`,
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

func TestEmbeddedTableCodeWrapFilterTargetsTablesOnly(t *testing.T) {
	for _, needle := range []string{
		`function Table(tbl)`,
		`Code = wrap_code`,
		`pandoc.RawInline("latex", "\\path" .. delim .. text .. delim)`,
	} {
		if !strings.Contains(tableCodeWrapFilter, needle) {
			t.Fatalf("embedded table code wrap filter missing %q", needle)
		}
	}
}

func TestEmbeddedTableAutoWidthFilterContainsExpectedMarkers(t *testing.T) {
	for _, needle := range []string{
		`md2pdf table auto-width filter`,
		`local function cell_score(compact_len, token_len)`,
		`local function token_floor(longest)`,
		`return #tbl.colspecs >= 2`,
		`local floor_budget = 0.55`,
		`tbl.colspecs = colspecs`,
	} {
		if !strings.Contains(tableAutoWidthFilter, needle) {
			t.Fatalf("embedded table auto-width filter missing %q", needle)
		}
	}
}

func TestWriteTableCodeWrapFilter(t *testing.T) {
	workDir := t.TempDir()
	path, err := writeTableCodeWrapFilter(workDir)
	if err != nil {
		t.Fatalf("writeTableCodeWrapFilter returned error: %v", err)
	}
	if !strings.HasPrefix(path, workDir) {
		t.Fatalf("expected filter path %q to be inside %q", path, workDir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written filter: %v", err)
	}
	if string(content) != tableCodeWrapFilter {
		t.Fatalf("written filter content does not match embedded filter")
	}
}

func TestWriteTableAutoWidthFilter(t *testing.T) {
	workDir := t.TempDir()
	path, err := writeTableAutoWidthFilter(workDir)
	if err != nil {
		t.Fatalf("writeTableAutoWidthFilter returned error: %v", err)
	}
	if !strings.HasPrefix(path, workDir) {
		t.Fatalf("expected filter path %q to be inside %q", path, workDir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written filter: %v", err)
	}
	if string(content) != tableAutoWidthFilter {
		t.Fatalf("written filter content does not match embedded filter")
	}
}

func TestEmbeddedColumnsFilterContainsUpstreamMarkers(t *testing.T) {
	for _, needle := range []string{
		`Columns - multiple column support in Pandoc's markdown.`,
		`@license MIT - see LICENSE file for details.`,
		`local target_formats = {`,
		`"latex",`,
		`\\begin{multicols}{`,
	} {
		if !strings.Contains(columnsFilter, needle) {
			t.Fatalf("embedded columns filter missing %q", needle)
		}
	}
}

func TestDefaultTemplateExposesPandocHeaderIncludes(t *testing.T) {
	for _, needle := range []string{
		`$for(header-includes)$`,
		`$header-includes$`,
		`$endfor$`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestWriteColumnsFilter(t *testing.T) {
	workDir := t.TempDir()
	path, err := writeColumnsFilter(workDir)
	if err != nil {
		t.Fatalf("writeColumnsFilter returned error: %v", err)
	}
	if !strings.HasPrefix(path, workDir) {
		t.Fatalf("expected filter path %q to be inside %q", path, workDir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written filter: %v", err)
	}
	if string(content) != columnsFilter {
		t.Fatalf("written filter content does not match embedded filter")
	}
}

func TestEmbeddedSideBySideFilterContainsExpectedMarkers(t *testing.T) {
	for _, needle := range []string{
		`md2pdf side-by-side layout filter`,
		`side-by-side`,
		`\begin{minipage}[`,
		`display: flex; align-items: `,
		`ratio`,
		`valign`,
	} {
		if !strings.Contains(sideBySideFilter, needle) {
			t.Fatalf("embedded side-by-side filter missing %q", needle)
		}
	}
}

func TestWriteSideBySideFilter(t *testing.T) {
	workDir := t.TempDir()
	path, err := writeSideBySideFilter(workDir)
	if err != nil {
		t.Fatalf("writeSideBySideFilter returned error: %v", err)
	}
	if !strings.HasPrefix(path, workDir) {
		t.Fatalf("expected filter path %q to be inside %q", path, workDir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written filter: %v", err)
	}
	if string(content) != sideBySideFilter {
		t.Fatalf("written filter content does not match embedded filter")
	}
}

func TestCompileCodeSymbolNormalizeFilterIncludesConfiguredAliases(t *testing.T) {
	filter, err := compileCodeSymbolNormalizeFilter(config.SymbolStyleConfig{
		Replace: map[string]string{
			"✅":      `\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}`,
			"💡":      `\mdtwosymbolglyph{✦}{\textasteriskcentered}`,
			"\uFE0F": ``,
		},
	})
	if err != nil {
		t.Fatalf("compileCodeSymbolNormalizeFilter returned error: %v", err)
	}

	for _, needle := range []string{
		`[utf8.char(0x2705)] = utf8.char(0x2611)`,
		`[utf8.char(0x1F4A1)] = utf8.char(0x2726)`,
		`[utf8.char(0xFE0F)] = ""`,
		`function Code(elem)`,
		`function CodeBlock(elem)`,
	} {
		if !strings.Contains(filter, needle) {
			t.Fatalf("expected filter to contain %q, got %q", needle, filter)
		}
	}
}

func TestWriteCodeSymbolNormalizeFilterSkipsWhenNoAliasesExist(t *testing.T) {
	path, err := writeCodeSymbolNormalizeFilter(t.TempDir(), config.SymbolStyleConfig{})
	if err != nil {
		t.Fatalf("writeCodeSymbolNormalizeFilter returned error: %v", err)
	}
	if path != "" {
		t.Fatalf("expected no filter path when there are no code symbol aliases, got %q", path)
	}
}

func TestMetadataArgsIncludesUnicodeSymbolPartial(t *testing.T) {
	cfg := config.Default()

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "unicode_symbols_partial=") {
		t.Fatalf("expected metadata to contain unicode symbol partial path, got %q", joined)
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

func TestMetadataArgsIncludesHiddenFigureCaptionFlag(t *testing.T) {
	cfg := config.Default()
	cfg.Style.Figures.CaptionEnabled = false

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "figure_caption_hidden=true") {
		t.Fatalf("expected metadata to contain figure caption visibility flag, got %q", joined)
	}
}

func TestMetadataArgsIncludesTableStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Style.Tables.RowSpacingFactor = 1.35
	cfg.Style.Tables.ZebraEnabled = true
	cfg.Style.Tables.ZebraColor = "#EFEFEF"

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"table_row_spacing_factor=1.35",
		"table_zebra_enabled=true",
		"table_zebra_color_model=HTML",
		"table_zebra_color_value=EFEFEF",
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

func TestMetadataArgsIncludesDocumentBackground(t *testing.T) {
	cfg := config.Default()
	cfg.Background.Image = "assets/background.png"
	cfg.Background.ImageFit = "stretch"
	cfg.Background.ApplyOn = "toc_and_body"

	args, err := metadataArgs(cfg, "/tmp/project", true, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"background_image=" + filepath.Clean("/tmp/project/assets/background.png"),
		"background_image_fit_stretch=true",
		"background_apply_on_toc_and_body=true",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
	if strings.Contains(joined, "background_apply_on_all_pages=true") {
		t.Fatalf("did not expect all_pages background metadata when toc_and_body is selected, got %q", joined)
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

func TestDefaultTemplateDefinesDocumentBackgroundHelpers(t *testing.T) {
	for _, needle := range []string{
		`\newcommand{\mdtwoplacebackgroundgraphic}[1]{`,
		`\newcommand{\mdtwoaddbackgroundimagecover}[1]{`,
		`\newcommand{\mdtwoaddbackgroundimagecontain}[1]{`,
		`\newcommand{\mdtwoaddbackgroundimagestretch}[1]{`,
		`\AddToShipoutPictureBG{`,
		`$if(background_apply_on_all_pages)$`,
		`$if(background_apply_on_toc_and_body)$`,
		`$if(background_apply_on_body_only)$`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestDefaultTemplateDefinesFigureCaptionHooks(t *testing.T) {
	for _, needle := range []string{
		`\DeclareCaptionFormat{mdtwonoop}{}`,
		`$if(figure_caption_hidden)$`,
		`\captionsetup[figure]{format=mdtwonoop,labelformat=empty,labelsep=none,skip=0pt}`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
}

func TestDefaultTemplateDefinesTableStyleHooks(t *testing.T) {
	for _, needle := range []string{
		`\usepackage[table]{xcolor}`,
		`\usepackage{etoolbox}`,
		`\usepackage{ragged2e}`,
		`\definecolor{mdtwotablezebracolor}{HTML}{EAEAEA}`,
		`\newcommand{\mdtwoapplytablestyle}{`,
		`\small%`,
		`\setlength{\tabcolsep}{3pt}%`,
		`\let\raggedright\RaggedRight%`,
		`\renewcommand{\bottomrule}{\noalign{\nobreak}\hiderowcolors\mdtwoorigbottomrule}%`,
		`\AtBeginEnvironment{longtable}{\mdtwoapplytablestyle}`,
		`\AtEndEnvironment{longtable}{\mdtworesettablestyle}`,
		`$if(table_zebra_enabled)$`,
		`$if(table_row_spacing_factor)$`,
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
		`\mdtwodocumenttitleblock`,
		`\textcolor{mdtwotitlecolor}{$title$}`,
	} {
		if !strings.Contains(defaultTemplate, needle) {
			t.Fatalf("default template missing %q", needle)
		}
	}
	if strings.Contains(defaultTemplate, `\maketitle`) {
		t.Fatalf("default template should not rely on \\maketitle anymore")
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

func TestMetadataArgsIncludesPrimaryColorAndFontMetadata(t *testing.T) {
	cfg := config.Default()
	cfg.Style.Colors.Primary = "#1F4E79"
	cfg.Style.Fonts.Body = "Calibri"
	cfg.Style.Fonts.Heading = "Cambria"

	args, err := metadataArgs(cfg, "/tmp", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{
		"color_primary=#1F4E79",
		"color_primary_model=HTML",
		"color_primary_value=1F4E79",
		"font_body=Calibri",
		"font_heading=Cambria",
		"heading_style_enabled=true",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("expected metadata to contain %q, got %q", needle, joined)
		}
	}
}

func TestMetadataArgsUsesAssetsLogoCoverAsBuiltinFallback(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.Mode = "builtin"
	cfg.Assets.LogoCover = "assets/logo-cover.png"

	args, err := metadataArgs(cfg, "/tmp/project", false, t.TempDir())
	if err != nil {
		t.Fatalf("metadataArgs returned error: %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "cover_logo="+filepath.Clean("/tmp/project/assets/logo-cover.png")) {
		t.Fatalf("expected builtin cover logo fallback in metadata, got %q", joined)
	}
}

func TestDefaultTemplateDefinesHeadingStyleHooks(t *testing.T) {
	for _, needle := range []string{
		`\usepackage{sectsty}`,
		`\usepackage{titlesec}`,
		`\newcommand{\mdtwoheadingfontstyle}{}`,
		`\setmainfont{$font_body$}`,
		`\newcommand{\mdtwohoneStyle}{`,
		`\newcommand{\mdtwoheadingbaseStyle}{`,
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

func TestCompileUnicodeSymbolPartialIncludesConfiguredMappings(t *testing.T) {
	partial, err := compileUnicodeSymbolPartial(config.SymbolStyleConfig{
		FallbackFont: "Noto Sans Symbols2",
		FallbackFor:  []string{"☐", "☑"},
		Replace: map[string]string{
			"✅": `\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}`,
			"❌": `\mdtwosymbolglyph{☒}{\ensuremath{\boxtimes}}`,
		},
	})
	if err != nil {
		t.Fatalf("compileUnicodeSymbolPartial returned error: %v", err)
	}

	for _, needle := range []string{
		`\providecommand{\mdtwosymbolglyph}[2]{#2}`,
		`\newfontfamily\mdtwosymbolfont{Noto Sans Symbols2}`,
		`\newunicodechar{☐}{\mdtwosymbolglyph{☐}{\ensuremath{\square}}}`,
		`\newunicodechar{☑}{\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}}`,
		`% U+2705`,
		`\newunicodechar{✅}{\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}}`,
		`% U+274C`,
		`\newunicodechar{❌}{\mdtwosymbolglyph{☒}{\ensuremath{\boxtimes}}}`,
	} {
		if !strings.Contains(partial, needle) {
			t.Fatalf("expected partial to contain %q, got %q", needle, partial)
		}
	}
}

func TestCompileUnicodeSymbolPartialSkipsFallbackEntryWhenExplicitReplacementExists(t *testing.T) {
	partial, err := compileUnicodeSymbolPartial(config.SymbolStyleConfig{
		FallbackFont: "Noto Sans Symbols2",
		FallbackFor:  []string{"☑"},
		Replace: map[string]string{
			"☑": `\textbf{DONE}`,
		},
	})
	if err != nil {
		t.Fatalf("compileUnicodeSymbolPartial returned error: %v", err)
	}

	if strings.Contains(partial, `\newunicodechar{☑}{\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}}`) {
		t.Fatalf("did not expect direct fallback definition when explicit replacement overrides it, got %q", partial)
	}
	if !strings.Contains(partial, `\newunicodechar{☑}{\textbf{DONE}}`) {
		t.Fatalf("expected explicit replacement to win, got %q", partial)
	}
}

func TestBuildCodeSymbolNormalizeMapExtractsSymbolGlyphTargets(t *testing.T) {
	mapping := buildCodeSymbolNormalizeMap(config.SymbolStyleConfig{
		Replace: map[string]string{
			"✅":      `\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}`,
			"💡":      `\mdtwosymbolglyph{✦}{\textasteriskcentered}`,
			"→":      `\ensuremath{\rightarrow}`,
			"\uFE0F": ``,
		},
	})

	for source, expected := range map[string]string{
		"✅":      "☑",
		"💡":      "✦",
		"\uFE0F": "",
	} {
		if got, ok := mapping[source]; !ok || got != expected {
			t.Fatalf("expected %q -> %q in code normalization map, got %q (present=%v)", source, expected, got, ok)
		}
	}
	if _, ok := mapping["→"]; ok {
		t.Fatalf("did not expect non-symbol-glyph replacement to be included in code normalization map")
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
		`mdtwotitlecolor`,
		`mdtwoheadingthemecolor`,
		`mdtwolinkcolor`,
		`mdtwotoclinkcolor`,
		`$if(color_primary_value)$`,
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
