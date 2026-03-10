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

func buildUnicodeSymbolMetadata(cfg config.Config, workDir string) ([][2]string, error) {
	if len(cfg.Style.Symbols.Replace) == 0 && len(cfg.Style.Symbols.FallbackFor) == 0 {
		return nil, nil
	}

	partial, err := compileUnicodeSymbolPartial(cfg.Style.Symbols)
	if err != nil {
		return nil, err
	}
	partialPath := filepath.Join(workDir, "md2pdf-unicode-symbols.tex")
	if err := os.WriteFile(partialPath, []byte(partial), 0o600); err != nil {
		return nil, fmt.Errorf("failed to write unicode symbol partial: %w", err)
	}

	return [][2]string{
		{"unicode_symbols_partial", partialPath},
	}, nil
}

func compileUnicodeSymbolPartial(style config.SymbolStyleConfig) (string, error) {
	var b strings.Builder

	b.WriteString("% md2pdf Unicode symbol helpers\n")
	b.WriteString("\\providecommand{\\mdtwosymbolglyph}[2]{#2}\n")
	if fontName := strings.TrimSpace(style.FallbackFont); fontName != "" {
		fmt.Fprintf(&b, "\\ifxetex\n  \\newfontfamily\\mdtwosymbolfont{%s}\n  \\renewcommand{\\mdtwosymbolglyph}[2]{{\\mdtwosymbolfont #1}}\n\\else\n  \\ifluatex\n    \\newfontfamily\\mdtwosymbolfont{%s}\n    \\renewcommand{\\mdtwosymbolglyph}[2]{{\\mdtwosymbolfont #1}}\n  \\fi\n\\fi\n", fontName, fontName)
	}

	replaced := make(map[string]struct{}, len(style.Replace))
	for symbol := range style.Replace {
		replaced[symbol] = struct{}{}
	}
	for _, symbol := range sortedUniqueSymbols(style.FallbackFor) {
		if _, overridden := replaced[symbol]; overridden {
			continue
		}
		replacement := defaultSymbolLatexFallback(symbol)
		r, _ := utf8.DecodeRuneInString(symbol)
		fmt.Fprintf(&b, "%% U+%04X\n", r)
		fmt.Fprintf(&b, "\\newunicodechar{%s}{\\mdtwosymbolglyph{%s}{%s}}\n", symbol, symbol, replacement)
	}

	for _, symbol := range sortedMapKeys(style.Replace) {
		replacement := style.Replace[symbol]
		r, _ := utf8.DecodeRuneInString(symbol)
		fmt.Fprintf(&b, "%% U+%04X\n", r)
		fmt.Fprintf(&b, "\\newunicodechar{%s}{%s}\n", symbol, replacement)
	}

	return b.String(), nil
}

func sortedUniqueSymbols(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	keys := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func defaultSymbolLatexFallback(symbol string) string {
	switch symbol {
	case "☐", "◻", "⬜":
		return `\ensuremath{\square}`
	case "☑", "✓", "✔":
		return `\ensuremath{\checkmark}`
	case "☒":
		return `\ensuremath{\boxtimes}`
	case "✗", "✖":
		return `\ensuremath{\times}`
	case "⚠":
		return `\textbf{!}`
	default:
		return symbol
	}
}
