package converter

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/neongreen/mono/printpdf/pkg/downloader"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

type TypstConverter struct {
	downloader *downloader.ToolDownloader
}

func NewTypstConverter() *TypstConverter {
	dl, _ := downloader.New()
	return &TypstConverter{downloader: dl}
}

func (t *TypstConverter) Name() string {
	return "typst"
}

func (t *TypstConverter) Convert(content []byte, contentType string, outputPath string, options PageOptions) error {
	// Get or download Typst
	typstPath, err := t.getTypst()
	if err != nil {
		return fmt.Errorf("failed to get typst: %w", err)
	}

	// Prepare input file
	var inputFile string
	if contentType == fetcher.ContentTypeMarkdown {
		// Convert markdown to Typst format
		inputFile, err = t.prepareMarkdownInput(content, options)
	} else {
		// For HTML, we need to convert to markdown first
		return fmt.Errorf("HTML conversion not yet supported for Typst")
	}
	if err != nil {
		return err
	}
	defer os.Remove(inputFile)

	// Run Typst
	cmd := exec.Command(typstPath, "compile", inputFile, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func (t *TypstConverter) getTypst() (string, error) {
	// Try system typst first
	if path, err := exec.LookPath("typst"); err == nil {
		return path, nil
	}

	// Download typst
	version := "0.12.0"
	downloadURL := t.getDownloadURL(version)

	return t.downloader.GetTool("typst", version, downloadURL)
}

func (t *TypstConverter) getDownloadURL(version string) string {
	platform := downloader.GetPlatform()

	// Typst release URLs
	// https://github.com/typst/typst/releases/download/v0.12.0/typst-x86_64-unknown-linux-musl.tar.xz
	switch platform {
	case "linux-amd64":
		return fmt.Sprintf("https://github.com/typst/typst/releases/download/v%s/typst-x86_64-unknown-linux-musl.tar.xz", version)
	case "linux-arm64":
		return fmt.Sprintf("https://github.com/typst/typst/releases/download/v%s/typst-aarch64-unknown-linux-musl.tar.xz", version)
	case "darwin-amd64":
		return fmt.Sprintf("https://github.com/typst/typst/releases/download/v%s/typst-x86_64-apple-darwin.tar.xz", version)
	case "darwin-arm64":
		return fmt.Sprintf("https://github.com/typst/typst/releases/download/v%s/typst-aarch64-apple-darwin.tar.xz", version)
	default:
		return ""
	}
}

func (t *TypstConverter) prepareMarkdownInput(content []byte, options PageOptions) (string, error) {
	// Convert Markdown to Typst format
	typstContent, err := convertMarkdownToTypst(content, options)
	if err != nil {
		return "", fmt.Errorf("failed to convert markdown to typst: %w", err)
	}

	// Create a temporary Typst file
	tmpFile, err := os.CreateTemp("", "printpdf-*.typ")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Write the converted Typst content
	if _, err := tmpFile.WriteString(typstContent); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
