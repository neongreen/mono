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

func (p *PrinceConverter) Convert(content []byte, contentType string, outputPath string, options PageOptions) error {
	// Get or download Prince
	princePath, err := p.getPrince()
	if err != nil {
		return fmt.Errorf("failed to get prince: %w", err)
	}

	// Prepare input file
	inputFile, cleanup, err := p.prepareInput(content, contentType, options)
	if err != nil {
		return err
	}
	if cleanup {
		defer os.Remove(inputFile)
	}

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

func (p *PrinceConverter) prepareInput(content []byte, contentType string, options PageOptions) (string, bool, error) {
	var ext string
	var htmlContent []byte
	var err error

	switch contentType {
	case fetcher.ContentTypeMarkdown:
		// Convert markdown to HTML for Prince
		ext = ".html"
		htmlContent, err = convertMarkdownToHTML(content, options)
		if err != nil {
			return "", false, fmt.Errorf("failed to convert markdown to HTML: %w", err)
		}
	case fetcher.ContentTypeHTML:
		ext = ".html"
		htmlContent = wrapHTMLWithPageOptions(content, options)
	default:
		return "", false, fmt.Errorf("unsupported content type: %s", contentType)
	}

	if options.KeepIntermediates && options.IntermediateDir != "" {
		if err := os.MkdirAll(options.IntermediateDir, 0o755); err != nil {
			return "", false, fmt.Errorf("failed to create intermediate directory %s: %w", options.IntermediateDir, err)
		}
		file, err := os.CreateTemp(options.IntermediateDir, "stage-*.html")
		if err != nil {
			return "", false, fmt.Errorf("failed to create intermediate HTML file: %w", err)
		}
		if _, err := file.Write(htmlContent); err != nil {
			file.Close()
			return "", false, fmt.Errorf("failed to write intermediate HTML: %w", err)
		}
		if err := file.Close(); err != nil {
			return "", false, fmt.Errorf("failed to close intermediate HTML file: %w", err)
		}
		return file.Name(), false, nil
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "printpdf-*"+ext)
	if err != nil {
		return "", false, err
	}
	if _, err := tmpFile.Write(htmlContent); err != nil {
		tmpFile.Close()
		return "", false, err
	}
	if err := tmpFile.Close(); err != nil {
		return "", false, err
	}

	return tmpFile.Name(), true, nil
}
