package converter

import (
	"strings"
	"testing"
)

func TestHTMLFootnoteRendering(t *testing.T) {
	markdown := []byte("Footnote note[^1].\n\n[^1]: Body\n")

	html, err := convertMarkdownToHTML(markdown, PageOptions{})
	if err != nil {
		t.Fatalf("convertMarkdownToHTML failed: %v", err)
	}

	result := string(html)
	if !strings.Contains(result, "class=\"footnote-ref\"") {
		t.Fatalf("expected footnote reference anchor, got: %s", result)
	}
	if !strings.Contains(result, "printpdf-footnote") {
		t.Fatalf("expected inline footnote span, got: %s", result)
	}
}
