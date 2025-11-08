package postlight

import (
	"context"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	// Skip if WASM file doesn't exist
	// This test requires the WASM module to be built first
	t.Skip("Requires WASM module to be built with 'make build-wasm'")

	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	// Sample HTML for testing
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
		<meta name="author" content="Test Author">
	</head>
	<body>
		<article>
			<h1>Test Article Title</h1>
			<p>This is a test article with some content.</p>
			<p>It has multiple paragraphs to test the parser.</p>
		</article>
	</body>
	</html>
	`

	article, err := parser.Parse(ctx, "https://example.com/test", html)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if article.Title == "" {
		t.Error("Expected title to be extracted")
	}

	if article.Content == "" {
		t.Error("Expected content to be extracted")
	}

	t.Logf("Parsed article: %+v", article)
}

func TestParser_ParseURL(t *testing.T) {
	// This test requires network access and the WASM module
	t.Skip("Requires WASM module and network access")

	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	// Use a known good article URL for testing
	// You can replace this with any article URL
	url := "https://example.com"

	article, err := parser.ParseURL(ctx, url)
	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if article == nil {
		t.Fatal("Expected article to be returned")
	}

	t.Logf("Parsed article: %+v", article)
}
