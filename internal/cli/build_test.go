package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julien/md2pdf/internal/render"
)

func TestNewBuildCmdDefinesCompressionFlags(t *testing.T) {
	cmd := newBuildCmd(&GlobalOptions{})

	compressFlag := cmd.Flags().Lookup("compress")
	if compressFlag == nil {
		t.Fatalf("expected build command to define --compress")
	}
	if compressFlag.DefValue != "false" {
		t.Fatalf("expected --compress default to be false, got %q", compressFlag.DefValue)
	}

	qualityFlag := cmd.Flags().Lookup("compress-quality")
	if qualityFlag == nil {
		t.Fatalf("expected build command to define --compress-quality")
	}
	if qualityFlag.DefValue != "printer" {
		t.Fatalf("expected --compress-quality default to be printer, got %q", qualityFlag.DefValue)
	}
}

func TestRunBuildWithoutCompressionRendersToFinalOutput(t *testing.T) {
	restore := stubBuildPipeline(t)
	defer restore()

	inputPath := writeTempMarkdown(t, "# Title\n\nBody.\n")
	outputPath := filepath.Join(t.TempDir(), "out.pdf")
	flags := &buildFlags{Output: outputPath}
	cmd := newBuildCmd(&GlobalOptions{})

	var renderedOutput string
	generatePDFFunc = func(_ context.Context, opts render.Options) error {
		renderedOutput = opts.OutputPath
		return nil
	}

	if err := runBuild(context.Background(), &GlobalOptions{}, cmd, flags, inputPath); err != nil {
		t.Fatalf("runBuild returned error: %v", err)
	}
	if renderedOutput != outputPath {
		t.Fatalf("expected direct render output %q, got %q", outputPath, renderedOutput)
	}
}

func TestRunBuildWithCompressionUsesIntermediatePDF(t *testing.T) {
	restore := stubBuildPipeline(t)
	defer restore()

	inputPath := writeTempMarkdown(t, "# Title\n\nBody.\n")
	outputPath := filepath.Join(t.TempDir(), "out.pdf")
	flags := &buildFlags{Output: outputPath, Compress: true, CompressQuality: "ebook"}
	cmd := newBuildCmd(&GlobalOptions{})

	var renderedOutput string
	var compressedInput string
	var compressedOutput string
	var compressedQuality string
	generatePDFFunc = func(_ context.Context, opts render.Options) error {
		renderedOutput = opts.OutputPath
		return os.WriteFile(opts.OutputPath, []byte("raw-pdf"), 0o600)
	}
	compressPDFFunc = func(input, output, quality string) error {
		compressedInput = input
		compressedOutput = output
		compressedQuality = quality
		if _, err := os.Stat(input); err != nil {
			return err
		}
		return os.WriteFile(output, []byte("compressed-pdf"), 0o600)
	}

	if err := runBuild(context.Background(), &GlobalOptions{}, cmd, flags, inputPath); err != nil {
		t.Fatalf("runBuild returned error: %v", err)
	}
	if renderedOutput == outputPath {
		t.Fatalf("expected compression flow to render to an intermediate PDF, got final output path %q", renderedOutput)
	}
	if compressedInput != renderedOutput {
		t.Fatalf("expected compression input %q to match render output %q", compressedInput, renderedOutput)
	}
	if compressedOutput != outputPath {
		t.Fatalf("expected compression output %q, got %q", outputPath, compressedOutput)
	}
	if compressedQuality != "ebook" {
		t.Fatalf("expected compression quality ebook, got %q", compressedQuality)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected compressed output file to exist, got %v", err)
	}
	if _, err := os.Stat(renderedOutput); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected intermediate PDF to be cleaned up, got err=%v", err)
	}
}

func TestRunBuildCompressionMapsGhostscriptErrorToDependencyError(t *testing.T) {
	restore := stubBuildPipeline(t)
	defer restore()

	inputPath := writeTempMarkdown(t, "# Title\n\nBody.\n")
	outputPath := filepath.Join(t.TempDir(), "out.pdf")
	flags := &buildFlags{Output: outputPath, Compress: true, CompressQuality: "printer"}
	cmd := newBuildCmd(&GlobalOptions{})

	generatePDFFunc = func(_ context.Context, opts render.Options) error {
		return os.WriteFile(opts.OutputPath, []byte("raw-pdf"), 0o600)
	}
	compressPDFFunc = func(_, _, _ string) error {
		return errors.New("ghostscript binary 'gs' is required for compression")
	}

	err := runBuild(context.Background(), &GlobalOptions{}, cmd, flags, inputPath)
	if err == nil {
		t.Fatalf("expected runBuild to fail when ghostscript is unavailable")
	}
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != ExitDependency {
		t.Fatalf("expected dependency error code %d, got %d", ExitDependency, appErr.Code)
	}
	if !strings.Contains(err.Error(), "ghostscript is required") {
		t.Fatalf("expected ghostscript dependency error, got %v", err)
	}
}

func stubBuildPipeline(t *testing.T) func() {
	t.Helper()
	oldGenerate := generatePDFFunc
	oldCompress := compressPDFFunc
	oldDeps := ensureBuildDependenciesFunc
	ensureBuildDependenciesFunc = func(string, bool) error { return nil }
	return func() {
		generatePDFFunc = oldGenerate
		compressPDFFunc = oldCompress
		ensureBuildDependenciesFunc = oldDeps
	}
}

func writeTempMarkdown(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.md")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("failed to write markdown fixture: %v", err)
	}
	return path
}
