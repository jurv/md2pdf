package config

import (
	"strings"
	"testing"
)

func TestSymbolStyleValidationRejectsMultiRuneKey(t *testing.T) {
	cfg := Default()
	cfg.Style.Symbols.Replace["ok"] = `\texttt{ok}`

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.symbols.replace keys") {
		t.Fatalf("expected style.symbols key validation error, got %v", err)
	}
}

func TestSymbolStyleValidationRejectsMultilineReplacement(t *testing.T) {
	cfg := Default()
	cfg.Style.Symbols.Replace["✅"] = "\\checkmark\n\\square"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.symbols.replace") {
		t.Fatalf("expected style.symbols replacement validation error, got %v", err)
	}
}

func TestSymbolStyleValidationAcceptsLatexSnippets(t *testing.T) {
	cfg := Default()
	cfg.Style.Symbols.Replace["✅"] = `\mdtwosymbolglyph{☑}{\ensuremath{\checkmark}}`

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid symbol replacement config, got %v", err)
	}
}

func TestSymbolStyleValidationRejectsMultiRuneFallbackEntry(t *testing.T) {
	cfg := Default()
	cfg.Style.Symbols.FallbackFor = append(cfg.Style.Symbols.FallbackFor, "ok")

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "style.symbols.fallback_for") {
		t.Fatalf("expected style.symbols fallback_for validation error, got %v", err)
	}
}

func TestDefaultSymbolStyleIncludesCommonCheckboxes(t *testing.T) {
	cfg := Default()

	if cfg.Style.Symbols.FallbackFont != "Noto Sans Symbols2" {
		t.Fatalf("expected default fallback font %q, got %q", "Noto Sans Symbols2", cfg.Style.Symbols.FallbackFont)
	}
	for _, symbol := range []string{"☐", "☑", "☒"} {
		found := false
		for _, fallback := range cfg.Style.Symbols.FallbackFor {
			if fallback == symbol {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected default symbol fallback set to include %q", symbol)
		}
	}
	for _, symbol := range []string{"💡", "🏁", "📋", "🔍", "🚀", "🚨", "🗹", "🗷", "🗸", "🗵"} {
		if _, ok := cfg.Style.Symbols.Replace[symbol]; !ok {
			t.Fatalf("expected default symbol replacements to include %q", symbol)
		}
	}
}
