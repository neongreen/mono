package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// sentenceBoundaryRegex matches sentence boundaries (. ! ?) followed by space or end of string
var sentenceBoundaryRegex = regexp.MustCompile(`([.!?])(\s+)`)

// formatMarkdown formats markdown content with one sentence per line
func formatMarkdown(input []byte) ([]byte, error) {
	parser := goldmark.DefaultParser()
	reader := text.NewReader(input)
	doc := parser.Parse(reader)

	var buf bytes.Buffer
	err := walkAndFormat(doc, input, &buf, 0)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// walkAndFormat walks the AST and formats it
func walkAndFormat(node ast.Node, source []byte, w io.Writer, depth int) error {
	// Handle different node types
	switch n := node.(type) {
	case *ast.Document:
		// Process all children
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, w, depth); err != nil {
				return err
			}
		}
		return nil

	case *ast.Heading:
		// Write heading prefix
		for i := 0; i < n.Level; i++ {
			w.Write([]byte("#"))
		}
		w.Write([]byte(" "))

		// Write heading content
		return writeInlineContent(n, source, w, false)

	case *ast.Paragraph:
		// For paragraphs, format with one sentence per line
		return writeInlineContent(n, source, w, true)

	case *ast.List:
		// Process list items
		itemNum := n.Start
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if listItem, ok := child.(*ast.ListItem); ok {
				// Write list marker
				if n.IsOrdered() {
					fmt.Fprintf(w, "%d. ", itemNum)
					itemNum++
				} else {
					w.Write([]byte("- "))
				}

				// Write list item content
				if err := writeListItemContent(listItem, source, w); err != nil {
					return err
				}
			}
		}
		// Add blank line after list
		if n.NextSibling() != nil {
			w.Write([]byte("\n"))
		}
		return nil

	case *ast.ListItem:
		// This should be handled by the List case above, but just in case
		return nil

	case *ast.TextBlock:
		// TextBlock is used in lists, should be handled by writeListItemContent
		// But if we encounter it directly, treat it like a paragraph
		return writeInlineContent(n, source, w, true)

	case *ast.FencedCodeBlock:
		// Write fenced code block as-is
		w.Write([]byte("```"))
		if n.Language(source) != nil {
			w.Write(n.Language(source))
		}
		w.Write([]byte("\n"))

		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			w.Write(line.Value(source))
		}
		w.Write([]byte("```\n\n"))
		return nil

	case *ast.CodeBlock:
		// Write indented code block
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			w.Write([]byte("    "))
			w.Write(line.Value(source))
		}
		w.Write([]byte("\n"))
		return nil

	case *ast.Blockquote:
		// Process blockquote content with prefix
		var buf bytes.Buffer
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, &buf, depth); err != nil {
				return err
			}
		}

		// Add > prefix to each line
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		for _, line := range lines {
			w.Write([]byte("> "))
			w.Write([]byte(line))
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\n"))
		return nil

	case *ast.ThematicBreak:
		w.Write([]byte("---\n\n"))
		return nil

	default:
		// For other block nodes, process children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, w, depth); err != nil {
				return err
			}
		}
		return nil
	}
}

// writeListItemContent writes the content of a list item with proper sentence splitting
func writeListItemContent(item *ast.ListItem, source []byte, w io.Writer) error {
	var buf bytes.Buffer

	// Collect all content from the list item
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.(type) {
		case *ast.Paragraph, *ast.TextBlock:
			// For paragraphs and text blocks in list items, collect inline text
			if err := collectInlineText(child, source, &buf); err != nil {
				return err
			}
		default:
			// For other types (nested lists, code blocks, etc.), format them separately
			// This is simplified - in a full implementation we'd handle indentation
			if err := walkAndFormat(child, source, &buf, 1); err != nil {
				return err
			}
		}
	}

	text := strings.TrimSpace(buf.String())

	// Split into sentences
	sentences := splitIntoSentences(text)

	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			if i > 0 {
				w.Write([]byte("  ")) // Indent continuation lines
			}
			w.Write([]byte(sentence))
			w.Write([]byte("\n"))
		}
	}

	return nil
}

// writeInlineContent writes inline content (text, emphasis, links, etc.)
func writeInlineContent(node ast.Node, source []byte, w io.Writer, splitSentences bool) error {
	var buf bytes.Buffer

	// Collect all inline content
	if err := collectInlineText(node, source, &buf); err != nil {
		return err
	}

	text := buf.String()

	if splitSentences {
		// Split into sentences
		sentences := splitIntoSentences(text)
		for i, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence != "" {
				w.Write([]byte(sentence))
				w.Write([]byte("\n"))
			}
			// Add blank line after last sentence if there's a next sibling
			if i == len(sentences)-1 && node.NextSibling() != nil {
				w.Write([]byte("\n"))
			}
		}
	} else {
		// Write as-is with newline
		w.Write([]byte(text))
		w.Write([]byte("\n"))
		if node.NextSibling() != nil {
			w.Write([]byte("\n"))
		}
	}

	return nil
}

// collectInlineText recursively collects inline text content
func collectInlineText(node ast.Node, source []byte, buf *bytes.Buffer) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			segment := n.Segment
			buf.Write(segment.Value(source))
			if n.SoftLineBreak() {
				buf.Write([]byte(" "))
			}
		case *ast.String:
			buf.Write(n.Value)
		case *ast.CodeSpan:
			buf.Write([]byte("`"))
			collectInlineText(child, source, buf)
			buf.Write([]byte("`"))
		case *ast.Emphasis:
			level := n.Level
			for i := 0; i < level; i++ {
				buf.Write([]byte("*"))
			}
			collectInlineText(child, source, buf)
			for i := 0; i < level; i++ {
				buf.Write([]byte("*"))
			}
		case *ast.Link:
			buf.Write([]byte("["))
			collectInlineText(child, source, buf)
			buf.Write([]byte("]("))
			buf.Write(n.Destination)
			if n.Title != nil {
				buf.Write([]byte(` "`))
				buf.Write(n.Title)
				buf.Write([]byte(`"`))
			}
			buf.Write([]byte(")"))
		case *ast.Image:
			buf.Write([]byte("!["))
			collectInlineText(child, source, buf)
			buf.Write([]byte("]("))
			buf.Write(n.Destination)
			if n.Title != nil {
				buf.Write([]byte(` "`))
				buf.Write(n.Title)
				buf.Write([]byte(`"`))
			}
			buf.Write([]byte(")"))
		case *ast.AutoLink:
			buf.Write([]byte("<"))
			buf.Write(n.URL(source))
			buf.Write([]byte(">"))
		default:
			// Recursively handle other inline nodes
			if child.HasChildren() {
				collectInlineText(child, source, buf)
			}
		}
	}
	return nil
}

// splitIntoSentences splits text into sentences
func splitIntoSentences(text string) []string {
	// Replace sentence boundaries with a special marker
	text = sentenceBoundaryRegex.ReplaceAllString(text, "$1\n\n")

	// Split by the marker
	parts := strings.Split(text, "\n\n")

	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			sentences = append(sentences, part)
		}
	}

	return sentences
}

func main() {
	writeFlag := flag.Bool("w", false, "write result to source file instead of stdout")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: markdown-format [-w] <file> [<file> ...]")
		fmt.Fprintln(os.Stderr, "       markdown-format -")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  -w    write result to source file(s) instead of stdout")
		os.Exit(1)
	}

	// Handle stdin
	if args[0] == "-" {
		if *writeFlag {
			fmt.Fprintln(os.Stderr, "Error: -w flag cannot be used with stdin")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Error: cannot specify files when reading from stdin")
			os.Exit(1)
		}
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
		return
	}

	// Handle file(s)
	if *writeFlag {
		// Format multiple files in place
		for _, filename := range args {
			input, err := os.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
				os.Exit(1)
			}
			output, err := formatMarkdown(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", filename, err)
				os.Exit(1)
			}
			err = os.WriteFile(filename, output, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", filename, err)
				os.Exit(1)
			}
		}
	} else {
		// Format single file to stdout
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Error: can only format one file to stdout (or use -w for multiple files)")
			os.Exit(1)
		}
		input, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		output, err := formatMarkdown(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
	}
}
