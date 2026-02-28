package frontmatter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/julien/md2pdf/internal/config"
	"gopkg.in/yaml.v3"
)

func ParseMarkdown(content []byte) (map[string]any, []byte, error) {
	text := string(content)
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")

	if !strings.HasPrefix(text, "---\n") {
		return nil, content, nil
	}

	lines := strings.Split(text, "\n")
	end := -1
	for i := 1; i < len(lines); i++ {
		marker := strings.TrimSpace(lines[i])
		if marker == "---" || marker == "..." {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, nil, fmt.Errorf("front matter start marker found but closing marker is missing")
	}

	front := strings.Join(lines[1:end], "\n")
	body := strings.Join(lines[end+1:], "\n")

	if strings.TrimSpace(front) == "" {
		return map[string]any{}, []byte(body), nil
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(front), &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid front matter yaml: %w", err)
	}

	return config.NormalizeMap(raw), []byte(body), nil
}

func Remove(content []byte) ([]byte, error) {
	_, body, err := ParseMarkdown(content)
	if err != nil {
		return nil, err
	}
	if body == nil {
		return bytes.Clone(content), nil
	}
	return body, nil
}
