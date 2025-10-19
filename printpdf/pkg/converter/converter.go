package converter

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

// PageOptions contains page layout options for PDF conversion
type PageOptions struct {
	Columns           int    // Number of columns (1 = no columns)
	Orientation       string // "portrait" or "landscape"
	Margin            string // Default page margin (e.g., "2cm", "1in")
	MarginTop         string // Optional top margin override
	MarginRight       string // Optional right margin override
	MarginBottom      string // Optional bottom margin override
	MarginLeft        string // Optional left margin override
	Zoom              int    // Zoom percentage (e.g., 80 for 80%, 100 is default)
	FirstPageGuide    string // Optional distance for a vertical guide on the first page (e.g., "3cm")
	KeepIntermediates bool   // When true, intermediate artifacts (HTML, Typst, etc.) are preserved on disk
	IntermediateDir   string // Directory where intermediate artifacts should be stored
}

// resolveMargins returns each side's margin along with the default value that was applied.
func (o PageOptions) resolveMargins() (top, right, bottom, left, uniform string) {
	uniform = o.Margin
	if uniform == "" {
		uniform = "2cm"
	}

	top = uniform
	if o.MarginTop != "" {
		top = o.MarginTop
	}

	right = uniform
	if o.MarginRight != "" {
		right = o.MarginRight
	}

	bottom = uniform
	if o.MarginBottom != "" {
		bottom = o.MarginBottom
	}

	left = uniform
	if o.MarginLeft != "" {
		left = o.MarginLeft
	}

	return top, right, bottom, left, uniform
}

// cssMarginValue returns the CSS margin shorthand that should be applied.
func (o PageOptions) cssMarginValue() string {
	top, right, bottom, left, uniform := o.resolveMargins()
	if top == uniform && right == uniform && bottom == uniform && left == uniform {
		return uniform
	}

	return strings.Join([]string{top, right, bottom, left}, " ")
}

// typstMarginValue returns the Typst margin expression.
func (o PageOptions) typstMarginValue() string {
	top, right, bottom, left, uniform := o.resolveMargins()
	if top == uniform && right == uniform && bottom == uniform && left == uniform {
		return uniform
	}

	return fmt.Sprintf("(top: %s, right: %s, bottom: %s, left: %s)", top, right, bottom, left)
}

// Converter interface for PDF conversion
type Converter interface {
	Name() string
	Convert(content []byte, contentType string, outputPath string, options PageOptions) error
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
