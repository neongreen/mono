package fetcher

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// extractReadableContent extracts the main content from HTML
// This is a simplified implementation of Mozilla's Readability algorithm
func extractReadableContent(htmlContent []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the main content
	content := findMainContent(doc)
	if content == nil {
		// Fallback to body
		content = findBody(doc)
	}

	if content == nil {
		return htmlContent, nil // Return as-is if we can't find content
	}

	// Extract and clean the content
	var buf bytes.Buffer
	if err := html.Render(&buf, content); err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	// Wrap in a clean HTML document
	cleanHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    max-width: 800px;
    margin: 40px auto;
    padding: 0 20px;
    line-height: 1.6;
    color: #24292e;
}
h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}
p {
    margin-top: 0;
    margin-bottom: 16px;
}
img {
    max-width: 100%%;
    height: auto;
}
a {
    color: #0366d6;
    text-decoration: none;
}
a:hover {
    text-decoration: underline;
}
</style>
</head>
<body>
%s
</body>
</html>`, buf.String())

	return []byte(cleanHTML), nil
}

// findMainContent tries to find the main content element
func findMainContent(n *html.Node) *html.Node {
	// Look for common content indicators
	if n.Type == html.ElementNode {
		// Check for semantic elements
		if n.Data == "main" || n.Data == "article" {
			return n
		}

		// Check for common content IDs and classes
		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "class" {
				val := strings.ToLower(attr.Val)
				if strings.Contains(val, "main") ||
					strings.Contains(val, "content") ||
					strings.Contains(val, "article") ||
					strings.Contains(val, "post") {
					return n
				}
			}
		}
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findMainContent(c); result != nil {
			return result
		}
	}

	return nil
}

// findBody finds the body element
func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findBody(c); result != nil {
			return result
		}
	}

	return nil
}
