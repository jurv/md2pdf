package render

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/julien/md2pdf/internal/config"
)

var atxHeadingRE = regexp.MustCompile(`^(\s{0,3})(#{1,6})[ \t]+(.+?)\s*$`)
var closeATXRE = regexp.MustCompile(`\s+#+\s*$`)

// ExtractFirstH1 extracts the first level-1 ATX heading found outside fenced code blocks.
// When stripFromBody is true, the heading line is removed from the returned markdown.
func ExtractFirstH1(markdown []byte, stripFromBody bool) (title string, out []byte, found bool) {
	lines := splitLinesPreserveNewline(markdown)
	insideFence := false
	fenceMarker := ""
	removedAt := -1

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if marker, ok := updateFenceState(line, insideFence, fenceMarker); ok {
			if insideFence {
				insideFence = false
				fenceMarker = ""
			} else {
				insideFence = true
				fenceMarker = marker
			}
			continue
		}
		if insideFence {
			continue
		}

		level, text, attr, indent, ok := parseHeadingLine(line)
		_ = attr
		_ = indent
		if !ok || level != 1 {
			continue
		}
		title = text
		found = true
		if stripFromBody {
			lines[i] = ""
			removedAt = i
		}
		break
	}

	if stripFromBody && removedAt >= 0 {
		// Remove one immediate blank line after stripped H1 to avoid leading whitespace block.
		if removedAt+1 < len(lines) && strings.TrimSpace(lines[removedAt+1]) == "" {
			lines[removedAt+1] = ""
		}
	}

	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
	}
	return title, buf.Bytes(), found
}

func ApplyHeadingPolicy(markdown []byte, cfg config.Config) ([]byte, error) {
	lines := splitLinesPreserveNewline(markdown)
	insideFence := false
	fenceMarker := ""
	counters := make([]int, 7) // 1..6

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if marker, ok := updateFenceState(line, insideFence, fenceMarker); ok {
			if insideFence {
				insideFence = false
				fenceMarker = ""
			} else {
				insideFence = true
				fenceMarker = marker
			}
			continue
		}
		if insideFence {
			continue
		}

		level, text, attr, indent, ok := parseHeadingLine(line)
		if !ok {
			continue
		}

		if level < cfg.Heading.FromLevel {
			for reset := cfg.Heading.FromLevel; reset <= 6; reset++ {
				counters[reset] = 0
			}
		}

		// Exclude headings outside ToC bounds.
		if level < cfg.TOC.FromLevel || level > cfg.TOC.ToLevel {
			attr = addAttributeClass(attr, "unlisted")
		}

		if cfg.Heading.Enabled && level >= cfg.Heading.FromLevel && level <= cfg.Heading.ToLevel {
			for reset := level + 1; reset <= 6; reset++ {
				counters[reset] = 0
			}
			if counters[level] == 0 {
				counters[level] = 1
			} else {
				counters[level]++
			}

			for parent := cfg.Heading.FromLevel; parent < level; parent++ {
				if counters[parent] == 0 {
					counters[parent] = 1
				}
			}

			prefix, err := formatHeadingPrefix(counters, level, cfg.Heading)
			if err != nil {
				return nil, err
			}
			if prefix != "" {
				text = buildNumberedHeadingText(text, prefix, cfg.Heading.Suffix)
			}
		}

		lines[i] = renderHeadingLine(indent, level, text, attr, hasTrailingNewline(line))
	}

	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
	}
	return buf.Bytes(), nil
}

func formatHeadingPrefix(counters []int, level int, cfg config.HeadingConfig) (string, error) {
	parts := make([]string, 0, level-cfg.FromLevel+1)
	for l := cfg.FromLevel; l <= level && l <= cfg.ToLevel; l++ {
		notation := cfg.Notation[strconv.Itoa(l)]
		if notation == "" {
			notation = "decimal"
		}
		parts = append(parts, formatCounter(counters[l], notation))
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, cfg.Separator) + cfg.Suffix, nil
}

func formatCounter(value int, notation string) string {
	if value <= 0 {
		value = 1
	}
	switch notation {
	case "roman_upper":
		return toRoman(value)
	case "roman_lower":
		return strings.ToLower(toRoman(value))
	case "alpha_upper":
		return toAlpha(value, true)
	case "alpha_lower":
		return toAlpha(value, false)
	default:
		return strconv.Itoa(value)
	}
}

func toRoman(v int) string {
	numerals := []struct {
		value int
		sym   string
	}{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}
	var b strings.Builder
	for _, n := range numerals {
		for v >= n.value {
			b.WriteString(n.sym)
			v -= n.value
		}
	}
	return b.String()
}

func toAlpha(v int, upper bool) string {
	if v <= 0 {
		v = 1
	}
	letters := "abcdefghijklmnopqrstuvwxyz"
	var out []byte
	for v > 0 {
		v--
		out = append([]byte{letters[v%26]}, out...)
		v /= 26
	}
	s := string(out)
	if upper {
		return strings.ToUpper(s)
	}
	return s
}

func splitLinesPreserveNewline(markdown []byte) []string {
	text := string(markdown)
	if text == "" {
		return []string{""}
	}
	parts := strings.SplitAfter(text, "\n")
	if !strings.HasSuffix(text, "\n") {
		return parts
	}
	return parts
}

func hasTrailingNewline(line string) bool {
	return strings.HasSuffix(line, "\n")
}

func parseHeadingLine(line string) (level int, text, attr, indent string, ok bool) {
	line = strings.TrimSuffix(line, "\n")
	match := atxHeadingRE.FindStringSubmatch(line)
	if match == nil {
		return 0, "", "", "", false
	}
	indent = match[1]
	level = len(match[2])
	raw := strings.TrimSpace(match[3])

	// Remove closing ATX markers if present.
	raw = closeATXRE.ReplaceAllString(raw, "")
	raw = strings.TrimSpace(raw)

	attr = ""
	if strings.HasSuffix(raw, "}") {
		if idx := strings.LastIndex(raw, "{"); idx > 0 {
			candidate := strings.TrimSpace(raw[idx:])
			if strings.HasPrefix(candidate, "{") && strings.HasSuffix(candidate, "}") {
				attr = candidate
				raw = strings.TrimSpace(raw[:idx])
			}
		}
	}

	return level, raw, attr, indent, true
}

func renderHeadingLine(indent string, level int, text, attr string, withNewline bool) string {
	line := indent + strings.Repeat("#", level) + " " + text
	if attr != "" {
		line += " " + attr
	}
	if withNewline {
		line += "\n"
	}
	return line
}

func addAttributeClass(attr, class string) string {
	classToken := "." + class
	if attr == "" {
		return "{" + classToken + "}"
	}
	if strings.Contains(attr, classToken) {
		return attr
	}
	trimmed := strings.TrimSpace(attr)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		body := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if body == "" {
			body = classToken
		} else {
			body = body + " " + classToken
		}
		return "{" + body + "}"
	}
	return attr
}

func buildNumberedHeadingText(text, prefix, suffix string) string {
	cleaned := strings.TrimSpace(text)
	candidates := []string{prefix}
	if suffix != "" && strings.HasSuffix(prefix, suffix) {
		candidates = append(candidates, strings.TrimSuffix(prefix, suffix))
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if stripped, ok := stripLeadingHeadingPrefix(cleaned, candidate); ok {
			cleaned = stripped
			break
		}
	}
	if cleaned == "" {
		return prefix
	}
	return prefix + " " + cleaned
}

func stripLeadingHeadingPrefix(text, prefix string) (string, bool) {
	if text == prefix {
		return "", true
	}
	if !strings.HasPrefix(text, prefix) {
		return "", false
	}
	rest := text[len(prefix):]
	if rest == "" {
		return "", true
	}
	runes := []rune(rest)
	if len(runes) == 0 {
		return "", true
	}
	if unicode.IsSpace(runes[0]) {
		return strings.TrimSpace(rest), true
	}
	// Accept separators like "1) Title" or "1: Title" but reject "1.2".
	if strings.ContainsRune(").:-", runes[0]) && len(runes) > 1 && unicode.IsSpace(runes[1]) {
		return strings.TrimSpace(string(runes[1:])), true
	}
	return "", false
}

func updateFenceState(line string, inside bool, currentMarker string) (marker string, toggled bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
		candidate := "```"
		if strings.HasPrefix(trimmed, "~~~") {
			candidate = "~~~"
		}
		if !inside {
			return candidate, true
		}
		if currentMarker == candidate {
			return candidate, true
		}
	}
	return "", false
}

func ValidateHeadingCompatibility(cfg config.Config) error {
	if cfg.Heading.FromLevel < 1 || cfg.Heading.ToLevel > 6 || cfg.Heading.FromLevel > cfg.Heading.ToLevel {
		return fmt.Errorf("invalid heading_numbering range")
	}
	if cfg.TOC.FromLevel < 1 || cfg.TOC.ToLevel > 6 || cfg.TOC.FromLevel > cfg.TOC.ToLevel {
		return fmt.Errorf("invalid toc range")
	}
	return nil
}
