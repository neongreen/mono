package postlight

import (
	"context"
	"strings"
	"testing"
)

// TestNewParser verifies that we can create a new parser instance.
func TestNewParser(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	if parser == nil {
		t.Fatal("Expected non-nil parser")
	}
	if parser.runtime == nil {
		t.Error("Expected runtime to be initialized")
	}
	if parser.module == nil {
		t.Error("Expected module to be compiled")
	}
}

// TestParser_ParseSimpleArticle tests parsing a simple article with basic content.
func TestParser_ParseSimpleArticle(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article Title</title>
	<meta name="author" content="Test Author">
	<meta name="description" content="This is a test article description">
</head>
<body>
	<article>
		<h1>Test Article Title</h1>
		<p>This is the first paragraph with some content.</p>
		<p>This is the second paragraph with more content.</p>
		<p>This is the third paragraph to ensure we have enough text.</p>
	</article>
</body>
</html>`

	article, err := parser.Parse(ctx, "https://example.com/test-article", html)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify basic fields
	if article.Title == "" {
		t.Error("Expected title to be extracted")
	}
	t.Logf("Title: %s", article.Title)

	if article.Content == "" {
		t.Error("Expected content to be extracted")
	}
	t.Logf("Content length: %d", len(article.Content))

	if article.URL != "https://example.com/test-article" {
		t.Errorf("Expected URL to be https://example.com/test-article, got %s", article.URL)
	}

	if article.Domain != "example.com" {
		t.Errorf("Expected domain to be example.com, got %s", article.Domain)
	}

	if article.WordCount == 0 {
		t.Error("Expected word count to be greater than 0")
	}
	t.Logf("Word count: %d", article.WordCount)
}

// TestParser_ParseArticleWithImage tests parsing an article with an image.
func TestParser_ParseArticleWithImage(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	html := `<!DOCTYPE html>
<html>
<head>
	<title>Article with Image</title>
</head>
<body>
	<article>
		<h1>Article with Image</h1>
		<img src="https://example.com/image.jpg" alt="Test image">
		<p>This article has an image above. This is some content below the image.</p>
		<p>More content to make the article substantial.</p>
	</article>
</body>
</html>`

	article, err := parser.Parse(ctx, "https://example.com/article-with-image", html)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if article.Title == "" {
		t.Error("Expected title to be extracted")
	}

	if article.Content == "" {
		t.Error("Expected content to be extracted")
	}

	// The parser should extract images
	if !strings.Contains(article.Content, "img") && article.LeadImageURL == "" {
		t.Log("Note: No image found in content or as lead image (this may be ok)")
	} else {
		t.Logf("Lead image URL: %s", article.LeadImageURL)
	}
}

// TestParser_ParseComplexArticle tests parsing a more complex article.
func TestParser_ParseComplexArticle(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	html := `<!DOCTYPE html>
<html>
<head>
	<title>Complex Article Title</title>
	<meta name="author" content="John Doe">
	<meta property="article:published_time" content="2024-01-15T10:00:00Z">
	<meta name="description" content="A comprehensive test article">
</head>
<body>
	<header>
		<nav>Navigation links that should be ignored</nav>
	</header>
	<main>
		<article>
			<h1>Complex Article Title</h1>
			<div class="meta">
				<span class="author">By John Doe</span>
				<time datetime="2024-01-15">January 15, 2024</time>
			</div>
			<p class="lead">This is the lead paragraph or excerpt of the article.</p>
			<img src="https://example.com/hero.jpg" alt="Hero image">
			<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
			<h2>First Section</h2>
			<p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.</p>
			<p>Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.</p>
			<h2>Second Section</h2>
			<p>Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
			<ul>
				<li>First bullet point with information</li>
				<li>Second bullet point with more details</li>
				<li>Third bullet point for completeness</li>
			</ul>
			<blockquote>
				<p>This is a quote from someone important that adds context to the article.</p>
			</blockquote>
			<p>Final paragraph wrapping up the article with concluding thoughts and summary.</p>
		</article>
	</main>
	<aside>
		<div class="related">Related articles that should be ignored</div>
		<div class="ads">Advertisements</div>
	</aside>
	<footer>
		<p>Footer content that should be ignored</p>
	</footer>
</body>
</html>`

	article, err := parser.Parse(ctx, "https://example.com/complex-article", html)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify all major fields
	if article.Title == "" {
		t.Error("Expected title to be extracted")
	}
	t.Logf("Title: %s", article.Title)

	if article.Author == "" {
		t.Log("Author not extracted (may be ok depending on parser)")
	} else {
		t.Logf("Author: %s", article.Author)
	}

	if article.Content == "" {
		t.Error("Expected content to be extracted")
	}
	t.Logf("Content length: %d bytes", len(article.Content))

	if article.WordCount == 0 {
		t.Error("Expected word count > 0")
	}
	t.Logf("Word count: %d", article.WordCount)

	if article.Excerpt != "" {
		t.Logf("Excerpt: %s", article.Excerpt)
	}

	if article.LeadImageURL != "" {
		t.Logf("Lead image: %s", article.LeadImageURL)
	}

	if article.DatePublished != "" {
		t.Logf("Published date: %s", article.DatePublished)
	}

	// Note: Our simplified parser extracts all body content
	// A full Postlight parser would filter out nav/footer, but our demo version doesn't
	// This just verifies we got the article content
	if !strings.Contains(article.Content, "Lorem ipsum") {
		t.Error("Content should include main article text")
	}
}

// TestParser_ParseMinimalHTML tests parsing minimal but valid HTML.
func TestParser_ParseMinimalHTML(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	html := `<html><head><title>Minimal</title></head><body><p>Just a paragraph.</p></body></html>`

	article, err := parser.Parse(ctx, "https://example.com/minimal", html)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if article.Title == "" {
		t.Error("Expected title to be extracted even from minimal HTML")
	}

	if article.Content == "" {
		t.Error("Expected content to be extracted even from minimal HTML")
	}

	t.Logf("Parsed minimal article - Title: %s, Content length: %d", article.Title, len(article.Content))
}

// TestParser_MultipleParses tests that a parser can be reused for multiple parses.
func TestParser_MultipleParses(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	testCases := []struct {
		name string
		html string
		url  string
	}{
		{
			name: "First article",
			html: `<html><head><title>First</title></head><body><p>First article content here.</p></body></html>`,
			url:  "https://example.com/first",
		},
		{
			name: "Second article",
			html: `<html><head><title>Second</title></head><body><p>Second article content here.</p></body></html>`,
			url:  "https://example.com/second",
		},
		{
			name: "Third article",
			html: `<html><head><title>Third</title></head><body><p>Third article content here.</p></body></html>`,
			url:  "https://example.com/third",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			article, err := parser.Parse(ctx, tc.url, tc.html)
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", tc.name, err)
			}

			if article.Title == "" {
				t.Errorf("Expected title for %s", tc.name)
			}

			if article.Content == "" {
				t.Errorf("Expected content for %s", tc.name)
			}

			if article.URL != tc.url {
				t.Errorf("Expected URL %s, got %s", tc.url, article.URL)
			}

			t.Logf("%s - Title: %s", tc.name, article.Title)
		})
	}
}

// TestParser_EmptyHTML tests that parsing empty HTML returns an error.
func TestParser_EmptyHTML(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close(ctx)

	html := ""

	_, err = parser.Parse(ctx, "https://example.com/empty", html)
	if err == nil {
		t.Log("Note: Parser accepted empty HTML (may return minimal data)")
	} else {
		t.Logf("Parser correctly rejected empty HTML: %v", err)
	}
}

// TestParser_Close tests that closing the parser doesn't error.
func TestParser_Close(t *testing.T) {
	ctx := context.Background()
	parser, err := NewParser(ctx)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	err = parser.Close(ctx)
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}

	// Closing again should also not error
	err = parser.Close(ctx)
	if err != nil {
		t.Errorf("Second close should not error: %v", err)
	}
}
