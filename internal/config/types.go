package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PDF          PDFConfig          `yaml:"pdf"`
	Metadata     MetadataConfig     `yaml:"metadata"`
	TOC          TOCConfig          `yaml:"toc"`
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

type TOCConfig struct {
	Mode  string `yaml:"mode"`
	Title string `yaml:"title"`
	Depth int    `yaml:"depth"`
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
		TOC: TOCConfig{
			Mode:  "auto",
			Title: "Table of Contents",
			Depth: 3,
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

	switch c.TOC.Mode {
	case "auto", "on", "off":
	default:
		return fmt.Errorf("invalid toc.mode %q (allowed: auto, on, off)", c.TOC.Mode)
	}

	if c.TOC.Depth <= 0 {
		return fmt.Errorf("toc.depth must be > 0")
	}

	switch c.Features.PlantUML {
	case "auto", "on", "off":
	default:
		return fmt.Errorf("invalid features.plantuml %q (allowed: auto, on, off)", c.Features.PlantUML)
	}

	return nil
}
