package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/julien/md2pdf/internal/config"
	"github.com/julien/md2pdf/internal/fs"
)

func buildHeaderFooterMetadata(cfg config.Config, baseDir, workDir string, tocEnabled bool) ([][2]string, error) {
	if !cfg.HeaderFooter.Enabled {
		return nil, nil
	}

	partial, err := compileHeaderFooterPartial(cfg, baseDir)
	if err != nil {
		return nil, err
	}
	partialPath := filepath.Join(workDir, "md2pdf-header-footer.tex")
	if err := os.WriteFile(partialPath, []byte(partial), 0o600); err != nil {
		return nil, fmt.Errorf("failed to write header/footer partial: %w", err)
	}

	pairs := [][2]string{
		{"hf_enabled", "true"},
		{"hf_partial", partialPath},
	}

	switch cfg.HeaderFooter.ApplyOn {
	case "all_pages":
		pairs = append(pairs, [2]string{"hf_activate_at_start", "true"})
	case "toc_and_body":
		if tocEnabled {
			pairs = append(pairs, [2]string{"hf_activate_before_toc", "true"})
		} else {
			pairs = append(pairs, [2]string{"hf_activate_before_body_no_toc", "true"})
		}
	case "body_only":
		if tocEnabled {
			pairs = append(pairs, [2]string{"hf_activate_after_toc", "true"})
		} else {
			pairs = append(pairs, [2]string{"hf_activate_before_body_no_toc", "true"})
		}
	}

	return pairs, nil
}

func compileHeaderFooterPartial(cfg config.Config, baseDir string) (string, error) {
	cfg = withDefaultHeaderLogoCell(cfg)

	headerContent, err := compileHeaderFooterRegion(
		cfg.HeaderFooter.Header,
		cfg.HeaderFooter.GlobalStyle,
		cfg.HeaderFooter.PageNumber,
		cfg.Metadata.Title,
		baseDir,
		false,
	)
	if err != nil {
		return "", err
	}
	footerContent, err := compileHeaderFooterRegion(
		cfg.HeaderFooter.Footer,
		cfg.HeaderFooter.GlobalStyle,
		cfg.HeaderFooter.PageNumber,
		cfg.Metadata.Title,
		baseDir,
		true,
	)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	headHeight := cfg.HeaderFooter.Header.HeightPt
	headSep := cfg.HeaderFooter.Header.SepPt
	if cfg.HeaderFooter.Header.RaisePt > 0 {
		headHeight += cfg.HeaderFooter.Header.RaisePt
		headSep += cfg.HeaderFooter.Header.RaisePt
	}
	if headHeight > 0 {
		fmt.Fprintf(&b, "\\setlength{\\headheight}{%spt}\n", formatFloatPt(headHeight))
	}
	if headSep > 0 {
		fmt.Fprintf(&b, "\\setlength{\\headsep}{%spt}\n", formatFloatPt(headSep))
	}
	footSkip := cfg.HeaderFooter.Footer.SkipPt
	if cfg.HeaderFooter.FooterReserveAbovePt > 0 {
		// Keep footer anchored while adding gap above it:
		// - increase footskip by reserve amount
		// - reduce textheight by the same amount
		footSkip += cfg.HeaderFooter.FooterReserveAbovePt
	}
	if footSkip > 0 {
		fmt.Fprintf(&b, "\\setlength{\\footskip}{%spt}\n", formatFloatPt(footSkip))
	}
	if cfg.HeaderFooter.FooterReserveAbovePt > 0 {
		fmt.Fprintf(&b, "\\addtolength{\\textheight}{-%spt}\n", formatFloatPt(cfg.HeaderFooter.FooterReserveAbovePt))
	}

	b.WriteString("\\fancypagestyle{mdtwohf}{\n")
	b.WriteString("\\fancyhf{}\n")
	b.WriteString("\\renewcommand{\\headrulewidth}{0pt}\n")
	b.WriteString("\\renewcommand{\\footrulewidth}{0pt}\n")
	if cfg.HeaderFooter.SideOffsetLeftPt > 0 {
		fmt.Fprintf(&b, "\\fancyhfoffset[L]{%spt}\n", formatFloatPt(cfg.HeaderFooter.SideOffsetLeftPt))
	}
	if cfg.HeaderFooter.SideOffsetRightPt > 0 {
		fmt.Fprintf(&b, "\\fancyhfoffset[R]{%spt}\n", formatFloatPt(cfg.HeaderFooter.SideOffsetRightPt))
	}
	if strings.TrimSpace(headerContent) != "" {
		if cfg.HeaderFooter.Header.RaisePt != 0 {
			// Keep box metrics at zero so raise_pt is a pure visual shift.
			fmt.Fprintf(&b, "\\fancyhead[C]{\\raisebox{%spt}[0pt][0pt]{\\begin{minipage}[t]{\\headwidth}\\vspace*{0pt}", formatFloatPt(cfg.HeaderFooter.Header.RaisePt))
		} else {
			b.WriteString("\\fancyhead[C]{\\begin{minipage}[t]{\\headwidth}\\vspace*{0pt}")
		}
		b.WriteString(headerContent)
		if cfg.HeaderFooter.Header.RaisePt != 0 {
			b.WriteString("\\end{minipage}}}\n")
		} else {
			b.WriteString("\\end{minipage}}\n")
		}
	}
	if strings.TrimSpace(footerContent) != "" {
		if cfg.HeaderFooter.Footer.RaisePt != 0 {
			// Keep box metrics at zero so raise_pt is a pure visual shift.
			fmt.Fprintf(&b, "\\fancyfoot[C]{\\raisebox{%spt}[0pt][0pt]{\\begin{minipage}[b]{\\headwidth}", formatFloatPt(cfg.HeaderFooter.Footer.RaisePt))
		} else {
			b.WriteString("\\fancyfoot[C]{\\begin{minipage}[b]{\\headwidth}")
		}
		b.WriteString(footerContent)
		if cfg.HeaderFooter.Footer.RaisePt != 0 {
			b.WriteString("\\end{minipage}}}\n")
		} else {
			b.WriteString("\\end{minipage}}\n")
		}
	}
	b.WriteString("}\n")
	b.WriteString("\\newcommand{\\mdtwohfactivate}{\\pagestyle{mdtwohf}\\fancypagestyle{plain}{\\pagestyle{mdtwohf}}}\n")
	return b.String(), nil
}

func withDefaultHeaderLogoCell(cfg config.Config) config.Config {
	if !cfg.HeaderFooter.Enabled {
		return cfg
	}
	if strings.TrimSpace(cfg.Assets.LogoHeader) == "" {
		return cfg
	}
	if len(cfg.HeaderFooter.Header.Grid.Cells) > 0 {
		return cfg
	}

	cfg.HeaderFooter.Header.Grid.Cells = []config.HeaderFooterCell{
		{
			Row:    1,
			Col:    1,
			AlignH: "left",
			AlignV: "top",
			Blocks: []config.HeaderFooterBlock{
				{
					Type:     "image",
					Path:     cfg.Assets.LogoHeader,
					HeightPt: 22,
				},
			},
		},
	}

	return cfg
}

func compileHeaderFooterRegion(
	region config.HeaderFooterRegion,
	globalStyle config.TextStyleConfig,
	pageCfg config.PageNumberConfig,
	docTitle string,
	baseDir string,
	isFooter bool,
) (string, error) {
	cols := normalizePositiveRatios(region.Grid.Columns)
	rows := normalizePositiveRatios(region.Grid.Rows)
	if len(cols) == 0 || len(rows) == 0 {
		return "", nil
	}

	cellsByPos := make(map[string]config.HeaderFooterCell, len(region.Grid.Cells))
	hasPageNumberBlock := false
	for _, cell := range region.Grid.Cells {
		key := cellKey(cell.Row, cell.Col)
		cellsByPos[key] = cell
		for _, block := range cell.Blocks {
			if block.Type == "page_number" {
				hasPageNumberBlock = true
			}
		}
	}

	if isFooter && pageCfg.Enabled && !hasPageNumberBlock {
		lastCol := len(cols)
		key := cellKey(1, lastCol)
		cell := cellsByPos[key]
		if cell.Row == 0 {
			cell.Row = 1
			cell.Col = lastCol
			cell.AlignH = "right"
			cell.AlignV = "bottom"
		}
		cell.Blocks = append(cell.Blocks, config.HeaderFooterBlock{
			Type:   "page_number",
			Format: pageCfg.Format,
		})
		cellsByPos[key] = cell
	}

	var b strings.Builder
	b.WriteString("\\setlength{\\tabcolsep}{0pt}\n")
	b.WriteString("\\renewcommand{\\arraystretch}{1}\n")
	tableAlign := "t"
	if isFooter {
		tableAlign = "b"
	}
	b.WriteString("\\begin{tabular}[")
	b.WriteString(tableAlign)
	b.WriteString("]{")
	for _, ratio := range cols {
		fmt.Fprintf(&b, "p{\\dimexpr%s\\linewidth\\relax}", trimTrailingZero(ratio))
	}
	b.WriteString("}\n")

	for row := 1; row <= len(rows); row++ {
		for col := 1; col <= len(cols); col++ {
			if col > 1 {
				b.WriteString(" & ")
			}
			cell := cellsByPos[cellKey(row, col)]
			cellContent, err := compileHeaderFooterCell(cell, globalStyle, pageCfg, docTitle, baseDir)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(cellContent) == "" {
				b.WriteString("~")
			} else {
				b.WriteString(cellContent)
			}
		}
		if row < len(rows) {
			b.WriteString(" \\\\\n")
		} else {
			b.WriteByte('\n')
		}
	}
	b.WriteString("\\end{tabular}")

	return b.String(), nil
}

func compileHeaderFooterCell(
	cell config.HeaderFooterCell,
	globalStyle config.TextStyleConfig,
	pageCfg config.PageNumberConfig,
	docTitle string,
	baseDir string,
) (string, error) {
	if len(cell.Blocks) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(cell.Blocks))
	for _, block := range cell.Blocks {
		blockLatex, err := compileHeaderFooterBlock(block, globalStyle, pageCfg, docTitle, baseDir)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(blockLatex) != "" {
			parts = append(parts, blockLatex)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}

	vAlign := "t"
	switch cell.AlignV {
	case "middle":
		vAlign = "c"
	case "bottom":
		vAlign = "b"
	}

	hAlignCmd := "\\raggedright"
	switch cell.AlignH {
	case "center":
		hAlignCmd = "\\centering"
	case "right":
		hAlignCmd = "\\raggedleft"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\\begin{minipage}[%s]{\\linewidth}", vAlign)
	if vAlign == "t" {
		b.WriteString("\\vspace*{0pt}")
	}
	b.WriteString("\\setlength{\\parskip}{0pt}")
	b.WriteString(hAlignCmd)
	b.WriteString(strings.Join(parts, "\\\\"))
	b.WriteString("\\end{minipage}")
	return b.String(), nil
}

func compileHeaderFooterBlock(
	block config.HeaderFooterBlock,
	globalStyle config.TextStyleConfig,
	pageCfg config.PageNumberConfig,
	docTitle string,
	baseDir string,
) (string, error) {
	switch block.Type {
	case "text":
		content := renderTextWithTokens(block.Value, docTitle)
		return applyTextStyle(content, mergeTextStyle(globalStyle, block.Style)), nil
	case "page_number":
		if !pageCfg.Enabled {
			return "", nil
		}
		format := block.Format
		if strings.TrimSpace(format) == "" {
			format = pageCfg.Format
		}
		content := renderTextWithTokens(format, docTitle)
		return applyTextStyle(content, mergeTextStyle(globalStyle, block.Style)), nil
	case "image":
		return renderImageBlock(block, baseDir), nil
	default:
		return "", fmt.Errorf("unsupported header/footer block type %q", block.Type)
	}
}

func renderImageBlock(block config.HeaderFooterBlock, baseDir string) string {
	path := fs.ResolveOptionalPath(baseDir, block.Path)
	opts := make([]string, 0, 3)
	if block.WidthPt > 0 {
		opts = append(opts, "width="+formatFloatPt(block.WidthPt)+"pt")
	}
	if block.HeightPt > 0 {
		opts = append(opts, "height="+formatFloatPt(block.HeightPt)+"pt")
	}
	opts = append(opts, "keepaspectratio")

	var b strings.Builder
	b.WriteString("\\includegraphics")
	if len(opts) > 0 {
		b.WriteString("[")
		b.WriteString(strings.Join(opts, ","))
		b.WriteString("]")
	}
	b.WriteString("{\\detokenize{")
	b.WriteString(path)
	b.WriteString("}}")
	return b.String()
}

func renderTextWithTokens(raw, docTitle string) string {
	const (
		placeholderPage      = "MD2PDFTOKPAGE"
		placeholderTotal     = "MD2PDFTOKTOTAL"
		placeholderDate      = "MD2PDFTOKDATE"
		placeholderTitleMark = "MD2PDFTOKTITLE"
	)

	text := strings.ReplaceAll(raw, "\r\n", "\n")
	text = strings.ReplaceAll(text, "{page}", placeholderPage)
	text = strings.ReplaceAll(text, "{total_pages}", placeholderTotal)
	text = strings.ReplaceAll(text, "{date}", placeholderDate)
	text = strings.ReplaceAll(text, "{title}", placeholderTitleMark)
	text = strings.ReplaceAll(text, placeholderTitleMark, docTitle)

	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = latexEscape(lines[i])
	}
	out := strings.Join(lines, `\\`)
	out = strings.ReplaceAll(out, placeholderPage, `\thepage`)
	out = strings.ReplaceAll(out, placeholderTotal, `\pageref{LastPage}`)
	out = strings.ReplaceAll(out, placeholderDate, `\today`)
	return out
}

func applyTextStyle(content string, style config.TextStyleConfig) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("{")
	if colorCmd := latexColorCommand(style.Color, style.Opacity); colorCmd != "" {
		b.WriteString(colorCmd)
	}
	if style.Font != "" {
		b.WriteString(`\fontspec{`)
		b.WriteString(latexEscape(style.Font))
		b.WriteString("}")
	}
	if style.SizePt > 0 {
		lineHeight := style.LineHeightPt
		if lineHeight <= 0 {
			lineHeight = style.SizePt * 1.2
		}
		b.WriteString(`\fontsize{`)
		b.WriteString(formatFloatPt(style.SizePt))
		b.WriteString("pt}{")
		b.WriteString(formatFloatPt(lineHeight))
		b.WriteString(`pt}\selectfont `)
	}
	if style.Weight == "bold" {
		b.WriteString(`\bfseries `)
	}
	b.WriteString(content)
	b.WriteString("}")
	return b.String()
}

func latexColorCommand(color string, opacity float64) string {
	trimmed := strings.TrimSpace(color)
	if trimmed == "" {
		return ""
	}
	model, value := latexColor(trimmed)
	if model == "HTML" {
		return `\color[HTML]{` + value + "}"
	}
	if opacity > 0 && opacity < 1 {
		percent := int(opacity * 100)
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
		return `\color{` + value + "!" + strconv.Itoa(percent) + "}"
	}
	return `\color{` + value + "}"
}

func mergeTextStyle(base, override config.TextStyleConfig) config.TextStyleConfig {
	out := base
	if strings.TrimSpace(override.Font) != "" {
		out.Font = strings.TrimSpace(override.Font)
	}
	if strings.TrimSpace(override.Color) != "" {
		out.Color = strings.TrimSpace(override.Color)
	}
	if override.SizePt > 0 {
		out.SizePt = override.SizePt
	}
	if override.LineHeightPt > 0 {
		out.LineHeightPt = override.LineHeightPt
	}
	if strings.TrimSpace(override.Weight) != "" {
		out.Weight = strings.TrimSpace(override.Weight)
	}
	if override.Opacity > 0 {
		out.Opacity = override.Opacity
	}
	return out
}

func normalizePositiveRatios(values []float64) []float64 {
	out := make([]float64, 0, len(values))
	sum := 0.0
	for _, v := range values {
		if v > 0 {
			sum += v
			out = append(out, v)
		}
	}
	if sum == 0 {
		return nil
	}
	for i := range out {
		out[i] = out[i] / sum
	}
	return out
}

func cellKey(row, col int) string {
	return fmt.Sprintf("%d:%d", row, col)
}

func formatFloatPt(v float64) string {
	return trimTrailingZero(v)
}

func trimTrailingZero(v float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 4, 64), "0"), ".")
}

func latexEscape(input string) string {
	replacer := strings.NewReplacer(
		"\\", `\textbackslash{}`,
		`{`, `\{`,
		`}`, `\}`,
		`$`, `\$`,
		`&`, `\&`,
		`#`, `\#`,
		`_`, `\_`,
		`%`, `\%`,
		`~`, `\textasciitilde{}`,
		`^`, `\textasciicircum{}`,
	)
	return replacer.Replace(input)
}
