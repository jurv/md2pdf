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
- Support multi-column authoring with the bundled `dialoa/columns` Pandoc Lua filter.
- Support paired two-pane layouts with the bundled `side-by-side` Lua filter (`ratio`, `gap`, `valign`).
- Style Markdown blockquotes in the default template (bar color, text color, bar width, spacing).
- Style heading colors and sizes in the default template (`style.headings.*`).
- Configure link colors globally and for ToC links (`style.links.*`).
- Add an optional cover page (builtin, external template, or full-bleed cover image with simple config).
- Add a document-wide background image with scoped activation (`background.*`).
- Configure rich header/footer layouts (multiline text, images, colors, page numbering) with a declarative grid model.

## Prerequisites

Install runtime tools on your machine:

- `pandoc`
- a PDF engine (`xelatex`, `lualatex`, or `pdflatex`)
- optional for diagrams: `pandoc-plantuml`, `plantuml`, `dot`
- optional utilities: `pdftk` (merge), `gs` (compress and `build --compress`)

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

Build and compress in one step:

```bash
md2pdf build notes.md -o notes.pdf --compress --compress-quality ebook
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

1. global config:
   - `--config <path>` if provided
   - otherwise `filepath.Join(os.UserConfigDir(), "md2pdf", "config.yaml")`
   - on Linux, this is typically `$XDG_CONFIG_HOME/md2pdf/config.yaml` or `~/.config/md2pdf/config.yaml`
   - on macOS, this is typically `~/Library/Application Support/md2pdf/config.yaml`
2. project config (`md2pdf.yaml` or `.md2pdf.yaml` in the working directory)
3. document front matter
4. CLI flags (highest priority)

See [docs/configuration.md](docs/configuration.md) for full details.
That page includes front matter syntax, merge behavior, multi-source rules, cover/background examples, and layout filters.

## Columns Layout

`md2pdf` bundles the upstream [`dialoa/columns`](https://github.com/dialoa/columns) Pandoc Lua filter and enables it automatically for all builds. Upstream usage and advanced options are documented in the project README:

- filter repository: <https://github.com/dialoa/columns>
- upstream documentation: <https://github.com/dialoa/columns/blob/master/README.md>

Minimal example:

```markdown
::: columns
::: column
![](assets/offre_a_propos_europe.png)
:::
::: column
Depuis **23 ans**, nous accompagnons nos clients.
:::
:::
```

Notes:

- use fenced Div syntax (`::: columns`, `::: column`), not raw HTML/CSS like `display:flex`
- prefer `![](image.png)` inside columns; standalone image captions create a `figure` environment, which is less predictable inside multi-column LaTeX layouts
- explicit column counts, gaps, rules, ragged columns, and column spans are supported by the upstream filter syntax

## Side-By-Side Layout

For paired content blocks such as “image left, text right”, use the bundled `side-by-side` filter instead of `columns`. It renders to LaTeX `minipage` blocks for PDF and supports width control plus optional vertical centering.

Example with ratio:

```markdown
::: {.side-by-side ratio="38:62" gap=20pt valign=center}
::: left
![](assets/offre_a_propos_europe.png)
:::
::: right
Depuis **23 ans**, nous accompagnons nos clients.
:::
:::
```

Example with percentages:

```markdown
::: {.side-by-side left=38% right=62% gap=20pt valign=center}
::: left
![](assets/offre_a_propos_europe.png)
:::
::: right
Depuis **23 ans**, nous accompagnons nos clients.
:::
:::
```

Notes:

- `valign` supports `top`, `center`, `bottom`
- `align` supports `left`, `center`, `right` and applies inside both panes
- if you include a standalone image with caption inside a pane, the filter flattens it to avoid LaTeX float issues inside `minipage`

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
