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
	if !strings.Contains(result, "<span class=\"printpdf-footnote\"") {
		t.Fatalf("expected inline footnote span, got: %s", result)
	}
	if strings.Contains(result, "<div class=\"printpdf-footnote\"") {
		t.Fatalf("unexpected block-level footnote container in output: %s", result)
	}
	if strings.Contains(result, "footnote-backref") {
		t.Fatalf("unexpected footnote backref found in output: %s", result)
	}
}
