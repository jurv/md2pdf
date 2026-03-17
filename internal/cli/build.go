package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julien/md2pdf/internal/config"
	"github.com/julien/md2pdf/internal/deps"
	"github.com/julien/md2pdf/internal/frontmatter"
	"github.com/julien/md2pdf/internal/fs"
	"github.com/julien/md2pdf/internal/pdf"
	"github.com/julien/md2pdf/internal/render"
	"github.com/spf13/cobra"
)

var (
	generatePDFFunc             = render.GeneratePDF
	compressPDFFunc             = pdf.Compress
	ensureBuildDependenciesFunc = deps.EnsureBuildDependencies
)

type buildFlags struct {
	Output          string
	GlobalConfig    string
	ProjectConfig   string
	Engine          string
	Template        string
	TOCMode         string
	TOCTitle        string
	TOCDepth        int
	Compress        bool
	CompressQuality string
}

func newBuildCmd(global *GlobalOptions) *cobra.Command {
	flags := &buildFlags{}
	cmd := &cobra.Command{
		Use:   "build <input.md>",
		Short: "Build a PDF document from a markdown source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(cmd.Context(), global, cmd, flags, args[0])
		},
	}

	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output PDF path (default: source basename + .pdf)")
	cmd.Flags().StringVar(&flags.GlobalConfig, "config", "", "Global YAML config path")
	cmd.Flags().StringVar(&flags.ProjectConfig, "project-config", "", "Project YAML config path")
	cmd.Flags().StringVar(&flags.Engine, "pdf-engine", "", "PDF engine (xelatex, lualatex, pdflatex)")
	cmd.Flags().StringVar(&flags.Template, "template", "", "Custom LaTeX template path")
	cmd.Flags().StringVar(&flags.TOCMode, "toc", "", "Table of contents mode (auto, on, off)")
	cmd.Flags().StringVar(&flags.TOCTitle, "toc-title", "", "Table of contents title")
	cmd.Flags().IntVar(&flags.TOCDepth, "toc-depth", 0, "Table of contents depth")
	cmd.Flags().BoolVar(&flags.Compress, "compress", false, "Compress the generated PDF with Ghostscript after rendering")
	cmd.Flags().StringVar(&flags.CompressQuality, "compress-quality", "printer", "Compression profile for --compress: screen, ebook, printer, prepress")

	return cmd
}

func runBuild(ctx context.Context, global *GlobalOptions, cmd *cobra.Command, flags *buildFlags, inputArg string) error {
	inputPath, err := filepath.Abs(inputArg)
	if err != nil {
		return userError("invalid input path", err)
	}

	inputContent, err := os.ReadFile(inputPath)
	if err != nil {
		return userError("failed to read input markdown", err)
	}
	frontCfg, inputBody, err := frontmatter.ParseMarkdown(inputContent)
	if err != nil {
		return userError("failed to parse front matter", err)
	}

	overrides := map[string]any{}
	if cmd.Flags().Changed("pdf-engine") {
		config.SetNestedValue(overrides, []string{"pdf", "engine"}, flags.Engine)
	}
	if cmd.Flags().Changed("template") {
		config.SetNestedValue(overrides, []string{"pdf", "template"}, flags.Template)
	}
	if cmd.Flags().Changed("toc") {
		config.SetNestedValue(overrides, []string{"toc", "mode"}, flags.TOCMode)
	}
	if cmd.Flags().Changed("toc-title") {
		config.SetNestedValue(overrides, []string{"toc", "title"}, flags.TOCTitle)
	}
	if cmd.Flags().Changed("toc-depth") {
		config.SetNestedValue(overrides, []string{"toc", "depth"}, flags.TOCDepth)
	}

	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  flags.GlobalConfig,
		ProjectPath: flags.ProjectConfig,
		BaseDir:     filepath.Dir(inputPath),
		FrontMatter: frontCfg,
		Overrides:   overrides,
	})
	if err != nil {
		return userError("configuration error", err)
	}
	if cfg.Heading.MirrorInTOC {
		cfg.TOC.FromLevel = cfg.Heading.FromLevel
		cfg.TOC.ToLevel = cfg.Heading.ToLevel
		cfg.TOC.Depth = cfg.TOC.ToLevel
	}
	if err := render.ValidateHeadingCompatibility(cfg); err != nil {
		return userError("invalid heading configuration", err)
	}

	sources, err := fs.ResolveSources(inputPath, cfg.Sources)
	if err != nil {
		return userError("failed to resolve markdown sources", err)
	}

	entryBody := inputBody
	if cfg.Title.Source == "entrypoint_h1" {
		extractedTitle, stripped, found := render.ExtractFirstH1(inputBody, cfg.Title.StripFromBody)
		if found {
			if cfg.Metadata.Title == "" {
				cfg.Metadata.Title = extractedTitle
			}
			if cfg.Title.StripFromBody {
				entryBody = stripped
			}
		}
	}

	chunks := make([][]byte, 0, len(sources))
	for _, source := range sources {
		var chunk []byte
		if source == inputPath {
			chunk = entryBody
		} else {
			blob, readErr := os.ReadFile(source)
			if readErr != nil {
				return userError("failed to read source file", readErr)
			}
			_, body, parseErr := frontmatter.ParseMarkdown(blob)
			if parseErr != nil {
				return userError(fmt.Sprintf("failed to parse front matter in %s", source), parseErr)
			}
			chunk = body
		}
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
	}

	merged := render.MergeMarkdownChunks(chunks)
	if len(merged) == 0 {
		return userError("empty markdown", fmt.Errorf("resolved sources produced no markdown content"))
	}
	merged, err = render.ApplyHeadingPolicy(merged, cfg)
	if err != nil {
		return userError("failed to apply heading policy", err)
	}

	enableTOC := render.ShouldEnableTOC(cfg.TOC.Mode, merged, cfg.TOC.FromLevel, cfg.TOC.ToLevel)
	enablePlantUML := shouldEnablePlantUML(cfg.Features.PlantUML, merged)

	if err := ensureBuildDependenciesFunc(cfg.PDF.Engine, enablePlantUML); err != nil {
		return dependencyError("dependency check failed", err)
	}

	outputPath := resolveOutputPath(inputPath, flags.Output, cfg.PDF.Output)
	renderOutputPath := outputPath
	if flags.Compress {
		tempFile, tempErr := os.CreateTemp("", "md2pdf-build-*.pdf")
		if tempErr != nil {
			return runtimeError("failed to prepare temporary PDF for compression", tempErr)
		}
		renderOutputPath = tempFile.Name()
		if err := tempFile.Close(); err != nil {
			_ = os.Remove(renderOutputPath)
			return runtimeError("failed to prepare temporary PDF for compression", err)
		}
		defer os.Remove(renderOutputPath)
	}

	err = generatePDFFunc(ctx, render.Options{
		InputPath:        inputPath,
		OutputPath:       renderOutputPath,
		Markdown:         merged,
		Config:           cfg,
		EnableTOC:        enableTOC,
		EnablePlantUML:   enablePlantUML,
		VerboseExecution: global.Verbose || global.Debug,
	})
	if err != nil {
		return runtimeError("build failed", err)
	}

	if flags.Compress {
		if err := compressPDFFunc(renderOutputPath, outputPath, flags.CompressQuality); err != nil {
			if strings.Contains(err.Error(), "ghostscript") {
				return dependencyError("ghostscript is required", err)
			}
			return runtimeError("compression failed", err)
		}
	}

	fmt.Printf("Generated PDF: %s\n", outputPath)
	return nil
}

func shouldEnablePlantUML(mode string, markdown []byte) bool {
	switch mode {
	case "on":
		return true
	case "off":
		return false
	default:
		return render.ContainsPlantUML(markdown)
	}
}

func resolveOutputPath(inputPath, cliOutput, cfgOutput string) string {
	if cliOutput != "" {
		resolved, err := filepath.Abs(cliOutput)
		if err == nil {
			return resolved
		}
		return cliOutput
	}
	if cfgOutput != "" {
		if filepath.IsAbs(cfgOutput) {
			return cfgOutput
		}
		return filepath.Clean(filepath.Join(filepath.Dir(inputPath), cfgOutput))
	}
	ext := filepath.Ext(inputPath)
	return strings.TrimSuffix(inputPath, ext) + ".pdf"
}
