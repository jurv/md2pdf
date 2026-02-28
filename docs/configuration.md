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
  title: "Project Notes"
  author: "Team"
toc:
  mode: auto
  title: "Table of Contents"
  depth: 3
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

Front matter is the YAML block at the top of a Markdown file. In `md2pdf`, it uses the same schema as regular config files (`pdf`, `metadata`, `toc`, `sources`, `assets`, `style`, `header_footer`, `features`).

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

| Key | Type | Default | Allowed values / validation | Runtime effect |
| --- | --- | --- | --- | --- |
| `pdf.engine` | string | `xelatex` | `xelatex`, `lualatex`, `pdflatex` | Selects the Pandoc PDF engine (`--pdf-engine`). |
| `pdf.template` | string or `null` | empty | File path (absolute or relative to entry markdown) | If set, uses custom template; otherwise uses embedded default template. |
| `pdf.output` | string or `null` | empty | File path | Default output path when `-o/--output` is not provided. |
| `metadata.title` | string or `null` | empty | Any string | Passed to Pandoc metadata as `title`. |
| `metadata.author` | string or `null` | empty | Any string | Passed to Pandoc metadata as `author`. |
| `metadata.subject` | string or `null` | empty | Any string | Passed to Pandoc metadata as `subject`. |
| `toc.mode` | string | `auto` | `auto`, `on`, `off` | Controls ToC generation. `auto` enables ToC when headings are detected. |
| `toc.title` | string or `null` | `Table of Contents` | Any string | Passed as `toc-title` metadata when ToC is enabled. |
| `toc.depth` | integer | `3` | Must be `> 0` | Passed to Pandoc as `--toc-depth`. |
| `sources.explicit` | list of strings | `[]` | Existing files expected | Ordered source list, processed first. |
| `sources.include` | list of strings | `[]` | Glob patterns | Additional sources matched by glob, sorted alphabetically, de-duplicated by canonical path. |
| `assets.search_paths` | list of strings | `[]` | Path list | Appended to Pandoc `--resource-path`. |
| `assets.logo_cover` | string or `null` | empty | Path or identifier | Passed as metadata key `logo_cover` (used only if template consumes it). |
| `assets.logo_header` | string or `null` | empty | Path or identifier | Passed as metadata key `logo_header` (used only if template consumes it). |
| `style.colors.primary` | string or `null` | empty | Any string (typically hex color) | Passed as metadata key `color_primary` (template-dependent). |
| `style.fonts.body` | string or `null` | empty | Any string | Passed as metadata key `font_body` (template-dependent). |
| `style.fonts.heading` | string or `null` | empty | Any string | Passed as metadata key `font_heading` (template-dependent). |
| `header_footer.header_left` | string or `null` | empty | Any string | Parsed and merged, currently not consumed by renderer/template. |
| `header_footer.header_right` | string or `null` | empty | Any string | Parsed and merged, currently not consumed by renderer/template. |
| `header_footer.footer_left` | string or `null` | empty | Any string | Parsed and merged, currently not consumed by renderer/template. |
| `header_footer.footer_right` | string or `null` | empty | Any string | Parsed and merged, currently not consumed by renderer/template. |
| `features.plantuml` | string | `auto` | `auto`, `on`, `off` | Controls PlantUML filter activation and dependency requirements. |

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
  title: "Functional Specification"
  author: "Project Team"
  subject: "Internal Document"
toc:
  mode: auto
  title: "Contents"
  depth: 3
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
