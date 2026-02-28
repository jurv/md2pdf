package pdf

import (
	"fmt"
	"os"
	"os/exec"
)

func Merge(output string, inputs []string) error {
	if len(inputs) < 2 {
		return fmt.Errorf("merge requires at least two input PDF files")
	}
	for _, input := range inputs {
		if _, err := os.Stat(input); err != nil {
			return fmt.Errorf("input PDF not found: %s", input)
		}
	}

	if _, err := exec.LookPath("pdftk"); err == nil {
		args := append([]string{}, inputs...)
		args = append(args, "cat", "output", output)
		cmd := exec.Command("pdftk", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("pdftk merge failed: %w (%s)", err, string(out))
		}
		return nil
	}

	if _, err := exec.LookPath("pdfunite"); err == nil {
		args := append([]string{}, inputs...)
		args = append(args, output)
		cmd := exec.Command("pdfunite", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("pdfunite merge failed: %w (%s)", err, string(out))
		}
		return nil
	}

	return fmt.Errorf("no PDF merge backend found (install pdftk or pdfunite)")
}

func Compress(input, output, quality string) error {
	if input == output {
		return fmt.Errorf("input and output paths must be different for compression")
	}
	if _, err := os.Stat(input); err != nil {
		return fmt.Errorf("input PDF not found: %s", input)
	}
	if _, err := exec.LookPath("gs"); err != nil {
		return fmt.Errorf("ghostscript binary 'gs' is required for compression")
	}

	setting := map[string]string{
		"screen":   "/screen",
		"ebook":    "/ebook",
		"printer":  "/printer",
		"prepress": "/prepress",
	}[quality]
	if setting == "" {
		return fmt.Errorf("invalid compression quality %q", quality)
	}

	args := []string{
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=" + setting,
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile=" + output,
		input,
	}
	cmd := exec.Command("gs", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ghostscript compression failed: %w (%s)", err, string(out))
	}
	return nil
}
