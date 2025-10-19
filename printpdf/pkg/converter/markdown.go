package converter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// convertMarkdownToHTML converts markdown content to HTML
func convertMarkdownToHTML(markdown []byte, options PageOptions) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,   // GitHub Flavored Markdown
			extension.Table, // Tables
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(markdown, &buf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Build page CSS with orientation, margin, and zoom
	margin := options.cssMarginValue()
	var pageCSS strings.Builder
	if options.Orientation == "landscape" {
		fmt.Fprintf(&pageCSS, "@page { size: A4 landscape; margin: %s; }\n", margin)
	} else {
		fmt.Fprintf(&pageCSS, "@page { size: A4 portrait; margin: %s; }\n", margin)
	}

	if guide := strings.TrimSpace(options.FirstPageGuide); guide != "" {
		fmt.Fprintf(&pageCSS, "@page:first {\n")
		fmt.Fprintf(&pageCSS, "    background-image: linear-gradient(90deg, #d0d7de, #d0d7de);\n")
		fmt.Fprintf(&pageCSS, "    background-size: 0.4pt 100%%;\n")
		fmt.Fprintf(&pageCSS, "    background-repeat: no-repeat;\n")
		fmt.Fprintf(&pageCSS, "    background-position: %s 0;\n", guide)
		fmt.Fprintf(&pageCSS, "}\n")
	}

	// Calculate zoom factor for font sizes
	zoom := options.Zoom
	if zoom == 0 {
		zoom = 100
	}
	zoomFactor := float64(zoom) / 100.0

	// Build body CSS - don't add max-width or auto margins when using page margins
	// because it conflicts with the user's margin settings
	bodyCSS := `body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    line-height: 1.6;
    color: #24292e;
}`

	// Wrap in a complete HTML document with nice styling
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
%s
html {
    font-size: %.0f%%;
}

%s

h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}

h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
h3 { font-size: 1.25em; }
h4 { font-size: 1em; }
h5 { font-size: 0.875em; }
h6 { font-size: 0.85em; color: #6a737d; }

code {
    background-color: transparent;
    border: 1px solid #d0d7de;
    border-radius: 3px;
    font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
    font-size: 85%%;
    margin: 0;
    padding: 0.2em 0.4em;
    font-feature-settings: "liga" 0, "kern" 0;
    font-variant-ligatures: none;
}

pre {
    background-color: transparent;
    border: 1px solid #d0d7de;
    border-radius: 3px;
    font-size: 85%%;
    line-height: 1.45;
    overflow: auto;
    padding: 16px;
}

pre code {
    background-color: transparent;
    border: 0;
    display: inline;
    line-height: inherit;
    margin: 0;
    overflow: visible;
    padding: 0;
    word-wrap: normal;
    font-feature-settings: "liga" 0, "kern" 0;
    font-variant-ligatures: none;
}

table {
    border-collapse: collapse;
    border-spacing: 0;
    margin-top: 0;
    margin-bottom: 16px;
}

table th {
    font-weight: 600;
    padding: 6px 13px;
    border: 1px solid #dfe2e5;
}

table td {
    padding: 6px 13px;
    border: 1px solid #dfe2e5;
}

table tr {
    background-color: #fff;
    border-top: 1px solid #c6cbd1;
}

table tr:nth-child(2n) {
    background-color: #f6f8fa;
}

blockquote {
    border-left: 0.25em solid #000;
    color: inherit;
    padding: 0 1em;
    margin-left: 0;
    font-style: italic;
}

ul, ol {
    padding-left: 2em;
    margin-top: 0;
    margin-bottom: 16px;
}

li + li {
    margin-top: 0.25em;
}

a {
    color: inherit;
    text-decoration: underline;
}

a:hover {
    text-decoration: underline;
}

img {
    max-width: 100%%;
    box-sizing: content-box;
}

hr {
    height: 0.25em;
    padding: 0;
    margin: 24px 0;
    background-color: #e1e4e8;
    border: 0;
}
</style>
</head>
<body>
%s
</body>
</html>`, pageCSS.String(), zoomFactor*100, bodyCSS, buf.String())

	// If columns are requested, wrap the content in a container with column CSS
	if options.Columns > 1 {
		// Add column support via CSS
		columnCSS := fmt.Sprintf(`<style>
body {
    column-count: %d;
    column-gap: 2em;
    column-rule: 1px solid #dfe2e5;
}
h1, h2, h3 {
    column-span: all;
}
</style>`, options.Columns)
		// Insert the column CSS before </head>
		htmlBytes := bytes.Replace([]byte(html), []byte("</head>"), []byte(columnCSS+"\n</head>"), 1)
		return htmlBytes, nil
	}

	return []byte(html), nil
}

// wrapHTMLWithPageOptions wraps raw HTML content with proper page styling and options
func wrapHTMLWithPageOptions(htmlContent []byte, options PageOptions) ([]byte, error) {
	// Build page CSS with orientation, margin, and zoom
	margin := options.cssMarginValue()
	var pageCSS strings.Builder
	if options.Orientation == "landscape" {
		fmt.Fprintf(&pageCSS, "@page { size: A4 landscape; margin: %s; }\n", margin)
	} else {
		fmt.Fprintf(&pageCSS, "@page { size: A4 portrait; margin: %s; }\n", margin)
	}

	if guide := strings.TrimSpace(options.FirstPageGuide); guide != "" {
		fmt.Fprintf(&pageCSS, "@page:first {\n")
		fmt.Fprintf(&pageCSS, "    background-image: linear-gradient(90deg, #d0d7de, #d0d7de);\n")
		fmt.Fprintf(&pageCSS, "    background-size: 0.4pt 100%%;\n")
		fmt.Fprintf(&pageCSS, "    background-repeat: no-repeat;\n")
		fmt.Fprintf(&pageCSS, "    background-position: %s 0;\n", guide)
		fmt.Fprintf(&pageCSS, "}\n")
	}

	// Calculate zoom factor for font sizes
	zoom := options.Zoom
	if zoom == 0 {
		zoom = 100
	}
	zoomFactor := float64(zoom) / 100.0

	// Check if the HTML already has <html>, <head> tags
	htmlStr := string(htmlContent)
	hasHTML := strings.Contains(strings.ToLower(htmlStr), "<html")
	hasHead := strings.Contains(strings.ToLower(htmlStr), "<head")

	var result []byte

	if hasHTML && hasHead {
		// HTML is already a complete document, inject our styles into <head>
		// Insert our page CSS before the closing </head> tag
		styleBlock := fmt.Sprintf(`<style>
%s
html {
    font-size: %.0f%%;
}
</style>`, pageCSS.String(), zoomFactor*100)

		if hasHead {
			// Insert before </head>
			result = bytes.Replace(htmlContent, []byte("</head>"), []byte(styleBlock+"\n</head>"), 1)
		} else {
			// Insert after <html> tag
			htmlTag := "<html"
			insertPos := strings.Index(strings.ToLower(htmlStr), htmlTag)
			if insertPos != -1 {
				// Find the end of the <html> tag
				endPos := strings.Index(htmlStr[insertPos:], ">") + insertPos + 1
				result = []byte(htmlStr[:endPos] + "\n<head>\n" + styleBlock + "\n</head>\n" + htmlStr[endPos:])
			} else {
				result = htmlContent
			}
		}
	} else {
		// HTML is a fragment, wrap it in a complete document
		wrappedHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
%s
html {
    font-size: %.0f%%;
}
</style>
</head>
<body>
%s
</body>
</html>`, pageCSS.String(), zoomFactor*100, htmlStr)
		result = []byte(wrappedHTML)
	}

	// Handle columns if requested
	if options.Columns > 1 {
		columnCSS := fmt.Sprintf(`<style>
body {
    column-count: %d;
    column-gap: 2em;
    column-rule: 1px solid #dfe2e5;
}
h1, h2, h3 {
    column-span: all;
}
</style>`, options.Columns)

		// Insert the column CSS before </head>
		result = bytes.Replace(result, []byte("</head>"), []byte(columnCSS+"\n</head>"), 1)
	}

	return result, nil
}
