package converter

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/neongreen/mono/printpdf/pkg/downloader"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

type WeasyPrintConverter struct {
	downloader *downloader.ToolDownloader
}

func NewWeasyPrintConverter() *WeasyPrintConverter {
	dl, _ := downloader.New()
	return &WeasyPrintConverter{downloader: dl}
}

func (w *WeasyPrintConverter) Name() string {
	return "weasyprint"
}

func (w *WeasyPrintConverter) Convert(content []byte, contentType string, outputPath string, options PageOptions) error {
	// Get or download WeasyPrint
	weasyPath, err := w.getWeasyPrint()
	if err != nil {
		return fmt.Errorf("failed to get weasyprint: %w", err)
	}

	// Prepare input file
	inputFile, err := w.prepareInput(content, contentType, options)
	if err != nil {
		return err
	}
	defer os.Remove(inputFile)

	// Run WeasyPrint
	cmd := exec.Command(weasyPath, inputFile, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("weasyprint failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func (w *WeasyPrintConverter) getWeasyPrint() (string, error) {
	// Try system weasyprint first
	if path, err := exec.LookPath("weasyprint"); err == nil {
		return path, nil
	}

	// WeasyPrint is a Python package, typically installed via pip
	// Check if it's available in common locations
	pythonPaths := []string{"python3", "python"}
	for _, pyPath := range pythonPaths {
		cmd := exec.Command(pyPath, "-m", "weasyprint", "--version")
		if err := cmd.Run(); err == nil {
			// Create a wrapper script
			return w.createPythonWrapper(pyPath)
		}
	}

	return "", fmt.Errorf("weasyprint not found. Please install with: pip install weasyprint")
}

func (w *WeasyPrintConverter) createPythonWrapper(pythonPath string) (string, error) {
	// Create a simple wrapper script that calls Python's weasyprint module
	wrapperPath := "/tmp/weasyprint-wrapper.sh"
	script := fmt.Sprintf(`#!/bin/sh
exec %s -m weasyprint "$@"
`, pythonPath)

	if err := os.WriteFile(wrapperPath, []byte(script), 0755); err != nil {
		return "", err
	}

	return wrapperPath, nil
}

func (w *WeasyPrintConverter) prepareInput(content []byte, contentType string, options PageOptions) (string, error) {
	var ext string
	var htmlContent []byte
	var err error

	switch contentType {
	case fetcher.ContentTypeMarkdown:
		// Convert markdown to HTML for WeasyPrint
		ext = ".html"
		htmlContent, err = convertMarkdownToHTML(content, options)
		if err != nil {
			return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
		}
	case fetcher.ContentTypeHTML:
		ext = ".html"
		htmlContent, err = wrapHTMLWithPageOptions(content, options)
		if err != nil {
			return "", fmt.Errorf("failed to wrap HTML with page options: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "printpdf-*"+ext)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(htmlContent); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
