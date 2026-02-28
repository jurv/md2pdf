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

type Options struct {
	InputPath        string
	OutputPath       string
	Markdown         []byte
	Config           config.Config
	EnableTOC        bool
	EnablePlantUML   bool
	VerboseExecution bool
}

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
		"--from=markdown",
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

	resourcePaths := []string{filepath.Dir(opts.InputPath)}
	for _, item := range opts.Config.Assets.SearchPaths {
		resourcePaths = append(resourcePaths, fs.ResolveOptionalPath(filepath.Dir(opts.InputPath), item))
	}
	args = append(args, "--resource-path="+strings.Join(resourcePaths, string(os.PathListSeparator)))

	args = append(args, metadataArgs(opts.Config, filepath.Dir(opts.InputPath))...)

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

func metadataArgs(cfg config.Config, baseDir string) []string {
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
	}
	if cfg.Style.Fonts.Body != "" {
		pairs = append(pairs, [2]string{"font_body", cfg.Style.Fonts.Body})
	}
	if cfg.Style.Fonts.Heading != "" {
		pairs = append(pairs, [2]string{"font_heading", cfg.Style.Fonts.Heading})
	}
	switch cfg.Title.RenderMode {
	case "inline":
		pairs = append(pairs, [2]string{"title_render_inline", "true"})
	case "separate_page":
		pairs = append(pairs, [2]string{"title_render_separate_page", "true"})
	case "none":
		pairs = append(pairs, [2]string{"title_render_none", "true"})
	}
	switch cfg.Cover.Mode {
	case "builtin":
		pairs = append(pairs, [2]string{"cover_mode_builtin", "true"})
		if cfg.Cover.Builtin.Logo != "" {
			pairs = append(pairs, [2]string{"cover_logo", fs.ResolveOptionalPath(baseDir, cfg.Cover.Builtin.Logo)})
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
	}

	out := make([]string, 0, len(pairs)*2)
	for _, pair := range pairs {
		out = append(out, "--metadata", pair[0]+"="+pair[1])
	}
	return out
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
