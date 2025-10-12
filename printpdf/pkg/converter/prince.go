package converter

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/neongreen/mono/printpdf/pkg/downloader"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

type PrinceConverter struct {
	downloader *downloader.ToolDownloader
}

func NewPrinceConverter() *PrinceConverter {
	dl, _ := downloader.New()
	return &PrinceConverter{downloader: dl}
}

func (p *PrinceConverter) Name() string {
	return "prince"
}

func (p *PrinceConverter) Convert(content []byte, contentType string, outputPath string) error {
	// Get or download Prince
	princePath, err := p.getPrince()
	if err != nil {
		return fmt.Errorf("failed to get prince: %w", err)
	}

	// Prepare input file
	inputFile, err := p.prepareInput(content, contentType)
	if err != nil {
		return err
	}
	defer os.Remove(inputFile)

	// Run Prince
	cmd := exec.Command(princePath, inputFile, "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("prince failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func (p *PrinceConverter) getPrince() (string, error) {
	// Try system prince first
	if path, err := exec.LookPath("prince"); err == nil {
		return path, nil
	}

	// Prince is commercial software and requires a license
	// We can't automatically download it
	return "", fmt.Errorf("prince not found in PATH. Please install Prince XML from https://www.princexml.com/")
}

func (p *PrinceConverter) prepareInput(content []byte, contentType string) (string, error) {
	var ext string
	var htmlContent []byte
	var err error
	
	switch contentType {
	case fetcher.ContentTypeMarkdown:
		// Convert markdown to HTML for Prince
		ext = ".html"
		htmlContent, err = convertMarkdownToHTML(content)
		if err != nil {
			return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
		}
	case fetcher.ContentTypeHTML:
		ext = ".html"
		htmlContent = content
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
