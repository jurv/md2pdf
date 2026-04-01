package render

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/julien/md2pdf/internal/config"
	"github.com/julien/md2pdf/internal/fs"
)

//go:embed templates/default.tex
var defaultTemplate string

//go:embed templates/table_code_wrap.lua
var tableCodeWrapFilter string

//go:embed templates/columns.lua
var columnsFilter string

//go:embed templates/side_by_side.lua
var sideBySideFilter string

type Options struct {
	InputPath        string
	OutputPath       string
	Markdown         []byte
	Config           config.Config
	EnableTOC        bool
	EnablePlantUML   bool
	VerboseExecution bool
}

// Accept common markdown written without an empty line before a blockquote or
// an ATX heading. A lot of existing project documents rely on that looser
// syntax, and keeping Pandoc's stricter blank-before-header rule causes later
// headings to be swallowed into the previous paragraph.
const pandocInputFormat = "markdown-blank_before_blockquote-blank_before_header"

func GeneratePDF(ctx context.Context, opts Options) error {
	if len(opts.Markdown) == 0 {
		return fmt.Errorf("cannot render empty markdown input")
	}
	outputPath := opts.OutputPath
	if !filepath.IsAbs(outputPath) {
		absOut, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to resolve output path: %w", err)
		}
		outputPath = absOut
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Run pandoc from an isolated temp workspace so transient artifacts
	// (like plantuml-images/) never pollute the user's document directory.
	workDir, err := os.MkdirTemp("", "md2pdf-work-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	tmpPath := filepath.Join(workDir, "input.md")
	if err := os.WriteFile(tmpPath, opts.Markdown, 0o600); err != nil {
		return fmt.Errorf("failed to write temporary markdown file: %w", err)
	}
	args := []string{
		tmpPath,
		"-o", outputPath,
		// Accept blockquotes in list items even without an extra blank line.
		"--from=" + pandocInputFormat,
		"--pdf-engine=" + opts.Config.PDF.Engine,
	}

	templatePath := ""
	if opts.Config.PDF.Template != "" {
		templatePath = fs.ResolveOptionalPath(filepath.Dir(opts.InputPath), opts.Config.PDF.Template)
	} else {
		templatePath, err = writeDefaultTemplate(workDir)
		if err != nil {
			return fmt.Errorf("failed to prepare default template: %w", err)
		}
	}
	args = append(args, "--template="+templatePath)

	tableCodeWrapFilterPath, err := writeTableCodeWrapFilter(workDir)
	if err != nil {
		return fmt.Errorf("failed to prepare inline code table filter: %w", err)
	}
	args = append(args, "--lua-filter="+tableCodeWrapFilterPath)

	codeSymbolNormalizeFilterPath, err := writeCodeSymbolNormalizeFilter(workDir, opts.Config.Style.Symbols)
	if err != nil {
		return fmt.Errorf("failed to prepare code symbol normalization filter: %w", err)
	}
	if codeSymbolNormalizeFilterPath != "" {
		args = append(args, "--lua-filter="+codeSymbolNormalizeFilterPath)
	}

	columnsFilterPath, err := writeColumnsFilter(workDir)
	if err != nil {
		return fmt.Errorf("failed to prepare columns filter: %w", err)
	}
	args = append(args, "--lua-filter="+columnsFilterPath)

	sideBySideFilterPath, err := writeSideBySideFilter(workDir)
	if err != nil {
		return fmt.Errorf("failed to prepare side-by-side filter: %w", err)
	}
	args = append(args, "--lua-filter="+sideBySideFilterPath)

	resourcePaths := []string{filepath.Dir(opts.InputPath)}
	for _, item := range opts.Config.Assets.SearchPaths {
		resourcePaths = append(resourcePaths, fs.ResolveOptionalPath(filepath.Dir(opts.InputPath), item))
	}
	args = append(args, "--resource-path="+strings.Join(resourcePaths, string(os.PathListSeparator)))

	metadata, err := metadataArgs(opts.Config, filepath.Dir(opts.InputPath), opts.EnableTOC, workDir)
	if err != nil {
		return fmt.Errorf("failed to prepare metadata: %w", err)
	}
	args = append(args, metadata...)

	if opts.EnableTOC {
		args = append(args, "--toc", "--toc-depth="+strconv.Itoa(opts.Config.TOC.ToLevel))
		if opts.Config.TOC.Title != "" {
			args = append(args, "--metadata", "toc-title="+opts.Config.TOC.Title)
		}
	}
	if opts.EnablePlantUML {
		args = append(args, "--filter", "pandoc-plantuml")
	}

	cmd := exec.CommandContext(ctx, "pandoc", args...)
	cmd.Dir = workDir
	if opts.EnablePlantUML {
		cmd.Env = withHeadlessJavaEnv(os.Environ())
	}
	if opts.VerboseExecution {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pandoc failed: %w", err)
		}
		return nil
	}

	combined, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc failed: %w (%s)", err, strings.TrimSpace(string(combined)))
	}
	return nil
}

func metadataArgs(cfg config.Config, baseDir string, tocEnabled bool, workDir string) ([]string, error) {
	pairs := make([][2]string, 0)
	pairs = append(pairs, [2]string{"link-citations", "true"})
	if cfg.Metadata.Title != "" {
		pairs = append(pairs, [2]string{"title", cfg.Metadata.Title})
	}
	if cfg.Metadata.Author != "" {
		pairs = append(pairs, [2]string{"author", cfg.Metadata.Author})
	}
	if cfg.Metadata.Subject != "" {
		pairs = append(pairs, [2]string{"subject", cfg.Metadata.Subject})
	}
	if cfg.Assets.LogoCover != "" {
		pairs = append(pairs, [2]string{"logo_cover", fs.ResolveOptionalPath(baseDir, cfg.Assets.LogoCover)})
	}
	if cfg.Assets.LogoHeader != "" {
		pairs = append(pairs, [2]string{"logo_header", fs.ResolveOptionalPath(baseDir, cfg.Assets.LogoHeader)})
	}
	if cfg.Style.Colors.Primary != "" {
		pairs = append(pairs, [2]string{"color_primary", cfg.Style.Colors.Primary})
		model, value := latexColor(cfg.Style.Colors.Primary)
		pairs = append(pairs, [2]string{"color_primary_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"color_primary_model", model})
		}
	}
	if cfg.Style.Fonts.Body != "" {
		pairs = append(pairs, [2]string{"font_body", cfg.Style.Fonts.Body})
	}
	if cfg.Style.Fonts.Heading != "" {
		pairs = append(pairs, [2]string{"font_heading", cfg.Style.Fonts.Heading})
	}
	if cfg.Style.Links.Color != "" {
		model, value := latexColor(cfg.Style.Links.Color)
		pairs = append(pairs, [2]string{"hyperref_link_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"hyperref_link_color_model", model})
		}
	}
	if cfg.Style.Links.URLColor != "" {
		model, value := latexColor(cfg.Style.Links.URLColor)
		pairs = append(pairs, [2]string{"hyperref_url_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"hyperref_url_color_model", model})
		}
	}
	if cfg.Style.Links.CitationColor != "" {
		model, value := latexColor(cfg.Style.Links.CitationColor)
		pairs = append(pairs, [2]string{"hyperref_cite_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"hyperref_cite_color_model", model})
		}
	}
	if cfg.Style.Links.TOCColor != "" {
		model, value := latexColor(cfg.Style.Links.TOCColor)
		pairs = append(pairs, [2]string{"hyperref_toc_link_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"hyperref_toc_link_color_model", model})
		}
	}
	if !cfg.Style.Figures.CaptionEnabled {
		pairs = append(pairs, [2]string{"figure_caption_hidden", "true"})
	}
	pairs = append(pairs, [2]string{"table_row_spacing_factor", strconv.FormatFloat(cfg.Style.Tables.RowSpacingFactor, 'f', -1, 64)})
	if cfg.Style.Tables.ZebraEnabled {
		pairs = append(pairs, [2]string{"table_zebra_enabled", "true"})
	}
	if strings.TrimSpace(cfg.Style.Tables.ZebraColor) != "" {
		model, value := latexColor(cfg.Style.Tables.ZebraColor)
		pairs = append(pairs, [2]string{"table_zebra_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"table_zebra_color_model", model})
		}
	}
	headingStyleEnabled := strings.TrimSpace(cfg.Style.Fonts.Heading) != "" || strings.TrimSpace(cfg.Style.Colors.Primary) != ""
	type headingStyleLevel struct {
		key string
		cfg config.HeadingLevelStyleConfig
	}
	for _, level := range []headingStyleLevel{
		{key: "h1", cfg: cfg.Style.Headings.H1},
		{key: "h2", cfg: cfg.Style.Headings.H2},
		{key: "h3", cfg: cfg.Style.Headings.H3},
		{key: "h4", cfg: cfg.Style.Headings.H4},
		{key: "h5", cfg: cfg.Style.Headings.H5},
		{key: "h6", cfg: cfg.Style.Headings.H6},
	} {
		if strings.TrimSpace(level.cfg.Color) != "" {
			model, value := latexColor(level.cfg.Color)
			pairs = append(pairs, [2]string{"heading_" + level.key + "_color_value", value})
			if model != "" {
				pairs = append(pairs, [2]string{"heading_" + level.key + "_color_model", model})
			}
			headingStyleEnabled = true
		}
		if level.cfg.SizePt != nil {
			pairs = append(pairs, [2]string{"heading_" + level.key + "_size_pt", strconv.FormatFloat(*level.cfg.SizePt, 'f', -1, 64)})
			headingStyleEnabled = true
		}
		if level.cfg.SpaceBeforePt != nil {
			pairs = append(pairs, [2]string{"heading_" + level.key + "_space_before_pt", strconv.FormatFloat(*level.cfg.SpaceBeforePt, 'f', -1, 64)})
			headingStyleEnabled = true
		}
		if level.cfg.SpaceAfterPt != nil {
			pairs = append(pairs, [2]string{"heading_" + level.key + "_space_after_pt", strconv.FormatFloat(*level.cfg.SpaceAfterPt, 'f', -1, 64)})
			headingStyleEnabled = true
		}
	}
	if headingStyleEnabled {
		pairs = append(pairs, [2]string{"heading_style_enabled", "true"})
	}
	if cfg.Style.BlockQuote.BarColor != "" {
		model, value := latexColor(cfg.Style.BlockQuote.BarColor)
		pairs = append(pairs, [2]string{"blockquote_bar_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"blockquote_bar_color_model", model})
		}
	}
	if cfg.Style.BlockQuote.TextColor != "" {
		model, value := latexColor(cfg.Style.BlockQuote.TextColor)
		pairs = append(pairs, [2]string{"blockquote_text_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"blockquote_text_color_model", model})
		}
	}
	if cfg.Style.BlockQuote.BackgroundColor != "" {
		model, value := latexColor(cfg.Style.BlockQuote.BackgroundColor)
		pairs = append(pairs, [2]string{"blockquote_background_color_value", value})
		if model != "" {
			pairs = append(pairs, [2]string{"blockquote_background_color_model", model})
		}
	}
	pairs = append(pairs,
		[2]string{"blockquote_bar_width_pt", strconv.FormatFloat(cfg.Style.BlockQuote.BarWidthPt, 'f', -1, 64)},
		[2]string{"blockquote_gap_pt", strconv.FormatFloat(cfg.Style.BlockQuote.GapPt, 'f', -1, 64)},
		[2]string{"blockquote_padding_pt", strconv.FormatFloat(cfg.Style.BlockQuote.PaddingPt, 'f', -1, 64)},
	)
	pairs = append(pairs,
		[2]string{"plantuml_align", cfg.Style.PlantUML.Align},
		[2]string{"plantuml_space_before_pt", strconv.FormatFloat(cfg.Style.PlantUML.SpaceBeforePt, 'f', -1, 64)},
		[2]string{"plantuml_space_after_pt", strconv.FormatFloat(cfg.Style.PlantUML.SpaceAfterPt, 'f', -1, 64)},
	)
	switch cfg.Title.RenderMode {
	case "inline":
		pairs = append(pairs, [2]string{"title_render_inline", "true"})
	case "separate_page":
		pairs = append(pairs, [2]string{"title_render_separate_page", "true"})
	case "none":
		pairs = append(pairs, [2]string{"title_render_none", "true"})
	}
	backgroundImage := strings.TrimSpace(cfg.Background.Image)
	if backgroundImage != "" {
		pairs = append(pairs, [2]string{"background_image", fs.ResolveOptionalPath(baseDir, backgroundImage)})
		switch effectiveBackgroundApplyOn(cfg) {
		case "toc_and_body":
			pairs = append(pairs, [2]string{"background_apply_on_toc_and_body", "true"})
		case "body_only":
			pairs = append(pairs, [2]string{"background_apply_on_body_only", "true"})
		default:
			pairs = append(pairs, [2]string{"background_apply_on_all_pages", "true"})
		}
		switch effectiveBackgroundImageFit(cfg) {
		case "contain":
			pairs = append(pairs, [2]string{"background_image_fit_contain", "true"})
		case "stretch":
			pairs = append(pairs, [2]string{"background_image_fit_stretch", "true"})
		default:
			pairs = append(pairs, [2]string{"background_image_fit_cover", "true"})
		}
	}
	coverMode := strings.TrimSpace(cfg.Cover.Mode)
	coverImage := strings.TrimSpace(cfg.Cover.Image)
	switch coverMode {
	case "builtin":
		pairs = append(pairs, [2]string{"cover_mode_builtin", "true"})
		if coverImage != "" {
			pairs = append(pairs, [2]string{"cover_image", fs.ResolveOptionalPath(baseDir, coverImage)})
			switch effectiveCoverImageFit(cfg) {
			case "contain":
				pairs = append(pairs, [2]string{"cover_image_fit_contain", "true"})
			case "stretch":
				pairs = append(pairs, [2]string{"cover_image_fit_stretch", "true"})
			default:
				pairs = append(pairs, [2]string{"cover_image_fit_cover", "true"})
			}
		}
		coverLogo := strings.TrimSpace(cfg.Cover.Builtin.Logo)
		if coverLogo == "" {
			coverLogo = strings.TrimSpace(cfg.Assets.LogoCover)
		}
		if coverLogo != "" {
			pairs = append(pairs, [2]string{"cover_logo", fs.ResolveOptionalPath(baseDir, coverLogo)})
		}
		if cfg.Cover.Builtin.TitleColor != "" {
			model, value := latexColor(cfg.Cover.Builtin.TitleColor)
			pairs = append(pairs, [2]string{"cover_title_color_value", value})
			if model != "" {
				pairs = append(pairs, [2]string{"cover_title_color_model", model})
			}
		}
		if cfg.Cover.Builtin.Subtitle != "" {
			pairs = append(pairs, [2]string{"cover_subtitle", cfg.Cover.Builtin.Subtitle})
		}
		if cfg.Cover.Builtin.BackgroundColor != "" {
			model, value := latexColor(cfg.Cover.Builtin.BackgroundColor)
			pairs = append(pairs, [2]string{"cover_background_color_value", value})
			if model != "" {
				pairs = append(pairs, [2]string{"cover_background_color_model", model})
			}
		}
		if cfg.Cover.Builtin.Align == "top" {
			pairs = append(pairs, [2]string{"cover_align_top", "true"})
		} else {
			pairs = append(pairs, [2]string{"cover_align_center", "true"})
		}
	case "external_template":
		pairs = append(pairs, [2]string{"cover_mode_external", "true"})
		if cfg.Cover.ExternalTemplate != "" {
			pairs = append(pairs, [2]string{"cover_external_template", fs.ResolveOptionalPath(baseDir, cfg.Cover.ExternalTemplate)})
		}
	default:
		// Simple cover image mode: use image as full-bleed first-page background
		// without inserting a dedicated extra cover page.
		if coverImage != "" {
			pairs = append(pairs,
				[2]string{"cover_mode_first_page_background", "true"},
				[2]string{"cover_image", fs.ResolveOptionalPath(baseDir, coverImage)},
			)
			switch effectiveCoverImageFit(cfg) {
			case "contain":
				pairs = append(pairs, [2]string{"cover_image_fit_contain", "true"})
			case "stretch":
				pairs = append(pairs, [2]string{"cover_image_fit_stretch", "true"})
			default:
				pairs = append(pairs, [2]string{"cover_image_fit_cover", "true"})
			}
		}
	}
	hfPairs, err := buildHeaderFooterMetadata(cfg, baseDir, workDir, tocEnabled)
	if err != nil {
		return nil, err
	}
	pairs = append(pairs, hfPairs...)
	symbolPairs, err := buildUnicodeSymbolMetadata(cfg, workDir)
	if err != nil {
		return nil, err
	}
	pairs = append(pairs, symbolPairs...)

	out := make([]string, 0, len(pairs)*2)
	for _, pair := range pairs {
		out = append(out, "--metadata", pair[0]+"="+pair[1])
	}
	return out, nil
}

func writeDefaultTemplate(workDir string) (string, error) {
	tmpFile, err := os.CreateTemp(workDir, "md2pdf-template-*.tex")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(defaultTemplate); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func writeTableCodeWrapFilter(workDir string) (string, error) {
	tmpFile, err := os.CreateTemp(workDir, "md2pdf-table-code-wrap-*.lua")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(tableCodeWrapFilter); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func writeColumnsFilter(workDir string) (string, error) {
	tmpFile, err := os.CreateTemp(workDir, "md2pdf-columns-*.lua")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(columnsFilter); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func writeSideBySideFilter(workDir string) (string, error) {
	tmpFile, err := os.CreateTemp(workDir, "md2pdf-side-by-side-*.lua")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(sideBySideFilter); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

var plantUMLFence = regexp.MustCompile("(?m)^\\s*```\\s*plantuml\\b")
var plantUMLStart = regexp.MustCompile("(?m)^\\s*@startuml\\b")

func ContainsPlantUML(markdown []byte) bool {
	return plantUMLFence.Match(markdown) || plantUMLStart.Match(markdown)
}

func ShouldEnableTOC(mode string, markdown []byte, fromLevel, toLevel int) bool {
	switch mode {
	case "on":
		return true
	case "off":
		return false
	default:
		lines := strings.Split(strings.ReplaceAll(string(markdown), "\r\n", "\n"), "\n")
		insideFence := false
		fenceMarker := ""
		for _, line := range lines {
			if marker, ok := updateFenceState(line, insideFence, fenceMarker); ok {
				if insideFence {
					insideFence = false
					fenceMarker = ""
				} else {
					insideFence = true
					fenceMarker = marker
				}
				continue
			}
			if insideFence {
				continue
			}
			level, _, _, _, ok := parseHeadingLine(line)
			if !ok {
				continue
			}
			if level >= fromLevel && level <= toLevel {
				return true
			}
		}
		return false
	}
}

func MergeMarkdownChunks(chunks [][]byte) []byte {
	if len(chunks) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for i, chunk := range chunks {
		buf.Write(bytes.TrimSpace(chunk))
		if i < len(chunks)-1 {
			buf.WriteString("\n\n")
		}
	}
	buf.WriteString("\n")
	return buf.Bytes()
}

func withHeadlessJavaEnv(base []string) []string {
	out := make([]string, 0, len(base)+1)
	found := false
	for _, item := range base {
		if strings.HasPrefix(item, "JAVA_TOOL_OPTIONS=") {
			found = true
			current := strings.TrimPrefix(item, "JAVA_TOOL_OPTIONS=")
			if !strings.Contains(current, "-Djava.awt.headless=true") {
				if strings.TrimSpace(current) == "" {
					current = "-Djava.awt.headless=true"
				} else {
					current = current + " -Djava.awt.headless=true"
				}
			}
			out = append(out, "JAVA_TOOL_OPTIONS="+current)
			continue
		}
		out = append(out, item)
	}
	if !found {
		out = append(out, "JAVA_TOOL_OPTIONS=-Djava.awt.headless=true")
	}
	return out
}

func effectiveCoverImageFit(cfg config.Config) string {
	switch strings.TrimSpace(cfg.Cover.ImageFit) {
	case "contain":
		return "contain"
	case "stretch":
		return "stretch"
	default:
		return "cover"
	}
}

func effectiveBackgroundImageFit(cfg config.Config) string {
	switch strings.TrimSpace(cfg.Background.ImageFit) {
	case "contain":
		return "contain"
	case "stretch":
		return "stretch"
	default:
		return "cover"
	}
}

func effectiveBackgroundApplyOn(cfg config.Config) string {
	switch strings.TrimSpace(cfg.Background.ApplyOn) {
	case "toc_and_body":
		return "toc_and_body"
	case "body_only":
		return "body_only"
	default:
		return "all_pages"
	}
}

var hexColorRE = regexp.MustCompile(`^#([0-9A-Fa-f]{6})$`)

func latexColor(value string) (model string, out string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ""
	}
	match := hexColorRE.FindStringSubmatch(trimmed)
	if match == nil {
		return "", trimmed
	}
	return "HTML", strings.ToUpper(match[1])
}
