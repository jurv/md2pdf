package config

import (
	"fmt"

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
	Colors ColorsConfig `yaml:"colors"`
	Fonts  FontsConfig  `yaml:"fonts"`
}

type ColorsConfig struct {
	Primary string `yaml:"primary"`
}

type FontsConfig struct {
	Body    string `yaml:"body"`
	Heading string `yaml:"heading"`
}

type HeaderFooterConfig struct {
	HeaderLeft  string `yaml:"header_left"`
	HeaderRight string `yaml:"header_right"`
	FooterLeft  string `yaml:"footer_left"`
	FooterRight string `yaml:"footer_right"`
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
			Mode: "none",
			Builtin: BuiltinCoverConfig{
				TitleColor:      "#000000",
				BackgroundColor: "#FFFFFF",
				Align:           "center",
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

	switch c.Cover.Builtin.Align {
	case "", "center", "top":
	default:
		return fmt.Errorf("invalid cover.builtin.align %q (allowed: center, top)", c.Cover.Builtin.Align)
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
