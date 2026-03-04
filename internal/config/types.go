package config

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PDF          PDFConfig          `yaml:"pdf"`
	Metadata     MetadataConfig     `yaml:"metadata"`
	Title        TitleConfig        `yaml:"title"`
	Heading      HeadingConfig      `yaml:"heading_numbering"`
	TOC          TOCConfig          `yaml:"toc"`
	Cover        CoverConfig        `yaml:"cover"`
	Sources      SourcesConfig      `yaml:"sources"`
	Assets       AssetsConfig       `yaml:"assets"`
	Style        StyleConfig        `yaml:"style"`
	HeaderFooter HeaderFooterConfig `yaml:"header_footer"`
	Features     FeaturesConfig     `yaml:"features"`
}

type PDFConfig struct {
	Engine   string `yaml:"engine"`
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
}

type MetadataConfig struct {
	Title   string `yaml:"title"`
	Author  string `yaml:"author"`
	Subject string `yaml:"subject"`
}

type TitleConfig struct {
	Source        string `yaml:"source"`
	StripFromBody bool   `yaml:"strip_from_body"`
	RenderMode    string `yaml:"render_mode"`
}

type HeadingConfig struct {
	Enabled     bool              `yaml:"enabled"`
	FromLevel   int               `yaml:"from_level"`
	ToLevel     int               `yaml:"to_level"`
	Separator   string            `yaml:"separator"`
	Suffix      string            `yaml:"suffix"`
	MirrorInTOC bool              `yaml:"mirror_in_toc"`
	Notation    map[string]string `yaml:"notation"`
}

type TOCConfig struct {
	Mode      string `yaml:"mode"`
	Title     string `yaml:"title"`
	Depth     int    `yaml:"depth"` // backward-compatible alias
	FromLevel int    `yaml:"from_level"`
	ToLevel   int    `yaml:"to_level"`
}

type CoverConfig struct {
	Mode             string             `yaml:"mode"`
	ExternalTemplate string             `yaml:"external_template"`
	Image            string             `yaml:"image"`
	ImageFit         string             `yaml:"image_fit"`
	Builtin          BuiltinCoverConfig `yaml:"builtin"`
}

type BuiltinCoverConfig struct {
	Logo            string `yaml:"logo"`
	TitleColor      string `yaml:"title_color"`
	Subtitle        string `yaml:"subtitle"`
	BackgroundColor string `yaml:"background_color"`
	Align           string `yaml:"align"`
}

type SourcesConfig struct {
	Explicit []string `yaml:"explicit"`
	Include  []string `yaml:"include"`
}

type AssetsConfig struct {
	SearchPaths []string `yaml:"search_paths"`
	LogoCover   string   `yaml:"logo_cover"`
	LogoHeader  string   `yaml:"logo_header"`
}

type StyleConfig struct {
	Colors     ColorsConfig          `yaml:"colors"`
	Fonts      FontsConfig           `yaml:"fonts"`
	Links      LinksStyleConfig      `yaml:"links"`
	Headings   HeadingStyleConfig    `yaml:"headings"`
	BlockQuote BlockQuoteStyleConfig `yaml:"blockquote"`
	PlantUML   PlantUMLStyleConfig   `yaml:"plantuml"`
}

type ColorsConfig struct {
	Primary string `yaml:"primary"`
}

type FontsConfig struct {
	Body    string `yaml:"body"`
	Heading string `yaml:"heading"`
}

type LinksStyleConfig struct {
	Color         string `yaml:"color"`
	URLColor      string `yaml:"url_color"`
	CitationColor string `yaml:"citation_color"`
	TOCColor      string `yaml:"toc_color"`
}

type HeadingStyleConfig struct {
	H1 HeadingLevelStyleConfig `yaml:"h1"`
	H2 HeadingLevelStyleConfig `yaml:"h2"`
	H3 HeadingLevelStyleConfig `yaml:"h3"`
	H4 HeadingLevelStyleConfig `yaml:"h4"`
	H5 HeadingLevelStyleConfig `yaml:"h5"`
	H6 HeadingLevelStyleConfig `yaml:"h6"`
}

type HeadingLevelStyleConfig struct {
	Color         string   `yaml:"color"`
	SizePt        *float64 `yaml:"size_pt"`
	SpaceBeforePt *float64 `yaml:"space_before_pt"`
	SpaceAfterPt  *float64 `yaml:"space_after_pt"`
}

type BlockQuoteStyleConfig struct {
	BarColor        string  `yaml:"bar_color"`
	TextColor       string  `yaml:"text_color"`
	BackgroundColor string  `yaml:"background_color"`
	BarWidthPt      float64 `yaml:"bar_width_pt"`
	GapPt           float64 `yaml:"gap_pt"`
	PaddingPt       float64 `yaml:"padding_pt"`
}

type PlantUMLStyleConfig struct {
	Align         string  `yaml:"align"`
	SpaceBeforePt float64 `yaml:"space_before_pt"`
	SpaceAfterPt  float64 `yaml:"space_after_pt"`
}

type HeaderFooterConfig struct {
	Enabled              bool               `yaml:"enabled"`
	ApplyOn              string             `yaml:"apply_on"`
	SideOffsetLeftPt     float64            `yaml:"side_offset_left_pt"`
	SideOffsetRightPt    float64            `yaml:"side_offset_right_pt"`
	FooterReserveAbovePt float64            `yaml:"footer_reserve_above_pt"`
	PageNumber           PageNumberConfig   `yaml:"page_number"`
	GlobalStyle          TextStyleConfig    `yaml:"global_style"`
	Header               HeaderFooterRegion `yaml:"header"`
	Footer               HeaderFooterRegion `yaml:"footer"`
}

type PageNumberConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Format     string `yaml:"format"`
	TotalPages bool   `yaml:"total_pages"`
}

type TextStyleConfig struct {
	Font         string  `yaml:"font"`
	Color        string  `yaml:"color"`
	SizePt       float64 `yaml:"size_pt"`
	LineHeightPt float64 `yaml:"line_height_pt"`
	Opacity      float64 `yaml:"opacity"`
	Weight       string  `yaml:"weight"`
}

type HeaderFooterRegion struct {
	HeightPt float64          `yaml:"height_pt"`
	SepPt    float64          `yaml:"sep_pt"`
	SkipPt   float64          `yaml:"skip_pt"`
	RaisePt  float64          `yaml:"raise_pt"`
	Grid     HeaderFooterGrid `yaml:"grid"`
}

type HeaderFooterGrid struct {
	Columns []float64          `yaml:"columns"`
	Rows    []float64          `yaml:"rows"`
	Cells   []HeaderFooterCell `yaml:"cells"`
}

type HeaderFooterCell struct {
	Row     int                 `yaml:"row"`
	Col     int                 `yaml:"col"`
	RowSpan int                 `yaml:"row_span"`
	ColSpan int                 `yaml:"col_span"`
	AlignH  string              `yaml:"align_h"`
	AlignV  string              `yaml:"align_v"`
	Blocks  []HeaderFooterBlock `yaml:"blocks"`
}

type HeaderFooterBlock struct {
	Type     string          `yaml:"type"`
	Value    string          `yaml:"value"`
	Path     string          `yaml:"path"`
	Format   string          `yaml:"format"`
	WidthPt  float64         `yaml:"width_pt"`
	HeightPt float64         `yaml:"height_pt"`
	Style    TextStyleConfig `yaml:"style"`
}

type FeaturesConfig struct {
	PlantUML string `yaml:"plantuml"`
}

func Default() Config {
	return Config{
		PDF: PDFConfig{
			Engine: "xelatex",
		},
		Title: TitleConfig{
			Source:        "entrypoint_h1",
			StripFromBody: true,
			RenderMode:    "inline",
		},
		Heading: HeadingConfig{
			Enabled:     true,
			FromLevel:   2,
			ToLevel:     3,
			Separator:   ".",
			Suffix:      "",
			MirrorInTOC: true,
			Notation: map[string]string{
				"2": "decimal",
				"3": "decimal",
			},
		},
		TOC: TOCConfig{
			Mode:      "auto",
			Title:     "Table of Contents",
			Depth:     3,
			FromLevel: 2,
			ToLevel:   3,
		},
		Cover: CoverConfig{
			Mode:     "none",
			ImageFit: "cover",
			Builtin: BuiltinCoverConfig{
				TitleColor:      "#000000",
				BackgroundColor: "#FFFFFF",
				Align:           "center",
			},
		},
		Style: StyleConfig{
			Links: LinksStyleConfig{
				Color:         "blue",
				URLColor:      "blue",
				CitationColor: "blue",
				TOCColor:      "",
			},
			BlockQuote: BlockQuoteStyleConfig{
				BarColor:        "#E6E6E6",
				TextColor:       "#6F6F6F",
				BackgroundColor: "#F7F7F7",
				BarWidthPt:      0.8,
				GapPt:           5,
				PaddingPt:       2,
			},
			PlantUML: PlantUMLStyleConfig{
				Align:         "center",
				SpaceBeforePt: 6,
				SpaceAfterPt:  0,
			},
		},
		HeaderFooter: HeaderFooterConfig{
			Enabled:              false,
			ApplyOn:              "toc_and_body",
			SideOffsetLeftPt:     20,
			SideOffsetRightPt:    20,
			FooterReserveAbovePt: 0,
			PageNumber: PageNumberConfig{
				Enabled:    true,
				Format:     "{page}",
				TotalPages: false,
			},
			GlobalStyle: TextStyleConfig{
				Color:        "#E0E0E0",
				SizePt:       7,
				LineHeightPt: 8,
				Opacity:      1,
				Weight:       "normal",
			},
			Header: HeaderFooterRegion{
				HeightPt: 36,
				SepPt:    22,
				RaisePt:  4,
				Grid: HeaderFooterGrid{
					Columns: []float64{0.38, 0.62},
					Rows:    []float64{1},
				},
			},
			Footer: HeaderFooterRegion{
				SkipPt:  24,
				RaisePt: 0,
				Grid: HeaderFooterGrid{
					Columns: []float64{0.92, 0.08},
					Rows:    []float64{1},
				},
			},
		},
		Features: FeaturesConfig{
			PlantUML: "auto",
		},
	}
}

func DefaultMap() map[string]any {
	cfg := Default()
	blob, err := yaml.Marshal(cfg)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := yaml.Unmarshal(blob, &out); err != nil {
		return map[string]any{}
	}
	return NormalizeMap(out)
}

func (c *Config) Validate() error {
	switch c.PDF.Engine {
	case "xelatex", "lualatex", "pdflatex":
	default:
		return fmt.Errorf("invalid pdf.engine %q (allowed: xelatex, lualatex, pdflatex)", c.PDF.Engine)
	}

	switch c.Title.Source {
	case "entrypoint_h1", "metadata_only", "none":
	default:
		return fmt.Errorf("invalid title.source %q (allowed: entrypoint_h1, metadata_only, none)", c.Title.Source)
	}

	switch c.Title.RenderMode {
	case "inline", "separate_page", "none":
	default:
		return fmt.Errorf("invalid title.render_mode %q (allowed: inline, separate_page, none)", c.Title.RenderMode)
	}

	switch c.TOC.Mode {
	case "auto", "on", "off":
	default:
		return fmt.Errorf("invalid toc.mode %q (allowed: auto, on, off)", c.TOC.Mode)
	}

	if err := validateLevelRange(c.TOC.FromLevel, c.TOC.ToLevel, "toc"); err != nil {
		return err
	}

	if c.TOC.Depth <= 0 {
		return fmt.Errorf("toc.depth/to_level must be > 0")
	}

	if err := validateLevelRange(c.Heading.FromLevel, c.Heading.ToLevel, "heading_numbering"); err != nil {
		return err
	}

	if c.Heading.Enabled {
		for levelKey, notation := range c.Heading.Notation {
			switch notation {
			case "decimal", "roman_upper", "roman_lower", "alpha_upper", "alpha_lower":
			default:
				return fmt.Errorf("invalid heading_numbering.notation[%s]=%q", levelKey, notation)
			}
		}
	}

	switch c.Cover.Mode {
	case "none", "builtin", "external_template":
	default:
		return fmt.Errorf("invalid cover.mode %q (allowed: none, builtin, external_template)", c.Cover.Mode)
	}

	if c.Cover.Mode == "external_template" && c.Cover.ExternalTemplate == "" {
		return fmt.Errorf("cover.external_template must be set when cover.mode=external_template")
	}
	if c.Cover.Mode == "external_template" && strings.TrimSpace(c.Cover.Image) != "" {
		return fmt.Errorf("cover.image cannot be set when cover.mode=external_template")
	}
	switch c.Cover.ImageFit {
	case "", "cover", "contain", "stretch":
	default:
		return fmt.Errorf("invalid cover.image_fit %q (allowed: cover, contain, stretch)", c.Cover.ImageFit)
	}

	switch c.Cover.Builtin.Align {
	case "", "center", "top":
	default:
		return fmt.Errorf("invalid cover.builtin.align %q (allowed: center, top)", c.Cover.Builtin.Align)
	}

	if err := validateBlockQuoteStyle(c.Style.BlockQuote, "style.blockquote"); err != nil {
		return err
	}
	if err := validatePlantUMLStyle(c.Style.PlantUML, "style.plantuml"); err != nil {
		return err
	}
	if err := validateLinksStyle(c.Style.Links, "style.links"); err != nil {
		return err
	}
	if err := validateHeadingStyle(c.Style.Headings, "style.headings"); err != nil {
		return err
	}

	switch c.HeaderFooter.ApplyOn {
	case "body_only", "toc_and_body", "all_pages":
	default:
		return fmt.Errorf("invalid header_footer.apply_on %q (allowed: body_only, toc_and_body, all_pages)", c.HeaderFooter.ApplyOn)
	}
	if c.HeaderFooter.SideOffsetLeftPt < 0 {
		return fmt.Errorf("header_footer.side_offset_left_pt must be >= 0")
	}
	if c.HeaderFooter.SideOffsetRightPt < 0 {
		return fmt.Errorf("header_footer.side_offset_right_pt must be >= 0")
	}
	if c.HeaderFooter.FooterReserveAbovePt < 0 {
		return fmt.Errorf("header_footer.footer_reserve_above_pt must be >= 0")
	}
	if !c.HeaderFooter.Enabled {
		if len(c.HeaderFooter.Header.Grid.Cells) > 0 || len(c.HeaderFooter.Footer.Grid.Cells) > 0 {
			return fmt.Errorf("header_footer.enabled must be true when header/footer cells are configured")
		}
	}

	if c.HeaderFooter.PageNumber.Enabled && strings.TrimSpace(c.HeaderFooter.PageNumber.Format) == "" {
		return fmt.Errorf("header_footer.page_number.format must not be empty when page numbering is enabled")
	}

	if err := validateTextStyle(c.HeaderFooter.GlobalStyle, "header_footer.global_style"); err != nil {
		return err
	}
	if err := validateHeaderFooterRegion(c.HeaderFooter.Header, "header_footer.header"); err != nil {
		return err
	}
	if err := validateHeaderFooterRegion(c.HeaderFooter.Footer, "header_footer.footer"); err != nil {
		return err
	}

	switch c.Features.PlantUML {
	case "auto", "on", "off":
	default:
		return fmt.Errorf("invalid features.plantuml %q (allowed: auto, on, off)", c.Features.PlantUML)
	}

	return nil
}

func validateLevelRange(fromLevel, toLevel int, key string) error {
	if fromLevel < 1 || fromLevel > 6 {
		return fmt.Errorf("%s.from_level must be between 1 and 6", key)
	}
	if toLevel < 1 || toLevel > 6 {
		return fmt.Errorf("%s.to_level must be between 1 and 6", key)
	}
	if fromLevel > toLevel {
		return fmt.Errorf("%s.from_level must be <= %s.to_level", key, key)
	}
	return nil
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
var namedColorPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func validateTextStyle(style TextStyleConfig, prefix string) error {
	if err := validateColorValue(style.Color, prefix+".color"); err != nil {
		return err
	}
	if style.SizePt < 0 {
		return fmt.Errorf("%s.size_pt must be >= 0", prefix)
	}
	if style.LineHeightPt < 0 {
		return fmt.Errorf("%s.line_height_pt must be >= 0", prefix)
	}
	if style.Opacity < 0 || style.Opacity > 1 {
		return fmt.Errorf("%s.opacity must be between 0 and 1", prefix)
	}
	switch style.Weight {
	case "", "normal", "bold":
	default:
		return fmt.Errorf("%s.weight must be normal or bold", prefix)
	}
	return nil
}

func validateBlockQuoteStyle(style BlockQuoteStyleConfig, prefix string) error {
	if err := validateColorValue(style.BarColor, prefix+".bar_color"); err != nil {
		return err
	}
	if err := validateColorValue(style.TextColor, prefix+".text_color"); err != nil {
		return err
	}
	if err := validateColorValue(style.BackgroundColor, prefix+".background_color"); err != nil {
		return err
	}
	if style.BarWidthPt < 0 {
		return fmt.Errorf("%s.bar_width_pt must be >= 0", prefix)
	}
	if style.GapPt < 0 {
		return fmt.Errorf("%s.gap_pt must be >= 0", prefix)
	}
	if style.PaddingPt < 0 {
		return fmt.Errorf("%s.padding_pt must be >= 0", prefix)
	}
	return nil
}

func validatePlantUMLStyle(style PlantUMLStyleConfig, prefix string) error {
	switch style.Align {
	case "", "left", "center", "right":
	default:
		return fmt.Errorf("%s.align must be left, center, or right", prefix)
	}
	if style.SpaceBeforePt < 0 {
		return fmt.Errorf("%s.space_before_pt must be >= 0", prefix)
	}
	if style.SpaceAfterPt < 0 {
		return fmt.Errorf("%s.space_after_pt must be >= 0", prefix)
	}
	return nil
}

func validateHeadingStyle(style HeadingStyleConfig, prefix string) error {
	levels := map[string]HeadingLevelStyleConfig{
		"h1": style.H1,
		"h2": style.H2,
		"h3": style.H3,
		"h4": style.H4,
		"h5": style.H5,
		"h6": style.H6,
	}
	for level, cfg := range levels {
		levelPrefix := prefix + "." + level
		if err := validateColorValue(cfg.Color, levelPrefix+".color"); err != nil {
			return err
		}
		if cfg.SizePt != nil && *cfg.SizePt <= 0 {
			return fmt.Errorf("%s.size_pt must be > 0", levelPrefix)
		}
		if cfg.SpaceBeforePt != nil && *cfg.SpaceBeforePt < 0 {
			return fmt.Errorf("%s.space_before_pt must be >= 0", levelPrefix)
		}
		if cfg.SpaceAfterPt != nil && *cfg.SpaceAfterPt < 0 {
			return fmt.Errorf("%s.space_after_pt must be >= 0", levelPrefix)
		}
	}
	return nil
}

func validateLinksStyle(style LinksStyleConfig, prefix string) error {
	if err := validateColorValue(style.Color, prefix+".color"); err != nil {
		return err
	}
	if err := validateColorValue(style.URLColor, prefix+".url_color"); err != nil {
		return err
	}
	if err := validateColorValue(style.CitationColor, prefix+".citation_color"); err != nil {
		return err
	}
	if err := validateColorValue(style.TOCColor, prefix+".toc_color"); err != nil {
		return err
	}
	return nil
}

func validateColorValue(raw, prefix string) error {
	color := strings.TrimSpace(raw)
	if color == "" {
		return nil
	}
	if !hexColorPattern.MatchString(color) && !namedColorPattern.MatchString(color) {
		return fmt.Errorf("%s must be a named color or #RRGGBB", prefix)
	}
	return nil
}

func validateHeaderFooterRegion(region HeaderFooterRegion, prefix string) error {
	if region.HeightPt < 0 {
		return fmt.Errorf("%s.height_pt must be >= 0", prefix)
	}
	if region.SepPt < 0 {
		return fmt.Errorf("%s.sep_pt must be >= 0", prefix)
	}
	if region.SkipPt < 0 {
		return fmt.Errorf("%s.skip_pt must be >= 0", prefix)
	}
	return validateHeaderFooterGrid(region.Grid, prefix+".grid")
}

func validateHeaderFooterGrid(grid HeaderFooterGrid, prefix string) error {
	if len(grid.Columns) == 0 {
		return fmt.Errorf("%s.columns must not be empty", prefix)
	}
	if len(grid.Rows) == 0 {
		return fmt.Errorf("%s.rows must not be empty", prefix)
	}
	for i, c := range grid.Columns {
		if c <= 0 {
			return fmt.Errorf("%s.columns[%d] must be > 0", prefix, i)
		}
	}
	for i, r := range grid.Rows {
		if r <= 0 {
			return fmt.Errorf("%s.rows[%d] must be > 0", prefix, i)
		}
	}

	seen := map[string]struct{}{}
	for i, cell := range grid.Cells {
		if cell.Row <= 0 || cell.Row > len(grid.Rows) {
			return fmt.Errorf("%s.cells[%d].row out of bounds", prefix, i)
		}
		if cell.Col <= 0 || cell.Col > len(grid.Columns) {
			return fmt.Errorf("%s.cells[%d].col out of bounds", prefix, i)
		}

		if cell.RowSpan == 0 {
			cell.RowSpan = 1
		}
		if cell.ColSpan == 0 {
			cell.ColSpan = 1
		}
		if cell.RowSpan != 1 || cell.ColSpan != 1 {
			return fmt.Errorf("%s.cells[%d] spans are currently unsupported (row_span and col_span must be 1)", prefix, i)
		}

		switch cell.AlignH {
		case "", "left", "center", "right":
		default:
			return fmt.Errorf("%s.cells[%d].align_h must be left, center, or right", prefix, i)
		}
		switch cell.AlignV {
		case "", "top", "middle", "bottom":
		default:
			return fmt.Errorf("%s.cells[%d].align_v must be top, middle, or bottom", prefix, i)
		}

		key := fmt.Sprintf("%d:%d", cell.Row, cell.Col)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%s.cells has duplicate cell at row=%d col=%d", prefix, cell.Row, cell.Col)
		}
		seen[key] = struct{}{}

		for j, block := range cell.Blocks {
			if err := validateHeaderFooterBlock(block, fmt.Sprintf("%s.cells[%d].blocks[%d]", prefix, i, j)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateHeaderFooterBlock(block HeaderFooterBlock, prefix string) error {
	switch block.Type {
	case "text":
		if strings.TrimSpace(block.Value) == "" {
			return fmt.Errorf("%s.value must not be empty for type=text", prefix)
		}
	case "image":
		if strings.TrimSpace(block.Path) == "" {
			return fmt.Errorf("%s.path must not be empty for type=image", prefix)
		}
	case "page_number":
		// Optional format; falls back to header_footer.page_number.format.
	default:
		return fmt.Errorf("%s.type must be one of: text, image, page_number", prefix)
	}

	if block.WidthPt < 0 {
		return fmt.Errorf("%s.width_pt must be >= 0", prefix)
	}
	if block.HeightPt < 0 {
		return fmt.Errorf("%s.height_pt must be >= 0", prefix)
	}

	return validateTextStyle(block.Style, prefix+".style")
}
