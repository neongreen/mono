package converter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	htmlrenderer "github.com/yuin/goldmark/renderer/html"
	stdhtml "golang.org/x/net/html"
	htmlatom "golang.org/x/net/html/atom"
)

const htmlFootnoteCSS = `sup[id^="fnref:"] {
    font-size: 0.75em;
    vertical-align: super;
}

a.footnote-ref {
    text-decoration: none;
    color: inherit;
}

.printpdf-footnote {
    float: footnote;
    font-size: 0.75em;
    color: #4a4a4a;
    margin-top: 0.5em;
    footnote-style-position: inside;
}

.printpdf-footnote::footnote-marker {
    font-weight: 600;
}
`

// convertMarkdownToHTML converts markdown content to HTML
func convertMarkdownToHTML(markdown []byte, options PageOptions) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Footnote,
			extension.Table, // Tables
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			htmlrenderer.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(markdown, &buf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	content, err := enhanceHTMLFootnotes(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to enhance footnotes: %w", err)
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
</html>`, pageCSS.String(), zoomFactor*100, bodyCSS, htmlFootnoteCSS, string(content))

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
	lower := strings.ToLower(htmlStr)
	hasHTML := strings.Contains(lower, "<html")
	hasHead := strings.Contains(lower, "<head")

	var result []byte

	if hasHTML && hasHead {
		styleBlock := fmt.Sprintf(`<style>
%s
html {
    font-size: %.0f%%;
}

%s
</style>
`, pageCSS.String(), zoomFactor*100, htmlFootnoteCSS)

		lower := strings.ToLower(htmlStr)
		closingIdx := strings.Index(lower, "</head>")
		if closingIdx >= 0 {
			closingLen := len("</head>")
			result = append(result, htmlContent[:closingIdx]...)
			result = append(result, []byte(styleBlock)...)
			result = append(result, htmlContent[closingIdx:closingIdx+closingLen]...)
			result = append(result, htmlContent[closingIdx+closingLen:]...)
		} else {
			result = make([]byte, len(htmlContent))
			copy(result, htmlContent)
			result = append(result, []byte(styleBlock)...)
		}
	} else {
		wrapped := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
%s
html {
    font-size: %.0f%%;
}

%s
</style>
</head>
<body>
%s
</body>
</html>`, pageCSS.String(), zoomFactor*100, htmlFootnoteCSS, htmlStr)
		result = []byte(wrapped)
	}

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
		result = bytes.Replace(result, []byte("</head>"), []byte(columnCSS+"\n</head>"), 1)
	}

	return result, nil
}

func attachHTMLFootnotes(root *stdhtml.Node, footnotes map[string][]*stdhtml.Node) int {
	if len(footnotes) == 0 {
		return 0
	}

	var attached int

	var walk func(*stdhtml.Node)
	walk = func(node *stdhtml.Node) {
		if node.Type == stdhtml.ElementNode && node.Data == "a" && hasClass(node, "footnote-ref") {
			target := strings.TrimPrefix(getAttr(node, "href"), "#")
			if target == "" {
				// Nothing to attach
			} else if content, ok := footnotes[target]; ok && len(content) > 0 {
				sup := node.Parent
				if sup != nil {
					removeAttr(node, "href")
					setDataAttribute(sup, "data-footnote-id", target)

					inlineContent, inline := buildInlineFootnoteContent(content)

					if inline && len(inlineContent) > 0 {
						container := &stdhtml.Node{Type: stdhtml.ElementNode, DataAtom: htmlatom.Span, Data: "span"}
						container.Attr = append(container.Attr, stdhtml.Attribute{Key: "class", Val: "printpdf-footnote"})
						container.Attr = append(container.Attr, stdhtml.Attribute{Key: "data-footnote-id", Val: target})

						for _, inlineNode := range inlineContent {
							container.AppendChild(inlineNode)
						}

						stripFootnoteBackrefs(container)

						if sup.Parent != nil {
							insertAfter(sup.Parent, container, sup)
							attached++
						}
					} else {
						container := &stdhtml.Node{Type: stdhtml.ElementNode, DataAtom: htmlatom.Div, Data: "div"}
						container.Attr = append(container.Attr, stdhtml.Attribute{Key: "class", Val: "printpdf-footnote"})
						container.Attr = append(container.Attr, stdhtml.Attribute{Key: "data-footnote-id", Val: target})

						for _, child := range content {
							container.AppendChild(cloneNode(child))
						}

						stripFootnoteBackrefs(container)

						if sup.Parent != nil {
							parent := sup.Parent
							insertionParent := parent.Parent
							if insertionParent != nil {
								insertAfter(insertionParent, container, parent)
							} else {
								insertAfter(parent, container, sup)
							}
							attached++
						}
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)
	return attached
}

func findFootnoteSection(root *stdhtml.Node) *stdhtml.Node {
	var section *stdhtml.Node
	var walk func(*stdhtml.Node)
	walk = func(node *stdhtml.Node) {
		if section != nil || node == nil {
			return
		}

		if node.Type == stdhtml.ElementNode {
			if (node.Data == "section" || node.Data == "div") && (hasClass(node, "footnotes") || hasAttr(node, "data-footnotes") || strings.EqualFold(getAttr(node, "role"), "doc-endnotes")) {
				section = node
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
			if section != nil {
				return
			}
		}
	}

	walk(root)
	return section
}

func enhanceHTMLFootnotes(input []byte) ([]byte, error) {
	container := &stdhtml.Node{Type: stdhtml.ElementNode, DataAtom: htmlatom.Div, Data: "div"}

	fragments, err := stdhtml.ParseFragment(bytes.NewReader(input), container)
	if err != nil {
		return nil, err
	}
	for _, node := range fragments {
		container.AppendChild(node)
	}

	footnotes, section := collectHTMLFootnoteDefinitions(container)
	if len(footnotes) == 0 {
		return input, nil
	}

	attached := attachHTMLFootnotes(container, footnotes)
	if attached == 0 {
		return input, nil
	}

	if section != nil {
		removeHTMLNode(section)
	}

	var buf bytes.Buffer
	for child := container.FirstChild; child != nil; child = child.NextSibling {
		if err := stdhtml.Render(&buf, child); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func collectHTMLFootnoteDefinitions(root *stdhtml.Node) (map[string][]*stdhtml.Node, *stdhtml.Node) {
	footnotes := make(map[string][]*stdhtml.Node)

	section := findFootnoteSection(root)
	if section == nil {
		return footnotes, nil
	}

	var list *stdhtml.Node
	for child := section.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == stdhtml.ElementNode && child.Data == "ol" {
			list = child
			break
		}
	}
	if list == nil {
		return footnotes, section
	}

	for item := list.FirstChild; item != nil; item = item.NextSibling {
		if item.Type != stdhtml.ElementNode || item.Data != "li" {
			continue
		}
		id := getAttr(item, "id")
		if id == "" {
			continue
		}

		clones := cloneChildrenWithoutBackrefs(item)
		clones = trimWhitespaceNodes(clones)
		if len(clones) == 0 {
			clones = []*stdhtml.Node{{Type: stdhtml.TextNode, Data: ""}}
		}

		footnotes[id] = clones
	}

	return footnotes, section
}

func cloneChildrenWithoutBackrefs(node *stdhtml.Node) []*stdhtml.Node {
	var clones []*stdhtml.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		clone := cloneNode(child)
		stripIDs(clone)
		stripFootnoteBackrefs(clone)
		clones = append(clones, clone)
	}
	return clones
}

func cloneNode(node *stdhtml.Node) *stdhtml.Node {
	if node == nil {
		return nil
	}

	clone := &stdhtml.Node{
		Type:      node.Type,
		DataAtom:  node.DataAtom,
		Data:      node.Data,
		Namespace: node.Namespace,
		Attr:      append([]stdhtml.Attribute(nil), node.Attr...),
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		clone.AppendChild(cloneNode(child))
	}

	return clone
}

func stripIDs(node *stdhtml.Node) {
	if node == nil {
		return
	}
	removeAttr(node, "id")
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		stripIDs(child)
	}
}

func stripFootnoteBackrefs(node *stdhtml.Node) {
	if node == nil {
		return
	}

	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == stdhtml.ElementNode && hasClass(child, "footnote-backref") {
			removeHTMLNode(child)
		} else {
			stripFootnoteBackrefs(child)
		}
		child = next
	}
}

func trimWhitespaceNodes(nodes []*stdhtml.Node) []*stdhtml.Node {
	for len(nodes) > 0 && isWhitespaceNode(nodes[0]) {
		nodes = nodes[1:]
	}
	for len(nodes) > 0 && isWhitespaceNode(nodes[len(nodes)-1]) {
		nodes = nodes[:len(nodes)-1]
	}
	return nodes
}

func isWhitespaceNode(node *stdhtml.Node) bool {
	if node == nil {
		return false
	}
	if node.Type == stdhtml.TextNode {
		return strings.TrimSpace(node.Data) == ""
	}
	return false
}

func hasClass(node *stdhtml.Node, class string) bool {
	if node == nil {
		return false
	}
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			for _, part := range strings.Fields(attr.Val) {
				if part == class {
					return true
				}
			}
		}
	}
	return false
}

func hasAttr(node *stdhtml.Node, key string) bool {
	if node == nil {
		return false
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

func getAttr(node *stdhtml.Node, key string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func removeAttr(node *stdhtml.Node, key string) {
	if node == nil {
		return
	}
	out := node.Attr[:0]
	for _, attr := range node.Attr {
		if attr.Key != key {
			out = append(out, attr)
		}
	}
	node.Attr = out
}

func setDataAttribute(node *stdhtml.Node, key, value string) {
	if node == nil {
		return
	}
	attrKey := key
	if !strings.HasPrefix(attrKey, "data-") {
		attrKey = "data-" + attrKey
	}
	for i, attr := range node.Attr {
		if attr.Key == attrKey {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, stdhtml.Attribute{Key: attrKey, Val: value})
}

func insertAfter(parent, newNode, existing *stdhtml.Node) {
	if parent == nil || newNode == nil || existing == nil {
		return
	}
	if existing.NextSibling != nil {
		parent.InsertBefore(newNode, existing.NextSibling)
	} else {
		parent.AppendChild(newNode)
	}
}

func removeHTMLNode(node *stdhtml.Node) {
	if node == nil || node.Parent == nil {
		return
	}
	parent := node.Parent
	if node.PrevSibling != nil {
		node.PrevSibling.NextSibling = node.NextSibling
	} else {
		parent.FirstChild = node.NextSibling
	}
	if node.NextSibling != nil {
		node.NextSibling.PrevSibling = node.PrevSibling
	} else {
		parent.LastChild = node.PrevSibling
	}
	node.Parent = nil
	node.PrevSibling = nil
	node.NextSibling = nil
}

func buildInlineFootnoteContent(nodes []*stdhtml.Node) ([]*stdhtml.Node, bool) {
	if len(nodes) == 0 {
		return nil, true
	}

	var result []*stdhtml.Node
	inline := true

	for _, node := range nodes {
		if node == nil {
			continue
		}

		if node.Type == stdhtml.ElementNode && node.Data == "p" {
			children := collectClonedChildren(node)
			childContent, childInline := buildInlineFootnoteContent(children)
			if len(childContent) == 0 {
				continue
			}

			if len(result) > 0 {
				result = append(result, newLineBreak())
			}

			result = append(result, childContent...)
			inline = inline && childInline
			continue
		}

		cloned := cloneNode(node)
		if node.Type == stdhtml.ElementNode && isBlockLevelElement(node.Data) {
			inline = false
		}

		result = append(result, cloned)

		if node.Type == stdhtml.TextNode && strings.Contains(node.Data, "\n") {
			inline = false
		}
	}

	return trimWhitespaceNodes(result), inline && !containsBlockLevelNode(result)
}

func collectClonedChildren(node *stdhtml.Node) []*stdhtml.Node {
	var children []*stdhtml.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, cloneNode(child))
	}
	return children
}

func containsBlockLevelNode(nodes []*stdhtml.Node) bool {
	for _, node := range nodes {
		if node != nil && node.Type == stdhtml.ElementNode && isBlockLevelElement(node.Data) {
			return true
		}
	}
	return false
}

func newLineBreak() *stdhtml.Node {
	return &stdhtml.Node{
		Type:     stdhtml.ElementNode,
		DataAtom: htmlatom.Br,
		Data:     "br",
	}
}

func isBlockLevelElement(name string) bool {
	switch strings.ToLower(name) {
	case "address", "article", "aside", "blockquote", "canvas", "dd", "div", "dl", "dt",
		"fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6",
		"header", "hr", "li", "main", "nav", "noscript", "ol", "output", "p", "pre", "section",
		"table", "tfoot", "ul", "video":
		return true
	default:
		return false
	}
}
