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
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "md2pdf-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary markdown file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, err := tmpFile.Write(opts.Markdown); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temporary markdown file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to flush temporary markdown file: %w", err)
	}

	args := []string{
		tmpPath,
		"-o", opts.OutputPath,
		"--from=markdown",
		"--pdf-engine=" + opts.Config.PDF.Engine,
		"--number-sections",
	}

	templatePath := ""
	if opts.Config.PDF.Template != "" {
		templatePath = fs.ResolveOptionalPath(filepath.Dir(opts.InputPath), opts.Config.PDF.Template)
	} else {
		templatePath, err = writeDefaultTemplate()
		if err != nil {
			return fmt.Errorf("failed to prepare default template: %w", err)
		}
		defer os.Remove(templatePath)
	}
	args = append(args, "--template="+templatePath)

	resourcePaths := []string{filepath.Dir(opts.InputPath)}
	for _, item := range opts.Config.Assets.SearchPaths {
		resourcePaths = append(resourcePaths, fs.ResolveOptionalPath(filepath.Dir(opts.InputPath), item))
	}
	args = append(args, "--resource-path="+strings.Join(resourcePaths, string(os.PathListSeparator)))

	args = append(args, metadataArgs(opts.Config)...)

	if opts.EnableTOC {
		args = append(args, "--toc", "--toc-depth="+strconv.Itoa(opts.Config.TOC.Depth))
		if opts.Config.TOC.Title != "" {
			args = append(args, "--metadata", "toc-title="+opts.Config.TOC.Title)
		}
	}
	if opts.EnablePlantUML {
		args = append(args, "--filter", "pandoc-plantuml")
	}

	cmd := exec.CommandContext(ctx, "pandoc", args...)
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

func metadataArgs(cfg config.Config) []string {
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
		pairs = append(pairs, [2]string{"logo_cover", cfg.Assets.LogoCover})
	}
	if cfg.Assets.LogoHeader != "" {
		pairs = append(pairs, [2]string{"logo_header", cfg.Assets.LogoHeader})
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

	out := make([]string, 0, len(pairs)*2)
	for _, pair := range pairs {
		out = append(out, "--metadata", pair[0]+"="+pair[1])
	}
	return out
}

func writeDefaultTemplate() (string, error) {
	tmpFile, err := os.CreateTemp("", "md2pdf-template-*.tex")
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

var tocHeadingPattern = regexp.MustCompile(`(?m)^\s*##{1,3}\s+`) // H2+ headings

func ShouldEnableTOC(mode string, markdown []byte) bool {
	switch mode {
	case "on":
		return true
	case "off":
		return false
	default:
		return tocHeadingPattern.Match(markdown)
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
