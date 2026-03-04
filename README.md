# md2pdf

`md2pdf` is a standalone CLI that generates PDF documents from Markdown using layered YAML configuration. It replaces project-specific shell scripts with a single, predictable workflow across Linux, macOS, and Windows.

## Features

- Build PDF files from single or multi-source Markdown inputs.
- Merge configuration from global, project, and document front matter.
- Control PDF engine, template, title extraction, heading numbering, table of contents, metadata, and assets.
- Render PlantUML blocks when the required toolchain is available.
- Run dependency diagnostics with `doctor`.
- Merge and compress existing PDF files.
- Apply safe default rendering guards (image auto-fit, code wrapping, and Unicode fallbacks).
- Style Markdown blockquotes in the default template (bar color, text color, bar width, spacing).
- Style heading colors and sizes in the default template (`style.headings.*`).
- Configure link colors globally and for ToC links (`style.links.*`).
- Add an optional cover page (builtin, external template, or full-bleed cover image with simple config).
- Configure rich header/footer layouts (multiline text, images, colors, page numbering) with a declarative grid model.

## Prerequisites

Install runtime tools on your machine:

- `pandoc`
- a PDF engine (`xelatex`, `lualatex`, or `pdflatex`)
- optional for diagrams: `pandoc-plantuml`, `plantuml`, `dot`
- optional utilities: `pdftk` (merge), `gs` (compress)

Use `md2pdf doctor` to validate your environment.

## Build from Source

```bash
go build -o md2pdf ./cmd/md2pdf
```

## Install

### Option 1: local install from source

```bash
go install ./cmd/md2pdf
```

### Option 2: prebuilt binaries

Download the matching binary for your OS/architecture from releases and place it in your `PATH`.

## Quick Start

Build a single Markdown file:

```bash
md2pdf build notes.md
```

Build with explicit output and TOC override:

```bash
md2pdf build notes.md -o notes.pdf --toc on --toc-title "Contents" --toc-depth 3
```

Check dependencies:

```bash
md2pdf doctor
md2pdf doctor --json
```

Merge PDFs:

```bash
md2pdf merge part1.pdf part2.pdf -o merged.pdf
```

Compress a PDF:

```bash
md2pdf compress merged.pdf -o merged-small.pdf --quality ebook
```

Generate a starter config:

```bash
md2pdf init --profile report
```

## Configuration

Configuration is YAML-based and supports cascade merging:

1. global config (`$XDG_CONFIG_HOME/md2pdf/config.yaml` or `~/.config/md2pdf/config.yaml`)
2. project config (`md2pdf.yaml` or `.md2pdf.yaml` in the working directory)
3. document front matter
4. CLI flags (highest priority)

See [docs/configuration.md](docs/configuration.md) for full details.
That page includes front matter syntax, merge behavior, and multi-source rules.

## Troubleshooting

Use `md2pdf doctor` first for missing dependencies and version checks. Additional guidance is available in [docs/troubleshooting.md](docs/troubleshooting.md).

## Minimal Code Map

- `cmd/md2pdf/main.go`: process entrypoint.
- `internal/cli`: Cobra commands and error codes.
- `internal/config`: YAML schema, merge logic, and validation.
- `internal/frontmatter`: Markdown front matter extraction.
- `internal/fs`: source and path resolution.
- `internal/render`: Pandoc orchestration and template handling.
- `internal/deps`: dependency inspection and `doctor` status.
- `internal/pdf`: merge and compression backends.
