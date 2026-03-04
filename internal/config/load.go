package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type LoadOptions struct {
	GlobalPath  string
	ProjectPath string
	BaseDir     string
	FrontMatter map[string]any
	Overrides   map[string]any
}

func Load(opts LoadOptions) (Config, error) {
	globalPath, err := resolveGlobalConfigPath(opts.GlobalPath)
	if err != nil {
		return Config{}, err
	}
	projectPath, err := resolveProjectConfigPath(opts.ProjectPath, opts.BaseDir)
	if err != nil {
		return Config{}, err
	}

	merged := DefaultMap()
	if globalPath != "" {
		cfgMap, err := LoadMap(globalPath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to load global config %q: %w", globalPath, err)
		}
		applyCompatibilityAliases(cfgMap)
		merged = MergeMap(merged, cfgMap)
	}
	if projectPath != "" {
		cfgMap, err := LoadMap(projectPath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to load project config %q: %w", projectPath, err)
		}
		applyCompatibilityAliases(cfgMap)
		merged = MergeMap(merged, cfgMap)
	}
	if len(opts.FrontMatter) > 0 {
		fm := NormalizeMap(opts.FrontMatter)
		applyCompatibilityAliases(fm)
		merged = MergeMap(merged, fm)
	}
	if len(opts.Overrides) > 0 {
		ov := NormalizeMap(opts.Overrides)
		applyCompatibilityAliases(ov)
		merged = MergeMap(merged, ov)
	}
	applyCompatibilityAliases(merged)
	if err := validateDeprecatedKeys(merged); err != nil {
		return Config{}, err
	}

	blob, err := yaml.Marshal(merged)
	if err != nil {
		return Config{}, fmt.Errorf("failed to serialize merged config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(blob, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode merged config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyCompatibilityAliases(merged map[string]any) {
	tocAny, ok := merged["toc"]
	if !ok {
		applyHeadingStyleAliases(merged)
		return
	}
	tocMap, ok := tocAny.(map[string]any)
	if !ok {
		applyHeadingStyleAliases(merged)
		return
	}

	if _, hasToLevel := tocMap["to_level"]; !hasToLevel {
		if depth, ok := tocMap["depth"]; ok {
			tocMap["to_level"] = depth
		}
	}
	if _, hasDepth := tocMap["depth"]; !hasDepth {
		if toLevel, ok := tocMap["to_level"]; ok {
			tocMap["depth"] = toLevel
		}
	}
	applyHeadingStyleAliases(merged)
}

func applyHeadingStyleAliases(merged map[string]any) {
	styleAny, ok := merged["style"]
	if !ok {
		return
	}
	styleMap, ok := styleAny.(map[string]any)
	if !ok {
		return
	}
	headingsAny, ok := styleMap["headings"]
	if !ok {
		return
	}
	headingsMap, ok := headingsAny.(map[string]any)
	if !ok {
		return
	}

	ensureLevelMap := func(level string) map[string]any {
		existing, ok := headingsMap[level]
		if ok {
			if cast, castOK := existing.(map[string]any); castOK {
				return cast
			}
		}
		fresh := map[string]any{}
		headingsMap[level] = fresh
		return fresh
	}

	if rawColor, ok := headingsMap["color"]; ok {
		for _, level := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
			levelMap := ensureLevelMap(level)
			if _, exists := levelMap["color"]; !exists {
				levelMap["color"] = rawColor
			}
		}
		delete(headingsMap, "color")
	}

	for _, level := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		legacySizeKey := level + "_size_pt"
		legacySizeValue, ok := headingsMap[legacySizeKey]
		if !ok {
			continue
		}
		levelMap := ensureLevelMap(level)
		if _, exists := levelMap["size_pt"]; !exists {
			levelMap["size_pt"] = legacySizeValue
		}
		delete(headingsMap, legacySizeKey)
	}
}

func validateDeprecatedKeys(merged map[string]any) error {
	hfAny, ok := merged["header_footer"]
	if !ok {
		return nil
	}
	hfMap, ok := hfAny.(map[string]any)
	if !ok {
		return nil
	}
	deprecated := []string{"header_left", "header_right", "footer_left", "footer_right"}
	for _, key := range deprecated {
		if _, exists := hfMap[key]; exists {
			return fmt.Errorf("deprecated key header_footer.%s detected; migrate to header_footer.header.grid.cells / header_footer.footer.grid.cells", key)
		}
	}
	return nil
}

func LoadMap(path string) (map[string]any, error) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(blob, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return map[string]any{}, nil
	}
	return NormalizeMap(raw), nil
}

func resolveGlobalConfigPath(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("global config not found: %w", err)
		}
		return explicit, nil
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidate := filepath.Join(xdg, "md2pdf", "config.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil
	}
	candidate := filepath.Join(home, ".config", "md2pdf", "config.yaml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", nil
}

func resolveProjectConfigPath(explicit, baseDir string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("project config not found: %w", err)
		}
		return explicit, nil
	}

	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return "", nil
		}
	}

	candidates := []string{
		filepath.Join(baseDir, "md2pdf.yaml"),
		filepath.Join(baseDir, ".md2pdf.yaml"),
	}
	for _, candidate := range candidates {
		_, err := os.Stat(candidate)
		if err == nil {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	return "", nil
}
