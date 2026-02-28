# Troubleshooting

## `dependency check failed`

Run:

```bash
md2pdf doctor
```

Install missing required tools (`pandoc`, selected PDF engine, and PlantUML stack if needed).

## Build fails on PlantUML documents

If your markdown contains PlantUML blocks, install:

- `pandoc-plantuml`
- `plantuml`
- `dot` (Graphviz)

Then rerun `md2pdf doctor`.

## Merge command fails

The merge command requires either:

- `pdftk`, or
- `pdfunite`

Install one of them and run the command again.

## Compress command fails

The compression command requires Ghostscript (`gs`). Install it and retry.

## Unexpected output path

By default, output is `<input-basename>.pdf` next to the input markdown file. Use `-o` to force a custom output path.

## Images overflow pages in custom templates

If you use a custom `pdf.template`, make sure it includes a global image policy such as:

```tex
\setkeys{Gin}{width=\linewidth,height=0.8\textheight,keepaspectratio}
```

The embedded default template already applies this safeguard.
