package render

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/julien/md2pdf/internal/config"
)

var emojiLookPath = exec.LookPath

func ContainsEmoji(markdown []byte) bool {
	return len(collectEmojiSequences(string(markdown))) > 0
}

type emojiImageSet struct {
	FilterPath string
	ImagesDir  string
}

func prepareEmojiImageFilter(ctx context.Context, workDir string, markdown []byte, cfg config.Config) (emojiImageSet, error) {
	style := cfg.Style.Emoji
	mode := strings.TrimSpace(style.Mode)
	if mode == "" {
		mode = "auto"
	}
	if mode == "none" {
		return emojiImageSet{}, nil
	}
	sequences := collectEmojiSequences(emojiScanInput(markdown, cfg))
	if len(sequences) == 0 {
		return emojiImageSet{}, nil
	}
	if _, err := emojiLookPath("pango-view"); err != nil {
		if mode == "auto" {
			return emojiImageSet{}, nil
		}
		return emojiImageSet{}, fmt.Errorf("emoji image rendering requires pango-view: %w", err)
	}

	imagesDir := filepath.Join(workDir, "md2pdf-emoji")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		return emojiImageSet{}, fmt.Errorf("failed to create emoji image directory: %w", err)
	}

	font := strings.TrimSpace(style.Font)
	if font == "" {
		font = "Noto Color Emoji 32"
	}
	mapping := make(map[string]string, len(sequences))
	for _, sequence := range sequences {
		name := emojiImageName(sequence)
		path := filepath.Join(imagesDir, name)
		if err := renderEmojiImage(ctx, font, sequence, path); err != nil {
			return emojiImageSet{}, err
		}
		mapping[sequence] = path
	}

	filter, err := compileEmojiImageFilter(mapping)
	if err != nil {
		return emojiImageSet{}, err
	}
	filterPath := filepath.Join(workDir, "md2pdf-emoji-images.lua")
	if err := os.WriteFile(filterPath, []byte(filter), 0o600); err != nil {
		return emojiImageSet{}, fmt.Errorf("failed to write emoji image filter: %w", err)
	}
	return emojiImageSet{FilterPath: filterPath, ImagesDir: imagesDir}, nil
}

func emojiScanInput(markdown []byte, cfg config.Config) string {
	parts := []string{string(markdown), cfg.Metadata.Title, cfg.Metadata.Author, cfg.Metadata.Subject, cfg.Cover.Builtin.Subtitle}
	return strings.Join(parts, "\n")
}

func renderEmojiImage(ctx context.Context, font, sequence, outputPath string) error {
	cmd := exec.CommandContext(ctx, "pango-view",
		"--no-display",
		"--background=transparent",
		"--font="+font,
		"--text="+sequence,
		"--output="+outputPath,
	)
	combined, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to render emoji %q with pango-view: %w (%s)", sequence, err, strings.TrimSpace(string(combined)))
	}
	return nil
}

func emojiImageName(sequence string) string {
	sum := sha256.Sum256([]byte(sequence))
	return fmt.Sprintf("emoji-%x.png", sum[:12])
}

func collectEmojiSequences(input string) []string {
	seen := map[string]struct{}{}
	var out []string
	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if !isEmojiBase(r) {
			i += size
			continue
		}
		start := i
		i += size
		i = consumeEmojiSuffix(input, i)
		for i < len(input) {
			zr, zsize := utf8.DecodeRuneInString(input[i:])
			if zr != '\u200d' {
				break
			}
			nextStart := i + zsize
			nr, nsize := utf8.DecodeRuneInString(input[nextStart:])
			if !isEmojiBase(nr) {
				break
			}
			i = consumeEmojiSuffix(input, nextStart+nsize)
		}
		sequence := input[start:i]
		if _, ok := seen[sequence]; !ok {
			seen[sequence] = struct{}{}
			out = append(out, sequence)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) == len(out[j]) {
			return out[i] < out[j]
		}
		return len(out[i]) > len(out[j])
	})
	return out
}

func consumeEmojiSuffix(input string, i int) int {
	for i < len(input) {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == '\ufe0f' || r == '\ufe0e' || r == '\u20e3' || (r >= 0x1F3FB && r <= 0x1F3FF) {
			i += size
			continue
		}
		break
	}
	return i
}

func isEmojiBase(r rune) bool {
	switch {
	case r >= 0x1F000 && r <= 0x1FAFF:
		return true
	case r == 0x2139:
		return true
	case r >= 0x2194 && r <= 0x21AA:
		return true
	case r >= 0x231A && r <= 0x23FF:
		return true
	case r >= 0x25AA && r <= 0x27BF:
		return true
	default:
		return false
	}
}

func compileEmojiImageFilter(mapping map[string]string) (string, error) {
	keys := make([]string, 0, len(mapping))
	for key := range mapping {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) == len(keys[j]) {
			return keys[i] < keys[j]
		}
		return len(keys[i]) > len(keys[j])
	})

	var b strings.Builder
	b.WriteString("local emoji = {\n")
	for _, key := range keys {
		fmt.Fprintf(&b, "  { source = %s, path = %q },\n", luaUTF8StringExpr(key), filepath.ToSlash(mapping[key]))
	}
	b.WriteString("}\n\n")
	b.WriteString(emojiImageFilterLua)
	return b.String(), nil
}

func luaUTF8StringExpr(value string) string {
	runes := []rune(value)
	parts := make([]string, 0, len(runes))
	for _, r := range runes {
		parts = append(parts, fmt.Sprintf("utf8.char(0x%X)", r))
	}
	return strings.Join(parts, " .. ")
}

const emojiImageFilterLua = `
local function is_latex()
  return FORMAT:match('latex') or FORMAT:match('pdf')
end

local function next_char_len(text, index)
  local b = text:byte(index)
  if not b then return 0 end
  if b < 0x80 then return 1 end
  if b < 0xE0 then return 2 end
  if b < 0xF0 then return 3 end
  return 4
end

local function latex_escape(text)
  local replacements = {
    ['\\'] = '\\textbackslash{}',
    ['{'] = '\\{',
    ['}'] = '\\}',
    ['#'] = '\\#',
    ['$'] = '\\$',
    ['%'] = '\\%',
    ['&'] = '\\&',
    ['_'] = '\\_',
    ['^'] = '\\textasciicircum{}',
    ['~'] = '\\textasciitilde{}',
  }
  return (text:gsub('[\\{}#%$%%&_\94~]', replacements))
end

local function latex_code_escape(text)
  local escaped = latex_escape(text)
  escaped = escaped:gsub(' ', '\\mdtwocodespace{}')
  escaped = escaped:gsub('\t', '\\mdtwocodetab{}')
  return escaped
end

local function find_match(text, index)
  for _, item in ipairs(emoji) do
    if text:sub(index, index + #item.source - 1) == item.source then
      return item
    end
  end
  return nil
end

local function has_emoji(text)
  local index = 1
  while index <= #text do
    local item = find_match(text, index)
    if item then return true end
    index = index + next_char_len(text, index)
  end
  return false
end

local function replace_str(text)
  local out = pandoc.List:new()
  local chunk_start = 1
  local index = 1
  while index <= #text do
    local item = find_match(text, index)
    if item then
      if chunk_start < index then
        out:insert(pandoc.Str(text:sub(chunk_start, index - 1)))
      end
      out:insert(pandoc.RawInline('latex', '\\mdtwoemoji{' .. item.path .. '}'))
      index = index + #item.source
      chunk_start = index
    else
      index = index + next_char_len(text, index)
    end
  end
  if chunk_start <= #text then
    out:insert(pandoc.Str(text:sub(chunk_start)))
  end
  return out
end

local function replace_code_inline(text)
  local out = pandoc.List:new()
  local chunk_start = 1
  local index = 1
  while index <= #text do
    local item = find_match(text, index)
    if item then
      if chunk_start < index then
        out:insert(pandoc.RawInline('latex', '\\texttt{' .. latex_code_escape(text:sub(chunk_start, index - 1)) .. '}'))
      end
      out:insert(pandoc.RawInline('latex', '\\mdtwoemoji{' .. item.path .. '}'))
      index = index + #item.source
      chunk_start = index
    else
      index = index + next_char_len(text, index)
    end
  end
  if chunk_start <= #text then
    out:insert(pandoc.RawInline('latex', '\\texttt{' .. latex_code_escape(text:sub(chunk_start)) .. '}'))
  end
  return out
end

local function replace_code_line(text)
  local out = {}
  local chunk_start = 1
  local index = 1
  while index <= #text do
    local item = find_match(text, index)
    if item then
      if chunk_start < index then
        table.insert(out, latex_code_escape(text:sub(chunk_start, index - 1)))
      end
      table.insert(out, '\\mdtwoemoji{' .. item.path .. '}')
      index = index + #item.source
      chunk_start = index
    else
      index = index + next_char_len(text, index)
    end
  end
  if chunk_start <= #text then
    table.insert(out, latex_code_escape(text:sub(chunk_start)))
  end
  return table.concat(out)
end

local function replace_meta_string(value)
  if not has_emoji(value) then
    return value
  end
  return pandoc.MetaInlines(replace_str(value))
end

local function replace_meta_inlines(value)
  local out = pandoc.List:new()
  for _, inline in ipairs(value) do
    if inline.t == 'Str' and has_emoji(inline.text) then
      out:extend(replace_str(inline.text))
    elseif inline.t == 'Code' and has_emoji(inline.text) then
      out:extend(replace_code_inline(inline.text))
    else
      out:insert(inline)
    end
  end
  return pandoc.MetaInlines(out)
end

function Meta(meta)
  for _, key in ipairs({ 'title', 'author', 'date', 'subtitle' }) do
    local value = meta[key]
    if type(value) == 'string' then
      meta[key] = replace_meta_string(value)
    elseif type(value) == 'table' and value.t == 'MetaInlines' then
      meta[key] = replace_meta_inlines(value)
    end
  end
  return meta
end

function Str(elem)
  if not is_latex() or not has_emoji(elem.text) then
    return nil
  end
  return replace_str(elem.text)
end

function Code(elem)
  if not is_latex() or not has_emoji(elem.text) then
    return nil
  end
  return replace_code_inline(elem.text)
end

function CodeBlock(elem)
  if not is_latex() or not has_emoji(elem.text) then
    return nil
  end
  local lines = {}
  for line in (elem.text .. '\n'):gmatch('(.-)\n') do
    table.insert(lines, '\\noindent ' .. replace_code_line(line) .. '\\par')
  end
  return pandoc.RawBlock('latex', '\\begin{mdtwoemojicodeblock}\n' .. table.concat(lines, '\n') .. '\n\\end{mdtwoemojicodeblock}')
end
`
