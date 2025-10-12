package converter

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// convertMarkdownToTypst converts Markdown content to Typst markup
func convertMarkdownToTypst(markdown []byte) (string, error) {
	// Parse the markdown
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	reader := text.NewReader(markdown)
	doc := md.Parser().Parse(reader)

	// Convert AST to Typst
	var buf bytes.Buffer

	// Write Typst document preamble
	buf.WriteString("#set page(paper: \"a4\", margin: (x: 2cm, y: 2cm))\n")
	buf.WriteString("#set text(font: \"Linux Libertine\", size: 11pt)\n")
	buf.WriteString("#set par(justify: false, leading: 0.65em)\n")
	buf.WriteString("#set heading(numbering: none)\n")
	buf.WriteString("\n")

	// Convert the document
	if err := renderNodeToTypst(&buf, doc, markdown); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderNodeToTypst recursively renders AST nodes to Typst markup
func renderNodeToTypst(buf *bytes.Buffer, node ast.Node, source []byte) error {
	switch n := node.(type) {
	case *ast.Document:
		return renderChildren(buf, n, source)

	case *ast.Heading:
		level := n.Level
		buf.WriteString(strings.Repeat("=", level))
		buf.WriteString(" ")
		if err := renderChildren(buf, n, source); err != nil {
			return err
		}
		buf.WriteString("\n\n")

	case *ast.Paragraph:
		if err := renderChildren(buf, n, source); err != nil {
			return err
		}
		buf.WriteString("\n\n")

	case *ast.Text:
		// Handle text content
		text := n.Text(source)
		// Escape special Typst characters
		escaped := escapeTypst(string(text))
		buf.WriteString(escaped)

		// Handle soft/hard line breaks
		if n.SoftLineBreak() {
			buf.WriteString(" ")
		} else if n.HardLineBreak() {
			buf.WriteString("\\\n")
		}

	case *ast.String:
		buf.WriteString(escapeTypst(string(n.Value)))

	case *ast.Emphasis:
		// Level 1 = italic (_text_), Level 2 = bold (*text*)
		if n.Level == 1 {
			buf.WriteString("_")
			if err := renderChildren(buf, n, source); err != nil {
				return err
			}
			buf.WriteString("_")
		} else if n.Level == 2 {
			buf.WriteString("*")
			if err := renderChildren(buf, n, source); err != nil {
				return err
			}
			buf.WriteString("*")
		} else {
			// Fallback for other levels
			if err := renderChildren(buf, n, source); err != nil {
				return err
			}
		}

	case *ast.Link:
		buf.WriteString("#link(\"")
		buf.WriteString(escapeTypst(string(n.Destination)))
		buf.WriteString("\")[")
		if err := renderChildren(buf, n, source); err != nil {
			return err
		}
		buf.WriteString("]")

	case *ast.CodeSpan:
		buf.WriteString("`")
		buf.WriteString(string(n.Text(source)))
		buf.WriteString("`")

	case *ast.FencedCodeBlock:
		lang := string(n.Language(source))
		buf.WriteString("```")
		if lang != "" {
			buf.WriteString(lang)
		}
		buf.WriteString("\n")

		// Get code content
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(source))
		}
		buf.WriteString("```\n\n")

	case *ast.CodeBlock:
		buf.WriteString("```\n")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(source))
		}
		buf.WriteString("```\n\n")

	case *ast.List:
		if err := renderList(buf, n, source, 0); err != nil {
			return err
		}
		buf.WriteString("\n")

	case *ast.ListItem:
		// Handled by renderList
		return renderChildren(buf, n, source)

	case *ast.Blockquote:
		// Typst doesn't have native blockquotes, use a styled block
		buf.WriteString("#block(inset: (left: 1em), stroke: (left: 2pt + gray))[\n")
		if err := renderChildren(buf, n, source); err != nil {
			return err
		}
		buf.WriteString("]\n\n")

	case *ast.ThematicBreak:
		buf.WriteString("#line(length: 100%)\n\n")

	case *ast.HTMLBlock:
		// Skip HTML blocks in Typst
		buf.WriteString("// HTML block omitted\n\n")

	case *ast.RawHTML:
		// Skip raw HTML
		return nil

	case *ast.Image:
		buf.WriteString("#image(\"")
		buf.WriteString(escapeTypst(string(n.Destination)))
		buf.WriteString("\"")
		if n.Title != nil {
			buf.WriteString(", alt: \"")
			buf.WriteString(escapeTypst(string(n.Title)))
			buf.WriteString("\"")
		}
		buf.WriteString(")\n\n")

	case *extast.Table:
		if err := renderTable(buf, n, source); err != nil {
			return err
		}

	case *extast.TableHeader:
		// Handled by renderTable
		return nil

	case *extast.TableRow:
		// Handled by renderTable
		return nil

	case *extast.TableCell:
		// Handled by renderTable
		return nil

	case *extast.Strikethrough:
		buf.WriteString("#strike[")
		if err := renderChildren(buf, n, source); err != nil {
			return err
		}
		buf.WriteString("]")

	default:
		// For unknown nodes, try to render children
		return renderChildren(buf, n, source)
	}

	return nil
}

// renderChildren renders all children of a node
func renderChildren(buf *bytes.Buffer, node ast.Node, source []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if err := renderNodeToTypst(buf, child, source); err != nil {
			return err
		}
	}
	return nil
}

// renderList renders a list (ordered or unordered)
func renderList(buf *bytes.Buffer, list *ast.List, source []byte, depth int) error {
	indent := strings.Repeat("  ", depth)

	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		listItem, ok := item.(*ast.ListItem)
		if !ok {
			continue
		}

		buf.WriteString(indent)
		if list.IsOrdered() {
			buf.WriteString("+ ")
		} else {
			buf.WriteString("- ")
		}

		// Render item content
		var itemBuf bytes.Buffer
		for child := listItem.FirstChild(); child != nil; child = child.NextSibling() {
			if childList, ok := child.(*ast.List); ok {
				// Nested list
				itemBuf.WriteString("\n")
				if err := renderList(&itemBuf, childList, source, depth+1); err != nil {
					return err
				}
			} else if para, ok := child.(*ast.Paragraph); ok {
				// Paragraph content - render inline
				for pChild := para.FirstChild(); pChild != nil; pChild = pChild.NextSibling() {
					if err := renderNodeToTypst(&itemBuf, pChild, source); err != nil {
						return err
					}
				}
			} else {
				if err := renderNodeToTypst(&itemBuf, child, source); err != nil {
					return err
				}
			}
		}

		// Write the item content, removing trailing newlines
		content := strings.TrimRight(itemBuf.String(), "\n")
		buf.WriteString(content)
		buf.WriteString("\n")
	}

	return nil
}

// renderTable renders a GFM table to Typst
func renderTable(buf *bytes.Buffer, table *extast.Table, source []byte) error {
	// Count columns
	var colCount int
	if table.FirstChild() != nil {
		if header, ok := table.FirstChild().(*extast.TableHeader); ok && header.FirstChild() != nil {
			if row, ok := header.FirstChild().(*extast.TableRow); ok {
				for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
					colCount++
				}
			}
		}
	}

	if colCount == 0 {
		return nil // Empty table
	}

	// Start table with column spec
	buf.WriteString("#table(\n")
	buf.WriteString("  columns: ")
	buf.WriteString(strings.Repeat("auto, ", colCount))
	buf.WriteString("\n")
	buf.WriteString("  stroke: 0.5pt,\n")
	buf.WriteString("  inset: 8pt,\n")

	// Render header
	if table.FirstChild() != nil {
		if header, ok := table.FirstChild().(*extast.TableHeader); ok {
			if row, ok := header.FirstChild().(*extast.TableRow); ok {
				for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
					buf.WriteString("  [*")
					var cellBuf bytes.Buffer
					if err := renderChildren(&cellBuf, cell, source); err != nil {
						return err
					}
					buf.WriteString(strings.TrimSpace(cellBuf.String()))
					buf.WriteString("*],\n")
				}
			}
		}
	}

	// Render body rows
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		if row, ok := child.(*extast.TableRow); ok {
			for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
				buf.WriteString("  [")
				var cellBuf bytes.Buffer
				if err := renderChildren(&cellBuf, cell, source); err != nil {
					return err
				}
				buf.WriteString(strings.TrimSpace(cellBuf.String()))
				buf.WriteString("],\n")
			}
		}
	}

	buf.WriteString(")\n\n")
	return nil
}

// escapeTypst escapes special characters for Typst
func escapeTypst(s string) string {
	// Typst special characters that need escaping in text
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"#", "\\#",
		"$", "\\$",
		"[", "\\[",
		"]", "\\]",
		"<", "\\<",
		">", "\\>",
		"@", "\\@",
	)
	return replacer.Replace(s)
}
