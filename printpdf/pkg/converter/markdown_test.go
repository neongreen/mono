package converter

import (
	"strings"
	"testing"

	stdhtml "golang.org/x/net/html"
)

// Test markdown-to-HTML body conversion (without CSS wrapping)
func TestMarkdownToHTMLBody(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:  "simple heading",
			input: "# Hello World",
			contains: []string{
				"<h1",
				"Hello World",
				"</h1>",
			},
			notContains: []string{
				"<html>",
				"<head>",
				"<style>",
			},
		},
		{
			name:  "paragraph",
			input: "This is a paragraph.",
			contains: []string{
				"<p>",
				"This is a paragraph.",
				"</p>",
			},
		},
		{
			name:  "code block",
			input: "```go\nfunc main() {}\n```",
			contains: []string{
				"<pre>",
				"<code",
				"func main()",
				"</code>",
				"</pre>",
			},
		},
		{
			name:  "inline code",
			input: "Use `fmt.Println()` to print.",
			contains: []string{
				"<code>",
				"fmt.Println()",
				"</code>",
			},
		},
		{
			name:  "table",
			input: "| Name | Age |\n|------|-----|\n| John | 30  |",
			contains: []string{
				"<table>",
				"<thead>",
				"<th>",
				"Name",
				"Age",
				"</th>",
				"<tbody>",
				"<td>",
				"John",
				"30",
				"</td>",
				"</table>",
			},
		},
		{
			name:  "unordered list",
			input: "- Item 1\n- Item 2\n- Item 3",
			contains: []string{
				"<ul>",
				"<li>",
				"Item 1",
				"Item 2",
				"Item 3",
				"</li>",
				"</ul>",
			},
		},
		{
			name:  "ordered list",
			input: "1. First\n2. Second\n3. Third",
			contains: []string{
				"<ol>",
				"<li>",
				"First",
				"Second",
				"Third",
				"</li>",
				"</ol>",
			},
		},
		{
			name:  "blockquote",
			input: "> This is a quote",
			contains: []string{
				"<blockquote>",
				"This is a quote",
				"</blockquote>",
			},
		},
		{
			name:  "link",
			input: "[GitHub](https://github.com)",
			contains: []string{
				"<a",
				"href=\"https://github.com\"",
				"GitHub",
				"</a>",
			},
		},
		{
			name:  "strikethrough (GFM)",
			input: "~~strikethrough~~",
			contains: []string{
				"<del>",
				"strikethrough",
				"</del>",
			},
		},
		{
			name:  "task list (GFM)",
			input: "- [x] Done\n- [ ] Todo",
			contains: []string{
				"<input",
				"type=\"checkbox\"",
				"checked",
				"Done",
				"Todo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := markdownToHTMLBody([]byte(tt.input))
			if err != nil {
				t.Fatalf("markdownToHTMLBody failed: %v", err)
			}

			resultStr := string(result)
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", expected, resultStr)
				}
			}

			for _, notExpected := range tt.notContains {
				if strings.Contains(resultStr, notExpected) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput: %s", notExpected, resultStr)
				}
			}
		})
	}
}

// Test footnote processing
func TestMarkdownToHTMLBodyFootnotes(t *testing.T) {
	markdown := "Text with footnote[^1].\n\n[^1]: Footnote content."

	result, err := markdownToHTMLBody([]byte(markdown))
	if err != nil {
		t.Fatalf("markdownToHTMLBody failed: %v", err)
	}

	resultStr := string(result)

	// Check for inline footnote span after enhancement
	if !strings.Contains(resultStr, "<span class=\"printpdf-footnote\"") {
		t.Errorf("Expected inline footnote span in output")
	}

	// Check that the footnote content is present
	if !strings.Contains(resultStr, "Footnote content") {
		t.Errorf("Expected footnote content in output")
	}

	// Check that backref is removed
	if strings.Contains(resultStr, "footnote-backref") {
		t.Errorf("Expected footnote backref to be removed")
	}

	// Check that footnote section is removed
	if strings.Contains(resultStr, "class=\"footnotes\"") {
		t.Errorf("Expected footnote section to be removed")
	}
}

// Test CSS generation functions
func TestGeneratePageCSS(t *testing.T) {
	tests := []struct {
		name     string
		options  PageOptions
		contains []string
	}{
		{
			name: "portrait with default margin",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
			},
			contains: []string{
				"@page",
				"size: A4 portrait",
				"margin: 2cm",
			},
		},
		{
			name: "landscape orientation",
			options: PageOptions{
				Orientation: "landscape",
				Margin:      "1in",
			},
			contains: []string{
				"@page",
				"size: A4 landscape",
				"margin: 1in",
			},
		},
		{
			name: "per-edge margins",
			options: PageOptions{
				Orientation:  "portrait",
				Margin:       "2cm",
				MarginTop:    "3cm",
				MarginRight:  "1cm",
				MarginBottom: "2cm",
				MarginLeft:   "1.5cm",
			},
			contains: []string{
				"@page",
				"margin: 3cm 1cm 2cm 1.5cm",
			},
		},
		{
			name: "with first page guide",
			options: PageOptions{
				Orientation:    "portrait",
				Margin:         "2cm",
				FirstPageGuide: "3cm",
			},
			contains: []string{
				"@page:first",
				"background-image",
				"background-position: 3cm 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css := generatePageCSS(tt.options)

			for _, expected := range tt.contains {
				if !strings.Contains(css, expected) {
					t.Errorf("Expected CSS to contain %q, but it didn't.\nCSS: %s", expected, css)
				}
			}
		})
	}
}

func TestGenerateBodyCSS(t *testing.T) {
	css := generateBodyCSS()

	expected := []string{
		"body",
		"font-family",
		"line-height",
		"color",
	}

	for _, exp := range expected {
		if !strings.Contains(css, exp) {
			t.Errorf("Expected body CSS to contain %q, but it didn't.\nCSS: %s", exp, css)
		}
	}
}

func TestGenerateContentCSS(t *testing.T) {
	css := generateContentCSS()

	expected := []string{
		"h1", "h2", "h3", "h4", "h5", "h6",
		"code",
		"pre",
		"table",
		"blockquote",
		"ul", "ol",
		"a",
		"img",
		"hr",
	}

	for _, exp := range expected {
		if !strings.Contains(css, exp) {
			t.Errorf("Expected content CSS to contain %q, but it didn't", exp)
		}
	}
}

func TestGenerateColumnCSS(t *testing.T) {
	tests := []struct {
		name     string
		columns  int
		contains []string
	}{
		{
			name:    "two columns",
			columns: 2,
			contains: []string{
				"column-count: 2",
				"column-gap",
				"column-rule",
			},
		},
		{
			name:    "three columns",
			columns: 3,
			contains: []string{
				"column-count: 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css := generateColumnCSS(tt.columns)

			for _, expected := range tt.contains {
				if !strings.Contains(css, expected) {
					t.Errorf("Expected column CSS to contain %q, but it didn't.\nCSS: %s", expected, css)
				}
			}
		})
	}
}

// Test HTML wrapping with document structure
func TestWrapHTMLBodyWithDocument(t *testing.T) {
	bodyContent := []byte("<h1>Hello</h1><p>World</p>")

	tests := []struct {
		name     string
		options  PageOptions
		contains []string
	}{
		{
			name: "basic document structure",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<html>",
				"<head>",
				"<meta charset=\"UTF-8\">",
				"<style>",
				"@page",
				"body {",
				"</style>",
				"</head>",
				"<body>",
				"<h1>Hello</h1>",
				"<p>World</p>",
				"</body>",
				"</html>",
			},
		},
		{
			name: "with zoom 150%",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        150,
			},
			contains: []string{
				"font-size: 150%",
			},
		},
		{
			name: "with columns",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
				Columns:     2,
			},
			contains: []string{
				"column-count: 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapHTMLBodyWithDocument(bodyContent, tt.options)
			resultStr := string(result)

			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Expected wrapped HTML to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

// Test the full pipeline: markdown -> HTML with document
func TestConvertMarkdownToHTMLFullPipeline(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		options  PageOptions
		contains []string
	}{
		{
			name:     "simple markdown with default options",
			markdown: "# Hello\n\nThis is **bold** text.",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<h1",
				"Hello",
				"</h1>",
				"<strong>",
				"bold",
				"</strong>",
				"@page",
				"size: A4 portrait",
			},
		},
		{
			name:     "markdown with code block",
			markdown: "# Code Example\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "1cm",
				Zoom:        100,
			},
			contains: []string{
				"<pre>",
				"<code",
				"func main()",
				"fmt.Println",
				"font-variant-ligatures: none",
			},
		},
		{
			name:     "markdown with table",
			markdown: "| Name | Value |\n|------|-------|\n| A    | 1     |\n| B    | 2     |",
			options: PageOptions{
				Orientation: "portrait",
				Margin:      "2cm",
				Zoom:        100,
			},
			contains: []string{
				"<table>",
				"<thead>",
				"<tbody>",
				"Name",
				"Value",
				"border-collapse",
			},
		},
		{
			name:     "landscape with custom zoom",
			markdown: "# Landscape Test",
			options: PageOptions{
				Orientation: "landscape",
				Margin:      "1.5cm",
				Zoom:        120,
			},
			contains: []string{
				"size: A4 landscape",
				"margin: 1.5cm",
				"font-size: 120%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertMarkdownToHTML([]byte(tt.markdown), tt.options)
			if err != nil {
				t.Fatalf("convertMarkdownToHTML failed: %v", err)
			}

			resultStr := string(result)
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Expected HTML to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

// Test HTML node manipulation utilities
func TestHTMLNodeManipulation(t *testing.T) {
	t.Run("hasClass", func(t *testing.T) {
		node := &stdhtml.Node{
			Type: stdhtml.ElementNode,
			Data: "div",
			Attr: []stdhtml.Attribute{
				{Key: "class", Val: "foo bar baz"},
			},
		}

		if !hasClass(node, "foo") {
			t.Error("Expected hasClass to return true for 'foo'")
		}
		if !hasClass(node, "bar") {
			t.Error("Expected hasClass to return true for 'bar'")
		}
		if hasClass(node, "qux") {
			t.Error("Expected hasClass to return false for 'qux'")
		}
		if hasClass(nil, "foo") {
			t.Error("Expected hasClass to return false for nil node")
		}
	})

	t.Run("getAttr", func(t *testing.T) {
		node := &stdhtml.Node{
			Type: stdhtml.ElementNode,
			Data: "a",
			Attr: []stdhtml.Attribute{
				{Key: "href", Val: "https://example.com"},
				{Key: "id", Val: "link1"},
			},
		}

		if val := getAttr(node, "href"); val != "https://example.com" {
			t.Errorf("Expected href to be 'https://example.com', got %q", val)
		}
		if val := getAttr(node, "id"); val != "link1" {
			t.Errorf("Expected id to be 'link1', got %q", val)
		}
		if val := getAttr(node, "class"); val != "" {
			t.Errorf("Expected class to be empty, got %q", val)
		}
		if val := getAttr(nil, "href"); val != "" {
			t.Errorf("Expected getAttr on nil node to return empty string, got %q", val)
		}
	})

	t.Run("hasAttr", func(t *testing.T) {
		node := &stdhtml.Node{
			Type: stdhtml.ElementNode,
			Data: "div",
			Attr: []stdhtml.Attribute{
				{Key: "id", Val: "test"},
			},
		}

		if !hasAttr(node, "id") {
			t.Error("Expected hasAttr to return true for 'id'")
		}
		if hasAttr(node, "class") {
			t.Error("Expected hasAttr to return false for 'class'")
		}
		if hasAttr(nil, "id") {
			t.Error("Expected hasAttr to return false for nil node")
		}
	})

	t.Run("isBlockLevelElement", func(t *testing.T) {
		blockElements := []string{"div", "p", "h1", "h2", "ul", "ol", "table", "blockquote", "pre"}
		inlineElements := []string{"span", "a", "strong", "em", "code"}

		for _, elem := range blockElements {
			if !isBlockLevelElement(elem) {
				t.Errorf("Expected %q to be a block-level element", elem)
			}
		}

		for _, elem := range inlineElements {
			if isBlockLevelElement(elem) {
				t.Errorf("Expected %q NOT to be a block-level element", elem)
			}
		}
	})

	t.Run("isWhitespaceNode", func(t *testing.T) {
		whitespaceNodes := []*stdhtml.Node{
			{Type: stdhtml.TextNode, Data: "   "},
			{Type: stdhtml.TextNode, Data: "\n\t"},
			{Type: stdhtml.TextNode, Data: ""},
		}

		for _, node := range whitespaceNodes {
			if !isWhitespaceNode(node) {
				t.Errorf("Expected node with data %q to be whitespace", node.Data)
			}
		}

		nonWhitespaceNode := &stdhtml.Node{Type: stdhtml.TextNode, Data: "text"}
		if isWhitespaceNode(nonWhitespaceNode) {
			t.Error("Expected node with text to NOT be whitespace")
		}

		if isWhitespaceNode(nil) {
			t.Error("Expected nil node to NOT be whitespace")
		}
	})

	t.Run("extractTextContent", func(t *testing.T) {
		// Create a simple node tree: <p>Hello <strong>World</strong>!</p>
		p := &stdhtml.Node{Type: stdhtml.ElementNode, Data: "p"}
		text1 := &stdhtml.Node{Type: stdhtml.TextNode, Data: "Hello "}
		strong := &stdhtml.Node{Type: stdhtml.ElementNode, Data: "strong"}
		text2 := &stdhtml.Node{Type: stdhtml.TextNode, Data: "World"}
		text3 := &stdhtml.Node{Type: stdhtml.TextNode, Data: "!"}

		p.AppendChild(text1)
		p.AppendChild(strong)
		strong.AppendChild(text2)
		p.AppendChild(text3)

		result := extractTextContent(p)
		expected := "Hello World!"
		if result != expected {
			t.Errorf("Expected text content to be %q, got %q", expected, result)
		}

		if extractTextContent(nil) != "" {
			t.Error("Expected extractTextContent on nil node to return empty string")
		}
	})
}

// Test edge cases
func TestMarkdownToHTMLBodyEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result, err := markdownToHTMLBody([]byte(""))
		if err != nil {
			t.Fatalf("Expected no error for empty input, got: %v", err)
		}
		// Empty input produces empty output or just whitespace
		_ = result
	})

	t.Run("raw HTML in markdown", func(t *testing.T) {
		input := "# Title\n\n<div class=\"custom\">Raw HTML</div>"
		result, err := markdownToHTMLBody([]byte(input))
		if err != nil {
			t.Fatalf("Failed to process raw HTML: %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, "<div class=\"custom\">") {
			t.Error("Expected raw HTML to be preserved")
		}
	})

	t.Run("multiple footnotes", func(t *testing.T) {
		input := "First[^1] and second[^2].\n\n[^1]: First note.\n[^2]: Second note."
		result, err := markdownToHTMLBody([]byte(input))
		if err != nil {
			t.Fatalf("Failed to process multiple footnotes: %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, "First note") || !strings.Contains(resultStr, "Second note") {
			t.Error("Expected both footnotes to be present")
		}
	})
}

// Test wrapping raw HTML with page options
func TestWrapHTMLWithPageOptions(t *testing.T) {
	t.Run("wrap simple HTML fragment", func(t *testing.T) {
		htmlContent := []byte("<h1>Title</h1><p>Content</p>")
		options := PageOptions{
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
		}

		result, err := wrapHTMLWithPageOptions(htmlContent, options)
		if err != nil {
			t.Fatalf("wrapHTMLWithPageOptions failed: %v", err)
		}

		resultStr := string(result)
		contains := []string{
			"<!DOCTYPE html>",
			"<html>",
			"<head>",
			"@page",
			"size: A4 portrait",
			"margin: 2cm",
			"<h1>Title</h1>",
			"<p>Content</p>",
		}

		for _, expected := range contains {
			if !strings.Contains(resultStr, expected) {
				t.Errorf("Expected wrapped HTML to contain %q", expected)
			}
		}
	})

	t.Run("inject CSS into complete HTML document", func(t *testing.T) {
		htmlContent := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
</head>
<body>
<h1>Title</h1>
</body>
</html>`)
		options := PageOptions{
			Orientation: "landscape",
			Margin:      "1in",
			Zoom:        120,
		}

		result, err := wrapHTMLWithPageOptions(htmlContent, options)
		if err != nil {
			t.Fatalf("wrapHTMLWithPageOptions failed: %v", err)
		}

		resultStr := string(result)
		contains := []string{
			"<title>Test</title>", // Original content preserved
			"@page",
			"size: A4 landscape",
			"margin: 1in",
			"font-size: 120%",
		}

		for _, expected := range contains {
			if !strings.Contains(resultStr, expected) {
				t.Errorf("Expected injected HTML to contain %q", expected)
			}
		}
	})

	t.Run("with columns", func(t *testing.T) {
		htmlContent := []byte("<h1>Title</h1><p>Content</p>")
		options := PageOptions{
			Orientation: "portrait",
			Margin:      "2cm",
			Zoom:        100,
			Columns:     3,
		}

		result, err := wrapHTMLWithPageOptions(htmlContent, options)
		if err != nil {
			t.Fatalf("wrapHTMLWithPageOptions failed: %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, "column-count: 3") {
			t.Error("Expected column CSS to be added")
		}
	})

	t.Run("with first page guide", func(t *testing.T) {
		htmlContent := []byte("<h1>Title</h1>")
		options := PageOptions{
			Orientation:    "portrait",
			Margin:         "2cm",
			Zoom:           100,
			FirstPageGuide: "5cm",
		}

		result, err := wrapHTMLWithPageOptions(htmlContent, options)
		if err != nil {
			t.Fatalf("wrapHTMLWithPageOptions failed: %v", err)
		}

		resultStr := string(result)
		contains := []string{
			"@page:first",
			"background-position: 5cm 0",
		}

		for _, expected := range contains {
			if !strings.Contains(resultStr, expected) {
				t.Errorf("Expected wrapped HTML to contain %q", expected)
			}
		}
	})
}
