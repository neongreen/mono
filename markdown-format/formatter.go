package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func formatMarkdown(input []byte) ( // formatMarkdown formats markdown content with one sentence per line
	[]byte, error) {
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

func walkAndFormat(node ast.Node, source []byte, w io.Writer, depth int) error {
	switch n := node.(type) {
	case *ast.Document:
		for child := n.
			FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, w, depth); err != nil {
				return err
			}
		}
		return nil
	case *ast.Heading:
		for i := 0; i < n.Level; i++ {
			w.Write([]byte("#"))
		}
		w.Write([]byte(" "))
		return writeInlineContent(n, source, w, false)
	case *ast.Paragraph:
		return writeInlineContent(n, source, w, true)
	case *ast.List:
		itemNum := n.Start
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if listItem, ok := child.(*ast.ListItem); ok {
				if n.IsOrdered() {
					fmt.Fprintf(w, "%d%c ", itemNum, n.Marker)
					itemNum++
				} else {
					fmt.Fprintf(w, "%c ", n.Marker)
				}
				if err := writeListItemContent(listItem, source, w); err != nil {
					return err
				}
			}
		}
		if n.NextSibling() != nil {
			w.Write([]byte("\n"))
		}
		return nil
	case *ast.ListItem:
		return nil
	case *ast.TextBlock:
		return writeInlineContent(n, source, w, true)
	case *ast.FencedCodeBlock:
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
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			w.Write([]byte("    "))
			w.Write(line.Value(source))
		}
		w.Write([]byte("\n"))
		return nil
	case *ast.Blockquote:
		var buf bytes.Buffer
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, &buf, depth); err != nil {
				return err
			}
		}
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
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if err := walkAndFormat(child, source, w, depth); err != nil {
				return err
			}
		}
		return nil
	}
}

func writeListItemContent(item *ast.ListItem, source []byte, w io.Writer) error {
	var textBuf bytes.Buffer
	var hasText bool

	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.(type) {
		case *ast.
			Paragraph, *ast.TextBlock:

			if err := collectInlineText(child,
				source, &textBuf); err !=
				nil {
				return err
			}
			hasText = true
		}
	}
	if hasText {
		text := strings.TrimSpace(
			textBuf.String())
		sentences := splitIntoSentences(text)
		for i, sentence := range sentences {
			sentence = strings.
				TrimSpace(sentence)
			if sentence != "" {
				if i > 0 {
					w.Write([]byte("  "))
				}
				w.Write([]byte(sentence))
				w.Write([]byte("\n"))
			}
		}
	}
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.(type) {
		case *ast.Paragraph,
			*ast.TextBlock:
			continue
		default:
			var nestedBuf bytes.
				Buffer

			if err := walkAndFormat(child,
				source, &nestedBuf, 1,
			); err !=
				nil {
				return err
			}

			nestedContent := nestedBuf.String()
			if strings.TrimSpace(nestedContent) != "" {
				lines := strings.SplitSeq(strings.
					TrimRight(nestedContent, "\n"),
					"\n")
				for line := range lines {
					w.Write([]byte("  "))
					w.Write(
						[]byte(line))
					w.Write([]byte("\n"))
				}
			}
		}
	}
	return nil
}

func writeInlineContent(node ast.Node, source []byte, w io.Writer, splitSentences bool) error {
	var buf bytes.Buffer
	if err := collectInlineText(node, source,
		&buf); err != nil {
		return err
	}
	text :=
		buf.String()
	if splitSentences {

		sentences := splitIntoSentences(text)
		for i, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence !=
				"" {
				w.Write(
					[]byte(sentence))
				w.Write([]byte("\n"))
			}
			if i == len(sentences)-1 && node.NextSibling() !=
				nil {
				w.
					Write([]byte("\n"))
			}
		}
	} else {
		w.Write([]byte(text))
		w.Write(
			[]byte("\n"))
		if node.
			NextSibling() != nil {
			w.Write([]byte("\n"))
		}
	}
	return nil
}

func collectInlineText(
	node ast.Node, source []byte, buf *bytes.Buffer) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			segment :=
				n.Segment
			buf.Write(segment.
				Value(source))
			if n.SoftLineBreak() {
				buf.Write([]byte(
					" "))
			}
		case
			*ast.String:
			buf.
				Write(n.Value)
		case *ast.
			CodeSpan:
			buf.Write([]byte("`"))
			collectInlineText(child, source, buf)
			buf.
				Write([]byte("`"))
		case
			*ast.Emphasis:
			level := n.Level
			for range level {
				buf.Write([]byte("*"))
			}
			collectInlineText(child, source, buf)
			for range level {
				buf.
					Write([]byte("*"))
			}
		case *ast.Link:
			buf.Write([]byte("["))
			collectInlineText(child, source,

				buf)
			buf.Write([]byte("]("))
			buf.Write(n.Destination)
			if n.Title !=
				nil {
				buf.Write([]byte(` "`))
				buf.Write(n.Title)
				buf.
					Write([]byte(`"`))
			}
			buf.Write(
				[]byte(")"))
		case *ast.Image:
			buf.Write([]byte("!["))
			collectInlineText(child, source, buf)
			buf.Write([]byte("]("))
			buf.Write(n.Destination)
			if n.Title != nil {
				buf.Write([]byte(` "`))
				buf.
					Write(n.Title)
				buf.Write([]byte(`"`))
			}
			buf.
				Write([]byte(")"))
		case *ast.AutoLink:
			buf.Write([]byte("<"))
			buf.Write(n.URL(source))
			buf.
				Write([]byte(">"))
		default:

			if child.HasChildren() {
				collectInlineText(child,
					source, buf)
			}
		}
	}
	return nil
}

func splitIntoSentences(
	text string) []string {

	protected := text
	replacements := make(map[string]string)
	for i, abbr := range commonAbbreviations {
		placeholder := fmt.Sprintf("\x00ABBR%d\x00",
			i)
		replacements[placeholder] = abbr
		protected = strings.ReplaceAll(protected, abbr,
			placeholder)
	}

	protected = sentenceBoundaryRegex.ReplaceAllString(protected, "$1\n\n$3")
	parts := strings.Split(protected, "\n\n")
	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			for placeholder, abbr := range replacements {
				part = strings.ReplaceAll(
					part, placeholder, abbr)

			}
			sentences = append(sentences, part)
		}
	}
	return sentences
}
