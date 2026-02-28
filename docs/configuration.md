# Configuration

`md2pdf` uses YAML for all configuration layers: global config, project config, and document front matter.

## Merge Rules

When multiple layers define the same key, values are resolved in this order (lowest to highest):

1. global config
2. project config
3. front matter
4. CLI flags

Merge behavior:

- objects are merged recursively
- scalar values replace inherited values
- lists replace inherited lists
- `null` removes an inherited key

## Example

```yaml
pdf:
  engine: xelatex
  template: null
metadata:
  title: null
  author: "Team"
title:
  source: entrypoint_h1
  strip_from_body: true
  render_mode: inline
heading_numbering:
  enabled: true
  from_level: 2
  to_level: 3
  notation:
    "2": decimal
    "3": decimal
toc:
  mode: auto
  title: "Table of Contents"
  from_level: 2
  to_level: 3
  depth: 3
cover:
  mode: none
sources:
  explicit:
    - "000-intro.md"
  include:
    - "sections/*.md"
assets:
  search_paths:
    - "assets"
  logo_cover: null
  logo_header: "assets/logo.png"
features:
  plantuml: auto
```

## Document Front Matter

Front matter is the YAML block at the top of a Markdown file. In `md2pdf`, it uses the same schema as regular config files (`pdf`, `metadata`, `title`, `heading_numbering`, `toc`, `cover`, `sources`, `assets`, `style`, `header_footer`, `features`).

Parsing rules:

1. the file must start with `---` on the first line
2. front matter ends with `---` or `...`
3. if no valid front matter block is found, the file is treated as plain Markdown

Example:

```markdown
---
metadata:
  title: "Weekly Report"
  author: "Delivery Team"
toc:
  mode: on
  title: "Contents"
  depth: 2
pdf:
  engine: xelatex
---

# Weekly Report

## Status
...
```

Because front matter is part of the cascade, it can override project/global values, and `null` can unset inherited keys:

```markdown
---
assets:
  logo_header: null
---
```

## Front Matter Key Reference

The front matter schema is identical to the config file schema. The table below documents each supported key, expected type, defaults, validation rules, and runtime effect.

| Key                                  | Type             | Default             | Allowed values / validation                                           | Runtime effect                                                                                                   |
|--------------------------------------|------------------|---------------------|-----------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `pdf.engine`                         | string           | `xelatex`           | `xelatex`, `lualatex`, `pdflatex`                                     | Selects the Pandoc PDF engine (`--pdf-engine`).                                                                  |
| `pdf.template`                       | string or `null` | empty               | File path (absolute or relative to entry markdown)                    | If set, uses custom template; otherwise uses embedded default template.                                          |
| `pdf.output`                         | string or `null` | empty               | File path                                                             | Default output path when `-o/--output` is not provided.                                                          |
| `metadata.title`                     | string or `null` | empty               | Any string                                                            | Explicit document title metadata.                                                                                |
| `metadata.author`                    | string or `null` | empty               | Any string                                                            | Passed to Pandoc metadata as `author`.                                                                           |
| `metadata.subject`                   | string or `null` | empty               | Any string                                                            | Passed to Pandoc metadata as `subject`.                                                                          |
| `title.source`                       | string           | `entrypoint_h1`     | `entrypoint_h1`, `metadata_only`, `none`                              | Chooses where `title` comes from; default extracts first `#` from entry markdown when `metadata.title` is empty. |
| `title.strip_from_body`              | boolean          | `true`              | `true` or `false`                                                     | Removes extracted entrypoint `#` heading from document body.                                                     |
| `title.render_mode`                  | string           | `inline`            | `inline`, `separate_page`, `none`                                     | Controls title rendering in the default template.                                                                |
| `heading_numbering.enabled`          | boolean          | `true`              | `true` or `false`                                                     | Enables heading number prefix generation in markdown before Pandoc.                                              |
| `heading_numbering.from_level`       | integer          | `2`                 | `1..6`, `<= to_level`                                                 | First heading level that receives numbering.                                                                     |
| `heading_numbering.to_level`         | integer          | `3`                 | `1..6`, `>= from_level`                                               | Last heading level that receives numbering.                                                                      |
| `heading_numbering.separator`        | string           | `.`                 | Any string                                                            | Separator between numbering components (example: `1.2.3`).                                                       |
| `heading_numbering.suffix`           | string           | empty               | Any string                                                            | Suffix appended to numbering prefix (example: `)` gives `1.2)`).                                                 |
| `heading_numbering.mirror_in_toc`    | boolean          | `true`              | `true` or `false`                                                     | If `true`, ToC bounds are forced to heading numbering bounds.                                                    |
| `heading_numbering.notation.<level>` | string           | `decimal`           | `decimal`, `roman_upper`, `roman_lower`, `alpha_upper`, `alpha_lower` | Per-level numbering style (`<level>` is `1` to `6`).                                                             |
| `toc.mode`                           | string           | `auto`              | `auto`, `on`, `off`                                                   | Controls ToC generation. `auto` enables ToC when headings are detected.                                          |
| `toc.title`                          | string or `null` | `Table of Contents` | Any string                                                            | Passed as `toc-title` metadata when ToC is enabled.                                                              |
| `toc.from_level`                     | integer          | `2`                 | `1..6`, `<= to_level`                                                 | Minimum heading level included in ToC auto-detection and filtering.                                              |
| `toc.to_level`                       | integer          | `3`                 | `1..6`, `>= from_level`                                               | Maximum heading level included in ToC and passed as Pandoc `--toc-depth`.                                        |
| `toc.depth`                          | integer          | `3`                 | Must be `> 0`                                                         | Backward-compatible alias of `toc.to_level`.                                                                     |
| `cover.mode`                         | string           | `none`              | `none`, `builtin`, `external_template`                                | Adds an optional cover page before title/body.                                                                   |
| `cover.external_template`            | string or `null` | empty               | File path                                                             | Required when `cover.mode: external_template`; LaTeX file included with `\\input`.                               |
| `cover.builtin.logo`                 | string or `null` | empty               | File path                                                             | Logo path for builtin cover.                                                                                     |
| `cover.builtin.title_color`          | string           | `#000000`           | color name or `#RRGGBB`                                               | Title color on builtin cover.                                                                                    |
| `cover.builtin.subtitle`             | string or `null` | empty               | Any string                                                            | Subtitle text on builtin cover.                                                                                  |
| `cover.builtin.background_color`     | string           | `#FFFFFF`           | color name or `#RRGGBB`                                               | Background color on builtin cover.                                                                               |
| `cover.builtin.align`                | string           | `center`            | `center`, `top`                                                       | Vertical alignment of builtin cover content.                                                                     |
| `sources.explicit`                   | list of strings  | `[]`                | Existing files expected                                               | Ordered source list, processed first.                                                                            |
| `sources.include`                    | list of strings  | `[]`                | Glob patterns                                                         | Additional sources matched by glob, sorted alphabetically, de-duplicated by canonical path.                      |
| `assets.search_paths`                | list of strings  | `[]`                | Path list                                                             | Appended to Pandoc `--resource-path`.                                                                            |
| `assets.logo_cover`                  | string or `null` | empty               | Path or identifier                                                    | Passed as metadata key `logo_cover` (used only if template consumes it).                                         |
| `assets.logo_header`                 | string or `null` | empty               | Path or identifier                                                    | Passed as metadata key `logo_header` (used only if template consumes it).                                        |
| `style.colors.primary`               | string or `null` | empty               | Any string (typically hex color)                                      | Passed as metadata key `color_primary` (template-dependent).                                                     |
| `style.fonts.body`                   | string or `null` | empty               | Any string                                                            | Passed as metadata key `font_body` (template-dependent).                                                         |
| `style.fonts.heading`                | string or `null` | empty               | Any string                                                            | Passed as metadata key `font_heading` (template-dependent).                                                      |
| `header_footer.header_left`          | string or `null` | empty               | Any string                                                            | Parsed and merged, currently not consumed by renderer/template.                                                  |
| `header_footer.header_right`         | string or `null` | empty               | Any string                                                            | Parsed and merged, currently not consumed by renderer/template.                                                  |
| `header_footer.footer_left`          | string or `null` | empty               | Any string                                                            | Parsed and merged, currently not consumed by renderer/template.                                                  |
| `header_footer.footer_right`         | string or `null` | empty               | Any string                                                            | Parsed and merged, currently not consumed by renderer/template.                                                  |
| `features.plantuml`                  | string           | `auto`              | `auto`, `on`, `off`                                                   | Controls PlantUML filter activation and dependency requirements.                                                 |

### Default heading behavior

Out of the box, the first entrypoint `#` heading is used as the document title and removed from body content. Heading numbering starts at `##` and ends at `###` (`1`, `1.1`, `2`, ...). By default, ToC bounds mirror numbering bounds (`h2..h3`), so `#` and `####+` headings are automatically marked as unlisted.

### Schema strictness

At the moment, unknown keys are ignored. To avoid silent mistakes, prefer keeping front matter limited to the keys listed above.

### Full front matter example

```markdown
---
pdf:
  engine: xelatex
  template: null
  output: null
metadata:
  title: null
  author: "Project Team"
  subject: "Internal Document"
title:
  source: entrypoint_h1
  strip_from_body: true
  render_mode: inline
heading_numbering:
  enabled: true
  from_level: 2
  to_level: 3
  separator: "."
  suffix: ""
  mirror_in_toc: true
  notation:
    "2": decimal
    "3": decimal
toc:
  mode: auto
  title: "Contents"
  from_level: 2
  to_level: 3
  depth: 3
cover:
  mode: none
  external_template: null
  builtin:
    logo: null
    title_color: "#000000"
    subtitle: null
    background_color: "#FFFFFF"
    align: center
sources:
  explicit:
    - "000-intro.md"
    - "010-scope.md"
  include:
    - "sections/*.md"
assets:
  search_paths:
    - "assets"
  logo_cover: "assets/logo-cover.png"
  logo_header: "assets/logo-header.png"
style:
  colors:
    primary: "#1F4E79"
  fonts:
    body: "Open Sans"
    heading: "Open Sans"
header_footer:
  header_left: null
  header_right: null
  footer_left: null
  footer_right: null
features:
  plantuml: auto
---
```

## Front Matter in Multi-source Mode

When `sources.explicit` and/or `sources.include` are used, `md2pdf` resolves and merges the Markdown files in deterministic order.

Important behavior:

1. only the entrypoint file front matter is used for configuration merging
2. front matter found in included source files is stripped from rendered content
3. included source front matter does not contribute additional config overrides

This keeps configuration deterministic and avoids hidden per-file overrides in large document sets.

## Multi-source Ordering

Source order is deterministic:

1. entries in `sources.explicit` in the exact declared order
2. entries matched by `sources.include` sorted alphabetically
3. duplicate files removed by canonical path

## Rendering Safety Defaults

The default embedded template includes behavior-focused safeguards to reduce rendering regressions across documents.

- Images are globally constrained to page-safe bounds: max width is text width, max height is `0.8 * page height`, aspect ratio preserved.
- Figures are forced near source location to avoid unexpected floating far from related text.
- Long code lines are soft-wrapped and line-numbered to avoid horizontal overflow.
- Pandoc table helper macros are enabled for safer table rendering.
- Common Unicode symbols are mapped to safe LaTeX equivalents to reduce build failures.

These defaults are active when `pdf.template` is not set. If you provide a custom template, you are responsible for carrying equivalent safeguards.

## PlantUML Artifact Handling

When PlantUML rendering is enabled, `md2pdf` runs Pandoc in an isolated temporary workspace. This prevents `plantuml-images` artifacts from being created in your source directory and keeps generated intermediates out of versioned content.

The workspace is removed after build completion (success or failure).
