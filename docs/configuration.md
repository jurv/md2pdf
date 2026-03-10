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
header_footer:
  enabled: false
  apply_on: toc_and_body
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
| `cover.mode`                         | string           | `none`              | `none`, `builtin`, `external_template`                                | Cover strategy. With `none` + `cover.image`, image is first-page background (no extra page). With `builtin`, a dedicated cover page is inserted. |
| `cover.external_template`            | string or `null` | empty               | File path                                                             | Required when `cover.mode: external_template`; LaTeX file included with `\\input`. Incompatible with `cover.image`. |
| `cover.image`                        | string or `null` | empty               | File path                                                             | Full-bleed image for page 1. With `cover.mode: none`, it decorates page 1 background without adding a new page.   |
| `cover.image_fit`                    | string           | `cover`             | `cover`, `contain`, `stretch`                                         | Fit behavior for `cover.image` (`cover` crops, `contain` preserves full image, `stretch` fills with distortion). |
| `cover.builtin.logo`                 | string or `null` | empty               | File path                                                             | Logo path for builtin cover.                                                                                     |
| `cover.builtin.title_color`          | string or `null` | empty               | color name or `#RRGGBB`                                               | Explicit title color on builtin cover. If unset, the embedded template falls back to `style.colors.primary`, then black. |
| `cover.builtin.subtitle`             | string or `null` | empty               | Any string                                                            | Subtitle text on builtin cover.                                                                                  |
| `cover.builtin.background_color`     | string           | `#FFFFFF`           | color name or `#RRGGBB`                                               | Background color on builtin cover.                                                                               |
| `cover.builtin.align`                | string           | `center`            | `center`, `top`                                                       | Vertical alignment of builtin cover content.                                                                     |
| `sources.explicit`                   | list of strings  | `[]`                | Existing files expected                                               | Ordered source list, processed first.                                                                            |
| `sources.include`                    | list of strings  | `[]`                | Glob patterns                                                         | Additional sources matched by glob, sorted alphabetically, de-duplicated by canonical path.                      |
| `assets.search_paths`                | list of strings  | `[]`                | Path list                                                             | Appended to Pandoc `--resource-path`.                                                                            |
| `assets.logo_cover`                  | string or `null` | empty               | Path or identifier                                                    | In the embedded template, fallback logo for `cover.mode: builtin` when `cover.builtin.logo` is unset. Also passed as metadata key `logo_cover`. |
| `assets.logo_header`                 | string or `null` | empty               | Path or identifier                                                    | In the embedded template, default header logo when `header_footer.enabled: true` and no explicit header cells are defined. Also passed as metadata key `logo_header`. |
| `style.colors.primary`               | string or `null` | empty               | Any string (typically hex color)                                      | Theme fallback color in the embedded template: headings, document title, builtin cover title, and link colors unless a more specific value is set. |
| `style.fonts.body`                   | string or `null` | empty               | Any string                                                            | Body font in the embedded template under `xelatex`/`lualatex`. Ignored by `pdflatex`.                          |
| `style.fonts.heading`                | string or `null` | empty               | Any string                                                            | Heading and document-title font in the embedded template under `xelatex`/`lualatex`. Ignored by `pdflatex`.   |
| `style.links.color`                  | string or `null` | empty               | named color or `#RRGGBB`                                              | Global color for internal links. If unset, the embedded template falls back to `style.colors.primary`, then blue. |
| `style.links.url_color`              | string or `null` | empty               | named color or `#RRGGBB`                                              | Color for URL links. If unset, the embedded template falls back to `style.colors.primary`, then blue.          |
| `style.links.citation_color`         | string or `null` | empty               | named color or `#RRGGBB`                                              | Color for citation links. If unset, the embedded template falls back to `style.colors.primary`, then blue.     |
| `style.links.toc_color`              | string or `null` | empty               | named color or `#RRGGBB`                                              | Color for ToC links only. If unset, ToC uses `style.links.color` fallback behavior.                            |
| `style.headings.h<1..6>.color`       | string or `null` | empty               | named color or `#RRGGBB`                                              | Per-level heading color (`h1`..`h6`) in the default template.                                                     |
| `style.headings.h<1..6>.size_pt`     | number or `null` | empty               | `> 0`                                                                 | Per-level heading size in points.                                                                                |
| `style.headings.h<1..6>.space_before_pt` | number or `null` | empty           | `>= 0`                                                                | Per-level vertical spacing before heading in points.                                                             |
| `style.headings.h<1..6>.space_after_pt`  | number or `null` | empty           | `>= 0`                                                                | Per-level vertical spacing after heading in points.                                                              |
| `style.blockquote.bar_color`         | string           | `#E6E6E6`           | named color or `#RRGGBB`                                              | Left bar color for Markdown blockquotes in the default template.                                                 |
| `style.blockquote.text_color`        | string           | `#6F6F6F`           | named color or `#RRGGBB`                                              | Text color for Markdown blockquotes in the default template.                                                     |
| `style.blockquote.background_color`  | string           | `#F7F7F7`           | named color or `#RRGGBB`                                              | Background color for Markdown blockquotes in the default template.                                               |
| `style.blockquote.bar_width_pt`      | number           | `0.8`               | `>= 0`                                                                | Left bar width for blockquotes (points).                                                                         |
| `style.blockquote.gap_pt`            | number           | `5`                 | `>= 0`                                                                | Horizontal gap between quote bar and quote text (points).                                                        |
| `style.blockquote.padding_pt`        | number           | `2`                 | `>= 0`                                                                | Inner padding applied to quote background (points).                                                              |
| `style.plantuml.align`               | string           | `center`            | `left`, `center`, `right`                                             | Horizontal alignment for PlantUML-generated images (`plantuml-images/...`) in the default template.            |
| `style.plantuml.space_before_pt`     | number           | `6`                 | `>= 0`                                                                | Extra vertical spacing inserted before PlantUML-generated images (points).                                       |
| `style.plantuml.space_after_pt`      | number           | `0`                 | `>= 0`                                                                | Extra vertical spacing inserted after PlantUML-generated images (points).                                        |
| `style.symbols.fallback_font`        | string or `null` | `Noto Sans Symbols2` | single-line font name                                                | Secondary font used by the embedded template for curated symbol characters under `xelatex`/`lualatex`.         |
| `style.symbols.fallback_for[]`       | list<string>     | built-in list       | each entry must be exactly one Unicode character                      | Characters rendered directly with `style.symbols.fallback_font`, with LaTeX-safe fallback behavior under `pdflatex`. |
| `style.symbols.replace.<char>`       | string           | built-in map        | key must be exactly one Unicode character; value must be a single-line LaTeX snippet | Explicit replacement override for Unicode symbols in the embedded template (example: `✅` -> `\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}`). |
| `header_footer.enabled`              | boolean          | `false`             | `true` or `false`                                                     | Enables rich header/footer rendering in the default template.                                                    |
| `header_footer.apply_on`             | string           | `toc_and_body`      | `body_only`, `toc_and_body`, `all_pages`                              | Controls where the header/footer style is activated.                                                             |
| `header_footer.side_offset_left_pt`  | number           | `20`                | `>= 0`                                                                | Extends header/footer rendering into the left page margin (`fancyhfoffset`).                                     |
| `header_footer.side_offset_right_pt` | number           | `20`                | `>= 0`                                                                | Extends header/footer rendering into the right page margin (`fancyhfoffset`).                                    |
| `header_footer.footer_reserve_above_pt` | number        | `0`                 | `>= 0`                                                                | Reserves extra space between body content and footer while keeping footer visually anchored near the page bottom. |
| `header_footer.page_number.enabled`  | boolean          | `true`              | `true` or `false`                                                     | Enables page-number block rendering.                                                                             |
| `header_footer.page_number.format`   | string           | `{page}`            | Non-empty string                                                      | Default format for `page_number` blocks (supports `{page}`, `{total_pages}`).                                   |
| `header_footer.page_number.total_pages` | boolean       | `false`             | `true` or `false`                                                     | Enables total-page token use in page-number formatting.                                                          |
| `header_footer.global_style.color`   | string           | `#E0E0E0`           | named color or `#RRGGBB`                                              | Default text color for header/footer blocks.                                                                     |
| `header_footer.global_style.size_pt` | number           | `7`                 | `>= 0`                                                                | Default font size in points.                                                                                     |
| `header_footer.global_style.line_height_pt` | number     | `8`                 | `>= 0`                                                                | Default line height in points.                                                                                   |
| `header_footer.header.height_pt`     | number           | `36`                | `>= 0`                                                                | Header box height (`\\headheight`).                                                                               |
| `header_footer.header.sep_pt`        | number           | `22`                | `>= 0`                                                                | Header/content separation (`\\headsep`).                                                                          |
| `header_footer.header.raise_pt`      | number           | `4`                 | any number                                                            | Vertical nudge for header content in points (positive = higher).                                                 |
| `header_footer.footer.skip_pt`       | number           | `24`                | `>= 0`                                                                | Footer/content separation (`\\footskip`). For tall footer content, LaTeX may enforce a minimum that makes small changes appear unchanged. |
| `header_footer.footer.raise_pt`      | number           | `0`                 | any number                                                            | Fine vertical nudge for footer content in points (positive = higher, negative = lower).                         |
| `header_footer.<header|footer>.grid.columns` | list<number> | `[0.38,0.62]` / `[0.92,0.08]` | positive numbers                                             | Declarative grid column ratios (normalized at render time).                                                      |
| `header_footer.<header|footer>.grid.rows` | list<number> | `[1]`               | positive numbers                                                      | Declarative grid row ratios (normalized at render time).                                                         |
| `header_footer.<header|footer>.grid.cells[].blocks[]` | list<object> | empty | `type: text|image|page_number`                                      | Cell content blocks (supports multiline text, image, and pagination).                                            |
| `features.plantuml`                  | string           | `auto`              | `auto`, `on`, `off`                                                   | Controls PlantUML filter activation and dependency requirements.                                                 |

### Default heading behavior

Out of the box, the first entrypoint `#` heading is used as the document title and removed from body content. Heading numbering starts at `##` and ends at `###` (`1`, `1.1`, `2`, ...). By default, ToC bounds mirror numbering bounds (`h2..h3`), so `#` and `####+` headings are automatically marked as unlisted.

Heading visual style remains LaTeX default unless `style.headings.*` is set. This means existing documents keep the same heading appearance when upgrading.
For backward compatibility, legacy keys (`style.headings.color`, `style.headings.h2_size_pt`, etc.) are still accepted and mapped to the new per-level format.
With LaTeX `article`, `h5` and `h6` are both rendered as `subparagraph`; if both are configured, `h6` settings take precedence.

### Theme colors, fonts, and logo fallbacks

In the embedded template, `style.colors.primary` is a fallback theme color, not a hard override. More specific keys keep priority: `style.headings.h*.color`, `style.links.*`, and `cover.builtin.title_color` win when set.

Likewise, `style.fonts.body` affects body text only, while `style.fonts.heading` affects headings and document title blocks. Custom fonts require `xelatex` or `lualatex` and must be installed on the host system.

For logos, `assets.logo_cover` is used as the builtin cover logo only if `cover.builtin.logo` is unset, and `assets.logo_header` is injected only when header/footer rendering is enabled and no explicit header cells are configured.

### Default blockquote behavior

With the embedded template, Markdown blockquotes (`>`) are rendered with a thin left bar, lighter text, and a subtle background. You can tune bar color, text color, background, bar width, spacing, and padding through `style.blockquote.*`.

### PlantUML image layout

With the embedded template, images produced by `pandoc-plantuml` (stored under `plantuml-images/`) can be styled through `style.plantuml.*`. By default they are centered and given a small spacing before the diagram.

### Unicode symbol replacements

With the embedded template, Unicode symbol handling is split into two layers. `style.symbols.fallback_for` routes curated checkbox/symbol characters through a secondary font (`style.symbols.fallback_font`) when the engine supports `fontspec`, while `style.symbols.replace` remains the explicit override mechanism for characters that need a custom LaTeX snippet. This avoids missing-glyph rectangles without requiring color-emoji support.

For inline code and fenced code blocks, md2pdf also normalizes emoji-like aliases when the replacement uses `\mdtwosymbolglyph{...}{...}`. This keeps code samples readable in the monospace font even though verbatim LaTeX does not support the same Unicode fallback mechanism as body text.

Example override:

```yaml
style:
  symbols:
    fallback_font: "Noto Sans Symbols2"
    fallback_for:
      - "☐"
      - "☑"
      - "☒"
      - "⚠"
    replace:
      "✅": '\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}'
      "❌": '\textbf{NO}'
```

Notes:

- `replace` wins if the same character appears in both `fallback_for` and `replace`.
- The default configuration strips the emoji variation selector (`U+FE0F`), so sequences like `⚠️` fall back to the base character `⚠`.
- The default configuration also aliases common checkbox variants used in editors, for example `✅` and `🗹` to `☑`, and `❌` and `🗷` to `☒` inside code blocks.
- The embedded template targets faithful monochrome symbols, not color emoji.

### Full-bleed cover image

For a simple full-page cover image, use only:

```yaml
cover:
  image: "assets/cover.png"
```

Behavior:

- page 1 uses the image as full-bleed background (no page margins)
- no dedicated extra cover page is inserted in this simple mode
- default fit is `cover`
- document flow stays unchanged (same first content page, just with background)

If you need a dedicated standalone cover page (with optional logo/title/subtitle overlays), use:

```yaml
cover:
  mode: builtin
  image: "assets/cover.png"
```

In `cover.mode: builtin`, the cover already renders title/author; default inline title rendering is skipped on the following pages to avoid duplicate title blocks.

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
  image: null
  image_fit: cover
  builtin:
    logo: null
    title_color: null
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
  links:
    color: null
    url_color: null
    citation_color: null
    toc_color: null
  headings:
    h1:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
    h2:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
    h3:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
    h4:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
    h5:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
    h6:
      color: null
      size_pt: null
      space_before_pt: null
      space_after_pt: null
  blockquote:
    bar_color: "#E6E6E6"
    text_color: "#6F6F6F"
    background_color: "#F7F7F7"
    bar_width_pt: 0.8
    gap_pt: 5
    padding_pt: 2
  plantuml:
    align: center
    space_before_pt: 6
    space_after_pt: 0
  symbols:
    fallback_font: "Noto Sans Symbols2"
    fallback_for:
      - "☐"
      - "☑"
      - "☒"
      - "⚠"
      - "✓"
      - "✔"
      - "✗"
      - "✖"
    replace:
      "💡": '\mdtwosymbolglyph{✦}{\textasteriskcentered}'
      "✅": '\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}'
      "❌": '\mdtwosymbolglyph{☒}{\ensuremath{\boxtimes}}'
      "🗹": '\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}'
      "🗷": '\mdtwosymbolglyph{☒}{\ensuremath{\boxtimes}}'
      "🗸": '\mdtwosymbolglyph{✓}{\ensuremath{\checkmark}}'
      "🗵": '\mdtwosymbolglyph{✗}{\ensuremath{\times}}'
      "◻": '\mdtwosymbolglyph{☐}{\ensuremath{\square}}'
      "⬜": '\mdtwosymbolglyph{☐}{\ensuremath{\square}}'
      "️": ''
header_footer:
  enabled: true
  apply_on: toc_and_body
  footer_reserve_above_pt: 0
  side_offset_left_pt: 20
  side_offset_right_pt: 20
  page_number:
    enabled: true
    format: "{page}"
    total_pages: false
  global_style:
    font: null
    color: "#E0E0E0"
    size_pt: 7
    line_height_pt: 8
    opacity: 1.0
    weight: normal
  header:
    height_pt: 36
    sep_pt: 22
    raise_pt: 4
    grid:
      columns: [0.38, 0.62]
      rows: [1]
      cells:
        - row: 1
          col: 1
          align_h: left
          align_v: top
          blocks:
            - type: image
              path: "assets/logo-header.png"
              height_pt: 22
        - row: 1
          col: 2
          align_h: right
          align_v: top
          blocks:
            - type: text
              value: |
                Integral Service
                31 rue Ampere
                69008 Chassieu
              style:
                color: "#16AFD2"
                weight: bold
  footer:
    skip_pt: 24
    raise_pt: 0
    grid:
      columns: [0.92, 0.08]
      rows: [1]
      cells:
        - row: 1
          col: 1
          align_h: left
          align_v: bottom
          blocks:
            - type: text
              value: |
                Confidential. Unauthorized use is prohibited.
                This document is intended for the addressee only.
              style:
                size_pt: 6
        - row: 1
          col: 2
          align_h: right
          align_v: bottom
          blocks:
            - type: page_number
              format: "{page}"
features:
  plantuml: auto
---
```

### Deprecated header/footer keys

Legacy flat keys are no longer accepted: `header_footer.header_left`, `header_footer.header_right`, `header_footer.footer_left`, `header_footer.footer_right`.
`md2pdf` now raises a configuration error when these keys are present to avoid silent no-op behavior.

### Header/Footer Grid Notes

- `grid.cells[].row` and `grid.cells[].col` are **1-based**.
- `row_span` and `col_span` keys exist for forward compatibility, but the current renderer only accepts `1`.
- Supported block types are `text`, `image`, and `page_number`.
- Text blocks support placeholders: `{page}`, `{total_pages}`, `{title}`, `{date}`.
- `footer_reserve_above_pt` adds gap above footer while keeping footer position stable (useful for multi-line legal footers).
- Use `footer.raise_pt` for precise vertical tuning; use `footer.skip_pt` for coarse body/footer separation.

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
