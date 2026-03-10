package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/julien/md2pdf/internal/config"
)

func writeCodeSymbolNormalizeFilter(workDir string, style config.SymbolStyleConfig) (string, error) {
	content, err := compileCodeSymbolNormalizeFilter(style)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	path := filepath.Join(workDir, "md2pdf-code-symbol-normalize.lua")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("failed to write code symbol normalization filter: %w", err)
	}
	return path, nil
}

func compileCodeSymbolNormalizeFilter(style config.SymbolStyleConfig) (string, error) {
	replacements := buildCodeSymbolNormalizeMap(style)
	if len(replacements) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("local replacements = {\n")
	for _, key := range keys {
		value := replacements[key]
		fmt.Fprintf(&b, "  [%s] = %s,\n", luaUTF8CharExpr(key), luaUTF8CharExpr(value))
	}
	b.WriteString("}\n\n")
	b.WriteString(`local function normalize(text)
  if text == nil or text == "" then
    return text
  end
  for source, target in pairs(replacements) do
    text = text:gsub(source, target)
  end
  return text
end

function Code(elem)
  elem.text = normalize(elem.text)
  return elem
end

function CodeBlock(elem)
  elem.text = normalize(elem.text)
  return elem
end
`)

	return b.String(), nil
}

func buildCodeSymbolNormalizeMap(style config.SymbolStyleConfig) map[string]string {
	out := make(map[string]string)
	for symbol, replacement := range style.Replace {
		if replacement == "" {
			out[symbol] = ""
			continue
		}
		target, ok := extractSymbolGlyphTarget(replacement)
		if !ok {
			continue
		}
		out[symbol] = target
	}
	return out
}

func extractSymbolGlyphTarget(replacement string) (string, bool) {
	const prefix = `\mdtwosymbolglyph{`
	if !strings.HasPrefix(replacement, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(replacement, prefix)
	end := strings.Index(rest, "}")
	if end <= 0 {
		return "", false
	}
	target := rest[:end]
	if utf8.RuneCountInString(target) != 1 {
		return "", false
	}
	return target, true
}

func luaUTF8CharExpr(value string) string {
	if value == "" {
		return `""`
	}
	runes := []rune(value)
	if len(runes) != 1 {
		return fmt.Sprintf("%q", value)
	}
	return fmt.Sprintf("utf8.char(0x%X)", runes[0])
}
