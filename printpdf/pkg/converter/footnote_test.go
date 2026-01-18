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
	if !strings.Contains(result, "<span class=\"printpdf-footnote\"") {
		t.Fatalf("expected inline footnote span, got: %s", result)
	}
	if strings.Contains(result, "<div class=\"printpdf-footnote\"") {
		t.Fatalf("unexpected block-level footnote container in output: %s", result)
	}
	if strings.Contains(result, "footnote-backref") {
		t.Fatalf("unexpected footnote backref found in output: %s", result)
	}
	if !strings.Contains(result, "sup id=\"fnref:1\"") {
		t.Fatalf("expected footnote superscript, got: %s", result)
	}
	if !strings.Contains(result, "data-footnote-id=\"fn:1\"") {
		t.Fatalf("expected data-footnote-id on superscript, got: %s", result)
	}
	// The superscript should be empty - Prince generates the number via ::footnote-call
	if !strings.Contains(result, "></sup>") {
		t.Fatalf("expected empty superscript (Prince generates number), got: %s", result)
	}
	if strings.Contains(result, "class=\"footnote-ref\"") {
		t.Fatalf("unexpected footnote anchor found in output: %s", result)
	}
	if !strings.Contains(result, "::footnote-call") {
		t.Fatalf("expected ::footnote-call CSS rule for Prince, got: %s", result)
	}
	if !strings.Contains(result, "span.printpdf-footnote::footnote-marker") {
		t.Fatalf("expected inline footnote marker override in CSS, got: %s", result)
	}
}
