package converter

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

// Converter interface for PDF conversion
type Converter interface {
	Name() string
	Convert(content []byte, contentType string, outputPath string) error
}

// ParseConverterList parses a comma-separated list of converter names
func ParseConverterList(convertersStr string) []Converter {
	if convertersStr == "all" {
		return []Converter{
			NewTypstConverter(),
			NewPrinceConverter(),
			NewWeasyPrintConverter(),
		}
	}

	var converters []Converter
	for _, name := range strings.Split(convertersStr, ",") {
		name = strings.TrimSpace(name)
		switch strings.ToLower(name) {
		case "typst":
			converters = append(converters, NewTypstConverter())
		case "prince":
			converters = append(converters, NewPrinceConverter())
		case "weasyprint":
			converters = append(converters, NewWeasyPrintConverter())
		}
	}
	return converters
}

// prepareContent converts content to the appropriate format for conversion
func prepareContent(content []byte, contentType string) ([]byte, error) {
	switch contentType {
	case fetcher.ContentTypeMarkdown:
		// Keep as markdown - converters will handle it
		return content, nil
	case fetcher.ContentTypeHTML:
		// Keep as HTML - converters will handle it
		return content, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}
